# Database Schema

All collections use PocketBase's built-in `id`, `created`, and `updated` fields.

Access rules unless noted: view/update require auth (`@request.auth.id != ''`), create is open.

## meta

Lookup table for tags, genres, definitions, and platforms. Referenced by `bookmarks`, `feeds`, `books`, `cds`, `games`, `movies`, `shows`, and `vinyls`.

> ID is pinned via `META_ID` env var so relation fields can reference it at migration time.

| Field  | Type   | Required | Constraints                                        |
|--------|--------|----------|----------------------------------------------------|
| `name` | text   | yes      |                                                    |
| `type` | select | no       | `definition`, `genre`, `platform`, `tags` (max: 1) |

## bookmarks

Saved articles, podcasts, and videos. Archived to Backblaze B2 on create.

| Field      | Type     | Required | Constraints                               |
|------------|----------|----------|-------------------------------------------|
| `title`    | text     | yes      |                                           |
| `creator`  | text     | yes      |                                           |
| `url`      | url      | yes      |                                           |
| `archive`  | url      | no       | Set automatically on create               |
| `tags`     | relation | yes      | → `meta`, max 5                           |
| `type`     | select   | yes      | `articles`, `podcasts`, `videos` (max: 1) |
| `dead`     | bool     | no       | Defaults to `false` on create             |
| `shared`   | bool     | no       | Defaults to `false` on create             |
| `comments` | text     | no       |                                           |

## feeds

RSS/podcast/YouTube feeds.

| Field      | Type     | Required | Constraints                                |
|------------|----------|----------|--------------------------------------------|
| `title`    | text     | yes      |                                            |
| `url`      | url      | yes      |                                            |
| `rss`      | url      | no       |                                            |
| `tags`     | relation | yes      | → `meta`, max 5                            |
| `type`     | select   | yes      | `podcasts`, `websites`, `youtube` (max: 1) |
| `dead`     | bool     | no       | Defaults to `false` on create              |
| `shared`   | bool     | no       | Defaults to `false` on create              |
| `comments` | text     | no       |                                            |

## books

Physical/digital book collection. Cover and year enriched from OpenLibrary on create.

| Field      | Type     | Required | Constraints                       |
|------------|----------|----------|-----------------------------------|
| `title`    | text     | yes      |                                   |
| `author`   | text     | yes      |                                   |
| `isbn`     | text     | no       | Used for OpenLibrary lookup       |
| `genre`    | relation | no       | → `meta` (type: `genre`), max 1   |
| `year`     | number   | no       | Set automatically if ISBN present |
| `cover`    | url      | no       | Set automatically (B2 URL)        |
| `comments` | text     | no       |                                   |

## cds

CD collection. Cover and year enriched from Discogs on create.

| Field      | Type     | Required | Constraints                     |
|------------|----------|----------|---------------------------------|
| `album`    | text     | yes      |                                 |
| `artist`   | text     | yes      |                                 |
| `barcode`  | text     | no       | Used for Discogs lookup         |
| `genre`    | relation | no       | → `meta` (type: `genre`), max 1 |
| `year`     | number   | no       | Set automatically               |
| `cover`    | url      | no       | Set automatically (B2 URL)      |
| `comments` | text     | no       |                                 |

## games

Game collection. Cover and year enriched from IGDB on create.

| Field       | Type     | Required | Constraints                        |
|-------------|----------|----------|------------------------------------|
| `title`     | text     | yes      |                                    |
| `publisher` | text     | no       |                                    |
| `barcode`   | text     | no       |                                    |
| `genre`     | relation | no       | → `meta` (type: `genre`), max 1    |
| `platform`  | relation | no       | → `meta` (type: `platform`), max 1 |
| `year`      | number   | no       | Set automatically                  |
| `cover`     | url      | no       | Set automatically (B2 URL)         |
| `comments`  | text     | no       |                                    |

## movies

Movie collection. Cover and year enriched from TMDB on create.

| Field        | Type     | Required | Constraints                           |
|--------------|----------|----------|---------------------------------------|
| `title`      | text     | yes      |                                       |
| `director`   | text     | no       |                                       |
| `barcode`    | text     | no       |                                       |
| `genre`      | relation | no       | → `meta` (type: `genre`), max 1       |
| `definition` | relation | no       | → `meta` (type: `definition`), max 1  |
| `year`       | number   | no       | Set automatically                     |
| `cover`      | url      | no       | Set automatically (B2 URL)            |
| `comments`   | text     | no       |                                       |

## shows

TV show collection. Cover and year enriched from TMDB on create.

| Field        | Type     | Required | Constraints                           |
|--------------|----------|----------|---------------------------------------|
| `title`      | text     | yes      |                                       |
| `director`   | text     | no       |                                       |
| `barcode`    | text     | no       |                                       |
| `genre`      | relation | no       | → `meta` (type: `genre`), max 1       |
| `definition` | relation | no       | → `meta` (type: `definition`), max 1  |
| `year`       | number   | no       | Set automatically                     |
| `cover`      | url      | no       | Set automatically (B2 URL)            |
| `comments`   | text     | no       |                                       |

## vinyls

Vinyl record collection. Cover and year enriched from Discogs on create.

| Field      | Type     | Required | Constraints                     |
|------------|----------|----------|---------------------------------|
| `album`    | text     | yes      |                                 |
| `artist`   | text     | yes      |                                 |
| `barcode`  | text     | no       | Used for Discogs lookup         |
| `genre`    | relation | no       | → `meta` (type: `genre`), max 1 |
| `year`     | number   | no       | Set automatically               |
| `cover`    | url      | no       | Set automatically (B2 URL)      |
| `comments` | text     | no       |                                 |

## mtg

Magic: The Gathering card collection. Enriched from Scryfall on create. Card images uploaded to B2.

| Field              | Type   | Required | Constraints                     |
|--------------------|--------|----------|---------------------------------|
| `name`             | text   | yes      |                                 |
| `set`              | text   | yes      | Set code (e.g. `dmu`)           |
| `collector_number` | number | yes      |                                 |
| `colors`           | text   | no       | Set automatically from Scryfall |
| `type`             | text   | no       | Set automatically from Scryfall |
| `set_name`         | text   | no       | Set automatically from Scryfall |
| `oracle_text`      | text   | no       | Set automatically from Scryfall |
| `flavor_text`      | text   | no       | Set automatically from Scryfall |
| `rarity`           | text   | no       | Set automatically from Scryfall |
| `artist`           | text   | no       | Set automatically from Scryfall |
| `released_at`      | text   | no       | Set automatically from Scryfall |
| `image`            | text   | no       | Set automatically (B2 URL)      |
| `back`             | text   | no       | Double-faced cards only (B2 URL)|

## github

Saved GitHub repositories. Enriched from GitHub API on create.

| Field         | Type | Required | Constraints       |
|---------------|------|----------|-------------------|
| `url`         | text | yes      | Unique index      |
| `name`        | text | no       | Set automatically |
| `owner`       | text | no       | Set automatically |
| `description` | text | no       | Set automatically |
| `language`    | text | no       | Set automatically |

## records

Work history / employment records.

| Field      | Type | Required | Constraints |
|------------|------|----------|-------------|
| `company`  | text | yes      |             |
| `position` | text | no       |             |
| `stack`    | json | no       |             |
| `start`    | date | yes      |             |
| `end`      | date | no       |             |
