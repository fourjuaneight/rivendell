// Package linkcheck checks the liveness of saved article/podcast/video URLs
// and flips a record's dead field when its link no longer resolves. See
// CheckCollection for the entry point used by the link_check cron job
// (registered in main.go); the rest of this file's exported surface is
// CheckURL, usable standalone for a single-URL liveness check.
package linkcheck

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// checkTimeout bounds each individual HTTP request (HEAD or GET). Long
// enough for slow servers, short enough that one dead host doesn't stall a
// worker for the rest of a collection's run.
const checkTimeout = 10 * time.Second

// workerLimit caps concurrent in-flight requests per CheckCollection call.
// Bounds total load on the checking host; does not throttle per-target-host
// (see docs/CRON.md's "Known limitation" note for link_check).
const workerLimit = 5

// isDeadStatus reports whether an HTTP status code indicates a dead link.
// 401/403/429 are treated as alive — anti-bot walls and rate-limits are not
// real link death.
func isDeadStatus(code int) bool {
	switch {
	case code >= 200 && code < 400:
		return false
	case code == 401 || code == 403 || code == 429:
		return false
	default:
		return true
	}
}

// doRequest issues a single HTTP request and returns its status code. Sets a
// browser-like User-Agent — Go's default UA gets blocked or misclassified
// (wrongly returning 403/404/5xx) by some hosts, which would otherwise show
// up as a false dead-link verdict.
func doRequest(method, url string) (int, error) {
	client := &http.Client{Timeout: checkTimeout}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; RivendellLinkChecker/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}

// attempt tries HEAD first; falls back to GET if the server doesn't support
// HEAD (405/501). Returns the final status code or a network-level error.
func attempt(url string) (int, error) {
	status, err := doRequest(http.MethodHead, url)
	if err != nil {
		return 0, err
	}
	if status == http.StatusMethodNotAllowed || status == http.StatusNotImplemented {
		return doRequest(http.MethodGet, url)
	}
	return status, nil
}

// CheckURL reports whether url is alive. On network error/timeout it retries
// once before giving up; a definitive HTTP status is never retried.
func CheckURL(url string) bool {
	status, err := attempt(url)
	if err != nil {
		status, err = attempt(url)
		if err != nil {
			return false
		}
	}
	return !isDeadStatus(status)
}

// CheckCollection checks the url field of every non-dead record in
// collectionName and flips dead=true on records whose link no longer
// resolves. Records are checked concurrently, capped at workerLimit
// in-flight requests. A single record's check or save failure is logged and
// skipped — it never aborts the rest of the collection; only a failure to
// query collectionName itself aborts (returned as an error so the caller can
// still run other collections). Every outcome — query failure, save
// failure, dead-flip, and a final checked/flipped summary — is logged via
// app.Logger() so it's visible in PocketBase's admin Logs panel.
func CheckCollection(app core.App, collectionName string) error {
	records, err := app.FindRecordsByFilter(collectionName, "dead = false", "", 0, 0)
	if err != nil {
		app.Logger().Error("link_check query failed", "collection", collectionName, "error", err.Error())
		return fmt.Errorf("[CheckCollection][%s]: %w", collectionName, err)
	}

	var checked, flipped atomic.Int64
	sem := make(chan struct{}, workerLimit)
	var wg sync.WaitGroup

	for _, r := range records {
		record := r
		wg.Add(1)
		sem <- struct{}{}

		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			checked.Add(1)

			url := record.GetString("url")
			if url == "" || CheckURL(url) {
				return
			}

			record.Set("dead", true)
			if err := app.Save(record); err != nil {
				app.Logger().Error("link_check save failed", "collection", collectionName, "record_id", record.Id, "error", err.Error())
				return
			}

			flipped.Add(1)
			app.Logger().Info("link_check marked dead", "collection", collectionName, "record_id", record.Id, "url", url)
		}()
	}

	wg.Wait()
	app.Logger().Info("link_check collection done", "collection", collectionName, "checked", checked.Load(), "flipped", flipped.Load())
	return nil
}
