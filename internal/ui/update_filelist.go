package ui

import (
	"context"
	"fmt"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
		// Check if already pruning
		if m.pruningFiles[filePath] {
			// Already pruning, skip
			return m, nil
		}
		// Check if file exists
		if _, ok := m.reviewCtx.FileContents[filePath]; !ok {
			return m, nil
		}
		// Mark file as pruning
		m.pruningFiles[filePath] = true
		// Create spinner for this file
		fileSpinner := spinner.New()
		fileSpinner.Spinner = spinner.Dot
		fileSpinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))
		m.pruningSpinners[filePath] = fileSpinner
		// Create new context for this command
		_, cancel := context.WithCancel(m.rootCtx)
		m.pruningCancels[filePath] = cancel
		// Update file list to show pruning indicator
		m.fileList = UpdateFileListModel(m.fileList, m.reviewCtx, m.pruningFiles)
		// Start spinner tick and prune command
		return m, tea.Batch(
			fileSpinner.Tick,
			// Prune functionality temporarily disabled - needs migration to coordinator
			// pruneFileCmd(ctx, m.flashClient, filePath, content),
			tea.Cmd(func() tea.Msg {
				return PruneFileMsg{
					FilePath: filePath,
					Summary:  "",
					Err:      fmt.Errorf("prune functionality is temporarily disabled during migration"),
				}
			}),
		)
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
