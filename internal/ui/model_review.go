package ui

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"golang.org/x/sync/errgroup"

	"github.com/trankhanh040147/plancli/internal/app"
	appcontext "github.com/trankhanh040147/plancli/internal/context"
	"github.com/trankhanh040147/plancli/internal/message"
	"github.com/trankhanh040147/plancli/internal/prompt"
)

// buildAttachments converts review context files to message attachments
func buildAttachments(reviewCtx *appcontext.ReviewContext) []message.Attachment {
	var attachments []message.Attachment
	for filePath, content := range reviewCtx.FileContents {
		attachments = append(attachments, message.Attachment{
			FilePath: filePath,
			FileName: filepath.Base(filePath),
			MimeType: "text/plain",
			Content:  []byte(content),
		})
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

		// Create separate context for message subscription that can be cancelled independently
		msgCtx, msgCancel := context.WithCancel(ctx)

		// Start coordinator in goroutine
		g, gCtx := errgroup.WithContext(ctx)

		// Subscribe to messages with separate context
		messageEvents := appInstance.Messages.Subscribe(msgCtx)
		messageReadBytes := make(map[string]int)
		var fullResponse strings.Builder
		var fullResponseMutex sync.Mutex

		// Message subscription goroutine - NOT in errgroup
		messageDone := make(chan struct{})
		go func() {
			defer close(messageDone)
			for {
				select {
				case <-msgCtx.Done():
					return
				case event, ok := <-messageEvents:
					if !ok {
						return
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
							fullResponseMutex.Lock()
							fullResponse.WriteString(chunk)
							fullResponseMutex.Unlock()
							messageReadBytes[msg.ID] = len(content)

							select {
							case chunkChan <- chunk:
							case <-msgCtx.Done():
								return
							}
						}
					}
				}
			}
		}()

		// Goroutine to run coordinator - this is the only one in errgroup
		var coordinatorErr error
		g.Go(func() error {
			_, err := appInstance.AgentCoordinator.Run(gCtx, sessionID, userPrompt, attachments...)
			if err != nil {
				coordinatorErr = err
				select {
				case errChan <- err:
				case <-gCtx.Done():
				}
				return err
			}
			return nil
		})

		// Wait for completion in background and send done message
		go func() {
			// Wait ONLY for coordinator to complete (not message goroutine)
			waitErr := g.Wait()

			// If coordinator completed successfully, wait a grace period for final messages
			if waitErr == nil && coordinatorErr == nil {
				time.Sleep(500 * time.Millisecond)
			}

			// Cancel message subscription to allow message goroutine to exit
			msgCancel()

			// Wait for message goroutine to exit (with timeout)
			select {
			case <-messageDone:
				// Message goroutine exited
			case <-time.After(1 * time.Second):
				// Timeout - message goroutine didn't exit, continue anyway
			}

			if waitErr != nil {
				// Check if it's a cancellation error
				if waitErr != context.Canceled {
					select {
					case errChan <- waitErr:
					case <-ctx.Done():
					}
				}
				return
			}

			// Fetch final assistant message from service as fallback to ensure completeness
			func() {
				fullResponseMutex.Lock()
				defer fullResponseMutex.Unlock()

				messages, err := appInstance.Messages.List(ctx, sessionID)
				if err == nil {
					// Find the last assistant message for this session
					for i := len(messages) - 1; i >= 0; i-- {
						msg := messages[i]
						if msg.SessionID == sessionID && msg.Role == message.Assistant {
							content := msg.Content().String()
							if content != "" {
								content = strings.TrimLeft(content, " \t")
								fullResponse.Reset()
								fullResponse.WriteString(content)
							}
							break
						}
					}
				}
			}()

			// Send final response when coordinator completes
			fullResponseMutex.Lock()
			response := fullResponse.String()
			fullResponseMutex.Unlock()

			select {
			case doneChan <- response:
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
