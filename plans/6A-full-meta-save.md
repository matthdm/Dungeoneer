---
plan-id: 6A-full-meta-save
status: queued
owner: unassigned
branch: plan/6A-full-meta-save
depends-on: [5C-boss-selection]
last-touched: 2026-04-30
---

# Plan: Phase 6A — Full Meta Save & Milestone Unlocks

## Goal

Extend `MetaSave` to track lifetime statistics and unlock game features as the player progresses across runs. The shop, upgrade station, echo shrine, and lore library are all milestone-gated — they appear in the hub only after the player has earned them. A toast notification fires once when a milestone unlocks.

When done, the hub evolves meaningfully as the player returns across runs. New players see an empty hub; veterans see a fully populated one.

## Scope

**In scope:**
- Extended `MetaSave` fields: `CompletedRuns int`, `TotalDeaths int`, `TotalRemnants int` (lifetime), `LoreUnlocked []string`, `HubState map[string]bool` (which features are visible), `Upgrades map[string]int`.
- `CheckMilestones()` called after each run; evaluates thresholds and sets `HubState` flags.
- Milestone definitions: shop (run 1 complete), upgrades (run 3 complete), echo shrine (first death), lore library (first NPC arc phase complete).
- Toast notification on first milestone unlock (minimal UI — a 3-second text overlay).
- `MetaSave` version field for forward-compatible loading.

**Out of scope (do not change in this plan):**
- The shop, upgrade station, or echo shrine UI themselves (Phase 8B, 7A).
- Lore content (Phase 6C).
- NG+ dialogue (Phase 6B).

## File envelope

**Touched:**
- `src/game/metasave.go` — extend struct, add version field, `CheckMilestones()`, migration for old saves
- `src/game/milestones.go` *(new)* — milestone definitions, threshold constants
- `src/ui/toast.go` *(new)* — 3-second text overlay toast (if not already present)
- `src/game/hub.go` — read `HubState` to conditionally show hub features
- `src/game/draw.game.go` — draw toast if active
- `src/game/game.go` — store active toast, tick it down
- `design-docs/roadmap.md` — mark 6.1, 6.2, 6.3 ✅ on completion
- `CLAUDE.md` — update Phase 6 status block
- `plans/_QUEUE.md` — move to Completed on finish

**Forbidden (do not modify in this plan):**
- `src/coords/`, `src/collision/` — offset plan
- `src/ui/shop.go`, `src/ui/upgrade_station.go`, `src/ui/echo_shrine.go` — don't create these; just read the HubState flag
- `src/dialogue/`, `src/dialogues/` — not relevant here

## Acceptance criteria

- [ ] `MetaSave` has `CompletedRuns`, `TotalDeaths`, `TotalRemnants`, `LoreUnlocked`, `HubState`, `Upgrades`, and `Version` fields.
- [ ] Old saves without the new fields load without crashing (zero-value defaults are correct).
- [ ] After the player's first completed run, `HubState["shop_unlocked"]` becomes true and a toast fires: "A merchant has arrived at the hub."
- [ ] After the player's third completed run, `HubState["upgrades_unlocked"]` becomes true with toast.
- [ ] After the player's first death, `HubState["echo_shrine_unlocked"]` becomes true with toast.
- [ ] After any major NPC reaches phase 1, `HubState["lore_library_unlocked"]` becomes true with toast.
- [ ] Toast renders as a 3-second centered text overlay, then fades. Only one toast at a time.
- [ ] `cd src && go build ./...` passes.

## Phases

### Phase 1: MetaSave extension
- [ ] 1.1 In `src/game/metasave.go`, add fields: `Version int`, `CompletedRuns int`, `TotalDeaths int`, `TotalRemnants int`, `LoreUnlocked []string`, `HubState map[string]bool`, `Upgrades map[string]int`.
- [ ] 1.2 Update `LoadMetaSave()` to initialize missing fields to zero values (handles old saves gracefully).
- [ ] 1.3 Update death and victory paths to increment `TotalDeaths` / `CompletedRuns` / `TotalRemnants` as appropriate.
- [ ] 1.4 `cd src && go build ./...` passes.

### Phase 2: Milestone system
- [ ] 2.1 Create `src/game/milestones.go`. Define milestone constants: `MilestoneShop`, `MilestoneUpgrades`, `MilestoneEchoShrine`, `MilestoneLoreLibrary`.
- [ ] 2.2 Implement `CheckMilestones(meta *MetaSave) []string` — returns a slice of newly-unlocked milestone IDs (those now meeting threshold for the first time). Sets `HubState[id] = true`.
- [ ] 2.3 Call `CheckMilestones()` at run end (death and victory), capture newly unlocked milestones, queue toasts.
- [ ] 2.4 `cd src && go build ./...` passes.

### Phase 3: Toast UI
- [ ] 3.1 Create `src/ui/toast.go`. `Toast` struct: `Message string`, `TTL float64`. `Draw(*ebiten.Image)` renders centered semi-transparent panel with text. `Update(dt float64) bool` ticks TTL down, returns true when expired.
- [ ] 3.2 In `src/game/game.go`, add `ActiveToast *ui.Toast`. In `Update()`, tick it; in `Draw()`, draw it if non-nil.
- [ ] 3.3 Queue milestone toasts via a `[]string` pending list; pop one to `ActiveToast` when `ActiveToast == nil`.
- [ ] 3.4 `cd src && go build ./...` passes.

### Phase 4: Hub state gating
- [ ] 4.1 In `src/game/hub.go`, wrap the shop NPC spawn, upgrade station spawn, echo shrine spawn, and lore library spawn in `if meta.HubState["X_unlocked"]` guards. (These entities don't exist yet — guards are no-ops for now but establish the pattern.)
- [ ] 4.2 `cd src && go build ./...` passes.

### Phase 5: Cleanup
- [ ] 5.1 Mark roadmap rows 6.1, 6.2, 6.3 ✅ in `design-docs/roadmap.md`.
- [ ] 5.2 Update `CLAUDE.md` Phase 6 status block.
- [ ] 5.3 Move this plan to `plans/COMPLETED/6A-full-meta-save.md`.
- [ ] 5.4 Update `plans/_QUEUE.md`.

## Progress log

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-04-30 | — | Drafted | Plan written; not yet started |

## What was NOT changed (intentional)

_None yet._

## Open questions

- **Save file migration.** If `Version` is absent (old save), treat as version 0 and apply all migrations. A simple `if meta.Version < 1 { /* init new fields */ meta.Version = 1 }` chain is sufficient. Affects 1.2.
- **Multiple simultaneous milestones.** If the player completes their 3rd run while also dying (not possible, but edge case), queue both toasts and show them sequentially. Affects 3.3.
