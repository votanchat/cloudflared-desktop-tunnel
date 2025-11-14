# Architecture Documentation

## Overview

Cloudflared Desktop Tunnel is a cross-platform desktop application built with Wails v2 that manages Cloudflare Tunnels. It embeds the `cloudflared` binary directly into the application, eliminating external dependencies.

## System Architecture

```
┌────────────────────────────────────────────────┐
│                  Frontend (React + TypeScript)               │
│  ┌──────────────────────────────────────────┐  │
│  │  TunnelManager  │  StatusDisplay  │  Settings  │  │
│  └──────────────────────────────────────────┘  │
└────────────────────────────────────────────────┘
                         │ Wails Bindings
                         │ (TypeScript ↔ Go)
                         │
┌────────────────────────────────────────────────┐
│                 Backend (Go)                              │
│  ┌──────────────────────────────────────────┐  │
│  │           App (Lifecycle)                    │  │
│  └──────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────┐  │
│  │        TunnelManager                      │  │
│  │  - Extract embedded binary              │  │
│  │  - Start/stop cloudflared process       │  │
│  │  - Monitor logs                         │  │
│  └──────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────┐  │
│  │        BackendClient                      │  │
│  │  - Fetch tunnel tokens                  │  │
│  │  - WebSocket for commands               │  │
│  │  - Periodic token refresh               │  │
│  └──────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────┐  │
│  │           Config                          │  │
│  │  - Load/save configuration              │  │
│  │  - Persistent storage                   │  │
│  └──────────────────────────────────────────┘  │
└────────────────────────────────────────────────┘
                         │
                         │ Process Execution
                         │
┌────────────────────────────────────────────────┐
│         Embedded cloudflared Binary                      │
│  (Platform-specific via go:embed + build tags)         │
└────────────────────────────────────────────────┘
                         │
                         │ Tunnel Connection
                         │
              ┌─────────────────────┐
              │  Cloudflare Edge   │
              └─────────────────────┘
```

## Component Details

### 1. Frontend Layer (React + TypeScript)

**Location**: `frontend/src/`

**Key Components**:
- **App.tsx**: Main application container with tab navigation
- **TunnelManager.tsx**: Tunnel control interface (start/stop)
- **StatusDisplay.tsx**: Real-time status and logs viewer
- **Settings.tsx**: Configuration management UI

**Communication**:
- Uses Wails runtime to call Go backend methods
- Automatically generated TypeScript bindings from Go code
- Type-safe communication with backend

### 2. Backend Layer (Go)

**Location**: `app/`

#### App (app.go)
- Manages application lifecycle (startup, shutdown)
- Coordinates between TunnelManager, BackendClient, and Config
- Exposes methods to frontend via Wails bindings

**Lifecycle Hooks**:
```go
Startup(ctx)   -> Initialize components
DomReady(ctx)  -> Frontend is ready
Shutdown(ctx)  -> Clean up resources
```

#### TunnelManager (tunnel.go)
- Manages cloudflared tunnel process
- Extracts embedded binary to temp directory
- Monitors process output and status
- Handles process lifecycle

**Key Methods**:
```go
Start(token string) error  // Start tunnel with token
Stop() error              // Stop tunnel
IsRunning() bool          // Check if running
GetLogs() []string        // Get recent logs
```

#### BackendClient (backend_client.go)
- HTTP client for REST API calls
- WebSocket client for real-time commands
- Automatic token refresh mechanism
- Command processor for remote operations

**Backend API Endpoints**:
```
GET  /api/token        -> Fetch tunnel token
POST /api/status       -> Report tunnel status
WS   /api/commands     -> Real-time commands
```

#### Config (config.go)
- Persistent configuration storage
- JSON-based config file
- Platform-specific config directory

**Config Location**:
- **Windows**: `%APPDATA%\cloudflared-desktop-tunnel\config.json`
- **macOS**: `~/Library/Application Support/cloudflared-desktop-tunnel/config.json`
- **Linux**: `~/.config/cloudflared-desktop-tunnel/config.json`

### 3. Binary Embedding Layer

**Location**: `binaries/`

**Strategy**:
- Uses Go 1.16+ `embed` directive
- Platform-specific files with build tags
- Runtime OS/architecture detection
- Automatic binary selection

**Build Tags**:
```go
//go:build windows   -> embed_windows.go
//go:build darwin    -> embed_darwin.go
//go:build linux     -> embed_linux.go
```

**Binary Selection Flow**:
```
1. Application starts
2. Build tag selects correct embed file
3. init() function detects GOARCH
4. Sets CloudflaredBinary variable
5. TunnelManager extracts to temp directory
6. Sets executable permissions (Unix)
7. Executes binary with token
```

## Data Flow

### Starting a Tunnel

```
1. User clicks "Start Tunnel" button
   ↓
2. TunnelManager.Start() called from frontend
   ↓
3. BackendClient.FetchToken() gets token from API
   ↓
4. TunnelManager extracts embedded binary
   ↓
5. cloudflared process starts with token
   ↓
6. Process logs streamed to frontend
   ↓
7. Status updates every 3 seconds
```

### Backend Commands

```
1. Backend sends command via WebSocket
   ↓
2. BackendClient receives message
   ↓
3. Command added to commandsCh channel
   ↓
4. processCommands() handles command
   ↓
5. Action executed (update, restart, etc.)
   ↓
6. Status reported back to backend
```

## Cross-Platform Considerations

### Binary Extraction
- **Windows**: Extract to `%TEMP%\cloudflared-windows-amd64.exe`
- **macOS**: Extract to `/tmp/cloudflared-darwin-{arch}`
- **Linux**: Extract to `/tmp/cloudflared-linux-{arch}`

### File Permissions
- Unix systems: `chmod +x` automatically applied
- Windows: No additional permissions needed

### Process Management
- Cross-platform using Go's `os/exec` package
- Graceful shutdown with process.Kill()
- Cleanup on application exit

## Security Considerations

1. **Token Storage**: Tokens are never persisted to disk
2. **Temporary Files**: Cleaned up on shutdown
3. **Binary Integrity**: Use official Cloudflare binaries only
4. **WebSocket Auth**: Implement authentication for production
5. **Config File**: Contains no sensitive data

## Performance

### Memory Usage
- Go backend: ~20-30 MB
- React frontend: ~50-70 MB
- Embedded binary: ~45 MB (per platform)
- cloudflared process: ~30-50 MB
- **Total**: ~150-200 MB

### Binary Size
- Without embedded binaries: ~15-20 MB
- With embedded binary: ~60-70 MB per platform
- Multiple architectures increase size proportionally

## Future Enhancements

1. **Auto-Update**: Download new cloudflared versions
2. **Multi-Tunnel**: Manage multiple tunnels simultaneously
3. **Advanced Logs**: Log filtering and search
4. **Metrics**: Display bandwidth and connection metrics
5. **System Tray**: Minimize to tray on close
6. **Notifications**: Desktop notifications for events
7. **Health Checks**: Automatic reconnection on failure
8. **Config Profiles**: Multiple configuration profiles
