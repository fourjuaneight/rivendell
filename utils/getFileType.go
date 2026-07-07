package utils

import (
	"fmt"
	"regexp"
)

type FileTypes struct {
	File string
	MIME string
}

// Compiled once at package init rather than per GetFileType call.
var (
	imgMatch = regexp.MustCompile(`(?i)^.*(png|jpg|jpeg|webp|gif|gifv)$`)
	vidMatch = regexp.MustCompile(`(?i)^.*(mp4|mov)$`)
)

func GetFileType(typeStr string, url string) FileTypes {
	isImg := imgMatch.MatchString(url)
	isVid := vidMatch.MatchString(url)

	var mediaType string

	switch {
	case isImg:
		mediaType = imgMatch.ReplaceAllString(url, "$1")
	case isVid:
		mediaType = vidMatch.ReplaceAllString(url, "$1")
	}

	fileType := map[string]FileTypes{
		"articles": {File: "md", MIME: "text/markdown"},
		"comics":   {File: mediaType, MIME: fmt.Sprintf("image/%s", mediaType)},
		"podcasts": {File: "mp3", MIME: "audio/mpeg"},
		"videos":   {File: "mp4", MIME: "video/mp4"},
	}

	return fileType[typeStr]
}
