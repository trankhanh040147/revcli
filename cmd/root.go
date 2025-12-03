package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Config holds global configuration
	apiKey string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-rev-cli",
	Short: "Gemini-powered code reviewer CLI for Go developers",
	Long: `go-rev-cli is a local command-line tool that acts as an intelligent peer reviewer.
It reads your local git changes and uses Google's Gemini LLM to analyze your code
for bugs, optimization opportunities, and idiomatic Go practices.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "Gemini API key (overrides GEMINI_API_KEY env var)")
}

// initConfig reads in config from environment variables
func initConfig() {
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	}

	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Warning: GEMINI_API_KEY not set. Set it via environment variable or --api-key flag.")
	}
}

// GetAPIKey returns the configured API key
func GetAPIKey() string {
	return apiKey
}

