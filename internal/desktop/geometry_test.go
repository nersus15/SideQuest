package desktop

import "testing"

var work = Rect{0, 0, 1920, 1040}

func TestPanelGrowsAwayFromNearestCorner(t *testing.T) {
	cases := []struct {
		name string
		cap  Rect
		want Rect
	}{
		{"top-left", Rect{100, 100, 256, 162}, Rect{100, 100, 480, 570}},
		{"top-right", Rect{1700, 100, 1856, 162}, Rect{1476, 100, 1856, 570}},
		{"bottom-left", Rect{100, 900, 256, 962}, Rect{100, 492, 480, 962}},
		{"bottom-right", Rect{1700, 900, 1856, 962}, Rect{1476, 492, 1856, 962}},
	}
	for _, c := range cases {
		a := AnchorFor(c.cap, work)
		got := PanelRect(c.cap, 380, 470, a)
		if got != c.want {
			t.Errorf("%s: panel %+v want %+v", c.name, got, c.want)
		}
		// The capsule sits exactly in the anchored corner of the panel.
		if back := CapsuleRect(got, c.cap.W(), c.cap.H(), a); back != c.cap {
			t.Errorf("%s: capsule %+v want %+v", c.name, back, c.cap)
		}
		if got.Left < work.Left || got.Top < work.Top || got.Right > work.Right || got.Bottom > work.Bottom {
			t.Errorf("%s: panel %+v leaves the screen", c.name, got)
		}
	}
}

func TestAnchorStaysStableNearCentre(t *testing.T) {
	// Just left of the centre line: the panel grows right and its own centre
	// ends up in the right half. The capsule must still come back to the
	// same spot because the anchor is kept, not re-guessed from the panel.
	cap := Rect{850, 100, 1006, 162}
	a := AnchorFor(cap, work)
	if a.Right {
		t.Fatal("expected a left anchor")
	}
	panel := PanelRect(cap, 380, 470, a)
	if AnchorFor(panel, work) == a {
		t.Fatal("test setup: panel should cross the centre line")
	}
	if back := CapsuleRect(panel, cap.W(), cap.H(), a); back != cap {
		t.Fatalf("capsule moved: %+v -> %+v", cap, back)
	}
}

func TestClampOffscreenSavedPosition(t *testing.T) {
	// A position saved on a monitor that no longer exists.
	got := Clamp(Rect{5000, -300, 5156, -238}, work)
	if got != (Rect{1764, 0, 1920, 62}) {
		t.Fatalf("got %+v", got)
	}
}

func TestDefaultSpot(t *testing.T) {
	got := DefaultSpot(156, 62, work, 16)
	if got != (Rect{1748, 48, 1904, 110}) {
		t.Fatalf("got %+v", got)
	}
}

func TestScale(t *testing.T) {
	if Scale(100, 144) != 150 || Scale(100, 96) != 100 || Scale(100, 0) != 100 {
		t.Fatal("bad scaling")
	}
}

func TestInsetAndRelative(t *testing.T) {
	if got := Inset(Rect{10, 20, 110, 80}, 8); got != (Rect{18, 28, 102, 72}) {
		t.Fatalf("Inset: %+v", got)
	}
	if got := Relative(Rect{110, 220, 250, 266}, Rect{100, 200, 480, 720}); got != (Rect{10, 20, 150, 66}) {
		t.Fatalf("Relative: %+v", got)
	}
}
