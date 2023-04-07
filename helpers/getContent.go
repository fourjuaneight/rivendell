package helpers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/fourjuaneight/rivendell/utils"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/proto"
	readability "github.com/go-shiori/go-readability"

	"golang.org/x/net/html"
)

// Get Markdown version of article from url.
func GetArticle(name string, urlString string) []byte {
	browser := rod.New().MustConnect().NoDefaultDevice()

	// visit url
	page := browser.MustPage(urlString).MustEmulate(devices.LaptopWithHiDPIScreen)
	page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:  1200,
		Height: 630,
	})

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
		page.MustElement(selector).MustRemove()
	}

	// get html
	htmlString, err := page.HTML()
	if err != nil {
		log.Fatal("[GetArticle][page.HTML] %w", err)
	}

	// get html node
	htmlNode, err := html.Parse(strings.NewReader(htmlString))
	if err != nil {
		log.Fatal("[GetArticle][html.Parse] %w", err)
	}

	// get url object
	pageURL, err := url.Parse(urlString)
	if err != nil {
		log.Fatal("[GetArticle][url.Parse] %w", err)
	}

	// get article and convert to markdown
	article, err := readability.FromDocument(htmlNode, pageURL)
	if err != nil {
		log.Fatal("[GetArticle][readability.FromReader] %w", err)
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

	finalMD := fmt.Sprintf("# %s\n\n%s", name, markdown)

	return []byte(finalMD)
}

// Get media file from source URL.
func GetMedia(name string, url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("[GetMedia][http.Get]: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("[GetMedia][io.ReadAll]: %w", err)
	}

	return body
}

// Get YouTube file from url.
func GetYTVid(name string, url string) []byte {
	fileName := utils.FileNameFmt(name)
	filePath := fileName + ".mp4"

	// download video with the ytdl function
	utils.YTDL(url, filePath)

	// read downloaded file into buffer
	buffer, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal("[GetYTVid][os.ReadFile]: %w", err)
	}

	utils.DeleteFiles([]string{filePath})

	return buffer
}

func GetContent(name string, url string, mediaType string) []byte {
	switch mediaType {
	case "articles":
		return GetArticle(name, url)
	case "videos":
		return GetYTVid(name, url)
	default:
		return GetMedia(name, url)
	}
}
