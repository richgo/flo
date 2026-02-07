package cmd

import (
	"fmt"
	"os"

	"github.com/richgo/flo/pkg/secrets"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management commands",
	Long:  `Commands for managing and displaying configuration and secrets.`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long: `Display current configuration including environment variables and secrets.
Secret values are masked for security.

Supported environment variables:
  - CLAUDE_API_KEY: API key for Claude backend
  - COPILOT_TOKEN: Token for GitHub Copilot backend
  - FLO_BACKEND: Default backend to use (claude/copilot)
  - FLO_MODEL: Default model to use`,
	RunE: runConfigShow,
}

func init() {
	configCmd.AddCommand(configShowCmd)
	rootCmd.AddCommand(configCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	// Load secrets from .env files
	manager, err := secrets.LoadDefault()
	if err != nil {
		return fmt.Errorf("failed to load secrets: %w", err)
	}

	fmt.Println("Flo Configuration")
	fmt.Println("=================")
	fmt.Println()

	// Display well-known environment variables
	fmt.Println("Environment Variables:")
	for _, key := range secrets.WellKnownKeys {
		value := manager.Get(key)
		maskedValue := secrets.Mask(value)
		fmt.Printf("  %-20s %s\n", key+":", maskedValue)
	}
	fmt.Println()

	// Check for .env files
	fmt.Println(".env Files:")
	checkEnvFile := func(path string) {
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("  ✓ %s (loaded)\n", path)
		} else if os.IsNotExist(err) {
			fmt.Printf("  - %s (not found)\n", path)
		}
	}

	checkEnvFile(".env")
	checkEnvFile(".flo/.env")
	fmt.Println()

	// Display backend info
	backend := manager.Get("FLO_BACKEND")
	if backend == "" {
		backend = "claude (default)"
	}
	fmt.Printf("Active Backend: %s\n", backend)

	model := manager.Get("FLO_MODEL")
	if model != "" {
		fmt.Printf("Model: %s\n", model)
	}

	return nil
}
