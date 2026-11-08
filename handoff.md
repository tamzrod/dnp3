# DNP3 — CURRENT CWAL Handoff

Repository: tamzrod/dnp3
Package: DNP3-POSTGATE-VERIFICATION-CORRECTION-001
Branch: work/mext-postgate-roadmap
CURRENT: DNP3-058
QUEUED: NONE
Status: TASK_READY

## Current corrective scope
DNP3-058 is an explicitly human-approved, verification-only correction to the MEXT-040/041 pre-PR audit. Previous docs-only tasks were archived and remotely checkpointed at `16ef1f7278b44f6a005edcfe076b376c0551bb38`, but Go checks were NOT_RUN in that executor. The owner requires actual tests, not an inferred pass. Task authority: `workflow/active_work/DNP3-058.md`.

## Executor requirements
Fetch safely; verify clean local/remote SHA and matching markers; run `bash scripts/cwal-check.sh`. Check `go version`; provision a compatible official Go toolchain in the executor/container if missing, without touching tracked repository files. Actually execute gofmt, diff check, `go vet ./...`, `go build ./...`, `go test ./... -count=1`, `go test -race ./internal/outstation/... -count=1`, `./scripts/verify-mvp.sh`, and `./scripts/verify-external-mvp.sh`. Record truthfully in `docs/verification/DNP3-058.md`. If any cannot run or fails, keep DNP3-058 CURRENT, mark TASK_BLOCKED, report and stop; no implementation fixes are authorized.

## Completed historical package and limits
MEXT-040/041 are still complete, their archival records and original NOT_RUN evidence remain unaltered. The new DNP3-058 report supplements them with actual Go verification. SBO and events/unsolicited remain DEFERRED / NOT SUPPORTED. Independent VEC-01 / third-party compatibility evidence is absent. MEXT-042 is unpromoted.

## Next action / stop
Execute CURRENT only, close with archive, matching `CURRENT: NONE` / `QUEUED: NONE` / `Status: WORK_PACKAGE_READY` only after every mandated check passes and remote checkpoint is confirmed. No PR, merge, write to main, deploy, field command or broader code change.
