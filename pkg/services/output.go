package services

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
	ShowPorts  bool
	ShowAll    bool
	ShowDetail bool
}

// Print prints services information in specified format
func Print(summary *ServicesSummary, format OutputFormat, opts PrintOptions) error {
	switch format {
	case FormatJSON:
		return PrintJSON(summary, opts)
	case FormatTable:
		PrintTable(summary, opts)
		return nil
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

// PrintTable prints services information in a human-readable table format with colors and emojis
func PrintTable(summary *ServicesSummary, opts PrintOptions) {
	// Define colors
	titleColor := color.New(color.FgCyan, color.Bold)
	sectionColor := color.New(color.FgYellow, color.Bold)
	labelColor := color.New(color.FgHiWhite)
	valueColor := color.New(color.FgWhite)

	// Print title
	titleColor.Println("  🛠️  Services Information")
	fmt.Println()

	// Print Summary
	sectionColor.Println("  📋 Summary")
	fmt.Println("  " + strings.Repeat("─", 40))
	printField(labelColor, valueColor, "Total Services", fmt.Sprintf("%d", summary.TotalServices))
	printField(labelColor, valueColor, "Running", fmt.Sprintf("%d", summary.Running))
	printField(labelColor, valueColor, "Stopped", fmt.Sprintf("%d", summary.Stopped))
	printField(labelColor, valueColor, "Failed", fmt.Sprintf("%d", summary.Failed))
	printField(labelColor, valueColor, "Listening Ports", fmt.Sprintf("%d", summary.TotalPorts))
	fmt.Println()

	// Filter services based on options
	services := summary.Services
	if !opts.ShowAll {
		services = filterByStatus(services, []string{"running"})
	}

	// Group services by status
	runningServices := filterByStatus(services, []string{"running"})
	failedServices := filterByStatus(services, []string{"failed"})
	stoppedServices := filterByStatus(services, []string{"stopped"})

	// Print Running Services
	if len(runningServices) > 0 {
		sectionColor.Println("  🔄 Running Services")
		printServicesTable(runningServices, opts.ShowPorts)
		fmt.Println()
	}

	// Print Failed Services
	if len(failedServices) > 0 {
		sectionColor.Println("  🔴 Failed Services")
		printServicesTable(failedServices, opts.ShowPorts)
		fmt.Println()
	}

	// Print Stopped Services
	if len(stoppedServices) > 0 && opts.ShowAll {
		sectionColor.Println("  ⏸️  Stopped Services")
		printServicesTable(stoppedServices, opts.ShowPorts)
		fmt.Println()
	}
}

// printServicesTable prints services in a table format
func printServicesTable(services []ServiceInfo, showPorts bool) {
	if len(services) == 0 {
		return
	}

	// Define table columns
	nameWidth := 20
	statusWidth := 10
	pidWidth := 8
	memoryWidth := 10
	cpuWidth := 8
	startTimeWidth := 18
	userWidth := 10
	portsWidth := 30

	// Print header
	fmt.Printf("  %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s\n",
		nameWidth, "NAME",
		statusWidth, "STATUS",
		pidWidth, "PID",
		memoryWidth, "MEMORY",
		cpuWidth, "CPU",
		startTimeWidth, "START TIME",
		userWidth, "USER",
		portsWidth, "PORTS")
	fmt.Println("  " + strings.Repeat("─", nameWidth+statusWidth+pidWidth+memoryWidth+cpuWidth+startTimeWidth+userWidth+portsWidth+7*3))

	// Print each service
	for _, service := range services {
		// Get status color and icon
		statusColor := getStatusColor(service.Status)
		statusIcon := getStatusIcon(service.Status)

		// Format memory
		memoryStr := formatMemory(service.MemoryMB)
		if service.PID <= 0 {
			memoryStr = "-"
		}

		// Format CPU
		cpuStr := formatCPU(service.CPUPercent)
		if service.PID <= 0 {
			cpuStr = "-"
		}

		// Format start time
		startTimeStr := service.StartTime
		if service.PID <= 0 {
			startTimeStr = "-"
		}

		// Format ports
		portsStr := "-"
		if showPorts && len(service.ListeningPorts) > 0 {
			var portStrs []string
			for _, port := range service.ListeningPorts {
				protocolIcon := getProtocolIcon(port.Protocol)
				portStrs = append(portStrs, fmt.Sprintf("%s%s:%d/%s",
					protocolIcon, port.LocalAddr, port.Port, port.Protocol))
			}
			portsStr = strings.Join(portStrs, ", ")
			if len(portsStr) > portsWidth {
				portsStr = portsStr[:portsWidth-3] + "..."
			}
		}

		// Print row
		fmt.Printf("  %-*s ", nameWidth, truncateString(service.Name, nameWidth))
		statusColor.Printf("%-*s ", statusWidth, statusIcon+" "+service.Status)
		fmt.Printf("%-*s ", pidWidth, formatPID(service.PID))
		fmt.Printf("%-*s ", memoryWidth, memoryStr)
		fmt.Printf("%-*s ", cpuWidth, cpuStr)
		fmt.Printf("%-*s ", startTimeWidth, startTimeStr)
		fmt.Printf("%-*s ", userWidth, truncateString(service.User, userWidth))
		fmt.Printf("%-*s\n", portsWidth, portsStr)
	}
}

// PrintDetail prints detailed information for a single service
func PrintDetail(service *ServiceInfo) {
	// Define colors
	titleColor := color.New(color.FgCyan, color.Bold)
	sectionColor := color.New(color.FgYellow, color.Bold)
	labelColor := color.New(color.FgHiWhite)
	valueColor := color.New(color.FgWhite)

	// Print title
	titleColor.Printf("  🛠️  Service Details: %s\n", service.Name)
	fmt.Println()

	// General Info Section
	sectionColor.Println("  📋 General Info")
	fmt.Println("  " + strings.Repeat("─", 40))
	printField(labelColor, valueColor, "Name", service.Name)
	printField(labelColor, valueColor, "Description", service.Description)

	statusColor := getStatusColor(service.Status)
	statusIcon := getStatusIcon(service.Status)
	fmt.Printf("  %-18s ", "Status:")
	statusColor.Printf("%s %s\n", statusIcon, service.Status)

	printField(labelColor, valueColor, "Loaded", service.Loaded)
	printField(labelColor, valueColor, "Active", service.Active)
	printField(labelColor, valueColor, "Sub State", service.SubState)
	printField(labelColor, valueColor, "Type", service.Type)
	fmt.Println()

	// Process Info Section
	sectionColor.Println("  📊 Process Info")
	fmt.Println("  " + strings.Repeat("─", 40))
	printField(labelColor, valueColor, "PID", formatPID(service.PID))
	printField(labelColor, valueColor, "User", service.User)
	printField(labelColor, valueColor, "Memory", formatMemory(service.MemoryMB))
	printField(labelColor, valueColor, "CPU", formatCPU(service.CPUPercent))
	printField(labelColor, valueColor, "Start Time", service.StartTime)
	printField(labelColor, valueColor, "Uptime", service.Uptime)
	printField(labelColor, valueColor, "Command", service.Command)
	fmt.Println()

	// Listening Ports Section
	if len(service.ListeningPorts) > 0 {
		sectionColor.Println("  🌐 Listening Ports")
		fmt.Println("  " + strings.Repeat("─", 40))
		for _, port := range service.ListeningPorts {
			protocolIcon := getProtocolIcon(port.Protocol)
			fmt.Printf("  %s %-6s %-20s %-8s %s (%d)\n",
				protocolIcon, port.Protocol, port.LocalAddr, port.State, port.ProcessName, port.PID)
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

// getStatusColor returns the appropriate color based on service status
func getStatusColor(status string) *color.Color {
	switch strings.ToLower(status) {
	case "running", "active":
		return color.New(color.FgGreen)
	case "stopped", "inactive":
		return color.New(color.FgYellow)
	case "failed", "dead":
		return color.New(color.FgRed)
	default:
		return color.New(color.FgWhite)
	}
}

// getStatusIcon returns emoji based on service status
func getStatusIcon(status string) string {
	switch strings.ToLower(status) {
	case "running", "active":
		return "🟢"
	case "stopped", "inactive":
		return "⏸️"
	case "failed", "dead":
		return "🔴"
	default:
		return "❓"
	}
}

// getProtocolIcon returns emoji based on protocol
func getProtocolIcon(protocol string) string {
	switch strings.ToLower(protocol) {
	case "tcp", "tcp6":
		return "🟢"
	case "udp", "udp6":
		return "🔵"
	default:
		return "⚪"
	}
}

// formatMemory formats memory usage in MB
func formatMemory(mb float64) string {
	if mb <= 0 {
		return "-"
	}
	return fmt.Sprintf("%.1f MB", mb)
}

// formatCPU formats CPU usage percentage
func formatCPU(percent float64) string {
	if percent <= 0 {
		return "-"
	}
	return fmt.Sprintf("%.1f%%", percent)
}

// formatPID formats PID
func formatPID(pid int) string {
	if pid <= 0 {
		return "-"
	}
	return fmt.Sprintf("%d", pid)
}

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// filterByStatus filters services by status
func filterByStatus(services []ServiceInfo, statuses []string) []ServiceInfo {
	var filtered []ServiceInfo
	for _, service := range services {
		for _, status := range statuses {
			if strings.EqualFold(service.Status, status) {
				filtered = append(filtered, service)
				break
			}
		}
	}
	return filtered
}

// PrintJSON prints services information in JSON format
func PrintJSON(summary *ServicesSummary, opts PrintOptions) error {
	// Filter services based on options
	services := summary.Services
	if !opts.ShowAll {
		services = filterByStatus(services, []string{"running"})
	}

	// Create output structure
	output := struct {
		*ServicesSummary
		Services []ServiceInfo `json:"services"`
	}{
		ServicesSummary: summary,
		Services:        services,
	}

	// Optionally filter ports
	if !opts.ShowPorts {
		for i := range output.Services {
			output.Services[i].ListeningPorts = nil
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}
