package desktop

import "context"

// Point is a top-left window position in physical screen pixels.
type Point struct {
	X, Y int
}

// Sizes are the capsule and panel window sizes in device-independent
// (CSS) pixels, including the transparent gutter around each.
type Sizes struct {
	CapsuleW, CapsuleH int
	PanelW, PanelH     int
	// Gutter is the transparent margin around the capsule and panel (for
	// their shadows); the radii are their corner radii. All in CSS pixels.
	Gutter                     int
	CapsuleRadius, PanelRadius int
}

// Window controls the single floating SideQuest window.
//
// The frontend pins the capsule to the anchor corner, so whichever way the
// platform implements expand/collapse, the capsule never moves on screen.
type Window interface {
	// Place puts the capsule at saved (its top-left, or a default spot),
	// keeps it on screen and shows nothing yet. Call before Show.
	Place(ctx context.Context, saved *Point)
	// Anchor returns the corner the panel grows from. It may re-anchor the
	// window if the capsule was dragged to another part of the screen.
	Anchor(ctx context.Context) Anchor
	// Expand makes room for the panel; Collapse goes back to the capsule.
	Expand(ctx context.Context)
	Collapse(ctx context.Context)
	// CapsuleSpot returns the capsule's top-left on screen.
	CapsuleSpot(ctx context.Context) (Point, bool)
	// Show reveals the window without taking focus from the user's work.
	Show(ctx context.Context)
}
