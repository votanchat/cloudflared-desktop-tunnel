package app

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Route represents a tunnel route configuration
type Route struct {
	Hostname string `json:"hostname"` // e.g., "myapp.example.com"
	Service  string `json:"service"`  // e.g., "http://localhost:3000"
}

// Config represents the application configuration
type Config struct {
	BackendURL      string  `json:"backendURL"`
	TunnelName      string  `json:"tunnelName"`
	AutoStart       bool    `json:"autoStart"`
	MinimizeToTray  bool    `json:"minimizeToTray"`
	RefreshInterval int     `json:"refreshInterval"` // in seconds
	WebServerPort   int     `json:"webServerPort"`   // Web server port (0 = random, >0 = fixed)
	Routes          []Route `json:"routes"`          // Domain routes for tunnel
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		BackendURL:      "https://api.example.com",
		TunnelName:      "my-tunnel",
		AutoStart:       false,
		MinimizeToTray:  true,
		RefreshInterval: 300,       // 5 minutes
		WebServerPort:   8080,      // Fixed port 8080 by default
		Routes:          []Route{}, // Empty routes by default
	}
}

// AddRoute adds a new route to the configuration
func (c *Config) AddRoute(hostname, service string) {
	// Check if route already exists
	for i := range c.Routes {
		if c.Routes[i].Hostname == hostname {
			// Update existing route
			c.Routes[i].Service = service
			return
		}
	}
	// Add new route
	c.Routes = append(c.Routes, Route{
		Hostname: hostname,
		Service:  service,
	})
}

// RemoveRoute removes a route by hostname
func (c *Config) RemoveRoute(hostname string) bool {
	for i := range c.Routes {
		if c.Routes[i].Hostname == hostname {
			c.Routes = append(c.Routes[:i], c.Routes[i+1:]...)
			return true
		}
	}
	return false
}

// GetRoute returns a route by hostname
func (c *Config) GetRoute(hostname string) *Route {
	for _, route := range c.Routes {
		if route.Hostname == hostname {
			return &route
		}
	}
	return nil
}

// getConfigPath returns the path to the config file
func getConfigPath() (string, error) {
	// Get user config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	// Create app config directory if it doesn't exist
	appConfigDir := filepath.Join(configDir, "cloudflared-desktop-tunnel")
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(appConfigDir, "config.json"), nil
}

// LoadConfig loads configuration from file
func LoadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	// If config doesn't exist, return default
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Save saves the configuration to file
func (c *Config) Save() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(configPath, data, 0644)
}
