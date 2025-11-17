package services

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Route represents a tunnel route configuration
type Route struct {
	Hostname string `json:"hostname"`
	Service  string `json:"service"`
}

// Config represents the application configuration
type Config struct {
	BackendURL      string  `json:"backendURL"`
	TunnelName      string  `json:"tunnelName"`
	AutoStart       bool    `json:"autoStart"`
	MinimizeToTray  bool    `json:"minimizeToTray"`
	RefreshInterval int     `json:"refreshInterval"`
	WebServerPort   int     `json:"webServerPort"`
	Routes          []Route  `json:"routes"`
}

// ConfigService manages application configuration
type ConfigService struct {
	config *Config
}

// NewConfigService creates a new config service
func NewConfigService() *ConfigService {
	return &ConfigService{}
}

// LoadConfig loads configuration from file
func (s *ConfigService) LoadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := s.DefaultConfig()
		s.config = config
		return config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	s.config = &config
	return &config, nil
}

// Save saves the configuration to file
func (s *ConfigService) Save(config *Config) error {
	s.config = config
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// DefaultConfig returns a default configuration
func (s *ConfigService) DefaultConfig() *Config {
	return &Config{
		BackendURL:      "https://api.example.com",
		TunnelName:      "my-tunnel",
		AutoStart:       false,
		MinimizeToTray:  true,
		RefreshInterval: 300,
		WebServerPort:   8080,
		Routes:          []Route{},
	}
}

// GetConfig returns the current configuration
func (s *ConfigService) GetConfig() *Config {
	if s.config == nil {
		return s.DefaultConfig()
	}
	return s.config
}

// AddRoute adds a new route to the configuration
func (s *ConfigService) AddRoute(hostname, service string) error {
	config := s.GetConfig()
	for i := range config.Routes {
		if config.Routes[i].Hostname == hostname {
			config.Routes[i].Service = service
			return s.Save(config)
		}
	}
	config.Routes = append(config.Routes, Route{
		Hostname: hostname,
		Service:  service,
	})
	return s.Save(config)
}

// RemoveRoute removes a route by hostname
func (s *ConfigService) RemoveRoute(hostname string) error {
	config := s.GetConfig()
	for i := range config.Routes {
		if config.Routes[i].Hostname == hostname {
			config.Routes = append(config.Routes[:i], config.Routes[i+1:]...)
			return s.Save(config)
		}
	}
	return nil
}

// GetRoutes returns all configured routes
func (s *ConfigService) GetRoutes() []Route {
	return s.GetConfig().Routes
}

// getConfigPath returns the path to the config file
func getConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	appConfigDir := filepath.Join(configDir, "cloudflared-desktop-tunnel-v3")
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(appConfigDir, "config.json"), nil
}

