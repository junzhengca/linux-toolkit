package gpu

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// GPUInfo contains detailed GPU information
type GPUInfo struct {
	// Basic Device Information
	CardName        string `json:"cardName"`
	DeviceName      string `json:"deviceName"`
	Driver          string `json:"driver"`
	Vendor          string `json:"vendor"`
	VendorID        string `json:"vendorId"`
	DeviceID        string `json:"deviceId"`
	SubsystemVendor string `json:"subsystemVendor"`
	SubsystemDevice string `json:"subsystemDevice"`
	Class           string `json:"class"`
	Revision        string `json:"revision"`

	// Display Information
	Enabled           bool     `json:"enabled"`
	Status            string   `json:"status"`
	Connectors        []string `json:"connectors"`
	EnabledConnectors []string `json:"enabledConnectors"`
	Modes             []string `json:"modes"`

	// Memory Information
	VRAMSize uint64 `json:"vramSize"`
	VRAMUsed uint64 `json:"vramUsed"`
	VRAMFree uint64 `json:"vramFree"`
	GARTSize uint64 `json:"gartSize"`
	GARTUsed uint64 `json:"gartUsed"`

	// Clock Information
	CoreClock      uint64 `json:"coreClock"`
	MemoryClock    uint64 `json:"memoryClock"`
	MaxCoreClock   uint64 `json:"maxCoreClock"`
	MaxMemoryClock uint64 `json:"maxMemoryClock"`

	// Power Information
	PowerUsage uint64 `json:"powerUsage"`
	PowerLimit uint64 `json:"powerLimit"`
	PowerCap   uint64 `json:"powerCap"`

	// Temperature Information
	Temperature     float64 `json:"temperature"`
	TemperatureCrit float64 `json:"temperatureCrit"`
	FanSpeed        uint64  `json:"fanSpeed"`
	FanSpeedPercent float64 `json:"fanSpeedPercent"`

	// Utilization Information
	GPUUtilPercent    float64 `json:"gpuUtilPercent"`
	MemoryUtilPercent float64 `json:"memoryUtilPercent"`

	// Bus Information
	BusID      string `json:"busId"`
	BusWidth   string `json:"busWidth"`
	PCIEGen    string `json:"pcieGen"`
	MaxPCIEGen string `json:"maxPcieGen"`

	// Firmware Information
	VBIOSVersion    string `json:"vbiosVersion"`
	FirmwareVersion string `json:"firmwareVersion"`

	// Additional Information
	DevicePath   string `json:"devicePath"`
	SysfsPath    string `json:"sysfsPath"`
	IsPrimary    bool   `json:"isPrimary"`
	GPUType      string `json:"gpuType"`
	ComputeUnits uint64 `json:"computeUnits"`
	CUDAcores    uint64 `json:"cudaCores"`
	Shaders      uint64 `json:"shaders"`
	TextureUnits uint64 `json:"textureUnits"`
	ROPs         uint64 `json:"rops"`
}

// ConnectorInfo contains connector information
type ConnectorInfo struct {
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Status     string            `json:"status"`
	Enabled    bool              `json:"enabled"`
	Properties map[string]string `json:"properties"`
}

// GPUReader reads GPU information from the Linux system
type GPUReader struct {
	sysfsDrmPath string
	sysfsPciPath string
	lspciPath    string
}

// NewGPUReader creates a new GPUReader
func NewGPUReader() *GPUReader {
	return &GPUReader{
		sysfsDrmPath: "/sys/class/drm",
		sysfsPciPath: "/sys/bus/pci/devices",
		lspciPath:    "/usr/bin/lspci",
	}
}

// ListGPUs returns a list of available GPU card names
func (r *GPUReader) ListGPUs() ([]string, error) {
	entries, err := os.ReadDir(r.sysfsDrmPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read drm directory: %w", err)
	}

	var cards []string
	for _, entry := range entries {
		// Look for card directories (card0, card1, etc.)
		if strings.HasPrefix(entry.Name(), "card") {
			// Skip card-N-... (connector subdirectories)
			if !strings.Contains(entry.Name(), "-") {
				cards = append(cards, entry.Name())
			}
		}
	}

	// Sort cards to ensure consistent ordering
	sort.Strings(cards)

	if len(cards) == 0 {
		return nil, fmt.Errorf("no GPUs found")
	}

	return cards, nil
}

// ReadGPU reads all GPU information for the specified card
func (r *GPUReader) ReadGPU(cardName string) (*GPUInfo, error) {
	info := &GPUInfo{
		CardName:   cardName,
		DevicePath: filepath.Join(r.sysfsDrmPath, cardName),
	}

	// Read basic device info
	if err := r.ReadBasicInfo(cardName, info); err != nil {
		return nil, fmt.Errorf("failed to read basic info: %w", err)
	}

	// Read PCI information
	if err := r.ReadPCIInfo(cardName, info); err == nil {
		// Non-fatal, continue
	}

	// Read lspci information
	if err := r.ReadLSPCIInfo(cardName, info); err == nil {
		// Non-fatal, continue
	}

	// Read memory information
	if err := r.ReadMemoryInfo(cardName, info); err == nil {
		// Non-fatal, continue
	}

	// Read clock information
	if err := r.ReadClockInfo(cardName, info); err == nil {
		// Non-fatal, continue
	}

	// Read power information
	if err := r.ReadPowerInfo(cardName, info); err == nil {
		// Non-fatal, continue
	}

	// Read temperature information
	if err := r.ReadTemperatureInfo(cardName, info); err == nil {
		// Non-fatal, continue
	}

	// Read utilization information
	if err := r.ReadUtilizationInfo(cardName, info); err == nil {
		// Non-fatal, continue
	}

	// Read bus information
	if err := r.ReadBusInfo(cardName, info); err == nil {
		// Non-fatal, continue
	}

	// Read firmware information
	if err := r.ReadFirmwareInfo(cardName, info); err == nil {
		// Non-fatal, continue
	}

	// Determine GPU type
	r.DetermineGPUType(info)

	return info, nil
}

// ReadBasicInfo reads basic device information from sysfs
func (r *GPUReader) ReadBasicInfo(cardName string, info *GPUInfo) error {
	cardPath := filepath.Join(r.sysfsDrmPath, cardName)
	devicePath := filepath.Join(cardPath, "device")

	info.SysfsPath = devicePath

	// Check if device is enabled
	if enabled, err := r.readBoolFile(filepath.Join(devicePath, "enable")); err == nil {
		info.Enabled = enabled
	}

	// Read card status
	if status, err := r.readStringFile(filepath.Join(cardPath, "status")); err == nil {
		info.Status = strings.TrimSpace(status)
	}

	// Read device name (e.g., /dev/dri/card0)
	if _, err := r.readStringFile(filepath.Join(cardPath, "dev")); err == nil {
		info.DeviceName = "/dev/dri/" + cardName
	}

	// Read driver from device symlink
	if driverPath, err := filepath.EvalSymlinks(filepath.Join(devicePath, "driver")); err == nil {
		info.Driver = filepath.Base(driverPath)
	}

	// Read connectors
	if connectors, err := r.ReadConnectors(cardName); err == nil {
		for _, conn := range connectors {
			info.Connectors = append(info.Connectors, conn.Name)
			if conn.Enabled {
				info.EnabledConnectors = append(info.EnabledConnectors, conn.Name)
			}
		}
	}

	// Read supported modes
	if modes, err := r.readStringFile(filepath.Join(cardPath, "modes")); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(modes))
		for scanner.Scan() {
			mode := strings.TrimSpace(scanner.Text())
			if mode != "" {
				info.Modes = append(info.Modes, mode)
			}
		}
	}

	return nil
}

// ReadPCIInfo reads PCI device information from sysfs
func (r *GPUReader) ReadPCIInfo(cardName string, info *GPUInfo) error {
	devicePath := filepath.Join(r.sysfsDrmPath, cardName, "device")

	// Read vendor ID
	if vendorID, err := r.readStringFile(filepath.Join(devicePath, "vendor")); err == nil {
		info.VendorID = strings.TrimPrefix(strings.TrimSpace(vendorID), "0x")
		info.Vendor = r.DetectVendor(info.VendorID)
	}

	// Read device ID
	if deviceID, err := r.readStringFile(filepath.Join(devicePath, "device")); err == nil {
		info.DeviceID = strings.TrimPrefix(strings.TrimSpace(deviceID), "0x")
	}

	// Read subsystem vendor
	if subVendor, err := r.readStringFile(filepath.Join(devicePath, "subsystem_vendor")); err == nil {
		info.SubsystemVendor = strings.TrimPrefix(strings.TrimSpace(subVendor), "0x")
	}

	// Read subsystem device
	if subDevice, err := r.readStringFile(filepath.Join(devicePath, "subsystem_device")); err == nil {
		info.SubsystemDevice = strings.TrimPrefix(strings.TrimSpace(subDevice), "0x")
	}

	// Read class
	if class, err := r.readStringFile(filepath.Join(devicePath, "class")); err == nil {
		info.Class = strings.TrimPrefix(strings.TrimSpace(class), "0x")
	}

	// Read revision
	if revision, err := r.readStringFile(filepath.Join(devicePath, "revision")); err == nil {
		info.Revision = strings.TrimPrefix(strings.TrimSpace(revision), "0x")
	}

	return nil
}

// ReadLSPCIInfo reads information from lspci command
func (r *GPUReader) ReadLSPCIInfo(cardName string, info *GPUInfo) error {
	// Get the bus ID from device symlink
	devicePath := filepath.Join(r.sysfsDrmPath, cardName, "device")
	busID, err := filepath.EvalSymlinks(devicePath)
	if err != nil {
		return err
	}
	busID = filepath.Base(busID)

	// Run lspci for this device
	cmd := exec.Command(r.lspciPath, "-v", "-s", busID)
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	// Parse lspci output
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// Extract subsystem info
		if strings.Contains(line, "Subsystem:") {
			parts := strings.SplitN(line, "Subsystem:", 2)
			if len(parts) == 2 {
				subsystem := strings.TrimSpace(parts[1])
				// Format: "Vendor Device [vendorID:deviceID]"
				if idx := strings.Index(subsystem, "["); idx > 0 {
					subsystem = strings.TrimSpace(subsystem[:idx])
				}
			}
		}
	}

	return nil
}

// ReadMemoryInfo reads memory information
func (r *GPUReader) ReadMemoryInfo(cardName string, info *GPUInfo) error {
	devicePath := filepath.Join(r.sysfsDrmPath, cardName, "device")

	// Try AMD-specific memory info
	if vramTotal, err := r.parseHexFile(filepath.Join(devicePath, "mem_info_vram_total")); err == nil {
		info.VRAMSize = vramTotal
	}
	if vramUsed, err := r.parseHexFile(filepath.Join(devicePath, "mem_info_vram_used")); err == nil {
		info.VRAMUsed = vramUsed
	}

	// Calculate VRAM free
	if info.VRAMSize > 0 && info.VRAMUsed > 0 {
		info.VRAMFree = info.VRAMSize - info.VRAMUsed
	}

	// Try GART size
	if gartSize, err := r.parseHexFile(filepath.Join(devicePath, "gart_size")); err == nil {
		info.GARTSize = gartSize
	}

	// Try NVIDIA-specific memory info
	nvidiaPath := filepath.Join(devicePath, "memory")
	if _, err := os.Stat(nvidiaPath); err == nil {
		// Read NVIDIA memory info from /proc/driver/nvidia/gpus/...
		// This would require additional parsing
	}

	return nil
}

// ReadClockInfo reads clock information
func (r *GPUReader) ReadClockInfo(cardName string, info *GPUInfo) error {
	devicePath := filepath.Join(r.sysfsDrmPath, cardName, "device")

	// Try AMD-specific clock info
	if _, err := r.parseDecFile(filepath.Join(devicePath, "gpu_busy_percent")); err == nil {
		// This is actually utilization, not clock
	}

	// Read AMD DPM clocks
	dpmSclkPath := filepath.Join(devicePath, "pp_dpm_sclk")
	if data, err := os.ReadFile(dpmSclkPath); err == nil {
		// Parse current clock from pp_dpm_sclk
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.Contains(line, "*") {
				// This is the current clock
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					// Format: "0: 500Mhz *" or "0: 500Mhz *"
					clockStr := strings.TrimSuffix(strings.TrimSpace(parts[1]), "Mhz")
					if clock, err := strconv.ParseUint(clockStr, 10, 64); err == nil {
						info.CoreClock = clock
					}
				}
			}
		}
	}

	// Read Intel-specific clock info
	if curFreq, err := r.parseDecFile(filepath.Join(devicePath, "gt_cur_freq_mhz")); err == nil {
		info.CoreClock = curFreq
	}
	if maxFreq, err := r.parseDecFile(filepath.Join(devicePath, "gt_RP0_freq_mhz")); err == nil {
		info.MaxCoreClock = maxFreq
	}

	return nil
}

// ReadPowerInfo reads power information
func (r *GPUReader) ReadPowerInfo(cardName string, info *GPUInfo) error {
	devicePath := filepath.Join(r.sysfsDrmPath, cardName, "device")

	// Try AMD-specific power info
	if _, err := r.parseDecFile(filepath.Join(devicePath, "gpu_busy_percent")); err == nil {
		// Not power, but we'll check other files
	}

	// Read power from hwmon
	hwmonPath := filepath.Join(devicePath, "hwmon")
	if entries, err := os.ReadDir(hwmonPath); err == nil {
		for _, entry := range entries {
			hwmonDir := filepath.Join(hwmonPath, entry.Name())
			// Try to read power files
			if power, err := r.parseDecFile(filepath.Join(hwmonDir, "power1_input")); err == nil {
				info.PowerUsage = power / 1000 // Convert microwatts to milliwatts
			}
			if powerLimit, err := r.parseDecFile(filepath.Join(hwmonDir, "power1_cap")); err == nil {
				info.PowerLimit = powerLimit / 1000
			}
		}
	}

	return nil
}

// ReadTemperatureInfo reads temperature information
func (r *GPUReader) ReadTemperatureInfo(cardName string, info *GPUInfo) error {
	devicePath := filepath.Join(r.sysfsDrmPath, cardName, "device")

	// Try hwmon for temperature
	hwmonPath := filepath.Join(devicePath, "hwmon")
	if entries, err := os.ReadDir(hwmonPath); err == nil {
		for _, entry := range entries {
			hwmonEntryDir := filepath.Join(hwmonPath, entry.Name())
			// Read temperature (in millidegrees Celsius)
			if temp, err := r.parseDecFile(filepath.Join(hwmonEntryDir, "temp1_input")); err == nil {
				info.Temperature = float64(temp) / 1000.0
			}
			// Read critical temperature
			if tempCrit, err := r.parseDecFile(filepath.Join(hwmonEntryDir, "temp1_crit")); err == nil {
				info.TemperatureCrit = float64(tempCrit) / 1000.0
			}
		}
	}

	// Try AMD-specific temperature
	if temp, err := r.parseDecFile(filepath.Join(devicePath, "gpu_temp")); err == nil {
		info.Temperature = float64(temp) / 1000.0
	}

	// Read fan speed
	if fanSpeed, err := r.parseDecFile(filepath.Join(devicePath, "fan_speed")); err == nil {
		info.FanSpeed = fanSpeed
	}
	if fanPercent, err := r.parseDecFile(filepath.Join(devicePath, "pwm1")); err == nil {
		// PWM is usually 0-255, convert to percentage
		info.FanSpeedPercent = float64(fanPercent) / 255.0 * 100.0
	}

	return nil
}

// ReadUtilizationInfo reads utilization information
func (r *GPUReader) ReadUtilizationInfo(cardName string, info *GPUInfo) error {
	devicePath := filepath.Join(r.sysfsDrmPath, cardName, "device")

	// Try AMD-specific utilization
	if gpuBusy, err := r.parseDecFile(filepath.Join(devicePath, "gpu_busy_percent")); err == nil {
		info.GPUUtilPercent = float64(gpuBusy)
	}
	if memBusy, err := r.parseDecFile(filepath.Join(devicePath, "mem_busy_percent")); err == nil {
		info.MemoryUtilPercent = float64(memBusy)
	}

	// Try NVIDIA-specific utilization
	if gpuUtil, err := r.parseDecFile(filepath.Join(devicePath, "utilization")); err == nil {
		info.GPUUtilPercent = float64(gpuUtil)
	}

	return nil
}

// ReadBusInfo reads bus information
func (r *GPUReader) ReadBusInfo(cardName string, info *GPUInfo) error {
	devicePath := filepath.Join(r.sysfsDrmPath, cardName, "device")

	// Get bus ID from symlink
	busID, err := filepath.EvalSymlinks(devicePath)
	if err == nil {
		info.BusID = filepath.Base(busID)
	}

	// Try to read PCIe width and generation
	if linkWidth, err := r.readStringFile(filepath.Join(devicePath, "max_link_width")); err == nil {
		info.BusWidth = "x" + strings.TrimSpace(linkWidth)
	}
	if linkSpeed, err := r.readStringFile(filepath.Join(devicePath, "max_link_speed")); err == nil {
		// Speed is in GT/s, convert to PCIe generation
		speed := strings.TrimSpace(linkSpeed)
		speed = strings.TrimSuffix(speed, "GT/s")
		if s, err := strconv.ParseFloat(speed, 64); err == nil {
			info.MaxPCIEGen = r.pcieSpeedToGen(s)
		}
	}

	// Read current link info
	if curWidth, err := r.readStringFile(filepath.Join(devicePath, "current_link_width")); err == nil {
		if info.BusWidth == "" {
			info.BusWidth = "x" + strings.TrimSpace(curWidth)
		}
	}
	if curSpeed, err := r.readStringFile(filepath.Join(devicePath, "current_link_speed")); err == nil {
		speed := strings.TrimSpace(curSpeed)
		speed = strings.TrimSuffix(speed, "GT/s")
		if s, err := strconv.ParseFloat(speed, 64); err == nil {
			info.PCIEGen = r.pcieSpeedToGen(s)
		}
	}

	return nil
}

// ReadFirmwareInfo reads firmware information
func (r *GPUReader) ReadFirmwareInfo(cardName string, info *GPUInfo) error {
	devicePath := filepath.Join(r.sysfsDrmPath, cardName, "device")

	// Try to read VBIOS version
	if vbios, err := r.readStringFile(filepath.Join(devicePath, "vbios_version")); err == nil {
		info.VBIOSVersion = strings.TrimSpace(vbios)
	}

	// Try firmware version
	if fw, err := r.readStringFile(filepath.Join(devicePath, "firmware_version")); err == nil {
		info.FirmwareVersion = strings.TrimSpace(fw)
	}

	return nil
}

// ReadConnectors reads connector information
func (r *GPUReader) ReadConnectors(cardName string) ([]ConnectorInfo, error) {
	cardPath := filepath.Join(r.sysfsDrmPath, cardName)
	entries, err := os.ReadDir(cardPath)
	if err != nil {
		return nil, err
	}

	var connectors []ConnectorInfo
	for _, entry := range entries {
		// Look for connector directories (card0-HDMI-A-1, card0-DP-1, etc.)
		if strings.HasPrefix(entry.Name(), cardName+"-") {
			connName := entry.Name()
			connPath := filepath.Join(cardPath, connName)

			conn := ConnectorInfo{
				Name: strings.TrimPrefix(connName, cardName+"-"),
			}

			// Read connector status
			if status, err := r.readStringFile(filepath.Join(connPath, "status")); err == nil {
				conn.Status = strings.TrimSpace(status)
			}

			// Read enabled status
			if enabled, err := r.readBoolFile(filepath.Join(connPath, "enabled")); err == nil {
				conn.Enabled = enabled
			}

			// Determine connector type from name
			conn.Type = r.getConnectorType(conn.Name)

			connectors = append(connectors, conn)
		}
	}

	return connectors, nil
}

// DetermineGPUType determines if the GPU is integrated or discrete
func (r *GPUReader) DetermineGPUType(info *GPUInfo) {
	// Check if it's Intel (usually integrated)
	if info.Vendor == "Intel" {
		info.GPUType = "integrated"
		return
	}

	// Check if it's AMD with specific device IDs (integrated APUs)
	if info.Vendor == "AMD" {
		// AMD APUs have device IDs in specific ranges
		if deviceID, err := strconv.ParseUint(info.DeviceID, 16, 64); err == nil {
			// Raven Ridge, Picasso, Renoir, Cezanne, etc. are APUs
			if (deviceID >= 0x15dd && deviceID <= 0x15df) ||
				(deviceID >= 0x1636 && deviceID <= 0x1646) ||
				(deviceID >= 0x1638 && deviceID <= 0x1646) ||
				(deviceID >= 0x1681 && deviceID <= 0x1683) {
				info.GPUType = "integrated"
				return
			}
		}
	}

	// Default to discrete
	info.GPUType = "discrete"
}

// DetectVendor detects the vendor name from vendor ID
func (r *GPUReader) DetectVendor(vendorID string) string {
	vendorMap := map[string]string{
		"1002": "AMD",
		"10de": "NVIDIA",
		"8086": "Intel",
		"102b": "Matrox",
		"1a03": "ASPEED",
		"15ad": "VMware",
	}

	if vendor, ok := vendorMap[strings.ToLower(vendorID)]; ok {
		return vendor
	}

	return "Unknown"
}

// Helper functions

func (r *GPUReader) getCardSysfsPath(cardName string) string {
	return filepath.Join(r.sysfsDrmPath, cardName)
}

func (r *GPUReader) getDevicePCIPath(cardName string) (string, error) {
	devicePath := filepath.Join(r.sysfsDrmPath, cardName, "device")
	return filepath.EvalSymlinks(devicePath)
}

func (r *GPUReader) parseHexFile(path string) (uint64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	value, err := strconv.ParseUint(strings.TrimSpace(string(data)), 0, 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (r *GPUReader) parseDecFile(path string) (uint64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	value, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (r *GPUReader) readStringFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (r *GPUReader) readBoolFile(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(data)) == "1", nil
}

func (r *GPUReader) getConnectorType(name string) string {
	switch {
	case strings.HasPrefix(name, "HDMI"):
		return "HDMI"
	case strings.HasPrefix(name, "DP"):
		return "DisplayPort"
	case strings.HasPrefix(name, "DVI"):
		return "DVI"
	case strings.HasPrefix(name, "VGA"):
		return "VGA"
	case strings.HasPrefix(name, "eDP"):
		return "eDP"
	case strings.HasPrefix(name, "LVDS"):
		return "LVDS"
	default:
		return "Unknown"
	}
}

func (r *GPUReader) pcieSpeedToGen(speed float64) string {
	switch {
	case speed >= 32.0:
		return "5.0"
	case speed >= 16.0:
		return "4.0"
	case speed >= 8.0:
		return "3.0"
	case speed >= 5.0:
		return "2.0"
	case speed >= 2.5:
		return "1.0"
	default:
		return "Unknown"
	}
}

// FormatBytes formats bytes to a human-readable string
func FormatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
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

// FormatMHz formats MHz to a human-readable string
func FormatMHz(mHz uint64) string {
	if mHz >= 1000 {
		return fmt.Sprintf("%.2f GHz", float64(mHz)/1000.0)
	}
	return fmt.Sprintf("%d MHz", mHz)
}
