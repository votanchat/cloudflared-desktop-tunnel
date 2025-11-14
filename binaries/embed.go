package binaries

// CloudflaredBinary holds the embedded cloudflared binary for the current platform
// This variable is populated by platform-specific build tag files:
// - embed_windows.go (for Windows)
// - embed_darwin.go (for macOS)
// - embed_linux.go (for Linux)
var CloudflaredBinary []byte
