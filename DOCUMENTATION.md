# Cloudflared Desktop Tunnel - Complete Documentation

**Table of Contents**
- [Overview](#overview)
- [Quick Start](#quick-start)
- [Setup Guide](#setup-guide)
- [Architecture](#architecture)
- [Backend API](#backend-api)
- [Manual Token Usage](#manual-token-usage)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [Code Optimizations](#code-optimizations)
- [Helper Scripts](#helper-scripts)
- [Binary Management](#binary-management)

---

## Overview

Cloudflared Desktop Tunnel is a cross-platform desktop application for managing Cloudflare Tunnels with embedded `cloudflared` binary support.

### Features

- **Cross-Platform**: Single codebase runs on Windows, macOS, and Linux
- **Flexible Token Management**: Auto-fetch tokens from backend API or use manual tokens
- **Backend Integration**: Connect to backend API for token management and remote commands
- **Modern UI**: Built with Wails v2 + React + TypeScript + Vite
- **Real-time Logs**: Stream cloudflared output directly in the UI
- **System Tray**: Minimize to system tray for background operation
- **Auto-Update**: Receive and apply updates from backend
- **Runtime Binary Download**: Automatically downloads cloudflared binaries from GitHub

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
│   ├── darwin/
│   └── linux/
├── frontend/              # React TypeScript UI
├── docs/                  # Documentation
├── scripts/               # Helper scripts
└── build/                 # Output directory
```

---

## Quick Start

Get up and running with Cloudflared Desktop Tunnel in 5 minutes!

### Prerequisites

- Go 1.21 or higher
- Node.js 18+ and npm
- Git
- Platform-specific dependencies:
  - **Windows**: WebView2 Runtime (pre-installed on Windows 10/11)
  - **macOS**: Xcode Command Line Tools
  - **Linux**: GTK3 and WebKit2GTK

### Step 1: Install Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

Verify installation:
```bash
wails version
```

### Step 2: Clone the Repository

```bash
git clone https://github.com/votanchat/cloudflared-desktop-tunnel.git
cd cloudflared-desktop-tunnel
```

### Step 3: Download Cloudflared Binaries

**Quick Download Script** (Linux/macOS):
```bash
./scripts/download-binaries.sh
```

**Windows PowerShell:**
```powershell
.\scripts\download-binaries.ps1
```

### Step 4: Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install frontend dependencies
cd frontend
npm install
cd ..
```

### Step 5: Run in Development Mode

```bash
wails dev
```

The application will open automatically with hot-reload enabled.

### Step 6: Configure Backend (Optional)

If you have a backend API:
1. Click the **Settings** tab
2. Enter your backend URL
3. Enter your tunnel name
4. Save settings

### Step 7: Start Your Tunnel

1. Click the **Tunnel** tab
2. Click **▶️ Start Tunnel**
3. The app will fetch a token and start the cloudflared process

### Building for Production

```bash
# Build for current platform
wails build

# Cross-platform builds
wails build -platform windows/amd64      # Windows
wails build -platform darwin/amd64       # macOS Intel
wails build -platform darwin/arm64       # macOS Apple Silicon
wails build -platform linux/amd64        # Linux
```

Built binaries will be in `build/bin/`

---

## Setup Guide

### Complete Setup Instructions

#### Bước 1: Check Prerequisites

```bash
# Check Go version (cần >= 1.21)
go version

# Check Node.js (cần >= 18)
node --version
npm --version
```

#### Bước 2: Install Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Kiểm tra
wails version

# Nếu lỗi "wails: command not found"
export PATH=$PATH:$(go env GOPATH)/bin
# Thêm vào ~/.bashrc hoặc ~/.zshrc để persist
```

#### Bước 3: Clone repo

```bash
git clone https://github.com/votanchat/cloudflared-desktop-tunnel.git
cd cloudflared-desktop-tunnel
```

#### Bước 4: Install dependencies

```bash
# Go dependencies
go mod download

# Frontend dependencies
cd frontend
npm install
cd ..
```

#### Bước 5: Download cloudflared binaries

**Option 1: Dùng script tự động (Khuyến nghị)**

```bash
# Linux/macOS
chmod +x scripts/download-binaries.sh
./scripts/download-binaries.sh

# Windows PowerShell
.\scripts\download-binaries.ps1
```

**Option 2: Manual download**

1. Go to https://github.com/cloudflare/cloudflared/releases/latest
2. Download files and place in correct directories:

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

3. Make executable (Unix):
```bash
chmod +x binaries/darwin/cloudflared-*
chmod +x binaries/linux/cloudflared-*
```

#### Bước 6: Verify binaries

```bash
# Linux/macOS
ls -lh binaries/*/*/*

# Windows PowerShell
Get-ChildItem -Recurse binaries

# Expected output:
# binaries/windows/cloudflared-windows-amd64.exe (~40MB)
# binaries/darwin/cloudflared-darwin-amd64 (~40MB)
# binaries/darwin/cloudflared-darwin-arm64 (~40MB)
# binaries/linux/cloudflared-linux-amd64 (~40MB)
```

#### Bước 7: Run development server

```bash
wails dev
```

### Common Setup Issues

#### Issue 1: Wails command not found

```bash
# Add to PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Make permanent (Linux/macOS)
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc
```

#### Issue 2: GTK errors (Linux)

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.0-dev

# Fedora
sudo dnf install gtk3-devel webkit2gtk3-devel

# Arch
sudo pacman -S gtk3 webkit2gtk
```

#### Issue 3: Frontend build fails

```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
cd ..
```

#### Issue 4: Port 5173 already in use

```bash
# Kill process using port
lsof -ti:5173 | xargs kill -9

# Or change port in frontend/vite.config.ts
```

### Testing Without Backend

#### Option 1: Use Manual Token

Get a tunnel token from Cloudflare:
```bash
cloudflared tunnel token my-tunnel
```

Then paste it directly in the app's Manual Token field.

#### Option 2: Mock Token

Create `app/tunnel_test.go`:

```go
package app

func (a *App) GetMockToken() string {
    return "eyJhIjoiMTIzNDU2Nzg5MGFiY2RlZiIsInQiOiJhYmNkZWYxMi0zNDU2LTc4OTAtYWJjZC1lZjEyMzQ1Njc4OTAiLCJzIjoiWldGaFpHVm1NVEl6TkRVMk56ZzVNR0ZpWTJSbFpnPT0ifQ=="
}
```

---

## Architecture

Cloudflared Desktop Tunnel is a cross-platform desktop application built with Wails v2 that manages Cloudflare Tunnels.

### System Architecture

```
┌────────────────────────────────────────────────┐
│          Frontend (React + TypeScript)         │
│  ┌──────────────────────────────────────────┐  │
│  │ TunnelManager │ StatusDisplay │ Settings │  │
│  └──────────────────────────────────────────┘  │
└────────────────────────────────────────────────┘
                   ↓ Wails Bindings
                   ↓ (TypeScript ↔ Go)
┌────────────────────────────────────────────────┐
│              Backend (Go)                      │
│  ┌──────────────────────────────────────────┐  │
│  │  App (Lifecycle)                         │  │
│  └──────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────┐  │
│  │  TunnelManager                           │  │
│  │  - Download binary from GitHub           │  │
│  │  - Start/stop cloudflared process        │  │
│  │  - Monitor logs                          │  │
│  └──────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────┐  │
│  │  BackendClient                           │  │
│  │  - Fetch tunnel tokens                   │  │
│  │  - WebSocket for commands                │  │
│  │  - Periodic token refresh                │  │
│  └──────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────┐  │
│  │  Config                                  │  │
│  │  - Load/save configuration               │  │
│  │  - Persistent storage                    │  │
│  └──────────────────────────────────────────┘  │
└────────────────────────────────────────────────┘
                   ↓ Process Execution
                   ↓
┌────────────────────────────────────────────────┐
│  Cloudflared Binary (runtime downloaded)       │
└────────────────────────────────────────────────┘
                   ↓ Tunnel Connection
                   ↓
           ┌──────────────────┐
           │  Cloudflare Edge │
           └──────────────────┘
```

### Component Details

#### Frontend Layer (React + TypeScript)

**Location**: `frontend/src/`

**Key Components**:
- **App.tsx**: Main application container with tab navigation
- **TunnelManager.tsx**: Tunnel control interface
- **StatusDisplay.tsx**: Real-time status and logs viewer
- **Settings.tsx**: Configuration management UI

#### Backend Layer (Go)

**Location**: `app/`

**App (app.go)**
- Manages application lifecycle (startup, shutdown)
- Coordinates between TunnelManager, BackendClient, and Config
- Exposes methods to frontend via Wails bindings

**Lifecycle Hooks**:
```go
Startup(ctx)   -> Initialize components
DomReady(ctx)  -> Frontend is ready
Shutdown(ctx)  -> Clean up resources
```

**TunnelManager (tunnel.go)**
- Manages cloudflared tunnel process
- Downloads binary from GitHub on first run
- Monitors process output and status
- Handles process lifecycle

**Key Methods**:
```go
Start(token string) error  // Start tunnel with token
Stop() error              // Stop tunnel
IsRunning() bool          // Check if running
GetLogs() []string        // Get recent logs
```

**BackendClient (backend_client.go)**
- HTTP client for REST API calls
- WebSocket client for real-time commands
- Automatic token refresh mechanism
- Command processor for remote operations

**Backend API Endpoints**:
```
GET  /api/token        -> Fetch tunnel token
POST /api/status       -> Report tunnel status
WS   /api/commands     -> Real-time commands
```

**Config (config.go)**
- Persistent configuration storage
- JSON-based config file
- Platform-specific config directory

**Config Location**:
- **Windows**: `%APPDATA%\cloudflared-desktop-tunnel\config.json`
- **macOS**: `~/Library/Application Support/cloudflared-desktop-tunnel/config.json`
- **Linux**: `~/.config/cloudflared-desktop-tunnel/config.json`

### Binary Download Strategy

The app **automatically downloads** cloudflared binaries at runtime from GitHub releases:

**Download Flow**:
```
1. Application starts
2. TunnelManager.ensureBinary() called
3. Checks cache directory for existing binary
4. If not found, fetches latest version from GitHub API
5. Downloads platform-specific binary
6. Extracts from .tgz (macOS) or saves directly
7. Sets executable permissions (Unix)
8. Caches binary for future use
9. Executes binary with token
```

**Platform Support**:
- Windows: Direct download of `.exe`
- macOS: Downloads `.tgz` and extracts binary
- Linux: Direct download of binary

### Data Flow

#### Starting a Tunnel

```
1. User clicks "Start Tunnel" button
   ↓
2. App checks for manual token:
   - If provided: use manual token
   - If not: fetch from backend API
   ↓
3. TunnelManager downloads/uses cached binary
   ↓
4. cloudflared process starts with token
   ↓
5. Process logs streamed to frontend
   ↓
6. Status updates displayed in real-time
```

#### Backend Commands

```
1. Backend sends command via WebSocket
   ↓
2. BackendClient receives message
   ↓
3. Command added to commandsCh channel
   ↓
4. processCommands() handles command
   ↓
5. Action executed (update, restart, etc.)
   ↓
6. Status reported back to backend
```

### Cross-Platform Considerations

**Binary Caching**:
- **Windows**: Cache to `%LOCALAPPDATA%\cloudflared-desktop-tunnel\`
- **macOS**: Cache to `~/Library/Caches/cloudflared-desktop-tunnel/`
- **Linux**: Cache to `~/.cache/cloudflared-desktop-tunnel/`

**File Permissions**:
- Unix systems: `chmod +x` automatically applied
- Windows: No additional permissions needed

**Process Management**:
- Cross-platform using Go's `os/exec` package
- Graceful shutdown with process.Kill()
- Cleanup on application exit

### Performance

**Memory Usage**:
- Go backend: ~20-30 MB
- React frontend: ~50-70 MB
- cloudflared process: ~30-50 MB
- **Total**: ~100-150 MB

**Binary Size**:
- Application (no embedded binaries): ~15-20 MB
- Cached cloudflared binary: ~45-50 MB
- Combined: ~60-70 MB per platform

---

## Backend API

This section describes the backend API that the Cloudflared Desktop Tunnel application connects to for token management and remote commands.

### Base URL

Configured in the application settings. Default: `https://api.example.com`

### Authentication

Currently, no authentication is implemented in the demo. In production, implement:
- API key authentication
- JWT tokens
- OAuth 2.0
- mTLS (mutual TLS)

### REST API Endpoints

#### 1. Get Tunnel Token

**Endpoint**: `GET /api/token`

**Description**: Fetches a Cloudflare tunnel token for the client to use.

**Request**:
```http
GET /api/token HTTP/1.1
Host: api.example.com
Content-Type: application/json
```

**Response**:
```json
{
  "token": "eyJhIjoiMTIzNDU2Nzg5MGFiY2RlZiIsInQiOiJhYmNkZWYxMi0zNDU2LTc4OTAtYWJjZC1lZjEyMzQ1Njc4OTAiLCJzIjoiWldGaFpHVm1NVEl6TkRVMk56ZzVNR0ZpWTJSbFpnPT0ifQ==",
  "expiresAt": "2025-11-15T12:00:00Z"
}
```

**Status Codes**:
- `200 OK`: Token fetched successfully
- `401 Unauthorized`: Authentication failed
- `500 Internal Server Error`: Server error

#### 2. Report Tunnel Status

**Endpoint**: `POST /api/status`

**Description**: Reports the current tunnel status to the backend.

**Request**:
```http
POST /api/status HTTP/1.1
Host: api.example.com
Content-Type: application/json

{
  "running": true,
  "tunnelName": "my-tunnel",
  "connections": 4,
  "uptime": 3600,
  "version": "1.0.0",
  "timestamp": "2025-11-14T10:00:00Z"
}
```

**Response**:
```json
{
  "status": "ok",
  "message": "Status received"
}
```

### WebSocket API

#### Commands WebSocket

**Endpoint**: `WS /api/commands`

**Description**: WebSocket connection for receiving real-time commands from the backend.

**Connection**:
```
ws://api.example.com/api/commands
wss://api.example.com/api/commands  (secure)
```

#### Command Format

All commands are sent as JSON messages:

```json
{
  "type": "command_type",
  "payload": {
    // Command-specific data
  }
}
```

#### Supported Commands

**Update Command**:
```json
{
  "type": "update",
  "payload": {
    "version": "1.1.0",
    "url": "https://releases.example.com/v1.1.0/app.exe",
    "checksum": "sha256:abc123...",
    "force": false
  }
}
```

**Restart Command**:
```json
{
  "type": "restart",
  "payload": {
    "reason": "Configuration updated",
    "delay": 5
  }
}
```

**Patch Command**:
```json
{
  "type": "patch",
  "payload": {
    "config": {
      "refreshInterval": 600,
      "backendURL": "https://new-api.example.com"
    }
  }
}
```

**Stop Command**:
```json
{
  "type": "stop",
  "payload": {
    "reason": "Maintenance window"
  }
}
```

**Fetch Logs Command**:
```json
{
  "type": "fetch_logs",
  "payload": {
    "lines": 100
  }
}
```

### Example Backend Implementation

#### Node.js + Express Example

```javascript
const express = require('express');
const expressWs = require('express-ws');

const app = express();
expressWs(app);

app.use(express.json());

const clients = new Set();

// GET /api/token
app.get('/api/token', (req, res) => {
  const token = generateTunnelToken();
  
  res.json({
    token: token,
    expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000)
  });
});

// POST /api/status
app.post('/api/status', (req, res) => {
  const status = req.body;
  console.log('Received status:', status);
  
  res.json({ status: 'ok', message: 'Status received' });
});

// WS /api/commands
app.ws('/api/commands', (ws, req) => {
  console.log('Client connected');
  clients.add(ws);
  
  ws.on('message', (msg) => {
    const message = JSON.parse(msg);
    console.log('Received:', message);
  });
  
  ws.on('close', () => {
    console.log('Client disconnected');
    clients.delete(ws);
  });
});

app.listen(3000, () => {
  console.log('Backend API running on port 3000');
});
```

### Security Recommendations

**Production Deployment**:
1. Use HTTPS/WSS for encrypted connections
2. Implement Authentication (JWT, OAuth 2.0, etc.)
3. Rate Limiting to prevent abuse
4. Input Validation for all incoming data
5. Token Rotation for security
6. Audit Logging for all API calls
7. CORS configuration

---

## Manual Token Usage

### Tổng quan

App hỗ trợ 2 cách lấy token để start tunnel:

1. **Tự động từ Backend** (mặc định) - App sẽ gọi API backend để lấy token
2. **Manual Token** - Bạn tự paste token vào UI

### Khi nào dùng Manual Token?

✅ **Nên dùng khi**:
- Testing/development mà chưa có backend
- Muốn kiểm soát token cụ thể
- Backend tạm thời không khả dụng
- Debug tunnel connection issues

❌ **Không nên dùng trong production**:
- Không secure để user paste token trực tiếp
- Khó quản lý token rotation
- Không có centralized control

### Cách lấy Cloudflare Tunnel Token

#### Option 1: Cloudflare Dashboard

1. Đăng nhập [Cloudflare Zero Trust Dashboard](https://one.dash.cloudflare.com/)
2. Vào **Networks** → **Tunnels**
3. Click vào tunnel của bạn
4. Tab **Configure** → scroll down → **Connector** → Click **View token**
5. Copy token

#### Option 2: CLI (Khuyến nghị)

```bash
# 1. Login Cloudflare
cloudflared tunnel login

# 2. Tạo tunnel mới (nếu chưa có)
cloudflared tunnel create my-test-tunnel

# 3. Lấy token
cloudflared tunnel token my-test-tunnel

# Copy token từ output
```

### Token Format

Token có dạng:
```
eyJhIjoiMTIzNDU2Nzg5MGFiY2RlZiIsInQiOiJhYmNkZWYxMi0zNDU2LTc4OTAtYWJjZC1lZjEyMzQ1Njc4OTAiLCJzIjoiWldGaFpHVm1NVEl6TkRVMk56ZzVNR0ZpWTJSbFpnPT0ifQ==
```

Đây là base64-encoded JSON chứa: Account ID, Tunnel ID, Secret

### Cách sử dụng trong App

1. **Mở Tab "Tunnel"**
2. **Click nút "✏️ Manual Token"** - section "🔑 Token Configuration" sẽ mở
3. **Paste token vào textarea**
4. **Click "▶️ Start Tunnel"**

App sẽ:
- ✅ Sử dụng token bạn vừa paste
- ✅ **KHÔNG** gọi backend API
- ✅ Start cloudflared với token đó
- ✅ Tự động xóa token khỏi input sau khi start (security)

### Flow Diagram

```
User clicks "Start Tunnel"
         |
         v
Manual Token field có giá trị?
         |
    +----+----+
    |         |
   YES       NO
    |         |
    v         v
Dùng      Gọi Backend
Manual    GET /api/token
Token          |
    |          v
    |     Dùng Backend
    |        Token
    |          |
    +----+-----+
         |
         v
    Start cloudflared
    với token
```

### Security Notes

✅ **Good Practices**:
1. Token chỉ dùng cho development/testing
2. Không commit token vào Git
3. Token tự động xóa khỏi UI sau khi start
4. Token không được lưu vào config file
5. Token không được log ra console

❌ **Không nên**:
1. Share token qua chat/email
2. Hardcode token trong code
3. Dùng production token cho testing
4. Để token trong clipboard lâu

---

## Troubleshooting

### Binary Extraction Issues

#### 1. "exec format error" - Wrong Architecture

**Error**:
```
failed to start tunnel: fork/exec /path/to/cloudflared-darwin-arm64: exec format error
```

**Nguyên nhân**: Binary không khớp với kiến trúc máy bạn

**Giải pháp**:

1. Check kiến trúc máy:
```bash
uname -m
# Output: x86_64 (Intel), arm64 (ARM64)
```

2. Check binary type:
```bash
file binaries/darwin/cloudflared-darwin-arm64
# Output phải có: "Mach-O 64-bit executable arm64"
```

3. Re-download đúng binary:
```bash
./scripts/download-binaries.sh
```

4. Clear cache:
```bash
# macOS
rm -rf ~/Library/Caches/cloudflared-desktop-tunnel/*

# Linux
rm -rf ~/.cache/cloudflared-desktop-tunnel/*
```

#### 2. Binary Caching

**App bây giờ cache binary để tăng performance**

**Cache location**:
- **macOS**: `~/Library/Caches/cloudflared-desktop-tunnel/`
- **Linux**: `~/.cache/cloudflared-desktop-tunnel/`
- **Windows**: `%LOCALAPPDATA%\cloudflared-desktop-tunnel\`

**Clear cache khi cần**:
```bash
# macOS
rm -rf ~/Library/Caches/cloudflared-desktop-tunnel

# Linux  
rm -rf ~/.cache/cloudflared-desktop-tunnel

# Windows (PowerShell)
Remove-Item -Recurse -Force "$env:LOCALAPPDATA\cloudflared-desktop-tunnel"
```

#### 3. "Binary is empty" Error

**Error**:
```
embedded binary is empty - did you download the cloudflared binaries?
```

**Giải pháp**:
```bash
./scripts/download-binaries.sh
ls -lh binaries/*/*/*  # Mỗi file phải > 40MB
```

#### 4. "Binary file too small" Error

**Giải pháp**:
```bash
rm binaries/darwin/cloudflared-*
./scripts/download-binaries.sh
du -h binaries/*/*/*  # Mỗi file phải ~40-50MB
```

#### 5. Permission Denied

**Giải pháp**:
```bash
chmod +x binaries/darwin/cloudflared-*
chmod +x binaries/linux/cloudflared-*
rm -rf ~/Library/Caches/cloudflared-desktop-tunnel
```

### Wails Runtime Issues

#### "Cannot read properties of undefined (reading 'app')"

**Nguyên nhân**: Wails runtime chưa initialize

**Giải pháp**:
```bash
# ĐÚNG:
wails dev

# SAI:
npm run dev  # ❌
```

### Backend Connection Issues

#### "Failed to fetch token from backend"

**Giải pháp 1: Dùng Manual Token**
```bash
cloudflared tunnel token my-tunnel
# Paste vào app UI
```

**Giải pháp 2: Check Backend**
```bash
curl http://localhost:3000/api/token
```

### Build Issues

#### Cross-compilation Failed

**macOS → Windows**:
```bash
brew install mingw-w64
wails build -platform windows/amd64
```

### Common Fixes Summary

| Issue | Fix |
|-------|-----|
| exec format error | Download correct architecture binary |
| Binary too small | Re-download binaries |
| Permission denied | `chmod +x binaries/**/*` |
| Wails undefined | Run `wails dev` not `npm run dev` |
| Backend connection | Use manual token |
| Slow start | Binary is cached after first run |

---

## Contributing

Thank you for your interest in contributing to Cloudflared Desktop Tunnel!

### Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally
3. **Create a feature branch**

### Development Setup

```bash
# Install Go dependencies
go mod download

# Install frontend dependencies
cd frontend
npm install
cd ..

# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### Download Cloudflared Binaries

```bash
./scripts/download-binaries.sh
```

### Running in Development Mode

```bash
wails dev
```

### Code Style

**Go Code**:
- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Run `go fmt` before committing
- Run `go vet` to check for common mistakes
- Add comments for exported functions and types

**TypeScript/React Code**:
- Use TypeScript for type safety
- Follow React best practices and hooks patterns
- Use functional components
- Add proper TypeScript types

### Testing

```bash
# Backend Tests
go test ./...

# Frontend Tests
cd frontend
npm test
```

### Pull Request Process

1. Update documentation if changing functionality
2. Add tests for new features
3. Ensure all tests pass
4. Update CHANGELOG.md with your changes
5. Create a pull request with a clear title and description

**PR Title Format** (Conventional Commits):
- `feat: Add new feature`
- `fix: Fix bug in tunnel manager`
- `docs: Update README`
- `style: Format code`
- `refactor: Refactor backend client`
- `test: Add tests for config`

### Reporting Bugs

Include:
1. Description of the issue
2. Steps to reproduce
3. Expected behavior
4. Actual behavior
5. System information (OS, Go version, Wails version)
6. Logs if applicable

### Feature Requests

1. Check existing issues to avoid duplicates
2. Clearly describe the feature and its use case
3. Explain why it would be beneficial
4. Provide examples if possible

---

## Code Optimizations

Applied targeted optimizations across the Go codebase to improve performance, reduce allocations, and follow Go best practices.

### Changes by File

#### app/tunnel.go

1. **readLogs() function (line 320)**
   - Added `defer pipe.Close()`: Ensures pipe is properly closed
   - Increased buffer from 1KB to 4KB: Reduces system calls for I/O operations
   - Optimized log retention: Changed from slice reslicing to `copy() + reslice` pattern, reducing memory allocations

2. **generateConfigFile() function (line 248)**
   - Pre-allocated `strings.Builder`: Used `Grow()` to reserve capacity, reducing internal allocations
   - Eliminated `fmt.Sprintf` calls: Replaced with direct `WriteString()` calls for better performance
   - Cached `yaml.String()` result: Stored in variable to avoid multiple conversions

#### app/config.go

1. **AddRoute() function (line 37)**
   - Optimized loop: Changed from `for i, route := range` to `for i := range` to avoid copying Route struct on each iteration

2. **RemoveRoute() function (line 54)**
   - Optimized loop: Changed from `for i, route := range` to `for i := range` to avoid unnecessary struct copies

#### binaries/downloader.go

1. **getLatestVersion() function (line 74)**
   - Moved HTTP client to module level: Created package-level `githubClient` variable instead of allocating new client on every call
   - Reuses connection pooling: HTTP client connection pools are now persistent across function calls

### Performance Impact

| Change | Type | Benefit |
|--------|------|---------|
| 4KB buffer in readLogs | I/O | Fewer read syscalls |
| Grow() in YAML builder | Memory | Reduced allocations during config generation |
| Eliminate fmt.Sprintf | CPU | Faster string formatting (no format parsing) |
| Avoid struct copies in loops | Memory/CPU | Less memory pressure, faster iterations |
| Module-level HTTP client | Memory | Connection pooling, reduced allocations |
| copy() for log retention | Memory | Avoids new slice allocation every 100 lines |

### Verification

All changes verified to:
- Compile without errors (`go build` successful)
- Maintain backward compatibility
- Follow Go idioms and best practices

---

## Helper Scripts

This directory contains helper scripts for the Cloudflared Desktop Tunnel project.

### Download Binaries Scripts

Automatically download the latest official cloudflared binaries from Cloudflare's GitHub releases.

#### For Linux/macOS

**Script**: `scripts/download-binaries.sh`

**Usage**:
```bash
chmod +x scripts/download-binaries.sh
./scripts/download-binaries.sh
```

This will download:
- Windows AMD64 binary
- macOS Intel (AMD64) binary
- macOS Apple Silicon (ARM64) binary
- Linux AMD64 binary
- Linux ARM64 binary

**Requirements**:
- `curl` or `wget` installed
- `tar` for extracting macOS binaries
- Internet connection

#### For Windows

**Script**: `scripts/download-binaries.ps1`

**Usage** (PowerShell):
```powershell
.\scripts\download-binaries.ps1
```

**Requirements**:
- PowerShell 5.0 or higher
- Internet connection

### Manual Download

If the scripts don't work, manually download from:
https://github.com/cloudflare/cloudflared/releases/latest

| Platform | Download File | Place In |
|----------|---------------|----------|
| Windows AMD64 | `cloudflared-windows-amd64.exe` | `binaries/windows/cloudflared-windows-amd64.exe` |
| macOS Intel | `cloudflared-darwin-amd64.tgz` | Extract to `binaries/darwin/cloudflared-darwin-amd64` |
| macOS ARM | `cloudflared-darwin-arm64.tgz` | Extract to `binaries/darwin/cloudflared-darwin-arm64` |
| Linux AMD64 | `cloudflared-linux-amd64` | `binaries/linux/cloudflared-linux-amd64` |
| Linux ARM64 | `cloudflared-linux-arm64` | `binaries/linux/cloudflared-linux-arm64` |

### Setting Permissions (Unix)

After downloading, make the binaries executable:

```bash
chmod +x binaries/darwin/cloudflared-darwin-*
chmod +x binaries/linux/cloudflared-linux-*
```

---

## Binary Management

Binaries are now **automatically downloaded at runtime** from GitHub releases when the application starts.

### Runtime Download

When the application starts and a tunnel is initiated, it will:

1. **Check cache**: Look for existing binary in platform-specific cache directory
2. **Download if needed**: If no valid binary found, download latest version from [Cloudflare's GitHub releases](https://github.com/cloudflare/cloudflared/releases)
3. **Cache for reuse**: Save the downloaded binary to avoid re-downloading on subsequent runs
4. **Verify and execute**: Validate the binary and set executable permissions

### Cache Locations

- **Windows**: `%LOCALAPPDATA%\cloudflared-desktop-tunnel\`
- **macOS**: `~/Library/Caches/cloudflared-desktop-tunnel/`
- **Linux**: `~/.cache/cloudflared-desktop-tunnel/` (or `$XDG_CACHE_HOME/cloudflared-desktop-tunnel/`)

### Platform Support

The downloader automatically detects the platform and downloads the appropriate binary:

- **Windows**: `cloudflared-windows-amd64.exe`
- **macOS Intel**: `cloudflared-darwin-amd64` (from .tgz archive)
- **macOS ARM (M1/M2/M3)**: `cloudflared-darwin-arm64` (from .tgz archive)
- **Linux AMD64**: `cloudflared-linux-amd64`
- **Linux ARM64**: `cloudflared-linux-arm64`

### Benefits

✅ **No manual setup**: Binaries are downloaded automatically  
✅ **Always up-to-date**: Gets the latest version from GitHub  
✅ **Smaller app size**: Application binary doesn't include embedded binaries  
✅ **Cross-platform**: Works on Windows, macOS, and Linux  
✅ **Cached**: Downloads only once, reuses cached binary  

### Network Requirements

The application requires internet access on first run to download the cloudflared binary. Subsequent runs will use the cached binary unless it's deleted or invalid.

---

## Resources

- 📚 [Wails Docs](https://wails.io/docs)
- 🔒 [Cloudflare Tunnel Docs](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/)
- 🐛 [GitHub Issues](https://github.com/votanchat/cloudflared-desktop-tunnel/issues)
- 💬 [Discussions](https://github.com/votanchat/cloudflared-desktop-tunnel/discussions)
- ⭐ [Repository](https://github.com/votanchat/cloudflared-desktop-tunnel)

---

## License

MIT License - see LICENSE file for details

## Acknowledgments

Built with:
- [Wails](https://wails.io) - Build desktop apps using Go & Web Technologies
- [Cloudflare Tunnel](https://www.cloudflare.com/products/tunnel/) - Secure tunnel service
- [React](https://react.dev) - Frontend library
