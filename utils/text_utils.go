package utils

import (
	"html"
	"regexp"
	"strings"
)

// StripTags removes all HTML tags from a string and decodes HTML entities.
func StripTags(s string) string {
	re := regexp.MustCompile("<[^>]*>")
	clean := re.ReplaceAllString(s, " ")
	return CleanText(html.UnescapeString(clean))
}

// DecodeHtml decodes HTML entities (like &oacute; to ó).
func DecodeHtml(s string) string {
	return html.UnescapeString(s)
}

// CleanText removes leading and trailing asterisks and extra whitespace.
func CleanText(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "* ")
	return s
}

// NormalizeCategory normalizes category names for case-insensitive and accent-insensitive matching.
func NormalizeCategory(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	r := strings.NewReplacer(
		"á", "a",
		"é", "e",
		"í", "i",
		"ó", "o",
		"ú", "u",
		"ü", "u",
	)
	return r.Replace(s)
}
