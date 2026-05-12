package schema

import (
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

var (
	metaIDOnce  sync.Once
	cachedMetaID string
)

func GetMetaID() string {
	metaIDOnce.Do(func() {
		path, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		if err := godotenv.Load(path + "/.env"); err != nil {
			log.Fatal(err)
		}

		id := os.Getenv("META_ID")
		if id == "" {
			log.Fatalln("META_ID env var not set")
		}

		cachedMetaID = id
	})

	return cachedMetaID
}

func BookmarksCollection() *core.Collection {
	collection := core.NewBaseCollection("bookmarks")
	// Access rules are filter expressions evaluated per request.
	// nil = deny all, types.Pointer("") = allow all, expression = conditional.
	collection.ViewRule = types.Pointer("@request.auth.id != ''")
	collection.CreateRule = types.Pointer("") // open: allows unauthenticated creates
	collection.UpdateRule = types.Pointer("@request.auth.id != ''")

	collection.Fields.Add(&core.TextField{Name: "title", Required: true})
	collection.Fields.Add(&core.TextField{Name: "creator", Required: true})
	collection.Fields.Add(&core.URLField{Name: "url", Required: true})
	collection.Fields.Add(&core.URLField{Name: "archive"})
	collection.Fields.Add(&core.RelationField{
		Name:         "tags",
		Required:     true,
		CollectionId: GetMetaID(), // RelationField requires the target collection's ID, not its name
		MaxSelect:    5,
	})
	collection.Fields.Add(&core.SelectField{
		Name:      "type",
		Required:  true,
		Values:    []string{"articles", "podcasts", "videos"},
		MaxSelect: 1,
	})
	collection.Fields.Add(&core.BoolField{Name: "dead"})
	collection.Fields.Add(&core.BoolField{Name: "shared"})
	collection.Fields.Add(&core.TextField{Name: "comments"})

	return collection
}

func FeedsCollection() *core.Collection {
	collection := core.NewBaseCollection("feeds")
	collection.ViewRule = types.Pointer("@request.auth.id != ''")
	collection.CreateRule = types.Pointer("")
	collection.UpdateRule = types.Pointer("@request.auth.id != ''")

	collection.Fields.Add(&core.TextField{Name: "title", Required: true})
	collection.Fields.Add(&core.URLField{Name: "url", Required: true})
	collection.Fields.Add(&core.URLField{Name: "rss"})
	collection.Fields.Add(&core.RelationField{
		Name:          "tags",
		Required:      true,
		CollectionId:  GetMetaID(),
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

func MediaCollection() *core.Collection {
	collection := core.NewBaseCollection("media")
	collection.ViewRule = types.Pointer("@request.auth.id != ''")
	collection.CreateRule = types.Pointer("")
	collection.UpdateRule = types.Pointer("@request.auth.id != ''")

	collection.Fields.Add(&core.TextField{Name: "title", Required: true})
	collection.Fields.Add(&core.TextField{Name: "creator", Required: true})
	collection.Fields.Add(&core.RelationField{
		Name:          "genre",
		Required:      true,
		CollectionId:  GetMetaID(),
		MaxSelect:     1,
		CascadeDelete: false,
	})
	collection.Fields.Add(&core.NumberField{Name: "year", Required: true})
	collection.Fields.Add(&core.SelectField{
		Name:      "type",
		Required:  true,
		Values:    []string{"books", "cds", "games", "movies", "shows", "vinyls"},
		MaxSelect: 1,
	})
	collection.Fields.Add(&core.TextField{Name: "barcode"})
	collection.Fields.Add(&core.URLField{Name: "cover"})
	collection.Fields.Add(&core.TextField{Name: "comments"})

	return collection
}

func MtgCollection() *core.Collection {
	collection := core.NewBaseCollection("mtg")
	collection.ViewRule = types.Pointer("@request.auth.id != ''")
	collection.CreateRule = types.Pointer("")
	collection.UpdateRule = types.Pointer("@request.auth.id != ''")

	collection.Fields.Add(&core.TextField{Name: "name", Required: true})
	collection.Fields.Add(&core.TextField{Name: "colors"})
	collection.Fields.Add(&core.TextField{Name: "type"})
	collection.Fields.Add(&core.TextField{Name: "set", Required: true})
	collection.Fields.Add(&core.TextField{Name: "set_name"})
	collection.Fields.Add(&core.TextField{Name: "oracle_text"})
	collection.Fields.Add(&core.TextField{Name: "flavor_text"})
	collection.Fields.Add(&core.TextField{Name: "rarity"})
	collection.Fields.Add(&core.NumberField{Name: "collector_number", Required: true})
	collection.Fields.Add(&core.TextField{Name: "artist"})
	collection.Fields.Add(&core.TextField{Name: "released_at"})
	collection.Fields.Add(&core.TextField{Name: "image"})
	collection.Fields.Add(&core.TextField{Name: "back"})

	return collection
}

func RecordsCollection() *core.Collection {
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

func GithubCollection() *core.Collection {
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


func MetaCollection() *core.Collection {
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

	// Pin the collection ID so RelationField.CollectionId in other collections can
	// reference it by a known value at schema definition time, before meta is created.
	collection.Id = GetMetaID()

	return collection
}
