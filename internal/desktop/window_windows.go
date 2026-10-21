//go:build windows

package desktop

import (
	"context"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

// On Windows the window never resizes when the panel opens. Resizing a
// window that grows up or left moves its top-left corner, and WebView2 keeps
// painting its old frame at that corner until it catches up, so the capsule
// would visibly jump. Instead the window always has the panel's size and a
// window region trims it down to the capsule. Areas outside the region are
// neither drawn nor clickable, so the desktop underneath stays usable.
// Expanding just removes the region.
//
// Transparency: on Windows 11 22H2 and later the window is truly
// transparent (WS_EX_NOREDIRECTIONBITMAP with no system backdrop), so the
// gutter around the capsule and panel shows the desktop and their shadows.
// Older versions cannot do that: Wails falls back to a blur-behind, and
// without it the gutter is painted dark. There the window region follows
// the exact rounded shapes instead, giving up the shadow but never showing
// a box.
var exactShapes = func() bool {
	v := windows.RtlGetVersion()
	return v.MajorVersion < 10 || (v.MajorVersion == 10 && v.BuildNumber < 22621)
}()

// Wails' own position APIs mix work-area-relative and absolute coordinates,
// so everything here talks to Win32 directly, in physical pixels.

var (
	user32                = windows.NewLazySystemDLL("user32.dll")
	gdi32                 = windows.NewLazySystemDLL("gdi32.dll")
	shcore                = windows.NewLazySystemDLL("shcore.dll")
	procFindWindowW       = user32.NewProc("FindWindowW")
	procGetWindowRect     = user32.NewProc("GetWindowRect")
	procSetWindowPos      = user32.NewProc("SetWindowPos")
	procSetWindowRgn      = user32.NewProc("SetWindowRgn")
	procMonitorFromRect   = user32.NewProc("MonitorFromRect")
	procGetMonitorInfoW   = user32.NewProc("GetMonitorInfoW")
	procGetDpiForWindow   = user32.NewProc("GetDpiForWindow")
	procGetWindowLongPtrW = user32.NewProc("GetWindowLongPtrW")
	procSetWindowLongPtrW = user32.NewProc("SetWindowLongPtrW")
	procShowWindow        = user32.NewProc("ShowWindow")
	procCreateRectRgn     = gdi32.NewProc("CreateRectRgn")
	procCreateRoundRectRg = gdi32.NewProc("CreateRoundRectRgn")
	procGetDpiForMonitor  = shcore.NewProc("GetDpiForMonitor")
)

const (
	gwlExStyle            = -20
	wsExToolWindow        = 0x00000080
	wsExAppWindow         = 0x00040000
	swpNoZOrder           = 0x0004
	swpNoActivate         = 0x0010
	swpFrameChanged       = 0x0020
	monitorDefaultNearest = 0x00000002
	mdtEffectiveDpi       = 0
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
	sizes     Sizes

	mu       sync.Mutex
	hwnd     uintptr
	anchor   Anchor
	expanded bool
	placed   bool
}

// New returns the Win32 window controller. className must match the
// WindowClassName passed to Wails.
func New(className string, sizes Sizes) Window {
	return &nativeWindow{className: className, sizes: sizes}
}

func (n *nativeWindow) handle() uintptr {
	if n.hwnd == 0 {
		cls, _ := windows.UTF16PtrFromString(n.className)
		n.hwnd, _, _ = procFindWindowW.Call(uintptr(unsafe.Pointer(cls)), 0)
	}
	return n.hwnd
}

func (n *nativeWindow) rect() Rect {
	var r winRect
	procGetWindowRect.Call(n.hwnd, uintptr(unsafe.Pointer(&r)))
	return Rect{int(r.Left), int(r.Top), int(r.Right), int(r.Bottom)}
}

// monitor returns the work area and DPI of the monitor nearest to r.
func monitor(r Rect) (work Rect, dpi int) {
	wr := winRect{int32(r.Left), int32(r.Top), int32(r.Right), int32(r.Bottom)}
	mon, _, _ := procMonitorFromRect.Call(uintptr(unsafe.Pointer(&wr)), monitorDefaultNearest)
	mi := monitorInfo{CbSize: uint32(unsafe.Sizeof(monitorInfo{}))}
	procGetMonitorInfoW.Call(mon, uintptr(unsafe.Pointer(&mi)))
	w := mi.RcWork
	work = Rect{int(w.Left), int(w.Top), int(w.Right), int(w.Bottom)}

	dpi = 96
	if procGetDpiForMonitor.Find() == nil { // Windows 8.1+
		var x, y uint32
		if hr, _, _ := procGetDpiForMonitor.Call(mon, mdtEffectiveDpi,
			uintptr(unsafe.Pointer(&x)), uintptr(unsafe.Pointer(&y))); hr == 0 && x > 0 {
			dpi = int(x)
		}
	}
	return work, dpi
}

func (n *nativeWindow) windowDPI() int {
	if procGetDpiForWindow.Find() == nil {
		if d, _, _ := procGetDpiForWindow.Call(n.hwnd); d > 0 {
			return int(d)
		}
	}
	return 96
}

// physical returns the capsule and panel sizes in pixels at dpi.
func (n *nativeWindow) physical(dpi int) (cw, ch, pw, ph int) {
	s := n.sizes
	return Scale(s.CapsuleW, dpi), Scale(s.CapsuleH, dpi), Scale(s.PanelW, dpi), Scale(s.PanelH, dpi)
}

// layout positions the panel-sized window so the capsule rect c sits in
// corner a, and trims the window to the capsule (or hides it entirely).
func (n *nativeWindow) layout(c Rect, a Anchor, pw, ph int, visible bool) {
	win := PanelRect(c, pw, ph, a)
	procSetWindowPos.Call(n.hwnd, hwndTopMost,
		uintptr(win.Left), uintptr(win.Top), uintptr(win.W()), uintptr(win.H()),
		swpNoActivate|swpFrameChanged)
	n.anchor = a
	if visible {
		n.trimToCapsule()
	} else {
		n.setRegion(Rect{}) // empty region: nothing drawn until Expand
	}
}

func (n *nativeWindow) trimToCapsule() {
	win := n.rect()
	dpi := n.windowDPI()
	cw, ch, _, _ := n.physical(dpi)
	c := Relative(CapsuleRect(win, cw, ch, n.anchor), win)
	if exactShapes {
		n.setRoundRegion(Inset(c, Scale(n.sizes.Gutter, dpi)), Scale(n.sizes.CapsuleRadius, dpi))
		return
	}
	n.setRegion(c)
}

// trimToPanel shows the whole panel: the full window where it is truly
// transparent, or the panel's rounded shape where it is not.
func (n *nativeWindow) trimToPanel() {
	if !exactShapes {
		procSetWindowRgn.Call(n.hwnd, 0, 1) // no region: the whole window
		return
	}
	win := n.rect()
	dpi := n.windowDPI()
	full := Rect{0, 0, win.W(), win.H()}
	n.setRoundRegion(Inset(full, Scale(n.sizes.Gutter, dpi)), Scale(n.sizes.PanelRadius, dpi))
}

// setRegion trims the window to r; the system owns the region afterwards.
func (n *nativeWindow) setRegion(r Rect) {
	rgn, _, _ := procCreateRectRgn.Call(uintptr(r.Left), uintptr(r.Top), uintptr(r.Right), uintptr(r.Bottom))
	procSetWindowRgn.Call(n.hwnd, rgn, 1)
}

// setRoundRegion trims the window to r with rounded corners of radius.
func (n *nativeWindow) setRoundRegion(r Rect, radius int) {
	// The right and bottom edges are exclusive, and the ellipse takes a
	// diameter.
	rgn, _, _ := procCreateRoundRectRg.Call(uintptr(r.Left), uintptr(r.Top), uintptr(r.Right+1), uintptr(r.Bottom+1),
		uintptr(2*radius), uintptr(2*radius))
	procSetWindowRgn.Call(n.hwnd, rgn, 1)
}

func (n *nativeWindow) Place(_ context.Context, saved *Point) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.handle() == 0 {
		return
	}

	// A floating toy should not crowd the taskbar or Alt+Tab.
	ex, _, _ := procGetWindowLongPtrW.Call(n.hwnd, negIndex(gwlExStyle))
	ex = (ex | wsExToolWindow) &^ wsExAppWindow
	procSetWindowLongPtrW.Call(n.hwnd, negIndex(gwlExStyle), ex)

	var c Rect
	if saved != nil {
		c = Rect{saved.X, saved.Y, saved.X + 1, saved.Y + 1}
	} else {
		c = n.rect() // where Wails put the window (centred on the primary monitor)
	}
	work, dpi := monitor(c)
	cw, ch, pw, ph := n.physical(dpi)
	if saved != nil {
		c = Clamp(Rect{saved.X, saved.Y, saved.X + cw, saved.Y + ch}, work)
	} else {
		c = DefaultSpot(cw, ch, work, Scale(16, dpi))
	}
	n.expanded = false
	n.layout(c, AnchorFor(c, work), pw, ph, true)
	n.placed = true
}

func (n *nativeWindow) Anchor(_ context.Context) Anchor {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.handle() == 0 || n.expanded || !n.placed {
		return n.anchor
	}
	// The capsule may have been dragged to another part of the screen, or
	// onto a monitor with a different DPI. If so, re-lay the window out
	// around it. The capsule stays hidden until Expand, which gives the UI
	// time to pin it to the new corner.
	win := n.rect()
	dpi := n.windowDPI()
	cw, ch, pw, ph := n.physical(dpi)
	c := CapsuleRect(win, cw, ch, n.anchor)
	work, _ := monitor(c)
	a := AnchorFor(c, work)
	if a != n.anchor || win.W() != pw || win.H() != ph {
		n.layout(c, a, pw, ph, false)
	}
	return n.anchor
}

func (n *nativeWindow) Expand(context.Context) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.handle() == 0 {
		return
	}
	n.expanded = true
	n.trimToPanel()
}

func (n *nativeWindow) Collapse(context.Context) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.handle() == 0 {
		return
	}
	n.expanded = false
	n.trimToCapsule()
}

func (n *nativeWindow) CapsuleSpot(context.Context) (Point, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.handle() == 0 {
		return Point{}, false
	}
	cw, ch, _, _ := n.physical(n.windowDPI())
	c := CapsuleRect(n.rect(), cw, ch, n.anchor)
	return Point{X: c.Left, Y: c.Top}, true
}

func (n *nativeWindow) Show(context.Context) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.handle() != 0 {
		procShowWindow.Call(n.hwnd, swShowNoActivate)
	}
}

// negIndex converts a negative int index into the uintptr Win32 expects.
func negIndex(i int) uintptr { return uintptr(int64(i)) }
