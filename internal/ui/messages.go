package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// ReviewStartMsg signals that a review has started
type ReviewStartMsg struct{}

// ReviewCompleteMsg contains the completed review response
type ReviewCompleteMsg struct {
	Response string
}

// ReviewErrorMsg contains an error from the review process
type ReviewErrorMsg struct {
	Err error
}

// StreamChunkMsg contains a chunk of streamed response
type StreamChunkMsg struct {
	Chunk string
}

// StreamDoneMsg signals that streaming is complete
type StreamDoneMsg struct {
	FullResponse string
}

// streamChunkCmd creates a command to listen for chunks from a channel
// It will block until a chunk arrives, so it should be used in a goroutine-safe way
func streamChunkCmd(chunkChan chan string) tea.Cmd {
	return func() tea.Msg {
		chunk, ok := <-chunkChan
		if !ok {
			// Channel closed, return nil to stop listening
			return nil
		}
		return StreamChunkMsg{Chunk: chunk}
	}
}

// streamDoneCmd creates a command to listen for completion from a channel
func streamDoneCmd(doneChan chan string) tea.Cmd {
	return func() tea.Msg {
		fullResponse, ok := <-doneChan
		if !ok {
			return nil
		}
		return StreamDoneMsg{FullResponse: fullResponse}
	}
}

// streamErrorCmd creates a command to listen for errors from a channel
func streamErrorCmd(errChan chan error) tea.Cmd {
	return func() tea.Msg {
		err, ok := <-errChan
		if !ok {
			return nil
		}
		return ReviewErrorMsg{Err: err}
	}
}

// ChatResponseMsg contains a response to a follow-up question
type ChatResponseMsg struct {
	Response string
}

// ChatErrorMsg contains an error from a chat interaction
type ChatErrorMsg struct {
	Err error
}

// SendMessageCmd creates a command to send a message to the LLM
func SendMessageCmd(message string, sendFunc func(string) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return sendFunc(message)
	}
}

// TickMsg is used for spinner animation
type TickMsg struct{}

// QuitMsg signals the program should quit
type QuitMsg struct{}

// YankType represents the type of yanked content
type YankType int

const (
	YankTypeReview YankType = iota
	YankTypeLastResponse
)

// String returns the string representation of YankType
func (t YankType) String() string {
	switch t {
	case YankTypeReview:
		return "review"
	case YankTypeLastResponse:
		return "last response"
	default:
		return "unknown"
	}
}

// YankMsg signals that content was yanked to clipboard
type YankMsg struct {
	Type YankType
}

// YankFeedbackMsg signals that yank feedback should be cleared
type YankFeedbackMsg struct{}

// ClearYankFeedbackCmd creates a command to clear yank feedback after a delay
func ClearYankFeedbackCmd(duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(t time.Time) tea.Msg {
		return YankFeedbackMsg{}
	})
}

// yankTimeoutMsg signals that the yank combo timeout has elapsed
type yankTimeoutMsg struct{}

// PruneFileMsg contains the result of pruning a file
type PruneFileMsg struct {
	FilePath string
	Summary  string
	Err      error
}
