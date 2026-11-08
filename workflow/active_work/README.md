# DNP3 — CURRENT Active Work

Repository: tamzrod/dnp3
Package: MEXT-THIRD-PARTY-CAPTURE-PROCEDURE-001
Branch: work
CURRENT: MEXT-042
QUEUED: NONE
Status: TASK_READY



## CURRENT — MEXT-042 (owner-promoted 2026-11-08)
MEXT-042 is the **documentation-only** independent third-party capture procedure placeholder, human-promoted after PR #18 merged. Permanent execution branch: `work`. Authoritative scope and executable verification: `workflow/active_work/MEXT-042.md`. **Do not execute third-party capture, issue live field controls, change supported behavior, or close R4/VEC-01**. Mandatory Go verification is executable and must PASS, including private Go installation fallback. No other CURRENT or QUEUED task is authorized. DNP3-059 remains closed and archived. This task is **TASK_READY only**, not implemented or checkpointed.


## Package closure
The post-PR #17 roadmap synchronization package is complete. DNP3-059 reconciled
the historical MEXT roadmap, the external-acceptance checklist and the
historical MEXT handoff with evidence already on `main`, without creating new
execution authority or changing code, profile, governance or verification
archives. It executed every mandatory Go gate on the exact branch tree at
`16f6f70ef70ab448a7b6a3a3b783902ee86f790e` with a provisioned official Go
1.22.12 toolchain: `gofmt -l .` clean, `git diff --check`, `go vet ./...`,
`go build ./...`, `go test ./... -count=1`,
`go test -race ./internal/outstation/... -count=1`, `./scripts/verify-mvp.sh`,
and `./scripts/verify-external-mvp.sh` (Tiers 1-3). Evidence is in
`docs/verification/DNP3-059.md`; the task is archived at
`archive/active-work/DNP3-059.md`. R4/VEC-01 third-party interoperability
remains OPEN; MEXT-042 remains unpromoted.

## Execution stop
CURRENT is NONE and QUEUED is NONE. No implementation, protocol, profile or
governance task may run without a new human-promoted bounded CWAL microtask with
matching markers in this file and root `handoff.md`. PR creation, merge, release
and field-device tests each require separate human authorization.

## Persistent working branch policy
The maintainer designates `work` as the sole persistent execution branch for the remaining active DNP3 roadmap. New CWAL tasks are promoted and closed on `work`; no task-specific branch is created unless the maintainer explicitly reverses this direction. Every new CURRENT still requires explicit human promotion and bounded scope. `main` is the PR integration baseline, not the CWAL execution branch.
