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

This applies **app-wide** — cron jobs, record create hooks (enrichers/preparers), and the archive pipeline all use `app.Logger()` with structured key-value attrs. Every log line includes at minimum `collection` and a record identifier (`title`, `album`, `name`, `url`, or `link` depending on collection) so failures can be pinpointed to a specific record.

Log levels:
- `Info` — lifecycle events (record create started/completed, enrich started/skipped, archive uploaded, link flipped dead, collection summary)
- `Error` — failures (prepare failed, enrich failed, save failed, archive failed, SingleFile capture/upload failed)

Pass structured key-value attrs, not formatted strings:

```go
app.Logger().Info("link_check marked dead", "collection", collectionName, "record_id", record.Id, "url", url)
app.Logger().Error("link_check save failed", "collection", collectionName, "record_id", record.Id, "error", err.Error())
app.Logger().Info("record created", "collection", collection, "record", label, "enriched", needsSave)
app.Logger().Error("record enrich failed", "collection", collection, "record", label, "error", err.Error())
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

### `backup` — daily at 4am (`0 4 * * *`)

Exports every non-system, non-auth collection to a pretty-printed JSON file on Backblaze B2 at `Backups/<collection>/<YYYY-MM-DD>.json` — one dated file per collection per run, never overwritten. Implementation: `backup/backup.go`.

- **Scope:** enumerated dynamically via `app.FindAllCollections()`; skips `collection.System` (e.g. `_superusers`, `_logs`) **and** `collection.IsAuth()` (e.g. `users`) — auth records can carry credentials/email and the files land off-site on B2. Keeps the ~16 data collections and `meta`. Collections added later are picked up automatically.
- **Per record:** `record.PublicExport()` then strip PB meta keys (`id`, plus a defensive `created`/`updated`/`collectionId`/`collectionName`/`expand` superset), and resolve relation fields from `meta` record IDs to their names (single relation → string, multi → array). A meta ID with no match is kept as-is. An empty collection still writes `[]`.
- **Format:** `json.MarshalIndent` (2-space) — diffable across dated snapshots.
- **Concurrency:** sequential across collections (~16, each one fetch + marshal + one B2 upload, off-peak).
- **Failure isolation:** a `buildMetaNames` or `FindAllCollections` failure aborts the whole run (every collection needs the meta map); a single collection's failure is logged and skipped (`continue`), never aborting the rest.
- **Logging:** `backup started`, per-collection `collection backed up` (`record_count`), per-collection failure, `backup done` (`backed_up`/`failed`) — see `backup.BackupAll`.
- **Restore:** `scripts/restore_collection.py <collection> --date <YYYY-MM-DD>` — relations are stored as names precisely so a restore matches the create API's input format. It diffs the backup against what's live, reports what's missing, and only writes with `--apply`. After each create it PATCHes back any non-relation field the preparers or enrichers overwrote (`prepareGame` forces `favorite = false`; `enrichGames` replaces `year`/`cover` from IGDB). Matching is by a natural key (`title` for most collections), so re-running is safe.
- **Why daily, not Mon/Wed/Fri:** the schedule used to be `0 4 * * 1,3,5`. On 2026-07-19 a superuser bulk-deleted 39 `games` records; 3 of them had been created earlier that same day and so appeared in no backup at all, and were lost permanently. A ≤24h window bounds that.

## Adding a new cron job

1. Put the job's logic in its own package (see `linkcheck/` as the template) — keep `main.go` to wiring, not business logic.
2. Use `app.Logger()` for anything you'd want visible in the admin UI, not `log.Printf`.
3. Decide the failure-isolation boundary up front: should one failure (e.g. one bad record, one bad sub-task) abort the whole job, or just be logged and skipped? `linkcheck.CheckCollection` logs-and-continues on a per-record failure but returns an error on a query-level failure — pick what's appropriate per job, not "log everything and always continue" by default.
4. Register with `app.Cron().MustAdd("<unique_id>", "<cron_expr>", handler)` in `main.go`, before `app.Start()`.
5. To test before shipping: temporarily change the cron expression to `"* * * * *"` (every minute), run `go run . serve`, and watch the admin UI Logs panel. Revert the expression before committing — don't ship a testing schedule.
