package helpers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	igdbBaseURL      = "https://api.igdb.com/v4"
	igdbImageBaseURL = "https://images.igdb.com/igdb/image/upload"
	twitchTokenURL   = "https://id.twitch.tv/oauth2/token"
)

var (
	igdbTokenMu  sync.Mutex
	igdbToken    string
	igdbTokenExp time.Time
)

type twitchTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type igdbCover struct {
	ImageID string `json:"image_id"`
}

type igdbCompany struct {
	Name string `json:"name"`
}

type igdbInvolvedCompany struct {
	Company   igdbCompany `json:"company"`
	Developer bool        `json:"developer"`
	Publisher bool        `json:"publisher"`
}

type igdbGame struct {
	ID                int                   `json:"id"`
	Name              string                `json:"name"`
	Cover             *igdbCover            `json:"cover"`
	FirstReleaseDate  int64                 `json:"first_release_date"`
	InvolvedCompanies []igdbInvolvedCompany `json:"involved_companies"`
}

type CleanGame struct {
	Title    string
	Creator  string
	Year     int
	CoverURL string
}

func getIGDBToken() (string, error) {
	igdbTokenMu.Lock()
	defer igdbTokenMu.Unlock()

	if igdbToken != "" && time.Now().Before(igdbTokenExp) {
		return igdbToken, nil
	}

	clientID, err := GetKeys("TWITCH_CLIENT_ID")
	if err != nil {
		return "", fmt.Errorf("[getIGDBToken]%w", err)
	}

	clientSecret, err := GetKeys("TWITCH_CLIENT_SECRET")
	if err != nil {
		return "", fmt.Errorf("[getIGDBToken]%w", err)
	}

	endpoint := fmt.Sprintf("%s?client_id=%s&client_secret=%s&grant_type=client_credentials",
		twitchTokenURL, clientID, clientSecret)

	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("[getIGDBToken][http.NewRequest]: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("[getIGDBToken][client.Do]: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("[getIGDBToken][io.ReadAll]: %w", err)
	}

	var tokenResp twitchTokenResponse
	if err = json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("[getIGDBToken][json.Unmarshal]: %w", err)
	}

	igdbToken = tokenResp.AccessToken
	igdbTokenExp = time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second)

	return igdbToken, nil
}

func GetGameInfo(title string, year int) (CleanGame, error) {
	token, err := getIGDBToken()
	if err != nil {
		return CleanGame{}, fmt.Errorf("[GetGameInfo]%w", err)
	}

	clientID, err := GetKeys("TWITCH_CLIENT_ID")
	if err != nil {
		return CleanGame{}, fmt.Errorf("[GetGameInfo]%w", err)
	}

	startOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	startOfNextYear := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

	query := strings.TrimSpace(fmt.Sprintf(`
search "%s";
fields name,cover.image_id,involved_companies.company.name,involved_companies.developer,involved_companies.publisher,first_release_date;
where version_parent = null & first_release_date >= %d & first_release_date < %d;
limit 1;
`, title, startOfYear, startOfNextYear))

	req, err := http.NewRequest("POST", igdbBaseURL+"/games", strings.NewReader(query))
	if err != nil {
		return CleanGame{}, fmt.Errorf("[GetGameInfo][http.NewRequest]: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Client-ID", clientID)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return CleanGame{}, fmt.Errorf("[GetGameInfo][client.Do]: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CleanGame{}, fmt.Errorf("[GetGameInfo][io.ReadAll]: %w", err)
	}

	var games []igdbGame
	if err = json.Unmarshal(body, &games); err != nil {
		return CleanGame{}, fmt.Errorf("[GetGameInfo][json.Unmarshal]: %w", err)
	}

	if len(games) == 0 {
		return CleanGame{}, fmt.Errorf("[GetGameInfo]: no results for %q (%d)", title, year)
	}

	game := games[0]

	var publishers []string
	for _, ic := range game.InvolvedCompanies {
		if ic.Publisher {
			publishers = append(publishers, ic.Company.Name)
		}
	}

	coverURL := ""
	if game.Cover != nil && game.Cover.ImageID != "" {
		coverURL = fmt.Sprintf("%s/t_cover_big/%s.jpg", igdbImageBaseURL, game.Cover.ImageID)
	}

	releaseYear := year
	if game.FirstReleaseDate != 0 {
		releaseYear = time.Unix(game.FirstReleaseDate, 0).UTC().Year()
	}

	return CleanGame{
		Title:    game.Name,
		Creator:  strings.Join(publishers, ", "),
		Year:     releaseYear,
		CoverURL: coverURL,
	}, nil
}
