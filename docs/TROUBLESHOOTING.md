# Troubleshooting Guide

## Binary Extraction Issues

### 1. "exec format error" - Wrong Architecture

**Error:**
```
failed to start tunnel: fork/exec /path/to/cloudflared-darwin-arm64: exec format error
```

**Nguyên nhân:**
- Binary không khớp với kiến trúc máy bạn
- Ví dụ: Chạy ARM64 binary trên Intel Mac, hoặc ngược lại

**Giải pháp:**

1. **Kiểm tra kiến trúc máy:**
```bash
# macOS/Linux
uname -m
# Output: 
# - x86_64 = Intel/AMD64
# - arm64 hoặc aarch64 = ARM64

# Trong app, check runtime
echo "GOOS: $GOOS, GOARCH: $GOARCH"
```

2. **Kiểm tra binary đã download:**
```bash
ls -lh binaries/darwin/
# Phải có:
# - cloudflared-darwin-amd64 (cho Intel Mac)
# - cloudflared-darwin-arm64 (cho M1/M2/M3 Mac)

# Check binary architecture
file binaries/darwin/cloudflared-darwin-arm64
# Output phải có: "Mach-O 64-bit executable arm64"

file binaries/darwin/cloudflared-darwin-amd64  
# Output phải có: "Mach-O 64-bit executable x86_64"
```

3. **Re-download đúng binary:**
```bash
# Xóa binaries cũ
rm -rf binaries/darwin/*

# Download lại
./scripts/download-binaries.sh

# Hoặc download manual từ:
# https://github.com/cloudflare/cloudflared/releases/latest
```

4. **Clear cache và thử lại:**
```bash
# macOS
rm -rf ~/Library/Caches/cloudflared-desktop-tunnel/*

# Linux
rm -rf ~/.cache/cloudflared-desktop-tunnel/*

# Restart app
wails dev
```

### 2. Binary Caching

**App bây giờ cache binary để tăng performance:**

**Cache location:**
- **macOS**: `~/Library/Caches/cloudflared-desktop-tunnel/`
- **Linux**: `~/.cache/cloudflared-desktop-tunnel/`
- **Windows**: `%LOCALAPPDATA%\cloudflared-desktop-tunnel\`

**Lợi ích:**
- ✅ Chỉ extract 1 lần
- ✅ Start tunnel nhanh hơn (không phải copy file mỗi lần)
- ✅ Verify binary integrity
- ✅ Auto re-extract nếu cache bị corrupt

**Clear cache khi cần:**
```bash
# macOS
rm -rf ~/Library/Caches/cloudflared-desktop-tunnel

# Linux  
rm -rf ~/.cache/cloudflared-desktop-tunnel

# Windows (PowerShell)
Remove-Item -Recurse -Force "$env:LOCALAPPDATA\cloudflared-desktop-tunnel"
```

### 3. "Binary is empty" Error

**Error:**
```
embedded binary is empty - did you download the cloudflared binaries?
```

**Nguyên nhân:**
Bạn chưa download cloudflared binaries.

**Giải pháp:**
```bash
# Run script để download
./scripts/download-binaries.sh

# Verify
ls -lh binaries/*/*/*
# Mỗi file phải > 40MB
```

### 4. "Binary file too small" Error

**Error trong logs:**
```
Binary file too small: 1024 bytes
```

**Nguyên nhân:**
- Download không hoàn tất
- File bị corrupt
- Placeholder file thay vì binary thật

**Giải pháp:**
```bash
# Remove corrupted file
rm binaries/darwin/cloudflared-*

# Re-download
./scripts/download-binaries.sh

# Verify size
du -h binaries/*/*/*
# Mỗi file phải ~40-50MB
```

### 5. Permission Denied

**Error:**
```
failed to start tunnel: permission denied
```

**Giải pháp:**
```bash
# Unix: Make binaries executable
chmod +x binaries/darwin/cloudflared-*
chmod +x binaries/linux/cloudflared-*

# Clear cache và để app re-extract
rm -rf ~/Library/Caches/cloudflared-desktop-tunnel
```

## Wails Runtime Issues

### "Cannot read properties of undefined (reading 'app')"

**Nguyên nhân:** Wails runtime chưa initialize.

**Giải pháp:**
```bash
# ĐÚNG:
wails dev

# SAI:
npm run dev  # ❌
cd frontend && npm run dev  # ❌
```

## Backend Connection Issues

### "Failed to fetch token from backend"

**Giải pháp 1: Dùng Manual Token**
```bash
# Get token từ Cloudflare
cloudflared tunnel token my-tunnel

# Paste vào app UI
# Tab Tunnel → ✏️ Manual Token
```

**Giải pháp 2: Check Backend**
```bash
# Test backend endpoint
curl http://localhost:3000/api/token

# Check backend logs
```

## Build Issues

### Cross-compilation Failed

**macOS → Windows:**
```bash
# Install MinGW
brew install mingw-w64

# Build
wails build -platform windows/amd64
```

**Linux → macOS:**
```bash
# Cần osxcross toolchain
# https://github.com/tpoechtrager/osxcross
```

## Performance Issues

### App Slow to Start Tunnel

**Có thể do:**
- Binary cache miss → re-extracting
- Network lag khi fetch token
- Large binary size

**Optimization:**
- Binary đã được cache sau lần đầu
- Dùng manual token để bypass backend call
- Binary chỉ extract 1 lần

## Debug Tips

### Enable Verbose Logging

**Check app logs:**
```bash
# Logs trong terminal khi chạy
wails dev

# Tìm các dòng:
# - "Using cloudflared binary: /path/to/binary"
# - "Runtime: GOOS=darwin, GOARCH=arm64"
# - "Binary size: X bytes"
# - "Using cached binary" hoặc "Extracting embedded binary"
```

### Verify Binary

```bash
# Check binary type
file ~/Library/Caches/cloudflared-desktop-tunnel/cloudflared-darwin-arm64

# Run binary directly
~/Library/Caches/cloudflared-desktop-tunnel/cloudflared-darwin-arm64 version

# Should output cloudflared version
```

### Check Process

```bash
# While tunnel is running
ps aux | grep cloudflared

# Should show:
# /path/to/cached/binary tunnel run --token ...
```

## Common Fixes Summary

| Issue | Fix |
|-------|-----|
| exec format error | Download correct architecture binary |
| Binary too small | Re-download binaries |
| Permission denied | `chmod +x binaries/**/*` |
| Wails undefined | Run `wails dev` not `npm run dev` |
| Backend connection | Use manual token |
| Slow start | Binary is cached after first run |
| Cache issues | Clear cache directory |

## Getting More Help

1. Check logs trong terminal
2. Xem [SETUP.md](../SETUP.md) cho detailed setup
3. Xem [MANUAL_TOKEN.md](./MANUAL_TOKEN.md) để bypass backend
4. Open issue trên GitHub với:
   - OS và architecture (`uname -a`)
   - Error message đầy đủ
   - Output của `ls -lh binaries/*/*/*`
   - Cloudflared version bạn download

## Related Docs

- [Setup Guide](../SETUP.md)
- [Manual Token Guide](./MANUAL_TOKEN.md)
- [Architecture Documentation](./ARCHITECTURE.md)
