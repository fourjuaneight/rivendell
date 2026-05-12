package helpers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type openLibraryAuthor struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

type openLibraryIdentifiers struct {
	ISBN10 []string `json:"isbn_10"`
	ISBN13 []string `json:"isbn_13"`
}

type openLibraryData struct {
	Title       string                 `json:"title"`
	Authors     []openLibraryAuthor    `json:"authors"`
	PublishDate string                 `json:"publish_date"`
	Identifiers openLibraryIdentifiers `json:"identifiers"`
}

type openLibraryRecord struct {
	Data openLibraryData `json:"data"`
}

type openLibraryReadResponse struct {
	Records map[string]openLibraryRecord `json:"records"`
}

type CleanBook struct {
	Title    string
	Creator  string
	Year     int
	ISBN10   string
	ISBN13   string
	CoverURL string
}

func GetBookInfo(isbn string) (CleanBook, error) {
	clean := strings.NewReplacer("-", "", " ", "").Replace(isbn)
	endpoint := fmt.Sprintf("https://openlibrary.org/api/volumes/brief/isbn/%s.json", clean)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return CleanBook{}, fmt.Errorf("[GetBookInfo][http.NewRequest]: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return CleanBook{}, fmt.Errorf("[GetBookInfo][client.Do]: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CleanBook{}, fmt.Errorf("[GetBookInfo][io.ReadAll]: %w", err)
	}

	var result openLibraryReadResponse
	if err = json.Unmarshal(body, &result); err != nil {
		return CleanBook{}, fmt.Errorf("[GetBookInfo][json.Unmarshal]: %w", err)
	}

	var record openLibraryRecord
	for _, r := range result.Records {
		record = r
		break
	}

	if record.Data.Title == "" {
		return CleanBook{}, fmt.Errorf("[GetBookInfo]: no record found for ISBN %s", isbn)
	}

	yearRe := regexp.MustCompile(`\d{4}`)
	var year int
	if match := yearRe.FindString(record.Data.PublishDate); match != "" {
		fmt.Sscanf(match, "%d", &year)
	}

	var authorNames []string
	for _, a := range record.Data.Authors {
		authorNames = append(authorNames, a.Name)
	}

	isbn10 := ""
	if len(record.Data.Identifiers.ISBN10) > 0 {
		isbn10 = record.Data.Identifiers.ISBN10[0]
	}
	isbn13 := ""
	if len(record.Data.Identifiers.ISBN13) > 0 {
		isbn13 = record.Data.Identifiers.ISBN13[0]
	}

	coverISBN := isbn13
	if coverISBN == "" {
		coverISBN = isbn10
	}
	coverURL := ""
	if coverISBN != "" {
		coverURL = fmt.Sprintf("https://covers.openlibrary.org/b/isbn/%s-L.jpg", coverISBN)
	}

	return CleanBook{
		Title:    record.Data.Title,
		Creator:  strings.Join(authorNames, ", "),
		Year:     year,
		ISBN10:   isbn10,
		ISBN13:   isbn13,
		CoverURL: coverURL,
	}, nil
}
