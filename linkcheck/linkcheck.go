package linkcheck

import (
	"net/http"
	"time"
)

const checkTimeout = 10 * time.Second

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
