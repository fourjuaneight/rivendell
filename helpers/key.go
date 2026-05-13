package helpers

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Get auth keys from .env file.
func GetKeys(key string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("[GetKeys][os.Getwd]: %w", err)
	}
	envPath := cwd + "/.env"
	err = godotenv.Load(envPath)
	if err != nil {
		return "", fmt.Errorf("[GetAuthKeys]: %w", err)
	}

	APP_KEY_ID := os.Getenv("B2_APP_KEY_ID")
	APP_KEY := os.Getenv("B2_APP_KEY")
	BUCKET_ID := os.Getenv("B2_BUCKET_ID")
	BUCKET_NAME := os.Getenv("B2_BUCKET_NAME")
	DISCOGS_TOKEN := os.Getenv("DISCOGS_TOKEN")
	GH_TOKEN := os.Getenv("GH_TOKEN")
	GH_USERNAME := os.Getenv("GH_USERNAME")
	META_ID := os.Getenv("META_ID")
	TMDB_KEY := os.Getenv("TMDB_KEY")
	TWITCH_CLIENT_ID := os.Getenv("TWITCH_CLIENT_ID")
	TWITCH_CLIENT_SECRET := os.Getenv("TWITCH_CLIENT_SECRET")
	YOUTUBE_KEY := os.Getenv("YOUTUBE_KEY")

	keys := map[string]string{
		"APP_KEY_ID":         APP_KEY_ID,
		"APP_KEY":            APP_KEY,
		"BUCKET_ID":          BUCKET_ID,
		"BUCKET_NAME":        BUCKET_NAME,
		"DISCOGS_TOKEN":      DISCOGS_TOKEN,
		"GH_TOKEN":           GH_TOKEN,
		"GH_USERNAME":        GH_USERNAME,
		"META_ID":            META_ID,
		"TMDB_KEY":           TMDB_KEY,
		"TWITCH_CLIENT_ID":   TWITCH_CLIENT_ID,
		"TWITCH_CLIENT_SECRET": TWITCH_CLIENT_SECRET,
		"YOUTUBE_KEY":        YOUTUBE_KEY,
	}

	return keys[key], nil
}
