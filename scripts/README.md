# Helper Scripts

This directory contains helper scripts for the Cloudflared Desktop Tunnel project.

## Download Binaries Scripts

These scripts automatically download the latest official cloudflared binaries from Cloudflare's GitHub releases.

### For Linux/macOS

**Script**: `download-binaries.sh`

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

All binaries will be placed in the correct directories under `binaries/`.

**Requirements**:
- `curl` or `wget` installed
- `tar` for extracting macOS binaries
- Internet connection

### For Windows

**Script**: `download-binaries.ps1`

**Usage** (PowerShell):
```powershell
.\scripts\download-binaries.ps1
```

This will download:
- Windows AMD64 binary

The binary will be placed in `binaries\windows\`.

**Requirements**:
- PowerShell 5.0 or higher
- Internet connection

## Manual Download

If the scripts don't work for you, manually download from:

https://github.com/cloudflare/cloudflared/releases/latest

### File Mapping

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

## Verify Downloads

After downloading, verify the files exist:

```bash
# Linux/macOS
ls -lh binaries/*/*/*

# Windows (PowerShell)
Get-ChildItem -Recurse binaries
```

## Troubleshooting

### wget not found (macOS)
```bash
brew install wget
```

### Permission denied
```bash
chmod +x scripts/download-binaries.sh
```

### Network errors
- Check your internet connection
- Try using a VPN if GitHub is blocked
- Download manually from the releases page

### Extraction errors (macOS binaries)
The macOS binaries come as `.tgz` files. The script extracts them automatically, but if you're doing it manually:

```bash
tar -xzf cloudflared-darwin-amd64.tgz
mv cloudflared binaries/darwin/cloudflared-darwin-amd64
chmod +x binaries/darwin/cloudflared-darwin-amd64
```

## Security Note

Always download binaries from official sources:
- Official GitHub: https://github.com/cloudflare/cloudflared/releases
- Verify checksums when available
- Check the release is signed by Cloudflare

Never download from unofficial or third-party sources.
