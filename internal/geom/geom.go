// Package geom holds small geometry helpers shared by logic and UI.
package geom

// Rect is an axis-aligned rectangle. Edges are half-open: [X, X+W).
type Rect struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	W float64 `json:"w"`
	H float64 `json:"h"`
}

// Centered returns a w×h rect centered on (cx, cy).
func Centered(cx, cy, w, h float64) Rect {
	return Rect{X: cx - w/2, Y: cy - h/2, W: w, H: h}
}

// Intersects reports whether r and o overlap. Touching edges do not count.
func (r Rect) Intersects(o Rect) bool {
	return r.X < o.X+o.W && o.X < r.X+r.W && r.Y < o.Y+o.H && o.Y < r.Y+r.H
}

func (r Rect) Contains(x, y float64) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

func (r Rect) ContainsRect(o Rect) bool {
	return o.X >= r.X && o.Y >= r.Y && o.X+o.W <= r.X+r.W && o.Y+o.H <= r.Y+r.H
}

func (r Rect) Center() (float64, float64) { return r.X + r.W/2, r.Y + r.H/2 }

// Expand grows the rect by d on every side (negative d shrinks it).
func (r Rect) Expand(d float64) Rect {
	return Rect{X: r.X - d, Y: r.Y - d, W: r.W + 2*d, H: r.H + 2*d}
}

// Logical screen size limits. Height is fixed so text and sprites stay a
// consistent size; width follows the device aspect ratio (T-6).
const (
	LogicalH    = 480
	MinLogicalW = 640
	MaxLogicalW = 1066
)

// LogicalSize maps the outside (window/CSS) size to the game's logical size.
func LogicalSize(outW, outH int) (int, int) {
	if outW <= 0 || outH <= 0 {
		return 720, LogicalH
	}
	w := LogicalH * outW / outH
	if w < MinLogicalW {
		w = MinLogicalW
	}
	if w > MaxLogicalW {
		w = MaxLogicalW
	}
	return w, LogicalH
}
