# DNP3 Interoperability Remediation: Action Checklist

## Human Status

**Overall: IN PROGRESS — do not archive yet.**

| Status | Meaning | Current items |
|---|---|---|
| COMPLETE | Implemented and verified | PROF-01..03, VEC-01..03, DLL-01..12, TCP-01..03, STACK-01..03, OBJ-01..06, API-01, SAFE-01..02 |
| BLOCKED | Requires a design decision | API-02 — choose TCP proxy capture or transport-observer instrumentation |
| NEXT | Next executable action | SAFE-03 — context cancellation tests |
| REMAINING | Not started | API-03, SAFE-04, CTRL-01..02, DOC-01, REL-01..02 |

Archive rule: move this file to `archive/` only when every task is `COMPLETE`
and the final verification gate passes.

## Scope Lock

Until every required task below is complete, only work that directly advances
this checklist is permitted. Do not add protocol features, UI work, refactors,
or unrelated cleanup.

## Completion Rules

- Complete tasks in dependency order.
- One pull request or commit may contain only one task unless a task explicitly
  says otherwise.
- Mark a task complete only after its listed verification command passes.
- If a task disproves an assumption, add a dated note under `## Decisions and
  Evidence` before starting the next task.
- Do not update marketing/status claims until Task DOC-01.

## Task Board

| ID | Action | Depends on | Deliverable | Pass criterion |
|----|--------|------------|-------------|----------------|
| PROF-01 | Write the initial supported-profile statement. | — | `active_work/supported-profile.md` | **Complete 2026-08-10.** Defines exactly one supported path: TCP, master/outstation, Class 0 static reads, and the currently intended direct-control subset. |
| PROF-02 | List every public configuration/API option outside that profile. | PROF-01 | Unsupported-options table in the profile | **Complete 2026-08-10.** Every inventoried public option is marked `Target`, `Reject`, or `Defer`; none are implied supported. |
| PROF-03 | Choose the authoritative sources for wire vectors. | PROF-01 | Sources section in the profile | **Complete 2026-08-10.** Official DNP references are recorded; independent raw capture remains required by VEC-01. |
| VEC-01 | Add one raw, externally sourced link-header fixture. | PROF-03 | `active_work/testdata/racom-dnp3-link-frame.hex` | **Complete 2026-08-10.** Fixture records source, expected fields, and exact bytes. |
| VEC-02 | Add a test that decodes the header fixture. | VEC-01 | Failing test in `internal/dll/frame` | **Complete 2026-08-10.** `TestDecodeRacomGoldenFrame` decodes the published unconfirmed-user-data frame. |
| VEC-03 | Add payload CRC fixtures at 0, 1, 16, 17, 249, and 250 bytes. | PROF-03 | `active_work/testdata/payload-crc-vectors.txt` | **Complete 2026-08-10.** Each boundary vector states full-frame length, header CRC, and payload-block CRCs. |
| DLL-01 | Correct primary link function-code constants only. | VEC-02 | `internal/dll/frame/frame.go` | **Complete 2026-08-10.** `go test ./internal/dll/frame -run TestControlByte -count=1` passes. |
| DLL-02 | Correct secondary link function-code constants only. | DLL-01 | `internal/dll/frame/frame.go` | **Complete 2026-08-10.** `TestSecondaryFunctionCodes` covers ACK, NACK, link status, not-supported, and user-data response codes. |
| DLL-03 | Encode destination addresses LSB-first. | DLL-01 | `frame.go` plus a focused test | **Complete 2026-08-10.** `TestAddressByteOrder` matches destination bytes. |
| DLL-04 | Decode destination addresses LSB-first. | DLL-03 | `frame.go` plus a focused test | **Complete 2026-08-10.** `TestAddressByteOrder` decodes destination correctly. |
| DLL-05 | Encode source addresses LSB-first. | DLL-03 | `frame.go` plus a focused test | **Complete 2026-08-10.** `TestAddressByteOrder` matches source bytes. |
| DLL-06 | Decode source addresses LSB-first. | DLL-05 | `frame.go` plus a focused test | **Complete 2026-08-10.** `TestAddressByteOrder` decodes source correctly. |
| DLL-07 | Replace split header CRC generation with one complete-header CRC. | DLL-03, DLL-05 | `frame.go` | **Complete 2026-08-10.** Encoder covers bytes 0–7 and matches the published header CRC. |
| DLL-08 | Replace split header CRC validation with complete-header validation. | DLL-07 | `frame.go` | **Complete 2026-08-10.** `TestDecodeRacomGoldenFrame` validates bytes 0–7 and `TestDecodeRejectsCorruptedHeader` rejects each header-byte mutation. |
| DLL-09 | Encode payload CRCs in 16-octet blocks. | VEC-03, DLL-07 | `frame.go` | **Complete 2026-08-10.** `TestPayloadCRCBoundaryVectors` matches all six payload boundary vectors. |
| DLL-10 | Decode and validate payload CRCs in 16-octet blocks. | DLL-09 | `frame.go` | **Complete 2026-08-10.** Boundary vectors decode and `TestDecodeRejectsCorruptedPayloadBlocks` rejects each block. |
| DLL-11 | Centralize the exact encoded-frame length calculation. | DLL-10 | Exported/internal helper with tests | **Complete 2026-08-10.** `EncodedSize` is used by encode/decode and boundary tests assert all six lengths. |
| DLL-12 | Remove obsolete two-byte CRC assumptions from DLL tests/comments. | DLL-11 | Updated DLL tests/docs | **Complete 2026-08-10.** `rg '16-bit quantity|2 bytes at a time' internal/dll/frame` returns no matches. |
| TCP-01 | Make TCP receive use the centralized wire-length calculation. | DLL-11 | `pkg/transport/tcp.go` | **Complete 2026-08-10.** `TestTCPReceiveGoldenFrame` receives the complete published frame over `net.Pipe`. |
| TCP-02 | Test two concatenated frames received in separate calls. | TCP-01 | Transport test | **Complete 2026-08-10.** `TestTCPReceiveConcatenatedFrames` returns exactly one frame per call. |
| TCP-03 | Test a golden frame delivered in fragmented writes. | TCP-01 | Transport test | **Complete 2026-08-10.** `TestTCPReceiveFragmentedWrite` returns one complete frame without loss or over-read. |
| STACK-01 | Make master inbound frame iteration use corrected frame length. | DLL-11 | `internal/master/master.go` | **Complete 2026-08-10.** Production loop uses `frame.EncodedSize`; integration offsets now use the same wire model. |
| STACK-02 | Make outstation inbound frame iteration use corrected frame length. | DLL-11 | `internal/outstation/outstation.go` | **Complete 2026-08-10.** Production loop uses `frame.EncodedSize`. |
| STACK-03 | Re-run all existing DLL/TL/AL stack tests against golden frames. | STACK-01, STACK-02 | Updated tests | **Complete 2026-08-10.** `go test ./test/integration -run TestProtocolStack -count=1` passes with corrected offsets. |
| OBJ-01 | Inventory byte order for every field in the supported profile. | PROF-02 | Table in `supported-profile.md` | **Complete 2026-08-10.** Table covers object header, qualifier, index, value, flags, time, CROB, and IIN. |
| OBJ-02 | Add one externally specified Class 0 binary-input fixture. | OBJ-01 | `active_work/testdata/class0-binary-input-vector.hex` | **Complete 2026-08-10.** Fixture records external references and expected point fields. |
| OBJ-03 | Correct binary input Variation 1 format and encode/decode byte order. | OBJ-02 | Focused source/test changes | **Complete 2026-08-10.** External packed-vector test passes; indexed Group 1 fields encode/decode LSB-first. |
| OBJ-04 | Add and fix one analog-input fixture for the supported variation. | OBJ-03 | Focused source/test changes | **Complete 2026-08-10.** `TestParseClass0AnalogInputVector` decodes index 0, value 1000, ONLINE. |
| OBJ-05 | Add and fix one counter fixture for the supported variation. | OBJ-04 | Focused source/test changes | Add `class0-counter-vector.hex`; parser test must decode index 0, value 1000, ONLINE. | 2026-08-10 | Complete |
| OBJ-06 | Add and fix one direct-control/CROB fixture. | OBJ-05 | Focused source/test changes | Exact request and response/status bytes pass. | 2026-08-10 | Complete |
| API-01 | Add an unskipped loopback test using public master and public outstation. | STACK-03, OBJ-06 | `test/integration` test | Connect, Class 0 read, and direct control all succeed over TCP. | 2026-08-10 | Complete |
| API-02 | Capture and assert the bytes exchanged by API-01. | API-01 | Test capture/assertions | Bytes match the relevant golden vectors. | 2026-08-10 | Blocked — public API exposes no transport capture seam; choose TCP proxy fixture or observer instrumentation |
| API-03 | Add a negative test for invalid CRC from the peer. | API-01 | Integration test | Client rejects the response without returning points. |
| SAFE-01 | Reject public TLS configuration with a clear unsupported error. | PROF-02 | Public master/server changes and tests | TLS cannot silently create plaintext TCP. | 2026-08-10 | Complete |
| SAFE-02 | Reject unsolicited enable/configuration with a clear unsupported error. | PROF-02 | Public master/server changes and tests | No successful enable result is returned without delivery support. | 2026-08-10 | Complete |
| SAFE-03 | Add context cancellation tests for public Connect, Read, and Stop. | API-01 | Public API tests | Canceled contexts return promptly and do not leave live connections. |
| SAFE-04 | Serialize public request, sequence, and reassembly state. | SAFE-03 | Public master changes/tests | `go test -race` concurrent-read test passes. |
| CTRL-01 | Parse per-point control status from an operate response. | OBJ-06 | Public master changes/tests | Rejected point does not report `ControlSuccess`. |
| CTRL-02 | Add select/operate timeout and mismatch integration tests. | CTRL-01 | Integration tests | Both failures surface as distinct errors/statuses. |
| DOC-01 | Rewrite README, STATUS, CONTRIBUTING, examples, and support matrix. | API-03, SAFE-02, CTRL-02 | Documentation changes | Every advertised capability has a passing test reference; unsupported items are explicit. |
| REL-01 | Create one reproducible verification command/script. | DOC-01 | `scripts/verify-interoperability.sh` or documented command | Runs unit, integration, raw-wire, and race checks. |
| REL-02 | Run the verification gate and record results. | REL-01 | Dated result in this file | All required commands pass; exclusions are listed. |

## Decisions and Evidence

Add entries here only when a task changes an assumption or establishes a
protocol interpretation that later tasks depend on.

| Date | Task | Decision or evidence | Reference |
|------|------|----------------------|-----------|
| 2026-08-10 | API-01 | Added `test/integration/public_api_loopback_test.go`; execution failed before server startup because the environment rejects `listen tcp 127.0.0.1:0` with `operation not permitted`. | `GOCACHE=/tmp/dnp3-go-cache go test ./test/integration -run TestPublicAPILoopbackReadAndDirectControl -count=1` |
| 2026-08-10 | API-01 | Consumed the Reset Link Stations acknowledgment during master handshake; public TCP loopback now passes Class 0 read and direct control. | `GOCACHE=/tmp/dnp3-go-cache go test ./test/integration -run TestPublicAPILoopbackReadAndDirectControl -count=1 -v` |
| 2026-08-10 | API-02 | Actual wire capture cannot be added cleanly through the public API without either a proxy-based test harness or a new transport-observer seam. | `pkg/dnp3/master/client.go` hides the transport behind the client implementation |
| 2026-08-10 | SAFE-01 | `NewClient` rejects TLS explicitly instead of constructing plaintext TCP; focused test passes. | `pkg/dnp3/master/safety_test.go` |
| 2026-08-10 | SAFE-02 | Public unsolicited enable/disable now return explicit unsupported errors; focused tests pass. | `pkg/dnp3/master/safety_unsolicited_test.go` |

## Current Task

`SAFE-03` — Add context cancellation tests for public Connect, Read, and Stop.
