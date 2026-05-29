---
plan-id: 11A-glyph-language
status: queued
owner: unassigned
branch: plan/11A-glyph-language
depends-on: [10C-world-expansion, 10B-abaddon-meta-narrative]
last-touched: 2026-05-27
---

# Plan: Phase 11A — Glyph Language System

## Goal

Introduce a discoverable cipher language as a NG+ reward layer. After completing their first run, players begin encountering Inscription Stone entities throughout the dungeon — interactable objects (like NPCs) that display encoded glyph messages. Each message is written in a 20-symbol geometric alphabet. Fragment items scattered across dungeon rooms each decode one symbol, growing the player's codex over successive runs. Early messages pay off in lore; later messages (requiring rarer symbols) unlock mechanical secrets: hidden shrine interactions, dormant objects, one-time boons. The system rewards players who want to dig deeper into the world's lore without blocking players who don't.

Inspired by Pokémon Gen 3's Braille/Unown mechanic — a learnable in-world script that rewards curiosity and pattern recognition.

## Design Reference

**The alphabet:** 20 geometric symbols (circles, lines, crosses, triangles, and combinations). Each symbol maps to a short word or concept (not individual letters — concept-glyphs for words like CHAIN, REMEMBER, ABOVE, BELOW, FIRST, NAME, BLOOD, etc.). Messages are 1–3 "words" per line, 2–4 lines per shrine.

**Fragment items:** `GlyphFragment` item type — not equippable, not consumable. Picked up like a regular item. Each fragment has an ID corresponding to one of the 20 symbols and reveals that symbol's meaning in the codex. Fragments drop from specific room types and biomes (not random — seeded by floor/biome so runs are consistent).

**Inscription Stones:** Interactable entities (E to interact). Show glyph tiles for each symbol in the message. Undecoded symbols render as the raw geometric shape. Decoded symbols show the concept-word below the shape. The panel renders the full message both encoded and decoded.

**Codex UI:** Hub panel (second tab in the lore library, or its own pedestal near the lore library). Shows all 20 symbols in a grid: decoded entries show the symbol + meaning; undecoded entries show the symbol as "???". Players can see which messages on known shrines they can now read.

**Tiered payoff:**
- Symbols 1–8 (common fragments, early floors): unlocking messages in this tier adds lore entries (cosmology, character backstories).
- Symbols 9–16 (uncommon fragments, mid floors): messages hint at hidden Inscription Stones in specific biomes.
- Symbols 17–20 (rare fragments, boss-adjacent rooms): fully decoded messages in this tier unlock mechanical secrets — a dormant shrine that gives a one-run ability modifier, a hidden cache, or a sealed door interaction.

**Message categories:**
- **Cosmological** — who built the dungeon, what Remnants are spiritually, the loop mechanism.
- **Character** — what NPCs were before the dungeon claimed them. Written in first person as if they left the inscriptions themselves.
- **Warnings** — messages addressed to "the one who returns," aware the reader is a repeat visitor.
- **Abaddon's Voice** — the last category, unlocked by symbols 17–20. Written differently — longer, stranger, aware of the player across runs. Directly links to Phase 10B meta-narrative content.

## Scope

**In scope:**
- `GlyphSymbol` and `GlyphFragment` data types.
- Fragment item type in the item registry — non-equippable, tracked in MetaSave.
- `InscriptionStone` entity — interactable, stores a `[]GlyphLine` message, renders decoded/undecoded.
- Codex UI panel in the hub.
- `data/glyphs.json` — 20 symbol definitions (ID, concept-word, visual descriptor for the renderer).
- `data/inscriptions.json` — all shrine message data (which symbols used, payoff type, payoff ID).
- Fragment drop seeding in level generation (biome/floor-specific, not random).
- NG+ gate — Inscription Stones and GlyphFragments only spawn when `MetaSave.CompletedRuns >= 1`.
- 25–30 Inscription Stone placements across all biomes.
- Mechanical payoff hookups for tier-4 symbols (shrine that grants ability modifier, hidden cache, sealed door).

**Out of scope (do not change in this plan):**
- Abaddon's dialogue content (authored in Phase 10B; this plan hooks into it but doesn't write it).
- New biomes (Phase 10C must be complete first so all biomes exist).
- Audio or particle effects on decoding (Phase 9C territory).
- Existing lore library content (Phase 6C).

## File envelope

**Touched:**
- `src/entities/inscription_stone.go` *(new)* — entity definition, interaction handler, glyph render
- `src/items/glyph_fragment.go` *(new)* — `GlyphFragment` item type, codex registration on pickup
- `src/game/lore.go` — extend with `GlyphSymbol` type, codex state helpers
- `src/game/metasave.go` — add `GlyphsDecoded []string`, `InscriptionsRead []string`
- `src/ui/codex.go` *(new)* — codex panel: 20-symbol grid, decoded/undecoded states
- `src/game/hub.go` — spawn codex pedestal entity; gate InscriptionStone spawns on NG+
- `src/game/game.go` — register codex UI state
- `src/game/draw.game.go` — draw dispatch for codex panel
- `src/levels/generate64.go` — fragment drop seeding per biome/floor
- `src/data/glyphs.json` *(new)* — 20 symbol definitions
- `src/data/inscriptions.json` *(new)* — shrine message data
- `design-docs/roadmap.md` — add Phase 11 section
- `plans/_QUEUE.md` — status update on completion

**Forbidden:**
- `src/coords/`, `src/collision/` — offset plan envelope
- `src/dialogues/` — do not modify existing dialogue trees in this plan
- `src/entities/player.go` — no player struct changes in this plan

## Acceptance criteria

- [ ] `MetaSave.CompletedRuns >= 1` gates all glyph content; first-run players see no Inscription Stones and find no GlyphFragments.
- [ ] 20 glyph symbols defined in `data/glyphs.json`, each with ID and concept-word.
- [ ] GlyphFragment items exist in the item registry, drop in biome/floor-seeded locations, register to `MetaSave.GlyphsDecoded` on pickup.
- [ ] InscriptionStone entity is interactable (E key, range check), renders glyph tiles with decoded/undecoded states.
- [ ] Codex panel opens from a hub pedestal, shows all 20 symbols with correct decoded/undecoded state.
- [ ] At least 25 Inscription Stones placed across the full biome set with authored messages.
- [ ] Tier 1–2 messages (symbols 1–8) hook into `unlock_lore` — reading a fully decoded tier-1 shrine adds a lore entry.
- [ ] At least 2 tier-4 messages have wired mechanical payoffs (ability modifier shrine, hidden cache).
- [ ] `cd src && go build ./...` passes.
- [ ] `cd src && go test ./...` passes for touched packages.

## Phases

### Phase 1: Data layer
- [ ] 1.1 Define `GlyphSymbol` struct in `src/game/lore.go`. Fields: ID, ConceptWord, VisualDescriptor string.
- [ ] 1.2 Create `src/data/glyphs.json` with all 20 symbols. Concept words: CHAIN, REMEMBER, ABOVE, BELOW, FIRST, NAME, BLOOD, SEAL, RETURN, MADE, BROKE, BEFORE, WITNESS, BURY, HOLLOW, CONSUME, SPEAK, THRONE, END, ALWAYS.
- [ ] 1.3 Create `src/data/inscriptions.json`. Each entry: ID, biome, floor range, symbols used ([]string), payoff type (lore_id | cache | ability_mod | none), payoff ID.
- [ ] 1.4 Extend `MetaSave` with `GlyphsDecoded []string` and `InscriptionsRead []string`. Update save/load.
- [ ] 1.5 `cd src && go build ./...` passes.

### Phase 2: Fragment items
- [ ] 2.1 Create `src/items/glyph_fragment.go`. `GlyphFragment` item — not equippable, not stackable beyond 1 per symbol. Fields: SymbolID string.
- [ ] 2.2 Register 20 GlyphFragment item templates in the item registry.
- [ ] 2.3 On pickup, add SymbolID to `MetaSave.GlyphsDecoded` (idempotent).
- [ ] 2.4 Seed fragment drops in `src/levels/generate64.go` — biome/floor threshold logic determines which symbol fragment can drop in which context. No random symbol assignment; each floor/biome seeds a fixed symbol so the player can plan.
- [ ] 2.5 Gate fragment spawns: only spawn when `MetaSave.CompletedRuns >= 1`.
- [ ] 2.6 `cd src && go build ./...` passes.

### Phase 3: Inscription Stone entity
- [ ] 3.1 Create `src/entities/inscription_stone.go`. Struct: position, `Message []GlyphLine` (each line is `[]string` symbol IDs), `PayoffType`, `PayoffID`, `HasBeenRead bool`.
- [ ] 3.2 Interaction handler (E key, range check, same pattern as NPC). Opens a glyph panel rather than dialogue panel.
- [ ] 3.3 Glyph panel renderer: grid of symbol tiles per line. Decoded symbols show the geometric shape + concept-word label below. Undecoded symbols show the shape only with "???" label.
- [ ] 3.4 On full decode of a message (all symbols in it are known): trigger payoff. Lore payoffs call `unlock_lore`. Cache payoffs spawn an item. Ability mod payoffs register a one-run modifier.
- [ ] 3.5 Add `InscriptionStoneID` to `MetaSave.InscriptionsRead` on first full decode (idempotent).
- [ ] 3.6 Place Inscription Stones in level generation, gated by `CompletedRuns >= 1`. Load placement data from `inscriptions.json`.
- [ ] 3.7 `cd src && go build ./...` passes.

### Phase 4: Codex UI
- [ ] 4.1 Create `src/ui/codex.go`. `Codex` struct: `Symbols []GlyphSymbol`, `Decoded []string` (from MetaSave).
- [ ] 4.2 Draw 20-symbol grid. 4 columns × 5 rows. Each cell: symbol geometric shape (pixel art), concept-word if decoded / "???" if not, subtle unlock indicator.
- [ ] 4.3 Keyboard/mouse navigation: scroll, close with Escape.
- [ ] 4.4 Spawn codex pedestal entity in hub (near lore library). Interaction opens codex panel. Gate pedestal spawn on `CompletedRuns >= 1`.
- [ ] 4.5 Register codex overlay state in `src/game/game.go` and `src/game/draw.game.go`.
- [ ] 4.6 `cd src && go build ./...` passes.

### Phase 5: Content — author all inscriptions
- [ ] 5.1 Write all 25–30 shrine messages in `inscriptions.json`. Distribute across categories: 8 cosmological, 8 character (one per named NPC), 7 warnings, 7 Abaddon (require symbols 17–20).
- [ ] 5.2 Wire tier-1 and tier-2 payoffs to existing lore IDs from Phase 6C `data/lore.json`.
- [ ] 5.3 Write 2 mechanical payoff shrines (ability modifier + hidden cache) and hook them up.
- [ ] 5.4 Verify all 20 symbols appear in at least 2 shrine messages so no symbol is a dead end.

### Phase 6: Cleanup
- [ ] 6.1 Add Phase 11 section to `design-docs/roadmap.md`.
- [ ] 6.2 Update `CLAUDE.md` phase status.
- [ ] 6.3 Move this plan to `plans/COMPLETED/11A-glyph-language.md`.
- [ ] 6.4 Update `plans/_QUEUE.md`.

## Progress log

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-05-27 | — | Drafted | Plan written from brainstorming session; not yet started |

## What was NOT changed (intentional)

_None yet._

## Open questions

- **Glyph pixel art.** The 20 geometric symbols need actual pixel art tiles. Either: (a) generate them procedurally from the VisualDescriptor string at runtime (circle, triangle, cross combinations), or (b) require a spritesheet. Option (a) is more flexible and avoids an art dependency. Recommendation: runtime geometry drawing using Ebiten vector primitives. Affects Phase 3.3.
- **Codex pedestal sprite.** No dedicated sprite likely exists. Use a stone/altar tile or tint an existing entity. Document the gap.
- **Abaddon messages.** Symbols 17–20 power the Abaddon message category. Those messages should be authored in coordination with Phase 10B's Abaddon dialogue to maintain voice consistency. Do not finalize Abaddon inscription content until 10B dialogue is written.
