package main

import (
	"embed"
	"log"

	"github.com/votanchat/cloudflared-desktop-tunnel-v3/logger"
	"github.com/votanchat/cloudflared-desktop-tunnel-v3/services"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Initialize file logging (only in build mode)
	if err := logger.InitFileLogging(); err != nil {
		// Continue with console logging only in dev mode
		logger.AppLogger.Debug("File logging disabled (dev mode)")
	}

	logger.AppLogger.Info("═══════════════════════════════════════════════════════")
	logger.AppLogger.Info("🚀 Cloudflared Desktop Tunnel - Starting Application")
	logger.AppLogger.Info("═══════════════════════════════════════════════════════")

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
			ApplicationShouldTerminateAfterLastWindowClosed: false, // Don't terminate when window closes
		},
	})

	// Create main window
	mainWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "Cloudflared Desktop Tunnel",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
	})

	// Handle window close - hide instead of close
	mainWindow.OnWindowEvent(events.Windows.WindowClosing, func(event *application.WindowEvent) {
		logger.AppLogger.Info("Window close requested, hiding to system tray...")
		event.Cancel()
		mainWindow.Hide()
	})

	// Create system tray
	systemTray := app.SystemTray.New()
	systemTray.SetTooltip("Cloudflared Desktop Tunnel")

	// Create system tray menu
	menu := application.NewMenu()
	showMenuItem := menu.Add("Hiển thị")
	showMenuItem.OnClick(func(ctx *application.Context) {
		logger.AppLogger.Info("Showing window from system tray...")
		mainWindow.Show()
		mainWindow.Focus()
	})
	menu.AddSeparator()
	quitMenuItem := menu.Add("Thoát")
	quitMenuItem.OnClick(func(ctx *application.Context) {
		logger.AppLogger.Info("Quit requested from system tray...")
		appService.Shutdown()
		logger.CloseFileLogging()
		app.Quit()
	})

	systemTray.SetMenu(menu)
	systemTray.Show() // Show system tray icon

	// Handle system tray click (toggle window visibility)
	systemTray.OnClick(func() {
		if mainWindow.IsVisible() {
			mainWindow.Hide()
		} else {
			mainWindow.Show()
			mainWindow.Focus()
		}
	})

	// Handle shutdown
	defer func() {
		logger.AppLogger.Info("═══════════════════════════════════════════════════════")
		logger.AppLogger.Info("🛑 Application shutting down...")
		logger.AppLogger.Info("═══════════════════════════════════════════════════════")
		appService.Shutdown()
		logger.CloseFileLogging()
	}()

	// Run the application
	err := app.Run()
	if err != nil {
		logger.AppLogger.Error("Error running application: %v", err)
		log.Fatal(err)
	}
}
