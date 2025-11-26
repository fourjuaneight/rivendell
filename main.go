package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/fourjuaneight/rivendell/helpers"
	"github.com/fourjuaneight/rivendell/utils"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
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
	collections := []*core.Collection{
		bookmarksCollection(),
		feedsCollection(),
		mediaCollection(),
		mtgCollection(),
		recordsCollection(),
		githubCollection(),
		stackExchangeCollection(),
		metaCollection(),
	}

	// manually declare schemas
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		for _, collection := range collections {
			existing, err := e.App.FindCollectionByNameOrId(collection.Name)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("[OnServe][FindCollectionByNameOrId]: %w", err)
			}

			if existing == nil {
				if err := e.App.Save(collection); err != nil {
					return fmt.Errorf("[OnServe][SaveCollection]: %w", err)
				}
			}
		}

		return e.Next()
	})

	// setup migrations
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: true,
	})

	// set default values
	app.OnRecordCreateRequest("bookmarks", "feeds").BindFunc(func(e *core.RecordRequestEvent) error {
		e.Record.Set("dead", false)
		e.Record.Set("shared", false)

		return e.Next()
	})

	// archive bookmarks and enrich records after creation
	app.OnRecordCreateRequest("bookmarks", "github", "stack_exchange").BindFunc(func(e *core.RecordRequestEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		var needsSave bool

		switch e.Collection.Name {
		case "bookmarks":
			name := e.Record.GetString("title")
			url := e.Record.GetString("url")
			typeName := e.Record.GetString("type")

			archiveURL, archiveErr := archive(name, url, typeName)
			if archiveErr != nil {
				return fmt.Errorf("[OnRecordCreateRequest][archiveErr]: %w", archiveErr)
			}

			e.Record.Set("archive", archiveURL)
			needsSave = true
		case "github":
			repoURL := e.Record.GetString("url")

			repo, repoErr := helpers.GetRepoInfo(repoURL)
			if repoErr != nil {
				return fmt.Errorf("[OnRecordCreateRequest][repoErr]: %w", repoErr)
			}

			e.Record.Set("name", repo.Name)
			e.Record.Set("owner", repo.Owner)
			e.Record.Set("description", repo.Description)
			e.Record.Set("language", repo.Language)
			needsSave = true
		case "stack_exchange":
			question := e.Record.GetString("question")

			questionInfo, questionInfoErr := helpers.GetQuestionInfo(question)
			if questionInfoErr != nil {
				return fmt.Errorf("[OnRecordCreateRequest][questionInfoErr]: %w", questionInfoErr)
			}

			e.Record.Set("title", questionInfo.Title)
			e.Record.Set("question", questionInfo.Question)
			e.Record.Set("answers", questionInfo.Answer)
			e.Record.Set("tags", questionInfo.Tags)
			needsSave = true
		}

		if needsSave {
			if err := e.App.Save(e.Record); err != nil {
				return fmt.Errorf("[OnRecordCreateRequest][save]: %w", err)
			}
		}

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal("[Start]: %w", err)
	}
}
