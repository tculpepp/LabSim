// Package decor is the My Lab decorating grid (P-6).
package decor

import (
	"errors"
	"sort"
)

var (
	ErrOutside = errors.New("outside the grid")
	ErrBlocked = errors.New("cell is blocked")
	ErrTaken   = errors.New("cell is taken")
)

type Cell struct{ X, Y int }

type Placement struct {
	ID   string `json:"id"`
	Cell Cell   `json:"cell"`
}

// Layout is a W×H grid; blocked cells hold furniture or doorways.
type Layout struct {
	W, H    int
	blocked map[Cell]bool
	cells   map[Cell]string
	at      map[string]Cell
}

// NewLayout restores saved placements, dropping any that no longer fit.
func NewLayout(w, h int, blocked []Cell, saved []Placement) *Layout {
	l := &Layout{W: w, H: h, blocked: map[Cell]bool{}, cells: map[Cell]string{}, at: map[string]Cell{}}
	for _, c := range blocked {
		l.blocked[c] = true
	}
	for _, p := range saved {
		if _, dup := l.at[p.ID]; !dup {
			l.Place(p.ID, p.Cell)
		}
	}
	return l
}

func (l *Layout) Inside(c Cell) bool { return c.X >= 0 && c.Y >= 0 && c.X < l.W && c.Y < l.H }

func (l *Layout) Blocked(c Cell) bool { return l.blocked[c] }

// Place puts id at c, moving it if it was already placed.
func (l *Layout) Place(id string, c Cell) error {
	switch {
	case !l.Inside(c):
		return ErrOutside
	case l.blocked[c]:
		return ErrBlocked
	}
	if other, ok := l.cells[c]; ok && other != id {
		return ErrTaken
	}
	l.Remove(id)
	l.cells[c], l.at[id] = id, c
	return nil
}

func (l *Layout) Remove(id string) {
	if c, ok := l.at[id]; ok {
		delete(l.cells, c)
		delete(l.at, id)
	}
}

func (l *Layout) At(c Cell) (string, bool) { id, ok := l.cells[c]; return id, ok }

func (l *Layout) Where(id string) (Cell, bool) { c, ok := l.at[id]; return c, ok }

// Placements returns everything placed, sorted by ID (stable saves).
func (l *Layout) Placements() []Placement {
	out := make([]Placement, 0, len(l.at))
	for id, c := range l.at {
		out = append(out, Placement{id, c})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
