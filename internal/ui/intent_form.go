package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	appcontext "github.com/trankhanh040147/plancli/internal/context"
)

// CollectIntent collects user intent using a huh form
// Returns nil if skipped (non-interactive mode)
func CollectIntent(interactive bool) (*appcontext.Intent, error) {
	if !interactive {
		return nil, nil
	}

	var customInstruction string
	var focusAreas []string
	var negativeConstraints string
	var webSearchEnabled bool = true // Default to enabled

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title("Custom Instructions (Optional)").
				Description("Provide specific instructions for the review. Examples: 'Focus on error handling', 'Check for race conditions', 'Review API design'").
				Value(&customInstruction).
				CharLimit(500).
				Validate(func(s string) error {
					trimmed := strings.TrimSpace(s)
					if len(trimmed) > 0 && len(trimmed) < 10 {
						return fmt.Errorf("instructions should be at least 10 characters (or leave empty)")
					}
					return nil
				}),

			huh.NewMultiSelect[string]().
				Title("Focus Areas").
				Description("Select areas to focus on (space to toggle, enter to confirm). Leave empty to review everything.").
				Options(
					huh.NewOption("Security", "security"),
					huh.NewOption("Performance", "performance"),
					huh.NewOption("Logic", "logic"),
					huh.NewOption("Style", "style"),
					huh.NewOption("Typo", "typo"),
					huh.NewOption("Naming", "naming"),
				).
				Value(&focusAreas),

			huh.NewText().
				Title("Negative Constraints (Optional)").
				Description("What should be ignored? Separate multiple items with commas. Examples: 'variable names', 'style issues', 'documentation'").
				Placeholder("e.g., 'variable names', 'style issues'").
				Value(&negativeConstraints).
				CharLimit(200),

			huh.NewConfirm().
				Title("Enable Web Search").
				Description("Allow Gemini to search the web for additional context (default: enabled)").
				Value(&webSearchEnabled),
		),
	).WithTheme(huh.ThemeCatppuccin()).
		WithWidth(80).
		WithShowHelp(true).
		WithShowErrors(true)

	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("failed to collect intent: %w", err)
	}

	// Parse negative constraints (split by comma or newline)
	var negativeList []string
	if negativeConstraints != "" {
		// Split by comma or newline
		parts := strings.FieldsFunc(negativeConstraints, func(r rune) bool {
			return r == ',' || r == '\n'
		})
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				negativeList = append(negativeList, part)
			}
		}
	}

	intent := &appcontext.Intent{
		CustomInstruction:   strings.TrimSpace(customInstruction),
		FocusAreas:          focusAreas,
		NegativeConstraints: negativeList,
		WebSearchEnabled:    webSearchEnabled,
	}

	return intent, nil
}
