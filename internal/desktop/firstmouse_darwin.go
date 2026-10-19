//go:build darwin

package desktop

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Cocoa -framework WebKit
void sqAcceptFirstMouse(void);
*/
import "C"

// acceptFirstMouse lets the very first click reach the page even when
// SideQuest is not the active app (always the case for a menu bar app).
// Without it macOS spends that click on activating the window.
func acceptFirstMouse() { C.sqAcceptFirstMouse() }
