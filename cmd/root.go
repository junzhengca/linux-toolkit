package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "linux-toolkit",
	Short: "A collection of useful Linux tools",
	Long: `linux-toolkit is a CLI utility that provides various useful tools
for Linux system administration and monitoring.`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
