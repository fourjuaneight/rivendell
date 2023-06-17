package helpers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

type NameOrTitle struct {
	En string `json:"en"`
}

type TagsAttributes struct {
	Name        NameOrTitle `json:"name"`
	Description string      `json:"description,omitempty"`
	Group       string      `json:"group"`
	Version     int         `json:"version"`
}

type TagsEntity struct {
	ID            string         `json:"id"`
	Type          string         `json:"type"`
	Attributes    TagsAttributes `json:"attributes"`
	Relationships []interface{}  `json:"relationships,omitempty"`
}

type AltTitlesEntity struct {
	Ja string `json:"ja,omitempty"`
	Zh string `json:"zh,omitempty"`
	Ru string `json:"ru,omitempty"`
	Fa string `json:"fa,omitempty"`
}

type Description struct {
	En   string `json:"en"`
	Ru   string `json:"ru"`
	PtBr string `json:"pt-br"`
}

type Links struct {
	Al    string `json:"al"`
	Ap    string `json:"ap"`
	Bw    string `json:"bw"`
	Kt    string `json:"kt"`
	Mu    string `json:"mu"`
	Amz   string `json:"amz"`
	Cdj   string `json:"cdj"`
	Ebj   string `json:"ebj"`
	Mal   string `json:"mal"`
	Raw   string `json:"raw"`
	Engtl string `json:"engtl"`
}

type MangaAttributes struct {
	Title                          NameOrTitle       `json:"title"`
	AltTitles                      []AltTitlesEntity `json:"altTitles,omitempty"`
	Description                    Description       `json:"description"`
	IsLocked                       bool              `json:"isLocked"`
	Links                          Links             `json:"links"`
	OriginalLanguage               string            `json:"originalLanguage"`
	LastVolume                     string            `json:"lastVolume"`
	LastChapter                    string            `json:"lastChapter"`
	PublicationDemographic         string            `json:"publicationDemographic"`
	Status                         string            `json:"status"`
	Year                           int               `json:"year"`
	ContentRating                  string            `json:"contentRating"`
	Tags                           []TagsEntity      `json:"tags,omitempty"`
	State                          string            `json:"state"`
	ChapterNumbersResetOnNewVolume bool              `json:"chapterNumbersResetOnNewVolume"`
	CreatedAt                      string            `json:"createdAt"`
	UpdatedAt                      string            `json:"updatedAt"`
	Version                        int               `json:"version"`
	AvailableTranslatedLanguages   []string          `json:"availableTranslatedLanguages,omitempty"`
	LatestUploadedChapter          string            `json:"latestUploadedChapter"`
}

type RelationshipsAttributes struct {
	Description string `json:"description"`
	Volume      string `json:"volume"`
	FileName    string `json:"fileName"`
	Locale      string `json:"locale"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	Version     int    `json:"version"`
}

type RelationshipsEntity struct {
	ID         string                   `json:"id"`
	Type       string                   `json:"type"`
	Attributes *RelationshipsAttributes `json:"attributes,omitempty"`
}

type MangaData struct {
	ID            string                `json:"id"`
	Type          string                `json:"type"`
	Attributes    MangaAttributes       `json:"attributes"`
	Relationships []RelationshipsEntity `json:"relationships,omitempty"`
}

type MangaResponse struct {
	Result   string    `json:"result"`
	Response string    `json:"response"`
	Data     MangaData `json:"data"`
}

type AuthorAttributes struct {
	Name      string `json:"name"`
	ImageURL  string `json:"imageUrl,omitempty"`
	Biography string `json:"biography"`
	Twitter   string `json:"twitter,omitempty"`
	Pixiv     string `json:"pixiv,omitempty"`
	MelonBook string `json:"melonBook,omitempty"`
	FanBox    string `json:"fanBox,omitempty"`
	Booth     string `json:"booth,omitempty"`
	NicoVideo string `json:"nicoVideo,omitempty"`
	Skeb      string `json:"skeb,omitempty"`
	Fantia    string `json:"fantia,omitempty"`
	Tumblr    string `json:"tumblr,omitempty"`
	Youtube   string `json:"youtube,omitempty"`
	Weibo     string `json:"weibo,omitempty"`
	Naver     string `json:"naver,omitempty"`
	Website   string `json:"website,omitempty"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Version   int    `json:"version"`
}

type AuthorData struct {
	ID            string                `json:"id"`
	Type          string                `json:"type"`
	Attributes    AuthorAttributes      `json:"attributes"`
	Relationships []RelationshipsEntity `json:"relationships,omitempty"`
}

type AuthorResponse struct {
	Result   string     `json:"result"`
	Response string     `json:"response"`
	Data     AuthorData `json:"data"`
}

type CleanManga struct {
	Title       string
	Description string
	Author      string
	Year        int
	Status      string
	Cover       string
	Url         string
}

const (
	API    = "https://api.mangadex.org"
	ASSETS = "https://uploads.mangadex.org"
)

func parseMDURL(url string) (string, error) {
	regex, err := regexp.Compile(`https?:\/\/[^/]+\/title\/([a-f0-9-]+)\/?.*`)
	if err != nil {
		return "", fmt.Errorf("[parseMDURL][regexp.Compile]: %w", err)

	}

	id := regex.ReplaceAllString(url, "$1")

	return id, nil
}

func getAuthor(id string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/author/%s", API, id))
	if err != nil {
		return "", fmt.Errorf("[getAuthor]: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("[fetch]: %d - %s", resp.StatusCode, resp.Status)
		return "", fmt.Errorf("[getAuthor]%w", err)
	}

	var response AuthorResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("[getAuthor]: %w", err)
	}

	return response.Data.Attributes.Name, nil
}

func GetMangaInfo(url string) (CleanManga, error) {
	id, err := parseMDURL(url)
	if err != nil {
		return CleanManga{}, fmt.Errorf("[GetMangaInfo]%w", err)
	}

	resp, err := http.Get(fmt.Sprintf("%s/manga/%s?limit=100&includes%%5B%%5D=cover_art&includes%%5B%%5D=scanlation_group&order%%5Bvolume%%5D=desc&order%%5Bchapter%%5D=desc&offset=0&contentRating%%5B%%5D=safe&contentRating%%5B%%5D=suggestive&contentRating%%5B%%5D=erotica&contentRating%%5B%%5D=pornographic&translatedLanguage%%5B%%5D=en", API, id))
	if err != nil {
		return CleanManga{}, fmt.Errorf("[GetMangaInfo][http.Get]: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("[fetch]: %d - %s (%s)", resp.StatusCode, resp.Status, id)
		return CleanManga{}, fmt.Errorf("[GetMangaInfo]%w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CleanManga{}, fmt.Errorf("[GetMangaInfo][io.ReadAll]: %w", err)
	}

	var response MangaResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return CleanManga{}, fmt.Errorf("[GetMangaInfo][json.Unmarshal]: %w", err)
	}

	var coverFile string
	for _, rel := range response.Data.Relationships {
		if rel.Type == "cover_art" && rel.Attributes != nil {
			coverFile = rel.Attributes.FileName
			break
		}
	}

	author, err := getAuthor(response.Data.Relationships[0].ID)
	if err != nil {
		return CleanManga{}, fmt.Errorf("[GetMangaInfo]%w", err)
	}

	return CleanManga{
		Title:       response.Data.Attributes.Title.En,
		Description: response.Data.Attributes.Description.En,
		Author:      author,
		Year:        response.Data.Attributes.Year,
		Status:      response.Data.Attributes.Status,
		Cover:       fmt.Sprintf("%s/covers/%s/%s", ASSETS, id, coverFile),
		Url:         url,
	}, nil
}
