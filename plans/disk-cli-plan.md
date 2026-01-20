# Disk Status Tool Implementation Plan

## Overview
Add a new CLI tool to display verbose disk status information, replacing the battery tool functionality. The tool will follow the same architectural pattern as the existing battery tool.

## Architecture

```mermaid
graph TD
    A[main.go] --> B[cmd/root.go]
    B --> C[cmd/disk.go]
    C --> D[pkg/disk/disk.go]
    C --> E[pkg/disk/output.go]
    D --> F[/sys/block/]
    D --> G[/etc/mtab]
    D --> H[/dev/disk/by-uuid/]
    D --> I[/proc/diskstats]
```

## Data Structures

### DiskInfo Struct
```go
type DiskInfo struct {
    Name           string  `json:"name"`            // Device name (e.g., sda1, nvme0n1p1)
    Path           string  `json:"path"`            // Full device path (e.g., /dev/sda1)
    Size           uint64  `json:"size"`            // Total size in bytes
    Used           uint64  `json:"used"`            // Used space in bytes
    Available      uint64  `json:"available"`       // Available space in bytes
    UsagePercent   float64 `json:"usagePercent"`     // Usage percentage
    FilesystemType string  `json:"filesystemType"`  // e.g., ext4, xfs, btrfs
    MountPoint     string  `json:"mountPoint"`      // Mount point
    UUID           string  `json:"uuid"`             // Filesystem UUID
    Label          string  `json:"label"`            // Filesystem label
    BlockSize      uint64  `json:"blockSize"`       // Block size in bytes
    Model          string  `json:"model"`            // Device model (from /sys/block/)
    Serial         string  `json:"serial"`           // Device serial number
    ReadOnly       bool    `json:"readOnly"`        // Whether mounted read-only
    Removable      bool    `json:"removable"`       // Whether device is removable
    Rotational     bool    `json:"rotational"`      // Whether device is rotational (HDD)
}
```

### I/O Stats Struct
```go
type IOStats struct {
    ReadsCompleted    uint64 `json:"readsCompleted"`
    ReadsMerged       uint64 `json:"readsMerged"`
    SectorsRead       uint64 `json:"sectorsRead"`
    ReadTimeMs        uint64 `json:"readTimeMs"`
    WritesCompleted   uint64 `json:"writesCompleted"`
    WritesMerged      uint64 `json:"writesMerged"`
    SectorsWritten    uint64 `json:"sectorsWritten"`
    WriteTimeMs       uint64 `json:"writeTimeMs"`
    IOInProgress      uint64 `json:"ioInProgress"`
    IOTimeMs          uint64 `json:"ioTimeMs"`
    WeightedIOTimeMs  uint64 `json:"weightedIOTimeMs"`
}
```

### InodeInfo Struct
```go
type InodeInfo struct {
    Total      uint64  `json:"total"`
    Used       uint64  `json:"used"`
    Free       uint64  `json:"free"`
    UsagePercent float64 `json:"usagePercent"`
}
```

## Implementation Steps

### Step 1: Create pkg/disk/disk.go
Create the core disk reading logic with the following components:

1. **DiskReader struct** - Main reader for disk information
   - `sysfsPath: /sys/block/`
   - `procDiskstatsPath: /proc/diskstats`
   - `etcMtabPath: /etc/mtab`

2. **Methods to implement:**
   - `ListDisks() ([]string, error)` - List all available disk devices
   - `ReadDisk(deviceName string) (*DiskInfo, error)` - Read disk information
   - `ReadIOStats(deviceName string) (*IOStats, error)` - Read I/O statistics
   - `ReadInodeInfo(mountPoint string) (*InodeInfo, error)` - Read inode usage
   - `GetDeviceModel(deviceName string) (string, error)` - Get device model from sysfs
   - `GetDeviceSerial(deviceName string) (string, error)` - Get device serial
   - `GetUUID(devicePath string) (string, error)` - Get filesystem UUID
   - `GetLabel(devicePath string) (string, error)` - Get filesystem label
   - `GetBlockSize(devicePath string) (uint64, error)` - Get block size

3. **Helper methods:**
   - `readMounts() ([]MountEntry, error)` - Parse /etc/mtab or /proc/mounts
   - `readProcDiskstats() (map[string]*IOStats, error)` - Parse /proc/diskstats
   - `parseSize(sizeStr string) (uint64, error)` - Parse size strings
   - `formatBytes(bytes uint64) string` - Format bytes to human-readable

### Step 2: Create pkg/disk/output.go
Create output formatting logic:

1. **OutputFormat type** - Table and JSON formats (same as battery)

2. **Functions to implement:**
   - `Print(disk *DiskInfo, ioStats *IOStats, inodeInfo *InodeInfo, format OutputFormat) error`
   - `PrintTable(disk *DiskInfo, ioStats *IOStats, inodeInfo *InodeInfo)` - Human-readable table with colors and emojis
   - `PrintJSON(disk *DiskInfo, ioStats *IOStats, inodeInfo *InodeInfo) error` - JSON output

3. **Helper functions:**
   - `getUsageColor(percent float64) *color.Color` - Color based on usage
   - `getFileSystemIcon(fsType string) string` - Emoji based on filesystem type
   - `printField()` - Consistent field formatting

### Step 3: Create cmd/disk.go
Create the CLI command:

1. **Command flags:**
   - `--device, -d` - Specific device name (e.g., sda1, nvme0n1p2)
   - `--mount, -m` - Filter by mount point
   - `--all, -a` - Show all disks (including unmounted)
   - `--format, -f` - Output format (table|json)
   - `--io-stats` - Include I/O statistics
   - `--inode-stats` - Include inode statistics

2. **Command logic:**
   - Auto-detect and display first mounted disk if no device specified
   - Support filtering by mount point
   - Support showing all available disks
   - Include optional I/O and inode stats

### Step 4: Update cmd/root.go
- Add disk command to rootCmd using `rootCmd.AddCommand(diskCmd)`

### Step 5: Update README.md
- Add documentation for the new disk command
- Include usage examples
- Show table and JSON output formats
- Document all command flags

## Table Output Format Example

```
  💾 Disk Information

  📱 Device Info
  ────────────────────────────────────────
  Device:            sda1
  Path:              /dev/sda1
  Model:             Samsung SSD 980
  Serial:            S4EWNX0M123456
  Filesystem:        ext4
  UUID:              a1b2-c3d4-e5f6
  Label:             root
  Mount Point:       /
  Block Size:        4096 bytes
  Type:              NVMe SSD (non-rotational)

  📊 Space Usage
  ────────────────────────────────────────
  Total Size:        500.11 GB
  Used Space:        234.56 GB
  Available Space:   240.23 GB
  Usage:             49.4%
  Inodes:            32,768,000 total / 1,234,567 used (3.8%)

  ⚡ I/O Statistics
  ────────────────────────────────────────
  Reads:             1,234,567 (2.34 TB)
  Writes:            987,654 (1.87 TB)
  Read Time:         12,345 ms
  Write Time:        9,876 ms
  I/O Time:          22,221 ms
```

## JSON Output Format Example

```json
{
  "name": "sda1",
  "path": "/dev/sda1",
  "size": 536870912000,
  "used": 251858739200,
  "available": 257931089920,
  "usagePercent": 49.4,
  "filesystemType": "ext4",
  "mountPoint": "/",
  "uuid": "a1b2-c3d4-e5f6",
  "label": "root",
  "blockSize": 4096,
  "model": "Samsung SSD 980",
  "serial": "S4EWNX0M123456",
  "readOnly": false,
  "removable": false,
  "rotational": false,
  "ioStats": {
    "readsCompleted": 1234567,
    "readsMerged": 12345,
    "sectorsRead": 4882812500,
    "readTimeMs": 12345,
    "writesCompleted": 987654,
    "writesMerged": 9876,
    "sectorsWritten": 3906250000,
    "writeTimeMs": 9876,
    "ioInProgress": 0,
    "ioTimeMs": 22221,
    "weightedIOTimeMs": 22221
  },
  "inodeInfo": {
    "total": 32768000,
    "used": 1234567,
    "free": 31533433,
    "usagePercent": 3.8
  }
}
```

## Color Scheme

| Usage % | Status | Color |
|---------|--------|-------|
| 0-50% | Healthy | 🟢 Green |
| 51-75% | Moderate | 🟡 Yellow |
| 76-90% | High | 🟠 Orange |
| 91-95% | Critical | 🔴 Red |
| 96-100% | Full | 🔴 Red (bold) |

## Filesystem Icons

| Type | Emoji |
|------|-------|
| ext4 | 📁 |
| xfs | 📂 |
| btrfs | 🗄️ |
| ntfs | 💿 |
| vfat | 💾 |
| swap | 🔄 |
| tmpfs | ⚡ |
| nfs | 🌐 |

## Dependencies
No new dependencies required. Uses only standard library and existing:
- `github.com/spf13/cobra` (already in go.mod)
- `github.com/fatih/color` (already in go.mod)

## Files to Create/Modify

### New Files:
1. `pkg/disk/disk.go` - Core disk reading logic
2. `pkg/disk/output.go` - Output formatting

### Files to Modify:
1. `cmd/disk.go` - Create new disk command
2. `cmd/root.go` - Add disk command to root
3. `README.md` - Update documentation

### Optional: Remove battery tool
- `cmd/battery.go` - Can be removed if replacing battery
- `pkg/battery/battery.go` - Can be removed if replacing battery
- `pkg/battery/output.go` - Can be removed if replacing battery

## Testing Considerations
- Test with various filesystem types (ext4, xfs, btrfs, ntfs)
- Test with mounted and unmounted devices
- Test with NVMe and SATA devices
- Test with removable devices (USB drives)
- Verify I/O stats parsing from /proc/diskstats
- Verify inode stats for different filesystems
