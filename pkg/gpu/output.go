package gpu

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

// Print prints GPU information in specified format
func Print(info *GPUInfo, format OutputFormat) error {
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

// PrintTable prints GPU information in a human-readable table format with colors and emojis
func PrintTable(info *GPUInfo) {
	// Define colors
	titleColor := color.New(color.FgCyan, color.Bold)
	sectionColor := color.New(color.FgYellow, color.Bold)
	labelColor := color.New(color.FgHiWhite)
	valueColor := color.New(color.FgWhite)

	// Vendor color
	vendorColor := getVendorColor(info.Vendor)

	// Print title
	vendorIcon := getVendorIcon(info.Vendor)
	titleColor.Printf("  %s GPU Information\n", vendorIcon)
	fmt.Println()

	// Device Info Section
	sectionColor.Println("  📋 Device Info")
	fmt.Println("  " + strings.Repeat("─", 40))
	printField(labelColor, valueColor, "Card Name", info.CardName)
	printField(labelColor, valueColor, "Device", info.DeviceName)
	printField(labelColor, valueColor, "Driver", info.Driver)
	fmt.Printf("  %-18s ", "Vendor:")
	vendorColor.Printf("%s\n", info.Vendor)
	printField(labelColor, valueColor, "Bus ID", info.BusID)
	printField(labelColor, valueColor, "PCIe Gen", fmt.Sprintf("%s (max: %s)", info.PCIEGen, info.MaxPCIEGen))
	printField(labelColor, valueColor, "Bus Width", info.BusWidth)
	printField(labelColor, valueColor, "Type", info.GPUType)
	fmt.Println()

	// Hardware Info Section
	sectionColor.Println("  🔧 Hardware")
	fmt.Println("  " + strings.Repeat("─", 40))
	printField(labelColor, valueColor, "Vendor ID", info.VendorID)
	printField(labelColor, valueColor, "Device ID", info.DeviceID)
	printField(labelColor, valueColor, "Class", info.Class)
	printField(labelColor, valueColor, "Revision", info.Revision)
	if info.VBIOSVersion != "" {
		printField(labelColor, valueColor, "VBIOS Version", info.VBIOSVersion)
	}
	if info.FirmwareVersion != "" {
		printField(labelColor, valueColor, "Firmware Version", info.FirmwareVersion)
	}
	fmt.Println()

	// Memory Section
	sectionColor.Println("  💾 Memory")
	fmt.Println("  " + strings.Repeat("─", 40))
	if info.VRAMSize > 0 {
		printField(labelColor, valueColor, "VRAM Total", FormatBytes(info.VRAMSize))
	}
	if info.VRAMUsed > 0 {
		printField(labelColor, valueColor, "VRAM Used", FormatBytes(info.VRAMUsed))
	}
	if info.VRAMFree > 0 {
		printField(labelColor, valueColor, "VRAM Free", FormatBytes(info.VRAMFree))
	}
	if info.GARTSize > 0 {
		printField(labelColor, valueColor, "GART Size", FormatBytes(info.GARTSize))
	}
	fmt.Println()

	// Clocks Section
	sectionColor.Println("  ⚡ Clocks")
	fmt.Println("  " + strings.Repeat("─", 40))
	if info.CoreClock > 0 {
		if info.MaxCoreClock > 0 {
			printField(labelColor, valueColor, "Core Clock", fmt.Sprintf("%s (max: %s)", FormatMHz(info.CoreClock), FormatMHz(info.MaxCoreClock)))
		} else {
			printField(labelColor, valueColor, "Core Clock", FormatMHz(info.CoreClock))
		}
	}
	if info.MemoryClock > 0 {
		if info.MaxMemoryClock > 0 {
			printField(labelColor, valueColor, "Memory Clock", fmt.Sprintf("%s (max: %s)", FormatMHz(info.MemoryClock), FormatMHz(info.MaxMemoryClock)))
		} else {
			printField(labelColor, valueColor, "Memory Clock", FormatMHz(info.MemoryClock))
		}
	}
	fmt.Println()

	// Temperature Section
	sectionColor.Println("  🌡️  Temperature")
	fmt.Println("  " + strings.Repeat("─", 40))
	if info.Temperature > 0 {
		tempColor := getTempColor(info.Temperature)
		fmt.Printf("  %-18s ", "Temperature:")
		tempColor.Printf("%.1f°C", info.Temperature)
		if info.TemperatureCrit > 0 {
			fmt.Printf(" (crit: %.1f°C)\n", info.TemperatureCrit)
		} else {
			fmt.Println()
		}
	}
	if info.FanSpeed > 0 {
		printField(labelColor, valueColor, "Fan Speed", fmt.Sprintf("%d RPM", info.FanSpeed))
	}
	if info.FanSpeedPercent > 0 {
		printField(labelColor, valueColor, "Fan Speed", fmt.Sprintf("%.1f%%", info.FanSpeedPercent))
	}
	fmt.Println()

	// Power Section
	sectionColor.Println("  🔌 Power")
	fmt.Println("  " + strings.Repeat("─", 40))
	if info.PowerUsage > 0 {
		printField(labelColor, valueColor, "Power Usage", fmt.Sprintf("%.1f W", float64(info.PowerUsage)/1000.0))
	}
	if info.PowerLimit > 0 {
		printField(labelColor, valueColor, "Power Limit", fmt.Sprintf("%.1f W", float64(info.PowerLimit)/1000.0))
	}
	if info.PowerCap > 0 {
		printField(labelColor, valueColor, "Power Cap", fmt.Sprintf("%.1f W", float64(info.PowerCap)/1000.0))
	}
	fmt.Println()

	// Utilization Section
	sectionColor.Println("  📊 Utilization")
	fmt.Println("  " + strings.Repeat("─", 40))
	if info.GPUUtilPercent > 0 {
		utilColor := getUtilColor(info.GPUUtilPercent)
		fmt.Printf("  %-18s ", "GPU Usage:")
		utilColor.Printf("%.1f%%\n", info.GPUUtilPercent)
	}
	if info.MemoryUtilPercent > 0 {
		utilColor := getUtilColor(info.MemoryUtilPercent)
		fmt.Printf("  %-18s ", "Memory Usage:")
		utilColor.Printf("%.1f%%\n", info.MemoryUtilPercent)
	}
	fmt.Println()

	// Connectors Section
	if len(info.Connectors) > 0 {
		sectionColor.Println("  🔌 Connectors")
		fmt.Println("  " + strings.Repeat("─", 40))
		for _, conn := range info.Connectors {
			status := "disconnected"
			statusColor := color.New(color.FgHiBlack)
			for _, enabled := range info.EnabledConnectors {
				if conn == enabled {
					status = "connected"
					statusColor = color.New(color.FgGreen)
					break
				}
			}
			fmt.Printf("  %-18s ", conn+":")
			statusColor.Printf("%s\n", status)
		}
		fmt.Println()
	}

	// Modes Section
	if len(info.Modes) > 0 {
		sectionColor.Println("  🖥️  Display Modes")
		fmt.Println("  " + strings.Repeat("─", 40))
		// Show first 10 modes
		maxModes := 10
		if len(info.Modes) < maxModes {
			maxModes = len(info.Modes)
		}
		for i := 0; i < maxModes; i++ {
			fmt.Printf("  %s\n", info.Modes[i])
		}
		if len(info.Modes) > maxModes {
			fmt.Printf("  ... and %d more modes\n", len(info.Modes)-maxModes)
		}
		fmt.Println()
	}
}

// PrintMultipleGPUs prints all GPUs in a compact table format
func PrintMultipleGPUs(gpus []*GPUInfo) {
	if len(gpus) == 0 {
		fmt.Println("No GPUs found.")
		return
	}

	titleColor := color.New(color.FgCyan, color.Bold)
	sectionColor := color.New(color.FgYellow, color.Bold)

	titleColor.Println("  🎮 All GPUs")
	fmt.Println()

	for i, gpu := range gpus {
		if i > 0 {
			fmt.Println()
		}
		sectionColor.Printf("  %s\n", gpu.CardName)
		fmt.Println("  " + strings.Repeat("─", 40))

		vendorColor := getVendorColor(gpu.Vendor)
		vendorIcon := getVendorIcon(gpu.Vendor)

		fmt.Printf("  %s ", vendorIcon)
		vendorColor.Printf("%s", gpu.Vendor)
		if gpu.Driver != "" {
			fmt.Printf(" | %s", gpu.Driver)
		}
		if gpu.BusID != "" {
			fmt.Printf(" | %s", gpu.BusID)
		}
		fmt.Println()

		if gpu.VRAMSize > 0 {
			fmt.Printf("  Memory: %s", FormatBytes(gpu.VRAMSize))
		}
		if gpu.Temperature > 0 {
			tempColor := getTempColor(gpu.Temperature)
			fmt.Printf(" | Temp: ")
			tempColor.Printf("%.1f°C", gpu.Temperature)
		}
		if gpu.GPUUtilPercent > 0 {
			utilColor := getUtilColor(gpu.GPUUtilPercent)
			fmt.Printf(" | Usage: ")
			utilColor.Printf("%.1f%%", gpu.GPUUtilPercent)
		}
		fmt.Println()

		if len(gpu.Connectors) > 0 {
			fmt.Printf("  Connectors: %s\n", strings.Join(gpu.Connectors, ", "))
		}
	}
	fmt.Println()
}

// printField prints a labeled field with consistent formatting
func printField(labelColor, valueColor *color.Color, label, value string) {
	if value == "" {
		return
	}
	fmt.Printf("  ")
	labelColor.Printf("%-18s", label+":")
	valueColor.Printf(" %s\n", value)
}

// getVendorColor returns the appropriate color based on vendor
func getVendorColor(vendor string) *color.Color {
	switch vendor {
	case "NVIDIA":
		return color.New(color.FgGreen)
	case "AMD":
		return color.New(color.FgRed)
	case "Intel":
		return color.New(color.FgBlue)
	default:
		return color.New(color.FgWhite)
	}
}

// getVendorIcon returns emoji based on vendor
func getVendorIcon(vendor string) string {
	switch vendor {
	case "NVIDIA":
		return "🟢"
	case "AMD":
		return "🔴"
	case "Intel":
		return "🔵"
	default:
		return "⚪"
	}
}

// getUtilColor returns the appropriate color based on utilization percentage
func getUtilColor(percent float64) *color.Color {
	switch {
	case percent >= 91:
		return color.New(color.FgRed)
	case percent >= 76:
		return color.New(color.FgHiYellow)
	case percent >= 51:
		return color.New(color.FgYellow)
	case percent >= 26:
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

// PrintJSON prints GPU information in JSON format
func PrintJSON(info *GPUInfo) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(info)
}
