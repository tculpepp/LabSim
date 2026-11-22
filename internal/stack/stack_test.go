package stack

import "testing"

type layer struct {
	name   string
	opaque bool
}

func (l layer) Opaque() bool { return l.opaque }

func names(ls []layer) []string {
	out := make([]string, len(ls))
	for i, l := range ls {
		out[i] = l.name
	}
	return out
}

func eq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestEmpty(t *testing.T) {
	var s Stack[layer]
	if _, ok := s.Top(); ok || s.Len() != 0 {
		t.Fatal("empty stack should have no top")
	}
	if _, ok := s.Pop(); ok {
		t.Fatal("Pop on empty should fail")
	}
}

func TestPushPopTop(t *testing.T) {
	var s Stack[layer]
	s.Push(layer{"play", true})
	s.Push(layer{"dialog", false})
	if top, _ := s.Top(); top.name != "dialog" {
		t.Fatalf("top = %s", top.name)
	}
	if p, ok := s.Pop(); !ok || p.name != "dialog" {
		t.Fatalf("pop = %v", p)
	}
	if top, _ := s.Top(); top.name != "play" {
		t.Fatalf("top = %s", top.name)
	}
}

func TestReplaceClears(t *testing.T) {
	var s Stack[layer]
	s.Push(layer{"title", true})
	s.Push(layer{"confirm", false})
	s.Replace(layer{"play", true})
	if s.Len() != 1 {
		t.Fatalf("len = %d", s.Len())
	}
}

func TestVisibleFromLastOpaque(t *testing.T) {
	var s Stack[layer]
	s.Push(layer{"title", true})
	s.Push(layer{"play", true})
	s.Push(layer{"pause", false})
	s.Push(layer{"confirm", false})
	got := names(s.Visible())
	if want := []string{"play", "pause", "confirm"}; !eq(got, want) {
		t.Fatalf("Visible = %v want %v", got, want)
	}
}

func TestVisibleNoOpaqueShowsAll(t *testing.T) {
	var s Stack[layer]
	s.Push(layer{"a", false})
	s.Push(layer{"b", false})
	if got := names(s.Visible()); !eq(got, []string{"a", "b"}) {
		t.Fatalf("Visible = %v", got)
	}
}
