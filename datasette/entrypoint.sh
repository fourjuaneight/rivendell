#!/bin/sh
# Entrypoint: serve a read-only snapshot of the PocketBase DB, refreshed periodically.
#
# Datasette loads its databases once at startup and (with --immutable) assumes the
# file never changes. So we can't overwrite the live file underneath it. Instead we
# resync, then (re)start Datasette against the fresh snapshot, looping every 5 min.
set -e

SRC="/data/pb_data/data.db"
DEST="/data/rivendell.db"

# Wait for PocketBase to create its DB (it may still be running migrations).
until [ -f "$SRC" ]; do
  echo "waiting for $SRC ..."
  sleep 5
done

while true; do
  # Refresh the snapshot. If a resync fails, keep serving the previous one.
  /app/sync.sh || echo "sync failed, serving previous snapshot"

  # Never start datasette on a missing snapshot: --immutable on a non-existent file
  # errors out and the loop would spin silently. Retry the sync instead.
  if [ ! -f "$DEST" ]; then
    echo "no snapshot available yet; retrying in 30s"
    sleep 30
    continue
  fi

  # --immutable takes the DB path as its value (not a boolean), so the path must
  # follow it directly and must NOT also be passed positionally.
  # No base_url: datasette is served at the root of its own Tailscale port (8443),
  # so every link stays on the datasette host. (base_url + Tailscale's path-strip
  # mismatched, dropping the prefix from request-path links like facets -> 404.)
  datasette serve \
    --immutable "$DEST" \
    --host 0.0.0.0 \
    --port 8001 \
    --metadata /app/metadata.yml &
  DSPID=$!

  # Serve for 5 minutes, then stop to pick up a fresh snapshot.
  sleep 300
  kill "$DSPID" 2>/dev/null || true
  wait "$DSPID" 2>/dev/null || true
done
