//go:build js && wasm

package scenes

import "syscall/js"

// hookLifecycle saves when the tab is hidden or closed. JS events only fire
// between frames, so game state is never half-updated here.
func (a *App) hookLifecycle() {
	doc := js.Global().Get("document")
	if doc.IsUndefined() {
		return
	}
	doc.Call("addEventListener", "visibilitychange", js.FuncOf(func(js.Value, []js.Value) any {
		if doc.Get("visibilityState").String() == "hidden" {
			a.SaveNow()
		}
		return nil
	}))
}
