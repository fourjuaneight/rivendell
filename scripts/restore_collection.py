#!/usr/bin/env python3
"""Restore records that exist in a B2 backup but are missing from PocketBase.

Written for the 2026-07-19 incident, where 39 `games` rows (every row with
rowid > 89) were deleted by a superuser action and the loss wasn't noticed until
after the next backup ran. It is deliberately generic: the same failure can hit
any collection.

How it works
------------
Backups store relation fields as meta *names*, not IDs (see backup/backup.go),
which is exactly the shape the create API expects — the preparers in
preparers.go resolve names back to IDs. So restoring is a plain POST per record.

Two wrinkles the POST alone doesn't handle, both worked around by a follow-up
PATCH:

  * The preparers overwrite some fields unconditionally. `prepareGame` sets
    `favorite = false` on every create, so a favorited game comes back
    unfavorited.
  * The enrichers re-query the external API and overwrite what they find. For
    games that's `year` and `cover` from IGDB, which can differ from what was
    stored.

After each create the script re-reads the record and PATCHes back any
non-relation field whose value drifted from the backup. Relation fields are left
alone: the preparers already resolved them to the correct IDs, and the update
API wants IDs rather than the names held in the backup.

Matching is by a single natural-key field (`title` for most collections, mirroring
recordLabel in enrichers.go), so re-running the script is safe — anything already
present is skipped.

Usage
-----
    # See what would be restored (default; writes nothing)
    ./scripts/restore_collection.py games --date 2026-07-17

    # Actually restore
    ./scripts/restore_collection.py games --date 2026-07-17 --apply

    # From a file you already downloaded
    ./scripts/restore_collection.py games --file ./games-2026-07-17.json --apply

Reads POCKETBASE_URL, B2_* and (optionally) PB_TOKEN from .env. Without a valid
PB_TOKEN it prompts for superuser credentials.
"""

import argparse
import base64
import getpass
import json
import os
import sys
import time
import urllib.error
import urllib.parse
import urllib.request

B2_AUTH_URL = "https://api.backblazeb2.com/b2api/v3/b2_authorize_account"

# Field used to tell one record from another, mirroring recordLabel in enrichers.go.
KEY_FIELD = {
    "cds": "album",
    "vinyls": "album",
    "github": "url",
    "watch_later": "link",
    "mtg": "name",
}
DEFAULT_KEY_FIELD = "title"

# B2 folder names, mirroring pathMap in helpers/uploadToB2.go. Backups are always
# written under the collection's own name, so this only needs the prefix.
BACKUP_PREFIX = "Backups"


def load_env(path):
    """Parse a KEY=VALUE .env file. Existing environment variables win.

    Values may be wrapped in single or double quotes, which godotenv (used by
    the Go side) strips — do the same here so both readers see the same value.
    """
    if not os.path.isfile(path):
        return
    with open(path) as fh:
        for line in fh:
            line = line.strip()
            if not line or line.startswith("#") or "=" not in line:
                continue
            key, _, value = line.partition("=")
            value = value.strip()
            if len(value) >= 2 and value[0] == value[-1] and value[0] in "\"'":
                value = value[1:-1]
            os.environ.setdefault(key.strip(), value)


def request_json(url, method="GET", body=None, headers=None):
    data = None
    headers = dict(headers or {})
    if body is not None:
        data = json.dumps(body).encode()
        headers["Content-Type"] = "application/json"

    req = urllib.request.Request(url, data=data, method=method, headers=headers)
    try:
        with urllib.request.urlopen(req, timeout=60) as resp:
            raw = resp.read()
    except urllib.error.HTTPError as err:
        detail = err.read().decode("utf-8", "replace")
        raise SystemExit(f"{method} {url} failed: {err.code} {detail}") from err
    return json.loads(raw) if raw else None


# --- backup retrieval -------------------------------------------------------


def fetch_backup_from_b2(collection, date):
    key_id = os.environ["B2_APP_KEY_ID"]
    app_key = os.environ["B2_APP_KEY"]
    bucket = os.environ["B2_BUCKET_NAME"]

    basic = base64.b64encode(f"{key_id}:{app_key}".encode()).decode()
    auth = request_json(B2_AUTH_URL, headers={"Authorization": f"Basic {basic}"})

    download_url = auth["apiInfo"]["storageApi"]["downloadUrl"]
    token = auth["authorizationToken"]
    name = f"{BACKUP_PREFIX}/{collection}/{date}.json"
    url = f"{download_url}/file/{bucket}/{urllib.parse.quote(name)}"

    return request_json(url, headers={"Authorization": token})


# --- PocketBase -------------------------------------------------------------


def pb_authenticate(base_url):
    """Return a superuser token, preferring PB_TOKEN if it still validates."""
    token = os.environ.get("PB_TOKEN", "").strip()
    if token:
        try:
            request_json(
                f"{base_url}/api/collections/_superusers/auth-refresh",
                method="POST",
                headers={"Authorization": token},
            )
            return token
        except SystemExit:
            print("PB_TOKEN rejected; falling back to password auth.", file=sys.stderr)

    identity = input("superuser email: ").strip()
    password = getpass.getpass("superuser password: ")
    result = request_json(
        f"{base_url}/api/collections/_superusers/auth-with-password",
        method="POST",
        body={"identity": identity, "password": password},
    )
    return result["token"]


def fetch_collection_schema(base_url, token, collection):
    return request_json(
        f"{base_url}/api/collections/{collection}",
        headers={"Authorization": token},
    )


def fetch_all_records(base_url, token, collection):
    records, page = [], 1
    while True:
        url = (
            f"{base_url}/api/collections/{collection}/records"
            f"?page={page}&perPage=500&skipTotal=1"
        )
        batch = request_json(url, headers={"Authorization": token})["items"]
        records.extend(batch)
        if len(batch) < 500:
            return records
        page += 1


# --- restore ----------------------------------------------------------------

# Never sent back to the API: PocketBase manages these.
SYSTEM_FIELDS = {"id", "collectionId", "collectionName", "created", "updated", "expand"}


def drifted_fields(backup_record, live_record, relation_fields):
    """Fields whose stored value no longer matches the backup, ignoring relations."""
    drifted = {}
    for name, want in backup_record.items():
        if name in SYSTEM_FIELDS or name in relation_fields:
            continue
        if live_record.get(name) != want:
            drifted[name] = want
    return drifted


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("collection", help="collection to restore into, e.g. games")
    source = parser.add_mutually_exclusive_group(required=True)
    source.add_argument("--date", help="backup date on B2, e.g. 2026-07-17")
    source.add_argument("--file", help="path to a backup JSON file")
    parser.add_argument(
        "--key",
        help="field used to match backup records against live ones "
        "(default: title, or album/url/link/name depending on collection)",
    )
    parser.add_argument(
        "--apply",
        action="store_true",
        help="perform the restore; without it the script only reports",
    )
    parser.add_argument(
        "--delay",
        type=float,
        default=1.0,
        help="seconds to wait between creates, to be kind to the enrichers' "
        "external APIs (default: 1.0)",
    )
    args = parser.parse_args()

    repo_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    load_env(os.path.join(repo_root, ".env"))

    base_url = os.environ.get("POCKETBASE_URL", "").rstrip("/")
    if not base_url:
        raise SystemExit("POCKETBASE_URL is not set")

    key_field = args.key or KEY_FIELD.get(args.collection, DEFAULT_KEY_FIELD)

    if args.file:
        with open(args.file) as fh:
            backup = json.load(fh)
    else:
        backup = fetch_backup_from_b2(args.collection, args.date)

    token = pb_authenticate(base_url)
    schema = fetch_collection_schema(base_url, token, args.collection)
    relation_fields = {f["name"] for f in schema["fields"] if f["type"] == "relation"}

    live = fetch_all_records(base_url, token, args.collection)
    live_keys = {r.get(key_field) for r in live}

    missing = [r for r in backup if r.get(key_field) not in live_keys]

    print(f"collection : {args.collection}")
    print(f"match on   : {key_field}")
    print(f"in backup  : {len(backup)}")
    print(f"live now   : {len(live)}")
    print(f"missing    : {len(missing)}")
    print()

    if not missing:
        print("Nothing to restore.")
        return

    for record in missing:
        print(f"  - {record.get(key_field)}")
    print()

    if not args.apply:
        print("Dry run. Re-run with --apply to restore these records.")
        return

    created, failed, repaired = 0, 0, 0
    for index, record in enumerate(missing, start=1):
        label = record.get(key_field)
        payload = {k: v for k, v in record.items() if k not in SYSTEM_FIELDS}

        try:
            new_record = request_json(
                f"{base_url}/api/collections/{args.collection}/records",
                method="POST",
                body=payload,
                headers={"Authorization": token},
            )
        except SystemExit as err:
            print(f"[{index}/{len(missing)}] FAILED {label}: {err}", file=sys.stderr)
            failed += 1
            continue

        created += 1

        # The preparers and enrichers ran during the create and may have
        # overwritten fields (favorite reset to false, year/cover replaced from
        # the external API). Put the backup's values back.
        drift = drifted_fields(record, new_record, relation_fields)
        if drift:
            try:
                request_json(
                    f"{base_url}/api/collections/{args.collection}/records/{new_record['id']}",
                    method="PATCH",
                    body=drift,
                    headers={"Authorization": token},
                )
                repaired += 1
                print(
                    f"[{index}/{len(missing)}] restored {label} "
                    f"(reverted {', '.join(sorted(drift))})"
                )
            except SystemExit as err:
                print(
                    f"[{index}/{len(missing)}] restored {label} but could not revert "
                    f"{', '.join(sorted(drift))}: {err}",
                    file=sys.stderr,
                )
        else:
            print(f"[{index}/{len(missing)}] restored {label}")

        if args.delay and index < len(missing):
            time.sleep(args.delay)

    print()
    print(f"created {created}, field-reverted {repaired}, failed {failed}")


if __name__ == "__main__":
    main()
