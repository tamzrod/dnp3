# DNP3 Implementation Handoff

**Last updated:** 2026-08-12  
**Roadmap:** `active_work/DNP3_MASTER_ROADMAP.md`  
**Profile:** `active_work/supported-profile.md`

## Status

- Planning complete.
- DNP3-001 through DNP3-045 complete. Implementation underway.
- Last checkpoint: DNP3-043/044/045 (error taxonomy + optional logging hooks +
  full-MVP public loopback against the simulator). All green incl. `-race`.
- Previous checkpoint: DNP3-040/041/042 (per-outstation outstanding-request
  tracking + timeout/retry config validation + optional idle-timeout keep-alive
  monitor). All green incl. `-race`.
- Prior checkpoint: DNP3-037/038/039 (commit 4c4ac0f, pushed to origin/main).
- Go 1.22 toolchain: reinstalled at `~/go-install/go/bin/go` (add to PATH:
  `export PATH=$HOME/go-install/go/bin:$PATH`).

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

### DNP3-013 — Basic IIN reaction (DeviceRestart / NeedTime)
- Commit message: `feat(master): react to critical IIN bits`
- `internal/master/master.go`: added `reactToIIN`, called from
  `processResponse` after the IIN update. On DeviceRestart (IIN1.7) the
  outstation is marked `NeedsIntegrity` + State="Restart" (clear local
  state / re-integrity); on NeedTime (IIN1.4) it is marked `NeedsTimeSync`
  (stub time-sync). Optional callbacks `SetDeviceRestartHandler` /
  `SetNeedTimeSyncHandler` provide the log/stub hook. Full integrity
  polling and time objects are later roadmap items.
- New Outstation accessors: `NeedsIntegrity()/ClearNeedsIntegrity()`,
  `NeedsTimeSync()/ClearNeedsTimeSync()` (thread-safe).
- Tests: `TestIINReactionDeviceRestart`, `TestIINReactionNeedTime`,
  `TestIINReactionBothBits`, `TestIINReactionClean` (injected IIN bits).
- Acceptance: documented reaction occurs; reaction tests green.

### DNP3-014 — Class-0 integrity request construction
- Commit message: `feat(master): canonical Class-0 integrity request`
- `internal/master/master.go`: `buildPollRequest` now constructs the
  Class-0 / integrity poll with the canonical all-objects qualifier
  (0x06) on Group 60 Variation 1 (was 0x07 count=0, which is semantically
  "read zero objects"). Single, deterministic request form for "read all
  Class-0 static data". The 4-byte header stride is preserved, so the
  outstation's `buildReadResponse` (which keys off group 60) is unaffected.
- Golden fixture `active_work/testdata/class0-integrity-request.hex`
  (G60V1 all-objects: `3C 01 06 00`).
- New helper `loadGoldenHex` (resolves `active_work/testdata` via
  runtime.Caller, strips `#` comments/whitespace, hex-decodes).
- Tests: updated `TestBuildPollRequest` (0x06 form), new
  `TestBuildPollRequestIntegrityGolden` (golden fixture match +
  PollClass0 == PollIntegrity form).
- Acceptance: integrity request is deterministic; request fixture test green.

### DNP3-015 — Multi-fragment Class-0 response handling
- Commit message: `fix(master): multi-fragment Class-0 reassembly`
- Verified the receive path (`processReceivedBytes`) already reassembles
  multi-fragment transport-layer responses into a complete APDU before
  parse: it loops over DLL frames, pushes each TL fragment to the
  `tl.Reassembler`, and only returns the message when FIN is received.
  `al.DecodeResponse` then parses the complete reassembled APDU.
- Golden fixture `active_work/testdata/class0-multifragment-apdu.hex`
  (single app message split across 2 transport fragments: G1V1 count=2
  in frag 1, G30V1 count=1 in frag 2; value=42).
- Test `TestMultiFragmentClass0Reassembly`: feeds a 2-fragment Class-0
  response through `processReceivedBytes`, asserts the reassembled APDU
  matches the golden, decodes it, and confirms both G1V1 (2 points) and
  G30V1 (1 point) headers are present (fragment 2 data not lost).
- Acceptance: all points present; multi-fragment test green.

### DNP3-016 — Binary Input G1V1 final correctness
- Commit message: `fix(objects): finalize Binary Input Variation 1`
- `pkg/dnp3/master/client.go`: G1V1 ("Binary Input - Packed Format", IEEE 1815)
  is now parsed as packed binary state bits for every qualifier valid for the
  packed format — count8 (0x07), count16 (0x27), all-objects (0x06), and
  range16 (0x28). Points are packed 8 per byte (bit 0 of byte 0 = point 0),
  carry no per-point quality byte, and are assigned ONLINE quality. The index
  base is 0 for count/all-objects qualifiers and `Start` for range16. The
  QualIndex8 (0x00) qualifier is not valid for packed format and still falls
  through to the legacy per-point path.
- New helper `packedBinaryRange` (delegates to `sequentialRange`) returns the
  index base + point count from the header qualifier.
- Tests: `TestG1V1PackedBoundary` (1/8/9/16 points, byte-boundary crossing,
  range16 base 5 and base 0), `TestG1V1PackedBitOrder` (LSB-first bit
  ordering: 0x01→point0, 0x80→point7), `TestG1V1PackedMultipleHeaders` (two
  packed headers, independent index bases). Existing `TestParseClass0PackedBinaryInputVector`
  remains the golden pass.
- Acceptance: golden pass; G1 vector suite green; loopback/integration green.

### DNP3-017 — Analog Input G30V1 final correctness
- Commit message: `fix(objects): finalize Analog Input Variation 1`
- `pkg/dnp3/master/client.go`: G30V1 (signed 32-bit value + 1-octet flags,
  5 octets/point, sequential) is now parsed for count8/count16/all-objects
  (index base 0) and range16 (index base = Start) via the new
  `sequentialRange` helper. `ReadAnalogInputs` sends a range16 (0x28) request;
  an external conformant outstation honoring that range returns sequential
  G30V1 points from Start, which the parser now handles (previously only 0x07
  was handled and range16 fell into the generic indexed loop, misreading a
  2-byte index per point).
- Tests (new file `pkg/dnp3/master/analog_input_test.go`):
  `TestG30V1Count8Boundary` (1000 / -1 signed), `TestG30V1SignedRange`
  (zero, MaxInt32, MinInt32, -1, -1000), `TestG30V1Range16Boundary`
  (start=5 stop=6), `TestG30V1LSBByteOrder`, `TestG30V1QualityByte`
  (ONLINE/RESTART/COMM_LOST).
- Acceptance: signed LSB decode + range16 + quality byte locked; suite green.

### DNP3-018 — Counter G20V1 final correctness
- Commit message: `fix(objects): finalize Counter Variation 1`
- `pkg/dnp3/master/client.go`: G20V1 (unsigned 32-bit counter + 1-octet
  flags, 5 octets/point, sequential) is now parsed for count8/count16/
  all-objects (index base 0) and range16 (index base = Start) via
  `sequentialRange`. Same gap closed as G30V1: `ReadCounters` sends range16
  and a conformant range response now decodes correctly.
- Tests (new file `pkg/dnp3/master/counter_test.go`):
  `TestG20V1Count8Boundary` (1000 / MaxUint32), `TestG20V1UnsignedRange`
  (zero, 1000, MaxUint32, high-bit-set unsigned), `TestG20V1Range16Boundary`
  (start=5 stop=6), `TestG20V1LSBByteOrder`, `TestG20V1QualityByte`.
- Acceptance: unsigned LSB decode + range16 + quality byte locked; suite green.

### DNP3-019 — CROB G12V1 request final correctness
- Commit message: `fix(objects): finalize CROB request encoding`
- `internal/master/master.go`: added named `CROBCode*` constants for the
  request encode/outstation decode (1..8 enum). `buildControlRequest` bool→
  code mapping now uses the named constants. The G12V1 wire LAYOUT is locked:
  header `0C 01 00 01`, 2-octet LSB index, then the 11-byte CROB value
  (code, count, onTime LSB, offTime LSB, status). Both `buildCROBRequest`
  (WriteBinaryOutput) and `buildControlRequest` (Operate) verified.
- Tests: extended `internal/master/control_vector_test.go` with
  `TestBuildCROBRequestLayout` (full 17-octet layout),
  `TestBuildCROBRequestIndexLSBHighByte` (index 0xABCD → CD AB),
  `TestBuildCROBRequestTimeLSBBoundary` (max-uint32 on/off times),
  `TestBuildControlRequestCROBBoolMapping` (true→LATCH_ON, false→LATCH_OFF,
  uint8 passthrough). Existing golden `direct-control-crob-vector.hex` pass.
- **Discovery (cross-layer, NOT fixed here):** the repository's CROB control-
  code VALUES use a 1..8 enum, NOT the IEEE 1815 G12V1 control-code bit field
  (0x01 NUL, 0x02 Pulse On, 0x04 Pulse Off, 0x08 Latch On, 0x10 Latch Off,
  0x80 Queue). This is consistent internally (master encode + outstation decode
  both use 1..8), so loopback works, but a real external outstation would NOT
  interpret code=7 as Latch On. The encode LAYOUT (single octet in position) is
  correct; reconciling the code VALUES with the IEEE 1815 bit field is a
  coordinated master+outstation+public-handler correction that should be a
  dedicated task before external interop verification (VEC-01+). Flagged for
  Grok/Rod.
- Acceptance: request bytes match golden; CROB request test green.

### DNP3-020 — CROB status parsing
- Commit message: `fix(master): parse CROB command status`
- `internal/master/master.go`: added `CommandStatus` type + constants
  (Success=0, Timeout=1, NoSelect=2, BadFormat=3, NotSupported=4,
  AlreadyActive=5, Blocked=6, Local=7, TooMany=8, NotAuthorized=9,
  Autonomous=10, Unknown=0xFF). New `parseCommandStatus(data)` scans the
  response object data for a G12V1 object and returns the per-point command
  status byte (CTRL-01); supports the index-only (0x00), count8 (0x07), and
  count16 (0x27) qualifiers. New `OperateWithStatus(...)` sends the request
  via `sendWithRetryAndGetResponse`, decodes the response, and returns the
  parsed status. A response with no parseable G12V1 status yields
  `CommandStatusUnknown` — never success.
- Tests (new file `internal/master/command_status_test.go`):
  `TestParseCommandStatusVectors` (all 11 status codes), missing-object and
  truncated-object → Unknown, `TestOperateWithStatusSuccessRejected`
  (success + 3 rejection codes via canned status-echo transport),
  `TestOperateWithStatusMissingObjectNotSuccess` (empty response → Unknown).
- Acceptance: failed point ≠ CommandStatusSuccess; status tests green.

### DNP3-021 — Public Operate response carries status
- Commit message: `feat(api): expose control status on OperateResponse`
- `pkg/dnp3/master/client.go`: public `client.Operate` now calls
  `internal.OperateWithStatus` and translates the result to
  `types.ControlStatus` via new `mapCommandStatus`. `OperateResponse.Status`
  carries the real per-point command status; a failed/unknown status is never
  reported as `ControlSuccess` (unknown → `ControlTimeout`).
- Tests (new file `pkg/dnp3/master/operate_status_test.go`):
  `TestPublicOperateSurfacesStatus` (success + 4 rejection codes through the
  public `Operate` path), `TestPublicOperateMissingStatusNotSuccess` (empty
  response → not ControlSuccess).
- Acceptance: caller sees real status; public control test green.

### DNP3-022 — Context cancellation on Connect
- Commit message: `fix(api): context cancellation on Connect`
- `pkg/dnp3/dnp3.go`: added `ErrContextCanceled` sentinel.
- `pkg/dnp3/master/client.go`: rewrote `Connect` to honor `ctx.Done()`. The
  blocking connect steps run in a goroutine; the main path selects on
  `ctx.Done()` vs completion. On cancellation, returns promptly with
  `ErrContextCanceled` and tears down any partially-established connection
  (`cleanupConnect` closes internal master + transport) so no live connection
  remains; state reset to Disconnected. `connectBlocking` also checks ctx
  between blocking steps so a cancellation landing between the dial and the
  handshake is observed without waiting for the next timeout.
- Tests (new file `pkg/dnp3/master/connect_cancel_test.go`):
  `TestConnectAlreadyCancelledContext` (pre-cancelled ctx → prompt,
  Disconnected), `TestConnectCancelledMidDial` (cancel while transport.Connect
  blocks → prompt, Disconnected, transport Close()d).
- **Note:** test mocks (`blockingTransport`, `slowCloseTransport`) make `Close`
  idempotent because the cancel-teardown path may call `Close` more than once
  (abort + connectBlocking error path); real `TCPTransport.Close`/`TLSTransport`
  .Close are already idempotent.
- Acceptance: prompt return, no live connection; tests green.

### DNP3-023 — Context cancellation on Read
- Commit message: `fix(api): context cancellation on Read`
- `pkg/dnp3/master/client.go`: `Read` now runs
  `SendRequestWithRetryAndGetResponse` in a goroutine and selects on
  `ctx.Done()`. On cancellation, returns promptly with `ErrContextCanceled`
  and a nil response — the in-flight result is discarded, so no partial points
  leak to the caller.
- Tests (new file `pkg/dnp3/master/read_cancel_test.go`):
  `TestReadAlreadyCancelledContext` (pre-cancelled → prompt, nil response),
  `TestReadCancelledMidRequest` (cancel while Receive blocks → prompt, nil
  response, no partial points).
- Acceptance: error + no partial points; tests green.

### DNP3-024 — Context cancellation on Operate / Disconnect
- Commit message: `fix(api): context cancellation on Operate and Disconnect`
- `pkg/dnp3/master/client.go`: `Operate` runs `OperateWithStatus` in a
  goroutine and selects on `ctx.Done()` (prompt `ErrContextCanceled`, nil
  response, in-flight result discarded). `Disconnect` runs
  internal.Disconnect + transport.Close in a goroutine and selects on
  `ctx.Done()`; on cancellation returns `ErrContextCanceled` but still resets
  state to Disconnected so the client is not left stuck (background teardown
  completes best-effort).
- Tests (new file `pkg/dnp3/master/operate_disconnect_cancel_test.go`):
  `TestOperateAlreadyCancelledContext`, `TestOperateCancelledMidRequest`,
  `TestDisconnectAlreadyCancelledContext`, `TestDisconnectCancelledMidTeardown`.
- Acceptance: all public entry points cancel cleanly; tests green.

### DNP3-025 — Race safety for sequence & reassembly
- Commit message: `fix(master): serialize request and reassembly state`
- `internal/master/master.go`: added `reqMu sync.Mutex` to the `Master`
  struct; `sendWithRetry` and `sendWithRetryAndGetResponse` acquire it for the
  whole request path (send → receive → reassembly) so concurrent requests on the
  same master do not race on the shared `reassembler` or interleave fragments on
  a single link. A DNP3 link is request/response, so serializing is both safe
  and correct. (`sequence` was already guarded by `m.mu`.)
- `pkg/dnp3/master/client.go`: public `c.sequence` read+advance in `Read` now
  guarded by `c.mu`.
- Tests (new file `pkg/dnp3/master/race_test.go`):
  `TestConcurrentReadsRaceFree` (50 concurrent Reads), `TestConcurrentOperateRaceFree`
  (50 concurrent Operates). New `pubReadEchoTransport` (mutex-guarded SEQ echo)
  and `buildPubReadResponse` helper.
- Acceptance: race detector clean; full `-race` suite green.

### DNP3-026 — Reject invalid CRC on receive
- Commit message: `test(master): reject invalid CRC frames`
- The DLL `frame.Decode` already validates header + per-block CRCs and returns
  an error; `processReceivedBytes` surfaces it (no points). DNP3-026 locks this
  with negative tests rather than new code.
- Tests (new file `pkg/dnp3/master/crc_reject_test.go`):
  `TestReadRejectsInvalidCRC`, `TestReadRejectsInvalidCRCIsNoPartial` (corrupted
  header CRC → error, nil response, no partial points),
  `TestALRejectsInvalidCRC` (header CRC + data-block CRC rejection at the frame
  layer). New `badCRCTransport` and `buildPubReadResponse` (shared).
- Acceptance: error, empty response; tests green.

### DNP3-027 — Reject truncated / oversize frames
- Commit message: `fix(master): reject truncated and oversize frames`
- `internal/dll/frame/frame.go`: added an explicit oversize guard in `Decode`
  (`dataLen > MaxDataSize` → error). Truncated (below `MinFrameSize`, fewer
  bytes than claimed length) and claimed-length mismatch were already handled;
  this adds the oversize defense-in-depth and locks all three with tests.
- Tests (new file `internal/dll/frame/truncation_test.go`):
  `TestDecodeRejectsTruncated` (below-min, truncated payload, no-panic on
  garbage), `TestDecodeRejectsClaimedLengthMismatch`,
  `TestDecodeRejectsOversize` (max 250-byte frame decodes; MaxDataSize == 250).
- Acceptance: error, no panic; tests green.

### DNP3-028 — Reject unsupported qualifiers
- Commit message: `fix(al): reject unsupported qualifiers`
- `internal/al/object_header.go`: added `ErrUnsupportedQualifier` sentinel;
  `Encode` and `DecodeObjectHeader` now wrap it (`%w`) so callers can
  `errors.Is`. The v0 allow-list is 0x06 (all-objects), 0x07 (8-bit count),
  0x00 (8-bit index), 0x28 (16-bit range), 0x27 (16-bit count); all other
  qualifier bytes return a clear error and no header.
- Tests (extended `internal/al/object_header_test.go`):
  `TestDecodeObjectHeaderUnsupportedQualifier` (table of 0x01,0x02,0x05,0x08,
  0x17,0x25,0x26,0x29,0xFE,0xFF → ErrUnsupportedQualifier, consumed=0, zero
  header), `TestEncodeObjectHeaderUnsupportedQualifier`.
- Acceptance: clear error, no data; tests green.

### DNP3-029 — Reject unsupported groups/variations on read
- Commit message: `fix(api): reject unsupported object groups for MVP`
- `pkg/dnp3/dnp3.go`: added `ErrUnsupportedGroup` sentinel.
- `pkg/dnp3/master/client.go`: `Read` validates each requested group against the
  v0 read profile (G1 Binary Input, G20 Counter, G30 Analog Input; variation 0
  "any" or 1) BEFORE any wire traffic. New `isSupportedReadGroup` helper.
- Tests (new file `pkg/dnp3/master/group_reject_test.go`):
  `TestReadRejectsUnsupportedGroups` (G2/G10/G13/G21/G32/G40/G60, bad
  variations, G0/G255 → ErrUnsupportedGroup),
  `TestReadAcceptsSupportedGroups` (G1/G20/G30 × v0/v1).
- Acceptance: explicit error before any request; tests green.

### DNP3-030 — Reject SBO / unsolicited / TLS — complete profile gate
- Commit message: `fix(api): complete supported-profile rejection matrix`
- `pkg/dnp3/dnp3.go`: added `ErrUnsupportedOption` sentinel.
- `pkg/dnp3/master/client.go`: `Operate` now rejects `SelectThenOperate` and
  `DirectOperateNoResponse` (only `DirectOperate` supported) with
  `ErrUnsupportedOption` before any wire traffic; `selectThenOperate` is
  hard-set to false for the v0 path. `EnableUnsolicited`/`DisableUnsolicited`
  return `ErrUnsupportedOption` (wrapped). TLS `NewClient` returns
  `ErrUnsupportedOption` (was a plain string error).
- Tests: updated `safety_test.go` (`TestNewClientRejectsUnsupportedTLS` now
  checks `errors.Is`); new file `pkg/dnp3/master/profile_reject_test.go`
  (`TestOperateRejectsSBO`, `TestOperateRejectsDirectOperateNoResponse`,
  `TestOperateAcceptsDirectOperate`, `TestUnsolicitedRejected`).
- Acceptance: no silent fallback; matrix tests green.

### DNP3-031 — Transport disconnect detection
- Commit message: `feat(master): detect transport disconnect and transition state`
- `pkg/transport/tcp.go`: added exported `IsDisconnect(err)` helper (matches
  `ErrClosed`, `io.EOF`/`io.ErrUnexpectedEOF`, `net.ErrClosed`).
- `internal/master/master.go`: added `ErrTransportDisconnected` sentinel and
  exported `IsDisconnectError`. New `markDisconnected(err)` helper wraps a
  transport close as `ErrTransportDisconnected` and sets `StateError`.
  `sendWithRetry` and `sendWithRetryAndGetResponse` now break the retry loop on
  a disconnect (no point retrying a dead link) at the Send, confirm, and
  receive sites; `waitForConfirmation` returns `ErrTransportDisconnected`
  (not `ErrConfirmTimeout`) on a peer close.
- `pkg/dnp3/master/client.go`: `Read` and `Operate` set the public client state
  to `StateDisconnected` when the result error is a transport disconnect, so a
  subsequent call fails fast with `ErrNotConnected`.
- Tests: new `pkg/dnp3/master/disconnect_test.go` —
  `TestReadDetectsTransportDisconnect`, `TestOperateDetectsTransportDisconnect`,
  `TestReadAfterDisconnectReturnsNotConnected` (peer-close transport returns
  `io.EOF`).
- Acceptance: peer close surfaces as disconnect; state transitions to
  Disconnected; no retry storm on a dead link.

### DNP3-032 — Reconnect + re-handshake
- Commit message: `feat(master): reconnect and re-handshake`
- `internal/master/master.go`: `Connect` now calls `resetForReconnect()` before
  `performLinkHandshake()`. New `resetForReconnect()` clears TL state
  (`reassembler.Reset()`, `fragmenter.Reset()`) so a reconnect after a
  mid-session drop re-handshakes (Reset Link Stations + Request Link Status)
  from a clean slate.
- The public `client.Connect` gate already permits reconnect after
  DNP3-031 sets state Disconnected; `AddOutstation` overwrites cleanly.
- Tests: new `internal/master/reconnect_test.go` —
  `TestReconnectReHandshakes` (drop mid-session, recover, re-handshake →
  Active), `TestResetForReconnectClearsTLState`.
- Acceptance: subsequent Read/Operate succeeds after a reconnect.

### DNP3-033 — Clear reassembly on reconnect
- Commit message: `fix(tl): reset reassembler on session restart`
- Implementation shared with DNP3-032 (`resetForReconnect` invoked by `Connect`).
- Tests: new `internal/master/reassembly_reset_test.go` —
  `TestReassemblerNoCrossSessionPollution` (partial FIR-only fragment from a
  dropped session does not leak into the next session's reassembled message),
  `TestResetForReconnectResetsFragmenter` (outgoing fragment Seq restarts at 0).
- Acceptance: no cross-session reassembly pollution.

### DNP3-034 — Retry policy refinement (distinct timeout/NACK/CRC handling)
- Commit message: `fix(master): refine retry policy by error class (DNP3-034)`
- `internal/master/master.go`:
  - Added `ErrLinkNACK` and `ErrCRCError` error sentinels.
  - Added `RetryClass` enum (`ClassTimeout`, `ClassNACK`, `ClassCRC`,
    `ClassDisconnect`, `ClassOther`) and `RetryPolicy` struct with per-class
    retry limits + delays (`DefaultRetryPolicy`).
  - Added `classifyRetryError(err) RetryClass` and `isCRCError(err)` helpers.
  - Added `retryPolicy` field on `Master` (init via `DefaultRetryPolicy` in
    `NewMaster`); `RetryPolicy()` / `SetRetryPolicy()` accessors.
  - `sendWithRetry` and `sendWithRetryAndGetResponse` classify errors via
    `classifyRetryError` and consult `RetryPolicy` for retry decision + delay.
  - `processReceivedBytes`: secondary NACK → `ErrLinkNACK`; CRC decode errors
    wrapped with `ErrCRCError`.
  - `processResponse`: inner error wrapped with `%w` (not `%v`) so
    `errors.Is` propagates to the NACK/CRC sentinels.
  - `retryAgain`: terminal error wrapped with `%w: %w` (ErrMaxRetries + inner)
    so the error class remains reachable via `errors.Is` after the budget is
    exhausted.
  - Added `strings` to imports.
- Tests: new `internal/master/retry_policy_test.go` — per-class retry tests
  (`TestRetryClassifiesTimeout/NACK/CRC`), recovery-after-failure tests
  (`TestRetryNACKRecoversOnSuccess`, `TestRetryCRCRecoversOnSuccess`),
  `TestRetryDisconnectNotRetried`, `TestRetryPerClassCounts`,
  `TestRetryDelayApplied`, `TestClassifyRetryError` (table), and
  `TestProcessReceivedBytesSurfacesNACK/CRC`.
- Acceptance: distinct retry handling per error class; tests green incl `-race`.

### DNP3-035 — Public Read returns only supported point types
- Commit message: `fix(api): restrict ReadResponse to MVP types`
- `pkg/dnp3/master/client.go`: `Read` now populates only the MVP-supported
  Class-0 slices (BinaryInputs G1, AnalogInputs G30, Counters G20).
  `BinaryOutputs`/`AnalogOutputs`/`FrozenCounters` are retained for forward
  compatibility but left nil by the v0 Read path (documented on the struct).
  The legacy `parseBinaryOutputs`/`parseAnalogOutputs` helpers are kept
  (tested) for deferred profiles but no longer wired into Read.
- Tests: new `pkg/dnp3/master/read_response_shape_test.go` —
  `TestReadResponsePopulatesMVPTypes` (per-group), `TestReadResponseDoesNotSurfaceUnsupportedTypes`,
  `TestReadResponseEmptyHasNoUnsupportedTypes`.
- Acceptance: no unsupported types surfaced; shape tests green.
- Discovery (not fixed here, out of scope): the legacy `skipGroupData` helper
  uses index-based byte counts that do not match the packed (G1V1) or
  sequential (G20V1/G30V1) response formats, so a single response carrying
  multiple MVP headers can mis-skip and lose points. The v0 Read path is
  exercised one group per response, so this does not affect production; the
  DNP3-035/036 tests use single-group fixtures to avoid it. Tracked for a later
  parser-robustness task.

### DNP3-036 — Deterministic outstation simulator (MVP profile)
- Commit message: `test: deterministic MVP outstation simulator`
- New `internal/testutils/simulator.go`: `MVPOutstationSimulator` implements
  the public `transport.Handler` interface. It answers the link handshake
  (Reset Link Stations → ACK; Request Link Status → Link Status), Class-0
  Read requests (G1/G20/G30 golden data via count8 qualifiers), and G12V1
  DirectOperate (CROB) with a configurable per-point command status. Responses
  echo the request's application SEQ and carry a configurable IIN. Concurrency-
  safe; records sent frames for assertions.
- New public test hook `pkg/dnp3/master.NewClientWithTransport(config, transport.Handler)`
  for driving the full public Connect → Read → Operate flow against a custom
  in-memory transport (no network I/O).
- Tests: new `internal/testutils/simulator_test.go` (handshake, golden Read
  encoding, Operate status echo, sent-frame recording) and new
  `test/integration/simulator_loopback_test.go`
  (`TestPublicLoopbackAgainstSimulator` full public flow against the simulator
  only, `TestPublicLoopbackSimulatorSurfacesCommandStatus`,
  `TestPublicLoopbackSimulatorStateTransitions`).
- Acceptance: public loopback green against simulator only (no real outstation
  process / no TCP).

### DNP3-037 — Integrity poll convenience method
- Commit message: `feat(api): add IntegrityPoll convenience method`
- `pkg/dnp3/master/client.go`: added `IntegrityPoll(ctx context.Context) (*ReadResponse, error)`
  to the public `Client` interface and the `client` struct. Issues a separate
  per-group Read for each MVP-supported Class-0 group (G1.1, G30.1, G20.1) and
  merges the results, instead of a single multi-group read. This sidesteps the
  known `skipGroupData` parser limitation (DNP3-035 discovery) where a single
  response carrying multiple MVP headers can lose points.
- Tests: new `pkg/dnp3/master/integrity_poll_test.go` —
  `TestIntegrityPollMatchesExplicitReads` (IntegrityPoll returns the same data
  as explicit per-group Reads against the simulator),
  `TestIntegrityPollNotConnected` (returns ErrNotConnected before Connect).
- Acceptance: IntegrityPoll returns merged MVP Class-0 data; tests green.

### DNP3-038 — Document supported-profile in code comments
- Commit message: `docs: annotate supported-profile dispositions`
- Annotated every public surface with Target/Reject/Defer dispositions
  (referencing `active_work/supported-profile.md`):
  - `pkg/dnp3/master/client.go`: `Config` fields and all `With*` option
    functions; `Client` interface methods (Read, IntegrityPoll, Operate,
    EnableUnsolicited, DisableUnsolicited, SetUnsolicitedHandler, Close).
  - `pkg/dnp3/types/commands.go`: `CommandType` constants, `NewBinaryControl`
    (Target, G12V1 only), `NewAnalogControl` (Reject), `NewPulseControl`
    (Reject), `NewReadRequest`, and the `ReadAllStatic`/`ReadAllEvents`/
    `ReadBinaryInputs`/`ReadAnalogInputs`/`ReadCounters` convenience vars.
  - `pkg/dnp3/outstation/server.go`: `With*` options (WithAddress, WithMasterAddress,
    WithTransport, WithTLS, WithMaxFragmentSize, WithUnsolicitedMode) and
    `Start`/`Stop`/`Close`/`SetDataHandler`/`SetCommandHandler`/
    `SetUnsolicitedHandler` methods.
- Doc-only change; no behavior. Build + full suite green.
- Acceptance: supported-profile dispositions documented in code.

### DNP3-039 — Master state machine formalization
- Commit message: `refactor(master): formalize state machine transitions`
- `internal/master/master.go`:
  - Added a state-machine doc comment (legal-transition diagram) on `State`.
  - Added `legalStateTransitions` table encoding every allowed (from -> to)
    transition.
  - Added `transitionTo(newState)` which validates against the table and returns
    the new `ErrIllegalStateTransition` sentinel on an illegal move; a
    self-transition is an idempotent no-op.
  - Added `isOperational()` (only StateInitialized/StateActive qualify) and
    `isLinkUp()` (Connected/Initialized/Active) guards.
  - Replaced the legacy `State() < StateInitialized` / `State() < StateConnected`
    numeric-ordinal guards on all operation methods (Poll, Operate,
    OperateWithStatus, TimeSync, WriteBinary/AnalogOutput, Read* family,
    Enable/DisableUnsolicited) with `isOperational()` / `isLinkUp()`.
  - **Key bug fix:** `StateError` has iota ordinal 5 (> StateActive 4), so the
    legacy `< StateInitialized` guard silently let operations through on a dead
    link. `isOperational()` explicitly excludes StateError, so operations on a
    dead link now correctly return ErrNotConnected.
  - `Connect()` now rejects a Connect from any state other than Disconnected or
    Error (e.g., a concurrent Connect while Connecting) with
    ErrIllegalStateTransition, instead of silently overwriting state; a failed
    handshake lands in StateError (not Disconnected) via the table.
  - `Initialize()` uses `transitionTo(StateInitialized)` (legal only from
    Connected/Active).
  - `Disconnect()` remains the unconditional teardown path (forces Disconnected
    from any state).
  - `markDisconnected()` uses `transitionTo(StateError)`.
  - `SetState()` retained as a documented escape hatch for tests/recovery.
- Tests: new `internal/master/state_machine_test.go` —
  `TestStateTransitionTableLegal`, `TestStateTransitionTableIllegal`,
  `TestStateTransitionSelfIsNoOp`, `TestOperationsRejectedInErrorState`
  (the central regression), `TestOperationsRejectedWhenDisconnected`,
  `TestOperationsAcceptedInOperationalStates`, `TestConnectRejectsFromConnecting`,
  `TestConnectReconnectFromError`. All green incl. `-race`.
- Acceptance: no silent illegal transitions; state transition table tests green.

### DNP3-040 — Request outstanding tracking
- Commit message: `feat(master): track outstanding request per outstation`
- `internal/master/master.go`:
  - Added `ErrRequestOutstanding` sentinel.
  - Added per-outstation outstanding-request tracking: `outstanding`
    `map[uint16]struct{}` + `outstandingMu`, with `beginRequest(id)` / `endRequest(id)`
    / `HasOutstandingRequest(id)` helpers. `beginRequest` returns a wrapped
    `ErrRequestOutstanding` if a request is already in flight for that same
    outstation; `endRequest` is idempotent.
  - `sendWithRetry` and `sendWithRetryAndGetResponse` call `beginRequest` BEFORE
    acquiring `reqMu` (deferred `endRequest`), so a concurrent same-outstation
    request is rejected immediately instead of blocking behind the global request
    lock indefinitely. Different outstations still queue via `reqMu`.
- `pkg/dnp3/master/race_test.go`: `TestConcurrentReadsRaceFree` /
  `TestConcurrentOperateRaceFree` updated to tolerate `ErrRequestOutstanding` as
  the defined concurrent same-outstation outcome (the race-free assertion — race
  detector quiet — still holds); only non-`ErrRequestOutstanding` errors fail.
- Tests: new `internal/master/outstanding_request_test.go` —
  `TestConcurrentSameOutstationRejected` (concurrent same-outstation request
  rejected with `ErrRequestOutstanding` without blocking; marker cleared after
  completion; a subsequent sequential request succeeds),
  `TestDistinctOutstationsNotRejectedByOutstandingTracking` (per-outstation
  independence), `TestBeginRequestIdempotentGuard` (idempotent endRequest +
  repeated begin rejected). All green incl. `-race`.
- Acceptance: defined concurrent behavior (reject, not silent queue); no
  corruption; concurrency tests green.

### DNP3-041 — Timeout configuration validation
- Commit message: `fix(api): validate timeout and retry config`
- The public `Config.Validate()` (already present) rejects non-positive
  `Timeout`, negative `RetryCount`, negative `RetryDelay`, and TLS-without-config,
  surfacing a `*dnp3.ConfigurationError` with the offending field. `NewClient`
  and `NewClientWithTransport` already call `Validate()` and return the error
  (and a nil Client) before constructing a master.
- Tests: expanded `pkg/dnp3/master/client_test.go` `TestConfigValidate` table
  with zero-timeout, negative-RetryDelay, and `WithTimeout(0)` / `WithRetry(-1, …)`
  / `WithRetry(3, -1)` cases; each rejection now asserts the error is a
  `*dnp3.ConfigurationError` via `errors.As`. Added
  `TestNewClientRejectsInvalidConfig` (DNP3-041 acceptance) verifying `NewClient`
  returns a non-nil error AND a nil Client (no leaked resources) and a
  `*ConfigurationError` for each invalid config.
- Acceptance: invalid config fails NewClient; config tests green.

### DNP3-042 — Keep-alive / idle detection (minimal)
- Commit message: `feat(master): optional idle timeout`
- `internal/master/master.go`:
  - Added `IdleTimeout int` (ms) to internal `master.Config`; `NewMaster` derives
    `idleTimeout time.Duration`.
  - Added `ErrIdleTimeout` sentinel.
  - Added `lastActivity time.Time` (under `mu`), `idleStop chan struct{}`,
    `idleWG sync.WaitGroup` fields; `recordActivity()` / `lastActivityAt()`
    helpers; `startIdleMonitor()` / `stopIdleMonitor()` lifecycle; and
    `idleMonitorLoop(stop)` which ticks at `idleTimeout/2` and, when no activity
    has been observed for the full idle timeout, transitions the session to
    Disconnected (bypassing the transition table, like Disconnect) and reports
    `ErrIdleTimeout` to the error handler.
  - `Connect()` calls `startIdleMonitor()` after a successful handshake;
    `Disconnect()` and `markDisconnected()` call `stopIdleMonitor()` so the
    goroutine does not outlive the session. Re-Connecting stops any prior monitor
    first (no leaked goroutines / double-close).
  - `sendWithRetry` / `sendWithRetryAndGetResponse` call `recordActivity()` on
    every successful send (post-`advanceSequence`) and every successful response.
- `pkg/dnp3/master/client.go`: added public `Config.IdleTimeout` field and
  `WithIdleTimeout(d)` option; wired into the internal `master.Config.IdleTimeout`
  in both `NewClient` and `NewClientWithTransport`.
- Tests: new `internal/master/idle_monitor_test.go` —
  `TestIdleMonitorClosesSessionToDisconnected` (idle → Disconnected),
  `TestIdleMonitorActivityPreventsClose` (periodic activity keeps state Active),
  `TestIdleMonitorDisabledByDefault` (IdleTimeout==0 → no monitor, no close),
  `TestIdleMonitorStopIsIdempotent`, `TestIdleMonitorRestartStopsPrevious`
  (re-Connect stops the prior monitor). All green incl. `-race`.
- Acceptance: state becomes Disconnected on idle close; idle tests green.

### DNP3-043 — Error type taxonomy
- Commit message: `feat(master): public error taxonomy and ClassifyError`
- `pkg/dnp3/dnp3.go`: added the public `ErrorCode` taxonomy
  (`ErrorCodeUnknown`, `ErrorCodeTimeout`, `ErrorCodeCRC`, `ErrorCodeSequence`,
  `ErrorCodeDisconnect`, `ErrorCodeBusy`, `ErrorCodeUnsupported`,
  `ErrorCodeCanceled`, `ErrorCodeConfiguration`, `ErrorCodeInvalid`) with
  `String()`, and `ClassifyError(err) ErrorCode` which walks the wrapped error
  chain via `errors.Is` against the public sentinels (canceled → configuration
  → unsupported → CRC → sequence → timeout → busy → disconnect → invalid →
  unknown). Added public sentinels `ErrCRC` and `ErrRequestOutstanding`
  (previously only internal), plus `ErrUnsupportedOption`.
- `pkg/dnp3/master/client.go`: added `wrapInternalError(prefix, err)` helper
  that wraps an internal error with the matching public sentinel while
  preserving the internal sentinel chain via `%w` (`%v` for the prefix text).
  Wired into the `Read` and `Operate` public boundaries so surfaced failures
  carry a classifiable public sentinel (CRC/sequence/busy/timeout/disconnect/
  unsupported).
- `internal/master/master.go`: `waitForResponse` now tags a non-disconnect
  receive error with `ErrTimeout` (disconnect errors pass through unchanged so
  `markDisconnected` wraps them with `ErrTransportDisconnected`), making a
  no-response-within-timeout classifiable as a timeout.
- `internal/testutils/mock_transport.go`: `TransportError` now carries an
  optional underlying cause exposed via `Unwrap`; `ErrTransportClosed` chains
  `io.EOF` so the master's `IsDisconnectError` recognizes a closed simulated
  peer exactly like a real TCP peer close (which surfaces `io.EOF`).
- Tests: `pkg/dnp3/error_taxonomy_test.go` (sentinel mapping, chain unwrap,
  precedence, ErrorCode String); `pkg/dnp3/master/error_classification_test.go`
  (boundary classification via transport mocks: badCRC, timeout, seqMismatch,
  peerClose). All green incl. `-race`.
- Acceptance: distinct public error types/codes for timeout, CRC, sequence,
  unsupported, disconnect; `ClassifyError` recognizes them through the public
  boundary wrapping.

### DNP3-044 — Logging hooks (optional, no-op default)
- Commit message: `feat(master): optional diagnostic logger hook`
- `pkg/dnp3/master/logger.go` (new): public `LogLevel` (`LogInfo`/`LogWarn`/
  `LogError`), `LogEvent` (Level/Op/Seq/Msg/Err), the `Logger` interface, a
  `NopLogger` no-op (the default), `FuncLogger(f)` adapter, `SeqNA` sentinel,
  and `diagAdapter` that bridges the public `Logger` to the internal
  `master.DiagHook`.
- `pkg/dnp3/master/client.go`: added `Config.Logger` field + `WithLogger(l)`
  option; the `client` stores `logger` and both `NewClient` and
  `NewClientWithTransport` install `diagAdapter(config.Logger)` on the
  internal master (nil logger → silent). `Connect` emits a public `connect`
  event (info on success, error on failure) via the nil-safe `emitLog` helper.
- `internal/master/master.go`: added `DiagLevel`/`DiagEvent`/`DiagHook` types,
  `SetDiagnosticHook`, and a nil-safe `diag(...)` helper that snapshots the
  hook under a read lock and invokes it OUTSIDE the master's own locks (so
  callbacks may safely query master state). `transitionTo` emits `state`
  events (info on legal transitions, error on illegal); `sendWithRetryAndGetResponse`
  emits `send`/`confirm`/`receive`/`retry` events with the application seq;
  the idle monitor emits a `state` error event on idle-driven close.
- Tests: `pkg/dnp3/master/logger_hook_test.go` — default-silent (no-logger
  path is safe on Read and on disconnect), hook-called on Read (send+receive,
  seq=0 on first request), hook-called on state transition (connect+state via
  the simulator), hook-called on failure (CRC → warn/error events),
  `NopLogger`/`FuncLogger`/`LogLevel.String` coverage. All green incl. `-race`.
- Acceptance: optional logger, default silent (no-op), hook called for
  frame/seq events.

### DNP3-045 — Public API loopback against simulator (full MVP)
- Commit message: `test(integration): full MVP public API loopback`
- `test/integration/mvp_loopback_test.go` (new): a single end-to-end full-MVP
  loopback against the deterministic in-memory simulator (no network I/O):
  `TestPublicMVPLoopbackFullLifecycle` exercises Connect → state Active →
  IntegrityPoll (all MVP Class-0 groups in one call) → assert 2 binary, 2
  analog (42, -7), 2 counter (100, 7) points + IIN == LastIIN → Operate (CROB
  DirectOperate) → assert ControlSuccess.
  `TestPublicMVPLoopbackOperateStatus` asserts a configured `ControlBlocked`
  outstation surfaces `ControlBlocked` through the public Operate path.
  `TestPublicMVPLoopbackErrorClassification` closes the simulated peer and
  asserts the next Read surfaces `dnp3.ErrNotConnected` and a non-`Unknown`
  `ClassifyError` category (validates the DNP3-043 taxonomy end-to-end).
- All green incl. `-race`.
- Acceptance: Connect → Integrity → Operate against the simulator only;
  points and command status asserted.

## Next READY Tasks

- **DNP3-049** — Master address configuration validation
- **DNP3-059** — Transport fragment size boundary tests
- **DNP3-065** — Double-check DLL EncodedSize usage
- **DNP3-072** — Master handoff.md template
- **DNP3-080, 084, 085, 087, 088, 098** (Outstation-side READY tasks)

## Recommended Next Task

**DNP3-049 — Master address configuration validation** (prereq: none).

If DNP3-049 is blocked, fall back to **DNP3-059 — Transport fragment size
boundary tests**.

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
- **DNP3-039 discovery (bug fixed):** `StateError` has iota ordinal 5, higher
  than `StateActive` (4). The legacy operation guards used
  `State() < StateInitialized` (ordinal 3), so a master in `StateError`
  (dead link) silently satisfied the readiness gate and operations were
  attempted on a dead link. Formalized to an explicit `isOperational()` allow-
  list (Initialized/Active only).
- **DNP3-037 workaround:** IntegrityPoll issues per-group reads rather than one
  multi-group read because the legacy `skipGroupData` response parser
  (DNP3-035 discovery) can lose points when a single response carries multiple
  MVP headers. Tracked for a later parser-robustness task.

## MVP Gate

MVP is declared complete only when **DNP3-056** passes.

```
TOTAL TASKS: 100
MASTER TASKS: 72
OUTSTATION TASKS: 28
MVP COMPLETE AT: DNP3-056
COMPLETED: DNP3-001 through DNP3-045
NEXT TASK: DNP3-049 — Master address configuration validation
```

## Test Status

- `go test ./...` — all packages green (including integration + simulator).
- `go test -race ./internal/master/... ./internal/testutils/... ./pkg/dnp3/... ./test/integration/...` — green (DNP3-043/044/045 verified).
- Pre-existing `go vet` "unreachable code" note in `internal/outstation/outstation.go:827` is NOT introduced by these tasks (confirmed on clean HEAD; `outstation.go` untouched by DNP3-043/044/045) and is out of scope.
- Checkpoint commits: `37277b3` (DNP3-016/017/018), `d45948d` (DNP3-019/020/021), `7ccd9cd` (DNP3-022/023/024), `22b1fe7` (DNP3-025/026/027), `c650a70` (DNP3-028/029/030), `ffa7908` (DNP3-031/032/033), DNP3-034/035/036, DNP3-037/038/039 (`4c4ac0f`), DNP3-040/041/042, then DNP3-043/044/045 (this checkpoint). All pushed to origin/main.
