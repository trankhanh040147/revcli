package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/trankhanh040147/revcli/internal/gemini"
)

// PruneFile summarizes a file using Gemini Flash model
// Returns the summary string and any error
// flashClient must be a non-nil client instance (typically gemini-2.5-flash)
func PruneFile(ctx context.Context, flashClient *gemini.Client, filePath, content string) (string, error) {
	if flashClient == nil {
		return "", fmt.Errorf("flash client is nil")
	}

	// Build summarization prompt
	prompt := fmt.Sprintf(PruneFilePromptTemplate, filePath, content)

	// Use a simple system prompt for summarization
	systemPrompt := "You are a code summarization assistant. Provide concise, one-sentence summaries of code files."

	// Generate summary
	summary, err := flashClient.GenerateContent(ctx, systemPrompt, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}

	// Clean up summary (remove quotes, trim whitespace)
	summary = strings.TrimSpace(summary)
	summary = strings.Trim(summary, `"`)

	return summary, nil
}
