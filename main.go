package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/fourjuaneight/rivendell/helpers"
	_ "github.com/fourjuaneight/rivendell/migrations"
	"github.com/fourjuaneight/rivendell/utils"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

func archive(name string, url string, typeName string) (string, error) {
	media, err := helpers.GetContent(name, url, typeName)
	if err != nil {
		return "", fmt.Errorf("[archive][GetContent]: %w", err)
	}

	typeOps := utils.GetFileType(typeName, url)
	list := utils.ToCapitalized(typeName)
	path := fmt.Sprintf("Bookmarks/%s/%s.%s", list, utils.FileNameFmt(name), typeOps.File)
	archiveUrl, err := helpers.UploadToB2(media, path, typeOps.MIME)
	if err != nil {
		return "", fmt.Errorf("[archive][UploadToB2]: %w", err)
	}

	// For articles, also upload a SingleFile HTML snapshot to B2 for later use.
	// Errors are non-fatal — the MD archive is the primary output.
	if typeName == "articles" {
		sfData, sfErr := helpers.GetSingleFile(url)
		if sfErr != nil {
			log.Printf("[archive][GetSingleFile]: %v", sfErr)
		} else {
			sfPath := fmt.Sprintf("Bookmarks/Articles/%s.html", utils.FileNameFmt(name))
			if _, sfUploadErr := helpers.UploadToB2(sfData, sfPath, "text/html"); sfUploadErr != nil {
				log.Printf("[archive][UploadToB2 SingleFile]: %v", sfUploadErr)
			}
		}
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
		music, err := helpers.GetMusicInfo(title, r.GetString("creator"), r.GetInt("year"), mediaType)
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

// resolveTagNames looks up meta records by name and returns their IDs.
// Allows callers to send tag names instead of opaque relation IDs.
func resolveTagNames(app core.App, names []string) ([]string, error) {
	if len(names) == 0 {
		return nil, nil
	}

	parts := make([]string, len(names))
	for i, name := range names {
		parts[i] = fmt.Sprintf(`name = "%s"`, name)
	}
	filter := fmt.Sprintf(`(%s) && type = "tags"`, strings.Join(parts, " || "))

	records, err := app.FindRecordsByFilter("meta", filter, "", 0, 0)
	if err != nil {
		return nil, fmt.Errorf("[resolveTagNames]: %w", err)
	}

	ids := make([]string, len(records))
	for i, r := range records {
		ids[i] = r.Id
	}
	return ids, nil
}

// resolveMetaName looks up a single meta record by name and type, returning its ID.
func resolveMetaName(app core.App, name, metaType string) (string, error) {
	filter := fmt.Sprintf(`name = "%s" && type = "%s"`, name, metaType)
	record, err := app.FindFirstRecordByFilter("meta", filter)
	if err != nil {
		return "", fmt.Errorf("[resolveMetaName] %q (%s): %w", name, metaType, err)
	}
	return record.Id, nil
}

func prepareBookmarkOrFeed(app core.App, r *core.Record) error {
	r.Set("dead", false)
	r.Set("shared", false)

	if tagNames := r.GetStringSlice("tags"); len(tagNames) > 0 {
		tagIDs, err := resolveTagNames(app, tagNames)
		if err != nil {
			return fmt.Errorf("[prepareBookmarkOrFeed]: %w", err)
		}
		r.Set("tags", tagIDs)
	}
	return nil
}

func prepareMedia(app core.App, r *core.Record) error {
	if genreName := r.GetString("genre"); genreName != "" {
		genreID, err := resolveMetaName(app, genreName, "genre")
		if err != nil {
			return fmt.Errorf("[prepareMedia]: %w", err)
		}
		r.Set("genre", genreID)
	}
	return nil
}

func main() {
	app := pocketbase.New()

	// Automigrate: on startup, applies any pending migrations in the migrations/ package.
	// In dev mode (binary built from source), also auto-generates migration files when
	// collections are modified via the admin UI.
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: true,
	})

	// preparers run before e.Next() — set defaults and resolve relation names to IDs.
	preparers := map[string]func(core.App, *core.Record) error{
		"bookmarks": prepareBookmarkOrFeed,
		"feeds":     prepareBookmarkOrFeed,
		"media":     prepareMedia,
	}

	// enrichers run after e.Next() — call external APIs and write enriched fields back.
	enrichers := map[string]func(*core.Record) (bool, error){
		"bookmarks": enrichBookmarks,
		"github":    enrichGithub,
		"mtg":       enrichMtg,
		"media":     enrichMedia,
	}

	app.OnRecordCreateRequest("bookmarks", "feeds", "github", "mtg", "media").BindFunc(func(e *core.RecordRequestEvent) error {
		if fn := preparers[e.Collection.Name]; fn != nil {
			if err := fn(e.App, e.Record); err != nil {
				return fmt.Errorf("[OnRecordCreateRequest]: %w", err)
			}
		}

		if err := e.Next(); err != nil {
			return err
		}

		fn := enrichers[e.Collection.Name]
		if fn == nil {
			return nil
		}

		needsSave, err := fn(e.Record)
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
