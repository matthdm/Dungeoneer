---
plan-id: W1-web-desktop-deploy
status: active
owner: unassigned
branch: main          # per standing instruction: work on main, no worktrees; only Matthew commits
depends-on: []
last-touched: 2026-10-08
---

# Plan: Web (Chrome) + desktop builds from one codebase

## Goal

Testers can play Dungeoneer in Chrome from a link, and the Windows build keeps working exactly as it does today, both produced from the same source tree. "Done" means: one script produces a web bundle (`index.html` + `wasm_exec.js` + `dungeoneer.wasm`) that boots to the main menu, plays a run, and keeps saves across a page reload; the Windows `.exe` runs with no data folders beside it; and CI builds both.

## Background (read before starting)

Measured on 2026-10-08, Go 1.23.3, Ebiten v2.8.8:

- `GOOS=js GOARCH=wasm go build .` from `src/` **already succeeds**: 17.7 MB `.wasm`, 6.4 MB gzipped. It was compiled only, never run in a browser.
- All sprites are embedded already (`src/images/embed.go`). Nothing in this plan touches image loading.
- What does not work in a browser is the filesystem. Two kinds of file access exist and they get different fixes:
  - **Read-only game data** (dialogues, `data/*.json`, `levels/hub.json`, `dev_settings.json`) is read with `os.ReadFile` / `os.ReadDir` relative to the working directory. Fix: embed it.
  - **Saves** (`meta.json`, `runsave.json`, `options.json`, `controls.json`, `echoes/run_N.json`) are written with `os.WriteFile`. Fix: a small storage package with a disk backend and a `localStorage` backend chosen by build tag.
- Every data loader in `game.NewGame` is non-fatal on a missing file (`src/game/game.go:264-282`), and `loadHub` falls back to a generated hub. So an un-fixed web build probably boots, but with no dialogue, lore, flavor text or hand-made hub. This is also true of the current Windows zip: `package.ps1` ships only the exe, `meta.json`, `controls.json` and `levels/hub.json`, so packaged builds today have no dialogues or lore. Phase 1 fixes that as a side effect.

## Design

Two new leaf packages. Neither imports Ebiten or any other project package.

**`src/gamedata` — read-only data, disk first, embedded fallback.**

```go
package gamedata

// Register installs the embedded copy of the game's data files.
func Register(fsys fs.FS)

// ReadFile returns the file at path. A file on disk wins; otherwise the
// embedded copy is used. path may use either slash direction.
func ReadFile(path string) ([]byte, error)

// ReadDir lists dir from disk if it exists there, otherwise from the embedded copy.
func ReadDir(dir string) ([]fs.DirEntry, error)
```

Disk-first is deliberate: it keeps live editing of JSON during development (no rebuild), and it keeps every existing loader test passing unchanged, because those tests pass temp-dir paths. In a browser the disk read always fails, so the embedded copy is used.

The embed directive has to live in a directory above `dialogues/` and `data/`, so it goes in package `main`:

```go
// src/embed.go
package main

import "embed"

//go:embed dialogues/*.json data/*.json levels/hub.json dev_settings.json
var dataFS embed.FS
```

and `main()` calls `gamedata.Register(dataFS)` before `game.NewGame`.

**`src/storage` — saves.**

```go
package storage

// ReadFile returns the saved blob called name (a slash-separated relative
// name such as "meta.json" or "echoes/run_3.json"). A missing blob returns
// an error for which errors.Is(err, fs.ErrNotExist) is true.
func ReadFile(name string) ([]byte, error)

// WriteFile stores data under name, creating parent directories on disk.
func WriteFile(name string, data []byte) error

// Remove deletes name. Removing a missing blob is not an error.
func Remove(name string) error
```

- `storage_desktop.go` (`//go:build !js`): `os.ReadFile` / `os.MkdirAll` + `os.WriteFile` / `os.Remove` on `filepath.FromSlash(name)`, relative to the working directory. This is the same location saves use today, so no existing save is orphaned and the existing tests that poke `meta.json` directly still pass.
- `storage_js.go` (`//go:build js`): `localStorage` under the key `"dungeoneer/" + name`. `syscall/js` panics when JavaScript throws (quota exceeded, storage blocked in a sandboxed iframe), so every call goes through one helper that recovers and returns an error:

```go
func call(method string, args ...any) (v js.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("storage: localStorage.%s: %v", method, r)
		}
	}()
	return js.Global().Get("localStorage").Call(method, args...), nil
}
```

  A `getItem` result that `IsNull()` returns `&fs.PathError{Op: "read", Path: name, Err: fs.ErrNotExist}`, which satisfies both `errors.Is(err, fs.ErrNotExist)` and the existing `os.IsNotExist(err)` checks in `runsave.go` and `metasave.go`.

**Quit.** A browser tab cannot exit. Two build-tagged files in `src/ui` export `const CanQuit` (`true` on desktop, `false` on web). The main menu and pause menu leave out "Exit Game" when it is false.

## Scope

**In scope:**
- Embed dialogues, `data/*.json`, `levels/hub.json`, `dev_settings.json`.
- Route meta, run, options, controls and echo saves through `storage`.
- Hide "Exit Game" on web.
- Web bundle: page shell, build script, local static server for testing.
- Chrome playtest: input, performance, save persistence. Fix what it finds inside the envelope.
- CI: fix the existing workflow and add a web job.

**Out of scope (do not change in this plan):**
- macOS and Linux builds. The storage package makes the later change small (one function decides the save directory), but nothing here builds or tests them.
- Moving desktop saves out of the working directory.
- Dev-only disk features on web: pause-menu Load/Save Level and Load/Save Player, level-editor F5/F6, the `--screenshot` flag, dev-menu benchmark scenarios. They already show an error line or print and continue when the read fails; leave them.
- Syncing saves between browser and desktop.
- Automated upload to itch.io (needs an API key; see Open questions).
- Firefox/Safari/mobile. Chrome on desktop is the target.

## File envelope

**Touched (modify freely within this plan):**
- New: `src/gamedata/gamedata.go`, `src/gamedata/gamedata_test.go`
- New: `src/storage/storage_desktop.go`, `src/storage/storage_js.go`, `src/storage/storage_test.go`
- New: `src/embed.go`
- New: `src/ui/platform_desktop.go`, `src/ui/platform_js.go`
- New: `src/web/index.html`, `src/cmd/webserve/main.go`, `build_web.ps1`
- `src/main.go` — register the embedded FS
- `src/dialogue/loader.go` — `os.ReadFile`/`os.ReadDir` → `gamedata`
- `src/game/lore.go`, `src/game/biome_flavor.go`, `src/game/enemy_flavor.go`, `src/game/event_room.go`, `src/game/combat_adapter.go`, `src/items/flavor.go` — the one `os.ReadFile` line in each loader
- `src/leveleditor/io.go` — **only** the read in `LoadLevelFromFile` (line 919)
- `src/game/metasave.go`, `src/game/runsave.go`, `src/game/options.go`, `src/game/echo_recorder.go`, `src/game/hub.go` (line 960 only)
- `src/controls/config.go`, `src/controls/config_test.go`
- `src/game/draw.game.go` (main-menu label list only; added 2026-10-08 with Matthew's approval)
- `src/game/game.go` (line 411 only), `src/game/handlers.game.go` (lines 59, 87 only), `src/ui/main_menu.go`, `src/ui/pause_menu.menu.go`
- `.github/workflows/build.yaml`, `.gitignore`
- `CLAUDE.md`, `README.md`, `design-docs/roadmap.md`, `plans/_QUEUE.md`

**Forbidden (do not modify in this plan, even if it would be convenient):**
- `src/coords/`, `src/collision/` — owned by `OFFSET_UNIFICATION_PLAN.md`
- `src/combat/`, `src/cmd/benchmarker/`, `src/cmd/combatvis/` — uncommitted benchmarker work is sitting in the tree; `combat.LoadScenario` stays on `os.ReadFile`
- `src/images/` — already embedded
- `src/leveleditor/layered_io.go`, `src/entities/player_io.go`, `src/ui/load_level_menu.go`, `src/ui/load_player_menu.go` — dev-only disk features, out of scope
- `package.ps1` — see Open questions

## Acceptance criteria

- [ ] `dungeoneer.exe` copied alone into an empty folder starts, shows Varn's dialogue, lore entries and the hand-made hub, and writes its saves into that folder.
- [ ] Running from `src/` still reads edited JSON from disk without a rebuild.
- [x] `build_web.ps1` produces `dist/web/` and `dungeoneer-web.zip`; served locally, the game reaches the main menu in Chrome with no errors in the DevTools console.
- [ ] In Chrome: start a run, reach floor 2, reload the page, and "Continue" resumes the run. Rebind a key, change an option, reload: both persist.
- [ ] In Chrome: die once, return to hub, and the echo from that run is listed at the Echo Shrine after a reload.
- [ ] No "Exit Game" entry on web; it is present and works on Windows.
- [ ] Findings table in Phase 4 is filled in (input conflicts, frame rate, floor generation time), with each row either fixed or recorded under Open questions.
- [ ] CI produces a Windows exe artifact and a web zip artifact from one workflow run.
- [x] `cd src && go build ./...` passes
- [x] `cd src && GOOS=js GOARCH=wasm go build -o NUL .` passes (PowerShell: `$env:GOOS='js'; $env:GOARCH='wasm'; go build -o NUL .`)
- [ ] `cd src && go test ./...` passes

## Phases

Each phase leaves the Windows build fully working. Run `cd src && go test ./...` before starting to record the passing baseline. No phase ends with a commit by an agent: stop, report, and Matthew commits.

### Phase 1: Embed read-only data

Tests first, in `src/gamedata/gamedata_test.go`, using `testing/fstest.MapFS` as the embedded copy and `t.TempDir()` for disk:

- [x] 1.1 `TestReadFile_EmbeddedFallback` — file absent on disk, present in the registered FS → embedded bytes returned.
- [x] 1.2 `TestReadFile_DiskWins` — same relative path present in both → disk bytes returned.
- [x] 1.3 `TestReadFile_BackslashPath` — `ReadFile("data\\lore.json")` finds the embedded `data/lore.json` (Windows callers build paths with `filepath.Join`).
- [x] 1.4 `TestReadFile_MissingEverywhere` — absent in both → error for which `errors.Is(err, fs.ErrNotExist)` is true.
- [x] 1.5 `TestReadFile_NothingRegistered` — `Register` never called, file absent on disk → the disk error is returned, no nil-pointer panic.
- [x] 1.6 `TestReadDir_EmbeddedFallback` and `TestReadDir_DiskWins` — same two cases for directories.
- [x] 1.7 Run `go test ./gamedata/`; confirm the tests fail to compile, then implement `src/gamedata/gamedata.go` per the Design section. Normalise with `path.Clean(filepath.ToSlash(p))` before the embedded lookup. Tests that change the registered FS must restore it with `t.Cleanup`.
- [x] 1.8 Add `src/embed.go` (Design section) and call `gamedata.Register(dataFS)` in `src/main.go` before `game.NewGame`.
- [x] 1.9 Swap the reads. Each is a one-line change from `os.ReadFile(path)` to `gamedata.ReadFile(path)`; drop the `os` import where it becomes unused:
  - `src/dialogue/loader.go:16`, `:36`, `:65`, and `os.ReadDir` → `gamedata.ReadDir` at `:56`
  - `src/game/lore.go:31`, `src/game/biome_flavor.go:15`, `src/game/enemy_flavor.go:15`, `src/game/event_room.go:54`, `src/items/flavor.go:18`
  - `src/game/combat_adapter.go:35` (`dev_settings.json`)
  - `src/leveleditor/io.go:919` (`LoadLevelFromFile`, used for `levels/hub.json`)
- [x] 1.10 `go build ./...` and `go test ./...` pass. The existing loader tests (`dialogue/loader_extra_test.go`, `game/lore_test.go`) must pass **unmodified**; if one fails, the disk-first rule is broken.
- [ ] 1.11 Manual: build the exe, copy it alone to an empty folder, run it. Talk to Varn (dialogue appears), open the lore library (entries have text), confirm the hub is the hand-made one. Then from `src/`, edit a line in a dialogue JSON, run without rebuilding, confirm the edit shows.

### Phase 2: Save storage

Tests first, in `src/storage/storage_test.go` (runs against the desktop backend; each test `t.Chdir`s — or `os.Chdir` + `t.Cleanup` on Go 1.23 — into `t.TempDir()`):

- [x] 2.1 `TestRoundTrip` — `WriteFile("a.json", x)` then `ReadFile("a.json")` returns `x`.
- [x] 2.2 `TestReadMissing` — `errors.Is(err, fs.ErrNotExist)` **and** `os.IsNotExist(err)` are both true (existing callers use the latter).
- [x] 2.3 `TestWriteCreatesParentDir` — `WriteFile("echoes/run_1.json", x)` succeeds in an empty directory and reads back.
- [x] 2.4 `TestOverwrite` — second write to the same name replaces the first.
- [x] 2.5 `TestRemove` — remove then read gives not-exist; `Remove` of a name that never existed returns nil.
- [x] 2.6 `TestWriteOverDirectoryFails` — a directory named `controls.json` exists → `WriteFile` and `ReadFile` return a non-nil error that is not not-exist.
- [x] 2.7 Implement `storage_desktop.go` and `storage_js.go` per the Design section. Type-check the web backend with `GOOS=js GOARCH=wasm go vet ./storage/` — it cannot be unit-tested without a browser; Phase 3 and 4 exercise it for real.
- [x] 2.8 `src/game/metasave.go`: `:68`, `:87` → `storage.ReadFile(metaSavePath)`; `:108` → `storage.WriteFile(metaSavePath, data)`. Keep the `os.IsNotExist` check at `:88`.
- [x] 2.9 `src/game/runsave.go`: `:36` → `storage.WriteFile`, `:42` → `storage.ReadFile`, `:61` → `storage.Remove`.
- [x] 2.10 `src/game/options.go`: `:29`, `:53`.
- [x] 2.11 `src/game/echo_recorder.go`: delete the `os.MkdirAll` block at `:84-87` (the storage package creates the directory), `:93` → `storage.WriteFile(path, data)`, `:103` → `storage.Remove(oldest)`, `:114` → `storage.ReadFile(path)`. `src/game/hub.go:960` → `storage.Remove(path)`. The names stored in `MetaSave.EchoFiles` (`echoes/run_N.json`) are already valid storage names; do not change them.
- [x] 2.12 `src/controls/config.go`: `SaveBindings` → `storage.WriteFile(configFileName, data)`; `LoadBindings` → `storage.ReadFile(configFileName)`, returning nil when `errors.Is(err, fs.ErrNotExist)`. Delete `GetConfigPath` (no remaining caller). In `src/controls/config_test.go`, delete `TestGetConfigPath_Error`: it tests the removed function, and it is already skipped on Windows. `TestSaveAndLoadBindings` and `TestSaveAndLoadBindings_FileErrors` must pass with only their `GetConfigPath()` calls replaced by `filepath.Join(tmpDir, configFileName)`.
- [x] 2.13 `go build ./...`, `go test ./...`, and the `GOOS=js` build all pass. `game/metasave_test.go`, `runsave_test.go`, `options_test.go`, `echo_recorder_test.go` must pass unmodified.
- [ ] 2.14 Manual on Windows: with an existing `meta.json` / `runsave.json` in `src/`, launch and confirm Continue still resumes the run and remnants/lore are intact.

### Phase 3: Web bundle and first boot in Chrome

- [x] 3.1 `src/ui/platform_desktop.go` (`//go:build !js`, `const CanQuit = true`) and `src/ui/platform_js.go` (`//go:build js`, `const CanQuit = false`).
- [x] 3.2 `src/ui/main_menu.go:50` and `src/ui/pause_menu.menu.go:67`: append "Exit Game" only when `CanQuit`. Before editing `main_menu.go`, check whether any code indexes the `Options` slice by position (the handlers in `src/game/handlers.game.go` switch on the text, which is safe); if something does, record it under Open questions and stop this sub-task. **Result: pause menu done; main menu blocked — see Open questions.** **Unblocked later the same day:** `MainMenu.Options` now comes from `mainMenuOptions(CanQuit)` and `draw.game.go` takes its labels from `MainMenu.Labels()`, so the two cannot drift apart.
- [x] 3.3 Because 3.2 was blocked for the main menu, the three `os.Exit` calls (`src/game/game.go:411`, `src/game/handlers.game.go:59`, `:87`) now go through `quitGame`, at the bottom of `handlers.game.go`, which does nothing when `ui.CanQuit` is false. Without this, clicking the still-visible main-menu entry would stop the Go program and freeze the page.
- [x] 3.4 `src/web/index.html`:

```html
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>Dungeoneer</title>
<style>html, body { margin: 0; height: 100%; background: #000; overflow: hidden; }</style>
</head>
<body>
<script src="wasm_exec.js"></script>
<script>
const go = new Go();
WebAssembly.instantiateStreaming(fetch("dungeoneer.wasm"), go.importObject)
  .then((result) => go.run(result.instance))
  .catch((err) => { document.body.style.color = "#fff"; document.body.textContent = "Failed to load: " + err; });
</script>
</body>
</html>
```

- [x] 3.5 `build_web.ps1` at the repo root: build `src` with `GOOS=js GOARCH=wasm` to `dist/web/dungeoneer.wasm`; copy `src/web/index.html`; copy `wasm_exec.js` from `$(go env GOROOT)/lib/wasm/` if it exists there, otherwise `$(go env GOROOT)/misc/wasm/` (the file moved in Go 1.24; it must come from the same Go that built the wasm); zip `dist/web/*` to `dungeoneer-web.zip` with `index.html` at the zip root. Restore `GOOS`/`GOARCH` afterwards. Add `dist/` and `dungeoneer-web.zip` to `.gitignore`.
- [x] 3.6 `src/cmd/webserve/main.go`: `http.FileServer(http.Dir(*dir))` on `localhost:8080`, `-dir` defaulting to `../dist/web`. Go's server sends `application/wasm`, which `instantiateStreaming` requires.
- [x] 3.7 Run `./build_web.ps1`, then `cd src; go run ./cmd/webserve`, open `http://localhost:8080` in Chrome with DevTools open. Record in the Progress log what happens: reaches main menu or not, and any console errors verbatim. Fix anything inside the envelope; anything outside goes to Open questions.

### Phase 4: Chrome playtest

Fill in this table. One row per finding; "OK" rows are still recorded.

| Area | Check | Result | Action |
|---|---|---|---|
| Saves | Reach floor 2, reload, Continue resumes | | |
| Saves | Rebind a key, toggle an option, reload | Option half verified 2026-10-08 (headless Chrome): volume 0.8→0.6 written to `dungeoneer/options.json`, survived reload, options screen showed it. Key rebind not exercised. | Rebind still to check by hand |
| Saves | Die, reload, echo listed at Echo Shrine | | |
| Saves | Banish an echo, reload, it stays gone | | |
| Input | Every default binding in `src/controls/` that uses Tab, Esc, F-keys, Ctrl or Alt: does Chrome act on it too? | | |
| Input | Right-click: browser context menu suppressed? | | |
| Input | Mouse wheel / zoom: page scroll or zoom suppressed? | | |
| Input | Fullscreen option: enters, Esc exits cleanly, option state stays correct | | |
| Display | Canvas fills the window, resizes, text is sharp | | |
| Audio | Sound starts after the first click | | |
| Perf | FPS/TPS in hub and on a generated floor with enemies (`ebiten.ActualFPS`/`ActualTPS`) | Main menu and hub: 60 TPS, ~165 FPS (headless Chrome, software GL, 1264×625). Generated floor with enemies not measured. | Measure a real floor by hand |
| Perf | Floor generation time; does the tab freeze visibly? | | |
| Load | Time from page open to main menu on a cold cache | Under 12 s from localhost with a fresh profile (upper bound only; not timed precisely, and no network latency). | Time it on itch.io |
| Robustness | Corrupt `meta.json` / `runsave.json` / `controls.json` blobs in localStorage | OK 2026-10-08: game boots and New Game reaches the hub. | — |
| Robustness | Click "Exit Game" on the main menu | OK 2026-10-08: nothing happens, game keeps running (Options opened afterwards). Entry is still visible — see Open questions. | — |
| Hosting | Same checks pass when the zip is uploaded to itch.io as a restricted HTML5 project (saves inside itch's iframe especially) | | |

- [ ] 4.1 Run the table against the local server.
- [ ] 4.2 Fix findings that are inside the envelope. Performance findings in `pathing`, `fov` or `levels` are outside it: record them under Open questions and reference `backlog-astar-heap.md` / `backlog-fov-dynamic-ray.md` where they apply.
- [ ] 4.3 Matthew uploads `dungeoneer-web.zip` to itch.io (HTML5, restricted/draft, "index.html" embed) and the Hosting row is run there.

### Phase 5: CI

- [x] 5.1 `.github/workflows/build.yaml`: set `defaults.run.working-directory: src`; replace `go-version: '1.21'` with `go-version-file: src/go.mod` (the module needs 1.23.3, and `go build` currently runs at the repo root where there is no `go.mod`).
- [x] 5.2 Windows job on `windows-latest`: `go build ./...`, `go test ./...`, then `go build -ldflags -H=windowsgui -o <name>.exe .`; upload the exe.
- [x] 5.3 Web job on `ubuntu-latest`: `GOOS=js GOARCH=wasm go build -o ../dist/web/dungeoneer.wasm .`, copy `index.html` and `wasm_exec.js` (same lib/misc fallback as 3.5); upload `dist/web` as the artifact.
- [ ] 5.4 Trigger the workflow manually once; confirm both artifacts download and the web one runs under `webserve`.

### Phase 6: Cleanup

- [x] 6.1 `README.md`: how to build and run the web version; link for testers once one exists.
- [ ] 6.2 `design-docs/roadmap.md`: add the row and mark it ✅.
- [x] 6.3 `CLAUDE.md`: status block entry; add `src/gamedata` and `src/storage` to the "Where to look" table under a "Loading data / saving" row, with the rule "never call `os.ReadFile`/`os.WriteFile` for game data or saves directly".
- [ ] 6.4 Move this plan to `plans/COMPLETED/` and update `plans/_QUEUE.md`.

## Review focus

Conditions the tests above do not cover, most likely first. Each is checked by hand in Phase 4 unless noted.

1. **`localStorage` unavailable or full** (blocked third-party storage in an iframe, private window, 5 MB quota). Expected: the game keeps running and behaves as if there is no save; it must not crash. The `call` helper's `recover` is what guarantees this — verify by running once in an Incognito window with third-party cookies blocked.
2. **A corrupt or truncated save blob.** Expected: same as desktop today (defaults for meta/options, error for run save). Unchanged code paths, but confirm once by editing the `dungeoneer/meta.json` key in DevTools → Application → Local Storage to `{`.
3. **Stale data folder next to a new exe.** Disk wins over embedded, so an old `dialogues/` folder shadows new embedded dialogue. Expected and accepted for developers; tell testers to unzip into a fresh folder.
4. **Tab closed mid-floor.** Run save is written on floor transition only, so progress on the current floor is lost. Same as killing the exe today; note it for testers.
5. **Browser keeps an old `.wasm` cached after a new upload.** Expected: testers get the new build. Check itch.io's behaviour on re-upload in 4.3; if stale, add a `?v=<build>` query to the fetch in `index.html`.

## Progress log

Append-only. Newest at the bottom. One row per session, not per commit.

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-10-08 | — | drafted | Plan written. Only evidence so far: the wasm target compiles (17.7 MB). Nothing has been run in a browser. |
| 2026-10-08 | 1, 2, 3, 5 | code complete; manual checks open | Test-first throughout. **Phase 1:** `gamedata` (8 tests) + 11 fallback tests across `dialogue`, `game`, `leveleditor`, `main`; all watched failing first. **Phase 2:** `storage` (7 tests, desktop backend); the save call-site swaps are a refactor with no desktop behaviour change, so they ran under the existing save tests plus two new `Finalize` characterization tests that passed before and after. The `localStorage` backend has no unit test (needs a browser); it was exercised in headless Chrome instead. **Phase 3:** `appendExitOption` (2 tests, failed first); `quitGame`, `webserve`, `index.html`, `build_web.ps1` have no unit tests. **First boot (3.7):** `build_web.ps1` → 18.6 MB wasm, 6.5 MB zip; headless Chrome via `webserve` reached the main menu, New Game loaded the hand-made hub from embedded data, only console line was "Loaded saved control bindings", no errors. **Phase 5:** workflow rewritten, not run (needs a push). **Baseline:** `go test ./...` had two failures before this work and has the same two after — `levels.TestGenerateForgottenSanctuary` and `levels.TestDoorPlacement`; `levels` is outside the envelope. The CI `test` job will be red until they are fixed, which is why it is a separate job from the two build jobs. |
| 2026-10-08 | 3 | 3.2 closed | Matthew approved adding `draw.game.go` to the envelope. Three tests written and watched failing first (`mainMenuOptions` with and without quit, `Labels` parallel to `Options`). Rebuilt web bundle shows four entries, no "Exit Game" (headless Chrome screenshot). The Windows half of that acceptance criterion (entry present and quits) is covered by the tests but not yet clicked by hand. |
| 2026-10-08 | 1, 4, 6 | partial | **1.11 (part):** fresh exe run alone in an empty folder with `--screenshot` boots and prints no "could not load" line, so lore, flavor and events came from the embedded copy. Dialogue and hub in-game were not looked at by hand (`dialogue.LoadAll` discards its error, so silence proves nothing there; the embed test and fallback tests cover it). **Phase 4 attempt:** tried to drive a run in headless Chrome through the DevTools protocol. Menu clicks and option keys work. In the hub, WASD changed the player's facing but the player did not move, and E at the portal did nothing; the player spawns diagonal to the portal (distance about 2.1 tiles against a 1.75 interaction radius). **Unresolved:** could be synthetic key events, the spawn nook, or a real movement problem on web or on both builds. No desktop comparison was possible from here. Needs a hand check in Chrome and on Windows before anything else in Phase 4. **Static input audit:** default bindings that Chrome may also act on are Tab (inventory), F1 (keybind overlay), F10 (HUD), plus hard-coded F2 (dev menu), F3, F12 (opens DevTools), F5/F6 (level editor; F5 reloads the page). Not tested in a real browser. **6.1, 6.3 done** (README web section; CLAUDE.md status, where-to-look row, build line). 6.2 and 6.4 wait for completion. |
| 2026-10-08 | 4 | 4.3 upload done | Matthew played the local web build ("worked for the most part" — specifics not yet recorded, so the movement/portal open question is not formally closed) and uploaded `dungeoneer-web.zip` to itch.io. The Hosting row of the Phase 4 table (saves inside itch's frame, saves surviving a re-upload) has not been reported back yet. |

## What was NOT changed (intentional)

- **Desktop save location stays the working directory.** Moving it to `os.UserConfigDir()` would orphan existing saves and break the tests that write `meta.json` directly. It becomes a one-function change in `storage_desktop.go` when macOS/Linux builds are wanted.
- **Loader signatures keep taking paths**, not an `fs.FS`. Changing them would rewrite every existing loader test for no behavioural gain.
- **`os.Exit` is now reached only through `quitGame`** (see 3.3). The original plan left the calls alone on the assumption that both menu entries would be hidden; the main-menu one could not be.
- **`TestGetConfigPath` was deleted along with `TestGetConfigPath_Error`.** The plan named only the second; both test the removed `GetConfigPath`.
- **No `gofmt` pass.** `gofmt -l` flags most of the repo (CRLF files), so edits were made in place preserving each file's line endings.
- **Dev-only disk tools are not ported to web** (level/player save-load menus, editor F5/F6, `--screenshot`, dev-menu scenarios). They fail with a visible error line and the game continues.

## Open questions

- **Does the player move and can they enter the portal in Chrome?** See the last Progress log row. First thing to check by hand; if it fails only on web it blocks every Phase 4 row below "Saves".
- **Two pre-existing `levels` test failures** (see Progress log) will make the CI `test` job red. Not caused by this plan and outside its envelope. Recommendation: fix or skip them in their own change before relying on CI.
- **`package.ps1` ships Matthew's own `meta.json` and `controls.json` to testers.** After Phase 1 the zip needs only the exe. Recommendation: drop both so testers start fresh, unless shipping a pre-progressed save is intentional. Blocks nothing; `package.ps1` is outside the envelope until this is answered.
- **Automated itch.io upload (butler) needs an API key stored as a GitHub secret.** Recommendation: upload by hand until builds go out more than weekly. Blocks nothing.
- **Queue position.** Added at #1 because it has no dependencies and unblocks outside testers, but the Void Rift pilot is still the active plan with one manual step left. Matthew decides whether this jumps ahead of `8A-item-sets`.
