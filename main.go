package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

// buildTime 由 build.ps1 通过 -ldflags -X 在打包时注入（本地时间）；
// wails dev / 直接 go run 时为空，标题保持原样
var buildTime string

// appTitle 窗口标题：打包产物带上编译时间，如 "FFMpeg Tools (build 2026-11-03 22:06:00)"
func appTitle() string {
	title := "FFMpeg Tools"
	if buildTime != "" {
		title += " (build " + buildTime + ")"
	}
	return title
}

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     appTitle(),
		Width:     1280,
		Height:    800,
		MinWidth:  960,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.Startup,
		Bind: []interface{}{
			app,
		},
		// 允许向窗口拖拽文件；禁用 Wails 默认的整窗 drop 高亮，
		// 保留 WebView 原生 drop 事件使 dataTransfer.files[].path 可用
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: false,
			CSSDropProperty:    "--wails-drop-target",
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
