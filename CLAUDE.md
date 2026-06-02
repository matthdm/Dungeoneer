# CLAUDE.md — Agent Entry Point

> Read this first. It is intentionally short and points you at the right doc for the task at hand instead of the whole codebase. Update it when the project's current focus changes.

YOU MAY NOT COMMIT ANYTHING EVER, ONLY MATTHEW MORALES WILL COMMIT.

## What this is

Dungeoneer is a 2D isometric dark-fantasy roguelike in Go on Ebiten v2.8. Real-time tile combat, procedurally generated floors, item-gated abilities, and NPCs that ascend into the run's final boss. Single-player, offline. Source lives in `src/` (~24K lines of Go); the `dungeoneer` Go module sits there too.

## Current state — last updated 2026-06-02

- **Phases 1, 2, 3: complete.** Run loop, combat depth (6 enemy roles, status effects, multi-phase boss), NPCs & dialogue (room-tag placement, branching JSON trees, Varn arc through phase 2 + boss).
- **Phase 4: complete.** Ability gating, 13 ability items, stat modifiers, gold economy, item quality tiers, loot refinement, chest variants, RunState serialization, mid-run save/load.
  - **Active stabilization work: the coordinate unification refactor.** See `OFFSET_UNIFICATION_PLAN.md` — phases 0–2 done, 3–5 partial.
- **Phase 5: complete.** NPC Phase Tracker, Varn arc (4 phases + boss fight with pre/post dialogue), boss selection engine, hub NPC quarter (Varn appears in hub after first meeting).
- **Phase 6: implemented, ⚠️ UNTESTED.** Code is written and unit-tested (46 tests pass) but manual testing per `design-docs/test-cases.md` T1–T9 has not been done. See manual test plan in session context.
  - 6A: MetaSave v1 (`CompletedRuns`, `TotalDeaths`, `TotalRemnants`, `LoreUnlocked`, `HubState`, `Upgrades`, `Betrayed`), milestone system (4 milestones), toast UI.
  - 6B: `meta_flag_gte/equals` conditions, `SelectTree` NG+ branching (betrayed → ng{N} → phase), `varn_ng1/2/3.json` + `varn_betrayed.json` with new quest content, `set_betrayed` cross-run persistence.
  - 6C: Lore registry, 15 lore entries in `data/lore.json`, `unlock_lore` action, lore library UI, hub lore pedestal (milestone-gated), `NPC.OnInteract`.
- **Phases 7–10: not started.** Echoes, living dungeon AI, item sets, hub shop/upgrades, polish, additional NPCs, Abaddon.

## Where to look (by task type)

| Task type | Read this first |
|---|---|
| Coordinate, hit detection, render offset | `OFFSET_UNIFICATION_PLAN.md`, `src/coords/worldpos.go` |
| Add or modify an ability/item | `design-docs/ability-items.md`, `src/items/load.go`, `src/entities/player.go` (`RefreshAbilities`, `EquipStarter`) |
| NPC or dialogue work | `design-docs/dialogue-system.md`, `src/dialogue/`, `src/dialogues/*.json`, `src/game/npc.game.go`, `src/game/npc_data.go` |
| Boss / Varn arc | `design-docs/boss-system.md`, `src/entities/varn_boss.go`, `src/game/boss.game.go` |
| Level generation / biomes / room tags | `design-docs/biome-system.md`, `design-docs/room-tagging.md`, `src/levels/generate64.go`, `src/levels/room_tagger.go`, `src/game/biome.go` |
| Pathing / FOV / movement | `src/pathing/astar.go`, `src/fov/`, `src/movement/controller.movement.go`, `src/collision/box.collision.go` |
| Phase planning, what's next | `design-docs/roadmap.md` |
| Test scenarios / acceptance | `design-docs/test-cases.md` |

## How agents pick up work

Work is chunked into **plans** under `plans/`. One plan = one branch = one bounded scope. An agent never needs to load the whole repo — the plan tells them what to read, what to touch, and what to leave alone.

To start a session: read this file, then `plans/_QUEUE.md`, then the **Active plan** named there. Follow `plans/_PROTOCOL.md` for the start-and-end ritual. Build must pass on every commit.

When the active plan is empty, promote the top-priority queued plan. When a session's work is incomplete, leave the unchecked boxes and append to the plan's progress log so the next agent resumes cleanly.

## Working contract

- **Engineering principles:** `.github/agents/instructions.md`. Performance-first, surgical edits, no per-frame allocations in hot paths, idiomatic Go.
- **Bound your scope.** Each plan has a **File envelope: Touched / Forbidden**. Do not modify forbidden files even if it would be convenient. Out-of-envelope changes go in **Open questions**, not the diff.
- **`OFFSET_UNIFICATION_PLAN.md` is the prototype** for the plan format and predates the `plans/` directory. It is followed under the same conventions and treated as a concurrent in-flight plan.
- **End every session by updating the active plan doc.** The next agent reads it cold; if your reasoning isn't in the progress log or "What was NOT changed (intentional)" section, the context is gone.
- **No tests exist yet.** Adding `_test.go` for pure-logic packages (`coords`, `pathing`, `levels/room_tagger`, `entities/effects`) is the highest-leverage thing you can do for AI maintainability — it's the only feedback loop available without launching the game.

## Coordinate invariant (do not violate)

Combat checks, hit detection, spell origins, and effect anchoring use `coords.WorldPos.BodyCenter()`. Never raw `TileX/TileY` (lags during interpolation) and never raw `InterpX/InterpY` (skips body-center offset). The "Golden rule" doc comment at the top of `src/coords/worldpos.go` is authoritative; the offset plan exists because parts of the codebase predate it and are still being migrated.

## Build & run

- `cd src && go build ./...` must pass before any commit.
- `src/build_and_run.bat` builds and launches on Windows.
- Tests: none yet. Add them as you change pure-logic code.

## Files to keep current

- This file (`CLAUDE.md`) — current phase status and the "where to look" table.
- `README.md` — public-facing status, feature checklist, bug bounties.
- `design-docs/roadmap.md` — phase planning and `Last updated` date.
- `plans/_QUEUE.md` — active and queued plans. Always reflects reality.
- `OFFSET_UNIFICATION_PLAN.md` — progress log for the in-flight refactor.

If those agree, an agent starting cold can be productive in one session.

# Coding Agent Standards

These rules apply to all coding work across the project. Follow them in every session.

## Commit Discipline

- Atomic commits: one logical change per commit.
- Commit messages: imperative mood, under 72 chars for the subject line. Body explains WHY, not WHAT.
- Always commit and push before marking work as complete. Unpushed work does not count.
- Never force-push to main/master. Feature branches only.
- Never skip pre-commit hooks (`--no-verify`). If a hook fails, fix the issue.

## Branch Naming

- Feature: `feature/<issue-number>-<short-description>`
- Bugfix: `fix/<issue-number>-<short-description>`
- Chore: `chore/<short-description>`
- Always branch from the latest main. Rebase before PR if behind.

## Pull Requests

- Title: short, under 70 characters.
- Body: summary of changes, test plan, and link to the GitHub issue.
- One PR per issue. Do not bundle unrelated changes.
- All CI checks must pass before requesting review.

## Tool Usage

### RTK (Rust Token Killer)

RTK runs transparently via a PreToolUse hook. You do not need to invoke it manually.
If `rtk --version` fails or returns unexpected output, stop and report.

### LSP

Use LSP for code navigation: go-to-definition, find-references, hover info.
Enabled via `ENABLE_LSP_TOOL=1`. Available for Python (pyright), Go (gopls), TypeScript.

### jcodemunch

Use for structural code queries: symbol search, file outlines, blast radius analysis.
Must run `index_repo` before queries return results for a repo.

### jdocmunch

Use for navigating large markdown docs by section rather than reading entire files.
Must index the repo or directory first via `doc_index_repo` or `index_local`.

## Memory and Knowledge Graphs

Before asking clarifying questions, check these sources in order:

### claude-mem (cross-session memory)

Search prior session context via the `mcp__plugin_claude-mem_mcp-search__*` tools:
1. `search(query)` → get index with IDs
2. `timeline(anchor=ID)` → surrounding context
3. `get_observations([IDs])` → full details

### graphify knowledge graphs

Query in this order (most specific to broadest):
1. **Project-local**: `./graphify-out/` — repo-specific code and docs
2. **Wiki graph**: `~/workspace/data-worklog/graphify-out/` — Partseeker docs, ADRs, domain knowledge
3. **Supergraph**: `~/workspace/.partseeker-supergraph/` — cross-repo relationships

Start with `GRAPH_REPORT.md` for god nodes and community structure.

### When to ask the user

Only ask if the answer is genuinely not in memory or graphs. When you do ask, state what you already checked.

## Code Standards

Follow the standards in `core/docs/standards/`. Key files:
- `code-quality.md` — structure, naming, dependencies
- `testing.md` — test coverage requirements
- `error-handling.md` — logging and error patterns
- `security.md` — input validation, auth, TLS
- `workflow-discipline.md` — how to approach work

Do not deviate from these standards without explicit user approval.

## Completion Protocol

When your task is done:
1. All tests pass.
2. All changes are committed and pushed.
3. Write a completion report to `completions/` in the worklog repo (if applicable).
4. Include: what was done, evidence of verification, any architectural escalations.

## Project-Specific Context

Load repo-specific overrides, paths, and conventions:

```
@.claude/repo-context.md
```

---

# Project Standards

## Language & Technology Defaults

- Prefer Go. Python 3 acceptable with strict type hints (mypy strict). TypeScript strict mode for frontend.
- When in doubt, choose the strongest type system available.
- AWS serverless over self-managed (SQS over Kafka, Aurora Serverless over self-hosted Postgres, Secrets Manager over Vault).
- Open source over proprietary at comparable quality.
- OpenTelemetry for observability. Structured JSON logging only.

## Behavioural Rules

- Propose test cases before writing implementation. Get approval.
- Before modifying existing code, run existing tests to confirm a passing baseline.
- When asked to "make it work" or "just get it done", do not skip error handling, logging, validation, or tests.
- If a required external dependency is missing (secrets manager, auth provider, log collector, CI pipeline), stop and flag it to the user. Do not work around it with placeholders or local substitutes.
- When generating files in a specific domain (API, database, auth), read the relevant pillar doc in `docs/standards/` first.
- Use conventional commits format for commit messages.
- Prefer small, reviewable changesets over large monolithic changes.
- State assumptions before acting. When ambiguous, present interpretations and ask — do not guess silently.
- Minimum code that solves the problem. No speculative features, premature abstractions, or error handling for impossible scenarios.
- Touch only what the task requires. Do not improve, refactor, or restyle adjacent code you were not asked to change.
- Close the loop: define verifiable success criteria before starting, run tests to confirm each step, do not mark done until verified.

## Critical Rules (always apply)

### Security

- Never hardcode secrets, tokens, keys, or connection strings. No exceptions. No "temporary" ones.
- Parameterised queries only. No string concatenation for SQL or NoSQL queries.
- Input validation at every boundary (API, queue consumer, file ingestion). Allowlists over denylists.
- Centralised auth middleware. Never check permissions inside individual handlers.

### Data Privacy

- Every data access must be scoped by tenant. No unscoped queries, no optional tenant filters.
- PII fields annotated in schema. PII masked in all log output.
- Soft delete with retention hooks. No hard deletes of user data without explicit policy.

### Testing

- No feature code merged without tests. Tests are not optional under time pressure.
- Every test suite must include: happy path, edge cases, error/failure conditions, auth enforcement, tenant isolation.
- Specifically test that user A cannot access user B's data.

### Error Handling

- No swallowed exceptions. Every catch block must handle, log, or re-throw.
- User-facing errors: generic, safe, no internal details. Internal errors: specific, contextual, actionable.
- Every log entry: structured JSON, correlation ID, timestamp, service name.

### MCP Server Safety

- All tools and integrations are read-only by default. Write, mutate, or destroy operations require explicit user approval and must be logged (who approved, what changed, which environment).
- AWS MCP must be read-only. Propose all changes as CloudFormation templates with a cost summary.
- Never create, modify, or delete cloud resources without explicit user confirmation. Describe what will happen, state the estimated cost, and confirm the target environment (dev/staging/prod) first.
- Cheapest viable option. Use the smallest resource that works. Flag costs before the user deploys anything.
- Tag every cloud resource and IaC template: `Project`, `Environment`, `CreatedBy: claude-code`.
- Read `docs/standards/mcp-safety.md` before any MCP server interaction that involves write operations.

## PR Merge Checklist

Before accepting any PR, run all three:

1. `/review-pr` — code review
2. `/security-review` — security review
3. `/simplify` — code simplifier

## Detailed Standards

Read the relevant file before working in that domain:

- [Security](docs/standards/security.md)
- [Data Privacy & Tenant Isolation](docs/standards/data-privacy.md)
- [Secrets Management](docs/standards/secrets.md)
- [Testing](docs/standards/testing.md)
- [Error Handling & Logging](docs/standards/error-handling.md)
- [Observability](docs/standards/observability.md)
- [API Design](docs/standards/api-design.md)
- [Database](docs/standards/database.md)
- [Code Quality](docs/standards/code-quality.md)
- [Scaffolding & External Dependencies](docs/standards/scaffolding.md)
- [MCP Server Safety](docs/standards/mcp-safety.md)
- [Workflow Discipline](docs/standards/workflow-discipline.md)
