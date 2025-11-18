package main

import (
	"embed"
	"log"
	"sync"
	"time"

	"github.com/votanchat/cloudflared-desktop-tunnel-v3/logger"
	"github.com/votanchat/cloudflared-desktop-tunnel-v3/services"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var trayIconBytes []byte

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

	// Store window options for recreation if needed
	windowOptions := application.WebviewWindowOptions{
		Title: "Cloudflared Desktop Tunnel",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
	}

	// Create main window - use pointer to allow reassignment
	var mainWindow *application.WebviewWindow

	// Helper function to create or recreate window
	createWindow := func() *application.WebviewWindow {
		window := app.Window.NewWithOptions(windowOptions)
		// Handle window close - hide instead of close
		window.OnWindowEvent(events.Common.WindowClosing, func(event *application.WindowEvent) {
			logger.AppLogger.Info("Window close requested, hiding to system tray...")
			// Cancel the close event immediately to prevent window destruction
			event.Cancel()
			// Hide the window instead of closing
			application.InvokeAsync(func() {
				defer func() {
					if r := recover(); r != nil {
						logger.AppLogger.Error("Error hiding window: %v", r)
					}
				}()
				window.Hide()
			})
		})
		return window
	}

	// Create initial window
	mainWindow = createWindow()

	// Create system tray
	systemTray := app.SystemTray.New()
	systemTray.SetTooltip("Cloudflared Desktop Tunnel")

	// Set system tray icon (use template icon for macOS for better appearance)
	if len(trayIconBytes) > 0 {
		systemTray.SetTemplateIcon(trayIconBytes)
		logger.AppLogger.Info("✅ System tray icon set")
	} else {
		logger.AppLogger.Warn("⚠️  System tray icon not found, using default")
	}

	// Helper function to ensure window exists and create if needed
	ensureWindow := func() *application.WebviewWindow {
		// First check if mainWindow is nil
		if mainWindow == nil {
			logger.AppLogger.Info("Window is nil, creating new window...")
			mainWindow = createWindow()
			// Re-attach to system tray
			systemTray.AttachWindow(mainWindow)
			logger.AppLogger.Info("New window created and attached to system tray")
			return mainWindow
		}

		// Check if window still exists in window manager by ID
		windowID := mainWindow.ID()
		_, exists := app.Window.GetByID(windowID)
		if !exists {
			logger.AppLogger.Info("Window #%d not found in window manager, recreating...", windowID)
			mainWindow = createWindow()
			// Re-attach to system tray
			systemTray.AttachWindow(mainWindow)
			logger.AppLogger.Info("Window recreated and attached to system tray")
			return mainWindow
		}

		// Also check if window is accessible (not destroyed internally)
		var windowAccessible bool
		func() {
			defer func() {
				if r := recover(); r != nil {
					windowAccessible = false
					logger.AppLogger.Info("Window check failed, window may be destroyed: %v", r)
				}
			}()
			// Try to check if window exists and is accessible
			_ = mainWindow.IsVisible()
			windowAccessible = true
		}()

		// If window is not accessible, recreate it
		if !windowAccessible {
			logger.AppLogger.Info("Window is not accessible, recreating...")
			mainWindow = createWindow()
			// Re-attach to system tray
			systemTray.AttachWindow(mainWindow)
			logger.AppLogger.Info("Window recreated and attached to system tray")
		}
		return mainWindow
	}

	// Attach window to system tray - this handles show/hide automatically
	systemTray.AttachWindow(mainWindow)

	// Create system tray menu
	menu := application.NewMenu()
	showMenuItem := menu.Add("Hiển thị")
	showMenuItem.OnClick(func(ctx *application.Context) {
		logger.AppLogger.Info("Showing window from system tray menu...")
		application.InvokeAsync(func() {
			defer func() {
				if r := recover(); r != nil {
					logger.AppLogger.Error("Error showing window: %v", r)
					// Try to recreate window if error occurred
					ensureWindow()
				}
			}()
			window := ensureWindow()
			if window.IsMinimised() {
				window.UnMinimise()
			}
			window.Show()
			// Focus after a small delay
			go func() {
				time.Sleep(100 * time.Millisecond)
				application.InvokeAsync(func() {
					defer func() {
						if r := recover(); r != nil {
							logger.AppLogger.Error("Error focusing window: %v", r)
						}
					}()
					window := ensureWindow()
					if window.IsVisible() {
						window.Focus()
					}
				})
			}()
		})
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

	// Handle dock icon click (macOS) - show window when clicking dock icon
	var dockClickInProgress bool
	var dockClickMutex sync.Mutex
	app.Event.OnApplicationEvent(events.Mac.ApplicationShouldHandleReopen, func(event *application.ApplicationEvent) {
		// Prevent multiple simultaneous calls
		dockClickMutex.Lock()
		if dockClickInProgress {
			dockClickMutex.Unlock()
			logger.AppLogger.Debug("Dock icon click already in progress, skipping...")
			return
		}
		dockClickInProgress = true
		dockClickMutex.Unlock()

		logger.AppLogger.Info("Dock icon clicked, ensuring window exists and showing...")
		application.InvokeAsync(func() {
			defer func() {
				dockClickMutex.Lock()
				dockClickInProgress = false
				dockClickMutex.Unlock()
				if r := recover(); r != nil {
					logger.AppLogger.Error("Error showing window from dock icon: %v", r)
					// Try to recreate window if error occurred
					ensureWindow()
				}
			}()
			// Ensure window exists first (create if needed)
			window := ensureWindow()
			if window == nil {
				logger.AppLogger.Error("Failed to create window")
				return
			}
			// Unminimize if minimized
			if window.IsMinimised() {
				window.UnMinimise()
			}
			// Show window - AttachWindow will handle positioning
			window.Show()
			// Focus after a small delay to ensure window is ready
			go func() {
				time.Sleep(100 * time.Millisecond)
				application.InvokeAsync(func() {
					defer func() {
						if r := recover(); r != nil {
							logger.AppLogger.Error("Error focusing window from dock icon: %v", r)
						}
					}()
					window := ensureWindow()
					if window != nil {
						window.Focus()
					}
				})
			}()
		})
	})

	// Override default click handler to ensure window exists
	systemTray.OnClick(func() {
		application.InvokeAsync(func() {
			defer func() {
				if r := recover(); r != nil {
					logger.AppLogger.Error("Error in system tray click handler: %v", r)
					// Try to recreate window if error occurred
					ensureWindow()
				}
			}()
			window := ensureWindow()
			if window.IsVisible() {
				window.Hide()
			} else {
				if window.IsMinimised() {
					window.UnMinimise()
				}
				window.Show()
				// Focus after a small delay
				go func() {
					time.Sleep(100 * time.Millisecond)
					application.InvokeAsync(func() {
						defer func() {
							if r := recover(); r != nil {
								logger.AppLogger.Error("Error focusing window from tray click: %v", r)
							}
						}()
						window := ensureWindow()
						if window.IsVisible() {
							window.Focus()
						}
					})
				}()
			}
		})
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
