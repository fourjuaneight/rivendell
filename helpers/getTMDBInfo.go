package helpers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
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
	Title   string
	Creator string
	Genre   string
	Year    string
	Type    string
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

	endpoint := fmt.Sprintf("https://api.themoviedb.org/3/%s/%s/credits", category, id)

	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("[getCredits][http.NewRequest]: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

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
		return "", fmt.Errorf("[GetRepoInfo][json.Unmarshal]: %w", err)
	}

	var creators []string
	for _, crew := range results.Crew {
		if crew.Job == "Director" {
			creators = append(creators, crew.Name)
		}
	}

	return strings.Join(creators, ", "), nil
}

func GetTMDBInfo(url string) (CleanMedia, error) {
	token, err := GetKeys("TMDB_KEY")
	if err != nil {
		return CleanMedia{}, fmt.Errorf("[GetTMDBInfo]%w", err)
	}

	data, err := parseTMDBURL(url)
	if err != nil {
		return CleanMedia{}, fmt.Errorf("[GetTMDBInfo]%w", err)
	}

	endpoint := fmt.Sprintf("https://api.themoviedb.org/3/%s/%s", data.category, data.id)

	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		return CleanMedia{}, fmt.Errorf("[GetTMDBInfo][http.NewRequest]: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return CleanMedia{}, fmt.Errorf("[GetTMDBInfo][client.Do]: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CleanMedia{}, fmt.Errorf("[GetTMDBInfo][io.ReadAll]: %w", err)
	}

	creator, err := getDirector(data.category, data.id)
	if err != nil {
		return CleanMedia{}, fmt.Errorf("[GetTMDBInfo]%w", err)
	}

	if data.category == "movie" {
		var movie Movie
		err = json.Unmarshal(body, &movie)
		if err != nil {
			return CleanMedia{}, fmt.Errorf("[GetTMDBInfo][json.Unmarshal]: %w", err)
		}

		year := movie.ReleaseDate[:4]

		return CleanMedia{
			Title:   movie.Title,
			Creator: creator,
			Genre:   "",
			Year:    year,
			Type:    "movie",
		}, nil
	}

	var tv TVShow
	err = json.Unmarshal(body, &tv)
	if err != nil {
		return CleanMedia{}, fmt.Errorf("[GetTMDBInfo][json.Unmarshal]: %w", err)
	}

	year := tv.FirstAirDate[:4]

	return CleanMedia{
		Title:   tv.Name,
		Creator: creator,
		Genre:   "",
		Year:    year,
		Type:    "tv",
	}, nil
}
