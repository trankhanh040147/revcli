package ui

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

// YankReview yanks the entire review content to clipboard
func YankReview(reviewResponse string, chatHistory []ChatMessage) tea.Cmd {
	return func() tea.Msg {
		var content strings.Builder

		// Use raw review response (markdown without ANSI codes)
		if reviewResponse != "" {
			content.WriteString(reviewResponse)
		}

		// Add raw chat history
		if len(chatHistory) > 0 {
			content.WriteString("\n\n---\n\n## Follow-up Chat\n\n")
			for _, msg := range chatHistory {
				if msg.Role == ChatRoleUser {
					content.WriteString("**You:** ")
					content.WriteString(msg.Content)
				} else {
					content.WriteString("**Assistant:**\n")
					content.WriteString(msg.Content)
				}
				content.WriteString("\n\n")
			}
		}

		result := content.String()
		if result == "" {
			return nil
		}

		err := clipboard.WriteAll(result)
		if err != nil {
			return ChatErrorMsg{Err: fmt.Errorf("failed to copy to clipboard: %w", err)}
		}

		return YankMsg{Type: YankTypeReview}
	}
}

// YankLastResponse yanks only the last/most recent assistant response
func YankLastResponse(reviewResponse string, chatHistory []ChatMessage) tea.Cmd {
	return func() tea.Msg {
		var content string

		// If there's chat history, get the last assistant message
		if len(chatHistory) > 0 {
			for i := len(chatHistory) - 1; i >= 0; i-- {
				if chatHistory[i].Role == ChatRoleAssistant {
					content = chatHistory[i].Content
					break
				}
			}
		}

		// If no chat history or no assistant message found, use the initial review
		if content == "" {
			content = reviewResponse
		}

		if content == "" {
			return nil
		}

		err := clipboard.WriteAll(content)
		if err != nil {
			return ChatErrorMsg{Err: fmt.Errorf("failed to copy to clipboard: %w", err)}
		}

		return YankMsg{Type: YankTypeLastResponse}
	}
}
