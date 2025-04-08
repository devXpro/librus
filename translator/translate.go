package translator

import (
	"context"
	"fmt"
	"librus/mongo/cache"
	"strings"
)

// TranslateText is the main function to translate text with placeholders handling and caching
func TranslateText(targetLanguage, text string) (string, error) {
	ctx := context.Background()
	translator := translatorFactory()
	translatorType := fmt.Sprintf("%T", translator)

	// Handle placeholders
	textWithPlaceholders, placeholders := preprocessText(text)

	// Check cache first
	cacheKey := generateCacheKey(translatorType, targetLanguage, textWithPlaceholders)
	cachedTranslation, err := cache.GetCachedTranslationByKey(cacheKey)
	if err != nil {
		return "", fmt.Errorf("cache.GetCachedTranslationByKey: %v", err)
	}
	if cachedTranslation != "" {
		return replacePlaceholders(cachedTranslation, placeholders), nil
	}

	// Translate if not in cache
	translatedText, err := translator.Translate(ctx, textWithPlaceholders, targetLanguage)
	if err != nil {
		return "", err
	}

	// Save to cache
	if err := cache.SetCachedTranslationByKey(cacheKey, translatedText); err != nil {
		return "", fmt.Errorf("cache.SetCachedTranslationByKey: %v", err)
	}

	// Replace placeholders with original content
	translatedWithPlaceholders := replacePlaceholders(translatedText, placeholders)

	// Clean up any remaining <nt> and </nt> tags
	return cleanupNoTranslateTags(translatedWithPlaceholders), nil
}

// cleanupNoTranslateTags removes <nt> and </nt> tags from a string
func cleanupNoTranslateTags(text string) string {
	text = strings.ReplaceAll(text, "<nt>", "")
	text = strings.ReplaceAll(text, "</nt>", "")
	return text
}
