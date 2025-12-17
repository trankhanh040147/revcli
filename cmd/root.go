package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/trankhanh040147/revcli/internal/config"
)

// Version information (set at build time)
var (
	version = "0.3.0"
	commit  = "dev"
	date    = "unknown"
)

var (
	// Config holds global configuration
	apiKey string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "revcli",
	Short:   "Gemini-powered code reviewer CLI",
	Version: version,
	Long: `revcli is a local command-line tool that acts as an intelligent peer reviewer.
It reads your local git changes and uses Google's Gemini LLM to analyze your code
for bugs, optimization opportunities, and best practices.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&apiKey, "api-key", "k", "", fmt.Sprintf("Gemini API key (overrides %s env var)", config.EnvGeminiAPIKey))

	// Custom version template
	rootCmd.SetVersionTemplate(fmt.Sprintf("revcli version %s (commit: %s, built: %s)\n", version, commit, date))
}

// initConfig reads in config from environment variables
func initConfig() {
	if apiKey == "" {
		apiKey = os.Getenv(config.EnvGeminiAPIKey)
	}

	if apiKey == "" {
		fmt.Fprintf(os.Stderr, "Warning: %s not set. Set it via environment variable or --api-key flag.\n", config.EnvGeminiAPIKey)
	}
}

// GetAPIKey returns the configured API key
func GetAPIKey() string {
	return apiKey
}
