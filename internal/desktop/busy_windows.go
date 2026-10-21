//go:build windows

package desktop

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var procSHQueryUserNotificationState = windows.NewLazySystemDLL("shell32.dll").NewProc("SHQueryUserNotificationState")

// Busy reports whether a fullscreen app, game or presentation is in front.
func Busy() bool {
	if procSHQueryUserNotificationState.Find() != nil {
		return false
	}
	var state int32
	hr, _, _ := procSHQueryUserNotificationState.Call(uintptr(unsafe.Pointer(&state)))
	if hr != 0 { // not S_OK
		return false
	}
	return busyState(int(state))
}
