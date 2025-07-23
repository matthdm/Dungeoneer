Character-Driven Quest System Design Document for The Dungeon

1. Overview

Feature Name:Character Ascension Quest System

Purpose:Enable NPCs to grow across layers of the Dungeon through player interaction, ultimately becoming powerful figures—often end bosses—who retain fragments of memory and identity across playthroughs.

Core Themes:

Manufactured purpose

Futility of power

Recurrence and decay

Philosophy as identity

2. System Summary

Element

Description

World

The Dungeon (a purgatorial realm)

NPC Type

Ascendant Characters (Major Entities)

Player Role

Agent of change who indirectly elevates others

Outcome

NPCs become final bosses or faded memories depending on choices

Meta Progression

NPCs retain memory fragments across NG+ cycles

3. Core Loop (Per NPC)

Layer 1–2:

NPC encounter: introduces personality/philosophy

Optional aid: small task or philosophical choice

Layer 3–4:

NPC gains followers/influence based on help

New ability/item unlock

Starts clashing with rival NPCs

Layer 5–7:

Turning point decision (side with them or not)

Morality and motive become blurred

Influence visibly alters world/floor

Layer 8–10:

Final form confrontation if supported

NPC becomes End Boss with powers shaped by player choices

NG+ Cycle:

May reference old playthrough ("Why do I remember you?")

May change philosophy slightly if repeatedly dethroned

4. Example NPC Arc (Template)

Name:

Varn, The Chainkeeper

Initial Personality:Stoic, obsessed with order. Believes rules are salvation from madness.

Philosophy:"Chaos is not freedom. Chains are not punishment. They are memory."

PHASE 1: Introduction (Layer 1–2)

Found guarding a gate of rusted iron.

Dialogue explores his fear of forgetting identity.

PHASE 2: First Task (Layer 2–3)

You help him “bind” a wandering entity (metaphor for fragmenting mind).

Gain passive buff: Order Within (less sanity decay)

PHASE 3: Conflict (Layer 5–6)

He requests you sabotage The Gilded Lie (another NPC).

Can choose to aid, betray, or ignore.

PHASE 4: Ascension (Layer 8+)

If supported, he becomes Warden Varn, Binder of Thought.

Boss battle uses chain-based memory-stealing mechanics.

NG+ Variant:

If defeated multiple times:

Starts to unravel.

Philosophy mutates: “Chains are lies we tell ourselves to be stable.”

5. System Mechanics

Feature

Design

Quest Flags

Boolean or integer state per NPC per run

Choice Tracking

Dialog/quest branches logged per playthrough

End Boss Pool

Final floor draws from ascended NPCs based on flags

NG+ Memory

Meta-log tracks defeat count, alters dialog/appearance

Philosophical Alignment

Player choices tag them (e.g. Order, Deceit, Despair)

6. Emotional Arc Goals

Tragedy: All NPCs must fall, no matter how noble.

Reflection: Each NG+ deepens your understanding of them.

Irony: You are both savior and destroyer.

Growth: The player, not the characters, changes with each loop.

7. Implementation Checklist

Task

Status

Create 5 Ascendant NPCs with defined philosophies

[ ]

Design 3–4 phases per character

[ ]

Implement choice flag tracking system

[ ]

Build NG+ memory dialogue hook system

[ ]

Tie final boss pool to player choice history

[ ]

Add visual indicators of ascension per character

[ ]

