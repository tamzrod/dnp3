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
- **Last completed task:** MEXT-016 — Qualifier allow-list vs 1815 for v0 path
- **Last checkpoint commit:** `5ffff47` (MEXT-012/014/016 checkpoint) — pushed to origin/main
- **Current task:** none (idle) — next READY is MEXT-015
- **Test status:** Internal `./scripts/verify-mvp.sh` must remain exit 0. External gate after MEXT-021/033.
- **Internal MVP baseline sha:** `53b40fb` (`53b40fb2f8df3ef6a682f091c6664c9aef64bde2`) — `./scripts/verify-mvp.sh` exit 0 pinned here before external changes (MEXT-003).

## Completed Tasks

- **MEXT-001** — Archive v1 series + open MEXT handoff. Archive pointers under `active_work/archive/`; live `MEXT_MASTER_ROADMAP.md`, `handoff.md`, `external-acceptance.md`, v1 roadmap path is stub. Full v1 bodies: `git show c4ce51c:active_work/DNP3_MASTER_ROADMAP.md` and `.../handoff.md`.
- **MEXT-002** — Record residuals R1–R5 in supported-profile. Added "External Residuals" section to `active_work/supported-profile.md` with R1–R5 table (residual, impact, resolving MEXT task). Docs-only; no code change; no tests.
- **MEXT-003** — Baseline commit hash + verify-mvp lock. Re-ran `./scripts/verify-mvp.sh` → exit 0 on HEAD `53b40fb` (`53b40fb2f8df3ef6a682f091c6664c9aef64bde2`). Pinned the green baseline sha in handoff before external changes.
- **MEXT-004** — External acceptance criteria checklist file. Confirmed `active_work/external-acceptance.md` already present and matches roadmap §4 (six gate items: verify-mvp, verify-external-mvp, CROB 1815 goldens, Operate no-timeout, multi-header Class-0, README claims). No edits needed.
- **MEXT-005** — README external-claim lock. Added explicit "External interop status (MEXT series lock)" block to `README.md` Current Status: external interop NOT claimed; internal use only; claim blocked until MEXT-035; points at MEXT roadmap, external-acceptance, supported-profile residuals. Resolves R5 over-claim risk. Docs-only.
- **MEXT-010** — CROB control-code IEEE 1815 bitfield audit. Wrote `active_work/crob-code-audit.md`: confirmed R2. CROB constants in `internal/master/master.go:2310-2319` are a 1..8 enum, not the 1815 bitfield (0x01/0x02/0x04/0x08/0x10/0x80). Only NUL matches by coincidence. Wire LAYOUT is correct; only constant VALUES + bool mapping diverge. Outstation decode switch (`outstation.go:410-432`) uses same off-spec enum. Listed affected goldens in `control_vector_test.go` for MEXT-011. Read-only; no code change.
- **MEXT-011** — CROB control-code correction + golden vectors. Realigned CROB constants in `internal/master/master.go` to the IEEE 1815 bitfield (NUL=0x01, PulseOn=0x02, PulseOff=0x04, LatchOn=0x08, LatchOff=0x10, Queue=0x80; removed non-spec Close/Open/Trip). Updated outstation decode switch (`internal/outstation/outstation.go` WriteBinaryOutput) and public outstation bool mapping (`pkg/dnp3/outstation/server.go`) to 1815 values. Updated goldens in `internal/master/control_vector_test.go` (code byte 0x07->0x08) and `internal/testutils/functional_test.go` (raw Code:2 -> CROBCodeLatchOn). Updated supported-profile.md (CROB control-code wire-field row + R2 marked resolved). go test ./... + verify-mvp.sh green. **R2 resolved.**
- **MEXT-012** — Direct-Operate response: status object optional path. Fixed R1 parse side in `internal/master/master.go` `OperateWithStatus`: replaced direct `parseCommandStatus` call with `resolveOperateStatus(resp)`. New rules: (a) G12V1 status=0 -> success; (b) no G12V1 object (IIN-only) + clear IIN -> success; (c) IIN-only + error IIN -> failure (never success; mapped FuncUnknown/ObjectUnknown->NotSupported, ParameterError/BadConfig->BadFormat, LocalControl->Local, AlreadyExecuting->AlreadyActive); (d) truncated G12V1 (header present, status byte missing) -> Unknown (never success). Added `containsG12V1Header` and `commandStatusFromIIN` helpers. Updated tests: `command_status_test.go` (renamed MissingObjectNotSuccess -> IINOnlyClearSuccess; added IINOnlyError + TruncatedNotSuccess), `pkg/dnp3/master/operate_status_test.go` (renamed + added error-IIN variant). Updated supported-profile.md R1 row + external-acceptance.md. go test ./... + verify-mvp.sh green. **R1 parse-side resolved; real-TCP proof pending MEXT-013.**
- **MEXT-014** — Multi-object-header Class-0 parse harden. Fixed R3 parse side in `pkg/dnp3/master/client.go`: rewrote `skipGroupData` to be qualifier-aware (new signature `(offset, data, al.ObjectHeader)`). Added helpers `objectPointCount`, `isSequentialQualifier`, `objectIndexPrefixBytes`, `objectValueBytes`. count8/count16/range16 now skip with NO per-point index prefix; index8 (0x00) uses a 2-octet per-point index (stack convention); G1V1 packed uses ceil(n/8). Previously skipGroupData assumed a 2-octet index for ALL qualifiers, so a single APDU carrying G1+G20+G30 lost G20/G30 (offset overshoot broke the header scan). Updated 5 callers. Added `pkg/dnp3/master/multi_header_test.go` (TestReadMultiHeaderReturnsAllGroups): one APDU with G1(2pts)+G20(1pt)+G30(1pt) returns all points. Updated supported-profile.md R3 row + external-acceptance.md. The per-group `IntegrityPoll` workaround remains as primary until MEXT-015. go test ./... + verify-mvp.sh green. **R3 parse-side resolved; workaround removal pending MEXT-015.**
- **MEXT-016** — Qualifier allow-list vs 1815 for v0 path. Locked the v0 request-side qualifier allow-list {0x06 all-objects (read/integrity), 0x00 index8 (operate/select), 0x28 range16 (ranged read), 0x07 count8 (event-class poll), 0x27 count16}. Added `internal/master/qualifier_golden_test.go`: TestPollRequestQualifierAllowList (every buildPollRequest header qualifier is in the allow-list; integrity poll golden G60V1 0x06), TestControlRequestQualifierAllowList (G12V1 control request qualifier 0x00, count 1), TestRangedReadRequestQualifierAllowList (range16 0x28, start/stop). Reject-others already enforced by `al.ObjectHeader.Encode` (ErrUnsupportedQualifier) and locked by `internal/al/object_header_test.go`. Updated supported-profile.md (request qualifier allow-list row). go test ./... + verify-mvp.sh green. **Goldens committed.**

## Current Checkpoint Batch

- [x] MEXT-001 — Archive v1 series + open MEXT handoff
- [x] MEXT-002 — Record residuals R1–R5 in supported-profile
- [x] MEXT-003 — Baseline commit hash + verify-mvp lock
- [x] MEXT-004 — External acceptance criteria checklist file
- [x] MEXT-005 — README external-claim lock
- [x] MEXT-010 — CROB control-code IEEE 1815 bitfield audit
- [x] MEXT-011 — CROB control-code correction + golden vectors
- [x] MEXT-012 — Direct-Operate response: status object optional path
- [x] MEXT-014 — Multi-object-header Class-0 parse harden
- [x] MEXT-016 — Qualifier allow-list vs 1815 for v0 path

## Next READY Tasks

- **MEXT-015** — IntegrityPoll single multi-header path (prereq MEXT-014, done)
- MEXT-013 — Operate real-TCP vs in-repo outstation (prereq MEXT-012, done) — proves R1 fix on real TCP
- MEXT-017 — Link handshake external frame vectors (prereq MEXT-003, done)

## Recommended Next Task

**MEXT-015 — IntegrityPoll single multi-header path**. Now that MEXT-014 hardened multi-header parse, make `IntegrityPoll` issue a single Class-0 / multi-group read as the primary path (remove the per-group workaround) while keeping a documented fallback. Assert the full MVP set is returned from one exchange. verify-mvp.sh must stay green.

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
- **R1:** Real outstation Direct-Operate may omit G12 status echo → ControlTimeout (MEXT-012/013). **Parse-side RESOLVED in MEXT-012** — IIN-only clear response now success; real-TCP proof pending MEXT-013.
- **R2:** CROB control-code values may not match IEEE 1815 bitfield (MEXT-010/011). **RESOLVED in MEXT-011** — constants now 1815 bitfield (0x01/0x02/0x04/0x08/0x10/0x80).
- **R3:** Multi-object-header Class-0 parse can lose points (MEXT-014/015). **Parse-side RESOLVED in MEXT-014** — skipGroupData qualifier-aware; multi-header G1+G20+G30 one APDU parsed without loss. Workaround removal pending MEXT-015.
- **R4:** No VEC-01 external capture proof yet.

## Blockers / Risks

- None for MEXT-002–004 (docs/baseline).
- Do not implement SBO/unsolicited/full Level-3 in this series.
- Do not claim external interop in README until MEXT-035.

## Next Action

1. Read `active_work/MEXT_MASTER_ROADMAP.md` (MEXT-015).
2. Implement **MEXT-015** (IntegrityPoll single multi-header path).
3. Checkpoint after MEXT-015 (1 task) — run go test ./... + verify-mvp.sh, commit, push.

## MVP Gate

```
TOTAL TASKS: 40
EXTERNAL MVP COMPLETE AT: MEXT-035
NEXT TASK: MEXT-015 — IntegrityPoll single multi-header path
```
