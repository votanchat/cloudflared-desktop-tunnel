//go:build windows

package binaries

import (
	_ "embed"
)

// CloudflaredBinary is the embedded cloudflared binary for Windows AMD64
//
//go:embed cloudflared-windows-amd64.exe
var CloudflaredBinary []byte
