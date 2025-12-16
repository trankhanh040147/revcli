package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// updateKeyMsgSearching handles key messages in search mode
func (m *Model) updateKeyMsgSearching(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.SearchEsc):
		// Exit search mode, keep results
		m.state = m.previousState
		m.searchInput.Blur()
		m.viewport.Height = CalculateViewportHeight(m.height, m.state, m.yankFeedback != "")
		return m, nil
	case key.Matches(msg, m.keys.SearchEnter):
		// Confirm search and exit search mode
		m.search.Query = m.searchInput.Value()
		m.search.Search(m.rawContent)
		m.state = m.previousState
		m.searchInput.Blur()
		m.viewport.Height = CalculateViewportHeight(m.height, m.state, m.yankFeedback != "")
		UpdateViewportWithSearch(&m.viewport, m.rawContent, m.search)
		return m, nil
	case key.Matches(msg, m.keys.ToggleMode):
		// Toggle search mode
		m.search.ToggleMode()
		UpdateViewportWithSearch(&m.viewport, m.rawContent, m.search)
		return m, nil
	default:
		// Update search input
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		// Live search as user types
		m.search.Query = m.searchInput.Value()
		m.search.Search(m.rawContent)
		UpdateViewportWithSearch(&m.viewport, m.rawContent, m.search)
		return m, cmd
	}
}

// handleYank handles yank commands and returns (model, cmd, handled)
func (m *Model) handleYank(msg tea.KeyMsg) (*Model, tea.Cmd, bool) {
	switch {
	case key.Matches(msg, m.keys.YankReview):
		if m.state == StateReviewing && !m.lastKeyWasY {
			// First 'y' press - wait to see if followed by another 'y' (yy)
			m.lastKeyWasY = true
			return m, tea.Tick(300*time.Millisecond, func(t time.Time) tea.Msg {
				return yankTimeoutMsg{}
			}), true
		} else if m.lastKeyWasY {
			// Second 'y' press (yy) - yank entire review
			m.lastKeyWasY = false
			return m, YankReview(m.reviewResponse, m.chatHistory), true
		}
	case key.Matches(msg, m.keys.YankLast):
		// Yank only the last/most recent response
		if m.state == StateReviewing {
			m.lastKeyWasY = false
			return m, YankLastResponse(m.reviewResponse, m.chatHistory), true
		}
		m.lastKeyWasY = false
		return m, nil, true
	}
	return m, nil, false
}

// handleNavigation handles navigation commands
func (m *Model) handleNavigation(msg tea.KeyMsg) (*Model, tea.Cmd) {
	if m.state != StateReviewing {
		return m, nil
	}

	switch {
	case key.Matches(msg, m.keys.Down):
		m.viewport.LineDown(1)
	case key.Matches(msg, m.keys.Up):
		m.viewport.LineUp(1)
	case key.Matches(msg, m.keys.Top):
		m.viewport.GotoTop()
	case key.Matches(msg, m.keys.Bottom):
		m.viewport.GotoBottom()
	case key.Matches(msg, m.keys.HalfPageDown):
		m.viewport.HalfViewDown()
	case key.Matches(msg, m.keys.HalfPageUp):
		m.viewport.HalfViewUp()
	case key.Matches(msg, m.keys.PageDown):
		m.viewport.ViewDown()
	case key.Matches(msg, m.keys.PageUp):
		m.viewport.ViewUp()
	}
	return m, nil
}

// handleSearchNavigation handles search navigation commands
func (m *Model) handleSearchNavigation(msg tea.KeyMsg) (*Model, tea.Cmd) {
	if m.state != StateReviewing || m.search.Query == "" {
		return m, nil
	}

	switch {
	case key.Matches(msg, m.keys.NextMatch):
		m.search.NextMatch()
		UpdateViewportWithSearch(&m.viewport, m.rawContent, m.search)
		ScrollToCurrentMatch(&m.viewport, m.search)
		m.resetYankChord()
	case key.Matches(msg, m.keys.PrevMatch):
		m.search.PrevMatch()
		UpdateViewportWithSearch(&m.viewport, m.rawContent, m.search)
		ScrollToCurrentMatch(&m.viewport, m.search)
		m.resetYankChord()
	}
	return m, nil
}

// updateKeyMsgReviewing handles key messages in reviewing mode
func (m *Model) updateKeyMsgReviewing(msg tea.KeyMsg) (*Model, tea.Cmd) {
	// Check for yank commands first (they handle their own state)
	if newM, cmd, handled := m.handleYank(msg); handled {
		return newM, cmd
	}

	switch {
	case key.Matches(msg, m.keys.Quit), key.Matches(msg, m.keys.ForceQuit):
		m.cancel()
		return m, tea.Quit
	case key.Matches(msg, m.keys.Search):
		m.previousState = m.state
		m.state = StateSearching
		m.searchInput.SetValue(m.search.Query)
		m.searchInput.Focus()
		m.viewport.Height = CalculateViewportHeight(m.height, m.state, m.yankFeedback != "")
		return m, textinput.Blink
	case key.Matches(msg, m.keys.Help):
		m.previousState = m.state
		m.state = StateHelp
		return m, nil
	case key.Matches(msg, m.keys.EnterChat):
		m.state = StateChatting
		m.textarea.Focus()
		m.viewport.Height = CalculateViewportHeight(m.height, m.state, m.yankFeedback != "")
		return m, nil
	default:
		// Try navigation
		m.handleNavigation(msg)
		// Try search navigation
		m.handleSearchNavigation(msg)
		// Reset yank chord for any other key
		m.resetYankChord()
	}
	return m, nil
}

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

// updateKeyMsgHelp handles key messages in help mode
func (m *Model) updateKeyMsgHelp(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Help), key.Matches(msg, m.keys.SearchEsc):
		m.state = m.previousState
		m.viewport.Height = CalculateViewportHeight(m.height, m.state, m.yankFeedback != "")
		return m, nil
	}
	return m, nil
}

// updateNonKeyMsg handles non-key messages
func (m *Model) updateNonKeyMsg(msg tea.Msg) (*Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		viewportHeight := CalculateViewportHeight(m.height, m.state, m.yankFeedback != "")
		if !m.ready {
			m.viewport = viewport.New(msg.Width, viewportHeight)
			m.viewport.Style = lipgloss.NewStyle().Padding(0, 2)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = viewportHeight
		}
		m.textarea.SetWidth(msg.Width - 4)

	case spinner.TickMsg:
		if m.state == StateLoading || m.streaming {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case ReviewCompleteMsg:
		m.state = StateReviewing
		m.reviewResponse = msg.Response
		m.streaming = false
		m.updateViewport()

	case ReviewErrorMsg:
		m.state = StateError
		m.errorMsg = msg.Err.Error()

	case ChatResponseMsg:
		m.streaming = false
		m.chatHistory = append(m.chatHistory, ChatMessage{Role: ChatRoleAssistant, Content: msg.Response})
		m.updateViewportAndScroll()

	case ChatErrorMsg:
		m.streaming = false
		m.errorMsg = msg.Err.Error()
		m.chatHistory = append(m.chatHistory, ChatMessage{Role: ChatRoleAssistant, Content: "Error: " + msg.Err.Error()})
		m.updateViewportAndScroll()

	case yankTimeoutMsg:
		// Timeout for yank combo - if lastKeyWasY is still true, yank entire review
		if m.lastKeyWasY {
			m.lastKeyWasY = false
			return m, YankReview(m.reviewResponse, m.chatHistory)
		}

	case YankMsg:
		m.yankFeedback = fmt.Sprintf("✓ Copied %s to clipboard!", msg.Type)
		m.viewport.Height = CalculateViewportHeight(m.height, m.state, m.yankFeedback != "")
		return m, ClearYankFeedbackCmd(2 * time.Second)

	case YankFeedbackMsg:
		m.yankFeedback = ""
		m.viewport.Height = CalculateViewportHeight(m.height, m.state, m.yankFeedback != "")
	}

	// Update textarea in chat mode
	if m.state == StateChatting && !m.streaming {
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Update viewport
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Route to state-specific handlers
		switch m.state {
		case StateSearching:
			return m.updateKeyMsgSearching(msg)
		case StateReviewing:
			return m.updateKeyMsgReviewing(msg)
		case StateChatting:
			return m.updateKeyMsgChatting(msg)
		case StateHelp:
			return m.updateKeyMsgHelp(msg)
		default:
			// Loading/Error states don't handle keys
			return m.updateNonKeyMsg(msg)
		}
	default:
		return m.updateNonKeyMsg(msg)
	}
}
