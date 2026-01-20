# Linux-Toolkit Server Feature Implementation Plan

## Executive Summary

This plan details the implementation of an HTTP server mode for linux-toolkit that exposes existing CLI tools as JSON APIs and provides a simple HTML UI for monitoring system metrics.

**Author**: AI Planner  
**Date**: 2026-01-20  
**Status**: Draft  
**Estimated Implementation Time**: 8-12 hours

---

## 1. Requirements Analysis

### User Requirements
- Command-line option to spin up a server
- API endpoints returning JSON output for all existing tools
- Simple HTML UI to view system metrics
- Configurable port via command-line argument

### Current Codebase State
- **Language**: Go 1.25.5
- **CLI Framework**: Cobra (spf13/cobra v1.10.2)
- **Existing Tools**: disk, cpu, gpu, battery, services (5 tools)
- **Output Formats**: Table (default) + JSON (`--format json`)
- **JSON Structures**: Already defined with proper tags in all pkg packages
- **Testing**: No existing test infrastructure
- **Web Dependencies**: None (pure CLI tool)

---

## 2. Architecture Decisions

### 2.1 Web Framework Selection

**Selected Framework**: **Standard Library `net/http`**

**Rationale**:
1. **No additional dependencies** - Keep binary small, follow project's minimalist approach
2. **Adequate performance** - For this use case, standard library is sufficient
3. **Existing JSON structures** can be reused directly via `encoding/json`
4. **Simpler to maintain** - No framework-specific patterns to learn
5. **Production-ready** - Used by Kubernetes, Docker, and many major projects

**Alternatives Considered**:
- **Gin** - Popular, high-performance, but adds 20KB+ dependency
- **Echo** - Very fast, but similar tradeoffs to Gin
- **Chi** - Lightweight router, but still external dependency

### 2.2 Server Architecture

```
┌─────────────────────────────────────────────────────────┐
│                 linux-toolkit                      │
│                                                     │
│  CLI Mode (existing)                     Server Mode (new)│
│  ┌──────────────┐                 ┌────────────────┐  │
│  │   Commands   │                 │ HTTP Server   │  │
│  │              │                 │              │  │
│  │ - disk       │                 │ - Router     │  │
│  │ - cpu        │                 │ - Handlers   │  │
│  │ - gpu        │                 │ - Static UI  │  │
│  │ - battery    │                 │              │  │
│  │ - services   │                 └──────┬───────┘  │
│  └──────────────┘                        │          │
│                                         │ Reuses pkg/*│
│                                ┌──────────┴─────────┐ │
│                                │                    │ │
│                                │ pkg/disk            │ │
│                                │ pkg/cpu             │ │
│                                │ pkg/gpu             │ │
│                                │ pkg/battery          │ │
│                                │ pkg/services        │ │
│                                └────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

### 2.3 API Versioning

- **Base Path**: `/api/v1/`
- **Version**: v1 to allow future breaking changes
- **Semantics**: Follow REST conventions for clear, predictable URLs

---

## 3. API Endpoints Design

### 3.1 Endpoints Overview

| Method | Path | Description | Response Format |
|---------|-------|-------------|-----------------|
| GET | `/api/v1/health` | Health check | JSON status |
| GET | `/api/v1/disk` | All disks info | JSON array |
| GET | `/api/v1/disk/{device}` | Specific disk | JSON object |
| GET | `/api/v1/cpu` | CPU information | JSON object |
| GET | `/api/v1/cpu/cores` | Per-core details | JSON array |
| GET | `/api/v1/gpu` | All GPUs | JSON array |
| GET | `/api/v1/gpu/{card}` | Specific GPU | JSON object |
| GET | `/api/v1/battery` | Battery status | JSON object |
| GET | `/api/v1/services` | All services | JSON object |
| GET | `/api/v1/services/{name}` | Specific service | JSON object |
| GET | `/api/v1/summary` | All metrics summary | JSON object |

### 3.2 Query Parameters (Optional)

Support existing CLI flags as query params:

**Disk**:
- `?device={name}` - Filter by device
- `?mount={path}` - Filter by mount point
- `?io-stats=true` - Include I/O stats
- `?inode-stats=true` - Include inode stats

**CPU**:
- `?show-cores=true` - Include per-core data
- `?show-flags=true` - Include CPU flags
- `?show-temp=true` - Include temperature

**GPU**:
- `?card={name}` - Filter by card
- `?show-connectors=true` - Include connector info
- `?show-modes=true` - Include display modes

**Services**:
- `?name={filter}` - Filter by name
- `?status={running\|stopped\|failed}` - Filter by status
- `?user={name}` - Filter by user
- `?port={number}` - Filter by port
- `?show-ports=true` - Include listening ports

### 3.3 Response Format Standards

**Success Response**:
```json
{
  "status": "success",
  "data": { ... },
  "timestamp": "2026-01-20T10:30:00Z"
}
```

**Error Response**:
```json
{
  "status": "error",
  "error": "Error message",
  "code": "NOT_FOUND"
}
```

**HTTP Status Codes**:
- `200 OK` - Successful request
- `400 Bad Request` - Invalid parameters
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

---

## 4. HTML UI Design

### 4.1 UI Architecture

- **Single Page Application** (SPA) using vanilla JavaScript
- **Auto-refresh** with configurable interval (default 5s)
- **Real-time updates** via polling (WebSocket optional future enhancement)
- **Responsive design** for desktop, tablet, mobile

### 4.2 UI Pages

**Dashboard (`/`)**:
- Summary cards showing key metrics
  - CPU usage and temperature
  - Memory usage
  - Disk usage for primary disks
  - GPU status
  - Active services count
  - Battery level (if available)
- Visual graphs/bars for usage percentages
- Color-coded status indicators (green/yellow/red)

**Detailed Views**:
- `/cpu` - Detailed CPU info with core breakdown
- `/disks` - All disks with detailed I/O stats
- `/gpus` - GPU details with clock speeds
- `/services` - Service list with status and ports
- `/battery` - Battery health and charge status

### 4.3 Static File Structure

```
pkg/server/static/
├── index.html          # Main dashboard
├── cpu.html           # CPU detail page
├── disk.html          # Disk detail page
├── gpu.html           # GPU detail page
├── services.html       # Services detail page
├── battery.html       # Battery detail page
├── css/
│   └── style.css     # Main stylesheet
└── js/
    └── app.js        # JavaScript logic
```

### 4.4 UI Mockup Description

**Header**:
- Title: "Linux Toolkit Dashboard"
- Refresh interval dropdown (5s, 10s, 30s, manual)
- Auto-refresh toggle button

**Main Content (Grid Layout)**:
```
┌─────────────────┬─────────────────┬─────────────────┐
│  CPU 🖥️      │  GPU 🎮       │  Battery 🔋    │
│  Usage: 45%     │  Usage: 78%    │  85%           │
│  Temp: 52°C     │  Temp: 62°C     │  Charging ⚡    │
└─────────────────┴─────────────────┴─────────────────┘
┌───────────────────────────────────────────────────────┐
│  Disk 💾                                      │
│  /dev/sda1: 49.4% (234 GB / 500 GB)          │
│  ██████████████████████████░░░░░░░░░░░░░░   │
│  ─────────────────────────────────────────────────── │
│  /dev/sda2: 47.7% (45 GB / 100 GB)           │
│  ██████████████████████░░░░░░░░░░░░░░░░░░░░   │
└───────────────────────────────────────────────────────┘
┌───────────────────────────────────────────────────────┐
│  Services 🛠️  (12 running, 1 failed)           │
│  🟢 sshd         🟢 nginx   🟢 docker       │
│  🟢 cron         🔴 my-service (failed)         │
└───────────────────────────────────────────────────────┘
```

---

## 5. Implementation Plan

### Phase 1: Server Core Infrastructure (2-3 hours)

#### 5.1.1 Create Server Package Structure

```bash
mkdir -p pkg/server/{static/css,static/js}
touch pkg/server/server.go
touch pkg/server/handlers.go
touch pkg/server/routes.go
```

**File: `pkg/server/server.go`**
- `Server` struct with http.Server, port, bind address
- `NewServer(port, bind string) *Server`
- `Start() error` - Start HTTP server
- `Stop() error` - Graceful shutdown
- `healthCheckHandler()` - Health endpoint

**File: `pkg/server/handlers.go`**
- JSON response wrapper functions
- API handlers for each tool (disk, cpu, gpu, battery, services)
- Static file serving
- CORS middleware (optional)
- Error handling

**File: `pkg/server/routes.go`**
- Route registration
- Path-to-handler mapping
- Query parameter parsing

#### 5.1.2 Add Server Command

**File: `cmd/server.go`**
```go
var serverCmd = &cobra.Command{
    Use:   "server",
    Short: "Start HTTP server for web access",
    Long:  "Start HTTP server that exposes system metrics as JSON API with HTML UI",
    Run:   runServer,
}
```

**Flags**:
- `--port, -p`: Port number (default: 8080)
- `--bind, -b`: Bind address (default: 0.0.0.0)
- `--interval, -i`: UI refresh interval (default: 5s)
- `--no-ui, -n`: Start API only without UI (optional)

#### 5.1.3 Update go.mod

No additional dependencies needed (using `net/http` only).

---

### Phase 2: API Handlers Implementation (3-4 hours)

#### 5.2.1 Reuse Existing JSON Output

The existing `pkg/*/output.go` files already have JSON serialization logic:

**Strategy**:
1. Import tool packages in `handlers.go`
2. Create tool readers and call existing functions
3. Wrap responses with standard API format

**Example for Disk Handler**:
```go
func handleDisk(c *gin.Context) {
    device := c.Query("device")
    mount := c.Query("mount")
    ioStats := c.Query("io-stats") == "true"
    inodeStats := c.Query("inode-stats") == "true"
    
    reader := disk.NewDiskReader()
    
    if device != "" {
        info, io, inode := reader.ReadDiskInfo(device)
        c.JSON(200, apiResponse{Status: "success", Data: map[string]interface{}{
            "disk": info, "ioStats": io, "inodeInfo": inode,
        }})
    } else {
        disks := reader.ListMountedDisks()
        c.JSON(200, apiResponse{Status: "success", Data: disks})
    }
}
```

#### 5.2.2 Implement All Handlers

- `handleHealth()` - Simple health check
- `handleDisk()` - Disk metrics
- `handleCPU()` - CPU metrics
- `handleGPU()` - GPU metrics
- `handleBattery()` - Battery metrics
- `handleServices()` - Services list
- `handleSummary()` - Combined metrics

#### 5.2.3 Add Query Parameter Support

Map query params to existing CLI flags:
- Parse `?io-stats=true` → set `includeIO = true`
- Parse `?device=sda1` → set `diskName = "sda1"`
- Apply same logic as existing CLI commands

---

### Phase 3: Static HTML UI Implementation (3-4 hours)

#### 5.3.1 Create Base HTML Template

**File: `pkg/server/static/index.html`**
- HTML5 boilerplate
- Responsive meta tags
- Embedded CSS or link to `css/style.css`
- Link to `js/app.js`

#### 5.3.2 Implement Dashboard JavaScript

**File: `pkg/server/static/js/app.js`**

**Features**:
- `fetchMetrics()` function to call API endpoints
- `updateDashboard()` to update DOM elements
- `autoRefresh()` with setInterval
- `formatBytes()` for disk sizes
- `formatPercentage()` for usage bars

**Functions**:
```javascript
async function fetchAllMetrics() {
    const [cpu, disk, gpu, services, battery] = await Promise.all([
        fetch('/api/v1/cpu').then(r => r.json()),
        fetch('/api/v1/disk').then(r => r.json()),
        fetch('/api/v1/gpu').then(r => r.json()),
        fetch('/api/v1/services').then(r => r.json()),
        fetch('/api/v1/battery').then(r => r.json())
    ]);
    
    return { cpu, disk, gpu, services, battery };
}

function updateDashboard(metrics) {
    // Update CPU card
    document.getElementById('cpu-usage').textContent = metrics.cpu.data.usagePercent + '%';
    // ... update other elements
    
    // Update progress bars
    document.getElementById('cpu-bar').style.width = metrics.cpu.data.usagePercent + '%';
    // ... update other bars
}
```

#### 5.3.3 Create CSS Styling

**File: `pkg/server/static/css/style.css`**

**Styles**:
- Modern, dark theme (matches CLI colors)
- Grid layout for cards
- Progress bars with color gradients
- Responsive breakpoints (mobile, tablet, desktop)
- Status indicators (green/yellow/red)

**Colors** (matching CLI theme):
- Success: `#00ff00` (green)
- Warning: `#ffff00` (yellow)
- Critical: `#ff0000` (red)
- Background: `#1e1e1e` (dark gray)
- Cards: `#2d2d2d` (slightly lighter)

#### 5.3.4 Create Detail Pages

- `cpu.html` - Core breakdown table
- `disk.html` - All disks with I/O stats table
- `gpu.html` - GPU details with VRAM usage
- `services.html` - Service list with status icons
- `battery.html` - Battery health metrics

---

### Phase 4: Integration and Testing (1-2 hours)

#### 5.4.1 Update root.go

Add server subcommand to root command:
```go
func init() {
    rootCmd.AddCommand(serverCmd)
}
```

#### 5.4.2 Build and Test

```bash
# Build
go build -o linux-toolkit

# Test CLI mode (existing)
./linux-toolkit cpu
./linux-toolkit disk --format json

# Test server mode
./linux-toolkit server --port 8080

# Test API endpoints
curl http://localhost:8080/api/v1/health
curl http://localhost:8080/api/v1/cpu
curl http://localhost:8080/api/v1/disk?device=sda1&io-stats=true
```

#### 5.4.3 Test HTML UI

- Open browser to `http://localhost:8080`
- Verify auto-refresh works
- Check responsive design (resize window)
- Test all detail pages
- Verify JSON data displays correctly

---

## 6. Code Changes Summary

### New Files

```
pkg/server/
├── server.go           # Server struct and main loop
├── handlers.go        # API endpoint handlers
├── routes.go          # Route registration
└── static/
    ├── index.html      # Main dashboard
    ├── cpu.html       # CPU detail page
    ├── disk.html      # Disk detail page
    ├── gpu.html       # GPU detail page
    ├── services.html   # Services detail page
    ├── battery.html   # Battery detail page
    ├── css/
    │   └── style.css # Stylesheets
    └── js/
        └── app.js    # JavaScript logic

cmd/
└── server.go          # Server CLI command
```

### Modified Files

```
cmd/root.go            # Add server subcommand
go.mod                # No changes needed
README.md             # Update with server usage
```

### Lines of Code Estimate

- `server.go`: ~150 lines
- `handlers.go`: ~400 lines (6 handlers)
- `routes.go`: ~100 lines
- `server.go` (cmd): ~80 lines
- HTML templates: ~800 lines total (6 pages)
- CSS: ~300 lines
- JavaScript: ~300 lines
- **Total new code**: ~2,130 lines

---

## 7. Error Handling Strategy

### 7.1 System Reading Errors

- Permission denied when reading `/proc` or `/sys`
  - Return 403 with message: "Insufficient permissions to read system data"
  - Recommend running as root or with appropriate privileges

- Device not found
  - Return 404 with message: "Device not found"
  - Example: GPU card that doesn't exist

### 7.2 API Error Responses

```go
type APIError struct {
    Status  string `json:"status"`
    Error   string `json:"error"`
    Code    string `json:"code"`
}

func respondError(w http.ResponseWriter, code int, message string, errCode string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(APIError{
        Status: "error",
        Error:  message,
        Code:    errCode,
    })
}
```

### 7.3 UI Error Handling

- Display error message to user
- Show retry button for failed API calls
- Gracefully handle missing data (e.g., no GPU)
- Clear console logs for debugging

---

## 8. Security Considerations

### 8.1 Access Control

- **Default**: Bind to `0.0.0.0` (all interfaces)
- **Recommendation**: Add `--bind 127.0.0.1` flag for localhost-only
- **Future**: Add basic authentication or API tokens

### 8.2 Input Validation

- Validate port number (1024-65535)
- Validate device names (prevent path traversal)
- Sanitize query parameters

### 8.3 Rate Limiting (Optional Enhancement)

- Implement simple rate limiter to prevent abuse
- Limit: 100 requests per minute per IP
- Return 429 status when exceeded

### 8.4 CORS (Optional Enhancement)

If accessing UI from different domain:
```go
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
        next.ServeHTTP(w, r)
    })
}
```

---

## 9. Performance Considerations

### 9.1 System Reading Optimization

- **Cache results** for 1-2 seconds to reduce system calls
- Read `/proc` and `/sys` files only once per refresh cycle
- Use sync.Mutex for concurrent request handling

### 9.2 HTTP Performance

- **Connection pooling**: Standard library handles automatically
- **Response caching**: Add `Cache-Control` header (max-age=5s)
- **Compression**: Optional gzip compression for JSON responses

### 9.3 UI Performance

- **Lazy loading**: Only fetch data for visible tabs/pages
- **Debouncing**: Prevent rapid API calls from user interactions
- **LocalStorage**: Cache data to reduce requests

---

## 10. Testing Strategy

### 10.1 Unit Tests (New - No Existing Infrastructure)

**Test Files**:
- `pkg/server/server_test.go` - Server lifecycle tests
- `pkg/server/handlers_test.go` - Handler tests
- `pkg/server/routes_test.go` - Routing tests

**Test Cases**:
```go
func TestNewServer(t *testing.T) {
    srv := server.NewServer(8080, "127.0.0.1")
    if srv == nil {
        t.Fatal("NewServer returned nil")
    }
    if srv.Port != 8080 {
        t.Errorf("Expected port 8080, got %d", srv.Port)
    }
}

func TestHealthHandler(t *testing.T) {
    req := httptest.NewRequest("GET", "/api/v1/health", nil)
    w := httptest.NewRecorder()
    
    handler(w, req)
    
    if w.Code != 200 {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
    
    var resp map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &resp)
    
    if resp["status"] != "ok" {
        t.Errorf("Expected status 'ok', got %v", resp["status"])
    }
}
```

### 10.2 Integration Tests

**Test Scenarios**:
1. Start server and verify all endpoints return 200
2. Test query parameter parsing
3. Test error cases (invalid device, permission denied)
4. Test concurrent requests (simulate multiple clients)
5. Test graceful shutdown

### 10.3 Manual Testing Checklist

- [ ] Server starts on specified port
- [ ] Health endpoint returns 200
- [ ] All API endpoints return valid JSON
- [ ] Query parameters work correctly
- [ ] HTML UI loads in browser
- [ ] Auto-refresh functions properly
- [ ] All detail pages work
- [ ] Responsive design works on mobile
- [ ] Error messages display correctly
- [ ] Server shuts down gracefully (Ctrl+C)

---

## 11. Documentation Updates

### 11.1 README.md Changes

Add new section:

```markdown
## Server Mode

Start an HTTP server to access system metrics via web browser or API.

### Usage

```bash
# Start server on default port (8080)
linux-toolkit server

# Start server on custom port
linux-toolkit server --port 3000

# Bind to localhost only
linux-toolkit server --bind 127.0.0.1 --port 8080

# Set UI refresh interval (seconds)
linux-toolkit server --interval 10

# Start API only (no UI)
linux-toolkit server --no-ui
```

### Accessing the Server

**Web UI**: Open browser to `http://localhost:8080`

**API Endpoints**:
- `GET /api/v1/health` - Health check
- `GET /api/v1/cpu` - CPU information
- `GET /api/v1/disk` - Disk information
- `GET /api/v1/gpu` - GPU information
- `GET /api/v1/battery` - Battery information
- `GET /api/v1/services` - Services information
- `GET /api/v1/summary` - All metrics

### API Example

```bash
# Get CPU info as JSON
curl http://localhost:8080/api/v1/cpu

# Get specific disk with I/O stats
curl "http://localhost:8080/api/v1/disk?device=sda1&io-stats=true"

# Get services with ports
curl "http://localhost:8080/api/v1/services?show-ports=true"
```
```

### 11.2 Help Text

Add command help:
```bash
linux-toolkit server --help
```

Output:
```
Start HTTP server for web access to system metrics

Usage:
  linux-toolkit server [flags]

Flags:
  -b, --bind string      Bind address (default "0.0.0.0")
  -p, --port int        Port number (default 8080)
  -i, --interval int    UI refresh interval in seconds (default 5)
  -n, --no-ui          Start API only, no HTML UI
  -h, --help            help for server
```

---

## 12. Future Enhancements

### 12.1 WebSocket Support

Replace polling with real-time updates:
- Push new metrics when system state changes
- Reduce server load
- More responsive UI

### 12.2 Historical Data

- Store metrics history (last hour, day, week)
- Visualize trends with charts
- Export data for analysis

### 12.3 Authentication

- Basic auth for API access
- Token-based authentication
- User permissions (read-only vs admin)

### 12.4 Export Formats

- CSV export
- PDF reports
- Prometheus metrics endpoint (`/metrics`)

### 12.5 Mobile App

- PWA (Progressive Web App)
- Offline support
- Push notifications for alerts

---

## 13. Success Criteria

The implementation is complete when:

1. ✅ `linux-toolkit server --port 8080` starts successfully
2. ✅ Browser accesses `http://localhost:8080` and shows dashboard
3. ✅ API endpoints return valid JSON matching existing structures
4. ✅ Query parameters work (e.g., `?device=sda1`, `?io-stats=true`)
5. ✅ HTML UI auto-refreshes every 5 seconds by default
6. ✅ All 5 tools (disk, cpu, gpu, battery, services) accessible via API
7. ✅ Responsive design works on desktop and mobile
8. ✅ Error handling covers permission denied, not found, and server errors
9. ✅ Server shuts down gracefully on Ctrl+C
10. ✅ README updated with server usage examples
11. ✅ Unit tests added for server package
12. ✅ No breaking changes to existing CLI functionality

---

## 14. Implementation Checklist

### Phase 1: Server Infrastructure
- [ ] Create `pkg/server/` directory structure
- [ ] Implement `pkg/server/server.go` with Server struct
- [ ] Implement graceful shutdown logic
- [ ] Create `cmd/server.go` with Cobra command
- [ ] Add flags: `--port`, `--bind`, `--interval`, `--no-ui`
- [ ] Add server command to `cmd/root.go`

### Phase 2: API Handlers
- [ ] Create `pkg/server/handlers.go`
- [ ] Implement `handleHealth()` handler
- [ ] Implement `handleDisk()` handler with query params
- [ ] Implement `handleCPU()` handler with query params
- [ ] Implement `handleGPU()` handler with query params
- [ ] Implement `handleBattery()` handler
- [ ] Implement `handleServices()` handler with query params
- [ ] Implement `handleSummary()` handler
- [ ] Add error response helper functions
- [ ] Create `pkg/server/routes.go` with route registration

### Phase 3: HTML UI
- [ ] Create `pkg/server/static/` directory
- [ ] Create `index.html` dashboard page
- [ ] Create CSS file `css/style.css`
- [ ] Create JavaScript file `js/app.js`
- [ ] Implement `fetchAllMetrics()` function
- [ ] Implement `updateDashboard()` function
- [ ] Implement auto-refresh with `setInterval()`
- [ ] Create detail pages: `cpu.html`, `disk.html`, `gpu.html`, `services.html`, `battery.html`
- [ ] Add responsive breakpoints

### Phase 4: Testing
- [ ] Create `pkg/server/server_test.go`
- [ ] Create `pkg/server/handlers_test.go`
- [ ] Write tests for all handlers
- [ ] Test with `httptest` for HTTP requests
- [ ] Build and run integration tests

### Phase 5: Documentation
- [ ] Update `README.md` with server section
- [ ] Add usage examples for all API endpoints
- [ ] Update command help text

---

## 15. Notes and Assumptions

### Assumptions
1. User wants a simple, functional UI without external frameworks (React/Vue)
2. No authentication required initially (security enhancement later)
3. Server will run on same machine monitoring itself (localhost access)
4. Existing JSON output structures are suitable for API responses
5. No need for database or data persistence (real-time only)

### Constraints
1. Keep implementation simple and maintainable
2. Reuse existing code as much as possible
3. No external dependencies (use standard library only)
4. Follow existing code patterns (Cobra, file structure)
5. Binary size should remain reasonable

### Risks and Mitigations

| Risk | Impact | Mitigation |
|-------|---------|-------------|
| High CPU usage reading /sys files | Performance degradation | Implement caching, limit refresh rate |
| Multiple concurrent API calls | Server overload | Add connection pooling, rate limiting |
| Permission issues reading system data | Data unavailable | Clear error messages, suggest running as root |
| Browser compatibility issues | UI doesn't work | Use vanilla JS, test in multiple browsers |
| Port already in use | Server won't start | Check port availability, suggest alternative port |

---

## 16. References

- [Golang net/http package](https://pkg.go.dev/net/http)
- [Cobra CLI documentation](https://github.com/spf13/cobra)
- [REST API best practices](https://restfulapi.net/)
- [HTTP handler patterns](https://go.dev/doc/effective_go#http_serve)

---

**End of Plan**
