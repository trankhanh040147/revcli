package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	appcontext "github.com/trankhanh040147/go-rev-cli/internal/context"
	"github.com/trankhanh040147/go-rev-cli/internal/filter"
	"github.com/trankhanh040147/go-rev-cli/internal/gemini"
	"github.com/trankhanh040147/go-rev-cli/internal/ui"
)

var (
	staged      bool
	model       string
	force       bool
	interactive bool
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
  go-rev-cli review --staged

  # Review all uncommitted changes with a specific model
  go-rev-cli review --model gemini-1.5-pro

  # Non-interactive mode (just print the review)
  go-rev-cli review --no-interactive

  # Skip secret detection check
  go-rev-cli review --force`,
	RunE: runReview,
}

func init() {
	rootCmd.AddCommand(reviewCmd)

	reviewCmd.Flags().BoolVar(&staged, "staged", false, "Review only staged changes (git diff --staged)")
	reviewCmd.Flags().StringVar(&model, "model", "gemini-1.5-flash", "Gemini model to use (gemini-1.5-flash or gemini-1.5-pro)")
	reviewCmd.Flags().BoolVar(&force, "force", false, "Skip secret detection and proceed anyway")
	reviewCmd.Flags().BoolVar(&interactive, "interactive", true, "Enable interactive chat mode")
	reviewCmd.Flags().BoolVar(&interactive, "no-interactive", false, "Disable interactive chat mode")
}

func runReview(cmd *cobra.Command, args []string) error {
	// Check for API key
	apiKey := GetAPIKey()
	if apiKey == "" {
		return fmt.Errorf("GEMINI_API_KEY is required. Set it via environment variable or --api-key flag")
	}

	// Handle --no-interactive flag
	if cmd.Flags().Changed("no-interactive") {
		interactive = false
	}

	// Create context
	ctx := context.Background()

	// Step 1: Build the review context
	fmt.Println(ui.RenderTitle("🔍 Go Code Review"))
	fmt.Println()
	fmt.Println("Gathering code changes...")

	builder := appcontext.NewBuilder(staged, force)
	reviewCtx, err := builder.Build()
	if err != nil {
		// Check if it's a secrets error
		if reviewCtx != nil && len(reviewCtx.SecretsFound) > 0 {
			printSecretsWarning(reviewCtx.SecretsFound)
			return fmt.Errorf("review aborted due to potential secrets")
		}
		return fmt.Errorf("failed to build review context: %w", err)
	}

	// Check if there are changes to review
	if !reviewCtx.HasChanges() {
		fmt.Println(ui.RenderWarning("No changes detected. Make sure you have uncommitted changes."))
		return nil
	}

	// Print summary
	fmt.Println(ui.RenderSuccess("Changes collected!"))
	fmt.Println(reviewCtx.Summary())
	fmt.Println()

	// Step 2: Initialize Gemini client
	fmt.Println("Connecting to Gemini API...")
	client, err := gemini.NewClient(ctx, apiKey, model)
	if err != nil {
		return fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer client.Close()

	fmt.Println(ui.RenderSuccess(fmt.Sprintf("Connected to %s", client.GetModelID())))
	fmt.Println()

	// Step 3: Run the review
	if interactive {
		// Interactive TUI mode
		return ui.Run(reviewCtx, client)
	}

	// Non-interactive mode
	return ui.RunSimple(ctx, reviewCtx, client)
}

// printSecretsWarning prints a warning about detected secrets
func printSecretsWarning(secrets []filter.SecretMatch) {
	fmt.Println(ui.RenderError("Potential secrets detected in your code!"))
	fmt.Println()

	for _, s := range secrets {
		fmt.Printf("  • %s (line %d): %s\n", s.FilePath, s.Line, s.Match)
	}

	fmt.Println()
	fmt.Println(ui.RenderWarning("Review aborted to prevent sending secrets to external API."))
	fmt.Println(ui.RenderHelp("Use --force to proceed anyway (not recommended)"))
	fmt.Println()

	os.Exit(1)
}
