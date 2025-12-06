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

	// Code block state
	codeBlocks           []CodeBlock // All detected code blocks
	activeCodeBlockIndex int         // Currently highlighted code block (-1 for none)
	currentViewportLine  int         // Current line in viewport

	// Prompt history state
	promptHistory      []string // History of sent prompts
	promptHistoryIndex int      // Current position in history (-1 for new prompt)
}

// ChatMessage represents a message in the chat history
type ChatMessage struct {
	Role    string // "user" or "assistant"
	Content string
}

// CodeBlock represents a code block in the markdown content
type CodeBlock struct {
	StartLine int    // Start line number (0-indexed)
	EndLine   int    // End line number (0-indexed)
	Language  string // Programming language
	Content   string // Code content without fences
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
		state:                StateLoading,
		reviewCtx:            reviewCtx,
		client:               client,
		ctx:                  ctx,
		cancel:               cancel,
		preset:               p,
		spinner:              s,
		textarea:             ta,
		searchInput:          si,
		search:               NewSearchState(),
		renderer:             renderer,
		ready:                false,
		streaming:            false,
		codeBlocks:           []CodeBlock{},
		activeCodeBlockIndex: -1,
		currentViewportLine:  0,
		promptHistory:        []string{},
		promptHistoryIndex:   -1,
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
		case "ctrl+x":
			// Cancel streaming request while retaining partial response
			if m.streaming {
				m.cancel()
				m.streaming = false
				// Create new context for future requests
				m.ctx, m.cancel = context.WithCancel(context.Background())
				// Keep any partial response that was already received
				m.yankFeedback = "Request cancelled"
				m.viewport.Height = m.calculateViewportHeight()
				return m, ClearYankFeedbackCmd(2 * time.Second)
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
					// Save to prompt history (avoid duplicates)
					if len(m.promptHistory) == 0 || m.promptHistory[len(m.promptHistory)-1] != question {
						m.promptHistory = append(m.promptHistory, question)
					}
					m.promptHistoryIndex = -1 // Reset to new prompt
					m.textarea.Reset()
					m.streaming = true
					m.chatHistory = append(m.chatHistory, ChatMessage{Role: "user", Content: question})
					return m, m.sendChatMessage(question)
				}
			}
		case "ctrl+p":
			// Previous prompt in chat mode
			if m.state == StateChatting && !m.streaming && len(m.promptHistory) > 0 {
				if m.promptHistoryIndex == -1 {
					// Start from last prompt
					m.promptHistoryIndex = len(m.promptHistory) - 1
				} else if m.promptHistoryIndex > 0 {
					m.promptHistoryIndex--
				}
				// Load prompt from history
				m.textarea.SetValue(m.promptHistory[m.promptHistoryIndex])
				m.textarea.CursorEnd()
			}
		case "ctrl+n":
			// Next prompt in chat mode
			if m.state == StateChatting && !m.streaming {
				if m.promptHistoryIndex >= 0 {
					m.promptHistoryIndex++
					if m.promptHistoryIndex >= len(m.promptHistory) {
						// Beyond last, clear textarea
						m.promptHistoryIndex = -1
						m.textarea.SetValue("")
					} else {
						// Load next prompt
						m.textarea.SetValue(m.promptHistory[m.promptHistoryIndex])
						m.textarea.CursorEnd()
					}
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
				m.currentViewportLine = m.getCurrentViewportLine()
				m.activeCodeBlockIndex = m.findCodeBlockAtViewport()
			}
		case "k", "up":
			if m.state == StateReviewing {
				m.viewport.LineUp(1)
				m.currentViewportLine = m.getCurrentViewportLine()
				m.activeCodeBlockIndex = m.findCodeBlockAtViewport()
			}
		case "g", "home":
			if m.state == StateReviewing {
				m.viewport.GotoTop()
				m.currentViewportLine = m.getCurrentViewportLine()
				m.activeCodeBlockIndex = m.findCodeBlockAtViewport()
			}
		case "G", "end":
			if m.state == StateReviewing {
				m.viewport.GotoBottom()
				m.currentViewportLine = m.getCurrentViewportLine()
				m.activeCodeBlockIndex = m.findCodeBlockAtViewport()
			}
		case "ctrl+d":
			if m.state == StateReviewing {
				m.viewport.HalfViewDown()
				m.currentViewportLine = m.getCurrentViewportLine()
				m.activeCodeBlockIndex = m.findCodeBlockAtViewport()
			}
		case "ctrl+u":
			if m.state == StateReviewing {
				m.viewport.HalfViewUp()
				m.currentViewportLine = m.getCurrentViewportLine()
				m.activeCodeBlockIndex = m.findCodeBlockAtViewport()
			}
		case "ctrl+f", "pgdown":
			if m.state == StateReviewing {
				m.viewport.ViewDown()
				m.currentViewportLine = m.getCurrentViewportLine()
				m.activeCodeBlockIndex = m.findCodeBlockAtViewport()
			}
		case "ctrl+b", "pgup":
			if m.state == StateReviewing {
				m.viewport.ViewUp()
				m.currentViewportLine = m.getCurrentViewportLine()
				m.activeCodeBlockIndex = m.findCodeBlockAtViewport()
			}

		// Code block navigation
		case "[":
			if m.state == StateReviewing {
				m.prevCodeBlock()
			}
		case "]":
			if m.state == StateReviewing {
				m.nextCodeBlock()
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
	m.codeBlocks = []CodeBlock{} // Reset code blocks

	// Render the main review response
	if m.reviewResponse != "" {
		rendered, err := m.renderer.RenderMarkdown(m.reviewResponse)
		if err != nil {
			content.WriteString(m.reviewResponse)
		} else {
			content.WriteString(rendered)
		}

		// Parse code blocks from rendered review response
		// Line numbers will be relative to the rendered content
		reviewCodeBlocks := parseCodeBlocksFromRendered(content.String())
		m.codeBlocks = append(m.codeBlocks, reviewCodeBlocks...)
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
				// Render assistant message
				rendered, err := m.renderer.RenderMarkdown(msg.Content)
				if err != nil {
					rendered = msg.Content
				}

				// Parse code blocks from rendered assistant message
				// Adjust line numbers to account for content before this message
				currentLineOffset := strings.Count(content.String(), "\n")
				assistantCodeBlocks := parseCodeBlocksFromRendered(rendered)
				for i := range assistantCodeBlocks {
					assistantCodeBlocks[i].StartLine += currentLineOffset
					assistantCodeBlocks[i].EndLine += currentLineOffset
					m.codeBlocks = append(m.codeBlocks, assistantCodeBlocks[i])
				}

				content.WriteString(rendered)
				content.WriteString("\n")
			}
		}
	}

	m.rawContent = content.String()

	// Update current viewport line based on scroll position
	m.currentViewportLine = m.getCurrentViewportLine()

	// Detect which code block is at current viewport position (if not explicitly set)
	if m.activeCodeBlockIndex < 0 {
		m.activeCodeBlockIndex = m.findCodeBlockAtViewport()
	}

	// Apply code block highlighting if there's an active code block
	finalContent := m.rawContent
	if m.activeCodeBlockIndex >= 0 && m.activeCodeBlockIndex < len(m.codeBlocks) {
		finalContent = m.highlightCodeBlockInContent(m.rawContent)
	}

	m.viewport.SetContent(finalContent)

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

	var displayContent string
	if m.search.Query == "" || len(m.search.Matches) == 0 {
		displayContent = m.rawContent
	} else {
		if m.search.Mode == SearchModeFilter {
			displayContent = m.search.FilterContent(m.rawContent)
		} else {
			displayContent = m.search.HighlightContent(m.rawContent)
		}
	}

	// Apply code block highlighting if there's an active code block
	// This preserves code block borders even when search highlighting is applied
	if m.activeCodeBlockIndex >= 0 && m.activeCodeBlockIndex < len(m.codeBlocks) {
		displayContent = m.highlightCodeBlockInContent(displayContent)
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

// prevCodeBlock navigates to the previous code block
func (m *Model) prevCodeBlock() {
	if len(m.codeBlocks) == 0 {
		return
	}

	if m.activeCodeBlockIndex < 0 {
		// If no active block, go to last one
		m.activeCodeBlockIndex = len(m.codeBlocks) - 1
	} else if m.activeCodeBlockIndex > 0 {
		m.activeCodeBlockIndex--
	} else {
		// Already at first block, wrap to last
		m.activeCodeBlockIndex = len(m.codeBlocks) - 1
	}

	m.scrollToCodeBlock(m.activeCodeBlockIndex)
	m.updateViewport()
}

// nextCodeBlock navigates to the next code block
func (m *Model) nextCodeBlock() {
	if len(m.codeBlocks) == 0 {
		return
	}

	if m.activeCodeBlockIndex < 0 {
		// If no active block, go to first one
		m.activeCodeBlockIndex = 0
	} else {
		m.activeCodeBlockIndex = (m.activeCodeBlockIndex + 1) % len(m.codeBlocks)
	}

	m.scrollToCodeBlock(m.activeCodeBlockIndex)
	m.updateViewport()
}

// scrollToCodeBlock scrolls the viewport to show the specified code block
func (m *Model) scrollToCodeBlock(index int) {
	if index < 0 || index >= len(m.codeBlocks) {
		return
	}

	block := m.codeBlocks[index]

	// Calculate target line (middle of code block)
	targetLine := (block.StartLine + block.EndLine) / 2

	// Scroll to show the code block in the center of viewport
	viewportHeight := m.viewport.Height
	scrollLine := targetLine - viewportHeight/2
	if scrollLine < 0 {
		scrollLine = 0
	}

	m.viewport.GotoTop()
	m.viewport.LineDown(scrollLine)
	m.currentViewportLine = m.getCurrentViewportLine()
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

// yankCodeBlock yanks the active/highlighted code block, or falls back to last block
func (m *Model) yankCodeBlock() tea.Cmd {
	return func() tea.Msg {
		// Check if there's an active code block
		if m.activeCodeBlockIndex >= 0 && m.activeCodeBlockIndex < len(m.codeBlocks) {
			block := m.codeBlocks[m.activeCodeBlockIndex]
			if block.Content != "" {
				err := clipboard.WriteAll(block.Content)
				if err != nil {
					return ChatErrorMsg{Err: fmt.Errorf("failed to copy to clipboard: %w", err)}
				}
				return YankMsg{Content: block.Content, Type: "code block"}
			}
		}

		// Fallback to last code block if no active block
		if len(m.codeBlocks) > 0 {
			lastBlock := m.codeBlocks[len(m.codeBlocks)-1]
			if lastBlock.Content != "" {
				err := clipboard.WriteAll(lastBlock.Content)
				if err != nil {
					return ChatErrorMsg{Err: fmt.Errorf("failed to copy to clipboard: %w", err)}
				}
				return YankMsg{Content: lastBlock.Content, Type: "code block"}
			}
		}

		// No code blocks found
		return YankMsg{Content: "", Type: "no code blocks found"}
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

// getCurrentViewportLine returns the approximate current line in the viewport
func (m *Model) getCurrentViewportLine() int {
	// YOffset gives us the number of lines scrolled from top
	// Add half the viewport height to get the middle line
	return m.viewport.YOffset + m.viewport.Height/2
}

// findCodeBlockAtViewport finds which code block (if any) is at the current viewport position
// Returns the index of the code block, or -1 if none
func (m *Model) findCodeBlockAtViewport() int {
	if len(m.codeBlocks) == 0 {
		return -1
	}

	// For now, we'll use a simple approach: check if viewport line is within any code block
	// This is approximate since rendered content may have different line counts
	viewportLine := m.getCurrentViewportLine()

	for i, block := range m.codeBlocks {
		// Check if viewport line is within the code block range
		// We use a tolerance since rendered content may differ
		if viewportLine >= block.StartLine-5 && viewportLine <= block.EndLine+5 {
			return i
		}
	}

	return -1
}

// highlightCodeBlockInContent applies border highlighting to the active code block in rendered content
func (m *Model) highlightCodeBlockInContent(content string) string {
	if m.activeCodeBlockIndex < 0 || m.activeCodeBlockIndex >= len(m.codeBlocks) {
		return content
	}

	block := m.codeBlocks[m.activeCodeBlockIndex]

	// Find the code block in the rendered content
	// Code blocks in rendered markdown typically have ANSI codes, so we search for the content
	// We'll search for a portion of the code content to locate it
	if block.Content == "" {
		return content
	}

	// Try to find the code block by searching for a unique portion of its content
	// Take first 20 characters as a search key
	searchKey := block.Content
	if len(searchKey) > 20 {
		searchKey = searchKey[:20]
	}

	// Search for the code block in the content
	lines := strings.Split(content, "\n")
	blockStartIdx := -1
	blockEndIdx := -1

	// Find lines that contain the search key
	for i, line := range lines {
		if strings.Contains(line, searchKey) && blockStartIdx == -1 {
			blockStartIdx = i
		}
		// Estimate end based on code block size
		if blockStartIdx >= 0 && i-blockStartIdx > block.EndLine-block.StartLine {
			blockEndIdx = i
			break
		}
	}

	if blockStartIdx == -1 {
		return content
	}

	if blockEndIdx == -1 {
		blockEndIdx = blockStartIdx + (block.EndLine - block.StartLine)
		if blockEndIdx >= len(lines) {
			blockEndIdx = len(lines) - 1
		}
	}

	// Extract the code block lines
	codeBlockLines := lines[blockStartIdx : blockEndIdx+1]
	codeBlockContent := strings.Join(codeBlockLines, "\n")

	// Apply border style
	highlightedBlock := codeBlockBorderStyle.Render(codeBlockContent)

	// Reconstruct content with highlighted block
	var result strings.Builder
	result.WriteString(strings.Join(lines[:blockStartIdx], "\n"))
	if blockStartIdx > 0 {
		result.WriteString("\n")
	}
	result.WriteString(highlightedBlock)
	if blockEndIdx < len(lines)-1 {
		result.WriteString("\n")
		result.WriteString(strings.Join(lines[blockEndIdx+1:], "\n"))
	}

	return result.String()
}

// parseCodeBlocks parses raw markdown content to extract all code blocks with their positions
// This is used for yanking code blocks from raw markdown
func parseCodeBlocks(content string) []CodeBlock {
	if content == "" {
		return []CodeBlock{}
	}

	// Regex to match code blocks: ```language\ncontent\n```
	// Language identifier can contain letters, numbers, and common special characters (+, #, -, etc.)
	codeBlockRegex := regexp.MustCompile("(?s)```([a-zA-Z0-9+\\-#]*)\n(.*?)```")
	matches := codeBlockRegex.FindAllStringSubmatch(content, -1)

	if len(matches) == 0 {
		return []CodeBlock{}
	}

	var codeBlocks []CodeBlock

	// Track current position in content
	currentPos := 0
	for _, match := range matches {
		// Find the start position of this match in the original content
		startPos := strings.Index(content[currentPos:], match[0])
		if startPos == -1 {
			continue
		}
		startPos += currentPos

		// Calculate line numbers
		startLine := strings.Count(content[:startPos], "\n")
		endPos := startPos + len(match[0])
		endLine := strings.Count(content[:endPos], "\n")

		language := strings.TrimSpace(match[1])
		codeContent := strings.TrimSpace(match[2])

		codeBlocks = append(codeBlocks, CodeBlock{
			StartLine: startLine,
			EndLine:   endLine,
			Language:  language,
			Content:   codeContent,
		})

		currentPos = endPos
	}

	return codeBlocks
}

// parseCodeBlocksFromRendered parses rendered content (with ANSI codes) to extract code blocks
// This finds code blocks in the rendered output and calculates their line positions
func parseCodeBlocksFromRendered(renderedContent string) []CodeBlock {
	if renderedContent == "" {
		return []CodeBlock{}
	}

	// Regex to match code blocks in rendered content
	// Language identifier can contain letters, numbers, and common special characters (+, #, -, etc.)
	// Note: Rendered content may have ANSI codes, but code block fences should still be visible
	codeBlockRegex := regexp.MustCompile("(?s)```([a-zA-Z0-9+\\-#]*)\n(.*?)```")
	matches := codeBlockRegex.FindAllStringSubmatch(renderedContent, -1)

	if len(matches) == 0 {
		return []CodeBlock{}
	}

	var codeBlocks []CodeBlock

	// Track current position in content
	currentPos := 0
	for _, match := range matches {
		// Find the start position of this match in the rendered content
		startPos := strings.Index(renderedContent[currentPos:], match[0])
		if startPos == -1 {
			continue
		}
		startPos += currentPos

		// Calculate line numbers in rendered content
		startLine := strings.Count(renderedContent[:startPos], "\n")
		endPos := startPos + len(match[0])
		endLine := strings.Count(renderedContent[:endPos], "\n")

		language := strings.TrimSpace(match[1])
		codeContent := strings.TrimSpace(match[2])

		codeBlocks = append(codeBlocks, CodeBlock{
			StartLine: startLine,
			EndLine:   endLine,
			Language:  language,
			Content:   codeContent,
		})

		currentPos = endPos
	}

	return codeBlocks
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
		} else if m.activeCodeBlockIndex >= 0 && m.activeCodeBlockIndex < len(m.codeBlocks) {
			blockNum := m.activeCodeBlockIndex + 1
			totalBlocks := len(m.codeBlocks)
			helpText = fmt.Sprintf("[ ]: navigate blocks (%d/%d) • yb: copy block • /: search • ?: help • q: quit",
				blockNum, totalBlocks)
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
