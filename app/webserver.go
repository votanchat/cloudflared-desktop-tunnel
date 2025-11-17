package app

import (
	"fmt"
	"net"
	"sync"

	"github.com/gin-gonic/gin"
)

// WebServerManager manages the Gin web server
type WebServerManager struct {
	mu       sync.RWMutex
	engine   *gin.Engine
	running  bool
	port     int
	listener net.Listener
}

// NewWebServerManager creates a new web server manager
func NewWebServerManager() *WebServerManager {
	// Disable Gin debug logging in production
	gin.SetMode(gin.ReleaseMode)

	return &WebServerManager{
		engine: gin.Default(),
	}
}

// StartWithPort starts the web server on a specific port
func (ws *WebServerManager) StartWithPort(port int) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.running {
		return fmt.Errorf("web server is already running")
	}

	if port <= 0 {
		return fmt.Errorf("invalid port number: %d (must be > 0)", port)
	}

	// Try to listen on the specified port
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start web server on port %d: port in use or unavailable", port)
	}

	ws.port = port
	ws.listener = listener

	serverLogger.Info("Web server will use port: %d", port)
	ws.setupRoutes()

	go func() {
		serverLogger.Info("Starting Gin web server on port %d", port)
		if err := ws.engine.RunListener(listener); err != nil && err.Error() != "http: Server closed" {
			serverLogger.Error("Web server error: %v", err)
		}
	}()

	ws.running = true
	return nil
}

// Start starts the web server on a random available port (deprecated, use StartWithPort)
func (ws *WebServerManager) Start() (int, error) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.running {
		return 0, fmt.Errorf("web server is already running")
	}

	// Find an available port
	listener, err := net.Listen("tcp", ":0") // 0 means OS assigns a random available port
	if err != nil {
		return 0, fmt.Errorf("failed to find available port: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	ws.port = port
	ws.listener = listener

	serverLogger.Info("Web server will use port: %d", port)
	ws.setupRoutes()

	go func() {
		serverLogger.Info("Starting Gin web server on port %d", port)
		if err := ws.engine.RunListener(listener); err != nil && err.Error() != "http: Server closed" {
			serverLogger.Error("Web server error: %v", err)
		}
	}()

	ws.running = true
	return port, nil
}

// Stop stops the web server gracefully
func (ws *WebServerManager) Stop() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if !ws.running {
		return fmt.Errorf("web server is not running")
	}

	serverLogger.Info("Stopping web server on port %d", ws.port)
	ws.running = false
	return nil
}

// IsRunning returns true if the web server is running
func (ws *WebServerManager) IsRunning() bool {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.running
}

// GetPort returns the port the web server is running on
func (ws *WebServerManager) GetPort() int {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.port
}

// setupRoutes sets up all HTTP routes
func (ws *WebServerManager) setupRoutes() {
	// Health check endpoint
	ws.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"port":   ws.port,
		})
	})

	// Main status page
	ws.engine.GET("/", ws.statusPageHandler)

	// Catch-all for any other request
	ws.engine.NoRoute(ws.statusPageHandler)
}

// statusPageHandler returns the HTML status page
func (ws *WebServerManager) statusPageHandler(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	port := ws.port
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<title>Cloudflared Desktop Tunnel</title>
	<style>
		body {
			font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
			text-align: center;
			margin: 0;
			padding: 0;
			background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
			min-height: 100vh;
			display: flex;
			align-items: center;
			justify-content: center;
		}
		.container {
			background: white;
			padding: 40px;
			border-radius: 10px;
			box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
			max-width: 500px;
		}
		h1 {
			color: #333;
			margin-top: 0;
		}
		.info {
			background: #f0f0f0;
			padding: 20px;
			border-radius: 5px;
			margin: 20px 0;
		}
		.port-number {
			font-size: 24px;
			font-weight: bold;
			color: #667eea;
		}
		.status-badge {
			display: inline-block;
			background: #4caf50;
			color: white;
			padding: 8px 16px;
			border-radius: 20px;
			margin-top: 15px;
			font-size: 14px;
		}
	</style>
</head>
<body>
	<div class="container">
		<h1>🚀 Web Server Running</h1>
		<div class="info">
			<p><strong>Port:</strong> <span class="port-number">%d</span></p>
			<p>Your cloudflared tunnel is active and forwarding traffic to this server.</p>
			<div class="status-badge">✓ Active</div>
		</div>
	</div>
</body>
</html>`, port)
	c.String(200, html)
}

// setupHTMLTemplate is called after server start to set up routes
func (ws *WebServerManager) setupHTMLTemplate() {
	// Routes are already set up in setupRoutes()
	// This function is kept for API compatibility
	serverLogger.Debug("Web server routes configured for port %d", ws.port)
}
