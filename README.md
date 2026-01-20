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

### Battery Statistics

Display detailed battery information including capacity, status, voltage, current, power, health, manufacturer, model, serial number, technology, cycle count, design capacity, and estimated time remaining.

```bash
# Display battery stats in table format (default) with colors and emojis
linux-toolkit battery

# Display battery stats in JSON format
linux-toolkit battery --format json

# Display specific battery device
linux-toolkit battery --battery BAT0
```

#### Table Format Output (with colors and emojis)

```
  🔋 Battery Information

  📱 Device Info
  ────────────────────────────────────────
  Device:            BAT0
  Manufacturer:      ASUSTeK
  Model:             ASUS Battery
  Serial Number:     
  Technology:        Li-ion

  📊 Status
  ────────────────────────────────────────
  Capacity:          48%
  Status             📉 Discharging
  Health             63.8% (Fair)
  Cycle Count:       94

  ⚡ Capacity
  ────────────────────────────────────────
  Design Capacity:   63.04 Wh
  Current Full Cap:  40.22 Wh

  🔌 Power
  ────────────────────────────────────────
  Voltage:           11.98 V (min: 11.98 V)
  Current:           0.00 A
  Power:             11.36 W

  ⏱️  Time Remaining
  ────────────────────────────────────────
                     ~1h 42m
```

#### JSON Format Output

```json
{
  "name": "BAT0",
  "capacity": 48,
  "status": "Discharging",
  "voltage": 11.985,
  "current": 0,
  "power": 11.026,
  "health": "Fair",
  "healthPercent": 63.80133563871131,
  "timeRemaining": "~1h 46m",
  "manufacturer": "ASUSTeK",
  "modelName": "ASUS Battery",
  "serialNumber": "",
  "technology": "Li-ion",
  "cycleCount": 94,
  "designCapacity": 63.041,
  "currentFullCap": 40.221,
  "voltageMinDesign": 11.985
}
```

#### Battery Health Status

The health status is calculated based on current full capacity compared to design capacity:

| Health Percentage | Status | Color |
|------------------|--------|-------|
| 90%+ | Excellent | 🟢 Green |
| 75%-89% | Good | 🟢 Light Green |
| 50%-74% | Fair | 🟡 Yellow |
| 25%-49% | Poor | 🟡 Light Yellow |
| <25% | Critical | 🔴 Red |

#### Status Indicators

| Status | Emoji |
|--------|-------|
| Charging | ⚡ |
| Discharging | 📉 |
| Full | ✅ |
| Unknown | ❓ |

### Command Options

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--battery` | `-b` | Battery device name (e.g., BAT0) | auto-detect |
| `--format` | `-f` | Output format (table\|json) | table |
| `--help` | `-h` | Help for battery command | - |

### Global Help

```bash
# Show main help
linux-toolkit --help

# Show battery command help
linux-toolkit battery --help
```

## Requirements

- Linux operating system
- Go 1.16 or higher (for building from source)
- `/sys/class/power_supply/` directory (standard on Linux)

## Project Structure

```
linux-toolkit/
├── cmd/
│   ├── root.go          # Root command
│   └── battery.go       # Battery stats subcommand
├── pkg/
│   └── battery/
│       ├── battery.go   # Battery data structures and reader
│       └── output.go    # Output formatting (table/JSON)
├── main.go              # Entry point
├── go.mod               # Go module definition
└── README.md            # This file
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License

## Future Tools

This toolkit is designed to be extensible. Future tools may include:
- CPU statistics
- Memory statistics
- Disk usage
- Network statistics
- And more...
