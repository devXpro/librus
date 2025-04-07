package translator

import (
	"context"
	"regexp"
	"strconv"
	"strings"
)

// Translator defines the interface for any translation service
type Translator interface {
	Translate(ctx context.Context, text, targetLanguage string) (string, error)
}

// preprocessText handles the replacement of non-translatable content with placeholders
func preprocessText(text string) (string, []string) {
	placeholderPrefix := "NO_TRANSLATE_PLACEHOLDER"
	var placeholders []string
	re := regexp.MustCompile(`<nt>([^<]+)</nt>`)
	textWithPlaceholders := re.ReplaceAllStringFunc(text, func(match string) string {
		content := re.ReplaceAllString(match, "$1")
		placeholder := placeholderPrefix + strconv.Itoa(len(placeholders))
		placeholders = append(placeholders, content)
		return placeholder
	})
	return textWithPlaceholders, placeholders
}

// replacePlaceholders restores the original non-translatable content after translation
func replacePlaceholders(text string, placeholders []string) string {
	for i, placeholder := range placeholders {
		text = strings.ReplaceAll(text, "NO_TRANSLATE_PLACEHOLDER"+strconv.Itoa(i), placeholder)
	}
	return text
}
