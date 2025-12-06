package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// Styles for the UI
var (
	// Title style
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			MarginBottom(1)

	// Subtitle style
	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			MarginBottom(1)

	// Error style
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)

	// Success style
	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981"))

	// Warning style
	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B"))

	// Input prompt style
	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B82F6")).
			Bold(true)

	// Spinner style
	spinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED"))

	// Help style
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			MarginTop(1)

	// Border style for content boxes
	boxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4B5563")).
		Padding(1, 2)

	// Code block border style for active highlighting
	codeBlockBorderStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		BorderTop(true).
		BorderBottom(true).
		BorderLeft(true).
		BorderRight(true)

	// Divider
	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#374151"))
)

// Renderer handles markdown rendering
type Renderer struct {
	glamour *glamour.TermRenderer
}

// NewRenderer creates a new markdown renderer
func NewRenderer() (*Renderer, error) {
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		return nil, err
	}

	return &Renderer{glamour: r}, nil
}

// RenderMarkdown renders markdown content for the terminal
func (r *Renderer) RenderMarkdown(content string) (string, error) {
	if r == nil || r.glamour == nil {
		return content, nil // Return raw content if renderer not available
	}
	return r.glamour.Render(content)
}

// RenderTitle renders a title
func RenderTitle(text string) string {
	return titleStyle.Render(text)
}

// RenderSubtitle renders a subtitle
func RenderSubtitle(text string) string {
	return subtitleStyle.Render(text)
}

// RenderError renders an error message
func RenderError(text string) string {
	return errorStyle.Render("✗ " + text)
}

// RenderSuccess renders a success message
func RenderSuccess(text string) string {
	return successStyle.Render("✓ " + text)
}

// RenderWarning renders a warning message
func RenderWarning(text string) string {
	return warningStyle.Render("⚠ " + text)
}

// RenderPrompt renders an input prompt
func RenderPrompt() string {
	return promptStyle.Render("❯ ")
}

// RenderHelp renders help text
func RenderHelp(text string) string {
	return helpStyle.Render(text)
}

// RenderBox renders content in a box
func RenderBox(content string) string {
	return boxStyle.Render(content)
}

// RenderDivider renders a divider line
func RenderDivider(width int) string {
	return dividerStyle.Render(strings.Repeat("─", width))
}

// RenderSecretWarning renders a warning about detected secrets
func RenderSecretWarning(secrets []SecretInfo) string {
	var sb strings.Builder

	sb.WriteString(errorStyle.Render("🔐 Potential secrets detected!\n\n"))

	for _, s := range secrets {
		sb.WriteString(warningStyle.Render("  • "))
		sb.WriteString(s.FilePath)
		sb.WriteString(" (line ")
		sb.WriteString(string(rune(s.Line + '0')))
		sb.WriteString("): ")
		sb.WriteString(s.Match)
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(helpStyle.Render("Use --force to proceed anyway (not recommended)"))

	return sb.String()
}

// SecretInfo contains information about a detected secret
type SecretInfo struct {
	FilePath string
	Line     int
	Match    string
}

// RenderLoadingDots renders animated loading dots
func RenderLoadingDots(tick int) string {
	dots := []string{"", ".", "..", "..."}
	return spinnerStyle.Render("Analyzing" + dots[tick%4])
}

// RenderProgress renders a simple progress indicator
func RenderProgress(current, total int) string {
	percentage := float64(current) / float64(total) * 100
	return subtitleStyle.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			"Progress: ",
			successStyle.Render(strings.Repeat("█", current)),
			dividerStyle.Render(strings.Repeat("░", total-current)),
			fmt.Sprintf(" %d%%", int(percentage)),
		),
	)
}

// AppHeader renders the application header
func AppHeader() string {
	header := `
 ██████╗  ██████╗       ██████╗ ███████╗██╗   ██╗
██╔════╝ ██╔═══██╗      ██╔══██╗██╔════╝██║   ██║
██║  ███╗██║   ██║█████╗██████╔╝█████╗  ██║   ██║
██║   ██║██║   ██║╚════╝██╔══██╗██╔══╝  ╚██╗ ██╔╝
╚██████╔╝╚██████╔╝      ██║  ██║███████╗ ╚████╔╝ 
 ╚═════╝  ╚═════╝       ╚═╝  ╚═╝╚══════╝  ╚═══╝  
`
	return titleStyle.Render(header)
}

// HelpFooter renders the help footer
func HelpFooter() string {
	return helpStyle.Render("q: quit • enter: send message • esc: exit chat mode")
}

// RenderTokenUsage renders token usage information
func RenderTokenUsage(prompt, completion, total int32) string {
	return subtitleStyle.Render(fmt.Sprintf(
		"📊 Token Usage: %d prompt + %d completion = %d total",
		prompt, completion, total,
	))
}
