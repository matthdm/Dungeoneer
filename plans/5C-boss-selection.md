---
plan-id: 5C-boss-selection
status: active
owner: claude
branch: main
depends-on: [5B-varn-arc]
last-touched: 2026-05-30
---

# Plan: Phase 5C — Boss Selection Engine & Varn Boss Fight

## Goal

Replace the current single-boss setup with a dynamic boss selection engine that picks the highest-phase major NPC as the final boss, then implement Varn's full three-phase boss fight: chain whip (melee), chain eruptions (AoE), and unchained frenzy (fast melee). Pre-fight and post-fight dialogue trees vary based on trust level and betrayal status.

When done, players who helped Varn through his arc fight a fully realized, story-integrated boss. Players who skipped or betrayed Varn still face a boss — either a degraded Varn fight or the existing generic fallback.

## Scope

**In scope:**
- `BossSelectionEngine` in a new `game/boss_selection.go`: evaluates all major NPC quest flags, selects the highest-phase eligible NPC as boss, falls back to the existing generic boss.
- Varn boss entity in `entities/bosses/varn.go` with 3 phases, phase-specific attack patterns.
- Boss arena theming hook: Varn's arena uses chain tile decorations (even if they are a recolor/overlay until art ships).
- Pre-fight dialogue tree (`dialogues/varn_boss_pre.json`) with trust/betrayal branching.
- Post-fight dialogue tree (`dialogues/varn_boss_post.json`) with first-defeat vs. NG+ branching.

**Out of scope (do not change in this plan):**
- Other NPC boss forms (Seris, Mira, Kael — Phase 10).
- NG+ post-defeat dialogue expansions (Phase 6B).
- Hub NPC appearances (Phase 5D).

## File envelope

**Touched:**
- `src/game/boss_selection.go` *(new)* — selection engine
- `src/entities/bosses/` *(new directory)* — `varn.go`
- `src/entities/boss.go` — hook for custom boss types
- `src/game/boss.game.go` — integrate selection engine, pre/post dialogue triggers
- `src/levels/generate64.go` — Varn arena decoration pass (chain tile overlay)
- `src/dialogues/varn_boss_pre.json` *(new)*
- `src/dialogues/varn_boss_post.json` *(new)*
- `design-docs/roadmap.md` — mark 5.10–5.14 ✅ on completion
- `CLAUDE.md` — update Phase 5 status block
- `plans/_QUEUE.md` — move to Completed on finish

**Forbidden (do not modify in this plan):**
- `src/coords/`, `src/collision/` — offset plan
- `src/dialogue/loader.go`, `src/dialogue/types.go` — phase tracker plan scope
- `src/dialogues/varn_phase*.json` — Varn arc plan scope
- `src/game/npc_data.go` — Varn arc plan scope

## Acceptance criteria

- [ ] `BossSelectionEngine.Select(RunState) BossType` returns `BossVarn` when `varn_p3_done >= 1`, otherwise returns `BossGeneric`.
- [ ] Varn boss has 3 phases with distinct attack patterns: chain whip melee (phase 1), chain eruptions AoE (phase 2, triggered at 66% HP), unchained frenzy fast melee (phase 3, triggered at 33% HP).
- [ ] Phase transitions change Varn's sprite and trigger a visual effect (reuse existing boss phase-change logic).
- [ ] Pre-fight dialogue plays before combat. Dialogue text differs for `trust_gte 50` (respected ally) vs. `varn_betrayed = 1` (cold and furious) vs. default (confused recognition).
- [ ] Post-fight dialogue plays after defeat. `NPCMeta["varn"].DefeatCount` increments. Text differs on first defeat vs. subsequent.
- [ ] Varn's arena has a chain decoration pass (tile overlay or sprite) — placeholder accepted if no art asset exists yet.
- [ ] `cd src && go build ./...` passes.

## Phases

### Phase 1: Boss selection engine
- [x] 1.1 Create `src/game/boss_selection.go`. Define `BossType` enum: `BossGeneric`, `BossVarn` (expandable for future NPCs).
- [x] 1.2 Implement `SelectBoss(rs RunState) BossType`: check `rs.QuestFlags["varn_p3_done"] >= 1` → `BossVarn`; else `BossGeneric`.
- [x] 1.3 In `src/game/boss.game.go`, call `SelectBoss()` when the boss room is entered; spawn the appropriate boss entity.
- [x] 1.4 `cd src && go build ./...` passes.

### Phase 2: Varn boss entity
- [ ] 2.1 Create `src/entities/bosses/` directory and `varn.go`. Define `VarnBoss` struct embedding or referencing the existing `Boss` struct.
- [ ] 2.2 Phase 1 attack (0%–66% HP): chain whip — melee swing with extended range (2.5 tiles), 3-hit combo, 0.8s cooldown.
- [ ] 2.3 Phase 2 attack (66% HP trigger): chain eruptions — 3 ground-targeted AoE circles in a spread pattern, 1.5s warning before detonation, moderate damage.
- [ ] 2.4 Phase 3 attack (33% HP trigger): unchained frenzy — move speed +50%, attack speed ×2, shorter range but rapid hits.
- [ ] 2.5 Phase transitions: sprite swap, flash effect (reuse existing boss phase-change code from `entities/boss.go` or `game/boss.game.go`).
- [ ] 2.6 `cd src && go build ./...` passes.

### Phase 3: Arena theming
- [ ] 3.1 In `src/levels/generate64.go`, after boss room generation, if `BossType == BossVarn`, run a decoration pass that places chain overlay tiles on the perimeter walls. Use an existing tile ID or a placeholder tint if no chain tile exists — document in "What was NOT changed."
- [ ] 3.2 `cd src && go build ./...` passes.

### Phase 4: Pre- and post-fight dialogue
- [ ] 4.1 Write `src/dialogues/varn_boss_pre.json`. Three root nodes: `trust_gte 50` (ally recognition, sad), `varn_betrayed = 1` (fury, cold), default (confused, resolute). Each is a 2–3 line monologue with no player response — combat begins on dialogue close.
- [ ] 4.2 Write `src/dialogues/varn_boss_post.json`. Two root nodes: `defeat_count = 0` (first defeat — raw, emotional), `defeat_count >= 1` (subsequent — wearier, more aware). Defeat count read from `NPCMeta["varn"].DefeatCount`.
- [ ] 4.3 In `src/game/boss.game.go`: on boss room entry, trigger pre-fight dialogue before combat begins; on boss death, increment `NPCMeta["varn"].DefeatCount`, trigger post-fight dialogue.
- [ ] 4.4 `cd src && go build ./...` passes.

### Phase 5: Cleanup
- [ ] 5.1 Mark roadmap rows 5.10–5.14 ✅ in `design-docs/roadmap.md`.
- [ ] 5.2 Update `CLAUDE.md` Phase 5 status block.
- [ ] 5.3 Move this plan to `plans/COMPLETED/5C-boss-selection.md`.
- [ ] 5.4 Update `plans/_QUEUE.md`.

## Progress log

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-04-30 | — | Drafted | Plan written; not yet started |
| 2026-05-30 | 1 | Complete | `boss_selection.go` created with `BossType`/`SelectBoss`; `setupBossFloor` now calls `SelectBoss(g.RunState)` instead of inline `varn_phase >= 3` check; build passes. |

## What was NOT changed (intentional)

_None yet._

## Open questions

- **Chain eruption projectile vs. zone.** AoE warning circles — implement as a timed `DamageZone` entity (simpler) or as a projectile with a zero-velocity and delayed detonation? Recommendation: `DamageZone` with a `Lifetime` and `Armed bool` (armed after 1.5s). Affects 2.3.
- **Chain tile asset.** If no chain tile exists in the spritesheet, use the existing wall tile with a green/gray tint via Ebiten's `ColorScale`. Document the art gap. Affects 3.1.
