# Testing

Unit tests cover all pure functions — logic with no external I/O, no API calls, no filesystem. External integrations (B2, GitHub, Scryfall, TMDB, etc.) are not tested here.

## Running tests

```sh
go test ./...
```

Run with verbose output:

```sh
go test ./utils/... ./datetime/... ./helpers/... ./linkcheck/... ./backup/... -v
```

## Test files

### `utils/utils_test.go`

Tests string transformation and file type utilities.

| Function | Cases | What's verified |
|----------|-------|-----------------|
| `FileNameFmt` | 22 | Spaces → underscores; separators (` - `, ` :: `, ` — `, ` : `) → dashes; `&` → `_and_`; trailing `.`/`?`/`!` stripped; emojis removed; special chars stripped; pipes normalized; smart quotes stripped; period-space/comma-space → dash; combined transforms; unicode non-spacing marks stripped |
| `ToCapitalized` | 5 | Lowercase → title case; already-capitalized passthrough; empty string |
| `ConvertEmoji` | 4 | Single emoji → `U+XXXX` code point; empty string → empty; non-emoji char |
| `EmojiUnicode` | 3 | Emoji → `U+XXXX` format; non-emoji passthrough; multiple emoji |
| `GetFileType` | 7 | `articles` → `md`/`text/markdown`; `podcasts` → `mp3`; `videos` → `mp4`; `comics` with image URL → correct extension and MIME type; unknown type → zero value |

### `datetime/datetime_test.go`

Tests date/time parsing and arithmetic.

| Function | Cases | What's verified |
|----------|-------|-----------------|
| `ParseISO` | 4 | Valid ISO 8601 strings with UTC and negative offsets; invalid format errors; empty string error |
| `SubDays` | 4 | Zero days (no change); 1 day; 7 days; month boundary wrap (leap year) |
| `SubHours` | 4 | Zero hours; mid-day subtraction; boundary to midnight; day rollover |
| `IsAfter` | 3 | Later date is after earlier; earlier is not after later; equal dates return false |

### `helpers/helpers_test.go`

Tests URL parsing functions used to extract IDs and metadata before API calls. All functions are package-private; tests live in `package helpers` for direct access.

| Function | Cases | What's verified |
|----------|-------|-----------------|
| `parseGHURL` | 5 | Standard `github.com/owner/repo`; URL with trailing slash; URL with path suffix; non-GitHub URL errors; empty string errors |
| `parseMTGURL` | 3 | Valid Scryfall oEmbed URL extracts card UUID; URL without `/oembed` path errors; empty string errors |
| `parseMDURL` | 3 | MangaDex title URL extracts chapter UUID; URL with slug after ID; non-matching URL returns input |
| `parseTMDBURL` | 4 | Movie URL extracts ID and `movie` category; TV URL extracts ID and `tv` category; URL without slug; non-matching URL returns raw input |
| `cleanYTURL` | 4 | Short `youtu.be` URL; full `youtube.com/watch?v=` URL; `youtube.com` without `www`; `feature=share` param stripped — all extract same video ID |
| `escapeText` | 4 | Newlines escaped to `\n` literals; no-newline passthrough; multiple newlines; empty string |
| `parseDiscogsTitle` | 5 | Standard `Artist - Album` format; artist with dash in name; album with dash (preserves remainder after first separator); no separator returns empty artist and full string as album; empty string |

### `linkcheck/linkcheck_test.go`

Tests the HTTP status code classification logic used by the `link_check` cron job.

| Function | Cases | What's verified |
|----------|-------|-----------------|
| `isDeadStatus` | 11 | 2xx/3xx → alive; 401/403/429 → alive (anti-bot/rate-limit, not real death); 400/404/410/5xx → dead |
| `CheckURL` | 5 | Live HTTP servers: 200 alive, 404/500 dead, 403/429 alive |
| `CheckURL` (HEAD fallback) | 1 | Falls back to GET when server returns 405 on HEAD |
| `CheckURL` (unreachable) | 1 | Returns false for connection-refused host |

### `backup/backup_test.go`

Tests the pure record-transformation logic used by the `backup` cron job. The DB/B2 I/O functions (`buildMetaNames`, `backupCollection`, `BackupAll`) are not unit-tested, per convention.

| Function | Cases | What's verified |
|----------|-------|-----------------|
| `stripAndResolve` | 5 | PB meta keys (`id`/`created`/`updated`/`collectionId`/`collectionName`/`expand`) stripped; single relation ID → name; multi relation `[]ID` → `[]name`; unresolved (orphan) ID kept as-is; non-relation fields untouched |

## Bugs found during testing

`ConvertEmoji` in `utils/emojiUnicode.go` panicked on any emoji. The original code used JavaScript surrogate-pair math (`runeValue[0] + runeValue[1]`) but Go's `[]rune` decodes UTF-8 directly to Unicode code points — emoji are a single rune, not two. Fixed to use `runeValue[0]` directly.
