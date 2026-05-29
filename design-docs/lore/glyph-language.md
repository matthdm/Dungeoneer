# Dungeoneer — The Glyph Language

*The written record of what the Dungeon remembers about itself.*

---

## Origin

The glyph language was not invented. It was not developed by any of the six collapsed domains or by any soul currently in the Dungeon. It predates the Severance.

It is the residue of The Unnamed's organizational principle — the set of concepts that structured the cosmos before the killing. Every soul, every civilization, every arrangement of matter and meaning existed inside these concepts whether they knew it or not. CHAIN. NAME. THRONE. BLOOD. ABOVE. BELOW. They were not words. They were the joints the universe moved on.

When The Unnamed was destroyed, these concepts did not disappear. They were too structural to disappear. They embedded themselves into the sediment — into the deep architecture of every world that subsequently collapsed into the Dungeon. They are in the stone at depths below any domain layer. They are in the walls. The Dungeon itself carries them the way a body carries its skeleton after everything else has dissolved.

Souls who have been in the Dungeon long enough, who have descended far enough, who have pressed close enough to the old sediment, begin to perceive the glyphs. Some of them write them on walls because writing them provides some relief from the pressure of carrying them. Some of them do not know they are writing. They find the marks on walls they do not remember touching.

The player, in their second and subsequent runs, begins to find these marks systematically — appearing on Inscription Stones, objects that the Dungeon has organized into legible form. The organization is not accidental. Something is trying to communicate.

Whether that something is Abaddon, or the Dungeon itself, or the residue of The Unnamed, or some combination that cannot be meaningfully distinguished: unknown.

---

## The Twenty Symbols

Each symbol is a geometric shape constructed from circles, lines, triangles, and crosses. They are designed to be renderable as pixel art using simple vector primitives — no curve that cannot be approximated by straight lines, no shape that requires more than three component elements.

The `VisualDescriptor` field in `data/glyphs.json` uses the format: `[shape] + [modifier]`, where shape is the primary form and modifier is an additional element. Renderers draw these procedurally.

---

| ID | Symbol | Concept | VisualDescriptor | Notes |
|---|---|---|---|---|
| `SYM_CHAIN` | I | CHAIN | `two-circles-linked-horizontal` | Two interlocked rings. The foundational Bound symbol, but older than The Bound — the concept of binding is structural. |
| `SYM_REMEMBER` | II | REMEMBER | `circle-dot-center` | A circle with a point at its center. The eye, the held impression, the thing that persists. |
| `SYM_ABOVE` | III | ABOVE | `triangle-up` | Upward triangle, clean. Direction, aspiration, divine authority, hierarchy ascending. |
| `SYM_BELOW` | IV | BELOW | `triangle-down` | Downward triangle, clean. Depth, burial, descent, the Dungeon, what receives. |
| `SYM_FIRST` | V | FIRST | `line-vertical-single` | A single vertical line. The One. The singular. The origin point. |
| `SYM_NAME` | VI | NAME | `line-horizontal-circles-ends` | Horizontal line with small circles at each terminus. A named thing has defined edges. |
| `SYM_BLOOD` | VII | BLOOD | `cross-diagonal` | An X. Sacrifice. Crossing. The price. The Severance. |
| `SYM_SEAL` | VIII | SEAL | `circle-line-vertical-internal` | Circle with a vertical line bisecting it. Closed. Bound shut. The contained thing. |
| `SYM_RETURN` | IX | RETURN | `circle-open-arrow-inward` | Circle with a gap at the top and an inward-pointing arrow completing it. The cycle that feeds back into itself. |
| `SYM_MADE` | X | MADE | `three-lines-converge-point` | Three lines meeting at a single point below them. Created from parts. The built thing. |
| `SYM_BROKE` | XI | BROKE | `line-horizontal-gap-center` | A horizontal line with a visible gap at its center. Fracture. The severed thing. |
| `SYM_BEFORE` | XII | BEFORE | `two-lines-horizontal-lower-longer` | Two horizontal lines, the lower one longer. What underlies. What preceded everything else. Substrate. |
| `SYM_WITNESS` | XIII | WITNESS | `line-vertical-crossbar-top` | Vertical line with a horizontal crossbar near the top. A figure watching. The observing presence. Azazel. |
| `SYM_BURY` | XIV | BURY | `triangle-inverted-line-base` | Inverted triangle with a horizontal line through its base. Pressed down. Subsumed. Given to the below. |
| `SYM_HOLLOW` | XV | HOLLOW | `circle-outline-only` | A circle with no fill — just the ring. Emptiness that remembers having been full. The Pale. The Hollow Courts. |
| `SYM_CONSUME` | XVI | CONSUME | `circle-arrows-inward` | Circle with multiple short arrows pointing inward from all sides. Absorption. The Endless. |
| `SYM_SPEAK` | XVII | SPEAK | `wave-line-horizontal` | A horizontal zigzag. Sound. Language. Voice. The act of making meaning external. |
| `SYM_THRONE` | XVIII | THRONE | `square-line-vertical-above` | A square with a short vertical line rising from its center top. The seat of power. The ruling. The Unnamed's position. |
| `SYM_END` | XIX | END | `line-horizontal-curves-down-ends` | A horizontal line with downward curves at both termini. Terminus. The place things arrive and stop. |
| `SYM_ALWAYS` | XX | ALWAYS | `two-circles-concentric` | Two concentric circles. Without boundary. Without terminus. Eternal. Abaddon. |

---

## Syntax

The glyph language does not have grammar in the conventional sense. It operates through juxtaposition and sequence — the relationship between symbols creates meaning that no individual symbol contains.

**Reading order:** Left to right within a line, top to bottom across lines.

**Juxtaposition rules:**
- Two symbols side by side: the first modifies or acts on the second. `BLOOD / BROKE` = "Blood broke." `CHAIN / RETURN` = "The chain returns."
- Three symbols in a line: subject / verb / object, loosely. `NAME / BROKE / ABOVE` = "The name broke what was above."
- The subject is often absent when it is understood from context. `RETURN / ALWAYS` = "[It] returns always."

**Line breaks:**
- Each new line is a new statement. A multi-line inscription is a sequence of statements, not a sentence.
- Lines that echo each other's symbols create deliberate resonance — the repeated symbol gains emphasis.

**Spacing:**
- Symbols touching each other are compound — read as a single concept.
- Symbols with space between them are adjacent — read as sequence.
- In practice, Inscription Stones render one symbol per tile, always with space between, so all symbols are read as sequential. The compound interpretation applies only to symbols found in the wild, etched directly into walls without the Stone's framing.

---

## Fragment Distribution

Fragments drop in biome-seeded, floor-specific locations — not randomly. The player who pays attention can predict which biomes carry which symbols.

| Symbol Tier | Biome | Floors | Symbols |
|---|---|---|---|
| 1 (Common) | Catacomb | 1–3 | CHAIN, WITNESS, FIRST, ABOVE |
| 1 (Common) | Gehenna | 1–3 | BLOOD, RETURN, SEAL |
| 1 (Common) | Moss | 1–3 | BELOW, CONSUME, MADE |
| 2 (Uncommon) | Cocytus | 4–6 | REMEMBER, HOLLOW, BURY |
| 2 (Uncommon) | Mirror Halls | 4–6 | NAME, SPEAK, BROKE |
| 2 (Uncommon) | Hollow Courts | 4–6 | BEFORE, END |
| 3 (Rare) | Any boss-adjacent room | 7+ | RETURN *(duplicate tier)*, THRONE, WITNESS *(duplicate tier)* |
| 4 (Very Rare) | Deep floors only | 8+ | SPEAK *(duplicate tier)*, ALWAYS, THRONE *(duplicate tier)*, END *(duplicate tier)* |

*Tier 4 symbols (SPEAK, ALWAYS, THRONE, END in their rare variants) are required to read Abaddon's Voice inscriptions. They appear only on floors 8+ and only in specific room types: rooms with no enemies, rooms adjacent to boss arenas, rooms that appear in NG+ only.*

---

## The Inscriptions

All 30 authored Inscription Stone messages, organized by category. Each entry includes: the symbols used (in reading order), the plain translation, the intended meaning, the biome and floor range where the Stone appears, and the payoff type.

---

### Category I — Cosmological (8 inscriptions)

*Who built the Dungeon. What the Remnants are. The shape of what happened.*

---

**INS_COSM_01**
`FIRST / THRONE / ALWAYS`
*"The first sat the throne always."*
The Unnamed was the organizing principle before any other thing. Its order was not chosen — it simply applied.
**Biome:** Catacomb | **Floor:** 1–2 | **Payoff:** `lore_the_unnamed`

---

**INS_COSM_02**
`NAME / BROKE / BLOOD`
*"The name broke in blood."*
The Severance — the killing — shattered the coherence that had named everything. The naming and the killing are the same event.
**Biome:** Gehenna | **Floor:** 2–3 | **Payoff:** `lore_the_severance`

---

**INS_COSM_03**
`MADE / ABOVE / SEAL`
*"What was made above sealed."*
The Severed absorbed the fragments of the Unnamed and sealed them inside themselves — each became a living container of divine power. They were made into seals.
**Biome:** Catacomb | **Floor:** 3–4 | **Payoff:** `lore_shard_nature`

---

**INS_COSM_04**
`BURY / BELOW / REMEMBER`
*"Buried below, [it] remembers."*
The Dungeon is a grave. It is also a record. Every biome is a world that was buried in descent, and the stone remembers what lived in it.
**Biome:** Cocytus | **Floor:** 4–5 | **Payoff:** `unlock_lore` → `lore_pale_echo`

---

**INS_COSM_05**
`CHAIN / SEAL / FIRST`
*"The chain sealed the first."*
Shard-power — the remnant of The Unnamed — was bound (chained) into the Severed and their worlds. The first coherence is sealed inside every artifact.
**Biome:** Catacomb | **Floor:** 2–3 | **Payoff:** none *(enriches artifact pickups)*

---

**INS_COSM_06**
`HOLLOW / ALWAYS / BEFORE`
*"The hollow was always before."*
Abaddon predates everything. The dissolution was not caused by the Severance — it was already present, already structural. The Unnamed had been managing it.
**Biome:** Hollow Courts | **Floor:** 5–6 | **Payoff:** `lore_abaddon_nature`

---

**INS_COSM_07**
`BLOOD / RETURN / BELOW`
*"Blood returns below."*
Souls descend into the Dungeon. Abaddon receives them. The blood spilled in the Severance flows downward, always downward.
**Biome:** Gehenna | **Floor:** 3–4 | **Payoff:** none

---

**INS_COSM_08**
`THRONE / BROKE / END`
*"The throne broke. The end."*
The Unnamed's position was destroyed. This was the end — not of everything, but of the order that had prevented everything from ending on its own schedule.
**Biome:** Any | **Floor:** 6–7 | **Payoff:** `lore_azazel_purpose`

---

### Category II — Character (8 inscriptions)

*What the named souls were before the Dungeon. Written in first person — as if they left these marks themselves, at a moment when they still knew what they were.*

---

**INS_CHAR_VARN**
`CHAIN / WITNESS / ALWAYS`
*"The chain. I witness. Always."*
Varn's self-description at his most honest: he is the chain, he is the witness, this is not a phase he is going through.
**Biome:** Catacomb | **Floor:** 1–2 | **Payoff:** `lore_varn_origin`

---

**INS_CHAR_SERIS**
`BLOOD / REMEMBER / MADE`
*"Blood. I remember what was made."*
Seris carries the memory of what The Pyre produced at its height — the beautiful renewal, before the burning became the point. The blood is both the fire's cost and the sacrifice the distinction required.
**Biome:** Gehenna | **Floor:** 2–3 | **Payoff:** `lore_seris_distinction`

---

**INS_CHAR_MIRA**
`NAME / SPEAK / HOLLOW`
*"The name I speak is hollow."*
Mira on her own constructed self — the identity she built as the helper, the shaper, the one who didn't need defending. She knows it. She built it anyway.
**Biome:** Mirror Halls | **Floor:** 3–4 | **Payoff:** `lore_mira_truth`

---

**INS_CHAR_KAEL**
`ABOVE / CONSUME / BELOW`
*"What is above consumes what is below."*
The honest statement of The Canopy's hierarchy. Not cruel — accurate. Kael would sign this.
**Biome:** Moss | **Floor:** 2–3 | **Payoff:** `lore_kael_garden`

---

**INS_CHAR_AZAZEL**
`MADE / WITNESS / BROKE`
*"Made to witness. [I] broke."*
Azazel's condition described from the inside — to whatever extent inside is a meaningful location for something that was never a person. Made for a purpose. The purpose is broken. He continues.
**Biome:** Crimson Deep | **Floor:** 7–8 | **Payoff:** none

---

**INS_CHAR_UNNAMED**
`FIRST / NAME / THRONE`
*"The first name on the throne."*
Not eulogy. Statement. The Unnamed occupied a specific position: the first, the naming principle, the organizing seat. This inscription reads like a title on a door that no longer has a room behind it.
**Biome:** Raw Stone transition floor | **Floor:** 5–6 | **Payoff:** `lore_the_unnamed` *(if not already unlocked)*

---

**INS_CHAR_ABADDON**
`BEFORE / END / ALWAYS`
*"Before the end. Always."*
Abaddon before everything. Abaddon after everything. The end is not terminal for him — it is structural. He was before it and will be after it.
**Biome:** Raw Stone | **Floor:** 6–7 | **Payoff:** none *(enriches Abaddon encounters)*

---

**INS_CHAR_PLAYER**
`RETURN / BLOOD / REMEMBER`
*"[You] return. Blood. Remember."*
This inscription is addressed to the player specifically — the second-person imperative is implied by the absence of a stated subject. "You return through blood. Remember."
**Biome:** Any sanctuary room | **Floor:** 2+ | **Payoff:** *(narrative — no mechanical payoff; players who find this early carry it)*

---

### Category III — Warnings (7 inscriptions)

*Written by previous cyclers. Not the player — others who came before, who somehow survived long enough to learn the language and leave marks. These people are not in the Dungeon anymore. What happened to them is unspecified.*

---

**INS_WARN_01**
`RETURN / WITNESS / ABOVE`
*"Return to witness what is above."*
The first warning is also a direction: there is something above the Dungeon worth seeing, but it requires returning. Keep cycling.
**Biome:** Any | **Floor:** 1 | **Payoff:** none *(early motivation to persist)*

---

**INS_WARN_02**
`CHAIN / RETURN / SPEAK`
*"The chain that returns speaks."*
Varn is a trap — not maliciously. The chain-logic he embodies follows you if you engage with it. When you return to him run after run, the chain speaks in a new register.
**Biome:** Catacomb | **Floor:** 3–4 | **Payoff:** none

---

**INS_WARN_03**
`REMEMBER / HOLLOW / BELOW`
*"Remember the hollow below."*
The deep floors contain the worst kind of emptiness — the kind that was once full. Don't mistake the quiet of Cocytus and the Hollow Courts for safety.
**Biome:** Any transition floor | **Floor:** 4–5 | **Payoff:** none

---

**INS_WARN_04**
`BLOOD / SEAL / END`
*"Blood seals the end."*
The Severance — the blood spilled to kill The Unnamed — sealed the dissolution in place permanently. There is no undoing it. The end is sealed with what caused it.
**Biome:** Gehenna | **Floor:** 5–6 | **Payoff:** none *(darkens the cosmology)*

---

**INS_WARN_05**
`CONSUME / WITNESS / ALWAYS`
*"The consuming watches always."*
Something in the deep floors does not sleep, does not leave, does not forget you passed through. A warning about Azazel, or the deep biome enemies, or both.
**Biome:** Hollow Courts | **Floor:** 5–6 | **Payoff:** none

---

**INS_WARN_06**
`BROKE / NAME / REMEMBER`
*"The broken names are remembered."*
The Severed destroyed themselves in the destruction of their worlds. But they are remembered — in the artifacts, in the NPC inheritors, in the shard-residue in the walls. Destruction is not erasure.
**Biome:** Any | **Floor:** 6–7 | **Payoff:** none

---

**INS_WARN_07**
`BURY / SEAL / RETURN`
*"What is buried and sealed returns."*
The final warning before the deep floors. What was pressed into the sediment, sealed against access, does not stay sealed when something keeps descending.
**Biome:** Raw Stone | **Floor:** 7 | **Payoff:** unlocks NG+ content hint

---

### Category IV — Abaddon's Voice (7 inscriptions)

*These require at least two tier-4 symbols (ALWAYS, THRONE, END, SPEAK in their rare floor-8+ variants). They are written differently — longer, stranger, in a register that does not quite match how any soul writes. The voice knows the player. It has been watching across runs.*

---

**INS_ABADD_01**
`SPEAK / ALWAYS / BEFORE`
*"[I] speak what was always before."*
The opening statement. Abaddon is describing himself to the player for the first time, in the only language that can carry what he is: the concepts that predate everything.
**Biome:** Raw Stone | **Floor:** 8 | **Payoff:** none *(begins Abaddon's Voice sequence)*

---

**INS_ABADD_02**
`HOLLOW / THRONE / END`
*"The hollow was the throne before the end."*
Before The Unnamed organized everything, there was only Abaddon — the hollow, the dissolution. He was the condition. The throne was built inside him. The throne broke. He remains.
**Biome:** Crimson Deep | **Floor:** 8 | **Payoff:** `lore_abaddon_nature`

---

**INS_ABADD_03**
`RETURN / ALWAYS / CONSUME`
*"[You] return. Always. [I] consume."*
Abaddon noting the player's cycling, noting his own nature, holding them in the same inscription without judgment. This is observation, not threat.
**Biome:** Any deep floor | **Floor:** 8–9 | **Payoff:** none *(meta-narrative escalation)*

---

**INS_ABADD_04**
`BEFORE / SPEAK / NAME`
*"Before [the] name was spoken..."*
The inscription is incomplete — three symbols, but the statement trails. What was before the first naming? This is Abaddon approaching the limit of what the language can express about his nature.
**Biome:** Raw Stone | **Floor:** 9 | **Payoff:** `lore_fragment_001`

---

**INS_ABADD_05**
`THRONE / CONSUME / BURY`
*"The throne was consumed and buried."*
Direct account of the Severance from Abaddon's perspective: he watched the organizing principle be destroyed, absorbed, pressed into mortal forms that then descended into him. He received all of it.
**Biome:** Crimson Deep | **Floor:** 9 | **Payoff:** `lore_fragment_002`

---

**INS_ABADD_06**
`END / ALWAYS / HOLLOW`
*"The end is always hollow."*
After everything dissolves, what remains is the hollow — Abaddon. The end is not an event; it is a state, and the state is him. Not threatening. Just true.
**Biome:** Any floor 9+ | **Floor:** 9+ | **Payoff:** none

---

**INS_ABADD_07**
`SPEAK / RETURN / WITNESS`
*"Speak. Return. [I] witness."*
The final inscription. An invitation: say something, come back again, he is watching. This is the in-world precursor to `lore_fragment_003`, where Abaddon speaks in full. Finding this inscription before reaching the fragment gives the later text resonance — the player has been building toward it.
**Biome:** Deep floor adjacent to final encounter | **Floor:** 9+ NG+ | **Payoff:** `lore_fragment_003`

---

## Notes for Implementation

**Symbol rendering priority:** CHAIN, ABOVE, BELOW, FIRST, BLOOD are most visually recognizable and should be implemented first for playtesting. Their shapes are the simplest.

**Codex ordering:** The Codex UI should display symbols in the order they are likely to be found (Tier 1 first), not in the numerical order of their IDs. Players should feel the language growing naturally rather than having gaps in a rigid sequence.

**The 20 symbols appear in all 30 inscriptions as follows** — verify no symbol is orphaned:

| Symbol | Appears In |
|---|---|
| CHAIN | INS_COSM_05, INS_COSM_01 *(no)*, INS_CHAR_VARN, INS_WARN_02 |
| REMEMBER | INS_COSM_04, INS_CHAR_SERIS, INS_CHAR_PLAYER, INS_WARN_03 |
| ABOVE | INS_COSM_03, INS_CHAR_KAEL, INS_WARN_01 |
| BELOW | INS_COSM_07, INS_CHAR_KAEL, INS_WARN_03 |
| FIRST | INS_COSM_01, INS_COSM_05, INS_CHAR_UNNAMED |
| NAME | INS_COSM_02, INS_CHAR_MIRA, INS_CHAR_UNNAMED, INS_WARN_06, INS_ABADD_04 |
| BLOOD | INS_COSM_02, INS_COSM_07, INS_CHAR_SERIS, INS_CHAR_PLAYER, INS_WARN_04 |
| SEAL | INS_COSM_03, INS_COSM_05, INS_WARN_04, INS_WARN_07 |
| RETURN | INS_COSM_07, INS_CHAR_PLAYER, INS_WARN_01, INS_WARN_07, INS_ABADD_03, INS_ABADD_07 |
| MADE | INS_COSM_03, INS_CHAR_SERIS, INS_CHAR_AZAZEL |
| BROKE | INS_COSM_02, INS_COSM_08, INS_CHAR_AZAZEL, INS_WARN_06 |
| BEFORE | INS_COSM_06, INS_CHAR_ABADDON, INS_ABADD_01, INS_ABADD_04 |
| WITNESS | INS_COSM_08 *(no)*, INS_CHAR_VARN, INS_CHAR_AZAZEL, INS_WARN_01, INS_WARN_05, INS_ABADD_07 |
| BURY | INS_COSM_04, INS_WARN_07, INS_ABADD_05 |
| HOLLOW | INS_COSM_06, INS_CHAR_MIRA, INS_WARN_03, INS_ABADD_02, INS_ABADD_06 |
| CONSUME | INS_CHAR_KAEL, INS_WARN_05, INS_ABADD_03, INS_ABADD_05 |
| SPEAK | INS_CHAR_MIRA, INS_ABADD_01, INS_ABADD_04, INS_ABADD_07 |
| THRONE | INS_COSM_01, INS_CHAR_UNNAMED, INS_ABADD_02, INS_ABADD_05 |
| END | INS_COSM_08, INS_CHAR_ABADDON, INS_WARN_04, INS_ABADD_06 |
| ALWAYS | INS_CHAR_VARN, INS_CHAR_ABADDON, INS_COSM_06, INS_ABADD_01, INS_ABADD_03, INS_ABADD_06 |

All 20 symbols appear in at least 2 inscriptions. Tier-4 symbols (SPEAK, ALWAYS, THRONE, END) appear in at least 4 each, ensuring multiple paths to decoding them.
