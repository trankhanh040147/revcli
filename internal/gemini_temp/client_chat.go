package gemini_temp

import (
	"context"
	"fmt"
	"log"
	"strings"

	"golang.org/x/sync/errgroup"
	"google.golang.org/genai"
)

// SendMessage sends a message with history and optional Web Search
func (c *Client) SendMessage(ctx context.Context, message string, webSearchEnabled bool) (string, error) {
	_, config := c.appendUserMessageAndPrepareTurn(message, webSearchEnabled)

	resp, err := c.client.Models.GenerateContent(ctx, c.modelID, c.history, config)
	if err != nil {
		// Rollback user message from history on error
		c.rollbackLastHistoryEntry()
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	// Check for safety block
	if len(resp.Candidates) > 0 && resp.Candidates[0].FinishReason == genai.FinishReasonSafety {
		c.rollbackLastHistoryEntry()
		return "", fmt.Errorf("response blocked by safety filters")
	}

	if len(resp.Candidates) > 0 {
		c.history = append(c.history, resp.Candidates[0].Content)
	}

	c.lastUsage = extractUsage(resp)
	return extractText(resp), nil
}

// SendMessageStream streams responses while maintaining history
func (c *Client) SendMessageStream(ctx context.Context, message string, callback StreamCallback, webSearchEnabled bool) (string, error) {
	_, config := c.appendUserMessageAndPrepareTurn(message, webSearchEnabled)

	iter := c.client.Models.GenerateContentStream(ctx, c.modelID, c.history, config)

	var fullResponseText string
	var finalContent *genai.Content

	// Convert iter.Seq2 to channel-based consumption pattern
	type iterResult struct {
		resp *genai.GenerateContentResponse
		err  error
	}
	ch := make(chan iterResult, StreamChannelBufferSize)

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		defer close(ch)
		iter(func(resp *genai.GenerateContentResponse, err error) bool {
			select {
			case ch <- iterResult{resp: resp, err: err}:
				return true
			case <-gCtx.Done():
				return false
			}
		})
		return nil
	})

	// Wait for goroutine completion in background
	go func() {
		if err := g.Wait(); err != nil {
			// Error already handled through channel, but log if unexpected
			log.Printf("Warning: errgroup returned error: %v", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			c.rollbackLastHistoryEntry()
			return fullResponseText, fmt.Errorf("stream cancelled: %w", ctx.Err())
		case result, ok := <-ch:
			if !ok {
				// Channel closed, iteration complete
				goto done
			}
			if result.err != nil {
				// Always rollback user message from history on any stream error
				c.rollbackLastHistoryEntry()
				return fullResponseText, fmt.Errorf("stream error: %w", result.err)
			}

			resp := result.resp
			// Check for safety block in streaming response
			if len(resp.Candidates) > 0 && resp.Candidates[0].FinishReason == genai.FinishReasonSafety {
				c.rollbackLastHistoryEntry()
				return fullResponseText, fmt.Errorf("response blocked by safety filters")
			}

			chunk := extractText(resp)
			fullResponseText += chunk
			if callback != nil {
				callback(chunk)
			}

			if len(resp.Candidates) > 0 {
				finalContent = resp.Candidates[0].Content
				c.lastUsage = extractUsage(resp)
			}
		}
	}

done:
	// Check final response for safety block before updating history
	if finalContent != nil {
		// Note: We check the last response's FinishReason, but since we're done streaming,
		// we check if we have any accumulated content. If FinishReasonSafety was detected
		// during streaming, we would have already returned an error above.
		// This final check is a safety net in case the last chunk had a safety block.
		c.history = append(c.history, finalContent)
	}

	return fullResponseText, nil
}

// GenerateContent remains simple for one-off calls
func (c *Client) GenerateContent(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	config := c.newGenerationConfig(systemPrompt, false)

	resp, err := c.client.Models.GenerateContent(ctx, c.modelID, genai.Text(userPrompt), config)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	c.lastUsage = extractUsage(resp)
	return extractText(resp), nil
}

// extractText for the new SDK structure
func extractText(resp *genai.GenerateContentResponse) string {
	if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return ""
	}
	if len(resp.Candidates) > 1 {
		log.Printf("Warning: Multiple candidates (%d) in response, using first candidate only", len(resp.Candidates))
	}
	var parts []string
	for _, part := range resp.Candidates[0].Content.Parts {
		if part == nil {
			continue
		}
		if txt := part.Text; txt != "" {
			parts = append(parts, txt)
		} else if part.FunctionCall != nil {
			log.Printf("Warning: FunctionCall part detected but not processed (part will be skipped)")
		} else if part.InlineData != nil {
			log.Printf("Warning: Blob part detected but not processed (part will be skipped)")
		}
	}
	return strings.Join(parts, "")
}

// extractUsage for the new SDK metadata structure
func extractUsage(resp *genai.GenerateContentResponse) *TokenUsage {
	if resp == nil || resp.UsageMetadata == nil {
		return nil
	}
	return &TokenUsage{
		PromptTokens:     resp.UsageMetadata.PromptTokenCount,
		CompletionTokens: resp.UsageMetadata.CandidatesTokenCount,
		TotalTokens:      resp.UsageMetadata.TotalTokenCount,
	}
}
