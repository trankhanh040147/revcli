package ui

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	appcontext "github.com/trankhanh040147/revcli/internal/context"
	"github.com/trankhanh040147/revcli/internal/gemini"
	"github.com/trankhanh040147/revcli/internal/preset"
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
	reviewResponse string
	rawContent     string // Original content without search highlighting
	errorMsg       string
	chatHistory    []ChatMessage

	// Dimensions
	width  int
	height int

	// Flags
	ready     bool
	streaming bool

	// Yank state
	yankFeedback string // Feedback message for yank
	lastKeyWasY  bool   // For detecting "yy" combo to yank entire review (code block navigation removed in v0.3.1)

	// Prompt history state
	promptHistory      []string // History of sent prompts
	promptHistoryIndex int      // Current position in history (-1 for new prompt)

	// Keybindings
	keys KeyMap
}

// ChatRole represents the role of a chat message
type ChatRole int

const (
	ChatRoleUser ChatRole = iota
	ChatRoleAssistant
)

// String returns the string representation of ChatRole
func (r ChatRole) String() string {
	switch r {
	case ChatRoleUser:
		return "user"
	case ChatRoleAssistant:
		return "assistant"
	default:
		return "unknown"
	}
}

// ChatMessage represents a message in the chat history
type ChatMessage struct {
	Role    ChatRole
	Content string
}

// NewModel creates a new application model
func NewModel(reviewCtx *appcontext.ReviewContext, client *gemini.Client, p *preset.Preset) *Model {
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
		fmt.Fprintf(os.Stderr, "warning: failed to initialize markdown renderer: %v\n", err)
		renderer = &Renderer{}
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	return &Model{
		state:              StateLoading,
		reviewCtx:          reviewCtx,
		client:             client,
		ctx:                ctx,
		cancel:             cancel,
		preset:             p,
		spinner:            s,
		textarea:           ta,
		searchInput:        si,
		search:             NewSearchState(),
		renderer:           renderer,
		ready:              false,
		streaming:          false,
		promptHistory:      []string{},
		promptHistoryIndex: -1,
		keys:               DefaultKeyMap(),
	}
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.startReview(),
	)
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
func RunSimple(ctx context.Context, w io.Writer, reviewCtx *appcontext.ReviewContext, client *gemini.Client, p *preset.Preset) error {
	// Initialize chat with system prompt (with preset if specified)
	systemPrompt := appcontext.GetSystemPrompt()
	if p != nil {
		systemPrompt = appcontext.GetSystemPromptWithPreset(p.Prompt, p.Replace)
	}
	client.StartChat(systemPrompt)

	fmt.Fprintln(w, RenderTitle("🔍 Code Review"))
	fmt.Fprintln(w, RenderSubtitle(reviewCtx.Summary()))
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Analyzing your code changes...")
	fmt.Fprintln(w)

	// Create renderer
	renderer, err := NewRenderer()
	if err != nil {
		return fmt.Errorf("failed to create renderer: %w", err)
	}

	// Stream the response
	startTime := time.Now()
	response, err := client.SendMessageStream(ctx, reviewCtx.UserPrompt, func(chunk string) {
		fmt.Fprint(w, chunk)
	})
	if err != nil {
		return fmt.Errorf("review failed: %w", err)
	}

	// Render the full response with markdown
	fmt.Fprintln(w)
	fmt.Fprintln(w, RenderDivider(80))
	fmt.Fprintln(w)

	rendered, err := renderer.RenderMarkdown(response)
	if err != nil {
		fmt.Fprintln(w, response)
	} else {
		fmt.Fprintln(w, rendered)
	}

	elapsed := time.Since(startTime)
	fmt.Fprintln(w)
	fmt.Fprintln(w, RenderSuccess(fmt.Sprintf("Review completed in %s", elapsed.Round(time.Millisecond))))

	// Display token usage
	if usage := client.GetLastUsage(); usage != nil {
		fmt.Fprintln(w, RenderTokenUsage(usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens))
	}

	return nil
}
