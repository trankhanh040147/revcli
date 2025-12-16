package ui

import (
	"github.com/charmbracelet/bubbles/key"
)

// KeyMap defines all keybindings for the TUI
type KeyMap struct {
	// Global
	Quit      key.Binding
	ForceQuit key.Binding

	// Navigation (reviewing mode)
	Up           key.Binding
	Down         key.Binding
	Top          key.Binding
	Bottom       key.Binding
	HalfPageDown key.Binding
	HalfPageUp   key.Binding
	PageDown     key.Binding
	PageUp       key.Binding

	// Search
	Search      key.Binding
	NextMatch   key.Binding
	PrevMatch   key.Binding
	ToggleMode  key.Binding
	SearchEnter key.Binding
	SearchEsc   key.Binding

	// Help
	Help key.Binding

	// Chat
	EnterChat     key.Binding
	ExitChat      key.Binding
	SendMessage   key.Binding
	PrevPrompt    key.Binding
	NextPrompt    key.Binding
	CancelRequest key.Binding

	// Yank
	YankReview key.Binding
	YankLast   key.Binding
}

// DefaultKeyMap returns the default keymap
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Global
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
		ForceQuit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "force quit"),
		),

		// Navigation
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/↑", "scroll up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/↓", "scroll down"),
		),
		Top: key.NewBinding(
			key.WithKeys("g", "home"),
			key.WithHelp("g/Home", "go to top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G", "end"),
			key.WithHelp("G/End", "go to bottom"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "half page down"),
		),
		HalfPageUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("ctrl+u", "half page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("ctrl+f", "pgdown"),
			key.WithHelp("ctrl+f/PgDn", "page down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("ctrl+b", "pgup"),
			key.WithHelp("ctrl+b/PgUp", "page up"),
		),

		// Search
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "start search"),
		),
		NextMatch: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "next match"),
		),
		PrevMatch: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("N", "previous match"),
		),
		ToggleMode: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "toggle mode"),
		),
		SearchEnter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm search"),
		),
		SearchEsc: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "exit search"),
		),

		// Help
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "show help"),
		),

		// Chat
		EnterChat: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "enter chat"),
		),
		ExitChat: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "exit chat"),
		),
		SendMessage: key.NewBinding(
			key.WithKeys("alt+enter"),
			key.WithHelp("alt+enter", "send message"),
		),
		PrevPrompt: key.NewBinding(
			key.WithKeys("ctrl+p"),
			key.WithHelp("ctrl+p", "previous prompt"),
		),
		NextPrompt: key.NewBinding(
			key.WithKeys("ctrl+n"),
			key.WithHelp("ctrl+n", "next prompt"),
		),
		CancelRequest: key.NewBinding(
			key.WithKeys("ctrl+x"),
			key.WithHelp("ctrl+x", "cancel request"),
		),

		// Yank
		YankReview: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "yank review"),
		),
		YankLast: key.NewBinding(
			key.WithKeys("Y"),
			key.WithHelp("Y", "yank last"),
		),
	}
}
