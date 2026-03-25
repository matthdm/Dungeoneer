---
name: dungeoneer-expert
description: "Use when: asking about Dungeoneer project roadmap, current bugs, code architecture, controls, item database, NPC systems, saves/progression, development blockers, or authoritative project state. Provides citation-based answers from design docs and source code."
---

# Dungeoneer Expert Skill

Your one-stop shop for all Dungeoneer project knowledge: roadmap phases, active bugs, gameplay systems, code internals, and live development state.

## Core Domains

| Domain | Example Query | Source |
|--------|--------------|--------|
| **Roadmap & Phases** | "What phase are we in?" | `design-docs/roadmap.md` |
| **Active Bugs** | "What are the current bugs?" | `BUG_TRACKER.md` (skill local) + README.md Known Issues |
| **Gameplay Systems** | "How does the FOV system work?" | `design-docs/*.md` + `src/fov/` |
| **Controls & Input** | "What hotkey loads a level?" | `README.md` Hotkeys + `src/controls/` |
| **Item Database** | "What does [item-name] do?" | `src/items/` + `src/images/items_structured_effects.json` |
| **Architecture** | "How are monsters spawned?" | `src/entities/` + `src/game/` + `src/levels/` |
| **Save/Progression** | "How is meta-progression saved?" | `src/game/metasave.go` + `src/meta.json` |
| **NPCs & Quest** | "What NPC dialogue trees exist?" | `design-docs/dialogue-system.md` + quest/dialogue data |
| **Level Editor** | "How to access the tile editor?" | `README.md` Hotkeys + `src/leveleditor/` |
| **Test Cases** | "What tests should I run for Phase 3?" | `design-docs/test-cases.md` |
| **Blockers** | "What's preventing next phase?" | `BUG_TRACKER.md` (skill local) + roadmap dependencies |

---

## Trigger Phrases

Use any of these to invoke this skill in chat:

- `/dungeoneer` (generic expert query)
- `/dungeoneer status` (current roadmap phase + blockers)
- `/dungeoneer bugs` (list tracked bugs)
- `/dungeoneer sync` (refresh internal context)
- `/dungeoneer item <name>` (item lookup)
- `/dungeoneer controls` (hotkeys reference)
- `/dungeoneer architecture` (code structure Q&A)
- `/dungeoneer tests` (list test cases for current or specified phase)
- `/dungeoneer tests phase3` (list Phase 3 test cases)

---

## Answer Format: Citations Required

When answering, always cite sources in this format:

**For code**: `src/path/file.go` with line range if specific
**For design**: `design-docs/system-name.md` with section
**For project state**: `README.md` section or `.github/BUG_TRACKER.md`
**For data**: `src/data/file.json` with key path if applicable

Example answer:
```
The FOV system uses raycasting with corner-casting optimization (📍 src/fov/fov.go, src/fov/walls.go).
The rendering pipeline is in src/fov/render.go and integrated into the game loop via 
📍 src/game/draw.game.go (FOV raycasting is triggered on entity movement).
```

---

## Knowledge Base Structure

The skill draws from these canonical sources:

### Roadmap & Planning
- `design-docs/roadmap.md` (7 phases, dependencies, scope)
- `README.md` (current WIP dev notes, feature checklist)
- `BUG_TRACKER.md` (local to skill; tracked bugs, attempts, blockers)
- `PROJECT_SNAPSHOT.json` (local to skill; cached metadata)

### Systems Documentation
- `design-docs/dialogue-system.md`
- `design-docs/inventory-system.md`
- `design-docs/quest-system.md`
- `design-docs/npc-system.md`
- `design-docs/enemy-placement.md`
- `design-docs/room-tagging.md`
- `design-docs/spawn-placement.md`
- `design-docs/living-dungeon-ai.md`
- `design-docs/meta-progression.md`
- `design-docs/player-stats.md`
- `design-docs/layered-world-system.md`

### Quality
- `design-docs/test-cases.md` (manual test scenarios by phase, blocking/visual/regression)

### Code Structure
- `src/main.go` (entry point, game loop bootstrap)
- `src/game/game.go` (Game struct, state machine)
- `src/game/progression.game.go` (EXP, leveling)
- `src/game/runstate.go` (per-run state)
- `src/game/metasave.go` (cross-run persistence)
- `src/entities/` (player, monsters, items, NPCs)
- `src/levels/` (generation, loading, layout)
- `src/fov/` (visibility, raycasting, rendering)
- `src/items/` (registry, effects, types)
- `src/controls/` (input handling)
- `src/ui/` (menus, HUD, dialogue)

### Data Files
- `src/meta.json` (meta-progression snapshot)
- `src/test_level.json` (dev test level definition)
- `src/images/items_structured_effects.json` (item catalog)
- `src/levels/hub.json` (hub world layout)

---

## Context Auto-Refresh Workflow

When project state changes (code edits, design doc updates, bug fixes), trigger:

```
/dungeoneer sync
```

This prompts the skill to:
1. **Re-scan roadmap**: Check `design-docs/roadmap.md` for phase/milestone updates
2. **Re-read bugs**: Parse `README.md` Known Issues + `.github/BUG_TRACKER.md`
3. **Index items**: Regenerate item catalog from `src/items/registry.go` + `items_structured_effects.json`
4. **Validate controls**: Refresh hotkey binding from `README.md` + `src/controls/controls.go`
5. **Audit code structure**: Confirm latest `src/game/game.go` + entity/system files exist and are reachable
6. **Update README**: If state has drifted from design docs, flag inconsistencies and prepare README patch

**Automated trigger** (optional Git hook setup):
- Add `.github/hooks/post-commit.sh` to run `dungeoneer sync` after commits touching design-docs or src/game/
- Add `Makefile` target: `make dungeoneer-sync` → runs context refresh

---

## Supported Query Patterns

| Pattern | Example | Behavior |
|---------|---------|----------|
| System deep-dive | "How does pathfinding work?" | Cite design doc + code locations, explain flow |
| Bug diagnosis | "Why does right-click spam cause stutter?" | Cite `BUG_TRACKER.md`, link movement/input code |
| Roadmap timeline | "What's blocking Phase 3?" | Cite roadmap dependencies, list pre-requisites |
| Backlog / optimization | "What's in the backlog?" | Cite `design-docs/roadmap.md` Backlog section, list items with severity |
| Test cases | "What tests for Phase 3?" | Cite `design-docs/test-cases.md`, filter by phase/tag |
| Item lookup | "What does Sword of Chaos do?" | Cite `items_structured_effects.json`, stat block, effects |
| Control reference | "How do I spawn a monster?" | Cite `README.md` Hotkeys, link `src/leveleditor/` or spawn logic |
| Architecture query | "Where is NPC dialogue handled?" | Cite `src/ui/`, `dialogue/` files, explain state flow |
| Feature status | "Is save/load implemented?" | Cite `src/game/metasave.go`, phase in roadmap, what's done |
| Blocker analysis | "Why can't we start Phase 4 yet?" | Link to blocking tasks in roadmap, cite bug #s, PRs, or code TBD sections |

---

## README Maintenance

The skill **actively scans the entire repository** to keep `README.md` synchronized:

- **Feature Checklist**: Cross-checked against actual code implementation status (files exist? complete? tested?)
- **Known Issues**: Auto-pulled from `BUG_TRACKER.md` (local skill directory) with full attempt history
- **Hotkeys**: Scanned from `src/controls/controls.go` and `src/controls/config.json` to validate README accuracy
- **Dev Notes**: Updated by parsing `design-docs/roadmap.md` phases + tracking current progress
- **Architecture**: Validates code module structure against actual `src/` directory
- **Item Database**: Indexed from `src/items/registry.go` and `items_structured_effects.json`

When you commit code changes (bug fix, feature add, refactor), run:
```
/dungeoneer sync
```

The skill performs a **comprehensive repo scan** (line-by-line where needed) and:
1. Detects code changes vs. README state
2. Identifies feature status drift
3. Flags new bugs or old attempts
4. Updates BUG_TRACKER.md + PROJECT_SNAPSHOT.json
5. Generates README patch with evidence citations
6. Shows user the diff; they approve & commit

---

## Asset Organization

```
.github/skills/dungeoneer-expert/
├── SKILL.md                         (this file)
├── dungeoneer-expert.instructions.md (context injection + repo scanning)
├── BUG_TRACKER.md                   (tracked bugs + attempts + blockers)
├── PROJECT_SNAPSHOT.json            (cached metadata + status indices)
├── REPO_SCAN_RESULTS.json           (auto-generated from full repo scan)
└── README_SYNC_CACHE.md             (tracking README updates)
```

### BUG_TRACKER.md Format

Each bug entry includes:
- ID + title
- Description
- Current status (Open/In Progress/Blocked)
- Attempts tried + why they failed
- Blocker (if any)
- Relevant code links
- Reproduction steps

Example:
```markdown
## Bug #1: Right-Click Movement Stutter

**Status**: Open / Priority: HIGH

**Description**: Spamming RMB causes player to stutter in place or revert to earlier tile.

**Attempts**:
1. Tried input debouncing in movement/controller.movement.go — failed because...
2. Tried pathfinding queue flush — caused...

**Blocker**: Needs investigation into movement state machine in src/entities/player_io.go

**Code**: 📍 src/entities/player_io.go (RMB handler), src/movement/controller.movement.go (pathfinding)
```

### PROJECT_SNAPSHOT.json Format

```json
{
  "phase": "Phase 1-2 (In Progress)",
  "last_sync": "2026-03-22T12:34:56Z",
  "item_count": 42,
  "known_bugs": 2,
  "roadmap_blockers": ["Save/Load Completion", "Run Loop Stability"],
  "controls_count": 15,
  "design_docs": 10,
  "code_modules": ["entities", "fov", "levels", "game", "ui", "items", ...]
}
```

---

## Developer Notes

- **Keep `.github/BUG_TRACKER.md` updated** as you work through bugs. Reference attempts, blockers, and code paths.
- **Run `/dungeoneer sync` after major commits** to keep internal context fresh.
- **README mirrors state**: Feature checklist ↔ roadmap phase ↔ code status.
- **Citations enable trust**: Always cite exact files/sections; vague answers indicate stale context (re-sync).
- **Blocker tracking**: Use roadmap dependencies + bug blocker field to answer "what's stopping phase N?"
- **Item database**: Dynamically indexed from `items_structured_effects.json` on each sync; easy item lookups.

---

## Quick Reference: File Paths

| Purpose | Path |
|---------|------|
| Roadmap | `design-docs/roadmap.md` |
| Bugs | `.github/skills/dungeoneer-expert/BUG_TRACKER.md` |
| Project state | `README.md` + `.github/skills/dungeoneer-expert/PROJECT_SNAPSHOT.json` |
| Repo state | `.github/skills/dungeoneer-expert/REPO_SCAN_RESULTS.json` |
| FOV/Rendering | `src/fov/` |
| Movement/Entities | `src/entities/` + `src/movement/` |
| Game loop | `src/game/game.go` + `src/game/runstate.go` |
| UI/Menus | `src/ui/` |
| Items | `src/items/` + `src/images/items_structured_effects.json` |
| Level editor | `src/leveleditor/` |
| Controls | `src/controls/` + `README.md` Hotkeys |
| Saves | `src/game/metasave.go` + `src/meta.json` |
