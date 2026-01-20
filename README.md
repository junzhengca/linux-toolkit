# linux-toolkit

A collection of useful Linux tools implemented in Go.

## Installation

### From Source

```bash
git clone https://github.com/jun/linux-toolkit.git
cd linux-toolkit
go build -o linux-toolkit
sudo mv linux-toolkit /usr/local/bin/
```

## Usage

### Disk Status

Display verbose disk information including device name, size, used/available space, usage percentage, filesystem type, mount point, UUID, label, block size, I/O statistics, inode usage, and device model/serial number.

```bash
# Display disk stats in table format (default) with colors and emojis
linux-toolkit disk

# Display disk stats in JSON format
linux-toolkit disk --format json

# Display specific disk device
linux-toolkit disk --device sda1

# Display disk by mount point
linux-toolkit disk --mount /

# Show all mounted disks
linux-toolkit disk --all

# Include I/O statistics
linux-toolkit disk --io-stats

# Include inode statistics
linux-toolkit disk --inode-stats

# Combine options
linux-toolkit disk --device nvme0n1p1 --format json --io-stats --inode-stats
```

#### Table Format Output (with colors and emojis)

```
  💾 Disk Information

  📱 Device Info
  ────────────────────────────────────────
  Device:            sda1
  Path:              /dev/sda1
  Model:             Samsung SSD 980
  Serial:            S4EWNX0M123456
  Filesystem:        📁 ext4
  UUID:              a1b2-c3d4-e5f6
  Label:             root
  Mount Point:       /
  Block Size:        4096 bytes
  Type:              SSD (non-rotational)

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

#### JSON Format Output

```json
{
  "disk": {
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
    "rotational": false
  },
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

#### All Disks Output

```
  💾 All Disks

  ─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
  Device          Size         Used         Avail        Usage%          Mount Point           Filesystem             Type
  ─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
  sda1            500.11 GB    234.56 GB    240.23 GB    49.4%           /                     📁 ext4                SSD
  sda2            100.00 GB    45.67 GB     50.12 GB     47.7%           /home                 📁 ext4                SSD
  sda3            2.00 GB      0.00 GB      2.00 GB      0.0%            [SWAP]                🔄 swap                SSD
  ─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
```

#### Usage Status Colors

| Usage % | Status | Color |
|---------|--------|-------|
| 0-50% | Healthy | 🟢 Green |
| 51-75% | Moderate | 🟡 Yellow |
| 76-90% | High | 🟠 Orange |
| 91-95% | Critical | 🔴 Red |
| 96-100% | Full | 🔴 Red (bold) |

#### Filesystem Icons

| Type | Emoji |
|------|-------|
| ext4, ext3, ext2 | 📁 |
| xfs | 📂 |
| btrfs | 🗄️ |
| ntfs | 💿 |
| vfat, fat32, exfat | 💾 |
| swap | 🔄 |
| tmpfs | ⚡ |
| nfs, cifs, smb | 🌐 |
| zfs | 🌊 |

### CPU Information

Display verbose CPU information including model name, architecture, vendor ID, CPU family, stepping, cache sizes, core counts, clock speeds, flags/features, temperature, load averages, and per-core usage.

```bash
# Display CPU information in table format (default) with colors and emojis
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

# Combine options
linux-toolkit cpu --format json --show-cores --show-temp
```

#### Table Format Output (with colors and emojis)

```
  🖥️  CPU Information

  📋 General Info
  ────────────────────────────────────────
  Model:             AMD Ryzen 7 5800X 8-Core Processor
  Vendor:            AuthenticAMD
  Architecture:      🖥️ x86_64
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

#### JSON Format Output

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
  ],
  "showCores": false,
  "showFlags": false,
  "showTemp": true
}
```

#### Usage Status Colors

| Usage % | Status | Color |
|---------|--------|-------|
| 0-25% | Idle | 🟢 Green |
| 26-50% | Light | 🟢 HiGreen |
| 51-75% | Moderate | 🟡 Yellow |
| 76-90% | High | 🟠 HiYellow |
| 91-100% | Critical | 🔴 Red |

#### Temperature Colors

| Temperature | Status | Color |
|-------------|--------|-------|
| 0-45°C | Cool | 🟢 Green |
| 46-60°C | Normal | 🟢 HiGreen |
| 61-75°C | Warm | 🟡 Yellow |
| 76-85°C | Hot | 🟠 HiYellow |
| 86°C+ | Critical | 🔴 Red |

#### Architecture Icons

| Architecture | Emoji |
|--------------|-------|
| x86_64, amd64 | 🖥️ |
| arm64, aarch64 | 📱 |
| riscv64 | 🔬 |

### GPU Information

Display verbose GPU information including device name, vendor, driver, memory, clocks, temperature, power usage, utilization, and connector information.

```bash
# Display GPU information in table format (default) with colors and emojis
linux-toolkit gpu

# Display GPU information in JSON format
linux-toolkit gpu --format json

# Display information for specific GPU card
linux-toolkit gpu --card card1

# Show all GPUs
linux-toolkit gpu --all

# Show connector information
linux-toolkit gpu --show-connectors

# Show supported display modes
linux-toolkit gpu --show-modes

# Show all verbose information
linux-toolkit gpu --all --show-connectors --show-modes

# Combine options
linux-toolkit gpu --card card0 --format json --show-connectors
```

#### Table Format Output (with colors and emojis)

```
  🎮 GPU Information

  📋 Device Info
  ────────────────────────────────────────
  Card Name:         card0
  Device:            /dev/dri/card0
  Driver:            amdgpu
  Vendor:            AMD
  Bus ID:            0000:01:00.0
  PCIe Gen:          4.0 (max: 4.0)
  Bus Width:         x16
  Type:              Discrete

  🔧 Hardware
  ────────────────────────────────────────
  Vendor ID:         1002
  Device ID:         73df
  Class:             0300
  Revision:          c4
  VBIOS Version:     113-D05201-011

  💾 Memory
  ────────────────────────────────────────
  VRAM Total:        16.00 GB
  VRAM Used:         4.23 GB
  VRAM Free:         11.77 GB
  GART Size:         512.00 MB

  ⚡ Clocks
  ────────────────────────────────────────
  Core Clock:        1.83 GHz (max: 2.25 GHz)
  Memory Clock:      16.00 GHz (max: 16.00 GHz)

  🌡️  Temperature
  ────────────────────────────────────────
  Temperature:        62°C (crit: 95°C)
  Fan Speed:          1200 RPM

  🔌 Power
  ────────────────────────────────────────
  Power Usage:        245 W
  Power Limit:        300 W

  📊 Utilization
  ────────────────────────────────────────
  GPU Usage:          78%
  Memory Usage:       26%

  🔌 Connectors
  ────────────────────────────────────────
  HDMI-A-1:          disconnected
  DisplayPort-1:      connected
  DisplayPort-2:      disconnected

  🖥️  Display Modes
  ────────────────────────────────────────
  1920x1080
  2560x1440
  3840x2160
  ... and 7 more modes
```

#### JSON Format Output

```json
{
  "cardName": "card0",
  "deviceName": "/dev/dri/card0",
  "driver": "amdgpu",
  "vendor": "AMD",
  "vendorId": "1002",
  "deviceId": "73df",
  "subsystemVendor": "1043",
  "subsystemDevice": "0482",
  "class": "0300",
  "revision": "c4",
  "enabled": true,
  "status": "active",
  "connectors": [
    "HDMI-A-1",
    "DisplayPort-1",
    "DisplayPort-2"
  ],
  "enabledConnectors": [
    "DisplayPort-1"
  ],
  "vramSize": 17179869184,
  "vramUsed": 4540267520,
  "vramFree": 12639601664,
  "gartSize": 536870912,
  "coreClock": 1825000,
  "memoryClock": 16000000,
  "maxCoreClock": 2250000,
  "maxMemoryClock": 16000000,
  "powerUsage": 245000,
  "powerLimit": 300000,
  "temperature": 62.0,
  "temperatureCrit": 95.0,
  "fanSpeed": 1200,
  "fanSpeedPercent": 45.0,
  "gpuUtilPercent": 78.0,
  "memoryUtilPercent": 26.0,
  "busId": "0000:01:00.0",
  "busWidth": "x16",
  "pcieGen": "4.0",
  "maxPcieGen": "4.0",
  "vbiosVersion": "113-D05201-011",
  "firmwareVersion": "",
  "devicePath": "/sys/class/drm/card0",
  "sysfsPath": "/sys/class/drm/card0/device",
  "isPrimary": true,
  "gpuType": "discrete",
  "computeUnits": 72,
  "cudaCores": 0,
  "shaders": 4608,
  "textureUnits": 288,
  "rops": 128
}
```

#### All GPUs Output

```
  🎮 All GPUs

  card0
  ────────────────────────────────────────
  🔴 AMD | amdgpu | 0000:01:00.0
  Memory: 16.00 GB | Temp: 62°C | Usage: 78%
  Connectors: HDMI-A-1, DisplayPort-1, DisplayPort-2

  card1
  ────────────────────────────────────────
  🟢 NVIDIA | nvidia | 0000:02:00.0
  Memory: 8.00 GB | Temp: 58°C | Usage: 45%
  Connectors: HDMI-A-1, DisplayPort-1
```

#### Vendor Colors

| Vendor | Color |
|--------|-------|
| NVIDIA | 🟢 Green |
| AMD | 🔴 Red |
| Intel | 🔵 Blue |
| Other | ⚪ White |

#### Usage Status Colors

| Usage % | Status | Color |
|---------|--------|-------|
| 0-25% | Idle | 🟢 Green |
| 26-50% | Light | 🟢 HiGreen |
| 51-75% | Moderate | 🟡 Yellow |
| 76-90% | High | 🟠 HiYellow |
| 91-100% | Critical | 🔴 Red |

#### Temperature Colors

| Temperature | Status | Color |
|-------------|--------|-------|
| 0-50°C | Cool | 🟢 Green |
| 51-65°C | Normal | 🟢 HiGreen |
| 66-80°C | Warm | 🟡 Yellow |
| 81-90°C | Hot | 🟠 HiYellow |
| 91°C+ | Critical | 🔴 Red |

### Services Information

Display services running on the system, including daemons and web services that bind to ports. Shows service name, status, PID, memory usage, CPU usage, start time, uptime, user, command, and listening ports.

```bash
# Display all running services in table format (default) with colors and emojis
linux-toolkit services

# Display all services including stopped ones
linux-toolkit services --show-all

# Display services in JSON format
linux-toolkit services --format json

# Show services filtered by status
linux-toolkit services --status running
linux-toolkit services --status failed

# Show service by name (detailed view)
linux-toolkit services --name nginx

# Show services filtered by user
linux-toolkit services --user nginx

# Show services listening on specific port
linux-toolkit services --port 80

# Show detailed port information
linux-toolkit services --show-ports

# Combine options
linux-toolkit services --status running --show-ports --format json
```

#### Table Format Output (with colors and emojis)

```
  🛠️  Services Information

  📋 Summary
  ────────────────────────────────────────
  Total Services:    15
  Running:           12
  Stopped:           2
  Failed:            1
  Listening Ports:   8

  🔄 Running Services
  ─────────────────────────────────────────────────────────────────────────────────────────────────────
  NAME              STATUS    PID    MEMORY    CPU    START TIME      USER      PORTS
  ─────────────────────────────────────────────────────────────────────────────────────────────────────
  🟢 sshd           running   1234   12.3 MB   0.1%   Jan 15 09:30    root      🟢:22/tcp
  🟢 nginx          running   2345   45.6 MB   0.5%   Jan 15 09:31    nginx     🟢:80/tcp, 🟢:443/tcp
  🟢 docker         running   3456   234.5 MB  1.2%   Jan 15 09:32    root      🟢:2375/tcp
  🟢 systemd-journald running  4567   8.2 MB    0.0%   Jan 15 09:30    root      -
  🟢 cron           running   5678   2.1 MB    0.0%   Jan 15 09:30    root      -
  ─────────────────────────────────────────────────────────────────────────────────────────────────────

  🔴 Failed Services
  ─────────────────────────────────────────────────────────────────────────────────────────────────────
  NAME              STATUS    PID    MEMORY    CPU    START TIME      USER      PORTS
  ─────────────────────────────────────────────────────────────────────────────────────────────────────
  🔴 my-service     failed    -      -         -      -               root      -
  ─────────────────────────────────────────────────────────────────────────────────────────────────────
```

#### Detailed Service Output

```
  🛠️  Service Details: nginx

  📋 General Info
  ────────────────────────────────────────
  Name:              nginx
  Description:       A high performance web server and a reverse proxy server
  Status:            🟢 running
  Loaded:            loaded
  Active:            active
  Sub State:         running
  Type:              systemd

  📊 Process Info
  ────────────────────────────────────────
  PID:               2345
  User:              nginx
  Memory:            45.6 MB
  CPU:               0.5%
  Start Time:        Jan 15 09:31:15
  Uptime:            5 days, 2 hours, 15 minutes
  Command:           nginx: master process /usr/sbin/nginx -g daemon off;

  🌐 Listening Ports
  ────────────────────────────────────────
  🟢 tcp  0.0.0.0:80    LISTEN   nginx (2345)
  🟢 tcp  0.0.0.0:443   LISTEN   nginx (2345)
  🟢 tcp6 [::]:80       LISTEN   nginx (2345)
  🟢 tcp6 [::]:443      LISTEN   nginx (2345)
```

#### JSON Format Output

```json
{
  "totalServices": 15,
  "running": 12,
  "stopped": 2,
  "failed": 1,
  "totalPorts": 8,
  "services": [
    {
      "name": "sshd",
      "description": "OpenSSH server daemon",
      "status": "running",
      "loaded": "loaded",
      "active": "active",
      "subState": "running",
      "pid": 1234,
      "memoryMB": 12.3,
      "cpuPercent": 0.1,
      "startTime": "Jan 15 09:30:00",
      "uptime": "5 days, 2 hours, 30 minutes",
      "user": "root",
      "command": "sshd: /usr/sbin/sshd -D [listener] 0 of 10-100 startups",
      "listeningPorts": [
        {
          "protocol": "tcp",
          "localAddr": "0.0.0.0:22",
          "port": 22,
          "state": "LISTEN",
          "processName": "sshd",
          "pid": 1234
        }
      ],
      "type": "systemd"
    },
    {
      "name": "nginx",
      "description": "A high performance web server and a reverse proxy server",
      "status": "running",
      "loaded": "loaded",
      "active": "active",
      "subState": "running",
      "pid": 2345,
      "memoryMB": 45.6,
      "cpuPercent": 0.5,
      "startTime": "Jan 15 09:31:15",
      "uptime": "5 days, 2 hours, 15 minutes",
      "user": "nginx",
      "command": "nginx: master process /usr/sbin/nginx -g daemon off;",
      "listeningPorts": [
        {
          "protocol": "tcp",
          "localAddr": "0.0.0.0:80",
          "port": 80,
          "state": "LISTEN",
          "processName": "nginx",
          "pid": 2345
        },
        {
          "protocol": "tcp",
          "localAddr": "0.0.0.0:443",
          "port": 443,
          "state": "LISTEN",
          "processName": "nginx",
          "pid": 2345
        }
      ],
      "type": "systemd"
    }
  ]
}
```

#### Status Colors

| Status | Icon | Color |
|--------|-------|-------|
| running, active | 🟢 | Green |
| stopped, inactive | ⏸️ | Yellow |
| failed, dead | 🔴 | Red |
| unknown | ❓ | White |

#### Protocol Icons

| Protocol | Emoji |
|----------|-------|
| tcp, tcp6 | 🟢 |
| udp, udp6 | 🔵 |

### Command Options

#### Disk Command Options

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--device` | `-d` | Disk device name (e.g., sda1, nvme0n1p1) | auto-detect first mounted |
| `--mount` | `-m` | Filter by mount point | - |
| `--all` | `-a` | Show all mounted disks | false |
| `--format` | `-f` | Output format (table\|json) | table |
| `--io-stats` | - | Include I/O statistics | false |
| `--inode-stats` | - | Include inode statistics | false |
| `--help` | `-h` | Help for disk command | - |

#### CPU Command Options

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--format` | `-f` | Output format (table\|json) | table |
| `--show-cores` | - | Show per-core usage details | false |
| `--show-flags` | - | Show all CPU flags/features | false |
| `--show-temp` | - | Include temperature information | false |
| `--help` | `-h` | Help for cpu command | - |

#### GPU Command Options

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--card` | `-c` | GPU card name (e.g., card0, card1) | auto-detect first GPU |
| `--all` | `-a` | Show all GPUs | false |
| `--format` | `-f` | Output format (table\|json) | table |
| `--show-connectors` | - | Show connector information | false |
| `--show-modes` | - | Show supported display modes | false |
| `--help` | `-h` | Help for gpu command | - |

#### Services Command Options

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--name` | `-n` | Filter by service name (supports partial match) | all |
| `--status` | `-s` | Filter by status (running\|stopped\|failed) | all |
| `--user` | `-u` | Filter by user | all |
| `--port` | `-p` | Filter by port number | all |
| `--show-ports` | - | Show listening ports for each service | false |
| `--show-all` | `-a` | Show all services including stopped ones | false |
| `--systemd-only` | - | Show only systemd services | false |
| `--format` | `-f` | Output format (table\|json) | table |
| `--help` | `-h` | Help for services command | - |

### Global Help

```bash
# Show main help
linux-toolkit --help

# Show disk command help
linux-toolkit disk --help

# Show cpu command help
linux-toolkit cpu --help

# Show gpu command help
linux-toolkit gpu --help

# Show services command help
linux-toolkit services --help
```

## Requirements

- Linux operating system
- Go 1.16 or higher (for building from source)
- `/sys/block/` directory (standard on Linux)
- `/proc/diskstats` file (standard on Linux)
- `/etc/mtab` or `/proc/mounts` file (standard on Linux)
- `/proc/cpuinfo` file (standard on Linux)
- `/proc/stat` file (standard on Linux)
- `/sys/devices/system/cpu/` directory (standard on Linux)
- `/sys/class/thermal/` directory (standard on Linux)
- `/proc/loadavg` file (standard on Linux)
- `/sys/class/drm/` directory (standard on Linux)
- `/sys/bus/pci/devices/` directory (standard on Linux)
- `/usr/bin/lspci` command (optional, for extended GPU info)
- `/usr/bin/systemctl` command (for systemd services)
- `/proc/[pid]/` directory (standard on Linux)
- `/proc/net/` directory (standard on Linux)

## Project Structure

```
linux-toolkit/
 ├── cmd/
 │   ├── root.go          # Root command
 │   ├── disk.go          # Disk stats subcommand
 │   ├── cpu.go           # CPU info subcommand
 │   ├── gpu.go           # GPU info subcommand
 │   └── services.go      # Services info subcommand
 ├── pkg/
 │   ├── disk/
 │   │   ├── disk.go      # Disk data structures and reader
 │   │   └── output.go    # Output formatting (table/JSON)
 │   ├── cpu/
 │   │   ├── cpu.go       # CPU data structures and reader
 │   │   └── output.go     # Output formatting (table/JSON)
 │   ├── gpu/
 │   │   ├── gpu.go       # GPU data structures and reader
 │   │   └── output.go     # Output formatting (table/JSON)
 │   └── services/
 │       ├── services.go  # Services data structures and reader
 │       └── output.go    # Output formatting (table/JSON)
 ├── main.go              # Entry point
 ├── go.mod               # Go module definition
 └── README.md            # This file
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License
