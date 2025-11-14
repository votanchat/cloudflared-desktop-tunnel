# Contributing to Cloudflared Desktop Tunnel

Thank you for your interest in contributing to Cloudflared Desktop Tunnel! This document provides guidelines and instructions for contributing.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/cloudflared-desktop-tunnel.git
   cd cloudflared-desktop-tunnel
   ```
3. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Node.js 18+ and npm
- Wails CLI v2
- Platform-specific dependencies (see README.md)

### Install Dependencies

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

Before running the app, download the official cloudflared binaries from [Cloudflare's releases](https://github.com/cloudflare/cloudflared/releases) and place them in the `binaries/` directory according to the structure in `binaries/README.md`.

### Running in Development Mode

```bash
wails dev
```

This will start the app in development mode with hot-reload enabled for both frontend and backend.

## Code Style

### Go Code

- Follow the [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Run `go fmt` before committing
- Run `go vet` to check for common mistakes
- Add comments for exported functions and types

### TypeScript/React Code

- Use TypeScript for type safety
- Follow React best practices and hooks patterns
- Use functional components
- Add proper TypeScript types for props and state

## Testing

### Backend Tests

```bash
go test ./...
```

### Frontend Tests

```bash
cd frontend
npm test
```

## Building

### Build for Current Platform

```bash
wails build
```

### Cross-Platform Builds

```bash
# Windows
wails build -platform windows/amd64

# macOS Intel
wails build -platform darwin/amd64

# macOS ARM
wails build -platform darwin/arm64

# Linux
wails build -platform linux/amd64
```

## Pull Request Process

1. **Update documentation** if you're changing functionality
2. **Add tests** for new features
3. **Ensure all tests pass** before submitting
4. **Update CHANGELOG.md** with your changes
5. **Create a pull request** with a clear title and description
6. **Link any related issues** in your PR description

### PR Title Format

Use conventional commits format:

- `feat: Add new feature`
- `fix: Fix bug in tunnel manager`
- `docs: Update README`
- `style: Format code`
- `refactor: Refactor backend client`
- `test: Add tests for config`
- `chore: Update dependencies`

## Reporting Bugs

When reporting bugs, please include:

1. **Description** of the issue
2. **Steps to reproduce**
3. **Expected behavior**
4. **Actual behavior**
5. **System information** (OS, Go version, Wails version)
6. **Logs** if applicable

## Feature Requests

We welcome feature requests! Please:

1. **Check existing issues** to avoid duplicates
2. **Clearly describe** the feature and its use case
3. **Explain why** it would be beneficial
4. **Provide examples** if possible

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers
- Accept constructive criticism
- Focus on what's best for the community

## Questions?

If you have questions, feel free to:

- Open an issue on GitHub
- Join our discussions
- Contact the maintainers

Thank you for contributing! 🚀
