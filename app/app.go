package app

import (
	"context"
	"fmt"
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

	// Initialize file logging first (only in build mode)
	// This must be called before any logging
	if err := InitFileLogging(); err != nil {
		// If file logging fails, continue with console logging only
		// This is expected in dev mode, so we don't log the error
	}

	appLogger.Info("Application starting up...")

	// Initialize configuration
	var err error
	a.config, err = LoadConfig()
	if err != nil {
		appLogger.Warn("Could not load config, using defaults: %v", err)
		a.config = DefaultConfig()
	}

	// Initialize backend client
	a.backendClient = NewBackendClient(a.config.BackendURL)
	go a.backendClient.Start(ctx)

	// Initialize tunnel manager
	a.tunnel = NewTunnelManager(a.config.TunnelName)
	a.tunnel.SetConfig(a.config)

	// Initialize web server manager
	a.webServer = NewWebServerManager()

	// Set callback to auto-start web server when tunnel starts successfully
	a.tunnel.SetOnTunnelStart(a.autoStartWebServer)

	// Auto-start tunnel if configured
	if a.config.AutoStart {
		go a.autoStartTunnel()
	}
}

// autoStartWebServer starts the web server when tunnel starts
func (a *App) autoStartWebServer() error {
	if a.webServer.IsRunning() {
		return nil
	}

	port := a.config.WebServerPort
	if port <= 0 {
		port = 8080
	}

	if err := a.webServer.StartWithPort(port); err != nil {
		appLogger.Warn("Failed to auto-start web server on port %d: %v", port, err)
		return nil
	}

	a.webServer.setupHTMLTemplate()
	appLogger.Info("Web server auto-started successfully on port %d", port)
	return nil
}

// autoStartTunnel automatically starts the tunnel if configured
func (a *App) autoStartTunnel() {
	token, err := a.backendClient.FetchToken()
	if err != nil {
		appLogger.Error("Failed to fetch token: %v", err)
		return
	}

	if err := a.tunnel.Start(token); err != nil {
		appLogger.Error("Failed to auto-start tunnel: %v", err)
	}
}

// DomReady is called after the front-end dom has been loaded
func (a *App) DomReady(ctx context.Context) {
	appLogger.Info("DOM is ready")
}

// Shutdown is called when the app is closing
func (a *App) Shutdown(ctx context.Context) {
	appLogger.Info("Application shutting down...")

	// Stop web server if running
	if a.webServer != nil && a.webServer.IsRunning() {
		if err := a.webServer.Stop(); err != nil {
			appLogger.Error("Error stopping web server: %v", err)
		}
	}

	// Stop tunnel if running
	if a.tunnel != nil {
		if a.tunnel.IsRunning() {
			if err := a.tunnel.Stop(); err != nil {
				appLogger.Error("Error stopping tunnel: %v", err)
			}
		}
		a.tunnel.Cleanup()
	}

	// Stop backend client
	if a.backendClient != nil {
		a.backendClient.Stop()
	}

	// Save configuration
	if a.config != nil {
		if err := a.config.Save(); err != nil {
			appLogger.Error("Error saving config: %v", err)
		}
	}

	// Close file logging
	CloseFileLogging()
}

// StartTunnel starts the cloudflared tunnel
// If manualToken is provided and not empty, it will be used instead of fetching from backend
func (a *App) StartTunnel(manualToken string) error {
	if a.tunnel.IsRunning() {
		return fmt.Errorf("tunnel is already running")
	}

	token, err := a.getToken(manualToken)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	return a.tunnel.Start(token)
}

// getToken retrieves token from manual input or backend
func (a *App) getToken(manualToken string) (string, error) {
	if manualToken != "" {
		appLogger.Info("Using manually provided token")
		return manualToken, nil
	}

	appLogger.Info("Fetching token from backend...")
	token, err := a.backendClient.FetchToken()
	if err != nil {
		return "", fmt.Errorf("failed to fetch token from backend: %w", err)
	}
	return token, nil
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

// Greet returns a greeting for the given name (kept for API compatibility)
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

	if err := a.StartTunnel(manualToken); err != nil {
		return nil, fmt.Errorf("failed to start tunnel: %w", err)
	}

	appLogger.Info("Tunnel started successfully, starting web server...")

	if a.webServer.IsRunning() {
		return nil, fmt.Errorf("web server is already running")
	}

	port := a.getWebServerPort()
	if err := a.webServer.StartWithPort(port); err != nil {
		appLogger.Warn("Failed to start web server on port %d: %v", port, err)
		return a.buildStatusResponse(true, false, port, err.Error()), nil
	}

	a.webServer.setupHTMLTemplate()
	appLogger.Info("Web server started successfully on port %d", port)

	return a.buildStatusResponse(true, true, port, ""), nil
}

// getWebServerPort returns the configured web server port or default
func (a *App) getWebServerPort() int {
	if a.config.WebServerPort > 0 {
		return a.config.WebServerPort
	}
	return 8080
}

// buildStatusResponse creates a standardized status response
func (a *App) buildStatusResponse(tunnelRunning, webServerRunning bool, port int, errorMsg string) map[string]interface{} {
	status := map[string]interface{}{
		"tunnelRunning":    tunnelRunning,
		"webServerRunning": webServerRunning,
		"port":             port,
	}

	if tunnelRunning {
		status["tunnel"] = "running"
	} else {
		status["tunnel"] = "stopped"
	}

	if webServerRunning {
		status["webServer"] = "running"
		status["status"] = "running"
	} else {
		status["webServer"] = "failed"
		if errorMsg != "" {
			status["error"] = errorMsg
		}
	}

	return status
}

// StartWebServerWithTunnel starts a web server with Gin and creates a tunnel to it
func (a *App) StartWebServerWithTunnel(manualToken string) (map[string]interface{}, error) {
	if a.webServer.IsRunning() {
		return nil, fmt.Errorf("web server is already running")
	}

	port, err := a.webServer.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start web server: %w", err)
	}

	a.webServer.setupHTMLTemplate()

	if err := a.StartTunnel(manualToken); err != nil {
		a.webServer.Stop()
		return nil, fmt.Errorf("failed to start tunnel: %w", err)
	}

	appLogger.Info("Web server and tunnel started successfully on port %d", port)

	return map[string]interface{}{
		"port":   port,
		"status": "running",
		"url":    fmt.Sprintf("http://localhost:%d", port),
	}, nil
}

// StopWebServerWithTunnel stops both the web server and tunnel
func (a *App) StopWebServerWithTunnel() error {
	if a.tunnel != nil && a.tunnel.IsRunning() {
		if err := a.tunnel.Stop(); err != nil {
			appLogger.Error("Error stopping tunnel: %v", err)
		}
	}

	if a.webServer != nil && a.webServer.IsRunning() {
		if err := a.webServer.Stop(); err != nil {
			appLogger.Error("Error stopping web server: %v", err)
		}
	}

	appLogger.Info("Web server and tunnel stopped")
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
