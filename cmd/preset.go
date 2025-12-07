package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/trankhanh040147/revcli/internal/preset"
	"github.com/trankhanh040147/revcli/internal/ui"
)

var (
	presetNameFlag    string
	presetDescription string
	presetPrompt      string
)

// presetCmd represents the preset command
var presetCmd = &cobra.Command{
	Use:   "preset",
	Short: "Manage review presets",
	Long: `Manage custom review presets for code reviews.

Presets allow you to customize the review style and focus. Built-in presets
cannot be modified, but you can create custom presets in ~/.config/revcli/presets/`,
}

// presetListCmd lists all available presets
var presetListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available presets",
	Long:  `List all built-in and custom presets with their descriptions.`,
	RunE:  runPresetList,
}

// presetCreateCmd creates a new custom preset
var presetCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new custom preset",
	Long: `Create a new custom preset interactively.

You will be prompted for:
- Name: Unique identifier for the preset
- Description: Brief description of what the preset focuses on
- Prompt: The review instructions/prompt text`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPresetCreate,
}

// presetDeleteCmd deletes a custom preset
var presetDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a custom preset",
	Long:  `Delete a custom preset. Built-in presets cannot be deleted.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runPresetDelete,
}

// presetShowCmd shows details of a preset
var presetShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show preset details",
	Long:  `Display the full details of a preset including its prompt.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runPresetShow,
}

func init() {
	rootCmd.AddCommand(presetCmd)
	presetCmd.AddCommand(presetListCmd)
	presetCmd.AddCommand(presetCreateCmd)
	presetCmd.AddCommand(presetDeleteCmd)
	presetCmd.AddCommand(presetShowCmd)

	presetCreateCmd.Flags().StringVarP(&presetNameFlag, "name", "n", "", "Preset name")
	presetCreateCmd.Flags().StringVarP(&presetDescription, "description", "d", "", "Preset description")
	presetCreateCmd.Flags().StringVarP(&presetPrompt, "prompt", "p", "", "Preset prompt text")
}

func runPresetList(cmd *cobra.Command, args []string) error {
	fmt.Println(ui.RenderTitle("📋 Available Presets"))
	fmt.Println()

	// List built-in presets
	builtInPresets := preset.List()
	if len(builtInPresets) > 0 {
		fmt.Println(ui.RenderSubtitle("Built-in Presets:"))
		for _, p := range builtInPresets {
			fmt.Printf("  • %s\n", ui.RenderSuccess(p.Name))
			if p.Description != "" {
				fmt.Printf("    %s\n", p.Description)
			}
			fmt.Println()
		}
	}

	// List custom presets
	customPresets, err := listCustomPresets()
	if err != nil {
		fmt.Printf("Warning: Could not list custom presets: %v\n", err)
	} else if len(customPresets) > 0 {
		fmt.Println(ui.RenderSubtitle("Custom Presets:"))
		for _, name := range customPresets {
			p, err := preset.Get(name)
			if err == nil {
				fmt.Printf("  • %s\n", ui.RenderSuccess(name))
				if p.Description != "" {
					fmt.Printf("    %s\n", p.Description)
				}
				fmt.Println()
			}
		}
	} else {
		fmt.Println(ui.RenderSubtitle("No custom presets found."))
		fmt.Println("Use 'revcli preset create' to create one.")
	}

	return nil
}

func runPresetCreate(cmd *cobra.Command, args []string) error {
	var name string
	if len(args) > 0 {
		name = args[0]
	} else if presetNameFlag != "" {
		name = presetNameFlag
	} else {
		fmt.Print("Preset name: ")
		fmt.Scanln(&name)
	}

	if name == "" {
		return fmt.Errorf("preset name is required")
	}

	// Check if preset already exists (built-in or custom)
	_, err := preset.Get(name)
	if err == nil {
		return fmt.Errorf("preset '%s' already exists", name)
	}

	// Get description
	description := presetDescription
	if description == "" {
		fmt.Print("Description: ")
		reader := bufio.NewReader(os.Stdin)
		desc, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read description: %w", err)
		}
		description = strings.TrimSpace(desc)
	}

	// Get prompt
	promptText := presetPrompt
	if promptText == "" {
		fmt.Println("Enter the preset prompt (press Enter twice or Ctrl+D to finish):")
		reader := bufio.NewReader(os.Stdin)
		var lines []string
		emptyCount := 0
		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				// EOF: add any partial line content, then break
				line = strings.TrimSuffix(line, "\n")
				if line != "" {
					lines = append(lines, line)
				}
				// If we have no content at all, user cancelled
				if len(lines) == 0 {
					return fmt.Errorf("prompt input cancelled")
				}
				break
			}
			if err != nil {
				return fmt.Errorf("failed to read prompt: %w", err)
			}
			// Trim the newline character
			line = strings.TrimSuffix(line, "\n")
			if line == "" {
				emptyCount++
				if emptyCount >= 2 {
					break
				}
			} else {
				emptyCount = 0
				lines = append(lines, line)
			}
		}
		promptText = strings.Join(lines, "\n")
	}

	if promptText == "" {
		return fmt.Errorf("prompt text is required")
	}

	// Create preset
	p := preset.Preset{
		Name:        name,
		Description: description,
		Prompt:      promptText,
	}

	// Save to file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	presetDir := filepath.Join(homeDir, ".config", "revcli", "presets")
	if err := os.MkdirAll(presetDir, 0755); err != nil {
		return fmt.Errorf("failed to create preset directory: %w", err)
	}

	// Normalize preset name to lowercase for consistency with Get() function
	normalizedName := strings.ToLower(name)
	presetPath := filepath.Join(presetDir, normalizedName+".yaml")
	data, err := yaml.Marshal(&p)
	if err != nil {
		return fmt.Errorf("failed to marshal preset: %w", err)
	}

	if err := os.WriteFile(presetPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write preset file: %w", err)
	}

	fmt.Println()
	fmt.Println(ui.RenderSuccess(fmt.Sprintf("Preset '%s' created successfully!", normalizedName)))
	fmt.Printf("Location: %s\n", presetPath)

	return nil
}

func runPresetDelete(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Check if it's a built-in preset
	if _, ok := preset.BuiltInPresets[strings.ToLower(name)]; ok {
		return fmt.Errorf("cannot delete built-in preset '%s'", name)
	}

	// Check if custom preset exists
	_, err := preset.Get(name)
	if err != nil {
		return fmt.Errorf("preset '%s' not found: %w", name, err)
	}

	// Get preset file path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Normalize preset name to lowercase for consistency with Get() function
	normalizedName := strings.ToLower(name)
	presetPath := filepath.Join(homeDir, ".config", "revcli", "presets", normalizedName+".yaml")

	// Confirm deletion
	fmt.Printf("Are you sure you want to delete preset '%s'? (y/N): ", name)
	var confirm string
	fmt.Scanln(&confirm)
	if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
		fmt.Println("Deletion cancelled.")
		return nil
	}

	// Delete file
	if err := os.Remove(presetPath); err != nil {
		return fmt.Errorf("failed to delete preset file: %w", err)
	}

	fmt.Println()
	fmt.Println(ui.RenderSuccess(fmt.Sprintf("Preset '%s' deleted successfully!", name)))

	return nil
}

func runPresetShow(cmd *cobra.Command, args []string) error {
	name := args[0]

	p, err := preset.Get(name)
	if err != nil {
		return err
	}

	fmt.Println(ui.RenderTitle(fmt.Sprintf("📝 Preset: %s", p.Name)))
	fmt.Println()

	if p.Description != "" {
		fmt.Println(ui.RenderSubtitle("Description:"))
		fmt.Println(p.Description)
		fmt.Println()
	}

	fmt.Println(ui.RenderSubtitle("Prompt:"))
	fmt.Println(p.Prompt)

	return nil
}

// listCustomPresets returns a list of custom preset names
func listCustomPresets() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	presetDir := filepath.Join(homeDir, ".config", "revcli", "presets")

	// Check if directory exists
	if _, err := os.Stat(presetDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	files, err := os.ReadDir(presetDir)
	if err != nil {
		return nil, err
	}

	var presets []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
			name := strings.TrimSuffix(file.Name(), ".yaml")
			presets = append(presets, name)
		}
	}

	return presets, nil
}
