package app

import (
	"context"
	"fmt"
	"log"
)

// App struct
type App struct {
	ctx           context.Context
	tunnel        *TunnelManager
	config        *Config
	backendClient *BackendClient
	webServer     *WebServerManager
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// Startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	log.Println("Application starting up...")

	// Initialize configuration
	var err error
	a.config, err = LoadConfig()
	if err != nil {
		log.Printf("Warning: Could not load config, using defaults: %v", err)
		a.config = DefaultConfig()
	}

	// Initialize backend client
	a.backendClient = NewBackendClient(a.config.BackendURL)
	go a.backendClient.Start(ctx)

	// Initialize tunnel manager
	a.tunnel = NewTunnelManager(a.config.TunnelName)
	// Set config reference for route management
	a.tunnel.SetConfig(a.config)

	// Initialize web server manager
	a.webServer = NewWebServerManager()

	// Set callback to auto-start web server when tunnel starts successfully
	a.tunnel.SetOnTunnelStart(func() error {
		if !a.webServer.IsRunning() {
			// Use fixed port from config
			port := a.config.WebServerPort
			if port <= 0 {
				port = 8080 // Fallback to default port
			}

			err := a.webServer.StartWithPort(port)
			if err != nil {
				log.Printf("Warning: Failed to auto-start web server on port %d: %v", port, err)
				return nil // Don't fail tunnel if web server fails
			}
			a.webServer.setupHTMLTemplate()
			log.Printf("Web server auto-started successfully on port %d", port)
		}
		return nil
	})

	// Auto-start tunnel if configured
	if a.config.AutoStart {
		go func() {
			token, err := a.backendClient.FetchToken()
			if err != nil {
				log.Printf("Failed to fetch token: %v", err)
				return
			}
			if err := a.tunnel.Start(token); err != nil {
				log.Printf("Failed to auto-start tunnel: %v", err)
			}
		}()
	}
}

// DomReady is called after the front-end dom has been loaded
func (a *App) DomReady(ctx context.Context) {
	log.Println("DOM is ready")
}

// Shutdown is called when the app is closing
func (a *App) Shutdown(ctx context.Context) {
	log.Println("Application shutting down...")

	// Stop web server if running
	if a.webServer != nil && a.webServer.IsRunning() {
		if err := a.webServer.Stop(); err != nil {
			log.Printf("Error stopping web server: %v", err)
		}
	}

	// Stop tunnel if running
	if a.tunnel != nil {
		if a.tunnel.IsRunning() {
			if err := a.tunnel.Stop(); err != nil {
				log.Printf("Error stopping tunnel: %v", err)
			}
		}
		// Clean up cached binary
		a.tunnel.Cleanup()
	}

	// Stop backend client
	if a.backendClient != nil {
		a.backendClient.Stop()
	}

	// Save configuration
	if a.config != nil {
		if err := a.config.Save(); err != nil {
			log.Printf("Error saving config: %v", err)
		}
	}
}

// StartTunnel starts the cloudflared tunnel
// If manualToken is provided and not empty, it will be used instead of fetching from backend
func (a *App) StartTunnel(manualToken string) error {
	if a.tunnel.IsRunning() {
		return fmt.Errorf("tunnel is already running")
	}

	var token string
	var err error

	// Check if manual token is provided
	if manualToken != "" {
		log.Println("Using manually provided token")
		token = manualToken
	} else {
		// Fetch token from backend
		log.Println("Fetching token from backend...")
		token, err = a.backendClient.FetchToken()
		if err != nil {
			return fmt.Errorf("failed to fetch token from backend: %w", err)
		}
	}

	return a.tunnel.Start(token)
}

// StopTunnel stops the cloudflared tunnel
func (a *App) StopTunnel() error {
	return a.tunnel.Stop()
}

// GetTunnelStatus returns the current tunnel status
func (a *App) GetTunnelStatus() map[string]interface{} {
	tunnelURL := a.tunnel.GetTunnelURL()
	return map[string]interface{}{
		"running":    a.tunnel.IsRunning(),
		"tunnelName": a.config.TunnelName,
		"tunnelURL":  tunnelURL,
		"logs":       a.tunnel.GetLogs(),
	}
}

// GetConfig returns the current configuration
func (a *App) GetConfig() *Config {
	return a.config
}

// UpdateConfig updates the configuration
func (a *App) UpdateConfig(config *Config) error {
	a.config = config
	return a.config.Save()
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, welcome to Cloudflared Desktop Tunnel!", name)
}

// AddRoute adds a route to the configuration
func (a *App) AddRoute(hostname, service string) error {
	a.config.AddRoute(hostname, service)
	if err := a.config.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	// Update tunnel config reference
	a.tunnel.SetConfig(a.config)
	return nil
}

// RemoveRoute removes a route from the configuration
func (a *App) RemoveRoute(hostname string) error {
	if !a.config.RemoveRoute(hostname) {
		return fmt.Errorf("route not found: %s", hostname)
	}
	if err := a.config.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	// Update tunnel config reference
	a.tunnel.SetConfig(a.config)
	return nil
}

// GetRoutes returns all configured routes
func (a *App) GetRoutes() []Route {
	return a.config.Routes
}

// StartTunnelWithWebServer starts tunnel and automatically starts web server on success
func (a *App) StartTunnelWithWebServer(manualToken string) (map[string]interface{}, error) {
	if a.tunnel.IsRunning() {
		return nil, fmt.Errorf("tunnel is already running")
	}

	// Start tunnel first (which also sets up the web server route)
	if err := a.StartTunnel(manualToken); err != nil {
		return nil, fmt.Errorf("failed to start tunnel: %w", err)
	}

	// Tunnel started successfully, now start web server
	log.Println("Tunnel started successfully, starting web server...")

	if a.webServer.IsRunning() {
		return nil, fmt.Errorf("web server is already running")
	}

	// Start web server on configured port
	port := a.config.WebServerPort
	if port <= 0 {
		port = 8080
	}

	err := a.webServer.StartWithPort(port)
	if err != nil {
		log.Printf("Warning: Failed to start web server on port %d: %v", port, err)
		return map[string]interface{}{
			"tunnel":           "running",
			"webServer":        "failed",
			"port":             0,
			"error":            err.Error(),
			"tunnelRunning":    true,
			"webServerRunning": false,
		}, nil
	}

	// Setup HTML routes for the web server
	a.webServer.setupHTMLTemplate()

	log.Printf("Web server started successfully on port %d", port)

	return map[string]interface{}{
		"tunnel":           "running",
		"webServer":        "running",
		"port":             port,
		"status":           "running",
		"tunnelRunning":    true,
		"webServerRunning": true,
	}, nil
}

// StartWebServerWithTunnel starts a web server with Gin and creates a tunnel to it
func (a *App) StartWebServerWithTunnel(manualToken string) (map[string]interface{}, error) {
	if a.webServer.IsRunning() {
		return nil, fmt.Errorf("web server is already running")
	}

	// Start web server
	port, err := a.webServer.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start web server: %w", err)
	}

	// Setup HTML routes for the web server
	a.webServer.setupHTMLTemplate()

	// Start the tunnel
	tunnelService := fmt.Sprintf("http://localhost:%d", port)
	if err := a.StartTunnel(manualToken); err != nil {
		a.webServer.Stop()
		return nil, fmt.Errorf("failed to start tunnel: %w", err)
	}

	log.Printf("Web server and tunnel started successfully on port %d", port)

	return map[string]interface{}{
		"port":   port,
		"status": "running",
		"url":    tunnelService,
	}, nil
}

// StopWebServerWithTunnel stops both the web server and tunnel
func (a *App) StopWebServerWithTunnel() error {
	// Stop tunnel first
	if a.tunnel != nil && a.tunnel.IsRunning() {
		if err := a.tunnel.Stop(); err != nil {
			log.Printf("Error stopping tunnel: %v", err)
		}
	}

	// Stop web server
	if a.webServer != nil && a.webServer.IsRunning() {
		if err := a.webServer.Stop(); err != nil {
			log.Printf("Error stopping web server: %v", err)
		}
	}

	log.Println("Web server and tunnel stopped")
	return nil
}

// GetWebServerStatus returns the current web server status
func (a *App) GetWebServerStatus() map[string]interface{} {
	running := false
	port := 0

	if a.webServer != nil {
		running = a.webServer.IsRunning()
		if running {
			port = a.webServer.GetPort()
		}
	}

	return map[string]interface{}{
		"running": running,
		"port":    port,
	}
}
