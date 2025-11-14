# Cloudflared Binaries

This directory contains the embedded cloudflared binaries for different platforms.

## Download Binaries

Download the official cloudflared binaries from the [Cloudflare releases page](https://github.com/cloudflare/cloudflared/releases) and place them in the appropriate directories:

### Windows
```
windows/cloudflared-windows-amd64.exe
```

Download: `cloudflared-windows-amd64.exe` from the latest release

### macOS (Darwin)
```
darwin/cloudflared-darwin-amd64
darwin/cloudflared-darwin-arm64
```

Download:
- `cloudflared-darwin-amd64` for Intel Macs
- `cloudflared-darwin-arm64` for Apple Silicon (M1/M2/M3)

### Linux
```
linux/cloudflared-linux-amd64
linux/cloudflared-linux-arm64
```

Download:
- `cloudflared-linux-amd64` for x86_64 systems
- `cloudflared-linux-arm64` for ARM64 systems

## Platform Detection

The application uses Go's `runtime.GOOS` and `runtime.GOARCH` to detect the platform at runtime and extract the correct binary from the embedded files.

## Build Tags

Each platform has its own embed file with build tags to ensure only the correct binary is included in the final executable:

- `embed_windows.go` - `//go:build windows`
- `embed_darwin.go` - `//go:build darwin`
- `embed_linux.go` - `//go:build linux`

## Example Directory Structure

```
binaries/
├── README.md
├── embed.go (common interface)
├── windows/
│   ├── embed_windows.go
│   └── cloudflared-windows-amd64.exe
├── darwin/
│   ├── embed_darwin.go
│   ├── cloudflared-darwin-amd64
│   └── cloudflared-darwin-arm64
└── linux/
    ├── embed_linux.go
    ├── cloudflared-linux-amd64
    └── cloudflared-linux-arm64
```

## Important Notes

1. **File Permissions**: After extracting the binary on Unix systems (Linux/macOS), the application automatically sets executable permissions (`chmod +x`).

2. **Temporary Files**: The extracted binary is written to the system's temporary directory and is cleaned up when the application exits.

3. **Architecture Detection**: The application detects both OS and architecture to select the correct binary variant (e.g., ARM64 vs AMD64).

4. **Binary Size**: The embedded binaries will increase the application's file size. The cloudflared binary is approximately 40-50MB per platform.

## Security

- Always download binaries from official Cloudflare sources
- Verify checksums when available
- Keep binaries updated to the latest stable version
