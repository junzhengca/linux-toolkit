package battery

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

// Print prints battery information in specified format
func Print(info *BatteryInfo, format OutputFormat) error {
	switch format {
	case FormatJSON:
		return PrintJSON(info)
	case FormatTable:
		PrintTable(info)
		return nil
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

// PrintTable prints battery information in a human-readable table format with colors and emojis
func PrintTable(info *BatteryInfo) {
	// Define colors
	titleColor := color.New(color.FgCyan, color.Bold)
	sectionColor := color.New(color.FgYellow, color.Bold)
	labelColor := color.New(color.FgHiWhite)
	valueColor := color.New(color.FgWhite)

	// Health colors
	healthColor := getHealthColor(info.HealthPercent)

	// Status emoji and color
	statusEmoji, statusColor := getStatusEmojiAndColor(info.Status)

	// Print title
	titleColor.Println("  🔋 Battery Information")
	fmt.Println()

	// Device Info Section
	sectionColor.Println("  📱 Device Info")
	fmt.Println("  " + strings.Repeat("─", 40))
	printField(labelColor, valueColor, "Device", info.Name)
	printField(labelColor, valueColor, "Manufacturer", info.Manufacturer)
	printField(labelColor, valueColor, "Model", info.ModelName)
	printField(labelColor, valueColor, "Serial Number", info.SerialNumber)
	printField(labelColor, valueColor, "Technology", info.Technology)
	fmt.Println()

	// Status Section
	sectionColor.Println("  📊 Status")
	fmt.Println("  " + strings.Repeat("─", 40))
	printField(labelColor, valueColor, "Capacity", fmt.Sprintf("%d%%", info.Capacity))

	// Status with emoji and color
	fmt.Printf("  %-18s ", "Status")
	statusColor.Printf("%s %s\n", statusEmoji, info.Status)

	// Health with color
	fmt.Printf("  %-18s ", "Health")
	healthColor.Printf("%.1f%% (%s)\n", info.HealthPercent, info.Health)

	printField(labelColor, valueColor, "Cycle Count", fmt.Sprintf("%d", info.CycleCount))
	fmt.Println()

	// Capacity Section
	sectionColor.Println("  ⚡ Capacity")
	fmt.Println("  " + strings.Repeat("─", 40))
	printField(labelColor, valueColor, "Design Capacity", fmt.Sprintf("%.2f Wh", info.DesignCapacity))
	printField(labelColor, valueColor, "Current Full Cap", fmt.Sprintf("%.2f Wh", info.CurrentFullCap))
	fmt.Println()

	// Power Section
	sectionColor.Println("  🔌 Power")
	fmt.Println("  " + strings.Repeat("─", 40))
	printField(labelColor, valueColor, "Voltage", fmt.Sprintf("%.2f V (min: %.2f V)", info.Voltage, info.VoltageMinDesign))
	printField(labelColor, valueColor, "Current", fmt.Sprintf("%.2f A", info.Current))
	printField(labelColor, valueColor, "Power", fmt.Sprintf("%.2f W", info.Power))
	fmt.Println()

	// Time Remaining Section
	sectionColor.Println("  ⏱️  Time Remaining")
	fmt.Println("  " + strings.Repeat("─", 40))
	fmt.Printf("  %-18s ", "")
	valueColor.Printf("%s\n", info.TimeRemaining)
	fmt.Println()
}

// printField prints a labeled field with consistent formatting
func printField(labelColor, valueColor *color.Color, label, value string) {
	fmt.Printf("  ")
	labelColor.Printf("%-18s", label+":")
	valueColor.Printf(" %s\n", value)
}

// getHealthColor returns the appropriate color based on health percentage
func getHealthColor(healthPercent float64) *color.Color {
	switch {
	case healthPercent >= 90:
		return color.New(color.FgGreen)
	case healthPercent >= 75:
		return color.New(color.FgHiGreen)
	case healthPercent >= 50:
		return color.New(color.FgYellow)
	case healthPercent >= 25:
		return color.New(color.FgHiYellow)
	default:
		return color.New(color.FgRed)
	}
}

// getStatusEmojiAndColor returns emoji and color based on battery status
func getStatusEmojiAndColor(status string) (string, *color.Color) {
	switch status {
	case "Charging":
		return "⚡", color.New(color.FgGreen)
	case "Discharging":
		return "📉", color.New(color.FgYellow)
	case "Full":
		return "✅", color.New(color.FgGreen)
	case "Unknown":
		return "❓", color.New(color.FgHiWhite)
	default:
		return "⚠️", color.New(color.FgRed)
	}
}

// PrintJSON prints battery information in JSON format
func PrintJSON(info *BatteryInfo) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(info)
}
