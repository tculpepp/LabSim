# LabSim — Requirements (v0.3, designer answers added)

Designed by: our 8-year-old lead game designer
Working name: **LabSim**
Date: 2026-11-22

## 1. Vision

> "Like Among Us where you walk around a map and go into rooms where there are experiments — except no one is trying to kill anyone. You cannot use bad language."

A friendly, top-down 2D science lab. You design your own scientist, walk around the lab, go into rooms, run experiments, and test your theories. You discover things, do daily jobs, get paid in crystals, and spend them on tools, outfits, new rooms, and decorations for your own lab. The game never ends. You just keep discovering and collecting.

- **Audience:** ages 8–27
- **Tone:** cheerful and silly. No violence, no scary stuff, no bad language.

## 2. Platform & Tech

| ID | Requirement |
|----|-------------|
| T-1 | Written in Go and compiled to WebAssembly (WASM). |
| T-2 | Uses **Ebitengine** for rendering, input, and audio. |
| T-3 | Runs in modern browsers: Chrome (including Chromebook), Safari (iPad/iPhone), Edge, and Firefox. |
| T-4 | Installable as a **PWA** (manifest plus service worker) and **playable offline** once it is installed. |
| T-5 | Supports **keyboard + mouse** (WASD or arrow keys to move, E or click to interact) **and touch** (on-screen joystick plus an interact button). |
| T-6 | The layout adapts to the screen, from phone portrait/landscape up to desktop. Touch controls appear only on touch devices. |
| T-7 | Game logic (economy, inventory, experiments, jobs, day clock, save model) lives in pure Go packages with no Ebitengine dependency, so it can be unit-tested with TDD. |
| T-8 | Static hosting only for v1. No server or accounts. |
| T-9 | The architecture keeps a clean split between game state and simulation, so **online multiplayer can be added later** (see §12). |
| T-10 | Performance target: 60 fps on a mid-range Chromebook and iPad. The first download is under 15 MB compressed (gzip/brotli). At v1.0.0-rc.1 the game is about 4.9 MB gzipped and 21 MB raw (`make size`). |

## 3. Player & Avatar

| ID | Requirement |
|----|-------------|
| A-1 | The first launch opens the **avatar creator** before play begins. |
| A-2 | Customizable: body shape, skin tone, hairstyle, hair color. |
| A-3 | Customizable: lab coat color, shirt, shoes, and patterns. |
| A-4 | Accessories: goggles, hats, gloves, backpacks. Some are free at the start and more can be bought in the shop. |
| A-5 | **Pet sidekick** that follows the player. During avatar creation the player picks a starter pet: **Mini robot, Lab cat, Slime buddy, or Puppy**. More pets and pet outfits are sold in the shop. |
| A-6 | The avatar can be re-edited at any time from the player's own lab (a "mirror" or wardrobe). |
| A-7 | Name entry: a **preset name builder** (e.g., "Doctor" + "Sparkle" + "Beaker") **and** an optional free-text name passed through the language filter (§9). |
| A-8 | Difficulty is chosen at creation: **Junior** or **Expert** (§8). It can be changed later in settings. |

## 4. Map & Movement

| ID | Requirement |
|----|-------------|
| M-1 | Top-down 2D map of connected rooms and hallways in a cute cartoon style. |
| M-2 | **10 rooms** (confirmed by the designer): |
|     | 1. Main Hall (hub/spawn) |
|     | 2. Dr. Nova's Office (daily jobs and pay) |
|     | 3. Shop (run by Tom) |
|     | 4. Chemistry Lab |
|     | 5. Physics Lab |
|     | 6. Greenhouse / Biology Lab |
|     | 7. Space Lab |
|     | 8. Storage Room (supplies and star dust spawn here) |
|     | 9. My Lab (the player's personal, decoratable room) |
|     | 10. **Robot Workshop**: locked at the start. It costs about **3 in-game days of job pay** to unlock (see RW-1–RW-5). |
| M-3 | The player walks with collision against walls and furniture. |
| M-4 | Walking into a room's experiment station and pressing interact opens that experiment. |
| M-5 | A mini-map or full map view shows the rooms, the player's location, locked rooms, and job targets. |
| M-6 | Locked rooms are visible but have a "locked" door showing the price to unlock. |

## 5. Items & Inventory

| ID | Requirement |
|----|-------------|
| I-1 | Items appear on the map (supplies, **star dust**, samples, parts). The player walks over them or presses interact to **grab** them. |
| I-2 | Items go into a **backpack** with limited slots. Backpack upgrades are sold in the shop. |
| I-3 | Item uses: **ingredients** for experiments, **deliveries** for jobs, or **sold** at the shop for coins. |
| I-4 | Supplies respawn every in-game day. |
| I-5 | An inventory screen shows the items, their counts, and a short description. |

## 6. Experiments

General rules:

| ID | Requirement |
|----|-------------|
| E-1 | Each experiment lives at a station inside a room. |
| E-2 | Three play styles, mixed across rooms: |
|     | **Sandbox**: try anything, and results get logged. |
|     | **Predict → Test**: the game asks "What happens if…?", the player predicts, runs the test, and sees whether they were right. |
|     | **Mini-game task**: an Among Us–style puzzle (connect wires, measure, fix). |
| E-3 | Science is **mostly real, with some magic**. Real results are accurate, and magic content is clearly fantasy. |
| E-4 | **Failure = funny boom + hint**: a cartoon poof or explosion, soot on the avatar's face, then a helpful clue. There is no punishment beyond the ingredients that were used. |
| E-5 | Each real-science result can show a **Fun Fact card**. |
| E-6 | **First-time discoveries** are added to the **Science Notebook** and pay a discovery bonus. |
| E-7 | Some experiments need a **lab tool upgrade** to unlock harder variants, such as a better microscope. |

v1 experiments (one per science):

| ID | Room | Experiment | Style | Notes |
|----|------|-----------|-------|-------|
| E-CHEM-1 | Chemistry | **Glowing Slime** (magic) | Sandbox | Combine ingredients to discover slimes: **Glow-in-the-dark** (lights up the room at night), **Bouncy**, **Rainbow** (changes colors), **Stretchy/Sticky**. Each new slime is a discovery. |
| E-PHYS-1 | Physics | **Circuit Builder** | Mini-game + Predict | Connect a battery, wires, a switch, and a bulb. Expert mode adds series vs. parallel circuits. |
| E-BIO-1 | Greenhouse | **Microscope Explorer** | Predict → Test | Look at slides (pond water, leaf, hair, onion skin) and identify what you see. A microscope upgrade unlocks more slides. |
| E-SPACE-1 | Space Lab | **Planet Gravity** | Predict → Test + Sandbox | Jump or drop objects on the Moon, Earth, Mars, and Jupiter, then compare jump height and fall time. |

## 6a. Robot Workshop (unlockable room)

| ID | Requirement |
|----|-------------|
| RW-1 | **Unlock price:** about 3 in-game days of job pay. The exact number of crystals lives in a data file. |
| RW-2 | **Build your own robot:** snap together a head, body, arms, and wheels. The finished robot can follow you and help around the lab. |
| RW-3 | **Fix broken robots:** a puzzle mini-game where broken robots arrive and you repair them. Fixing one pays crystals. |
| RW-4 | **Program a robot:** give it arrow-step instructions to get it through a maze (simple coding). Expert mode adds loops and repeats. |
| RW-5 | **Robot races:** race the robot you built. Its parts change its speed, and winning pays crystals. |
| RW-6 | Not required for the v1 first playable. It is the first content to add after v1. |

## 7. Economy, Jobs & Progression

| ID | Requirement |
|----|-------------|
| P-1 | The currency is **crystals** (shown with a crystal icon). They are different from the **star dust** item you pick up on the map. |
| P-2 | Ways to earn: **daily jobs**, **first-time discoveries**, and **selling items**. |
| P-3 | **In-game day**: about 10 real minutes by default, and configurable. There is a day/night tint, and the morning starts a new day. |
| P-4 | Each in-game morning, **Dr. Nova** gives out 3 jobs, for example "Deliver 3 jars of star dust to the Space Lab" or "Light the bulb in the Physics Lab". Players are paid when a job is done. |
| P-5 | Shop categories: **lab tool upgrades**, **avatar outfits/accessories/pets**, **room unlocks**, **decorations for My Lab**, and **backpack upgrades**. |
| P-6 | **My Lab**: the player places and moves decorations and trophies on a grid. |
| P-7 | The game **never ends**. There is no final win state. Long-term goals are filling the notebook, collecting outfits and decorations, and unlocking every room. |
| P-8 | Prices and payouts live in data files (JSON/YAML) so they can be tuned without code changes. |

## 8. Ages & Difficulty

| ID | Requirement |
|----|-------------|
| D-1 | **Junior** mode: simple words at roughly a 2nd–3rd grade reading level, bigger hints, and simpler puzzles. |
| D-2 | **Expert** mode: harder puzzle variants, more scientific detail on Fun Fact cards, and fewer hints. |
| D-3 | All text is short, and icons are used wherever possible. |

## 9. Safety & Language

| ID | Requirement |
|----|-------------|
| S-1 | **No bad language anywhere** in game content. |
| S-2 | All free-text input (player name, lab name, pet name) passes through a **profanity filter** that handles leetspeak, spacing tricks, and substrings. Blocked text shows a friendly message: "Let's pick a different name!" |
| S-3 | The preset name builder is always available as the safe default. |
| S-4 | No violence, no killing, no scary content. Booms are silly, never harmful. |
| S-5 | No ads, no real-money purchases, no data collection, and no external network calls in v1. |
| S-6 | Lab animals are always treated kindly. Biology interactions are observing, feeding, and housing animals, never harming them. |

## 10. Characters (computer-controlled)

| ID | Character | Role |
|----|-----------|------|
| C-1 | **Dr. Nova** (Boss Scientist) | Loves space and always smiles. Gives daily jobs, pays the player in crystals, and congratulates them. |
| C-2 | **Tom** (Shopkeeper) | Runs the shop and loves to help people become better scientists. Gives friendly tips about which upgrades help. **Look:** tall, light skin, blue eyes, brown hair that is a little lopsided, short beard. He is right-handed, so he holds things in his right hand. |
| C-3 | **Gizmo** (Helper Owl) | **Look:** a round, bright green cartoon owl with big friendly eyes and lab goggles pushed up on its forehead. Floats next to the player, flapping little wings. Runs the tutorial, gives hints on request, and points the way to job targets. |
| C-4 | **Lab Animals** | **Hamster/mouse**, **frog/axolotl** (in a tank), **bunny** (hops in the halls), **butterflies/bugs** (in the Greenhouse). They wander the map, and some are part of biology content. |

## 11. Saving, Audio, Settings

| ID | Requirement |
|----|-------------|
| SV-1 | **Auto-save** to browser storage (IndexedDB/localStorage) on room change, job completion, purchase, and every 30 s. |
| SV-2 | **One scientist per device.** A "Start over" option asks for confirmation first. |
| SV-3 | The save format is versioned so later updates don't break saves. |
| AU-1 | **Sound effects only** in v1 (bubbles, zaps, booms, coin ding, footsteps). No music. (Synthesized in code; footsteps left out as too repetitive.) |
| AU-2 | Mute toggle and volume slider. |
| ST-1 | Settings: difficulty, sound, day length, control style. |

## 12. Future (not in v1, but design for it)

- Online multiplayer, where friends share the lab (chat model still to be decided: emotes, preset phrases, or filtered text)
- More experiments per science
- Music
- Read-aloud text for young readers
- Multiple save slots and save export/import
- The designer's own drawings as special items or pets

## 13. Art

| ID | Requirement |
|----|-------------|
| AR-1 | Bright, cute cartoon style with a top-down view and round, friendly characters. |
| AR-2 | ~~v1 uses free CC0 asset packs (e.g., Kenney.nl)~~ **Changed in M9:** all art is drawn in code (vector shapes), so it is one consistent style, tiny to download, and free of licence questions. Asset packs remain an option later. |
| AR-3 | Avatar parts are layered (body, hair, clothes, accessories), so combinations are cheap to add. (Drawn as layered vector shapes rather than sprite images.) |

## 14. v1 "Done" Definition (first playable)

1. Create an avatar (with a safe name) and pick a difficulty.
2. Walk around the 8–10 room map on both keyboard and touch.
3. Grab items into the backpack.
4. Play all 4 experiments, including success, funny-boom failure, hints, and notebook discoveries.
5. Receive and complete daily jobs from Dr. Nova and get paid in crystals.
6. Buy at least one item in each shop category, and place a decoration in My Lab.
7. Auto-save and resume.
8. Install as a PWA and play offline.

## 15. Open Questions for the Designer

None right now. New questions will come up during play-testing.
