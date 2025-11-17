package main

import (
	"embed"
	"log"

	"github.com/votanchat/cloudflared-desktop-tunnel-v3/services"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Initialize file logging (only in build mode)
	if err := InitFileLogging(); err != nil {
		// Continue with console logging only in dev mode
	}

	appLogger.Info("Application starting up...")

	// Create app service (orchestrates all services)
	appService := services.NewAppService()

	// Create Wails application
	app := application.New(application.Options{
		Name:        "Cloudflared Desktop Tunnel",
		Description: "Cross-platform desktop app for managing Cloudflare Tunnels",
		Services: []application.Service{
			application.NewService(appService),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// Create main window
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "Cloudflared Desktop Tunnel",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
	})

	// Handle shutdown
	defer func() {
		appLogger.Info("Application shutting down...")
		appService.Shutdown()
		CloseFileLogging()
	}()

	// Run the application
	err := app.Run()
	if err != nil {
		appLogger.Error("Error running application: %v", err)
		log.Fatal(err)
	}
}
