//go:build !(js && wasm)

package scenes

// hookLifecycle is a no-op on desktop; App.Update saves when the window closes.
func (a *App) hookLifecycle() {}
