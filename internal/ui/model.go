package ui

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
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
	StateFileList
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
	apiKey string // API key for prune operations

	// Review preset
	preset *preset.Preset

	// UI components
	spinner     spinner.Model
	viewport    viewport.Model
	textarea    textarea.Model
	searchInput textinput.Model
	fileList    list.Model
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

	// Streaming channels (set during StreamStartMsg)
	streamChunkChan chan string
	streamErrChan   chan error
	streamDoneChan  chan string

	// Yank state
	yankFeedback string // Feedback message for yank
	lastKeyWasY  bool   // For detecting "yy" combo to yank entire review (code block navigation removed in v0.3.1)

	// Prompt history state
	promptHistory      []string // History of sent prompts
	promptHistoryIndex int      // Current position in history (-1 for new prompt)

	// Keybindings
	keys KeyMap
}

// NewModel creates a new application model
func NewModel(reviewCtx *appcontext.ReviewContext, client *gemini.Client, p *preset.Preset, apiKey string) *Model {
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

	// Create file list
	fileListModel := NewFileListModel(reviewCtx)

	return &Model{
		state:              StateLoading,
		reviewCtx:          reviewCtx,
		client:             client,
		ctx:                ctx,
		cancel:             cancel,
		apiKey:             apiKey,
		preset:             p,
		spinner:            s,
		textarea:           ta,
		searchInput:        si,
		fileList:           fileListModel,
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
func Run(reviewCtx *appcontext.ReviewContext, client *gemini.Client, p *preset.Preset, apiKey string) error {
	model := NewModel(reviewCtx, client, p, apiKey)
	program := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := program.Run(); err != nil {
		return fmt.Errorf("error running UI: %w", err)
	}

	return nil
}
