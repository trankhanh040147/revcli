package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/key"
)

// updateKeyMsgHelp handles key messages in help mode
func (m *Model) updateKeyMsgHelp(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Help), key.Matches(msg, m.keys.SearchEsc):
		m.returnToPreviousState()
		return m, nil
	}
	return m, nil
}

