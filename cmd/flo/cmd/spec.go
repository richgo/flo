package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/richgo/flo/pkg/spec"
	"github.com/spf13/cobra"
)

var specCmd = &cobra.Command{
	Use:   "spec",
	Short: "Spec management commands",
	Long:  `Commands for managing and validating SPEC.md files.`,
}

var specValidateCmd = &cobra.Command{
	Use:   "validate [path]",
	Short: "Validate a SPEC.md file",
	Long: `Validate that a SPEC.md file contains all required sections (Goal, Context, Success Criteria)
and follows proper markdown structure.

If no path is provided, validates .flo/SPEC.md in the current directory.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSpecValidate,
}

func init() {
	specCmd.AddCommand(specValidateCmd)
	rootCmd.AddCommand(specCmd)
}

func runSpecValidate(cmd *cobra.Command, args []string) error {
	// Determine spec file path
	specPath := ".flo/SPEC.md"
	if len(args) > 0 {
		specPath = args[0]
	}

	// Make path absolute
	absPath, err := filepath.Abs(specPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("spec file not found: %s", absPath)
	}

	// Validate the spec
	validator := spec.NewValidator()
	result, err := validator.ValidateFile(absPath)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Display results
	fmt.Printf("Validating: %s\n\n", absPath)

	if result.Valid {
		fmt.Println("✓ Spec is valid!")
		return nil
	}

	// Show validation errors
	fmt.Println("✗ Spec validation failed:")
	fmt.Println()

	if len(result.MissingSections) > 0 {
		fmt.Println("Missing required sections:")
		for _, section := range result.MissingSections {
			fmt.Printf("  - %s\n", section)
		}
		fmt.Println()
	}

	if len(result.Errors) > 0 {
		fmt.Println("Errors:")
		for _, err := range result.Errors {
			fmt.Printf("  - %s\n", err)
		}
		fmt.Println()
	}

	return fmt.Errorf("spec validation failed")
}
