package main

import "time"

type BuildEvent struct {
	ID         string
	FinishedAt time.Time
	Status     string
}
type ReleaseOperation struct {
	BuildID    string
	ReleasedAt time.Time
}
type Diagnostic struct {
	BuildID string
	Reason  string
}

// staleBuilds is the compliance decision: unfinished builds older than the retention window are swept.
func staleBuilds(events []BuildEvent, now time.Time, retention time.Duration) []Diagnostic {
	var out []Diagnostic
	for _, event := range events {
		if event.Status != "succeeded" && now.Sub(event.FinishedAt) > retention {
			out = append(out, Diagnostic{BuildID: event.ID, Reason: "stale developer-tool record"})
		}
	}
	return out
}
