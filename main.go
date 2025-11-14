package main

import (
	"embed"
	"log"

	"github.com/votanchat/cloudflared-desktop-tunnel/app"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	appInstance := app.NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "Cloudflared Desktop Tunnel",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        appInstance.Startup,
		OnDomReady:       appInstance.DomReady,
		OnShutdown:       appInstance.Shutdown,
		Bind: []interface{}{
			appInstance,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarDefault(),
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			About: &mac.AboutInfo{
				Title:   "Cloudflared Desktop Tunnel",
				Message: "Cross-platform desktop app for managing Cloudflare Tunnels",
			},
		},
		Linux: &linux.Options{
			Icon: []byte{}, // You can add your icon bytes here
		},
	})

	if err != nil {
		log.Fatal("Error starting application:", err)
	}
}
