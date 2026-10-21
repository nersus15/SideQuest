// Package desktop handles the floating window: anchoring, sizing and
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

// PanelRect returns the w×h rect that shares corner a with the capsule
// rect c, i.e. the panel grows away from the screen corner the capsule is in.
func PanelRect(c Rect, w, h int, a Anchor) Rect {
	r := Rect{Left: c.Left, Top: c.Top}
	if a.Right {
		r.Left = c.Right - w
	}
	if a.Bottom {
		r.Top = c.Bottom - h
	}
	r.Right, r.Bottom = r.Left+w, r.Top+h
	return r
}

// CapsuleRect returns the w×h rect in corner a of the window rect win.
// It is the inverse of PanelRect.
func CapsuleRect(win Rect, w, h int, a Anchor) Rect {
	r := Rect{Left: win.Left, Top: win.Top}
	if a.Right {
		r.Left = win.Right - w
	}
	if a.Bottom {
		r.Top = win.Bottom - h
	}
	r.Right, r.Bottom = r.Left+w, r.Top+h
	return r
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

// Inset shrinks r by d on every side.
func Inset(r Rect, d int) Rect {
	return Rect{r.Left + d, r.Top + d, r.Right - d, r.Bottom - d}
}

// Relative expresses r relative to the top-left corner of origin, as
// window regions expect.
func Relative(r, origin Rect) Rect {
	return Rect{r.Left - origin.Left, r.Top - origin.Top, r.Right - origin.Left, r.Bottom - origin.Top}
}
