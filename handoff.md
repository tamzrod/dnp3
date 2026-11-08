# DNP3 — CURRENT CWAL Handoff

Repository: tamzrod/dnp3
Package: DNP3-MEXT042-PROVENANCE-CORRECTION-001
Branch: work
CURRENT: DNP3-060
QUEUED: NONE
Status: TASK_READY




## CURRENT DNP3-060 — limited historical evidence correction
Owner-promoted on 2026-11-08, permanent branch `work`. The only active task is `workflow/active_work/DNP3-060.md`: correct the two inaccurate PR #18 base SHA labels in MEXT-042 evidence/archive without changing historical test results. The underlying MEXT-042 procedure remains completed and third-party VEC-01/R4 remains OPEN. Every mandated Go check must actually run and PASS, including private Go installation fallback. No other implementation, PR or merge is authorized by this promotion; this is TASK_READY only, not a completed correction.

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
matching markers in `workflow/active_work/README.md` and this file. Any actual
VEC-01 campaign requires a separately owner-approved task and safety gates. PR
creation, merge, release and field-device tests each require separate human
authorization.

## Persistent branch rule (owner instruction, 2026-11-08)
Use `work` for the remaining active roadmap, including subsequent separately human-approved CWAL packages. Do not switch branches by package. A stable branch name does not bypass task authorization, PR audit against `main`, CI, or merge approval. The superseded `docs/sync-mext-post-pr17-20261108` ref is not the active CWAL branch; retire it after confirming the new `work` tip.
