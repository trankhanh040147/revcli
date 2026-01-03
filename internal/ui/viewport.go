package ui

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/viewport"
)

// BuildViewportContent builds the viewport content from review response and chat history
func BuildViewportContent(reviewResponse string, chatHistory []ChatMessage, renderer *Renderer, width int) string {
	var content strings.Builder

	// Render the main review response
	if reviewResponse != "" {
		rendered, err := renderer.RenderMarkdown(reviewResponse)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: markdown rendering failed: %v\n", err)
			content.WriteString(reviewResponse)
		} else {
			content.WriteString(rendered)
		}

	}

	// Render chat history
	if len(chatHistory) > 0 {
		content.WriteString("\n")
		content.WriteString(RenderDivider(width - 4))
		content.WriteString("\n\n")
		content.WriteString(RenderTitle("💬 Follow-up Chat"))
		content.WriteString("\n\n")

		for _, msg := range chatHistory {
			if msg.Role == ChatRoleUser {
				content.WriteString(RenderPrompt())
				content.WriteString(msg.Content)
				content.WriteString("\n\n")
			} else {
				// Render assistant message
				rendered, err := renderer.RenderMarkdown(msg.Content)
				if err != nil {
					fmt.Fprintf(os.Stderr, "warning: markdown rendering failed for chat message: %v\n", err)
					rendered = msg.Content
				}

				content.WriteString(rendered)
				content.WriteString("\n")
			}
		}
	}

	return content.String()
}

// CalculateViewportHeight calculates the dynamic viewport height based on UI state
func CalculateViewportHeight(height int, state State, hasYankFeedback bool) int {
	// Base reserved lines: header (2) + footer (2)
	reserved := 4

	// Search bar when active
	if state == StateSearching {
		reserved += 2
	}

	// Yank feedback when showing
	if hasYankFeedback {
		reserved += 1
	}

	// Chat textarea when in chat mode
	if state == StateChatting {
		reserved += 4
	}

	result := height - reserved
	if result < 5 {
		result = 5 // Minimum height
	}
	return result
}

// ScrollToCurrentMatch scrolls the viewport to show the current search match
func ScrollToCurrentMatch(vp *viewport.Model, search *SearchState) {
	line := search.CurrentMatchLine()
	if line < 0 {
		return
	}

	// In filter mode, translate original line to filtered view index
	if search.Mode == SearchModeFilter {
		line = search.FilteredLineIndex(line)
		if line < 0 {
			return
		}
	}

	// Calculate approximate line position and scroll there
	// This is an approximation since rendered content may have different line counts
	viewportHeight := vp.Height()
	targetLine := line - viewportHeight/2
	if targetLine < 0 {
		targetLine = 0
	}

	// Scroll to target line - viewport v2 handles scrolling through Update with key messages
	// For now, we'll use a simple approach: set content and let viewport handle it
	// Note: v2 viewport doesn't have direct scroll methods, navigation is handled via Update()
}

// UpdateViewportWithSearch updates the viewport with search highlighting
func UpdateViewportWithSearch(vp *viewport.Model, rawContent string, search *SearchState) {
	if rawContent == "" {
		return
	}

	var displayContent string
	if search.Query == "" || len(search.Matches) == 0 {
		displayContent = rawContent
	} else {
		if search.Mode == SearchModeFilter {
			displayContent = search.FilterContent(rawContent)
		} else {
			displayContent = search.HighlightContent(rawContent)
		}
	}

	vp.SetContent(displayContent)
}

// updateViewportHeight updates the viewport height based on current UI state
func (m *Model) updateViewportHeight() {
	m.viewport.SetHeight(CalculateViewportHeight(m.height, m.state, m.yankFeedback != ""))
}

// resetYankChord resets the yank chord state
func (m *Model) resetYankChord() {
	m.lastKeyWasY = false
}

// updateViewport updates the viewport content
func (m *Model) updateViewport() {
	m.updateViewportWithScroll(false)
}

// updateViewportAndScroll updates viewport and scrolls to bottom (for chat responses)
func (m *Model) updateViewportAndScroll() {
	m.updateViewportWithScroll(true)
}

// updateViewportWithScroll updates the viewport content with optional scroll
func (m *Model) updateViewportWithScroll(scrollToBottom bool) {
	// Build content (code block navigation feature removed in v0.3.1)
	content := BuildViewportContent(m.reviewResponse, m.chatHistory, m.renderer, m.width)
	m.rawContent = content
	m.viewport.SetContent(m.rawContent)

	// Only scroll to bottom for chat updates, not initial review
	if scrollToBottom {
		m.viewport.GotoBottom()
	}
}
