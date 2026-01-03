package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/trankhanh040147/revcli/internal/preset"
	"github.com/trankhanh040147/revcli/internal/prompt"
	"github.com/trankhanh040147/revcli/internal/ui"
)

var (
	presetNameFlag    string
	presetDescription string
	presetPrompt      string
	presetUnsetFlag   bool
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

// presetEditCmd edits an existing custom preset
var presetEditCmd = &cobra.Command{
	Use:   "edit [name]",
	Short: "Edit a custom preset",
	Long:  `Edit an existing custom preset. Built-in presets cannot be edited.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runPresetEdit,
}

// presetOpenCmd opens a preset file in editor or preset directory in file manager
var presetOpenCmd = &cobra.Command{
	Use:   "open [name]",
	Short: "Open preset file or directory",
	Long: `Open a preset file in your default editor, or open the preset directory in your file manager.
	
If a preset name is provided, opens that preset file in $EDITOR (or vi as fallback).
If no name is provided, opens the preset directory in your file manager.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPresetOpen,
}

// presetPathCmd shows the path to a preset file or directory
var presetPathCmd = &cobra.Command{
	Use:   "path [name]",
	Short: "Show preset file or directory path",
	Long: `Show the file path for a specific preset, or the preset directory path if no name is provided.
	
Useful for manual editing or scripting.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPresetPath,
}

// presetDefaultCmd manages the default preset
var presetDefaultCmd = &cobra.Command{
	Use:   "default [name]",
	Short: "Set or show the default preset",
	Long: `Set the default preset to use when --preset flag is not provided, or show the current default preset.
	
Examples:
  revcli preset default quick          # Set 'quick' as default
  revcli preset default                # Show current default
  revcli preset default --unset        # Clear default preset`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPresetDefault,
}

// presetSystemCmd manages the system prompt
var presetSystemCmd = &cobra.Command{
	Use:   "system",
	Short: "Manage system prompt",
	Long: `Manage the custom system prompt used for code reviews.
	
The system prompt defines the AI's role and review style. By default, a built-in
system prompt is used. You can customize it by creating a custom system prompt.`,
}

// presetSystemShowCmd shows the current system prompt
var presetSystemShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current system prompt",
	Long:  `Display the current system prompt (custom or default).`,
	RunE:  runPresetSystemShow,
}

// presetSystemEditCmd edits the system prompt
var presetSystemEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit system prompt",
	Long: `Edit the system prompt interactively or open in editor.
	
If a custom system prompt exists, it will be loaded for editing.
Otherwise, the default system prompt will be used as a starting point.`,
	RunE: runPresetSystemEdit,
}

// presetSystemResetCmd resets the system prompt to default
var presetSystemResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset system prompt to default",
	Long:  `Remove the custom system prompt file to restore the default system prompt.`,
	RunE:  runPresetSystemReset,
}

func init() {
	rootCmd.AddCommand(presetCmd)
	presetCmd.AddCommand(presetListCmd)
	presetCmd.AddCommand(presetCreateCmd)
	presetCmd.AddCommand(presetDeleteCmd)
	presetCmd.AddCommand(presetShowCmd)
	presetCmd.AddCommand(presetEditCmd)
	presetCmd.AddCommand(presetOpenCmd)
	presetCmd.AddCommand(presetPathCmd)
	presetCmd.AddCommand(presetDefaultCmd)
	presetCmd.AddCommand(presetSystemCmd)

	presetSystemCmd.AddCommand(presetSystemShowCmd)
	presetSystemCmd.AddCommand(presetSystemEditCmd)
	presetSystemCmd.AddCommand(presetSystemResetCmd)

	presetCreateCmd.Flags().StringVarP(&presetNameFlag, "name", "n", "", "Preset name")
	presetCreateCmd.Flags().StringVarP(&presetDescription, "description", "d", "", "Preset description")
	presetCreateCmd.Flags().StringVarP(&presetPrompt, "prompt", "p", "", "Preset prompt text")
	presetDefaultCmd.Flags().BoolVar(&presetUnsetFlag, "unset", false, "Clear the default preset")
}

// editMultilineText opens the current text in an external editor and returns the edited content
func editMultilineText(prompt string, currentValue string) (string, error) {
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "revcli-edit-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath) // Clean up temp file

	// Write current value to temp file
	if _, err := tmpFile.WriteString(currentValue); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to write to temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	// Get editor from environment or use fallback
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi" // Default fallback
	}

	// Show prompt message
	if prompt != "" {
		fmt.Println(prompt)
	}
	fmt.Printf("Opening editor (%s)... Press Ctrl+X then Y to save and exit (vim), or Ctrl+O then Enter (nano).\n", editor)

	// Open file in editor
	editCmd := exec.Command(editor, tmpPath)
	editCmd.Stdin = os.Stdin
	editCmd.Stdout = os.Stdout
	editCmd.Stderr = os.Stderr

	if err := editCmd.Run(); err != nil {
		return "", fmt.Errorf("editor exited with error: %w", err)
	}

	// Read back edited content
	editedData, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", fmt.Errorf("failed to read edited file: %w", err)
	}

	// Remove trailing line endings (handle both \r\n and \n)
	editedText := string(editedData)
	editedText = strings.TrimSuffix(editedText, "\r\n") // Windows line ending
	editedText = strings.TrimSuffix(editedText, "\n")   // Unix line ending

	return editedText, nil
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

	// Create a single reader for stdin to avoid data loss when input is piped
	// Multiple readers from the same stdin can cause buffered data to be lost
	stdinReader := bufio.NewReader(os.Stdin)

	// Get description
	description := presetDescription
	if description == "" {
		fmt.Print("Description: ")
		desc, err := stdinReader.ReadString('\n')
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read description: %w", err)
		}
		description = strings.TrimSpace(desc)
	}

	// Get prompt
	promptText := presetPrompt
	if promptText == "" {
		var err error
		promptText, err = editMultilineText("Enter the preset prompt:", "")
		if err != nil {
			return fmt.Errorf("failed to edit prompt: %w", err)
		}
		if strings.TrimSpace(promptText) == "" {
			return fmt.Errorf("prompt input cancelled or empty")
		}
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

func runPresetEdit(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Check if it's a built-in preset
	if _, ok := preset.BuiltInPresets[strings.ToLower(name)]; ok {
		return fmt.Errorf("cannot edit built-in preset '%s'. Use 'revcli preset create' to create a custom preset", name)
	}

	// Load existing preset
	p, err := preset.Get(name)
	if err != nil {
		return fmt.Errorf("preset '%s' not found: %w", name, err)
	}

	// Verify it's a custom preset by checking if file exists
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	normalizedName := strings.ToLower(name)
	presetPath := filepath.Join(homeDir, ".config", "revcli", "presets", normalizedName+".yaml")
	if _, err := os.Stat(presetPath); os.IsNotExist(err) {
		return fmt.Errorf("preset '%s' is not a custom preset and cannot be edited", name)
	}

	fmt.Println(ui.RenderTitle(fmt.Sprintf("✏️  Editing Preset: %s", p.Name)))
	fmt.Println()
	fmt.Printf("Name: %s (read-only)\n", p.Name)
	fmt.Println()
	fmt.Println("Current values are shown in brackets. Press Enter to keep, or type a new value.")
	fmt.Println()

	// Create a single reader for stdin to avoid data loss when input is piped
	stdinReader := bufio.NewReader(os.Stdin)

	// Edit description
	fmt.Printf("Description [%s] (press Enter to keep): ", p.Description)
	descInput, err := stdinReader.ReadString('\n')
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read description: %w", err)
	}
	newDescription := strings.TrimSpace(descInput)
	if newDescription == "" {
		newDescription = p.Description
	}

	// Edit prompt using external editor
	fmt.Println("Current prompt:")
	fmt.Println(p.Prompt)
	fmt.Println()
	newPrompt, err := editMultilineText("Enter new prompt (or leave empty to keep current):", p.Prompt)
	if err != nil {
		return fmt.Errorf("failed to edit prompt: %w", err)
	}
	if strings.TrimSpace(newPrompt) == "" {
		newPrompt = p.Prompt
	}

	if newPrompt == "" {
		return fmt.Errorf("prompt text is required")
	}

	// Update preset (name remains unchanged)
	p.Description = newDescription
	p.Prompt = newPrompt

	// Save to file (using original path since name doesn't change)
	presetDir := filepath.Join(homeDir, ".config", "revcli", "presets")
	if err := os.MkdirAll(presetDir, 0755); err != nil {
		return fmt.Errorf("failed to create preset directory: %w", err)
	}

	data, err := yaml.Marshal(&p)
	if err != nil {
		return fmt.Errorf("failed to marshal preset: %w", err)
	}

	if err := os.WriteFile(presetPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write preset file: %w", err)
	}

	fmt.Println()
	fmt.Println(ui.RenderSuccess(fmt.Sprintf("Preset '%s' updated successfully!", normalizedName)))
	fmt.Printf("Location: %s\n", presetPath)

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

func runPresetOpen(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	presetDir := filepath.Join(homeDir, ".config", "revcli", "presets")

	if len(args) > 0 {
		// Open specific preset file in editor
		name := args[0]

		// Check if it's a built-in preset
		if _, ok := preset.BuiltInPresets[strings.ToLower(name)]; ok {
			return fmt.Errorf("cannot open built-in preset '%s'. Built-in presets are not stored as files", name)
		}

		// Check if custom preset exists
		_, err := preset.Get(name)
		if err != nil {
			return fmt.Errorf("preset '%s' not found: %w", name, err)
		}

		normalizedName := strings.ToLower(name)
		presetPath := filepath.Join(presetDir, normalizedName+".yaml")

		// Check if file exists
		if _, err := os.Stat(presetPath); os.IsNotExist(err) {
			return fmt.Errorf("preset file '%s' does not exist", presetPath)
		}

		// Get editor from environment or use fallback
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi" // Default fallback
		}

		// Open file in editor
		editCmd := exec.Command(editor, presetPath)
		editCmd.Stdin = os.Stdin
		editCmd.Stdout = os.Stdout
		editCmd.Stderr = os.Stderr

		if err := editCmd.Run(); err != nil {
			return fmt.Errorf("failed to open editor: %w", err)
		}

		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Opened preset '%s' in %s", name, editor)))
	} else {
		// Open preset directory in file manager
		// Ensure directory exists
		if err := os.MkdirAll(presetDir, 0755); err != nil {
			return fmt.Errorf("failed to create preset directory: %w", err)
		}

		var openCmd *exec.Cmd
		switch runtime.GOOS {
		case "linux":
			openCmd = exec.Command("xdg-open", presetDir)
		case "darwin":
			openCmd = exec.Command("open", presetDir)
		case "windows":
			openCmd = exec.Command("explorer", presetDir)
		default:
			return fmt.Errorf("unsupported operating system: %s. Please open the directory manually: %s", runtime.GOOS, presetDir)
		}

		if err := openCmd.Run(); err != nil {
			return fmt.Errorf("failed to open file manager: %w. Directory: %s", err, presetDir)
		}

		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Opened preset directory: %s", presetDir)))
	}

	return nil
}

func runPresetPath(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	presetDir := filepath.Join(homeDir, ".config", "revcli", "presets")

	if len(args) > 0 {
		// Show path to specific preset file
		name := args[0]

		// Check if it's a built-in preset
		if _, ok := preset.BuiltInPresets[strings.ToLower(name)]; ok {
			return fmt.Errorf("built-in preset '%s' does not have a file path. Built-in presets are embedded in the application", name)
		}

		// Check if custom preset exists
		_, err := preset.Get(name)
		if err != nil {
			return fmt.Errorf("preset '%s' not found: %w", name, err)
		}

		normalizedName := strings.ToLower(name)
		presetPath := filepath.Join(presetDir, normalizedName+".yaml")

		fmt.Println(presetPath)
	} else {
		// Show preset directory path
		fmt.Println(presetDir)
	}

	return nil
}

func runPresetDefault(cmd *cobra.Command, args []string) error {
	// Handle --unset flag
	if presetUnsetFlag {
		if err := preset.ClearDefaultPreset(); err != nil {
			return fmt.Errorf("failed to clear default preset: %w", err)
		}
		fmt.Println(ui.RenderSuccess("Default preset cleared."))
		return nil
	}

	// If name provided, set default
	if len(args) > 0 {
		name := args[0]

		// Validate preset exists
		_, err := preset.Get(name)
		if err != nil {
			return fmt.Errorf("preset '%s' not found: %w", name, err)
		}

		if err := preset.SetDefaultPreset(name); err != nil {
			return fmt.Errorf("failed to set default preset: %w", err)
		}

		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Default preset set to '%s'", name)))
		return nil
	}

	// No args: show current default
	defaultPreset, err := preset.GetDefaultPreset()
	if err != nil {
		return fmt.Errorf("failed to load default preset: %w", err)
	}

	if defaultPreset == "" {
		fmt.Println("No default preset is set.")
		fmt.Println("Use 'revcli preset default <name>' to set one.")
	} else {
		fmt.Printf("Default preset: %s\n", ui.RenderSuccess(defaultPreset))
	}

	return nil
}

func runPresetSystemShow(cmd *cobra.Command, args []string) error {
	// Load system prompt (custom or default)
	customPrompt, found, err := preset.LoadSystemPrompt()
	if err != nil {
		return fmt.Errorf("failed to load system prompt: %w", err)
	}

	if found {
		fmt.Println(ui.RenderTitle("📝 Custom System Prompt"))
		fmt.Println()
		fmt.Println(ui.RenderSubtitle("Current system prompt (custom):"))
		fmt.Println(customPrompt)
		fmt.Println()
		systemPromptPath, _ := preset.GetSystemPromptPath()
		fmt.Printf("Location: %s\n", systemPromptPath)
	} else {
		fmt.Println(ui.RenderTitle("📝 System Prompt"))
		fmt.Println()
		fmt.Println(ui.RenderSubtitle("Current system prompt (default):"))
		fmt.Println(prompt.SystemPrompt)
		fmt.Println()
		fmt.Println("No custom system prompt found. Use 'revcli preset system edit' to create one.")
	}

	return nil
}

func runPresetSystemEdit(cmd *cobra.Command, args []string) error {
	// Load current system prompt (custom or default)
	customPrompt, found, err := preset.LoadSystemPrompt()
	if err != nil {
		return fmt.Errorf("failed to load system prompt: %w", err)
	}

	// Get current prompt (custom or default)
	var currentPrompt string
	if found {
		currentPrompt = customPrompt
	} else {
		currentPrompt = prompt.SystemPrompt
	}

	fmt.Println(ui.RenderTitle("✏️  Editing System Prompt"))
	fmt.Println()
	if currentPrompt != "" {
		fmt.Println("Current system prompt:")
		fmt.Println(currentPrompt)
		fmt.Println()
	}
	newPrompt, err := editMultilineText("Enter new system prompt:", currentPrompt)
	if err != nil {
		return fmt.Errorf("failed to edit system prompt: %w", err)
	}
	if strings.TrimSpace(newPrompt) == "" {
		return fmt.Errorf("system prompt input cancelled or empty")
	}

	if newPrompt == "" {
		return fmt.Errorf("system prompt text is required")
	}

	// Save custom system prompt
	if err := preset.SaveSystemPrompt(newPrompt); err != nil {
		return fmt.Errorf("failed to save system prompt: %w", err)
	}

	fmt.Println()
	systemPromptPath, _ := preset.GetSystemPromptPath()
	fmt.Println(ui.RenderSuccess("System prompt updated successfully!"))
	fmt.Printf("Location: %s\n", systemPromptPath)
	fmt.Println()
	fmt.Println("The custom system prompt will be used in all future reviews.")
	fmt.Println("Use 'revcli preset system reset' to restore the default.")

	return nil
}

func runPresetSystemReset(cmd *cobra.Command, args []string) error {
	// Check if custom system prompt exists
	_, found, err := preset.LoadSystemPrompt()
	if err != nil {
		return fmt.Errorf("failed to check system prompt: %w", err)
	}

	if !found {
		fmt.Println("No custom system prompt found. Already using default system prompt.")
		return nil
	}

	// Confirm deletion
	fmt.Printf("Are you sure you want to reset the system prompt to default? (y/N): ")
	var confirm string
	fmt.Scanln(&confirm)
	if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
		fmt.Println("Reset cancelled.")
		return nil
	}

	// Delete custom system prompt file
	if err := preset.DeleteSystemPrompt(); err != nil {
		return fmt.Errorf("failed to reset system prompt: %w", err)
	}

	fmt.Println()
	fmt.Println(ui.RenderSuccess("System prompt reset to default successfully!"))
	fmt.Println("The default system prompt will be used in all future reviews.")

	return nil
}
