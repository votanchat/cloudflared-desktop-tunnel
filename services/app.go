package services

import (
	"context"
	"fmt"
)

// AppService orchestrates all services
type AppService struct {
	configService   *ConfigService
	tunnelService   *TunnelService
	backendService  *BackendService
	webServerService *WebServerService
	ctx             context.Context
}

// NewAppService creates a new app service
func NewAppService() *AppService {
	configService := NewConfigService()
	config, _ := configService.LoadConfig()

	tunnelService := NewTunnelService(config.TunnelName, configService)
	backendService := NewBackendService(config.BackendURL)
	webServerService := NewWebServerService()

	// Set up tunnel callback to auto-start web server
	tunnelService.SetOnTunnelStart(func() error {
		if webServerService.IsRunning() {
			return nil
		}
		port := config.WebServerPort
		if port <= 0 {
			port = 8080
		}
		return webServerService.Start(port)
	})

	app := &AppService{
		configService:   configService,
		tunnelService:   tunnelService,
		backendService:  backendService,
		webServerService: webServerService,
		ctx:             context.Background(),
	}

	// Start backend client
	backendService.Start()

	// Auto-start tunnel if configured
	if config.AutoStart {
		go app.autoStartTunnel()
	}

	return app
}

// autoStartTunnel automatically starts the tunnel if configured
func (s *AppService) autoStartTunnel() {
	token, err := s.backendService.FetchToken()
	if err != nil {
		return
	}

	if err := s.tunnelService.Start(token); err != nil {
		// Log error
	}
}

// StartTunnel starts the cloudflared tunnel
func (s *AppService) StartTunnel(manualToken string) error {
	if s.tunnelService.IsRunning() {
		return fmt.Errorf("tunnel is already running")
	}

	token, err := s.getToken(manualToken)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	return s.tunnelService.Start(token)
}

// getToken retrieves token from manual input or backend
func (s *AppService) getToken(manualToken string) (string, error) {
	if manualToken != "" {
		return manualToken, nil
	}

	token, err := s.backendService.FetchToken()
	if err != nil {
		return "", fmt.Errorf("failed to fetch token from backend: %w", err)
	}
	return token, nil
}

// StopTunnel stops the cloudflared tunnel
func (s *AppService) StopTunnel() error {
	return s.tunnelService.Stop()
}

// GetTunnelStatus returns the current tunnel status
func (s *AppService) GetTunnelStatus() map[string]interface{} {
	return s.tunnelService.GetStatus()
}

// GetConfig returns the current configuration
func (s *AppService) GetConfig() *Config {
	return s.configService.GetConfig()
}

// UpdateConfig updates the configuration
func (s *AppService) UpdateConfig(config *Config) error {
	return s.configService.Save(config)
}

// StartWebServer starts the web server
func (s *AppService) StartWebServer(port int) error {
	if port <= 0 {
		config := s.configService.GetConfig()
		port = config.WebServerPort
		if port <= 0 {
			port = 8080
		}
	}
	return s.webServerService.Start(port)
}

// StopWebServer stops the web server
func (s *AppService) StopWebServer() error {
	return s.webServerService.Stop()
}

// GetWebServerStatus returns the current web server status
func (s *AppService) GetWebServerStatus() map[string]interface{} {
	return s.webServerService.GetStatus()
}

// StartTunnelWithWebServer starts tunnel and automatically starts web server on success
func (s *AppService) StartTunnelWithWebServer(manualToken string) (map[string]interface{}, error) {
	if s.tunnelService.IsRunning() {
		return nil, fmt.Errorf("tunnel is already running")
	}

	if err := s.StartTunnel(manualToken); err != nil {
		return nil, fmt.Errorf("failed to start tunnel: %w", err)
	}

	config := s.configService.GetConfig()
	port := config.WebServerPort
	if port <= 0 {
		port = 8080
	}

	if s.webServerService.IsRunning() {
		return s.buildStatusResponse(true, true, port, ""), nil
	}

	if err := s.webServerService.Start(port); err != nil {
		return s.buildStatusResponse(true, false, port, err.Error()), nil
	}

	return s.buildStatusResponse(true, true, port, ""), nil
}

// buildStatusResponse creates a standardized status response
func (s *AppService) buildStatusResponse(tunnelRunning, webServerRunning bool, port int, errorMsg string) map[string]interface{} {
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

// StopAll stops both tunnel and web server
func (s *AppService) StopAll() error {
	if s.tunnelService != nil && s.tunnelService.IsRunning() {
		if err := s.tunnelService.Stop(); err != nil {
			// Log error
		}
	}

	if s.webServerService != nil && s.webServerService.IsRunning() {
		if err := s.webServerService.Stop(); err != nil {
			// Log error
		}
	}

	return nil
}

// Shutdown performs cleanup when app is shutting down
func (s *AppService) Shutdown() {
	s.StopAll()

	if s.backendService != nil {
		s.backendService.Stop()
	}

	if s.tunnelService != nil {
		s.tunnelService.Cleanup()
	}

	if s.configService != nil {
		config := s.configService.GetConfig()
		s.configService.Save(config)
	}
}

