package geom

import "testing"

func TestIntersects(t *testing.T) {
	a := Rect{0, 0, 10, 10}
	cases := []struct {
		name string
		b    Rect
		want bool
	}{
		{"overlap", Rect{5, 5, 10, 10}, true},
		{"inside", Rect{2, 2, 2, 2}, true},
		{"touching edge is not overlap", Rect{10, 0, 5, 5}, false},
		{"apart", Rect{20, 20, 1, 1}, false},
	}
	for _, c := range cases {
		if got := a.Intersects(c.b); got != c.want {
			t.Errorf("%s: Intersects = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestContains(t *testing.T) {
	r := Rect{10, 10, 20, 20}
	if !r.Contains(10, 10) || !r.Contains(29.9, 29.9) {
		t.Fatal("expected inside")
	}
	if r.Contains(30, 15) || r.Contains(9, 15) {
		t.Fatal("expected outside")
	}
}

func TestContainsRect(t *testing.T) {
	outer := Rect{0, 0, 100, 100}
	if !outer.ContainsRect(Rect{0, 0, 100, 100}) {
		t.Fatal("equal rect should be contained")
	}
	if outer.ContainsRect(Rect{50, 50, 60, 10}) {
		t.Fatal("overflowing rect should not be contained")
	}
}

func TestCenterAndExpand(t *testing.T) {
	r := Rect{10, 20, 40, 60}
	if x, y := r.Center(); x != 30 || y != 50 {
		t.Fatalf("Center = %v,%v", x, y)
	}
	if e := r.Expand(5); e != (Rect{5, 15, 50, 70}) {
		t.Fatalf("Expand = %+v", e)
	}
}

func TestCentered(t *testing.T) {
	if r := Centered(50, 50, 10, 20); r != (Rect{45, 40, 10, 20}) {
		t.Fatalf("Centered = %+v", r)
	}
}

func TestLogicalSize(t *testing.T) {
	cases := []struct {
		name         string
		outW, outH   int
		wantW, wantH int
	}{
		{"3:2 desktop window", 960, 640, 720, 480},
		{"16:9 laptop", 1920, 1080, 853, 480},
		{"ultra wide clamps", 3000, 900, MaxLogicalW, 480},
		{"portrait clamps to min width", 390, 844, MinLogicalW, 480},
		{"zero size falls back", 0, 0, 720, 480},
	}
	for _, c := range cases {
		w, h := LogicalSize(c.outW, c.outH)
		if w != c.wantW || h != c.wantH {
			t.Errorf("%s: LogicalSize(%d,%d) = %d,%d want %d,%d", c.name, c.outW, c.outH, w, h, c.wantW, c.wantH)
		}
	}
}
