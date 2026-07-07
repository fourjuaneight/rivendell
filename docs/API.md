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

Pass as `Authorization: Bearer {token}` on all read/update requests.

## Relation name resolution

For `genre`, `definition`, `platform`, and `status` fields, pass the **name string** (e.g. `"rock"`, `"4k"`, `"ps5"`, `"in_progress"`). The server looks up the matching `meta` record and replaces it with the ID before saving. Passing a raw meta ID does **not** work — it fails to match any name and the create is rejected with an error.

For `tags` fields on `articles`, `podcasts`, `videos`, `feeds`, `read_later`, and `watch_later`, pass an array of **tag name strings** (e.g. `["programming", "go"]`). The server resolves each to its `meta` record ID before saving. Every name must match an existing `meta` record of type `tags` — if **any** name is unknown (including a raw meta ID, which matches no name), the create is rejected with an error listing the unmatched name(s). Create the tag via the `meta` collection first (see below).

## Collections

### Auto-enriched on create

The server fetches and fills additional fields automatically after the record is saved.

#### articles

Send: `title`, `creator`, `url`, `tags` (names) — optionally `year`, `comments`
Server sets: `dead = false`, `shared = false`, `favorite = false`, `archive` (Markdown content archived to B2; a SingleFile HTML snapshot is also uploaded to B2 as a best-effort extra)

```sh
curl -X POST '{BASE_URL}/api/collections/articles/records' \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Some Article",
    "creator": "Author Name",
    "url": "https://example.com/article",
    "tags": ["programming", "go"]
  }'
```

```js
const res = await fetch(`${BASE_URL}/api/collections/articles/records`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    title: 'Some Article',
    creator: 'Author Name',
    url: 'https://example.com/article',
    tags: ['programming', 'go'],
  }),
});
const record = await res.json();
```

#### podcasts

Send: `title`, `creator`, `url`, `tags` (names) — optionally `year`, `comments`
Server sets: `dead = false`, `shared = false`, `favorite = false`, `archive` (episode archived to B2). Enrichment is skipped if `archive` is already set on create.

```sh
curl -X POST '{BASE_URL}/api/collections/podcasts/records' \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Some Episode",
    "creator": "Show Name",
    "url": "https://example.com/episode.mp3",
    "tags": ["tech"]
  }'
```

```js
const res = await fetch(`${BASE_URL}/api/collections/podcasts/records`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    title: 'Some Episode',
    creator: 'Show Name',
    url: 'https://example.com/episode.mp3',
    tags: ['tech'],
  }),
});
const record = await res.json();
```

#### videos

Send: `url`, `tags` (names) — optionally `comments`
Server sets: `title`, `creator`, `url`, `year` (from YouTube), `dead = false`, `shared = false`, `favorite = false`, `archive` (video downloaded and archived to B2). Enrichment is skipped if `archive` is already set on create.

```sh
curl -X POST '{BASE_URL}/api/collections/videos/records' \
  -H 'Content-Type: application/json' \
  -d '{
    "url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
    "tags": ["music"]
  }'
```

```js
const res = await fetch(`${BASE_URL}/api/collections/videos/records`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    url: 'https://www.youtube.com/watch?v=dQw4w9WgXcQ',
    tags: ['music'],
  }),
});
const record = await res.json();
```

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
Server sets: all card fields + `image` and `back` (uploaded to B2) from Scryfall

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

#### books

Send: `title`, `author` — optionally `isbn`, `genre` (name), `status` (name), `year`, `comments`
Server sets: `year` (from OpenLibrary if ISBN provided), `cover` (B2 URL), `favorite = false`, `status` (defaults to `not_started` if omitted)

```sh
curl -X POST '{BASE_URL}/api/collections/books/records' \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Nineteen Eighty-Four",
    "author": "George Orwell",
    "isbn": "9780451524935",
    "genre": "fiction"
  }'
```

```js
const res = await fetch(`${BASE_URL}/api/collections/books/records`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    title: 'Nineteen Eighty-Four',
    author: 'George Orwell',
    isbn: '9780451524935',
    genre: 'fiction',
  }),
});
const record = await res.json();
```

#### cds

Send: `album`, `artist` — optionally `barcode`, `genre` (name), `year`, `comments`
Server sets: `year` and `cover` (from Discogs, B2 URL)

```sh
curl -X POST '{BASE_URL}/api/collections/cds/records' \
  -H 'Content-Type: application/json' \
  -d '{
    "album": "Lateralus",
    "artist": "Tool",
    "barcode": "0828768199121",
    "genre": "rock"
  }'
```

```js
const res = await fetch(`${BASE_URL}/api/collections/cds/records`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    album: 'Lateralus',
    artist: 'Tool',
    barcode: '0828768199121',
    genre: 'rock',
  }),
});
const record = await res.json();
```

#### games

Send: `title` — optionally `publisher`, `barcode`, `genre` (name), `platform` (name), `status` (name), `year`, `comments`
Server sets: `year` and `cover` (from IGDB, B2 URL), `favorite = false`, `status` (defaults to `not_started` if omitted)

```sh
curl -X POST '{BASE_URL}/api/collections/games/records' \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Hollow Knight",
    "genre": "action",
    "platform": "pc"
  }'
```

```js
const res = await fetch(`${BASE_URL}/api/collections/games/records`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    title: 'Hollow Knight',
    genre: 'action',
    platform: 'pc',
  }),
});
const record = await res.json();
```

#### movies

Send: `title` — optionally `director`, `barcode`, `genre` (name), `definition` (name), `status` (name), `year`, `comments`
Server sets: `year` and `cover` (from TMDB, B2 URL), `favorite = false`, `status` (defaults to `not_started` if omitted)

```sh
curl -X POST '{BASE_URL}/api/collections/movies/records' \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Blade Runner 2049",
    "director": "Denis Villeneuve",
    "genre": "sci-fi",
    "definition": "4k"
  }'
```

```js
const res = await fetch(`${BASE_URL}/api/collections/movies/records`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    title: 'Blade Runner 2049',
    director: 'Denis Villeneuve',
    genre: 'sci-fi',
    definition: '4k',
  }),
});
const record = await res.json();
```

#### shows

Send: `title` — optionally `director`, `barcode`, `genre` (name), `definition` (name), `status` (name), `season`, `year`, `comments`
Server sets: `year` and `cover` (from TMDB season poster if `season` provided, else show poster, B2 URL), `favorite = false`, `status` (defaults to `not_started` if omitted)

```sh
curl -X POST '{BASE_URL}/api/collections/shows/records' \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Severance",
    "genre": "drama",
    "definition": "1080p"
  }'
```

```js
const res = await fetch(`${BASE_URL}/api/collections/shows/records`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    title: 'Severance',
    genre: 'drama',
    definition: '1080p',
  }),
});
const record = await res.json();
```

#### vinyls

Send: `album`, `artist` — optionally `barcode`, `genre` (name), `year`, `comments`
Server sets: `year` and `cover` (from Discogs, B2 URL)

```sh
curl -X POST '{BASE_URL}/api/collections/vinyls/records' \
  -H 'Content-Type: application/json' \
  -d '{
    "album": "Kind of Blue",
    "artist": "Miles Davis",
    "genre": "jazz"
  }'
```

```js
const res = await fetch(`${BASE_URL}/api/collections/vinyls/records`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    album: 'Kind of Blue',
    artist: 'Miles Davis',
    genre: 'jazz',
  }),
});
const record = await res.json();
```

#### watch_later

Send: `link`, `tags` (names) — `title` and `channel` are auto-filled from YouTube
Server sets: `title`, `channel` (from YouTube)

```sh
curl -X POST '{BASE_URL}/api/collections/watch_later/records' \
  -H 'Content-Type: application/json' \
  -d '{
    "link": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
    "tags": ["music"]
  }'
```

```js
const res = await fetch(`${BASE_URL}/api/collections/watch_later/records`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    link: 'https://www.youtube.com/watch?v=dQw4w9WgXcQ',
    tags: ['music'],
  }),
});
const record = await res.json();
```

### Manual entry

All fields must be provided by the caller.

#### feeds

Send: `title`, `url`, `type`, `tags` (names) — optionally `rss`, `comments`
Server sets: `dead = false`, `shared = false`

```sh
curl -X POST '{BASE_URL}/api/collections/feeds/records' \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Feed Name",
    "url": "https://example.com",
    "rss": "https://example.com/feed.xml",
    "type": "website",
    "tags": ["tech"]
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
    type: 'website',
    tags: ['tech'],
  }),
});
const record = await res.json();
```

`type` options: `podcast` · `website` · `youtube`

#### read_later

Send: `title`, `link`, `tags` (names)

```sh
curl -X POST '{BASE_URL}/api/collections/read_later/records' \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Some Article",
    "link": "https://example.com/article",
    "tags": ["programming"]
  }'
```

```js
const res = await fetch(`${BASE_URL}/api/collections/read_later/records`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    title: 'Some Article',
    link: 'https://example.com/article',
    tags: ['programming'],
  }),
});
const record = await res.json();
```

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

Lookup values for `tags`, `genre`, `definition`, `platform`, and `status` fields used by other collections.

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

`type` options: `definition` · `genre` · `platform` · `status` · `tags`

You reference these by **name**, not by the returned `id`. For `tags` on `articles`, `podcasts`, `videos`, `feeds`, `read_later`, and `watch_later`, pass the tag name strings created here — the server resolves them to IDs. Likewise, pass `genre`, `definition`, `platform`, and `status` name strings directly on the media collections that use them.

## Updating records

All updates require auth.

```sh
curl -X PATCH '{BASE_URL}/api/collections/{collection}/records/{id}' \
  -H 'Authorization: Bearer {token}' \
  -H 'Content-Type: application/json' \
  -d '{"field": "new_value"}'
```

```js
const res = await fetch(`${BASE_URL}/api/collections/${collection}/records/${id}`, {
  method: 'PATCH',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  },
  body: JSON.stringify({ field: 'new_value' }),
});
const updated = await res.json();
```

Common update use cases:
- `articles` / `podcasts` / `videos` / `feeds`: toggle `dead` or `shared`
- `articles` / `podcasts` / `videos` / `books` / `games` / `movies` / `shows`: toggle `favorite`
- `books` / `games` / `movies` / `shows`: update `status` (pass a `status` meta name)
- `articles`: update `comments`
- `records`: set `end` date when leaving a position

## Listing / querying records

```sh
curl '{BASE_URL}/api/collections/{collection}/records?filter=genre%3D"rock"&sort=-created&page=1&perPage=30' \
  -H 'Authorization: Bearer {token}'
```

```js
const params = new URLSearchParams({
  filter: 'genre = "rock"',
  sort: '-created',
  page: 1,
  perPage: 30,
});
const res = await fetch(`${BASE_URL}/api/collections/${collection}/records?${params}`, {
  headers: { 'Authorization': `Bearer ${token}` },
});
const { items, totalItems, totalPages } = await res.json();
```

Filter operators: `=` `!=` `>` `<` `>=` `<=` `~` (contains) `!~` (not contains). Combine with `&&` / `||`.
Sort prefix `-` = descending (e.g. `-created` = newest first).
