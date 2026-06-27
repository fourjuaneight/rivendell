package linkcheck

import "testing"

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
