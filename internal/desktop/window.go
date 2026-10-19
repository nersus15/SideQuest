package desktop

import "context"

// Point is a top-left window position in physical screen pixels.
type Point struct {
	X, Y int
}

// Window controls the single floating SideQuest window. Sizes are in
// device-independent pixels (CSS pixels).
type Window interface {
	// Place sizes the window, puts it at saved (or a default spot) and keeps
	// it on screen. Call before showing the window.
	Place(ctx context.Context, saved *Point, w, h int)
	// Resize changes the size, growing away from the nearest screen corner.
	Resize(ctx context.Context, w, h int)
	// Spot returns where the window's top-left would be at size w×h,
	// e.g. where the capsule will sit after collapsing.
	Spot(ctx context.Context, w, h int) (Point, bool)
	// Show reveals the window without taking focus from the user's work.
	Show(ctx context.Context)
}
