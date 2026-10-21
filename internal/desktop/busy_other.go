//go:build !windows

package desktop

// Busy is not detected on this platform yet, so Boo never stays quiet.
func Busy() bool { return false }
