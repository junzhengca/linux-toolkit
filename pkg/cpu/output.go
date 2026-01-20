package cpu

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

// OutputFormat represents output format type
type OutputFormat string

const (
	FormatTable OutputFormat = "table"
	FormatJSON  OutputFormat = "json"
)

// PrintOptions controls what information to display
type PrintOptions struct {
	ShowCores bool
	ShowFlags bool
	ShowTemp  bool
}

// Print prints CPU information in specified format
func Print(info *CPUInfo, format OutputFormat, opts PrintOptions) error {
	switch format {
	case FormatJSON:
		return PrintJSON(info, opts)
	case FormatTable:
		PrintTable(info, opts)
		return nil
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

// PrintTable prints CPU information in a human-readable table format with colors and emojis
func PrintTable(info *CPUInfo, opts PrintOptions) {
	// Define colors
	titleColor := color.New(color.FgCyan, color.Bold)
	sectionColor := color.New(color.FgYellow, color.Bold)
	labelColor := color.New(color.FgHiWhite)
	valueColor := color.New(color.FgWhite)

	// Architecture icon
	archIcon := getArchIcon(info.Architecture)

	// Print title
	titleColor.Printf("  %s CPU Information\n", archIcon)
	fmt.Println()

	// General Info Section
	sectionColor.Println("  📋 General Info")
	fmt.Println("  " + strings.Repeat("─", 40))
	printField(labelColor, valueColor, "Model", info.ModelName)
	printField(labelColor, valueColor, "Vendor", info.VendorID)
	printField(labelColor, valueColor, "Architecture", fmt.Sprintf("%s %s", archIcon, info.Architecture))
	printField(labelColor, valueColor, "Mode", info.CPUMode)
	printField(labelColor, valueColor, "Family", fmt.Sprintf("%d", info.CPUFamily))
	printField(labelColor, valueColor, "Model", fmt.Sprintf("%d", info.Model))
	printField(labelColor, valueColor, "Stepping", fmt.Sprintf("%d", info.Stepping))
	if info.Bogomips > 0 {
		printField(labelColor, valueColor, "Bogomips", fmt.Sprintf("%.2f", info.Bogomips))
	}
	fmt.Println()

	// Core Configuration Section
	sectionColor.Println("  🔧 Core Configuration")
	fmt.Println("  " + strings.Repeat("─", 40))
	printField(labelColor, valueColor, "Physical Cores", fmt.Sprintf("%d", info.PhysicalCores))
	printField(labelColor, valueColor, "Logical Cores", fmt.Sprintf("%d", info.LogicalCores))
	printField(labelColor, valueColor, "Threads per Core", fmt.Sprintf("%d", info.ThreadsPerCore))
	printField(labelColor, valueColor, "Sockets", fmt.Sprintf("%d", info.Sockets))
	fmt.Println()

	// Frequency Section
	sectionColor.Println("  ⚡ Frequency")
	fmt.Println("  " + strings.Repeat("─", 40))
	if info.MinFrequency > 0 {
		printField(labelColor, valueColor, "Min Frequency", FormatMHz(info.MinFrequency))
	}
	if info.MaxFrequency > 0 {
		printField(labelColor, valueColor, "Max Frequency", FormatMHz(info.MaxFrequency))
	}
	if info.CurrentFrequency > 0 {
		printField(labelColor, valueColor, "Current Frequency", FormatMHz(info.CurrentFrequency))
	}
	fmt.Println()

	// Cache Sizes Section
	sectionColor.Println("  💾 Cache Sizes")
	fmt.Println("  " + strings.Repeat("─", 40))
	if info.CacheL1d > 0 {
		printField(labelColor, valueColor, "L1 Data Cache", formatCacheSize(info.CacheL1d))
	}
	if info.CacheL1i > 0 {
		printField(labelColor, valueColor, "L1 Instruction", formatCacheSize(info.CacheL1i))
	}
	if info.CacheL2 > 0 {
		printField(labelColor, valueColor, "L2 Cache", formatCacheSize(info.CacheL2))
	}
	if info.CacheL3 > 0 {
		printField(labelColor, valueColor, "L3 Cache", formatCacheSize(info.CacheL3))
	}
	fmt.Println()

	// System Load Section
	sectionColor.Println("  📊 System Load")
	fmt.Println("  " + strings.Repeat("─", 40))
	printField(labelColor, valueColor, "Load Average (1m)", fmt.Sprintf("%.2f", info.LoadAvg1))
	printField(labelColor, valueColor, "Load Average (5m)", fmt.Sprintf("%.2f", info.LoadAvg5))
	printField(labelColor, valueColor, "Load Average (15m)", fmt.Sprintf("%.2f", info.LoadAvg15))
	if info.TotalProcesses > 0 {
		printField(labelColor, valueColor, "Total Processes", fmt.Sprintf("%d", info.TotalProcesses))
	}
	if info.RunningProcesses > 0 {
		printField(labelColor, valueColor, "Running Processes", fmt.Sprintf("%d", info.RunningProcesses))
	}
	fmt.Println()

	// Temperature Section
	if opts.ShowTemp && info.CPUTemperature > 0 {
		sectionColor.Println("  🌡️  Temperature")
		fmt.Println("  " + strings.Repeat("─", 40))
		tempColor := getTempColor(info.CPUTemperature)
		fmt.Printf("  %-18s ", "CPU Temperature")
		tempColor.Printf("%.1f°C\n", info.CPUTemperature)
		fmt.Println()
	}

	// Core Usage Section
	if opts.ShowCores && len(info.Cores) > 0 {
		sectionColor.Println("  🔌 Core Usage")
		fmt.Println("  " + strings.Repeat("─", 40))
		for _, core := range info.Cores {
			usageColor := getUsageColor(core.UsagePercent)
			fmt.Printf("  %-18s ", fmt.Sprintf("Core %d (CPU%d):", core.CoreID, core.ProcessorID))
			usageColor.Printf("%.1f%%", core.UsagePercent)
			if core.Frequency > 0 {
				valueColor.Printf(" @ %s\n", FormatMHz(core.Frequency))
			} else {
				fmt.Println()
			}
		}
		fmt.Println()
	}

	// CPU Flags Section
	if opts.ShowFlags && len(info.Flags) > 0 {
		sectionColor.Println("  🏷️  CPU Flags")
		fmt.Println("  " + strings.Repeat("─", 40))
		maxFlags := 180
		if len(info.Flags) < maxFlags {
			maxFlags = len(info.Flags)
		}
		displayFlags := info.Flags[:maxFlags]
		fmt.Printf("  Showing %d of %d flags:\n", maxFlags, len(info.Flags))
		for i := 0; i < len(displayFlags); i += 10 {
			end := i + 10
			if end > len(displayFlags) {
				end = len(displayFlags)
			}
			fmt.Printf("  %s\n", strings.Join(displayFlags[i:end], " "))
		}
		if len(info.Flags) > maxFlags {
			fmt.Printf("  ... and %d more flags\n", len(info.Flags)-maxFlags)
		}
		fmt.Println()
	}
}

// printField prints a labeled field with consistent formatting
func printField(labelColor, valueColor *color.Color, label, value string) {
	fmt.Printf("  ")
	labelColor.Printf("%-18s", label+":")
	valueColor.Printf(" %s\n", value)
}

// getUsageColor returns the appropriate color based on usage percentage
func getUsageColor(usagePercent float64) *color.Color {
	switch {
	case usagePercent >= 91:
		return color.New(color.FgRed)
	case usagePercent >= 76:
		return color.New(color.FgHiYellow)
	case usagePercent >= 51:
		return color.New(color.FgYellow)
	case usagePercent >= 26:
		return color.New(color.FgHiGreen)
	default:
		return color.New(color.FgGreen)
	}
}

// getTempColor returns the appropriate color based on temperature
func getTempColor(temp float64) *color.Color {
	switch {
	case temp >= 86:
		return color.New(color.FgRed)
	case temp >= 76:
		return color.New(color.FgHiYellow)
	case temp >= 61:
		return color.New(color.FgYellow)
	case temp >= 46:
		return color.New(color.FgHiGreen)
	default:
		return color.New(color.FgGreen)
	}
}

// getArchIcon returns emoji based on architecture
func getArchIcon(arch string) string {
	switch strings.ToLower(arch) {
	case "x86_64", "amd64":
		return "🖥️"
	case "arm64", "aarch64":
		return "📱"
	case "riscv64":
		return "🔬"
	case "arm", "armv7l":
		return "📱"
	default:
		return "💻"
	}
}

// formatCacheSize formats cache size in a human-readable format
func formatCacheSize(kb int) string {
	if kb >= 1024 {
		return fmt.Sprintf("%d MB (%d KB)", kb/1024, kb)
	}
	return fmt.Sprintf("%d KB", kb)
}

// PrintJSON prints CPU information in JSON format
func PrintJSON(info *CPUInfo, opts PrintOptions) error {
	output := struct {
		*CPUInfo
		ShowCores bool `json:"showCores,omitempty"`
		ShowFlags bool `json:"showFlags,omitempty"`
		ShowTemp  bool `json:"showTemp,omitempty"`
	}{
		CPUInfo:   info,
		ShowCores: opts.ShowCores,
		ShowFlags: opts.ShowFlags,
		ShowTemp:  opts.ShowTemp,
	}

	// Optionally filter output based on options
	if !opts.ShowCores {
		output.Cores = nil
	}
	if !opts.ShowFlags {
		output.Flags = nil
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}
