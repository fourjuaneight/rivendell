# Rivendell

Personal bookmarking and archiving database powered by [PocketBase](https://pocketbase.io/).

## Requirements

- Go ≥ 1.25
- Docker + Docker Compose (for containerized runs)
- Access to the external APIs used by the helpers:
  - Backblaze B2: `B2_APP_KEY_ID`, `B2_APP_KEY`, `B2_BUCKET_ID`, `B2_BUCKET_NAME`
  - GitHub GraphQL: `GH_TOKEN`, `GH_USERNAME`
  - The Movie Database: `TMDB_KEY`
  - YouTube Data API v3: `YOUTUBE_KEY`
  - PocketBase meta collection id: `META_ID`

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
EOF
```

3. Pull Go dependencies:

```sh
go mod download
```

## Local development

Run the PocketBase app directly:

```sh
go run . serve
```

- The HTTP API listens on `http://127.0.0.1:8090` by default.
- Data created through PocketBase UI or API will be stored inside the local `pb_data/` directory.

## Docker workflow

Build and start the stack (PocketBase + Caddy):

```sh
docker compose up --build
```

- The PocketBase service binds to `8090`, proxied by Caddy on ports `80/443`.
- Persistent data lives in the `pb_data` bind mount; keep it if you need to retain records.

Stop and remove the containers:

```sh
docker compose down
```

## Deployment

The repo ships with a convenience script that pulls the latest code and rebuilds the Docker services:

```sh
./deploy.sh
```

Run it from the server where the stack should stay up-to-date.