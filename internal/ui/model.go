package ui

import (
	"context"
	"fmt"
	"log"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/trankhanh040147/revcli/internal/app"
	appcontext "github.com/trankhanh040147/revcli/internal/context"
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

	// App and session
	app          *app.App
	sessionID    string
	rootCtx      context.Context    // Root context for cancellation chain
	activeCancel context.CancelFunc // Cancel function for currently active command

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
	ready            bool
	streaming        bool
	webSearchEnabled bool // Web search toggle for follow-up questions (default: true, resets per question)

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

	// Pruning state (per-file tracking)
	pruningFiles    map[string]bool               // Track which files are currently being pruned
	pruningSpinners map[string]spinner.Model      // Spinners for each file being pruned
	pruningCancels  map[string]context.CancelFunc // Cancel functions for each pruning operation

	// Keybindings
	keys KeyMap
}

// NewModel creates a new application model
func NewModel(reviewCtx *appcontext.ReviewContext, appInstance *app.App, sessionID string, p *preset.Preset) *Model {
	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))

	// todo: check cannot type on this text area
	// Create textarea for chat input
	ta := textarea.New()
	ta.Placeholder = "Ask a follow-up question..."
	ta.Focus()
	ta.CharLimit = 1000
	ta.SetWidth(80)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false

	// Custom textarea styling to avoid white background and row highlighting
	ta.SetStyles(textarea.Styles{
		Focused: textarea.StyleState{
			Base: lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7C3AED")),
			CursorLine: lipgloss.NewStyle(),
		},
		Blurred: textarea.StyleState{
			Base: lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#4B5563")),
			CursorLine: lipgloss.NewStyle(),
		},
	})

	// Create search input
	si := textinput.New()
	si.Placeholder = "Search..."
	si.CharLimit = 100
	si.SetWidth(40)

	// Create renderer (with fallback if it fails)
	renderer, err := NewRenderer()
	if err != nil {
		log.Printf("warning: failed to initialize markdown renderer: %v", err)
		renderer = &Renderer{}
	}

	// Create root context
	rootCtx := context.Background()

	// Create file list
	fileListModel := NewFileListModel(reviewCtx, nil)

	return &Model{
		state:              StateLoading,
		reviewCtx:          reviewCtx,
		app:                appInstance,
		sessionID:          sessionID,
		rootCtx:            rootCtx,
		activeCancel:       nil,
		preset:             p,
		spinner:            s,
		textarea:           ta,
		searchInput:        si,
		fileList:           fileListModel,
		search:             NewSearchState(),
		renderer:           renderer,
		ready:              false,
		streaming:          false,
		webSearchEnabled:   true, // Default to enabled
		promptHistory:      []string{},
		promptHistoryIndex: -1,
		pruningFiles:       make(map[string]bool),
		pruningSpinners:    make(map[string]spinner.Model),
		pruningCancels:     make(map[string]context.CancelFunc),
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
func Run(reviewCtx *appcontext.ReviewContext, appInstance *app.App, sessionID string, p *preset.Preset) error {
	model := NewModel(reviewCtx, appInstance, sessionID, p)
	program := tea.NewProgram(model)

	// Subscribe app events to TUI
	go appInstance.Subscribe(program)

	if _, err := program.Run(); err != nil {
		return fmt.Errorf("error running UI: %w", err)
	}

	return nil
}

// resetStreamState resets streaming state and clears all stream channels
func (m *Model) resetStreamState() {
	m.streaming = false
	m.streamChunkChan = nil
	m.streamErrChan = nil
	m.streamDoneChan = nil
}

// transitionToErrorOnCancel handles state transition when a request is cancelled
func (m *Model) transitionToErrorOnCancel() {
	m.resetStreamState()
	// Always transition to error state with appropriate message
	if m.reviewResponse != "" {
		m.state = StateError
		m.errorMsg = "Request cancelled (partial response available)"
		m.updateViewport() // Update viewport to show partial response
	} else {
		m.state = StateError
		m.errorMsg = "Request cancelled"
	}
}
