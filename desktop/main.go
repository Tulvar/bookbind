package main

import (
	"embed"
	"os"
	"path/filepath"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

// Release builds enable this with -ldflags "-X main.bundledFFmpeg=1".
var bundledFFmpeg string

func main() {
	ffmpegPath, ffprobePath := "ffmpeg", "ffprobe"
	if bundledFFmpeg == "1" {
		executable, err := os.Executable()
		if err != nil {
			println("Error locating the bookbind desktop executable:", err.Error())
			return
		}
		ffmpegPath, ffprobePath = bundledToolPaths(executable, runtime.GOOS)
	}
	app := NewApp(ffmpegPath, ffprobePath)

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "bookbind",
		Width:  1180,
		Height: 760,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 245, G: 247, B: 250, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func bundledToolPaths(executable, goos string) (ffmpegPath, ffprobePath string) {
	dir := filepath.Dir(executable)
	extension := ""
	switch goos {
	case "darwin":
		dir = filepath.Join(dir, "..", "Resources")
	case "windows":
		extension = ".exe"
	}
	dir = filepath.Join(dir, "bin")
	return filepath.Join(dir, "ffmpeg"+extension), filepath.Join(dir, "ffprobe"+extension)
}
