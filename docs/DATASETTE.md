# Datasette Analytics

A read-only [Datasette](https://datasette.io/) service that surfaces correlations, patterns, and aggregates over the Rivendell collections — counts by year/creator, tag frequency, genre/platform/status breakdowns, and cross-collection summaries — as canned SQL queries and visual dashboards.

It runs as a separate container in `docker-compose.yml` and serves a periodically-refreshed **snapshot** of PocketBase's SQLite database, so it never touches the live data.

## Access

Behind Tailscale, on its **own dedicated node/hostname**, served at the root on `:443`:

```
https://rivendell-datasette.<your-tailnet>.ts.net/
```

- `/` — Datasette home (databases, tables, canned queries)
- `/-/dashboards` — dashboard index
- `/-/dashboards/<id>` — an individual dashboard

For local development the container also binds `http://127.0.0.1:8001` (see `ports` in `docker-compose.yml`).

There is no auth — access is gated by Tailscale (tailnet-only), same as the app.

> **Why a dedicated node instead of a `/analytics/` path prefix on the main host?** Tailscale's path handlers *strip* the prefix before proxying, while Datasette's `base_url` assumes the proxy doesn't. The mismatch left request-path links (facets, sort) pointing at `/rivendell/...` with no prefix, which fell through to PocketBase → 404. Giving Datasette its own hostname at the root removes `base_url` entirely, so no prefix can ever be lost. (A non-standard HTTPS port like `:8443` works too, but `http://…:8443` foot-guns and the URL is uglier — a dedicated node keeps a clean `:443` URL.)

## Architecture

```
PocketBase ──writes──> pb_data/data.db (+ -wal/-shm, WAL mode, live, locked)
                               │  bind-mounted READ-ONLY at /data/pb_data
                               ▼
sync.sh: cp data.db (+ -wal/-shm) ─> /tmp/pbsnap ─VACUUM INTO─> /tmp/pbsnap/raw.db
                               │  drop system + auth tables, VACUUM INTO again
                               ▼
                        /data/rivendell.db
                               │  snapshot in the datasette-data volume
                               ▼
datasette serve --immutable /data/rivendell.db        (no base_url — root)
                               │  datasette:8001 (internal network)
                               ▼
tailscale-datasette node  :443 ──/──>  rivendell-datasette.<tailnet>.ts.net
```

- **Read-only source.** `pb_data` is a read-only bind mount, so the container physically cannot write your live data.
- **Why copy-then-`VACUUM INTO`, not `sqlite3 .backup`.** PocketBase keeps the live DB open with locks (WAL mode). Pointing `sqlite3` directly at it over the read-only mount fails with `Error: database is locked`, leaving an empty `rivendell.db` and a 502. So `sync.sh` instead `cp`s the file-set (`data.db` + `-wal`/`-shm`) to a writable temp and runs `VACUUM INTO` on that *private copy*. That sidesteps the lock entirely and means sqlite never even opens the live DB. (sqlite replays the copied `-wal` to produce a consistent snapshot.)
- **The snapshot is stripped before it is served.** A raw copy of `data.db` carries secrets: `_params` holds the app settings — B2/S3 credentials in the clear unless `PB_ENCRYPTION_KEY` is set — while `_superusers` and `users` hold password hashes and token keys, and `_collections` holds any OAuth2 client secrets. Datasette is reachable by anyone on the tailnet, so `sync.sh` drops every underscore-prefixed (system) table plus `users`, matching the same policy `backup/backup.go` applies to B2 backups (`collection.System || collection.IsAuth()`). It then runs a **second** `VACUUM INTO`: `DROP TABLE` only moves pages to the freelist, so without it the dropped rows are still readable as raw bytes in the served file. A final check refuses to publish a snapshot in which `_params`, `_superusers`, or `users` survived.
- **Immutable + restart loop.** Datasette serves `--immutable` (fast; assumes the file never changes) and loads DBs only at startup, so the entrypoint **resyncs then restarts** Datasette every 5 minutes to pick up new data — it never overwrites a file Datasette has open. A guard skips serving if no snapshot exists yet (a failed sync can't leave it crash-looping on `--immutable <missing-file>`).
- **Routing.** A second Tailscale container (`tailscale-datasette`, hostname `rivendell-datasette`) proxies `/` on `:443` → `datasette:8001` via `serve-datasette.json`. Datasette runs with **no `base_url`** (it owns the whole host), so every link — assets, facets, sort, JS — stays on this host. (Earlier a `/analytics/` path prefix + `base_url` on the main host broke facet links: Tailscale strips the prefix but Datasette's request-path links don't re-add it → 404 into PocketBase.) Each Tailscale node has its own auth key (`TS_AUTHKEY` for the app node, `TS_AUTHKEY_DATASETTE` for this one); single-use keys are fine since `TS_AUTH_ONCE` + the state volume persist a node after first auth.

## Directory layout

```
datasette/
├── Dockerfile      # python:3.12-slim + datasette + pip plugins + sqlite3
├── entrypoint.sh   # wait-for-db, sync + serve loop (5 min refresh), missing-snapshot guard
├── sync.sh         # copy data.db file-set to temp, VACUUM INTO -> rivendell.db, then denormalize
├── denormalize.py  # rewrite snapshot relation IDs -> meta names (tags + single relations)
├── metadata.yml    # table metadata, ~50 canned queries, 10 dashboards
└── plugins/
    └── render_meta_names.py   # render_cell hook: prettify the tags array cell (["go","rust"] -> go, rust)
```

### Plugins

Installed in the `Dockerfile`:

- [`datasette-dashboards`](https://github.com/rclement/datasette-dashboards) — declarative dashboards of charts, defined in `metadata.yml`
- [`datasette-vega`](https://github.com/simonw/datasette-vega) — interactive charts from any query result
- [`datasette-copyable`](https://github.com/simonw/datasette-copyable) — copy/export table or query rows as CSV/TSV/Markdown

## metadata.yml

This is the heart of the analytics layer. Two parts:

- **`databases.rivendell.queries`** — ~50 named, read-only canned queries (one URL each, e.g. `/rivendell/articles_tags`).
- **top-level `plugins.datasette-dashboards`** — 10 dashboards (overview + one per collection group), each a set of vega-lite charts.

> **Placement matters:** dashboard definitions must live under the **top-level** `plugins:` key, not under `databases.<db>.plugins`. Datasette-dashboards only reads the top-level location; nesting it under a database silently registers nothing.

### Relation fields

PocketBase stores relation fields as meta record IDs. To make the analytics copy human-readable **everywhere** — table cells, facets, and filters — `sync.sh` runs `denormalize.py` after each snapshot, replacing IDs with their `meta` `name`:

- `tags` (multi-select) → a JSON array of names, e.g. `["go","rust"]`
- `genre` / `platform` / `status` / `definition` (single-select) → the name string

So the snapshot's columns already hold names. Canned queries group on the column directly — **no `meta` join**:

```sql
-- tag frequency
SELECT je.value AS tag, COUNT(*) AS count
FROM articles a, json_each(a.tags) je
GROUP BY je.value
ORDER BY count DESC

-- single relation
SELECT genre, COUNT(*) AS count
FROM books
WHERE genre IS NOT NULL AND genre != ''
GROUP BY genre
ORDER BY count DESC
```

`_facet_array=tags` and column facets show names too, because the stored values are names.

**Why denormalize instead of joining at query time?** Datasette renders facet values from the raw column and has **no hook to relabel them** (`render_cell` only affects table cells) — so the only way to get names in facets is to store names. The snapshot is a throwaway read-only copy, so dropping the IDs there is free.

The `plugins/render_meta_names.py` `render_cell` hook then just prettifies the `tags` *cell* (`["go","rust"]` → `go, rust`); it's a harmless no-op for the single-value columns now that they already hold names.

### Dashboards

| Dashboard | Covers |
|-----------|--------|
| `rivendell-overview` | collection sizes, all-media-by-year, favorites by collection |
| `articles-dashboard` | by year, top creators, top tags, top domains |
| `podcasts-dashboard` / `videos-dashboard` | by year, top creators, top tags |
| `books-dashboard` | by year, top authors, by genre, by status |
| `games-dashboard` | by year, by platform, by genre, by status, top publishers |
| `movies-dashboard` | by year, top directors, by genre, by definition, by status |
| `shows-dashboard` | by year, top directors, by genre, by status, by season |
| `music-dashboard` | CDs & vinyls: by year, top artists, by genre |
| `code-feeds-dashboard` | GitHub by language / top owners; feeds by type / tags |

## Running

The service starts with the rest of the stack:

```sh
docker compose up --build
```

Order is `tailscale` → `app` → `datasette` (Datasette `depends_on` the app being healthy, so `pb_data/data.db` exists before the first sync). The first sync waits for the DB to appear, so a cold start won't crash-loop.

## Adding a query or chart

1. **Canned query** — add an entry under `databases.rivendell.queries` in `metadata.yml`. It's instantly available at `/rivendell/<name>`.
2. **Chart** — add a chart under a dashboard's `charts:` in the top-level `plugins.datasette-dashboards` block. Charts use an inline `query:` plus a vega-lite `display:` spec (`mark` + `encoding`).
3. Rebuild the container (`metadata.yml` is copied at build time): `docker compose up --build datasette`.

Snapshot columns already hold names (see Relation fields), so group on the column directly — no `meta` join.

## Deployment notes

- **Two gitignored serve configs** (they hold environment-specific tailnet hosts):
  - `serve.json` — the main node (`rivendell`), `/` → app:

    ```json
    { "TCP": { "443": { "HTTPS": true } },
      "Web": { "rivendell.<your-tailnet>.ts.net:443": { "Handlers": { "/": { "Proxy": "http://rivendell-app:8090" } } } } }
    ```

  - `serve-datasette.json` — the datasette node (`rivendell-datasette`), `/` → datasette:

    ```json
    { "TCP": { "443": { "HTTPS": true } },
      "Web": { "rivendell-datasette.<your-tailnet>.ts.net:443": { "Handlers": { "/": { "Proxy": "http://datasette:8001" } } } } }
    ```

- **Set `TS_AUTHKEY_DATASETTE`** in `.env` — the datasette node's own auth key (separate from the app node's `TS_AUTHKEY`). Single-use is fine.
- **Tailscale reads serve config only at startup.** After changing either file, restart that node: `docker compose restart tailscale` / `docker compose restart tailscale-datasette`.
- The snapshot refreshes every 5 minutes. Brand-new records appear after the next resync (and Datasette restart), not instantly.

## Troubleshooting

**`HTTP ERROR 502`** at the datasette URL means Tailscale reached the container but Datasette isn't listening on 8001. Diagnose from the **datasette** container (the tailscale logs only show the symptom):

```sh
docker compose exec datasette sh -c 'ls -la /data/rivendell.db; (cat /proc/net/tcp | grep -q ":1F41 " && echo LISTENING || echo NOT-LISTENING)'
docker compose logs datasette --tail=40
```

- **`rivendell.db` is 0 bytes / `NOT-LISTENING`** — the snapshot is empty, so Datasette has nothing to serve. Run the sync by hand (traced) to see the real error:

  ```sh
  docker compose exec datasette sh -c 'sh -x /app/sync.sh; echo RC=$?'
  ```

  Originally this failed with `Error: database is locked` (sqlite was pointed straight at the live, locked DB) — fixed by the copy-then-`VACUUM INTO` design above. If it ever returns, confirm the `cp` step can read `/data/pb_data/data.db`.

- **Edits to `sync.sh` / `entrypoint.sh` / `metadata.yml` seem ignored** — they're baked into the image at build time. Rebuild: `docker compose up -d --build datasette` (add `--no-cache` if a layer is stale).

- **Stale data / new records missing** — the snapshot only refreshes every 5 minutes. Force it now: `docker compose restart datasette`.

- **502 immediately after `up`** — first sync + boot takes a few seconds (Datasette waits for the app to be healthy first). Give it ~20s.

- **A Datasette link 404s with a PocketBase JSON body** (`{"data":{},"message":"The requested resource wasn't found.","status":404}`) — the request hit the app node, not the datasette node, so a link lost its host. This is the path-prefix bug we eliminated: keep Datasette on its own hostname with **no** `--setting base_url`. Don't reintroduce a `/analytics/` path prefix on the main host.
