# MEXT-040 — Post-gate SBO profile backlog stub

Status: CURRENT — HUMAN_PROMOTED
Package: MEXT-POSTGATE-ROADMAP-001
Branch: work/mext-postgate-roadmap
Human approver: repository owner (explicit instruction in conversation on 2026-11-08)
KDE design approval: Not applicable to this documentation-only deferral. **No SBO design or implementation is approved.**
Roadmap: `active_work/MEXT_MASTER_ROADMAP.md` Phase 4, MEXT-040
Predecessor: MEXT-035 completed; in-repo external-MVP verified; independent third-party R4/VEC-01 still open.

## Objective
Create a **backlog stub only** for a future Select-Before-Operate (SBO) profile. Explicitly defer implementation and any unsupported capability claim, document questions that would require an approved future design, and retain the v0 Direct Operate only supported-profile boundary.

## Minimal context
- `active_work/MEXT_MASTER_ROADMAP.md` → Phase 4 / MEXT-040 excerpt (read only)
- `active_work/supported-profile.md` → supported and rejected Master control modes (read only)
- `docs/architecture/` or KDE governance → only the relevant approval rule, if necessary
- `active_work/postgate-backlog.md` → create/update SBO section (only implementation artifact)

## Allowed edits
- `active_work/postgate-backlog.md` (new document; SBO section only)
- `docs/verification/MEXT-040.md` (actual evidence, including docs checks)
- `workflow/active_work/README.md`, `handoff.md` (matching CWAL markers and concise handoff)
- `archive/active-work/MEXT-040.md` (archive on success only; remove active task only after archive)

## Scope and acceptance
1. Document SBO as **DEFERRED / NOT SUPPORTED** in a distinct MEXT-040 section; describe prerequisites and design decisions to be researched (Select→Operate transaction, command status and timeout handling, sequence/state, negative cases).
2. Keep the **current v0 supported profile unchanged**. Clarify that SBO behavior is *not* implemented, verified or approved.
3. Reference future required independent design approval, protocol specification evidence, golden-vector tests, timeout/disconnect/ambiguous-command negatives and non-production validation — planning only.
4. Do not modify unrelated backlog sections (MEXT-041 is QUEUED, MEXT-042 unpromoted).
5. Verify markdown scope with `git diff --check`, inspect `git diff --name-only`, `git status --short`, and confirm both marker blocks match. Record PASS/FAIL/NOT_RUN truthfully; optionally run `go test ./... -count=1` and `go vet ./...` if the execution environment supports it, but do not claim these tests ran unless they did.
6. Archive completed CURRENT, add verification evidence, push and read back remote SHA. **Only after the MEXT-040 remote checkpoint is verified**, and only under the already approved package, the executor may promote QUEUED MEXT-041 to CURRENT by changing both marker files atomically in one checkpoint. If any gate fails, retain MEXT-040 CURRENT and stop as TASK_BLOCKED.

## Invariants / negative cases
- No changes to Go files, supported profile, DNP3 wire framing, control behavior, CI, dependencies, architecture approvals or live SCADA/field-device connections.
- No Select-Before-Operate control request may be sent; no third-party interoperability/full IEEE 1815 conformance claims.
- Stop and report on stale remote branch, dirty checkout, mismatched markers, pending conflicting PR #16 changes, or documentation scope expansion.

## CWAL completion boundary
Once MEXT-040 passes, checkpoint and then promote the *already human-prepromoted* MEXT-041 per `operation cwal.md`. Never silently start MEXT-042. No PR creation, merge, release or direct main write is authorized by this microtask.
