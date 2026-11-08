# DNP3 — CURRENT Active Work

Repository: tamzrod/dnp3
Package: MEXT-POSTGATE-ROADMAP-001
Branch: work/mext-postgate-roadmap
CURRENT: NONE
QUEUED: NONE
Status: WORK_PACKAGE_READY

## Package closure
The post-MEXT-035 documentation-only backlog package is complete. MEXT-040 (SBO
profile deferral backlog stub) and MEXT-041 (events/unsolicited deferral backlog
stub) both passed documentation-only acceptance, were archived under
`archive/active-work/`, and have evidence in `docs/verification/MEXT-040.md` and
`docs/verification/MEXT-041.md`. Both stubs live in
`active_work/postgate-backlog.md` and record their capabilities as
DEFERRED / NOT SUPPORTED. MEXT-042 (third-party capture procedure) remains
unpromoted and requires separate human approval. `CURRENT: NONE` forbids further
implementation until a human promotes a new package.

## Work branch policy
`work/mext-postgate-roadmap` is the persistent non-main execution branch for this
post-gate roadmap series. A long-lived branch is **not** blanket authorization
for future roadmap tasks. New packages require fresh human promotion and scope,
regardless of branch reuse. No automatic PR, merge, main push, production
deployment or field command.

## Dependency and merge hygiene
Started from `main` at `b38d7691aeb1adccf52fe665e368d4d77766977c` after PR #15
merged with passing main CI. Historical roadmap sync PR #16
(`docs/sync-mext-roadmap-post-pr15`) was **open** when this branch was prepared.
If PR #16 merges later, reconcile base safely before PR creation; do not silently
overwrite its roadmap/handoff changes. This package must NOT edit the MEXT roadmap
or `active_work/handoff.md`.

## Support boundary
MEXT-035 recorded in-repo real-TCP v0 Master MVP completion. Unsupported SBO and
unsolicited/event functionality remain deferred. No independent VEC-01/third-party
interoperability proof or complete IEEE 1815 conformance claim. DNP3-056/057
correction tasks are completed and archived; do not reopen them.

## Verification state
PASS: `bash scripts/cwal-check.sh`, `git diff --check`, documentation scope
checks. NOT_RUN: `go test ./... -count=1`, `go vet ./...` (Go toolchain absent;
changes are docs-only). See `docs/verification/MEXT-040.md` and
`docs/verification/MEXT-041.md` for exact commands and results.

## Remaining limitations
No third-party interoperability/capture evidence exists; no IEEE 1815 conformance
claim is made. PR creation, merge, release and field-device tests each require
separate human authorization.
