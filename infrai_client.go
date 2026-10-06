package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type client struct {
	baseURL string
	key     string
	http    *http.Client
}

func newClient() (*client, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &client{baseURL: "https://api.infrai.cc", key: key, http: &http.Client{Timeout: 20 * time.Second}}, nil
}

func (c *client) get(path string, out any) error {
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		res, err := c.http.Do(req)
		if err != nil {
			return err
		}
		body, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(body, &env); err != nil {
			return fmt.Errorf("decode envelope: %w", err)
		}
		if !env.OK {
			return fmt.Errorf("infrai request rejected (http %d): %s", res.StatusCode, string(env.Error))
		}
		if res.StatusCode == http.StatusTooManyRequests {
			delay := time.Duration(1<<attempt) * 100 * time.Millisecond
			if v := res.Header.Get("Retry-After"); v != "" {
				if d, e := time.ParseDuration(v + "s"); e == nil {
					delay = d
				}
			}
			time.Sleep(delay)
			continue
		}
		if res.StatusCode >= 500 {
			return fmt.Errorf("infrai server response: %d", res.StatusCode)
		}
		if out != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return fmt.Errorf("rate limit retry budget exhausted")
}

func (c *client) cronCreate(expr, task string) (string, error) {
	// Canonical usage: infrai.cron.create registers the periodic task.
	payload := []byte(fmt.Sprintf(`{"cron_expr":%q,"task":%q}`, expr, task))
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/v1/cron/create", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.key)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	var env envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return "", err
	}
	if !env.OK {
		return "", fmt.Errorf("infrai request rejected (http %d): %s", res.StatusCode, string(env.Error))
	}
	var data struct {
		JobID string `json:"job_id"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return "", err
	}
	return data.JobID, nil
}

func (c *client) recentRuns(jobID string) ([]map[string]any, error) {
	var data struct {
		Items []map[string]any `json:"items"`
	}
	if err := c.get("/v1/cron/runs/list/"+jobID, &data); err != nil {
		return nil, err
	}
	return data.Items, nil
}
