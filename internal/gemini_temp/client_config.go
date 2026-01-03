package gemini_temp

import (
	"github.com/samber/lo"
	"google.golang.org/genai"
)

// buildSafetySettings constructs safety settings from threshold string
func (c *Client) buildSafetySettings() []*genai.SafetySetting {
	threshold := mapSafetyThreshold(c.safetyThreshold)
	categories := getAllHarmCategories()
	settings := lo.Map(categories, func(category genai.HarmCategory, _ int) *genai.SafetySetting {
		return &genai.SafetySetting{
			Category:  category,
			Threshold: threshold,
		}
	})
	return settings
}

// newGenerationConfig creates a GenerateContentConfig with the given system prompt
func (c *Client) newGenerationConfig(systemPrompt string, webSearchEnabled bool) *genai.GenerateContentConfig {
	temp := c.modelConfig.Temperature
	topP := c.modelConfig.TopP
	topK := float32(c.modelConfig.TopK)

	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{{Text: systemPrompt}},
		},
		Temperature:    &temp,
		TopP:           &topP,
		TopK:           &topK,
		SafetySettings: c.buildSafetySettings(),
	}

	if webSearchEnabled {
		config.Tools = []*genai.Tool{
			{GoogleSearch: &genai.GoogleSearch{}},
		}
	}

	return config
}

// appendUserMessageAndPrepareTurn creates user message, appends it to history,
// and prepares the generation config
func (c *Client) appendUserMessageAndPrepareTurn(message string, webSearchEnabled bool) (*genai.Content, *genai.GenerateContentConfig) {
	userMsg := &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{{Text: message}},
	}

	config := c.newGenerationConfig(c.systemPrompt, webSearchEnabled)

	c.history = append(c.history, userMsg)
	return userMsg, config
}
