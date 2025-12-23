package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Styles for web search indicator (defined once at package level)
var (
	webSearchIndicatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6B7280")).
				MarginBottom(0)
	webSearchCheckboxEnabledStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#10B981"))
	webSearchCheckboxDisabledStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#6B7280"))
)

// viewLoading renders the loading state
func (m *Model) viewLoading() string {
	var s strings.Builder
	s.WriteString(RenderTitle("🔍 LLM Review"))
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
	s.WriteString(RenderTitle("🔍 LLM Review"))
	s.WriteString("\n")
	s.WriteString(RenderError(m.errorMsg))
	s.WriteString("\n")
	// Show partial response if available (e.g., when request was cancelled)
	if m.reviewResponse != "" {
		s.WriteString("\n")
		s.WriteString(m.viewport.View())
		s.WriteString("\n")
	}
	s.WriteString(RenderHelp("q: quit"))
	return s.String()
}

// viewMain renders the main reviewing/chatting/searching state
func (m *Model) viewMain() string {
	var s strings.Builder
	s.WriteString(RenderTitle("🔍 LLM Review"))
	s.WriteString("\n")
	s.WriteString(m.viewport.View())
	s.WriteString("\n")

	if m.state == StateChatting {
		if m.streaming {
			s.WriteString(m.spinner.View())
			s.WriteString(" Thinking...\n")
		} else {
			// Render web search indicator
			s.WriteString(m.renderWebSearchIndicator())
			s.WriteString("\n")
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

// renderWebSearchIndicator renders the web search toggle indicator
func (m *Model) renderWebSearchIndicator() string {
	var checkbox string
	if m.webSearchEnabled {
		checkbox = webSearchCheckboxEnabledStyle.Render("[✓]")
	} else {
		checkbox = webSearchCheckboxDisabledStyle.Render("[ ]")
	}

	return webSearchIndicatorStyle.Render(fmt.Sprintf("%s Web Search (Ctrl+w to toggle)", checkbox))
}

// viewFooter renders the footer help text based on current state
func (m *Model) viewFooter() string {
	switch m.state {
	case StateSearching:
		return RenderCompactHelp("searching")
	case StateReviewing:
		if m.search.Query != "" && m.search.MatchCount() > 0 {
			return RenderHelp(fmt.Sprintf("n/N: next/prev (%d/%d) • /: search • ?: help • q: quit",
				m.search.CurrentMatch+1, m.search.MatchCount()))
		}
		return RenderCompactHelp("reviewing")
	case StateChatting:
		return RenderCompactHelp("chatting")
	case StateFileList:
		return RenderCompactHelp("filelist")
	default:
		return RenderCompactHelp("")
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
	case StateFileList:
		return m.viewFileList()
	case StateReviewing, StateChatting, StateSearching:
		return m.viewMain()
	default:
		return ""
	}
}

// viewFileList renders the file list view
func (m *Model) viewFileList() string {
	var s strings.Builder
	s.WriteString(RenderTitle("📁 Files to Review"))
	s.WriteString("\n\n")
	s.WriteString(m.fileList.View())
	s.WriteString("\n")

	// Yank feedback
	if m.yankFeedback != "" {
		s.WriteString("\n")
		s.WriteString(RenderSuccess(m.yankFeedback))
	}

	// Footer
	s.WriteString("\n")
	s.WriteString(RenderCompactHelp("filelist"))

	return s.String()
}
