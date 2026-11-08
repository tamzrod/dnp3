<!-- Archived after verified acceptance. Original task body preserved verbatim below. -->

Archived: 2026-11-08
Execution commit SHA: 4b6eecc81a7581677f91655c69d1d0050cfd8228
Base SHA (merged PR #18 main): b334c8cd3d8d1f84842c952c84ced4bef42efdd7
Verification evidence: docs/verification/MEXT-042.md
Go toolchain: go1.22.12 linux/amd64 (private install $HOME/sdk/go; official archive sha256 4fa4f869b0f7fc6bb1eb2660e74657fbf04cdd290b5aef905585c86051b34d43)

# MEXT-042 — Independent third-party DNP3 capture procedure placeholder

Status: CURRENT — HUMAN_PROMOTED
Package: MEXT-THIRD-PARTY-CAPTURE-PROCEDURE-001
Branch: work
Owner authorization: explicit user selection to human-promote MEXT-042 on 2026-11-08
Design approval: N/A for documentation-only, non-executable procedure planning; no live third-party tests or new supported protocol behavior approved
Predecessor: MEXT-035 in-repo v0 real-TCP gate complete; MEXT-040/041 deferral stubs and DNP3-059 roadmap sync completed; PR #18 merged into main at `b334c8cd3d8d1f84842c952c84ced4bef42efdd7`.
Promotion baseline: `b334c8cd3d8d1f84842c952c84ced4bef42efdd7` on permanent branch `work`

## Objective — bounded documentation only
Write a **procedure placeholder**, not an executed test, defining the prerequisites, reproducible collection steps, capture artifacts, pass/fail criteria and evidence provenance for independent third-party DNP3 v0 interoperability validation (VEC-01 / residual R4). Do not close R4 or label third-party interoperability proven. This task must not implement a test harness, add a working integration, send controls, or change code.

## Startup and authority
Read `AGENTS.md`, `operation cwal.md`, `BLACK_SHEEP_WALL.md`, and only the matching top markers in `workflow/active_work/README.md` and root `handoff.md`. Run `bash scripts/cwal-check.sh`; verify CURRENT=MEXT-042, designated branch `work`, clean and remote-synchronized checkout, and correct task body. No other task is promoted, and `QUEUED: NONE`.

## Context — use minimum needed (BLACK SHEEP WALL)
- `active_work/MEXT_MASTER_ROADMAP.md`, section for MEXT-042 and MEXT-035 post-gate definition
- `active_work/supported-profile.md`, authoritative v0 allowed/unsupported boundary
- `active_work/external-acceptance.md`, completed in-repo acceptance vs missing third-party R4
- `active_work/testdata/external/FORMAT.md` and `active_work/testdata/external/sample-vec01-placeholder.vec`, existing VEC-01 fixture conventions (READ ONLY)
- `scripts/verify-external-mvp.sh` and `scripts/README.md`, existing advisory vs required evidence boundary (READ ONLY)
- `docs/verification/DNP3-059.md` only for last known prior checkpoint, not as proof of independent interoperability.

## Exact deliverable and allowlist
- Create `active_work/third-party-capture-procedure.md` (human-readable planning document ONLY). It must explain: independent third-party peer selection and provenance; isolated non-production test topology; IP/port/address configuration; v0-only operation matrix (TCP connect/link handshake, Class-0 G1V1/G30V1/G20V1 reads including multi-header, G12V1 CROB Direct Operate with explicitly authorized simulated outputs); frame/PCAP collection and clock correlation; original capture + redacted reproducible sidecar format; CRC/LSB-first, SEQ/CON and application status inspection; deterministic negative cases (bad CRC, mismatched address, dropped connection/timeout, command reject or unknown control outcome); bounded timing and safe stop; reviewer-independent acceptance/rejection mapping; artifact hashes, stack implementation/version, software configuration, tool versions, timestamps and exact executed SHA; and limitations.
- On completion, update only `workflow/active_work/README.md` and root `handoff.md` to archive/close CURRENT, add `docs/verification/MEXT-042.md`, and archive the verbatim active body at `archive/active-work/MEXT-042.md`; delete active `workflow/active_work/MEXT-042.md` after archival.
- During preparation the only authorized changed paths are this CURRENT task file and those two marker files.
- **No changes** to Go source/tests, .kde governance, code architecture, `go.mod`, `go.sum`, Go workspace, supported profile, existing capture fixtures, scripts, CI, README marketing claims, roadmap status, release files or any historical task/evidence.

## Acceptance and negative checks
1. Explicit planning-only status: **VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY INTEROP NOT PROVEN**. A procedural document alone cannot satisfy external acceptance.
2. Partner must be truly independently implemented (a second copy of this DNP3 library / its own outstation does NOT count). Include version/license/provenance evidence.
3. Scope restricted to approved v0 TCP and static groups, Direct Operate only with non-production simulated outputs and prior separate permission. No live field control, production peer, serial/TLS/SBO, unsolicited events or unsupported groups.
4. Clearly distinguish evidence collection from acceptance. Define command/status checks (including ambiguous outcomes and never blind retry) and evidence integrity checks; do not manufacture PCAP or test results.
5. Include failure/timeout/capture-incomplete disposition and bounded nonproduction safe-stop criteria. No unsupported feature can silently become supported.
6. No claims of certified IEEE 1815 conformance, production readiness or universal DNP3 compatibility.
7. Match all six CWAL authority fields; only allowlisted Markdown files may change; maintain a clean, reproducible remote checkpoint and accurate verification record.

## Mandatory executable verification — REQUIRED, never waive as docs-only
JR must run the following commands on the exact verified `work` checkout, record **exit status, Go version, execution SHA and meaningful output**:
```sh
bash scripts/cwal-check.sh
test -z "$(gofmt -l .)"
git diff --check
go version
go vet ./...
go build ./...
go test ./... -count=1
go test -race ./internal/outstation/... -count=1
./scripts/verify-mvp.sh
./scripts/verify-external-mvp.sh
```
If Go is missing, try installing official compatible Go system-wide, then privately in JR's user-owned directory with PATH, then isolated Go runner/container with the exact checkout. Verify official archive hash if downloading. **NOT_RUN is unacceptable for mandatory checks.** Running the existing external MVP script is only in-repo Tiers 1–3 and is NEVER VEC-01 third-party proof.

## Closeout and stop conditions
- If and only if the deliverable, scope, mandatory Go checks, CWAL checks and evidence are all PASS: archive CURRENT verbatim, write `docs/verification/MEXT-042.md`, set `CURRENT: NONE`, `QUEUED: NONE`, `Status: WORK_PACKAGE_READY` in both marker files, commit and push to `work`, fetch/read back exact remote SHA and markers/archive/evidence, report `TASK_CHECKPOINTED`.
- If any mandatory test cannot be executed or fails, or further changes/tests demand new authority: keep `CURRENT: MEXT-042`, set `Status: TASK_BLOCKED`, record exact reason and STOP. No implementation code fix under this task.
- No PR creation, merge, main push, force push, deployment, production device traffic, third-party lab execution or field controls are authorized. Any actual VEC-01 test campaign needs separately owner-approved task and safety gates.
