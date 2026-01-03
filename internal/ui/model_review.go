package ui

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/sync/errgroup"

	appcontext "github.com/trankhanh040147/revcli/internal/context"
	"github.com/trankhanh040147/revcli/internal/app"
	"github.com/trankhanh040147/revcli/internal/message"
	"github.com/trankhanh040147/revcli/internal/prompt"
)

// buildAttachments converts review context files to message attachments
func buildAttachments(reviewCtx *appcontext.ReviewContext) []message.Attachment {
	var attachments []message.Attachment
	for filePath, content := range reviewCtx.FileContents {
		attachments = append(attachments, message.NewTextAttachment(filePath, content))
	}
	return attachments
}

// startReview initiates the code review with streaming support
func (m *Model) startReview() tea.Cmd {
	// Rebuild prompt with pruned files if any
	userPrompt := m.reviewCtx.UserPrompt
	if len(m.reviewCtx.PrunedFiles) > 0 {
		userPrompt = prompt.BuildReviewPromptWithPruning(
			m.reviewCtx.RawDiff,
			m.reviewCtx.FileContents,
			m.reviewCtx.PrunedFiles,
		)
	}

	// Build attachments
	attachments := buildAttachments(m.reviewCtx)

	// Create new context for this command
	ctx, cancel := context.WithCancel(m.rootCtx)
	m.activeCancel = cancel
	// Return command that starts streaming via coordinator
	return streamReviewCmd(ctx, m.app, m.sessionID, userPrompt, attachments)
}

// streamReviewCmd creates a command that streams the review response using coordinator
func streamReviewCmd(ctx context.Context, appInstance *app.App, sessionID, userPrompt string, attachments []message.Attachment) tea.Cmd {
	return func() tea.Msg {
		// Channel to send chunks from goroutine to tea program
		chunkChan := make(chan string, 100)
		errChan := make(chan error, 10)
		doneChan := make(chan string, 1)

		// Start coordinator in goroutine
		g, gCtx := errgroup.WithContext(ctx)
		
		// Subscribe to messages
		messageEvents := appInstance.Messages.Subscribe(gCtx)
		messageReadBytes := make(map[string]int)
		var fullResponse strings.Builder

		// Goroutine to handle message events
		g.Go(func() error {
			for {
				select {
				case <-gCtx.Done():
					return gCtx.Err()
				case event, ok := <-messageEvents:
					if !ok {
						return nil
					}
					msg := event.Payload
					// Filter by sessionID and assistant role
					if msg.SessionID == sessionID && msg.Role == message.Assistant && len(msg.Parts) > 0 {
						content := msg.Content().String()
						readBytes := messageReadBytes[msg.ID]

						if len(content) > readBytes {
							// New content available
							chunk := content[readBytes:]
							// Trim leading whitespace on first chunk
							if readBytes == 0 {
								chunk = strings.TrimLeft(chunk, " \t")
							}
							fullResponse.WriteString(chunk)
							messageReadBytes[msg.ID] = len(content)

							select {
							case chunkChan <- chunk:
							case <-gCtx.Done():
								return gCtx.Err()
							}
						}
					}
				}
			}
		})

		// Goroutine to run coordinator
		g.Go(func() error {
			_, err := appInstance.AgentCoordinator.Run(gCtx, sessionID, userPrompt, attachments...)
			if err != nil {
				select {
				case errChan <- err:
				case <-gCtx.Done():
				}
				return err
			}
			// Coordinator completed - signal done (fullResponse will be collected from final message)
			return nil
		})

		// Wait for completion in background and send done message
		go func() {
			if err := g.Wait(); err != nil {
				// Check if it's a cancellation error
				if err != context.Canceled {
					select {
					case errChan <- err:
					case <-ctx.Done():
					}
				}
				return
			}
			// Send final response when coordinator completes
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
