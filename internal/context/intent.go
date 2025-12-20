package context

import (
	"fmt"
	"strings"

	"github.com/trankhanh040147/revcli/internal/preset"
)

// Intent represents user's review intent and focus areas
type Intent struct {
	// CustomInstruction is optional user-provided custom instruction
	CustomInstruction string
	// FocusAreas are selected focus areas (security, performance, logic, style, typo, naming)
	FocusAreas []string
	// NegativeConstraints are things the user wants to ignore
	NegativeConstraints []string
	// WebSearchEnabled controls whether web search is enabled for Gemini requests (default: true)
	WebSearchEnabled bool
}

// BuildSystemPromptWithIntent builds the system prompt incorporating intent
func BuildSystemPromptWithIntent(basePrompt string, intent *Intent, presets map[string]*preset.Preset) string {
	if intent == nil {
		return basePrompt
	}

	var parts []string
	parts = append(parts, basePrompt)

	// Add focus area presets
	if len(intent.FocusAreas) > 0 {
		parts = append(parts, "\n\n---\n\n## Focus Areas\n")
		for _, area := range intent.FocusAreas {
			if preset, ok := presets[area]; ok {
				parts = append(parts, preset.Prompt)
				parts = append(parts, "\n\n")
			}
		}
	}

	// Add custom instruction
	if intent.CustomInstruction != "" {
		parts = append(parts, "\n\n---\n\n## Custom Instructions\n")
		parts = append(parts, intent.CustomInstruction)
		parts = append(parts, "\n")
	}

	// Add negative constraints
	if len(intent.NegativeConstraints) > 0 {
		parts = append(parts, "\n\n---\n\n## Negative Constraints\n")
		parts = append(parts, "User explicitly stated to ignore:\n")
		for _, constraint := range intent.NegativeConstraints {
			parts = append(parts, fmt.Sprintf("- %s\n", constraint))
		}
	}

	return strings.Join(parts, "")
}

// GetFocusAreaPresets returns preset map for focus areas
func GetFocusAreaPresets() (map[string]*preset.Preset, error) {
	presets := make(map[string]*preset.Preset)

	// Map focus areas to preset names
	focusAreaMap := map[string]string{
		"security":    "security",
		"performance": "performance",
		"logic":       "logic",
		"style":       "style",
		"typo":        "typo",
		"naming":      "naming",
	}

	for focusArea, presetName := range focusAreaMap {
		p, err := preset.Get(presetName)
		if err != nil {
			// If preset doesn't exist, skip it
			continue
		}
		presets[focusArea] = p
	}

	return presets, nil
}

