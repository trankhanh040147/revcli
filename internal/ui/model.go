package ui

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	appcontext "github.com/trankhanh040147/revcli/internal/context"
	"github.com/trankhanh040147/revcli/internal/gemini"
	"github.com/trankhanh040147/revcli/internal/preset"
	"github.com/trankhanh040147/revcli/internal/prompt"
)

// State represents the current state of the application
type State int

const (
	StateLoading State = iota
	StateReviewing
	StateChatting
	StateSearching
	StateHelp
	StateError
	StateQuitting
)

// Model represents the application state for Bubbletea
type Model struct {
	// State machine
	state         State
	previousState State // For returning from search/help overlays

	// Review context
	reviewCtx *appcontext.ReviewContext

	// Gemini client
	client *gemini.Client
	ctx    context.Context
	cancel context.CancelFunc

	// Review preset
	preset *preset.Preset

	// UI components
	spinner     spinner.Model
	viewport    viewport.Model
	textarea    textarea.Model
	searchInput textinput.Model
	renderer    *Renderer

	// Search state
	search *SearchState

	// Content
	reviewResponse  string
	rawContent      string // Original content without search highlighting
	streamedContent string
	errorMsg        string
	chatHistory     []ChatMessage

	// Dimensions
	width  int
	height int

	// Flags
	ready     bool
	streaming bool

	// Yank state
	yankFeedback string // Feedback message for yank
	lastKeyWasY  bool   // For detecting "yb" combo
}

// ChatMessage represents a message in the chat history
type ChatMessage struct {
	Role    string // "user" or "assistant"
	Content string
}

// NewModel creates a new application model
func NewModel(reviewCtx *appcontext.ReviewContext, client *gemini.Client, p *preset.Preset) Model {
	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))

	// Create textarea for chat input
	ta := textarea.New()
	ta.Placeholder = "Ask a follow-up question..."
	ta.Focus()
	ta.CharLimit = 1000
	ta.SetWidth(80)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false

	// Custom textarea styling to avoid white background and row highlighting
	ta.FocusedStyle.Base = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED"))
	ta.BlurredStyle.Base = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4B5563"))
	// Remove cursor line highlighting
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.BlurredStyle.CursorLine = lipgloss.NewStyle()

	// Create search input
	si := textinput.New()
	si.Placeholder = "Search..."
	si.CharLimit = 100
	si.Width = 40

	// Create renderer (with fallback if it fails)
	renderer, err := NewRenderer()
	if err != nil {
		// Create a basic renderer as fallback
		renderer = &Renderer{}
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	return Model{
		state:       StateLoading,
		reviewCtx:   reviewCtx,
		client:      client,
		ctx:         ctx,
		cancel:      cancel,
		preset:      p,
		spinner:     s,
		textarea:    ta,
		searchInput: si,
		search:      NewSearchState(),
		renderer:    renderer,
		ready:       false,
		streaming:   false,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.startReview(),
	)
}

// startReview initiates the code review
func (m Model) startReview() tea.Cmd {
	return func() tea.Msg {
		// Initialize chat with system prompt (with preset if specified)
		systemPrompt := appcontext.GetSystemPrompt()
		if m.preset != nil {
			systemPrompt = appcontext.GetSystemPromptWithPreset(m.preset.Prompt)
		}
		m.client.StartChat(systemPrompt)

		// Send the review request with streaming
		var fullResponse strings.Builder

		_, err := m.client.SendMessageStream(m.ctx, m.reviewCtx.UserPrompt, func(chunk string) {
			fullResponse.WriteString(chunk)
		})
		if err != nil {
			return ReviewErrorMsg{Err: err}
		}

		return ReviewCompleteMsg{Response: fullResponse.String()}
	}
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle search mode input
		if m.state == StateSearching {
			switch msg.String() {
			case "esc":
				// Exit search mode, keep results
				m.state = m.previousState
				m.searchInput.Blur()
				m.viewport.Height = m.calculateViewportHeight()
				return m, nil
			case "enter":
				// Confirm search and exit search mode
				m.search.Query = m.searchInput.Value()
				m.search.Search(m.rawContent)
				m.state = m.previousState
				m.searchInput.Blur()
				m.viewport.Height = m.calculateViewportHeight()
				m.updateViewportWithSearch()
				return m, nil
			case "tab":
				// Toggle search mode
				m.search.ToggleMode()
				m.updateViewportWithSearch()
				return m, nil
			default:
				// Update search input
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				// Live search as user types
				m.search.Query = m.searchInput.Value()
				m.search.Search(m.rawContent)
				m.updateViewportWithSearch()
				return m, cmd
			}
		}

		switch msg.String() {
		case "ctrl+c", "q":
			if m.state != StateChatting || m.textarea.Value() == "" {
				m.cancel()
				return m, tea.Quit
			}
		case "esc":
			if m.state == StateHelp {
				m.state = m.previousState
				m.viewport.Height = m.calculateViewportHeight()
				return m, nil
			}
			if m.state == StateChatting {
				m.state = StateReviewing
				m.viewport.Height = m.calculateViewportHeight()
				return m, nil
			}
		case "enter":
			if m.state == StateReviewing {
				m.state = StateChatting
				m.textarea.Focus()
				m.viewport.Height = m.calculateViewportHeight()
				return m, nil
			}
			// In chat mode, let textarea handle Enter for newlines
		case "alt+enter":
			// Send message with Alt+Enter in chat mode
			if m.state == StateChatting && !m.streaming {
				question := strings.TrimSpace(m.textarea.Value())
				if question != "" {
					m.textarea.Reset()
					m.streaming = true
					m.chatHistory = append(m.chatHistory, ChatMessage{Role: "user", Content: question})
					return m, m.sendChatMessage(question)
				}
			}

		// Help overlay (only in reviewing mode, not chatting - let textarea handle ?)
		case "?":
			if m.state == StateReviewing {
				m.previousState = m.state
				m.state = StateHelp
				return m, nil
			} else if m.state == StateHelp {
				m.state = m.previousState
				return m, nil
			}

		// Search mode
		case "/":
			if m.state == StateReviewing {
				m.previousState = m.state
				m.state = StateSearching
				m.searchInput.SetValue(m.search.Query)
				m.searchInput.Focus()
				m.viewport.Height = m.calculateViewportHeight()
				return m, textinput.Blink
			}
		case "n":
			// Next match
			if m.state == StateReviewing && m.search.Query != "" {
				m.search.NextMatch()
				m.updateViewportWithSearch()
				m.scrollToCurrentMatch()
			}
			m.lastKeyWasY = false
		case "N":
			// Previous match
			if m.state == StateReviewing && m.search.Query != "" {
				m.search.PrevMatch()
				m.updateViewportWithSearch()
				m.scrollToCurrentMatch()
			}
			m.lastKeyWasY = false

		// Yank commands
		case "y":
			if m.state == StateReviewing && !m.lastKeyWasY {
				// First 'y' press - wait to see if followed by 'b'
				m.lastKeyWasY = true
				return m, tea.Tick(300*time.Millisecond, func(t time.Time) tea.Msg {
					return yankTimeoutMsg{}
				})
			} else if m.lastKeyWasY {
				// Second 'y' press (yy) - yank entire review
				m.lastKeyWasY = false
				return m, m.yankReview()
			}
		case "b":
			if m.state == StateReviewing && m.lastKeyWasY {
				// 'yb' combo - yank code block
				m.lastKeyWasY = false
				return m, m.yankCodeBlock()
			}
			m.lastKeyWasY = false
		case "Y":
			// Yank only the last/most recent response
			if m.state == StateReviewing {
				m.lastKeyWasY = false
				return m, m.yankLastResponse()
			}
			m.lastKeyWasY = false

		// Vim-style navigation (only in reviewing state, not chatting)
		case "j", "down":
			if m.state == StateReviewing {
				m.viewport.LineDown(1)
			}
		case "k", "up":
			if m.state == StateReviewing {
				m.viewport.LineUp(1)
			}
		case "g", "home":
			if m.state == StateReviewing {
				m.viewport.GotoTop()
			}
		case "G", "end":
			if m.state == StateReviewing {
				m.viewport.GotoBottom()
			}
		case "ctrl+d":
			if m.state == StateReviewing {
				m.viewport.HalfViewDown()
			}
		case "ctrl+u":
			if m.state == StateReviewing {
				m.viewport.HalfViewUp()
			}
		case "ctrl+f", "pgdown":
			if m.state == StateReviewing {
				m.viewport.ViewDown()
			}
		case "ctrl+b", "pgup":
			if m.state == StateReviewing {
				m.viewport.ViewUp()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		viewportHeight := m.calculateViewportHeight()
		if !m.ready {
			m.viewport = viewport.New(msg.Width, viewportHeight)
			m.viewport.Style = lipgloss.NewStyle().Padding(0, 2)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = viewportHeight
		}
		m.textarea.SetWidth(msg.Width - 4)

	case spinner.TickMsg:
		if m.state == StateLoading || m.streaming {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case ReviewCompleteMsg:
		m.state = StateReviewing
		m.reviewResponse = msg.Response
		m.streaming = false
		m.updateViewport()

	case ReviewErrorMsg:
		m.state = StateError
		m.errorMsg = msg.Err.Error()

	case ChatResponseMsg:
		m.streaming = false
		m.chatHistory = append(m.chatHistory, ChatMessage{Role: "assistant", Content: msg.Response})
		m.updateViewportAndScroll()

	case ChatErrorMsg:
		m.streaming = false
		m.errorMsg = msg.Err.Error()
		m.chatHistory = append(m.chatHistory, ChatMessage{Role: "assistant", Content: "Error: " + msg.Err.Error()})
		m.updateViewportAndScroll()

	case yankTimeoutMsg:
		// Timeout for yank combo - if lastKeyWasY is still true, yank entire review
		if m.lastKeyWasY {
			m.lastKeyWasY = false
			return m, m.yankReview()
		}

	case YankMsg:
		m.yankFeedback = fmt.Sprintf("✓ Copied %s to clipboard!", msg.Type)
		m.viewport.Height = m.calculateViewportHeight()
		return m, ClearYankFeedbackCmd(2 * time.Second)

	case YankFeedbackMsg:
		m.yankFeedback = ""
		m.viewport.Height = m.calculateViewportHeight()
	}

	// Update textarea in chat mode
	if m.state == StateChatting && !m.streaming {
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Update viewport
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// sendChatMessage sends a follow-up question
func (m Model) sendChatMessage(question string) tea.Cmd {
	return func() tea.Msg {
		followUp := prompt.BuildFollowUpPrompt(question)

		response, err := m.client.SendMessage(m.ctx, followUp)
		if err != nil {
			return ChatErrorMsg{Err: err}
		}

		return ChatResponseMsg{Response: response}
	}
}

// updateViewport updates the viewport content
// scrollToBottom determines whether to scroll to the bottom after update
func (m *Model) updateViewport() {
	m.updateViewportWithScroll(false)
}

// updateViewportAndScroll updates viewport and scrolls to bottom (for chat responses)
func (m *Model) updateViewportAndScroll() {
	m.updateViewportWithScroll(true)
}

// updateViewportWithScroll updates the viewport content with optional scroll
func (m *Model) updateViewportWithScroll(scrollToBottom bool) {
	var content strings.Builder

	// Render the main review response
	if m.reviewResponse != "" {
		rendered, err := m.renderer.RenderMarkdown(m.reviewResponse)
		if err != nil {
			content.WriteString(m.reviewResponse)
		} else {
			content.WriteString(rendered)
		}
	}

	// Render chat history
	if len(m.chatHistory) > 0 {
		content.WriteString("\n")
		content.WriteString(RenderDivider(m.width - 4))
		content.WriteString("\n\n")
		content.WriteString(RenderTitle("💬 Follow-up Chat"))
		content.WriteString("\n\n")

		for _, msg := range m.chatHistory {
			if msg.Role == "user" {
				content.WriteString(RenderPrompt())
				content.WriteString(msg.Content)
				content.WriteString("\n\n")
			} else {
				rendered, err := m.renderer.RenderMarkdown(msg.Content)
				if err != nil {
					content.WriteString(msg.Content)
				} else {
					content.WriteString(rendered)
				}
				content.WriteString("\n")
			}
		}
	}

	m.rawContent = content.String()
	m.viewport.SetContent(m.rawContent)

	// Only scroll to bottom for chat updates, not initial review
	if scrollToBottom {
		m.viewport.GotoBottom()
	}
}

// updateViewportWithSearch updates the viewport with search highlighting
func (m *Model) updateViewportWithSearch() {
	if m.rawContent == "" {
		return
	}

	if m.search.Query == "" || len(m.search.Matches) == 0 {
		m.viewport.SetContent(m.rawContent)
		return
	}

	var displayContent string
	if m.search.Mode == SearchModeFilter {
		displayContent = m.search.FilterContent(m.rawContent)
	} else {
		displayContent = m.search.HighlightContent(m.rawContent)
	}

	m.viewport.SetContent(displayContent)
}

// calculateViewportHeight calculates the dynamic viewport height based on UI state
func (m *Model) calculateViewportHeight() int {
	// Base reserved lines: header (2) + footer (2)
	reserved := 4

	// Search bar when active
	if m.state == StateSearching {
		reserved += 2
	}

	// Yank feedback when showing
	if m.yankFeedback != "" {
		reserved += 1
	}

	// Chat textarea when in chat mode
	if m.state == StateChatting {
		reserved += 4
	}

	height := m.height - reserved
	if height < 5 {
		height = 5 // Minimum height
	}
	return height
}

// scrollToCurrentMatch scrolls the viewport to show the current match
func (m *Model) scrollToCurrentMatch() {
	line := m.search.CurrentMatchLine()
	if line < 0 {
		return
	}

	// In filter mode, translate original line to filtered view index
	if m.search.Mode == SearchModeFilter {
		line = m.search.FilteredLineIndex(line)
		if line < 0 {
			return
		}
	}

	// Calculate approximate line position and scroll there
	// This is an approximation since rendered content may have different line counts
	viewportHeight := m.viewport.Height
	targetLine := line - viewportHeight/2
	if targetLine < 0 {
		targetLine = 0
	}

	m.viewport.GotoTop()
	m.viewport.LineDown(targetLine)
}

// yankReview yanks the entire review content to clipboard
func (m *Model) yankReview() tea.Cmd {
	return func() tea.Msg {
		var content strings.Builder

		// Use raw review response (markdown without ANSI codes)
		if m.reviewResponse != "" {
			content.WriteString(m.reviewResponse)
		}

		// Add raw chat history
		if len(m.chatHistory) > 0 {
			content.WriteString("\n\n---\n\n## Follow-up Chat\n\n")
			for _, msg := range m.chatHistory {
				if msg.Role == "user" {
					content.WriteString("**You:** ")
					content.WriteString(msg.Content)
				} else {
					content.WriteString("**Assistant:**\n")
					content.WriteString(msg.Content)
				}
				content.WriteString("\n\n")
			}
		}

		result := content.String()
		if result == "" {
			return nil
		}

		err := clipboard.WriteAll(result)
		if err != nil {
			return ChatErrorMsg{Err: fmt.Errorf("failed to copy to clipboard: %w", err)}
		}

		return YankMsg{Content: result, Type: "review"}
	}
}

// yankCodeBlock yanks the code block nearest to the current viewport position
func (m *Model) yankCodeBlock() tea.Cmd {
	return func() tea.Msg {
		content := m.reviewResponse
		if content == "" {
			return nil
		}

		// Find all code blocks (fenced with ```)
		codeBlockRegex := regexp.MustCompile("(?s)```[a-zA-Z]*\n(.*?)```")
		matches := codeBlockRegex.FindAllStringSubmatch(content, -1)

		if len(matches) == 0 {
			return YankMsg{Content: "", Type: "no code blocks found"}
		}

		// Get the first code block (or could be enhanced to find nearest to cursor)
		// For now, we'll get the last code block as it's likely the most relevant suggestion
		lastMatch := matches[len(matches)-1]
		codeContent := strings.TrimSpace(lastMatch[1])

		if codeContent == "" {
			return YankMsg{Content: "", Type: "empty code block"}
		}

		err := clipboard.WriteAll(codeContent)
		if err != nil {
			return ChatErrorMsg{Err: fmt.Errorf("failed to copy to clipboard: %w", err)}
		}

		return YankMsg{Content: codeContent, Type: "code block"}
	}
}

// yankLastResponse yanks only the last/most recent assistant response
func (m *Model) yankLastResponse() tea.Cmd {
	return func() tea.Msg {
		var content string

		// If there's chat history, get the last assistant message
		if len(m.chatHistory) > 0 {
			for i := len(m.chatHistory) - 1; i >= 0; i-- {
				if m.chatHistory[i].Role == "assistant" {
					content = m.chatHistory[i].Content
					break
				}
			}
		}

		// If no chat history or no assistant message found, use the initial review
		if content == "" {
			content = m.reviewResponse
		}

		if content == "" {
			return nil
		}

		err := clipboard.WriteAll(content)
		if err != nil {
			return ChatErrorMsg{Err: fmt.Errorf("failed to copy to clipboard: %w", err)}
		}

		return YankMsg{Content: content, Type: "last response"}
	}
}

// View renders the UI
func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	// If help overlay is active, render it
	if m.state == StateHelp {
		helpOverlay := NewHelpOverlay(m.width, m.height)
		return helpOverlay.Render()
	}

	var s strings.Builder

	// Header
	s.WriteString(RenderTitle("🔍 Go Code Review"))
	s.WriteString("\n")

	switch m.state {
	case StateLoading:
		s.WriteString(m.spinner.View())
		s.WriteString(" Analyzing your code changes...\n\n")
		s.WriteString(RenderSubtitle(m.reviewCtx.Summary()))

	case StateReviewing, StateChatting, StateSearching:
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

	case StateError:
		s.WriteString(RenderError(m.errorMsg))
	}

	// Yank feedback
	if m.yankFeedback != "" {
		s.WriteString("\n")
		s.WriteString(RenderSuccess(m.yankFeedback))
	}

	// Footer
	s.WriteString("\n")
	if m.state == StateSearching {
		s.WriteString(RenderHelp("enter: confirm • tab: toggle mode • esc: cancel"))
	} else if m.state == StateReviewing {
		var helpText string
		if m.search.Query != "" && m.search.MatchCount() > 0 {
			helpText = fmt.Sprintf("n/N: next/prev (%d/%d) • /: search • ?: help • q: quit",
				m.search.CurrentMatch+1, m.search.MatchCount())
		} else {
			helpText = "j/k: scroll • y: yank • /: search • ?: help • q: quit"
		}
		s.WriteString(RenderHelp(helpText))
	} else if m.state == StateChatting {
		s.WriteString(RenderHelp("alt+enter: send • esc: back • q: quit"))
	} else {
		s.WriteString(RenderHelp("q: quit"))
	}

	return s.String()
}

// Run starts the Bubbletea program
func Run(reviewCtx *appcontext.ReviewContext, client *gemini.Client, p *preset.Preset) error {
	model := NewModel(reviewCtx, client, p)
	program := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := program.Run(); err != nil {
		return fmt.Errorf("error running UI: %w", err)
	}

	return nil
}

// RunSimple runs a simple non-interactive review
func RunSimple(ctx context.Context, reviewCtx *appcontext.ReviewContext, client *gemini.Client, p *preset.Preset) error {
	// Initialize chat with system prompt (with preset if specified)
	systemPrompt := appcontext.GetSystemPrompt()
	if p != nil {
		systemPrompt = appcontext.GetSystemPromptWithPreset(p.Prompt)
	}
	client.StartChat(systemPrompt)

	fmt.Println(RenderTitle("🔍 Code Review"))
	fmt.Println(RenderSubtitle(reviewCtx.Summary()))
	fmt.Println()
	fmt.Println("Analyzing your code changes...")
	fmt.Println()

	// Create renderer
	renderer, err := NewRenderer()
	if err != nil {
		return fmt.Errorf("failed to create renderer: %w", err)
	}

	// Stream the response
	startTime := time.Now()
	response, err := client.SendMessageStream(ctx, reviewCtx.UserPrompt, func(chunk string) {
		fmt.Print(chunk)
	})
	if err != nil {
		return fmt.Errorf("review failed: %w", err)
	}

	// Render the full response with markdown
	fmt.Println()
	fmt.Println(RenderDivider(80))
	fmt.Println()

	rendered, err := renderer.RenderMarkdown(response)
	if err != nil {
		fmt.Println(response)
	} else {
		fmt.Println(rendered)
	}

	elapsed := time.Since(startTime)
	fmt.Println()
	fmt.Println(RenderSuccess(fmt.Sprintf("Review completed in %s", elapsed.Round(time.Millisecond))))

	// Display token usage
	if usage := client.GetLastUsage(); usage != nil {
		fmt.Println(RenderTokenUsage(usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens))
	}

	return nil
}
