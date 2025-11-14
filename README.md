# Cloudflared Desktop Tunnel

> Cross-platform desktop application for managing Cloudflare Tunnels with embedded `cloudflared` binary

## 🚀 Features

- **Embedded Binaries**: Bundle cloudflared binaries for Windows, macOS (Intel/ARM), and Linux (amd64/arm64) using Go's `embed` directive
- **Cross-Platform**: Single codebase runs on Windows, macOS, and Linux
- **Backend Integration**: Connect to backend API for token management and remote commands
- **Modern UI**: Built with Wails v2 + React + TypeScript + Vite
- **Auto-Update**: Receive and apply updates from backend
- **System Tray**: Minimize to system tray for background operation

## 📋 Prerequisites

- Go 1.21 or higher
- Node.js 18+ and npm/yarn
- Wails CLI v2
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### Platform-specific requirements:

**Windows:**
- WebView2 Runtime (usually pre-installed on Windows 10/11)
- GCC (MinGW-w64 for cross-compilation)

**macOS:**
- Xcode Command Line Tools

**Linux:**
- GTK3 and WebKit2GTK
```bash
sudo apt install libgtk-3-dev libwebkit2gtk-4.0-dev # Debian/Ubuntu
sudo dnf install gtk3-devel webkit2gtk3-devel # Fedora
```

## 🛠️ Installation

### 1. Clone the repository
```bash
git clone https://github.com/votanchat/cloudflared-desktop-tunnel.git
cd cloudflared-desktop-tunnel
```

### 2. Download cloudflared binaries

Download the official cloudflared binaries from [Cloudflare's releases](https://github.com/cloudflare/cloudflared/releases) and place them in the `binaries/` directory:

```
binaries/
├── windows/
│   └── cloudflared-windows-amd64.exe
├── darwin/
│   ├── cloudflared-darwin-amd64
│   └── cloudflared-darwin-arm64
└── linux/
    ├── cloudflared-linux-amd64
    └── cloudflared-linux-arm64
```

### 3. Install dependencies
```bash
# Install Go dependencies
go mod download

# Install frontend dependencies
cd frontend
npm install
cd ..
```

### 4. Run in development mode
```bash
wails dev
```

## 🏗️ Building

### Build for current platform
```bash
wails build
```

### Cross-platform builds
```bash
# Windows
wails build -platform windows/amd64

# macOS Intel
wails build -platform darwin/amd64

# macOS ARM (M1/M2)
wails build -platform darwin/arm64

# Linux
wails build -platform linux/amd64
```

Built binaries will be in `build/bin/`

## 📖 Architecture

### Binary Embedding Strategy

The app uses Go's `embed` directive with build tags to conditionally embed the correct binary:

```go
// binaries/windows/embed_windows.go
//go:build windows

package binaries

import _ "embed"

//go:embed cloudflared-windows-amd64.exe
var CloudflaredBinary []byte
```

At runtime, the app:
1. Detects OS and architecture using `runtime.GOOS` and `runtime.GOARCH`
2. Extracts the embedded binary to a temporary directory
3. Sets executable permissions (Unix systems)
4. Runs the binary with appropriate flags
5. Cleans up on application shutdown

### Backend Communication

The app connects to a backend API for:
- **Token Management**: Fetch and refresh Cloudflare tunnel tokens
- **Remote Commands**: Receive commands to update, restart, or configure the tunnel
- **Health Monitoring**: Report tunnel status and logs

```go
type BackendClient struct {
    baseURL string
    token   string
    ws      *websocket.Conn
}

// Backend API endpoints
// GET  /api/token       - Fetch tunnel token
// POST /api/status      - Report tunnel status
// WS   /api/commands    - Receive real-time commands
```

## 🎨 Frontend Components

- **TunnelManager**: Main control panel for starting/stopping tunnel
- **StatusDisplay**: Real-time tunnel status and connection info
- **Settings**: Configure backend URL and tunnel parameters
- **LogsViewer**: Display cloudflared output logs

## 🔧 Configuration

Create a `config.json` file or use the UI settings:

```json
{
  "backendURL": "https://your-backend.com",
  "tunnelName": "my-tunnel",
  "autoStart": false,
  "minimizeToTray": true
}
```

## 📝 Development

### Project Structure
```
.
├── main.go                 # Wails application entry point
├── app/
│   ├── app.go             # Main app lifecycle
│   ├── tunnel.go          # Tunnel management logic
│   ├── config.go          # Configuration handling
│   └── backend_client.go  # Backend API client
├── binaries/
│   ├── windows/
│   │   ├── embed_windows.go
│   │   └── cloudflared-windows-amd64.exe
│   ├── darwin/
│   │   ├── embed_darwin.go
│   │   ├── cloudflared-darwin-amd64
│   │   └── cloudflared-darwin-arm64
│   └── linux/
│       ├── embed_linux.go
│       ├── cloudflared-linux-amd64
│       └── cloudflared-linux-arm64
├── frontend/
│   ├── src/
│   │   ├── App.tsx
│   │   ├── components/
│   │   └── hooks/
│   └── package.json
└── build/                 # Output directory
```

### Generate TypeScript bindings
```bash
wails generate module
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

MIT License - see LICENSE file for details

## 🙏 Acknowledgments

- [Wails](https://wails.io) - Build desktop apps using Go & Web Technologies
- [Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/) - Secure tunnel service
- [React](https://react.dev) - Frontend library

## 📞 Support

For issues and questions, please use the [GitHub Issues](https://github.com/votanchat/cloudflared-desktop-tunnel/issues) page.
