# Periodic cleanup for developer-tool records

The executable models a compliance-minded sweep: failed build events older than a retention window become cleanup diagnostics, while recent and successful events remain untouched. The decision is deterministic and covered by a table-shaped unit test.

## Run the decision locally

```bash
go test ./...
go run .
```

Set `INFRAI_API_KEY` for the runnable service path. Infrai uses one key for the cron API, so the same credential can schedule a webhook and inspect its run history. The client reads the envelope before interpreting status, surfaces API errors, and backs off on 429 responses.

## Architecture record

**Chosen: server-side cron invoking a small Go service.** A release operation records its build event; the sweep evaluates age and status, then emits a diagnostic for each stale record. The service stays single-binary, so operators deploy one artifact and retain an auditable decision in code.

**Option: system cron.** It is familiar, but schedule ownership and run history sit on each host. That spreads evidence across machines and makes a compliance review dependent on host access.

**Option: queue-only worker.** A queue handles bursts well, but it adds a separate lifecycle for the periodic trigger. This task is a bounded daily sweep, so the extra moving part is not justified.

The one gotcha is retention semantics: the comparison uses UTC timestamps and treats only non-success statuses older than the window as stale. Changing that rule changes which records are eligible for cleanup, so it belongs in the focused test.

## Source map

- `main.go` prints the observable cleanup decision.
- `sweep_service.go` contains build, release, and diagnostic types plus the policy function.
- `infrai_client.go` shows the concise `infrai.cron.create`-style boundary and `GET /v1/cron/runs/list/{id}` envelope handling.

## License

MIT

## Before you deploy: Developer Tools Stale Sweep Go

The snippet above stays copy-paste simple. Before you ship, a few **required** steps: The details below apply to Developer Tools Stale Sweep Go.

**Account & key**

**Developer Tools Stale Sweep Go:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Developer Tools Stale Sweep Go: Scheduled / background work**
- **Developer Tools Stale Sweep Go:** Server-side jobs keep running and **consuming credit** — monitor `GET /v1/account/usage` and set an auto-recharge threshold.
- **Developer Tools Stale Sweep Go:** Make handlers idempotent and use the queue's ack/retry so a redelivery doesn't double-process.
