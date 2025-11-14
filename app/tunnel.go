package app

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	binaryPath string
}

// NewTunnelManager creates a new tunnel manager
func NewTunnelManager(tunnelName string) *TunnelManager {
	return &TunnelManager{
		tunnelName: tunnelName,
		logs:       make([]string, 0, 100),
	}
}

// Start starts the cloudflared tunnel with the given token
func (tm *TunnelManager) Start(token string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.running {
		return fmt.Errorf("tunnel is already running")
	}

	// Extract the embedded binary to a temporary location
	binaryPath, err := tm.extractBinary()
	if err != nil {
		return fmt.Errorf("failed to extract binary: %w", err)
	}
	tm.binaryPath = binaryPath

	// Prepare the cloudflared command
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

	// Clean up the temporary binary
	if tm.binaryPath != "" {
		os.Remove(tm.binaryPath)
	}

	tm.running = false
	log.Println("Tunnel stopped")

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

// extractBinary extracts the embedded binary to a temporary file
func (tm *TunnelManager) extractBinary() (string, error) {
	// Create a temporary file
	tmpDir := os.TempDir()
	binaryName := fmt.Sprintf("cloudflared-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	tmpFile := filepath.Join(tmpDir, binaryName)

	// Write the embedded binary to the temp file
	if err := os.WriteFile(tmpFile, binaries.CloudflaredBinary, 0755); err != nil {
		return "", fmt.Errorf("failed to write binary: %w", err)
	}

	log.Printf("Extracted binary to: %s", tmpFile)
	return tmpFile, nil
}

// readLogs reads logs from the given pipe and stores them
func (tm *TunnelManager) readLogs(pipe io.ReadCloser, source string) {
	buf := make([]byte, 1024)
	for {
		n, err := pipe.Read(buf)
		if n > 0 {
			line := string(buf[:n])
			log.Printf("[%s] %s", source, line)

			tm.mu.Lock()
			tm.logs = append(tm.logs, line)
			// Keep only last 100 lines
			if len(tm.logs) > 100 {
				tm.logs = tm.logs[len(tm.logs)-100:]
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

	// Clean up the temporary binary
	if tm.binaryPath != "" {
		os.Remove(tm.binaryPath)
	}
}
