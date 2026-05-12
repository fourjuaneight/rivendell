package main

import (
	"fmt"
	"log"

	"github.com/fourjuaneight/rivendell/helpers"
	_ "github.com/fourjuaneight/rivendell/migrations"
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

	// Automigrate: on startup, applies any pending migrations in the migrations/ package.
	// In dev mode (binary built from source), also auto-generates migration files when
	// collections are modified via the admin UI.
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: true,
	})

	// OnRecordCreateRequest hooks are middleware for the create API endpoint.
	// Fields set on e.Record before calling e.Next() are included in the initial DB insert.
	app.OnRecordCreateRequest("bookmarks", "feeds").BindFunc(func(e *core.RecordRequestEvent) error {
		e.Record.Set("dead", false)
		e.Record.Set("shared", false)

		return e.Next()
	})

	// e.Next() is called first so the record is saved before the external API calls.
	// Enriched fields are written back via a second app.Save() after e.Next() returns.
	app.OnRecordCreateRequest("bookmarks", "github", "mtg", "media").BindFunc(func(e *core.RecordRequestEvent) error {
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
		case "media":
			mediaType := e.Record.GetString("type")
			if mediaType == "movies" || mediaType == "shows" {
				title := e.Record.GetString("title")
				year := e.Record.GetInt("year")

				tmdbID, searchErr := helpers.SearchMedia(title, year, mediaType)
				if searchErr != nil {
					return fmt.Errorf("[OnRecordCreateRequest][SearchMedia]: %w", searchErr)
				}

				e.Record.Set("tmdb_id", tmdbID)
				needsSave = true
			}
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
