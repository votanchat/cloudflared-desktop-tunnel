//go:build darwin

package binaries

import (
	_ "embed"
	"runtime"
)

// Embedded binaries for macOS (both Intel and ARM)
var (
	//go:embed cloudflared-darwin-amd64
	amd64Binary []byte

	//go:embed cloudflared-darwin-arm64
	arm64Binary []byte
)

// init selects the correct binary based on architecture
func init() {
	switch runtime.GOARCH {
	case "amd64":
		Cloudflar edBinary = amd64Binary
	case "arm64":
		CloudflaredBinary = arm64Binary
	default:
		// Fallback to amd64 if architecture is unknown
		CloudflaredBinary = amd64Binary
	}
}
