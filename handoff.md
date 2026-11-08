# DNP3 — CURRENT CWAL Handoff

Repository: tamzrod/dnp3
Package: VEC-01-LAB-PREP-001
Branch: work
CURRENT: VEC-01-LAB-PREP-001
QUEUED: NONE
Status: TASK_READY





## CURRENT — VEC-01-LAB-PREP-001 (human-promoted 2026-11-08)
The **only** active bounded task is `workflow/active_work/VEC-01-LAB-PREP-001.md` on permanent branch `work`. Scope: planning-only independent OpenDNP3 peer preparation, isolated non-production topology, configuration and future validation/evidence plan. No clone, build, installation, peer execution, DNP3 packets, network capture, field-device control, implementation, or PR is authorized. Mandatory existing Go regression verification must actually PASS. VEC-01 is NOT EXECUTED; R4 remains OPEN. This promotion records TASK_READY, not implementation or completion.

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
matching markers in `workflow/active_work/README.md` and this file. Any actual
VEC-01 campaign requires a separately owner-approved task and safety gates. PR
creation, merge, release and field-device tests each require separate human
authorization.

## Persistent branch rule (owner instruction, 2026-11-08)
Use `work` for the remaining active roadmap, including subsequent separately human-approved CWAL packages. Do not switch branches by package. A stable branch name does not bypass task authorization, PR audit against `main`, CI, or merge approval. The superseded `docs/sync-mext-post-pr17-20261108` ref is not the active CWAL branch; retire it after confirming the new `work` tip.
