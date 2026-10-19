//go:build darwin

package tray

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Cocoa
#include <stdlib.h>
void trayStart(const char *tooltip, const char **labels, int count);
void trayStop(void);
*/
import "C"

import "unsafe"

// Start shows the menu bar item and turns SideQuest into an accessory app
// (no Dock icon, no app menu). Call it once the app has finished launching.
func Start(opts Options) {
	items = opts.Items

	tip := C.CString(opts.Tooltip)
	defer C.free(unsafe.Pointer(tip))

	labels := make([]*C.char, len(items))
	for i, it := range items {
		labels[i] = C.CString(it.Label)
	}
	defer func() {
		for _, l := range labels {
			C.free(unsafe.Pointer(l))
		}
	}()

	var first **C.char
	if len(labels) > 0 {
		first = &labels[0]
	}
	// The Objective-C side copies the strings before returning.
	C.trayStart(tip, first, C.int(len(labels)))
}

// Stop removes the menu bar item.
func Stop() { C.trayStop() }

//export goTrayItem
func goTrayItem(i C.int) { run(int(i)) }
