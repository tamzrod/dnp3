# DNP3 Implementation Handoff (MEXT Series)

**Last updated:** 2026-08-12  
**Roadmap:** `active_work/MEXT_MASTER_ROADMAP.md`  
**Profile:** `active_work/supported-profile.md`  
**Acceptance:** `active_work/external-acceptance.md`  
**Archived series:** `active_work/archive/`

## Status

- **Series:** MEXT (Master External Use)
- Planning complete for MEXT.
- **Internal MVP:** COMPLETE at DNP3-056 (archived). Do not reopen v1 task IDs.
- **External MVP:** NOT COMPLETE. Target close at **MEXT-035**.
- **Last completed task:** MEXT-010 — CROB control-code IEEE 1815 bitfield audit
- **Last checkpoint commit:** `73e02cd` (MEXT-001..003 checkpoint) — pushed to origin/main
- **Current task:** none (idle) — next READY is MEXT-011
- **Test status:** Internal `./scripts/verify-mvp.sh` must remain exit 0. External gate after MEXT-021/033.
- **Internal MVP baseline sha:** `53b40fb` (`53b40fb2f8df3ef6a682f091c6664c9aef64bde2`) — `./scripts/verify-mvp.sh` exit 0 pinned here before external changes (MEXT-003).

## Completed Tasks

- **MEXT-001** — Archive v1 series + open MEXT handoff. Archive pointers under `active_work/archive/`; live `MEXT_MASTER_ROADMAP.md`, `handoff.md`, `external-acceptance.md`, v1 roadmap path is stub. Full v1 bodies: `git show c4ce51c:active_work/DNP3_MASTER_ROADMAP.md` and `.../handoff.md`.
- **MEXT-002** — Record residuals R1–R5 in supported-profile. Added "External Residuals" section to `active_work/supported-profile.md` with R1–R5 table (residual, impact, resolving MEXT task). Docs-only; no code change; no tests.
- **MEXT-003** — Baseline commit hash + verify-mvp lock. Re-ran `./scripts/verify-mvp.sh` → exit 0 on HEAD `53b40fb` (`53b40fb2f8df3ef6a682f091c6664c9aef64bde2`). Pinned the green baseline sha in handoff before external changes.
- **MEXT-004** — External acceptance criteria checklist file. Confirmed `active_work/external-acceptance.md` already present and matches roadmap §4 (six gate items: verify-mvp, verify-external-mvp, CROB 1815 goldens, Operate no-timeout, multi-header Class-0, README claims). No edits needed.
- **MEXT-005** — README external-claim lock. Added explicit "External interop status (MEXT series lock)" block to `README.md` Current Status: external interop NOT claimed; internal use only; claim blocked until MEXT-035; points at MEXT roadmap, external-acceptance, supported-profile residuals. Resolves R5 over-claim risk. Docs-only.
- **MEXT-010** — CROB control-code IEEE 1815 bitfield audit. Wrote `active_work/crob-code-audit.md`: confirmed R2. CROB constants in `internal/master/master.go:2310-2319` are a 1..8 enum, not the 1815 bitfield (0x01/0x02/0x04/0x08/0x10/0x80). Only NUL matches by coincidence. Wire LAYOUT is correct; only constant VALUES + bool mapping diverge. Outstation decode switch (`outstation.go:410-432`) uses same off-spec enum. Listed affected goldens in `control_vector_test.go` for MEXT-011. Read-only; no code change.

## Current Checkpoint Batch

- [x] MEXT-001 — Archive v1 series + open MEXT handoff
- [x] MEXT-002 — Record residuals R1–R5 in supported-profile
- [x] MEXT-003 — Baseline commit hash + verify-mvp lock
- [x] MEXT-004 — External acceptance criteria checklist file
- [x] MEXT-005 — README external-claim lock
- [x] MEXT-010 — CROB control-code IEEE 1815 bitfield audit

## Next READY Tasks

- **MEXT-011** — CROB control-code correction + golden vectors (prereq MEXT-010, done)
- MEXT-012 — Direct-Operate response: status object optional path (prereq MEXT-003, done)
- MEXT-014 — Multi-header Class-0 parse fix (prereq MEXT-003, done)
- MEXT-016 — IIN bit map freeze for external v0 (prereq MEXT-003, done)

## Recommended Next Task

**MEXT-011 — CROB control-code correction + golden vectors**. Realign CROB constants in `internal/master/master.go` to the IEEE 1815 bitfield (0x01/0x02/0x04/0x08/0x10/0x80); update outstation decode switch + goldens in `control_vector_test.go`; keep public API stable via bool path. Fix R2. verify-mvp.sh must stay green.

## Test Commands (baseline)

```bash
export PATH=$HOME/go-install/go/bin:$PATH
go test ./...
go test -race ./internal/master/... ./pkg/dnp3/... ./test/integration/...
./scripts/verify-mvp.sh
# ./scripts/verify-external-mvp.sh   # after MEXT-021/033
```

## Code State (this batch)

- `active_work/archive/*`: v1 archive pointers + README
- `active_work/MEXT_MASTER_ROADMAP.md`: full MEXT roadmap (40 tasks)
- `active_work/DNP3_MASTER_ROADMAP.md`: archived stub pointing to MEXT
- `active_work/external-acceptance.md`: external gate checklist
- `active_work/HANDOFF_TEMPLATE.md`: MEXT template

## Implementation Discoveries (carry forward from v1)

- DNP3 multi-octet wire fields are **LSB-first**.
- `frame.EncodedSize` = header + data + 2*ceil(dataLen/16) CRC bytes.
- **R1:** Real outstation Direct-Operate may omit G12 status echo → ControlTimeout (MEXT-012/013).
- **R2:** CROB control-code values may not match IEEE 1815 bitfield (MEXT-010/011).
- **R3:** Multi-object-header Class-0 parse can lose points (MEXT-014/015).
- **R4:** No VEC-01 external capture proof yet.

## Blockers / Risks

- None for MEXT-002–004 (docs/baseline).
- Do not implement SBO/unsolicited/full Level-3 in this series.
- Do not claim external interop in README until MEXT-035.

## Next Action

1. Read `active_work/crob-code-audit.md` + `active_work/MEXT_MASTER_ROADMAP.md` (MEXT-011).
2. Implement **MEXT-011** (CROB control-code correction + golden vectors).
3. Checkpoint after MEXT-011/012/014 (3 tasks) — run go test ./... + verify-mvp.sh, commit, push.

## MVP Gate

```
TOTAL TASKS: 40
EXTERNAL MVP COMPLETE AT: MEXT-035
NEXT TASK: MEXT-011 — CROB control-code correction + golden vectors
```
