# LabSim

A science-lab game designed by an 8-year-old scientist. Make your own
scientist, explore a 10-room lab, collect things, do Dr. Nova's daily jobs,
run real(ish) experiments, and spend your crystals in Tom's shop.

Written in Go with [Ebitengine](https://ebitengine.org/), compiled to
WebAssembly, and installable as a PWA that works offline.

- Requirements: [`REQUIREMENTS.md`](REQUIREMENTS.md)
- Build plan and status: [`BUILD_PLAN.md`](BUILD_PLAN.md)
- Play-test checklist: [`PLAYTEST.md`](PLAYTEST.md)
- Credits: [`CREDITS.md`](CREDITS.md)

## Quick start

Needs Go (see `go.mod`) and, for `make test-wasm`, Node.js.

| Command | What it does |
|---|---|
| `make run` | Play in a desktop window (fastest for development). |
| `make serve` | Build the web version into `dist/` and serve it at http://localhost:8080. |
| `make test` | `go vet` for desktop and WebAssembly, then all tests. |
| `make test-wasm` | Run the logic tests compiled to WebAssembly under Node. |
| `make size` | Build and report the download size (budget: 15 MB gzipped). |
| `go run ./tools/snapshot -scene shop out.png` | Render one screen to a PNG (see `internal/scenes/debug.go` for scene names). |

## Controls

| | Keyboard | Touch |
|---|---|---|
| Walk | Arrows or WASD | Drag anywhere (joystick) |
| Use / grab / talk | E, Space, or Enter | **Use** button |
| Backpack | B or I | **Bag** |
| Map | M | **Map** |
| Science Notebook | N | **Notes** |
| Job list | J | Tap the job list |
| Pause / settings | Esc or P | **II** |

## Deploying

`.github/workflows/pages.yml` tests, builds, and publishes `dist/` to
GitHub Pages on every push to `main`. In the repository settings, set
**Pages → Source** to **GitHub Actions**. Each build gets a new service
worker cache name, so installed copies update the next time they're online.

## Layout

```
cmd/labsim        entry point
data/             game content (JSON) + validation tests
internal/<pkg>    pure game logic, built test-first (no Ebitengine)
internal/ui       drawing helpers, widgets, procedural art
internal/scenes   screens: title, creator, play, experiments, shop, ...
internal/sfx      synthesized sound effects
tools/            icon generator, dev server, snapshot renderer
web/              index.html, manifest, service worker
```
