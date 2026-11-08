# Gin Rummy

A club car Gin Rummy game by Ironite, written in Go with
[Ebitengine](https://ebitengine.org). Take a seat in the club car of the Starlight
Limited as it crosses the prairie at night: play a passenger, share one computer
with a friend, or play together online. It ships on its own and inside the
[Ironite Launcher](https://github.com/adamglassner/launcher).

## How to play

- Two players, ten cards each. The next card turns up to start the discard pile.
- Melds: a set is three or four of a kind, a run is three or more in a row of one suit. Aces are low.
- Deadwood is everything not in a meld: aces 1, number cards their number, faces 10.
- On your turn draw from the stock or take the top discard, then discard one card.
- Knock when your deadwood is 10 or less. The other player lays off on your melds; you score the difference. If their deadwood is the same or lower, they undercut you: the difference plus 25.
- Gin (no deadwood) scores their deadwood plus 25 and they can't lay off; big gin (all eleven cards) is plus 31.
- Play to 100 (or 50, or 150). The winner gets a 100-point game bonus, everyone gets 25 per hand won, and a shutout doubles the winner's total.
- F2 starts a new game, M toggles music, F11 toggles full screen.

### Ways to play

- **Deal** – you against a passenger in the club car. Every game makes up a new one, with a face, a name and a manner that fit it, and a playing style (some knock early, some hold out for gin).
- **Play Together → Same Computer** – two people take turns at one PC with a "pass the computer" screen between turns.
- **Play Together → Online** – one player hosts, the other joins by IP address (TCP port 47250). The host runs the table and sends the guest only what they can see. A guest who leaves is replaced by a computer player.

Saves and settings are kept in `%APPDATA%\Ironite\GinRummy\save.json`, outside the install folder.

## Play

The easiest way is **[Ironite Launcher](https://github.com/adamglassner/launcher)**: it downloads
Gin Rummy from this repository's releases, installs it and keeps it up to date.

Or download from the [latest release](../../releases/latest): `GinRummy-Setup-v<version>.exe`
installs it with Start menu and desktop shortcuts, or unzip `GinRummy-<version>-windows.zip`
anywhere and run `GinRummy.exe`.

## Code

| File | What it is |
|---|---|
| `main.go` | Window setup and the entry point. |
| `game.go` | Game state, card zones laid out from each player's seat, the hand script (dealing, the upcard, drawing, discarding, knocking, laying off, scoring), results. |
| `internal/gin/` | The rules (melds, deadwood, laying off, scoring) and the computer players (no graphics; has tests). |
| `net.go` | Online play: host/guest protocol (JSON lines over TCP). |
| `setup.go`, `setup_ui.go` | Play Together: seating, lobby, pass screen, text fields, face picker. |
| `faces.go`, `cameo.go` | The passengers: random faces, fitting names, table talk and playing styles, and their portraits drawn in Go. |
| `ui.go`, `panels.go` | The table, the top bar, the turn box and banners; title, rules, credits and result panels. |
| `draw.go`, `glyphs.go`, `assets.go` | Vector drawing, text and suit symbols, embedded art. |
| `sound.go`, `save.go`, `common.go` | Music and sound effects, the save file, shared bits. |
| `assets/` | Card faces, props, the baize, fonts, music and sound effects. |
| `store/` | Cover, banner and icon art for store pages and the launcher. |
| `tools/bake/` | Draws the card art and store art and synthesizes the music and sounds into `assets/` (`node tools/bake/bake.mjs`, needs Playwright and ffmpeg). |
| `tools/mkres/` | Writes the Windows icon and version info into the .exe. |
| `installer/` | Inno Setup scripts for the standalone `GinRummy-Setup-v<version>.exe`. |

`headless_test_main.go` and `sound_stub.go` are only used by test builds:
`go run -tags headless .` plays four solo games, two online games between a host
and a guest over TCP on this computer (including the guest leaving mid-game and
another joining), and a shared-computer game through the pass screen.
`go test ./internal/gin` checks the rules and plays 120 computer hands.

## Building

You need Go 1.24 or newer and bash (Git Bash works on Windows). No C compiler is needed.

```bash
./build.sh
```

This writes `dist/GinRummy/GinRummy.exe` with its README, CREDITS, CHANGELOG and license files.
The version comes from the top entry of `CHANGELOG.md`. For the installer, install
Inno Setup 6.3 or newer, then: `ISCC.exe /DGameVersion=<version> installer\GinRummySetup.iss`

## Releasing

Add a new version section at the top of `CHANGELOG.md` and push to `main`. The
release workflow builds the game, publishes `GinRummy-<version>-windows.zip` as a
GitHub release (Ironite Launcher installs and updates from it), then builds and
attaches the standalone `GinRummy-Setup-v<version>.exe`.

(c) 2026 Ironite
