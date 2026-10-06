package main

import (
	"testing"
	"time"
)

func TestStaleBuildsKeepsRecentAndSuccessful(t *testing.T) {
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	events := []BuildEvent{
		{ID: "old-failed", FinishedAt: now.Add(-48 * time.Hour), Status: "failed"},
		{ID: "recent-failed", FinishedAt: now.Add(-2 * time.Hour), Status: "failed"},
		{ID: "old-success", FinishedAt: now.Add(-48 * time.Hour), Status: "succeeded"},
	}
	got := staleBuilds(events, now, 24*time.Hour)
	if len(got) != 1 || got[0].BuildID != "old-failed" {
		t.Fatalf("got %#v", got)
	}
}
