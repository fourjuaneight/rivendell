package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

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

func downloadCover(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("[downloadCover][http.Get]: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[downloadCover][io.ReadAll]: %w", err)
	}

	return data, nil
}

func enrichMedia(r *core.Record) (bool, error) {
	title := r.GetString("title")
	mediaType := r.GetString("type")

	var coverURL string

	switch mediaType {
	case "movies", "shows":
		url, err := helpers.SearchMedia(title, r.GetInt("year"), mediaType)
		if err != nil {
			return false, fmt.Errorf("[enrichMedia]: %w", err)
		}
		coverURL = url
	case "books":
		if isbn := r.GetString("barcode"); isbn != "" {
			book, err := helpers.GetBookInfo(isbn)
			if err != nil {
				return false, fmt.Errorf("[enrichMedia]: %w", err)
			}
			coverURL = book.CoverURL
		}
	case "games":
		game, err := helpers.GetGameInfo(title, r.GetInt("year"))
		if err != nil {
			return false, fmt.Errorf("[enrichMedia]: %w", err)
		}
		coverURL = game.CoverURL
	case "cds", "vinyls":
		music, err := helpers.GetMusicInfo(title, r.GetString("creator"), r.GetInt("year"))
		if err != nil {
			return false, fmt.Errorf("[enrichMedia]: %w", err)
		}
		coverURL = music.CoverURL
	}

	if coverURL == "" {
		return false, nil
	}

	imageData, err := downloadCover(coverURL)
	if err != nil {
		return false, fmt.Errorf("[enrichMedia]: %w", err)
	}

	path := fmt.Sprintf("Media/%s/%s.jpeg", utils.ToCapitalized(mediaType), utils.FileNameFmt(title))
	b2URL, err := helpers.UploadToB2(imageData, path, "image/jpeg")
	if err != nil {
		return false, fmt.Errorf("[enrichMedia]: %w", err)
	}

	r.Set("cover", b2URL)
	return true, nil
}

func main() {
	app := pocketbase.New()

	// Automigrate: on startup, applies any pending migrations in the migrations/ package.
	// In dev mode (binary built from source), also auto-generates migration files when
	// collections are modified via the admin UI.
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: true,
	})

	// Pre-save hook: fields set before e.Next() are included in the initial DB insert.
	app.OnRecordCreateRequest("bookmarks", "feeds").BindFunc(func(e *core.RecordRequestEvent) error {
		e.Record.Set("dead", false)
		e.Record.Set("shared", false)

		return e.Next()
	})

	// Post-save hook: e.Next() commits the record first, then external API calls enrich it
	// and a second app.Save() writes the additional fields back. Separate from the pre-save
	// hook because the API calls are fallible and shouldn't block the initial insert.
	app.OnRecordCreateRequest("bookmarks", "github", "mtg", "media").BindFunc(func(e *core.RecordRequestEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		// Dispatch table maps collection name to its enrichment function.
		// Each enricher sets fields on the record and returns true if a save is needed.
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
