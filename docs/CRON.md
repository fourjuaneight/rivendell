# Cron Jobs

Scheduled background jobs, registered via PocketBase's built-in `app.Cron()` scheduler. They run in-process — no external scheduler, no separate binary.

## How it works

Jobs are registered in `main.go`, after the `OnRecordCreateRequest` hook wiring and before `app.Start()`:

```go
app.Cron().MustAdd("link_check", "0 3 * * *", func() {
	app.Logger().Info("link_check started")
	for _, name := range []string{"articles", "podcasts", "videos"} {
		linkcheck.CheckCollection(app, name)
	}
})
```

- `app.Cron().MustAdd(id, cronExpr, handler)` — panics on an invalid cron expression (fails fast at startup instead of silently registering nothing). Use `.Add(...)` instead if you want the error returned rather than a panic.
- `id` must be unique — re-registering the same `id` replaces the job. System jobs use `__pb*__` IDs (logs cleanup, backups) — don't `Remove`/`RemoveAll`/`Stop` those by accident.
- Standard 5-field cron syntax: minute hour day month weekday. Supports lists, ranges, steps, and the usual macros.
- Jobs start automatically when `app.Start()` runs (no separate `migrate up`-style step) and each firing runs in its own goroutine.

## Logging

Use `app.Logger()` (PocketBase's structured `slog.Logger`), not `log.Printf` — `app.Logger()` persists to the `_logs` table and shows up in the admin UI's Logs panel (`/_/` → Logs). Plain `log.Printf` only goes to stdout, which means checking on a job means SSHing into the container.

Log levels used by `linkcheck`:
- `Info` — normal lifecycle events (job started, a record's `dead` flag flipped, a collection's run summary)
- `Error` — failures (a DB query failed, a record save failed)

Pass structured key-value attrs, not formatted strings:

```go
app.Logger().Info("link_check marked dead", "collection", collectionName, "record_id", record.Id, "url", url)
app.Logger().Error("link_check save failed", "collection", collectionName, "record_id", record.Id, "error", err.Error())
```

## Existing jobs

### `link_check` — daily at 3am (`0 3 * * *`)

Checks the `url` field on every non-`dead` record in `articles`, `podcasts`, and `videos`. Flips `dead` to `true` when a link no longer resolves. Implementation: `linkcheck/linkcheck.go`.

- **Request:** HEAD first; falls back to GET only if the server returns 405/501. 10s timeout. Sets a browser-like `User-Agent` (the Go default UA gets blocked/misclassified by some hosts).
- **Retry:** once, on network error/timeout only — a definitive HTTP status is never retried.
- **Verdict:** 2xx/3xx → alive. 401/403/429 → alive (anti-bot wall or rate-limit, not real link death). 404/410/5xx/timeout/connection error → dead.
- **Concurrency:** worker pool capped at 5 per collection (the three collections run sequentially, so total concurrency across a full job run stays at 5).
- **Logging:** job start, per-collection query failure, per-record save failure, per-record dead-flip, per-collection summary (`checked`/`flipped` counts) — see `linkcheck.CheckCollection`.
- **Known limitation:** the concurrency cap bounds *total* in-flight requests, not *per-host* requests — if several records share a host, up to 5 can hit it at once. Mitigated by the daily schedule rather than per-host throttling; acceptable for a personal-scale collection.

## Adding a new cron job

1. Put the job's logic in its own package (see `linkcheck/` as the template) — keep `main.go` to wiring, not business logic.
2. Use `app.Logger()` for anything you'd want visible in the admin UI, not `log.Printf`.
3. Decide the failure-isolation boundary up front: should one failure (e.g. one bad record, one bad sub-task) abort the whole job, or just be logged and skipped? `linkcheck.CheckCollection` logs-and-continues on a per-record failure but returns an error on a query-level failure — pick what's appropriate per job, not "log everything and always continue" by default.
4. Register with `app.Cron().MustAdd("<unique_id>", "<cron_expr>", handler)` in `main.go`, before `app.Start()`.
5. To test before shipping: temporarily change the cron expression to `"* * * * *"` (every minute), run `go run . serve`, and watch the admin UI Logs panel. Revert the expression before committing — don't ship a testing schedule.
