// Package backup exports every non-system PocketBase collection to a
// pretty-printed JSON file on Backblaze B2, one file per collection under a
// dated path (Backups/<collection>/<YYYY-MM-DD>.json). Relation fields are
// rewritten from meta record IDs to names so a backup survives a meta rebuild
// and matches the create API's input format. See BackupAll for the cron entry
// point (registered in main.go).
package backup

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fourjuaneight/rivendell/helpers"

	"github.com/pocketbase/pocketbase/core"
)

// metaKeys are PocketBase-managed fields stripped from every backed-up record.
// Only "id" is actually present in this app's schema today; the rest are a
// defensive superset so the backup stays correct if created/updated or the
// API-response meta keys ever appear.
var metaKeys = []string{"id", "created", "updated", "collectionId", "collectionName", "expand"}

// resolveID returns the meta name for id, or id itself if there's no match.
func resolveID(id string, metaNames map[string]string) string {
	if name, ok := metaNames[id]; ok {
		return name
	}
	return id
}

// stripAndResolve removes PocketBase meta fields from a single record's
// exported field map and rewrites relation fields from meta IDs to names. It is
// pure: no DB, B2, or network access. relFields is the set of relation field
// names for the record's collection; metaNames maps meta record ID to name. An
// ID with no entry in metaNames is left unchanged so data is never silently
// dropped. The input map is mutated in place and returned.
func stripAndResolve(fields map[string]any, relFields []string, metaNames map[string]string) map[string]any {
	for _, k := range metaKeys {
		delete(fields, k)
	}

	for _, name := range relFields {
		switch v := fields[name].(type) {
		case string:
			fields[name] = resolveID(v, metaNames)
		case []string:
			resolved := make([]string, len(v))
			for i, id := range v {
				resolved[i] = resolveID(id, metaNames)
			}
			fields[name] = resolved
		}
	}

	return fields
}

// buildMetaNames queries every meta record once and returns a map of record ID
// to name, used to resolve relation fields during backup.
func buildMetaNames(app core.App) (map[string]string, error) {
	records, err := app.FindAllRecords("meta")
	if err != nil {
		return nil, fmt.Errorf("[buildMetaNames]: %w", err)
	}

	names := make(map[string]string, len(records))
	for _, r := range records {
		names[r.Id] = r.GetString("name")
	}
	return names, nil
}

// relationFieldNames returns the names of every relation field in a collection.
func relationFieldNames(collection *core.Collection) []string {
	var names []string
	for _, f := range collection.Fields {
		if _, ok := f.(*core.RelationField); ok {
			names = append(names, f.GetName())
		}
	}
	return names
}

// backupCollection fetches all records in a collection, strips meta fields,
// resolves relations to names, and uploads a pretty-printed JSON array to B2 at
// Backups/<collection>/<date>.json. An empty collection still uploads "[]".
func backupCollection(app core.App, collection *core.Collection, metaNames map[string]string, date string) error {
	records, err := app.FindAllRecords(collection.Name)
	if err != nil {
		return fmt.Errorf("[backupCollection][%s]: %w", collection.Name, err)
	}

	relFields := relationFieldNames(collection)

	rows := make([]map[string]any, 0, len(records))
	for _, r := range records {
		rows = append(rows, stripAndResolve(r.PublicExport(), relFields, metaNames))
	}

	data, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return fmt.Errorf("[backupCollection][%s]: %w", collection.Name, err)
	}

	filename := fmt.Sprintf("%s/%s.json", collection.Name, date)
	if _, err := helpers.UploadToB2(data, "backups", filename, "application/json"); err != nil {
		return fmt.Errorf("[backupCollection][%s]: %w", collection.Name, err)
	}

	app.Logger().Info("collection backed up",
		"collection", collection.Name,
		"record_count", len(records),
	)
	return nil
}

// BackupAll is the cron entry point. It exports every non-system collection to
// B2 as a dated JSON file, sequentially. One collection's failure is logged and
// skipped — it never aborts the rest of the run. A failure to build the meta
// map or list collections aborts the whole run, since every collection needs
// the meta map to resolve relations.
func BackupAll(app core.App) error {
	date := time.Now().Format("2006-01-02")

	metaNames, err := buildMetaNames(app)
	if err != nil {
		app.Logger().Error("backup aborted, meta lookup failed", "error", err.Error())
		return err
	}

	collections, err := app.FindAllCollections()
	if err != nil {
		app.Logger().Error("backup aborted, collection list failed", "error", err.Error())
		return err
	}

	app.Logger().Info("backup started", "date", date, "collection_count", len(collections))

	var backedUp, failed int
	for _, collection := range collections {
		// Skip system collections (e.g. _superusers, _logs) and auth
		// collections (e.g. users): auth records can carry credentials/email,
		// and backups land off-site on B2, so they are deliberately excluded.
		if collection.System || collection.IsAuth() {
			continue
		}
		if err := backupCollection(app, collection, metaNames, date); err != nil {
			app.Logger().Error("collection backup failed",
				"collection", collection.Name,
				"error", err.Error(),
			)
			failed++
			continue
		}
		backedUp++
	}

	app.Logger().Info("backup done", "backed_up", backedUp, "failed", failed)
	return nil
}
