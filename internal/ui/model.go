package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	appcontext "github.com/trankhanh040147/go-rev-cli/internal/context"
	"github.com/trankhanh040147/go-rev-cli/internal/gemini"
	"github.com/trankhanh040147/go-rev-cli/internal/prompt"
)

// State represents the current state of the application
type State int

const (
	StateLoading State = iota
	StateReviewing
	StateChatting
	StateError
	StateQuitting
)

// Model represents the application state for Bubbletea
type Model struct {
	// State machine
	state State

	// Review context
	reviewCtx *appcontext.ReviewContext

	// Gemini client
	client *gemini.Client
	ctx    context.Context
	cancel context.CancelFunc

	// UI components
	spinner  spinner.Model
	viewport viewport.Model
	textarea textarea.Model
	renderer *Renderer

	// Content
	reviewResponse  string
	streamedContent string
	errorMsg        string
	chatHistory     []ChatMessage

	// Dimensions
	width  int
	height int

	// Flags
	ready     bool
	streaming bool
}

// ChatMessage represents a message in the chat history
type ChatMessage struct {
	Role    string // "user" or "assistant"
	Content string
}

// NewModel creates a new application model
func NewModel(reviewCtx *appcontext.ReviewContext, client *gemini.Client) Model {
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

	// Create renderer
	renderer, _ := NewRenderer()

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	return Model{
		state:     StateLoading,
		reviewCtx: reviewCtx,
		client:    client,
		ctx:       ctx,
		cancel:    cancel,
		spinner:   s,
		textarea:  ta,
		renderer:  renderer,
		ready:     false,
		streaming: false,
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
		// Initialize chat with system prompt
		m.client.StartChat(appcontext.GetSystemPrompt())

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
		switch msg.String() {
		case "ctrl+c", "q":
			if m.state != StateChatting || m.textarea.Value() == "" {
				m.cancel()
				return m, tea.Quit
			}
		case "esc":
			if m.state == StateChatting {
				m.state = StateReviewing
				return m, nil
			}
		case "enter":
			if m.state == StateReviewing {
				m.state = StateChatting
				m.textarea.Focus()
				return m, nil
			}
			if m.state == StateChatting && !m.streaming {
				question := strings.TrimSpace(m.textarea.Value())
				if question != "" {
					m.textarea.Reset()
					m.streaming = true
					m.chatHistory = append(m.chatHistory, ChatMessage{Role: "user", Content: question})
					return m, m.sendChatMessage(question)
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-10)
			m.viewport.Style = lipgloss.NewStyle().Padding(0, 2)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 10
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
		m.updateViewport()

	case ChatErrorMsg:
		m.streaming = false
		m.errorMsg = msg.Err.Error()
		m.chatHistory = append(m.chatHistory, ChatMessage{Role: "assistant", Content: "Error: " + msg.Err.Error()})
		m.updateViewport()
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
func (m *Model) updateViewport() {
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

	m.viewport.SetContent(content.String())
	m.viewport.GotoBottom()
}

// View renders the UI
func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
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

	case StateReviewing, StateChatting:
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

	case StateError:
		s.WriteString(RenderError(m.errorMsg))
	}

	// Footer
	s.WriteString("\n")
	if m.state == StateReviewing {
		s.WriteString(RenderHelp("enter: ask follow-up • q: quit"))
	} else if m.state == StateChatting {
		s.WriteString(RenderHelp("enter: send • esc: back • q: quit"))
	} else {
		s.WriteString(RenderHelp("q: quit"))
	}

	return s.String()
}

// Run starts the Bubbletea program
func Run(reviewCtx *appcontext.ReviewContext, client *gemini.Client) error {
	model := NewModel(reviewCtx, client)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running UI: %w", err)
	}

	return nil
}

// RunSimple runs a simple non-interactive review
func RunSimple(ctx context.Context, reviewCtx *appcontext.ReviewContext, client *gemini.Client) error {
	// Initialize chat
	client.StartChat(appcontext.GetSystemPrompt())

	fmt.Println(RenderTitle("🔍 Go Code Review"))
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

	return nil
}

