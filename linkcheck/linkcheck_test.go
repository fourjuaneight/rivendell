package linkcheck

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsDeadStatus(t *testing.T) {
	cases := []struct {
		name string
		code int
		want bool
	}{
		{"200 OK is alive", 200, false},
		{"301 redirect is alive", 301, false},
		{"399 boundary is alive", 399, false},
		{"400 bad request is dead", 400, true},
		{"401 unauthorized is alive (anti-bot)", 401, false},
		{"403 forbidden is alive (anti-bot)", 403, false},
		{"404 not found is dead", 404, true},
		{"410 gone is dead", 410, true},
		{"429 too many requests is alive (rate-limit)", 429, false},
		{"500 server error is dead", 500, true},
		{"503 unavailable is dead", 503, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := isDeadStatus(c.code)
			if got != c.want {
				t.Errorf("isDeadStatus(%d) = %v, want %v", c.code, got, c.want)
			}
		})
	}
}

func TestCheckURL(t *testing.T) {
	cases := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"alive 200", http.StatusOK, true},
		{"dead 404", http.StatusNotFound, false},
		{"dead 500", http.StatusInternalServerError, false},
		{"alive 403 anti-bot", http.StatusForbidden, true},
		{"alive 429 rate-limit", http.StatusTooManyRequests, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(c.statusCode)
			}))
			defer srv.Close()

			got := CheckURL(srv.URL)
			if got != c.want {
				t.Errorf("CheckURL() with status %d = %v, want %v", c.statusCode, got, c.want)
			}
		})
	}
}

func TestCheckURLHeadFallbackToGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	if !CheckURL(srv.URL) {
		t.Error("CheckURL should fall back to GET when HEAD returns 405")
	}
}

func TestCheckURLUnreachable(t *testing.T) {
	if CheckURL("http://127.0.0.1:1") {
		t.Error("CheckURL should return false for unreachable host")
	}
}
