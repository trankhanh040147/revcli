package preset

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/trankhanh040147/revcli/internal/config"
	"gopkg.in/yaml.v3"
)

// getConfigPath returns the path to the config file
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, config.ConfigDirName, config.AppDirName, "config.yaml"), nil
}

// LoadConfig loads the configuration from ~/.config/revcli/config.yaml
func LoadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return defaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults for missing Gemini config
	applyGeminiDefaults(&config)

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

// applyGeminiDefaults applies default values to Gemini configuration
func applyGeminiDefaults(config *Config) {
	if config.Gemini == nil {
		config.Gemini = defaultGeminiConfig()
	} else {
		if config.Gemini.ModelParams == nil {
			config.Gemini.ModelParams = defaultModelParams()
		} else {
			applyModelParamDefaults(config.Gemini.ModelParams)
		}
		if config.Gemini.SafetySettings == nil {
			config.Gemini.SafetySettings = defaultSafetySettings()
		} else if config.Gemini.SafetySettings.Threshold == "" {
			config.Gemini.SafetySettings.Threshold = defaultSafetySettings().Threshold
		}
	}
}

// defaultConfig returns a Config with default values
func defaultConfig() *Config {
	return &Config{
		Gemini: defaultGeminiConfig(),
	}
}

// defaultGeminiConfig returns a GeminiConfig with default values
func defaultGeminiConfig() *GeminiConfig {
	return &GeminiConfig{
		ModelParams:    defaultModelParams(),
		SafetySettings: defaultSafetySettings(),
	}
}

// defaultModelParams returns ModelParams with default values
func defaultModelParams() *ModelParams {
	return &ModelParams{
		Temperature: 0.3,
		TopP:        0.95,
		TopK:        40,
	}
}

// applyModelParamDefaults applies default values to ModelParams fields that are zero
func applyModelParamDefaults(p *ModelParams) {
	defaultParams := defaultModelParams()
	if p.Temperature == 0 {
		p.Temperature = defaultParams.Temperature
	}
	if p.TopP == 0 {
		p.TopP = defaultParams.TopP
	}
	if p.TopK == 0 {
		p.TopK = defaultParams.TopK
	}
}

// defaultSafetySettings returns SafetySettings with default values
func defaultSafetySettings() *SafetySettings {
	return &SafetySettings{
		Threshold: DefaultSafetyThreshold, // Maps to HarmBlockThresholdBlockOnlyHigh
	}
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

