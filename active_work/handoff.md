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
- **Last completed task:** MEXT-022 — Real-TCP full MVP path test
- **Last checkpoint commit:** `d1ef2a0` (MEXT-022 — real-TCP full MVP master path) — pushed to origin/main
- **Current task:** none (idle) — next READY is MEXT-023
- **Test status:** Internal `./scripts/verify-mvp.sh` exit 0 (green after MEXT-022). External gate `scripts/verify-external-mvp.sh` Tier 1 (internal real-TCP, now incl. TestRealTCPFullMVPPath) green; Tier 2 (VEC-01) fail-closed by design until MEXT-033.
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
- **MEXT-015** — IntegrityPoll single multi-header path. Removed the per-group workaround as the primary path in `pkg/dnp3/master/client.go` `runIntegrityPoll`: the primary is now ONE Class-0 multi-group Read (G1/G20/G30 all-objects headers in a single APDU); the MEXT-014 qualifier-aware parsers populate every group from that single exchange. Retained a documented per-group fallback (`integrityPollPerGroup`) used only when the primary exchange errors (peer that rejects a multi-group Class-0 read or transport failure). Updated doc comments on `IntegrityPoll` (interface + impl). Added `pkg/dnp3/master/integrity_poll_test.go`: TestIntegrityPollSingleMultiHeaderExchange (asserts exactly ONE application request frame is sent for the poll and the full MVP set G1+G20+G30 is returned from that one exchange), TestIntegrityPollFallbackPerGroup (multi-group primary errors via seq-mismatch transport → per-group fallback returns the full set). Updated supported-profile.md (R3 row marked resolved; verification table row for the multi-header exchange + fallback) and external-acceptance.md. go test ./... + verify-mvp.sh green. **R3 resolved.**
- **MEXT-013** — Operate real-TCP vs in-repo outstation. Proved the MEXT-012 R1 fix on a real TCP master↔outstation loopback (not the in-memory simulator). The in-repo outstation's `handleDirectOperate` returns an IIN-only response (outstation IIN bytes + no G12V1 control-status echo); before MEXT-012 this left the master with no parseable status → ControlTimeout (the DNP3-091 discovery). With MEXT-012's `resolveOperateStatus`, an IIN-only response with clear IIN is CommandStatusSuccess. Added `test/integration/operate_real_tcp_test.go`: TestOperateRealTCPSuccess (real-TCP Connect → DirectOperate CROB → ControlSuccess, not ControlTimeout; outstation-side dispatch symmetry asserted), TestOperateRealTCPBlockedStatus (rejected command → IIN.ParameterError → classified failure, never ControlSuccess/ControlTimeout). Updated outstation_side_gate_test.go discovery note (no longer claims ControlTimeout). Updated supported-profile.md (R1 row marked resolved; verification table row) and external-acceptance.md (R1 checklist item checked). go test ./... + verify-mvp.sh green. **R1 resolved (real-TCP proven). Observed response shape: IIN-only (no G12V1 echo) on success; IIN.ParameterError on rejection.**
- **MEXT-017** — Link handshake external frame vectors. Locked the IEEE 1815 wire shape of the master's link-layer handshake request frames against external-style golden byte vectors. Added golden fixtures `active_work/testdata/link-reset-link-stations.hex` (control 0xC0, master 0x0003 → outstation 0x0004), `link-request-link-status.hex` (0xC9), `link-secondary-ack.hex` (0x00), `link-secondary-link-status.hex` (0x02), `link-secondary-nack.hex` (0x01). Added `internal/master/link_handshake_vectors_test.go`: TestLinkHandshakeRequestVectors (the master's first two emitted frames during Connect equal the golden request bytes; decode-level field assertions on control byte/func/addrs), TestLinkHandshakeRequiresBothExchanges (valid ACK + valid Link Status → Connect succeeds, state Active), TestLinkHandshakeNACKFailsConnect (NACK on reset → Connect fails, state Error; NACKRetries=0 so terminal on first NACK), TestLinkHandshakeWrongFuncOnLinkStatusFailsConnect (valid ACK then ACK-instead-of-Link-Status → Connect fails), TestLinkHandshakeMissingLinkStatusFailsConnect (transport closes after ACK, no second exchange → Connect fails), TestLinkHandshakeGoldenResponseDecode (secondary golden response fixtures decode to IEEE 1815 fields independent of the master encoder). Added a `scriptedTransport` test helper (queued responses + sent-frame recording). **Connect requires both exchanges (ACK then Link Status); any mismatch fails Connect → StateError.** go test ./... + verify-mvp.sh green.
- **MEXT-018** — Application SEQ + CON on solicited path audit. Audited the master's solicited-path application sequence/confirm handling against IEEE 1815 for the v0 path. The unit-level invariants (SEQ stream 0-15 wrap; advance only on successful send; no advance on transport Send failure; processResponse SEQ match/mismatch → ErrResponseSeqMismatch; waitForConfirmation match/mismatch/timeout; CON=1 response triggers an application confirm) were already covered across master_test.go / app_confirm_test.go / confirm_timeout_test.go / fcb_test.go. Filled the remaining end-to-end gaps in `internal/master/seq_con_audit_test.go`: TestSolicitedCONConfirmAndResponseMatchingSeq (CON=1 request → dedicated confirm with matching SEQ → response with matching SEQ → success; SEQ advances exactly once), TestSolicitedCONConfirmWrongSeqFails (CON=1 with a wrong-SEQ dedicated confirm → ErrConfirmSeqMismatch, terminal), TestSolicitedResponseSeqMismatchFailsEndToEnd (response SEQ mismatch → ErrResponseSeqMismatch, no data surfaced), TestSolicitedRetryReusesOrAdvancesSeq (characterization: a retry after a response-SEQ mismatch carries an INCREMENTED SEQ because the master advances at send time). Added `scriptedSeqTransport`, `retryEchoSeqTransport`, and `terminalRetryPolicy` test helpers. **Discovery (ticketed, not fixed — behavior change beyond audit scope):** `sendWithRetry`/`sendWithRetryAndGetResponse` advance the application SEQ at send time (after `transport.Send` succeeds, before the response/confirm is validated), so a retry of the same logical request after a response failure uses the NEXT SEQ. The doc comment ("retries reuse the same value; it advances only on a successful send") describes send-success, not full-transaction success. IEEE 1815 is ambiguous on retry SEQ; the current behavior is internally consistent for the success path. Changing to advance only on full-transaction success is a behavior change deferred to architect decision — follow-up tracked here. go test ./... + verify-mvp.sh green.
- **MEXT-019** — IIN table freeze vs 1815. Froze the IIN bit map in `internal/al/application.go` against IEEE 1815-2012 for the external v0 interop claim. The IIN table (flag → octet/position) was already correct and documented; added an explicit `MEXT-019 FREEZE` note to the `IIN` doc comment stating the mapping MUST NOT change without a spec-continuity review. Added `internal/al/iin_freeze_test.go`: TestIINKnownMasksFreeze (named critical masks pinned to their verified [IIN1,IIN2] bytes — all-clear, NeedTime, DeviceRestart, DeviceTrouble, AllStations+DeviceRestart, Class1+2+3 events, FuncUnknown, ObjectUnknown, ParameterError, all-IIN2-errors 0xFC, command-rejected), TestIINRoundTripAllMasks (every 16-bit mask round-trips losslessly through Bytes/SetIIN — freezes the entire table against drift), TestIINDecodeIINAllMasks (public DecodeIIN/EncodeIIN inverse for all masks), TestIINReservedBitsRoundTrip (characterizes the two reserved IIN2 bits: the encoder/decoder preserve wire bytes rather than masking reserved bits; locks current behavior so a future force-to-0 is a deliberate decision). **Table location:** `internal/al/application.go` `IIN` struct + doc comment (lines ~187-236). Existing IIN tests (TestIINBytes/SetIIN/BitPositions/DecodeIINTooShort) remain green. go test ./... + verify-mvp.sh green.
- **MEXT-020** — VEC-01 capture fixture format. Defined the external capture fixture format for the R4 (VEC-01 / independent PCAP / third-party stack) proof. Created `active_work/testdata/external/` with: `FORMAT.md` (the `.vec` line-oriented text format — metadata `key: value` header + ordered `@ <direction> <layer>` frame records with hex bytes, optional `.pcap`/`.pcapng` sidecar, comment lines, loader rules when one is added in MEXT-022+), `README.md` (purpose, R4 risk, status — placeholder until MEXT-022/033), and `sample-vec01-placeholder.vec` (a placeholder fixture: link-handshake bytes matching the MEXT-017 goldens for master 0x0003 → outstation 0x0004, then an illustrative Class-0 G60V1 request and G1/G20/G30 response). No code/loader added this task (format documented + directory exists, per scope). go test ./... + verify-mvp.sh green. **Format spec location:** `active_work/testdata/external/FORMAT.md`.
- **MEXT-021** — verify-external-mvp.sh skeleton. Added the external-MVP gate script `scripts/verify-external-mvp.sh` (executable), a two-tier gate for the R4/VEC-01 external interop claim. **Tier 1 (internal real-TCP):** `go build ./...` + the real-TCP loopback tests (real TCP transport + real DNP3 wire framing vs the in-repo outstation) — `TestTCPMasterOutstationRead`/`TestTCPDirectCommunication`/`TestMasterOutstationEndToEndComprehensive`, `TestOperateRealTCPSuccess`/`TestOperateRealTCPBlockedStatus`, `TestPublicMVPLoopbackFullLifecycle`/`...OperateStatus`/`...ErrorClassification`. Tier 1 passes today. **Tier 2 (external/third-party VEC-01): fail-closed.** Looks for a genuine (non-placeholder) `.vec` capture fixture under `active_work/testdata/external/` and/or an external interop test; until such proof lands (MEXT-022 real-TCP full MVP path, MEXT-033 third-party stack capture) it refuses to pass (exit 1) so the external interop claim cannot be made prematurely. `ALLOW_NO_EXTERNAL=1` runs Tier 1 only (does NOT satisfy the external claim). Wired with TODO markers for MEXT-022/MEXT-033 to add the real external test command. Documented the script in `scripts/README.md`. **Verified:** Tier-1-only path exit 0; fail-closed path exit 1 with a clear message. verify-mvp.sh still green.
- **MEXT-022** — Real-TCP full MVP path test. Added `test/integration/real_tcp_full_mvp_test.go`: `TestRealTCPFullMVPPath` exercises the complete v0 public API over a REAL TCP master↔outstation loopback with NO simulator transport — Connect → IntegrityPoll (assert all MVP points: 2 binary [true,false], 2 analog [42,-7], 2 counters [100,7]) → Operate (DirectOperate CROB Latch On index 0 → ControlSuccess + outstation-side dispatch symmetry) → Disconnect → terminal StateDisconnected. Uses the shared `recordingDataHandler`/`recordingCommandHandler`. Wired into `scripts/verify-external-mvp.sh` Tier 1. **Discovery + fix (outstation, real-path gap MEXT-022 surfaced):** the master's MEXT-015 multi-group integrity read requests the default variation per group (G1V0/G20V0/G30V0), but the real outstation's static builders only special-cased `variation==1` and emitted a malformed response for V0 (the in-memory simulator handled V0, masking the bug). Per IEEE 1815 "variation 0 = default variation," added `normalizeStaticVariation` in `internal/outstation/outstation.go` (dispatch in `buildReadResponse`) to serve V0 as V1 for the static groups (1/10/20/30/40); event groups (2/21/31) are not normalized (their empty-buffer stub behavior is unchanged). Locked the fix with `internal/outstation/static_variation_test.go`: `TestReadStaticVariationZeroServedAsDefault` (V0 response object data == V1 for G1/G20/G30) and `TestNormalizeStaticVariation` (table test for the helper, incl. event groups passthrough). go test ./... + verify-mvp.sh + verify-external-mvp.sh Tier 1 green.

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
- [x] MEXT-015 — IntegrityPoll single multi-header path
- [x] MEXT-013 — Operate real-TCP vs in-repo outstation
- [x] MEXT-017 — Link handshake external frame vectors
- [x] MEXT-018 — Application SEQ + CON on solicited path audit
- [x] MEXT-019 — IIN table freeze vs 1815
- [x] MEXT-020 — VEC-01 capture fixture format
- [x] MEXT-021 — verify-external-mvp.sh skeleton

## Next READY Tasks

- **MEXT-022** — Real-TCP full MVP path test (prereqs MEXT-013, MEXT-015 — done)

## Recommended Next Task

**MEXT-022 — Real-TCP full MVP path test**. Add an end-to-end real-TCP test (Connect → IntegrityPoll → Operate → Close) over TCP with no simulator transport, asserting points + operate policy. Must be green locally and referenced by `scripts/verify-external-mvp.sh` Tier 1. verify-mvp.sh must stay green.

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
- **R1:** Real outstation Direct-Operate may omit G12 status echo → ControlTimeout (MEXT-012/013). **RESOLVED in MEXT-012/013** — IIN-only clear response now success; proven on real TCP (DirectOperate against in-repo outstation returns ControlSuccess).
- **R2:** CROB control-code values may not match IEEE 1815 bitfield (MEXT-010/011). **RESOLVED in MEXT-011** — constants now 1815 bitfield (0x01/0x02/0x04/0x08/0x10/0x80).
- **R3:** Multi-object-header Class-0 parse can lose points (MEXT-014/015). **RESOLVED in MEXT-014/015** — `skipGroupData` is qualifier-aware; `IntegrityPoll` now uses a single Class-0 multi-group read as primary path (one exchange returns the full set) with a per-group fallback for peers that reject the multi-group exchange.
- **R4:** No VEC-01 external capture proof yet.

## Blockers / Risks

- None for MEXT-002–004 (docs/baseline).
- Do not implement SBO/unsolicited/full Level-3 in this series.
- Do not claim external interop in README until MEXT-035.

## Next Action

1. Read `active_work/MEXT_MASTER_ROADMAP.md` (MEXT-022).
2. Implement **MEXT-022** (Real-TCP full MVP path test).
3. Run go test ./... + verify-mvp.sh + verify-external-mvp.sh Tier 1; commit, push.

## MVP Gate

```
TOTAL TASKS: 40
EXTERNAL MVP COMPLETE AT: MEXT-035
NEXT TASK: MEXT-022 — Real-TCP full MVP path test
```
