package ui

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"charm.land/fantasy"

	appcontext "github.com/trankhanh040147/revcli/internal/context"
	"github.com/trankhanh040147/revcli/internal/app"
	"github.com/trankhanh040147/revcli/internal/message"
	"github.com/trankhanh040147/revcli/internal/preset"
)

// RunSimple runs a simple non-interactive review using coordinator
func RunSimple(ctx context.Context, w io.Writer, reviewCtx *appcontext.ReviewContext, appInstance *app.App, sessionID string, p *preset.Preset) error {
	fmt.Fprintln(w, RenderTitle("🔍 Code Review"))
	fmt.Fprintln(w, RenderSubtitle(reviewCtx.Summary()))
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Analyzing your code changes...")
	fmt.Fprintln(w)

	// Build prompt and attachments
	prompt := reviewCtx.UserPrompt
	attachments := buildAttachments(reviewCtx)

	// Use app.RunNonInteractive which handles streaming
	// Note: RunNonInteractive doesn't support attachments, so we include file contents in prompt if needed
	// For now, we'll use the coordinator directly for better control
	startTime := time.Now()

	// Subscribe to messages for streaming
	messageEvents := appInstance.Messages.Subscribe(ctx)
	messageReadBytes := make(map[string]int)

	// Run coordinator in goroutine
	type response struct {
		result *fantasy.AgentResult
		err    error
	}
	done := make(chan response, 1)

	go func() {
		result, err := appInstance.AgentCoordinator.Run(ctx, sessionID, prompt, attachments...)
		done <- response{result: result, err: err}
	}()

	// Stream messages
	for {
		select {
		case result := <-done:
			if result.err != nil {
				return fmt.Errorf("review failed: %w", result.err)
			}
			// Wait a moment for final message updates
			time.Sleep(100 * time.Millisecond)
			
			// Render the full response with markdown
			fmt.Fprintln(w)
			fmt.Fprintln(w, RenderDivider(80))
			fmt.Fprintln(w)

			// Get final content from result
			finalContent := result.result.Response.Content.Text()
			renderer, err := NewRenderer()
			if err == nil {
				rendered, err := renderer.RenderMarkdown(finalContent)
				if err == nil {
					fmt.Fprintln(w, rendered)
				} else {
					fmt.Fprintln(w, finalContent)
				}
			} else {
				fmt.Fprintln(w, finalContent)
			}

			elapsed := time.Since(startTime)
			fmt.Fprintln(w)
			fmt.Fprintln(w, RenderSuccess(fmt.Sprintf("Review completed in %s", elapsed.Round(time.Millisecond))))
			return nil

		case event := <-messageEvents:
			msg := event.Payload
			if msg.SessionID == sessionID && msg.Role == message.Assistant && len(msg.Parts) > 0 {
				content := msg.Content().String()
				readBytes := messageReadBytes[msg.ID]

				if len(content) > readBytes {
					part := content[readBytes:]
					if readBytes == 0 {
						part = strings.TrimLeft(part, " \t")
					}
					fmt.Fprint(w, part)
					messageReadBytes[msg.ID] = len(content)
				}
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
