package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	appcontext "github.com/trankhanh040147/revcli/internal/context"
)

// startReview initiates the code review
func (m *Model) startReview() tea.Cmd {
	return func() tea.Msg {
		// Initialize chat with system prompt (with preset if specified)
		systemPrompt := appcontext.GetSystemPrompt()
		if m.preset != nil {
			systemPrompt = appcontext.GetSystemPromptWithPreset(m.preset.Prompt, m.preset.Replace)
		}
		m.client.StartChat(systemPrompt)

		// Send the review request with streaming
		var fullResponse strings.Builder

		_, err := m.client.SendMessageStream(m.ctx, m.reviewCtx.UserPrompt, func(chunk string) {
			fullResponse.WriteString(chunk)
		})
		if err != nil {
			return ReviewErrorMsg{Err: err}
		}

		return ReviewCompleteMsg{Response: fullResponse.String()}
	}
}
