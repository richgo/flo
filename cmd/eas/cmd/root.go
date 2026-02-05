package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "eas",
	Short: "Enterprise AI SDLC - AI-powered development workflow",
	Long: `EAS (Enterprise AI SDLC) orchestrates AI agents for spec-driven,
test-driven development across multiple repositories.

Initialize a feature workspace, create tasks, and let AI agents
implement them with TDD enforcement.`,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(taskCmd)
	rootCmd.AddCommand(statusCmd)
}
