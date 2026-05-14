package helpers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"strings"
)

const discogsBaseURL = "https://api.discogs.com"

type discogsSearchResult struct {
	Title      string   `json:"title"`
	Year       string   `json:"year"`
	CoverImage string   `json:"cover_image"`
	Genre      []string `json:"genre"`
}

type discogsSearchResponse struct {
	Results []discogsSearchResult `json:"results"`
}

type CleanMusic struct {
	Title    string
	Creator  string
	Year     string
	CoverURL string
}

// parseDiscogsTitle splits the Discogs search result title format "Artist - Album Title"
// into its components. Takes the first " - " separator; album may contain additional dashes.
func parseDiscogsTitle(title string) (artist, album string) {
	artist, album, found := strings.Cut(title, " - ")
	if !found {
		return "", title
	}
	return artist, album
}

// DOCS: https://www.discogs.com/developers
func GetMusicInfo(title, artist string, year int) (CleanMusic, error) {
	token, err := GetKeys("DISCOGS_TOKEN")
	if err != nil {
		return CleanMusic{}, fmt.Errorf("[GetMusicInfo]%w", err)
	}

	params := neturl.Values{}
	params.Set("release_title", title)
	params.Set("artist", artist)
	params.Set("year", fmt.Sprintf("%d", year))
	// master = canonical release; avoids duplicate pressing-specific results
	params.Set("type", "master")
	params.Set("per_page", "1")

	endpoint := fmt.Sprintf("%s/database/search?%s", discogsBaseURL, params.Encode())

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return CleanMusic{}, fmt.Errorf("[GetMusicInfo][http.NewRequest]: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Discogs token=%s", token))
	req.Header.Set("User-Agent", "Rivendell/1.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return CleanMusic{}, fmt.Errorf("[GetMusicInfo][client.Do]: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CleanMusic{}, fmt.Errorf("[GetMusicInfo][io.ReadAll]: %w", err)
	}

	var result discogsSearchResponse
	if err = json.Unmarshal(body, &result); err != nil {
		return CleanMusic{}, fmt.Errorf("[GetMusicInfo][json.Unmarshal]: %w", err)
	}

	if len(result.Results) == 0 {
		return CleanMusic{}, fmt.Errorf("[GetMusicInfo]: no results for %q by %q (%d)", title, artist, year)
	}

	r := result.Results[0]
	parsedArtist, parsedTitle := parseDiscogsTitle(r.Title)

	releaseYear := r.Year
	if releaseYear == "" {
		releaseYear = fmt.Sprintf("%d", year)
	}

	resolvedArtist := parsedArtist
	if resolvedArtist == "" {
		resolvedArtist = artist
	}

	return CleanMusic{
		Title:    parsedTitle,
		Creator:  resolvedArtist,
		Year:     releaseYear,
		CoverURL: r.CoverImage,
	}, nil
}
