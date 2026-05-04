# API Usage

Base URL: `http://127.0.0.1:8090`

All endpoints require auth except record creation (create rule is open).

## Authentication

```sh
curl -X POST '{BASE_URL}/api/collections/users/auth-with-password' \
  -H 'Content-Type: application/json' \
  -d '{"identity": "you@example.com", "password": "yourpassword"}'
```

Returns `token`. Pass as `Authorization: Bearer {token}` on all subsequent requests.

## Collections

### Auto-enriched on create

These collections require minimal input — the server fetches and fills remaining fields automatically.

#### bookmarks

Send: `title`, `creator`, `url`, `type`, `tags`, `comments` (optional)
Server sets: `dead = false`, `shared = false`, `archive` (fetches content, uploads to B2)

```sh
curl -X POST '{BASE_URL}/api/collections/bookmarks/records' \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Some Article",
    "creator": "Author Name",
    "url": "https://example.com/article",
    "type": "articles",
    "tags": ["meta_record_id_1", "meta_record_id_2"]
  }'
```

`type` options: `articles` · `podcasts` · `videos`

#### github

Send: `url` only
Server sets: `name`, `owner`, `description`, `language` (fetched from GitHub API)

```sh
curl -X POST '{BASE_URL}/api/collections/github/records' \
  -H 'Content-Type: application/json' \
  -d '{"url": "https://github.com/owner/repo"}'
```

#### mtg

Send: `name`, `set` (set code), `collector_number`
Server sets: `colors`, `type`, `set_name`, `oracle_text`, `flavor_text`, `rarity`, `artist`, `released_at`, `image`, `back` (fetched from Scryfall)

```sh
curl -X POST '{BASE_URL}/api/collections/mtg/records' \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Black Lotus",
    "set": "lea",
    "collector_number": 233
  }'
```

### Full manual entry

All fields must be provided by the caller.

#### feeds

Send: `title`, `url`, `type`, `tags` — and optionally `rss`, `comments`
Server sets: `dead = false`, `shared = false`

```sh
curl -X POST '{BASE_URL}/api/collections/feeds/records' \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Feed Name",
    "url": "https://example.com",
    "rss": "https://example.com/feed.xml",
    "type": "websites",
    "tags": ["meta_record_id_1"]
  }'
```

`type` options: `podcasts` · `websites` · `youtube`

#### media

```sh
curl -X POST '{BASE_URL}/api/collections/media/records' \
  -H 'Authorization: Bearer {token}' \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Album Title",
    "creator": "Artist Name",
    "genre": "meta_record_id",
    "year": 1991,
    "type": "cds"
  }'
```

`type` options: `books` · `cds` · `games` · `movies` · `shows` · `vinyls`

#### records

```sh
curl -X POST '{BASE_URL}/api/collections/records/records' \
  -H 'Authorization: Bearer {token}' \
  -H 'Content-Type: application/json' \
  -d '{
    "company": "Acme Corp",
    "position": "Engineer",
    "stack": ["Go", "Postgres"],
    "start": "2022-01-01 00:00:00",
    "end": "2024-06-01 00:00:00"
  }'
```

`end` is optional (omit for current position).

#### meta

Lookup values for `tags` and `genre` fields used by other collections.

```sh
curl -X POST '{BASE_URL}/api/collections/meta/records' \
  -H 'Content-Type: application/json' \
  -d '{"name": "programming", "type": "tags"}'
```

`type` options: `tags` · `genre`

The `id` returned here is what you pass as the relation value in `bookmarks.tags`, `feeds.tags`, `media.genre`.

## Updating records

All updates require auth.

```sh
curl -X PATCH '{BASE_URL}/api/collections/{collection}/records/{id}' \
  -H 'Authorization: Bearer {token}' \
  -H 'Content-Type: application/json' \
  -d '{"field": "new_value"}'
```

Common update use cases:
- `bookmarks` / `feeds`: toggle `dead` or `shared`
- `bookmarks`: update `comments`
- `records`: set `end` date when leaving a position

## Listing / querying records

```sh
curl '{BASE_URL}/api/collections/{collection}/records' \
  -H 'Authorization: Bearer {token}'
```

Supports `?filter=`, `?sort=`, `?page=`, `?perPage=` query params per PocketBase docs.
