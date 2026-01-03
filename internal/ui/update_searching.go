package ui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
)

// updateKeyMsgSearching handles key messages in search mode
func (m *Model) updateKeyMsgSearching(msg tea.KeyMsg) (*Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.SearchEsc):
		// Exit search mode, keep results
		m.returnToPreviousState()
		m.searchInput.Blur()
		return m, nil
	case key.Matches(msg, m.keys.SearchEnter):
		// Confirm search and exit search mode
		m.search.Query = m.searchInput.Value()
		m.search.Search(m.rawContent)
		m.returnToPreviousState()
		m.searchInput.Blur()
		UpdateViewportWithSearch(&m.viewport, m.rawContent, m.search)
		return m, nil
	case key.Matches(msg, m.keys.ToggleMode):
		// Toggle search mode
		m.search.ToggleMode()
		UpdateViewportWithSearch(&m.viewport, m.rawContent, m.search)
		return m, nil
	default:
		// Update search input
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		// Live search as user types
		m.search.Query = m.searchInput.Value()
		m.search.Search(m.rawContent)
		UpdateViewportWithSearch(&m.viewport, m.rawContent, m.search)
		return m, cmd
	}
}

