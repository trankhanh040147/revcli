package ui

import (
	"context"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/key"
)

// updateKeyMsgChatting handles key messages in chatting mode
func (m *Model) updateKeyMsgChatting(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit), key.Matches(msg, m.keys.ForceQuit):
		if m.textarea.Value() == "" {
			m.cancel()
			return m, tea.Quit
		}
	case key.Matches(msg, m.keys.ExitChat):
		m.state = StateReviewing
		m.viewport.Height = CalculateViewportHeight(m.height, m.state, m.yankFeedback != "")
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
				return m, SendChatMessage(m.ctx, m.client, question)
			}
		}
	case key.Matches(msg, m.keys.PrevPrompt):
		if !m.streaming && len(m.promptHistory) > 0 {
			_, newIndex, promptText := NavigatePromptHistory(m.promptHistory, m.promptHistoryIndex, -1)
			m.promptHistoryIndex = newIndex
			if promptText != "" {
				m.textarea.SetValue(promptText)
				m.textarea.CursorEnd()
			}
		}
	case key.Matches(msg, m.keys.NextPrompt):
		if !m.streaming {
			_, newIndex, promptText := NavigatePromptHistory(m.promptHistory, m.promptHistoryIndex, 1)
			m.promptHistoryIndex = newIndex
			if promptText != "" {
				m.textarea.SetValue(promptText)
				m.textarea.CursorEnd()
			} else {
				m.textarea.SetValue("")
			}
		}
	case key.Matches(msg, m.keys.CancelRequest):
		if m.streaming {
			m.cancel()
			m.streaming = false
			m.ctx, m.cancel = context.WithCancel(context.Background())
			m.yankFeedback = "Request cancelled"
			m.viewport.Height = CalculateViewportHeight(m.height, m.state, m.yankFeedback != "")
			return m, ClearYankFeedbackCmd(2 * time.Second)
		}
	}
	return m, nil
}

