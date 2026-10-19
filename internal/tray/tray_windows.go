//go:build windows

package tray

import (
	"runtime"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	shell32  = windows.NewLazySystemDLL("shell32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procRegisterClassExW         = user32.NewProc("RegisterClassExW")
	procCreateWindowExW          = user32.NewProc("CreateWindowExW")
	procDefWindowProcW           = user32.NewProc("DefWindowProcW")
	procDestroyWindow            = user32.NewProc("DestroyWindow")
	procGetMessageW              = user32.NewProc("GetMessageW")
	procTranslateMessage         = user32.NewProc("TranslateMessage")
	procDispatchMessageW         = user32.NewProc("DispatchMessageW")
	procPostMessageW             = user32.NewProc("PostMessageW")
	procPostQuitMessage          = user32.NewProc("PostQuitMessage")
	procRegisterWindowMessageW   = user32.NewProc("RegisterWindowMessageW")
	procCreatePopupMenu          = user32.NewProc("CreatePopupMenu")
	procAppendMenuW              = user32.NewProc("AppendMenuW")
	procTrackPopupMenu           = user32.NewProc("TrackPopupMenu")
	procDestroyMenu              = user32.NewProc("DestroyMenu")
	procGetCursorPos             = user32.NewProc("GetCursorPos")
	procSetForegroundWindow      = user32.NewProc("SetForegroundWindow")
	procCreateIconFromResourceEx = user32.NewProc("CreateIconFromResourceEx")
	procLoadIconW                = user32.NewProc("LoadIconW")
	procDestroyIcon              = user32.NewProc("DestroyIcon")
	procGetSystemMetrics         = user32.NewProc("GetSystemMetrics")
	procShellNotifyIconW         = shell32.NewProc("Shell_NotifyIconW")
	procGetModuleHandleW         = kernel32.NewProc("GetModuleHandleW")
)

const (
	wmDestroy     = 0x0002
	wmClose       = 0x0010
	wmNull        = 0x0000
	wmLButtonUp   = 0x0202
	wmRButtonUp   = 0x0205
	wmContextMenu = 0x007B
	wmTrayIcon    = 0x8001 // WM_APP + 1

	nimAdd    = 0
	nimDelete = 2

	nifMessage = 0x1
	nifIcon    = 0x2
	nifTip     = 0x4

	mfString    = 0x0
	mfSeparator = 0x800

	tpmReturnCmd   = 0x0100
	tpmNoNotify    = 0x0080
	tpmRightButton = 0x0002

	hwndMessage    = ^uintptr(2) // (HWND)-3: message-only window
	idiApplication = 32512
	smCxSmIcon     = 49
)

type wndClassEx struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   windows.Handle
	Icon       windows.Handle
	Cursor     windows.Handle
	Background windows.Handle
	MenuName   *uint16
	ClassName  *uint16
	IconSm     windows.Handle
}

type notifyIconData struct {
	Size            uint32
	Wnd             uintptr
	ID              uint32
	Flags           uint32
	CallbackMessage uint32
	Icon            uintptr
	Tip             [128]uint16
	State           uint32
	StateMask       uint32
	Info            [256]uint16
	Version         uint32
	InfoTitle       [64]uint16
	InfoFlags       uint32
	GuidItem        windows.GUID
	BalloonIcon     uintptr
}

type msg struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

var (
	mu             sync.Mutex
	hwnd           uintptr
	nid            notifyIconData
	onClick        func()
	taskbarCreated uintptr
)

// Start adds the notification-area icon. It runs its own message loop on a
// dedicated OS thread, independent of the Wails UI thread.
func Start(opts Options) {
	items = opts.Items
	onClick = opts.OnClick
	ready := make(chan struct{})
	go loop(opts, ready)
	<-ready
}

// Stop removes the icon and ends the tray message loop.
func Stop() {
	mu.Lock()
	h := hwnd
	mu.Unlock()
	if h != 0 {
		procPostMessageW.Call(h, wmClose, 0, 0)
	}
}

func loop(opts Options, ready chan struct{}) {
	runtime.LockOSThread() // window messages are delivered per thread

	inst, _, _ := procGetModuleHandleW.Call(0)
	className, _ := windows.UTF16PtrFromString("SideQuestTray")
	wc := wndClassEx{
		WndProc:   windows.NewCallback(wndProc),
		Instance:  windows.Handle(inst),
		ClassName: className,
	}
	wc.Size = uint32(unsafe.Sizeof(wc))
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	h, _, _ := procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(className)), 0, 0,
		0, 0, 0, 0, hwndMessage, 0, inst, 0)
	if h == 0 {
		close(ready)
		return
	}
	tbc, _ := windows.UTF16PtrFromString("TaskbarCreated")
	taskbarCreated, _, _ = procRegisterWindowMessageW.Call(uintptr(unsafe.Pointer(tbc)))

	nid = notifyIconData{
		Wnd:             h,
		ID:              1,
		Flags:           nifMessage | nifIcon | nifTip,
		CallbackMessage: wmTrayIcon,
		Icon:            loadIcon(opts.Icon),
	}
	nid.Size = uint32(unsafe.Sizeof(nid))
	tip, _ := windows.UTF16FromString(opts.Tooltip)
	copy(nid.Tip[:len(nid.Tip)-1], tip)

	mu.Lock()
	hwnd = h
	mu.Unlock()
	addIcon()
	close(ready)

	var m msg
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}

func addIcon() {
	procShellNotifyIconW.Call(nimAdd, uintptr(unsafe.Pointer(&nid)))
}

func wndProc(h, message, wParam, lParam uintptr) uintptr {
	switch {
	case message == wmTrayIcon:
		switch lParam & 0xffff {
		case wmLButtonUp:
			if onClick != nil {
				go onClick()
			}
		case wmRButtonUp, wmContextMenu:
			showMenu(h)
		}
		return 0
	case taskbarCreated != 0 && message == taskbarCreated:
		addIcon() // Explorer restarted; put the icon back
		return 0
	case message == wmDestroy:
		procShellNotifyIconW.Call(nimDelete, uintptr(unsafe.Pointer(&nid)))
		if nid.Icon != 0 {
			procDestroyIcon.Call(nid.Icon)
		}
		mu.Lock()
		hwnd = 0
		mu.Unlock()
		procPostQuitMessage.Call(0)
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(h, message, wParam, lParam)
	return r
}

func showMenu(h uintptr) {
	menu, _, _ := procCreatePopupMenu.Call()
	if menu == 0 {
		return
	}
	defer procDestroyMenu.Call(menu)
	for i, it := range items {
		if it.Label == "" {
			procAppendMenuW.Call(menu, mfSeparator, 0, 0)
			continue
		}
		label, _ := windows.UTF16PtrFromString(it.Label)
		// Command ids start at 1; 0 means "menu dismissed".
		procAppendMenuW.Call(menu, mfString, uintptr(i+1), uintptr(unsafe.Pointer(label)))
	}

	var pt struct{ X, Y int32 }
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	// Required so the menu closes when the user clicks elsewhere.
	procSetForegroundWindow.Call(h)
	cmd, _, _ := procTrackPopupMenu.Call(menu, tpmReturnCmd|tpmNoNotify|tpmRightButton,
		uintptr(pt.X), uintptr(pt.Y), 0, h, 0)
	procPostMessageW.Call(h, wmNull, 0, 0)
	if cmd > 0 {
		run(int(cmd) - 1)
	}
}

// loadIcon picks the small-icon-sized image from an .ico file, falling back
// to the stock application icon.
func loadIcon(ico []byte) uintptr {
	want, _, _ := procGetSystemMetrics.Call(smCxSmIcon)
	if data := pickIcon(ico, int(want)); data != nil {
		h, _, _ := procCreateIconFromResourceEx.Call(uintptr(unsafe.Pointer(&data[0])), uintptr(len(data)),
			1, 0x00030000, want, want, 0)
		if h != 0 {
			return h
		}
	}
	h, _, _ := procLoadIconW.Call(0, idiApplication)
	return h
}
