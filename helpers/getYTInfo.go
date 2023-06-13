package helpers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/fourjuaneight/rivendell/utils"
)

type YouTubeAPIEndpoint struct {
	Endpoint string
	Link     string
}

type YouTubeResponse struct {
	Kind  string `json:"kind"`
	Etag  string `json:"etag"`
	Items []struct {
		Kind    string `json:"kind"`
		Etag    string `json:"etag"`
		ID      string `json:"id"`
		Snippet struct {
			PublishedAt string `json:"publishedAt"`
			ChannelID   string `json:"channelId"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Thumbnails  struct {
				Default struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"default"`
				Medium struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"medium"`
				High struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"high"`
				Standard struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"standard"`
				Maxres struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"maxres"`
			} `json:"thumbnails"`
			ChannelTitle         string   `json:"channelTitle"`
			Tags                 []string `json:"tags"`
			CategoryID           string   `json:"categoryId"`
			LiveBroadcastContent string   `json:"liveBroadcastContent"`
			DefaultLanguage      string   `json:"defaultLanguage"`
			Localized            struct {
				Title       string `json:"title"`
				Description string `json:"description"`
			} `json:"localized"`
			DefaultAudioLanguage string `json:"defaultAudioLanguage"`
		} `json:"snippet"`
	} `json:"items"`
	PageInfo struct {
		TotalResults   int `json:"totalResults"`
		ResultsPerPage int `json:"resultsPerPage"`
	} `json:"pageInfo"`
}

type CleanYT struct {
	Title   string
	Creator string
	URL     string
	Tags    []string
}

func cleanYTURL(url string) YouTubeAPIEndpoint {
	re := regexp.MustCompile(`(https:\/\/)(www\.)?(youtu.*)\.(be|com)\/(watch\?v=)?`)
	extractedID := re.ReplaceAllString(url, "")
	extractedID = strings.ReplaceAll(extractedID, "&feature=share", "")
	endpoint := fmt.Sprintf("https://youtube.googleapis.com/youtube/v3/videos?part=snippet&id=%s", extractedID)
	link := fmt.Sprintf("https://youtu.be/%s", extractedID)
	data := YouTubeAPIEndpoint{Endpoint: endpoint, Link: link}

	return data
}

func GetYTInfo(url string) (CleanYT, error) {
	key, err := GetKeys("YOUTUBE_KEY")
	if err != nil {
		return CleanYT{}, fmt.Errorf("[GetYTInfo]%w", err)
	}

	urls := cleanYTURL(url)

	resp, err := http.Get(fmt.Sprintf("%s&key=%s", urls.Endpoint, key))
	if err != nil {
		return CleanYT{}, fmt.Errorf("[GetYTInfo][http.Get]: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("[fetch]: %d - %s (%s)", resp.StatusCode, resp.Status, urls.Link)
		return CleanYT{}, fmt.Errorf("[GetYTInfo]%w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CleanYT{}, fmt.Errorf("[GetYTInfo][io.ReadAll]: %w", err)
	}

	var response YouTubeResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return CleanYT{}, fmt.Errorf("[GetYTInfo][json.Unmarshal]: %w", err)
	}

	video := response.Items[0].Snippet

	return CleanYT{
		Title:   utils.FileNameFmt(video.Title),
		Creator: utils.FileNameFmt(video.ChannelTitle),
		URL:     urls.Link,
	}, nil
}
