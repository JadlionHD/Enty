package main

import (
	"context"
	"embed"

	"github.com/JadlionHD/Enty/internal/utils"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()
	utils := utils.Utils()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "Enty",
		Width:  800,
		Height: 600,

		MinWidth:  800,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Frameless:        true,
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			utils.Start(ctx)
		},
		Bind: []interface{}{
			app,
			utils,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
