package main

import (
	"fmt"
	"log"

	"github.com/fourjuaneight/rivendell/backup"
	"github.com/fourjuaneight/rivendell/linkcheck"
	_ "github.com/fourjuaneight/rivendell/migrations"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

// logMaxDays is the request/app log retention window, in days. Long enough that
// a data-loss incident noticed weeks later still has attributable log entries.
const logMaxDays = 90

func main() {
	app := pocketbase.New()

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: true,
	})

	// Log retention is the only record of who deleted what. The PocketBase
	// default (5 days) is shorter than the gap between noticing data loss and
	// investigating it, and LogAuthId is off by default, so entries can't be
	// attributed. Enforced on every boot rather than in a migration: the
	// migrations directory is gitignored, so a migration wouldn't survive a
	// fresh clone.
	app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		settings := e.App.Settings()
		if settings.Logs.MaxDays == logMaxDays && settings.Logs.LogAuthId {
			return nil
		}

		settings.Logs.MaxDays = logMaxDays
		settings.Logs.LogAuthId = true

		if err := e.App.Save(settings); err != nil {
			return fmt.Errorf("[OnBootstrap][settings]: %w", err)
		}

		e.App.Logger().Info("log settings enforced",
			"max_days", logMaxDays,
			"log_auth_id", true,
		)

		return nil
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

	// Audit trail for deletes. Nothing in this app deletes records, so every
	// delete is an external superuser action (dashboard or API token). Without
	// this, a bulk delete leaves no attributable trace once the request log
	// ages out. Logged before e.Next() so an attempt is recorded even if the
	// delete later fails.
	app.OnRecordDeleteRequest().BindFunc(func(e *core.RecordRequestEvent) error {
		collection := e.Collection.Name

		var authID string
		if e.Auth != nil {
			authID = e.Auth.Id
		}

		app.Logger().Warn("record delete requested",
			"collection", collection,
			"record", recordLabel(collection, e.Record),
			"record_id", e.Record.Id,
			"auth_id", authID,
			"ip", e.RealIP(),
		)

		return e.Next()
	})

	app.Cron().MustAdd("link_check", "0 3 * * *", func() {
		app.Logger().Info("link_check started")
		for _, name := range []string{"articles", "podcasts", "videos"} {
			linkcheck.CheckCollection(app, name)
		}
	})

	app.Cron().MustAdd("backup", "0 4 * * *", func() {
		backup.BackupAll(app)
	})

	if err := app.Start(); err != nil {
		log.Fatal("[Start]: %w", err)
	}
}
