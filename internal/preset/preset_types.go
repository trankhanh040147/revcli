package preset

import "gopkg.in/yaml.v3"

// Preset defines a review preset configuration
type Preset struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Prompt      string `yaml:"prompt"`
	Replace     bool   `yaml:"replace,omitempty"` // If true, replace base prompt instead of appending
}

// MarshalYAML implements custom YAML marshaling to use literal block scalars for multiline prompts
func (p *Preset) MarshalYAML() (interface{}, error) {
	// Build the mapping node manually to control formatting
	root := &yaml.Node{
		Kind:    yaml.MappingNode,
		Content: []*yaml.Node{},
	}

	// Add name field
	root.Content = append(root.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "name"},
		&yaml.Node{Kind: yaml.ScalarNode, Value: p.Name},
	)

	// Add description field
	root.Content = append(root.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "description"},
		&yaml.Node{Kind: yaml.ScalarNode, Value: p.Description},
	)

	// Add prompt field with literal style for multiline content
	promptNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: p.Prompt,
		Style: yaml.LiteralStyle, // Use | literal block scalar
	}
	root.Content = append(root.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "prompt"},
		promptNode,
	)

	// Add replace field only if true
	if p.Replace {
		root.Content = append(root.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "replace"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: "true"},
		)
	}

	// Return root node directly - yaml.Marshal() will wrap it in a document automatically
	return root, nil
}

// Config defines the revcli configuration
type Config struct {
	DefaultPreset string        `yaml:"default_preset,omitempty"`
	Gemini        *GeminiConfig `yaml:"gemini,omitempty"`
}

// GeminiConfig defines Gemini API client configuration
type GeminiConfig struct {
	ModelParams     *ModelParams     `yaml:"model_params,omitempty"`
	SafetySettings  *SafetySettings  `yaml:"safety_settings,omitempty"`
}

// ModelParams defines model generation parameters
type ModelParams struct {
	Temperature float32 `yaml:"temperature,omitempty"`
	TopP         float32 `yaml:"top_p,omitempty"`
	TopK         int     `yaml:"top_k,omitempty"`
}

// SafetySettings defines safety filtering configuration
type SafetySettings struct {
	Threshold string `yaml:"threshold,omitempty"` // "HIGH", "MEDIUM_AND_ABOVE", "LOW_AND_ABOVE", "NONE", "OFF"
}

// DefaultSafetyThreshold is the default safety threshold value
const DefaultSafetyThreshold = "HIGH"

