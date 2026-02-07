package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version information set via ldflags during build
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display the version, build commit, and build date of flo.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("flo version %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
