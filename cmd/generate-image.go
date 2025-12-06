package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/trankhanh040147/revcli/internal/gemini"
	"github.com/trankhanh040147/revcli/internal/image"
)

var (
	imagePrompt      string
	imageOutput      string
	imageAspectRatio string
	imageResolution  string
	imageInteractive bool
)

// generateImageCmd represents the generate-image command
var generateImageCmd = &cobra.Command{
	Use:   "generate-image",
	Short: "Generate images using Gemini 3 Pro Image Preview",
	Long: `Generate images from text prompts using Google's Gemini 3 Pro Image Preview model.

This command uses the gemini-3-pro-image-preview model to create images based on
text descriptions. You can specify aspect ratio, resolution, and output location.

Examples:
  # Generate an image with default settings
  revcli generate-image --prompt "A beautiful sunset over mountains"

  # Generate with custom aspect ratio and resolution
  revcli generate-image -p "A cat wearing sunglasses" -a 16:9 -r 2K -o cat.png

  # Generate in interactive mode
  revcli generate-image --interactive`,
	RunE:    runGenerateImage,
	Aliases: []string{"image", "img"},
}

func init() {
	rootCmd.AddCommand(generateImageCmd)

	generateImageCmd.Flags().StringVarP(&imagePrompt, "prompt", "p", "", "Text prompt for image generation (required)")
	generateImageCmd.Flags().StringVarP(&imageOutput, "output", "o", image.DefaultOutputFile, "Output file path")
	generateImageCmd.Flags().StringVarP(&imageAspectRatio, "aspect-ratio", "a", image.DefaultAspectRatio, "Aspect ratio (1:1, 2:3, 3:2, 4:3, 3:4, 16:9, 9:16)")
	generateImageCmd.Flags().StringVarP(&imageResolution, "resolution", "r", image.DefaultResolution, "Image resolution (1K, 2K, 4K)")
	generateImageCmd.Flags().BoolVarP(&imageInteractive, "interactive", "i", false, "Enable interactive TUI mode")
	generateImageCmd.Flags().BoolP("no-interactive", "I", false, "Disable interactive TUI mode")
}

func runGenerateImage(cmd *cobra.Command, args []string) error {
	// Check for API key
	apiKey := GetAPIKey()
	if apiKey == "" {
		return fmt.Errorf("GEMINI_API_KEY is required. Set it via environment variable or --api-key flag")
	}

	// Handle --no-interactive flag
	if cmd.Flags().Changed("no-interactive") {
		imageInteractive = false
	}

	// Create context
	ctx := context.Background()

	// Check if prompt is provided (unless in interactive mode)
	if imagePrompt == "" && !imageInteractive {
		return fmt.Errorf("prompt is required. Use --prompt flag or --interactive mode")
	}

	// Initialize Gemini client for image generation
	client, err := gemini.NewClient(ctx, apiKey, image.DefaultModel)
	if err != nil {
		return fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer client.Close()

	// Create image generator
	generator := image.NewGenerator(client, apiKey)

	// Run in interactive or non-interactive mode
	if imageInteractive {
		// TODO: Implement interactive TUI mode
		// For now, fall back to non-interactive if prompt is provided
		if imagePrompt == "" {
			return fmt.Errorf("interactive mode not yet implemented. Please provide --prompt flag")
		}
	}

	// Estimate cost before generation
	estimatedCost := image.EstimateCost(imagePrompt, imageResolution)
	if estimatedCost != nil {
		fmt.Println("Estimated cost:")
		fmt.Printf("  $%.6f (Input: $%.6f, Output: $%.6f)\n",
			estimatedCost.TotalCost,
			estimatedCost.InputCost,
			estimatedCost.OutputCost,
		)
		fmt.Printf("  ~%.1f VND (Input: ~%.1f VND, Output: ~%.1f VND)\n",
			estimatedCost.TotalCostVND,
			estimatedCost.InputCostVND,
			estimatedCost.OutputCostVND,
		)
		fmt.Println()
	}

	// Generate image
	fmt.Println("Generating image...")
	fmt.Printf("Prompt: %s\n", imagePrompt)
	fmt.Printf("Aspect Ratio: %s\n", imageAspectRatio)
	fmt.Printf("Resolution: %s\n", imageResolution)
	fmt.Println()

	result, err := generator.GenerateImage(ctx, imagePrompt, imageAspectRatio, imageResolution, imageOutput)
	if err != nil {
		return fmt.Errorf("failed to generate image: %w", err)
	}

	// Print success message
	fmt.Printf("✓ Image generated successfully!\n")
	fmt.Printf("  Saved to: %s\n", imageOutput)
	if result.Description != "" {
		fmt.Printf("\nDescription: %s\n", result.Description)
	}

	// Display token usage
	if result.Usage != nil {
		fmt.Println()
		fmt.Printf("📊 Token Usage: %d prompt + %d completion = %d total\n",
			result.Usage.PromptTokens,
			result.Usage.CompletionTokens,
			result.Usage.TotalTokens,
		)
	}

	// Display cost
	if result.Cost != nil {
		fmt.Printf("💰 Cost: $%.6f (Input: $%.6f, Output: $%.6f)\n",
			result.Cost.TotalCost,
			result.Cost.InputCost,
			result.Cost.OutputCost,
		)
		fmt.Printf("   ~%.1f VND (Input: ~%.1f VND, Output: ~%.1f VND)\n",
			result.Cost.TotalCostVND,
			result.Cost.InputCostVND,
			result.Cost.OutputCostVND,
		)
	}

	return nil
}
