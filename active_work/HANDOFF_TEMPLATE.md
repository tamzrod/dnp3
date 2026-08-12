# Handoff Template (MEXT Series)

Copy forward from `active_work/handoff.md` after each micro-task.
Roadmap: `active_work/MEXT_MASTER_ROADMAP.md`

Keep entries concise and factual. The repository + handoff are the source of
truth. Do not assume a task failed just because an agent hung — verify from
git + this file.

## Status

- Last completed task: <!-- MEXT-NNN (title) -->
- Last checkpoint commit: <!-- <sha> (pushed to origin/main) -->
- Current task: <!-- MEXT-NNN (title) — IN PROGRESS / BLOCKED / DONE-uncommitted -->
- External MVP status: NOT COMPLETE until MEXT-035 (internal MVP archived at DNP3-056).
- Test status: <!-- verify-mvp.sh exit 0 / verify-external-mvp.sh exit 0 / failing -->

## Completed Tasks

- MEXT-NNN — <!-- one-line summary; tests: <names>; acceptance: <criterion> — met -->

## Current Checkpoint Batch

- [ ] MEXT-NNN — <!-- title -->
- [ ] MEXT-NNN — <!-- title -->
- [ ] MEXT-NNN — <!-- title -->

## Next READY Tasks

- MEXT-NNN — <!-- title --> (prereq MEXT-NNN, done)
- MEXT-NNN — <!-- title -->

## Recommended Next Task

**MEXT-NNN — <!-- title -->**. <!-- one-sentence what + why. -->

If blocked, fall back to **MEXT-NNN — <!-- title -->**.

## Test Commands (baseline)

```bash
export PATH=$HOME/go-install/go/bin:$PATH
go test ./...
go test -race ./internal/master/... ./pkg/dnp3/... ./test/integration/...
./scripts/verify-mvp.sh
# ./scripts/verify-external-mvp.sh   # after MEXT-021/033
```

## Code State (this batch)

- `path/to/file.go`: <!-- what changed -->

## Implementation Discoveries

- <!-- discoveries -->

## Blockers / Risks

- None.

## Next Action

1. Verify `git status` + `git log --oneline -5` match the checkpoint above.
2. Read this file + `active_work/MEXT_MASTER_ROADMAP.md`.
3. Implement the Recommended Next Task; run its tests; update this file.
4. Every 3 completed tasks: verify → update this file → commit → push to main.
