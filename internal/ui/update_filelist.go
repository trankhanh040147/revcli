package ui

import (
	"context"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// updateKeyMsgFileList handles key messages in file list mode
func (m *Model) updateKeyMsgFileList(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		m.returnToPreviousState()
		return m, nil
	case key.Matches(msg, m.keys.FileListPrune):
		// Prune the selected file
		filePath, ok := GetSelectedFile(m.fileList)
		if !ok {
			return m, nil
		}
		// Check if already pruned (PrunedFiles is always initialized in builder.go:91)
		if _, pruned := m.reviewCtx.PrunedFiles[filePath]; pruned {
			// Already pruned, skip
			return m, nil
		}
		// Get file content
		content, ok := m.reviewCtx.FileContents[filePath]
		if !ok {
			return m, nil
		}
		// Create new context for this command
		ctx, cancel := context.WithCancel(m.rootCtx)
		m.activeCancel = cancel
		// Trigger prune command
		return m, pruneFileCmd(ctx, m.flashClient, filePath, content)
	case key.Matches(msg, m.keys.SelectFile):
		// View selected file (for now, just go back)
		m.returnToPreviousState()
		return m, nil
	default:
		// Update file list
		var cmd tea.Cmd
		m.fileList, cmd = m.fileList.Update(msg)
		return m, cmd
	}
}
