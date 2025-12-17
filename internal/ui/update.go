package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// returnToPreviousState returns to the previous state and updates viewport height
func (m *Model) returnToPreviousState() {
	m.state = m.previousState
	m.viewport.Height = CalculateViewportHeight(m.height, m.state, m.yankFeedback != "")
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
		// Update file list dimensions
		m.fileList.SetWidth(msg.Width - 4)
		m.fileList.SetHeight(msg.Height - 4)

	case spinner.TickMsg:
		if m.state == StateLoading || m.streaming {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

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
		)

	case StreamChunkMsg:
		// Append chunk to response and update viewport incrementally
		m.reviewResponse += msg.Chunk
		m.updateViewport()
		// Continue listening for more chunks, errors, and completion
		return m, tea.Batch(
			streamChunkCmd(m.streamChunkChan),
			streamErrorCmd(m.streamErrChan),
			streamDoneCmd(m.streamDoneChan),
		)

	case StreamDoneMsg:
		// Streaming complete: set final response and transition to reviewing state
		m.state = StateReviewing
		m.reviewResponse = msg.FullResponse
		m.streaming = false
		// Clear streaming channels
		m.streamChunkChan = nil
		m.streamErrChan = nil
		m.streamDoneChan = nil
		m.updateViewport()

	case ReviewCompleteMsg:
		m.state = StateReviewing
		m.reviewResponse = msg.Response
		m.streaming = false
		m.updateViewport()

	case ReviewErrorMsg:
		m.state = StateError
		m.errorMsg = msg.Err.Error()
		m.streaming = false
		// Clear streaming channels
		m.streamChunkChan = nil
		m.streamErrChan = nil
		m.streamDoneChan = nil

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

	case PruneFileMsg:
		if msg.Err != nil {
			m.yankFeedback = fmt.Sprintf("Error pruning file: %v", msg.Err)
			return m, ClearYankFeedbackCmd(3 * time.Second)
		}
		// Update pruned files map (PrunedFiles is always initialized in builder.go)
		m.reviewCtx.PrunedFiles[msg.FilePath] = msg.Summary
		// Update file list to show pruned indicator
		m.fileList = UpdateFileListModel(m.fileList, m.reviewCtx)
		m.yankFeedback = fmt.Sprintf("✓ Pruned %s", msg.FilePath)
		return m, ClearYankFeedbackCmd(2 * time.Second)
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
		default:
			// Loading/Error states don't handle keys
			return m.updateNonKeyMsg(msg)
		}
	default:
		return m.updateNonKeyMsg(msg)
	}
}
