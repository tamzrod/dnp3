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

The v0 MVP path below is verified by the repository's own deterministic
loopback and unit tests (internal verification, not external
interoperability). External interoperability (VEC-01) remains pending. Every
"Target" capability in the disposition table now has at least one in-repo test
reference:

| Capability | Test reference | Task |
|------------|----------------|------|
| LSB-first wire encoding of MVP point indices/values/CROB times | `pkg/dnp3/master/object_vector_test.go`, `internal/master/control_vector_test.go`, `internal/master/master_test.go`, `test/integration/tcp_test.go` | DNP3-001 |
| Object-header encode/decode model | `internal/al/object_header_test.go` | DNP3-002/003 |
| Master read path (Class-0 headers, range16 qualifier) | `pkg/dnp3/master/client_test.go` (`TestBuildReadRequestGolden`), `internal/master/master_test.go` | DNP3-004/005 |
| Link-layer handshake (reset link + link status) | `internal/dll/link/link_test.go`, `test/integration/tcp_test.go` | DNP3-006/007 |
| `Connect` / `Disconnect` / `Close` lifecycle + context cancellation | `pkg/dnp3/master/client_test.go`, `test/integration/close_reuse_test.go` | DNP3-022/024/050 |
| Public `Read` / `IntegrityPoll` (Class-0 G1/G30/G20) | `pkg/dnp3/master/client_test.go`, `test/integration/mvp_loopback_test.go` | DNP3-036/037/045 |
| Public `Operate` (Direct Operate G12V1 CROB) + command status | `test/integration/mvp_loopback_test.go` | DNP3-021/045 |
| Retry / timeout / outstanding-request tracking | `internal/master/retry_policy_test.go`, `internal/master/outstanding_request_test.go` | DNP3-031/032/040 |
| Idle-timeout keep-alive close | `internal/master/idle_monitor_test.go` | DNP3-042 |
| Public error taxonomy + `ClassifyError` | `pkg/dnp3/error_taxonomy_test.go`, `pkg/dnp3/master/error_classification_test.go` | DNP3-043 |
| Optional diagnostic logger hook (default silent) | `pkg/dnp3/master/logger_hook_test.go` | DNP3-044 |
| Full MVP loopback (Connect → Integrity → Operate) against simulator | `test/integration/mvp_loopback_test.go` | DNP3-045 |
| Master/outstation address validation | `pkg/dnp3/master/client_test.go` (`TestConfigValidate`, `TestNewClientRejectsInvalidConfig`) | DNP3-049 |
| Client reusable after `Close` (Close → Connect again) | `test/integration/close_reuse_test.go` | DNP3-050 |
| Auto-integrity re-poll on DeviceRestart IIN (opt-in) | `test/integration/auto_integrity_test.go` | DNP3-053 |
| Transport fragment-size boundaries (0/249/250) | `internal/tl/boundary_test.go` | DNP3-059 |
| MVP acceptance gate (single command) | `scripts/verify-mvp.sh` | DNP3-052/056 |

No capability in this profile is verified for **external** interoperability
yet; an independent raw capture (VEC-01) is still required for that claim.

### MVP Acceptance Gate Record (DNP3-056)

The Master MVP acceptance gate was run on this commit via
`./scripts/verify-mvp.sh` and passed:

- `go build ./...` — OK
- `go vet` (MVP packages, excluding the pre-existing out-of-scope
  `internal/outstation` unreachable-code note) — OK
- `go test -count=1` (MVP unit + integration) — all OK
- `go test -race -count=1` (MVP race-relevant packages) — all OK
- `verify-mvp.sh` exit code: 0

**MVP COMPLETE** (internal verification; external VEC-01 interoperability
remains pending). The gate is reproducible: any session may re-run
`./scripts/verify-mvp.sh` to confirm.

## Object Wire-Field Inventory

| Field | Required v0 encoding | Current implementation status |
|-------|----------------------|-------------------------------|
| Object group, variation, qualifier, count | One-octet fields | Present; verified by `internal/al/object_header_test.go` (DNP3-002/003). |
| Point index | Two-octet LSB-first unsigned integer | LSB-first; verified by `pkg/dnp3/master/object_vector_test.go` and `internal/master/master_test.go` (DNP3-001). |
| Binary Input Variation 1 value | One packed state bit | Present; verified by the MVP loopback in `test/integration/mvp_loopback_test.go` (DNP3-045). |
| Analog Input Variation 1 value | Four-octet signed integer with one quality octet, LSB-first value order | LSB-first int32 + quality; verified by `pkg/dnp3/master/object_vector_test.go` and the MVP loopback (DNP3-001/045). |
| Counter Variation 1 value | Four-octet unsigned integer, LSB-first octet order | LSB-first uint32; verified by `pkg/dnp3/master/object_vector_test.go` and the MVP loopback (DNP3-001/045). |
| Quality flags | One octet | Present; surfaced through the MVP loopback assertions (DNP3-045). |
| CROB index | Two-octet LSB-first unsigned integer | LSB-first; verified by `internal/master/control_vector_test.go` and the Operate loopback (DNP3-001/045). |
| CROB times | Four-octet unsigned integers, LSB-first octet order | LSB-first; verified by `internal/master/control_vector_test.go` (DNP3-001). |
| IIN | Two octets in application response | Present; `IntegrityPoll` response IIN equals `LastIIN()` asserted in the MVP loopback (DNP3-045). |

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
