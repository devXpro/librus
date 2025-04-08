package translator

import (
	"crypto/md5"
	"encoding/hex"
	"librus/helper"
)

// translatorFactory creates the appropriate translator based on environment configuration
func translatorFactory() Translator {
	openAIKey := helper.GetEnv("OPEN_AI_KEY", "")
	if openAIKey != "" {
		return &ChatGPTTranslator{apiKey: openAIKey}
	}
	return &GoogleTranslator{}
}

// generateCacheKey creates a unique key for caching translations
func generateCacheKey(translatorType, targetLanguage, text string) string {
	hash := md5.Sum([]byte("v2" + translatorType + targetLanguage + text))
	return hex.EncodeToString(hash[:])
}
