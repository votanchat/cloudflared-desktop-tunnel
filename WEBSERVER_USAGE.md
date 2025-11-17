# Web Server with Tunnel - Using Gin Framework

This guide explains the new separation of concerns for web server and tunnel management.

## Architecture

The application now has two independent, well-separated components:

### 1. TunnelManager (tunnel.go)
- Manages cloudflared tunnel lifecycle
- Handles binary download and caching
- Manages tunnel processes and logs
- Responsibilities: Tunnel authentication, routing, process management

### 2. WebServerManager (webserver.go)
- Manages Gin HTTP server
- Auto-detects available ports
- Handles HTTP routing with Gin
- Responsibilities: Web server lifecycle, HTTP handling, status pages

### 3. App (app.go)
- Orchestrates both TunnelManager and WebServerManager
- Provides unified API for combined operations
- Manages integration between components

## Features

- **Gin Framework**: Modern, fast HTTP framework with middleware support
- **Automatic Port Assignment**: OS assigns random available ports
- **Clean Separation**: Tunnel and web server are independent
- **Easy Maintenance**: Each component has a single responsibility
- **Beautiful Status Page**: Styled HTML response showing server status
- **Health Endpoint**: `/health` endpoint for monitoring

## JavaScript/Frontend Usage

```javascript
// Start web server with tunnel
async function startTunnel() {
  try {
    const result = await window.go.app.App.StartWebServerWithTunnel("your-token-here");
    console.log("Tunnel started:");
    console.log("  Port:", result.port);
    console.log("  Service URL:", result.url);
    console.log("  Status:", result.status);
  } catch (error) {
    console.error("Failed to start tunnel:", error);
  }
}

// Get web server status
async function checkStatus() {
  const status = await window.go.app.App.GetWebServerStatus();
  console.log("Running:", status.running);
  console.log("Port:", status.port);
}

// Stop web server and tunnel
async function stopTunnel() {
  try {
    await window.go.app.App.StopWebServerWithTunnel();
    console.log("Tunnel stopped");
  } catch (error) {
    console.error("Failed to stop tunnel:", error);
  }
}
```

## Go Backend Usage

```go
// In your app struct methods

// Start both web server and tunnel
result, err := a.StartWebServerWithTunnel("your-token-here")
if err != nil {
  log.Printf("Error: %v", err)
  return
}
log.Printf("Server running on port: %v", result["port"])

// Check status
status := a.GetWebServerStatus()
if status["running"].(bool) {
  log.Printf("Port: %d", status["port"].(int))
}

// Stop
err = a.StopWebServerWithTunnel()
```

## API Reference

### WebServerManager Methods

#### Start() (int, error)
Starts the Gin web server on a random available port.

**Returns:**
- Port number (int)
- Error (if any)

```go
ws := NewWebServerManager()
port, err := ws.Start()
if err != nil {
  log.Fatal(err)
}
log.Printf("Server running on port: %d", port)
```

#### Stop() error
Gracefully stops the web server.

#### IsRunning() bool
Returns whether the server is currently running.

#### GetPort() int
Returns the port number the server is running on.

#### setupHTMLTemplate()
Configures HTML routes (called automatically by App).

### App Methods

#### StartWebServerWithTunnel(manualToken string) (map[string]interface{}, error)
Starts both web server and tunnel in one call.

**Parameters:**
- `manualToken`: Cloudflare tunnel token (if empty, fetches from backend)

**Returns:**
- `port` (int): The assigned port
- `status` (string): "running"
- `url` (string): Local service URL

#### StopWebServerWithTunnel() error
Stops both web server and tunnel gracefully.

#### GetWebServerStatus() map[string]interface{}
Returns current web server status.

**Returns:**
- `running` (bool): Is server running
- `port` (int): Current port (0 if not running)

## Endpoints

### GET /
Status page with beautiful UI showing:
- Port number
- Running status indicator
- Tunnel information

### GET /health
JSON health check endpoint
```json
{
  "status": "ok",
  "port": 54321
}
```

### GET /* (catch-all)
Any other path returns the status page

## Configuration & Customization

### Customize Status Page
Edit `statusPageHandler()` in `webserver.go` to customize the HTML response:

```go
func (ws *WebServerManager) statusPageHandler(c *gin.Context) {
	// Your custom HTML here
}
```

### Add Custom Routes
Extend `setupRoutes()` to add more endpoints:

```go
ws.engine.POST("/api/data", func(c *gin.Context) {
	// Handle POST request
})
```

### Enable Gin Debug Mode
For development:

```go
gin.SetMode(gin.DebugMode)
```

## Example Flow

1. **Initialize**: App creates WebServerManager and TunnelManager
2. **Start Server**: WebServerManager starts Gin on random port (e.g., 54321)
3. **Configure Route**: App adds tunnel route to `http://localhost:54321`
4. **Start Tunnel**: TunnelManager starts cloudflared with configuration
5. **Access**: Users access the tunnel URL, traffic is routed to the web server
6. **Stop**: Both components shut down gracefully on app exit

## File Structure

```
app/
├── app.go              # Orchestration layer
├── tunnel.go           # TunnelManager (tunnel logic)
├── webserver.go        # WebServerManager (web server logic)
├── config.go           # Configuration
├── backend_client.go   # Backend integration
```

## Benefits of This Architecture

- **Testability**: Each component can be tested independently
- **Reusability**: WebServerManager can be used without tunnels
- **Maintainability**: Changes to web server don't affect tunnel logic
- **Scalability**: Easy to add features to either component
- **Clarity**: Clear separation of concerns

## Notes

- Web server uses Gin framework for performance and middleware support
- Tunnel and web server run in separate goroutines
- Both components handle cleanup properly on shutdown
- Thread-safe operations using sync.RWMutex
