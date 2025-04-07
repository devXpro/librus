package translator

import (
	"context"
	"fmt"

	"cloud.google.com/go/translate"
	"golang.org/x/text/language"
)

// GoogleTranslator implements the Translator interface using Google Translate API
type GoogleTranslator struct{}

func (g *GoogleTranslator) Translate(ctx context.Context, text, targetLanguage string) (string, error) {
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

	return resp[0].Text, nil
}
