# API Usage

Base URL: `http://127.0.0.1:8090`

All endpoints require auth except record creation (create rule is open).

## Authentication

PocketBase uses **superuser impersonate tokens** as API keys. These are non-renewable (non-expiring) and grant full access — use only for internal server-to-server calls.

### Get a token

**Option 1 — Admin UI:** `/_/` → Collections → `_superusers` → select user → Impersonate → copy token

**Option 2 — API** (requires an existing superuser session token):

```sh
curl -X POST '{BASE_URL}/api/collections/_superusers/impersonate/{superuserId}' \
  -H 'Authorization: {superuser-session-token}' \
  -H 'Content-Type: application/json' \
  -d '{"duration": 0}'
```

```js
const res = await fetch(`${BASE_URL}/api/collections/_superusers/impersonate/${superuserId}`, {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': superuserSessionToken,
  },
  body: JSON.stringify({ duration: 0 }),
});
const { token } = await res.json();
```

`duration: 0` = no expiry.

### Use the token

Pass as `Authorization: {token}` — no `Bearer` prefix — on all read/update requests.

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

```js
const res = await fetch(`${BASE_URL}/api/collections/bookmarks/records`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    title: 'Some Article',
    creator: 'Author Name',
    url: 'https://example.com/article',
    type: 'articles',
    tags: ['meta_record_id_1', 'meta_record_id_2'],
  }),
});
const record = await res.json();
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

```js
const res = await fetch(`${BASE_URL}/api/collections/github/records`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ url: 'https://github.com/owner/repo' }),
});
const record = await res.json();
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

```js
const res = await fetch(`${BASE_URL}/api/collections/mtg/records`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ name: 'Black Lotus', set: 'lea', collector_number: 233 }),
});
const record = await res.json();
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

```js
const res = await fetch(`${BASE_URL}/api/collections/feeds/records`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    title: 'Feed Name',
    url: 'https://example.com',
    rss: 'https://example.com/feed.xml',
    type: 'websites',
    tags: ['meta_record_id_1'],
  }),
});
const record = await res.json();
```

`type` options: `podcasts` · `websites` · `youtube`

#### media

```sh
curl -X POST '{BASE_URL}/api/collections/media/records' \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Album Title",
    "creator": "Artist Name",
    "genre": "meta_record_id",
    "year": 1991,
    "type": "cds"
  }'
```

```js
const res = await fetch(`${BASE_URL}/api/collections/media/records`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    title: 'Album Title',
    creator: 'Artist Name',
    genre: 'meta_record_id',
    year: 1991,
    type: 'cds',
  }),
});
const record = await res.json();
```

`type` options: `books` · `cds` · `games` · `movies` · `shows` · `vinyls`

#### records

```sh
curl -X POST '{BASE_URL}/api/collections/records/records' \
  -H 'Content-Type: application/json' \
  -d '{
    "company": "Acme Corp",
    "position": "Engineer",
    "stack": ["Go", "Postgres"],
    "start": "2022-01-01 00:00:00",
    "end": "2024-06-01 00:00:00"
  }'
```

```js
const res = await fetch(`${BASE_URL}/api/collections/records/records`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    company: 'Acme Corp',
    position: 'Engineer',
    stack: ['Go', 'Postgres'],
    start: '2022-01-01 00:00:00',
    end: '2024-06-01 00:00:00',
  }),
});
const record = await res.json();
```

`end` is optional (omit for current position).

#### meta

Lookup values for `tags` and `genre` fields used by other collections.

```sh
curl -X POST '{BASE_URL}/api/collections/meta/records' \
  -H 'Content-Type: application/json' \
  -d '{"name": "programming", "type": "tags"}'
```

```js
const res = await fetch(`${BASE_URL}/api/collections/meta/records`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ name: 'programming', type: 'tags' }),
});
const record = await res.json();
```

`type` options: `tags` · `genre`

The `id` returned here is what you pass as the relation value in `bookmarks.tags`, `feeds.tags`, `media.genre`.

## Updating records

All updates require auth.

```sh
curl -X PATCH '{BASE_URL}/api/collections/{collection}/records/{id}' \
  -H 'Authorization: {token}' \
  -H 'Content-Type: application/json' \
  -d '{"field": "new_value"}'
```

```js
const res = await fetch(`${BASE_URL}/api/collections/${collection}/records/${id}`, {
  method: 'PATCH',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': token,
  },
  body: JSON.stringify({ field: 'new_value' }),
});
const updated = await res.json();
```

Common update use cases:
- `bookmarks` / `feeds`: toggle `dead` or `shared`
- `bookmarks`: update `comments`
- `records`: set `end` date when leaving a position

## Listing / querying records

```sh
curl '{BASE_URL}/api/collections/{collection}/records?filter=type%3D"movies"&sort=-created&page=1&perPage=30' \
  -H 'Authorization: {token}'
```

```js
const params = new URLSearchParams({
  filter: 'type = "movies"',
  sort: '-created',
  page: 1,
  perPage: 30,
});
const res = await fetch(`${BASE_URL}/api/collections/${collection}/records?${params}`, {
  headers: { 'Authorization': token },
});
const { items, totalItems, totalPages } = await res.json();
```

Filter operators: `=` `!=` `>` `<` `>=` `<=` `~` (contains) `!~` (not contains). Combine with `&&` / `||`.
Sort prefix `-` = descending (e.g. `-created` = newest first).
