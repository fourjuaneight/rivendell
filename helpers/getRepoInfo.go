package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

type Response struct {
	Data struct {
		Repository struct {
			Name  string `json:"name"`
			Owner struct {
				Login string `json:"login"`
			} `json:"owner"`
			Description     *string `json:"description"`
			PrimaryLanguage struct {
				Name *string `json:"name"`
			} `json:"primaryLanguage"`
		} `json:"repository"`
	} `json:"data"`
}

type CleanRepo struct {
	Name        string
	Owner       string
	Description string
	URL         string
	Language    string
}

func parseURL(url string) (string, string, error) {
	regex, err := regexp.Compile(`github\.com/([^/]+)/([^/]+)`)
	if err != nil {
		return "", "", fmt.Errorf("[parseURL][regexp.Compile]: %w", err)

	}

	matches := regex.FindStringSubmatch(url)
	if len(matches) == 3 {
		return matches[1], matches[2], nil
	}

	return "", "", fmt.Errorf("[parseURL]: No matches found%w", nil)

}

func GetRepoInfo(url string) (CleanRepo, error) {
	token, err := GetKeys("GH_TOKEN")
	if err != nil {
		return CleanRepo{}, fmt.Errorf("[GetRepoInfo][GetKeys](GH_TOKEN): %w", err)
	}

	owner, repo, err := parseURL(url)
	if err != nil {
		return CleanRepo{}, fmt.Errorf("[GetRepoInfo]%w", err)
	}

	query := fmt.Sprintf(`
		query {
			repository(owner: "%s", name: "%s") {
				name
				owner {
					login
				}
				description
				primaryLanguage {
					name
				}
			}
			}
		}
	`, owner, repo)

	options := struct {
		Query string `json:"query"`
	}{
		Query: query,
	}

	jsonData, err := json.Marshal(options)
	if err != nil {
		return CleanRepo{}, fmt.Errorf("[GetRepoInfo][json.Marshal]: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBuffer(jsonData))
	if err != nil {
		return CleanRepo{}, fmt.Errorf("[GetRepoInfo][http.NewRequest]: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return CleanRepo{}, fmt.Errorf("[GetRepoInfo][client.Do]: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CleanRepo{}, fmt.Errorf("[GetRepoInfo][io.ReadAll]: %w", err)
	}

	var results Response
	err = json.Unmarshal(body, &results)
	if err != nil {
		return CleanRepo{}, fmt.Errorf("[GetRepoInfo][json.Unmarshal]: %w", err)
	}

	description := ""
	if results.Data.Repository.Description != nil {
		description = *results.Data.Repository.Description
	}

	language := "Markdown"
	if results.Data.Repository.PrimaryLanguage.Name != nil {
		language = *results.Data.Repository.PrimaryLanguage.Name
	}

	cleanRepo := CleanRepo{
		Name:        results.Data.Repository.Name,
		Owner:       results.Data.Repository.Owner.Login,
		Description: description,
		URL:         url,
		Language:    language,
	}

	return cleanRepo, nil
}
