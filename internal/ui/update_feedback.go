package ui

import (
	"context"
	"errors"
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/trankhanh040147/revcli/internal/agent"
)

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
		// Handle agent package specific errors
		if errors.Is(msg.Err, agent.ErrRequestCancelled) {
			m.errorMsg = "Request cancelled"
		} else if errors.Is(msg.Err, agent.ErrSessionBusy) {
			m.errorMsg = "Session is busy processing another request"
		} else if errors.Is(msg.Err, agent.ErrEmptyPrompt) {
			m.errorMsg = "Prompt is empty"
		} else if errors.Is(msg.Err, agent.ErrSessionMissing) {
			m.errorMsg = "Session is missing"
		} else {
			m.errorMsg = msg.Err.Error()
		}
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

	filePath := pruneMsg.FilePath

	// Clear pruning state for this file
	delete(m.pruningFiles, filePath)
	delete(m.pruningSpinners, filePath)
	if cancel, ok := m.pruningCancels[filePath]; ok {
		delete(m.pruningCancels, filePath)
		// Cancel function is no longer needed, but don't call it (operation is done)
		_ = cancel
	}

	if pruneMsg.Err != nil {
		// Check if error is cancellation (errors.Is works with %w wrapped errors)
		if errors.Is(pruneMsg.Err, context.Canceled) {
			m.yankFeedback = fmt.Sprintf("Pruning cancelled: %s", filePath)
		} else {
			m.yankFeedback = fmt.Sprintf("Error pruning file: %v", pruneMsg.Err)
		}
		// Update file list to remove pruning indicator
		m.fileList = UpdateFileListModel(m.fileList, m.reviewCtx, m.pruningFiles)
		return m, ClearYankFeedbackCmd(PruneErrorFeedbackDuration), true
	}
	// Update pruned files map (PrunedFiles is always initialized in builder.go)
	m.reviewCtx.PrunedFiles[filePath] = pruneMsg.Summary
	// Update file list to show pruned indicator (and remove pruning indicator)
	m.fileList = UpdateFileListModel(m.fileList, m.reviewCtx, m.pruningFiles)
	m.yankFeedback = fmt.Sprintf("✓ Pruned %s", filePath)
	return m, ClearYankFeedbackCmd(YankFeedbackDuration), true
}

