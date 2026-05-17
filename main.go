package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
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
	filename := fmt.Sprintf("%s/%s.%s", list, utils.FileNameFmt(name), typeOps.File)
	archiveUrl, err := helpers.UploadToB2(media, "bookmarks", filename, typeOps.MIME)
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
			sfFilename := fmt.Sprintf("Articles/%s.html", utils.FileNameFmt(name))
			if _, sfUploadErr := helpers.UploadToB2(sfData, "bookmarks", sfFilename, "text/html"); sfUploadErr != nil {
				log.Printf("[archive][UploadToB2 SingleFile]: %v", sfUploadErr)
			}
		}
	}

	return archiveUrl, nil
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

func uploadCoverToB2(coverURL, collection, filename string) (string, error) {
	data, err := downloadCover(coverURL)
	if err != nil {
		return "", fmt.Errorf("[uploadCoverToB2]: %w", err)
	}
	return helpers.UploadToB2(data, collection, filename, "image/jpeg")
}

// ── Enrichers ────────────────────────────────────────────────────────────────

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

	// Only overwrite card fields when caller didn't provide full data.
	// rarity is a reliable sentinel — always set by Scryfall, never by the caller alone.
	if r.GetString("rarity") == "" {
		r.Set("colors", card.Colors)
		r.Set("type", card.Type)
		r.Set("set_name", card.SetName)
		r.Set("oracle_text", card.OracleText)
		r.Set("flavor_text", card.FlavorText)
		r.Set("rarity", card.Rarity)
		r.Set("artist", card.Artist)
		r.Set("released_at", card.ReleasedAt)
		if card.Back != nil {
			r.Set("back", card.Back)
		}
	}

	// Download front image from Scryfall, upload to B2, store B2 URL.
	if card.Image != "" {
		imageFile := fmt.Sprintf("%s/%s.jpeg", r.GetString("set"), utils.FileNameFmt(r.GetString("name")))
		b2ImageURL, err := uploadCoverToB2(card.Image, "mtg", imageFile)
		if err != nil {
			return false, fmt.Errorf("[enrichMtg]: %w", err)
		}
		r.Set("image", b2ImageURL)
	}

	// Same for back face when present.
	if card.Back != nil && *card.Back != "" {
		backFile := fmt.Sprintf("%s/%s-back.jpeg", r.GetString("set"), utils.FileNameFmt(r.GetString("name")))
		b2BackURL, err := uploadCoverToB2(*card.Back, "mtg", backFile)
		if err != nil {
			return false, fmt.Errorf("[enrichMtg]: %w", err)
		}
		r.Set("back", b2BackURL)
	}

	return true, nil
}

func enrichBooks(r *core.Record) (bool, error) {
	isbn := r.GetString("isbn")
	if isbn == "" {
		return false, nil
	}

	book, err := helpers.GetBookInfo(isbn)
	if err != nil {
		return false, fmt.Errorf("[enrichBooks]: %w", err)
	}

	var needsSave bool
	if book.Year != 0 {
		r.Set("year", book.Year)
		needsSave = true
	}
	if book.CoverURL != "" {
		b2URL, err := uploadCoverToB2(book.CoverURL, "books", fmt.Sprintf("%s.jpeg", utils.FileNameFmt(r.GetString("title"))))
		if err != nil {
			return false, fmt.Errorf("[enrichBooks]: %w", err)
		}
		r.Set("cover", b2URL)
		needsSave = true
	}
	return needsSave, nil
}

func enrichCds(r *core.Record) (bool, error) {
	album := r.GetString("album")
	music, err := helpers.GetMusicInfo(album, r.GetString("artist"), r.GetInt("year"), r.GetString("barcode"), "cds")
	if err != nil {
		return false, fmt.Errorf("[enrichCds]: %w", err)
	}

	var needsSave bool
	if music.Year != "" {
		if y, err := strconv.Atoi(music.Year); err == nil && y != 0 {
			r.Set("year", y)
			needsSave = true
		}
	}
	if music.CoverURL != "" {
		b2URL, err := uploadCoverToB2(music.CoverURL, "cds", fmt.Sprintf("%s.jpeg", utils.FileNameFmt(album)))
		if err != nil {
			return false, fmt.Errorf("[enrichCds]: %w", err)
		}
		r.Set("cover", b2URL)
		needsSave = true
	}
	return needsSave, nil
}

func enrichGames(r *core.Record) (bool, error) {
	title := r.GetString("title")
	game, err := helpers.GetGameInfo(title, r.GetInt("year"))
	if err != nil {
		return false, fmt.Errorf("[enrichGames]: %w", err)
	}

	var needsSave bool
	if game.Year != 0 {
		r.Set("year", game.Year)
		needsSave = true
	}
	if game.CoverURL != "" {
		b2URL, err := uploadCoverToB2(game.CoverURL, "games", fmt.Sprintf("%s.jpeg", utils.FileNameFmt(title)))
		if err != nil {
			return false, fmt.Errorf("[enrichGames]: %w", err)
		}
		r.Set("cover", b2URL)
		needsSave = true
	}
	return needsSave, nil
}

func enrichMovies(r *core.Record) (bool, error) {
	title := r.GetString("title")
	media, err := helpers.SearchMedia(title, r.GetInt("year"), 0, "movies")
	if err != nil {
		return false, fmt.Errorf("[enrichMovies]: %w", err)
	}

	var needsSave bool
	if media.Year != "" {
		if y, err := strconv.Atoi(media.Year); err == nil && y != 0 {
			r.Set("year", y)
			needsSave = true
		}
	}
	if media.CoverURL != "" {
		b2URL, err := uploadCoverToB2(media.CoverURL, "movies", fmt.Sprintf("%s.jpeg", utils.FileNameFmt(title)))
		if err != nil {
			return false, fmt.Errorf("[enrichMovies]: %w", err)
		}
		r.Set("cover", b2URL)
		needsSave = true
	}
	return needsSave, nil
}

func enrichShows(r *core.Record) (bool, error) {
	title := r.GetString("title")
	media, err := helpers.SearchMedia(title, r.GetInt("year"), r.GetInt("season"), "shows")
	if err != nil {
		return false, fmt.Errorf("[enrichShows]: %w", err)
	}

	var needsSave bool
	if media.Year != "" {
		if y, err := strconv.Atoi(media.Year); err == nil && y != 0 {
			r.Set("year", y)
			needsSave = true
		}
	}
	if media.CoverURL != "" {
		b2URL, err := uploadCoverToB2(media.CoverURL, "shows", fmt.Sprintf("%s.jpeg", utils.FileNameFmt(title)))
		if err != nil {
			return false, fmt.Errorf("[enrichShows]: %w", err)
		}
		r.Set("cover", b2URL)
		needsSave = true
	}
	return needsSave, nil
}

func enrichVinyls(r *core.Record) (bool, error) {
	album := r.GetString("album")
	music, err := helpers.GetMusicInfo(album, r.GetString("artist"), r.GetInt("year"), r.GetString("barcode"), "vinyls")
	if err != nil {
		return false, fmt.Errorf("[enrichVinyls]: %w", err)
	}

	var needsSave bool
	if music.Year != "" {
		if y, err := strconv.Atoi(music.Year); err == nil && y != 0 {
			r.Set("year", y)
			needsSave = true
		}
	}
	if music.CoverURL != "" {
		b2URL, err := uploadCoverToB2(music.CoverURL, "vinyls", fmt.Sprintf("%s.jpeg", utils.FileNameFmt(album)))
		if err != nil {
			return false, fmt.Errorf("[enrichVinyls]: %w", err)
		}
		r.Set("cover", b2URL)
		needsSave = true
	}
	return needsSave, nil
}

// ── Meta name resolvers ───────────────────────────────────────────────────────

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

// ── Preparers ─────────────────────────────────────────────────────────────────

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

func prepareWithGenre(app core.App, r *core.Record) error {
	if name := r.GetString("genre"); name != "" {
		id, err := resolveMetaName(app, name, "genre")
		if err != nil {
			return fmt.Errorf("[prepareWithGenre]: %w", err)
		}
		r.Set("genre", id)
	}
	return nil
}

func prepareMovieOrShow(app core.App, r *core.Record) error {
	if err := prepareWithGenre(app, r); err != nil {
		return err
	}
	if name := r.GetString("definition"); name != "" {
		id, err := resolveMetaName(app, name, "definition")
		if err != nil {
			return fmt.Errorf("[prepareMovieOrShow]: %w", err)
		}
		r.Set("definition", id)
	}
	return nil
}

func prepareGame(app core.App, r *core.Record) error {
	if err := prepareWithGenre(app, r); err != nil {
		return err
	}
	if name := r.GetString("platform"); name != "" {
		id, err := resolveMetaName(app, name, "platform")
		if err != nil {
			return fmt.Errorf("[prepareGame]: %w", err)
		}
		r.Set("platform", id)
	}
	return nil
}

// ── Main ──────────────────────────────────────────────────────────────────────

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
		"books":     prepareWithGenre,
		"cds":       prepareWithGenre,
		"games":     prepareGame,
		"movies":    prepareMovieOrShow,
		"shows":     prepareMovieOrShow,
		"vinyls":    prepareWithGenre,
	}

	// enrichers run after e.Next() — call external APIs and write enriched fields back.
	enrichers := map[string]func(*core.Record) (bool, error){
		"bookmarks": enrichBookmarks,
		"github":    enrichGithub,
		"mtg":       enrichMtg,
		"books":     enrichBooks,
		"cds":       enrichCds,
		"games":     enrichGames,
		"movies":    enrichMovies,
		"shows":     enrichShows,
		"vinyls":    enrichVinyls,
	}

	app.OnRecordCreateRequest(
		"bookmarks", "feeds", "github", "mtg",
		"books", "cds", "games", "movies", "shows", "vinyls",
	).BindFunc(func(e *core.RecordRequestEvent) error {
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
