//go:build !windows

package desktop

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// runtimeWindow is a simple fallback for development on macOS and Linux.
// The Wails runtime reports positions and screen sizes in differing units
// across platforms, so it keeps the top-left corner fixed instead of doing
// the corner-aware anchoring used on Windows.
type runtimeWindow struct{}

// New returns the runtime-based window controller.
func New(string) Window { return runtimeWindow{} }

func (runtimeWindow) Place(ctx context.Context, saved *Point, w, h int) {
	runtime.WindowSetSize(ctx, w, h)
	if saved != nil {
		runtime.WindowSetPosition(ctx, saved.X, saved.Y)
	}
	runtime.WindowSetAlwaysOnTop(ctx, true)
}

func (runtimeWindow) Resize(ctx context.Context, w, h int) {
	runtime.WindowSetSize(ctx, w, h)
}

func (runtimeWindow) Show(ctx context.Context) {
	runtime.WindowShow(ctx)
}

func (runtimeWindow) Spot(ctx context.Context, _, _ int) (Point, bool) {
	x, y := runtime.WindowGetPosition(ctx)
	return Point{x, y}, true
}
