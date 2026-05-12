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

func enrichBookmarks(r *core.Record) (bool, error) {
	archiveURL, err := archive(r.GetString("title"), r.GetString("url"), r.GetString("type"))
	if err != nil {
		return false, fmt.Errorf("[enrichBookmarks]: %w", err)
	}

	r.Set("archive", archiveURL)
	return true, nil
}

func enrichGithub(r *core.Record) (bool, error) {
	repo, err := helpers.GetRepoInfo(r.GetString("url"))
	if err != nil {
		return false, fmt.Errorf("[enrichGithub]: %w", err)
	}

	r.Set("name", repo.Name)
	r.Set("owner", repo.Owner)
	r.Set("description", repo.Description)
	r.Set("language", repo.Language)
	return true, nil
}

func enrichMtg(r *core.Record) (bool, error) {
	cardSelection, err := helpers.SearchCard(r.GetString("name"), r.GetString("set"), r.GetInt("collector_number"))
	if err != nil {
		return false, fmt.Errorf("[enrichMtg]: %w", err)
	}

	var card helpers.MTGItem
	for _, c := range cardSelection {
		card = c
		break
	}

	r.Set("colors", card.Colors)
	r.Set("type", card.Type)
	r.Set("set_name", card.SetName)
	r.Set("oracle_text", card.OracleText)
	r.Set("flavor_text", card.FlavorText)
	r.Set("rarity", card.Rarity)
	r.Set("artist", card.Artist)
	r.Set("released_at", card.ReleasedAt)
	r.Set("image", card.Image)
	if card.Back != nil {
		r.Set("back", card.Back)
	}
	return true, nil
}

func enrichMedia(r *core.Record) (bool, error) {
	switch r.GetString("type") {
	case "movies", "shows":
		coverURL, err := helpers.SearchMedia(r.GetString("title"), r.GetInt("year"), r.GetString("type"))
		if err != nil {
			return false, fmt.Errorf("[enrichMedia]: %w", err)
		}
		if coverURL != "" {
			r.Set("cover", coverURL)
			return true, nil
		}
	case "books":
		if isbn := r.GetString("barcode"); isbn != "" {
			book, err := helpers.GetBookInfo(isbn)
			if err != nil {
				return false, fmt.Errorf("[enrichMedia]: %w", err)
			}
			if book.CoverURL != "" {
				r.Set("cover", book.CoverURL)
				return true, nil
			}
		}
	}
	return false, nil
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

		enrichers := map[string]func(*core.Record) (bool, error){
			"bookmarks": enrichBookmarks,
			"github":    enrichGithub,
			"mtg":       enrichMtg,
			"media":     enrichMedia,
		}

		enrich, ok := enrichers[e.Collection.Name]
		if !ok {
			return nil
		}

		needsSave, err := enrich(e.Record)
		if err != nil {
			return fmt.Errorf("[OnRecordCreateRequest]: %w", err)
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
