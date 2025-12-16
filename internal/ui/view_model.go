package ui

import (
	"fmt"
	"strings"
)

// viewLoading renders the loading state
func (m *Model) viewLoading() string {
	var s strings.Builder
	s.WriteString(RenderTitle("🔍 Go Code Review"))
	s.WriteString("\n")
	s.WriteString(m.spinner.View())
	s.WriteString(" Analyzing your code changes...\n\n")
	s.WriteString(RenderSubtitle(m.reviewCtx.Summary()))
	s.WriteString("\n")
	s.WriteString(RenderHelp("q: quit"))
	return s.String()
}

// viewError renders the error state
func (m *Model) viewError() string {
	var s strings.Builder
	s.WriteString(RenderTitle("🔍 Go Code Review"))
	s.WriteString("\n")
	s.WriteString(RenderError(m.errorMsg))
	s.WriteString("\n")
	s.WriteString(RenderHelp("q: quit"))
	return s.String()
}

// viewMain renders the main reviewing/chatting/searching state
func (m *Model) viewMain() string {
	var s strings.Builder
	s.WriteString(RenderTitle("🔍 Go Code Review"))
	s.WriteString("\n")
	s.WriteString(m.viewport.View())
	s.WriteString("\n")

	if m.state == StateChatting {
		if m.streaming {
			s.WriteString(m.spinner.View())
			s.WriteString(" Thinking...\n")
		} else {
			s.WriteString(m.textarea.View())
		}
	}

	if m.state == StateSearching {
		s.WriteString(RenderSearchInput(
			m.searchInput.Value(),
			m.search.MatchCount(),
			m.search.CurrentMatch,
			m.search.Mode,
		))
		s.WriteString("\n")
	}

	// Yank feedback
	if m.yankFeedback != "" {
		s.WriteString("\n")
		s.WriteString(RenderSuccess(m.yankFeedback))
	}

	// Footer
	s.WriteString("\n")
	s.WriteString(m.viewFooter())

	return s.String()
}

// viewFooter renders the footer help text based on current state
func (m *Model) viewFooter() string {
	switch m.state {
	case StateSearching:
		return RenderHelp("enter: confirm • tab: toggle mode • esc: cancel")
	case StateReviewing:
		if m.search.Query != "" && m.search.MatchCount() > 0 {
			return RenderHelp(fmt.Sprintf("n/N: next/prev (%d/%d) • /: search • ?: help • q: quit",
				m.search.CurrentMatch+1, m.search.MatchCount()))
		}
		return RenderHelp("j/k: scroll • y: yank • /: search • ?: help • q: quit")
	case StateChatting:
		return RenderHelp("alt+enter: send • esc: back • q: quit")
	default:
		return RenderHelp("q: quit")
	}
}

// View renders the UI
func (m *Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	// If help overlay is active, render it
	if m.state == StateHelp {
		helpOverlay := NewHelpOverlay(m.width, m.height)
		return helpOverlay.Render()
	}

	switch m.state {
	case StateLoading:
		return m.viewLoading()
	case StateError:
		return m.viewError()
	case StateReviewing, StateChatting, StateSearching:
		return m.viewMain()
	default:
		return ""
	}
}
