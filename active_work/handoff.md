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
- **Last completed task:** MEXT-003 — Baseline commit hash + verify-mvp lock
- **Last checkpoint commit:** e9c65cca (archive) / dd071f18 (MEXT roadmap) — pushed to origin/main
- **Current task:** none (idle) — next READY is MEXT-004
- **Test status:** Internal `./scripts/verify-mvp.sh` must remain exit 0. External gate after MEXT-021/033.
- **Internal MVP baseline sha:** `53b40fb` (`53b40fb2f8df3ef6a682f091c6664c9aef64bde2`) — `./scripts/verify-mvp.sh` exit 0 pinned here before external changes (MEXT-003).

## Completed Tasks

- **MEXT-001** — Archive v1 series + open MEXT handoff. Archive pointers under `active_work/archive/`; live `MEXT_MASTER_ROADMAP.md`, `handoff.md`, `external-acceptance.md`, v1 roadmap path is stub. Full v1 bodies: `git show c4ce51c:active_work/DNP3_MASTER_ROADMAP.md` and `.../handoff.md`.
- **MEXT-002** — Record residuals R1–R5 in supported-profile. Added "External Residuals" section to `active_work/supported-profile.md` with R1–R5 table (residual, impact, resolving MEXT task). Docs-only; no code change; no tests.
- **MEXT-003** — Baseline commit hash + verify-mvp lock. Re-ran `./scripts/verify-mvp.sh` → exit 0 on HEAD `53b40fb` (`53b40fb2f8df3ef6a682f091c6664c9aef64bde2`). Pinned the green baseline sha in handoff before external changes.

## Current Checkpoint Batch

- [x] MEXT-001 — Archive v1 series + open MEXT handoff
- [x] MEXT-002 — Record residuals R1–R5 in supported-profile
- [x] MEXT-003 — Baseline commit hash + verify-mvp lock

## Next READY Tasks

- **MEXT-004** — External acceptance checklist (file already present; confirm/align)
- MEXT-005 — README external-claim lock (prereq MEXT-002, done)
- MEXT-010 — CROB control-code IEEE 1815 bitfield audit (prereq MEXT-003, now done)

## Recommended Next Task

**MEXT-004 — External acceptance criteria checklist file**. Confirm `active_work/external-acceptance.md` matches roadmap §4 (six gate items). This completes the 3-task checkpoint batch (001→003); commit and push after.

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

1. Read `active_work/MEXT_MASTER_ROADMAP.md` → MEXT-004.
2. Implement **MEXT-004** (confirm external-acceptance.md matches roadmap §4).
3. Checkpoint: 3 tasks (001→003) done → commit → push main.
4. Then proceed to MEXT-005 / MEXT-010.

## MVP Gate

```
TOTAL TASKS: 40
EXTERNAL MVP COMPLETE AT: MEXT-035
NEXT TASK: MEXT-004 — External acceptance criteria checklist file
```
