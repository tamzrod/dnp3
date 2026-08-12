# Implementation Handoff (TEMPLATE)

<!--
DNP3-072 — Master handoff.md template for implementers.

Purpose: a canonical, copy-paste skeleton an implementer agent fills in so the
NEXT session can recover state from the repository alone (this file + git +
the roadmap), without relying on a prior conversation.

Usage:
  cp active_work/HANDOFF_TEMPLATE.md active_work/handoff.md
  # then fill in the bracketed placeholders below.

Keep entries concise and factual. The repository + handoff are the source of
truth. Do not assume a task failed just because an agent hung — verify from
git + this file.
-->

## Status

- Last completed task: <!-- DNP3-NNN (title) -->
- Last checkpoint commit: <!-- <sha> (pushed to origin/main) -->
- Current task: <!-- DNP3-NNN (title) — IN PROGRESS / BLOCKED / DONE-uncommitted -->
- MVP status: MVP COMPLETE at DNP3-056 (internal verification; external VEC-01 pending).
- Test status: <!-- full suite green / -race green / verify-mvp.sh exit 0 / failing (which) -->

## Completed Tasks

<!-- One line per completed micro-task, newest last. Group by checkpoint batch. -->
- DNP3-NNN — <!-- one-line summary; tests: <names>; acceptance: <criterion> — met -->

## Current Checkpoint Batch

<!-- The 3 micro-tasks forming the in-progress checkpoint (commit every 3). -->
- [ ] DNP3-NNN — <!-- title -->
- [ ] DNP3-NNN — <!-- title -->
- [ ] DNP3-NNN — <!-- title -->

## Next READY Tasks

<!-- The next implementable micro-tasks (prerequisites satisfied). -->
- DNP3-NNN — <!-- title --> (prereq DNP3-NNN, done)
- DNP3-NNN — <!-- title -->

## Recommended Next Task

**DNP3-NNN — <!-- title -->**. <!-- one-sentence what + why. -->

If blocked, fall back to **DNP3-NNN — <!-- title -->**.

## Test Commands (baseline)

```bash
export PATH=$HOME/go-install/go/bin:$PATH   # Go toolchain location
go test ./...
go test -race ./internal/master/... ./pkg/dnp3/... ./test/integration/...
./scripts/verify-mvp.sh    # single pre-merge gate
```

## Code State (this batch)

<!-- Which files changed and the essential shape of the change. -->
- `path/to/file.go`: <!-- what changed -->
- `path/to/file_test.go` (new): <!-- test names -->

## Implementation Discoveries

<!-- Non-obvious invariants, gotchas, spec clarifications discovered. -->
- <!-- e.g. DNP3 wire fields are LSB-first; frame.EncodedSize = HeaderSize + dataLen + 2*ceil(dataLen/16). -->

## Blockers / Risks

<!-- Genuine blockers requiring planner/architect input. Leave "None." if clear. -->
- None.

## Next Action

<!-- The single concrete next step the next agent should take. -->
1. Verify `git status` + `git log --oneline -5` match the checkpoint above.
2. Read this file + `active_work/DNP3_MASTER_ROADMAP.md`.
3. Implement the Recommended Next Task; run its tests; update this file.
4. Every 3 completed tasks: verify → update this file → commit → push to main.
