#!/bin/bash

# Script to download cloudflared binaries for all platforms
# This downloads the latest release from Cloudflare's official GitHub repository

set -e

COLOR_RESET="\033[0m"
COLOR_GREEN="\033[0;32m"
COLOR_BLUE="\033[0;34m"
COLOR_YELLOW="\033[1;33m"

echo -e "${COLOR_BLUE}Cloudflared Binary Downloader${COLOR_RESET}"
echo -e "${COLOR_BLUE}==============================${COLOR_RESET}"
echo ""

# Get the latest version
echo -e "${COLOR_YELLOW}Fetching latest cloudflared version...${COLOR_RESET}"
LATEST_VERSION=$(curl -s https://api.github.com/repos/cloudflare/cloudflared/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
echo -e "${COLOR_GREEN}Latest version: ${LATEST_VERSION}${COLOR_RESET}"
echo ""

# Base URL for downloads
BASE_URL="https://github.com/cloudflare/cloudflared/releases/download/${LATEST_VERSION}"

# Create directories
mkdir -p binaries/windows
mkdir -p binaries/darwin
mkdir -p binaries/linux

# Download Windows binary
echo -e "${COLOR_YELLOW}Downloading Windows binary...${COLOR_RESET}"
wget -q --show-progress "${BASE_URL}/cloudflared-windows-amd64.exe" -O binaries/windows/cloudflared-windows-amd64.exe
echo -e "${COLOR_GREEN}✓ Windows binary downloaded${COLOR_RESET}"
echo ""

# Download macOS binaries
echo -e "${COLOR_YELLOW}Downloading macOS Intel binary...${COLOR_RESET}"
wget -q --show-progress "${BASE_URL}/cloudflared-darwin-amd64.tgz" -O /tmp/cloudflared-darwin-amd64.tgz
tar -xzf /tmp/cloudflared-darwin-amd64.tgz -C binaries/darwin/
mv binaries/darwin/cloudflared binaries/darwin/cloudflared-darwin-amd64
rm /tmp/cloudflared-darwin-amd64.tgz
chmod +x binaries/darwin/cloudflared-darwin-amd64
echo -e "${COLOR_GREEN}✓ macOS Intel binary downloaded${COLOR_RESET}"
echo ""

echo -e "${COLOR_YELLOW}Downloading macOS ARM binary...${COLOR_RESET}"
wget -q --show-progress "${BASE_URL}/cloudflared-darwin-arm64.tgz" -O /tmp/cloudflared-darwin-arm64.tgz
tar -xzf /tmp/cloudflared-darwin-arm64.tgz -C binaries/darwin/
mv binaries/darwin/cloudflared binaries/darwin/cloudflared-darwin-arm64
rm /tmp/cloudflared-darwin-arm64.tgz
chmod +x binaries/darwin/cloudflared-darwin-arm64
echo -e "${COLOR_GREEN}✓ macOS ARM binary downloaded${COLOR_RESET}"
echo ""

# Download Linux binaries
echo -e "${COLOR_YELLOW}Downloading Linux AMD64 binary...${COLOR_RESET}"
wget -q --show-progress "${BASE_URL}/cloudflared-linux-amd64" -O binaries/linux/cloudflared-linux-amd64
chmod +x binaries/linux/cloudflared-linux-amd64
echo -e "${COLOR_GREEN}✓ Linux AMD64 binary downloaded${COLOR_RESET}"
echo ""

echo -e "${COLOR_YELLOW}Downloading Linux ARM64 binary...${COLOR_RESET}"
wget -q --show-progress "${BASE_URL}/cloudflared-linux-arm64" -O binaries/linux/cloudflared-linux-arm64
chmod +x binaries/linux/cloudflared-linux-arm64
echo -e "${COLOR_GREEN}✓ Linux ARM64 binary downloaded${COLOR_RESET}"
echo ""

echo -e "${COLOR_GREEN}==============================${COLOR_RESET}"
echo -e "${COLOR_GREEN}All binaries downloaded successfully!${COLOR_RESET}"
echo -e "${COLOR_GREEN}==============================${COLOR_RESET}"
echo ""
echo "Binary locations:"
echo "  Windows:     binaries/windows/cloudflared-windows-amd64.exe"
echo "  macOS Intel: binaries/darwin/cloudflared-darwin-amd64"
echo "  macOS ARM:   binaries/darwin/cloudflared-darwin-arm64"
echo "  Linux AMD64: binaries/linux/cloudflared-linux-amd64"
echo "  Linux ARM64: binaries/linux/cloudflared-linux-arm64"
echo ""
echo "You can now run: wails dev"
