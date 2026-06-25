#!/bin/sh
# Snapshot the live PocketBase SQLite DB into a standalone copy for Datasette.
#
# We never let sqlite open the live DB directly: PocketBase holds locks on it and
# the pb_data mount is read-only, so an in-place ".backup" fails with
# "database is locked". Instead we copy the file-set (data.db + WAL sidecars) to a
# writable temp dir, then VACUUM INTO a clean single-file DB from that private copy.
# This avoids the lock entirely and never writes to (or even opens) your live data.

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
rm -f "$DEST.tmp"
sqlite3 "$TMP/data.db" "VACUUM INTO '$DEST.tmp'"

# Denormalize relation columns (meta IDs -> names) so Datasette shows names in
# table cells AND facets/filters. Done on the temp file before the atomic swap.
SNAPSHOT="$DEST.tmp" python3 /app/denormalize.py

# Swap atomically so Datasette never sees a partial/pre-denormalized file.
mv -f "$DEST.tmp" "$DEST"
rm -rf "$TMP"

echo "Synced $(date -Iseconds)"
