package ui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/trankhanh040147/revcli/internal/gemini"
)

// pruneFileCmd creates a command to prune a file
func pruneFileCmd(ctx context.Context, flashClient *gemini.Client, filePath, content string) tea.Cmd {
	return func() tea.Msg {
		if flashClient == nil {
			return PruneFileMsg{
				FilePath: filePath,
				Err:      fmt.Errorf("flash client not initialized"),
			}
		}

		summary, err := PruneFile(ctx, flashClient, filePath, content)
		return PruneFileMsg{
			FilePath: filePath,
			Summary:  summary,
			Err:      err,
		}
	}
}

