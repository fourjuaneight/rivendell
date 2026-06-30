// Package backup exports every non-system PocketBase collection to a
// pretty-printed JSON file on Backblaze B2, one file per collection under a
// dated path (Backups/<collection>/<YYYY-MM-DD>.json). Relation fields are
// rewritten from meta record IDs to names so a backup survives a meta rebuild
// and matches the create API's input format. See BackupAll for the cron entry
// point (registered in main.go).
package backup

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
