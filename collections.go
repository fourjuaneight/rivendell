package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
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

func bookmarksCollection() *models.Collection {
	collection := &models.Collection{
		Name:       "bookmarks",
		Type:       models.CollectionTypeBase,
		ListRule:   nil,
		ViewRule:   types.Pointer("@request.auth.id != ''"),
		CreateRule: types.Pointer(""),
		UpdateRule: types.Pointer("@request.auth.id != ''"),
		DeleteRule: nil,
		Schema: schema.NewSchema(
			&schema.SchemaField{
				Name:     "title",
				Type:     schema.FieldTypeText,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "creator",
				Type:     schema.FieldTypeText,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "url",
				Type:     schema.FieldTypeUrl,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "archive",
				Type:     schema.FieldTypeUrl,
				Required: false,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "tags",
				Type:     schema.FieldTypeRelation,
				Required: true,
				Options: &schema.RelationOptions{
					MaxSelect:     types.Pointer(5),
					CollectionId:  getMetaID(),
					CascadeDelete: false,
				},
			},
			&schema.SchemaField{
				Name:     "type",
				Type:     schema.FieldTypeText,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "dead",
				Type:     schema.FieldTypeBool,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "shared",
				Type:     schema.FieldTypeBool,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "comments",
				Type:     schema.FieldTypeText,
				Required: false,
				Unique:   false,
			},
		),
	}

	return collection
}

func feedsCollection() *models.Collection {
	collection := &models.Collection{
		Name:       "feeds",
		Type:       models.CollectionTypeBase,
		ListRule:   nil,
		ViewRule:   types.Pointer("@request.auth.id != ''"),
		CreateRule: types.Pointer(""),
		UpdateRule: types.Pointer("@request.auth.id != ''"),
		DeleteRule: nil,
		Schema: schema.NewSchema(
			&schema.SchemaField{
				Name:     "title",
				Type:     schema.FieldTypeText,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "url",
				Type:     schema.FieldTypeUrl,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "rss",
				Type:     schema.FieldTypeUrl,
				Required: false,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "tags",
				Type:     schema.FieldTypeRelation,
				Required: true,
				Options: &schema.RelationOptions{
					MaxSelect:     types.Pointer(5),
					CollectionId:  getMetaID(),
					CascadeDelete: false,
				},
			},
			&schema.SchemaField{
				Name:     "type",
				Type:     schema.FieldTypeText,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "dead",
				Type:     schema.FieldTypeBool,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "shared",
				Type:     schema.FieldTypeBool,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "comments",
				Type:     schema.FieldTypeText,
				Required: false,
				Unique:   false,
			},
		),
	}

	return collection
}

func mediaCollection() *models.Collection {
	collection := &models.Collection{
		Name:       "media",
		Type:       models.CollectionTypeBase,
		ListRule:   nil,
		ViewRule:   types.Pointer("@request.auth.id != ''"),
		CreateRule: types.Pointer(""),
		UpdateRule: types.Pointer("@request.auth.id != ''"),
		DeleteRule: nil,
		Schema: schema.NewSchema(
			&schema.SchemaField{
				Name:     "title",
				Type:     schema.FieldTypeText,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "creator",
				Type:     schema.FieldTypeText,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "genre",
				Type:     schema.FieldTypeRelation,
				Required: true,
				Options: &schema.RelationOptions{
					MaxSelect:     types.Pointer(1),
					CollectionId:  getMetaID(),
					CascadeDelete: false,
				},
			},
			&schema.SchemaField{
				Name:     "year",
				Type:     schema.FieldTypeNumber,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "rating",
				Type:     schema.FieldTypeNumber,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "physical",
				Type:     schema.FieldTypeBool,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "shelf",
				Type:     schema.FieldTypeBool,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "type",
				Type:     schema.FieldTypeText,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "shared",
				Type:     schema.FieldTypeBool,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "comments",
				Type:     schema.FieldTypeText,
				Required: false,
				Unique:   false,
			},
		),
	}

	return collection
}

func musicCollection() *models.Collection {
	collection := &models.Collection{
		Name:       "music",
		Type:       models.CollectionTypeBase,
		ListRule:   nil,
		ViewRule:   types.Pointer("@request.auth.id != ''"),
		CreateRule: types.Pointer(""),
		UpdateRule: types.Pointer("@request.auth.id != ''"),
		DeleteRule: nil,
		Schema: schema.NewSchema(
			&schema.SchemaField{
				Name:     "title",
				Type:     schema.FieldTypeText,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "artist",
				Type:     schema.FieldTypeText,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "genre",
				Type:     schema.FieldTypeRelation,
				Required: true,
				Options: &schema.RelationOptions{
					MaxSelect:     types.Pointer(1),
					CollectionId:  getMetaID(),
					CascadeDelete: false,
				},
			},
			&schema.SchemaField{
				Name:     "year",
				Type:     schema.FieldTypeNumber,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "rating",
				Type:     schema.FieldTypeNumber,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "playlist",
				Type:     schema.FieldTypeText,
				Required: true,
				Unique:   false,
			},
		),
	}

	return collection
}

func mtgCollection() *models.Collection {
	collection := &models.Collection{
		Name:       "mtg",
		Type:       models.CollectionTypeBase,
		ListRule:   nil,
		ViewRule:   types.Pointer("@request.auth.id != ''"),
		CreateRule: types.Pointer(""),
		UpdateRule: types.Pointer("@request.auth.id != ''"),
		DeleteRule: nil,
		Schema: schema.NewSchema(
			&schema.SchemaField{
				Name:     "name",
				Type:     schema.FieldTypeText,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "colors",
				Type:     schema.FieldTypeText,
				Required: false,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "type",
				Type:     schema.FieldTypeText,
				Required: false,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "set",
				Type:     schema.FieldTypeText,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "set_name",
				Type:     schema.FieldTypeText,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "oracle_text",
				Type:     schema.FieldTypeText,
				Required: false,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "flavor_text",
				Type:     schema.FieldTypeText,
				Required: false,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "rarity",
				Type:     schema.FieldTypeText,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "collector_number",
				Type:     schema.FieldTypeText,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "artist",
				Type:     schema.FieldTypeText,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "released_at",
				Type:     schema.FieldTypeText,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "image",
				Type:     schema.FieldTypeText,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "back",
				Type:     schema.FieldTypeText,
				Required: false,
				Unique:   false,
			},
		),
	}

	return collection
}

func recordsCollection() *models.Collection {
	collection := &models.Collection{
		Name:       "records",
		Type:       models.CollectionTypeBase,
		ListRule:   nil,
		ViewRule:   types.Pointer("@request.auth.id != ''"),
		CreateRule: types.Pointer(""),
		UpdateRule: types.Pointer("@request.auth.id != ''"),
		DeleteRule: nil,
		Schema: schema.NewSchema(
			&schema.SchemaField{
				Name:     "company",
				Type:     schema.FieldTypeText,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "position",
				Type:     schema.FieldTypeText,
				Required: false,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "stack",
				Type:     schema.FieldTypeJson,
				Required: false,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "start",
				Type:     schema.FieldTypeDate,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "end",
				Type:     schema.FieldTypeDate,
				Required: false,
				Unique:   false,
			},
		),
	}

	return collection
}

func metaCollection() *models.Collection {
	collection := &models.Collection{
		Name:       "meta",
		Type:       models.CollectionTypeBase,
		ListRule:   nil,
		ViewRule:   types.Pointer("@request.auth.id != ''"),
		CreateRule: types.Pointer(""),
		UpdateRule: types.Pointer("@request.auth.id != ''"),
		DeleteRule: nil,
		Schema: schema.NewSchema(
			&schema.SchemaField{
				Name:     "name",
				Type:     schema.FieldTypeText,
				Required: true,
				Unique:   false,
			},
			&schema.SchemaField{
				Name:     "type",
				Type:     schema.FieldTypeSelect,
				Required: false,
				Unique:   false,
				Options: &schema.SelectOptions{
					MaxSelect: 1,
					Values: []string{
						"tags",
						"genre",
					},
				},
			},
		),
	}

	collection.SetId(getMetaID())

	return collection
}
