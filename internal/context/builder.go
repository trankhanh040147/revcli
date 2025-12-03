package context

import (
	"fmt"

	"github.com/trankhanh040147/go-rev-cli/internal/filter"
	"github.com/trankhanh040147/go-rev-cli/internal/git"
	"github.com/trankhanh040147/go-rev-cli/internal/prompt"
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
	staged bool
	force  bool
}

// NewBuilder creates a new context builder
func NewBuilder(staged, force bool) *Builder {
	return &Builder{
		staged: staged,
		force:  force,
	}
}

// Build gathers git changes and assembles the review context
func (b *Builder) Build() (*ReviewContext, error) {
	// Step 1: Get git diff and file contents
	diffResult, err := git.GetDiff(b.staged)
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
func GetSystemPrompt() string {
	return prompt.SystemPrompt
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

// HasChanges returns true if there are changes to review
func (rc *ReviewContext) HasChanges() bool {
	return len(rc.FileContents) > 0 || rc.RawDiff != ""
}

