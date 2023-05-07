package translator

import (
	"cloud.google.com/go/translate"
	"context"
	"fmt"
	"golang.org/x/text/language"
	"librus/mongo/cache"
	"regexp"
	"strconv"
	"strings"
)

func TranslateText(targetLanguage, text string) (string, error) {
	placeholderPrefix := "NO_TRANSLATE_PLACEHOLDER"
	var placeholders []string
	re := regexp.MustCompile(`<nt>([^<]+)</nt>`)
	textWithPlaceholders := re.ReplaceAllStringFunc(text, func(match string) string {
		content := re.ReplaceAllString(match, "$1")
		placeholder := placeholderPrefix + strconv.Itoa(len(placeholders))
		placeholders = append(placeholders, content)
		return placeholder
	})
	translatedTextWithPlaceholders, err := translateText(targetLanguage, textWithPlaceholders)
	if err != nil {
		return text, err
	}

	return replacePlaceholders(translatedTextWithPlaceholders, placeholders), nil
}

func replacePlaceholders(text string, placeholders []string) string {
	for i, placeholder := range placeholders {
		text = strings.ReplaceAll(text, "NO_TRANSLATE_PLACEHOLDER"+strconv.Itoa(i), placeholder)
	}
	return text
}

func translateText(targetLanguage string, text string) (string, error) {
	ctx := context.Background()

	cachedTranslation, err := cache.GetCachedTranslation(targetLanguage, text)
	if err != nil {
		return "", fmt.Errorf("cache.GetCachedTranslation: %v", err)
	}
	if cachedTranslation != "" {
		return cachedTranslation, nil
	}

	lang, err := language.Parse(targetLanguage)
	if err != nil {
		return "", fmt.Errorf("language.Parse: %v", err)
	}

	translateClient, err := translate.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer translateClient.Close()

	resp, err := translateClient.Translate(ctx, []string{text}, lang, &translate.Options{Format: "text"})
	if err != nil {
		return "", fmt.Errorf("translate: %v", err)
	}
	if len(resp) == 0 {
		return "", fmt.Errorf("translate returned empty response to text: %s", text)
	}

	translation := resp[0].Text

	// Save the translation to cache
	if err := cache.SetCachedTranslation(targetLanguage, text, translation); err != nil {
		return "", fmt.Errorf("cache.SetCachedTranslation: %v", err)
	}

	return translation, nil
}
