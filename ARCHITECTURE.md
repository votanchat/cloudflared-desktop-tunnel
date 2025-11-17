# Architecture: Auto Web Server Start

## Overview

When tunnel starts successfully, the web server **automatically starts** on a random available port.

## How It Works

### 1. Initialization (app.go - Startup)

```
App.Startup()
├── Create TunnelManager
├── Create WebServerManager
└── Register Callback: TunnelManager.SetOnTunnelStart()
    └── Callback: Auto-start WebServer when tunnel starts
```

### 2. User Starts Tunnel (Frontend)

```javascript
window.go.app.App.StartTunnel(token)
```

### 3. Tunnel Start Process

```
TunnelManager.Start(token)
├── Download/prepare cloudflared binary
├── Start cloudflared process
├── Set running = true
├── Start log readers
├── Call callback: onTunnelStart()
│   └── WebServerManager.Start() → Auto-start web server
│       └── WebServerManager.setupHTMLTemplate() → Setup routes
└── Return success
```

### 4. Auto Web Server Start

When `onTunnelStart` callback is invoked:

```go
// In app.go Startup()
a.tunnel.SetOnTunnelStart(func() error {
    port, err := a.webServer.Start()
    // Setup routes
    a.webServer.setupHTMLTemplate()
    return nil
})
```

This runs in a **separate goroutine** so:
- Tunnel startup completes immediately
- Web server starts in parallel
- No blocking between components

## Flow Diagram

```
User clicks "Start Tunnel"
    ↓
Frontend calls: App.StartTunnel(token)
    ↓
Backend: TunnelManager.Start(token)
    ├── Start cloudflared process
    ├── Set running = true
    └── Call callback in goroutine
        ↓
        WebServerManager.Start()
        ├── Find available port
        ├── Start Gin server
        ├── setupHTMLTemplate()
        └── Ready for requests
    ↓
Tunnel + Web Server both running
```

## Key Methods

### TunnelManager

```go
// Set callback to run when tunnel starts
func (tm *TunnelManager) SetOnTunnelStart(callback OnTunnelStart)

// Type definition
type OnTunnelStart func() error
```

### WebServerManager

```go
// Start web server on random port
func (ws *WebServerManager) Start() (int, error)

// Setup HTML routes
func (ws *WebServerManager) setupHTMLTemplate()
```

### App

```go
// Manual start both: web server → tunnel
func (a *App) StartWebServerWithTunnel(token string) (map[string]interface{}, error)

// Start tunnel, auto-starts web server via callback
func (a *App) StartTunnel(token string) error

// Start tunnel + auto web server (same as StartTunnel with callback)
func (a *App) StartTunnelWithWebServer(token string) (map[string]interface{}, error)

// Stop both gracefully
func (a *App) StopWebServerWithTunnel() error
```

## Usage Scenarios

### Scenario 1: Simple Tunnel Start (Default)
User clicks "Start Tunnel" → Tunnel starts → Web server auto-starts

```javascript
await window.go.app.App.StartTunnel(token);
// Both tunnel and web server are now running
```

### Scenario 2: Start Web Server Without Tunnel
(For development/testing)

```go
port, err := a.webServer.Start()
a.webServer.setupHTMLTemplate()
// Web server running on specified port, no tunnel
```

### Scenario 3: Manual Control (Web Server First)
Start web server, then add tunnel route, then start tunnel

```javascript
const result = await window.go.app.App.StartWebServerWithTunnel(token);
// Web server starts first, then tunnel
```

## Thread Safety

- **TunnelManager.mu** - RWMutex protects running state
- **WebServerManager.mu** - RWMutex protects running state
- **Callback** runs in separate goroutine to avoid blocking

## Error Handling

```go
// If web server fails to start, tunnel continues running
a.tunnel.SetOnTunnelStart(func() error {
    port, err := a.webServer.Start()
    if err != nil {
        log.Printf("Warning: Web server failed: %v", err)
        return nil  // Don't fail tunnel
    }
    return nil
})
```

## Shutdown Process

```
App.Shutdown()
├── Stop web server (if running)
├── Stop tunnel (if running)
├── Cleanup cached binary
├── Stop backend client
└── Save configuration
```

## Benefits

1. **User Experience**: One click starts both components
2. **Separation of Concerns**: Each component handles its responsibility
3. **Flexibility**: Can start them independently if needed
4. **Error Resilience**: Web server failure doesn't break tunnel
5. **Clean Code**: Callback pattern is clear and maintainable

## Future Enhancements

- Add config option to disable auto web server start
- Add metrics/monitoring for both components
- Add health check endpoint that monitors both
- Support multiple web server instances
