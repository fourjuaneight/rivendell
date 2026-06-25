"""Datasette render_cell plugin: show meta names instead of raw record IDs.

PocketBase relation fields store meta record IDs, so table views show opaque IDs
like ["ty8l2szit8dk7vw", ...]. This resolves them to the human-readable `name`
from the `meta` collection at display time.

Display-only: filters, facets, and the JSON API still use the raw IDs, so nothing
else changes.
"""

import json

from datasette import hookimpl

# Columns holding a JSON array of meta record IDs (multi-select relations).
ARRAY_RELATIONS = {"tags"}
# Columns holding a single meta record ID (single-select relations).
SINGLE_RELATIONS = {"genre", "platform", "status", "definition"}

# {database_name: {meta_id: name}}. The snapshot is immutable and the datasette
# process restarts every ~5 min, so a per-process cache stays fresh.
_meta_cache = {}


async def _meta_map(datasette, database):
    if database not in _meta_cache:
        try:
            rows = await datasette.get_database(database).execute(
                "select id, name from meta"
            )
            _meta_cache[database] = {row["id"]: row["name"] for row in rows}
        except Exception:
            # No meta table in this database — cache empty so we fall back to IDs.
            _meta_cache[database] = {}
    return _meta_cache[database]


@hookimpl
def render_cell(value, column, table, database, datasette):
    if column in ARRAY_RELATIONS:

        async def render_array():
            if not value:
                return value
            try:
                ids = json.loads(value)
            except (TypeError, ValueError):
                return value
            if not isinstance(ids, list):
                return value
            names = await _meta_map(datasette, database)
            return ", ".join(str(names.get(i, i)) for i in ids)

        return render_array()

    if column in SINGLE_RELATIONS:

        async def render_single():
            if not value:
                return value
            names = await _meta_map(datasette, database)
            return names.get(value, value)

        return render_single()

    # Not a relation column — let datasette render it normally.
    return None
