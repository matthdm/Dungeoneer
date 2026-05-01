# Agent Session Protocol

> Every agent — human or AI — that picks up work on this repo follows this ritual. It is short on purpose. The plan doc carries the rest.

## Session start

1. Read `/CLAUDE.md`.
2. Open `plans/_QUEUE.md`. Find the **Active plan**.
3. Read the active plan **in full**, including its Progress Log and "What was NOT changed (intentional)" sections.
4. Run `cd src && go build ./...` to confirm a clean baseline.
5. State in chat which checkboxes you intend to close in this session and why.

If `_QUEUE.md` has no active plan, your first job is to promote the highest-priority queued plan to active. Do not start unstructured work.

## Session end

1. `cd src && go build ./...` must pass. If it doesn't, your session is not done — fix it or back the change out.
2. `cd src && go test ./...` must pass when tests exist for the touched packages. If you added tests, they must run.
3. Update the plan's **Progress log** table with date, phase touched, status, and a one-line note.
4. If you skipped or deferred a sub-task, add an entry to **What was NOT changed (intentional)** explaining why. The next agent must not re-litigate it.
5. Commit on the plan's branch with format:

   ```
   [<plan-id>] <phase>: <imperative summary>
   ```

   Example: `[4F-loot-refinement] phase-1: add ChestVariant enum and Locked field`.

6. If the plan's Acceptance criteria are now all checked: move the plan doc to `plans/COMPLETED/`, update `_QUEUE.md` (active → next, append to Completed), and update the Phase status block in `CLAUDE.md`.

## Branching

- One branch per plan: `plan/<plan-id>`.
- Don't rebase across sessions; the progress log lives in the branch's history.
- Squash-merge to `main` only when the plan is fully complete.

## Scope discipline

- Touch only files listed under the plan's **File envelope: Touched**.
- If you discover a needed change outside the envelope, stop. Add it to **Open questions** with your recommended fix and stop work on the affected sub-task. Do not expand scope.
- If the plan is wrong, fix the plan first (in a separate commit), then resume work.

## Escalation

- If a sub-task requires a design decision the plan doesn't answer, add it to **Open questions** with your recommendation and stop work on that sub-task only. Continue with other unblocked work.
- Never fabricate intent. The plan is the contract; if the contract is silent, escalate.

## Spawning new plans

A new plan doc is required when:
- The work would span more than one session, OR
- The work touches three or more packages, OR
- The work is on the roadmap (`design-docs/roadmap.md`) but isn't covered by an existing plan.

Copy `_TEMPLATE.md`, fill it in, add an entry to `_QUEUE.md`. Don't start coding until the plan is reviewed (by a human or by your own honest read of whether the file envelope holds).
