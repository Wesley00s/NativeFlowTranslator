package utils

import (
	"regexp"
	"strings"
)

var languageMap = map[string]string{
	"brazilian portuguese":   "ptbr",
	"united states english":  "enus",
	"british english":        "enuk",
	"european spanish":       "es",
	"latin american spanish": "eslat",
	"canadian french":        "frca",

	"portuguese": "ptpt",
	"english":    "en",
	"spanish":    "es",
	"french":     "fr",
	"german":     "de",
	"italian":    "it",
	"japanese":   "ja",
	"chinese":    "zh",
	"russian":    "ru",
}

func NormalizeLangCode(lang string) string {
	lower := strings.ToLower(strings.TrimSpace(lang))

	if code, exists := languageMap[lower]; exists {
		return code
	}

	reg := regexp.MustCompile("[^a-zA-Z0-9]+")
	clean := reg.ReplaceAllString(lower, "")

	return clean
}
