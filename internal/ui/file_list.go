package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	appcontext "github.com/trankhanh040147/revcli/internal/context"
)

// FileListItem represents an item in the file list
type FileListItem struct {
	Path   string
	Size   int
	Pruned bool
}

// Title returns the display title for the item
func (f FileListItem) Title() string {
	prunedIndicator := ""
	if f.Pruned {
		prunedIndicator = " ✓"
	}
	return fmt.Sprintf("%s%s", f.Path, prunedIndicator)
}

// Description returns the description (file size)
func (f FileListItem) Description() string {
	return formatFileSize(f.Size)
}

// FilterValue returns the value to filter by
func (f FileListItem) FilterValue() string {
	return f.Path
}

// formatFileSize formats bytes into human readable format
func formatFileSize(bytes int) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := int64(bytes) / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// NewFileListModel creates a new file list model from ReviewContext
func NewFileListModel(reviewCtx *appcontext.ReviewContext) list.Model {
	items := make([]list.Item, 0, len(reviewCtx.FileContents))

	for path, content := range reviewCtx.FileContents {
		// PrunedFiles is always initialized in builder.go:91
		_, pruned := reviewCtx.PrunedFiles[path]
		items = append(items, FileListItem{
			Path:   path,
			Size:   len(content),
			Pruned: pruned,
		})
	}

	// Create list with custom styling
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Files to Review"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	// Custom styles
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED")).
		Bold(true).
		MarginBottom(1)

	l.Styles.NoItems = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Italic(true)

	return l
}

// UpdateFileListModel updates the file list model with current pruned state
func UpdateFileListModel(l list.Model, reviewCtx *appcontext.ReviewContext) list.Model {
	items := make([]list.Item, 0, len(reviewCtx.FileContents))

	for path, content := range reviewCtx.FileContents {
		// PrunedFiles is always initialized in builder.go:91
		_, pruned := reviewCtx.PrunedFiles[path]
		items = append(items, FileListItem{
			Path:   path,
			Size:   len(content),
			Pruned: pruned,
		})
	}

	l.SetItems(items)
	return l
}

// GetSelectedFile returns the currently selected file path
func GetSelectedFile(l list.Model) (string, bool) {
	item := l.SelectedItem()
	if item == nil {
		return "", false
	}
	fileItem, ok := item.(FileListItem)
	if !ok {
		return "", false
	}
	return fileItem.Path, true
}
