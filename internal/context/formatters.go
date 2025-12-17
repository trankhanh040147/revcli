package context

import (
	"fmt"
	"strings"

	"github.com/trankhanh040147/revcli/internal/prompt"
)

// Summary returns a summary of what will be reviewed
func (rc *ReviewContext) Summary() string {
	fileCount := len(rc.FileContents)
	ignoredCount := len(rc.IgnoredFiles)

	summary := "📋 Review Context:\n"
	summary += fmt.Sprintf("   • Files to review: %d\n", fileCount)
	if ignoredCount > 0 {
		summary += fmt.Sprintf("   • Files ignored: %d\n", ignoredCount)
	}
	summary += fmt.Sprintf("   • Estimated tokens: ~%d\n", rc.EstimatedTokens)

	// Token warning
	if warning := prompt.MaxTokenWarning(rc.UserPrompt, 100000); warning != "" {
		summary += fmt.Sprintf("   ⚠️  %s\n", warning)
	}

	return summary
}

// DetailedSummary returns a detailed summary including file list
func (rc *ReviewContext) DetailedSummary() string {
	var sb strings.Builder

	sb.WriteString("📋 Review Context\n")
	sb.WriteString("─────────────────\n\n")

	// Files to review
	sb.WriteString("📁 Files to review:\n")
	if len(rc.FileContents) == 0 {
		sb.WriteString("   (none)\n")
	} else {
		totalSize := 0
		for path, content := range rc.FileContents {
			size := len(content)
			totalSize += size
			sb.WriteString(fmt.Sprintf("   • %s (%s)\n", path, formatBytes(size)))
		}
		sb.WriteString(fmt.Sprintf("\n   Total: %d files, %s\n", len(rc.FileContents), formatBytes(totalSize)))
	}

	// Ignored files
	if len(rc.IgnoredFiles) > 0 {
		sb.WriteString("\n🚫 Ignored files:\n")
		for _, path := range rc.IgnoredFiles {
			sb.WriteString(fmt.Sprintf("   • %s\n", path))
		}
	}

	// Token estimate
	sb.WriteString(fmt.Sprintf("\n📊 Token Estimate: ~%d tokens\n", rc.EstimatedTokens))

	// Token warning
	if warning := prompt.MaxTokenWarning(rc.UserPrompt, 100000); warning != "" {
		sb.WriteString(fmt.Sprintf("⚠️  %s\n", warning))
	}

	return sb.String()
}

// formatBytes formats bytes into human readable format
func formatBytes(bytes int) string {
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

