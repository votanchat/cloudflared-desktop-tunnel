# PowerShell script to download cloudflared binaries for Windows
# Run this script on Windows to download the required binary

Write-Host "Cloudflared Binary Downloader" -ForegroundColor Blue
Write-Host "==============================" -ForegroundColor Blue
Write-Host ""

# Get the latest version
Write-Host "Fetching latest cloudflared version..." -ForegroundColor Yellow
$latestRelease = Invoke-RestMethod -Uri "https://api.github.com/repos/cloudflare/cloudflared/releases/latest"
$latestVersion = $latestRelease.tag_name
Write-Host "Latest version: $latestVersion" -ForegroundColor Green
Write-Host ""

# Base URL for downloads
$baseUrl = "https://github.com/cloudflare/cloudflared/releases/download/$latestVersion"

# Create directories
New-Item -ItemType Directory -Force -Path "binaries\windows" | Out-Null

# Download Windows binary
Write-Host "Downloading Windows binary..." -ForegroundColor Yellow
$downloadUrl = "$baseUrl/cloudflared-windows-amd64.exe"
$outputPath = "binaries\windows\cloudflared-windows-amd64.exe"

Invoke-WebRequest -Uri $downloadUrl -OutFile $outputPath
Write-Host "✓ Windows binary downloaded" -ForegroundColor Green
Write-Host ""

Write-Host "==============================" -ForegroundColor Green
Write-Host "Binary downloaded successfully!" -ForegroundColor Green
Write-Host "==============================" -ForegroundColor Green
Write-Host ""
Write-Host "Binary location: binaries\windows\cloudflared-windows-amd64.exe"
Write-Host ""
Write-Host "You can now run: wails dev"
