//go:build !darwin && !windows

package tray

// Start is a no-op on platforms without a tray implementation.
func Start(opts Options) { items = opts.Items }

// Stop is a no-op on platforms without a tray implementation.
func Stop() {}
