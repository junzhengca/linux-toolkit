package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/jun/linux-toolkit/pkg/battery"
	"github.com/jun/linux-toolkit/pkg/cpu"
	"github.com/jun/linux-toolkit/pkg/disk"
	"github.com/jun/linux-toolkit/pkg/gpu"
	"github.com/jun/linux-toolkit/pkg/services"
)

type APIResponse struct {
	Status    string      `json:"status"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp string      `json:"timestamp"`
}

type APIError struct {
	Status string `json:"status"`
	Error  string `json:"error"`
	Code   string `json:"code"`
}

func respondJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(APIResponse{
		Status: "success",
		Data:   data,
	})
}

func respondError(w http.ResponseWriter, message string, code int, errCode string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(APIError{
		Status: "error",
		Error:  message,
		Code:   errCode,
	})
}

func (s *Server) healthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, "Method not allowed", http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED")
			return
		}

		respondJSON(w, map[string]string{"message": "Server is running"}, http.StatusOK)
	}
}

func (s *Server) handleDisk() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, "Method not allowed", http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED")
			return
		}

		query := r.URL.Query()
		deviceName := query.Get("device")
		mountPoint := query.Get("mount")
		showAll := query.Get("all") == "true"
		includeIO := query.Get("io-stats") == "true"
		includeInode := query.Get("inode-stats") == "true"

		reader := disk.NewDiskReader()

		if showAll {
			mountedDisks, err := reader.ListMountedDisks()
			if err != nil {
				respondError(w, fmt.Sprintf("Failed to list disks: %v", err), http.StatusInternalServerError, "INTERNAL_ERROR")
				return
			}

			var disks []*disk.DiskInfo
			for _, devicePath := range mountedDisks {
				diskInfo, err := reader.ReadDisk(devicePath)
				if err != nil {
					continue
				}
				disks = append(disks, diskInfo)
			}

			respondJSON(w, disks, http.StatusOK)
			return
		}

		var devicePath string
		if deviceName != "" {
			if !strings.HasPrefix(deviceName, "/dev/") {
				devicePath = "/dev/" + deviceName
			} else {
				devicePath = deviceName
			}
		} else if mountPoint != "" {
			mountedDisks, err := reader.ListMountedDisks()
			if err != nil {
				respondError(w, fmt.Sprintf("Failed to read mounts: %v", err), http.StatusInternalServerError, "INTERNAL_ERROR")
				return
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
				respondError(w, fmt.Sprintf("No device found mounted at %s", mountPoint), http.StatusNotFound, "NOT_FOUND")
				return
			}
		} else {
			mountedDisks, err := reader.ListMountedDisks()
			if err != nil {
				respondError(w, fmt.Sprintf("Failed to find mounted disks: %v", err), http.StatusInternalServerError, "INTERNAL_ERROR")
				return
			}
			if len(mountedDisks) == 0 {
				respondError(w, "No disks found", http.StatusNotFound, "NOT_FOUND")
				return
			}
			devicePath = mountedDisks[0]
		}

		diskInfo, err := reader.ReadDisk(devicePath)
		if err != nil {
			respondError(w, fmt.Sprintf("Failed to read disk information: %v", err), http.StatusInternalServerError, "INTERNAL_ERROR")
			return
		}

		response := map[string]interface{}{
			"disk": diskInfo,
		}

		if includeIO {
			ioStats, err := reader.ReadIOStats(devicePath)
			if err != nil {
				response["ioStats"] = nil
			} else {
				response["ioStats"] = ioStats
			}
		}

		if includeInode && diskInfo.MountPoint != "" {
			inodeInfo, err := reader.ReadInodeInfo(diskInfo.MountPoint)
			if err != nil {
				response["inodeInfo"] = nil
			} else {
				response["inodeInfo"] = inodeInfo
			}
		}

		respondJSON(w, response, http.StatusOK)
	}
}

func (s *Server) handleCPU() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, "Method not allowed", http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED")
			return
		}

		query := r.URL.Query()
		showCores := query.Get("show-cores") == "true"
		showFlags := query.Get("show-flags") == "true"
		showTemp := query.Get("show-temp") == "true"

		reader := cpu.NewCPUReader()
		cpuInfo, err := reader.ReadCPU()
		if err != nil {
			respondError(w, fmt.Sprintf("Failed to read CPU information: %v", err), http.StatusInternalServerError, "INTERNAL_ERROR")
			return
		}

		if !showCores {
			cpuInfo.Cores = nil
		}

		if !showFlags {
			cpuInfo.Flags = nil
		}

		if !showTemp {
			cpuInfo.CPUTemperature = 0
		}

		respondJSON(w, cpuInfo, http.StatusOK)
	}
}

func (s *Server) handleCores() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, "Method not allowed", http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED")
			return
		}

		reader := cpu.NewCPUReader()
		cpuInfo, err := reader.ReadCPU()
		if err != nil {
			respondError(w, fmt.Sprintf("Failed to read CPU information: %v", err), http.StatusInternalServerError, "INTERNAL_ERROR")
			return
		}

		respondJSON(w, cpuInfo.Cores, http.StatusOK)
	}
}

func (s *Server) handleGPU() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, "Method not allowed", http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED")
			return
		}

		query := r.URL.Query()
		cardName := query.Get("card")
		showConnectors := query.Get("show-connectors") == "true"
		showModes := query.Get("show-modes") == "true"

		reader := gpu.NewGPUReader()

		if cardName == "" {
			cards, err := reader.ListGPUs()
			if err != nil {
				respondError(w, fmt.Sprintf("Failed to list GPUs: %v", err), http.StatusInternalServerError, "INTERNAL_ERROR")
				return
			}

			var gpus []*gpu.GPUInfo
			for _, card := range cards {
				gpuInfo, err := reader.ReadGPU(card)
				if err != nil {
					continue
				}

				if !showConnectors {
					gpuInfo.Connectors = nil
				}

				if !showModes {
					gpuInfo.Modes = nil
				}

				gpus = append(gpus, gpuInfo)
			}

			respondJSON(w, gpus, http.StatusOK)
			return
		}

		gpuInfo, err := reader.ReadGPU(cardName)
		if err != nil {
			respondError(w, fmt.Sprintf("Failed to read GPU information: %v", err), http.StatusInternalServerError, "INTERNAL_ERROR")
			return
		}

		if !showConnectors {
			gpuInfo.Connectors = nil
		}

		if !showModes {
			gpuInfo.Modes = nil
		}

		respondJSON(w, gpuInfo, http.StatusOK)
	}
}

func (s *Server) handleBattery() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, "Method not allowed", http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED")
			return
		}

		reader := battery.NewBatteryReader()
		batteries, err := reader.ListBatteries()
		if err != nil {
			respondError(w, fmt.Sprintf("Failed to list batteries: %v", err), http.StatusInternalServerError, "INTERNAL_ERROR")
			return
		}

		if len(batteries) == 0 {
			respondError(w, "No batteries found", http.StatusNotFound, "NOT_FOUND")
			return
		}

		batteryInfo, err := reader.ReadBattery(batteries[0])
		if err != nil {
			respondError(w, fmt.Sprintf("Failed to read battery information: %v", err), http.StatusInternalServerError, "INTERNAL_ERROR")
			return
		}

		respondJSON(w, batteryInfo, http.StatusOK)
	}
}

func (s *Server) handleServices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, "Method not allowed", http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED")
			return
		}

		query := r.URL.Query()
		serviceName := query.Get("name")
		serviceStatus := query.Get("status")
		serviceUser := query.Get("user")
		portStr := query.Get("port")

		reader := services.NewServicesReader()

		if serviceName != "" {
			serviceInfo, err := reader.GetServiceDetails(serviceName)
			if err != nil {
				respondError(w, fmt.Sprintf("Failed to read service: %v", err), http.StatusInternalServerError, "INTERNAL_ERROR")
				return
			}
			respondJSON(w, serviceInfo, http.StatusOK)
			return
		}

		summary, err := reader.ReadAllServices()
		if err != nil {
			respondError(w, fmt.Sprintf("Failed to read services: %v", err), http.StatusInternalServerError, "INTERNAL_ERROR")
			return
		}

		var port int
		if portStr != "" {
			port, err = strconv.Atoi(portStr)
			if err != nil {
				respondError(w, "Invalid port number", http.StatusBadRequest, "INVALID_PARAMETER")
				return
			}
		}

		filteredServices := reader.FilterServices(
			summary.Services,
			serviceName,
			serviceStatus,
			serviceUser,
			port,
		)

		summary.Services = filteredServices

		running := 0
		stopped := 0
		failed := 0
		for _, s := range filteredServices {
			switch s.Status {
			case "running":
				running++
			case "stopped":
				stopped++
			case "failed":
				failed++
			}
		}

		summary.Running = running
		summary.Stopped = stopped
		summary.Failed = failed

		respondJSON(w, summary, http.StatusOK)
	}
}

func (s *Server) handleSummary() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, "Method not allowed", http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED")
			return
		}

		summary := make(map[string]interface{})

		cpuReader := cpu.NewCPUReader()
		cpuInfo, err := cpuReader.ReadCPU()
		if err == nil {
			summary["cpu"] = cpuInfo
		}

		diskReader := disk.NewDiskReader()
		mountedDisks, err := diskReader.ListMountedDisks()
		if err == nil && len(mountedDisks) > 0 {
			var disks []*disk.DiskInfo
			for _, devicePath := range mountedDisks {
				diskInfo, err := diskReader.ReadDisk(devicePath)
				if err == nil {
					disks = append(disks, diskInfo)
				}
			}
			summary["disks"] = disks
		}

		gpuReader := gpu.NewGPUReader()
		cards, err := gpuReader.ListGPUs()
		if err == nil && len(cards) > 0 {
			var gpus []*gpu.GPUInfo
			for _, card := range cards {
				gpuInfo, err := gpuReader.ReadGPU(card)
				if err == nil {
					gpus = append(gpus, gpuInfo)
				}
			}
			summary["gpus"] = gpus
		}

		batteryReader := battery.NewBatteryReader()
		batteries, _ := batteryReader.ListBatteries()
		if len(batteries) > 0 {
			batteryInfo, err := batteryReader.ReadBattery(batteries[0])
			if err == nil {
				summary["battery"] = batteryInfo
			}
		}

		servicesReader := services.NewServicesReader()
		servicesSummary, err := servicesReader.ReadAllServices()
		if err == nil {
			summary["services"] = servicesSummary
		}

		respondJSON(w, summary, http.StatusOK)
	}
}
