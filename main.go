package main

import (
	"fmt"
	"log"

	"github.com/fourjuaneight/rivendell/linkcheck"
	_ "github.com/fourjuaneight/rivendell/migrations"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

func main() {
	app := pocketbase.New()

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: true,
	})

	// preparers run before e.Next() — set defaults and resolve relation names to IDs.
	preparers := map[string]func(core.App, *core.Record) error{
		"articles":    prepareArticle,
		"podcasts":    prepareMedia,
		"videos":      prepareMedia,
		"feeds":       prepareFeed,
		"books":       prepareBook,
		"cds":         prepareWithGenre,
		"games":       prepareGame,
		"movies":      prepareMovieOrShow,
		"shows":       prepareMovieOrShow,
		"vinyls":      prepareWithGenre,
		"read_later":  prepareTags,
		"watch_later": prepareTags,
	}

	// enrichers run after e.Next() — call external APIs and write enriched fields back.
	enrichers := map[string]func(core.App, *core.Record) (bool, error){
		"articles":    enrichArticles,
		"podcasts":    enrichPodcasts,
		"videos":      enrichVideos,
		"github":      enrichGithub,
		"mtg":         enrichMtg,
		"books":       enrichBooks,
		"cds":         enrichCds,
		"games":       enrichGames,
		"movies":      enrichMovies,
		"shows":       enrichShows,
		"vinyls":      enrichVinyls,
		"watch_later": enrichWatchLater,
	}

	app.OnRecordCreateRequest(
		"articles", "podcasts", "videos", "feeds", "github", "mtg",
		"books", "cds", "games", "movies", "shows", "vinyls",
		"read_later", "watch_later",
	).BindFunc(func(e *core.RecordRequestEvent) error {
		collection := e.Collection.Name
		label := recordLabel(collection, e.Record)

		app.Logger().Info("record create started",
			"collection", collection,
			"record", label,
		)

		if fn := preparers[collection]; fn != nil {
			if err := fn(e.App, e.Record); err != nil {
				app.Logger().Error("record prepare failed",
					"collection", collection,
					"record", label,
					"error", err.Error(),
				)
				return fmt.Errorf("[OnRecordCreateRequest]: %w", err)
			}
		}

		if err := e.Next(); err != nil {
			return err
		}

		fn := enrichers[collection]
		if fn == nil {
			app.Logger().Info("record created (no enrichment)",
				"collection", collection,
				"record", label,
			)
			return nil
		}

		needsSave, err := fn(e.App, e.Record)
		if err != nil {
			app.Logger().Error("record enrich failed",
				"collection", collection,
				"record", label,
				"error", err.Error(),
			)
			return fmt.Errorf("[OnRecordCreateRequest]: %w", err)
		}

		if needsSave {
			if err := e.App.Save(e.Record); err != nil {
				app.Logger().Error("record enrich save failed",
					"collection", collection,
					"record", label,
					"error", err.Error(),
				)
				return fmt.Errorf("[OnRecordCreateRequest][save]: %w", err)
			}
		}

		app.Logger().Info("record created",
			"collection", collection,
			"record", label,
			"enriched", needsSave,
		)

		return nil
	})

	app.Cron().MustAdd("link_check", "0 3 * * *", func() {
		app.Logger().Info("link_check started")
		for _, name := range []string{"articles", "podcasts", "videos"} {
			linkcheck.CheckCollection(app, name)
		}
	})

	if err := app.Start(); err != nil {
		log.Fatal("[Start]: %w", err)
	}
}
