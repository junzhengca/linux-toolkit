# Linux Toolkit API Documentation

Complete REST API reference for the Linux Toolkit HTTP server.

## Table of Contents

- [Overview](#overview)
- [Base URL](#base-url)
- [Response Format](#response-format)
- [Endpoints](#endpoints)
  - [Health Check](#health-check)
  - [CPU Information](#cpu-information)
  - [CPU Cores](#cpu-cores)
  - [Disk Information](#disk-information)
  - [GPU Information](#gpu-information)
  - [Battery Information](#battery-information)
  - [Services Information](#services-information)
  - [System Summary](#system-summary)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)
- [Best Practices](#best-practices)

## Overview

The Linux Toolkit API provides programmatic access to system metrics including CPU, disk, GPU, battery, and services information. All endpoints return JSON data with a consistent response structure.

### Features

- **RESTful design** with predictable resource-oriented URLs
- **JSON responses** for easy integration with any programming language
- **No authentication required** by default (suitable for local monitoring)
- **Configurable query parameters** for filtering and customization
- **Error responses** with descriptive messages and codes

## Base URL

```
http://localhost:8080/api/v1
```

Replace `localhost:8080` with your server's actual address and port.

## Response Format

All API responses follow a consistent JSON structure:

### Success Response

```json
{
  "status": "success",
  "data": {
    /* Actual metric data */
  },
  "timestamp": "2026-01-20T12:00:00Z"
}
```

### Error Response

```json
{
  "status": "error",
  "error": "Error message description",
  "code": "ERROR_CODE",
  "timestamp": "2026-01-20T12:00:00Z"
}
```

### HTTP Status Codes

| Code | Description |
|------|-------------|
| `200 OK` | Request succeeded |
| `400 Bad Request` | Invalid parameters provided |
| `404 Not Found` | Resource not found |
| `405 Method Not Allowed` | HTTP method not supported for endpoint |
| `500 Internal Server Error` | Server error occurred |

---

## Endpoints

### Health Check

Check if the server is running and responding.

**Endpoint:** `GET /api/v1/health`

**Description:** Returns a simple health check message to verify server availability.

**Query Parameters:** None

**Response:**

```json
{
  "status": "success",
  "data": {
    "message": "Server is running"
  },
  "timestamp": "2026-01-20T12:00:00Z"
}
```

**Example:**

```bash
curl http://localhost:8080/api/v1/health
```

---

### CPU Information

Retrieve detailed CPU information including model, architecture, cores, cache, temperature, and load averages.

**Endpoint:** `GET /api/v1/cpu`

**Description:** Returns comprehensive CPU metrics with optional details.

**Query Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|-------|----------|----------|-------------|
| `show-cores` | boolean | No | false | Include per-core usage details |
| `show-flags` | boolean | No | false | Include CPU flags/features list |
| `show-temp` | boolean | No | false | Include CPU temperature |

**Response Schema:**

```json
{
  "status": "success",
  "data": {
    "modelName": "string",
    "vendorId": "string",
    "architecture": "string",
    "cpuMode": "string",
    "cpuFamily": number,
    "model": number,
    "stepping": number,
    "physicalCores": number,
    "logicalCores": number,
    "threadsPerCore": number,
    "sockets": number,
    "minFrequency": number,
    "maxFrequency": number,
    "currentFrequency": number,
    "bogomips": number,
    "flags": ["string"],
    "cacheL1d": number,
    "cacheL1i": number,
    "cacheL2": number,
    "cacheL3": number,
    "cpuTemperature": number,
    "loadAvg1": number,
    "loadAvg5": number,
    "loadAvg15": number,
    "totalProcesses": number,
    "runningProcesses": number,
    "cores": [
      {
        "coreId": number,
        "physicalId": number,
        "processorId": number,
        "frequency": number,
        "usagePercent": number
      }
    ]
  },
  "timestamp": "2026-01-20T12:00:00Z"
}
```

**Field Descriptions:**

| Field | Type | Description |
|-------|-------|-------------|
| `modelName` | string | CPU model name (e.g., "AMD Ryzen 7 5800X") |
| `vendorId` | string | Vendor identifier (e.g., "AuthenticAMD", "GenuineIntel") |
| `architecture` | string | CPU architecture (e.g., "x86_64", "arm64") |
| `cpuMode` | string | CPU mode (e.g., "64-bit") |
| `physicalCores` | number | Number of physical CPU cores |
| `logicalCores` | number | Number of logical CPU cores (including hyperthreading) |
| `threadsPerCore` | number | Threads per core (usually 1 or 2) |
| `minFrequency` | number | Minimum CPU frequency in Hz |
| `maxFrequency` | number | Maximum CPU frequency in Hz |
| `currentFrequency` | number | Current CPU frequency in Hz |
| `cpuTemperature` | number | Current CPU temperature in Celsius (0 if unavailable) |
| `loadAvg1` | number | 1-minute load average |
| `loadAvg5` | number | 5-minute load average |
| `loadAvg15` | number | 15-minute load average |
| `totalProcesses` | number | Total number of processes |
| `runningProcesses` | number | Number of currently running processes |
| `cores` | array | Array of per-core details (only if show-cores=true) |
| `flags` | array | Array of CPU feature flags (only if show-flags=true) |

**Examples:**

```bash
# Basic CPU information
curl http://localhost:8080/api/v1/cpu

# CPU with per-core usage and temperature
curl "http://localhost:8080/api/v1/cpu?show-cores=true&show-temp=true"

# CPU with all details
curl "http://localhost:8080/api/v1/cpu?show-cores=true&show-flags=true&show-temp=true"
```

---

### CPU Cores

Retrieve per-core CPU usage details only.

**Endpoint:** `GET /api/v1/cpu/cores`

**Description:** Returns an array of CPU core information with usage and frequency.

**Query Parameters:** None

**Response Schema:**

```json
{
  "status": "success",
  "data": [
    {
      "coreId": 0,
      "physicalId": 0,
      "processorId": 0,
      "frequency": number,
      "usagePercent": number
    }
  ],
  "timestamp": "2026-01-20T12:00:00Z"
}
```

**Field Descriptions:**

| Field | Type | Description |
|-------|-------|-------------|
| `coreId` | number | Core identifier |
| `physicalId` | number | Physical CPU socket ID |
| `processorId` | number | Processor ID (logical core) |
| `frequency` | number | Core frequency in Hz |
| `usagePercent` | number | Core usage percentage (0-100) |

**Example:**

```bash
curl http://localhost:8080/api/v1/cpu/cores
```

---

### Disk Information

Retrieve disk information including size, usage, filesystem type, I/O statistics, and inode information.

**Endpoint:** `GET /api/v1/disk`

**Description:** Returns disk metrics with optional details. Can return single disk or all disks.

**Query Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|-------|----------|----------|-------------|
| `device` | string | No | auto-detect | Disk device name (e.g., "sda1", "nvme0n1p1") |
| `mount` | string | No | - | Filter by mount point (e.g., "/", "/home") |
| `all` | boolean | No | false | Show all mounted disks |
| `io-stats` | boolean | No | false | Include I/O statistics |
| `inode-stats` | boolean | No | false | Include inode statistics |

**Note:** If `all=true`, the response structure changes (see "All Disks" response below).

**Response Schema (Single Disk):**

```json
{
  "status": "success",
  "data": {
    "disk": {
      "name": "string",
      "path": "string",
      "size": number,
      "used": number,
      "available": number,
      "usagePercent": number,
      "filesystemType": "string",
      "mountPoint": "string",
      "uuid": "string",
      "label": "string",
      "blockSize": number,
      "model": "string",
      "serial": "string",
      "readOnly": boolean,
      "removable": boolean,
      "rotational": boolean
    },
    "ioStats": {
      "readsCompleted": number,
      "readsMerged": number,
      "sectorsRead": number,
      "readTimeMs": number,
      "writesCompleted": number,
      "writesMerged": number,
      "sectorsWritten": number,
      "writeTimeMs": number,
      "ioInProgress": number,
      "ioTimeMs": number,
      "weightedIOTimeMs": number
    },
    "inodeInfo": {
      "total": number,
      "used": number,
      "free": number,
      "usagePercent": number
    }
  },
  "timestamp": "2026-01-20T12:00:00Z"
}
```

**Response Schema (All Disks - when all=true):**

```json
{
  "status": "success",
  "data": {
    "data": [
      {
        "name": "string",
        "path": "string",
        "size": number,
        "used": number,
        "available": number,
        "usagePercent": number,
        "filesystemType": "string",
        "mountPoint": "string",
        "uuid": "string",
        "label": "string",
        "blockSize": number,
        "model": "string",
        "serial": "string",
        "readOnly": boolean,
        "removable": boolean,
        "rotational": boolean
      }
    ],
    "ioStats": {
      "sda1": {
        "readsCompleted": number,
        "readsMerged": number,
        "sectorsRead": number,
        "readTimeMs": number,
        "writesCompleted": number,
        "writesMerged": number,
        "sectorsWritten": number,
        "writeTimeMs": number,
        "ioInProgress": number,
        "ioTimeMs": number,
        "weightedIOTimeMs": number
      }
    },
    "inodeInfo": {
      "sda1": {
        "total": number,
        "used": number,
        "free": number,
        "usagePercent": number
      }
    }
  },
  "timestamp": "2026-01-20T12:00:00Z"
}
```

**Field Descriptions:**

| Field | Type | Description |
|-------|-------|-------------|
| `name` | string | Disk device name (e.g., "sda1") |
| `path` | string | Full device path (e.g., "/dev/sda1") |
| `size` | number | Total disk size in bytes |
| `used` | number | Used disk space in bytes |
| `available` | number | Available disk space in bytes |
| `usagePercent` | number | Disk usage percentage (0-100) |
| `filesystemType` | string | Filesystem type (e.g., "ext4", "xfs", "ntfs") |
| `mountPoint` | string | Mount point path (empty if not mounted) |
| `uuid` | string | Filesystem UUID |
| `label` | string | Filesystem label |
| `blockSize` | number | Filesystem block size in bytes |
| `model` | string | Disk model name |
| `serial` | string | Disk serial number |
| `readOnly` | boolean | Whether disk is read-only |
| `removable` | boolean | Whether disk is removable |
| `rotational` | boolean | Whether disk is rotational (HDD) or non-rotational (SSD) |
| `ioStats` | object | I/O statistics (only if io-stats=true) |
| `inodeInfo` | object | Inode usage statistics (only if inode-stats=true) |

**I/O Statistics Fields:**

| Field | Type | Description |
|-------|-------|-------------|
| `readsCompleted` | number | Number of reads completed successfully |
| `readsMerged` | number | Number of reads merged |
| `sectorsRead` | number | Number of sectors read |
| `readTimeMs` | number | Total read time in milliseconds |
| `writesCompleted` | number | Number of writes completed successfully |
| `writesMerged` | number | Number of writes merged |
| `sectorsWritten` | number | Number of sectors written |
| `writeTimeMs` | number | Total write time in milliseconds |
| `ioInProgress` | number | Number of I/O operations in progress |
| `ioTimeMs` | number | Total I/O time in milliseconds |
| `weightedIOTimeMs` | number | Weighted I/O time in milliseconds |

**Examples:**

```bash
# Get first mounted disk
curl http://localhost:8080/api/v1/disk

# Get specific disk
curl "http://localhost:8080/api/v1/disk?device=nvme0n1p1"

# Get disk by mount point
curl "http://localhost:8080/api/v1/disk?mount=/"

# Get disk with I/O and inode stats
curl "http://localhost:8080/api/v1/disk?device=sda1&io-stats=true&inode-stats=true"

# Get all mounted disks
curl "http://localhost:8080/api/v1/disk?all=true"

# Get all disks with full statistics
curl "http://localhost:8080/api/v1/disk?all=true&io-stats=true&inode-stats=true"
```

---

### GPU Information

Retrieve GPU information including device name, vendor, driver, memory, clocks, temperature, power, and utilization.

**Endpoint:** `GET /api/v1/gpu`

**Description:** Returns GPU metrics with optional details. Can return single GPU or all GPUs.

**Query Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|-------|----------|----------|-------------|
| `card` | string | No | auto-detect | GPU card name (e.g., "card0", "card1") |
| `show-connectors` | boolean | No | false | Include connector information |
| `show-modes` | boolean | No | false | Include supported display modes |

**Note:** If `card` parameter is omitted, returns an array of all GPUs.

**Response Schema (Single GPU):**

```json
{
  "status": "success",
  "data": {
    "cardName": "string",
    "deviceName": "string",
    "driver": "string",
    "vendor": "string",
    "vendorId": "string",
    "deviceId": "string",
    "subsystemVendor": "string",
    "subsystemDevice": "string",
    "class": "string",
    "revision": "string",
    "enabled": boolean,
    "status": "string",
    "connectors": ["string"],
    "enabledConnectors": ["string"],
    "vramSize": number,
    "vramUsed": number,
    "vramFree": number,
    "gartSize": number,
    "coreClock": number,
    "memoryClock": number,
    "maxCoreClock": number,
    "maxMemoryClock": number,
    "powerUsage": number,
    "powerLimit": number,
    "temperature": number,
    "temperatureCrit": number,
    "fanSpeed": number,
    "fanSpeedPercent": number,
    "gpuUtilPercent": number,
    "memoryUtilPercent": number,
    "busId": "string",
    "busWidth": "string",
    "pcieGen": "string",
    "maxPcieGen": "string",
    "vbiosVersion": "string",
    "firmwareVersion": "string",
    "devicePath": "string",
    "sysfsPath": "string",
    "isPrimary": boolean,
    "gpuType": "string",
    "computeUnits": number,
    "cudaCores": number,
    "shaders": number,
    "textureUnits": number,
    "rops": number,
    "modes": [
      {
        "width": number,
        "height": number,
        "refresh": number,
        "bits": number
      }
    ]
  },
  "timestamp": "2026-01-20T12:00:00Z"
}
```

**Field Descriptions:**

| Field | Type | Description |
|-------|-------|-------------|
| `cardName` | string | GPU card name (e.g., "card0") |
| `deviceName` | string | Device path (e.g., "/dev/dri/card0") |
| `driver` | string | GPU driver name (e.g., "amdgpu", "nvidia", "i915") |
| `vendor` | string | GPU vendor (e.g., "AMD", "NVIDIA", "Intel") |
| `vendorId` | string | PCI vendor ID |
| `deviceId` | string | PCI device ID |
| `subsystemVendor` | string | Subsystem vendor ID |
| `subsystemDevice` | string | Subsystem device ID |
| `class` | string | PCI class |
| `revision` | string | Device revision |
| `enabled` | boolean | Whether GPU is enabled |
| `status` | string | GPU status (e.g., "active", "connected") |
| `connectors` | array | All available connectors (only if show-connectors=true) |
| `enabledConnectors` | array | Currently enabled connectors |
| `vramSize` | number | Total VRAM size in bytes |
| `vramUsed` | number | Used VRAM in bytes |
| `vramFree` | number | Free VRAM in bytes |
| `gartSize` | number | GART size in bytes |
| `coreClock` | number | Current core clock in Hz |
| `memoryClock` | number | Current memory clock in Hz |
| `maxCoreClock` | number | Maximum core clock in Hz |
| `maxMemoryClock` | number | Maximum memory clock in Hz |
| `powerUsage` | number | Current power usage in milliwatts |
| `powerLimit` | number | Power limit in milliwatts |
| `temperature` | number | Current temperature in Celsius |
| `temperatureCrit` | number | Critical temperature in Celsius |
| `fanSpeed` | number | Fan speed in RPM |
| `fanSpeedPercent` | number | Fan speed percentage (0-100) |
| `gpuUtilPercent` | number | GPU utilization percentage (0-100) |
| `memoryUtilPercent` | number | Memory utilization percentage (0-100) |
| `busId` | string | PCI bus ID (e.g., "0000:01:00.0") |
| `busWidth` | string | Bus width (e.g., "x16", "x8") |
| `pcieGen` | string | Current PCIe generation (e.g., "4.0") |
| `maxPcieGen` | string | Maximum PCIe generation |
| `vbiosVersion` | string | VBIOS version |
| `firmwareVersion` | string | Firmware version |
| `devicePath` | string | Device path in sysfs |
| `sysfsPath` | string | Sysfs path |
| `isPrimary` | boolean | Whether this is the primary GPU |
| `gpuType` | string | GPU type ("discrete" or "integrated") |
| `computeUnits` | number | Number of compute units (AMD) |
| `cudaCores` | number | Number of CUDA cores (NVIDIA) |
| `shaders` | number | Number of shaders |
| `textureUnits` | number | Number of texture units |
| `rops` | number | Number of raster operations units |
| `modes` | array | Supported display modes (only if show-modes=true) |

**Examples:**

```bash
# Get first GPU
curl http://localhost:8080/api/v1/gpu

# Get specific GPU
curl "http://localhost:8080/api/v1/gpu?card=card1"

# Get all GPUs
curl "http://localhost:8080/api/v1/gpu

# Get GPU with connector information
curl "http://localhost:8080/api/v1/gpu?show-connectors=true"

# Get all GPUs with full details
curl "http://localhost:8080/api/v1/gpu?show-connectors=true&show-modes=true"
```

---

### Battery Information

Retrieve battery information including capacity, status, voltage, current, power, and health.

**Endpoint:** `GET /api/v1/battery`

**Description:** Returns battery status and health information. Returns information for the first available battery.

**Query Parameters:** None

**Response Schema:**

```json
{
  "status": "success",
  "data": {
    "name": "string",
    "path": "string",
    "manufacturer": "string",
    "modelName": "string",
    "serial": "string",
    "technology": "string",
    "status": "string",
    "capacity": number,
    "designCapacity": number,
    "currentCapacity": number,
    "chargePercent": number,
    "voltage": number,
    "current": number,
    "power": number,
    "health": number,
    "cycleCount": number,
    "timeToEmpty": number,
    "timeToFull": number
  },
  "timestamp": "2026-01-20T12:00:00Z"
}
```

**Field Descriptions:**

| Field | Type | Description |
|-------|-------|-------------|
| `name` | string | Battery name (e.g., "BAT0") |
| `path` | string | Battery path in sysfs |
| `manufacturer` | string | Battery manufacturer |
| `modelName` | string | Battery model name |
| `serial` | string | Battery serial number |
| `technology` | string | Battery technology (e.g., "Li-ion") |
| `status` | string | Battery status ("charging", "discharging", "full", "unknown") |
| `capacity` | number | Current capacity in watt-hours |
| `designCapacity` | number | Design capacity in watt-hours |
| `currentCapacity` | number | Current capacity in watt-hours |
| `chargePercent` | number | Charge percentage (0-100) |
| `voltage` | number | Current voltage in volts |
| `current` | number | Current in amperes |
| `power` | number | Current power in watts |
| `health` | number | Battery health percentage (0-100) |
| `cycleCount` | number | Number of charge cycles |
| `timeToEmpty` | number | Estimated time to empty in minutes (0 if charging) |
| `timeToFull` | number | Estimated time to full in minutes (0 if discharging) |

**Example:**

```bash
curl http://localhost:8080/api/v1/battery
```

---

### Services Information

Retrieve system services information including status, resource usage, and listening ports.

**Endpoint:** `GET /api/v1/services`

**Description:** Returns services metrics with filtering options. Can return all services or specific service details.

**Query Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|-------|----------|----------|-------------|
| `name` | string | No | - | Filter by service name (partial match) |
| `status` | string | No | all | Filter by status ("running", "stopped", "failed") |
| `user` | string | No | all | Filter by user running the service |
| `port` | integer | No | - | Filter by listening port number |

**Note:** If `name` parameter is provided, returns detailed information for a single service. Otherwise, returns a summary of filtered services.

**Response Schema (Summary):**

```json
{
  "status": "success",
  "data": {
    "totalServices": number,
    "running": number,
    "stopped": number,
    "failed": number,
    "services": [
      {
        "name": "string",
        "description": "string",
        "status": "string",
        "loaded": "string",
        "active": "string",
        "subState": "string",
        "pid": number,
        "memoryMB": number,
        "cpuPercent": number,
        "startTime": "string",
        "uptime": "string",
        "user": "string",
        "command": "string",
        "listeningPorts": [
          {
            "protocol": "string",
            "localAddr": "string",
            "port": number,
            "state": "string",
            "processName": "string",
            "pid": number
          }
        ],
        "type": "string"
      }
    ]
  },
  "timestamp": "2026-01-20T12:00:00Z"
}
```

**Response Schema (Single Service - when name parameter is provided):**

```json
{
  "status": "success",
  "data": {
    "name": "string",
    "description": "string",
    "status": "string",
    "loaded": "string",
    "active": "string",
    "subState": "string",
    "pid": number,
    "memoryMB": number,
    "cpuPercent": number,
    "startTime": "string",
    "uptime": "string",
    "user": "string",
    "command": "string",
    "listeningPorts": [
      {
        "protocol": "string",
        "localAddr": "string",
        "port": number,
        "state": "string",
        "processName": "string",
        "pid": number
      }
    ],
    "type": "string"
  },
  "timestamp": "2026-01-20T12:00:00Z"
}
```

**Field Descriptions:**

| Field | Type | Description |
|-------|-------|-------------|
| `totalServices` | number | Total number of services returned |
| `running` | number | Number of running services |
| `stopped` | number | Number of stopped services |
| `failed` | number | Number of failed services |
| `name` | string | Service name |
| `description` | string | Service description |
| `status` | string | Service status ("running", "stopped", "failed") |
| `loaded` | string | Loaded state |
| `active` | string | Active state |
| `subState` | string | Sub-state (e.g., "running", "dead", "exited") |
| `pid` | number | Process ID |
| `memoryMB` | number | Memory usage in megabytes |
| `cpuPercent` | number | CPU usage percentage |
| `startTime` | string | Start time (e.g., "Jan 15 09:30:00") |
| `uptime` | string | Human-readable uptime |
| `user` | string | User running the service |
| `command` | string | Full command line |
| `listeningPorts` | array | Array of listening ports |
| `type` | string | Service type ("systemd", "init") |

**Listening Port Fields:**

| Field | Type | Description |
|-------|-------|-------------|
| `protocol` | string | Protocol ("tcp", "tcp6", "udp", "udp6") |
| `localAddr` | string | Local address |
| `port` | number | Port number |
| `state` | string | Socket state |
| `processName` | string | Process name |
| `pid` | number | Process ID |

**Examples:**

```bash
# Get all running services
curl "http://localhost:8080/api/v1/services?status=running"

# Get services by name (detailed view)
curl "http://localhost:8080/api/v1/services?name=nginx"

# Get services listening on port 80
curl "http://localhost:8080/api/v1/services?port=80"

# Get services by user
curl "http://localhost:8080/api/v1/services?user=root"

# Get all services
curl "http://localhost:8080/api/v1/services"
```

---

### System Summary

Retrieve all system metrics in a single request.

**Endpoint:** `GET /api/v1/summary`

**Description:** Returns a comprehensive summary of all system metrics including CPU, disks, GPUs, battery, and services.

**Query Parameters:** None

**Response Schema:**

```json
{
  "status": "success",
  "data": {
    "cpu": { /* CPU object */ },
    "disks": [
      { /* Disk object */ }
    ],
    "gpus": [
      { /* GPU object */ }
    ],
    "battery": { /* Battery object */ },
    "services": { /* Services summary object */ }
  },
  "timestamp": "2026-01-20T12:00:00Z"
}
```

**Note:** This endpoint returns a subset of available data compared to individual endpoints. For detailed information, use the specific endpoint.

**Example:**

```bash
curl http://localhost:8080/api/v1/summary
```

---

## Error Handling

All error responses follow a consistent format:

```json
{
  "status": "error",
  "error": "Error message description",
  "code": "ERROR_CODE",
  "timestamp": "2026-01-20T12:00:00Z"
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `METHOD_NOT_ALLOWED` | 405 | HTTP method not supported |
| `INVALID_PARAMETER` | 400 | Invalid query parameter |
| `NOT_FOUND` | 404 | Resource not found |
| `INTERNAL_ERROR` | 500 | Internal server error |

### Common Error Scenarios

**Invalid Parameter:**
```bash
curl "http://localhost:8080/api/v1/disk?port=abc"
```

Response:
```json
{
  "status": "error",
  "error": "Invalid port number",
  "code": "INVALID_PARAMETER",
  "timestamp": "2026-01-20T12:00:00Z"
}
```

**Resource Not Found:**
```bash
curl "http://localhost:8080/api/v1/disk?device=nonexistent"
```

Response:
```json
{
  "status": "error",
  "error": "No disk found",
  "code": "NOT_FOUND",
  "timestamp": "2026-01-20T12:00:00Z"
}
```

**Internal Error:**
```json
{
  "status": "error",
  "error": "Failed to read CPU information: permission denied",
  "code": "INTERNAL_ERROR",
  "timestamp": "2026-01-20T12:00:00Z"
}
```

---

## Rate Limiting

Currently, the Linux Toolkit API does not implement rate limiting. This may change in future versions for production deployments.

**Recommendation:** Implement rate limiting at the reverse proxy level (e.g., Nginx, HAProxy) for production deployments.

---

## Best Practices

### 1. Caching

Many metrics don't change frequently. Implement client-side caching to reduce server load:

- CPU metrics: Cache for 5-10 seconds
- Disk metrics: Cache for 30-60 seconds
- GPU metrics: Cache for 5-10 seconds
- Battery metrics: Cache for 30-60 seconds
- Services metrics: Cache for 30-60 seconds

### 2. Error Handling

Always check the `status` field in responses:

```python
import requests

response = requests.get('http://localhost:8080/api/v1/cpu')
data = response.json()

if data['status'] == 'success':
    # Process data
    cpu_info = data['data']
else:
    # Handle error
    print(f"Error: {data['error']}")
```

### 3. Request Filtering

Use query parameters to retrieve only the data you need:

```bash
# Bad - fetches all CPU data
curl http://localhost:8080/api/v1/cpu

# Good - fetches only essential data
curl "http://localhost:8080/api/v1/cpu?show-cores=true"
```

### 4. Batch Requests

When needing multiple metrics, use the `/api/v1/summary` endpoint instead of multiple individual requests:

```bash
# Bad - multiple requests
curl http://localhost:8080/api/v1/cpu
curl http://localhost:8080/api/v1/disk
curl http://localhost:8080/api/v1/gpu

# Good - single request
curl http://localhost:8080/api/v1/summary
```

### 5. Timeouts

Set appropriate timeouts for API requests:

```python
import requests

try:
    response = requests.get(
        'http://localhost:8080/api/v1/cpu',
        timeout=10  # 10 second timeout
    )
except requests.Timeout:
    print("Request timed out")
```

### 6. Pagination

Not currently applicable as all endpoints return complete datasets. Future versions may implement pagination for large result sets.

### 7. Monitoring

Monitor the `/api/v1/health` endpoint periodically to check server availability:

```bash
while true; do
    curl http://localhost:8080/api/v1/health
    sleep 5
done
```

### 8. Security

**Important:** The API currently has no authentication. For production deployments:

1. **Use reverse proxy** (Nginx, Apache) to handle authentication
2. **Restrict network access** using firewall rules
3. **Implement API keys** or OAuth if needed
4. **Use HTTPS** with valid certificates

Example Nginx configuration with basic authentication:

```nginx
location /api/ {
    auth_basic "Restricted Access";
    auth_basic_user_file /etc/nginx/.htpasswd;

    proxy_pass http://localhost:8080/api/;
    proxy_set_header Host $host;
}
```

### 9. Performance

- Use connection pooling for multiple requests
- Compress responses (enable gzip compression in reverse proxy)
- Minimize unnecessary parameters
- Monitor response times and adjust polling intervals accordingly

### 10. Development

For development and testing:

- Use `--no-ui` flag to start API-only mode
- Test endpoints with tools like `curl`, `httpie`, or Postman
- Validate JSON responses with `jq`:

```bash
curl -s http://localhost:8080/api/v1/cpu | jq '.data.modelName'
```

---

## Support

For issues, questions, or contributions, please visit:
- GitHub: https://github.com/jun/linux-toolkit
- Issues: https://github.com/jun/linux-toolkit/issues

---

**Last Updated:** January 20, 2026
