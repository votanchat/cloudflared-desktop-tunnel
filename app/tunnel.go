package app

import (
	"bufio"
	"fmt"
	"io"
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

	binaryPath, err := tm.ensureBinary()
	if err != nil {
		return fmt.Errorf("failed to prepare binary: %w", err)
	}
	tm.binaryPath = binaryPath

	tunnelLogger.Info("Using cloudflared binary: %s", binaryPath)
	tunnelLogger.Debug("Runtime: GOOS=%s, GOARCH=%s", runtime.GOOS, runtime.GOARCH)
	tunnelLogger.Info("Using token mode (routes managed via Cloudflare Dashboard)")

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
	tunnelLogger.Info("Tunnel started with PID %d", tm.cmd.Process.Pid)

	go tm.readLogs(stdout, "stdout")
	go tm.readLogs(stderr, "stderr")
	go tm.monitorProcess()

	if tm.onTunnelStart != nil {
		go func() {
			if err := tm.onTunnelStart(); err != nil {
				tunnelLogger.Error("Error in onTunnelStart callback: %v", err)
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
	tunnelLogger.Info("Tunnel stopped")
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

// GetTunnelURL extracts tunnel URL from logs
func (tm *TunnelManager) GetTunnelURL() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	urlPatterns := []string{"trycloudflare.com", "cloudflare.com"}

	for _, logLine := range tm.logs {
		if !strings.Contains(logLine, "https://") {
			continue
		}

		for _, pattern := range urlPatterns {
			if !strings.Contains(logLine, pattern) {
				continue
			}

			start := strings.Index(logLine, "https://")
			if start == -1 {
				continue
			}

			end := strings.Index(logLine[start:], " ")
			if end == -1 {
				end = len(logLine[start:])
			}

			url := logLine[start : start+end]
			if strings.Contains(url, ".") {
				return url
			}
		}
	}
	return ""
}

// ensureBinary ensures the cloudflared binary is downloaded and ready to use
func (tm *TunnelManager) ensureBinary() (string, error) {
	if tm.binaryPath != "" {
		if _, err := os.Stat(tm.binaryPath); err == nil {
			if tm.isBinaryValid(tm.binaryPath) {
				tunnelLogger.Debug("Using cached binary: %s", tm.binaryPath)
				return tm.binaryPath, nil
			}
			tunnelLogger.Warn("Cached binary is invalid, will re-download")
		} else {
			tunnelLogger.Debug("Cached binary not found, will download")
		}
	}

	cacheDir, err := getCacheDir()
	if err != nil {
		return "", fmt.Errorf("failed to get cache dir: %w", err)
	}

	// Set logger for binaries package
	binaries.SetLogger(binaryLogger)

	binaryPath, err := binaries.DownloadCloudflared(cacheDir)
	if err != nil {
		return "", fmt.Errorf("failed to download binary: %w", err)
	}

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

	const minSize = 10 * 1024 * 1024 // 10MB
	if info.Size() < minSize {
		tunnelLogger.Warn("Binary file too small: %d bytes", info.Size())
		return false
	}

	if runtime.GOOS != "windows" && info.Mode()&0111 == 0 {
		tunnelLogger.Debug("Binary is not executable, fixing permissions...")
		if err := os.Chmod(path, 0755); err != nil {
			tunnelLogger.Error("Failed to set executable permission: %v", err)
			return false
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

	scanner := bufio.NewScanner(pipe)
	const maxLogLines = 100

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		tunnelLogger.Debug("[%s] %s", source, line)

		tm.mu.Lock()
		tm.logs = append(tm.logs, line)
		if len(tm.logs) > maxLogLines {
			tm.logs = tm.logs[len(tm.logs)-maxLogLines:]
		}
		tm.mu.Unlock()
	}

	if err := scanner.Err(); err != nil {
		tunnelLogger.Error("Error reading %s: %v", source, err)
	}
}

// monitorProcess monitors the tunnel process and handles exit
func (tm *TunnelManager) monitorProcess() {
	err := tm.cmd.Wait()

	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.running = false

	if err != nil {
		tunnelLogger.Error("Tunnel process exited with error: %v", err)
		tm.logs = append(tm.logs, fmt.Sprintf("Process exited: %v", err))
	} else {
		tunnelLogger.Info("Tunnel process exited normally")
	}
}

// Cleanup should be called when the app is shutting down to clean up cached binary
func (tm *TunnelManager) Cleanup() {
	if tm.binaryPath != "" {
		tunnelLogger.Debug("Cleaning up cached binary: %s", tm.binaryPath)
		os.Remove(tm.binaryPath)
	}
}
