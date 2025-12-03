package gemini

import (
	"context"
	"fmt"
	"io"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps the Gemini API client
type Client struct {
	client  *genai.Client
	model   *genai.GenerativeModel
	chat    *genai.ChatSession
	modelID string
}

// StreamCallback is called for each chunk of streamed response
type StreamCallback func(chunk string)

// NewClient creates a new Gemini client
func NewClient(ctx context.Context, apiKey, modelID string) (*Client, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	model := client.GenerativeModel(modelID)

	// Configure the model for code review
	model.SetTemperature(0.3) // Lower temperature for more focused responses
	model.SetTopP(0.95)
	model.SetTopK(40)

	// Set safety settings to be less restrictive for code content
	model.SafetySettings = []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockOnlyHigh,
		},
		{
			Category:  genai.HarmCategoryHateSpeech,
			Threshold: genai.HarmBlockOnlyHigh,
		},
		{
			Category:  genai.HarmCategoryDangerousContent,
			Threshold: genai.HarmBlockOnlyHigh,
		},
		{
			Category:  genai.HarmCategorySexuallyExplicit,
			Threshold: genai.HarmBlockOnlyHigh,
		},
	}

	return &Client{
		client:  client,
		model:   model,
		modelID: modelID,
	}, nil
}

// StartChat initializes a chat session with the system prompt
func (c *Client) StartChat(systemPrompt string) {
	c.model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemPrompt)},
	}
	c.chat = c.model.StartChat()
}

// SendMessage sends a message and returns the full response
func (c *Client) SendMessage(ctx context.Context, message string) (string, error) {
	if c.chat == nil {
		return "", fmt.Errorf("chat session not initialized, call StartChat first")
	}

	resp, err := c.chat.SendMessage(ctx, genai.Text(message))
	if err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	return extractText(resp), nil
}

// SendMessageStream sends a message and streams the response
func (c *Client) SendMessageStream(ctx context.Context, message string, callback StreamCallback) (string, error) {
	if c.chat == nil {
		return "", fmt.Errorf("chat session not initialized, call StartChat first")
	}

	iter := c.chat.SendMessageStream(ctx, genai.Text(message))

	var fullResponse string
	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fullResponse, fmt.Errorf("stream error: %w", err)
		}

		chunk := extractText(resp)
		fullResponse += chunk
		if callback != nil {
			callback(chunk)
		}
	}

	return fullResponse, nil
}

// GenerateContent sends a one-off generation request (no chat history)
func (c *Client) GenerateContent(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	// Create a temporary model with system instruction
	model := c.client.GenerativeModel(c.modelID)
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemPrompt)},
	}

	resp, err := model.GenerateContent(ctx, genai.Text(userPrompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	return extractText(resp), nil
}

// GenerateContentStream sends a one-off generation request with streaming
func (c *Client) GenerateContentStream(ctx context.Context, systemPrompt, userPrompt string, callback StreamCallback) (string, error) {
	// Create a temporary model with system instruction
	model := c.client.GenerativeModel(c.modelID)
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemPrompt)},
	}
	model.SetTemperature(0.3)

	iter := model.GenerateContentStream(ctx, genai.Text(userPrompt))

	var fullResponse string
	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fullResponse, fmt.Errorf("stream error: %w", err)
		}

		chunk := extractText(resp)
		fullResponse += chunk
		if callback != nil {
			callback(chunk)
		}
	}

	return fullResponse, nil
}

// Close closes the client connection
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// extractText extracts text content from a GenerateContentResponse
func extractText(resp *genai.GenerateContentResponse) string {
	if resp == nil || len(resp.Candidates) == 0 {
		return ""
	}

	var text string
	for _, candidate := range resp.Candidates {
		if candidate.Content != nil {
			for _, part := range candidate.Content.Parts {
				if t, ok := part.(genai.Text); ok {
					text += string(t)
				}
			}
		}
	}

	return text
}

// StreamWriter wraps an io.Writer for streaming responses
type StreamWriter struct {
	Writer io.Writer
}

// Write implements StreamCallback for writing to an io.Writer
func (sw *StreamWriter) Write(chunk string) {
	sw.Writer.Write([]byte(chunk))
}

// GetModelID returns the current model ID
func (c *Client) GetModelID() string {
	return c.modelID
}

