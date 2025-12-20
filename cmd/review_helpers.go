package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/trankhanh040147/revcli/internal/config"
	appcontext "github.com/trankhanh040147/revcli/internal/context"
	"github.com/trankhanh040147/revcli/internal/gemini"
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

// initializeClient initializes and returns a Gemini API client with the given model
func initializeClient(ctx context.Context, apiKey, model string) (*gemini.Client, error) {
	cfg, err := preset.LoadConfig()
	if err != nil {
		log.Printf("warn: failed to load configuration, proceeding with defaults: %v", err)
		cfg = nil
	}
	client, err := gemini.NewClient(ctx, apiKey, model, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	return client, nil
}

// initializeAPIClient initializes and returns a Gemini API client
func initializeAPIClient(ctx context.Context, apiKey, model string) (*gemini.Client, error) {
	return initializeClient(ctx, apiKey, model)
}

// initializeFlashClient initializes and returns a Gemini Flash API client for prune operations
func initializeFlashClient(ctx context.Context, apiKey string) (*gemini.Client, error) {
	return initializeClient(ctx, apiKey, config.ModelGeminiFlash)
}
