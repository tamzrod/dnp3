# DNP3 Master Completion Roadmap

**Status:** Planning complete — ready for implementation agents  
**Created:** 2026-08-12  
**Primary objective:** Spec-correct, interoperable DNP3 Master (MVP first)  
**Total tasks:** 100  
**Master tasks:** 72  
**Outstation tasks:** 28  
**MVP complete at:** DNP3-056  
**Next task:** DNP3-001 — Residual endian audit

This document is the single source of truth for implementation agents.  
Do **not** implement outside the defined micro-tasks.  
Do **not** invent requirements.  
Update `active_work/handoff.md` after every completed task.

---

## 1. CURRENT DNP3 LIBRARY STATE

Native Go library. Layers present and partially integrated:

- **DLL** (`internal/dll/frame`, `crc`): Frame encode/decode, LSB addresses, single header CRC + 16-octet block CRCs, primary/secondary FCs, golden vectors. Largely correct after remediation.
- **TL** (`internal/tl`): Fragmenter/Reassembler, FIR/FIN/seq (0-63), max 249 data octets. Functional for single-session use.
- **AL** (`internal/al`): APDU header (FIR/FIN/CON/UNS/SEQ), function codes, basic IIN struct + encode/decode, Response with IIN. No formal object-header model; object parse/build lives in master/client parsers.
- **Master** (`internal/master` + `pkg/dnp3/master`): State machine, link reset + ACK consume, AL→TL→DLL send / reverse receive, retries, Class-0 style polls, direct-operate CROB path, public Client API. Sequence/confirm/IIN reaction incomplete. Residual big-endian paths remain in some object parsers/builders.
- **Outstation** (`internal/outstation` + `pkg/dnp3/outstation`): Parallel stack, data/command handlers, Class-0 responses for supported objects. Sufficient for loopback; not full event/unsolicited.
- **Transport**: TCP only (TLS explicitly rejected). Serial absent.
- **SA**: Present, unused by Master MVP path.
- **Tests**: Unit (CRC/frame/TL/AL), golden object vectors, protocol-stack integration, public loopback (Class-0 + direct control). Deterministic simulator partial. Race/cancellation incomplete.
- **Profile lock**: `active_work/supported-profile.md` defines v0 (TCP, Class-0 G1.1/G30.1/G20.1, G12.1 direct operate).

No Ingestor/Kafka/NamelessSCADA code exists in-repo. Library is pure protocol.

---

## 2. SPECIFICATION AUDIT

| Area | Status | Notes |
|------|--------|-------|
| DLL frame structure / length / sync | COMPLETE | Matches golden |
| DLL addresses LSB | COMPLETE | Remediated |
| DLL control / primary-secondary FCs | PARTIAL | Core FCs present; full FCB/FCV/DFC incomplete |
| DLL ACK/NACK/link status/reset | PARTIAL | Reset + basic ACK; full secondary incomplete |
| DLL CRC (header + blocks) | COMPLETE | Golden + boundary vectors |
| TL FIR/FIN/seq / segmentation | COMPLETE | Functional |
| TL reassembly / malformed / isolation | PARTIAL | Concurrent isolation & edge cases weak |
| AL APDU header / FIR/FIN/CON/UNS/SEQ | PARTIAL | Encode/decode present; sequence continuity & CON timeout incomplete |
| AL confirmations | PARTIAL | Logic present; matching & timeout rigor incomplete |
| AL function codes | PARTIAL | Constants present; only Read/DirectOperate exercised |
| AL IIN | PARTIAL | Struct exists; bit semantics & Master reaction incomplete (verify vs IEEE 1815) |
| Object headers / qualifiers / indices | PARTIAL | Ad-hoc; 0x06/0x07 used; full qualifier set incomplete |
| Binary Input G1V1 | PARTIAL | Golden present; residual quality/packing edges |
| Analog Input G30V1 | PARTIAL | Golden present; residual endian in non-0x07 paths |
| Counter G20V1 | PARTIAL | Golden present; residual endian |
| CROB G12V1 | PARTIAL | Direct operate works; status parsing incomplete |
| Other static objects | MISSING / EXPLICITLY UNSUPPORTED for v0 | |
| Event objects / Class 1-3 | EXPLICITLY UNSUPPORTED | |
| Unsolicited | EXPLICITLY UNSUPPORTED | Public API rejects |
| Select-before-operate | EXPLICITLY UNSUPPORTED for v0 | |
| Time objects / sync | MISSING | |
| Master session establishment | PARTIAL | Reset works; full link-status exchange incomplete |
| Master integrity / Class-0 poll | PARTIAL | Works for supported objects |
| Master retries / timeouts | PARTIAL | Present; cancellation & race incomplete |
| Master reconnect / recovery | MISSING | |
| Master IIN-triggered behavior | MISSING | |
| Malformed / invalid CRC / bad qualifier / bad seq | PARTIAL | Some negative tests |
| Concurrency / resource cleanup | PARTIAL | Race detector not fully exercised |
| Public API profile lock | PARTIAL | TLS/unsol rejected; remaining options need tightening |

---

## 3. TARGET MASTER ARCHITECTURE

```
pkg/dnp3/master.Client          // public, context-aware, profile-gated
        │
internal/master.Master          // owns session state, request orchestration,
                                // sequence, retries, recovery
        │
internal/al                     // APDU + IIN + formal object headers
        │
internal/tl                     // fragment / reassemble (session-scoped)
        │
internal/dll/frame + crc        // framing only
        │
pkg/transport.Handler           // TCP only for MVP
```

**State ownership**
- Master: connection state, outstation map, app sequence, pending confirm, reassembler, timeout/retry config.
- TL Fragmenter/Reassembler: per-session, reset on reconnect.
- DLL: stateless encode/decode.
- Public Client: thin wrapper; no protocol state beyond config and handlers.

No Kafka, no SCADA models, no topic names inside the library.

---

## 4. MASTER MVP

### MUST HAVE
- Correct TCP + link reset/handshake with validated ACK
- AL→TL→DLL and reverse path with multi-fragment support
- Class-0 integrity read of G1V1, G30V1, G20V1 with correct qualifiers, LSB fields, quality
- Direct-operate G12V1 CROB with correct status parsing
- Application sequence continuity + confirm handling
- IIN storage + basic reaction (NeedsTimeSync / DeviceRestart once verified)
- Timeouts, retries, context cancellation
- Malformed frame / bad CRC / invalid qualifier rejection
- Clean reconnect after transport failure
- Public API rejects everything outside supported-profile v0

### SHOULD HAVE
- Formal object-header encode/decode helper
- Link-status exchange
- Race-safe concurrent Reads
- Deterministic simulator fully driven by golden vectors

### LATER (explicitly deferred)
- Unsolicited / Class 1-3 events
- Select-before-operate
- Time sync / G50
- Additional variations & groups
- Serial / TLS
- Secure Authentication activation
- Multi-outstation concurrent sessions beyond basic map
- File transfer / attributes

---

## 5. MASTER MICRO-TASK ROADMAP (DNP3-001 … DNP3-072)

### DNP3-001 — Residual endian audit
- **Purpose:** Close remaining multi-octet BE paths in MVP objects.
- **Prerequisites:** none
- **Scope:** `pkg/dnp3/master` parsers + `internal/master` builders for G1/G20/G30/G12
- **Work:** Inventory every multi-octet field; convert residual BE to LSB.
- **Tests:** Existing golden vectors + new BE-negative cases.
- **Acceptance:** No BE remains in v0 object paths.
- **Commit message:** `fix(objects): eliminate residual big-endian in MVP paths`
- **Handoff:** List of fixed functions.
- **Stop condition:** Object vector tests green.

### DNP3-002 — Formal object-header model (encode)
- **Purpose:** Replace ad-hoc header construction.
- **Prerequisites:** none
- **Scope:** `internal/al` new types only
- **Work:** ObjectHeader struct (group, variation, qualifier, range/count); Encode.
- **Tests:** Round-trip encode of 0x06/0x07 headers used by Class-0.
- **Acceptance:** Encode matches golden request headers.
- **Commit message:** `feat(al): object header encode model`
- **Handoff:** API surface.
- **Stop condition:** Unit tests pass; no caller changes yet.

### DNP3-003 — Formal object-header model (decode)
- **Purpose:** Symmetric decode + reject invalid.
- **Prerequisites:** DNP3-002
- **Scope:** `internal/al`
- **Work:** Decode + validation of supported qualifiers only.
- **Tests:** Valid + invalid qualifier/count.
- **Acceptance:** Invalid combinations return error.
- **Commit message:** `feat(al): object header decode and validation`
- **Handoff:** Decoder API.
- **Stop condition:** Tests green.

### DNP3-004 — Wire Master read path to object-header model
- **Purpose:** Eliminate ad-hoc header bytes in requests.
- **Prerequisites:** DNP3-001, DNP3-003
- **Scope:** `internal/master` + public Read builder
- **Work:** Use ObjectHeader for Class-0 requests.
- **Tests:** Public loopback still passes.
- **Acceptance:** Generated request headers match prior golden.
- **Commit message:** `refactor(master): use object-header model for reads`
- **Handoff:** Request path updated.
- **Stop condition:** Loopback green.

### DNP3-005 — Wire Master response parse to object-header model
- **Purpose:** Consistent parse.
- **Prerequisites:** DNP3-003, DNP3-004
- **Scope:** Response parsers in client/master
- **Work:** Decode headers then points.
- **Tests:** Golden Class-0 responses.
- **Acceptance:** Same points returned as before.
- **Commit message:** `refactor(master): object-header based response parse`
- **Handoff:** Parser location.
- **Stop condition:** Vector + loopback green.

### DNP3-006 — Link handshake ACK validation
- **Purpose:** Spec-correct session start.
- **Prerequisites:** none
- **Scope:** `internal/master` performLinkHandshake / sendResetLink
- **Work:** Validate secondary ACK (FC, addresses, CRC); reject others.
- **Tests:** Good ACK / bad FC / wrong address.
- **Acceptance:** Handshake fails on invalid ACK.
- **Commit message:** `fix(master): validate link-layer ACK`
- **Handoff:** Handshake state.
- **Stop condition:** Handshake unit tests green.

### DNP3-007 — Link-status request after reset
- **Purpose:** Complete basic link establishment.
- **Prerequisites:** DNP3-006
- **Scope:** Master connect path
- **Work:** Issue Request Link Status; accept Link Status response.
- **Tests:** Full handshake sequence.
- **Acceptance:** Connect succeeds only after both reset ACK + status.
- **Commit message:** `feat(master): link-status exchange on connect`
- **Handoff:** Connect sequence.
- **Stop condition:** Integration handshake test green.

### DNP3-008 — Application sequence continuity
- **Purpose:** Correct SEQ handling.
- **Prerequisites:** none
- **Scope:** Master request orchestration
- **Work:** Maintain 0-15 sequence; increment only on successful send.
- **Tests:** Sequence advances correctly across requests.
- **Acceptance:** Observed SEQs match expected stream.
- **Commit message:** `fix(master): application sequence continuity`
- **Handoff:** Sequence owner.
- **Stop condition:** Sequence unit tests green.

### DNP3-009 — Confirmation wait + timeout
- **Purpose:** CON bit correctness.
- **Prerequisites:** DNP3-008
- **Scope:** send path when CON=1
- **Work:** Wait for matching confirm; timeout → retry/fail.
- **Tests:** Confirm arrives / times out / wrong SEQ.
- **Acceptance:** Spec-required confirm behavior.
- **Commit message:** `fix(master): confirmation timeout and matching`
- **Handoff:** Confirm logic.
- **Stop condition:** Confirm tests green.

### DNP3-010 — Response sequence matching
- **Purpose:** Reject mismatched responses.
- **Prerequisites:** DNP3-008
- **Scope:** processResponse
- **Work:** Match response SEQ to outstanding request.
- **Tests:** Matching / mismatch.
- **Acceptance:** Mismatch returns error, no points.
- **Commit message:** `fix(master): response sequence validation`
- **Handoff:** Matching rules.
- **Stop condition:** Tests green.

### DNP3-011 — IIN bit semantics verification & correction
- **Purpose:** Align IIN with specification.
- **Prerequisites:** none
- **Scope:** `internal/al` IIN only
- **Work:** Verify bit positions against authoritative source; correct names/flags if wrong. If verification unavailable, document “requires verification” and freeze current mapping.
- **Tests:** Encode/decode of known IIN values.
- **Acceptance:** Bits match verified mapping (or frozen with note).
- **Commit message:** `fix(al): correct IIN bit semantics`
- **Handoff:** Updated IIN definition.
- **Stop condition:** IIN tests green.

### DNP3-012 — Master IIN storage & exposure
- **Purpose:** Surface IIN to callers.
- **Prerequisites:** DNP3-011
- **Scope:** Master + public ReadResponse
- **Work:** Always update and return latest IIN.
- **Tests:** IIN present after Class-0.
- **Acceptance:** Public response carries IIN.
- **Commit message:** `feat(master): expose IIN on responses`
- **Handoff:** IIN path.
- **Stop condition:** Loopback asserts IIN.

### DNP3-013 — Basic IIN reaction (DeviceRestart / NeedsTime)
- **Purpose:** Minimal required Master reaction.
- **Prerequisites:** DNP3-012
- **Scope:** After response processing
- **Work:** If DeviceRestart set → clear local state / re-integrity; if NeedsTime set → log or stub time-sync request (no full time objects yet).
- **Tests:** Injected IIN bits.
- **Acceptance:** Documented reaction occurs.
- **Commit message:** `feat(master): react to critical IIN bits`
- **Handoff:** Reaction policy.
- **Stop condition:** Reaction tests green.

### DNP3-014 — Class-0 integrity request construction
- **Purpose:** Canonical integrity poll.
- **Prerequisites:** DNP3-004
- **Scope:** Poll / Read path
- **Work:** Prefer G60V1 or explicit G1/G20/G30 headers; single consistent form.
- **Tests:** Request bytes match golden.
- **Acceptance:** Integrity request is deterministic.
- **Commit message:** `feat(master): canonical Class-0 integrity request`
- **Handoff:** Request builder.
- **Stop condition:** Request fixture test green.

### DNP3-015 — Multi-fragment Class-0 response handling
- **Purpose:** Large static data.
- **Prerequisites:** DNP3-005 (TL already present)
- **Scope:** Receive path
- **Work:** Ensure reassembler delivers complete APDU before parse.
- **Tests:** Multi-fragment golden response.
- **Acceptance:** All points present.
- **Commit message:** `fix(master): multi-fragment Class-0 reassembly`
- **Handoff:** Reassembly usage.
- **Stop condition:** Multi-fragment test green.

### DNP3-016 — Binary Input G1V1 final correctness
- **Purpose:** Lock G1V1.
- **Prerequisites:** DNP3-001, DNP3-005
- **Scope:** G1V1 only
- **Work:** Packed vs indexed, quality bits, index LSB.
- **Tests:** All existing + new boundary vectors.
- **Acceptance:** Golden pass; no residual deviations.
- **Commit message:** `fix(objects): finalize Binary Input Variation 1`
- **Handoff:** G1V1 locked.
- **Stop condition:** G1 vector suite green.

### DNP3-017 — Analog Input G30V1 final correctness
- **Purpose:** Lock G30V1.
- **Prerequisites:** DNP3-001, DNP3-005
- **Scope:** G30V1 only
- **Work:** 32-bit signed + quality, LSB, sequential vs indexed.
- **Tests:** Golden + boundaries.
- **Acceptance:** Matches supported-profile.
- **Commit message:** `fix(objects): finalize Analog Input Variation 1`
- **Handoff:** G30V1 locked.
- **Stop condition:** G30 vector suite green.

### DNP3-018 — Counter G20V1 final correctness
- **Purpose:** Lock G20V1.
- **Prerequisites:** DNP3-001, DNP3-005
- **Scope:** G20V1 only
- **Work:** 32-bit unsigned + quality, LSB.
- **Tests:** Golden + boundaries.
- **Acceptance:** Matches supported-profile.
- **Commit message:** `fix(objects): finalize Counter Variation 1`
- **Handoff:** G20V1 locked.
- **Stop condition:** G20 vector suite green.

### DNP3-019 — CROB G12V1 request final correctness
- **Purpose:** Lock control request.
- **Prerequisites:** DNP3-001
- **Scope:** CROB encode
- **Work:** Code, count, on/off times, status, index LSB.
- **Tests:** Golden direct-control vector.
- **Acceptance:** Request bytes match.
- **Commit message:** `fix(objects): finalize CROB request encoding`
- **Handoff:** CROB request locked.
- **Stop condition:** CROB request test green.

### DNP3-020 — CROB status parsing
- **Purpose:** Per-point command status (CTRL-01).
- **Prerequisites:** DNP3-019
- **Scope:** Operate response path
- **Work:** Parse status; map to ControlStatus; never claim success on failure.
- **Tests:** Success + rejected status vectors.
- **Acceptance:** Failed point ≠ ControlSuccess.
- **Commit message:** `fix(master): parse CROB command status`
- **Handoff:** Status mapping.
- **Stop condition:** Control status tests green.

### DNP3-021 — Public Operate response carries status
- **Purpose:** Surface status.
- **Prerequisites:** DNP3-020
- **Scope:** Public OperateResponse
- **Work:** Populate Status from parsed result.
- **Tests:** Public Operate asserts status.
- **Acceptance:** Caller sees real status.
- **Commit message:** `feat(api): expose control status on OperateResponse`
- **Handoff:** Public field.
- **Stop condition:** Public control test green.

### DNP3-022 — Context cancellation on Connect
- **Purpose:** SAFE-03 partial.
- **Prerequisites:** none
- **Scope:** Client.Connect
- **Work:** Honor ctx.Done(); abort transport + handshake.
- **Tests:** Cancelled context.
- **Acceptance:** Returns promptly, no live connection.
- **Commit message:** `fix(api): context cancellation on Connect`
- **Handoff:** Cancel path.
- **Stop condition:** Cancel test green.

### DNP3-023 — Context cancellation on Read
- **Purpose:** SAFE-03.
- **Prerequisites:** DNP3-022
- **Scope:** Client.Read
- **Work:** Cancel outstanding request.
- **Tests:** Cancel mid-read.
- **Acceptance:** Error, no partial points leaked.
- **Commit message:** `fix(api): context cancellation on Read`
- **Handoff:** Read cancel.
- **Stop condition:** Cancel test green.

### DNP3-024 — Context cancellation on Operate / Disconnect
- **Purpose:** Complete SAFE-03.
- **Prerequisites:** DNP3-023
- **Scope:** Operate + Disconnect/Close
- **Work:** Same pattern.
- **Tests:** Cancel cases.
- **Acceptance:** All public entry points cancel cleanly.
- **Commit message:** `fix(api): context cancellation on Operate and Disconnect`
- **Handoff:** Full cancel surface.
- **Stop condition:** Cancel suite green.

### DNP3-025 — Race safety for sequence & reassembly
- **Purpose:** SAFE-04.
- **Prerequisites:** DNP3-008, DNP3-015
- **Scope:** Master shared state
- **Work:** Mutex or serialize request path.
- **Tests:** `go test -race` concurrent Reads.
- **Acceptance:** Race detector clean.
- **Commit message:** `fix(master): serialize request and reassembly state`
- **Handoff:** Locking model.
- **Stop condition:** Race test green.

### DNP3-026 — Reject invalid CRC on receive
- **Purpose:** Robustness.
- **Prerequisites:** DLL already validates
- **Scope:** Master receive path
- **Work:** Surface CRC failure as error; no points.
- **Tests:** Injected bad CRC frame.
- **Acceptance:** Error, empty response.
- **Commit message:** `test(master): reject invalid CRC frames`
- **Handoff:** Error path.
- **Stop condition:** Negative test green.

### DNP3-027 — Reject truncated / oversize frames
- **Purpose:** Robustness.
- **Prerequisites:** DNP3-026
- **Scope:** Receive path
- **Work:** Length checks before parse.
- **Tests:** Truncated, claimed-length mismatch.
- **Acceptance:** Error, no panic.
- **Commit message:** `fix(master): reject truncated and oversize frames`
- **Handoff:** Length guards.
- **Stop condition:** Boundary tests green.

### DNP3-028 — Reject unsupported qualifiers
- **Purpose:** Spec safety.
- **Prerequisites:** DNP3-003
- **Scope:** Object-header decode
- **Work:** Only 0x06/0x07 (and any other v0-approved) accepted.
- **Tests:** Unsupported qualifier → error.
- **Acceptance:** Clear error, no data.
- **Commit message:** `fix(al): reject unsupported qualifiers`
- **Handoff:** Qualifier allow-list.
- **Stop condition:** Reject tests green.

### DNP3-029 — Reject unsupported groups/variations on read
- **Purpose:** Profile lock.
- **Prerequisites:** DNP3-014
- **Scope:** Public Read + internal
- **Work:** Only G1.1/G20.1/G30.1 (or G60.1) accepted for MVP.
- **Tests:** Other groups → unsupported error.
- **Acceptance:** Explicit error.
- **Commit message:** `fix(api): reject unsupported object groups for MVP`
- **Handoff:** Allow-list.
- **Stop condition:** Negative API tests green.

### DNP3-030 — Reject SBO / unsolicited / TLS already present — tighten remaining options
- **Purpose:** Complete profile gate.
- **Prerequisites:** none
- **Scope:** Public Config options
- **Work:** Ensure every non-v0 option returns clear error.
- **Tests:** Matrix of options.
- **Acceptance:** No silent fallback.
- **Commit message:** `fix(api): complete supported-profile rejection matrix`
- **Handoff:** Rejection table.
- **Stop condition:** Safety tests green.

### DNP3-031 — Transport disconnect detection
- **Purpose:** Recovery foundation.
- **Prerequisites:** none
- **Scope:** Master + transport adapter
- **Work:** Detect closed connection; set state Error/Disconnected.
- **Tests:** Peer close.
- **Acceptance:** State transitions correctly.
- **Commit message:** `feat(master): detect transport disconnect`
- **Handoff:** State transition.
- **Stop condition:** Disconnect test green.

### DNP3-032 — Reconnect + re-handshake
- **Purpose:** Recovery.
- **Prerequisites:** DNP3-007, DNP3-031
- **Scope:** Connect path after failure
- **Work:** Clear TL state, re-issue link reset/status.
- **Tests:** Drop mid-session then reconnect.
- **Acceptance:** Subsequent Read succeeds.
- **Commit message:** `feat(master): reconnect and re-handshake`
- **Handoff:** Recovery sequence.
- **Stop condition:** Recovery test green.

### DNP3-033 — Clear reassembly on reconnect
- **Purpose:** No stale fragments.
- **Prerequisites:** DNP3-032
- **Scope:** TL reassembler
- **Work:** Reset on reconnect.
- **Tests:** Partial fragment then reconnect.
- **Acceptance:** No cross-session pollution.
- **Commit message:** `fix(tl): reset reassembler on session restart`
- **Handoff:** Reset points.
- **Stop condition:** Isolation test green.

### DNP3-034 — Retry policy refinement
- **Purpose:** Spec-aligned retries.
- **Prerequisites:** DNP3-009
- **Scope:** sendWithRetry
- **Work:** Distinct handling for timeout vs NACK vs CRC error.
- **Tests:** Each failure class.
- **Acceptance:** Retry counts and delays correct.
- **Commit message:** `fix(master): refine retry policy by error class`
- **Handoff:** Retry table.
- **Stop condition:** Retry tests green.

### DNP3-035 — Public Read returns only supported point types
- **Purpose:** Clean API.
- **Prerequisites:** DNP3-016, DNP3-017, DNP3-018
- **Scope:** ReadResponse
- **Work:** Populate only Binary/Analog/Counter for MVP; others empty or omitted.
- **Tests:** Response shape.
- **Acceptance:** No unsupported types returned.
- **Commit message:** `fix(api): restrict ReadResponse to MVP types`
- **Handoff:** Response shape.
- **Stop condition:** Shape test green.

### DNP3-036 — Deterministic outstation simulator (MVP profile)
- **Purpose:** Test without physical device.
- **Prerequisites:** none
- **Scope:** `internal/testutils` or laboratory
- **Work:** Simulator answers Class-0 + CROB with golden data.
- **Tests:** Public loopback switches to simulator.
- **Acceptance:** Loopback green against simulator only.
- **Commit message:** `test: deterministic MVP outstation simulator`
- **Handoff:** Simulator API.
- **Stop condition:** Simulator-driven loopback green.

### DNP3-037 — Integrity poll convenience method
- **Purpose:** Clean Master surface.
- **Prerequisites:** DNP3-014
- **Scope:** Public Client
- **Work:** IntegrityPoll(ctx) → Read with Class-0.
- **Tests:** Convenience method.
- **Acceptance:** Same data as explicit Read.
- **Commit message:** `feat(api): IntegrityPoll convenience method`
- **Handoff:** Method.
- **Stop condition:** Test green.

### DNP3-038 — Document supported-profile in code comments
- **Purpose:** Traceability.
- **Prerequisites:** DNP3-030
- **Scope:** Public package docs
- **Work:** Comment every public option with Target/Reject/Defer.
- **Tests:** none (doc only)
- **Acceptance:** Comments match profile.
- **Commit message:** `docs(api): annotate supported-profile decisions`
- **Handoff:** Doc locations.
- **Stop condition:** Review only.

### DNP3-039 — Master state machine formalization
- **Purpose:** Clear states.
- **Prerequisites:** DNP3-007, DNP3-032
- **Scope:** internal/master State
- **Work:** Explicit transitions; illegal transitions return error.
- **Tests:** State transition table.
- **Acceptance:** No silent illegal transitions.
- **Commit message:** `refactor(master): formalize state machine transitions`
- **Handoff:** State diagram in comment.
- **Stop condition:** Transition tests green.

### DNP3-040 — Request outstanding tracking
- **Purpose:** One outstanding request per outstation for MVP.
- **Prerequisites:** DNP3-010
- **Scope:** Master
- **Work:** Track pending request; reject concurrent same-outstation request (or queue).
- **Tests:** Concurrent same-outstation.
- **Acceptance:** Defined behavior, no corruption.
- **Commit message:** `feat(master): track outstanding request per outstation`
- **Handoff:** Tracking model.
- **Stop condition:** Concurrency test green.

### DNP3-041 — Timeout configuration validation
- **Purpose:** Safe defaults.
- **Prerequisites:** none
- **Scope:** Config
- **Work:** Reject non-positive timeouts; sensible defaults.
- **Tests:** Config validation.
- **Acceptance:** Invalid config fails NewClient.
- **Commit message:** `fix(api): validate timeout and retry config`
- **Handoff:** Validation rules.
- **Stop condition:** Config tests green.

### DNP3-042 — Keep-alive / idle detection (minimal)
- **Purpose:** Detect dead peers.
- **Prerequisites:** DNP3-031
- **Scope:** Optional idle timeout
- **Work:** If configured, close idle connection.
- **Tests:** Idle close.
- **Acceptance:** State becomes Disconnected.
- **Commit message:** `feat(master): optional idle timeout`
- **Handoff:** Config flag.
- **Stop condition:** Idle test green.

### DNP3-043 — Error type taxonomy
- **Purpose:** Caller can distinguish failures.
- **Prerequisites:** none
- **Scope:** pkg/dnp3 errors + master
- **Work:** Distinct types/codes for timeout, CRC, sequence, unsupported, disconnect.
- **Tests:** Error classification.
- **Acceptance:** Errors inspectable.
- **Commit message:** `feat(api): structured protocol error types`
- **Handoff:** Error set.
- **Stop condition:** Error tests green.

### DNP3-044 — Logging hooks (optional, no-op default)
- **Purpose:** Observability without coupling.
- **Prerequisites:** none
- **Scope:** Master
- **Work:** Optional logger interface for frame/seq events.
- **Tests:** Hook called.
- **Acceptance:** Default silent.
- **Commit message:** `feat(master): optional diagnostic logger hook`
- **Handoff:** Interface.
- **Stop condition:** Hook test green.

### DNP3-045 — Public API loopback against simulator (full MVP)
- **Purpose:** End-to-end gate.
- **Prerequisites:** DNP3-016…021, DNP3-036
- **Scope:** test/integration
- **Work:** Single test: Connect → Integrity → Operate → assert points + status.
- **Tests:** That test.
- **Acceptance:** Passes deterministically.
- **Commit message:** `test(integration): full MVP public API loopback`
- **Handoff:** Gate test.
- **Stop condition:** Gate green.

### DNP3-046 — Negative: unexpected function code in response
- **Purpose:** Robustness.
- **Prerequisites:** DNP3-010
- **Scope:** processResponse
- **Work:** Reject non-RESPONSE/non-CONFIRM.
- **Tests:** Injected bad FC.
- **Acceptance:** Error.
- **Commit message:** `fix(master): reject unexpected response function codes`
- **Handoff:** Check.
- **Stop condition:** Test green.

### DNP3-047 — Negative: FIR/FIN inconsistency at AL
- **Purpose:** Robustness.
- **Prerequisites:** DNP3-015
- **Scope:** AL decode after reassembly
- **Work:** Single-fragment must be FIR+FIN; multi must match.
- **Tests:** Bad FIR/FIN.
- **Acceptance:** Error.
- **Commit message:** `fix(al): validate FIR/FIN consistency`
- **Handoff:** Validation.
- **Stop condition:** Test green.

### DNP3-048 — Negative: TL sequence gap
- **Purpose:** TL robustness.
- **Prerequisites:** none
- **Scope:** Reassembler
- **Work:** Already errors; ensure Master surfaces it.
- **Tests:** Gap injection.
- **Acceptance:** Error to caller.
- **Commit message:** `fix(master): surface TL sequence gap errors`
- **Handoff:** Error path.
- **Stop condition:** Test green.

### DNP3-049 — Master address configuration validation
- **Purpose:** Safety.
- **Prerequisites:** none
- **Scope:** Config
- **Work:** Reject reserved/broadcast where inappropriate.
- **Tests:** Invalid addresses.
- **Acceptance:** Clear error.
- **Commit message:** `fix(api): validate master and outstation addresses`
- **Handoff:** Rules.
- **Stop condition:** Test green.

### DNP3-050 — Clean resource cleanup on Close
- **Purpose:** No leaks.
- **Prerequisites:** DNP3-024
- **Scope:** Client.Close
- **Work:** Cancel, disconnect, nil transport, clear maps.
- **Tests:** Close then Connect again works.
- **Acceptance:** Reusable after Close.
- **Commit message:** `fix(api): thorough Close cleanup`
- **Handoff:** Cleanup order.
- **Stop condition:** Lifecycle test green.

### DNP3-051 — MVP documentation lock (README + SUPPORTED)
- **Purpose:** No over-claim.
- **Prerequisites:** DNP3-045
- **Scope:** docs
- **Work:** Every claimed capability has test reference; unsupported listed.
- **Tests:** none
- **Acceptance:** Docs match reality.
- **Commit message:** `docs: lock claims to verified Master MVP`
- **Handoff:** Doc files.
- **Stop condition:** Review.

### DNP3-052 — Verification script for MVP gate
- **Purpose:** REL-01 style.
- **Prerequisites:** DNP3-045
- **Scope:** scripts/
- **Work:** Single command runs unit + integration + race for MVP.
- **Tests:** Script itself.
- **Acceptance:** Exit 0 on clean tree.
- **Commit message:** `test: add MVP verification script`
- **Handoff:** Script path.
- **Stop condition:** Script green.

### DNP3-053 — IIN-triggered integrity re-poll (DeviceRestart)
- **Purpose:** Correct Master behavior.
- **Prerequisites:** DNP3-013
- **Scope:** After response
- **Work:** Auto-schedule integrity if restart bit set (configurable).
- **Tests:** Injected restart bit.
- **Acceptance:** Integrity follows.
- **Commit message:** `feat(master): auto integrity after DeviceRestart IIN`
- **Handoff:** Policy.
- **Stop condition:** Test green.

### DNP3-054 — Confirm generation for CON responses (Master side)
- **Purpose:** When Master must confirm.
- **Prerequisites:** DNP3-009
- **Scope:** Receive path
- **Work:** If response CON=1, send confirm before processing data.
- **Tests:** CON response.
- **Acceptance:** Confirm sent with matching SEQ.
- **Commit message:** `feat(master): send application confirm when required`
- **Handoff:** Confirm send.
- **Stop condition:** Test green.

### DNP3-055 — Session isolation for multiple outstations (basic)
- **Purpose:** Map correctness.
- **Prerequisites:** DNP3-040
- **Scope:** Master outstation map
- **Work:** Per-outstation sequence / pending.
- **Tests:** Two outstations sequential.
- **Acceptance:** No cross-talk.
- **Commit message:** `feat(master): per-outstation request state`
- **Handoff:** Map structure.
- **Stop condition:** Multi-outstation test green.

### DNP3-056 — Final MVP acceptance gate
- **Purpose:** Declare MVP complete.
- **Prerequisites:** DNP3-045, DNP3-051, DNP3-052
- **Scope:** All prior
- **Work:** Run verification script; record results.
- **Tests:** Full suite.
- **Acceptance:** All green; profile documented.
- **Commit message:** `test: Master MVP acceptance gate passed`
- **Handoff:** Gate record.
- **Stop condition:** Gate green. **MVP COMPLETE.**

### DNP3-057 — Link FCB/FCV handling (Master primary)
- **Purpose:** Later correctness.
- **Prerequisites:** DNP3-007
- **Scope:** DLL control on confirmed user data
- **Work:** Maintain FCB; set FCV correctly.
- **Tests:** FCB toggle.
- **Acceptance:** Matches expected pattern.
- **Commit message:** `feat(dll/master): FCB/FCV on confirmed user data`
- **Handoff:** FCB state.
- **Stop condition:** FCB tests green.

### DNP3-058 — Secondary NACK handling
- **Purpose:** Robustness.
- **Prerequisites:** DNP3-006
- **Scope:** Handshake / data path
- **Work:** Treat NACK as failure; retry or error.
- **Tests:** NACK injection.
- **Acceptance:** Defined behavior.
- **Commit message:** `feat(master): handle link-layer NACK`
- **Handoff:** NACK policy.
- **Stop condition:** Test green.

### DNP3-059 — Transport fragment size boundary tests
- **Purpose:** TL completeness.
- **Prerequisites:** none
- **Scope:** TL
- **Work:** Exact 249 / 250 / 0 sizes.
- **Tests:** Boundary fragmentize/reassemble.
- **Acceptance:** No off-by-one.
- **Commit message:** `test(tl): fragment size boundary vectors`
- **Handoff:** Vectors.
- **Stop condition:** Tests green.

### DNP3-060 — AL sequence wrap-around
- **Purpose:** Correctness.
- **Prerequisites:** DNP3-008
- **Scope:** Sequence
- **Work:** 15→0 wrap.
- **Tests:** Wrap.
- **Acceptance:** Continues correctly.
- **Commit message:** `fix(master): application sequence wrap-around`
- **Handoff:** Wrap logic.
- **Stop condition:** Test green.

### DNP3-061 — Malformed object data rejection
- **Purpose:** Robustness.
- **Prerequisites:** DNP3-005
- **Scope:** Parsers
- **Work:** Truncated point data → error.
- **Tests:** Truncated points.
- **Acceptance:** Error, no partial list.
- **Commit message:** `fix(master): reject truncated object data`
- **Handoff:** Guard.
- **Stop condition:** Test green.

### DNP3-062 — Quality flag mapping finalization
- **Purpose:** Interop.
- **Prerequisites:** DNP3-016…018
- **Scope:** Quality bits
- **Work:** Map ONLINE/RESTART/etc. to public flags; document.
- **Tests:** Quality vectors.
- **Acceptance:** Matches profile notes.
- **Commit message:** `fix(objects): finalize quality flag mapping`
- **Handoff:** Mapping table.
- **Stop condition:** Tests green.

### DNP3-063 — Public API example (minimal Master)
- **Purpose:** Usability.
- **Prerequisites:** DNP3-056
- **Scope:** examples/
- **Work:** Short Connect/Integrity/Operate example.
- **Tests:** Builds.
- **Acceptance:** Compiles against public API.
- **Commit message:** `docs: minimal Master example`
- **Handoff:** Example path.
- **Stop condition:** Build green.

### DNP3-064 — Master metrics hooks (optional)
- **Purpose:** Observability.
- **Prerequisites:** DNP3-044
- **Scope:** Counters for requests/timeouts/CRC errors
- **Work:** Optional metrics interface.
- **Tests:** Incremented.
- **Acceptance:** Default no-op.
- **Commit message:** `feat(master): optional metrics hooks`
- **Handoff:** Interface.
- **Stop condition:** Test green.

### DNP3-065 — Double-check DLL EncodedSize usage everywhere
- **Purpose:** No length bugs.
- **Prerequisites:** none
- **Scope:** All receive loops
- **Work:** Audit offset advancement.
- **Tests:** Concatenated frames.
- **Acceptance:** No over/under-read.
- **Commit message:** `fix: consistent EncodedSize usage on receive`
- **Handoff:** Audit result.
- **Stop condition:** Concat tests green.

### DNP3-066 — Confirm timeout distinct from response timeout
- **Purpose:** Spec nuance.
- **Prerequisites:** DNP3-009
- **Scope:** Timeouts
- **Work:** Separate config or documented relation.
- **Tests:** Distinct values.
- **Acceptance:** Behavior documented + tested.
- **Commit message:** `feat(master): distinct confirm timeout`
- **Handoff:** Config.
- **Stop condition:** Test green.

### DNP3-067 — Unsupported variation explicit error messages
- **Purpose:** Diagnostics.
- **Prerequisites:** DNP3-029
- **Scope:** Errors
- **Work:** Message names group/variation.
- **Tests:** Error text.
- **Acceptance:** Clear.
- **Commit message:** `fix(api): descriptive unsupported object errors`
- **Handoff:** Messages.
- **Stop condition:** Test green.

### DNP3-068 — Master restart after Close
- **Purpose:** Lifecycle.
- **Prerequisites:** DNP3-050
- **Scope:** NewClient after Close of previous
- **Work:** Ensure no global state leakage.
- **Tests:** Sequential clients.
- **Acceptance:** Independent.
- **Commit message:** `fix(api): clean restart after Close`
- **Handoff:** Lifecycle.
- **Stop condition:** Test green.

### DNP3-069 — Integration test matrix documentation
- **Purpose:** Traceability.
- **Prerequisites:** DNP3-056
- **Scope:** laboratory or docs
- **Work:** Table of capability → test name.
- **Tests:** none
- **Acceptance:** Complete for MVP.
- **Commit message:** `docs: Master MVP capability-to-test matrix`
- **Handoff:** Matrix.
- **Stop condition:** Review.

### DNP3-070 — Final Master residual gap scan
- **Purpose:** Catch anything missed.
- **Prerequisites:** DNP3-056
- **Scope:** All Master code
- **Work:** Code review against audit table; open tasks if needed.
- **Tests:** none
- **Acceptance:** No untracked CRITICAL/PARTIAL.
- **Commit message:** `docs: Master residual gap scan results`
- **Handoff:** Scan report.
- **Stop condition:** Report committed.

### DNP3-071 — CROB multiple points (optional batch)
- **Purpose:** Future-proof.
- **Prerequisites:** DNP3-020
- **Scope:** Operate path
- **Work:** Support count >1 if already partially present; otherwise reject cleanly.
- **Tests:** Single + multi (or reject).
- **Acceptance:** Defined.
- **Commit message:** `feat(master): multi-point CROB or explicit single-only`
- **Handoff:** Policy.
- **Stop condition:** Test green.

### DNP3-072 — Master handoff.md template for implementers
- **Purpose:** Recoverability.
- **Prerequisites:** none
- **Scope:** active_work or root
- **Work:** Template recording completed task IDs, next READY, test commands.
- **Tests:** none
- **Acceptance:** Usable by next agent.
- **Commit message:** `docs: implementation handoff template`
- **Handoff:** Template.
- **Stop condition:** File present.

---

## 6. OUTSTATION REMAINING-CAPACITY TASKS (DNP3-073 … DNP3-100)

### DNP3-073 — Outstation Class-0 response builder uses object-header model
- **Prerequisites:** DNP3-002, DNP3-003
- **Scope:** Outstation response path
- **Work:** Emit headers via model.
- **Tests:** Golden match.
- **Acceptance:** Bytes match Master expectations.
- **Commit message:** `refactor(outstation): object-header Class-0 responses`
- **Stop condition:** Vector green.

### DNP3-074 — Outstation G1V1 / G20V1 / G30V1 final lock
- **Prerequisites:** DNP3-016…018
- **Scope:** Builders
- **Work:** Same wire rules as Master parsers.
- **Tests:** Round-trip with Master.
- **Acceptance:** Loopback perfect.
- **Commit message:** `fix(outstation): finalize static object builders`
- **Stop condition:** Round-trip green.

### DNP3-075 — Outstation CROB status response
- **Prerequisites:** DNP3-020
- **Scope:** Command handler response
- **Work:** Return correct status codes.
- **Tests:** Success + fail.
- **Acceptance:** Master parses correctly.
- **Commit message:** `feat(outstation): CROB status in operate response`
- **Stop condition:** Control loopback green.

### DNP3-076 — Outstation link ACK generation
- **Prerequisites:** DNP3-006
- **Scope:** Link layer responses
- **Work:** Correct secondary ACK for reset.
- **Tests:** Handshake with Master.
- **Acceptance:** Handshake succeeds.
- **Commit message:** `fix(outstation): correct link ACK generation`
- **Stop condition:** Handshake green.

### DNP3-077 — Outstation link-status response
- **Prerequisites:** DNP3-007
- **Scope:** Link
- **Work:** Respond to Request Link Status.
- **Tests:** Full handshake.
- **Acceptance:** Master Connect succeeds.
- **Commit message:** `feat(outstation): link-status response`
- **Stop condition:** Test green.

### DNP3-078 — Outstation IIN population
- **Prerequisites:** DNP3-011
- **Scope:** Response
- **Work:** Set IIN bits from internal state.
- **Tests:** IIN visible to Master.
- **Acceptance:** Master receives expected bits.
- **Commit message:** `feat(outstation): populate IIN`
- **Stop condition:** IIN test green.

### DNP3-079 — Outstation multi-fragment response
- **Prerequisites:** DNP3-015
- **Scope:** TL send
- **Work:** Fragment large Class-0.
- **Tests:** Multi-fragment receive by Master.
- **Acceptance:** Complete data.
- **Commit message:** `feat(outstation): multi-fragment Class-0`
- **Stop condition:** Test green.

### DNP3-080 — Outstation reject unsupported function codes
- **Prerequisites:** none
- **Scope:** AL dispatch
- **Work:** Return IIN or error response for unsupported.
- **Tests:** Bad FC.
- **Acceptance:** Master sees clear failure.
- **Commit message:** `fix(outstation): reject unsupported function codes`
- **Stop condition:** Test green.

### DNP3-081 — Outstation simulator configuration (points)
- **Prerequisites:** DNP3-036
- **Scope:** Simulator
- **Work:** Configurable point counts/values for tests.
- **Tests:** Different configs.
- **Acceptance:** Tests can vary data.
- **Commit message:** `test: configurable simulator point sets`
- **Stop condition:** Tests green.

### DNP3-082 — Outstation session reset on link reset
- **Prerequisites:** DNP3-076
- **Scope:** State
- **Work:** Clear buffers on reset.
- **Tests:** Reset then new request.
- **Acceptance:** Clean state.
- **Commit message:** `fix(outstation): reset state on link reset`
- **Stop condition:** Test green.

### DNP3-083 — Outstation FCB tracking (secondary)
- **Prerequisites:** DNP3-057
- **Scope:** Link
- **Work:** Track Master FCB; detect repeats.
- **Tests:** FCB repeat.
- **Acceptance:** Defined behavior.
- **Commit message:** `feat(outstation): secondary FCB tracking`
- **Stop condition:** Test green.

### DNP3-084 — Outstation concurrent connection rejection (MVP single)
- **Prerequisites:** none
- **Scope:** Server
- **Work:** Reject second connection or document single.
- **Tests:** Second connect.
- **Acceptance:** Defined.
- **Commit message:** `fix(outstation): single-connection MVP policy`
- **Stop condition:** Test green.

### DNP3-085 — Outstation context cancellation on Start/Stop
- **Prerequisites:** none
- **Scope:** Server API
- **Work:** Honor ctx.
- **Tests:** Cancel.
- **Acceptance:** Clean stop.
- **Commit message:** `fix(outstation): context cancellation`
- **Stop condition:** Test green.

### DNP3-086 — Outstation malformed frame rejection
- **Prerequisites:** DNP3-026
- **Scope:** Receive
- **Work:** Same CRC/length guards.
- **Tests:** Bad frames.
- **Acceptance:** No crash, no bad response.
- **Commit message:** `fix(outstation): reject malformed frames`
- **Stop condition:** Test green.

### DNP3-087 — Outstation public API profile lock
- **Prerequisites:** DNP3-030
- **Scope:** Server options
- **Work:** Reject unsolicited/TLS/etc.
- **Tests:** Matrix.
- **Acceptance:** Clear errors.
- **Commit message:** `fix(outstation): supported-profile rejection`
- **Stop condition:** Tests green.

### DNP3-088 — Outstation deterministic event buffer stub (empty)
- **Prerequisites:** none
- **Scope:** Future events
- **Work:** Empty Class 1-3; return empty or IIN.
- **Tests:** Class 1 request → empty/IIN.
- **Acceptance:** No crash.
- **Commit message:** `feat(outstation): stub empty event classes`
- **Stop condition:** Test green.

### DNP3-089 — Outstation data handler interface hardening
- **Prerequisites:** DNP3-074
- **Scope:** DataHandler
- **Work:** Document required methods for MVP objects only.
- **Tests:** Interface compliance.
- **Acceptance:** Compiles with minimal handler.
- **Commit message:** `refactor(outstation): MVP data handler contract`
- **Stop condition:** Compile + test green.

### DNP3-090 — Outstation command handler interface hardening
- **Prerequisites:** DNP3-075
- **Scope:** CommandHandler
- **Work:** CROB only for MVP.
- **Tests:** Analog command rejected.
- **Acceptance:** Clear error.
- **Commit message:** `refactor(outstation): MVP command handler contract`
- **Stop condition:** Test green.

### DNP3-091 — Outstation integration test symmetry
- **Prerequisites:** DNP3-045
- **Scope:** test/integration
- **Work:** Mirror Master gate from Outstation side.
- **Tests:** Same scenarios.
- **Acceptance:** Both directions green.
- **Commit message:** `test(integration): Outstation-side MVP gate`
- **Stop condition:** Gate green.

### DNP3-092 — Outstation recovery after Master reconnect
- **Prerequisites:** DNP3-032
- **Scope:** Session
- **Work:** Accept new link reset after disconnect.
- **Tests:** Master reconnect.
- **Acceptance:** Second session works.
- **Commit message:** `fix(outstation): accept post-reconnect link reset`
- **Stop condition:** Test green.

### DNP3-093 — Outstation logging/metrics hooks
- **Prerequisites:** DNP3-044
- **Scope:** Parallel to Master
- **Work:** Optional hooks.
- **Tests:** Called.
- **Acceptance:** Default silent.
- **Commit message:** `feat(outstation): optional diagnostic hooks`
- **Stop condition:** Test green.

### DNP3-094 — Outstation EncodedSize receive audit
- **Prerequisites:** DNP3-065
- **Scope:** Receive loop
- **Work:** Same as Master.
- **Tests:** Concat frames.
- **Acceptance:** Correct.
- **Commit message:** `fix(outstation): consistent EncodedSize usage`
- **Stop condition:** Test green.

### DNP3-095 — Outstation example
- **Prerequisites:** DNP3-063
- **Scope:** examples/
- **Work:** Minimal server.
- **Tests:** Builds.
- **Acceptance:** Compiles.
- **Commit message:** `docs: minimal Outstation example`
- **Stop condition:** Build green.

### DNP3-096 — Outstation residual gap scan
- **Prerequisites:** DNP3-091
- **Scope:** Outstation code
- **Work:** Scan vs audit table.
- **Tests:** none
- **Acceptance:** Report.
- **Commit message:** `docs: Outstation residual gap scan`
- **Stop condition:** Report committed.

### DNP3-097 — Shared test fixture library cleanup
- **Prerequisites:** DNP3-036, DNP3-081
- **Scope:** testutils
- **Work:** Common golden loaders, frame builders.
- **Tests:** Used by both.
- **Acceptance:** No duplication.
- **Commit message:** `test: shared fixture helpers`
- **Stop condition:** Tests still green.

### DNP3-098 — Conformance test enablement (DLL/TL/AL existing)
- **Prerequisites:** none
- **Scope:** test/conformance
- **Work:** Ensure existing suites run and pass against current code.
- **Tests:** Those suites.
- **Acceptance:** CI-runnable.
- **Commit message:** `test: enable existing conformance suites`
- **Stop condition:** Suites green.

### DNP3-099 — Final dual-role verification script
- **Prerequisites:** DNP3-052, DNP3-091
- **Scope:** scripts/
- **Work:** Master + Outstation gates.
- **Tests:** Script.
- **Acceptance:** Exit 0.
- **Commit message:** `test: dual-role verification script`
- **Stop condition:** Script green.

### DNP3-100 — Project-level handoff completion
- **Prerequisites:** all prior
- **Scope:** handoff.md
- **Work:** Record 100/100, next steps (events, SBO, time, SA).
- **Tests:** none
- **Acceptance:** Fresh agent can continue from handoff alone.
- **Commit message:** `docs: complete implementation handoff for 100 tasks`
- **Stop condition:** Handoff final.

---

## 7. DEPENDENCY / READY MAP

**READY immediately** (no prereqs):  
001, 002, 006, 008, 011, 022, 026 (partial), 030, 031, 036, 038, 041, 043, 044, 049, 059, 065, 072, 080, 084, 085, 087, 088, 098

**Key chains**  
002 → 003 → 004/005 → 014/015/016-018  
006 → 007 → 032  
008 → 009/010 → 040  
011 → 012 → 013 → 053  
016-021 + 036 → 045 → 056 (MVP)  
Master MVP gate at **DNP3-056**.  
Outstation tasks mostly after corresponding Master locks (073+ after 002/016 etc.).

Agents may take any READY task, implement, test, update handoff, repeat.

Every 3 completed micro-tasks: verify → update handoff.md → commit → push to main.

---

## 8. ROADMAP AUDIT

- Exactly 100 tasks, sequential IDs 001–100.
- Every required field present.
- Every CRITICAL/PARTIAL Master gap mapped to a task or explicit LATER decision.
- No silent omission of protocol-critical items (CRC, seq, confirm, objects, handshake, recovery, rejection).
- Dependencies real; many parallel READY.
- Master prioritized (72 tasks); Outstation uses remaining 28 for correctness + Master test support.
- MVP explicitly ends at DNP3-056.
- Unsupported capabilities documented.
- Protocol-critical tasks carry tests.
- No unnecessary feature work (events/SBO/time deferred).
- Handoff design supports multi-agent continuation.

---

## 9. IMPLEMENTATION START POINT

Begin with any READY task.  
**Recommended first:** DNP3-001 (endian audit) or DNP3-002 (object-header model).

```
TOTAL TASKS: 100
MASTER TASKS: 72
OUTSTATION TASKS: 28
MVP COMPLETE AT: DNP3-056
NEXT TASK: DNP3-001 — Residual endian audit
```

---

*End of roadmap. Implementation agents must update `active_work/handoff.md` after every task.*
