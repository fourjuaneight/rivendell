package main

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Turns first letter of a string to uppercase, capitalizing the string
func toCapitalized(str string) string {
	return cases.Title(language.English, cases.Compact).String(str)
}
