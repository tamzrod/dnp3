# Supported Interoperability Profile v0

## Target Path

The first externally interoperable release supports only this path:

1. A DNP3 master opens one TCP connection to one outstation.
2. The master performs a Class 0 read.
3. The outstation returns static Binary Input, Analog Input, and Counter data.
4. The master sends one direct-operate control request.
5. The outstation returns the control result.

## Required Wire Behavior

- TCP transport only.
- Standard DNP3 link-layer framing.
- Standard DNP3 link function codes.
- LSB-first DNP3 multi-octet wire fields.
- One CRC for the complete link header.
- One CRC for each 16-octet payload block, including the final partial block.
- Maximum link payload: 250 octets.

## Included Object Subset

| Operation | Object |
|-----------|--------|
| Class 0 read | Group 1 Variation 1 Binary Input |
| Class 0 read | Group 30 Variation 1 Analog Input |
| Class 0 read | Group 20 Variation 1 Counter |
| Direct control | Group 12 Variation 1 CROB |

## Explicitly Unsupported

The public API must reject these options until a separately tested profile adds
them:

- TLS
- Serial transport
- Unsolicited responses and event delivery
- Secure authentication
- Time synchronization and time objects
- File transfer
- Device attributes
- Restart, delay measurement, and freeze operations
- Select-before-operate
- Object groups or variations not listed above

## Public API Disposition

`Target` means it belongs to the initial interoperability path but is not yet
externally verified. `Reject` means the public API must return a clear
unsupported error until a later profile adds it. `Defer` means it remains
available only after its dedicated verification task proves its behavior.

| Public option | Disposition | Boundary for v0 |
|---------------|-------------|-----------------|
| Master `WithMasterAddress` | Target | One configured master address. |
| Master `WithOutstationAddress` | Target | One configured outstation address. |
| Master `WithTransport(TCP, address, port)` | Target | One TCP peer. |
| Master `WithTransport(TLS, ...)` and `WithTLS` | Reject | No plaintext fallback. |
| Master `WithTimeout`, `WithRetry` | Defer | Requires lifecycle/timeout verification. |
| Master `Connect`, `Disconnect`, `Close` | Target | One connection lifecycle. |
| Master `Read` | Target | Class 0 static Groups 1.1, 30.1, and 20.1 only. |
| Master `Operate` | Target | Direct operate, Group 12 Variation 1 only. |
| Master select-before-operate and direct-operate-no-response | Reject | Not part of v0. |
| Master `EnableUnsolicited`, `DisableUnsolicited`, `SetUnsolicitedHandler` | Reject | No unsolicited delivery path. |
| Outstation `WithAddress`, `WithMasterAddress` | Target | One configured outstation/master address pair. |
| Outstation `WithTransport(TCP, address, port)` | Target | TCP listener only. |
| Outstation `WithTransport(TLS, ...)` and `WithTLS` | Reject | No TLS listener. |
| Outstation `WithMaxFragmentSize` | Defer | Must be reconciled with tested fragmentation limits. |
| Outstation `WithUnsolicitedMode`, `SetUnsolicitedHandler` | Reject | No unsolicited delivery path. |
| Outstation `Start`, `Stop`, `Close` | Target | One TCP listener lifecycle. |
| Outstation `SetDataHandler` | Target | Static Groups 1.1, 30.1, and 20.1 only. |
| Outstation `SetCommandHandler` | Target | Direct Group 12 Variation 1 control only. |
| Data handler binary/analog output status and frozen-counter methods | Reject | Not read by the v0 profile. |
| Command handler analog controls | Reject | Group 41 and related variations are out of scope. |
| `types.ReadAllStatic` | Defer | Its variation-zero requests require object-selection verification. |
| `types.ReadAllEvents` | Reject | Event objects are out of scope. |
| `types.ReadBinaryInputs`, `ReadAnalogInputs`, `ReadCounters` | Target | Limited to the listed Group/Variation pairs. |
| `types.NewBinaryControl(..., DirectOperate)` | Target | Group 12 Variation 1 only. |
| `types.NewAnalogControl`, `NewPulseControl` | Reject | Analog and pulse control profiles are out of scope. |
| Timestamp conversion APIs | Reject | Time objects are out of scope. |
| Direct `pkg/transport` TCP API | Defer | Tested only as implementation support for the public API. |
| Direct `pkg/transport` TLS API | Reject | Not a supported transport profile. |

## Current Verification State

No capability in this profile is verified for external interoperability yet.
Tasks VEC-01 through API-03 establish that verification.

## Object Wire-Field Inventory

| Field | Required v0 encoding | Current implementation status |
|-------|----------------------|-------------------------------|
| Object group, variation, qualifier, count | One-octet fields | Present; requires fixture verification. |
| Point index | Two-octet LSB-first unsigned integer | Current public parsers use big-endian; correction pending. |
| Binary Input Variation 1 value | One packed state bit | Current parser/builders use repository-specific bit handling; fixture required. |
| Analog Input Variation 1 value | Four-octet signed integer with one quality octet, LSB-first value order | Current parser/builders use float32/big-endian; correction pending. |
| Counter Variation 1 value | Four-octet unsigned integer, LSB-first octet order | Current parser/builders use big-endian; correction pending. |
| Quality flags | One octet | Present; semantic mapping requires fixture verification. |
| CROB index | Two-octet LSB-first unsigned integer | Current control paths require fixture verification. |
| CROB times | Four-octet unsigned integers, LSB-first octet order | Current control paths require fixture verification. |
| IIN | Two octets in application response | Current response path requires fixture verification. |

## Authoritative Wire References

These references define the wire behavior used by the remediation fixtures:

1. DNP Users Group, *A DNP3 Protocol Primer*, Revision A: link-frame layout,
   250-octet payload limit, and one CRC pair per 16 payload octets.
   [Public PDF](https://www.dnp.org/Portals/0/AboutUs/DNP3%20Primer%20Rev%20A.pdf)
2. DNP Users Group, *AN2013-004b Validation of Incoming DNP3 Data*: link
   control function codes, LSB/MSB address representation, and header layout.
   [Public PDF](https://www.dnp.org/Portals/0/Public%20Documents/DNP3%20AN2013-004b%20Validation%20of%20Incoming%20DNP3%20Data.pdf)
3. An independent raw capture or independently generated frame is required
   for VEC-01. Repository self-encoded frames are not acceptable evidence.
