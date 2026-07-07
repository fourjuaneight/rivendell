package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/fourjuaneight/rivendell/helpers"
	"github.com/fourjuaneight/rivendell/utils"

	"github.com/pocketbase/pocketbase/core"
)

// recordLabel returns a human-readable identifier for a record based on its collection.
func recordLabel(collection string, r *core.Record) string {
	switch collection {
	case "cds", "vinyls":
		return r.GetString("album")
	case "github":
		return r.GetString("url")
	case "watch_later":
		return r.GetString("link")
	case "mtg":
		return r.GetString("name")
	default:
		return r.GetString("title")
	}
}

func archive(app core.App, name string, url string, typeName string, creator string, year int) (string, error) {
	app.Logger().Info("archive started",
		"collection", typeName,
		"title", name,
		"url", url,
	)

	media, err := helpers.GetContent(name, url, typeName, creator, year)
	if err != nil {
		return "", fmt.Errorf("[archive][GetContent]: %w", err)
	}

	typeOps := utils.GetFileType(typeName, url)
	list := utils.ToCapitalized(typeName)
	filename := fmt.Sprintf("%s/%s.%s", list, utils.FileNameFmt(name), typeOps.File)
	archiveUrl, err := helpers.UploadToB2(media, typeName, filename, typeOps.MIME)
	if err != nil {
		return "", fmt.Errorf("[archive][UploadToB2]: %w", err)
	}

	app.Logger().Info("archive uploaded",
		"collection", typeName,
		"title", name,
		"b2_path", filename,
	)

	// For articles, also upload a SingleFile HTML snapshot to B2 for later use.
	// Errors are non-fatal — the MD archive is the primary output.
	if typeName == "articles" {
		sfData, sfErr := helpers.GetSingleFile(url)
		if sfErr != nil {
			app.Logger().Error("archive SingleFile capture failed",
				"collection", typeName,
				"title", name,
				"url", url,
				"error", sfErr.Error(),
			)
		} else {
			sfFilename := fmt.Sprintf("Articles/%s.html", utils.FileNameFmt(name))
			if _, sfUploadErr := helpers.UploadToB2(sfData, "articles", sfFilename, "text/html"); sfUploadErr != nil {
				app.Logger().Error("archive SingleFile upload failed",
					"collection", typeName,
					"title", name,
					"error", sfUploadErr.Error(),
				)
			} else {
				app.Logger().Info("archive SingleFile uploaded",
					"collection", typeName,
					"title", name,
					"b2_path", sfFilename,
				)
			}
		}
	}

	return archiveUrl, nil
}

func downloadCover(url string) ([]byte, error) {
	resp, err := helpers.MediaClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("[downloadCover][http.Get]: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[downloadCover][resp]: %d - %s", resp.StatusCode, resp.Status)
	}

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

func enrichArticles(app core.App, r *core.Record) (bool, error) {
	archiveURL, err := archive(app, r.GetString("title"), r.GetString("url"), "articles", "", 0)
	if err != nil {
		return false, fmt.Errorf("[enrichArticles]: %w", err)
	}
	r.Set("archive", archiveURL)
	return true, nil
}

func enrichPodcasts(app core.App, r *core.Record) (bool, error) {
	if r.GetString("archive") != "" {
		app.Logger().Info("enrich skipped, archive already set",
			"collection", "podcasts",
			"title", r.GetString("title"),
		)
		return false, nil
	}
	archiveURL, err := archive(app, r.GetString("title"), r.GetString("url"), "podcasts", r.GetString("creator"), r.GetInt("year"))
	if err != nil {
		return false, fmt.Errorf("[enrichPodcasts]: %w", err)
	}
	r.Set("archive", archiveURL)
	return true, nil
}

func enrichVideos(app core.App, r *core.Record) (bool, error) {
	if r.GetString("archive") != "" {
		app.Logger().Info("enrich skipped, archive already set",
			"collection", "videos",
			"url", r.GetString("url"),
		)
		return false, nil
	}

	app.Logger().Info("enrich fetching YouTube metadata",
		"collection", "videos",
		"url", r.GetString("url"),
	)

	yt, err := helpers.GetYTInfo(r.GetString("url"))
	if err != nil {
		return false, fmt.Errorf("[enrichVideos]: %w", err)
	}
	r.Set("title", yt.Title)
	r.Set("creator", yt.Creator)
	r.Set("url", yt.URL)
	r.Set("year", yt.Year)

	app.Logger().Info("enrich YouTube metadata resolved",
		"collection", "videos",
		"title", yt.Title,
		"creator", yt.Creator,
	)

	archiveURL, err := archive(app, yt.Title, yt.URL, "videos", yt.Creator, yt.Year)
	if err != nil {
		return false, fmt.Errorf("[enrichVideos]: %w", err)
	}
	r.Set("archive", archiveURL)

	return true, nil
}

func enrichGithub(app core.App, r *core.Record) (bool, error) {
	url := r.GetString("url")
	app.Logger().Info("enrich fetching repo info",
		"collection", "github",
		"url", url,
	)

	repo, err := helpers.GetRepoInfo(url)
	if err != nil {
		return false, fmt.Errorf("[enrichGithub]: %w", err)
	}
	r.Set("name", repo.Name)
	r.Set("owner", repo.Owner)
	r.Set("description", repo.Description)
	r.Set("language", repo.Language)

	app.Logger().Info("enrich repo resolved",
		"collection", "github",
		"name", repo.Name,
		"owner", repo.Owner,
		"language", repo.Language,
	)
	return true, nil
}

func enrichMtg(app core.App, r *core.Record) (bool, error) {
	name := r.GetString("name")
	set := r.GetString("set")
	number := r.GetInt("collector_number")

	app.Logger().Info("enrich searching Scryfall",
		"collection", "mtg",
		"name", name,
		"set", set,
		"collector_number", number,
	)

	cardSelection, err := helpers.SearchCard(name, set, number)
	if err != nil {
		return false, fmt.Errorf("[enrichMtg]: %w", err)
	}

	if len(cardSelection) == 0 {
		return false, fmt.Errorf("[enrichMtg]: no card found for %s/%d", set, number)
	}

	var card helpers.MTGItem
	for _, c := range cardSelection {
		card = c
		break
	}

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

	if card.Image != "" {
		imageFile := fmt.Sprintf("%s/%s.jpeg", set, utils.FileNameFmt(name))
		b2ImageURL, err := uploadCoverToB2(card.Image, "mtg", imageFile)
		if err != nil {
			return false, fmt.Errorf("[enrichMtg]: %w", err)
		}
		r.Set("image", b2ImageURL)
		app.Logger().Info("enrich card image uploaded",
			"collection", "mtg",
			"name", name,
			"b2_path", imageFile,
		)
	}

	if card.Back != nil && *card.Back != "" {
		backFile := fmt.Sprintf("%s/%s-back.jpeg", set, utils.FileNameFmt(name))
		b2BackURL, err := uploadCoverToB2(*card.Back, "mtg", backFile)
		if err != nil {
			return false, fmt.Errorf("[enrichMtg]: %w", err)
		}
		r.Set("back", b2BackURL)
		app.Logger().Info("enrich card back image uploaded",
			"collection", "mtg",
			"name", name,
			"b2_path", backFile,
		)
	}

	return true, nil
}

func enrichBooks(app core.App, r *core.Record) (bool, error) {
	title := r.GetString("title")
	isbn := r.GetString("isbn")
	if isbn == "" {
		app.Logger().Info("enrich skipped, no ISBN",
			"collection", "books",
			"title", title,
		)
		return false, nil
	}

	app.Logger().Info("enrich fetching book info",
		"collection", "books",
		"title", title,
		"isbn", isbn,
	)

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
		b2URL, err := uploadCoverToB2(book.CoverURL, "books", fmt.Sprintf("%s.jpeg", utils.FileNameFmt(title)))
		if err != nil {
			return false, fmt.Errorf("[enrichBooks]: %w", err)
		}
		r.Set("cover", b2URL)
		needsSave = true
		app.Logger().Info("enrich cover uploaded",
			"collection", "books",
			"title", title,
		)
	}
	return needsSave, nil
}

// enrichMusic handles both CDs and vinyls — same Discogs lookup, different collection name.
func enrichMusic(app core.App, r *core.Record, collection string) (bool, error) {
	album := r.GetString("album")
	artist := r.GetString("artist")

	app.Logger().Info("enrich fetching music info",
		"collection", collection,
		"album", album,
		"artist", artist,
	)

	music, err := helpers.GetMusicInfo(album, artist, r.GetInt("year"), r.GetString("barcode"), collection)
	if err != nil {
		return false, fmt.Errorf("[enrich%s]: %w", utils.ToCapitalized(collection), err)
	}

	var needsSave bool
	if music.Year != "" {
		if y, err := strconv.Atoi(music.Year); err == nil && y != 0 {
			r.Set("year", y)
			needsSave = true
		}
	}
	if music.CoverURL != "" {
		b2URL, err := uploadCoverToB2(music.CoverURL, collection, fmt.Sprintf("%s.jpeg", utils.FileNameFmt(album)))
		if err != nil {
			return false, fmt.Errorf("[enrich%s]: %w", utils.ToCapitalized(collection), err)
		}
		r.Set("cover", b2URL)
		needsSave = true
		app.Logger().Info("enrich cover uploaded",
			"collection", collection,
			"album", album,
		)
	}
	return needsSave, nil
}

func enrichCds(app core.App, r *core.Record) (bool, error) {
	return enrichMusic(app, r, "cds")
}

func enrichVinyls(app core.App, r *core.Record) (bool, error) {
	return enrichMusic(app, r, "vinyls")
}

func enrichGames(app core.App, r *core.Record) (bool, error) {
	title := r.GetString("title")

	app.Logger().Info("enrich fetching game info",
		"collection", "games",
		"title", title,
	)

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
		app.Logger().Info("enrich cover uploaded",
			"collection", "games",
			"title", title,
		)
	}
	return needsSave, nil
}

// enrichTMDB handles both movies and shows — same TMDB search, different cover file naming.
func enrichTMDB(app core.App, r *core.Record, collection string) (bool, error) {
	title := r.GetString("title")
	season := r.GetInt("season")

	app.Logger().Info("enrich fetching TMDB info",
		"collection", collection,
		"title", title,
		"season", season,
		"director", r.GetString("director"),
	)

	media, err := helpers.SearchMedia(title, r.GetInt("year"), season, collection, r.GetString("director"))
	if err != nil {
		return false, fmt.Errorf("[enrich%s]: %w", utils.ToCapitalized(collection), err)
	}

	var needsSave bool
	if media.Year != "" {
		if y, err := strconv.Atoi(media.Year); err == nil && y != 0 {
			r.Set("year", y)
			needsSave = true
		}
	}
	if media.CoverURL != "" {
		var coverFile string
		if collection == "shows" && season > 0 {
			coverFile = fmt.Sprintf("%s-s%02d.jpeg", utils.FileNameFmt(title), season)
		} else if collection == "movies" {
			coverFile = fmt.Sprintf("%s-%s.jpeg", utils.FileNameFmt(title), media.Year)
		} else {
			coverFile = fmt.Sprintf("%s.jpeg", utils.FileNameFmt(title))
		}

		b2URL, err := uploadCoverToB2(media.CoverURL, collection, coverFile)
		if err != nil {
			return false, fmt.Errorf("[enrich%s]: %w", utils.ToCapitalized(collection), err)
		}
		r.Set("cover", b2URL)
		needsSave = true
		app.Logger().Info("enrich cover uploaded",
			"collection", collection,
			"title", title,
		)
	}
	return needsSave, nil
}

func enrichMovies(app core.App, r *core.Record) (bool, error) {
	return enrichTMDB(app, r, "movies")
}

func enrichShows(app core.App, r *core.Record) (bool, error) {
	return enrichTMDB(app, r, "shows")
}

func enrichWatchLater(app core.App, r *core.Record) (bool, error) {
	link := r.GetString("link")
	app.Logger().Info("enrich fetching YouTube info",
		"collection", "watch_later",
		"link", link,
	)

	yt, err := helpers.GetYTInfo(link)
	if err != nil {
		return false, fmt.Errorf("[enrichWatchLater]: %w", err)
	}
	r.Set("title", yt.Title)
	r.Set("channel", yt.Creator)

	app.Logger().Info("enrich YouTube info resolved",
		"collection", "watch_later",
		"title", yt.Title,
		"channel", yt.Creator,
	)
	return true, nil
}
