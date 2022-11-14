package utils

import (
	"regexp"
)

// Sanitize Removes redundant spaces and line breaks on text
func Sanitize(text string) string {
	matchRedundantSpaces := regexp.MustCompile(`\s\s+`)
	matchLineBreaks := regexp.MustCompile(`\\+n`)

	text = matchRedundantSpaces.ReplaceAllString(text, " ")
	text = matchLineBreaks.ReplaceAllString(text, "\n")
	return text
}
