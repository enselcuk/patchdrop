package main

import (
	"embed"
	"os"
	"strconv"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"

	"localshare/internal/node"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	opts := node.Options{Profile: os.Getenv("LOCALSHARE_PROFILE")}
	if p, err := strconv.Atoi(os.Getenv("LOCALSHARE_PORT")); err == nil {
		opts.Port = p
	}

	if len(os.Args) > 1 && isCLICommand(os.Args[1]) {
		os.Exit(runCLI(opts, os.Args[1:]))
	}

	app := NewApp(opts)
	title := "LocalShare"
	if opts.Profile != "" {
		title += " (" + opts.Profile + ")"
	}
	err := wails.Run(&options.App{
		Title:     title,
		Width:     980,
		Height:    660,
		MinWidth:  760,
		MinHeight: 500,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 246, G: 247, B: 249, A: 1},
		OnStartup:        app.startup,
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: true,
		},
		Mac: &mac.Options{
			About: &mac.AboutInfo{Title: "LocalShare", Message: "Fast file sharing over your local network."},
		},
		Bind: []interface{}{app},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
