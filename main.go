package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"SideQuest/internal/desktop"
	"SideQuest/internal/store"
)

//go:embed all:frontend/dist
var assets embed.FS

const windowClass = "SideQuestWindow"

func dataDir() string {
	base, err := os.UserConfigDir()
	if err != nil {
		base = os.TempDir()
	}
	return filepath.Join(base, "SideQuest")
}

func newStore(dir string) store.File {
	return store.File{Path: filepath.Join(dir, "state.json")}
}

func main() {
	app, err := NewApp(dataDir(), desktop.New(windowClass))
	if err != nil {
		log.Fatal(err)
	}

	err = wails.Run(&options.App{
		Title:            "SideQuest",
		Width:            capsuleW,
		Height:           capsuleH,
		Frameless:        true,
		DisableResize:    true,
		AlwaysOnTop:      true,
		StartHidden:      true, // shown once positioned, so it never jumps
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},
		AssetServer:      &assetserver.Options{Assets: assets},
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnBeforeClose:    app.beforeClose,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "com.kamscode.sidequest",
			OnSecondInstanceLaunch: func(options.SecondInstanceData) {
				if app.ctx != nil {
					runtime.WindowShow(app.ctx)
					runtime.EventsEmit(app.ctx, "sidequest:poke")
				}
			},
		},
		Bind: []any{app},
		Windows: &windows.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  false,
			BackdropType:         windows.None,
			// No DWM border or shadow around the transparent window area.
			DisableFramelessWindowDecorations: true,
			WindowClassName:                   windowClass,
			Theme:                             windows.Dark,
			DisablePinchZoom:                  true,
		},
		Mac: &mac.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  false,
			TitleBar:             mac.TitleBarHiddenInset(),
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
