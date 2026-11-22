// Package scenes wires game state, logic packages, and UI into Ebitengine scenes.
package scenes

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"labsim/data"
	"labsim/internal/avatar"
	"labsim/internal/content"
	"labsim/internal/geom"
	"labsim/internal/items"
	"labsim/internal/names"
	"labsim/internal/save"
	"labsim/internal/stack"
	"labsim/internal/state"
	"labsim/internal/ui"
	"labsim/internal/world"
)

const autosaveInterval = 30 * time.Second // SV-1

// Scene is one screen or overlay. Only the top scene updates; Visible scenes draw.
type Scene interface {
	Update(a *App) error
	Draw(a *App, screen *ebiten.Image)
	Opaque() bool
}

// App is the ebiten.Game implementation and owns shared resources.
type App struct {
	W, H       int // logical screen size
	outW, outH int // outside (window/CSS) size
	Map        *world.Map
	Text       *content.Catalog
	Avatars    *avatar.Catalog
	Names      *names.Builder
	Items      *items.Catalog
	Spawns     []items.Rule
	RoomRects  map[string]geom.Rect
	Pointer    ui.Pointer

	// State is the current game, nil on the title screen.
	State    *state.State
	Store    save.Store
	Codec    save.Codec
	autosave *save.Autosaver

	saved  *save.Data // existing save found at startup (enables Continue)
	notice string     // message for the title screen, e.g. corrupt save

	scenes stack.Stack[Scene]
	sfx    func(name string)
}

func NewApp(store save.Store) (*App, error) {
	m, err := world.Load(data.Map)
	if err != nil {
		return nil, fmt.Errorf("load map: %w", err)
	}
	st, err := world.LoadStations(data.Stations)
	if err == nil {
		err = m.AddStations(st)
	}
	if err != nil {
		return nil, fmt.Errorf("load stations: %w", err)
	}
	txt, err := content.Load(data.Text)
	if err != nil {
		return nil, fmt.Errorf("load text: %w", err)
	}
	av, err := avatar.Load(data.Avatar)
	if err != nil {
		return nil, fmt.Errorf("load avatar: %w", err)
	}
	nb, err := names.Load(data.Names)
	if err != nil {
		return nil, fmt.Errorf("load names: %w", err)
	}
	ic, err := items.Load(data.Items)
	if err != nil {
		return nil, fmt.Errorf("load items: %w", err)
	}
	rooms := map[string]geom.Rect{}
	for _, r := range m.Rooms {
		rooms[r.ID] = r.Bounds
	}
	rules, err := items.LoadRules(data.Spawns, ic, rooms)
	if err != nil {
		return nil, fmt.Errorf("load spawns: %w", err)
	}
	a := &App{W: 720, H: geom.LogicalH, Map: m, Text: txt, Avatars: av, Names: nb,
		Items: ic, Spawns: rules, RoomRects: rooms, Store: store, Codec: save.DefaultCodec()}
	a.autosave = save.NewAutosaver(autosaveInterval, a.writeSave)
	a.checkSave()
	a.scenes.Push(NewTitle(a))
	a.hookLifecycle()
	return a, nil
}

// checkSave looks for an existing save. Unreadable saves are backed up, never discarded.
func (a *App) checkSave() {
	a.saved, a.notice = nil, ""
	b, err := a.Store.Load()
	switch {
	case errors.Is(err, save.ErrNoSave):
		return
	case err != nil:
		log.Printf("load save: %v", err)
		a.notice = "title.storageBroken"
		return
	}
	d, err := a.Codec.Decode(b)
	if err != nil {
		log.Printf("decode save: %v", err)
		if berr := a.Store.Backup(b); berr != nil {
			log.Printf("backup save: %v", berr)
		}
		a.notice = "title.corrupt"
		if errors.Is(err, save.ErrFutureVersion) {
			a.notice = "title.future"
		}
		return
	}
	a.saved = &d
}

// HasSave reports whether Continue is available.
func (a *App) HasSave() bool { return a.saved != nil }

// Level is the current difficulty (Junior before a game starts).
func (a *App) Level() content.Level {
	if a.State != nil {
		return a.State.Level
	}
	return content.Junior
}

// T returns player-facing text for the current difficulty.
func (a *App) T(key string) string { return a.Text.Get(key, a.Level()) }

// Tf formats player-facing text.
func (a *App) Tf(key string, args ...any) string { return fmt.Sprintf(a.T(key), args...) }

func (a *App) Push(s Scene)    { a.scenes.Push(s) }
func (a *App) Pop()            { a.scenes.Pop() }
func (a *App) Replace(s Scene) { a.scenes.Replace(s) }

// Say opens a dialog from a character.
func (a *App) Say(who, nameKey, text string) {
	a.Push(NewDialog(a, who, a.T(nameKey), text))
}

// Confirm asks a yes/no question; onYes runs if the player agrees.
func (a *App) Confirm(questionKey string, onYes func(a *App)) {
	a.Push(NewConfirm(a, questionKey, onYes))
}

func (a *App) writeSave() error {
	if a.State == nil {
		return nil
	}
	b, err := a.Codec.Encode(a.State.Snapshot())
	if err != nil {
		return err
	}
	return a.Store.Save(b)
}

// SaveNow saves immediately (room change, job done, purchase, closing).
func (a *App) SaveNow() {
	a.autosave.Now()
	if err := a.autosave.LastErr(); err != nil {
		log.Printf("save: %v", err)
	}
}

// SaveFailing reports whether the last save attempt failed.
func (a *App) SaveFailing() bool { return a.autosave.LastErr() != nil }

// StartNewGame begins play with a fresh scientist, replacing any old save.
func (a *App) StartNewGame(name string, lvl content.Level, av avatar.Avatar) {
	a.State = state.New(name, lvl, a.Map.Spawn.X, a.Map.Spawn.Y)
	a.State.Avatar = a.Avatars.Normalize(av, nil)
	a.Replace(NewPlay(a))
	a.SaveNow()
	a.Say("gizmo", "gizmo.name", a.Tf("gizmo.welcome", name))
	a.State.SetFlag("welcomed")
}

// ContinueGame restores the saved scientist.
func (a *App) ContinueGame() error {
	if a.saved == nil {
		return save.ErrNoSave
	}
	s, err := state.FromSave(*a.saved)
	if err != nil {
		if b, eerr := a.Codec.Encode(*a.saved); eerr == nil {
			_ = a.Store.Backup(b)
		}
		return err
	}
	// A save from an older map could put the player inside a wall.
	box := geom.Centered(s.X, s.Y, playerSize, playerSize)
	if !(geom.Rect{W: a.Map.Width, H: a.Map.Height}).ContainsRect(box) || a.Map.Collides(box, s.IsUnlocked) {
		s.X, s.Y = a.Map.Spawn.X, a.Map.Spawn.Y
	}
	s.Avatar = a.Avatars.Normalize(s.Avatar, s.Owns)
	a.State = s
	a.Replace(NewPlay(a))
	return nil
}

// StartOver deletes the scientist and returns to the title (SV-2).
func (a *App) StartOver() {
	if err := a.Store.Clear(); err != nil {
		log.Printf("clear save: %v", err)
	}
	a.State = nil
	a.checkSave()
	a.Replace(NewTitle(a))
}

func (a *App) Update() error {
	if ebiten.IsWindowBeingClosed() {
		a.SaveNow()
		return ebiten.Termination
	}
	a.Pointer.Update()
	if top, ok := a.scenes.Top(); ok {
		return top.Update(a)
	}
	return nil
}

func (a *App) Draw(screen *ebiten.Image) {
	for _, s := range a.scenes.Visible() {
		s.Draw(a, screen)
	}
	if a.Pointer.TouchDevice() && a.outW < a.outH {
		ui.Fill(screen, geom.Rect{W: float64(a.W), H: float64(a.H)}, 0, 0, ui.ColWall)
		ui.TextCentered(screen, a.T("rotate.hint"), ui.Bold(ui.SizeHeading), float64(a.W)/2, float64(a.H)/2, ui.ColTextLight)
	}
}

func (a *App) Layout(outW, outH int) (int, int) {
	a.outW, a.outH = outW, outH
	a.W, a.H = geom.LogicalSize(outW, outH)
	return a.W, a.H
}

// center returns the screen center.
func (a *App) center() (float64, float64) { return float64(a.W) / 2, float64(a.H) / 2 }

// ItemName returns an item's display name (or its ID if unknown).
func (a *App) ItemName(id string) string {
	if it, ok := a.Items.Get(id); ok {
		return it.Name
	}
	return id
}

// DrawItem draws an item's icon.
func (a *App) DrawItem(screen *ebiten.Image, id string, x, y, size float64) {
	it, ok := a.Items.Get(id)
	if !ok {
		ui.Diamond(screen, x, y, size/2, ui.ColLocked)
		return
	}
	ui.ItemIcon(screen, it.Icon, ui.Hex(it.Color), x, y, size)
}

// Sound plays a named sound effect (no-op until audio is loaded).
func (a *App) Sound(name string) {
	if a.sfx != nil {
		a.sfx(name)
	}
}
