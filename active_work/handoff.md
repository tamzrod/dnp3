# DNP3 Implementation Handoff (MEXT Series)

**Last updated:** 2026-08-12  
**Roadmap:** `active_work/MEXT_MASTER_ROADMAP.md`  
**Profile:** `active_work/supported-profile.md`  
**Acceptance:** `active_work/external-acceptance.md`  
**Archived series:** `active_work/archive/DNP3_MASTER_ROADMAP_v1_internal_mvp.md`

## Status

- **Series:** MEXT (Master External Use)
- Planning complete for MEXT.
- **Internal MVP:** COMPLETE at DNP3-056 (archived series). Do not reopen v1 task IDs for new work.
- **External MVP:** NOT COMPLETE. Target close at **MEXT-035**.
- Last completed task: *(none in MEXT yet — series opened)*
- Last checkpoint commit: *(pending archive body commit)*
- Current task: **MEXT-001 — Archive v1 series + open MEXT handoff** (docs)
- Test status: Internal `./scripts/verify-mvp.sh` must remain exit 0. External gate after MEXT-021/033.

## Completed Tasks

*(MEXT series — none yet)*

### Archived series note

V1 tasks DNP3-001 … ~DNP3-098 delivered **internal** Master MVP. Full history: `active_work/archive/handoff_v1_internal_mvp.md`.

## Current Checkpoint Batch

- [ ] MEXT-001 — Archive v1 series + open MEXT handoff
- [ ] MEXT-002 — Record residuals R1–R5 in supported-profile
- [ ] MEXT-003 — Baseline commit hash + verify-mvp lock

## Next READY Tasks

- **MEXT-001** — Archive v1 series + open MEXT handoff (no prereq)
- MEXT-002 — Record residuals R1–R5 (prereq MEXT-001)
- MEXT-003 — Baseline sha + verify-mvp lock (prereq MEXT-001)
- MEXT-004 — External acceptance checklist file (prereq MEXT-001; file may already exist)

## Recommended Next Task

**MEXT-001 — Archive v1 series + open MEXT handoff**. Confirm archive file bodies under `active_work/archive/`, then mark MEXT-001 done and proceed to **MEXT-002**.

## Test Commands (baseline)

```bash
export PATH=$HOME/go-install/go/bin:$PATH
go test ./...
go test -race ./internal/master/... ./pkg/dnp3/... ./test/integration/...
./scripts/verify-mvp.sh
# ./scripts/verify-external-mvp.sh   # after MEXT-021/033
```

## Code State (this batch)

- `active_work/MEXT_MASTER_ROADMAP.md`: new series source of truth
- `active_work/handoff.md`: this file (MEXT)
- `active_work/HANDOFF_TEMPLATE.md`: MEXT template
- `active_work/external-acceptance.md`: external gate checklist
- `active_work/DNP3_MASTER_ROADMAP.md`: archive pointer stub
- `active_work/archive/*`: full v1 roadmap/handoff bodies

## Implementation Discoveries (carry forward from v1)

- DNP3 multi-octet wire fields are **LSB-first**.
- `frame.EncodedSize` = header + data + 2*ceil(dataLen/16) CRC bytes.
- **R1:** Real outstation Direct-Operate may omit G12 status echo → ControlTimeout (MEXT-012/013).
- **R2:** CROB control-code values may not match IEEE 1815 bitfield (MEXT-010/011).
- **R3:** Multi-object-header Class-0 parse can lose points (MEXT-014/015).
- **R4:** No VEC-01 external capture proof yet.

## Blockers / Risks

- None for MEXT-001–004 (docs/baseline).
- Do not implement SBO/unsolicited/full Level-3 in this series.
- Do not claim external interop in README until MEXT-035.

## Next Action

1. Ensure `active_work/archive/` holds full v1 bodies.
2. Read `active_work/MEXT_MASTER_ROADMAP.md`.
3. Complete MEXT-001 if archive bodies missing; else MEXT-002.
4. Every 3 tasks: verify → update handoff → commit → push main.

## MVP Gate

```
TOTAL TASKS: 40
EXTERNAL MVP COMPLETE AT: MEXT-035
NEXT TASK: MEXT-001 — Archive v1 series + open MEXT handoff
```
