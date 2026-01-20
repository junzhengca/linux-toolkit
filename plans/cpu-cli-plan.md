# CPU Information Tool Implementation Plan

## Overview
Add a new CLI tool to display verbose CPU information. The tool will follow the same architectural pattern as the existing battery and disk tools.

## Architecture

```mermaid
graph TD
    A[main.go] --> B[cmd/root.go]
    B --> C[cmd/cpu.go]
    C --> D[pkg/cpu/cpu.go]
    C --> E[pkg/cpu/output.go]
    D --> F[/proc/cpuinfo]
    D --> G[/sys/devices/system/cpu/]
    D --> H[/proc/stat]
    D --> I[/sys/class/thermal/]
```

## Data Structures

### CPUInfo Struct
```go
type CPUInfo struct {
    ModelName        string   `json:"modelName"`         // CPU model name
    VendorID         string   `json:"vendorId"`          // Vendor ID (e.g., GenuineIntel)
    Architecture     string   `json:"architecture"`       // CPU architecture (e.g., x86_64)
    CPUMode          string   `json:"cpuMode"`           // 32-bit or 64-bit
    CPUFamily        int      `json:"cpuFamily"`         // CPU family number
    Model            int      `json:"model"`             // Model number
    Stepping         int      `json:"stepping"`          // Stepping ID
    PhysicalCores    int      `json:"physicalCores"`     // Number of physical cores
    LogicalCores     int      `json:"logicalCores"`      // Number of logical cores (threads)
    ThreadsPerCore   int      `json:"threadsPerCore"`    // Threads per core
    Sockets          int      `json:"sockets"`           // Number of CPU sockets
    MinFrequency     int64    `json:"minFrequency"`      // Minimum frequency in kHz
    MaxFrequency     int64    `json:"maxFrequency"`      // Maximum frequency in kHz
    CurrentFrequency int64    `json:"currentFrequency"`  // Current frequency in kHz
    Bogomips         float64  `json:"bogomips"`          // Bogomips value
    Flags            []string `json:"flags"`             // CPU flags/features
    CacheL1d         int      `json:"cacheL1d"`          // L1 data cache size in KB
    CacheL1i         int      `json:"cacheL1i"`          // L1 instruction cache size in KB
    CacheL2          int      `json:"cacheL2"`           // L2 cache size in KB
    CacheL3          int      `json:"cacheL3"`           // L3 cache size in KB
    CPUTemperature   float64  `json:"cpuTemperature"`    // CPU temperature in Celsius
    LoadAvg1         float64  `json:"loadAvg1"`          // 1-minute load average
    LoadAvg5         float64  `json:"loadAvg5"`          // 5-minute load average
    LoadAvg15        float64  `json:"loadAvg15"`         // 15-minute load average
    TotalProcesses   int      `json:"totalProcesses"`    // Total number of processes
    RunningProcesses int      `json:"runningProcesses"`  // Number of running processes
    Cores            []CoreInfo `json:"cores"`          // Per-core information
}
```

### CoreInfo Struct
```go
type CoreInfo struct {
    CoreID       int     `json:"coreId"`       // Core ID
    PhysicalID   int     `json:"physicalId"`   // Physical CPU ID (socket)
    ProcessorID  int     `json:"processorId"`  // Processor ID (logical thread)
    Frequency    int64   `json:"frequency"`    // Current frequency in kHz
    UsagePercent float64 `json:"usagePercent"` // Current usage percentage
}
```

## Implementation Steps

### Step 1: Create pkg/cpu/cpu.go
Create the core CPU reading logic with the following components:

1. **CPUReader struct** - Main reader for CPU information
   - `procCpuinfoPath: /proc/cpuinfo`
   - `procStatPath: /proc/stat`
   - `sysCpuPath: /sys/devices/system/cpu/`
   - `sysThermalPath: /sys/class/thermal/`

2. **Methods to implement:**
   - `ReadCPU() (*CPUInfo, error)` - Read all CPU information
   - `ReadCpuinfo() (*CPUInfo, error)` - Parse /proc/cpuinfo
   - `ReadProcStat() (*CPUInfo, error)` - Parse /proc/stat for load and usage
   - `ReadCpuFreq() (*CPUInfo, error)` - Read CPU frequencies from sysfs
   - `ReadCpuCache() (*CPUInfo, error)` - Read cache sizes from sysfs
   - `ReadCpuTemp() (float64, error)` - Read CPU temperature from thermal zone
   - `ReadLoadAvg() (float64, float64, float64, error)` - Read load averages
   - `ReadCoreInfo() ([]CoreInfo, error)` - Read per-core information

3. **Helper methods:**
   - `parseCpuinfoLine(line string) (key, value string)`
   - `parseFrequency(freqStr string) (int64, error)`
   - `formatFrequency(kHz int64) string`
   - `getThermalZone() (string, error)` - Find CPU thermal zone
   - `calculateCoreUsage(prevStat, currStat map[string]uint64) float64`

### Step 2: Create pkg/cpu/output.go
Create output formatting logic:

1. **OutputFormat type** - Table and JSON formats (same as battery/disk)

2. **Functions to implement:**
   - `Print(cpu *CPUInfo, format OutputFormat) error`
   - `PrintTable(cpu *CPUInfo)` - Human-readable table with colors and emojis
   - `PrintJSON(cpu *CPUInfo) error` - JSON output

3. **Helper functions:**
   - `getUsageColor(percent float64) *color.Color` - Color based on usage
   - `getTempColor(temp float64) *color.Color` - Color based on temperature
   - `getArchIcon(arch string) string` - Emoji based on architecture
   - `printField()` - Consistent field formatting
   - `formatMHz(kHz int64) string` - Format kHz to MHz/GHz

### Step 3: Create cmd/cpu.go
Create the CLI command:

1. **Command flags:**
   - `--format, -f` - Output format (table|json)
   - `--show-cores` - Show per-core usage details
   - `--show-flags` - Show all CPU flags/features
   - `--show-temp` - Include temperature information

2. **Command logic:**
   - Read all CPU information
   - Optionally filter output based on flags
   - Display in requested format

### Step 4: Update cmd/root.go
- Add CPU command to rootCmd using `rootCmd.AddCommand(cpuCmd)`

### Step 5: Update README.md
- Add documentation for the new CPU command
- Include usage examples
- Show table and JSON output formats
- Document all command flags

## Table Output Format Example

```
  🖥️  CPU Information

  📋 General Info
  ────────────────────────────────────────
  Model:             AMD Ryzen 7 5800X 8-Core Processor
  Vendor:            AuthenticAMD
  Architecture:      x86_64
  Mode:              64-bit
  Family:            25
  Model:             8
  Stepping:          0
  Bogomips:          6786.72

  🔧 Core Configuration
  ────────────────────────────────────────
  Physical Cores:    8
  Logical Cores:     16
  Threads per Core:  2
  Sockets:           1

  ⚡ Frequency
  ────────────────────────────────────────
  Min Frequency:     2200 MHz
  Max Frequency:     4800 MHz
  Current Frequency: 3800 MHz

  💾 Cache Sizes
  ────────────────────────────────────────
  L1 Data Cache:     32 KB
  L1 Instruction:    32 KB
  L2 Cache:          512 KB
  L3 Cache:          32768 KB (32 MB)

  📊 System Load
  ────────────────────────────────────────
  Load Average (1m):  2.45
  Load Average (5m):  2.12
  Load Average (15m): 1.89
  Total Processes:    324
  Running Processes:  3

  🌡️  Temperature
  ────────────────────────────────────────
  CPU Temperature:    52°C

  🔌 Core Usage
  ────────────────────────────────────────
  Core 0 (CPU0):      15.2% @ 3800 MHz
  Core 1 (CPU1):      12.8% @ 3800 MHz
  Core 2 (CPU2):      18.5% @ 3800 MHz
  Core 3 (CPU3):      14.1% @ 3800 MHz
  Core 4 (CPU4):      22.3% @ 3800 MHz
  Core 5 (CPU5):      16.7% @ 3800 MHz
  Core 6 (CPU6):      19.4% @ 3800 MHz
  Core 7 (CPU7):      13.9% @ 3800 MHz

  🏷️  CPU Flags (showing first 20 of 180)
  ────────────────────────────────────────
  fpu vme de pse tsc msr pae mce cx8 apic
  sep mtrr pge mca cmov pat pse36 clflush mmx
  fxsr sse sse2 ss ht syscall nx pdpe1gb rdtscp
  ...
```

## JSON Output Format Example

```json
{
  "modelName": "AMD Ryzen 7 5800X 8-Core Processor",
  "vendorId": "AuthenticAMD",
  "architecture": "x86_64",
  "cpuMode": "64-bit",
  "cpuFamily": 25,
  "model": 8,
  "stepping": 0,
  "physicalCores": 8,
  "logicalCores": 16,
  "threadsPerCore": 2,
  "sockets": 1,
  "minFrequency": 2200000,
  "maxFrequency": 4800000,
  "currentFrequency": 3800000,
  "bogomips": 6786.72,
  "flags": [
    "fpu", "vme", "de", "pse", "tsc", "msr", "pae", "mce",
    "cx8", "apic", "sep", "mtrr", "pge", "mca", "cmov", "pat",
    ...
  ],
  "cacheL1d": 32,
  "cacheL1i": 32,
  "cacheL2": 512,
  "cacheL3": 32768,
  "cpuTemperature": 52.0,
  "loadAvg1": 2.45,
  "loadAvg5": 2.12,
  "loadAvg15": 1.89,
  "totalProcesses": 324,
  "runningProcesses": 3,
  "cores": [
    {
      "coreId": 0,
      "physicalId": 0,
      "processorId": 0,
      "frequency": 3800000,
      "usagePercent": 15.2
    },
    ...
  ]
}
```

## Color Schemes

### Usage Percentage
| Usage % | Status | Color |
|---------|--------|-------|
| 0-25% | Idle | 🟢 Green |
| 26-50% | Light | 🟢 HiGreen |
| 51-75% | Moderate | 🟡 Yellow |
| 76-90% | High | 🟠 HiYellow |
| 91-100% | Critical | 🔴 Red |

### Temperature
| Temperature | Status | Color |
|-------------|--------|-------|
| 0-45°C | Cool | 🟢 Green |
| 46-60°C | Normal | 🟢 HiGreen |
| 61-75°C | Warm | 🟡 Yellow |
| 76-85°C | Hot | 🟠 HiYellow |
| 86°C+ | Critical | 🔴 Red |

### Architecture Icons
| Architecture | Emoji |
|--------------|-------|
| x86_64 | 🖥️ |
| arm64 | 📱 |
| aarch64 | 📱 |
| riscv64 | 🔬 |

## Dependencies
No new dependencies required. Uses only standard library and existing:
- `github.com/spf13/cobra` (already in go.mod)
- `github.com/fatih/color` (already in go.mod)

## Files to Create/Modify

### New Files:
1. `pkg/cpu/cpu.go` - Core CPU reading logic
2. `pkg/cpu/output.go` - Output formatting
3. `cmd/cpu.go` - CPU command

### Files to Modify:
1. `cmd/root.go` - Add CPU command to root
2. `README.md` - Update documentation

## Linux System Information Sources

### /proc/cpuinfo
Contains CPU model, vendor, family, model, stepping, flags, bogomips, cache sizes

### /proc/stat
Contains CPU statistics for load calculation:
- First line: `cpu user nice system idle iowait irq softirq steal guest guest_nice`
- Per-core lines: `cpu0 user nice system idle ...`

### /sys/devices/system/cpu/
- `cpu0/cpufreq/scaling_cur_freq` - Current frequency
- `cpu0/cpufreq/scaling_min_freq` - Minimum frequency
- `cpu0/cpufreq/scaling_max_freq` - Maximum frequency
- `cpu0/cache/index0/size` - L1 data cache size
- `cpu0/cache/index1/size` - L1 instruction cache size
- `cpu0/cache/index2/size` - L2 cache size
- `cpu0/cache/index3/size` - L3 cache size
- `cpu0/topology/core_id` - Core ID
- `cpu0/topology/physical_package_id` - Physical socket ID

### /sys/class/thermal/
- `thermal_zone0/temp` - Temperature in millidegrees Celsius
- Need to identify which zone corresponds to CPU

### /proc/loadavg
Contains load averages: `1min 5min 15min running/total last_pid`

## Usage Examples

```bash
# Display CPU information in table format (default)
linux-toolkit cpu

# Display CPU information in JSON format
linux-toolkit cpu --format json

# Show per-core usage details
linux-toolkit cpu --show-cores

# Show all CPU flags/features
linux-toolkit cpu --show-flags

# Include temperature information
linux-toolkit cpu --show-temp

# Show all verbose information
linux-toolkit cpu --show-cores --show-flags --show-temp

# Show help
linux-toolkit cpu --help
```

## Testing Considerations
- Test on different CPU architectures (x86_64, ARM64)
- Test with single-core and multi-core systems
- Test with single-socket and multi-socket systems
- Verify frequency scaling on systems with dynamic frequency
- Verify temperature reading on systems with thermal zones
- Test load calculation accuracy
- Verify cache size parsing
- Test with various CPU vendors (Intel, AMD, ARM)
