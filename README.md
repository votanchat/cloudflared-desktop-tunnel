# Cloudflared Desktop Tunnel - Wails v3

This is the migrated Wails v3 version of the Cloudflared Desktop Tunnel application.

## Project Structure

```
cloudflared-desktop-tunnel-v3/
├── main.go                 # Application entry point
├── logger.go               # Logging utilities
├── services/               # Service layer (v3 architecture)
│   ├── app.go             # Main orchestrator service
│   ├── config.go          # Configuration service
│   ├── tunnel.go          # Tunnel management service
│   ├── backend.go         # Backend API client service
│   └── webserver.go       # Web server service
├── binaries/              # Binary downloader
│   └── downloader.go
└── frontend/             # React TypeScript frontend
    └── src/
        └── App.tsx       # Minimalistic UI
```

## Key Changes from v2 to v3

### Architecture
- **v2**: Single `App` struct with lifecycle hooks (`Startup`, `DomReady`, `Shutdown`)
- **v3**: Service-based architecture with independent services registered via `application.NewService()`

### Service Registration
- **v2**: `Bind: []interface{}{appInstance}`
- **v3**: `Services: []application.Service{application.NewService(&Service{})}`

### Frontend Bindings
- **v2**: `window.go.app.MethodName()`
- **v3**: `ServiceName.MethodName()` from generated bindings (auto-generated on build)

### Lifecycle Management
- **v2**: Lifecycle hooks in App struct
- **v3**: Initialization in `main.go`, cleanup via defer in `main()`

## Services

### AppService
Main orchestrator service that coordinates all other services. Provides high-level methods:
- `StartTunnel(manualToken string) error`
- `StopTunnel() error`
- `GetTunnelStatus() map[string]interface{}`
- `StartWebServer(port int) error`
- `StopWebServer() error`
- `GetWebServerStatus() map[string]interface{}`
- `GetConfig() *Config`
- `UpdateConfig(config *Config) error`

### ConfigService
Manages application configuration:
- `LoadConfig() (*Config, error)`
- `Save(config *Config) error`
- `GetConfig() *Config`
- `AddRoute(hostname, service string) error`
- `RemoveRoute(hostname string) error`

### TunnelService
Manages cloudflared tunnel process:
- `Start(token string) error`
- `Stop() error`
- `IsRunning() bool`
- `GetStatus() map[string]interface{}`
- `GetLogs() []string`

### BackendService
Handles backend API communication:
- `FetchToken() (string, error)`
- `ReportStatus(status map[string]interface{}) error`
- `Start()` - Starts WebSocket connection and token refresh
- `Stop()` - Stops all connections

### WebServerService
Manages Gin web server:
- `Start(port int) error`
- `Stop() error`
- `IsRunning() bool`
- `GetStatus() map[string]interface{}`

## Development

### Prerequisites
- Go 1.24+
- Node.js and npm
- Wails v3 CLI

### Running in Development Mode

1. Install dependencies:
```bash
cd frontend
npm install
```

2. Run in dev mode:
```bash
wails3 dev
```

This will:
- Start the Go backend
- Start the Vite dev server for the frontend
- Generate TypeScript bindings automatically
- Hot-reload on changes

### Building

```bash
wails3 build
```

This will:
- Build the frontend
- Generate TypeScript bindings
- Compile the Go application
- Create platform-specific binaries

## Frontend

The frontend is a minimalistic React TypeScript application with:
- **StatusDisplay**: Shows tunnel and web server status
- **TunnelControls**: Start/Stop tunnel buttons
- **WebServerControls**: Start/Stop web server buttons

### Using Services in Frontend

After building, TypeScript bindings are generated. Import services like:

```typescript
import { AppService } from "../bindings/github.com/votanchat/cloudflared-desktop-tunnel-v3/services"

// Call service methods
const status = await AppService.GetTunnelStatus()
await AppService.StartTunnel(token)
```

## Configuration

Configuration is stored in:
- **macOS**: `~/Library/Application Support/cloudflared-desktop-tunnel-v3/config.json`
- **Linux**: `~/.config/cloudflared-desktop-tunnel-v3/config.json`
- **Windows**: `%AppData%\cloudflared-desktop-tunnel-v3\config.json`

## Logging

Logs are written to:
- **Dev mode**: Console only
- **Build mode**: `~/Library/Application Support/cloudflared-desktop-tunnel-v3/logs/app-YYYY-MM-DD.log`

## Migration Notes

### What Was Migrated
- ✅ All application logic from v2
- ✅ Configuration management
- ✅ Tunnel management
- ✅ Backend client
- ✅ Web server manager
- ✅ Binary downloader
- ✅ Logging system

### What Was Changed
- ❌ Removed v2 UI boilerplate
- ✅ Created minimalistic v3 UI
- ✅ Converted to service-based architecture
- ✅ Updated to v3 API patterns

### What Was Removed
- Old v2 frontend components (replaced with minimal UI)
- v2 lifecycle hooks (replaced with service initialization)
- v2 binding patterns (replaced with v3 bindings)

## Next Steps

1. **Generate Bindings**: Run `wails3 dev` or `wails3 build` to generate TypeScript bindings
2. **Update Frontend Imports**: After bindings are generated, update the import path in `App.tsx` if needed
3. **Test Functionality**: Verify all features work as expected
4. **Customize UI**: Add any additional UI elements as needed

## Troubleshooting

### Bindings Not Generated
- Run `wails3 dev` or `wails3 build` to generate bindings
- Check `frontend/src/bindings/` directory

### Import Errors in Frontend
- Ensure bindings are generated first
- Check the module path matches in `go.mod` and frontend imports

### Service Methods Not Available
- Ensure services are registered in `main.go`
- Check service methods are exported (capitalized)
- Verify bindings are up to date
