# DNP3 — CURRENT Active Work

Repository: tamzrod/dnp3
Package: VEC-01-LAB-PREP-001
Branch: work
CURRENT: NONE
QUEUED: NONE
Status: WORK_PACKAGE_READY





## Package closure — VEC-01-LAB-PREP-001 (2026-11-08)
The OpenDNP3 laboratory preparation package is complete as **documentation
only**. It added `active_work/vec-01-opendnp3-lab-preparation.md`, a planning
placeholder covering peer provenance (independently implemented OpenDNP3 C++,
Apache-2.0, archived upstream, pinned tag 3.1.2), reproducible preparation
design, isolated non-production topology, the v0 configuration matrix, a future
test matrix (not execution), capture/artifact plan, and authorization gates. No
peer was cloned, built, installed or run; no DNP3 packets, capture, field
control or production interaction occurred. **VEC-01 remains NOT EXECUTED, R4
remains OPEN, and third-party interoperability is NOT PROVEN.** All mandatory
regression gates PASSED on the exact `work` tree at
`08d3882a3b63010e21be6d4021593dda0c5cc4e3` with a provisioned official Go
1.22.12 toolchain: `gofmt -l .` clean, `git diff --check`, `go vet ./...`,
`go build ./...`, `go test ./... -count=1`,
`go test -race ./internal/outstation/... -count=1`, `./scripts/verify-mvp.sh`,
and `./scripts/verify-external-mvp.sh` (Tiers 1-3). Evidence:
`docs/verification/VEC-01-LAB-PREP-001.md`; archive:
`archive/active-work/VEC-01-LAB-PREP-001.md`.

## Package closure — DNP3-060 (2026-11-08)
The MEXT-042 provenance correction is complete. DNP3-060 corrected the two
inaccurate PR #18 base-SHA labels in `docs/verification/MEXT-042.md` and
`archive/active-work/MEXT-042.md` from `95ac1028…` (which is actually merged
PR #17) to the correct merged PR #18 SHA
`b334c8cd3d8d1f84842c952c84ced4bef42efdd7` — exactly one line changed per file.
Historical MEXT-042 execution SHA, test exit results and the archived task body
are unchanged. All mandatory checks PASSED on the exact `work` tree at
`da00a24e42d93e0025033d32037421864cc6cf3a` with a provisioned official Go
1.22.12 toolchain: `gofmt -l .` clean, `git diff --check`, `go vet ./...`,
`go build ./...`, `go test ./... -count=1`,
`go test -race ./internal/outstation/... -count=1`, `./scripts/verify-mvp.sh`,
and `./scripts/verify-external-mvp.sh` (Tiers 1-3). Evidence:
`docs/verification/DNP3-060.md`; archive: `archive/active-work/DNP3-060.md`.
R4/VEC-01 remains OPEN.

## Package closure — MEXT-042 (2026-11-08)
The independent third-party capture procedure package is complete as
**documentation only**. MEXT-042 added `active_work/third-party-capture-procedure.md`,
a planning placeholder defining peer provenance, isolated topology, the v0-only
operation matrix, capture/artifact format, wire inspection checks, negative
cases, safe-stop and reviewer acceptance criteria. No capture was executed, no
harness or code was added, and no supported behavior changed. **VEC-01 remains
NOT EXECUTED, R4 remains OPEN, and third-party interoperability is NOT PROVEN.**
All mandatory checks PASSED on the exact `work` tree at
`4b6eecc81a7581677f91655c69d1d0050cfd8228` with a provisioned official Go
1.22.12 toolchain: `gofmt -l .` clean, `git diff --check`, `go vet ./...`,
`go build ./...`, `go test ./... -count=1`,
`go test -race ./internal/outstation/... -count=1`, `./scripts/verify-mvp.sh`,
and `./scripts/verify-external-mvp.sh` (Tiers 1-3). Evidence:
`docs/verification/MEXT-042.md`; archive: `archive/active-work/MEXT-042.md`.

## Execution stop
CURRENT is NONE and QUEUED is NONE. No implementation, protocol, profile or
governance task may run without a new human-promoted bounded CWAL microtask with
matching markers in this file and root `handoff.md`. Any actual VEC-01 campaign
requires a separately owner-approved task and safety gates. PR creation, merge,
release and field-device tests each require separate human authorization.

## Persistent working branch policy
The maintainer designates `work` as the sole persistent execution branch for the remaining active DNP3 roadmap. New CWAL tasks are promoted and closed on `work`; no task-specific branch is created unless the maintainer explicitly reverses this direction. Every new CURRENT still requires explicit human promotion and bounded scope. `main` is the PR integration baseline, not the CWAL execution branch.
