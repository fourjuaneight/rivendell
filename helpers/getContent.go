package helpers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/fourjuaneight/rivendell/utils"

	query "github.com/PuerkitoBio/goquery"
	readability "github.com/go-shiori/go-readability"

	"golang.org/x/net/html"
)

// Get Markdown version of article from url.
func GetArticle(name string, urlString string) ([]byte, error) {
	// get html from url
	resp, err := http.Get(urlString)
	if err != nil {
		return nil, fmt.Errorf("[GetArticle][http.Get] %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		mgs := fmt.Sprintf("%d - %s", resp.StatusCode, resp.Status)

		return nil, fmt.Errorf("[GetArticle][resp] %s", mgs)
	}

	// parse html
	doc, err := query.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[GetArticle][query.NewDocumentFromReader] %w", err)
	}

	// remove annoyances
	cleanDoc := []string{
		// WIRED
		"div.newsletter-subscribe-form",
		"div[class^='RecircMostPopularContiner']",
		"div[data-attr-viewport-monitor]",
		"div[class^='NewsletterSubscribeFormWrapper']",
		"div[data-testid='NewsletterSubscribeFormWrapper']",
		"div[class^='GenericCalloutWrapper']",
		"div[data-testid='GenericCallout']",
		"aside[class^='Sidebar']",
		"aside[data-testid='SidebarEmbed']",
		"div[class^='ContributorsWrapper']",
		"div[data-testid='Contributors']",
		// The Atlantic
		"p[class^='ArticleRelatedContentLink']",
		"div[class^='ArticleRelatedContentModule']",
		"div[class^='ArticleBooksModule']",
		// Ars Technica
		"div.gallery",
		"div.story-sidebar",
		// Media
		"img",
		"picture",
		"figure",
		"video",
		"iframe",
	}
	for _, selector := range cleanDoc {
		doc.Find(selector).Each(func(i int, s *query.Selection) {
			s.Remove()
		})
	}

	// get html
	htmlString, err := doc.Html()
	if err != nil {
		return nil, fmt.Errorf("[GetArticle][doc.Html] %w", err)
	}

	// get html node
	htmlNode, err := html.Parse(strings.NewReader(htmlString))
	if err != nil {
		return nil, fmt.Errorf("[GetArticle][html.Parse] %w", err)
	}

	// get url object
	pageURL, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("[GetArticle][url.Parse] %w", err)
	}

	// get article and convert to markdown
	article, err := readability.FromDocument(htmlNode, pageURL)
	if err != nil {
		return nil, fmt.Errorf("[GetArticle][readability.FromReader] %w", err)
	}
	markdown := article.Content

	// clean markdown
	re1 := regexp.MustCompile(`([‘’]+)`)
	re2 := regexp.MustCompile(`([“”]+)`)
	markdown = re1.ReplaceAllString(markdown, `'`)
	markdown = re2.ReplaceAllString(markdown, `"`)

	if strings.Contains(urlString, "wired") {
		re3 := regexp.MustCompile(`([—]+)`)
		markdown = re3.ReplaceAllString(markdown, "")
	}

	media := fmt.Sprintf("# %s\n\n%s", name, markdown)

	return []byte(media), nil
}

// Get media file from source URL.
func GetMedia(name string, url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("[GetMedia][http.Get]: %w", err)
	}

	defer resp.Body.Close()

	media, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[GetMedia][io.ReadAll]: %w", err)
	}

	return media, nil
}

// Get YouTube file from url.
func GetYTVid(name string, url string) ([]byte, error) {
	fileName := utils.FileNameFmt(name)
	filePath := fileName + ".mp4"

	// download video with the ytdl function
	ytdlErr := utils.YTDL(url, filePath)
	if ytdlErr != nil {
		return nil, fmt.Errorf("[GetYTVid][YTDL]: %w", ytdlErr)
	}

	// read downloaded file into buffer
	media, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("[GetYTVid][os.ReadFile]: %w", err)
	}

	dfErr := utils.DeleteFiles([]string{filePath})
	if dfErr != nil {
		return nil, fmt.Errorf("[GetYTVid][DeleteFiles]: %w", dfErr)
	}

	return media, nil
}

func GetContent(name string, url string, mediaType string) ([]byte, error) {
	switch mediaType {
	case "articles":
		return GetArticle(name, url)
	case "videos":
		return GetYTVid(name, url)
	default:
		return GetMedia(name, url)
	}
}
