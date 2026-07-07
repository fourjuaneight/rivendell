package helpers

import (
	"net/http"
	"time"
)

// Shared HTTP clients with timeouts. Enrichment runs synchronously inside the
// record-create request, so an unbounded client would let one hung upstream
// host stall the request forever. Every external call in this package — and the
// cover download in enrichers.go — goes through one of these instead of
// http.Get or a bare &http.Client{}.
var (
	// HTTPClient bounds metadata/API calls (JSON lookups, small responses).
	HTTPClient = &http.Client{Timeout: 30 * time.Second}

	// MediaClient bounds large transfers — podcast/cover downloads and B2
	// uploads, which can carry a full video file — so it uses a longer timeout
	// than the metadata client.
	MediaClient = &http.Client{Timeout: 300 * time.Second}
)
