package ui

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// returnToPreviousState returns to the previous state and updates viewport height
func (m *Model) returnToPreviousState() {
	m.state = m.previousState
	m.updateViewportHeight()
}

// handleChatCompletion handles chat response completion (success or error)
// Note: webSearchEnabled is reset in update_chatting.go when sending the message,
// so it's already set to true when this handler runs
func (m *Model) handleChatCompletion(content string, isError bool) {
	m.streaming = false
	m.chatHistory = append(m.chatHistory, ChatMessage{Role: ChatRoleAssistant, Content: content})
	m.textarea.Focus() // Ensure textarea remains focused after response/error
	m.updateViewportAndScroll()
}

// handleWindowSize handles window resize messages
func (m *Model) handleWindowSize(msg tea.WindowSizeMsg) {
	m.width = msg.Width
	m.height = msg.Height

	m.updateViewportHeight()
	if !m.ready {
		m.viewport = viewport.New()
		m.viewport.SetWidth(msg.Width)
		m.viewport.SetHeight(CalculateViewportHeight(msg.Height, m.state, m.yankFeedback != ""))
		m.viewport.Style = lipgloss.NewStyle().Padding(0, 2)
		m.ready = true
	} else {
		m.viewport.SetWidth(msg.Width)
	}
	m.textarea.SetWidth(msg.Width - 4)
	// Update file list dimensions
	m.fileList.SetWidth(msg.Width - 4)
	m.fileList.SetHeight(msg.Height - 4)
}

// updateNonKeyMsg handles non-key messages
func (m *Model) updateNonKeyMsg(msg tea.Msg) (*Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle window size
	if windowMsg, ok := msg.(tea.WindowSizeMsg); ok {
		m.handleWindowSize(windowMsg)
	}

	// Handle spinner tick
	if tickMsg, ok := msg.(spinner.TickMsg); ok {
		if cmd := m.handleSpinnerTick(tickMsg); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Handle stream messages (may return early)
	if newM, cmd, shouldReturn := m.handleStreamMessages(msg); shouldReturn {
		return newM, cmd
	}

	// Handle review messages
	m.handleReviewMessages(msg)

	// Handle chat messages
	m.handleChatMessages(msg)

	// Handle yank messages (may return early)
	if newM, cmd, shouldReturn := m.handleYankMessages(msg); shouldReturn {
		return newM, cmd
	}

	// Handle prune messages (may return early)
	if newM, cmd, shouldReturn := m.handlePruneMessages(msg); shouldReturn {
		return newM, cmd
	}

	// Update file list in file list mode
	if m.state == StateFileList {
		var cmd tea.Cmd
		m.fileList, cmd = m.fileList.Update(msg)
		cmds = append(cmds, cmd)
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
		// Handle quit keys globally (including during loading/streaming)
		if key.Matches(msg, m.keys.Quit) || key.Matches(msg, m.keys.ForceQuit) {
			if m.activeCancel != nil {
				m.activeCancel()
				m.activeCancel = nil
			}
			return m, tea.Quit
		}
		// Handle cancel request globally for all long-running operations
		if key.Matches(msg, m.keys.CancelRequest) {
			// Cancel main active operation
			if m.activeCancel != nil {
				m.activeCancel()
				m.activeCancel = nil
			}
			// Cancel all active pruning operations
			for filePath, cancel := range m.pruningCancels {
				cancel()
				delete(m.pruningCancels, filePath)
				delete(m.pruningFiles, filePath)
				delete(m.pruningSpinners, filePath)
			}
			// Update file list to remove pruning indicators
			if m.state == StateFileList {
				m.fileList = UpdateFileListModel(m.fileList, m.reviewCtx, m.pruningFiles)
			}
			// Handle cancellation based on state
			if m.state == StateLoading || m.streaming {
				// For loading/streaming, transition to error state
				m.transitionToErrorOnCancel()
				return m, nil
			} else if m.state == StateFileList {
				// For file list (pruning), show feedback but stay in file list
				m.yankFeedback = "Pruning cancelled"
				m.updateViewportHeight()
				return m, ClearYankFeedbackCmd(YankFeedbackDuration)
			}
			// If no active operation, just show feedback
			m.yankFeedback = "No active operation to cancel"
			m.updateViewportHeight()
			return m, ClearYankFeedbackCmd(YankFeedbackDuration)
		}
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
		case StateFileList:
			return m.updateKeyMsgFileList(msg)
		case StateError:
			return m.updateKeyMsgError(msg)
		default:
			// Loading state doesn't handle other keys
			return m.updateNonKeyMsg(msg)
		}
	default:
		return m.updateNonKeyMsg(msg)
	}
}
