package main

import (
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

func main() {
	app := pocketbase.New()
	collections := []*models.Collection{
		bookmarksCollection(),
		feedsCollection(),
		mediaCollection(),
		musicCollection(),
		mtgCollection(),
		recordsCollection(),
		metaCollection(),
	}

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		for _, collection := range collections {
			existing, _ := app.Dao().FindCollectionByNameOrId(collection.Name)

			if existing == nil {
				if err := app.Dao().SaveCollection(collection); err != nil {
					log.Fatal("[OnBeforeServe]: %w", err)
				}
			}
		}

		return nil
	})

	migratecmd.MustRegister(app, app.RootCmd, &migratecmd.Options{
		Automigrate: true,
	})

	app.OnRecordBeforeCreateRequest().Add(func(e *core.RecordCreateEvent) error {
		record := e.Record
		collection := record.Collection().Name

		if collection == "bookmarks" || collection == "feeds" {
			record.Set("dead", false)
			record.Set("shared", false)
		}

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal("[Start]: %w", err)
	}
}
