package translator

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// ChatGPTTranslator implements the Translator interface using OpenAI's GPT model
type ChatGPTTranslator struct {
	apiKey string
}

func (c *ChatGPTTranslator) Translate(ctx context.Context, text, targetLanguage string) (string, error) {
	client := openai.NewClient(c.apiKey)

	prompt := fmt.Sprintf("Translate the following text to %s. "+
		"Only return the translated text without any additional explanation or notes:\n\n%s", targetLanguage, text)

	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: "gpt-4o-mini",
			Messages: []openai.ChatCompletionMessage{
				{
					Role: "system",
					Content: "You are a professional translator. " +
						"Translate ABSOLUTELY EVERYTHING in the text to the target language, including ALL names, " +
						"titles, places, and any other proper nouns. " +
						"Leave NO words in the original language. " +
						"The ONLY exception is content between <nt> and </nt> tags, which must remain untouched. " +
						"Translate personal names and surnames fully into the target language. " +
						"For example, 'Jan Kowalski' could become 'John Smith' in English or 'Іван Коваль' in Ukrainian. " +
						"ALL text should look completely native and natural in the target language.",
				},
				{
					Role:    "user",
					Content: prompt,
				},
			},
			Temperature: 0.1, // Low temperature for more predictable translations
		},
	)

	if err != nil {
		return "", fmt.Errorf("openai API error: %v", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("openai returned empty response to text: %s", text)
	}

	return resp.Choices[0].Message.Content, nil
}
