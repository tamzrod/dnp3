# DNP3 — CURRENT Active Work

Repository: tamzrod/dnp3
Package: MEXT-POSTGATE-ROADMAP-001
Branch: work/mext-postgate-roadmap
CURRENT: MEXT-040
QUEUED: MEXT-041
Status: TASK_READY

## Current package authority
The repository owner approved the **post-MEXT-035 documentation-only backlog package** on 2026-11-08. MEXT-040 (SBO profile deferral backlog stub) is the single CURRENT task. Its bounded task body is `workflow/active_work/MEXT-040.md`. MEXT-041 (events/unsolicited deferral backlog stub) is already human-prepromoted as QUEUED but **must not be read or executed** until MEXT-040 is verified and remotely checkpointed. MEXT-042 (third-party capture procedure) is not promoted.

## Work branch policy
`work/mext-postgate-roadmap` is the persistent non-main execution branch for this post-gate roadmap series. A long-lived branch is **not** blanket authorization for future roadmap tasks. New packages require fresh human promotion and scope, regardless of branch reuse. No automatic PR, merge, main push, production deployment or field command.

## Dependency and merge hygiene
Started from `main` at `b38d7691aeb1adccf52fe665e368d4d77766977c` after PR #15 merged with passing main CI. Historical roadmap sync PR #16 (`docs/sync-mext-roadmap-post-pr15`) was **open** when this branch was prepared. If PR #16 merges later, reconcile base safely before PR creation; do not silently overwrite its roadmap/handoff changes. This package must NOT edit the MEXT roadmap or `active_work/handoff.md`.

## Support boundary
MEXT-035 recorded in-repo real-TCP v0 Master MVP completion. Unsupported SBO and unsolicited/event functionality remain deferred. No independent VEC-01/third-party interoperability proof or complete IEEE 1815 conformance claim. DNP3-056/057 correction tasks are completed and archived; do not reopen them.

## Initial state
Only task preparation has been committed; **no MEXT-040 execution or verification has run**. Run `bash scripts/cwal-check.sh` on a clean checked-out branch, fetch and confirm exact remote SHA, then execute only CURRENT. Follow `operation cwal.md` for checkpoint, queue promotion, and final package closure.
