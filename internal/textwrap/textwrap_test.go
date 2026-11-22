package textwrap

import (
	"reflect"
	"testing"
)

// mono measures 1 unit per rune.
func mono(s string) float64 { return float64(len([]rune(s))) }

func TestWrap(t *testing.T) {
	cases := []struct {
		name string
		text string
		max  float64
		want []string
	}{
		{"fits", "hello world", 20, []string{"hello world"}},
		{"breaks at space", "hello big world", 10, []string{"hello big", "world"}},
		{"long word own line", "a supercalifragilistic b", 10, []string{"a", "supercalifragilistic", "b"}},
		{"newline kept", "one\ntwo three", 20, []string{"one", "two three"}},
		{"blank line kept", "one\n\ntwo", 20, []string{"one", "", "two"}},
		{"extra spaces collapse", "a   b", 20, []string{"a b"}},
		{"empty", "", 10, []string{""}},
	}
	for _, c := range cases {
		if got := Wrap(c.text, c.max, mono); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: Wrap = %q want %q", c.name, got, c.want)
		}
	}
}

func TestPages(t *testing.T) {
	lines := []string{"1", "2", "3", "4", "5"}
	got := Pages(lines, 2)
	want := [][]string{{"1", "2"}, {"3", "4"}, {"5"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Pages = %v", got)
	}
	if got := Pages(nil, 3); len(got) != 1 || len(got[0]) != 0 {
		t.Fatalf("Pages(nil) = %v, want one empty page", got)
	}
}
