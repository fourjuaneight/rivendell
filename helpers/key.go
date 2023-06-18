package helpers

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Get auth keys from .env file.
func GetKeys(key string) (string, error) {
	envPath := os.Getenv("PWD") + "/.env"
	err := godotenv.Load(envPath)
	if err != nil {
		return "", fmt.Errorf("[GetAuthKeys]: %w", err)
	}

	APP_KEY_ID := os.Getenv("B2_APP_KEY_ID")
	APP_KEY := os.Getenv("B2_APP_KEY")
	BUCKET_ID := os.Getenv("B2_BUCKET_ID")
	BUCKET_NAME := os.Getenv("B2_BUCKET_NAME")
	GH_TOKEN := os.Getenv("GH_TOKEN")
	GH_USERNAME := os.Getenv("GH_USERNAME")
	TMDB_KEY := os.Getenv("TMDB_KEY")
	YOUTUBE_KEY := os.Getenv("YOUTUBE_KEY")

	keys := map[string]string{
		"APP_KEY_ID":  APP_KEY_ID,
		"APP_KEY":     APP_KEY,
		"BUCKET_ID":   BUCKET_ID,
		"BUCKET_NAME": BUCKET_NAME,
		"GH_TOKEN":    GH_TOKEN,
		"GH_USERNAME": GH_USERNAME,
		"TMDB_KEY":    TMDB_KEY,
		"YOUTUBE_KEY": YOUTUBE_KEY,
	}

	return keys[key], nil
}
