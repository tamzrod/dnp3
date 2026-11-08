# DNP3 — CURRENT Active Work

Repository: tamzrod/dnp3
Package: MEXT-POSTGATE-ROADMAP-001
Branch: work/mext-postgate-roadmap
CURRENT: MEXT-041
QUEUED: NONE
Status: TASK_READY

## Current package authority
The repository owner approved the **post-MEXT-035 documentation-only backlog package** on 2026-11-08. MEXT-040 (SBO profile deferral backlog stub) completed, was archived to `archive/active-work/MEXT-040.md`, and is recorded in `docs/verification/MEXT-040.md`. Under the same approved package, the already human-prepromoted MEXT-041 (events/unsolicited deferral backlog stub) is now promoted to CURRENT; its bounded task body is `workflow/active_work/MEXT-041.md`. MEXT-042 (third-party capture procedure) is not promoted and must not be started.

## Work branch policy
`work/mext-postgate-roadmap` is the persistent non-main execution branch for this post-gate roadmap series. A long-lived branch is **not** blanket authorization for future roadmap tasks. New packages require fresh human promotion and scope, regardless of branch reuse. No automatic PR, merge, main push, production deployment or field command.

## Dependency and merge hygiene
Started from `main` at `b38d7691aeb1adccf52fe665e368d4d77766977c` after PR #15 merged with passing main CI. Historical roadmap sync PR #16 (`docs/sync-mext-roadmap-post-pr15`) was **open** when this branch was prepared. If PR #16 merges later, reconcile base safely before PR creation; do not silently overwrite its roadmap/handoff changes. This package must NOT edit the MEXT roadmap or `active_work/handoff.md`.

## Support boundary
MEXT-035 recorded in-repo real-TCP v0 Master MVP completion. Unsupported SBO and unsolicited/event functionality remain deferred. No independent VEC-01/third-party interoperability proof or complete IEEE 1815 conformance claim. DNP3-056/057 correction tasks are completed and archived; do not reopen them.

## MEXT-040 checkpoint
MEXT-040 archived after verified documentation-only acceptance: the SBO deferral stub in `active_work/postgate-backlog.md` records SBO as DEFERRED / NOT SUPPORTED with no runtime, profile, or wire changes. Go build/test/vet were NOT_RUN (toolchain absent); the change is docs-only. Follow `operation cwal.md` for the MEXT-041 checkpoint and final package closure.
