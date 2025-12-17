package ui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// pruneFileCmd creates a command to prune a file
func pruneFileCmd(ctx context.Context, apiKey, filePath, content string) tea.Cmd {
	return func() tea.Msg {
		if apiKey == "" {
			return PruneFileMsg{
				FilePath: filePath,
				Err:      fmt.Errorf("API key not set"),
			}
		}

		summary, err := PruneFile(apiKey, filePath, content)
		return PruneFileMsg{
			FilePath: filePath,
			Summary:  summary,
			Err:      err,
		}
	}
}

