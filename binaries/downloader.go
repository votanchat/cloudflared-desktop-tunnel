package binaries

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Logger interface for binary downloader
type Logger interface {
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Debug(format string, args ...interface{})
}

var binaryLogger Logger = &defaultLogger{}

type defaultLogger struct{}

func (l *defaultLogger) Info(format string, args ...interface{}) {
	log.Printf("[BINARY] "+format, args...)
}

func (l *defaultLogger) Warn(format string, args ...interface{}) {
	log.Printf("[BINARY] WARN: "+format, args...)
}

func (l *defaultLogger) Error(format string, args ...interface{}) {
	log.Printf("[BINARY] ERROR: "+format, args...)
}

func (l *defaultLogger) Debug(format string, args ...interface{}) {
	log.Printf("[BINARY] DEBUG: "+format, args...)
}

// SetLogger sets a custom logger for the binaries package
func SetLogger(logger Logger) {
	binaryLogger = logger
}

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

// DownloadCloudflared downloads the cloudflared binary for the current platform
// and saves it to the cache directory
func DownloadCloudflared(cacheDir string) (string, error) {
	// Ensure cache directory exists
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Determine binary name based on OS and architecture
	binaryName := fmt.Sprintf("cloudflared-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	binaryPath := filepath.Join(cacheDir, binaryName)

	const minSize = 10 * 1024 * 1024 // 10MB

	if info, err := os.Stat(binaryPath); err == nil {
		if info.Size() >= minSize {
			binaryLogger.Info("Valid binary found in cache: %s", binaryPath)
			return binaryPath, nil
		}
		binaryLogger.Warn("Cached binary is too small, will re-download")
		os.Remove(binaryPath)
	}

	binaryLogger.Info("Fetching latest cloudflared version from GitHub...")
	version, err := getLatestVersion()
	if err != nil {
		return "", fmt.Errorf("failed to get latest version: %w", err)
	}
	binaryLogger.Info("Latest version: %s", version)

	binaryLogger.Info("Downloading cloudflared binary for %s/%s...", runtime.GOOS, runtime.GOARCH)
	if err := downloadBinary(version, binaryPath); err != nil {
		return "", fmt.Errorf("failed to download binary: %w", err)
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(binaryPath, 0755); err != nil {
			return "", fmt.Errorf("failed to set executable permissions: %w", err)
		}
	}

	binaryLogger.Info("Binary downloaded successfully: %s", binaryPath)
	return binaryPath, nil
}

var githubClient = &http.Client{
	Timeout: 30 * time.Second,
}

// getLatestVersion fetches the latest cloudflared version from GitHub
func getLatestVersion() (string, error) {
	resp, err := githubClient.Get("https://api.github.com/repos/cloudflare/cloudflared/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return release.TagName, nil
}

// downloadBinary downloads the cloudflared binary for the current platform
func downloadBinary(version, outputPath string) error {
	baseURL := fmt.Sprintf("https://github.com/cloudflare/cloudflared/releases/download/%s", version)

	var downloadURL string
	var needsExtraction bool

	switch runtime.GOOS {
	case "windows":
		downloadURL = fmt.Sprintf("%s/cloudflared-windows-amd64.exe", baseURL)
		needsExtraction = false
	case "darwin":
		// macOS uses .tgz files
		if runtime.GOARCH == "arm64" {
			downloadURL = fmt.Sprintf("%s/cloudflared-darwin-arm64.tgz", baseURL)
		} else {
			downloadURL = fmt.Sprintf("%s/cloudflared-darwin-amd64.tgz", baseURL)
		}
		needsExtraction = true
	case "linux":
		if runtime.GOARCH == "arm64" {
			downloadURL = fmt.Sprintf("%s/cloudflared-linux-arm64", baseURL)
		} else {
			downloadURL = fmt.Sprintf("%s/cloudflared-linux-amd64", baseURL)
		}
		needsExtraction = false
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	// Download the file
	downloadClient := &http.Client{
		Timeout: 5 * time.Minute, // Binary downloads can take time
	}

	binaryLogger.Debug("Downloading from: %s", downloadURL)
	resp, err := downloadClient.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	if needsExtraction {
		// Extract from .tgz (macOS)
		return extractTgz(resp.Body, outputPath)
	}

	// Direct download (Windows, Linux)
	return writeFile(resp.Body, outputPath)
}

// extractTgz extracts the cloudflared binary from a .tgz archive
func extractTgz(r io.Reader, outputPath string) error {
	// Create temporary file for the archive
	tmpFile, err := os.CreateTemp("", "cloudflared-*.tgz")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write archive to temp file
	if _, err := io.Copy(tmpFile, r); err != nil {
		return fmt.Errorf("failed to write archive: %w", err)
	}

	// Close and reopen for reading
	tmpFile.Close()

	// Open the archive
	file, err := os.Open(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	// Create tar reader
	tr := tar.NewReader(gzr)

	// Extract the cloudflared binary
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		// Look for the cloudflared binary (usually just "cloudflared" in the archive)
		if header.Typeflag == tar.TypeReg && (header.Name == "cloudflared" || filepath.Base(header.Name) == "cloudflared") {
			// Write to output path
			outFile, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, tr); err != nil {
				return fmt.Errorf("failed to extract binary: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("cloudflared binary not found in archive")
}

// writeFile writes the content from reader to the output path
func writeFile(r io.Reader, outputPath string) error {
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, r); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

