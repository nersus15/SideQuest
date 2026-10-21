// Package tray puts SideQuest in the macOS menu bar or the Windows
// notification area. Wails v2 has no tray support, so each platform has a
// small native implementation.
package tray

import (
	"encoding/binary"
	"sync"
)

// Item is one menu entry. An empty Label adds a separator.
type Item struct {
	Label   string
	OnClick func()
}

// Options configure the tray icon.
type Options struct {
	Tooltip string
	// Icon is a Windows .ico file. macOS draws its own template image.
	Icon  []byte
	Items []Item
	// OnClick runs when the icon itself is clicked (Windows only; macOS
	// always opens the menu, as menu bar apps should).
	OnClick func()
}

var (
	itemsMu sync.Mutex
	items   []Item
)

func setItems(in []Item) {
	itemsMu.Lock()
	items = append([]Item(nil), in...)
	itemsMu.Unlock()
}

// snapshot returns a copy of the menu items.
func snapshot() []Item {
	itemsMu.Lock()
	defer itemsMu.Unlock()
	return append([]Item(nil), items...)
}

// run calls an item's handler off the UI thread, so handlers may freely
// call back into Wails (which dispatches to the UI thread itself).
func run(i int) {
	list := snapshot()
	if i >= 0 && i < len(list) && list[i].OnClick != nil {
		go list[i].OnClick()
	}
}

// SetLabels renames the menu items in order, e.g. after the player
// switches language. Separators (empty labels) stay separators.
func SetLabels(labels []string) {
	itemsMu.Lock()
	for i := range items {
		if i < len(labels) && items[i].Label != "" {
			items[i].Label = labels[i]
		}
	}
	itemsMu.Unlock()
	relabel(snapshot())
}

// pickIcon returns the image data of the smallest entry that is at least
// size pixels wide (or the largest one if none is).
func pickIcon(ico []byte, size int) []byte {
	if len(ico) < 6 || binary.LittleEndian.Uint16(ico[2:]) != 1 {
		return nil
	}
	count := int(binary.LittleEndian.Uint16(ico[4:]))
	var best []byte
	bestW := 0
	for i := 0; i < count; i++ {
		e := 6 + 16*i
		if e+16 > len(ico) {
			return best
		}
		w := int(ico[e])
		if w == 0 {
			w = 256
		}
		n := int(binary.LittleEndian.Uint32(ico[e+8:]))
		off := int(binary.LittleEndian.Uint32(ico[e+12:]))
		if off < 0 || n <= 0 || off+n > len(ico) {
			continue
		}
		better := best == nil ||
			(w >= size && (bestW < size || w < bestW)) ||
			(w < size && bestW < size && w > bestW)
		if better {
			best, bestW = ico[off:off+n], w
		}
	}
	return best
}
