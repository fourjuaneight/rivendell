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

var discogsFormatMap = map[string]string{
	"cds":    "CD",
	"vinyls": "Vinyl",
}

func discogsSearch(token string, params neturl.Values) ([]discogsSearchResult, error) {
	endpoint := fmt.Sprintf("%s/database/search?%s", discogsBaseURL, params.Encode())
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("[discogsSearch][http.NewRequest]: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Discogs token=%s", token))
	req.Header.Set("User-Agent", "Rivendell/1.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[discogsSearch][client.Do]: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[discogsSearch][io.ReadAll]: %w", err)
	}

	var result discogsSearchResponse
	if err = json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("[discogsSearch][json.Unmarshal]: %w", err)
	}
	return result.Results, nil
}

// DOCS: https://www.discogs.com/developers
func GetMusicInfo(title, artist string, year int, barcode, mediaType string) (CleanMusic, error) {
	token, err := GetKeys("DISCOGS_TOKEN")
	if err != nil {
		return CleanMusic{}, fmt.Errorf("[GetMusicInfo]%w", err)
	}

	var results []discogsSearchResult

	// Barcode search first — identifies the exact pressing.
	if barcode != "" {
		barcodeParams := neturl.Values{}
		barcodeParams.Set("barcode", barcode)
		barcodeParams.Set("type", "release")
		barcodeParams.Set("per_page", "1")
		results, err = discogsSearch(token, barcodeParams)
		if err != nil {
			return CleanMusic{}, fmt.Errorf("[GetMusicInfo]%w", err)
		}
	}

	// Fall back to title+artist search if barcode yielded nothing.
	if len(results) == 0 {
		params := neturl.Values{}
		params.Set("release_title", title)
		params.Set("artist", artist)
		if year != 0 {
			params.Set("year", fmt.Sprintf("%d", year))
		}
		params.Set("per_page", "1")

		// Use format-specific release search when media type is known; otherwise fall back to master.
		// Masters are format-agnostic so format= has no effect when type=master.
		if format, ok := discogsFormatMap[mediaType]; ok {
			params.Set("type", "release")
			params.Set("format", format)
		} else {
			// master = canonical release; avoids duplicate pressing-specific results
			params.Set("type", "master")
		}

		results, err = discogsSearch(token, params)
		if err != nil {
			return CleanMusic{}, fmt.Errorf("[GetMusicInfo]%w", err)
		}
	}

	if len(results) == 0 {
		return CleanMusic{}, fmt.Errorf("[GetMusicInfo]: no results for %q by %q (%d)", title, artist, year)
	}

	r := results[0]
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
