// Package stack is a generic scene stack with overlay support.
package stack

// Layer is anything that can sit on the stack. Opaque layers hide those below.
type Layer interface{ Opaque() bool }

type Stack[T Layer] struct{ items []T }

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }

func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	v := s.items[len(s.items)-1]
	s.items[len(s.items)-1] = zero
	s.items = s.items[:len(s.items)-1]
	return v, true
}

func (s *Stack[T]) Top() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

// Replace clears the stack and pushes v.
func (s *Stack[T]) Replace(v T) {
	clear(s.items)
	s.items = append(s.items[:0], v)
}

func (s *Stack[T]) Len() int { return len(s.items) }

// Visible returns the layers to draw, bottom to top: from the topmost opaque layer up.
func (s *Stack[T]) Visible() []T {
	start := 0
	for i := len(s.items) - 1; i >= 0; i-- {
		if s.items[i].Opaque() {
			start = i
			break
		}
	}
	return s.items[start:]
}
