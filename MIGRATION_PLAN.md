# Wails v2 to v3 Migration Plan

## Overview
This document outlines the step-by-step migration plan from the Wails v2 project to the new Wails v3 project structure.

## Project Structure Comparison

### Wails v2 Structure
```
cloudflared-desktop-tunnel/
├── app/
│   ├── app.go          # Main app struct with lifecycle hooks
│   ├── config.go       # Configuration management
│   ├── tunnel.go       # Tunnel manager
│   ├── backend_client.go # Backend API client
│   ├── webserver.go    # Web server manager
│   └── logger.go       # Logging system
├── binaries/
│   └── downloader.go   # Binary downloader
├── main.go             # Entry point with wails.Run()
└── frontend/           # React frontend
```

### Wails v3 Structure
```
cloudflared-desktop-tunnel-v3/
├── main.go             # Entry point with application.New()
├── services/           # Service files (one per service)
│   ├── config.go
│   ├── tunnel.go
│   ├── backend.go
│   ├── webserver.go
│   └── logger.go
├── binaries/
│   └── downloader.go
└── frontend/           # React TypeScript frontend
```

## Migration Steps

### Step 1: Update Module Name
- [x] Update `go.mod` module name from `changeme` to `github.com/votanchat/cloudflared-desktop-tunnel-v3`

### Step 2: Migrate Logger (No Service Needed)
- [ ] Copy `app/logger.go` to root directory
- [ ] Keep as package-level utilities (not a service)
- [ ] Update package name from `app` to `main`

### Step 3: Migrate Config Management
- [ ] Copy `app/config.go` to `services/config.go`
- [ ] Convert to a service struct: `type ConfigService struct {}`
- [ ] Convert methods to service methods:
  - `LoadConfig()` → `(s *ConfigService) LoadConfig() (*Config, error)`
  - `Save()` → `(s *ConfigService) Save(config *Config) error`
  - `DefaultConfig()` → `(s *ConfigService) DefaultConfig() *Config`
- [ ] Keep Route management methods as service methods

### Step 4: Migrate Tunnel Manager
- [ ] Copy `app/tunnel.go` to `services/tunnel.go`
- [ ] Convert to service: `type TunnelService struct { config *ConfigService }`
- [ ] Convert methods:
  - `Start(token string) error`
  - `Stop() error`
  - `GetStatus() map[string]interface{}`
  - `GetLogs() []string`
- [ ] Update to use ConfigService instead of direct Config access

### Step 5: Migrate Backend Client
- [ ] Copy `app/backend_client.go` to `services/backend.go`
- [ ] Convert to service: `type BackendService struct {}`
- [ ] Convert methods:
  - `FetchToken() (string, error)`
  - `ReportStatus(status map[string]interface{}) error`
- [ ] Keep WebSocket connection logic internal

### Step 6: Migrate Web Server Manager
- [ ] Copy `app/webserver.go` to `services/webserver.go`
- [ ] Convert to service: `type WebServerService struct {}`
- [ ] Convert methods:
  - `Start(port int) error`
  - `Stop() error`
  - `GetStatus() map[string]interface{}`

### Step 7: Migrate Binary Downloader
- [ ] Copy `binaries/downloader.go` to `binaries/downloader.go`
- [ ] Keep package structure, update imports if needed

### Step 8: Create Main App Service (Orchestrator)
- [ ] Create `services/app.go` to coordinate all services
- [ ] Handle startup/shutdown lifecycle
- [ ] Manage service dependencies
- [ ] Provide high-level methods for frontend

### Step 9: Update main.go
- [ ] Register all services in `application.New()`
- [ ] Set up window options
- [ ] Remove v2-specific code (OnStartup, OnShutdown, etc.)
- [ ] Use v3 application lifecycle

### Step 10: Create Minimalistic Frontend
- [ ] Create simple React components:
  - `StatusDisplay.tsx` - Show tunnel/web server status
  - `TunnelControls.tsx` - Start/Stop tunnel buttons
  - `Settings.tsx` - Basic settings (minimal)
- [ ] Remove all v2 boilerplate UI
- [ ] Use clean, minimal design
- [ ] Connect to services via generated bindings

## Key Architectural Changes

### v2 → v3 Differences

1. **Lifecycle Hooks**
   - v2: `Startup(ctx)`, `DomReady(ctx)`, `Shutdown(ctx)` methods on App struct
   - v3: No lifecycle hooks, use service initialization in main.go

2. **Service Registration**
   - v2: `Bind: []interface{}{appInstance}`
   - v3: `Services: []application.Service{application.NewService(&Service{})}`

3. **Frontend Bindings**
   - v2: `window.go.app.MethodName()`
   - v3: `ServiceName.MethodName()` from generated bindings

4. **Events**
   - v2: `runtime.EventsEmit()`
   - v3: `application.RegisterEvent[T]()` and `app.Event.Emit()`

5. **Window Management**
   - v2: Single window via options
   - v3: `app.Window.NewWithOptions()` for explicit window creation

## Files to Copy/Adapt

### Direct Copy (with package name change)
- `app/logger.go` → root `logger.go`
- `binaries/downloader.go` → `binaries/downloader.go`

### Convert to Services
- `app/config.go` → `services/config.go`
- `app/tunnel.go` → `services/tunnel.go`
- `app/backend_client.go` → `services/backend.go`
- `app/webserver.go` → `services/webserver.go`

### New Files
- `services/app.go` - Main orchestrator service
- `frontend/src/components/StatusDisplay.tsx`
- `frontend/src/components/TunnelControls.tsx`
- `frontend/src/components/Settings.tsx`

### Update Existing
- `main.go` - Complete rewrite for v3
- `frontend/src/App.tsx` - Minimalistic UI
- `go.mod` - Update module name and dependencies

## Dependencies to Add

Add to `go.mod`:
- `github.com/gin-gonic/gin` - Web server
- `github.com/gorilla/websocket` - WebSocket client

## Implementation Order

1. ✅ Create v3 project structure
2. ⏳ Update module name
3. ⏳ Migrate logger (standalone)
4. ⏳ Migrate config service
5. ⏳ Migrate tunnel service
6. ⏳ Migrate backend service
7. ⏳ Migrate webserver service
8. ⏳ Migrate binary downloader
9. ⏳ Create app orchestrator service
10. ⏳ Update main.go
11. ⏳ Create minimalistic frontend
12. ⏳ Test and refine

## Notes

- Keep UI minimal - only essential components
- Remove all v2 boilerplate
- Services should be independent where possible
- Use dependency injection for service dependencies
- Maintain the same functionality but with v3 architecture

