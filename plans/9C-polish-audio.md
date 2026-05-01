---
plan-id: 9C-polish-audio
status: queued
owner: unassigned
branch: plan/9C-polish-audio
depends-on: [9B-polish-visual-effects]
last-touched: 2026-04-30
---

# Plan: Phase 9C — Audio Engine, SFX & Music

## Goal

Add audio to the game: a simple SFX playback engine via Ebiten's audio package, combat and UI sound effects, per-biome ambient music loops with crossfade, and a unique boss music track. Also wire the Master Volume slider from the options menu to actual audio output.

This plan assumes audio asset files (`.ogg` or `.wav`) are provided externally or placeholder-generated; the plan wires the engine and hookup points.

## Scope

**In scope:**
- `audio.Engine` wrapping Ebiten audio context: play SFX (concurrent, capped), play/crossfade music.
- Hookup points for: melee hit, spell cast, spell impact, enemy death, player damage, item pickup, level up.
- Ambient music loop per biome; crossfade on floor change.
- Boss music track, triggered on boss room entry.
- UI SFX: menu open/close, dialogue tick, purchase confirm.
- Master volume from `OptionsData.MasterVolume` applied to audio engine.

**Out of scope (do not change in this plan):**
- Audio asset creation (assets are a separate concern — stubs accepted).
- Per-channel volume control (just master volume for now).
- 3D/positional audio.

## File envelope

**Touched:**
- `src/audio/audio.go` *(new)* — `Engine`, SFX playback, music playback, volume control
- `src/audio/sfx/` *(new directory)* — placeholder or real `.ogg` files for each SFX
- `src/audio/music/` *(new directory)* — placeholder or real music tracks
- `src/game/game.go` — initialize audio engine; pass to subsystems
- `src/game/options.go` — wire `MasterVolume` to `audio.Engine.SetVolume()`
- `src/entities/monster.go` — play hit/death SFX
- `src/entities/player.go` — play damage SFX
- `src/spells/` — play cast and impact SFX
- `src/game/handlers.game.go` — play pickup SFX
- `src/game/boss.game.go` — trigger boss music on boss room entry
- `src/game/hub.go` — trigger biome music on floor entry; crossfade on transition
- `src/ui/` — play UI SFX on open/close/confirm
- `design-docs/roadmap.md` — mark Phase 9 audio rows ✅ on completion
- `CLAUDE.md` — update Phase 9 status block
- `plans/_QUEUE.md` — move to Completed on finish

**Forbidden (do not modify in this plan):**
- `src/coords/`, `src/collision/` — offset plan
- `src/game/particles.go` — visual effects plan scope
- Gameplay logic files

## Acceptance criteria

- [ ] `audio.Engine` can play SFX with at most 8 concurrent voices; new plays beyond cap are silently dropped (no crash).
- [ ] Melee hit, spell cast/impact, enemy death, player damage, item pickup all play distinct SFX (placeholder files accepted).
- [ ] Biome ambient music loops on floor entry; crossfades to new track on floor change (0.5s crossfade).
- [ ] Boss music triggers on boss room entry; ambient music fades out.
- [ ] UI SFX plays on dialogue panel open, typewriter tick (every 3rd character), shop purchase confirm.
- [ ] Master Volume slider in options is wired to `audio.Engine.SetVolume()` and takes effect immediately.
- [ ] Game does not crash if audio files are missing — fallback to silence with a logged warning.
- [ ] `cd src && go build ./...` passes.

## Phases

### Phase 1: Audio engine
- [ ] 1.1 Create `src/audio/audio.go`. `Engine` struct wrapping `ebiten/v2/audio.Context`.
- [ ] 1.2 `PlaySFX(id string)` — load from `sfx/` cache, play via a pool of 8 players. Silently skip if pool is exhausted.
- [ ] 1.3 `PlayMusic(id string)` — load from `music/`, loop, fade in over 0.5s.
- [ ] 1.4 `CrossfadeTo(id string)` — fade out current track while fading in new one over 0.5s.
- [ ] 1.5 `SetVolume(v float64)` — apply to all active players.
- [ ] 1.6 Graceful missing-file handling: log a warning, return without crash.
- [ ] 1.7 Create placeholder 0.1s silence `.ogg` files for each SFX/music ID so the system boots without real assets.
- [ ] 1.8 `cd src && go build ./...` passes.

### Phase 2: Combat and gameplay SFX hookups
- [ ] 2.1 Wire `audio.PlaySFX("hit_melee")` in melee hit path.
- [ ] 2.2 Wire `audio.PlaySFX("spell_cast")` in spell cast path (or per-spell IDs if variety is desired).
- [ ] 2.3 Wire `audio.PlaySFX("spell_impact")` in spell hit callback.
- [ ] 2.4 Wire `audio.PlaySFX("enemy_death")` in monster death path.
- [ ] 2.5 Wire `audio.PlaySFX("player_damage")` in player hit path.
- [ ] 2.6 Wire `audio.PlaySFX("item_pickup")` in pickup handler.
- [ ] 2.7 `cd src && go build ./...` passes.

### Phase 3: Music hookups
- [ ] 3.1 In `src/game/hub.go` floor entry, call `audio.CrossfadeTo(biome.MusicID)` (add `MusicID string` to `BiomeConfig`).
- [ ] 3.2 In `src/game/boss.game.go` boss room entry, call `audio.CrossfadeTo("boss_theme")`.
- [ ] 3.3 Hub world plays `"hub_ambient"` on load.
- [ ] 3.4 `cd src && go build ./...` passes.

### Phase 4: UI SFX and volume wiring
- [ ] 4.1 Play `"ui_open"` / `"ui_close"` in dialogue panel and shop open/close paths.
- [ ] 4.2 Play `"ui_tick"` every 3rd character in the dialogue typewriter renderer.
- [ ] 4.3 Play `"ui_purchase"` on shop and upgrade station confirm.
- [ ] 4.4 In `src/game/options.go` `Apply()`, call `audio.Engine.SetVolume(o.MasterVolume)`.
- [ ] 4.5 `cd src && go build ./...` passes.

### Phase 5: Cleanup
- [ ] 5.1 Mark Phase 9 audio roadmap rows ✅ in `design-docs/roadmap.md`.
- [ ] 5.2 Update `CLAUDE.md` Phase 9 status block.
- [ ] 5.3 Move this plan to `plans/COMPLETED/9C-polish-audio.md`.
- [ ] 5.4 Update `plans/_QUEUE.md`.

## Progress log

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-04-30 | — | Drafted | Plan written; not yet started |

## What was NOT changed (intentional)

_None yet._

## Open questions

- **Audio file format.** Ebiten supports `.ogg` and `.wav` natively. `.ogg` is preferred for music (smaller), `.wav` for SFX (lower latency). Confirm build constraints for the target platform. Affects 1.7.
- **BiomeConfig.MusicID.** Add this field to `BiomeConfig` in `game/biome.go`. Each biome needs a music ID string. Neutral fallback: `"dungeon_ambient"` for all biomes until individual tracks exist. Affects 3.1.
