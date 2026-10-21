//go:build !darwin && !windows

package tray

// Start is a no-op on platforms without a tray implementation.
func Start(opts Options) { setItems(opts.Items) }

func relabel([]Item) {}

// Stop is a no-op on platforms without a tray implementation.
func Stop() {}
