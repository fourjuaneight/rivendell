package helpers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/fourjuaneight/rivendell/utils"

	query "github.com/PuerkitoBio/goquery"
	readability "github.com/go-shiori/go-readability"

	"golang.org/x/net/html"
)

// Get Markdown version of article from url.
func GetArticle(name string, urlString string) ([]byte, error) {
	// get html from url
	resp, err := HTTPClient.Get(urlString)
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
	resp, err := MediaClient.Get(url)
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

// tagMediaMetadata writes record metadata into a media file (MP4 video or MP3 audio)
// via ffmpeg. Uses stream copy (no re-encode), so it's fast. Mirrors tag_videos.py:
// the creator maps to both artist and album_artist, and genre is caller-supplied
// ("YouTube" for videos, "Podcast" for podcasts). Empty/zero fields are skipped.
// ffmpeg writes the correct tag format (iTunes atoms for MP4, ID3 frames for MP3)
// based on the output container, inferred from the file extension.
func tagMediaMetadata(path string, title string, creator string, year int, genre string) error {
	taggedPath := path + ".tagged" + filepath.Ext(path)

	args := []string{"-y", "-i", path, "-map", "0", "-c", "copy"}
	if title != "" {
		args = append(args, "-metadata", "title="+title)
	}
	if creator != "" {
		args = append(args, "-metadata", "artist="+creator, "-metadata", "album_artist="+creator)
	}
	if year != 0 {
		args = append(args, "-metadata", "date="+strconv.Itoa(year))
	}
	args = append(args, "-metadata", "genre="+genre, taggedPath)

	if err := utils.CMD("ffmpeg", args...); err != nil {
		os.Remove(taggedPath) // best-effort cleanup of partial output
		return fmt.Errorf("[tagMediaMetadata][ffmpeg]: %w", err)
	}

	// Replace the original download with the tagged copy.
	if err := os.Rename(taggedPath, path); err != nil {
		os.Remove(taggedPath)
		return fmt.Errorf("[tagMediaMetadata][os.Rename]: %w", err)
	}

	return nil
}

// Get YouTube file from url.
func GetYTVid(name string, url string, creator string, year int) ([]byte, error) {
	fileName := utils.FileNameFmt(name)
	filePath := fileName + ".mp4"

	// download video with the ytdl function
	ytdlErr := utils.YTDL(url, filePath)
	if ytdlErr != nil {
		return nil, fmt.Errorf("[GetYTVid][YTDL]: %w", ytdlErr)
	}

	// embed record metadata (title, creator, year, genre) into the MP4
	if err := tagMediaMetadata(filePath, name, creator, year, "YouTube"); err != nil {
		os.Remove(filePath) // best-effort cleanup of the untagged download
		return nil, fmt.Errorf("[GetYTVid][tagMediaMetadata]: %w", err)
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

// GetPodcast downloads a podcast MP3 from url, embeds record metadata, and returns
// the tagged bytes. ffmpeg needs the file on disk, so unlike GetMedia this stages the
// download to a temp file before tagging. Genre is fixed to "Podcast".
func GetPodcast(name string, url string, creator string, year int) ([]byte, error) {
	// reuse GetMedia to fetch the bytes, then stage to disk for ffmpeg
	data, err := GetMedia(name, url)
	if err != nil {
		return nil, fmt.Errorf("[GetPodcast]: %w", err)
	}

	filePath := utils.FileNameFmt(name) + ".mp3"
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return nil, fmt.Errorf("[GetPodcast][os.WriteFile]: %w", err)
	}

	// embed record metadata (title, creator, year, genre) into the MP3
	if err := tagMediaMetadata(filePath, name, creator, year, "Podcast"); err != nil {
		os.Remove(filePath) // best-effort cleanup of the untagged download
		return nil, fmt.Errorf("[GetPodcast][tagMediaMetadata]: %w", err)
	}

	media, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("[GetPodcast][os.ReadFile]: %w", err)
	}

	dfErr := utils.DeleteFiles([]string{filePath})
	if dfErr != nil {
		return nil, fmt.Errorf("[GetPodcast][DeleteFiles]: %w", dfErr)
	}

	return media, nil
}

// GetSingleFile captures a full-page HTML snapshot via single-file-cli and returns the bytes.
// Requires chromium and single-file-cli installed in the runtime environment.
//
// Note: single-file hardcodes "--single-process" when launching chromium, which crashes the
// renderer on chromium >= ~131 so the remote-debugging port never opens. The Dockerfile strips
// that flag from single-file's browser.js. We pass only "--no-sandbox" (the container runs as
// root) and let single-file manage headless mode itself.
func GetSingleFile(urlString string) ([]byte, error) {
	cmd := exec.Command("single-file",
		"--browser-executable-path=/usr/bin/chromium-browser",
		`--browser-args=["--no-sandbox"]`,
		"--dump-content",
		urlString,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("[GetSingleFile][cmd.Run]: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	// single-file exits 0 even when the capture fails (e.g. it logs "fetch failed" to stderr
	// and emits nothing). Treat empty output as an error so callers never archive an empty file.
	output := stdout.Bytes()
	if len(output) == 0 {
		return nil, fmt.Errorf("[GetSingleFile]: empty output: %s", strings.TrimSpace(stderr.String()))
	}

	return output, nil
}

func GetContent(name string, url string, mediaType string, creator string, year int) ([]byte, error) {
	switch mediaType {
	case "articles":
		return GetArticle(name, url)
	case "videos":
		return GetYTVid(name, url, creator, year)
	case "podcasts":
		return GetPodcast(name, url, creator, year)
	default:
		return GetMedia(name, url)
	}
}
