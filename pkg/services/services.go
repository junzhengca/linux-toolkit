package services

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ServiceInfo contains detailed service information
type ServiceInfo struct {
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	Status         string     `json:"status"`
	Loaded         string     `json:"loaded"`
	Active         string     `json:"active"`
	SubState       string     `json:"subState"`
	PID            int        `json:"pid"`
	MemoryMB       float64    `json:"memoryMB"`
	CPUPercent     float64    `json:"cpuPercent"`
	StartTime      string     `json:"startTime"`
	Uptime         string     `json:"uptime"`
	User           string     `json:"user"`
	Command        string     `json:"command"`
	ListeningPorts []PortInfo `json:"listeningPorts"`
	Type           string     `json:"type"`
}

// PortInfo contains port binding information
type PortInfo struct {
	Protocol    string `json:"protocol"`
	LocalAddr   string `json:"localAddr"`
	Port        int    `json:"port"`
	State       string `json:"state"`
	ProcessName string `json:"processName"`
	PID         int    `json:"pid"`
}

// ServicesSummary contains a summary of all services
type ServicesSummary struct {
	TotalServices int           `json:"totalServices"`
	Running       int           `json:"running"`
	Stopped       int           `json:"stopped"`
	Failed        int           `json:"failed"`
	TotalPorts    int           `json:"totalPorts"`
	Services      []ServiceInfo `json:"services"`
}

// ServicesReader reads services information from the Linux system
type ServicesReader struct {
	systemctlPath string
	procPath      string
	procNetPath   string
}

// NewServicesReader creates a new ServicesReader
func NewServicesReader() *ServicesReader {
	return &ServicesReader{
		systemctlPath: "/usr/bin/systemctl",
		procPath:      "/proc",
		procNetPath:   "/proc/net",
	}
}

// ReadAllServices reads all services information
func (r *ServicesReader) ReadAllServices() (*ServicesSummary, error) {
	summary := &ServicesSummary{
		Services: []ServiceInfo{},
	}

	// Read systemd services
	systemdServices, err := r.ReadSystemdServices()
	if err != nil {
		// Non-fatal, continue with process-based services
	} else {
		summary.Services = append(summary.Services, systemdServices...)
	}

	// Read listening ports
	ports, err := r.ReadListeningPorts()
	if err == nil {
		summary.TotalPorts = len(ports)
		r.MatchPortsToServices(summary.Services, ports)
	}

	// Calculate summary statistics
	r.calculateSummary(summary)

	return summary, nil
}

// ReadSystemdServices reads systemd services using systemctl
func (r *ServicesReader) ReadSystemdServices() ([]ServiceInfo, error) {
	// Check if systemctl is available
	if _, err := os.Stat(r.systemctlPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("systemctl not found")
	}

	// Get list of all services
	cmd := exec.Command(r.systemctlPath, "list-units", "--type=service", "--all", "--no-pager")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	services, err := r.parseSystemctlListOutput(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse systemctl output: %w", err)
	}

	// Get detailed information for each service
	for i := range services {
		details, err := r.GetServiceDetails(services[i].Name)
		if err == nil {
			services[i] = *details
		}
	}

	return services, nil
}

// parseSystemctlListOutput parses systemctl list-units output
func (r *ServicesReader) parseSystemctlListOutput(output string) ([]ServiceInfo, error) {
	var services []ServiceInfo

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		// Skip header, footer, and empty/whitespace-only lines
		if strings.HasPrefix(line, "UNIT") || strings.HasPrefix(line, "LOAD") || strings.TrimSpace(line) == "" || strings.HasPrefix(line, "Legend:") || strings.HasPrefix(line, "To") {
			continue
		}

		// Parse line: UNIT LOAD ACTIVE SUB DESCRIPTION
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		var nameIndex int
		var loadedIndex int
		var activeIndex int
		var subStateIndex int
		var descStartIndex int

		if fields[0] == "●" {
			nameIndex = 1
			loadedIndex = 2
			activeIndex = 3
			subStateIndex = 4
			descStartIndex = 5
		} else {
			nameIndex = 0
			loadedIndex = 1
			activeIndex = 2
			subStateIndex = 3
			descStartIndex = 4
		}

		if fields[nameIndex] == "UNIT" || fields[nameIndex] == "LOAD" || fields[nameIndex] == "ACTIVE" || fields[nameIndex] == "SUB" {
			continue
		}

		service := ServiceInfo{
			Name:     fields[nameIndex],
			Loaded:   fields[loadedIndex],
			Active:   fields[activeIndex],
			SubState: fields[subStateIndex],
			Type:     "systemd",
		}

		// Set status based on active state
		if service.Active == "active" {
			service.Status = "running"
		} else if service.Active == "inactive" {
			service.Status = "stopped"
		} else if service.Active == "failed" {
			service.Status = "failed"
		} else {
			service.Status = "unknown"
		}

		// Description is rest of the fields
		if len(fields) > descStartIndex {
			service.Description = strings.Join(fields[descStartIndex:], " ")
		}

		services = append(services, service)
	}

	return services, nil
}

// GetServiceDetails gets detailed information for a specific service
func (r *ServicesReader) GetServiceDetails(name string) (*ServiceInfo, error) {
	cmd := exec.Command(r.systemctlPath, "show", name, "--no-pager")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get service details: %w", err)
	}

	details := &ServiceInfo{
		Name: name,
		Type: "systemd",
	}

	// Parse systemctl show output
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		switch key {
		case "Description":
			details.Description = value
		case "MainPID":
			if pid, err := strconv.Atoi(value); err == nil && pid > 0 {
				details.PID = pid
				// Get process info
				r.populateProcessInfo(details, pid)
			}
		case "LoadState":
			details.Loaded = value
		case "ActiveState":
			details.Active = value
		case "SubState":
			details.SubState = value
		case "ExecStart":
			// Parse command
			details.Command = r.parseExecStart(value)
		}
	}

	// Set status based on active state
	if details.Active == "active" {
		details.Status = "running"
	} else if details.Active == "inactive" {
		details.Status = "stopped"
	} else if details.Active == "failed" {
		details.Status = "failed"
	} else {
		details.Status = "unknown"
	}

	return details, nil
}

// parseExecStart parses ExecStart value from systemctl
func (r *ServicesReader) parseExecStart(value string) string {
	// ExecStart format: path={path} argv0={argv0} argv1={argv1} ...
	// or just the command path
	if strings.HasPrefix(value, "path=") {
		// Parse systemd format
		parts := strings.Fields(value)
		if len(parts) > 0 {
			return strings.TrimPrefix(parts[0], "path=")
		}
	}
	return value
}

// populateProcessInfo populates process information for a PID
func (r *ServicesReader) populateProcessInfo(service *ServiceInfo, pid int) {
	// Get user
	if user, err := r.getProcessUser(pid); err == nil {
		service.User = user
	}

	// Get memory
	if memMB, err := r.getProcessMemoryMB(pid); err == nil {
		service.MemoryMB = memMB
	}

	// Get CPU
	if cpuPercent, err := r.getProcessCPUPercent(pid); err == nil {
		service.CPUPercent = cpuPercent
	}

	// Get start time and uptime
	if startTime, uptime, err := r.getProcessTime(pid); err == nil {
		service.StartTime = startTime
		service.Uptime = uptime
	}

	// Get command line
	if cmdline, err := r.getProcessCmdline(pid); err == nil {
		if service.Command == "" {
			service.Command = cmdline
		}
	}
}

// getProcessUser gets the user running the process
func (r *ServicesReader) getProcessUser(pid int) (string, error) {
	statusPath := filepath.Join(r.procPath, strconv.Itoa(pid), "status")
	file, err := os.Open(statusPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Uid:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				uid := fields[1]
				// Try to get username from /etc/passwd
				if username, err := r.getUsernameFromUID(uid); err == nil {
					return username, nil
				}
				return uid, nil
			}
		}
	}

	return "", fmt.Errorf("user not found")
}

// getUsernameFromUID gets username from UID
func (r *ServicesReader) getUsernameFromUID(uid string) (string, error) {
	file, err := os.Open("/etc/passwd")
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ":")
		if len(fields) >= 3 && fields[2] == uid {
			return fields[0], nil
		}
	}

	return "", fmt.Errorf("username not found")
}

// getProcessMemoryMB gets memory usage in MB
func (r *ServicesReader) getProcessMemoryMB(pid int) (float64, error) {
	statmPath := filepath.Join(r.procPath, strconv.Itoa(pid), "statm")
	data, err := os.ReadFile(statmPath)
	if err != nil {
		return 0, err
	}

	fields := strings.Fields(string(data))
	if len(fields) < 2 {
		return 0, fmt.Errorf("invalid statm format")
	}

	// Resident set size in pages
	rss, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return 0, err
	}

	// Get page size
	pageSize := float64(os.Getpagesize())

	// Convert to MB
	return (rss * pageSize) / (1024 * 1024), nil
}

// getProcessCPUPercent gets CPU usage percentage
func (r *ServicesReader) getProcessCPUPercent(pid int) (float64, error) {
	statPath := filepath.Join(r.procPath, strconv.Itoa(pid), "stat")
	data, err := os.ReadFile(statPath)
	if err != nil {
		return 0, err
	}

	fields := strings.Fields(string(data))
	if len(fields) < 17 {
		return 0, fmt.Errorf("invalid stat format")
	}

	// Parse CPU times: utime, stime, cutime, cstime
	utime, _ := strconv.ParseFloat(fields[13], 64)
	stime, _ := strconv.ParseFloat(fields[14], 64)
	cutime, _ := strconv.ParseFloat(fields[15], 64)
	cstime, _ := strconv.ParseFloat(fields[16], 64)

	totalTime := utime + stime + cutime + cstime

	// Get system uptime
	uptime, err := r.getSystemUptime()
	if err != nil {
		return 0, err
	}

	// Get process start time
	startTime, err := strconv.ParseFloat(fields[21], 64)
	if err != nil {
		return 0, err
	}

	// Calculate CPU percentage
	elapsed := uptime - (startTime / float64(os.Getpagesize()/os.Getpagesize()))
	if elapsed > 0 {
		return (totalTime / elapsed) * 100, nil
	}

	return 0, nil
}

// getSystemUptime gets system uptime in seconds
func (r *ServicesReader) getSystemUptime() (float64, error) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, err
	}

	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return 0, fmt.Errorf("invalid uptime format")
	}

	return strconv.ParseFloat(fields[0], 64)
}

// getProcessTime gets process start time and uptime
func (r *ServicesReader) getProcessTime(pid int) (string, string, error) {
	statPath := filepath.Join(r.procPath, strconv.Itoa(pid), "stat")
	data, err := os.ReadFile(statPath)
	if err != nil {
		return "", "", err
	}

	fields := strings.Fields(string(data))
	if len(fields) < 22 {
		return "", "", fmt.Errorf("invalid stat format")
	}

	startTime, err := strconv.ParseInt(fields[21], 10, 64)
	if err != nil {
		return "", "", err
	}

	// Convert start time to seconds
	startTimeSec := startTime / int64(os.Getpagesize())

	// Get system uptime
	uptime, err := r.getSystemUptime()
	if err != nil {
		return "", "", err
	}

	// Calculate process uptime
	elapsed := uptime - float64(startTimeSec)

	// Format start time
	startTimeUnix := time.Now().Unix() - int64(elapsed)
	startTimeStr := time.Unix(startTimeUnix, 0).Format("Jan 02 15:04:05")

	// Format uptime
	uptimeStr := formatDuration(elapsed)

	return startTimeStr, uptimeStr, nil
}

// formatDuration formats seconds in a human-readable format
func formatDuration(seconds float64) string {
	duration := time.Duration(seconds * float64(time.Second))

	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%d days, %d hours, %d minutes", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%d hours, %d minutes", hours, minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%d minutes", minutes)
	}
	return "< 1 minute"
}

// getProcessCmdline gets the command line for a process
func (r *ServicesReader) getProcessCmdline(pid int) (string, error) {
	cmdlinePath := filepath.Join(r.procPath, strconv.Itoa(pid), "cmdline")
	data, err := os.ReadFile(cmdlinePath)
	if err != nil {
		return "", err
	}

	// Replace null bytes with spaces
	cmdline := strings.ReplaceAll(string(data), "\x00", " ")
	return strings.TrimSpace(cmdline), nil
}

// ReadListeningPorts reads all listening ports from /proc/net
func (r *ServicesReader) ReadListeningPorts() ([]PortInfo, error) {
	var ports []PortInfo

	// Read TCP ports
	tcpPorts, err := r.parseProcNet("tcp")
	if err == nil {
		ports = append(ports, tcpPorts...)
	}

	// Read TCP6 ports
	tcp6Ports, err := r.parseProcNet("tcp6")
	if err == nil {
		ports = append(ports, tcp6Ports...)
	}

	// Read UDP ports
	udpPorts, err := r.parseProcNet("udp")
	if err == nil {
		ports = append(ports, udpPorts...)
	}

	// Read UDP6 ports
	udp6Ports, err := r.parseProcNet("udp6")
	if err == nil {
		ports = append(ports, udp6Ports...)
	}

	return ports, nil
}

// parseProcNet parses /proc/net/{tcp,tcp6,udp,udp6}
func (r *ServicesReader) parseProcNet(protocol string) ([]PortInfo, error) {
	var ports []PortInfo

	path := filepath.Join(r.procNetPath, protocol)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip header
		if strings.HasPrefix(line, "  sl") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		// Parse local address (format: local_addr:port)
		localAddr := fields[1]
		addrParts := strings.Split(localAddr, ":")
		if len(addrParts) != 2 {
			continue
		}

		// Parse port (hex)
		portHex := addrParts[1]
		port, err := strconv.ParseInt(portHex, 16, 32)
		if err != nil {
			continue
		}

		// Parse state
		state := fields[3]
		stateNum, _ := strconv.Atoi(state)
		stateStr := r.getSocketState(stateNum)

		// Only include listening sockets
		if stateNum != 10 && stateNum != 11 { // TCP_LISTEN = 10, TCP_CLOSE = 11
			continue
		}

		// Parse address
		addrHex := addrParts[0]
		ip := r.parseIPAddr(addrHex, protocol)

		// Get inode
		inode := fields[9]

		// Get process info from inode
		processName, pid := r.getProcessFromInode(inode)

		portInfo := PortInfo{
			Protocol:    protocol,
			LocalAddr:   fmt.Sprintf("%s:%d", ip, port),
			Port:        int(port),
			State:       stateStr,
			ProcessName: processName,
			PID:         pid,
		}

		ports = append(ports, portInfo)
	}

	return ports, nil
}

// parseIPAddr parses IP address from hex format
func (r *ServicesReader) parseIPAddr(hexAddr string, protocol string) string {
	// IPv4: 0100007F -> 127.0.0.1
	// IPv6: 00000000000000000000000000000000 -> ::

	if strings.Contains(protocol, "6") {
		// IPv6 - 32 hex characters
		if len(hexAddr) < 32 {
			return "0.0.0.0"
		}
		// Convert to IPv6 format
		var parts []string
		for i := 0; i < 32; i += 4 {
			part := hexAddr[i : i+4]
			// Convert to integer
			val, _ := strconv.ParseInt(part, 16, 32)
			parts = append(parts, fmt.Sprintf("%02x", val))
		}
		// Format as IPv6
		ipv6 := strings.Join(parts, ":")
		return ipv6
	}

	// IPv4 - 8 hex characters
	if len(hexAddr) < 8 {
		return "0.0.0.0"
	}

	// Parse as little-endian
	val, _ := strconv.ParseUint(hexAddr, 16, 32)
	ip := make(net.IP, 4)
	ip[0] = byte(val & 0xFF)
	ip[1] = byte((val >> 8) & 0xFF)
	ip[2] = byte((val >> 16) & 0xFF)
	ip[3] = byte((val >> 24) & 0xFF)

	return ip.String()
}

// getSocketState returns socket state string from state number
func (r *ServicesReader) getSocketState(state int) string {
	states := map[int]string{
		1:  "ESTABLISHED",
		2:  "SYN_SENT",
		3:  "SYN_RECV",
		4:  "FIN_WAIT1",
		5:  "FIN_WAIT2",
		6:  "TIME_WAIT",
		7:  "CLOSE",
		8:  "CLOSE_WAIT",
		9:  "LAST_ACK",
		10: "LISTEN",
		11: "CLOSING",
	}

	if stateStr, ok := states[state]; ok {
		return stateStr
	}
	return "UNKNOWN"
}

// getProcessFromInode gets process name and PID from socket inode
func (r *ServicesReader) getProcessFromInode(inode string) (string, int) {
	// Search /proc/[pid]/fd for the inode
	entries, err := os.ReadDir(r.procPath)
	if err != nil {
		return "", 0
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		// Check /proc/[pid]/fd
		fdPath := filepath.Join(r.procPath, entry.Name(), "fd")
		fdEntries, err := os.ReadDir(fdPath)
		if err != nil {
			continue
		}

		for _, fdEntry := range fdEntries {
			linkPath := filepath.Join(fdPath, fdEntry.Name())
			linkTarget, err := os.Readlink(linkPath)
			if err != nil {
				continue
			}

			// Check if link contains the inode
			if strings.Contains(linkTarget, "socket:["+inode+"]") {
				// Get process name
				cmdline, _ := r.getProcessCmdline(pid)
				if cmdline == "" {
					commPath := filepath.Join(r.procPath, entry.Name(), "comm")
					if commData, err := os.ReadFile(commPath); err == nil {
						cmdline = strings.TrimSpace(string(commData))
					}
				}
				return cmdline, pid
			}
		}
	}

	return "", 0
}

// MatchPortsToServices matches ports to services by PID
func (r *ServicesReader) MatchPortsToServices(services []ServiceInfo, ports []PortInfo) {
	for i := range services {
		if services[i].PID <= 0 {
			continue
		}

		for _, port := range ports {
			if port.PID == services[i].PID {
				services[i].ListeningPorts = append(services[i].ListeningPorts, port)
			}
		}
	}
}

// calculateSummary calculates summary statistics
func (r *ServicesReader) calculateSummary(summary *ServicesSummary) {
	summary.TotalServices = len(summary.Services)

	for _, service := range summary.Services {
		switch service.Status {
		case "running":
			summary.Running++
		case "stopped":
			summary.Stopped++
		case "failed":
			summary.Failed++
		}
	}
}

// FilterServices filters services based on criteria
func (r *ServicesReader) FilterServices(services []ServiceInfo, name, status, user string, port int) []ServiceInfo {
	var filtered []ServiceInfo

	for _, service := range services {
		// Filter by name
		if name != "" {
			if !strings.Contains(strings.ToLower(service.Name), strings.ToLower(name)) {
				continue
			}
		}

		// Filter by status
		if status != "" {
			if service.Status != status {
				continue
			}
		}

		// Filter by user
		if user != "" {
			if service.User != user {
				continue
			}
		}

		// Filter by port
		if port > 0 {
			hasPort := false
			for _, p := range service.ListeningPorts {
				if p.Port == port {
					hasPort = true
					break
				}
			}
			if !hasPort {
				continue
			}
		}

		filtered = append(filtered, service)
	}

	return filtered
}

// ReadServiceByName reads a specific service by name
func (r *ServicesReader) ReadServiceByName(name string) (*ServiceInfo, error) {
	details, err := r.GetServiceDetails(name)
	if err != nil {
		return nil, err
	}

	// Read listening ports
	ports, err := r.ReadListeningPorts()
	if err == nil && details.PID > 0 {
		for _, port := range ports {
			if port.PID == details.PID {
				details.ListeningPorts = append(details.ListeningPorts, port)
			}
		}
	}

	return details, nil
}
