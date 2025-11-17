package main

// This file is kept for backward compatibility
// New code should use the logger package instead
import "github.com/votanchat/cloudflared-desktop-tunnel-v3/logger"

// Re-export logger functions and variables for backward compatibility
var (
	appLogger     = logger.AppLogger
	tunnelLogger  = logger.TunnelLogger
	backendLogger = logger.BackendLogger
	serverLogger  = logger.ServerLogger
	binaryLogger  = logger.BinaryLogger
)

var (
	InitFileLogging  = logger.InitFileLogging
	CloseFileLogging = logger.CloseFileLogging
	GetLogger        = logger.GetLogger
	SetLevel         = logger.SetLevel
)
