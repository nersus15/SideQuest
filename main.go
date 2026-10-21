package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"
	goruntime "runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"SideQuest/internal/desktop"
	"SideQuest/internal/store"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/windows/icon.ico
var trayIcon []byte

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

// appMenu replaces Wails' default macOS menu, which binds Cmd+Q to quit.
// SideQuest only quits from its menu bar item; Edit keeps copy/paste keys.
func appMenu() *menu.Menu {
	if goruntime.GOOS != "darwin" {
		return nil // a menu would add a menu bar to the Windows window
	}
	return menu.NewMenuFromItems(menu.EditMenu())
}

func main() {
	window := desktop.New(windowClass, desktop.Sizes{
		CapsuleW: capsuleW, CapsuleH: capsuleH,
		PanelW: panelW, PanelH: panelH,
		Gutter:        gutter,
		CapsuleRadius: capsuleRadius,
		PanelRadius:   panelRadius,
	})
	app, err := NewApp(dataDir(), window)
	if err != nil {
		log.Fatal(err)
	}

	err = wails.Run(&options.App{
		Title:         "SideQuest",
		Width:         capsuleW,
		Height:        capsuleH,
		Frameless:     true,
		DisableResize: true,
		AlwaysOnTop:   true,
		StartHidden:   true, // shown once positioned, so it never jumps
		// Closing the window (e.g. Alt+F4) only hides it; quitting happens
		// from the tray or menu bar.
		HideWindowOnClose: true,
		Menu:              appMenu(),
		BackgroundColour:  &options.RGBA{R: 0, G: 0, B: 0, A: 0},
		AssetServer:       &assetserver.Options{Assets: assets},
		OnStartup:         app.startup,
		OnDomReady:        app.domReady,
		OnBeforeClose:     app.beforeClose,
		OnShutdown:        app.shutdown,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               "com.kamscode.sidequest",
			OnSecondInstanceLaunch: func(options.SecondInstanceData) { app.open() },
		},
		Bind: []any{app},
		Windows: &windows.Options{
			// Translucent with no backdrop gives a truly transparent window on
			// Windows 11 22H2+ (see internal/desktop/window_windows.go for
			// older versions).
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
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
