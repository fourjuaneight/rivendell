# Rivendell

Personal bookmarking and archiving database powered by [PocketBase](https://pocketbase.io/).

## Docs

- [SCHEMA.md](SCHEMA.md) — collection definitions, field types, and constraints
- [API.md](API.md) — how to create and query records; which fields are manual vs. auto-filled
- [MIGRATIONS.md](MIGRATIONS.md) — how schema migrations work and how to write new ones
- [TESTING.md](TESTING.md) — what's tested, how to run tests, and bugs found during testing

---

## Requirements

- Go ≥ 1.25
- Docker + Docker Compose (for containerized runs)
- Tailscale (for production networking)
- Access to the external APIs used by the helpers:
  - Backblaze B2: `B2_APP_KEY_ID`, `B2_APP_KEY`, `B2_BUCKET_ID`, `B2_BUCKET_NAME`
  - GitHub GraphQL: `GH_TOKEN`, `GH_USERNAME`
  - The Movie Database: `TMDB_KEY`
  - YouTube Data API v3: `YOUTUBE_KEY`
  - PocketBase meta collection ID: `META_ID`
  - Tailscale auth key: `TS_AUTHKEY`

## Setup

1. Clone the repository and move into it.
2. Create a `.env` file in the project root with all required variables:

```sh
cat <<'EOF' > .env
B2_APP_KEY_ID=
B2_APP_KEY=
B2_BUCKET_ID=
B2_BUCKET_NAME=
GH_TOKEN=
GH_USERNAME=
TMDB_KEY=
YOUTUBE_KEY=
META_ID=
TS_AUTHKEY=
EOF
```

`META_ID` must be a valid 15-character PocketBase ID. Generate one:
```sh
cat /dev/urandom | LC_ALL=C tr -dc 'a-z0-9' | head -c 15
```

See [SCHEMA.md](SCHEMA.md) for why this value is required.

3. Pull Go dependencies:

```sh
go mod download
```

## Local development

Run the PocketBase app directly:

```sh
go run . serve
```

- HTTP API listens on `http://127.0.0.1:8090` by default.
- Schema migrations run automatically on startup.
- Data stored in `pb_data/` directory.

See [API.md](API.md) for request examples.

## Docker workflow

Build and start the stack (PocketBase + Tailscale):

```sh
docker compose up --build
```

- PocketBase binds to port `8090` on the internal Docker network.
- Tailscale proxies HTTPS traffic from your tailnet to the app — no public ports exposed.
- Persistent data lives in the `pb_data` bind mount.

Stop and remove the containers:

```sh
docker compose down
```

## Migrations

Schema is managed via versioned migration files in `migrations/`. They run automatically on `serve` startup — no manual steps needed. See [MIGRATIONS.md](MIGRATIONS.md) for how to write new ones.

## Deployment

The repo ships with a convenience script that pulls the latest code and rebuilds the Docker services:

```sh
./deploy.sh
```

Run it from the server where the stack should stay up-to-date.
