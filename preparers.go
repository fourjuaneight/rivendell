package main

import (
	"fmt"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// resolveTagNames looks up meta records by name and returns their IDs.
func resolveTagNames(app core.App, names []string) ([]string, error) {
	if len(names) == 0 {
		return nil, nil
	}

	parts := make([]string, len(names))
	for i, name := range names {
		parts[i] = fmt.Sprintf(`name = "%s"`, name)
	}
	filter := fmt.Sprintf(`(%s) && type = "tags"`, strings.Join(parts, " || "))

	records, err := app.FindRecordsByFilter("meta", filter, "", 0, 0)
	if err != nil {
		return nil, fmt.Errorf("[resolveTagNames]: %w", err)
	}

	ids := make([]string, len(records))
	for i, r := range records {
		ids[i] = r.Id
	}
	return ids, nil
}

// resolveMetaName looks up a single meta record by name and type, returning its ID.
func resolveMetaName(app core.App, name, metaType string) (string, error) {
	filter := fmt.Sprintf(`name = "%s" && type = "%s"`, name, metaType)
	record, err := app.FindFirstRecordByFilter("meta", filter)
	if err != nil {
		return "", fmt.Errorf("[resolveMetaName] %q (%s): %w", name, metaType, err)
	}
	return record.Id, nil
}

func prepareTags(app core.App, r *core.Record) error {
	if tagNames := r.GetStringSlice("tags"); len(tagNames) > 0 {
		tagIDs, err := resolveTagNames(app, tagNames)
		if err != nil {
			return fmt.Errorf("[prepareTags]: %w", err)
		}
		r.Set("tags", tagIDs)
	}
	return nil
}

func prepareArticle(app core.App, r *core.Record) error {
	r.Set("dead", false)
	r.Set("shared", false)
	r.Set("favorite", false)
	return prepareTags(app, r)
}

func prepareMedia(app core.App, r *core.Record) error {
	r.Set("dead", false)
	r.Set("shared", false)
	r.Set("favorite", false)
	return prepareTags(app, r)
}

func prepareFeed(app core.App, r *core.Record) error {
	r.Set("dead", false)
	r.Set("shared", false)
	return prepareTags(app, r)
}

func prepareWithGenre(app core.App, r *core.Record) error {
	if name := r.GetString("genre"); name != "" {
		id, err := resolveMetaName(app, name, "genre")
		if err != nil {
			return fmt.Errorf("[prepareWithGenre]: %w", err)
		}
		r.Set("genre", id)
	}
	return nil
}

func prepareWithStatus(app core.App, r *core.Record) error {
	name := r.GetString("status")
	if name == "" {
		name = "not_started"
	}
	id, err := resolveMetaName(app, name, "status")
	if err != nil {
		return fmt.Errorf("[prepareWithStatus]: %w", err)
	}
	r.Set("status", id)
	return nil
}

func prepareBook(app core.App, r *core.Record) error {
	if err := prepareWithGenre(app, r); err != nil {
		return err
	}
	r.Set("favorite", false)
	return prepareWithStatus(app, r)
}

func prepareMovieOrShow(app core.App, r *core.Record) error {
	if err := prepareWithGenre(app, r); err != nil {
		return err
	}
	if name := r.GetString("definition"); name != "" {
		id, err := resolveMetaName(app, name, "definition")
		if err != nil {
			return fmt.Errorf("[prepareMovieOrShow]: %w", err)
		}
		r.Set("definition", id)
	}
	r.Set("favorite", false)
	return prepareWithStatus(app, r)
}

func prepareGame(app core.App, r *core.Record) error {
	if err := prepareWithGenre(app, r); err != nil {
		return err
	}
	if name := r.GetString("platform"); name != "" {
		id, err := resolveMetaName(app, name, "platform")
		if err != nil {
			return fmt.Errorf("[prepareGame]: %w", err)
		}
		r.Set("platform", id)
	}
	r.Set("favorite", false)
	return prepareWithStatus(app, r)
}
