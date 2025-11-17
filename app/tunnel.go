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

// OnTunnelStart is a callback function when tunnel starts successfully
type OnTunnelStart func() error

// TunnelManager manages the cloudflared tunnel process
type TunnelManager struct {
	mu            sync.RWMutex
	running       bool
	cmd           *exec.Cmd
	tunnelName    string
	logs          []string
	binaryPath    string        // Cached binary path
	config        *Config       // Reference to config for routes
	onTunnelStart OnTunnelStart // Callback when tunnel starts
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

// SetOnTunnelStart sets the callback function to be called when tunnel starts successfully
func (tm *TunnelManager) SetOnTunnelStart(callback OnTunnelStart) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.onTunnelStart = callback
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

	// Always use token mode (routes managed via Cloudflare Dashboard)
	log.Println("Using token mode (routes managed via Cloudflare Dashboard)")
	tm.cmd = exec.Command(binaryPath, "tunnel", "run", "--token", token)

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

	// Call onTunnelStart callback if set
	if tm.onTunnelStart != nil {
		go func() {
			if err := tm.onTunnelStart(); err != nil {
				log.Printf("Error in onTunnelStart callback: %v", err)
			}
		}()
	}

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

// GetTunnelURL extracts tunnel URL from logs (looks for cURL URL pattern)
func (tm *TunnelManager) GetTunnelURL() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	// Look for URL in logs (cloudflared prints URLs like "https://xxx.trycloudflare.com")
	for _, log := range tm.logs {
		// Match patterns like https://xxx.trycloudflare.com or https://xxx.cloudflare.com
		if strings.Contains(log, "https://") && (strings.Contains(log, "trycloudflare.com") || strings.Contains(log, "cloudflare.com")) {
			// Extract URL from log line
			start := strings.Index(log, "https://")
			if start != -1 {
				// Find end of URL (space or end of line)
				end := strings.Index(log[start:], " ")
				if end == -1 {
					end = len(log[start:])
				}
				url := log[start : start+end]
				// Validate URL format
				if strings.Contains(url, ".") {
					return url
				}
			}
		}
	}
	return ""
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
