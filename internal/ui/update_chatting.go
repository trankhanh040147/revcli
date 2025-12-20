package ui

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// navigatePromptHistory navigates through prompt history in the given direction
func (m *Model) navigatePromptHistory(direction int) {
	if m.streaming {
		return
	}
	if direction < 0 && len(m.promptHistory) == 0 {
		return
	}
	_, newIndex, promptText := NavigatePromptHistory(m.promptHistory, m.promptHistoryIndex, direction)
	m.promptHistoryIndex = newIndex
	if promptText != "" {
		m.textarea.SetValue(promptText)
		m.textarea.CursorEnd()
	} else if direction > 0 {
		m.textarea.SetValue("")
	}
}

// updateKeyMsgChatting handles key messages in chatting mode
func (m *Model) updateKeyMsgChatting(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit), key.Matches(msg, m.keys.ForceQuit):
		if m.textarea.Value() == "" {
			if m.activeCancel != nil {
				m.activeCancel()
				m.activeCancel = nil
			}
			return m, tea.Quit
		}
	case key.Matches(msg, m.keys.ExitChat):
		m.state = StateReviewing
		m.updateViewportHeight()
		return m, nil
	case key.Matches(msg, m.keys.SendMessage):
		if !m.streaming {
			question := strings.TrimSpace(m.textarea.Value())
			if question != "" {
				m.promptHistory = UpdatePromptHistory(m.promptHistory, question)
				m.promptHistoryIndex = -1
				m.textarea.Reset()
				m.streaming = true
				m.chatHistory = append(m.chatHistory, ChatMessage{Role: ChatRoleUser, Content: question})
				// Capture current webSearchEnabled setting for this request, then reset for next question
				webSearchEnabled := m.webSearchEnabled
				m.webSearchEnabled = true // Reset to default (true) for next question
				// Create new context for this command
				ctx, cancel := context.WithCancel(m.rootCtx)
				m.activeCancel = cancel
				return m, SendChatMessage(ctx, m.client, question, webSearchEnabled)
			}
		}
	case key.Matches(msg, m.keys.PrevPrompt):
		m.navigatePromptHistory(-1)
	case key.Matches(msg, m.keys.NextPrompt):
		m.navigatePromptHistory(1)
	case key.Matches(msg, m.keys.CancelRequest):
		if m.streaming {
			if m.activeCancel != nil {
				m.activeCancel()
				m.activeCancel = nil
			}
			m.resetStreamState()
			m.yankFeedback = "Request cancelled"
			m.updateViewportHeight()
			return m, ClearYankFeedbackCmd(YankFeedbackDuration)
		}
	case key.Matches(msg, m.keys.ToggleWebSearch):
		if !m.streaming {
			m.webSearchEnabled = !m.webSearchEnabled
			return m, nil
		}
	default:
		// Pass through typing keys to textarea when not streaming
		if !m.streaming {
			var cmd tea.Cmd
			m.textarea, cmd = m.textarea.Update(msg)
			return m, cmd
		}
		return m, nil
	}
	return m, nil
}
