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
						"Translate the text without adding or removing information. " +
						"Preserve formatting, placeholders, and special markers.",
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
