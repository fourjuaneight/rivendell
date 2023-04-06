package main

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

func fileNameFmt(name string) string {
	cleanName := emojiRange.ReplaceAllString(name, "")
	cleanName = strings.Trim(cleanName, " \t\n\r\v\f")
	cleanName = strings.TrimSuffix(cleanName, ".")
	cleanName = strings.TrimSuffix(cleanName, "?")
	cleanName = strings.TrimSuffix(cleanName, "!")
	cleanName = strings.ReplaceAll(cleanName, ". ", "-")
	cleanName = strings.ReplaceAll(cleanName, ", ", "-")
	cleanName = strings.ReplaceAll(cleanName, " :: ", "-")
	cleanName = strings.ReplaceAll(cleanName, " : ", "-")
	cleanName = strings.ReplaceAll(cleanName, ": ", "-")
	cleanName = strings.ReplaceAll(cleanName, " - ", "-")
	cleanName = strings.ReplaceAll(cleanName, " -- ", "-")
	cleanName = strings.ReplaceAll(cleanName, " – ", "-")
	cleanName = strings.ReplaceAll(cleanName, " –– ", "-")
	cleanName = strings.ReplaceAll(cleanName, " — ", "-")
	cleanName = strings.ReplaceAll(cleanName, " —— ", "-")
	cleanName = strings.ReplaceAll(cleanName, "… ", "_")
	cleanName = regexp.MustCompile(`[-|\\]+`).ReplaceAllString(cleanName, "-")
	cleanName = strings.ReplaceAll(cleanName, " & ", "_and_")
	cleanName = strings.ReplaceAll(cleanName, "&", "_and_")
	cleanName = strings.ReplaceAll(cleanName, "?", "")
	cleanName = regexp.MustCompile(`[^a-zA-Z0-9_\- ]+`).ReplaceAllString(cleanName, "")
	cleanName = strings.ReplaceAll(cleanName, " ", "_")
	cleanName = strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) {
			return -1
		}
		return r
	}, cleanName)
	cleanName = strings.Map(func(r rune) rune {
		if !utf8.ValidRune(r) {
			return -1
		}
		return r
	}, cleanName)

	return cleanName
}
