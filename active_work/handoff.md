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

## Next READY Tasks

- **DNP3-004** — Wire Master read path to object-header model  *(prereqs: DNP3-001 ✓, DNP3-003 ✓)*
- **DNP3-006** — Link handshake ACK validation
- **DNP3-008** — Application sequence continuity
- **DNP3-011** — IIN bit semantics verification & correction
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

**DNP3-004 — Wire Master read path to object-header model**

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
COMPLETED: DNP3-001, DNP3-002, DNP3-003
NEXT TASK: DNP3-004 — Wire Master read path to object-header model
```

## Test Status

- `go test ./...` — all packages green (including integration).
- `go test -race ./internal/master/... ./pkg/dnp3/master/...` — green.
- Commit hash: (pending commit — see git log after commit)
