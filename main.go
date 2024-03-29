package main

import (
	"fmt"
	"log"

	"github.com/fourjuaneight/rivendell/helpers"
	"github.com/fourjuaneight/rivendell/utils"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

func archive(name string, url string, typeName string) (string, error) {
	media, getcontentErr := helpers.GetContent(name, url, typeName)
	if getcontentErr != nil {
		return "", fmt.Errorf("[archive][GetContent]: %w", getcontentErr)
	}

	typeOps := utils.GetFileType(typeName, url)
	list := utils.ToCapitalized(typeName)
	path := fmt.Sprintf("Bookmarks/%s/%s.%s", list, utils.FileNameFmt(name), typeOps.File)
	archiveUrl, uploadtob2Err := helpers.UploadToB2(media, path, typeOps.MIME)
	if uploadtob2Err != nil {
		return "", fmt.Errorf("[archive][UploadToB2]: %w", uploadtob2Err)
	}

	return archiveUrl, nil
}

func main() {
	app := pocketbase.New()
	collections := []*models.Collection{
		bookmarksCollection(),
		feedsCollection(),
		mediaCollection(),
		musicCollection(),
		mtgCollection(),
		recordsCollection(),
		githubCollection(),
		stackExchangeCollection(),
		metaCollection(),
	}

	// manually declare schemas
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

	// setup migrations
	migratecmd.MustRegister(app, app.RootCmd, &migratecmd.Options{
		Automigrate: true,
	})

	// set default values
	app.OnRecordBeforeCreateRequest().Add(func(e *core.RecordCreateEvent) error {
		record := e.Record
		collection := record.Collection().Name

		if collection == "bookmarks" || collection == "feeds" {
			record.Set("dead", false)
			record.Set("shared", false)
		}

		return nil
	})

	// archive bookmarks
	app.OnRecordAfterCreateRequest().Add(func(e *core.RecordCreateEvent) error {
		record := e.Record
		collection := record.Collection().Name

		switch collection {
		case "bookmarks":
			// archive bookmarks
			name := record.SchemaData()["title"].(string)
			url := record.SchemaData()["url"].(string)
			typeName := record.SchemaData()["type"].(string)

			archive, archiveErr := archive(name, url, typeName)
			if archiveErr != nil {
				return fmt.Errorf("[OnRecordAfterCreateRequest][archiveErr]: %w", archiveErr)
			}

			record.Set("archive", archive)
		case "github":
			// query repository info
			url := record.SchemaData()["url"].(string)

			repo, repoErr := helpers.GetRepoInfo(url)
			if repoErr != nil {
				return fmt.Errorf("[OnRecordAfterCreateRequest][repoErr]: %w", repoErr)
			}

			record.Set("name", repo.Name)
			record.Set("owner", repo.Owner)
			record.Set("description", repo.Description)
			record.Set("language", repo.Language)
		case "stack_exchange":
			// query repository info
			question := record.SchemaData()["question"].(string)

			questionInfo, questionInfoErr := helpers.GetQuestionInfo(question)
			if questionInfoErr != nil {
				return fmt.Errorf("[OnRecordAfterCreateRequest][questionInfoErr]: %w", questionInfoErr)
			}

			record.Set("title", questionInfo.Title)
			record.Set("question", questionInfo.Question)
			record.Set("answers", questionInfo.Answer)
			record.Set("tags", questionInfo.Tags)
		default:
		}

		app.Dao().SaveRecord(record)

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal("[Start]: %w", err)
	}
}
