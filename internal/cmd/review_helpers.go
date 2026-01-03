package cmd

import (
	"path/filepath"

	appcontext "github.com/trankhanh040147/revcli/internal/context"
	"github.com/trankhanh040147/revcli/internal/message"
	"github.com/trankhanh040147/revcli/internal/preset"
)

// loadActivePreset loads the active preset based on presetName or default preset
func loadActivePreset(presetName string, presetReplace bool) (*preset.Preset, error) {
	var activePreset *preset.Preset
	if presetName != "" {
		var err error
		activePreset, err = preset.Get(presetName)
		if err != nil {
			return nil, err
		}
	} else {
		// Try to load default preset
		defaultPresetName, err := preset.GetDefaultPreset()
		if err == nil && defaultPresetName != "" {
			activePreset, err = preset.Get(defaultPresetName)
			if err != nil {
				// Default preset doesn't exist anymore, ignore
				activePreset = nil
			}
		}
	}

	// Apply --preset-replace flag override if set
	if activePreset != nil && presetReplace {
		activePreset.Replace = true
	}

	return activePreset, nil
}

// buildReviewContext builds the review context from the builder and intent
func buildReviewContext(builder *appcontext.Builder, intent *appcontext.Intent) (*appcontext.ReviewContext, error) {
	if intent != nil {
		builder.WithIntent(intent)
	}
	return builder.Build()
}

// buildReviewPrompt builds the review prompt from context and preset
// TODO: using preset later
func buildReviewPrompt(reviewCtx *appcontext.ReviewContext, preset *preset.Preset) string {
	// Start with the user prompt from context
	prompt := reviewCtx.UserPrompt

	// If there are pruned files, the prompt should already include them
	// (handled by prompt.BuildReviewPromptWithPruning)
	return prompt
}

// buildAttachments converts review context files to message attachments
func buildAttachments(reviewCtx *appcontext.ReviewContext) []message.Attachment {
	var attachments []message.Attachment
	for filePath, content := range reviewCtx.FileContents {
		attachments = append(attachments, message.Attachment{
			FilePath: filePath,
			FileName: filepath.Base(filePath),
			MimeType: "text/plain",
			Content:  []byte(content),
		})
	}
	return attachments
}
