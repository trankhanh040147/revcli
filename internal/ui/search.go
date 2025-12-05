package ui

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// SearchMode represents the search behavior
type SearchMode int

const (
	SearchModeHighlight SearchMode = iota // Highlight matches, show all content
	SearchModeFilter                      // Show only lines with matches
)

// SearchState holds the current search state
type SearchState struct {
	Active       bool
	Query        string
	Mode         SearchMode
	Matches      []SearchMatch
	CurrentMatch int
}

// SearchMatch represents a match in the content
type SearchMatch struct {
	Line      int    // Line number (0-indexed)
	StartCol  int    // Start column in the line
	EndCol    int    // End column in the line
	LineText  string // The full line text
	MatchText string // The matched text
}

// NewSearchState creates a new search state
func NewSearchState() *SearchState {
	return &SearchState{
		Active:       false,
		Query:        "",
		Mode:         SearchModeHighlight,
		Matches:      nil,
		CurrentMatch: -1,
	}
}

// Search performs a case-insensitive search and returns matches
func (s *SearchState) Search(content string) {
	s.Matches = nil
	s.CurrentMatch = -1

	if s.Query == "" {
		return
	}

	// Create case-insensitive regex
	pattern, err := regexp.Compile("(?i)" + regexp.QuoteMeta(s.Query))
	if err != nil {
		return
	}

	lines := strings.Split(content, "\n")
	for lineNum, line := range lines {
		matches := pattern.FindAllStringIndex(line, -1)
		for _, match := range matches {
			s.Matches = append(s.Matches, SearchMatch{
				Line:      lineNum,
				StartCol:  match[0],
				EndCol:    match[1],
				LineText:  line,
				MatchText: line[match[0]:match[1]],
			})
		}
	}

	// Set to first match if any found
	if len(s.Matches) > 0 {
		s.CurrentMatch = 0
	}
}

// NextMatch moves to the next match
func (s *SearchState) NextMatch() {
	if len(s.Matches) == 0 {
		return
	}
	s.CurrentMatch = (s.CurrentMatch + 1) % len(s.Matches)
}

// PrevMatch moves to the previous match
func (s *SearchState) PrevMatch() {
	if len(s.Matches) == 0 {
		return
	}
	s.CurrentMatch--
	if s.CurrentMatch < 0 {
		s.CurrentMatch = len(s.Matches) - 1
	}
}

// CurrentMatchLine returns the line number of the current match
func (s *SearchState) CurrentMatchLine() int {
	if s.CurrentMatch < 0 || s.CurrentMatch >= len(s.Matches) {
		return -1
	}
	return s.Matches[s.CurrentMatch].Line
}

// FilteredLineIndex translates an original line number to its index in the filtered view.
// Returns -1 if the line is not in the filtered set.
func (s *SearchState) FilteredLineIndex(originalLine int) int {
	if len(s.Matches) == 0 {
		return -1
	}

	// Collect unique line numbers with matches in order
	seen := make(map[int]bool)
	var uniqueLines []int
	for _, match := range s.Matches {
		if !seen[match.Line] {
			seen[match.Line] = true
			uniqueLines = append(uniqueLines, match.Line)
		}
	}

	// Find the index of the original line in the filtered view
	for idx, line := range uniqueLines {
		if line == originalLine {
			return idx
		}
	}

	return -1
}

// ToggleMode switches between highlight and filter modes
func (s *SearchState) ToggleMode() {
	if s.Mode == SearchModeHighlight {
		s.Mode = SearchModeFilter
	} else {
		s.Mode = SearchModeHighlight
	}
}

// Reset clears the search state
func (s *SearchState) Reset() {
	s.Active = false
	s.Query = ""
	s.Matches = nil
	s.CurrentMatch = -1
}

// MatchCount returns the number of matches
func (s *SearchState) MatchCount() int {
	return len(s.Matches)
}

// MatchStatus returns a string describing the current match status
func (s *SearchState) MatchStatus() string {
	if len(s.Matches) == 0 {
		if s.Query != "" {
			return "No matches"
		}
		return ""
	}
	return formatMatchStatus(s.CurrentMatch+1, len(s.Matches), "")
}

// HighlightContent highlights all matches in the content
func (s *SearchState) HighlightContent(content string) string {
	if s.Query == "" || len(s.Matches) == 0 {
		return content
	}

	// Create highlight style
	highlightStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#FACC15")).
		Foreground(lipgloss.Color("#000000"))

	currentHighlightStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#F97316")).
		Foreground(lipgloss.Color("#000000")).
		Bold(true)

	// Create case-insensitive regex
	pattern, err := regexp.Compile("(?i)" + regexp.QuoteMeta(s.Query))
	if err != nil {
		return content
	}

	lines := strings.Split(content, "\n")
	matchIdx := 0

	for lineNum := range lines {
		line := lines[lineNum]
		matches := pattern.FindAllStringIndex(line, -1)

		if len(matches) == 0 {
			continue
		}

		// Rebuild line with highlights (process from end to preserve indices)
		for i := len(matches) - 1; i >= 0; i-- {
			match := matches[i]
			matchText := line[match[0]:match[1]]

			// Check if this is the current match
			var highlighted string
			if matchIdx == s.CurrentMatch {
				highlighted = currentHighlightStyle.Render(matchText)
			} else {
				highlighted = highlightStyle.Render(matchText)
			}

			line = line[:match[0]] + highlighted + line[match[1]:]
			matchIdx++
		}

		lines[lineNum] = line
	}

	return strings.Join(lines, "\n")
}

// FilterContent returns only lines containing matches
func (s *SearchState) FilterContent(content string) string {
	if s.Query == "" || len(s.Matches) == 0 {
		return content
	}

	// Collect unique line numbers with matches
	lineSet := make(map[int]bool)
	for _, match := range s.Matches {
		lineSet[match.Line] = true
	}

	lines := strings.Split(content, "\n")
	var filteredLines []string

	for lineNum, line := range lines {
		if lineSet[lineNum] {
			filteredLines = append(filteredLines, line)
		}
	}

	// Highlight the filtered content
	filteredContent := strings.Join(filteredLines, "\n")
	return s.HighlightContent(filteredContent)
}

// RenderSearchInput renders the search input bar
func RenderSearchInput(query string, matchCount, currentMatch int, mode SearchMode) string {
	searchStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3B82F6")).
		Bold(true)

	modeText := "highlight"
	if mode == SearchModeFilter {
		modeText = "filter"
	}

	var status string
	if query == "" {
		status = ""
	} else if matchCount == 0 {
		status = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Render(" (no matches)")
	} else {
		status = formatMatchStatus(currentMatch+1, matchCount, modeText)
	}

	return searchStyle.Render("/") + query + status
}

// formatMatchStatus formats the match status string
func formatMatchStatus(current, total int, mode string) string {
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF"))
	if mode == "" {
		return statusStyle.Render(" (" + itoa(current) + "/" + itoa(total) + ")")
	}
	return statusStyle.Render(" (" + itoa(current) + "/" + itoa(total) + " " + mode + ")")
}

// itoa converts int to string (simple implementation)
func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	var result []byte
	negative := n < 0
	if negative {
		n = -n
	}

	for n > 0 {
		result = append([]byte{byte('0' + n%10)}, result...)
		n /= 10
	}

	if negative {
		result = append([]byte{'-'}, result...)
	}

	return string(result)
}
