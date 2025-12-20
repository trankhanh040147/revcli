package ui

import "time"

// UI feedback durations
const (
	YankFeedbackDuration      = 2 * time.Second
	PruneErrorFeedbackDuration = 3 * time.Second
	YankChordTimeout          = 300 * time.Millisecond
)

// PruneFilePromptTemplate is the template for file summarization prompts
const PruneFilePromptTemplate = `Summarize this code file in one sentence. Focus on what the file does, its main purpose, and key functionality. Be concise.

File: %s

%s`

