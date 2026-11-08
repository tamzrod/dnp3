# Verification — MEXT-042

Package: MEXT-THIRD-PARTY-CAPTURE-PROCEDURE-001
Branch: work
Task: MEXT-042 — Independent third-party DNP3 capture procedure placeholder
Execution commit SHA: `4b6eecc81a7581677f91655c69d1d0050cfd8228`
Base SHA (merged PR #18 `main`): `95ac10289ead46c09264a9da2bb877f0e0507410`
Promotion baseline: `b334c8cd3d8d1f84842c952c84ced4bef42efdd7` (per task body)

## Purpose

Documentation-only. MEXT-042 writes a **procedure placeholder** for an
independent third-party DNP3 v0 interoperability campaign (VEC-01 / residual
R4). It defines prerequisites, reproducible collection steps, capture
artifacts, pass/fail criteria and evidence provenance. It does **not** execute
a capture, implement a harness, add a working integration, send controls, or
change any supported behavior. VEC-01 remains NOT EXECUTED, R4 remains OPEN,
and third-party interoperability is NOT PROVEN.

## Deliverable

- `active_work/third-party-capture-procedure.md` (new, planning only). Sections:
  purpose/non-goals; independent peer selection and provenance; isolated
  non-production topology; IP/port/address configuration; v0-only operation
  matrix; capture collection/clock correlation/artifact format; wire inspection
  checks (CRC, LSB-first, SEQ/CON, application status); deterministic negative
  cases; bounded timing and safe stop; evidence provenance/integrity; reviewer
  acceptance/rejection mapping; authorization gates; limitations.
- No Go source/tests, `.kde` governance, architecture, `go.mod`/`go.sum`/Go
  workspace, `active_work/supported-profile.md`, existing capture fixtures,
  scripts, CI, README claims, roadmap status, release files, or historical
  task/evidence files were changed.
- Lifecycle-only changes for closure: `docs/verification/MEXT-042.md`,
  `archive/active-work/MEXT-042.md`, removal of `workflow/active_work/MEXT-042.md`,
  and matched marker updates in `workflow/active_work/README.md` and root
  `handoff.md`.

## Toolchain provisioning

- `go.mod` / `go.work` require `go 1.22`; CI (`.github/workflows/ci.yml`) uses
  `go-version: '1.22'`.
- No system Go was installed (`command -v go` empty). Provisioned privately per
  the task's fallback (no root required):
  - Downloaded the official archive `go1.22.12.linux-amd64.tar.gz` from
    `https://go.dev/dl/`.
  - Verified SHA-256 against official go.dev metadata:
    `4fa4f869b0f7fc6bb1eb2660e74657fbf04cdd290b5aef905585c86051b34d43` — match.
  - Extracted to `$HOME/sdk/go`; prepended `$HOME/sdk/go/bin` to `PATH`.
- `go version` → `go version go1.22.12 linux/amd64`.
- Provisioning modified no tracked repository file; `go.sum`/`go.work.sum`
  unchanged.

## Mandatory checks — all executed, all PASS

Environment: Linux x86_64, Go 1.22.12, execution SHA
`4b6eecc81a7581677f91655c69d1d0050cfd8228`.

| # | Command | Result | Exit |
|---|---------|--------|------|
| 1 | `bash scripts/cwal-check.sh` | PASS at startup on the clean, remote-synchronized checkout (`CWAL CHECK OK: CURRENT=MEXT-042 branch=work`) | 0 |
| 2 | `test -z "$(gofmt -l .)"` | PASS — empty output | 0 |
| 3 | `git diff --check` | PASS — no whitespace/conflict errors | 0 |
| 4 | `go version` | `go1.22.12 linux/amd64` | 0 |
| 5 | `go vet ./...` | PASS — no findings | 0 |
| 6 | `go build ./...` | PASS | 0 |
| 7 | `go test ./... -count=1` | PASS — all listed packages `ok` | 0 |
| 8 | `go test -race ./internal/outstation/... -count=1` | PASS — `ok dnp3/internal/outstation` | 0 |
| 9 | `./scripts/verify-mvp.sh` | PASS — vet + unit/integration + race all green | 0 |
| 10 | `./scripts/verify-external-mvp.sh` | PASS — Tiers 1–3 green; VEC-01 advisory only | 0 |

Note: `scripts/cwal-check.sh` requires a clean working tree, so it necessarily
FAILS while the new uncommitted deliverable is present (mid-preparation). It
PASSED on the clean checkout before editing and is re-run on the clean tree
after the checkpoint commit; this is expected, not a skipped check. All nine
other mandatory checks ran on the exact `work` tree and PASSED.

Selected package results from check 7 (all `ok`): `internal/al`,
`internal/dll/crc`, `internal/dll/frame`, `internal/dll/link`, `internal/master`,
`internal/outstation`, `internal/sa*`, `internal/testutils*`, `internal/tl`,
`pkg/dnp3*`, `pkg/transport`, `test/conformance/*`, `test/integration`.

## Acceptance mapping

1. Planning-only status stated up front: **VEC-01 NOT EXECUTED / R4 OPEN /
   THIRD-PARTY INTEROP NOT PROVEN**; the document explicitly cannot satisfy
   external acceptance.
2. Independent peer must be truly independently implemented; a second copy of
   this library / its own outstation is explicitly recorded as INELIGIBLE, with
   version/license/provenance fields required.
3. Scope restricted to v0 TCP and static groups, Direct-Operate G12V1 only,
   against non-production simulated outputs with prior separate permission; no
   live field control, production peer, serial/TLS/SBO, unsolicited events or
   unsupported groups.
4. Evidence collection distinguished from acceptance; command/status checks
   (including ambiguous outcomes and never-blind-retry) and evidence-integrity
   checks defined; fabrication of PCAP/results explicitly prohibited.
5. Failure/timeout/capture-incomplete disposition and bounded non-production
   safe-stop criteria included; no unsupported feature becomes supported.
6. No claim of certified IEEE 1815 conformance, production readiness or
   universal DNP3 compatibility.
7. All six CWAL authority fields match; only allowlisted Markdown files changed;
   clean, reproducible remote checkpoint and accurate verification record below.

## Scope / hygiene

- `git status --porcelain` clean before edits (after fast-forward to `4b6eecc`).
- New deliverable: `active_work/third-party-capture-procedure.md` only.
- `git diff --check` PASS; `gofmt -l .` empty.

## Limitations / non-claims

- Documentation reconciliation only; no functionality added or changed.
- `verify-external-mvp.sh` PASS covers in-repo Tiers 1–3 (real-TCP MVP, operate
  matrix, multi-header) and is never VEC-01 third-party proof.
- VEC-01 is NOT EXECUTED; R4 remains OPEN; no third-party interoperability or
  IEEE 1815 conformance claim is made.
- PR creation, merge, release, third-party lab execution and field-device tests
  each require separate human authorization.
