# Hướng dẫn Setup Chi tiết

## Vấn đề 1: Lỗi "Cannot read properties of undefined (reading 'app')"

### Nguyên nhân
Wails runtime chưa được khởi tạo hoặc app không chạy trong môi trường Wails.

### Giải pháp

✅ **Đã fix** trong code mới nhất:
- Thêm check `window.go` trước khi gọi methods
- Thêm loading state cho Wails runtime
- Error messages rõ ràng hơn

### Cách chạy đúng

```bash
# KHÔNG chạy như web app thông thường
# npm run dev  ❌ SAI

# PHẢI chạy qua Wails
wails dev  ✅ ĐÚNG
```

## Vấn đề 2: Không tự động download cloudflared binary

### Giải thích

Repo này **không tự động download** binary vì:
1. Binary size lớn (~40-50MB per platform)
2. Security - chỉ nên dùng official binaries
3. Git không phù hợp để lưu large binaries

### Giải pháp: Download thủ công

#### Option 1: Dùng script tự động (Khuyến nghị)

**Linux/macOS:**
```bash
chmod +x scripts/download-binaries.sh
./scripts/download-binaries.sh
```

**Windows (PowerShell):**
```powershell
.\scripts\download-binaries.ps1
```

#### Option 2: Download thủ công

1. Vào https://github.com/cloudflare/cloudflared/releases/latest

2. Download các file tương ứng:

**Windows:**
```
Download: cloudflared-windows-amd64.exe
Đặt vào:  binaries/windows/cloudflared-windows-amd64.exe
```

**macOS:**
```bash
# Intel Mac
Download: cloudflared-darwin-amd64.tgz
Giải nén: tar -xzf cloudflared-darwin-amd64.tgz
Đặt vào:  binaries/darwin/cloudflared-darwin-amd64
Chmod:    chmod +x binaries/darwin/cloudflared-darwin-amd64

# Apple Silicon (M1/M2/M3)
Download: cloudflared-darwin-arm64.tgz
Giải nén: tar -xzf cloudflared-darwin-arm64.tgz
Đặt vào:  binaries/darwin/cloudflared-darwin-arm64
Chmod:    chmod +x binaries/darwin/cloudflared-darwin-arm64
```

**Linux:**
```bash
# AMD64
Download: cloudflared-linux-amd64
Đặt vào:  binaries/linux/cloudflared-linux-amd64
Chmod:    chmod +x binaries/linux/cloudflared-linux-amd64

# ARM64 (optional)
Download: cloudflared-linux-arm64
Đặt vào:  binaries/linux/cloudflared-linux-arm64
Chmod:    chmod +x binaries/linux/cloudflared-linux-arm64
```

## Setup từ đầu (Complete Guide)

### Bước 1: Prerequisites

```bash
# Check Go version (cần >= 1.21)
go version

# Check Node.js (cần >= 18)
node --version
npm --version
```

### Bước 2: Install Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Kiểm tra
wails version

# Nếu lỗi "wails: command not found"
export PATH=$PATH:$(go env GOPATH)/bin
# Thêm vào ~/.bashrc hoặc ~/.zshrc để persist
```

### Bước 3: Clone repo

```bash
git clone https://github.com/votanchat/cloudflared-desktop-tunnel.git
cd cloudflared-desktop-tunnel
```

### Bước 4: Install dependencies

```bash
# Go dependencies
go mod download

# Frontend dependencies
cd frontend
npm install
cd ..
```

### Bước 5: Download cloudflared binaries

```bash
# Linux/macOS
./scripts/download-binaries.sh

# Windows PowerShell
.\scripts\download-binaries.ps1

# Hoặc download thủ công theo hướng dẫn ở trên
```

### Bước 6: Verify binaries

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

### Bước 7: Run development server

```bash
wails dev
```

App sẽ tự động mở với hot-reload enabled.

## Testing Without Backend

### Option 1: Mock Backend (Khuyến nghị cho testing)

Tạo file `app/tunnel_test.go`:

```go
package app

import "fmt"

// GetMockToken returns a mock token for testing
func (a *App) GetMockToken() string {
    // Lấy token thật từ:
    // cloudflared tunnel token <tunnel-id>
    return "eyJhIjoiMTIzNDU2Nzg5MGFiY2RlZiIsInQiOiJhYmNkZWYxMi0zNDU2LTc4OTAtYWJjZC1lZjEyMzQ1Njc4OTAiLCJzIjoiWldGaFpHVm1NVEl6TkRVMk56ZzVNR0ZpWTJSbFpnPT0ifQ=="
}
```

Sửa `app/app.go`, method `StartTunnel()`:

```go
func (a *App) StartTunnel() error {
    if a.tunnel.IsRunning() {
        return fmt.Errorf("tunnel is already running")
    }

    // For testing: use mock token instead of backend
    // token, err := a.backendClient.FetchToken()
    token := a.GetMockToken() // Use mock token
    var err error = nil
    
    if err != nil {
        return fmt.Errorf("failed to fetch token: %w", err)
    }

    return a.tunnel.Start(token)
}
```

### Option 2: Tạo tunnel token thật

```bash
# Login Cloudflare
cloudflared tunnel login

# Tạo tunnel mới
cloudflared tunnel create my-test-tunnel

# Lấy token
cloudflared tunnel token my-test-tunnel

# Copy token và paste vào GetMockToken()
```

## Troubleshooting

### 1. Wails command not found

```bash
# Add to PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Make permanent (Linux/macOS)
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc
```

### 2. GTK errors (Linux)

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.0-dev

# Fedora
sudo dnf install gtk3-devel webkit2gtk3-devel

# Arch
sudo pacman -S gtk3 webkit2gtk
```

### 3. Frontend build fails

```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
cd ..
```

### 4. Binary extraction fails

**Error:** `permission denied`

```bash
# Unix systems: make sure binaries are executable
chmod +x binaries/darwin/cloudflared-darwin-*
chmod +x binaries/linux/cloudflared-linux-*
```

### 5. Port 5173 already in use

```bash
# Kill process using port
lsof -ti:5173 | xargs kill -9

# Or change port in frontend/vite.config.ts
```

### 6. "Failed to start tunnel" error

**Nguyên nhân:**
- Backend không accessible
- Token không hợp lệ
- Binary không extract được

**Debug:**
```bash
# Check logs trong app
# Hoặc xem console trong DevTools (Wails dev mode)
```

## Build cho Production

### Build current platform

```bash
wails build
```

Output: `build/bin/cloudflared-tunnel` (hoặc `.exe` trên Windows)

### Cross-platform builds

```bash
# Windows từ Linux/macOS
wails build -platform windows/amd64

# macOS từ Linux (requires osxcross)
wails build -platform darwin/amd64
wails build -platform darwin/arm64

# Linux từ macOS/Windows
wails build -platform linux/amd64
```

**Lưu ý:** Cross-compilation có thể yêu cầu setup thêm toolchains.

## Build với GitHub Actions

Push tag để trigger CI/CD:

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

GitHub Actions sẽ tự động build cho cả 3 platforms và tạo release.

## Cấu trúc Binary sau khi Build

```
build/bin/
├── cloudflared-tunnel           # macOS/Linux executable
└── cloudflared-tunnel.exe       # Windows executable

Size: ~60-70MB (đã bao gồm embedded cloudflared binary)
```

## Next Steps

1. ✅ Setup xong? → Test start/stop tunnel
2. ✅ Muốn connect backend thật? → Xem `docs/BACKEND_API.md`
3. ✅ Muốn customize UI? → Edit `frontend/src/`
4. ✅ Muốn thêm features? → Xem `CONTRIBUTING.md`

## Getting Help

- 📚 [Wails Docs](https://wails.io/docs)
- 🔒 [Cloudflare Tunnel Docs](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/)
- 🐛 [GitHub Issues](https://github.com/votanchat/cloudflared-desktop-tunnel/issues)
- 💬 [Discussions](https://github.com/votanchat/cloudflared-desktop-tunnel/discussions)
