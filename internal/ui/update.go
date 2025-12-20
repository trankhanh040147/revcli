package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
		m.viewport = viewport.New(msg.Width, m.viewport.Height)
		m.viewport.Style = lipgloss.NewStyle().Padding(0, 2)
		m.ready = true
	} else {
		m.viewport.Width = msg.Width
	}
	m.textarea.SetWidth(msg.Width - 4)
	// Update file list dimensions
	m.fileList.SetWidth(msg.Width - 4)
	m.fileList.SetHeight(msg.Height - 4)
}

// handleSpinnerTick handles spinner animation tick messages
func (m *Model) handleSpinnerTick(msg spinner.TickMsg) tea.Cmd {
	if m.state == StateLoading || m.streaming {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return cmd
	}
	return nil
}

// handleStreamMessages handles streaming-related messages
// Returns (model, cmd, shouldReturnEarly)
func (m *Model) handleStreamMessages(msg tea.Msg) (*Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case StreamStartMsg:
		// Start streaming: store channels and begin listening
		m.streaming = true
		m.reviewResponse = "" // Reset response buffer
		m.streamChunkChan = msg.ChunkChan
		m.streamErrChan = msg.ErrChan
		m.streamDoneChan = msg.DoneChan
		// Return commands to listen for chunks, errors, and completion
		return m, tea.Batch(
			streamChunkCmd(m.streamChunkChan),
			streamErrorCmd(m.streamErrChan),
			streamDoneCmd(m.streamDoneChan),
		), true

	case StreamChunkMsg:
		// Append chunk to response and update viewport incrementally
		m.reviewResponse += msg.Chunk
		m.updateViewport()
		// Continue listening for more chunks, errors, and completion
		return m, tea.Batch(
			streamChunkCmd(m.streamChunkChan),
			streamErrorCmd(m.streamErrChan),
			streamDoneCmd(m.streamDoneChan),
		), true

	case StreamDoneMsg:
		// Streaming complete: set final response and transition to reviewing state
		m.state = StateReviewing
		m.reviewResponse = msg.FullResponse
		m.resetStreamState()
		// Clear active cancel (command completed)
		m.activeCancel = nil
		m.updateViewport()
		return m, nil, false
	}
	return m, nil, false
}

// handleReviewMessages handles review completion and error messages
func (m *Model) handleReviewMessages(msg tea.Msg) {
	switch msg := msg.(type) {
	case ReviewCompleteMsg:
		m.state = StateReviewing
		m.reviewResponse = msg.Response
		m.streaming = false
		// Clear active cancel (command completed)
		m.activeCancel = nil
		m.updateViewport()

	case ReviewErrorMsg:
		m.state = StateError
		m.errorMsg = msg.Err.Error()
		m.resetStreamState()
		// Clear active cancel (command completed/errored)
		m.activeCancel = nil
	}
}

// handleChatMessages handles chat response and error messages
func (m *Model) handleChatMessages(msg tea.Msg) {
	switch msg := msg.(type) {
	case ChatResponseMsg:
		// Clear active cancel (command completed)
		m.activeCancel = nil
		m.handleChatCompletion(msg.Response, false)

	case ChatErrorMsg:
		// Clear active cancel (command errored)
		m.activeCancel = nil
		m.errorMsg = msg.Err.Error()
		m.handleChatCompletion("Error: "+msg.Err.Error(), true)
	}
}

// handleYankMessages handles yank-related messages
// Returns (model, cmd, shouldReturnEarly)
func (m *Model) handleYankMessages(msg tea.Msg) (*Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case yankTimeoutMsg:
		// Timeout for yank combo - if lastKeyWasY is still true, yank entire review
		if m.lastKeyWasY {
			m.lastKeyWasY = false
			return m, YankReview(m.reviewResponse, m.chatHistory), true
		}
		return m, nil, false

	case YankMsg:
		m.yankFeedback = fmt.Sprintf("✓ Copied %s to clipboard!", msg.Type)
		m.updateViewportHeight()
		return m, ClearYankFeedbackCmd(YankFeedbackDuration), true

	case YankFeedbackMsg:
		m.yankFeedback = ""
		m.updateViewportHeight()
		return m, nil, false
	}
	return m, nil, false
}

// handlePruneMessages handles file pruning messages
// Returns (model, cmd, shouldReturnEarly)
func (m *Model) handlePruneMessages(msg tea.Msg) (*Model, tea.Cmd, bool) {
	pruneMsg, ok := msg.(PruneFileMsg)
	if !ok {
		return m, nil, false
	}

	// Clear active cancel (command completed)
	m.activeCancel = nil

	if pruneMsg.Err != nil {
		m.yankFeedback = fmt.Sprintf("Error pruning file: %v", pruneMsg.Err)
		return m, ClearYankFeedbackCmd(PruneErrorFeedbackDuration), true
	}
	// Update pruned files map (PrunedFiles is always initialized in builder.go)
	m.reviewCtx.PrunedFiles[pruneMsg.FilePath] = pruneMsg.Summary
	// Update file list to show pruned indicator
	m.fileList = UpdateFileListModel(m.fileList, m.reviewCtx)
	m.yankFeedback = fmt.Sprintf("✓ Pruned %s", pruneMsg.FilePath)
	return m, ClearYankFeedbackCmd(YankFeedbackDuration), true
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
		// Handle cancel request globally when streaming (including during loading)
		if key.Matches(msg, m.keys.CancelRequest) && m.streaming {
			if m.activeCancel != nil {
				m.activeCancel()
				m.activeCancel = nil
			}
			m.transitionToErrorOnCancel()
			return m, nil
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
