package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// handleYank handles yank commands and returns (model, cmd, handled)
func (m *Model) handleYank(msg tea.KeyMsg) (*Model, tea.Cmd, bool) {
	switch {
	case key.Matches(msg, m.keys.YankReview):
		if m.state == StateReviewing && !m.lastKeyWasY {
			// First 'y' press - wait to see if followed by another 'y' (yy)
			m.lastKeyWasY = true
			return m, tea.Tick(YankChordTimeout, func(t time.Time) tea.Msg {
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
		if m.activeCancel != nil {
			m.activeCancel()
			m.activeCancel = nil
		}
		return m, tea.Quit
	case key.Matches(msg, m.keys.CancelRequest):
		if m.streaming && m.activeCancel != nil {
			m.activeCancel()
			m.activeCancel = nil
		}
		m.transitionToErrorOnCancel()
		return m, nil
	case key.Matches(msg, m.keys.Search):
		m.previousState = m.state
		m.state = StateSearching
		m.searchInput.SetValue(m.search.Query)
		m.searchInput.Focus()
		m.updateViewportHeight()
		return m, textinput.Blink
	case key.Matches(msg, m.keys.Help):
		m.previousState = m.state
		m.state = StateHelp
		return m, nil
	case key.Matches(msg, m.keys.EnterChat):
		m.state = StateChatting
		m.textarea.Focus()
		m.updateViewportHeight()
		return m, nil
	case key.Matches(msg, m.keys.FileList):
		m.previousState = m.state
		m.state = StateFileList
		// Update file list with current pruned state
		m.fileList = UpdateFileListModel(m.fileList, m.reviewCtx)
		// Set file list dimensions
		m.fileList.SetWidth(m.width - 4)
		m.fileList.SetHeight(m.height - 4)
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
