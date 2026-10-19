//go:build windows

package desktop

import (
	"context"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Wails' own WindowSetPosition is relative to the monitor work area while
// WindowGetPosition is absolute, and sizes are DPI-scaled inconsistently, so
// on Windows we drive the native window directly with Win32.

var (
	user32                = windows.NewLazySystemDLL("user32.dll")
	procFindWindowW       = user32.NewProc("FindWindowW")
	procGetWindowRect     = user32.NewProc("GetWindowRect")
	procSetWindowPos      = user32.NewProc("SetWindowPos")
	procMonitorFromWindow = user32.NewProc("MonitorFromWindow")
	procGetMonitorInfoW   = user32.NewProc("GetMonitorInfoW")
	procGetDpiForWindow   = user32.NewProc("GetDpiForWindow")
	procGetWindowLongPtrW = user32.NewProc("GetWindowLongPtrW")
	procSetWindowLongPtrW = user32.NewProc("SetWindowLongPtrW")
	procShowWindow        = user32.NewProc("ShowWindow")
)

const (
	gwlExStyle            = -20
	wsExToolWindow        = 0x00000080
	wsExAppWindow         = 0x00040000
	swpNoSize             = 0x0001
	swpNoZOrder           = 0x0004
	swpNoActivate         = 0x0010
	swpFrameChanged       = 0x0020
	monitorDefaultNearest = 0x00000002
	hwndTopMost           = ^uintptr(0) // (HWND)-1
	swShowNoActivate      = 4
)

type winRect struct{ Left, Top, Right, Bottom int32 }

type monitorInfo struct {
	CbSize    uint32
	RcMonitor winRect
	RcWork    winRect
	DwFlags   uint32
}

type nativeWindow struct {
	className string
	mu        sync.Mutex
	hwnd      uintptr
}

// New returns the Win32 window controller. className must match the
// WindowClassName passed to Wails.
func New(className string) Window {
	return &nativeWindow{className: className}
}

func (n *nativeWindow) handle() uintptr {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.hwnd == 0 {
		cls, _ := windows.UTF16PtrFromString(n.className)
		h, _, _ := procFindWindowW.Call(uintptr(unsafe.Pointer(cls)), 0)
		n.hwnd = h
	}
	return n.hwnd
}

func (n *nativeWindow) rect(h uintptr) Rect {
	var r winRect
	procGetWindowRect.Call(h, uintptr(unsafe.Pointer(&r)))
	return Rect{int(r.Left), int(r.Top), int(r.Right), int(r.Bottom)}
}

func (n *nativeWindow) work(h uintptr) Rect {
	mon, _, _ := procMonitorFromWindow.Call(h, monitorDefaultNearest)
	mi := monitorInfo{CbSize: uint32(unsafe.Sizeof(monitorInfo{}))}
	procGetMonitorInfoW.Call(mon, uintptr(unsafe.Pointer(&mi)))
	w := mi.RcWork
	return Rect{int(w.Left), int(w.Top), int(w.Right), int(w.Bottom)}
}

func (n *nativeWindow) dpi(h uintptr) int {
	if procGetDpiForWindow.Find() != nil { // pre Windows 10 1607
		return 96
	}
	d, _, _ := procGetDpiForWindow.Call(h)
	if d == 0 {
		return 96
	}
	return int(d)
}

func (n *nativeWindow) move(h uintptr, r Rect) {
	procSetWindowPos.Call(h, 0,
		uintptr(r.Left), uintptr(r.Top), uintptr(r.W()), uintptr(r.H()),
		swpNoZOrder|swpNoActivate)
}

func (n *nativeWindow) Place(_ context.Context, saved *Point, w, h int) {
	hw := n.handle()
	if hw == 0 {
		return
	}

	// A floating toy should not crowd the taskbar or Alt+Tab.
	ex, _, _ := procGetWindowLongPtrW.Call(hw, negIndex(gwlExStyle))
	ex = (ex | wsExToolWindow) &^ wsExAppWindow
	procSetWindowLongPtrW.Call(hw, negIndex(gwlExStyle), ex)

	if saved != nil {
		// Move first so DPI and work area come from the saved monitor.
		procSetWindowPos.Call(hw, 0, uintptr(saved.X), uintptr(saved.Y), 0, 0,
			swpNoZOrder|swpNoActivate|swpNoSize)
	}
	dpi := n.dpi(hw)
	work := n.work(hw)
	pw, ph := Scale(w, dpi), Scale(h, dpi)

	var r Rect
	if saved != nil {
		r = Clamp(Rect{saved.X, saved.Y, saved.X + pw, saved.Y + ph}, work)
	} else {
		r = DefaultSpot(pw, ph, work, Scale(16, dpi))
	}
	procSetWindowPos.Call(hw, hwndTopMost,
		uintptr(r.Left), uintptr(r.Top), uintptr(r.W()), uintptr(r.H()),
		swpNoActivate|swpFrameChanged)
}

func (n *nativeWindow) Resize(_ context.Context, w, h int) {
	hw := n.handle()
	if hw == 0 {
		return
	}
	dpi := n.dpi(hw)
	n.move(hw, Resize(n.rect(hw), Scale(w, dpi), Scale(h, dpi), n.work(hw)))
}

func (n *nativeWindow) Show(_ context.Context) {
	if hw := n.handle(); hw != 0 {
		procShowWindow.Call(hw, swShowNoActivate)
	}
}

func (n *nativeWindow) Spot(_ context.Context, w, h int) (Point, bool) {
	hw := n.handle()
	if hw == 0 {
		return Point{}, false
	}
	dpi := n.dpi(hw)
	r := Resize(n.rect(hw), Scale(w, dpi), Scale(h, dpi), n.work(hw))
	return Point{X: r.Left, Y: r.Top}, true
}

// negIndex converts a negative int index into the uintptr Win32 expects.
func negIndex(i int) uintptr { return uintptr(int64(i)) }
