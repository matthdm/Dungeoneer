---
name: dungeoneer-expert-context
description: "Internal instructions for dungeoneer-expert skill. Auto-loaded when skill is invoked. Manages context injection, comprehensive repo scanning, and README synchronization."
applyTo: "**"
---

# Dungeoneer Expert Skill: Context Injection & Repo Scanning

**These rules auto-load whenever the `dungeoneer-expert` skill is invoked.**

---

## Context Loading Behavior

When you see a query matching the dungeoneer skill triggers (roadmap, bugs, item lookup, controls, architecture), immediately:

1. **Load these files into working context** (don't just reference them):
   - `c:\Github\Dungeoneer\README.md` (feature status, hotkeys, bugs)
   - `c:\Github\Dungeoneer\design-docs\roadmap.md` (phases and dependencies)
   - `.github/skills/dungeoneer-expert/BUG_TRACKER.md` (active bugs, attempts, blockers — LOCAL SKILL DIRECTORY)
   - `.github/skills/dungeoneer-expert/PROJECT_SNAPSHOT.json` (cached scan results)

2. **For item lookups**, also load:
   - `src/images/items_structured_effects.json` (item catalog with effects)
   - `src/items/registry.go` (item definitions)

3. **For architecture queries**, be ready to scan:
   - `src/game/game.go` (Game struct, main state machine)
   - `src/entities/*` (entity types: player, monster, NPC, etc.)
   - `src/fov/fov.go` + `src/fov/render.go` (render pipeline)

4. **Cross-reference and cite** with line numbers or sections whenever possible.

---

## Comprehensive Repository Scanning

**CRITICAL**: When `/dungeoneer sync` is invoked, perform a **full repository scan** (not just quick checks).

### Scan Phase 1: Code Module Indexing

Scan `src/` directory exhaustively:

1. **src/game/** — Parse every `*.go` file:
   - `game.go`: Identify Game struct, state fields, main Update/Draw methods
   - `progression.game.go`: EXP reward logic, player leveling
   - `runstate.go`: RunState struct, per-run tracking fields
   - `metasave.go`: Save/Load logic; check if complete or TODO
   - `deathscreen.go`, `debug.game.go`, other features: index status
   - Look for `// TODO`, `// FIXME`, `// BUG` comments → add to BUG_TRACKER

2. **src/entities/** — Index all entity types:
   - `player.go`: Player struct, HP, stats, level
   - `monster.go`, `ambush.monster.go`, `roaming.monster.go`: Monster AI types
   - `npc.go`: Does it exist? If not, Phase 3 not started
   - `itemdrop.go`: Loot system
   - `inventory_ops.go`: Inventory API surface
   - Look for incomplete or stubbed functions

3. **src/items/** — Item registry:
   - `registry.go`: Load all items; count total
   - `types.go`: Item struct definition
   - `load.go`: How items are loaded from JSON
   - Cross-check against `src/images/items_structured_effects.json`: verify completeness

4. **src/levels/** — Level generation:
   - `generate64.go`: Procedural logic; check for floor templates, biome assignment
   - `maze.go`, `throats.go`: Generation algorithms
   - `pathutil.go`: Level pathfinding utilities
   - Check for TODO sections indicating incomplete phases

5. **src/ui/** — UI system:
   - Scan all `*.go` files in `ui/` directory
   - Identify which menus are implemented: `devmenu.go`, `inventory.go`, `heropanel.go`, etc.
   - Check for TODOs or incomplete features

6. **src/controls/** — Input system:
   - `controls.go`: Read hotkey bindings
   - `config.go`: Parse control configuration
   - Compare actual hotkeys in code vs. README.md hotkey table → flag drift

### Scan Phase 2: Design Document Validation

Read every design doc in `design-docs/`:

1. **roadmap.md**: Parse all 7 phases
   - Identify which phase tasks are complete (code evidence)
   - Which are in-progress (partial implementation)
   - Which are not started
   - Compare against PROJECT_SNAPSHOT.json `phase_completion` field
   - Flag misalignments

2. **system-*.md files** (dialogue, inventory, quest, npc, etc.):
   - Scan for references to code files
   - Check if those files exist and are implemented
   - Flag if design is ahead of code (design describes feature not implemented)
   - Flag if code is ahead of design (code implements feature not in design doc)

### Scan Phase 3: Feature Status Cross-Check

For each feature in README.md Feature Checklist:

1. **Check if feature code exists**:
   - Map feature name to source file(s)
   - Does the file exist? Is it >100 lines? Is it >1000 lines?
   - Size heuristic: <100 lines = stub; 100-500 = basic; 500+ = substantial

2. **Check for TODO/FIXME comments**:
   - Run grep across src/ for // TODO (count)
   - Add uncaptured TODOs to BUG_TRACKER

3. **Compare code status vs. README checkbox**:
   - If README says ✓ Done but code has TODOs → flag drift
   - If README says [ ] Not started but code exists → flag drift
   - Generate patch with corrections

### Scan Phase 4: Bug Tracking

For each bug in BUG_TRACKER.md:

1. **Check code references**:
   - Are the cited files still at the cited line ranges?
   - Have they changed? (If yes, update citations)

2. **Look for new bugs**:
   - Scan for // BUG comments in code
   - Scan for TODO BLOCKER / TODO CRITICAL
   - If found and not in BUG_TRACKER, add new entry

3. **Check if bugs are fixed**:
   - For "Open" bugs, check if referenced code has been edited recently
   - If heavily refactored since bug report, flag for re-testing

### Scan Phase 5: Data File Inventory

1. **src/items/registry.go** + `src/images/items_structured_effects.json`:
   - Count total items
   - Check if counts match
   - Validate item effect descriptions are present

2. **src/meta.json**:
   - Parse to find meta-progression state
   - Check if save/load logic in metasave.go matches schema

3. **src/controls.json**:
   - Parse and compare to README hotkey table
   - Flag any new hotkeys or removed hotkeys

4. **Level JSON files** (test_level.json, hub.json):
   - Verify they exist and are parseable
   - Check if hub.json NPC counts match design doc expectations

### Scan Phase 6: Git & Temporal Signals

1. **Check git log** (if available):
   - Recent commits touching src/game/ or src/entities/ → code is active
   - Date of last major commit → indicates stale modules

2. **File modification times**:
   - Recently modified files likely have active bugs or features
   - Stale files may need cleanup or archiving

---

## README Synchronization Workflow

When `/dungeoneer sync` completes scanning, **generate README patch** with this algorithm:

### Step 1: Feature Checklist Updates

For each feature in README.md checklist:

```
If code_status != readme_status:
  Generate patch updating checkbox:
    ✓ (Done), ~ (WIP), ⚠ (Planned), or empty (not started)
  Include evidence citation in patch comment: "Evidence: src/file.go lines X-Y shows implementation"
```

### Step 2: Known Issues Section Updates

```
For each bug in BUG_TRACKER.md:
  If status == "Open" and priority == "HIGH":
    Add/update entry in README Known Issues
    Format: "- [ ] **Bug #N**: [Title] — [Impact] (Evidence: BUG_TRACKER.md)"
  
  If status == "Fixed" and still in README:
    Remove from Known Issues
    Move to fixed section: "- [x] **Bug #N**: [Title] (Fixed)"
```

### Step 3: Architecture Summary

```
Scan src/ directory structure.
For each module with substantial code (>500 lines):
  Add/update entry in README "Where to look first":
    "- src/module/ — [brief description] (~N lines)"
```

### Step 4: Hotkeys Validation

```
Parse src/controls/controls.go for hotkey bindings.
Compare to README Hotkeys table.

If mismatch:
  Generate patch updating README table with new/changed hotkeys
  Include citation: "Evidence: src/controls/controls.go line X"
```

### Step 5: Phase & Blocker Summary

```
Read design-docs/roadmap.md latest phase.
Cross-check against code implementation.

If current_phase_progress has changed:
  Update README "Developer Notes" with new focus
  Update phase % in PROJECT_SNAPSHOT.json
  If phase is complete, mark in README milestone
```

### Step 6: Generate Patch Diff

Output final patch as markdown:

```
## README Sync Patch (Generated by /dungeoneer sync)

**Last Scan**: 2026-03-22 HH:MM:SS

### Changes
- [ ] Feature Checklist: X items updated
- [ ] Known Issues: Y bugs added/removed
- [ ] Hotkeys: Z entries validated
- [ ] Phase Progress: Updated to Phase N (M%)

### Evidence

**Feature Checklist Changes**:
```
- src/game/metasave.go: 850 lines, substantial save/load impl → marking Phase 1.7 as [~] WIP
- src/features/feature_x.go: Does NOT exist → marking as [ ] Not started
```

**Hotkey Changes**:
```
- F key: src/controls/controls.go line 234 — not in README, adding "F: Debug overlay"
```

**Phase Update**:
```
- Phase 1 was 80%, Phase 1.7 (Save/Load) now 60% complete (src/game/metasave.go)
```

**Bugs Summary**:
```
- Bug #1: Still Open, HIGH priority
- Bug #2: Fixed ✓ (remove from Known Issues)
- Bug #3: Fixed ✓ (remove from Known Issues)
- Bug #4: Fixed ✓ (remove from Known Issues)
- New bug found: src/movement/controller.movement.go line 145 has TODO BLOCKER
```

---

Approval: User reviews patch, approves, commits.
```

---

## Stale Context Recovery

If `/dungeoneer sync` hasn't been run in >1 day:

1. Suggest to user: "Context is 1+ day old. Run `/dungeoneer sync` to refresh."
2. Use cached PROJECT_SNAPSHOT.json but flag uncertainty
3. Cite timestamp: "Last scan: 2026-03-21 at 10:00 UTC — may be stale"

---

## Special Cases

### Pattern: Roadmap + Phase Questions
**Example**: "What phase are we in?" "What are the blockers for Phase 4?"

1. Parse `design-docs/roadmap.md` → find current completed/in-progress phase
2. Cite the roadmap table(s) with task dependencies
3. If blocked, cross-reference `BUG_TRACKER.md` blocker field
4. Link to any relevant code proving status (e.g., "Phase 1.7 says save/load, and we have `src/game/metasave.go`")

### Pattern: Bug Questions
**Example**: "What are the current bugs?" "Why does the right-click thing happen?"

1. Load `README.md` Known Issues section + entire `.github/BUG_TRACKER.md`
2. For each bug, cite:
   - ID + title
   - Status (Open/Blocked/In Progress)
   - Attempts tried + why they failed
   - Code locations (📍 file.go lines)
3. If asking about a specific bug, link to the tracker entry + any related code

### Pattern: Item Lookup
**Example**: "What does [item] do?" "List all items and their effects"

1. Load `src/images/items_structured_effects.json`
2. Search by item name (case-insensitive)
3. Return: name, stats (ATK, DEF, HP, etc.), effects (text description), rarity, type
4. Cite the JSON path: `items_structured_effects.json` → "items.[category].[item-name]"

### Pattern: Control / Hotkey Lookup
**Example**: "How do I spawn a monster?" "What's the hotkey to load a level?"

1. Check `README.md` Hotkeys table
2. If complex (level loading), explain the sequence citing code
3. Example answer: "Press **Q** to generate a forest level (📍 README.md Hotkeys). Behind the scenes, this calls `src/leveleditor/procgen.go` MakeForest() (📍 src/levels/generate64.go linked)."

### Pattern: Architecture Deep-Dive
**Example**: "How does monster pathfinding work?" "Where is NPC dialogue handled?"

1. Cite the design doc first (e.g., `design-docs/npc-system.md`)
2. Then walk through code flow: `src/entities/monster.go` → `src/pathing/astar.go` → movement controller
3. Include relevant structs, key functions, call chains
4. Example: "Monster AI pathfinding (📍 design-docs/enemy-placement.md) is implemented as follows: 1) MonsterUpdate() in entities/monster.go (line X) calls PathTo(target) 2) PathTo uses A* in pathing/astar.go (line Y) ..."

### Pattern: Feature Status
**Example**: "Is save/load done?" "What's implemented for Phase 2?"

1. Cite roadmap task + current status
2. Look for code evidence: does the file exist? Is it complete or partial?
3. Example: "Phase 1.7 (Meta Save Basic) is partially done. We have 📍 `src/game/metasave.go` which saves Remnants. But TODO: full run state serialization (see code lines X-Y marked with TODO)."

### Pattern: Blocker Analysis
**Example**: "What's blocking Phase 3?" "Why can't we move to the next phase?"

1. Parse roadmap → find all Phase N tasks listed as dependencies
2. Cross-check against `BUG_TRACKER.md` → see which blockers are marked
3. Link: code → bug → roadmap dependency
4. Example answer: "Phase 3 depends on Phase 1 completion (📍 roadmap.md). Phase 1 is blocked by Bug #1 (right-click stutter) which prevents reliable pathfinding testing. See BUG_TRACKER.md Bug #1, attempts 1-2 failed because [reasons]. Code: src/movement/controller.movement.go."

---

## Query Handling Rules

**All answersfor dungeoneer queries MUST be based on comprehensive repo scan results.**

### Pattern: Roadmap + Phase Questions
**Example**: "What phase are we in?" "What are the blockers for Phase 4?"

1. Parse `design-docs/roadmap.md` → find current completed/in-progress phase
2. Cross-reference code: check if Phase files exist and are implemented
3. Cite the roadmap table(s) with task dependencies + code evidence
4. If blocked, cross-reference `BUG_TRACKER.md` (local skill dir) blocker field
5. Link to any relevant code proving status (e.g., "Phase 1.7 says save/load, and we have `src/game/metasave.go` at 800+ lines")

### Pattern: Bug Questions
**Example**: "What are the current bugs?" "Why does the right-click thing happen?"

1. Load full `BUG_TRACKER.md` from skill directory
2. Also check README.md Known Issues section
3. For each bug, provide:
   - ID + title + status (Open/Blocked/In Progress/Fixed)
   - Priority level
   - Attempts tried + why they failed
   - Code locations (📍 file.go lines)
4. Rank by priority to user

### Pattern: Item Lookup
**Example**: "What does [item] do?" "List all items and their effects"

1. Parse `src/images/items_structured_effects.json` for full item catalog
2. Cross-check against `src/items/registry.go` for code definitions
3. Return: name, stats (ATK, DEF, HP, special effects), rarity, type, location
4. Cite: `items_structured_effects.json → [category].[item-name]`

### Pattern: Control / Hotkey Lookup
**Example**: "How do I spawn a monster?" "What's the hotkey to load a level?"

1. Parse `src/controls/controls.go` + `src/controls/config.json` for current bindings
2. Compare against README.md Hotkeys table — flag if stale
3. For complex sequences, explain flow with code citations
4. Example: "Press **Q** to generate forest (📍 src/leveleditor/procgen.go). Code path: input handler → leveleditor.procgen.MakeForest() → levels/generate64.go"

### Pattern: Architecture Deep-Dive
**Example**: "How does monster pathfinding work?" "Where is NPC dialogue handled?"

1. Cite the design doc first (`design-docs/enemy-placement.md`, `design-docs/dialogue-system.md`)
2. Walk through full code flow with file + line references
3. Include struct definitions, key function names, call chains
4. Example: "Monster AI (📍 design-docs/enemy-placement.md): 1) MonsterUpdate() in 📍 src/entities/monster.go (line X) calls PathTo(target) 2) PathTo uses A* in 📍 src/pathing/astar.go (line Y) 3) Movement executed in 📍 src/movement/controller.movement.go"

### Pattern: Feature Status
**Example**: "Is save/load done?" "What's implemented for Phase 2?"

1. Cite roadmap task ID + expected phase
2. Check if code file exists + measure completeness (file size, TODO count)
3. Return: status (Done/WIP/Partial/Not Started) + % complete + code citations
4. Example: "Phase 1.7 (Meta Save): ~60% done. We have 📍 src/game/metasave.go (850 lines) saving Remnants. Incomplete: full run state serialization (see code lines X-Y marked //TODO)"

### Pattern: Blocker Analysis
**Example**: "What's blocking Phase 3?" "Why can't we move to the next phase?"

1. Parse `design-docs/roadmap.md` → identify Phase N dependencies
2. Cross-check against `BUG_TRACKER.md` → which bugs block which tasks?
3. Link: code → bug ID → roadmap task → phase dependency
4. Example: "Phase 3 depends on Phase 1 completion (📍 roadmap.md Phase 3.1). **Blocker**: Bug #1 (right-click stutter, 📍 BUG_TRACKER.md) prevents reliable pathfinding testing (🔴 AFFECTS PHASE 1 STABILIZATION). Attempts 1-2 failed (see BUG_TRACKER for details). Code: 📍 src/movement/controller.movement.go"

---

## Citation Format (ENFORCED)

When answering, ALL claims must cite exact sources:

✅ **GOOD**:
```
The FOV system uses raycasting with corner-casting (📍 design-docs/layered-world-system.md 
Rendering section) implemented in 📍 src/fov/fov.go (lines 45-120) and rendered via 
📍 src/fov/render.go (RenderFOV function). It's integrated into the game loop in 
📍 src/game/draw.game.go (line 234).
```

❌ **BAD** (too vague):
```
The FOV system renders via raycasting. See the fov files for details.
```

---

## New Bug Filing Helper

When user reports a new bug:

1. Confirm: bug title + exact reproduction steps
2. Add entry to `.github/skills/dungeoneer-expert/BUG_TRACKER.md`
3. Ask: which source files are involved? (for code references)
4. Set status: "🔴 Open"
5. Note: "Will be reflected in README.md on next `/dungeoneer sync`"

---

## Roadmap Phase Completion Helper

When all tasks for a phase are done:

1. Update `design-docs/roadmap.md` → mark phase tasks complete
2. Update README.md "Developer Notes" → new phase focus
3. Run `/dungeoneer sync` → auto-update PROJECT_SNAPSHOT.json
4. Suggest commit message: "Phase X complete: [summary]. Skill sync updated tracking."

---

## Design Doc Drift Detection

If design doc describes feature not yet implemented:

1. Flag: "Design ahead of code"
2. References: cite doc section + check if code files exist
3. Suggest: add to roadmap as upcoming task

If code implements feature not in design doc:

1. Flag: "Code ahead of design"
2. Add to backlog: "Update design-docs to capture implementation"

---

## Performance & Context Efficiency

- **Cache results**: PROJECT_SNAPSHOT.json stores scan results; re-use if <1 hour old
- **Incremental scans**: If user says "only check src/game/", scan just that dir
- **Parallelizable scans**: Code module scan + design doc scan can run in parallel
- **Batch citations**: When answering, combine multiple citations into compact format

---

## Testing the Skill

To verify skill is working correctly, ask:

```
/dungeoneer: What is the current project status? Generate a full status report.
```

Expected response:
- Current phase + % complete
- All open bugs (from BUG_TRACKER.md)
- Feature checklist status
- Next blockers
- Hotkey reference
- **ALL cited with exact file links**
