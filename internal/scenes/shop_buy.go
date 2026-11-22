package scenes

import (
	"github.com/hajimehoshi/ebiten/v2"

	"labsim/internal/geom"
)

// Buy side of the shop. Filled in by M8.

func (sh *Shop) listTop() float64 { return 112 }

func (sh *Shop) addCategoryTabs(a *App) {}

func (sh *Shop) drawCategoryMarker(screen *ebiten.Image) {}

func (sh *Shop) buyRows(a *App) []shopRow { return nil }

var _ = geom.Rect{}
