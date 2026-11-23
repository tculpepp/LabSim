# LabSim Build Plan

The source of truth for requirements is `REQUIREMENTS.md`, and IDs like `E-CHEM-1` refer to it. Each milestone ends with a playable build, a play-test by the designer, and a commit.

## Architecture rules
- **Pure logic in `internal/<domain>`**, with no Ebitengine imports. Built test-first: write a failing test, then the code, then refactor.
- **Rendering and input in `internal/ui` and `internal/scenes`**. These packages are thin and call into the logic packages.
- **Game state is one struct** (`internal/state`) that is changed only through methods or commands. This keeps saves simple and lets multiplayer be added later (T-9).
- **Content is data-driven:** `data/*.json` (items, jobs, shop catalog, experiments, prices) is embedded with `go:embed` and validated by tests (P-8).
- **Platform differences live behind interfaces** using build tags: storage (`localStorage` on js/wasm, a file on desktop) and touch vs. keyboard.

Existing pieces to reuse: `economy.Wallet`, `inventory.Backpack`, `clock.Clock`, `jobs.Board`/`jobs.Deal`, `cmd/labsim/main.go` (the room placeholders move into world data), and the Makefile targets `test`/`run`/`wasm`/`serve`.

## Milestones

**M0: Foundations ✅ (done)**
Go module, wallet, backpack, day clock, job board (all tested), Ebitengine placeholder walk, WASM build, PWA manifest and service worker, icon generator, dev server.


**M1: World, movement, controls ✅ (v0.2.0)**
- `internal/world`: room and wall rectangles, doorways, axis-aligned box collision with sliding, room lookup at a point (`RoomAt`), and locked doors (M-1–M-6). The map is stored as JSON in `data/map.json`, laid out from the confirmed 10-room list.
- Camera follows the player, and there is a mini-map overlay (M-5).
- `internal/input`: one action layer (move vector plus interact) shared by keyboard and touch, with an on-screen joystick and interact button shown only on touch devices (T-5, T-6).
- Tests: collision edge cases, door blocking, room lookup, and input mapping.
- Play-test: walk every room on a laptop and an iPad.

**M2: Scenes, text, UI kit ✅ (v0.3.0)**
- Scene manager: Title, then Avatar Creator, then Play, with overlay scenes (dialog, inventory, shop, notebook, settings).
- Text rendering with `text/v2` and an embedded friendly open-licence font. Dialog boxes with character name and portrait. HUD showing the day, crystals, and backpack.
- Junior/Expert text variants loaded from data (D-1–D-3).

**M3: Save system ✅ (v0.4.0)**
- `internal/save`: a versioned JSON snapshot of the game state, with a migration hook (SV-3).
- Storage interface: `localStorage` through `syscall/js` on wasm, and a file in the user config directory on desktop.
- Autosave on room change, job completion, purchase, and every 30 seconds. "Start over" asks for confirmation (SV-1, SV-2).
- Tests: save round-trip, loading an old version, and corrupt-save fallback.

**M4: Avatar creator and safe names ✅ (v0.5.0)**
- Layered avatar (body, skin, hair, coat, clothes, accessories) and choice of starter pet (A-1–A-5, A-8).
- `internal/filter`: profanity filter that handles leetspeak, spacing, repeated letters, and substrings, plus an allow-list so innocent names aren't blocked by accident (S-2). This package gets heavy table-driven tests.
- Preset name builder that combines a title, an adjective, and a science word (A-7, S-3). The wardrobe in My Lab re-opens the creator (A-6).

**M5: Items and pickup ✅ (v0.6.0)**
- Item catalog in `data/items.json`, including star dust and experiment ingredients.
- Item spawns per room, re-rolled each new day from the day number (I-4). Grab an item by walking over it or pressing interact (I-1). Inventory screen (I-5).
- Selling items to Tom (I-3).

**M6: Characters and daily jobs ✅ (v0.7.0)**
- NPCs: Dr. Nova (office), Tom (shop), Gizmo (floats beside the player and runs the tutorial), and wandering lab animals (C-1–C-4, S-6).
- Each morning, deal 3 jobs from `data/jobs.json` seeded by the day number (P-3, P-4). Delivery jobs check the backpack and room. Experiment jobs listen for experiment-completed events. Pay goes to the wallet.
- Gizmo's first-day tutorial: walk, grab, deliver, get paid.

**M7: Experiment framework and the 4 experiments ✅ (v0.8.0)**
- As built: Expert circuit adds an ON/OFF switch (closed-loop idea) instead of series vs. parallel; gravity adds Pluto with the telescope. 26 discoveries, each with a Junior/Expert fun fact; +5 crystal bonus the first time; Science Notebook (N / Notes button).
- `internal/experiment`: a shared interface for the three styles (Sandbox, Predict→Test, Task); a result type (success / funny boom + hint); first-time discovery tracking in the Science Notebook with a bonus payout (E-1–E-7).
- Build one experiment at a time. Each has its logic tested first, then its UI:
  - **M7a** Glowing Slime: ingredient combinations lead to Glow / Bouncy / Rainbow / Stretchy-Sticky slime (E-CHEM-1).
  - **M7b** Circuit Builder: circuit graph, then check whether the bulb lights. Expert mode adds series vs. parallel (E-PHYS-1).
  - **M7c** Microscope Explorer: slides and identification, with slides gated behind the microscope upgrade (E-BIO-1).
  - **M7d** Planet Gravity: jump height and fall time from real `g` values for Moon, Earth, Mars, and Jupiter (E-SPACE-1).
- Fun Fact cards with Junior and Expert text (E-5).

**M8: Shop, upgrades, My Lab**
- `internal/shop`: catalog in `data/shop.json` with categories for tools, outfits/pets, room unlocks (Robot Workshop at about 3 days of pay), decorations, and backpack upgrades (P-5). Buying spends from the wallet and grants the item. Tool upgrades unlock experiment variants (E-7).
- My Lab: place, move, and remove decorations on a grid (P-6).
- Tests: purchase rules, unlock effects, and grid placement.

**M9: Art and audio pass**
- Replace the placeholders with Kenney CC0 packs, recolored for one consistent look (AR-1–AR-3). Credits go in `CREDITS.md`.
- Sprite sheets for the avatar layers, NPCs, Tom (as she described him), and Gizmo (round and floating, screen face). Walk animations and a night tint.
- Sound effects (bubbles, zaps, booms, coin ding, footsteps) with mute and volume controls (AU-1, AU-2).

**M10: Polish and release v1**
- Settings screen: difficulty, sound, day length, controls (ST-1).
- Performance pass at 60 fps on a Chromebook and an iPad, and a check of the compressed bundle size (T-10).
- Offline PWA test: install, go into airplane mode, play (T-4). Add a service worker version bump to the Makefile.
- Deploy as a static site (GitHub Pages by default).
- Walk through the full §14 checklist with the designer.

**After v1**
1. Robot Workshop: build, fix, program, race (RW-1–RW-6).
2. 2–3 more experiments per science.
3. Music and a read-aloud option.
4. Multiple save slots and export.
5. Online multiplayer (chat model still to be decided).

## Dev tools
- `make test` vets for desktop and WASM, then runs the tests.
- `make test-wasm` runs the logic tests compiled to WASM under Node, including `localStorage` saving.
- `go run ./tools/snapshot -scene <name> out.png` saves a screenshot of one screen (see `scenes.DebugScene` for the names).

## Per-milestone definition of done
- `make test` is green, with new logic covered by tests written first.
- `make serve` works in Chrome and Safari, and `make run` works on desktop.
- The designer play-tests it. Her feedback becomes requirement updates or backlog items.
- The diff summary is shown to the user and committed only after approval.

## Risks and mitigations
- **Touch on small phones:** test the joystick and font sizes early, in M1–M2.
- **WASM size and load time:** watch the size every milestone and compress assets.
- **Profanity filter false positives:** use an allow-list, and the preset name builder is always available as a fallback.
- **Scope creep from new ideas:** add them to the "After v1" list unless they replace something already planned.
