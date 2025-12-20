package ui

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/sync/errgroup"

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

	// Get web search setting from intent (default to true if intent is nil)
	webSearchEnabled := true
	if m.reviewCtx.Intent != nil {
		webSearchEnabled = m.reviewCtx.Intent.WebSearchEnabled
	}

	// Create new context for this command
	ctx, cancel := context.WithCancel(m.rootCtx)
	m.activeCancel = cancel
	// Return command that starts streaming in goroutine
	return streamReviewCmd(ctx, m.client, userPrompt, webSearchEnabled)
}

// streamReviewCmd creates a command that streams the review response
func streamReviewCmd(ctx context.Context, client *gemini.Client, userPrompt string, webSearchEnabled bool) tea.Cmd {
	return func() tea.Msg {
		// Channel to send chunks from goroutine to tea program
		chunkChan := make(chan string, gemini.ChunkChannelBufferSize)
		errChan := make(chan error, gemini.ErrorChannelBufferSize)
		doneChan := make(chan string, gemini.DoneChannelBufferSize)

		// Start streaming using errgroup for proper error propagation
		g, gCtx := errgroup.WithContext(ctx)
		g.Go(func() error {
			var fullResponse strings.Builder

			_, err := client.SendMessageStream(gCtx, userPrompt, func(chunk string) {
				fullResponse.WriteString(chunk)
				select {
				case chunkChan <- chunk:
				case <-gCtx.Done():
					return
				}
			}, webSearchEnabled)

			if err != nil {
				select {
				case errChan <- err:
				case <-gCtx.Done():
				}
				return err
			}

			select {
			case doneChan <- fullResponse.String():
			case <-gCtx.Done():
			}
			return nil
		})

		// Wait for completion in background and handle errors
		go func() {
			if err := g.Wait(); err != nil {
				// Error already sent through errChan, but ensure it's sent if errgroup fails
				select {
				case errChan <- err:
				case <-ctx.Done():
				}
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
