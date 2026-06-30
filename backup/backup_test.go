package backup

import (
	"reflect"
	"testing"
)

func TestStripAndResolve(t *testing.T) {
	metaNames := map[string]string{
		"id_go":   "go",
		"id_rust": "rust",
		"id_rock": "rock",
	}

	t.Run("strips PB meta fields", func(t *testing.T) {
		in := map[string]any{
			"id":             "rec123",
			"created":        "2026-01-01 00:00:00Z",
			"updated":        "2026-01-02 00:00:00Z",
			"collectionId":   "col1",
			"collectionName": "articles",
			"expand":         map[string]any{},
			"title":          "Hello",
		}
		got := stripAndResolve(in, nil, metaNames)
		want := map[string]any{"title": "Hello"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("resolves single relation ID to name", func(t *testing.T) {
		in := map[string]any{"genre": "id_rock", "title": "X"}
		got := stripAndResolve(in, []string{"genre"}, metaNames)
		if got["genre"] != "rock" {
			t.Errorf("genre = %v, want rock", got["genre"])
		}
	})

	t.Run("resolves multi relation IDs to names", func(t *testing.T) {
		in := map[string]any{"tags": []string{"id_go", "id_rust"}}
		got := stripAndResolve(in, []string{"tags"}, metaNames)
		want := []string{"go", "rust"}
		if !reflect.DeepEqual(got["tags"], want) {
			t.Errorf("tags = %v, want %v", got["tags"], want)
		}
	})

	t.Run("keeps unresolved orphan ID as-is", func(t *testing.T) {
		in := map[string]any{"genre": "id_missing"}
		got := stripAndResolve(in, []string{"genre"}, metaNames)
		if got["genre"] != "id_missing" {
			t.Errorf("genre = %v, want id_missing (kept as-is)", got["genre"])
		}
	})

	t.Run("leaves non-relation fields untouched", func(t *testing.T) {
		in := map[string]any{"title": "Hello", "year": 2026, "dead": false}
		got := stripAndResolve(in, []string{"genre", "tags"}, metaNames)
		want := map[string]any{"title": "Hello", "year": 2026, "dead": false}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}
