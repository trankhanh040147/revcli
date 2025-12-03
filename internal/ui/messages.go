package ui

import (
	tea "github.com/charmbracelet/bubbletea"
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

