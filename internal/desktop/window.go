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
	// Anchor returns the corner the window should grow from right now.
	Anchor(ctx context.Context) Anchor
	// Resize changes the size, keeping corner a in place.
	Resize(ctx context.Context, w, h int, a Anchor)
	// Spot returns where the window's top-left would be at size w×h with
	// corner a kept in place, e.g. where the capsule will sit after collapsing.
	Spot(ctx context.Context, w, h int, a Anchor) (Point, bool)
	// Show reveals the window without taking focus from the user's work.
	Show(ctx context.Context)
}
