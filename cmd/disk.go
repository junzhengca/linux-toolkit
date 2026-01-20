package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jun/linux-toolkit/pkg/disk"
	"github.com/spf13/cobra"
)

var (
	diskName         string
	mountPoint       string
	showAll          bool
	diskOutputFormat string
	includeIO        bool
	includeInode     bool
)

var diskCmd = &cobra.Command{
	Use:   "disk",
	Short: "Display verbose disk status",
	Long: `Display detailed disk information including device name, size, used/available space,
usage percentage, filesystem type, mount point, UUID, label, block size, I/O statistics,
inode usage, and device model/serial number.`,
	Run: func(cmd *cobra.Command, args []string) {
		reader := disk.NewDiskReader()

		// Handle --all flag
		if showAll {
			showAllDisks(reader)
			return
		}

		var devicePath string

		// Determine which disk to show
		if diskName != "" {
			// User specified a device name
			if !strings.HasPrefix(diskName, "/dev/") {
				devicePath = "/dev/" + diskName
			} else {
				devicePath = diskName
			}
		} else if mountPoint != "" {
			// User specified a mount point, find the device
			// We'll read all mounted disks and find the one matching the mount point
			mountedDisks, err := reader.ListMountedDisks()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to read mounts: %v\n", err)
				os.Exit(1)
			}
			for _, devPath := range mountedDisks {
				diskInfo, err := reader.ReadDisk(devPath)
				if err != nil {
					continue
				}
				if diskInfo.MountPoint == mountPoint {
					devicePath = devPath
					break
				}
			}
			if devicePath == "" {
				fmt.Fprintf(os.Stderr, "Error: No device found mounted at %s\n", mountPoint)
				os.Exit(1)
			}
		} else {
			// Auto-detect first mounted disk
			mountedDisks, err := reader.ListMountedDisks()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to find mounted disks: %v\n", err)
				os.Exit(1)
			}
			devicePath = mountedDisks[0]
		}

		// Read disk information
		diskInfo, err := reader.ReadDisk(devicePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to read disk information: %v\n", err)
			os.Exit(1)
		}

		// Read I/O statistics if requested
		var ioStats *disk.IOStats
		if includeIO {
			ioStats, err = reader.ReadIOStats(devicePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to read I/O statistics: %v\n", err)
			}
		}

		// Read inode information if requested and disk is mounted
		var inodeInfo *disk.InodeInfo
		if includeInode && diskInfo.MountPoint != "" {
			inodeInfo, err = reader.ReadInodeInfo(diskInfo.MountPoint)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to read inode information: %v\n", err)
			}
		}

		// Parse output format
		format := disk.FormatTable
		if diskOutputFormat == "json" {
			format = disk.FormatJSON
		}

		// Print disk information
		if err := disk.Print(diskInfo, ioStats, inodeInfo, format); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to print disk information: %v\n", err)
			os.Exit(1)
		}
	},
}

// showAllDisks displays all available disks
func showAllDisks(reader *disk.DiskReader) {
	var disks []*disk.DiskInfo

	// Get all mounted disks
	mountedDisks, err := reader.ListMountedDisks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to find mounted disks: %v\n", err)
		os.Exit(1)
	}

	// Read information for each disk
	for _, devicePath := range mountedDisks {
		diskInfo, err := reader.ReadDisk(devicePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to read disk %s: %v\n", devicePath, err)
			continue
		}
		disks = append(disks, diskInfo)
	}

	if len(disks) == 0 {
		fmt.Println("No disks found.")
		return
	}

	// Print in table format
	disk.PrintMultipleDisks(disks, includeIO)
}

func init() {
	rootCmd.AddCommand(diskCmd)

	diskCmd.Flags().StringVarP(&diskName, "device", "d", "", "Disk device name (e.g., sda1, nvme0n1p1)")
	diskCmd.Flags().StringVarP(&mountPoint, "mount", "m", "", "Filter by mount point")
	diskCmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all mounted disks")
	diskCmd.Flags().StringVarP(&diskOutputFormat, "format", "f", "table", "Output format (table|json)")
	diskCmd.Flags().BoolVar(&includeIO, "io-stats", false, "Include I/O statistics")
	diskCmd.Flags().BoolVar(&includeInode, "inode-stats", false, "Include inode statistics")
}
