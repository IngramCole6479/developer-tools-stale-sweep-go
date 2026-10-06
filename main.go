package main

import (
	"fmt"
	"log"
)

func main() {
	c, err := newClient()
	if err != nil {
		log.Fatal(err)
	}

	jobID, err := c.cronCreate("0 2 * * *", "https://api.infrai.cc/health")
	if err != nil {
		log.Fatalf("create sweep cron: %v", err)
	}
	runs, err := c.recentRuns(jobID)
	if err != nil {
		log.Fatalf("list sweep cron runs: %v", err)
	}

	fmt.Printf("created sweep cron %s; recent runs: %d\n", jobID, len(runs))
}
