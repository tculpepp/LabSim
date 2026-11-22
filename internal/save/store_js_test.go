//go:build js && wasm

package save

import (
	"syscall/js"
	"testing"
)

func TestLocalStore(t *testing.T) {
	if ls := js.Global().Get("localStorage"); ls.IsUndefined() || ls.IsNull() {
		t.Skip("no localStorage in this JS runtime")
	}
	testStore(t, LocalStore{Key: "labsim-test"})
}

func TestLocalStoreMissingIsError(t *testing.T) {
	// Simulate a browser that blocks storage: the call must error, not panic.
	ls := js.Global().Get("localStorage")
	js.Global().Set("localStorage", js.Undefined())
	defer js.Global().Set("localStorage", ls)
	if v := js.Global().Get("localStorage"); !v.IsUndefined() {
		t.Skip("runtime does not allow removing localStorage")
	}
	if err := (LocalStore{Key: "x"}).Save([]byte("a")); err == nil {
		t.Fatal("expected error with no localStorage")
	}
	if _, err := (LocalStore{Key: "x"}).Load(); err == nil {
		t.Fatal("expected error with no localStorage")
	}
}
