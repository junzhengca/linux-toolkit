package cmd

import (
	"fmt"
	"os"

	"github.com/jun/linux-toolkit/pkg/cpu"
	"github.com/spf13/cobra"
)

var (
	cpuOutputFormat string
	showCores       bool
	showFlags       bool
	showTemp        bool
)

var cpuCmd = &cobra.Command{
	Use:   "cpu",
	Short: "Display verbose CPU information",
	Long: `Display detailed CPU information including model name, architecture, vendor ID,
CPU family, stepping, cache sizes, core counts, clock speeds, flags/features,
temperature, load averages, and per-core usage.`,
	Run: func(cmd *cobra.Command, args []string) {
		reader := cpu.NewCPUReader()

		// Read CPU information
		info, err := reader.ReadCPU()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to read CPU information: %v\n", err)
			os.Exit(1)
		}

		// Parse output format
		format := cpu.FormatTable
		if cpuOutputFormat == "json" {
			format = cpu.FormatJSON
		}

		// Set print options
		opts := cpu.PrintOptions{
			ShowCores: showCores,
			ShowFlags: showFlags,
			ShowTemp:  showTemp,
		}

		// Print CPU information
		if err := cpu.Print(info, format, opts); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to print CPU information: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(cpuCmd)

	cpuCmd.Flags().StringVarP(&cpuOutputFormat, "format", "f", "table", "Output format (table|json)")
	cpuCmd.Flags().BoolVar(&showCores, "show-cores", false, "Show per-core usage details")
	cpuCmd.Flags().BoolVar(&showFlags, "show-flags", false, "Show all CPU flags/features")
	cpuCmd.Flags().BoolVar(&showTemp, "show-temp", false, "Include temperature information")
}
