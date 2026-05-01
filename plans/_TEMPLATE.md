---
plan-id: short-slug-matching-filename
status: queued        # queued | active | blocked | complete
owner: unassigned     # agent name or "human"
branch: plan/<plan-id>
depends-on: []        # other plan-ids that must be complete first
last-touched: YYYY-MM-DD
---

# Plan: <Human-readable title>

## Goal

One paragraph. What does "done" look like? A reader who has never seen this codebase should understand the outcome.

## Scope

**In scope:**
- specific outcome 1
- specific outcome 2

**Out of scope (do not change in this plan):**
- list anything that's adjacent and tempting but belongs elsewhere

## File envelope

**Touched (modify freely within this plan):**
- `src/foo/bar.go`
- `src/baz/qux.go`

**Forbidden (do not modify in this plan, even if it would be convenient):**
- everything not in Touched, especially:
- `src/coords/` — owned by `OFFSET_UNIFICATION_PLAN.md`
- any package not relevant to this plan's goal

## Acceptance criteria

The plan is complete when every box here is checked.

- [ ] Testable claim 1 (something a reader could verify by reading code or running a test)
- [ ] Testable claim 2
- [ ] `cd src && go build ./...` passes
- [ ] `cd src && go test ./...` passes (skip if no tests exist for touched packages)

## Phases

Each phase is a single session's worth of work. Sub-tasks are checkboxes. An agent should be able to close every checkbox in one phase without leaving the file envelope.

### Phase 1: <name>
- [ ] 1.1 task description with file path
- [ ] 1.2 task description

### Phase 2: <name>
- [ ] 2.1 task description

### Phase N: Cleanup
- [ ] N.1 mark roadmap rows ✅ in `design-docs/roadmap.md`
- [ ] N.2 update `CLAUDE.md` Phase status block if needed
- [ ] N.3 move this plan to `plans/COMPLETED/`

## Progress log

Append-only. Newest at the bottom. One row per session, not per commit.

| Date | Phase | Status | Notes |
|------|-------|--------|-------|

## What was NOT changed (intentional)

Append-only. When you defer or skip a sub-task, document it here so the next agent doesn't redo your reasoning.

_None yet._

## Open questions

Add when you encounter a decision the plan doesn't answer. Include your recommendation and which sub-task is blocked.

_None yet._
