#!/usr/bin/env python3
"""Replace meta record IDs with their names in the Datasette snapshot.

PocketBase relation fields store meta record IDs. Datasette renders the raw stored
value in table cells AND in facets/filters, and there's no hook to relabel facet
values — so the only way to show names in facets is to put names in the column.
This rewrites the snapshot in place: `tags` (JSON array of IDs) -> array of names;
`genre`/`platform`/`status`/`definition` (single ID) -> name.

It runs only on the read-only analytics copy (never the live PocketBase DB), so
dropping the raw IDs here is fine. Canned queries then group by the names directly
(no `meta` join needed).
"""
import json
import os
import sqlite3

# Tables whose `tags` column holds a JSON array of meta IDs.
ARRAY_TABLES = ["articles", "podcasts", "videos", "feeds", "read_later", "watch_later"]

# Tables -> single-relation columns holding one meta ID.
SINGLE_COLS = {
    "books": ["genre", "status"],
    "cds": ["genre"],
    "games": ["genre", "platform", "status"],
    "movies": ["genre", "definition", "status"],
    "shows": ["genre", "definition", "status"],
    "vinyls": ["genre"],
}


def main():
    db = sqlite3.connect(os.environ["SNAPSHOT"])
    db.row_factory = sqlite3.Row
    meta = {r["id"]: r["name"] for r in db.execute("SELECT id, name FROM meta")}

    def table_exists(t):
        return db.execute(
            "SELECT 1 FROM sqlite_master WHERE type='table' AND name=?", (t,)
        ).fetchone() is not None

    def has_col(t, c):
        return c in {row[1] for row in db.execute(f'PRAGMA table_info("{t}")')}

    # tags: JSON array of IDs -> JSON array of names (unknown IDs kept as-is)
    for t in ARRAY_TABLES:
        if not (table_exists(t) and has_col(t, "tags")):
            continue
        updates = []
        for row in db.execute(f'SELECT rowid AS rid, tags FROM "{t}"'):
            raw = row["tags"]
            if not raw:
                continue
            try:
                ids = json.loads(raw)
            except (TypeError, ValueError):
                continue
            if not isinstance(ids, list):
                continue
            updates.append((json.dumps([meta.get(i, i) for i in ids]), row["rid"]))
        if updates:
            db.executemany(f'UPDATE "{t}" SET tags = ? WHERE rowid = ?', updates)

    # single relations: ID -> name (only rows whose value is a known meta ID)
    for t, cols in SINGLE_COLS.items():
        if not table_exists(t):
            continue
        for c in cols:
            if not has_col(t, c):
                continue
            db.execute(
                f'UPDATE "{t}" SET "{c}" = (SELECT name FROM meta WHERE id = "{t}"."{c}") '
                f'WHERE "{c}" IN (SELECT id FROM meta)'
            )

    db.commit()
    db.close()


if __name__ == "__main__":
    main()
