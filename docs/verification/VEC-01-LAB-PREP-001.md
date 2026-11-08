# Verification — VEC-01-LAB-PREP-001

Package: VEC-01-LAB-PREP-001
Branch: work
Task: VEC-01-LAB-PREP-001 — Independent OpenDNP3 peer laboratory preparation plan
Execution commit SHA: `08d3882a3b63010e21be6d4021593dda0c5cc4e3`
Promotion baseline: `f165e8509436153dfa53ab68bb67c51be9a88b10` (merged PR #19)

## Purpose

Documentation-only. VEC-01-LAB-PREP-001 writes a **planning-only** laboratory
preparation document describing how an independent, pinned **OpenDNP3** C++
outstation could later be prepared for VEC-01 interoperability testing against
the `tamzrod/dnp3` Go master. It is not a deployment and not an executed test.
No peer was cloned, downloaded, built, installed or run; no DNP3 packets, network
capture, or field control occurred. **VEC-01 remains NOT EXECUTED; R4 remains
OPEN; third-party interoperability is NOT PROVEN.**

## Deliverable

- `active_work/vec-01-opendnp3-lab-preparation.md` (new, planning only) covering
  all seven required areas: (1) independent peer provenance; (2) reproducible
  preparation design; (3) isolation design; (4) configuration matrix; (5) future
  test matrix (not execution); (6) capture and artifact plan; (7) authorization
  gates and risks. It is explicitly labeled **PLANNING ONLY / VEC-01 NOT
  EXECUTED / R4 OPEN / THIRD-PARTY INTEROPERABILITY NOT PROVEN**.
- No Go source/tests, fixtures, `go.mod`/`go.sum`, CI/scripts, Compose files,
  peer binaries, dependency locks, `.kde` governance, roadmap, supported profile,
  historical verification, or README claims were changed.
- Lifecycle-only changes for closure: `docs/verification/VEC-01-LAB-PREP-001.md`,
  `archive/active-work/VEC-01-LAB-PREP-001.md`, removal of
  `workflow/active_work/VEC-01-LAB-PREP-001.md`, and matched marker updates in
  `workflow/active_work/README.md` and root `handoff.md`.

## Provenance facts verified (read-only)

Facts about the candidate peer were verified via the public GitHub API
(read-only) rather than asserted from memory:

| Field | Verified value |
|-------|----------------|
| Upstream | `https://github.com/dnp3/opendnp3` (formerly `automatak/dnp3`, repo id 7551306) |
| License | Apache-2.0 (SPDX `Apache-2.0`) |
| Maintenance | `archived: true`; last push `2022-05-18`; no open issues |
| Latest release | tag `3.1.2`, published `2022-04-22` |
| Pinned source SHA | `26b4c01e4839bbbda8866655e086471c4917ee53` (tag `3.1.2`) |
| Lineage | Independently implemented C++ stack, distinct from this Go repository |

The release tarball SHA-256 is intentionally **not** recorded (it would require a
download, which this task prohibits); the plan directs recording it at future
acquisition time. No release/variation/feature support is claimed as verified.

## Toolchain provisioning

- `go.mod` / `go.work` require `go 1.22`; CI uses `go-version: '1.22'`.
- No system Go was present. Provisioned privately per the task fallback (no root):
  downloaded the official archive `go1.22.12.linux-amd64.tar.gz` from
  `https://go.dev/dl/`, verified SHA-256
  `4fa4f869b0f7fc6bb1eb2660e74657fbf04cdd290b5aef905585c86051b34d43` (match),
  extracted to `$HOME/sdk/go`, prepended `$HOME/sdk/go/bin` to `PATH`.
- `go version` → `go version go1.22.12 linux/amd64`.
- Provisioning modified no tracked repository file; `go.sum`/`go.work.sum`
  unchanged.

## Mandatory executable regression gates — all executed, all PASS

Environment: Linux x86_64, Go 1.22.12, execution SHA
`08d3882a3b63010e21be6d4021593dda0c5cc4e3`.

| # | Command | Result | Exit |
|---|---------|--------|------|
| 1 | `bash scripts/cwal-check.sh` | PASS on clean startup (`CWAL CHECK OK: CURRENT=VEC-01-LAB-PREP-001 branch=work`); re-run PASS post-commit | 0 |
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
FAILS while the new uncommitted plan is present (mid-edit dirty tree); that is
**not** an authorized PASS gate. It PASSED on the clean startup checkout and is
re-run on the clean tree after the checkpoint commit. All other mandatory gates
ran on the exact `work` tree and PASSED. The external MVP gate's Tiers 1–3 are
in-repo real-TCP tests and are **not** third-party VEC-01 proof.

Selected package results from check 7 (all `ok`): `internal/al`,
`internal/dll/crc`, `internal/dll/frame`, `internal/dll/link`, `internal/master`,
`internal/outstation`, `internal/sa*`, `internal/testutils*`, `internal/tl`,
`pkg/dnp3*`, `pkg/transport`, `test/conformance/*`, `test/integration`.

## Acceptance mapping

1. Standalone plan covers all seven deliverable requirements and is explicitly
   labeled planning-only with VEC-01 NOT EXECUTED / R4 OPEN / interop not proven.
2. OpenDNP3 is independently sourced (C++ lineage, Apache-2.0, archived upstream)
   and reviewed at the plan level; unverified release/variation/setup behavior is
   marked `TO VERIFY`; no invented hashes or licensing terms.
3. Zero peer deployment, capture, field control, or production network
   interaction; all changed tracked paths are in the allowlist; the six CWAL
   authority fields match across both state files.
4. All mandatory regression gates actually PASS with new evidence recorded here.
5. On completion the task body is archived verbatim, evidence added, and both
   markers closed to `CURRENT: NONE`, `QUEUED: NONE`,
   `Status: WORK_PACKAGE_READY`, committed and pushed only to `work`.

## Limitations / non-claims

- Documentation-only; no functionality, peer, lab, or capture was created.
- VEC-01 is NOT EXECUTED; R4 remains OPEN; no third-party interoperability,
  IEEE 1815 conformance, certification or production-readiness claim is made.
- `verify-external-mvp.sh` PASS covers in-repo Tiers 1–3 only and is never VEC-01
  evidence.
- Any future lab deployment and VEC-01 execution require separate human-promoted
  tasks and safety gates; PR creation, merge, release and field-device tests each
  require separate human authorization.
