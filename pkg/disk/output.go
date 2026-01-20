package disk

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

// Print prints disk information in specified format
func Print(disk *DiskInfo, ioStats *IOStats, inodeInfo *InodeInfo, format OutputFormat) error {
	switch format {
	case FormatJSON:
		return PrintJSON(disk, ioStats, inodeInfo)
	case FormatTable:
		PrintTable(disk, ioStats, inodeInfo)
		return nil
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

// PrintTable prints disk information in a human-readable table format with colors and emojis
func PrintTable(disk *DiskInfo, ioStats *IOStats, inodeInfo *InodeInfo) {
	// Define colors
	titleColor := color.New(color.FgCyan, color.Bold)
	sectionColor := color.New(color.FgYellow, color.Bold)
	labelColor := color.New(color.FgHiWhite)
	valueColor := color.New(color.FgWhite)

	// Usage colors
	usageColor := getUsageColor(disk.UsagePercent)

	// Filesystem icon
	fsIcon := getFileSystemIcon(disk.FilesystemType)

	// Device type
	deviceType := getDeviceType(disk)

	// Print title
	titleColor.Println("  💾 Disk Information")
	fmt.Println()

	// Device Info Section
	sectionColor.Println("  📱 Device Info")
	fmt.Println("  " + strings.Repeat("─", 40))
	printField(labelColor, valueColor, "Device", disk.Name)
	printField(labelColor, valueColor, "Path", disk.Path)
	printField(labelColor, valueColor, "Model", disk.Model)
	printField(labelColor, valueColor, "Serial", disk.Serial)
	printField(labelColor, valueColor, "Filesystem", fmt.Sprintf("%s %s", fsIcon, disk.FilesystemType))
	printField(labelColor, valueColor, "UUID", disk.UUID)
	printField(labelColor, valueColor, "Label", disk.Label)
	printField(labelColor, valueColor, "Mount Point", disk.MountPoint)
	printField(labelColor, valueColor, "Block Size", fmt.Sprintf("%d bytes", disk.BlockSize))
	printField(labelColor, valueColor, "Type", deviceType)

	// Flags
	var flags []string
	if disk.ReadOnly {
		flags = append(flags, "read-only")
	}
	if disk.Removable {
		flags = append(flags, "removable")
	}
	if len(flags) > 0 {
		printField(labelColor, valueColor, "Flags", strings.Join(flags, ", "))
	}
	fmt.Println()

	// Space Usage Section
	sectionColor.Println("  📊 Space Usage")
	fmt.Println("  " + strings.Repeat("─", 40))
	printField(labelColor, valueColor, "Total Size", FormatBytes(disk.Size))
	printField(labelColor, valueColor, "Used Space", FormatBytes(disk.Used))
	printField(labelColor, valueColor, "Available Space", FormatBytes(disk.Available))

	// Usage with color
	fmt.Printf("  %-18s ", "Usage")
	usageColor.Printf("%.1f%%\n", disk.UsagePercent)

	// Inode Info
	if inodeInfo != nil {
		fmt.Printf("  %-18s ", "Inodes")
		valueColor.Printf("%s total / %s used (%.1f%%)\n",
			formatNumber(inodeInfo.Total),
			formatNumber(inodeInfo.Used),
			inodeInfo.UsagePercent)
	}
	fmt.Println()

	// I/O Statistics Section
	if ioStats != nil {
		sectionColor.Println("  ⚡ I/O Statistics")
		fmt.Println("  " + strings.Repeat("─", 40))
		printField(labelColor, valueColor, "Reads", fmt.Sprintf("%s (%s)",
			formatNumber(ioStats.ReadsCompleted),
			FormatSectors(ioStats.SectorsRead)))
		printField(labelColor, valueColor, "Writes", fmt.Sprintf("%s (%s)",
			formatNumber(ioStats.WritesCompleted),
			FormatSectors(ioStats.SectorsWritten)))
		printField(labelColor, valueColor, "Read Time", fmt.Sprintf("%d ms", ioStats.ReadTimeMs))
		printField(labelColor, valueColor, "Write Time", fmt.Sprintf("%d ms", ioStats.WriteTimeMs))
		printField(labelColor, valueColor, "I/O Time", fmt.Sprintf("%d ms", ioStats.IOTimeMs))
		if ioStats.IOInProgress > 0 {
			printField(labelColor, valueColor, "I/O In Progress", fmt.Sprintf("%d", ioStats.IOInProgress))
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
	case usagePercent >= 96:
		return color.New(color.FgRed, color.Bold)
	case usagePercent >= 91:
		return color.New(color.FgRed)
	case usagePercent >= 76:
		return color.New(color.FgHiYellow)
	case usagePercent >= 51:
		return color.New(color.FgYellow)
	default:
		return color.New(color.FgGreen)
	}
}

// getFileSystemIcon returns emoji based on filesystem type
func getFileSystemIcon(fsType string) string {
	switch strings.ToLower(fsType) {
	case "ext4", "ext3", "ext2":
		return "📁"
	case "xfs":
		return "📂"
	case "btrfs":
		return "🗄️"
	case "ntfs":
		return "💿"
	case "vfat", "fat32", "exfat":
		return "💾"
	case "swap":
		return "🔄"
	case "tmpfs":
		return "⚡"
	case "nfs", "cifs", "smb":
		return "🌐"
	case "zfs":
		return "🌊"
	default:
		return "📦"
	}
}

// getDeviceType returns a human-readable device type string
func getDeviceType(disk *DiskInfo) string {
	if disk.Rotational {
		return "HDD (rotational)"
	}
	return "SSD (non-rotational)"
}

// formatNumber formats a large number with commas
func formatNumber(n uint64) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}

	var result []rune
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, r)
	}
	return string(result)
}

// PrintJSON prints disk information in JSON format
func PrintJSON(disk *DiskInfo, ioStats *IOStats, inodeInfo *InodeInfo) error {
	output := struct {
		Disk      *DiskInfo  `json:"disk"`
		IOStats   *IOStats   `json:"ioStats,omitempty"`
		InodeInfo *InodeInfo `json:"inodeInfo,omitempty"`
	}{
		Disk:      disk,
		IOStats:   ioStats,
		InodeInfo: inodeInfo,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// PrintMultipleDisks prints multiple disks in table format
func PrintMultipleDisks(disks []*DiskInfo, includeIOStats bool) {
	titleColor := color.New(color.FgCyan, color.Bold)
	headerColor := color.New(color.FgHiWhite, color.Bold)

	// Print title
	titleColor.Println("  💾 All Disks")
	fmt.Println()

	// Print header
	fmt.Println("  " + strings.Repeat("─", 120))
	headerColor.Printf("  %-15s %-12s %-12s %-12s %-15s %-20s %-15s %-10s\n",
		"Device", "Size", "Used", "Avail", "Usage%", "Mount Point", "Filesystem", "Type")
	fmt.Println("  " + strings.Repeat("─", 120))

	// Print each disk
	for _, disk := range disks {
		usageColor := getUsageColor(disk.UsagePercent)
		fsIcon := getFileSystemIcon(disk.FilesystemType)
		deviceType := "SSD"
		if disk.Rotational {
			deviceType = "HDD"
		}

		fmt.Printf("  %-15s %-12s %-12s %-12s ",
			disk.Name,
			FormatBytes(disk.Size),
			FormatBytes(disk.Used),
			FormatBytes(disk.Available))

		usageColor.Printf("%.1f%%", disk.UsagePercent)

		fmt.Printf(" %-15s %-20s %-15s %-10s\n",
			"",
			disk.MountPoint,
			fmt.Sprintf("%s %s", fsIcon, disk.FilesystemType),
			deviceType)
	}

	fmt.Println("  " + strings.Repeat("─", 120))
}
