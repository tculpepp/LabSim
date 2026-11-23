// Package experiment holds the science behind the four v1 experiments and
// the Science Notebook (E-1..E-7). No UI here: everything is testable.
package experiment

import "sort"

// Notebook records first-time discoveries (E-6).
type Notebook struct{ found map[string]bool }

func NewNotebook(found []string) *Notebook {
	n := &Notebook{found: map[string]bool{}}
	for _, id := range found {
		n.found[id] = true
	}
	return n
}

// Add records a discovery; returns true if it is new.
func (n *Notebook) Add(id string) bool {
	if n.found[id] {
		return false
	}
	n.found[id] = true
	return true
}

func (n *Notebook) Has(id string) bool { return n.found[id] }

func (n *Notebook) Found() []string {
	out := make([]string, 0, len(n.found))
	for id := range n.found {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}
