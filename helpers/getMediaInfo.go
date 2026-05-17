package helpers

import (
	"encoding/json"
	"fmt"
	"io"

	"net/http"
	neturl "net/url"
	"regexp"
	"strings"

	"github.com/sahilm/fuzzy"
)

type Movie struct {
	Adult               bool   `json:"adult"`
	BackdropPath        string `json:"backdrop_path"`
	BelongsToCollection struct {
		ID           int    `json:"id"`
		Name         string `json:"name"`
		PosterPath   string `json:"poster_path"`
		BackdropPath string `json:"backdrop_path"`
	} `json:"belongs_to_collection"`
	Budget int `json:"budget"`
	Genres []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"genres"`
	Homepage            string  `json:"homepage"`
	ID                  int     `json:"id"`
	ImdbID              string  `json:"imdb_id"`
	OriginalLanguage    string  `json:"original_language"`
	OriginalTitle       string  `json:"original_title"`
	Overview            string  `json:"overview"`
	Popularity          float64 `json:"popularity"`
	PosterPath          string  `json:"poster_path"`
	ProductionCompanies []struct {
		ID            int    `json:"id"`
		LogoPath      string `json:"logo_path"`
		Name          string `json:"name"`
		OriginCountry string `json:"origin_country"`
	} `json:"production_companies"`
	ProductionCountries []struct {
		Iso31661 string `json:"iso_3166_1"`
		Name     string `json:"name"`
	} `json:"production_countries"`
	ReleaseDate     string `json:"release_date"`
	Revenue         int    `json:"revenue"`
	Runtime         int    `json:"runtime"`
	SpokenLanguages []struct {
		EnglishName string `json:"english_name"`
		Iso6391     string `json:"iso_639_1"`
		Name        string `json:"name"`
	} `json:"spoken_languages"`
	Status      string  `json:"status"`
	Tagline     string  `json:"tagline"`
	Title       string  `json:"title"`
	Video       bool    `json:"video"`
	VoteAverage float64 `json:"vote_average"`
	VoteCount   int     `json:"vote_count"`
}

type TVShow struct {
	Adult        bool   `json:"adult"`
	BackdropPath string `json:"backdrop_path"`
	CreatedBy    []struct {
		ID          int    `json:"id"`
		CreditID    string `json:"credit_id"`
		Name        string `json:"name"`
		Gender      int    `json:"gender"`
		ProfilePath string `json:"profile_path"`
	} `json:"created_by"`
	EpisodeRunTime []int  `json:"episode_run_time"`
	FirstAirDate   string `json:"first_air_date"`
	Genres         []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"genres"`
	Homepage         string   `json:"homepage"`
	ID               int      `json:"id"`
	InProduction     bool     `json:"in_production"`
	Languages        []string `json:"languages"`
	LastAirDate      string   `json:"last_air_date"`
	LastEpisodeToAir struct {
		ID             int     `json:"id"`
		Name           string  `json:"name"`
		Overview       string  `json:"overview"`
		VoteAverage    float64 `json:"vote_average"`
		VoteCount      int     `json:"vote_count"`
		AirDate        string  `json:"air_date"`
		EpisodeNumber  int     `json:"episode_number"`
		ProductionCode string  `json:"production_code"`
		Runtime        int     `json:"runtime"`
		SeasonNumber   int     `json:"season_number"`
		ShowID         int     `json:"show_id"`
		StillPath      string  `json:"still_path"`
	} `json:"last_episode_to_air"`
	Name     string `json:"name"`
	Networks []struct {
		ID            int    `json:"id"`
		LogoPath      string `json:"logo_path"`
		Name          string `json:"name"`
		OriginCountry string `json:"origin_country"`
	} `json:"networks"`
	NumberOfEpisodes    int      `json:"number_of_episodes"`
	NumberOfSeasons     int      `json:"number_of_seasons"`
	OriginCountry       []string `json:"origin_country"`
	OriginalLanguage    string   `json:"original_language"`
	OriginalName        string   `json:"original_name"`
	Overview            string   `json:"overview"`
	Popularity          float64  `json:"popularity"`
	PosterPath          string   `json:"poster_path"`
	ProductionCompanies []struct {
		ID            int    `json:"id"`
		LogoPath      string `json:"logo_path"`
		Name          string `json:"name"`
		OriginCountry string `json:"origin_country"`
	} `json:"production_companies"`
	ProductionCountries []struct {
		ISO31661 string `json:"iso_3166_1"`
		Name     string `json:"name"`
	} `json:"production_countries"`
	Seasons []struct {
		AirDate      string `json:"air_date"`
		EpisodeCount int    `json:"episode_count"`
		ID           int    `json:"id"`
		Name         string `json:"name"`
		Overview     string `json:"overview"`
		PosterPath   string `json:"poster_path"`
		SeasonNumber int    `json:"season_number"`
	} `json:"seasons"`
	SpokenLanguages []struct {
		EnglishName string `json:"english_name"`
		ISO6391     string `json:"iso_639_1"`
		Name        string `json:"name"`
	} `json:"spoken_languages"`
	Status      string  `json:"status"`
	Tagline     string  `json:"tagline"`
	Type        string  `json:"type"`
	VoteAverage float64 `json:"vote_average"`
	VoteCount   int     `json:"vote_count"`
}

type Credits struct {
	ID   int `json:"id"`
	Cast []struct {
		Adult              bool    `json:"adult"`
		Gender             int     `json:"gender"`
		ID                 int     `json:"id"`
		KnownForDepartment string  `json:"known_for_department"`
		Name               string  `json:"name"`
		OriginalName       string  `json:"original_name"`
		Popularity         float64 `json:"popularity"`
		ProfilePath        string  `json:"profile_path"`
		CastID             int     `json:"cast_id"`
		Character          string  `json:"character"`
		CreditID           string  `json:"credit_id"`
		Order              int     `json:"order"`
	} `json:"cast"`
	Crew []struct {
		Adult              bool    `json:"adult"`
		Gender             int     `json:"gender"`
		ID                 int     `json:"id"`
		KnownForDepartment string  `json:"known_for_department"`
		Name               string  `json:"name"`
		OriginalName       string  `json:"original_name"`
		Popularity         float64 `json:"popularity"`
		ProfilePath        *string `json:"profile_path"`
		CreditID           string  `json:"credit_id"`
		Department         string  `json:"department"`
		Job                string  `json:"job"`
	} `json:"crew"`
}

type TypeData struct {
	id       string
	category string
}

type CleanMedia struct {
	Title    string
	Creator  string
	Genre    string
	Year     string
	Type     string
	CoverURL string
}

type searchResult struct {
	Results []struct {
		ID    int    `json:"id"`
		Title string `json:"title"` // movies
		Name  string `json:"name"`  // tv
	} `json:"results"`
}

type seasonImages struct {
	Posters []struct {
		FilePath string `json:"file_path"`
	} `json:"posters"`
}

func parseTMDBURL(url string) (TypeData, error) {
	regex, err := regexp.Compile(`https?:\/\/[^/]+\/(movie|tv)\/([0-9]+)-?.*`)
	if err != nil {
		return TypeData{}, fmt.Errorf("[parseTMDBURL][regexp.Compile]: %w", err)

	}

	id := regex.ReplaceAllString(url, "$2")
	category := regex.ReplaceAllString(url, "$1")

	return TypeData{
		id:       id,
		category: category,
	}, nil
}

func getDirector(category string, id string) (string, error) {
	token, err := GetKeys("TMDB_KEY")
	if err != nil {
		return "", fmt.Errorf("[getCredits]%w", err)
	}

	// DOCS: https://developer.themoviedb.org/reference/movie-credits (movie)
	//       https://developer.themoviedb.org/reference/tv-series-credits (tv)
	endpoint := fmt.Sprintf("https://api.themoviedb.org/3/%s/%s/credits?api_key=%s", category, id, token)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("[getCredits][http.NewRequest]: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("[getCredits][client.Do]: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("[getCredits][io.ReadAll]: %w", err)
	}

	var results Credits
	err = json.Unmarshal(body, &results)
	if err != nil {
		return "", fmt.Errorf("[getCredits][json.Unmarshal]: %w", err)
	}

	var creators []string
	for _, crew := range results.Crew {
		if crew.Job == "Director" {
			creators = append(creators, crew.Name)
		}
	}

	return strings.Join(creators, ", "), nil
}

func GetMediaInfo(url string) (CleanMedia, error) {
	token, err := GetKeys("TMDB_KEY")
	if err != nil {
		return CleanMedia{}, fmt.Errorf("[GetMediaInfo]%w", err)
	}

	data, err := parseTMDBURL(url)
	if err != nil {
		return CleanMedia{}, fmt.Errorf("[GetMediaInfo]%w", err)
	}

	// DOCS: https://developer.themoviedb.org/reference/movie-details (movie)
	//       https://developer.themoviedb.org/reference/tv-series-details (tv)
	endpoint := fmt.Sprintf("https://api.themoviedb.org/3/%s/%s?api_key=%s", data.category, data.id, token)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return CleanMedia{}, fmt.Errorf("[GetMediaInfo][http.NewRequest]: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return CleanMedia{}, fmt.Errorf("[GetMediaInfo][client.Do]: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CleanMedia{}, fmt.Errorf("[GetMediaInfo][io.ReadAll]: %w", err)
	}

	creator, err := getDirector(data.category, data.id)
	if err != nil {
		return CleanMedia{}, fmt.Errorf("[GetMediaInfo]: %w", err)
	}

	if data.category == "movie" {
		var movie Movie
		err = json.Unmarshal(body, &movie)
		if err != nil {
			return CleanMedia{}, fmt.Errorf("[GetMediaInfo][json.Unmarshal]: %w", err)
		}

		year := movie.ReleaseDate[:4]

		coverURL := ""
		if movie.PosterPath != "" {
			coverURL = "https://image.tmdb.org/t/p/original" + movie.PosterPath
		}

		return CleanMedia{
			Title:    movie.Title,
			Creator:  creator,
			Genre:    "",
			Year:     year,
			Type:     "movie",
			CoverURL: coverURL,
		}, nil
	}

	var tv TVShow
	err = json.Unmarshal(body, &tv)
	if err != nil {
		return CleanMedia{}, fmt.Errorf("[GetMediaInfo][json.Unmarshal]: %w", err)
	}

	year := tv.FirstAirDate[:4]

	coverURL := ""
	if tv.PosterPath != "" {
		coverURL = "https://image.tmdb.org/t/p/original" + tv.PosterPath
	}

	return CleanMedia{
		Title:    tv.Name,
		Creator:  creator,
		Genre:    "",
		Year:     year,
		Type:     "tv",
		CoverURL: coverURL,
	}, nil
}

func tmdbGet(token, endpoint string) ([]byte, error) {
	sep := "?"
	if strings.Contains(endpoint, "?") {
		sep = "&"
	}
	url := endpoint + sep + "api_key=" + token

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("[tmdbGet][http.NewRequest]: %w", err)
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("[tmdbGet][client.Do]: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[tmdbGet][io.ReadAll]: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[tmdbGet]: %s", resp.Status)
	}

	return body, nil
}

// SearchMedia searches TMDB by title/year and returns a full CleanMedia for the best result.
// For shows, fetches the season-specific poster; season=0 means the whole series, uses season 1.
// Used by the create hook. GetMediaInfo is available for URL-based full detail lookups.
func SearchMedia(title string, year int, season int, mediaType string) (CleanMedia, error) {
	token, err := GetKeys("TMDB_KEY")
	if err != nil {
		return CleanMedia{}, fmt.Errorf("[SearchMedia]: [GetKeys]: %w", err)
	}

	category := "movie"
	if mediaType == "shows" {
		category = "tv"
	}

	// DOCS: https://developer.themoviedb.org/reference/search-movie (movies)
	//       https://developer.themoviedb.org/reference/search-tv (shows)
	searchEndpoint := fmt.Sprintf("https://api.themoviedb.org/3/search/%s?query=%s", category, neturl.QueryEscape(title))
	if year != 0 {
		searchEndpoint += fmt.Sprintf("&year=%d", year)
	}

	searchBody, err := tmdbGet(token, searchEndpoint)
	if err != nil {
		return CleanMedia{}, fmt.Errorf("[SearchMedia]: %w", err)
	}

	var results searchResult
	if err = json.Unmarshal(searchBody, &results); err != nil {
		return CleanMedia{}, fmt.Errorf("[SearchMedia][json.Unmarshal]: %w", err)
	}

	if len(results.Results) == 0 {
		return CleanMedia{}, fmt.Errorf("[SearchMedia]: no results for %q (%d)", title, year)
	}

	// Fuzzy-match the input title against all returned titles to handle
	// minor differences between stored and TMDB titles.
	titles := make([]string, len(results.Results))
	for i, r := range results.Results {
		if category == "movie" {
			titles[i] = r.Title
		} else {
			titles[i] = r.Name
		}
	}
	bestIdx := 0
	if matches := fuzzy.Find(title, titles); len(matches) > 0 {
		bestIdx = matches[0].Index
	}

	// DOCS: https://developer.themoviedb.org/reference/movie-details (movie)
	//       https://developer.themoviedb.org/reference/tv-series-details (tv)
	detailEndpoint := fmt.Sprintf("https://api.themoviedb.org/3/%s/%d", category, results.Results[bestIdx].ID)
	detailBody, err := tmdbGet(token, detailEndpoint)
	if err != nil {
		return CleanMedia{}, fmt.Errorf("[SearchMedia]: %w", err)
	}

	if category == "movie" {
		var movie Movie
		if err = json.Unmarshal(detailBody, &movie); err != nil {
			return CleanMedia{}, fmt.Errorf("[SearchMedia][json.Unmarshal movie]: %w", err)
		}
		genre := ""
		if len(movie.Genres) > 0 {
			genre = movie.Genres[0].Name
		}
		coverURL := ""
		if movie.PosterPath != "" {
			coverURL = "https://image.tmdb.org/t/p/original" + movie.PosterPath
		}
		releaseYear := year
		if len(movie.ReleaseDate) >= 4 {
			releaseYear = 0
			for _, c := range movie.ReleaseDate[:4] {
				releaseYear = releaseYear*10 + int(c-'0')
			}
		}
		return CleanMedia{
			Title:    movie.Title,
			Genre:    genre,
			Year:     fmt.Sprintf("%d", releaseYear),
			Type:     "movies",
			CoverURL: coverURL,
		}, nil
	}

	var tv TVShow
	if err = json.Unmarshal(detailBody, &tv); err != nil {
		return CleanMedia{}, fmt.Errorf("[SearchMedia][json.Unmarshal tv]: %w", err)
	}
	genre := ""
	if len(tv.Genres) > 0 {
		genre = tv.Genres[0].Name
	}
	coverURL := ""
	if tv.PosterPath != "" {
		coverURL = "https://image.tmdb.org/t/p/original" + tv.PosterPath
	}
	// DOCS: https://developer.themoviedb.org/reference/tv-season-images
	// season == 0 means whole series — fall back to season 1 poster.
	seasonNum := season
	if seasonNum == 0 {
		seasonNum = 1
	}
	{
		imgEndpoint := fmt.Sprintf("https://api.themoviedb.org/3/tv/%d/season/%d/images", tv.ID, seasonNum)
		imgBody, imgErr := tmdbGet(token, imgEndpoint)
		if imgErr == nil {
			var imgs seasonImages
			if jsonErr := json.Unmarshal(imgBody, &imgs); jsonErr == nil && len(imgs.Posters) > 0 {
				coverURL = "https://image.tmdb.org/t/p/original" + imgs.Posters[0].FilePath
			}
		}
	}
	releaseYear := year
	if len(tv.FirstAirDate) >= 4 {
		releaseYear = 0
		for _, c := range tv.FirstAirDate[:4] {
			releaseYear = releaseYear*10 + int(c-'0')
		}
	}
	return CleanMedia{
		Title:    tv.Name,
		Genre:    genre,
		Year:     fmt.Sprintf("%d", releaseYear),
		Type:     "shows",
		CoverURL: coverURL,
	}, nil
}
