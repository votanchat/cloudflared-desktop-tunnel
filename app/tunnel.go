package app

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/votanchat/cloudflared-desktop-tunnel/binaries"
)

// TunnelManager manages the cloudflared tunnel process
type TunnelManager struct {
	mu         sync.RWMutex
	running    bool
	cmd        *exec.Cmd
	tunnelName string
	logs       []string
	binaryPath string  // Cached binary path
	config     *Config // Reference to config for routes
}

// NewTunnelManager creates a new tunnel manager
func NewTunnelManager(tunnelName string) *TunnelManager {
	return &TunnelManager{
		tunnelName: tunnelName,
		logs:       make([]string, 0, 100),
	}
}

// SetConfig sets the config reference for route management
func (tm *TunnelManager) SetConfig(config *Config) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.config = config
}

// Start starts the cloudflared tunnel with the given token
func (tm *TunnelManager) Start(token string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.running {
		return fmt.Errorf("tunnel is already running")
	}

	// Extract or get cached binary
	binaryPath, err := tm.ensureBinary()
	if err != nil {
		return fmt.Errorf("failed to prepare binary: %w", err)
	}
	tm.binaryPath = binaryPath

	// Log binary info for debugging
	log.Printf("Using cloudflared binary: %s", binaryPath)
	log.Printf("Runtime: GOOS=%s, GOARCH=%s", runtime.GOOS, runtime.GOARCH)

	// Check if we have routes configured
	if tm.config != nil && len(tm.config.Routes) > 0 {
		// Use config file with routes
		configPath, err := tm.generateConfigFile(token)
		if err != nil {
			return fmt.Errorf("failed to generate config file: %w", err)
		}
		log.Printf("Using config file with %d route(s): %s", len(tm.config.Routes), configPath)
		for _, route := range tm.config.Routes {
			log.Printf("  Route: %s -> %s", route.Hostname, route.Service)
		}
		// Use both --token and --config (token for auth, config for routes)
		tm.cmd = exec.Command(binaryPath, "tunnel", "run", "--token", token, "--config", configPath)
	} else {
		// Use token directly (no routes, routes managed via Cloudflare Dashboard)
		log.Println("Using token mode (routes managed via Cloudflare Dashboard)")
		tm.cmd = exec.Command(binaryPath, "tunnel", "run", "--token", token)
	}

	// Capture stdout and stderr
	stdout, err := tm.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := tm.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the process
	if err := tm.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start tunnel: %w", err)
	}

	tm.running = true
	log.Printf("Tunnel started with PID %d", tm.cmd.Process.Pid)

	// Start goroutines to read logs
	go tm.readLogs(stdout, "stdout")
	go tm.readLogs(stderr, "stderr")

	// Monitor process
	go tm.monitorProcess()

	return nil
}

// Stop stops the cloudflared tunnel
func (tm *TunnelManager) Stop() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if !tm.running {
		return fmt.Errorf("tunnel is not running")
	}

	if tm.cmd != nil && tm.cmd.Process != nil {
		if err := tm.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
		tm.cmd.Wait() // Wait for process to exit
	}

	tm.running = false
	log.Println("Tunnel stopped")

	// Note: We DON'T delete binaryPath here anymore
	// It's cached for next start

	return nil
}

// IsRunning returns true if the tunnel is currently running
func (tm *TunnelManager) IsRunning() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.running
}

// GetLogs returns the last 100 log lines
func (tm *TunnelManager) GetLogs() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.logs
}

// ensureBinary ensures the cloudflared binary is downloaded and ready to use
// It caches the binary and only downloads once
func (tm *TunnelManager) ensureBinary() (string, error) {
	// If we already have a cached binary path, verify it exists
	if tm.binaryPath != "" {
		if _, err := os.Stat(tm.binaryPath); err == nil {
			if tm.isBinaryValid(tm.binaryPath) {
				log.Printf("Using cached binary: %s", tm.binaryPath)
				return tm.binaryPath, nil
			}
			log.Printf("Cached binary is invalid, will re-download")
		} else {
			log.Printf("Cached binary not found, will download")
		}
	}

	// Create app-specific cache directory
	cacheDir, err := getCacheDir()
	if err != nil {
		return "", fmt.Errorf("failed to get cache dir: %w", err)
	}

	// Download binary from GitHub releases
	binaryPath, err := binaries.DownloadCloudflared(cacheDir)
	if err != nil {
		return "", fmt.Errorf("failed to download binary: %w", err)
	}

	// Verify the downloaded binary
	if !tm.isBinaryValid(binaryPath) {
		os.Remove(binaryPath)
		return "", fmt.Errorf("downloaded binary is not valid or not executable")
	}

	return binaryPath, nil
}

// isBinaryValid checks if the binary file is valid and executable
func (tm *TunnelManager) isBinaryValid(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Check file size (cloudflared should be at least 10MB)
	if info.Size() < 10*1024*1024 {
		log.Printf("Binary file too small: %d bytes", info.Size())
		return false
	}

	// Check if executable (Unix systems)
	if runtime.GOOS != "windows" {
		if info.Mode()&0111 == 0 {
			log.Printf("Binary is not executable, fixing permissions...")
			if err := os.Chmod(path, 0755); err != nil {
				log.Printf("Failed to set executable permission: %v", err)
				return false
			}
		}
	}

	return true
}

// getCacheDir returns the cache directory for storing the binary
func getCacheDir() (string, error) {
	// Use OS-specific cache directory
	var baseDir string
	var err error

	switch runtime.GOOS {
	case "darwin":
		homeDir, _ := os.UserHomeDir()
		baseDir = filepath.Join(homeDir, "Library", "Caches")
	case "windows":
		baseDir, err = os.UserCacheDir()
		if err != nil {
			return "", err
		}
	default: // linux
		homeDir, _ := os.UserHomeDir()
		xdgCache := os.Getenv("XDG_CACHE_HOME")
		if xdgCache != "" {
			baseDir = xdgCache
		} else {
			baseDir = filepath.Join(homeDir, ".cache")
		}
	}

	// Create app-specific subdirectory
	cacheDir := filepath.Join(baseDir, "cloudflared-desktop-tunnel")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	return cacheDir, nil
}

// generateConfigFile generates a cloudflared YAML config file with routes
func (tm *TunnelManager) generateConfigFile(token string) (string, error) {
	// Get config directory
	configDir, err := getConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config dir: %w", err)
	}

	configPath := filepath.Join(configDir, "tunnel-config.yaml")

	// Build YAML content with pre-allocated capacity
	yaml := strings.Builder{}
	yaml.Grow(256 + len(tm.config.Routes)*64) // Pre-allocate for efficiency
	yaml.WriteString("# Cloudflared tunnel configuration\n")
	yaml.WriteString("# Generated automatically by cloudflared-desktop-tunnel\n\n")

	// Note: cloudflared doesn't support token in config file
	// We'll use --token flag separately, config file only for ingress routes
	yaml.WriteString("# Tunnel will be authenticated via --token flag\n\n")

	// Add ingress routes
	yaml.WriteString("ingress:\n")
	for _, route := range tm.config.Routes {
		yaml.WriteString("  - hostname: ")
		yaml.WriteString(route.Hostname)
		yaml.WriteString("\n    service: ")
		yaml.WriteString(route.Service)
		yaml.WriteString("\n")
	}

	// Add catch-all rule (required by cloudflared)
	yaml.WriteString("  - service: http_status:404\n")

	// Write to file
	yamlBytes := []byte(yaml.String())
	if err := os.WriteFile(configPath, yamlBytes, 0644); err != nil {
		return "", fmt.Errorf("failed to write config file: %w", err)
	}

	log.Printf("Generated tunnel config file: %s", configPath)
	return configPath, nil
}

// getConfigDir returns the config directory for storing tunnel config files
func getConfigDir() (string, error) {
	// Use OS-specific config directory
	var baseDir string
	var err error

	switch runtime.GOOS {
	case "darwin":
		homeDir, _ := os.UserHomeDir()
		baseDir = filepath.Join(homeDir, "Library", "Application Support")
	case "windows":
		baseDir, err = os.UserConfigDir()
		if err != nil {
			return "", err
		}
	default: // linux
		homeDir, _ := os.UserHomeDir()
		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig != "" {
			baseDir = xdgConfig
		} else {
			baseDir = filepath.Join(homeDir, ".config")
		}
	}

	// Create app-specific subdirectory
	configDir := filepath.Join(baseDir, "cloudflared-desktop-tunnel")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return configDir, nil
}

// readLogs reads logs from the given pipe and stores them
func (tm *TunnelManager) readLogs(pipe io.ReadCloser, source string) {
	defer pipe.Close()
	buf := make([]byte, 4096) // Larger buffer for better I/O efficiency
	for {
		n, err := pipe.Read(buf)
		if n > 0 {
			line := string(buf[:n])
			log.Printf("[%s] %s", source, line)

			tm.mu.Lock()
			tm.logs = append(tm.logs, line)
			// Keep only last 100 lines - use slice reslicing instead of re-allocation
			if len(tm.logs) > 100 {
				copy(tm.logs, tm.logs[len(tm.logs)-100:])
				tm.logs = tm.logs[:100]
			}
			tm.mu.Unlock()
		}
		if err != nil {
			break
		}
	}
}

// monitorProcess monitors the tunnel process and handles exit
func (tm *TunnelManager) monitorProcess() {
	err := tm.cmd.Wait()

	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.running = false

	if err != nil {
		log.Printf("Tunnel process exited with error: %v", err)
		tm.logs = append(tm.logs, fmt.Sprintf("Process exited: %v", err))
	} else {
		log.Println("Tunnel process exited normally")
	}

	// Note: We keep the binary cached for next start
}

// Cleanup should be called when the app is shutting down to clean up cached binary
func (tm *TunnelManager) Cleanup() {
	if tm.binaryPath != "" {
		log.Printf("Cleaning up cached binary: %s", tm.binaryPath)
		os.Remove(tm.binaryPath)
	}
}
