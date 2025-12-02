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
	app.OnRecordCreateRequest("bookmarks", "github", "mtg", "stack_exchange").BindFunc(func(e *core.RecordRequestEvent) error {
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
				return fmt.Errorf("[OnRecordCreateRequest][archive]: %w", archiveErr)
			}

			e.Record.Set("archive", archiveURL)
			needsSave = true
		case "github":
			repoURL := e.Record.GetString("url")

			repo, repoErr := helpers.GetRepoInfo(repoURL)
			if repoErr != nil {
				return fmt.Errorf("[OnRecordCreateRequest][GetRepoInfo]: %w", repoErr)
			}

			e.Record.Set("name", repo.Name)
			e.Record.Set("owner", repo.Owner)
			e.Record.Set("description", repo.Description)
			e.Record.Set("language", repo.Language)
			needsSave = true
		case "mtg":
			name := e.Record.GetString("name")
			set := e.Record.GetString("set")
			number := e.Record.GetInt("collector_number")

			cardSelection, cardErr := helpers.SearchCard(name, set, number)
			if cardErr != nil {
				return fmt.Errorf("[OnRecordCreateRequest][SearchCard]: %w", cardErr)
			}

			// Get the first card from the selection map
			var card helpers.MTGItem
			for _, c := range cardSelection {
				card = c
				break
			}

			e.Record.Set("colors", card.Colors)
			e.Record.Set("type", card.Type)
			e.Record.Set("set_name", card.SetName)
			e.Record.Set("oracle_text", card.OracleText)
			e.Record.Set("flavor_text", card.FlavorText)
			e.Record.Set("rarity", card.Rarity)
			e.Record.Set("artist", card.Artist)
			e.Record.Set("released_at", card.ReleasedAt)
			e.Record.Set("image", card.Image)
			if card.Back != nil {
				e.Record.Set("back", card.Back)
			}
			needsSave = true
		case "stack_exchange":
			question := e.Record.GetString("question")

			questionInfo, questionInfoErr := helpers.GetQuestionInfo(question)
			if questionInfoErr != nil {
				return fmt.Errorf("[OnRecordCreateRequest][GetQuestionInfo]: %w", questionInfoErr)
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
