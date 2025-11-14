# Cloudflared Desktop Tunnel

> Cross-platform desktop application for managing Cloudflare Tunnels with embedded `cloudflared` binary

## ⚠️ IMPORTANT: Binary Download Required

**This repository does NOT include cloudflared binaries.** You must download them separately before building.

### Quick Setup

```bash
# Clone repository
git clone https://github.com/votanchat/cloudflared-desktop-tunnel.git
cd cloudflared-desktop-tunnel

# Download binaries using provided script
./scripts/download-binaries.sh  # Linux/macOS
# OR
.\scripts\download-binaries.ps1  # Windows PowerShell

# Install dependencies and run
go mod download
cd frontend && npm install && cd ..
wails dev
```

📚 **For detailed setup instructions, see [SETUP.md](./SETUP.md)**

## 🚀 Features

- **Embedded Binaries**: Bundle cloudflared binaries for Windows, macOS (Intel/ARM), and Linux (amd64/arm64) using Go's `embed` directive
- **Cross-Platform**: Single codebase runs on Windows, macOS, and Linux
- **Flexible Token Management**: 
  - Auto-fetch tokens from backend API
  - **Manual token input** for testing/development (no backend required!)
- **Backend Integration**: Connect to backend API for token management and remote commands
- **Modern UI**: Built with Wails v2 + React + TypeScript + Vite
- **Auto-Update**: Receive and apply updates from backend
- **Real-time Logs**: Stream cloudflared output directly in the UI
- **System Tray**: Minimize to system tray for background operation

## 📝 Quick Links

- 🚀 [**Setup Guide (Start Here!)**](./SETUP.md) - Complete setup with troubleshooting
- 🔑 [**Manual Token Guide**](./docs/MANUAL_TOKEN.md) - Use without backend for testing
- 🏛️ [Architecture Documentation](./docs/ARCHITECTURE.md) - System design details
- 🔌 [Backend API Specification](./docs/BACKEND_API.md) - How to build backend
- ⚡ [Quick Start](./docs/QUICKSTART.md) - Get running in 5 minutes

## 💫 New: Manual Token Support

**No backend? No problem!** You can now start tunnels with manual tokens:

1. Get your tunnel token from Cloudflare:
   ```bash
   cloudflared tunnel token <tunnel-name>
   ```

2. In the app, click **"✏️ Manual Token"** button

3. Paste your token and click **Start Tunnel**

**Perfect for:**
- ✅ Testing without backend infrastructure
- ✅ Development and debugging
- ✅ Quick demos and POCs
- ✅ Temporary setups

See [Manual Token Guide](./docs/MANUAL_TOKEN.md) for details.

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

**Option A: Use automated script (Recommended)**

```bash
# Linux/macOS
chmod +x scripts/download-binaries.sh
./scripts/download-binaries.sh

# Windows PowerShell
.\scripts\download-binaries.ps1
```

**Option B: Manual download**

Download official binaries from [Cloudflare's releases](https://github.com/cloudflare/cloudflared/releases/latest) and place them in:

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

**Make binaries executable (Linux/macOS):**
```bash
chmod +x binaries/darwin/cloudflared-*
chmod +x binaries/linux/cloudflared-*
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

# macOS ARM (M1/M2/M3)
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

### Token Management

The app supports two ways to get tunnel tokens:

**1. Backend API (Production)**
```go
// Automatic token fetch from backend
GET /api/token -> {"token": "...", "expiresAt": "..."}
```

**2. Manual Token (Development/Testing)**
```go
// User provides token directly in UI
// Perfect when backend is not available
StartTunnel(manualToken string)
```

See [Manual Token Guide](./docs/MANUAL_TOKEN.md) for details.

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

See [Backend API Documentation](./docs/BACKEND_API.md) for full API specification.

## 🎨 Frontend Components

- **TunnelManager**: Main control panel for starting/stopping tunnel with manual token support
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

Config file location:
- **Windows**: `%APPDATA%\cloudflared-desktop-tunnel\config.json`
- **macOS**: `~/Library/Application Support/cloudflared-desktop-tunnel/config.json`
- **Linux**: `~/.config/cloudflared-desktop-tunnel/config.json`

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
├── docs/                  # Documentation
├── scripts/               # Helper scripts
└── build/                 # Output directory
```

### Generate TypeScript bindings
```bash
wails generate module
```

## ⚠️ Common Issues

### "Cannot read properties of undefined (reading 'app')"

This means Wails runtime is not initialized. Make sure you're running:
```bash
wails dev  # ✅ Correct
```

NOT:
```bash
npm run dev  # ❌ Wrong - this won't work!
```

See [SETUP.md](./SETUP.md) for detailed troubleshooting.

### Binary not found errors

Make sure you've downloaded the cloudflared binaries:
```bash
./scripts/download-binaries.sh
```

Verify they exist:
```bash
ls -lh binaries/*/*/*
```

### Testing without backend

Use manual token feature! See [Manual Token Guide](./docs/MANUAL_TOKEN.md).

## 🔒 Security Notes

- Tokens are never persisted to disk
- Manual tokens are cleared from UI after use
- Temporary binary files are cleaned up on shutdown
- Always use official Cloudflare binaries from https://github.com/cloudflare/cloudflared/releases
- Implement authentication in your backend API for production use

## 🤝 Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](./CONTRIBUTING.md) for details.

## 📄 License

MIT License - see [LICENSE](./LICENSE) file for details

## 🙏 Acknowledgments

- [Wails](https://wails.io) - Build desktop apps using Go & Web Technologies
- [Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/) - Secure tunnel service
- [React](https://react.dev) - Frontend library

## 📞 Support

For issues and questions:
- 🐛 [GitHub Issues](https://github.com/votanchat/cloudflared-desktop-tunnel/issues)
- 💬 [GitHub Discussions](https://github.com/votanchat/cloudflared-desktop-tunnel/discussions)
- 📚 [Setup Guide](./SETUP.md)
- 🔑 [Manual Token Guide](./docs/MANUAL_TOKEN.md)
- 🏛️ [Architecture Docs](./docs/ARCHITECTURE.md)
