package linkcheck

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
