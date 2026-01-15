package utils

import (
	"strings"
)

func NormalizeLangCode(lang string) string {
	return strings.TrimSpace(lang)
}
