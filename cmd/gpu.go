package cmd

import (
	"fmt"
	"os"

	"github.com/jun/linux-toolkit/pkg/gpu"
	"github.com/spf13/cobra"
)

var (
	gpuCardName     string
	gpuOutputFormat string
	showAllGPUs     bool
	showConnectors  bool
	showModes       bool
)

var gpuCmd = &cobra.Command{
	Use:   "gpu",
	Short: "Display verbose GPU information",
	Long: `Display detailed GPU information including device name, vendor, driver, memory,
clocks, temperature, power usage, utilization, and connector information.`,
	Run: func(cmd *cobra.Command, args []string) {
		reader := gpu.NewGPUReader()

		// Handle --all flag
		if showAllGPUs {
			showAllGPUsFunc(reader)
			return
		}

		// Determine which GPU to show
		var cardName string
		if gpuCardName != "" {
			cardName = gpuCardName
		} else {
			// Auto-detect first GPU
			cards, err := reader.ListGPUs()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to find GPUs: %v\n", err)
				os.Exit(1)
			}
			cardName = cards[0]
		}

		// Read GPU information
		gpuInfo, err := reader.ReadGPU(cardName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to read GPU information: %v\n", err)
			os.Exit(1)
		}

		// Parse output format
		format := gpu.FormatTable
		if gpuOutputFormat == "json" {
			format = gpu.FormatJSON
		}

		// Print GPU information
		if err := gpu.Print(gpuInfo, format); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to print GPU information: %v\n", err)
			os.Exit(1)
		}
	},
}

// showAllGPUsFunc displays all available GPUs
func showAllGPUsFunc(reader *gpu.GPUReader) {
	var gpus []*gpu.GPUInfo

	// Get all GPU cards
	cards, err := reader.ListGPUs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to find GPUs: %v\n", err)
		os.Exit(1)
	}

	// Read information for each GPU
	for _, cardName := range cards {
		gpuInfo, err := reader.ReadGPU(cardName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to read GPU %s: %v\n", cardName, err)
			continue
		}
		gpus = append(gpus, gpuInfo)
	}

	if len(gpus) == 0 {
		fmt.Println("No GPUs found.")
		return
	}

	// Parse output format
	format := gpu.FormatTable
	if gpuOutputFormat == "json" {
		format = gpu.FormatJSON
	}

	// Print in appropriate format
	if format == gpu.FormatJSON {
		for _, gpuInfo := range gpus {
			if err := gpu.Print(gpuInfo, format); err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to print GPU information: %v\n", err)
				os.Exit(1)
			}
		}
	} else {
		gpu.PrintMultipleGPUs(gpus)
	}
}

func init() {
	rootCmd.AddCommand(gpuCmd)

	gpuCmd.Flags().StringVarP(&gpuCardName, "card", "c", "", "GPU card name (e.g., card0, card1)")
	gpuCmd.Flags().StringVarP(&gpuOutputFormat, "format", "f", "table", "Output format (table|json)")
	gpuCmd.Flags().BoolVarP(&showAllGPUs, "all", "a", false, "Show all GPUs")
	gpuCmd.Flags().BoolVar(&showConnectors, "show-connectors", false, "Show connector information")
	gpuCmd.Flags().BoolVar(&showModes, "show-modes", false, "Show supported display modes")
}
