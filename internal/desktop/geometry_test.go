package desktop

import "testing"

var work = Rect{0, 0, 1920, 1040}

func TestResizeGrowsAwayFromNearestCorner(t *testing.T) {
	cases := []struct {
		name string
		cur  Rect
		want Rect
	}{
		{"top-left", Rect{100, 100, 280, 170}, Rect{100, 100, 460, 520}},
		{"top-right", Rect{1700, 100, 1880, 170}, Rect{1520, 100, 1880, 520}},
		{"bottom-left", Rect{100, 900, 280, 970}, Rect{100, 550, 460, 970}},
		{"bottom-right", Rect{1700, 900, 1880, 970}, Rect{1520, 550, 1880, 970}},
	}
	for _, c := range cases {
		got := Resize(c.cur, 360, 420, work)
		if got != c.want {
			t.Errorf("%s: got %+v want %+v", c.name, got, c.want)
		}
		// Collapsing again returns to the original spot.
		if back := Resize(got, c.cur.W(), c.cur.H(), work); back != c.cur {
			t.Errorf("%s: collapse went to %+v, want %+v", c.name, back, c.cur)
		}
	}
}

func TestResizeStaysOnScreen(t *testing.T) {
	// A capsule pushed partly past the bottom-right corner.
	got := Resize(Rect{1900, 1000, 2080, 1070}, 360, 420, work)
	if got.Right > work.Right || got.Bottom > work.Bottom || got.Left < 0 || got.Top < 0 {
		t.Fatalf("window escaped work area: %+v", got)
	}
}

func TestClampOffscreenSavedPosition(t *testing.T) {
	// A position saved on a monitor that no longer exists.
	got := Clamp(Rect{5000, -300, 5180, -230}, work)
	if got != (Rect{1740, 0, 1920, 70}) {
		t.Fatalf("got %+v", got)
	}
}

func TestScale(t *testing.T) {
	if Scale(100, 144) != 150 || Scale(100, 96) != 100 || Scale(100, 0) != 100 {
		t.Fatal("bad scaling")
	}
}
