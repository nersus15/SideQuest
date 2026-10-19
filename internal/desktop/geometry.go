// Package desktop handles the floating window: sizing, anchoring and
// keeping it on screen. Geometry is pure and platform independent;
// the platform files only read and move the native window.
package desktop

// Rect is a screen rectangle in physical pixels. Right and Bottom are exclusive.
type Rect struct {
	Left, Top, Right, Bottom int
}

// W returns the width.
func (r Rect) W() int { return r.Right - r.Left }

// H returns the height.
func (r Rect) H() int { return r.Bottom - r.Top }

// Anchor is the window corner that stays put while the window resizes.
type Anchor struct {
	Right  bool `json:"right"`
	Bottom bool `json:"bottom"`
}

// AnchorFor picks the corner nearest to the closest screen corner, so a
// capsule parked bottom-right expands up and to the left.
func AnchorFor(cur Rect, work Rect) Anchor {
	return Anchor{
		Right:  cur.Left+cur.W()/2 > work.Left+work.W()/2,
		Bottom: cur.Top+cur.H()/2 > work.Top+work.H()/2,
	}
}

// Resize returns a w×h rect that keeps corner a of cur in place. Using the
// same anchor to expand and collapse returns the window to the same spot.
// The result always fits inside work.
func Resize(cur Rect, w, h int, work Rect, a Anchor) Rect {
	next := Rect{Left: cur.Left, Top: cur.Top}
	if a.Right {
		next.Left = cur.Right - w
	}
	if a.Bottom {
		next.Top = cur.Bottom - h
	}
	next.Right = next.Left + w
	next.Bottom = next.Top + h
	return Clamp(next, work)
}

// Clamp moves r so that it lies within work, keeping its size when it fits.
func Clamp(r Rect, work Rect) Rect {
	w, h := r.W(), r.H()
	if r.Right > work.Right {
		r.Left = work.Right - w
	}
	if r.Bottom > work.Bottom {
		r.Top = work.Bottom - h
	}
	if r.Left < work.Left {
		r.Left = work.Left
	}
	if r.Top < work.Top {
		r.Top = work.Top
	}
	r.Right = r.Left + w
	r.Bottom = r.Top + h
	return r
}

// DefaultSpot places a w×h window near the top-right of the work area.
func DefaultSpot(w, h int, work Rect, margin int) Rect {
	r := Rect{Left: work.Right - w - margin, Top: work.Top + margin*3}
	r.Right = r.Left + w
	r.Bottom = r.Top + h
	return Clamp(r, work)
}

// Scale converts device-independent pixels to physical pixels.
func Scale(dip, dpi int) int {
	if dpi <= 0 {
		dpi = 96
	}
	return (dip*dpi + 48) / 96
}
