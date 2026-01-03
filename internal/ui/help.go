package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// HelpOverlay renders the help overlay with all keybindings
type HelpOverlay struct {
	width  int
	height int
}

// NewHelpOverlay creates a new help overlay
func NewHelpOverlay(width, height int) *HelpOverlay {
	return &HelpOverlay{
		width:  width,
		height: height,
	}
}

// keybinding represents a single keybinding entry
type keybinding struct {
	key         string
	description string
}

// section represents a section of keybindings
type section struct {
	title    string
	bindings []keybinding
}

// Render renders the help overlay
func (h *HelpOverlay) Render() string {
	// Define all keybinding sections
	sections := []section{
		{
			title: "Navigation",
			bindings: []keybinding{
				{"j / ↓", "Scroll down one line"},
				{"k / ↑", "Scroll up one line"},
				{"g / Home", "Go to top"},
				{"G / End", "Go to bottom"},
				{"Ctrl+d", "Half page down"},
				{"Ctrl+u", "Half page up"},
				{"Ctrl+f / PgDn", "Full page down"},
				{"Ctrl+b / PgUp", "Full page up"},
			},
		},
		{
			title: "Search",
			bindings: []keybinding{
				{"/", "Start search"},
				{"n", "Next match"},
				{"N", "Previous match"},
				{"Tab", "Toggle highlight/filter mode"},
				{"Esc", "Exit search"},
			},
		},
		{
			title: "Clipboard",
			bindings: []keybinding{
				{"y / yy", "Yank entire review + chat history"},
				{"Y", "Yank only last response"},
			},
		},
		{
			title: "Chat",
			bindings: []keybinding{
				{"Enter", "Enter chat mode / Create newline"},
				{"Alt+Enter", "Send message"},
				{"Ctrl+P", "Previous prompt"},
				{"Ctrl+N", "Next prompt"},
				{"Ctrl+X", "Cancel request"},
				{"Ctrl+W", "Toggle web search"},
				{"Esc", "Exit chat mode"},
			},
		},
		{
			title: "File List",
			bindings: []keybinding{
				{"i", "Enter file list / Prune selected file"},
				{"j/k", "Navigate files"},
				{"Enter", "View selected file"},
				{"Esc", "Back to review"},
			},
		},
		{
			title: "General",
			bindings: []keybinding{
				{"?", "Toggle this help"},
				{"q", "Quit"},
				{"Ctrl+c", "Force quit"},
			},
		},
	}

	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED")).
		MarginBottom(1)

	sectionTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#3B82F6")).
		MarginTop(1)

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FACC15")).
		Bold(true).
		Width(16)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF"))

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4B5563")).
		Padding(1, 2)

	// Build content
	var content strings.Builder

	content.WriteString(titleStyle.Render("⌨️  Keyboard Shortcuts"))
	content.WriteString("\n\n")

	for _, sec := range sections {
		content.WriteString(sectionTitleStyle.Render(sec.title))
		content.WriteString("\n")

		for _, b := range sec.bindings {
			content.WriteString("  ")
			content.WriteString(keyStyle.Render(b.key))
			content.WriteString(descStyle.Render(b.description))
			content.WriteString("\n")
		}
	}

	content.WriteString("\n")
	content.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Italic(true).
		Render("Press ? or Esc to close"))

	// Render box content
	boxContent := borderStyle.Render(content.String())

	// Center the box using lipgloss.Place()
	return lipgloss.Place(h.width, h.height, lipgloss.Center, lipgloss.Center, boxContent)
}

// RenderCompact renders a compact version of the help for the footer
func RenderCompactHelp(state string) string {
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280"))

	switch state {
	case "reviewing":
		return helpStyle.Render("j/k: scroll • /: search • i: file list • ?: help • enter: chat • q: quit")
	case "chatting":
		return helpStyle.Render("enter: send • ctrl+w: toggle web search • esc: back • ?: help • q: quit")
	case "searching":
		return helpStyle.Render("enter: confirm • tab: mode • n/N: matches • esc: cancel")
	case "filelist":
		return helpStyle.Render("j/k: navigate • i: prune • Enter: view • Esc: back")
	case "help":
		return helpStyle.Render("?: close • esc: close")
	default:
		return helpStyle.Render("?: help • q: quit")
	}
}
