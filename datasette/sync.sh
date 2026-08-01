#!/bin/sh
# Snapshot the live PocketBase SQLite DB into a standalone copy for Datasette.
#
# We never let sqlite open the live DB directly: PocketBase holds locks on it and
# the pb_data mount is read-only, so an in-place ".backup" fails with
# "database is locked". Instead we copy the file-set (data.db + WAL sidecars) to a
# writable temp dir, then VACUUM INTO a clean single-file DB from that private copy.
# This avoids the lock entirely and never writes to (or even opens) your live data.
#
# The snapshot is then stripped of system and auth tables (see below) — Datasette
# serves it to the whole tailnet, so it must never contain credentials.

set -e

SRC="/data/pb_data/data.db"
DEST="/data/rivendell.db"
TMP="/tmp/pbsnap"

if [ ! -f "$SRC" ]; then
  echo "Source DB not found: $SRC"
  exit 1
fi

# Stage a private copy of the DB and its WAL sidecars (sqlite replays the -wal on
# the copy to reach a consistent state). -wal/-shm may not exist; that's fine.
rm -rf "$TMP"
mkdir -p "$TMP"
cp "$SRC" "$TMP/data.db"
[ -f "$SRC-wal" ] && cp "$SRC-wal" "$TMP/data.db-wal" || true
[ -f "$SRC-shm" ] && cp "$SRC-shm" "$TMP/data.db-shm" || true

# Snapshot the copy. VACUUM INTO requires the target not to exist.
rm -f "$TMP/raw.db" "$DEST.tmp"
sqlite3 "$TMP/data.db" "VACUUM INTO '$TMP/raw.db'"

# Strip everything that isn't analytics data before the snapshot is served.
# Datasette is reachable by anyone on the tailnet, and a raw copy of data.db
# carries secrets: _params holds the app settings (B2/S3 credentials in the
# clear unless PB_ENCRYPTION_KEY is set), _superusers and users hold password
# hashes and token keys, and _collections holds any OAuth2 client secrets.
# Same policy as backup/backup.go: drop system (underscore-prefixed) and auth
# collections, keep the rest.
sqlite3 "$TMP/raw.db" \
  "SELECT 'DROP TABLE IF EXISTS \"' || name || '\";'
     FROM sqlite_master
    WHERE type = 'table'
      AND (name LIKE '\_%' ESCAPE '\' OR name = 'users');" \
  | sqlite3 "$TMP/raw.db"

# Second VACUUM INTO, not an in-place VACUUM: DROP TABLE only moves pages to the
# freelist, so the dropped rows are still readable as raw bytes in the file. This
# writes a fresh file containing only live pages.
sqlite3 "$TMP/raw.db" "VACUUM INTO '$DEST.tmp'"

# Fail loudly rather than serve a snapshot that still has secrets in it.
for t in _params _superusers users; do
  if [ -n "$(sqlite3 "$DEST.tmp" "SELECT 1 FROM sqlite_master WHERE type='table' AND name='$t';")" ]; then
    echo "refusing to publish: sensitive table '$t' survived the strip"
    rm -f "$DEST.tmp"
    exit 1
  fi
done

# Denormalize relation columns (meta IDs -> names) so Datasette shows names in
# table cells AND facets/filters. Done on the temp file before the atomic swap.
SNAPSHOT="$DEST.tmp" python3 /app/denormalize.py

# Swap atomically so Datasette never sees a partial/pre-denormalized file.
mv -f "$DEST.tmp" "$DEST"
rm -rf "$TMP"

echo "Synced $(date -Iseconds)"
