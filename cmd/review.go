package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/trankhanh040147/revcli/internal/config"
	appcontext "github.com/trankhanh040147/revcli/internal/context"
	"github.com/trankhanh040147/revcli/internal/ui"
)

var (
	staged        bool
	model         string
	force         bool
	interactive   bool
	baseBranch    string
	presetName    string
	presetReplace bool
)

// reviewCmd represents the review command
var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Review code changes using Gemini AI",
	Long: `Analyzes your git diff and provides an intelligent code review.
Uses Google's Gemini LLM to detect bugs, suggest optimizations,
and ensure idiomatic Go practices.

Examples:
  # Review staged changes with interactive chat
  revcli review --staged

  # Review changes against main branch
  revcli review --base main

  # Review all uncommitted changes with a specific model
  revcli review --model gemini-2.5-pro

  # Non-interactive mode (just print the review)
  revcli review --no-interactive

  # Skip secret detection check
  revcli review --force

  # Use preset with replace mode (replaces base prompt)
  revcli review --preset quick --preset-replace
  revcli review -p quick -R`,
	RunE: runReview,
}

func init() {
	rootCmd.AddCommand(reviewCmd)

	reviewCmd.Flags().BoolVarP(&staged, "staged", "s", false, "Review only staged changes (git diff --staged)")
	reviewCmd.Flags().StringVarP(&baseBranch, "base", "b", "", "Base branch/commit to compare against (e.g., main, develop, abc123)")
	reviewCmd.Flags().StringVarP(&model, "model", "m", "gemini-2.5-pro", "Gemini model to use (gemini-2.5-pro, gemini-2.5-flash, etc.)")
	reviewCmd.Flags().BoolVarP(&force, "force", "f", false, "Skip secret detection and proceed anyway")
	reviewCmd.Flags().BoolVarP(&interactive, "interactive", "i", true, "Enable interactive chat mode")
	reviewCmd.Flags().BoolP("no-interactive", "I", false, "Disable interactive chat mode")
	reviewCmd.Flags().StringVarP(&presetName, "preset", "p", "", "Review preset (quick, strict, security, performance, logic, style, typo, naming)")
	reviewCmd.Flags().BoolVarP(&presetReplace, "preset-replace", "R", false, "Replace base prompt with preset prompt instead of appending")
}

func runReview(cmd *cobra.Command, args []string) error {
	// Check for API key
	apiKey := GetAPIKey()
	if apiKey == "" {
		return fmt.Errorf("%s is required. Set it via environment variable or --api-key flag", config.EnvGeminiAPIKey)
	}

	// Handle --no-interactive flag
	if cmd.Flags().Changed("no-interactive") {
		interactive = false
	}

	// Create context
	ctx := context.Background()

	// Validate mutually exclusive flags
	if staged && baseBranch != "" {
		return fmt.Errorf("cannot use --staged and --base together. Choose one")
	}

	// Load preset: use specified preset or default preset
	activePreset, err := loadActivePreset(presetName, presetReplace)
	if err != nil {
		return err
	}

	// Step 0: Collect intent (if interactive)
	var intent *appcontext.Intent
	if interactive {
		fmt.Println(ui.RenderTitle("🔍 Code Review"))
		fmt.Println()
		fmt.Println("Configure your review intent (press Ctrl+C to skip)...")
		fmt.Println()
		var err error
		intent, err = ui.CollectIntent(interactive)
		if err != nil {
			return fmt.Errorf("failed to collect intent: %w", err)
		}
		fmt.Println()
	}

	// Step 1: Build the review context
	printReviewHeader(os.Stdout, activePreset, baseBranch, staged)

	builder := appcontext.NewBuilder(staged, force, baseBranch)
	reviewCtx, err := buildReviewContext(builder, intent)
	if err != nil {
		// Check if it's a secrets error using errors.Is/As
		var secretsErr appcontext.SecretsError
		if errors.As(err, &secretsErr) {
			if printErr := printSecretsWarning(os.Stdout, secretsErr.Matches); printErr != nil {
				return printErr
			}
			return ErrSecretsDetected
		}
		return fmt.Errorf("failed to build review context: %w", err)
	}

	// Check if there are changes to review
	if !reviewCtx.HasChanges() {
		fmt.Println(ui.RenderWarning("No changes detected. Make sure you have uncommitted changes."))
		return nil
	}

	// Print detailed summary with file list
	printContextSummary(os.Stdout, reviewCtx)

	// Step 2: Initialize Gemini client
	fmt.Println("Connecting to Gemini API...")
	client, err := initializeAPIClient(ctx, apiKey, model)
	if err != nil {
		return err
	}

	fmt.Println(ui.RenderSuccess(fmt.Sprintf("Connected to %s", client.GetModelID())))
	fmt.Println()

	// Initialize flash client for prune operations
	flashClient, err := initializeFlashClient(ctx, apiKey)
	if err != nil {
		// Log warning but continue - pruning will fail gracefully if needed
		log.Printf("warning: failed to create flash client: %v", err)
		flashClient = nil
	}

	// Step 3: Run the review
	if interactive {
		// Interactive TUI mode
		return ui.Run(reviewCtx, client, flashClient, activePreset, apiKey)
	}

	// Non-interactive mode
	return ui.RunSimple(ctx, os.Stdout, reviewCtx, client, activePreset)
}
