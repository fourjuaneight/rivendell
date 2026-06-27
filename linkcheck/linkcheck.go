package linkcheck

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

const checkTimeout = 10 * time.Second
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

// doRequest issues a single HTTP request and returns its status code.
func doRequest(method, url string) (int, error) {
	client := &http.Client{Timeout: checkTimeout}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return 0, err
	}

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
// resolves. A single record's check or save failure is logged and skipped —
// it never aborts the rest of the collection.
func CheckCollection(app core.App, collectionName string) error {
	records, err := app.FindRecordsByFilter(collectionName, "dead = false", "", 0, 0)
	if err != nil {
		return fmt.Errorf("[CheckCollection][%s]: %w", collectionName, err)
	}

	sem := make(chan struct{}, workerLimit)
	var wg sync.WaitGroup

	for _, r := range records {
		record := r
		wg.Add(1)
		sem <- struct{}{}

		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			url := record.GetString("url")
			if url == "" || CheckURL(url) {
				return
			}

			record.Set("dead", true)
			if err := app.Save(record); err != nil {
				log.Printf("[CheckCollection][%s][save %s]: %v", collectionName, record.Id, err)
			}
		}()
	}

	wg.Wait()
	return nil
}
