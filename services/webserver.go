package services

import (
	"fmt"
	"net"
	"sync"

	"github.com/gin-gonic/gin"
)

// WebServerService manages the Gin web server
type WebServerService struct {
	mu       sync.RWMutex
	engine   *gin.Engine
	running  bool
	port     int
	listener net.Listener
}

// NewWebServerService creates a new web server service
func NewWebServerService() *WebServerService {
	gin.SetMode(gin.ReleaseMode)

	return &WebServerService{
		engine: gin.Default(),
	}
}

// Start starts the web server on a specific port
func (s *WebServerService) Start(port int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("web server is already running")
	}

	if port <= 0 {
		return fmt.Errorf("invalid port number: %d (must be > 0)", port)
	}

	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start web server on port %d: port in use or unavailable", port)
	}

	s.port = port
	s.listener = listener

	s.setupRoutes()

	go func() {
		if err := s.engine.RunListener(listener); err != nil && err.Error() != "http: Server closed" {
			// Log error
		}
	}()

	s.running = true
	return nil
}

// Stop stops the web server gracefully
func (s *WebServerService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("web server is not running")
	}

	s.running = false
	if s.listener != nil {
		s.listener.Close()
	}
	return nil
}

// IsRunning returns true if the web server is running
func (s *WebServerService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetPort returns the port the web server is running on
func (s *WebServerService) GetPort() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.port
}

// GetStatus returns the current web server status
func (s *WebServerService) GetStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"running": s.running,
		"port":    s.port,
	}
}

// setupRoutes sets up all HTTP routes
func (s *WebServerService) setupRoutes() {
	s.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"port":   s.port,
		})
	})

	s.engine.GET("/", s.statusPageHandler)
	s.engine.NoRoute(s.statusPageHandler)
}

// statusPageHandler returns the HTML status page
func (s *WebServerService) statusPageHandler(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	port := s.port
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

