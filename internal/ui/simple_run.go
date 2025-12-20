package ui

import (
	"context"
	"fmt"
	"io"
	"time"

	appcontext "github.com/trankhanh040147/revcli/internal/context"
	"github.com/trankhanh040147/revcli/internal/gemini"
	"github.com/trankhanh040147/revcli/internal/preset"
)

// RunSimple runs a simple non-interactive review
func RunSimple(ctx context.Context, w io.Writer, reviewCtx *appcontext.ReviewContext, client *gemini.Client, p *preset.Preset) error {
	// Initialize chat with system prompt (with preset and intent if specified)
	var presetPrompt string
	var presetReplace bool
	if p != nil {
		presetPrompt = p.Prompt
		presetReplace = p.Replace
	}
	systemPrompt := appcontext.GetSystemPromptWithIntent(reviewCtx.Intent, presetPrompt, presetReplace)
	client.StartChat(systemPrompt)

	fmt.Fprintln(w, RenderTitle("🔍 Code Review"))
	fmt.Fprintln(w, RenderSubtitle(reviewCtx.Summary()))
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Analyzing your code changes...")
	fmt.Fprintln(w)

	// Create renderer
	renderer, err := NewRenderer()
	if err != nil {
		return fmt.Errorf("failed to create renderer: %w", err)
	}

	// Get web search setting from intent (default to true if intent is nil)
	webSearchEnabled := true
	if reviewCtx.Intent != nil {
		webSearchEnabled = reviewCtx.Intent.WebSearchEnabled
	}

	// Stream the response
	startTime := time.Now()
	response, err := client.SendMessageStream(ctx, reviewCtx.UserPrompt, func(chunk string) {
		fmt.Fprint(w, chunk)
	}, webSearchEnabled)
	if err != nil {
		return fmt.Errorf("review failed: %w", err)
	}

	// Render the full response with markdown
	fmt.Fprintln(w)
	fmt.Fprintln(w, RenderDivider(80))
	fmt.Fprintln(w)

	rendered, err := renderer.RenderMarkdown(response)
	if err != nil {
		fmt.Fprintln(w, response)
	} else {
		fmt.Fprintln(w, rendered)
	}

	elapsed := time.Since(startTime)
	fmt.Fprintln(w)
	fmt.Fprintln(w, RenderSuccess(fmt.Sprintf("Review completed in %s", elapsed.Round(time.Millisecond))))

	// Display token usage
	if usage := client.GetLastUsage(); usage != nil {
		fmt.Fprintln(w, RenderTokenUsage(usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens))
	}

	return nil
}
