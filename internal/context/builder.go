package context

import (
	"fmt"

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
	// Intent is the user's review intent and focus areas
	Intent *Intent
	// PrunedFiles maps file paths to their summaries (for token optimization)
	PrunedFiles map[string]string
}

// Builder constructs the review context from git changes
type Builder struct {
	staged     bool
	force      bool
	baseBranch string
	intent     *Intent
}

// NewBuilder creates a new context builder
func NewBuilder(staged, force bool, baseBranch string) *Builder {
	return &Builder{
		staged:     staged,
		force:      force,
		baseBranch: baseBranch,
		intent:     nil,
	}
}

// WithIntent sets the intent for the builder
func (b *Builder) WithIntent(intent *Intent) *Builder {
	b.intent = intent
	return b
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
		return nil, SecretsError{Matches: filterResult.SecretsFound}
	}

	// Step 4: Filter the diff to remove ignored files
	filteredDiff := filter.FilterDiff(diffResult.RawDiff)

	// Step 5: Build the prompt (with pruning support)
	userPrompt := prompt.BuildReviewPromptWithPruning(filteredDiff, filterResult.FilteredFiles, nil)

	// Step 6: Estimate tokens
	estimatedTokens := prompt.EstimateTokens(userPrompt)

	return &ReviewContext{
		RawDiff:         filteredDiff,
		FileContents:    filterResult.FilteredFiles,
		IgnoredFiles:    filterResult.IgnoredFiles,
		SecretsFound:    filterResult.SecretsFound,
		UserPrompt:      userPrompt,
		EstimatedTokens: estimatedTokens,
		Intent:          b.intent,
		PrunedFiles:     make(map[string]string),
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

// GetSystemPromptWithIntent returns the system prompt incorporating intent and preset
func GetSystemPromptWithIntent(intent *Intent, presetPrompt string, presetReplace bool) string {
	// Start with base prompt or preset
	var basePrompt string
	if presetReplace && presetPrompt != "" {
		basePrompt = presetPrompt
	} else {
		basePrompt = GetSystemPrompt()
		if presetPrompt != "" && !presetReplace {
			basePrompt = basePrompt + "\n\n---\n\n" + presetPrompt
		}
	}

	// If no intent, return base prompt
	if intent == nil {
		return basePrompt
	}

	// Get focus area presets
	focusPresets, err := GetFocusAreaPresets()
	if err != nil {
		// If we can't get presets, just use base prompt with intent
		return BuildSystemPromptWithIntent(basePrompt, intent, nil)
	}

	return BuildSystemPromptWithIntent(basePrompt, intent, focusPresets)
}

// HasChanges returns true if there are changes to review
func (rc *ReviewContext) HasChanges() bool {
	return len(rc.FileContents) > 0 || rc.RawDiff != ""
}
