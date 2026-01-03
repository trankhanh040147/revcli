package ui

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

// handleSpinnerTick handles spinner animation tick messages
func (m *Model) handleSpinnerTick(msg spinner.TickMsg) tea.Cmd {
	var cmds []tea.Cmd

	// Handle main spinner (loading/streaming)
	if m.state == StateLoading || m.streaming {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Handle pruning spinners (update all active ones)
	for filePath, fileSpinner := range m.pruningSpinners {
		var cmd tea.Cmd
		updatedSpinner, cmd := fileSpinner.Update(msg)
		m.pruningSpinners[filePath] = updatedSpinner
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if len(cmds) > 0 {
		return tea.Batch(cmds...)
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

