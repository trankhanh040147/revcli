package preset

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Preset defines a review preset configuration
type Preset struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Prompt      string `yaml:"prompt"`
	Replace     bool   `yaml:"replace,omitempty"` // If true, replace base prompt instead of appending
}

// Config defines the revcli configuration
type Config struct {
	DefaultPreset string `yaml:"default_preset,omitempty"`
}

// BuiltInPresets contains all built-in review presets
var BuiltInPresets = map[string]Preset{
	"quick": {
		Name:        "quick",
		Description: "Fast, high-level review focusing on critical issues only",
		Prompt: `## Review Mode: Quick Review

Focus ONLY on:
- Critical bugs that would cause crashes or data loss
- Security vulnerabilities
- Obvious logic errors

Skip:
- Style suggestions
- Minor improvements
- Documentation issues
- Performance micro-optimizations

Keep your response concise - aim for the top 3-5 most important issues only.`,
	},
	"strict": {
		Name:        "strict",
		Description: "Thorough, nitpicky review covering all aspects",
		Prompt: `## Review Mode: Strict Review

Be extremely thorough and nitpicky. Review EVERYTHING:
- Every possible edge case
- All error handling paths
- Memory and resource management
- Thread safety and race conditions
- API design and consistency
- Documentation completeness
- Test coverage gaps
- Code style and formatting
- Naming conventions
- Dead code and unused imports

Flag even minor issues. Nothing is too small to mention.`,
	},
	"security": {
		Name:        "security",
		Description: "Security-focused analysis for vulnerabilities",
		Prompt: `## Review Mode: Security Audit

Focus exclusively on security concerns:
- Input validation and sanitization
- SQL injection vulnerabilities
- XSS and CSRF risks
- Authentication and authorization flaws
- Hardcoded secrets or credentials
- Insecure cryptographic practices
- Path traversal vulnerabilities
- Command injection risks
- Sensitive data exposure
- Insecure deserialization
- Dependency vulnerabilities
- Access control issues

Rate each finding by severity (Critical/High/Medium/Low).`,
	},
	"performance": {
		Name:        "performance",
		Description: "Performance optimization focused review",
		Prompt: `## Review Mode: Performance Review

Focus on performance implications:
- Unnecessary allocations and memory usage
- Inefficient algorithms (O(n²) when O(n) possible)
- Database query optimization (N+1 queries)
- Caching opportunities
- Goroutine and channel efficiency
- Buffer reuse opportunities
- Unnecessary copying of data
- Lock contention and synchronization
- I/O optimization
- Lazy loading opportunities
- Connection pooling

Suggest specific optimizations with expected impact.`,
	},
	"logic": {
		Name:        "logic",
		Description: "Logic and algorithm correctness verification",
		Prompt: `## Review Mode: Logic Review

Focus on correctness and logic:
- Algorithm correctness
- Edge case handling
- Boundary conditions
- Off-by-one errors
- Null/nil handling
- Boolean logic errors
- State machine correctness
- Control flow issues
- Loop termination conditions
- Recursive function base cases
- Mathematical operations accuracy
- Business logic correctness

Trace through the code logic step by step.`,
	},
	"style": {
		Name:        "style",
		Description: "Code style and formatting review",
		Prompt: `## Review Mode: Style Review

Focus on code style and readability:
- Consistent formatting
- Idiomatic patterns for the language
- Code organization
- Function and file length
- Comment quality and necessity
- Magic numbers and constants
- DRY principle violations
- SOLID principle adherence
- Clean code practices
- Readability improvements
- Consistent error handling patterns

Reference style guides where applicable.`,
	},
	"typo": {
		Name:        "typo",
		Description: "Typo and spelling error detection",
		Prompt: `## Review Mode: Typo Detection

Focus on finding typos and spelling errors:
- Variable name typos
- Function name typos
- Comment spelling errors
- String literal typos
- Documentation errors
- Inconsistent spelling (e.g., color vs colour)
- Common programming typos (recieve, occured, etc.)
- Copy-paste errors
- Wrong word usage

List each typo with its location and correct spelling.`,
	},
	"naming": {
		Name:        "naming",
		Description: "Variable and function naming review",
		Prompt: `## Review Mode: Naming Review

Focus on naming quality:
- Descriptive variable names
- Self-documenting function names
- Consistent naming conventions
- Avoiding abbreviations
- Domain-appropriate terminology
- Boolean naming (is/has/can prefixes)
- Collection naming (plural forms)
- Constant naming (UPPER_CASE where appropriate)
- Package/module naming
- Avoiding generic names (data, info, temp)
- Name length appropriateness

Suggest specific renamed alternatives.`,
	},
}

// Get returns a preset by name, checking built-in first then custom
func Get(name string) (*Preset, error) {
	name = strings.ToLower(name)

	// Check built-in presets first
	if preset, ok := BuiltInPresets[name]; ok {
		return &preset, nil
	}

	// Check custom presets
	preset, err := loadCustomPreset(name)
	if err != nil {
		// Try to find similar presets
		similar := FindSimilarPresets(name, 2)

		var suggestion string
		if len(similar) > 0 {
			if len(similar) == 1 {
				suggestion = fmt.Sprintf(". Did you mean '%s'?", similar[0])
			} else {
				suggestion = fmt.Sprintf(". Did you mean one of: %s?", strings.Join(similar, ", "))
			}
		}

		return nil, fmt.Errorf("preset '%s' not found%s. Available: %s", name, suggestion, ListNames())
	}

	return preset, nil
}

// FindSimilarPresets finds preset names similar to the given name using Levenshtein distance
// threshold: maximum edit distance to consider a preset as similar (lower = more strict)
func FindSimilarPresets(name string, threshold int) []string {
	allNames := GetAllPresetNames()
	similar := make([]string, 0)

	name = strings.ToLower(name)
	nameLen := len(name)

	for _, presetName := range allNames {
		presetNameLower := strings.ToLower(presetName)

		// Skip exact matches
		if presetNameLower == name {
			continue
		}

		// For very short inputs (1-2 chars), use prefix matching
		if nameLen <= 2 {
			if strings.HasPrefix(presetNameLower, name) {
				similar = append(similar, presetName)
				continue
			}
		}

		// Calculate Levenshtein distance
		distance := levenshteinDistance(name, presetNameLower)

		// For short inputs (3-4 chars), use relative similarity (percentage-based)
		// For longer inputs, use absolute threshold
		var isSimilar bool
		if nameLen <= 4 {
			// Use relative similarity: distance should be <= 50% of input length
			maxDistance := (nameLen + 1) / 2 // Allow up to 50% of input length as distance
			isSimilar = distance <= maxDistance
		} else {
			// Use absolute threshold for longer inputs
			isSimilar = distance <= threshold
		}

		// Also check if the input is a substring of the preset name (for partial matches)
		isSubstring := strings.HasPrefix(presetNameLower, name) || strings.Contains(presetNameLower, name)

		if isSimilar || (isSubstring && nameLen >= 3) {
			similar = append(similar, presetName)
		}
	}

	return similar
}

// levenshteinDistance calculates the edit distance between two strings
func levenshteinDistance(s1, s2 string) int {
	r1, r2 := []rune(s1), []rune(s2)
	column := make([]int, len(r1)+1)

	for y := 1; y <= len(r1); y++ {
		column[y] = y
	}

	for x := 1; x <= len(r2); x++ {
		column[0] = x
		lastDiag := x - 1
		for y := 1; y <= len(r1); y++ {
			oldDiag := column[y]
			cost := 0
			if r1[y-1] != r2[x-1] {
				cost = 1
			}
			column[y] = min(column[y]+1, column[y-1]+1, lastDiag+cost)
			lastDiag = oldDiag
		}
	}

	return column[len(r1)]
}

// min returns the minimum of three integers
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// loadCustomPreset attempts to load a preset from ~/.config/revcli/presets/
func loadCustomPreset(name string) (*Preset, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	presetPath := filepath.Join(homeDir, ".config", "revcli", "presets", name+".yaml")

	data, err := os.ReadFile(presetPath)
	if err != nil {
		return nil, err
	}

	var preset Preset
	if err := yaml.Unmarshal(data, &preset); err != nil {
		return nil, fmt.Errorf("failed to parse preset file: %w", err)
	}

	if preset.Name == "" {
		preset.Name = name
	}

	return &preset, nil
}

// ListNames returns a comma-separated list of available preset names (built-in + custom)
func ListNames() string {
	names := GetAllPresetNames()
	return strings.Join(names, ", ")
}

// GetAllPresetNames returns a list of all preset names (built-in + custom)
func GetAllPresetNames() []string {
	names := make([]string, 0, len(BuiltInPresets))

	// Add built-in preset names
	for name := range BuiltInPresets {
		names = append(names, name)
	}

	// Add custom preset names
	customPresets, err := listCustomPresetNames()
	if err == nil {
		names = append(names, customPresets...)
	}

	return names
}

// listCustomPresetNames returns a list of custom preset names
func listCustomPresetNames() ([]string, error) {
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

// List returns all built-in presets
func List() []Preset {
	presets := make([]Preset, 0, len(BuiltInPresets))
	for _, preset := range BuiltInPresets {
		presets = append(presets, preset)
	}
	return presets
}

// ApplyToPrompt appends the preset's prompt modifier to the base system prompt
func (p *Preset) ApplyToPrompt(basePrompt string) string {
	return basePrompt + "\n\n---\n\n" + p.Prompt
}

// getConfigPath returns the path to the config file
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "revcli", "config.yaml"), nil
}

// LoadConfig loads the configuration from ~/.config/revcli/config.yaml
func LoadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves the configuration to ~/.config/revcli/config.yaml
func SaveConfig(config *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetDefaultPreset returns the default preset name from config, or empty string if not set
func GetDefaultPreset() (string, error) {
	config, err := LoadConfig()
	if err != nil {
		return "", err
	}
	return config.DefaultPreset, nil
}

// SetDefaultPreset sets the default preset in the config file
func SetDefaultPreset(presetName string) error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	config.DefaultPreset = presetName
	return SaveConfig(config)
}

// ClearDefaultPreset removes the default preset from the config file
func ClearDefaultPreset() error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	config.DefaultPreset = ""
	return SaveConfig(config)
}

// GetSystemPromptPath returns the path to the system prompt preset file
func GetSystemPromptPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "revcli", "presets", "system.yaml"), nil
}

// LoadSystemPrompt loads the system prompt from custom file or returns empty string to use default
// Returns the prompt text and a boolean indicating if custom prompt was found
func LoadSystemPrompt() (string, bool, error) {
	systemPromptPath, err := GetSystemPromptPath()
	if err != nil {
		return "", false, err
	}

	// Check if custom system prompt file exists
	if _, err := os.Stat(systemPromptPath); os.IsNotExist(err) {
		return "", false, nil
	}

	// Load custom system prompt
	data, err := os.ReadFile(systemPromptPath)
	if err != nil {
		return "", false, fmt.Errorf("failed to read system prompt file: %w", err)
	}

	var preset Preset
	if err := yaml.Unmarshal(data, &preset); err != nil {
		return "", false, fmt.Errorf("failed to parse system prompt file: %w", err)
	}

	if preset.Prompt == "" {
		return "", false, nil
	}

	return preset.Prompt, true, nil
}

// SaveSystemPrompt saves a custom system prompt to file
func SaveSystemPrompt(promptText string) error {
	systemPromptPath, err := GetSystemPromptPath()
	if err != nil {
		return err
	}

	// Ensure preset directory exists
	presetDir := filepath.Dir(systemPromptPath)
	if err := os.MkdirAll(presetDir, 0755); err != nil {
		return fmt.Errorf("failed to create preset directory: %w", err)
	}

	preset := Preset{
		Name:        "system",
		Description: "Custom system prompt for code reviews",
		Prompt:      promptText,
	}

	data, err := yaml.Marshal(&preset)
	if err != nil {
		return fmt.Errorf("failed to marshal system prompt: %w", err)
	}

	if err := os.WriteFile(systemPromptPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write system prompt file: %w", err)
	}

	return nil
}

// DeleteSystemPrompt removes the custom system prompt file to restore default
func DeleteSystemPrompt() error {
	systemPromptPath, err := GetSystemPromptPath()
	if err != nil {
		return err
	}

	if err := os.Remove(systemPromptPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete system prompt file: %w", err)
	}

	return nil
}
