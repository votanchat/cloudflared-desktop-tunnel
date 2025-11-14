# Quick Start Guide

Get up and running with Cloudflared Desktop Tunnel in 5 minutes!

## Prerequisites

Before you begin, make sure you have:

- ✅ Go 1.21 or higher
- ✅ Node.js 18+ and npm
- ✅ Git
- ✅ Platform-specific dependencies:
  - **Windows**: WebView2 Runtime (pre-installed on Windows 10/11)
  - **macOS**: Xcode Command Line Tools
  - **Linux**: GTK3 and WebKit2GTK

## Step 1: Install Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

Verify installation:
```bash
wails version
```

## Step 2: Clone the Repository

```bash
git clone https://github.com/votanchat/cloudflared-desktop-tunnel.git
cd cloudflared-desktop-tunnel
```

## Step 3: Download Cloudflared Binaries

Download the official binaries from [Cloudflare's releases](https://github.com/cloudflare/cloudflared/releases/latest):

### For Windows
```bash
# Download cloudflared-windows-amd64.exe
# Place in: binaries/windows/cloudflared-windows-amd64.exe
```

### For macOS
```bash
# Download cloudflared-darwin-amd64 (Intel)
# Download cloudflared-darwin-arm64 (Apple Silicon)
# Place in: binaries/darwin/
```

### For Linux
```bash
# Download cloudflared-linux-amd64
# Download cloudflared-linux-arm64 (optional)
# Place in: binaries/linux/
```

**Quick Download Script** (Linux/macOS):
```bash
# Run from project root
./scripts/download-binaries.sh
```

## Step 4: Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install frontend dependencies
cd frontend
npm install
cd ..
```

## Step 5: Run in Development Mode

```bash
wails dev
```

The application will open automatically with:
- Hot-reload for frontend changes
- Automatic restart for backend changes
- DevTools enabled

## Step 6: Configure Backend (Optional)

If you have a backend API:

1. Click the **Settings** tab
2. Enter your backend URL (e.g., `https://api.example.com`)
3. Enter your tunnel name
4. Save settings

## Step 7: Start Your Tunnel

1. Click the **Tunnel** tab
2. Click **▶️ Start Tunnel**
3. The app will:
   - Fetch a token from your backend (or use a test token)
   - Start the cloudflared process
   - Display connection logs

## Building for Production

### Build for Current Platform

```bash
wails build
```

Output will be in `build/bin/`

### Build for All Platforms

```bash
# Windows
wails build -platform windows/amd64

# macOS
wails build -platform darwin/amd64  # Intel
wails build -platform darwin/arm64  # Apple Silicon

# Linux
wails build -platform linux/amd64
```

## Testing Without Backend

To test without a backend API:

1. Get a tunnel token manually from Cloudflare:
   ```bash
   cloudflared tunnel token <tunnel-id>
   ```

2. Modify `app/app.go` temporarily to use hardcoded token:
   ```go
   func (a *App) StartTunnel() error {
       // Use hardcoded token for testing
       token := "your-tunnel-token-here"
       return a.tunnel.Start(token)
   }
   ```

## Troubleshooting

### Wails CLI not found
```bash
# Add Go bin to PATH
export PATH=$PATH:$(go env GOPATH)/bin
```

### GTK errors on Linux
```bash
sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.0-dev
```

### Frontend build fails
```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
cd ..
```

### Binary not executable (Unix)
```bash
chmod +x build/bin/cloudflared-tunnel
```

### Port already in use
```bash
# Frontend dev server (default 5173)
# Change in frontend/vite.config.ts
```

## Next Steps

- 📚 Read the [Architecture Documentation](./ARCHITECTURE.md)
- 🔌 Implement your [Backend API](./BACKEND_API.md)
- 🐛 Report issues on [GitHub](https://github.com/votanchat/cloudflared-desktop-tunnel/issues)
- ⭐ Star the repository if you find it useful!

## Getting Help

- 💬 [GitHub Discussions](https://github.com/votanchat/cloudflared-desktop-tunnel/discussions)
- 🐛 [Issue Tracker](https://github.com/votanchat/cloudflared-desktop-tunnel/issues)
- 📝 [Wails Documentation](https://wails.io/docs/introduction)
- 🔒 [Cloudflare Tunnel Docs](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/)

## Demo Video

(Coming soon!)

## Credits

Built with:
- [Wails](https://wails.io) - Go + Web Technologies framework
- [React](https://react.dev) - UI library
- [Cloudflare Tunnel](https://www.cloudflare.com/products/tunnel/) - Secure tunneling service

Happy tunneling! 🚀
