package gemini

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bytedance/sonic"
	"github.com/trankhanh040147/revcli/internal/config"
	"github.com/trankhanh040147/revcli/internal/preset"
	"google.golang.org/genai"
)

// StreamCallback remains the same
type StreamCallback func(string)

// TokenUsage matches the new SDK field names
type TokenUsage struct {
	PromptTokens     int32
	CompletionTokens int32
	TotalTokens      int32
}

type Client struct {
	client     *genai.Client
	modelID    string
	lastUsage  *TokenUsage
	modelConfig *preset.ModelParams
	safetyThreshold string
	// In the new SDK, history is maintained as a slice of Content
	history []*genai.Content
	// System instructions are passed per request in the config
	systemPrompt string
}

func NewClient(ctx context.Context, apiKey, modelID string, config *preset.Config) (*Client, error) {
	// The new client automatically picks up API key from env if cfg is nil,
	// but we'll set it explicitly for your use case.
	cfg := &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	}
	client, err := genai.NewClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Load config or use defaults
	var modelConfig *preset.ModelParams
	var safetyThreshold string
	if config != nil && config.Gemini != nil {
		modelConfig = config.Gemini.ModelParams
		if config.Gemini.SafetySettings != nil {
			safetyThreshold = config.Gemini.SafetySettings.Threshold
		}
	}

	// Apply defaults if not provided
	if modelConfig == nil {
		modelConfig = &preset.ModelParams{
			Temperature: DefaultModelTemperature,
			TopP:         DefaultModelTopP,
			TopK:         DefaultModelTopK,
		}
	}
	if safetyThreshold == "" {
		safetyThreshold = preset.DefaultSafetyThreshold
	}

	return &Client{
		client:          client,
		modelID:         modelID,
		modelConfig:     modelConfig,
		safetyThreshold: safetyThreshold,
	}, nil
}

// StartChat sets the system prompt and clears history
func (c *Client) StartChat(systemPrompt string) {
	c.systemPrompt = systemPrompt
	c.history = make([]*genai.Content, 0, DefaultHistoryCapacity)
}

// rollbackLastHistoryEntry removes the last entry from history
func (c *Client) rollbackLastHistoryEntry() {
	if len(c.history) > 0 {
		c.history = c.history[:len(c.history)-1]
	}
}

func (c *Client) GetLastUsage() *TokenUsage {
	return c.lastUsage
}

func (c *Client) GetModelID() string {
	return c.modelID
}

// SessionData represents the serializable session state
type SessionData struct {
	SystemPrompt string            `json:"systemPrompt"`
	History      []*genai.Content  `json:"history"`
	ModelID      string            `json:"modelID"`
}

// getSessionPath returns the path to a session file
func getSessionPath(sessionName string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	sessionDir := filepath.Join(homeDir, config.ConfigDirName, config.AppDirName, config.SessionsDirName)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create sessions directory: %w", err)
	}
	return filepath.Join(sessionDir, sessionName+".json"), nil
}

// SaveSession saves the current session (history and system prompt) to disk
func (c *Client) SaveSession(sessionName string) error {
	if sessionName == "" {
		return fmt.Errorf("session name cannot be empty")
	}

	sessionPath, err := getSessionPath(sessionName)
	if err != nil {
		return err
	}

	sessionData := SessionData{
		SystemPrompt: c.systemPrompt,
		History:      c.history,
		ModelID:      c.modelID,
	}

	data, err := sonic.MarshalIndent(sessionData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	if err := os.WriteFile(sessionPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// LoadSession loads a session (history and system prompt) from disk
func (c *Client) LoadSession(sessionName string) error {
	if sessionName == "" {
		return fmt.Errorf("session name cannot be empty")
	}

	sessionPath, err := getSessionPath(sessionName)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(sessionPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("session '%s' not found", sessionName)
		}
		return fmt.Errorf("failed to read session file: %w", err)
	}

	var sessionData SessionData
	if err := sonic.Unmarshal(data, &sessionData); err != nil {
		return fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	// Restore session state
	c.systemPrompt = sessionData.SystemPrompt
	c.history = sessionData.History
	// Note: modelID is not changed as it's tied to the client instance

	return nil
}
