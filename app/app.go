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

	// Stop tunnel if running
	if a.tunnel != nil && a.tunnel.IsRunning() {
		if err := a.tunnel.Stop(); err != nil {
			log.Printf("Error stopping tunnel: %v", err)
		}
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
func (a *App) StartTunnel() error {
	if a.tunnel.IsRunning() {
		return fmt.Errorf("tunnel is already running")
	}

	// Fetch token from backend
	token, err := a.backendClient.FetchToken()
	if err != nil {
		return fmt.Errorf("failed to fetch token: %w", err)
	}

	return a.tunnel.Start(token)
}

// StopTunnel stops the cloudflared tunnel
func (a *App) StopTunnel() error {
	return a.tunnel.Stop()
}

// GetTunnelStatus returns the current tunnel status
func (a *App) GetTunnelStatus() map[string]interface{} {
	return map[string]interface{}{
		"running":    a.tunnel.IsRunning(),
		"tunnelName": a.config.TunnelName,
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
