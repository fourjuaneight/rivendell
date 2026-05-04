# Testing

Unit tests cover all pure functions — logic with no external I/O, no API calls, no filesystem. External integrations (B2, GitHub, Scryfall, TMDB, etc.) are not tested here.

## Running tests

```sh
go test ./...
```

Run with verbose output:

```sh
go test ./utils/... ./datetime/... ./helpers/... -v
```

## Test files

### `utils/utils_test.go`

Tests string transformation and file type utilities.

| Function | Cases | What's verified |
|----------|-------|-----------------|
| `FileNameFmt` | 17 | Spaces → underscores; separators (` - `, ` :: `, ` — `, ` : `) → dashes; `&` → `_and_`; trailing `.`/`?`/`!` stripped; emojis removed; special chars stripped; pipes normalized |
| `ToCapitalized` | 5 | Lowercase → title case; already-capitalized passthrough; empty string |
| `EmojiUnicode` | 3 | Emoji → `U+XXXX` format; non-emoji passthrough; multiple emoji |
| `GetFileType` | 6 | `articles` → `md`/`text/markdown`; `podcasts` → `mp3`; `videos` → `mp4`; `comics` with image URL → correct extension and MIME type |

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
| `parseMDURL` | 2 | MangaDex title URL extracts chapter UUID; URL with slug after ID |
| `parseTMDBURL` | 3 | Movie URL extracts ID and `movie` category; TV URL extracts ID and `tv` category; URL without slug |
| `cleanYTURL` | 3 | Short `youtu.be` URL; full `youtube.com/watch?v=` URL; `youtube.com` without `www` — all extract same video ID |
| `escapeText` | 4 | Newlines escaped to `\n` literals; no-newline passthrough; multiple newlines; empty string |

## Bugs found during testing

`ConvertEmoji` in `utils/emojiUnicode.go` panicked on any emoji. The original code used JavaScript surrogate-pair math (`runeValue[0] + runeValue[1]`) but Go's `[]rune` decodes UTF-8 directly to Unicode code points — emoji are a single rune, not two. Fixed to use `runeValue[0]` directly.
