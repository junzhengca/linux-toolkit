package battery

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// BatteryInfo contains detailed battery information
type BatteryInfo struct {
	Name             string  `json:"name"`
	Capacity         int     `json:"capacity"`
	Status           string  `json:"status"`
	Voltage          float64 `json:"voltage"`
	Current          float64 `json:"current"`
	Power            float64 `json:"power"`
	Health           string  `json:"health"`
	HealthPercent    float64 `json:"healthPercent"`
	TimeRemaining    string  `json:"timeRemaining"`
	Manufacturer     string  `json:"manufacturer"`
	ModelName        string  `json:"modelName"`
	SerialNumber     string  `json:"serialNumber"`
	Technology       string  `json:"technology"`
	CycleCount       int     `json:"cycleCount"`
	DesignCapacity   float64 `json:"designCapacity"`   // in Wh
	CurrentFullCap   float64 `json:"currentFullCap"`   // in Wh
	VoltageMinDesign float64 `json:"voltageMinDesign"` // in V
}

// BatteryReader reads battery information from the Linux sysfs
type BatteryReader struct {
	sysfsPath string
}

// NewBatteryReader creates a new BatteryReader
func NewBatteryReader() *BatteryReader {
	return &BatteryReader{
		sysfsPath: "/sys/class/power_supply",
	}
}

// ListBatteries returns a list of available battery device names
func (r *BatteryReader) ListBatteries() ([]string, error) {
	entries, err := os.ReadDir(r.sysfsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read power_supply directory: %w", err)
	}

	var batteries []string
	for _, entry := range entries {
		// Check if this is a battery device by looking for type file
		// Note: In /sys/class/power_supply/, entries are symlinks, so we don't check IsDir()
		typePath := filepath.Join(r.sysfsPath, entry.Name(), "type")
		typeData, err := os.ReadFile(typePath)
		if err == nil {
			typeStr := strings.TrimSpace(string(typeData))
			if typeStr == "Battery" {
				batteries = append(batteries, entry.Name())
			}
		}
	}

	if len(batteries) == 0 {
		return nil, fmt.Errorf("no batteries found")
	}

	return batteries, nil
}

// ReadBattery reads battery information for the specified device
func (r *BatteryReader) ReadBattery(deviceName string) (*BatteryInfo, error) {
	devicePath := filepath.Join(r.sysfsPath, deviceName)

	info := &BatteryInfo{
		Name: deviceName,
	}

	// Read capacity
	capacity, err := r.readIntFile(devicePath, "capacity")
	if err == nil {
		info.Capacity = capacity
	}

	// Read status
	status, err := r.readStringFile(devicePath, "status")
	if err == nil {
		info.Status = strings.TrimSpace(status)
	}

	// Read voltage (in microvolts, convert to volts)
	voltageUV, err := r.readIntFile(devicePath, "voltage_now")
	if err == nil {
		info.Voltage = float64(voltageUV) / 1_000_000
	}

	// Read current (in microamps, convert to amps)
	currentUA, err := r.readIntFile(devicePath, "current_now")
	if err == nil {
		info.Current = float64(currentUA) / 1_000_000
	}

	// Read power (in microwatts, convert to watts)
	powerUW, err := r.readIntFile(devicePath, "power_now")
	if err == nil {
		info.Power = float64(powerUW) / 1_000_000
	} else {
		// Calculate power from voltage and current if power_now not available
		if info.Voltage > 0 && info.Current > 0 {
			info.Power = info.Voltage * info.Current
		}
	}

	// Read manufacturer
	manufacturer, err := r.readStringFile(devicePath, "manufacturer")
	if err == nil {
		info.Manufacturer = strings.TrimSpace(manufacturer)
	}

	// Read model name
	modelName, err := r.readStringFile(devicePath, "model_name")
	if err == nil {
		info.ModelName = strings.TrimSpace(modelName)
	}

	// Read serial number
	serialNumber, err := r.readStringFile(devicePath, "serial_number")
	if err == nil {
		info.SerialNumber = strings.TrimSpace(serialNumber)
	}

	// Read technology
	technology, err := r.readStringFile(devicePath, "technology")
	if err == nil {
		info.Technology = strings.TrimSpace(technology)
	}

	// Read cycle count
	cycleCount, err := r.readIntFile(devicePath, "cycle_count")
	if err == nil {
		info.CycleCount = cycleCount
	}

	// Read design capacity (in microwatt-hours, convert to watt-hours)
	energyFullDesign, err := r.readIntFile(devicePath, "energy_full_design")
	if err == nil {
		info.DesignCapacity = float64(energyFullDesign) / 1_000_000
	}

	// Read current full capacity (in microwatt-hours, convert to watt-hours)
	energyFull, err := r.readIntFile(devicePath, "energy_full")
	if err == nil {
		info.CurrentFullCap = float64(energyFull) / 1_000_000
	}

	// Read voltage min design (in microvolts, convert to volts)
	voltageMinDesign, err := r.readIntFile(devicePath, "voltage_min_design")
	if err == nil {
		info.VoltageMinDesign = float64(voltageMinDesign) / 1_000_000
	}

	// Calculate health percentage
	if info.DesignCapacity > 0 {
		info.HealthPercent = (info.CurrentFullCap / info.DesignCapacity) * 100
	}

	// Determine health status string
	info.Health = r.determineHealthStatus(info.HealthPercent)

	// Calculate time remaining
	info.TimeRemaining = r.calculateTimeRemaining(info)

	return info, nil
}

// readIntFile reads an integer value from a file
func (r *BatteryReader) readIntFile(devicePath, filename string) (int, error) {
	data, err := os.ReadFile(filepath.Join(devicePath, filename))
	if err != nil {
		return 0, err
	}

	value, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, err
	}

	return value, nil
}

// readStringFile reads a string value from a file
func (r *BatteryReader) readStringFile(devicePath, filename string) (string, error) {
	data, err := os.ReadFile(filepath.Join(devicePath, filename))
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// calculateTimeRemaining estimates the time remaining based on current status and power
func (r *BatteryReader) calculateTimeRemaining(info *BatteryInfo) string {
	if info.Status == "Full" || info.Status == "Unknown" {
		return "N/A"
	}

	if info.Power <= 0 {
		return "Unknown"
	}

	// Read energy values for more accurate calculation
	devicePath := filepath.Join(r.sysfsPath, info.Name)
	energyNow, err1 := r.readIntFile(devicePath, "energy_now")
	energyFull, err2 := r.readIntFile(devicePath, "energy_full")

	if err1 != nil || err2 != nil {
		// Fallback: estimate based on capacity and power
		if info.Status == "Charging" {
			return "~Unknown"
		}
		// For discharging, estimate based on capacity
		hours := float64(100-info.Capacity) / (info.Power * 10) // Rough estimate
		if hours < 0 {
			return "Unknown"
		}
		return formatDuration(hours)
	}

	if info.Status == "Charging" {
		// Time to full
		remainingEnergy := energyFull - energyNow
		hours := float64(remainingEnergy) / float64(info.Power*1_000_000)
		if hours < 0 {
			return "Unknown"
		}
		return "~" + formatDuration(hours)
	}

	// Discharging - time to empty
	hours := float64(energyNow) / float64(info.Power*1_000_000)
	if hours < 0 {
		return "Unknown"
	}
	return "~" + formatDuration(hours)
}

// formatDuration formats hours in a human-readable format
func formatDuration(hours float64) string {
	minutes := int(hours * 60)
	h := minutes / 60
	m := minutes % 60

	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

// determineHealthStatus returns a human-readable health status based on health percentage
func (r *BatteryReader) determineHealthStatus(healthPercent float64) string {
	switch {
	case healthPercent >= 90:
		return "Excellent"
	case healthPercent >= 75:
		return "Good"
	case healthPercent >= 50:
		return "Fair"
	case healthPercent >= 25:
		return "Poor"
	default:
		return "Critical"
	}
}
