package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/trankhanh040147/revcli/internal/gemini"
)

// PruneFile summarizes a file using Gemini Flash model
// Returns the summary string and any error
func PruneFile(apiKey, filePath, content string) (string, error) {
	// Use Gemini Flash for cheap summarization
	flashClient, err := gemini.NewClient(context.Background(), apiKey, "gemini-2.5-flash")
	if err != nil {
		return "", fmt.Errorf("failed to create flash client: %w", err)
	}
	defer flashClient.Close()

	// Build summarization prompt
	prompt := fmt.Sprintf(`Summarize this code file in one sentence. Focus on what the file does, its main purpose, and key functionality. Be concise.

File: %s

%s`, filePath, content)

	// Use a simple system prompt for summarization
	systemPrompt := "You are a code summarization assistant. Provide concise, one-sentence summaries of code files."

	// Generate summary
	summary, err := flashClient.GenerateContent(context.Background(), systemPrompt, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}

	// Clean up summary (remove quotes, trim whitespace)
	summary = strings.TrimSpace(summary)
	summary = strings.Trim(summary, `"`)

	return summary, nil
}
