# CROB Control-Code Audit (MEXT-010)

**Task:** MEXT-010 — CROB control-code IEEE 1815 bitfield audit
**Residual:** R2
**Date:** 2026-08-12
**Scope:** Read-only audit of the G12V1 CROB encode path + constants. No code
change in this task.

## IEEE 1815 G12V1 control-code bit field (authoritative)

The Group 12 Variation 1 Control Relay Output Block encodes its first octet as a
**bit field**, not an enumeration. Per IEEE 1815 (and the DNP3 Technical
Reference / AN2013-004b referenced in
`active_work/supported-profile.md`), the control-code octet is:

| Bit | Mask | Meaning |
|-----|------|---------|
| 0   | 0x01 | NUL — no operation |
| 1   | 0x02 | Pulse On — energize for `OnTime` |
| 2   | 0x04 | Pulse Off — de-energize for `OffTime` |
| 3   | 0x08 | Latch On — set output to ON |
| 4   | 0x10 | Latch Off — set output to OFF |
| 7   | 0x80 | Queue — add to outstation queue (do not execute immediately) |

Bits 5 and 6 are reserved (0). Combinations are valid via OR (e.g. 0x0A =
Latch On + Pulse Off). The field is a bitmask; arbitrary 1..8 integer enum
values are not on-spec.

## Current implementation values (audit)

Source: `internal/master/master.go` (constants + encode) and
`internal/outstation/outstation.go` (decode/execute switch).

### Constants — `internal/master/master.go:2310-2319`

```go
const (
    CROBCodeNUL      uint8 = 1
    CROBCodeClose    uint8 = 2 // turn ON
    CROBCodeOpen     uint8 = 3 // turn OFF
    CROBCodeTrip     uint8 = 4
    CROBCodePulseOn  uint8 = 5
    CROBCodePulseOff uint8 = 6
    CROBCodeLatchOn  uint8 = 7 // Value = true
    CROBCodeLatchOff uint8 = 8 // Value = false
)
```

These are a **1..8 sequential enumeration**, not the IEEE 1815 bitfield. The
existing code comment already flags this divergence.

### Encode path — `internal/master/master.go:2122-2154`

`buildControlRequest` (Group 12) writes `code` as a single octet at byte 0 of
the 11-byte CROB value. The **layout** (octet position) is correct; only the
**values** placed into that octet are off-spec.

The public `Operate` bool mapping:
- `value == true`  -> `CROBCodeLatchOn`  (7) — should be `0x08`
- `value == false` -> `CROBCodeLatchOff` (8) — should be `0x10`

A `uint8`/`uint16` value is passed through unchanged, so a caller who already
knows the IEEE 1815 masks can send correct bytes — but the named constants and
the bool helper emit non-1815 values.

### Decode/execute path — `internal/outstation/outstation.go:410-432`

`WriteBinaryOutput` switches on `crob.Code` using the same 1..8 enum
(`case 1` … `case 8`). The outstation decodes the code octet raw
(`outstation.go:1533`) and dispatches by enum value. Because both sides use the
same off-spec enum, the in-repo simulator loopback is self-consistent — but a
real IED that interprets the octet as the 1815 bitfield will misread these
codes.

## Current vs IEEE 1815 — reconciliation table

| Operation | Current constant | Current value | IEEE 1815 value | Match? |
|-----------|------------------|---------------|-----------------|--------|
| NUL / no-op | `CROBCodeNUL` | 1 (0x01) | 0x01 | ✅ coincidental |
| Pulse On | `CROBCodePulseOn` | 5 (0x05) | 0x02 | ❌ |
| Pulse Off | `CROBCodePulseOff` | 6 (0x06) | 0x04 | ❌ |
| Latch On | `CROBCodeLatchOn` | 7 (0x07) | 0x08 | ❌ |
| Latch Off | `CROBCodeLatchOff` | 8 (0x08) | 0x10 | ❌ |
| Close (turn ON) | `CROBCodeClose` | 2 (0x02) | no direct 1815 name; 0x02 = Pulse On | ❌ ambiguous |
| Open (turn OFF) | `CROBCodeOpen` | 3 (0x03) | no direct 1815 name; 0x03 = NUL+Pulse On | ❌ ambiguous |
| Trip | `CROBCodeTrip` | 4 (0x04) | no direct 1815 name; 0x04 = Pulse Off | ❌ ambiguous |
| Queue | (none) | — | 0x80 | ❌ missing |

Only NUL matches the spec, and only by coincidence (the enum starts at 1).

## Wire layout (correct, locked by DNP3-019)

For reference, the CROB **octet layout** is correct and unchanged by this audit:

```
Offset | Field   | Size | Order
-------+---------+------+-----------
  0    | code    | 1    | bitfield
  1    | count   | 1    | —
  2-5  | onTime  | 4    | LSB-first
  6-9  | offTime | 4    | LSB-first
 10    | status  | 1    | —
```

Goldens in `internal/master/control_vector_test.go` lock the layout and the
LSB-first time/index encoding; the code **values** in those goldens (e.g.
`Code: 7`, expected byte `0x07`) will need updating in MEXT-011.

## Affected tests (to update in MEXT-011)

- `internal/master/control_vector_test.go`
  - `TestBuildCROBRequestGoldenVector`: `Code: 7` -> `0x08`, expected byte `0x07` -> `0x08`
  - `TestBuildCROBRequestLayout`: `CROBCodeLatchOn` value, `wantVal` first byte `0x07` -> `0x08`
  - `TestBuildCROBRequestIndexLSBHighByte`: `CROBCodeLatchOff` value
  - `TestBuildCROBRequestTimeLSBBoundary`: `CROBCodePulseOn` value
  - `TestBuildControlRequestCROBBoolMapping`: true->`0x08`, false->`0x10`
- Any loopback/operate test asserting the code byte on the wire
  (`test/integration/mvp_loopback_test.go`, simulator tests) — verify the code
  value echoed by the outstation.

## Public-API impact (for MEXT-011)

- The named constants change value. Callers using `CROBCodeLatchOn` etc. by name
  are source-compatible; callers hardcoding `7`/`8` are not.
- The `uint8`/`uint16` passthrough in `buildControlRequest` stays, so callers
  passing raw IEEE 1815 masks continue to work.
- `CROBCodeClose`/`CROBCodeOpen`/`CROBCodeTrip` have no clean IEEE 1815
  single-bit equivalents. MEXT-011 should keep the spec-defined names
  (NUL/PulseOn/PulseOff/LatchOn/LatchOff/Queue) and either remove or
  remap the non-spec names. The public `pkg/dnp3/types` API uses a `bool` value
  + `DirectOperate`, so it is unaffected at the API surface — only the
  internal wire byte changes.

## Conclusion

R2 is confirmed: the G12V1 control-code values are a 1..8 enum, not the IEEE
1815 bitfield. The wire **layout** is correct; only the constant **values** and
the bool mapping diverge. Fix scoped for MEXT-011: realign constants to the
1815 bitfield, update goldens + outstation decode switch, keep the public API
stable via the bool path. Keep `./scripts/verify-mvp.sh` green.
