---
plan-id: 6C-lore-system
status: queued
owner: unassigned
branch: plan/6C-lore-system
depends-on: [6B-ng-plus-dialogue]
last-touched: 2026-04-30
---

# Plan: Phase 6C — Lore System

## Goal

Build the lore registry, unlock action, and lore library UI. Lore entries are unlocked through dialogue, exploration, and high-trust NPC conversations. The lore library in the hub is a scrollable, categorized panel the player can browse between runs. Locked entries appear as "???" until earned.

When done, the game has a readable in-world lore canon that rewards engaged play. Approximately 15–20 entries covering Abaddon's nature, dungeon cosmology, NPC backstories, and cryptic fragments.

## Scope

**In scope:**
- `LoreDef` struct: ID, title, category (character / cosmology / history / fragment), body text.
- Lore registry loaded from `data/lore.json`.
- `unlock_lore "id"` dialogue action adds to `MetaSave.LoreUnlocked`.
- Lore library UI: scrollable panel, 4-tab category filter, locked entries shown as "???" with the entry title visible.
- 15–20 lore entries as content.
- Hook a few lore unlocks into existing Varn dialogue trees.

**Out of scope (do not change in this plan):**
- Abaddon-specific lore (Phase 10B).
- Alignment system (Phase 10B).
- Audio or particle effects on unlock.

## File envelope

**Touched:**
- `src/game/lore.go` *(new)* — `LoreDef`, registry loader, `IsUnlocked()` helper
- `src/data/lore.json` *(new)* — 15–20 lore entries
- `src/dialogue/types.go` — add `unlock_lore` action type
- `src/game/npc.game.go` — handle `unlock_lore` action (add to `MetaSave.LoreUnlocked`)
- `src/ui/lore_library.go` *(new)* — scrollable lore panel with tabs
- `src/game/game.go` — register lore library state, store open/close
- `src/game/draw.game.go` — draw dispatch for lore library
- `src/game/hub.go` — spawn lore library interaction point (a book/pedestal entity) when `HubState["lore_library_unlocked"]`
- `design-docs/roadmap.md` — mark 6.8–6.11 ✅ on completion
- `CLAUDE.md` — update Phase 6 status block to "complete"
- `plans/_QUEUE.md` — move to Completed on finish

**Forbidden (do not modify in this plan):**
- `src/coords/`, `src/collision/` — offset plan
- `src/dialogues/varn_phase*.json` — only add `unlock_lore` actions; do not restructure trees
- `src/entities/` — not relevant

## Acceptance criteria

- [ ] `LoreDef` struct exists with ID, Title, Category, Body fields.
- [ ] `data/lore.json` contains 15–20 entries across all 4 categories.
- [ ] `unlock_lore "lore_id"` dialogue action adds the ID to `MetaSave.LoreUnlocked` (idempotent).
- [ ] Lore library UI opens from a hub interaction point (book/pedestal), shows categorized tabs (Character / Cosmology / History / Fragment).
- [ ] Unlocked entries show full title and body. Locked entries show title as "???" and body as "Unlock through exploration or NPC dialogue."
- [ ] At least 3 lore entries are wired into existing Varn NG+ dialogue trees as `unlock_lore` actions.
- [ ] `cd src && go build ./...` passes.

## Phases

### Phase 1: Lore registry
- [ ] 1.1 Create `src/game/lore.go`. Define `LoreCategory` enum and `LoreDef` struct.
- [ ] 1.2 `LoadLoreRegistry(path string) ([]LoreDef, error)` — reads `data/lore.json`.
- [ ] 1.3 `IsUnlocked(meta *MetaSave, id string) bool` — checks `meta.LoreUnlocked`.
- [ ] 1.4 Create `src/data/lore.json` with 15–20 entries. Categories and suggested entries:
  - **Character (5):** Varn's past as a dungeon warden, Abaddon's origin (cryptic), the Hollow Monk's true identity, the Weeping Shade's guilt, the Mad Scholar's discovery.
  - **Cosmology (4):** What the dungeon is (a purgatorial memory space), how Remnants work spiritually, the loop mechanism, Abaddon's relationship to the loop.
  - **History (4):** The dungeon before Abaddon, the first player-character (implied), the chain war, the sealing.
  - **Fragment (4):** Cryptic one-paragraph pieces with no clear attribution — deliberately mysterious.
- [ ] 1.5 `cd src && go build ./...` passes.

### Phase 2: Unlock action
- [ ] 2.1 In `src/dialogue/types.go`, add action type `"unlock_lore"` with field `LoreID string`.
- [ ] 2.2 In `src/game/npc.game.go` action executor, handle `unlock_lore`: append `LoreID` to `MetaSave.LoreUnlocked` if not already present. Optionally queue a toast ("Lore unlocked: [Title]").
- [ ] 2.3 Add `unlock_lore` actions to `varn_ng2.json` (1 entry on trust >= 60 branch) and `varn_ng3.json` (1 entry on trust >= 80 branch).
- [ ] 2.4 `cd src && go build ./...` passes.

### Phase 3: Lore library UI
- [ ] 3.1 Create `src/ui/lore_library.go`. `LoreLibrary` struct: `Registry []LoreDef`, `Unlocked []string`, `ActiveCategory LoreCategory`, `ScrollOffset int`.
- [ ] 3.2 `Draw(*ebiten.Image)`: tab row at top (4 categories), scrollable entry list below. Unlocked entries: bold title + body text. Locked: grayed "???" title + "Unlock through..." hint.
- [ ] 3.3 `Update()`: tab click switches category, scroll wheel / arrow keys move `ScrollOffset`, Escape closes.
- [ ] 3.4 In `src/game/hub.go`, spawn a book/pedestal entity near the hub NPC quarter when `HubState["lore_library_unlocked"]`. Interaction opens the lore library.
- [ ] 3.5 In `src/game/game.go` and `src/game/draw.game.go`, register the lore library overlay state.
- [ ] 3.6 `cd src && go build ./...` passes.

### Phase 4: Cleanup
- [ ] 4.1 Mark roadmap rows 6.8–6.11 ✅ in `design-docs/roadmap.md`.
- [ ] 4.2 Update `CLAUDE.md` Phase 6 status block to "🟢 COMPLETE".
- [ ] 4.3 Move this plan to `plans/COMPLETED/6C-lore-system.md`.
- [ ] 4.4 Update `plans/_QUEUE.md`.

## Progress log

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-04-30 | — | Drafted | Plan written; not yet started |

## What was NOT changed (intentional)

_None yet._

## Open questions

- **Pedestal/book sprite.** If no suitable sprite exists in the spritesheet, use an existing chest sprite with a different tint. Document the gap. Affects 3.4.
- **Lore unlock toast vs. silent.** Recommendation: show a brief toast ("Lore Unlocked") but not the full entry text in the toast — the player should go to the library to read it. Affects 2.2.
