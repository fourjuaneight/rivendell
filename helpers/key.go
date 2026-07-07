package helpers

import (
	"fmt"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

// envOnce ensures the .env file is loaded exactly once. Previously GetKeys
// re-read and re-parsed .env on every call, and enrichment calls it several
// times per record. godotenv.Load does not overwrite already-set env vars, so a
// single load at first use is sufficient.
var (
	envOnce    sync.Once
	envLoadErr error
)

// keyToEnv maps the short logical key name callers pass to GetKeys to the
// underlying environment variable name.
var keyToEnv = map[string]string{
	"APP_KEY_ID":           "B2_APP_KEY_ID",
	"APP_KEY":              "B2_APP_KEY",
	"BUCKET_ID":            "B2_BUCKET_ID",
	"BUCKET_NAME":          "B2_BUCKET_NAME",
	"DISCOGS_TOKEN":        "DISCOGS_TOKEN",
	"GH_TOKEN":             "GH_TOKEN",
	"GH_USERNAME":          "GH_USERNAME",
	"META_ID":              "META_ID",
	"TMDB_KEY":             "TMDB_KEY",
	"TWITCH_CLIENT_ID":     "TWITCH_CLIENT_ID",
	"TWITCH_CLIENT_SECRET": "TWITCH_CLIENT_SECRET",
	"YOUTUBE_KEY":          "YOUTUBE_KEY",
}

// GetKeys returns the value of a configured secret by its logical name. The
// .env file is loaded once on first call. Returns an error for an unknown key
// name or an unset/empty value, so a missing secret fails fast here instead of
// surfacing as a cryptic upstream 401.
func GetKeys(key string) (string, error) {
	envOnce.Do(func() {
		cwd, err := os.Getwd()
		if err != nil {
			envLoadErr = fmt.Errorf("[GetKeys][os.Getwd]: %w", err)
			return
		}
		if err := godotenv.Load(cwd + "/.env"); err != nil {
			envLoadErr = fmt.Errorf("[GetKeys][godotenv.Load]: %w", err)
		}
	})
	if envLoadErr != nil {
		return "", envLoadErr
	}

	envName, ok := keyToEnv[key]
	if !ok {
		return "", fmt.Errorf("[GetKeys]: unknown key %q", key)
	}

	val := os.Getenv(envName)
	if val == "" {
		return "", fmt.Errorf("[GetKeys]: missing or empty key %q (%s)", key, envName)
	}

	return val, nil
}
