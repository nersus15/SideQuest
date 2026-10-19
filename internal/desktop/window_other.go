//go:build !windows

package desktop

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// runtimeWindow is used on macOS (and Linux). It resizes the window with the
// Wails runtime and always keeps the top-left corner fixed: the webview's
// origin is the window's top-left, so nothing moves on screen.
type runtimeWindow struct {
	sizes Sizes
}

// New returns the runtime-based window controller.
func New(_ string, sizes Sizes) Window { return &runtimeWindow{sizes: sizes} }

func (r *runtimeWindow) Place(ctx context.Context, saved *Point) {
	acceptFirstMouse()
	runtime.WindowSetSize(ctx, r.sizes.CapsuleW, r.sizes.CapsuleH)
	if saved != nil {
		runtime.WindowSetPosition(ctx, saved.X, saved.Y)
	}
	runtime.WindowSetAlwaysOnTop(ctx, true)
}

func (r *runtimeWindow) Anchor(context.Context) Anchor { return Anchor{} }

func (r *runtimeWindow) Expand(ctx context.Context) {
	runtime.WindowSetSize(ctx, r.sizes.PanelW, r.sizes.PanelH)
}

func (r *runtimeWindow) Collapse(ctx context.Context) {
	runtime.WindowSetSize(ctx, r.sizes.CapsuleW, r.sizes.CapsuleH)
}

func (r *runtimeWindow) CapsuleSpot(ctx context.Context) (Point, bool) {
	x, y := runtime.WindowGetPosition(ctx)
	return Point{x, y}, true
}

func (r *runtimeWindow) Show(ctx context.Context) {
	runtime.WindowShow(ctx)
}
