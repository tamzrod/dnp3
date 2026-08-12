# DNP3 Implementation Handoff

**Last updated:** 2026-08-12  
**Roadmap:** `active_work/DNP3_MASTER_ROADMAP.md`  
**Profile:** `active_work/supported-profile.md`

## Status

- Planning complete.
- DNP3-001, DNP3-002, DNP3-003 complete. Implementation underway.
- Go 1.22 toolchain installed at `~/go-install/go/bin/go` (not preinstalled in env).

## Completed Tasks

### DNP3-001 — Residual endian audit
- Commit message: `fix(objects): eliminate residual big-endian in MVP paths`
- `pkg/dnp3/master/client.go`: converted `parseAnalogInputs` (G30 V1/V2/V3/V5),
  `parseCounters` (G20 V1/V5/V6), `parseBinaryOutputs` (G10 index),
  `parseAnalogOutputs` (G40 index + V1/V2) from `binary.BigEndian` to
  `binary.LittleEndian`.
- `internal/master/master.go`: converted the 7 `Read*` range builders
  (G1/G3/G10/G20/G21/G30/G40) from BE start/stop to LSB-first; converted
  `buildControlRequest` G12 CROB OnTime/OffTime from BE to LSB.
- `test/integration/tcp_test.go`: fixed the test helper parsers to handle the
  v0 `0x07` count qualifier and use LSB. This fixed the pre-existing
  `TestMasterOutstationEndToEndComprehensive` failure.
- New BE-negative tests in `pkg/dnp3/master/object_vector_test.go`,
  `internal/master/control_vector_test.go`, `internal/master/master_test.go`.
- Out of scope (v0 profile rejects): G41-G44 analog output builders and
  `encodeDNP3Time` still use BE; explicitly rejected by supported-profile.md.

### DNP3-002 — Formal object-header model (encode)
- Commit message: `feat(al): object header encode model`
- New file `internal/al/object_header.go`: `ObjectHeader` struct
  (Group, Variation, Qualifier, Count, Start, Stop), `Encode`, `EncodedSize`,
  `EncodeObjectHeaders`, and qualifier constants (`QualAllObjects`=0x06,
  `QualCount8`=0x07, `QualIndex8`=0x00, `QualRange16`=0x28, `QualCount16`=0x27).
- No caller changes yet (new types only).
- Tests in `internal/al/object_header_test.go`: round-trip encode of 0x06/0x07
  Class-0 headers, append behavior, multi-header encode, unsupported-qualifier
  error, EncodedSize consistency, range16 LSB.

### DNP3-003 — Formal object-header model (decode)
- Commit message: `feat(al): object header decode and validation`
- Added `DecodeObjectHeader(data, offset)` and `ValidQualifier(q)` to
  `internal/al/object_header.go`. Returns `(ObjectHeader, consumed, error)`.
  Rejects unsupported qualifiers, truncated input.
- Tests: valid round-trip decode for all v0 qualifiers, offset decode,
  truncated rejection, unsupported-qualifier rejection, count8 LSB (golden
  G30V1 vector), range16 LSB, `ValidQualifier` coverage.

### DNP3-004 — Wire Master read path to object-header model
- Commit message: `refactor(master): use object-header model for reads`
- `internal/master/master.go`: `buildPollRequest` now constructs poll
  headers via `al.ObjectHeader`/`al.EncodeObjectHeaders` (preserving the
  prior `0x07 0x00` wire bytes for Class-0/event polls). The 7 `Read*`
  range builders now use a new `buildReadRangeRequest` helper that emits
  the `0x28` (range16) qualifier via `al.ObjectHeader`.
- `pkg/dnp3/master/client.go`: `buildReadRequest` now uses `al.ObjectHeader`
  with the `0x06` (all-objects) qualifier.
- New golden test `TestBuildReadRequestGolden` verifies the public Read
  builder produces the expected Class-0 request bytes.
- `TestReadRangeQualifierLSB` rewritten to verify `buildReadRangeRequest`
  emits the `0x28` range16 header LSB-first.
- Acceptance: generated request headers match golden; loopback green.

### DNP3-005 — Wire Master response parse to object-header model
- Commit message: `refactor(master): object-header based response parse`
- `pkg/dnp3/master/client.go`: `parseBinaryInputs`, `parseAnalogInputs`,
  `parseCounters`, `parseBinaryOutputs`, `parseAnalogOutputs` now decode
  each object header via `al.DecodeObjectHeader` instead of ad-hoc byte
  reads. Point parsing logic is unchanged.
- Acceptance: same points returned as before; golden Class-0 response
  vectors and loopback/integration all green.

### DNP3-006 — Link handshake ACK validation
- Commit message: `fix(master): validate link-layer ACK`
- `internal/master/master.go`: `sendResetLink` now decodes the secondary
  ACK frame and validates it via new `validateResetLinkACK`. The ACK must
  be a well-formed frame (sync + CRC via `frame.Decode`), DIR=0, PRM=0,
  function code ACK (0), SrcAddr=outstation, DestAddr=master. Any
  deviation fails the handshake.
- Tests: `TestValidateResetLinkACK` covers good ACK, bad FC (NACK),
  wrong DIR, wrong PRM, wrong source address, wrong destination address,
  and malformed frame (no sync).
- Acceptance: handshake fails on invalid ACK; integration tests (real
  outstation sends valid ACK) still green.

### DNP3-007 — Link-status request after reset
- Commit message: `fix(master): link status request after reset`
- `internal/master/master.go`: `performLinkHandshake` now issues a Request
  Link Status (FC=9) after the Reset Link ACK and validates the secondary
  Link Status response (FC=2) via new `sendLinkStatusRequest` +
  `validateLinkStatusResponse`. Connect succeeds only after both exchanges.
- Tests: `TestValidateLinkStatusResponse` covers good link status, wrong
  FC, wrong DIR, wrong PRM, wrong source/destination address, malformed.
- Acceptance: integration handshake (real outstation sends valid Link
  Status) green.

### DNP3-008 — Application sequence continuity
- Commit message: `fix(master): application sequence continuity`
- `internal/master/master.go`: added `sequence` field (0-15, mutex-guarded)
  with `nextSequence`/`advanceSequence`/`currentSequence`. `sendWithRetry`
  and `sendWithRetryAndGetResponse` allocate the sequence per request and
  advance it only on successful send (no advance on send failure). Retries
  reuse the same seq within one logical request.
- Tests: `TestSequenceStream` (0-15 wrap), `TestSendWithRetrySequenceAdvances`
  (observed SEQs 0,1,2), `TestSendWithRetrySequenceNoAdvanceOnSendFailure`.
- Acceptance: observed SEQs match the expected 0-15 stream.

### DNP3-009 — Confirmation wait + timeout
- Commit message: `fix(master): confirmation timeout and matching`
- `internal/master/master.go`: rewrote `waitForConfirmation`. A dedicated
  confirm (FuncResponse, IIN-only) must match the request SEQ or returns
  `ErrConfirmSeqMismatch`; a transport receive error surfaces as
  `ErrConfirmTimeout` (caller retries). Full responses acting as confirm
  are still accepted (strict response-seq matching deferred to DNP3-010).
- Tests: `TestWaitForConfirmationMatchingSeq`, `TestWaitForConfirmationWrongSeq`,
  `TestWaitForConfirmationTimeout`.
- Acceptance: spec-required confirm (match / mismatch / timeout) behavior.

### DNP3-010 — Response sequence matching
- Commit message: `fix(master): response sequence matching`
- `internal/master/master.go`: `processResponse` now takes `expectedSeq` and
  rejects responses whose application SEQ does not match the outstanding
  request (`ErrResponseSeqMismatch`, no data surfaced). Callers
  (`sendWithRetry`, `sendWithRetryAndGetResponse`) pass the request seq.
- `internal/outstation/outstation.go`: `ProcessRequest` now echoes the
  request's SEQ in solicited responses (spec-compliant outstation behavior).
- Tests: `TestProcessResponseMatchingSeq`, `TestProcessResponseMismatchSeq`,
  `TestSendWithRetrySequenceAdvances` (rewritten to use an echo-seq mock
  transport that returns a response matching each request's SEQ).
- Acceptance: matching SEQ accepted; mismatch rejected with no data.

### DNP3-011 — IIN bit semantics verification & correction
- Commit message: `fix(al): correct IIN bit semantics`
- `internal/al/application.go`: IIN struct field names/semantics corrected to
  the verified IEEE 1815-2012 mapping (bit POSITIONS/hex masks unchanged:
  IIN1.0=0x80 ... IIN1.7=0x01, IIN2.0=0x80 ... IIN2.7=0x01). Corrected names:
  IIN1: AllStations, Class1Events, Class2Events, Class3Events, NeedTime,
  LocalControl, DeviceTrouble, DeviceRestart.
  IIN2: FuncUnknown, ObjectUnknown, ParameterError, BufferOverflow,
  AlreadyExecuting, BadConfig, Reserved2_6, Reserved2_7.
  Previous names (AllStop/ByteOver/Limit64K/Busy/ParamUnavail/NeedsTimeSync/
  ConfigError/...) were a garbled, incorrect mapping.
- `internal/outstation/outstation.go`: fixed semantic misuse — buffer-full now
  sets `BufferOverflow` (IIN2.3, was incorrectly ByteOver=0x40=Class1Events);
  parameter unavailable now sets `ParameterError` (IIN2.2, was incorrectly
  ParamUnavail=0x01=DeviceRestart); generic error response now sets
  `FuncUnknown` (IIN2.0).
- Tests: corrected `internal/al/application_test.go` and
  `test/conformance/al/al_test.go` to assert the verified mapping; added
  `TestIINBitPositions` (per-bit golden). Updated `internal/outstation`,
  `internal/testutils`, `test/integration`, `benchmarks` references.
- Source verified: IEEE 1815-2012 §10.5.1 (via DNP Users Group device-profile
  IIN table — Hindlepower/ATevo manual reproduces the canonical Object 80
  bit index table).
- Acceptance: IIN bits match verified mapping; all IIN tests green.

### DNP3-012 — Master IIN storage & exposure
- Commit message: `feat(master): expose IIN on responses`
- `internal/master/master.go`: added thread-safe `Outstation.GetIIN() [2]byte`
  (IIN already updated on every `processResponse`).
- `pkg/dnp3/master/client.go`: added `LastIIN() [2]byte` to the public
  `Client` interface + implementation (returns the master's stored IIN for
  the configured outstation). The public `ReadResponse.IIN` already carried
  the per-response IIN; this exposes the stored copy.
- Tests: `TestProcessResponseStoresIIN` (internal: verifies the stored IIN
  equals the response IIN), loopback
  `TestPublicAPILoopbackReadAndDirectControl` asserts
  `resp.IIN == client.LastIIN()`.
- Acceptance: public response carries IIN; loopback asserts IIN.

## Next READY Tasks

- **DNP3-013** — Basic IIN reaction (DeviceRestart / NeedTime)  *(prereqs: DNP3-012 ✓)*
- **DNP3-022** — Context cancellation on Connect
- **DNP3-030** — Complete supported-profile rejection matrix
- **DNP3-031** — Transport disconnect detection
- **DNP3-036** — Deterministic outstation simulator (MVP profile)
- **DNP3-041** — Timeout configuration validation
- **DNP3-043** — Error type taxonomy
- **DNP3-044** — Logging hooks
- **DNP3-049** — Master address configuration validation
- **DNP3-059** — Transport fragment size boundary tests
- **DNP3-065** — Double-check DLL EncodedSize usage
- **DNP3-072** — Master handoff.md template
- **DNP3-080, 084, 085, 087, 088, 098** (Outstation-side READY tasks)

## Recommended Next Task

**DNP3-013 — Basic IIN reaction (DeviceRestart / NeedTime)**

After completing a task:

1. Run the relevant unit / golden / loopback tests.
2. Update this handoff (move task to Completed, note any new READY tasks, record test commands used).
3. Commit with the exact commit message from the roadmap.
4. Every 3 completed tasks: verify → update handoff.md → commit → push to main.

## Test Commands (baseline)

```bash
export PATH=$HOME/go-install/go/bin:$PATH   # Go toolchain location
go test ./...
go test -race ./internal/master/... ./pkg/dnp3/...
go test ./test/integration/...
```

## Notes for Agents

- Do **not** implement outside the defined micro-tasks.
- Do **not** invent requirements or add unsupported objects/features.
- Prefer deterministic fixtures and golden vectors.
- Keep changes commit-sized and independently testable.
- Master has priority; Outstation tasks exist only to support Master correctness and testing.
- DNP3 wire fields are LSB-first (little-endian) for all multi-octet values
  (indices, ranges, analog/counter values, CROB times). Confirmed by golden
  vectors in `active_work/testdata/` and `active_work/supported-profile.md`.

## Implementation Discoveries

- The pre-existing `TestMasterOutstationEndToEndComprehensive` failure was
  caused by the integration test's own helper parsers using `binary.BigEndian`
  and not handling the v0 `0x07` count qualifier. Fixed in DNP3-001.
- The outstation encoder (`internal/outstation/outstation.go`) already encodes
  G1V1/G30V1/G20V1 with the correct `0x07` qualifier and LSB values; the bug was
  only on the Master parse side and in test helpers.
- There are two CROB builders: `buildCROBRequest` (used by `WriteBinaryOutput`,
  already LSB) and `buildControlRequest` (used by `Operate`, was BE for
  OnTime/OffTime — now fixed to LSB).
- The DNP3 object-header qualifier byte is structured as high nibble (mode) +
  low nibble (size). For v0: 0x06 = all-objects (4-byte header, range byte=0x00),
  0x07 = 8-bit count (4-byte header), 0x00 = 8-bit index/count, 0x28 = 16-bit
  start/stop (7-byte header), 0x27 = 16-bit count (5-byte header).
- `ObjectHeader.EncodedSize()` returns the header+range byte count (excludes
  per-point object data). Use it to advance offsets when scanning response data.

## MVP Gate

MVP is declared complete only when **DNP3-056** passes.

```
TOTAL TASKS: 100
MASTER TASKS: 72
OUTSTATION TASKS: 28
MVP COMPLETE AT: DNP3-056
COMPLETED: DNP3-001 through DNP3-012
NEXT TASK: DNP3-013 — Basic IIN reaction (DeviceRestart / NeedTime)
```

## Test Status

- `go test ./...` — all 21 packages green (including integration).
- `go test -race ./internal/al/... ./internal/master/... ./internal/outstation/... ./pkg/dnp3/master/... ./test/integration/...` — green.
- Pre-existing `go vet` "unreachable code" notes in `internal/outstation/outstation.go:827` and `pkg/dnp3/master/client.go:322` are NOT introduced by these tasks (confirmed on clean HEAD) and are out of scope.
- Commit hash: f552577 (DNP3-010/011/012 checkpoint)
