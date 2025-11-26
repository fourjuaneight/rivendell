package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

func getMetaID() string {
	path, pathErr := os.Getwd()
	if pathErr != nil {
		log.Fatal(pathErr)
	}

	envPath := path + "/.env"
	envErr := godotenv.Load(envPath)
	if envErr != nil {
		log.Fatal(envErr)
	}

	META_ID := os.Getenv("META_ID")
	if META_ID == "" {
		log.Fatalln("Please provide a Meta collection ID.")
	}

	return META_ID
}

// - title (text, required)
// - creator (text, required)
// - url (url, required)
// - archive (url)
// - tags (relation to meta, required, max 5)
// - type (select: articles, podcasts, videos, required)
// - dead (bool)
// - shared (bool)
// - comments (text)
func bookmarksCollection() *core.Collection {
	collection := core.NewBaseCollection("bookmarks")
	collection.ViewRule = types.Pointer("@request.auth.id != ''")
	collection.CreateRule = types.Pointer("")
	collection.UpdateRule = types.Pointer("@request.auth.id != ''")

	collection.Fields.Add(&core.TextField{
		Name:     "title",
		Required: true,
	})
	collection.Fields.Add(&core.TextField{
		Name:     "creator",
		Required: true,
	})
	collection.Fields.Add(&core.URLField{
		Name:     "url",
		Required: true,
	})
	collection.Fields.Add(&core.URLField{
		Name: "archive",
	})
	collection.Fields.Add(&core.RelationField{
		Name:          "tags",
		Required:      true,
		CollectionId:  getMetaID(),
		MaxSelect:     5,
		CascadeDelete: false,
	})
	collection.Fields.Add(&core.SelectField{
		Name:      "type",
		Required:  true,
		Values:    []string{"articles", "podcasts", "videos"},
		MaxSelect: 1,
	})
	collection.Fields.Add(&core.BoolField{
		Name: "dead",
	})
	collection.Fields.Add(&core.BoolField{
		Name: "shared",
	})
	collection.Fields.Add(&core.TextField{
		Name: "comments",
	})

	return collection
}

// - title (text, required)
// - url (url, required)
// - rss (url)
// - tags (relation to meta, required, max 5)
// - type (select: podcasts, websites, youtube, required)
// - dead (bool)
// - shared (bool)
// - comments (text)
func feedsCollection() *core.Collection {
	collection := core.NewBaseCollection("feeds")
	collection.ViewRule = types.Pointer("@request.auth.id != ''")
	collection.CreateRule = types.Pointer("")
	collection.UpdateRule = types.Pointer("@request.auth.id != ''")

	collection.Fields.Add(&core.TextField{
		Name:     "title",
		Required: true,
	})
	collection.Fields.Add(&core.URLField{
		Name:     "url",
		Required: true,
	})
	collection.Fields.Add(&core.URLField{
		Name: "rss",
	})
	collection.Fields.Add(&core.RelationField{
		Name:          "tags",
		Required:      true,
		CollectionId:  getMetaID(),
		MaxSelect:     5,
		CascadeDelete: false,
	})
	collection.Fields.Add(&core.SelectField{
		Name:      "type",
		Required:  true,
		Values:    []string{"podcasts", "websites", "youtube"},
		MaxSelect: 1,
	})
	collection.Fields.Add(&core.BoolField{Name: "dead"})
	collection.Fields.Add(&core.BoolField{Name: "shared"})
	collection.Fields.Add(&core.TextField{Name: "comments"})

	return collection
}

// - title (text, required)
// - creator (text, required)
// - genre (relation to meta, required, max 1)
// - year (number, required)
// - rating (number, required)
// - physical (bool)
// - shelf (bool)
// - type (select: books, cds, games, movies, shows, vinyls required)
// - shared (bool)
// - comments (text)
func mediaCollection() *core.Collection {
	collection := core.NewBaseCollection("media")
	collection.ViewRule = types.Pointer("@request.auth.id != ''")
	collection.CreateRule = types.Pointer("")
	collection.UpdateRule = types.Pointer("@request.auth.id != ''")

	collection.Fields.Add(&core.TextField{Name: "title", Required: true})
	collection.Fields.Add(&core.TextField{Name: "creator", Required: true})
	collection.Fields.Add(&core.RelationField{
		Name:          "genre",
		Required:      true,
		CollectionId:  getMetaID(),
		MaxSelect:     1,
		CascadeDelete: false,
	})
	collection.Fields.Add(&core.NumberField{Name: "year", Required: true})
	collection.Fields.Add(&core.NumberField{Name: "rating", Required: true})
	collection.Fields.Add(&core.BoolField{Name: "physical"})
	collection.Fields.Add(&core.BoolField{Name: "shelf"})
	collection.Fields.Add(&core.SelectField{
		Name:      "type",
		Required:  true,
		Values:    []string{"books", "cds", "games", "movies", "shows", "vinyls"},
		MaxSelect: 1,
	})
	collection.Fields.Add(&core.BoolField{Name: "shared"})
	collection.Fields.Add(&core.TextField{Name: "comments"})

	return collection
}

// - title (text, required)
// - artist (text, required)
// - genre (relation to meta, required, max 1)
// - year (number, required)
// - rating (number, required)
// - playlist (text, required)
func musicCollection() *core.Collection {
	collection := core.NewBaseCollection("music")
	collection.ViewRule = types.Pointer("@request.auth.id != ''")
	collection.CreateRule = types.Pointer("")
	collection.UpdateRule = types.Pointer("@request.auth.id != ''")

	collection.Fields.Add(&core.TextField{Name: "title", Required: true})
	collection.Fields.Add(&core.TextField{Name: "artist", Required: true})
	collection.Fields.Add(&core.RelationField{
		Name:          "genre",
		Required:      true,
		CollectionId:  getMetaID(),
		MaxSelect:     1,
		CascadeDelete: false,
	})
	collection.Fields.Add(&core.NumberField{Name: "year", Required: true})
	collection.Fields.Add(&core.NumberField{Name: "rating", Required: true})
	collection.Fields.Add(&core.TextField{Name: "playlist", Required: true})

	return collection
}

// - name (text, required)
// - colors (text)
// - type (text)
// - set (text, required)
// - set_name (text, required)
// - oracle_text (text)
// - flavor_text (text)
// - rarity (text, required)
// - collector_number (text, required)
// - artist (text, required)
// - released_at (text, required)
// - image (text, required)
// - back (text)
func mtgCollection() *core.Collection {
	collection := core.NewBaseCollection("mtg")
	collection.ViewRule = types.Pointer("@request.auth.id != ''")
	collection.CreateRule = types.Pointer("")
	collection.UpdateRule = types.Pointer("@request.auth.id != ''")

	collection.Fields.Add(&core.TextField{Name: "name", Required: true})
	collection.Fields.Add(&core.TextField{Name: "colors"})
	collection.Fields.Add(&core.TextField{Name: "type"})
	collection.Fields.Add(&core.TextField{Name: "set", Required: true})
	collection.Fields.Add(&core.TextField{Name: "set_name", Required: true})
	collection.Fields.Add(&core.TextField{Name: "oracle_text"})
	collection.Fields.Add(&core.TextField{Name: "flavor_text"})
	collection.Fields.Add(&core.TextField{Name: "rarity", Required: true})
	collection.Fields.Add(&core.TextField{Name: "collector_number", Required: true})
	collection.Fields.Add(&core.TextField{Name: "artist", Required: true})
	collection.Fields.Add(&core.TextField{Name: "released_at", Required: true})
	collection.Fields.Add(&core.TextField{Name: "image", Required: true})
	collection.Fields.Add(&core.TextField{Name: "back"})

	return collection
}

// - company (text, required)
// - position (text)
// - stack (json)
// - start (date, required)
// - end (date)
func recordsCollection() *core.Collection {
	collection := core.NewBaseCollection("records")
	collection.ViewRule = types.Pointer("@request.auth.id != ''")
	collection.CreateRule = types.Pointer("")
	collection.UpdateRule = types.Pointer("@request.auth.id != ''")

	collection.Fields.Add(&core.TextField{Name: "company", Required: true})
	collection.Fields.Add(&core.TextField{Name: "position"})
	collection.Fields.Add(&core.JSONField{Name: "stack"})
	collection.Fields.Add(&core.DateField{Name: "start", Required: true})
	collection.Fields.Add(&core.DateField{Name: "end"})

	return collection
}

// - name (text)
// - owner (text)
// - description (text)
// - language (text)
// - url (text, required, unique index)
func githubCollection() *core.Collection {
	collection := core.NewBaseCollection("github")
	collection.ViewRule = types.Pointer("@request.auth.id != ''")
	collection.CreateRule = types.Pointer("")
	collection.UpdateRule = types.Pointer("@request.auth.id != ''")

	collection.Fields.Add(&core.TextField{Name: "name"})
	collection.Fields.Add(&core.TextField{Name: "owner"})
	collection.Fields.Add(&core.TextField{Name: "description"})
	collection.Fields.Add(&core.TextField{Name: "language"})
	collection.Fields.Add(&core.TextField{Name: "url", Required: true})
	collection.AddIndex("idx_github_url_unique", true, "url", "")

	return collection
}

// - title (text)
// - question (text, required)
// - answer (text)
// - tags (json)
func stackExchangeCollection() *core.Collection {
	collection := core.NewBaseCollection("stack_exchange")
	collection.ViewRule = types.Pointer("@request.auth.id != ''")
	collection.CreateRule = types.Pointer("")
	collection.UpdateRule = types.Pointer("@request.auth.id != ''")

	collection.Fields.Add(&core.TextField{Name: "title"})
	collection.Fields.Add(&core.TextField{Name: "question", Required: true})
	collection.Fields.Add(&core.TextField{Name: "answer"})
	collection.Fields.Add(&core.JSONField{Name: "tags"})

	return collection
}

// - name (text, required)
// - type (select: tags, genre, max 1)
func metaCollection() *core.Collection {
	collection := core.NewBaseCollection("meta")
	collection.ViewRule = types.Pointer("@request.auth.id != ''")
	collection.CreateRule = types.Pointer("")
	collection.UpdateRule = types.Pointer("@request.auth.id != ''")

	collection.Fields.Add(&core.TextField{Name: "name", Required: true})
	collection.Fields.Add(&core.SelectField{
		Name:      "type",
		Values:    []string{"tags", "genre"},
		MaxSelect: 1,
	})

	collection.Id = getMetaID()

	return collection
}
