# Architecture

Rivendell is a personal bookmarking, media collection, and archiving system. It stores records of articles, podcasts, videos, books, games, movies, TV shows, music (CDs and vinyl), Magic: The Gathering cards, GitHub repositories, RSS feeds, and work history. On record creation, the system automatically enriches entries with metadata and archives media files to cloud storage.

## Purpose

A single self-hosted database that:
1. Catalogs media across 15 collection types with a unified API
2. Automatically enriches records with metadata from external services (cover art, year, creator info)
3. Archives content to durable cloud storage (Backblaze B2) so it survives link rot
4. Detects dead links via a daily cron job
5. Provides read-only analytics via a separate Datasette service

## System overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│  Tailscale (private network)                                            │
│                                                                         │
│  rivendell.<tailnet>:443 ──────────► PocketBase app (:8090)             │
│  rivendell-datasette.<tailnet>:443 ─► Datasette (:8001)                 │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────┐       ┌──────────────────────────────────────────┐
│  Client (curl/app)  │──POST─│  PocketBase                              │
│                     │       │                                          │
│                     │       │  ┌─────────┐  ┌──────────┐  ┌────────┐  │
│                     │       │  │Preparers│─►│PB Save   │─►│Enricher│  │
│                     │       │  │(resolve │  │(validate │  │(API +  │  │
│                     │       │  │ names,  │  │ + write) │  │ upload)│  │
│                     │       │  │ defaults│  └──────────┘  └───┬────┘  │
│                     │       │  └─────────┘                    │       │
│                     │       │                                 ▼       │
│                     │       │                          ┌────────────┐ │
│                     │       │                          │ Save again │ │
│                     │       │                          │(enriched   │ │
│                     │       │                          │ fields)    │ │
│                     │       │                          └────────────┘ │
└─────────────────────┘       └──────────────────────────────────────────┘
                                          │
                                          ▼
                    ┌─────────────────────────────────────────┐
                    │         External APIs                    │
                    │                                         │
                    │  YouTube Data API    Backblaze B2       │
                    │  TMDB               OpenLibrary         │
                    │  IGDB (via Twitch)  Discogs             │
                    │  Scryfall           GitHub GraphQL      │
                    │  SingleFile CLI     yt-dlp              │
                    └─────────────────────────────────────────┘
```

## Technology stack

| Layer | Technology | Role |
|-------|-----------|------|
| Framework | PocketBase (Go) | REST API, SQLite DB, auth, admin UI |
| Database | SQLite (WAL mode) | Single-file, embedded, no external DB server |
| Storage | Backblaze B2 | Durable object storage for archived media |
| Networking | Tailscale | Private mesh VPN, HTTPS certs, no public ports |
| Analytics | Datasette (Python) | Read-only SQL dashboards on a DB snapshot |
| Containers | Docker Compose | 4 services: app, datasette, 2 tailscale nodes |
| Video download | yt-dlp | YouTube video archiving |
| HTML snapshot | single-file-cli + Chromium | Full-page article archiving |
| Media tagging | ffmpeg | Embed metadata into MP4/MP3 files |

## Record lifecycle

Every record flows through three phases on create:

### Phase 1: Prepare (before PocketBase save)

Runs synchronously before the record hits the database.

| Action | Collections | What happens |
|--------|------------|--------------|
| Set boolean defaults | articles, podcasts, videos, feeds | `dead=false`, `shared=false`, `favorite=false` |
| Resolve tag names → IDs | articles, podcasts, videos, feeds, read_later, watch_later | Lookup `meta` records by name, replace with IDs |
| Resolve single relation names → IDs | books, cds, games, movies, shows, vinyls | `genre`, `platform`, `definition`, `status` names → meta IDs |
| Default status | books, games, movies, shows | Set `status` to `not_started` if omitted |

### Phase 2: PocketBase save

The framework validates required fields, enforces relation constraints, and writes the record to SQLite.

### Phase 3: Enrich (after PocketBase save)

Runs after the record exists in the DB. Calls external APIs, downloads/uploads media, then saves enriched fields back.

| Collection | External API | Fields set | Media archived |
|-----------|-------------|-----------|----------------|
| articles | — | `archive` | Markdown (readability-extracted) + SingleFile HTML → B2 |
| podcasts | — | `archive` | MP3 download + ffmpeg metadata tag → B2 |
| videos | YouTube Data API | `title`, `creator`, `url`, `year`, `archive` | MP4 via yt-dlp + ffmpeg metadata tag → B2 |
| github | GitHub GraphQL | `name`, `owner`, `description`, `language` | None |
| mtg | Scryfall | all card fields + `image`, `back` | Card images → B2 |
| books | OpenLibrary | `year`, `cover` | Cover image → B2 |
| cds | Discogs | `year`, `cover` | Cover image → B2 |
| vinyls | Discogs | `year`, `cover` | Cover image → B2 |
| games | IGDB | `year`, `cover` | Cover image → B2 |
| movies | TMDB | `year`, `cover` | Cover image → B2 |
| shows | TMDB | `year`, `cover` | Cover image → B2 |
| watch_later | YouTube Data API | `title`, `channel` | None |

### Enrichment skip conditions

- `podcasts`, `videos`: skipped if `archive` field is already set on create
- `books`: skipped if no `isbn` provided
- `mtg`: card metadata fields skipped if `rarity` already set (sentinel for pre-filled data)

## Collection architecture

### The `meta` collection

Central lookup table referenced by all other collections. Stores tags, genres, platforms, statuses, and definitions as `(name, type)` pairs.

```
meta
├── type: "tags"        → articles, podcasts, videos, feeds, read_later, watch_later
├── type: "genre"       → books, cds, games, movies, shows, vinyls
├── type: "platform"    → games
├── type: "status"      → books, games, movies, shows
└── type: "definition"  → movies, shows
```

Its collection ID is pinned via `META_ID` env var so relation fields can reference it at schema definition time (before the collection exists in the DB).

### Name resolution

Callers send **name strings** (e.g. `"rock"`, `"ps5"`, `["programming", "go"]`), not opaque IDs. The server resolves names to meta record IDs before saving:

- **Tags** (multi-relation): query `meta` for all matching names with `type = "tags"`, replace field with ID array
- **Single relations** (genre, platform, status, definition): query `meta` for exact `name + type` match, replace field with single ID

Raw meta IDs do **not** work as input — they fail to match any name and the create is rejected.

### Collection groups

```
┌─────────────────────────────────────────────────────────────────┐
│  Archivable (content preserved to B2)                           │
│  articles, podcasts, videos                                     │
│  Fields: title, creator, url, tags, archive, dead, shared, fav  │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  Media (cover + year enriched)                                  │
│  books, cds, games, movies, shows, vinyls                       │
│  Fields: title/album, creator/artist, year, cover, genre        │
│  + collection-specific: isbn, barcode, platform, definition,    │
│    season, status, favorite                                     │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  Reference (metadata-only, no media archiving)                  │
│  github, mtg, feeds, read_later, watch_later, records           │
└─────────────────────────────────────────────────────────────────┘
```

## Source structure

```
rivendell/
├── main.go              # App init, hook wiring, cron, app.Start()
├── enrichers.go         # All enrichment logic (API calls, B2 uploads)
├── preparers.go         # Relation name resolution, field defaults
├── schema/
│   └── collections.go   # Collection field definitions (used by migrations)
├── helpers/
│   ├── getBookInfo.go   # OpenLibrary ISBN lookup
│   ├── getContent.go    # Content fetcher (article/video/podcast/media)
│   ├── getGameInfo.go   # IGDB game lookup via Twitch OAuth
│   ├── getMangaInfo.go  # MangaDex lookup (legacy, unused by hooks)
│   ├── getMediaInfo.go  # TMDB movie/show search + detail
│   ├── getMTGInfo.go    # Scryfall card lookup
│   ├── getMusicInfo.go  # Discogs music search (CDs + vinyls)
│   ├── getRepoInfo.go   # GitHub GraphQL repository info
│   ├── getYTInfo.go     # YouTube Data API video metadata
│   ├── key.go           # Env var loader (.env → key map)
│   └── uploadToB2.go    # Backblaze B2 auth + upload
├── linkcheck/
│   └── linkcheck.go     # Dead-link detection (HEAD/GET, retry, worker pool)
├── backup/
│   └── backup.go        # Collection → JSON export to B2 (backup cron job)
├── utils/
│   ├── cmd.go           # Shell command executor
│   ├── deleteFiles.go   # File cleanup
│   ├── emojiUnicode.go  # Emoji → U+XXXX conversion
│   ├── fileNameFmt.go   # Title → safe filename
│   ├── getFileType.go   # Collection → file extension + MIME
│   ├── toCapitalized.go # Title case
│   └── ytdl.go          # yt-dlp wrapper
├── datetime/            # Date parsing and arithmetic
├── migrations/          # Versioned schema migrations (auto-run on startup)
├── datasette/
│   ├── Dockerfile       # Python image + datasette + plugins
│   ├── entrypoint.sh    # Sync loop: resync snapshot → restart datasette (5 min)
│   ├── sync.sh          # Copy DB → VACUUM INTO → denormalize → atomic swap
│   ├── denormalize.py   # Replace meta IDs with names in snapshot
│   ├── metadata.yml     # ~50 canned queries + 10 dashboards
│   └── plugins/
│       └── render_meta_names.py  # Prettify tags cell display
├── Dockerfile           # Multi-stage Go build + runtime deps (chromium, ffmpeg, yt-dlp, single-file)
├── docker-compose.yml   # 4 services: tailscale, app, datasette, tailscale-datasette
└── deploy.sh            # git pull + docker compose up --build
```

## Hook execution model

PocketBase hooks use a middleware-chain pattern. The `OnRecordCreateRequest` hook wraps record creation:

```
Request arrives
    │
    ▼
Preparer runs (resolve names, set defaults)
    │
    ▼
e.Next() — PocketBase validates + saves the record
    │
    ▼
Enricher runs (external APIs, media upload)
    │
    ▼
e.App.Save(record) — persist enriched fields
    │
    ▼
Response returned to client
```

The client gets back the fully enriched record in the response. Enrichment is **synchronous** — the request blocks until all API calls and uploads complete. This is acceptable for a single-user personal system.

Because the request blocks on external I/O, all outbound HTTP calls use shared clients with timeouts (`helpers/http.go`): `HTTPClient` (30s) for metadata/API lookups and `MediaClient` (300s) for large transfers (media downloads, B2 uploads). A hung upstream host therefore fails the request on a deadline rather than stalling it indefinitely. The one exception is TMDB's `tmdbClient`, which keeps its own timeout plus gzip-EOF handling.

## Cron jobs

| Job ID | Schedule | What it does |
|--------|----------|--------------|
| `link_check` | Daily 3am | HEAD-check every non-dead URL in articles/podcasts/videos. Flip `dead=true` on failures. Worker pool of 5, one retry on network error. |
| `backup` | Mon/Wed/Fri 4am | Export every non-system, non-auth collection to `Backups/<collection>/<date>.json` on B2. Strips PB meta fields, resolves relation IDs to meta names. Sequential; one collection's failure is logged and skipped. |

See [CRON.md](CRON.md) for the full per-job detail and the convention for adding new jobs.

## Archiving pipeline

For collections that archive content to B2:

```
┌──────────┐     ┌──────────────┐     ┌──────────┐     ┌─────┐
│  Fetch   │────►│  Process     │────►│  Upload  │────►│ B2  │
│  content │     │  (convert/   │     │  to B2   │     │     │
│  from URL│     │   tag/snap)  │     │          │     │     │
└──────────┘     └──────────────┘     └──────────┘     └─────┘

Articles:  HTTP GET → readability extract → Markdown     → B2 (.md)
           HTTP GET → single-file-cli    → HTML snapshot → B2 (.html) [best-effort]

Videos:    yt-dlp download → ffmpeg metadata tag → B2 (.mp4)

Podcasts:  HTTP GET → ffmpeg metadata tag → B2 (.mp3)
```

### B2 path convention

Files are stored as: `{Collection}/{sanitized_filename}.{ext}`

Examples:
- `Articles/How_Postgres_Handles_Transactions.md`
- `Videos/Never_Gonna_Give_You_Up.mp4`
- `Books/Nineteen_Eighty-Four.jpeg`
- `MTG/dmu/Sheoldred_the_Apocalypse.jpeg`

The `fileNameFmt` utility sanitizes titles: spaces → underscores, separators → dashes, emojis removed, special characters stripped.

### Media metadata tagging

Videos and podcasts are tagged with metadata (title, artist, year, genre) via ffmpeg before upload. This embeds discoverable metadata into the archived file itself (iTunes atoms for MP4, ID3 frames for MP3), making the archive self-describing even outside PocketBase.

## Datasette analytics layer

A separate read-only service that runs SQL queries and dashboards over a periodically-refreshed snapshot of the main database.

```
PocketBase (live, locked) ──► cp + VACUUM INTO ──► denormalize.py ──► datasette serve --immutable
                                                                              │
                                 ◄── 5 min loop ──────────────────────────────┘
```

Key design decisions:
- **Snapshot, not live query**: PocketBase holds locks; direct access fails. Copy-then-VACUUM sidesteps.
- **Denormalized**: Relation IDs replaced with names so facets/filters show readable values.
- **Own Tailscale hostname**: Avoids path-prefix routing bugs. Datasette serves at root.
- **Immutable + restart loop**: Datasette caches at startup; restart to pick up new data.

## Logging

All application logging uses PocketBase's structured `slog.Logger` (`app.Logger()`), which persists to the `_logs` SQLite table and surfaces in the admin UI's Logs panel.

Every log line includes:
- `collection` — which collection the operation targets
- Record identifier — `title`, `album`, `name`, `url`, or `link` depending on collection type

Log points in the record lifecycle:
1. `record create started` — request received
2. `record prepare failed` — name resolution error
3. `archive started` / `archive uploaded` — content fetching and B2 upload
4. `enrich fetching *` — external API call initiated
5. `enrich cover uploaded` / `enrich * resolved` — API success
6. `record enrich failed` — enrichment error
7. `record enrich save failed` — DB write error after enrichment
8. `record created` — success, with `enriched: true/false`

## Access control

| Operation | Rule |
|-----------|------|
| Create | Open (no auth required) — enables simple API integrations |
| View | Auth required (`@request.auth.id != ''`) |
| Update | Auth required |
| Delete | Admin only (default PocketBase behavior) |

Auth uses PocketBase superuser impersonate tokens (non-expiring, full access). These are API keys for server-to-server calls, not user sessions.

## External API map

| Service | Used by | Auth | What's fetched |
|---------|---------|------|----------------|
| Backblaze B2 | All archivable collections | `B2_APP_KEY_ID` + `B2_APP_KEY` | Upload destination |
| YouTube Data API v3 | videos, watch_later | `YOUTUBE_KEY` | Title, channel, publish date |
| TMDB v3 | movies, shows | `TMDB_KEY` | Title, year, poster, director |
| IGDB (Twitch OAuth) | games | `TWITCH_CLIENT_ID` + `TWITCH_CLIENT_SECRET` | Title, year, cover, publisher |
| Discogs | cds, vinyls | `DISCOGS_TOKEN` | Year, cover image |
| OpenLibrary | books | None (public) | Year, cover image |
| Scryfall | mtg | None (public, User-Agent only) | All card fields + images |
| GitHub GraphQL | github | `GH_TOKEN` | Repo name, owner, description, language |

## Deployment topology

```
┌──────────────────────────────────────────────────────────┐
│  Docker host                                             │
│                                                          │
│  ┌────────────────┐   ┌──────────────────────────────┐  │
│  │ tailscale      │   │ app (PocketBase)             │  │
│  │ rivendell:443 ─┼──►│ :8090                        │  │
│  │                │   │ volumes: .env, pb_data/      │  │
│  └────────────────┘   └──────────────────────────────┘  │
│                                                          │
│  ┌────────────────┐   ┌──────────────────────────────┐  │
│  │ tailscale-     │   │ datasette                    │  │
│  │ datasette:443 ─┼──►│ :8001                        │  │
│  │                │   │ volumes: pb_data/ (RO),      │  │
│  └────────────────┘   │          datasette-data/     │  │
│                        └──────────────────────────────┘  │
│                                                          │
│  Network: internal (no public ports)                     │
│  Access: Tailscale tailnet only                          │
└──────────────────────────────────────────────────────────┘
```

Service startup order: `tailscale` → `app` (waits for tailscale healthy) → `datasette` (waits for app healthy) → `tailscale-datasette` (waits for datasette healthy).

## Implementing in another language/framework

The core concepts are framework-agnostic:

1. **Hook-based enrichment**: Intercept record creation, call external APIs, write back. Any framework with lifecycle hooks or middleware works (Django signals, Rails callbacks, Hono middleware, etc.).

2. **Name resolution layer**: Accept human-readable names in the API, resolve to internal IDs before save. A simple lookup table + pre-save hook.

3. **Synchronous archiving**: Block on content download + cloud upload during create. For multi-user systems, move to a job queue (but for single-user, synchronous is simpler and the response contains the final record).

4. **Shared meta table**: One table for all lookup values (tags, genres, platforms, etc.), typed by a `type` column. Avoids N tiny tables.

5. **Analytics via snapshot**: Copy the DB, denormalize relations for readability, serve read-only. Works with any SQL database + any dashboard tool.

Key decisions to replicate:
- Archive content **at creation time** — don't rely on the source URL persisting
- Store the B2/S3 URL on the record — the record is self-contained after enrichment
- Make creates open (no auth) for easy integrations; gate reads behind auth
- Use structured logging with collection + record identity on every log line
- Dead-link detection as a background job, not on every read
