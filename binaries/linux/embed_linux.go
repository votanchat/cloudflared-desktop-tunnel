//go:build linux

package binaries

import (
	_ "embed"
	"runtime"
)

// Embedded binaries for Linux (both AMD64 and ARM64)
var (
	//go:embed cloudflared-linux-amd64
	amd64Binary []byte

	//go:embed cloudflared-linux-arm64
	arm64Binary []byte
)

// init selects the correct binary based on architecture
func init() {
	switch runtime.GOARCH {
	case "amd64":
		CloudflaredBinary = amd64Binary
	case "arm64":
		CloudflaredBinary = arm64Binary
	default:
		// Fallback to amd64 if architecture is unknown
		CloudflaredBinary = amd64Binary
	}
}
