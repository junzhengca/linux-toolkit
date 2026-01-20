package disk

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// DiskInfo contains detailed disk information
type DiskInfo struct {
	Name           string  `json:"name"`
	Path           string  `json:"path"`
	Size           uint64  `json:"size"`
	Used           uint64  `json:"used"`
	Available      uint64  `json:"available"`
	UsagePercent   float64 `json:"usagePercent"`
	FilesystemType string  `json:"filesystemType"`
	MountPoint     string  `json:"mountPoint"`
	UUID           string  `json:"uuid"`
	Label          string  `json:"label"`
	BlockSize      uint64  `json:"blockSize"`
	Model          string  `json:"model"`
	Serial         string  `json:"serial"`
	ReadOnly       bool    `json:"readOnly"`
	Removable      bool    `json:"removable"`
	Rotational     bool    `json:"rotational"`
}

// IOStats contains disk I/O statistics
type IOStats struct {
	ReadsCompleted   uint64 `json:"readsCompleted"`
	ReadsMerged      uint64 `json:"readsMerged"`
	SectorsRead      uint64 `json:"sectorsRead"`
	ReadTimeMs       uint64 `json:"readTimeMs"`
	WritesCompleted  uint64 `json:"writesCompleted"`
	WritesMerged     uint64 `json:"writesMerged"`
	SectorsWritten   uint64 `json:"sectorsWritten"`
	WriteTimeMs      uint64 `json:"writeTimeMs"`
	IOInProgress     uint64 `json:"ioInProgress"`
	IOTimeMs         uint64 `json:"ioTimeMs"`
	WeightedIOTimeMs uint64 `json:"weightedIOTimeMs"`
}

// InodeInfo contains inode usage information
type InodeInfo struct {
	Total        uint64  `json:"total"`
	Used         uint64  `json:"used"`
	Free         uint64  `json:"free"`
	UsagePercent float64 `json:"usagePercent"`
}

// MountEntry represents a single mount entry
type MountEntry struct {
	Spec    string // Device spec (e.g., /dev/sda1)
	File    string // Mount point
	Vfstype string // Filesystem type
	Mntops  string // Mount options
}

// DiskReader reads disk information from the Linux system
type DiskReader struct {
	sysfsPath          string
	procDiskstatsPath  string
	etcMtabPath        string
	procMountsPath     string
	devDiskByUUIDPath  string
	devDiskByLabelPath string
}

// NewDiskReader creates a new DiskReader
func NewDiskReader() *DiskReader {
	return &DiskReader{
		sysfsPath:          "/sys/block",
		procDiskstatsPath:  "/proc/diskstats",
		etcMtabPath:        "/etc/mtab",
		procMountsPath:     "/proc/mounts",
		devDiskByUUIDPath:  "/dev/disk/by-uuid",
		devDiskByLabelPath: "/dev/disk/by-label",
	}
}

// ListDisks returns a list of available disk device names
func (r *DiskReader) ListDisks() ([]string, error) {
	entries, err := os.ReadDir(r.sysfsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read block devices directory: %w", err)
	}

	var disks []string
	for _, entry := range entries {
		// Skip loopback devices and non-device entries
		if strings.HasPrefix(entry.Name(), "loop") {
			continue
		}
		// Check if this is a block device
		devicePath := filepath.Join(r.sysfsPath, entry.Name())
		if _, err := os.Stat(filepath.Join(devicePath, "device")); err == nil {
			disks = append(disks, entry.Name())
		}
	}

	if len(disks) == 0 {
		return nil, fmt.Errorf("no disks found")
	}

	return disks, nil
}

// ListMountedDisks returns a list of mounted disk device paths
func (r *DiskReader) ListMountedDisks() ([]string, error) {
	mounts, err := r.readMounts()
	if err != nil {
		return nil, fmt.Errorf("failed to read mounts: %w", err)
	}

	var disks []string
	for _, mount := range mounts {
		// Skip virtual filesystems and special mounts
		if strings.HasPrefix(mount.Spec, "/dev/") || strings.HasPrefix(mount.Spec, "/dev/mapper/") {
			disks = append(disks, mount.Spec)
		}
	}

	if len(disks) == 0 {
		return nil, fmt.Errorf("no mounted disks found")
	}

	return disks, nil
}

// ReadDisk reads disk information for the specified device
func (r *DiskReader) ReadDisk(devicePath string) (*DiskInfo, error) {
	// Get device name from path
	deviceName := filepath.Base(devicePath)

	info := &DiskInfo{
		Name: deviceName,
		Path: devicePath,
	}

	// Read mount information
	mounts, err := r.readMounts()
	if err == nil {
		for _, mount := range mounts {
			if mount.Spec == devicePath {
				info.MountPoint = mount.File
				info.FilesystemType = mount.Vfstype
				info.ReadOnly = strings.Contains(mount.Mntops, "ro")
				break
			}
		}
	}

	// Read filesystem statistics
	if info.MountPoint != "" {
		var stat syscall.Statfs_t
		if err := syscall.Statfs(info.MountPoint, &stat); err == nil {
			info.BlockSize = uint64(stat.Bsize)
			info.Size = stat.Blocks * uint64(stat.Bsize)
			info.Available = stat.Bavail * uint64(stat.Bsize)
			info.Used = info.Size - stat.Bfree*uint64(stat.Bsize)
			if info.Size > 0 {
				info.UsagePercent = (float64(info.Used) / float64(info.Size)) * 100
			}
		}
	}

	// Read UUID
	info.UUID, _ = r.getUUID(devicePath)

	// Read Label
	info.Label, _ = r.getLabel(devicePath)

	// Read device model and serial
	baseDevice := r.getBaseDevice(deviceName)
	info.Model, _ = r.getDeviceModel(baseDevice)
	info.Serial, _ = r.getDeviceSerial(baseDevice)

	// Read removable status
	info.Removable, _ = r.getDeviceRemovable(baseDevice)

	// Read rotational status
	info.Rotational, _ = r.getDeviceRotational(baseDevice)

	return info, nil
}

// ReadIOStats reads I/O statistics for the specified device
func (r *DiskReader) ReadIOStats(devicePath string) (*IOStats, error) {
	stats := &IOStats{}

	// Get base device name (e.g., sda from sda1)
	baseDevice := r.getBaseDevice(filepath.Base(devicePath))

	file, err := os.Open(r.procDiskstatsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open diskstats: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 14 {
			continue
		}

		// Check if this line is for our device
		// Format: major minor device reads_completed reads_merged sectors_read read_time_ms ...
		if fields[2] == baseDevice {
			stats.ReadsCompleted = parseUint64(fields[3])
			stats.ReadsMerged = parseUint64(fields[4])
			stats.SectorsRead = parseUint64(fields[5])
			stats.ReadTimeMs = parseUint64(fields[6])
			stats.WritesCompleted = parseUint64(fields[7])
			stats.WritesMerged = parseUint64(fields[8])
			stats.SectorsWritten = parseUint64(fields[9])
			stats.WriteTimeMs = parseUint64(fields[10])
			stats.IOInProgress = parseUint64(fields[11])
			stats.IOTimeMs = parseUint64(fields[12])
			stats.WeightedIOTimeMs = parseUint64(fields[13])
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read diskstats: %w", err)
	}

	return stats, nil
}

// ReadInodeInfo reads inode usage information for the specified mount point
func (r *DiskReader) ReadInodeInfo(mountPoint string) (*InodeInfo, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(mountPoint, &stat); err != nil {
		return nil, fmt.Errorf("failed to stat filesystem: %w", err)
	}

	info := &InodeInfo{
		Total: stat.Files,
		Free:  stat.Ffree,
		Used:  stat.Files - stat.Ffree,
	}

	if info.Total > 0 {
		info.UsagePercent = (float64(info.Used) / float64(info.Total)) * 100
	}

	return info, nil
}

// readMounts reads mount information from /etc/mtab or /proc/mounts
func (r *DiskReader) readMounts() ([]MountEntry, error) {
	// Try /etc/mtab first, fall back to /proc/mounts
	mountsFile := r.etcMtabPath
	if _, err := os.Stat(mountsFile); err != nil {
		mountsFile = r.procMountsPath
	}

	file, err := os.Open(mountsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open mounts file: %w", err)
	}
	defer file.Close()

	var mounts []MountEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 4 {
			mounts = append(mounts, MountEntry{
				Spec:    fields[0],
				File:    fields[1],
				Vfstype: fields[2],
				Mntops:  fields[3],
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read mounts: %w", err)
	}

	return mounts, nil
}

// getBaseDevice returns the base device name (e.g., sda from sda1, nvme0n1 from nvme0n1p2)
func (r *DiskReader) getBaseDevice(deviceName string) string {
	// Handle nvme devices (e.g., nvme0n1p2 -> nvme0n1)
	if strings.HasPrefix(deviceName, "nvme") {
		parts := strings.Split(deviceName, "p")
		if len(parts) >= 2 {
			return parts[0]
		}
		return deviceName
	}

	// Handle mmc devices (e.g., mmcblk0p1 -> mmcblk0)
	if strings.HasPrefix(deviceName, "mmcblk") {
		parts := strings.Split(deviceName, "p")
		if len(parts) >= 2 {
			return parts[0]
		}
		return deviceName
	}

	// Handle standard devices (e.g., sda1 -> sda)
	for i := len(deviceName) - 1; i > 0; i-- {
		if deviceName[i] >= '0' && deviceName[i] <= '9' {
			continue
		}
		return deviceName[:i+1]
	}
	return deviceName
}

// getDeviceModel reads the device model from sysfs
func (r *DiskReader) getDeviceModel(deviceName string) (string, error) {
	modelPath := filepath.Join(r.sysfsPath, deviceName, "device", "model")
	data, err := os.ReadFile(modelPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// getDeviceSerial reads the device serial from sysfs
func (r *DiskReader) getDeviceSerial(deviceName string) (string, error) {
	// Try different serial file locations
	serialPaths := []string{
		filepath.Join(r.sysfsPath, deviceName, "device", "serial"),
		filepath.Join(r.sysfsPath, deviceName, "device", "wwid"),
	}

	for _, path := range serialPaths {
		data, err := os.ReadFile(path)
		if err == nil {
			serial := strings.TrimSpace(string(data))
			if serial != "" {
				return serial, nil
			}
		}
	}
	return "", fmt.Errorf("serial not found")
}

// getDeviceRemovable reads whether the device is removable
func (r *DiskReader) getDeviceRemovable(deviceName string) (bool, error) {
	removablePath := filepath.Join(r.sysfsPath, deviceName, "removable")
	data, err := os.ReadFile(removablePath)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(data)) == "1", nil
}

// getDeviceRotational reads whether the device is rotational
func (r *DiskReader) getDeviceRotational(deviceName string) (bool, error) {
	rotationalPath := filepath.Join(r.sysfsPath, deviceName, "queue", "rotational")
	data, err := os.ReadFile(rotationalPath)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(data)) == "1", nil
}

// getUUID reads the filesystem UUID from /dev/disk/by-uuid
func (r *DiskReader) getUUID(devicePath string) (string, error) {
	entries, err := os.ReadDir(r.devDiskByUUIDPath)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		linkPath := filepath.Join(r.devDiskByUUIDPath, entry.Name())
		target, err := os.Readlink(linkPath)
		if err != nil {
			continue
		}
		// Resolve the symlink target to get the actual device path
		resolvedPath := filepath.Join("/dev/disk/by-uuid", target)
		if strings.Contains(resolvedPath, filepath.Base(devicePath)) {
			return entry.Name(), nil
		}
	}

	return "", fmt.Errorf("UUID not found for device %s", devicePath)
}

// getLabel reads the filesystem label from /dev/disk/by-label
func (r *DiskReader) getLabel(devicePath string) (string, error) {
	entries, err := os.ReadDir(r.devDiskByLabelPath)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		linkPath := filepath.Join(r.devDiskByLabelPath, entry.Name())
		target, err := os.Readlink(linkPath)
		if err != nil {
			continue
		}
		resolvedPath := filepath.Join("/dev/disk/by-label", target)
		if strings.Contains(resolvedPath, filepath.Base(devicePath)) {
			return entry.Name(), nil
		}
	}

	return "", fmt.Errorf("label not found for device %s", devicePath)
}

// parseUint64 safely parses a string to uint64
func parseUint64(s string) uint64 {
	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return val
}

// FormatBytes formats bytes to a human-readable string
func FormatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
		PB = TB * 1024
	)

	switch {
	case bytes >= PB:
		return fmt.Sprintf("%.2f PB", float64(bytes)/float64(PB))
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// FormatSectors formats sectors to a human-readable string (assuming 512-byte sectors)
func FormatSectors(sectors uint64) string {
	bytes := sectors * 512
	return FormatBytes(bytes)
}
