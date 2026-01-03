package ui

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// updateKeyMsgError handles key messages in error state
func (m *Model) updateKeyMsgError(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit), key.Matches(msg, m.keys.ForceQuit):
		if m.activeCancel != nil {
			m.activeCancel()
			m.activeCancel = nil
		}
		return m, tea.Quit
	}
	return m, nil
}
