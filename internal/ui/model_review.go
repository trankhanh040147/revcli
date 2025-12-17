package ui

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	appcontext "github.com/trankhanh040147/revcli/internal/context"
	"github.com/trankhanh040147/revcli/internal/gemini"
	"github.com/trankhanh040147/revcli/internal/prompt"
)

// startReview initiates the code review with streaming support
func (m *Model) startReview() tea.Cmd {
	// Initialize chat with system prompt (with preset and intent if specified)
	var presetPrompt string
	var presetReplace bool
	if m.preset != nil {
		presetPrompt = m.preset.Prompt
		presetReplace = m.preset.Replace
	}
	systemPrompt := appcontext.GetSystemPromptWithIntent(m.reviewCtx.Intent, presetPrompt, presetReplace)
	m.client.StartChat(systemPrompt)

	// Rebuild prompt with pruned files if any
	userPrompt := m.reviewCtx.UserPrompt
	if len(m.reviewCtx.PrunedFiles) > 0 {
		userPrompt = prompt.BuildReviewPromptWithPruning(
			m.reviewCtx.RawDiff,
			m.reviewCtx.FileContents,
			m.reviewCtx.PrunedFiles,
		)
	}

	// Return command that starts streaming in goroutine
	return streamReviewCmd(m.ctx, m.client, userPrompt)
}

// streamReviewCmd creates a command that streams the review response
func streamReviewCmd(ctx context.Context, client *gemini.Client, userPrompt string) tea.Cmd {
	return func() tea.Msg {
		// Channel to send chunks from goroutine to tea program
		chunkChan := make(chan string, 10)
		errChan := make(chan error, 1)
		doneChan := make(chan string, 1)

		// Start streaming in goroutine
		go func() {
			var fullResponse strings.Builder

			_, err := client.SendMessageStream(ctx, userPrompt, func(chunk string) {
				fullResponse.WriteString(chunk)
				select {
				case chunkChan <- chunk:
				case <-ctx.Done():
					return
				}
			})

			if err != nil {
				select {
				case errChan <- err:
				case <-ctx.Done():
				}
				return
			}

			select {
			case doneChan <- fullResponse.String():
			case <-ctx.Done():
			}
		}()

		// Return initial message to start receiving chunks
		return StreamStartMsg{
			ChunkChan: chunkChan,
			ErrChan:   errChan,
			DoneChan:  doneChan,
		}
	}
}

// StreamStartMsg signals that streaming has started and provides channels
type StreamStartMsg struct {
	ChunkChan chan string
	ErrChan   chan error
	DoneChan  chan string
}
