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
		return nil, fmt.Errorf("preset '%s' not found. Available: %s", name, ListNames())
	}

	return preset, nil
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

// ListNames returns a comma-separated list of available preset names
func ListNames() string {
	names := make([]string, 0, len(BuiltInPresets))
	for name := range BuiltInPresets {
		names = append(names, name)
	}
	return strings.Join(names, ", ")
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
