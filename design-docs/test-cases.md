# Test Cases

Manual test scenarios for verifying Dungeoneer features. Each test has a unique ID, the phase it belongs to, steps to reproduce, and expected results.

Tests marked **[BLOCKING]** must pass before the phase is considered complete. Tests marked **[VISUAL]** require visual inspection. Tests marked **[REGRESSION]** should be re-run after changes to related systems.

---

## Phase 3: NPCs & Dialogue

### T3.01 — NPC Spawns on Dungeon Floor [BLOCKING]
**Steps:** Start a new run. Enter floor 1.
**Expected:**
- 1-2 NPCs visible in room centers
- NPC not in boss room, not on player spawn tile, not on exit tile
- NPC bobs with idle sine animation
- NPC faces player when within 6 tiles

### T3.02 — "[E] Talk" Hint Appears [BLOCKING]
**Steps:** Walk within 2 tiles of an NPC.
**Expected:**
- "[E] Talk" text appears centered above the NPC
- Hint disappears when player moves out of range
- Only one hint shows at a time (nearest NPC)
- Hint does not appear while dialogue panel is open

### T3.03 — Dialogue Opens and Typewriter Plays [BLOCKING]
**Steps:** Stand near NPC, press E.
**Expected:**
- Semi-transparent dark overlay covers the screen
- Bottom-center panel (~80% width, ~35% height) with dark red border
- NPC portrait at 2x scale on left side of panel
- Speaker name in gold above dialogue text
- Text types out character by character (typewriter effect)
- All game input blocked (cannot move, open inventory, etc.)

### T3.04 — Click to Skip Typewriter [BLOCKING]
**Steps:** During typewriter animation, click anywhere on the panel.
**Expected:**
- Full text appears immediately
- "[Click to continue]" hint or response options become visible

### T3.05 — Advance Through SimpleDialogue [BLOCKING]
**Steps:** Talk to any minor NPC. Click "[Continue]" through all lines.
**Expected:**
- Each click advances to the next line with a new typewriter animation
- After the final line, dialogue panel closes automatically
- Player can move again immediately

### T3.06 — Escape Closes Dialogue [BLOCKING]
**Steps:** Open dialogue with any NPC. Press Escape.
**Expected:**
- Dialogue panel closes immediately
- Game resumes normal input handling

### T3.07 — Biome Filtering [REGRESSION]
**Steps:** Start multiple runs until you encounter Crypt, Brick, and Gallery biome floors.
**Expected:**
- Crypt floors: Hollow Monk, Weeping Shade, or Scavenger
- Brick floors: Forgotten Soldier or Scavenger
- Gallery floors: Mad Scholar or Scavenger
- Scavenger can appear on any biome
- No biome-restricted NPC appears outside its valid biomes

### T3.08 — No NPC in Boss Room [BLOCKING]
**Steps:** Reach the boss floor.
**Expected:**
- No minor NPCs spawned inside the boss arena room
- Boss arena functions normally

### T3.09 — Hub NPC Spawning After Meeting [BLOCKING]
**Steps:** Meet an NPC during a dungeon run (open dialogue). Die or complete run. Return to hub.
**Expected:**
- The NPC you spoke to appears at one of the hub slots
- NPC is interactable in the hub with the same dialogue
- NPCs you have NOT met do not appear in the hub

### T3.10 — Dialogue Blocks All Input [BLOCKING]
**Steps:** Open dialogue with an NPC. Attempt: WASD movement, I (inventory), Tab, E, mouse clicks outside panel.
**Expected:**
- No game actions processed while dialogue is active
- Only dialogue panel clicks and Escape key are handled

### T3.11 — Multiple NPCs on Same Floor [VISUAL]
**Steps:** Play multiple floors until a floor spawns 2 NPCs.
**Expected:**
- Both NPCs render correctly with proper depth sorting
- Each has its own independent dialogue
- "[E] Talk" shows only for the nearest NPC in range

### T3.12 — NPC Not on Player Spawn or Exit [REGRESSION]
**Steps:** Start 5+ runs, observe NPC positions on each floor.
**Expected:**
- No NPC ever occupies the player's spawn tile
- No NPC ever occupies the exit portal tile
- Minimum visual separation between NPC and spawn/exit

### T3.13 — Dialogue Panel Resize [VISUAL]
**Steps:** Open dialogue with an NPC. Resize the game window (or toggle fullscreen).
**Expected:**
- Dialogue panel repositions correctly to bottom-center of new window size
- Text, portrait, and responses remain properly laid out

### T3.14 — NPC Rendering in FOV [VISUAL]
**Steps:** On a dungeon floor (not hub), find an NPC. Walk away until it leaves FOV.
**Expected:**
- NPC disappears when outside player's field of view
- NPC reappears when player returns within FOV
- In the hub (FullBright), NPCs are always visible regardless of distance

### T3.15 — Quest Flag Conditions (Future — Branching Dialogue)
**Steps:** Create a dialogue tree with a response gated by `flag_equals`. Set the flag via a prior dialogue action. Re-enter dialogue.
**Expected:**
- Gated response hidden when condition not met
- Gated response visible after flag is set
- Selecting a response with `OnSelect` actions fires those actions

### T3.16 — SimpleDialogue to Tree Conversion [REGRESSION]
**Steps:** Add a new SimpleDialogue JSON file to `src/dialogues/`. Restart the game.
**Expected:**
- File auto-loaded by `dialogue.LoadAll()`
- Converted to DialogueTree with `[Continue]` responses
- NPC with matching DialogueID can open this dialogue

---

## Adding New Test Cases

When adding tests for a new phase or feature:

1. Use the ID format `T{phase}.{number}` (e.g., T4.01 for the first Phase 4 test)
2. Mark blocking tests that must pass for the phase to ship
3. Mark visual tests that require manual screenshot/observation
4. Mark regression tests that should be re-run when related systems change
5. Include exact steps to reproduce — assume the tester has no context
6. State expected results precisely — "works correctly" is not a valid expected result
