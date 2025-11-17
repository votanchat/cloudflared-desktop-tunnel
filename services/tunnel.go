package services

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

	"github.com/votanchat/cloudflared-desktop-tunnel-v3/binaries"
	"github.com/votanchat/cloudflared-desktop-tunnel-v3/logger"
)

// TunnelService manages the cloudflared tunnel process
type TunnelService struct {
	mu            sync.RWMutex
	running       bool
	cmd           *exec.Cmd
	tunnelName    string
	logs          []string
	binaryPath    string
	configService *ConfigService
	onTunnelStart func() error
}

// NewTunnelService creates a new tunnel service
func NewTunnelService(tunnelName string, configService *ConfigService) *TunnelService {
	return &TunnelService{
		tunnelName:    tunnelName,
		logs:          make([]string, 0, 100),
		configService: configService,
	}
}

// SetOnTunnelStart sets the callback function to be called when tunnel starts successfully
func (s *TunnelService) SetOnTunnelStart(callback func() error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onTunnelStart = callback
}

// Start starts the cloudflared tunnel with the given token
func (s *TunnelService) Start(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("tunnel is already running")
	}

	binaryPath, err := s.ensureBinary()
	if err != nil {
		return fmt.Errorf("failed to prepare binary: %w", err)
	}
	s.binaryPath = binaryPath

	s.cmd = exec.Command(binaryPath, "tunnel", "run", "--token", token)

	stdout, err := s.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := s.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start tunnel: %w", err)
	}

	s.running = true
	logger.TunnelLogger.Info("✅ Tunnel started successfully")
	logger.TunnelLogger.Info("   🏷️  Name: %s", s.tunnelName)
	tokenPreview := token
	if len(token) > 10 {
		tokenPreview = token[:10] + "..."
	}
	logger.TunnelLogger.Info("   🔑 Token: %s", tokenPreview)

	go s.readLogs(stdout, "stdout")
	go s.readLogs(stderr, "stderr")
	go s.monitorProcess()

	if s.onTunnelStart != nil {
		go func() {
			if err := s.onTunnelStart(); err != nil {
				logger.TunnelLogger.Error("Failed to execute onTunnelStart callback: %v", err)
			}
		}()
	}

	return nil
}

// Stop stops the cloudflared tunnel
func (s *TunnelService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("tunnel is not running")
	}

	if s.cmd != nil && s.cmd.Process != nil {
		if err := s.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
		s.cmd.Wait()
	}

	s.running = false
	logger.TunnelLogger.Info("⏹️  Tunnel stopped")
	return nil
}

// IsRunning returns true if the tunnel is currently running
func (s *TunnelService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetLogs returns the last 100 log lines
func (s *TunnelService) GetLogs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.logs
}

// GetTunnelURL extracts tunnel URL from logs
func (s *TunnelService) GetTunnelURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	urlPatterns := []string{"trycloudflare.com", "cloudflare.com"}

	for _, logLine := range s.logs {
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

// GetStatus returns the current tunnel status
func (s *TunnelService) GetStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	config := s.configService.GetConfig()
	return map[string]interface{}{
		"running":    s.running,
		"tunnelName": config.TunnelName,
		"tunnelURL":  s.GetTunnelURL(),
		"logs":       s.logs,
	}
}

// ensureBinary ensures the cloudflared binary is downloaded and ready to use
func (s *TunnelService) ensureBinary() (string, error) {
	if s.binaryPath != "" {
		if _, err := os.Stat(s.binaryPath); err == nil {
			if s.isBinaryValid(s.binaryPath) {
				return s.binaryPath, nil
			}
		}
	}

	cacheDir, err := getCacheDir()
	if err != nil {
		return "", fmt.Errorf("failed to get cache dir: %w", err)
	}

	binaries.SetLogger(&binaryLoggerAdapter{})

	binaryPath, err := binaries.DownloadCloudflared(cacheDir)
	if err != nil {
		return "", fmt.Errorf("failed to download binary: %w", err)
	}

	if !s.isBinaryValid(binaryPath) {
		os.Remove(binaryPath)
		return "", fmt.Errorf("downloaded binary is not valid or not executable")
	}

	return binaryPath, nil
}

// isBinaryValid checks if the binary file is valid and executable
func (s *TunnelService) isBinaryValid(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	const minSize = 10 * 1024 * 1024 // 10MB
	if info.Size() < minSize {
		return false
	}

	if runtime.GOOS != "windows" && info.Mode()&0111 == 0 {
		if err := os.Chmod(path, 0755); err != nil {
			return false
		}
	}

	return true
}

// getCacheDir returns the cache directory for storing the binary
func getCacheDir() (string, error) {
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

	cacheDir := filepath.Join(baseDir, "cloudflared-desktop-tunnel-v3")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	return cacheDir, nil
}

// readLogs reads logs from the given pipe and stores them
func (s *TunnelService) readLogs(pipe io.ReadCloser, source string) {
	defer pipe.Close()

	scanner := bufio.NewScanner(pipe)
	const maxLogLines = 100

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		s.mu.Lock()
		s.logs = append(s.logs, line)
		if len(s.logs) > maxLogLines {
			s.logs = s.logs[len(s.logs)-maxLogLines:]
		}
		s.mu.Unlock()
	}

	if err := scanner.Err(); err != nil {
		logger.TunnelLogger.Error("Error reading tunnel logs from %s: %v", source, err)
	}
}

// monitorProcess monitors the tunnel process and handles exit
func (s *TunnelService) monitorProcess() {
	err := s.cmd.Wait()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.running = false

	if err != nil {
		s.logs = append(s.logs, fmt.Sprintf("Process exited: %v", err))
	}
}

// Cleanup should be called when the app is shutting down
func (s *TunnelService) Cleanup() {
	if s.binaryPath != "" {
		os.Remove(s.binaryPath)
	}
}

// binaryLoggerAdapter adapts our logger to the binaries package logger interface
type binaryLoggerAdapter struct{}

func (l *binaryLoggerAdapter) Info(format string, args ...interface{}) {
	// Logger will be set from main package
	fmt.Printf("[BINARY] "+format+"\n", args...)
}

func (l *binaryLoggerAdapter) Warn(format string, args ...interface{}) {
	fmt.Printf("[BINARY] WARN: "+format+"\n", args...)
}

func (l *binaryLoggerAdapter) Error(format string, args ...interface{}) {
	fmt.Printf("[BINARY] ERROR: "+format+"\n", args...)
}

func (l *binaryLoggerAdapter) Debug(format string, args ...interface{}) {
	fmt.Printf("[BINARY] DEBUG: "+format+"\n", args...)
}

