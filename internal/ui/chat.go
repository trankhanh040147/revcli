package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/trankhanh040147/revcli/internal/gemini"
	"github.com/trankhanh040147/revcli/internal/prompt"
)

// SendChatMessage sends a follow-up question
func SendChatMessage(ctx context.Context, client *gemini.Client, question string) tea.Cmd {
	return func() tea.Msg {
		followUp := prompt.BuildFollowUpPrompt(question)

		response, err := client.SendMessage(ctx, followUp)
		if err != nil {
			return ChatErrorMsg{Err: err}
		}

		return ChatResponseMsg{Response: response}
	}
}

// UpdatePromptHistory adds a question to prompt history, avoiding duplicates
func UpdatePromptHistory(history []string, question string) []string {
	if len(history) == 0 || history[len(history)-1] != question {
		return append(history, question)
	}
	return history
}

// NavigatePromptHistory navigates through prompt history
// Returns updated history, new index, and the prompt at that index (empty if new prompt)
func NavigatePromptHistory(history []string, currentIndex int, direction int) ([]string, int, string) {
	if len(history) == 0 {
		return history, -1, ""
	}

	newIndex := currentIndex
	if direction < 0 {
		// Previous (going backwards)
		if currentIndex == -1 {
			// Start from last prompt
			newIndex = len(history) - 1
		} else if currentIndex > 0 {
			newIndex = currentIndex - 1
		}
	} else if direction > 0 {
		// Next (going forwards)
		if currentIndex >= 0 {
			newIndex = currentIndex + 1
			if newIndex >= len(history) {
				// Beyond last, return empty for new prompt
				return history, -1, ""
			}
		}
	}

	if newIndex >= 0 && newIndex < len(history) {
		return history, newIndex, history[newIndex]
	}

	return history, -1, ""
}
