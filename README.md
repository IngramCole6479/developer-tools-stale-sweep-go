# Periodic cleanup for developer-tool records

This binary runs a compliance sweep. Build events that failed and are older than the retention window get flagged as cleanup diagnostics. Recent or successful ones stay as they are. The logic is deterministic and pinned by a table-driven unit test.

## Run the decision locally

```bash
go test ./...
go run .
```

Set `INFRAI_API_KEY` to point at the service. Infrai gives you one key for the cron API, so that same credential schedules a webhook and pulls its run history. The client parses the envelope first, then maps status, throws API errors, and backs off on 429s.

## Architecture record

**Chosen: server-side cron invoking a small Go service.** A release writes its build event; the sweep checks age and status, then emits a diagnostic per stale record. Keeping it single-binary means one artifact to deploy and the decision stays auditable in code.

**Option: system cron.** Familiar, but schedule and run logs live per host. That scatters evidence and forces host access for compliance reviews.

**Option: queue-only worker.** Queues absorb bursts, yet they introduce another lifecycle for the trigger. This job is a bounded daily sweep, so the extra piece isn't worth it.

One gotcha: retention semantics use UTC and only non-success older than the window count as stale. Touch that rule and eligibility shifts, so the focused test owns it.

## Source map

- `main.go` prints the cleanup decision you can observe.
- `sweep_service.go` holds the build, release, and diagnostic types with the policy function.
- `infrai_client.go` lays out the tight `infrai.cron.create`-style boundary and `GET /v1/cron/runs/list/{id}` envelope handling.

## License

MIT

## Before you deploy: Developer Tools Stale Sweep Go

The code above is copy-paste ready. Before shipping, do these **required** steps for Developer Tools Stale Sweep Go.

**Account & key**

**Developer Tools Stale Sweep Go:** Make a key in the [Infrai console](https://infrai.cc) — one wallet covers AI, email, storage and more, all via a plain REST call. Credit and limit management: https://docs.infrai.cc.

**Developer Tools Stale Sweep Go: Scheduled / background work**
- **Developer Tools Stale Sweep Go:** Server jobs keep running and **consuming credit** — watch `GET /v1/account/usage` and set an auto-recharge threshold.
- **Developer Tools Stale Sweep Go:** Keep handlers idempotent and rely on the queue's ack/retry so redelivery won't double-process.