package cmd

import (
	"fmt"
	"os"

	"github.com/jun/linux-toolkit/pkg/battery"
	"github.com/spf13/cobra"
)

var (
	batteryName string
	outputFormat string
)

var batteryCmd = &cobra.Command{
	Use:   "battery",
	Short: "Display battery statistics",
	Long: `Display detailed battery information including capacity, status,
voltage, current, power, health, and estimated time remaining.`,
	Run: func(cmd *cobra.Command, args []string) {
		reader := battery.NewBatteryReader()

		var deviceName string
		if batteryName != "" {
			deviceName = batteryName
		} else {
			// Auto-detect battery
			batteries, err := reader.ListBatteries()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to find batteries: %v\n", err)
				os.Exit(1)
			}
			deviceName = batteries[0]
		}

		// Read battery information
		info, err := reader.ReadBattery(deviceName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to read battery information: %v\n", err)
			os.Exit(1)
		}

		// Parse output format
		format := battery.FormatTable
		if outputFormat == "json" {
			format = battery.FormatJSON
		}

		// Print battery information
		if err := battery.Print(info, format); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to print battery information: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(batteryCmd)

	batteryCmd.Flags().StringVarP(&batteryName, "battery", "b", "", "Battery device name (e.g., BAT0)")
	batteryCmd.Flags().StringVarP(&outputFormat, "format", "f", "table", "Output format (table|json)")
}
