package context

import (
	"fmt"
	"strings"

	"github.com/trankhanh040147/revcli/internal/filter"
	"github.com/trankhanh040147/revcli/internal/git"
	"github.com/trankhanh040147/revcli/internal/preset"
	"github.com/trankhanh040147/revcli/internal/prompt"
)

// ReviewContext contains all the data needed for a code review
type ReviewContext struct {
	// RawDiff is the filtered git diff
	RawDiff string
	// FileContents maps file paths to their content
	FileContents map[string]string
	// IgnoredFiles lists files that were filtered out
	IgnoredFiles []string
	// SecretsFound contains any potential secrets detected
	SecretsFound []filter.SecretMatch
	// UserPrompt is the assembled prompt for the LLM
	UserPrompt string
	// EstimatedTokens is the rough token count
	EstimatedTokens int
}

// Builder constructs the review context from git changes
type Builder struct {
	staged     bool
	force      bool
	baseBranch string
}

// NewBuilder creates a new context builder
func NewBuilder(staged, force bool, baseBranch string) *Builder {
	return &Builder{
		staged:     staged,
		force:      force,
		baseBranch: baseBranch,
	}
}

// Build gathers git changes and assembles the review context
func (b *Builder) Build() (*ReviewContext, error) {
	// Step 1: Get git diff and file contents
	diffResult, err := git.GetDiff(b.staged, b.baseBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to get git diff: %w", err)
	}

	// Step 2: Filter files and scan for secrets
	filterResult := filter.Filter(diffResult.ModifiedFiles, diffResult.RawDiff)

	// Step 3: Check for secrets (unless force is enabled)
	if filterResult.HasSecrets() && !b.force {
		return &ReviewContext{
			SecretsFound: filterResult.SecretsFound,
		}, fmt.Errorf("potential secrets detected in code. Use --force to proceed anyway")
	}

	// Step 4: Filter the diff to remove ignored files
	filteredDiff := filter.FilterDiff(diffResult.RawDiff)

	// Step 5: Build the prompt
	userPrompt := prompt.BuildReviewPrompt(filteredDiff, filterResult.FilteredFiles)

	// Step 6: Estimate tokens
	estimatedTokens := prompt.EstimateTokens(userPrompt)

	return &ReviewContext{
		RawDiff:         filteredDiff,
		FileContents:    filterResult.FilteredFiles,
		IgnoredFiles:    filterResult.IgnoredFiles,
		SecretsFound:    filterResult.SecretsFound,
		UserPrompt:      userPrompt,
		EstimatedTokens: estimatedTokens,
	}, nil
}

// BuildFromDiff creates a review context from an existing diff string
func BuildFromDiff(rawDiff string, files map[string]string) *ReviewContext {
	filterResult := filter.Filter(files, rawDiff)
	filteredDiff := filter.FilterDiff(rawDiff)
	userPrompt := prompt.BuildReviewPrompt(filteredDiff, filterResult.FilteredFiles)
	estimatedTokens := prompt.EstimateTokens(userPrompt)

	return &ReviewContext{
		RawDiff:         filteredDiff,
		FileContents:    filterResult.FilteredFiles,
		IgnoredFiles:    filterResult.IgnoredFiles,
		SecretsFound:    filterResult.SecretsFound,
		UserPrompt:      userPrompt,
		EstimatedTokens: estimatedTokens,
	}
}

// GetSystemPrompt returns the system prompt for the LLM
// Checks for custom system prompt file first, falls back to default if not found
func GetSystemPrompt() string {
	customPrompt, found, err := preset.LoadSystemPrompt()
	if err == nil && found {
		return customPrompt
	}
	// Fallback to default system prompt
	return prompt.SystemPrompt
}

// GetSystemPromptWithPreset returns the system prompt modified by a preset
// If replace is true, returns only the preset prompt (replacing the base prompt)
// If replace is false, appends the preset prompt to the base prompt (default behavior)
func GetSystemPromptWithPreset(presetPrompt string, replace bool) string {
	if replace {
		return presetPrompt
	}
	return prompt.SystemPrompt + "\n\n---\n\n" + presetPrompt
}

// Summary returns a summary of what will be reviewed
func (rc *ReviewContext) Summary() string {
	fileCount := len(rc.FileContents)
	ignoredCount := len(rc.IgnoredFiles)

	summary := fmt.Sprintf("📋 Review Context:\n")
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

// HasChanges returns true if there are changes to review
func (rc *ReviewContext) HasChanges() bool {
	return len(rc.FileContents) > 0 || rc.RawDiff != ""
}
