package cpu

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// CPUInfo contains detailed CPU information
type CPUInfo struct {
	ModelName        string     `json:"modelName"`
	VendorID         string     `json:"vendorId"`
	Architecture     string     `json:"architecture"`
	CPUMode          string     `json:"cpuMode"`
	CPUFamily        int        `json:"cpuFamily"`
	Model            int        `json:"model"`
	Stepping         int        `json:"stepping"`
	PhysicalCores    int        `json:"physicalCores"`
	LogicalCores     int        `json:"logicalCores"`
	ThreadsPerCore   int        `json:"threadsPerCore"`
	Sockets          int        `json:"sockets"`
	MinFrequency     int64      `json:"minFrequency"`
	MaxFrequency     int64      `json:"maxFrequency"`
	CurrentFrequency int64      `json:"currentFrequency"`
	Bogomips         float64    `json:"bogomips"`
	Flags            []string   `json:"flags"`
	CacheL1d         int        `json:"cacheL1d"`
	CacheL1i         int        `json:"cacheL1i"`
	CacheL2          int        `json:"cacheL2"`
	CacheL3          int        `json:"cacheL3"`
	CPUTemperature   float64    `json:"cpuTemperature"`
	LoadAvg1         float64    `json:"loadAvg1"`
	LoadAvg5         float64    `json:"loadAvg5"`
	LoadAvg15        float64    `json:"loadAvg15"`
	TotalProcesses   int        `json:"totalProcesses"`
	RunningProcesses int        `json:"runningProcesses"`
	Cores            []CoreInfo `json:"cores"`
}

// CoreInfo contains per-core CPU information
type CoreInfo struct {
	CoreID       int     `json:"coreId"`
	PhysicalID   int     `json:"physicalId"`
	ProcessorID  int     `json:"processorId"`
	Frequency    int64   `json:"frequency"`
	UsagePercent float64 `json:"usagePercent"`
}

// CPUReader reads CPU information from the Linux system
type CPUReader struct {
	procCpuinfoPath string
	procStatPath    string
	sysCpuPath      string
	sysThermalPath  string
	procLoadavgPath string
}

// NewCPUReader creates a new CPUReader
func NewCPUReader() *CPUReader {
	return &CPUReader{
		procCpuinfoPath: "/proc/cpuinfo",
		procStatPath:    "/proc/stat",
		sysCpuPath:      "/sys/devices/system/cpu",
		sysThermalPath:  "/sys/class/thermal",
		procLoadavgPath: "/proc/loadavg",
	}
}

// ReadCPU reads all CPU information
func (r *CPUReader) ReadCPU() (*CPUInfo, error) {
	info := &CPUInfo{}

	// Read from /proc/cpuinfo
	cpuinfo, err := r.ReadCpuinfo()
	if err != nil {
		return nil, fmt.Errorf("failed to read cpuinfo: %w", err)
	}
	*info = *cpuinfo

	// Read from /sys/devices/system/cpu/ for frequencies and cache
	if err := r.ReadCpuFreq(info); err != nil {
		// Non-fatal, continue
	}

	if err := r.ReadCpuCache(info); err != nil {
		// Non-fatal, continue
	}

	// Read core information
	cores, err := r.ReadCoreInfo()
	if err == nil {
		info.Cores = cores
	}

	// Read from /proc/stat for load and usage
	if err := r.ReadProcStat(info); err != nil {
		return nil, fmt.Errorf("failed to read proc/stat: %w", err)
	}

	// Read temperature
	if temp, err := r.ReadCpuTemp(); err == nil {
		info.CPUTemperature = temp
	}

	// Read load averages
	if load1, load5, load15, err := r.ReadLoadAvg(); err == nil {
		info.LoadAvg1 = load1
		info.LoadAvg5 = load5
		info.LoadAvg15 = load15
	}

	return info, nil
}

// ReadCpuinfo parses /proc/cpuinfo
func (r *CPUReader) ReadCpuinfo() (*CPUInfo, error) {
	info := &CPUInfo{}

	file, err := os.Open(r.procCpuinfoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open cpuinfo: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	physicalIDs := make(map[int]bool)
	coreIDs := make(map[int]bool)
	processorCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "model name":
			info.ModelName = value
		case "vendor_id":
			info.VendorID = value
		case "CPU family":
			if val, err := strconv.Atoi(value); err == nil {
				info.CPUFamily = val
			}
		case "model":
			if val, err := strconv.Atoi(value); err == nil {
				info.Model = val
			}
		case "stepping":
			if val, err := strconv.Atoi(value); err == nil {
				info.Stepping = val
			}
		case "bogomips":
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				info.Bogomips = val
			}
		case "flags":
			info.Flags = strings.Fields(value)
		case "cpu cores":
			if val, err := strconv.Atoi(value); err == nil {
				info.PhysicalCores = val
			}
		case "siblings":
			if val, err := strconv.Atoi(value); err == nil {
				info.LogicalCores = val
			}
		case "physical id":
			if val, err := strconv.Atoi(value); err == nil {
				physicalIDs[val] = true
			}
		case "core id":
			if val, err := strconv.Atoi(value); err == nil {
				coreIDs[val] = true
			}
		case "processor":
			processorCount++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read cpuinfo: %w", err)
	}

	// Calculate derived values
	info.Sockets = len(physicalIDs)
	if info.Sockets == 0 {
		info.Sockets = 1
	}

	if info.LogicalCores == 0 {
		info.LogicalCores = processorCount
	}

	if info.PhysicalCores == 0 {
		info.PhysicalCores = processorCount
	}

	if info.ThreadsPerCore == 0 && info.PhysicalCores > 0 {
		info.ThreadsPerCore = info.LogicalCores / info.PhysicalCores
	}

	// Detect architecture from uname or cpuinfo
	info.Architecture = r.detectArchitecture()
	info.CPUMode = "64-bit"
	if strings.Contains(info.ModelName, "32-bit") || strings.Contains(info.Architecture, "386") {
		info.CPUMode = "32-bit"
	}

	return info, nil
}

// ReadProcStat parses /proc/stat for load and usage
func (r *CPUReader) ReadProcStat(info *CPUInfo) error {
	file, err := os.Open(r.procStatPath)
	if err != nil {
		return fmt.Errorf("failed to open proc/stat: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu ") {
			// Skip aggregate CPU line
			continue
		}
		if strings.HasPrefix(line, "cpu") {
			// Parse per-core CPU line
			fields := strings.Fields(line)
			if len(fields) < 8 {
				continue
			}

			// Format: cpu0 user nice system idle iowait irq softirq
			processorName := fields[0]
			processorID, err := strconv.Atoi(processorName[3:])
			if err != nil {
				continue
			}

			// Calculate usage percentage
			user, _ := strconv.ParseFloat(fields[1], 64)
			nice, _ := strconv.ParseFloat(fields[2], 64)
			system, _ := strconv.ParseFloat(fields[3], 64)
			idle, _ := strconv.ParseFloat(fields[4], 64)

			total := user + nice + system + idle
			var usagePercent float64
			if total > 0 {
				usagePercent = ((user + nice + system) / total) * 100
			}

			// Update core info
			for i := range info.Cores {
				if info.Cores[i].ProcessorID == processorID {
					info.Cores[i].UsagePercent = usagePercent
					break
				}
			}
		} else if strings.HasPrefix(line, "processes") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				if val, err := strconv.Atoi(fields[1]); err == nil {
					info.TotalProcesses = val
				}
			}
		} else if strings.HasPrefix(line, "procs_running") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				if val, err := strconv.Atoi(fields[1]); err == nil {
					info.RunningProcesses = val
				}
			}
		}
	}

	return scanner.Err()
}

// ReadCpuFreq reads CPU frequencies from sysfs
func (r *CPUReader) ReadCpuFreq(info *CPUInfo) error {
	cpu0Path := filepath.Join(r.sysCpuPath, "cpu0", "cpufreq")

	// Read min frequency
	if data, err := os.ReadFile(filepath.Join(cpu0Path, "scaling_min_freq")); err == nil {
		if val, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err == nil {
			info.MinFrequency = val
		}
	}

	// Read max frequency
	if data, err := os.ReadFile(filepath.Join(cpu0Path, "scaling_max_freq")); err == nil {
		if val, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err == nil {
			info.MaxFrequency = val
		}
	}

	// Read current frequency
	if data, err := os.ReadFile(filepath.Join(cpu0Path, "scaling_cur_freq")); err == nil {
		if val, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err == nil {
			info.CurrentFrequency = val
		}
	}

	return nil
}

// ReadCpuCache reads CPU cache sizes from sysfs
func (r *CPUReader) ReadCpuCache(info *CPUInfo) error {
	cpu0Path := filepath.Join(r.sysCpuPath, "cpu0", "cache")

	// Read cache sizes
	indices := []struct {
		index int
		field *int
	}{
		{0, &info.CacheL1d}, // L1 data cache
		{1, &info.CacheL1i}, // L1 instruction cache
		{2, &info.CacheL2},  // L2 cache
		{3, &info.CacheL3},  // L3 cache
	}

	for _, idx := range indices {
		cachePath := filepath.Join(cpu0Path, fmt.Sprintf("index%d", idx.index), "size")
		if data, err := os.ReadFile(cachePath); err == nil {
			sizeStr := strings.TrimSpace(string(data))
			// Parse size (e.g., "32K", "512K", "32M")
			var multiplier int = 1
			if strings.HasSuffix(sizeStr, "K") {
				multiplier = 1
				sizeStr = sizeStr[:len(sizeStr)-1]
			} else if strings.HasSuffix(sizeStr, "M") {
				multiplier = 1024
				sizeStr = sizeStr[:len(sizeStr)-1]
			}
			if val, err := strconv.Atoi(sizeStr); err == nil {
				*idx.field = val * multiplier
			}
		}
	}

	return nil
}

// ReadCoreInfo reads per-core information from sysfs
func (r *CPUReader) ReadCoreInfo() ([]CoreInfo, error) {
	var cores []CoreInfo

	entries, err := os.ReadDir(r.sysCpuPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cpu directory: %w", err)
	}

	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "cpu") {
			continue
		}

		processorName := entry.Name()
		processorID, err := strconv.Atoi(processorName[3:])
		if err != nil {
			continue
		}

		corePath := filepath.Join(r.sysCpuPath, processorName)

		core := CoreInfo{
			ProcessorID: processorID,
		}

		// Read core ID
		if data, err := os.ReadFile(filepath.Join(corePath, "topology", "core_id")); err == nil {
			if val, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
				core.CoreID = val
			}
		}

		// Read physical ID (socket)
		if data, err := os.ReadFile(filepath.Join(corePath, "topology", "physical_package_id")); err == nil {
			if val, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
				core.PhysicalID = val
			}
		}

		// Read frequency
		freqPath := filepath.Join(corePath, "cpufreq", "scaling_cur_freq")
		if data, err := os.ReadFile(freqPath); err == nil {
			if val, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err == nil {
				core.Frequency = val
			}
		}

		cores = append(cores, core)
	}

	// Sort by processor ID
	sort.Slice(cores, func(i, j int) bool {
		return cores[i].ProcessorID < cores[j].ProcessorID
	})

	return cores, nil
}

// ReadCpuTemp reads CPU temperature from thermal zone
func (r *CPUReader) ReadCpuTemp() (float64, error) {
	entries, err := os.ReadDir(r.sysThermalPath)
	if err != nil {
		return 0, fmt.Errorf("failed to read thermal directory: %w", err)
	}

	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "thermal_zone") {
			continue
		}

		zonePath := filepath.Join(r.sysThermalPath, entry.Name())

		// Check if this is a CPU thermal zone
		typePath := filepath.Join(zonePath, "type")
		typeData, err := os.ReadFile(typePath)
		if err != nil {
			continue
		}

		typeStr := strings.TrimSpace(string(typeData))
		if !strings.Contains(strings.ToLower(typeStr), "cpu") &&
			!strings.Contains(strings.ToLower(typeStr), "x86") &&
			!strings.Contains(strings.ToLower(typeStr), "acpi") {
			continue
		}

		// Read temperature
		tempPath := filepath.Join(zonePath, "temp")
		tempData, err := os.ReadFile(tempPath)
		if err != nil {
			continue
		}

		tempStr := strings.TrimSpace(string(tempData))
		if temp, err := strconv.ParseFloat(tempStr, 64); err == nil {
			// Temperature is in millidegrees Celsius
			return temp / 1000.0, nil
		}
	}

	return 0, fmt.Errorf("CPU thermal zone not found")
}

// ReadLoadAvg reads load averages from /proc/loadavg
func (r *CPUReader) ReadLoadAvg() (float64, float64, float64, error) {
	data, err := os.ReadFile(r.procLoadavgPath)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to read loadavg: %w", err)
	}

	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return 0, 0, 0, fmt.Errorf("invalid loadavg format")
	}

	load1, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, 0, 0, err
	}

	load5, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return 0, 0, 0, err
	}

	load15, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return 0, 0, 0, err
	}

	return load1, load5, load15, nil
}

// detectArchitecture detects the CPU architecture
func (r *CPUReader) detectArchitecture() string {
	// Try to read from /proc/cpuinfo first
	file, err := os.Open(r.procCpuinfoPath)
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "Flags:") {
				flags := strings.ToLower(line)
				if strings.Contains(flags, "lm") { // Long Mode = 64-bit
					return "x86_64"
				}
				return "x86"
			}
		}
	}

	// Default to x86_64 for modern systems
	return "x86_64"
}

// FormatMHz formats kHz to MHz or GHz
func FormatMHz(kHz int64) string {
	if kHz >= 1000000 {
		return fmt.Sprintf("%.2f GHz", float64(kHz)/1000000.0)
	}
	return fmt.Sprintf("%.0f MHz", float64(kHz)/1000.0)
}
