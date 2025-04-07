package translator

import (
	"strings"
	"testing"
	"time"
)

func TestTranslateText(t *testing.T) {
	// Test case 1: Basic translation from English to Spanish
	t.Run("Basic translation", func(t *testing.T) {
		sourceText := "Hello, my name is John."
		targetLang := "es" // Spanish

		// First translation
		startTime := time.Now()
		translatedText, err := TranslateText(targetLang, sourceText)
		firstTranslationTime := time.Since(startTime)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if translatedText == "" {
			t.Fatal("Expected non-empty translation, got empty string")
		}

		if translatedText == sourceText {
			t.Fatal("Translation is identical to source text, expected it to be different")
		}

		// Check for Spanish words that should be in the translation
		if !strings.Contains(strings.ToLower(translatedText), "hola") &&
			!strings.Contains(strings.ToLower(translatedText), "nombre") {
			t.Logf("Unexpected translation result: %s", translatedText)
			t.Logf("Expected Spanish words not found in the translation")
		}

		// Second call (should use cache)
		startTime = time.Now()
		secondTranslatedText, err := TranslateText(targetLang, sourceText)
		secondTranslationTime := time.Since(startTime)

		if err != nil {
			t.Fatalf("Unexpected error on second call: %v", err)
		}

		if secondTranslatedText != translatedText {
			t.Fatalf("Second translation differs from first: '%s' vs '%s'", secondTranslatedText, translatedText)
		}

		// The second call should be faster due to caching
		t.Logf("First translation time: %v", firstTranslationTime)
		t.Logf("Second translation time: %v", secondTranslationTime)

		if secondTranslationTime > firstTranslationTime {
			t.Logf("Warning: Second translation took longer than first, caching might not be working")
		}
	})

	// Test case 2: Translation with placeholders
	t.Run("Translation with placeholders", func(t *testing.T) {
		sourceText := "Hello, my name is <nt>John Doe</nt> and I am <nt>30</nt> years old."
		targetLang := "fr" // French

		translatedText, err := TranslateText(targetLang, sourceText)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if translatedText == "" {
			t.Fatal("Expected non-empty translation, got empty string")
		}

		// The placeholders should remain unchanged in the translation
		if !strings.Contains(translatedText, "John Doe") {
			t.Fatalf("Placeholder 'John Doe' not preserved in translation: %s", translatedText)
		}

		if !strings.Contains(translatedText, "30") {
			t.Fatalf("Placeholder '30' not preserved in translation: %s", translatedText)
		}

		// Check for French words
		if !strings.Contains(strings.ToLower(translatedText), "bonjour") &&
			!strings.Contains(strings.ToLower(translatedText), "ans") {
			t.Logf("Unexpected translation result: %s", translatedText)
			t.Logf("Expected French words not found in the translation")
		}
	})
}
