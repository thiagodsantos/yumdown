package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()
	if err := wails.Run(&options.App{
		Title:            "Yumdown",
		Width:            1120,
		Height:           780,
		MinWidth:         860,
		MinHeight:        640,
		BackgroundColour: &options.RGBA{R: 12, G: 16, B: 24, A: 255},
		AssetServer:      &assetserver.Options{Assets: assets},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind:             []interface{}{app},
	}); err != nil {
		println("Error:", err.Error())
	}
}
