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

`Target` means it belongs to the initial interoperability path (the tested v0
subset now carries scoped external evidence as of 2026-11-11; broader claims are
not made). `Reject` means the public API must return a clear
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
loopback and unit tests (internal verification). Scoped **external** v0
evidence now also exists: the owner-accepted, isolated-lab VEC-01 campaign
(2026-11-11, `work` checkpoint `6c297df0fb7c702f66818e506008e00e9d716ea4`)
demonstrated one Go Master interoperating with an independently implemented
**OpenDNP3 3.1.2** outstation for this v0 path only. See
[`../docs/verification/VEC-01-LAB-EXEC-001.md`](../docs/verification/VEC-01-LAB-EXEC-001.md)
§§50–55 and the durable evidence under
[`testdata/external/vec01-20261111/`](testdata/external/vec01-20261111/README.md).
Every "Target" capability in the disposition table now has at least one in-repo
test reference:

| Capability | Test reference | Task |
|------------|----------------|------|
| LSB-first wire encoding of MVP point indices/values/CROB times | `pkg/dnp3/master/object_vector_test.go`, `internal/master/control_vector_test.go`, `internal/master/master_test.go`, `test/integration/tcp_test.go` | DNP3-001 |
| Object-header encode/decode model | `internal/al/object_header_test.go` | DNP3-002/003 |
| Master read path (Class-0 headers, range16 qualifier) | `pkg/dnp3/master/client_test.go` (`TestBuildReadRequestGolden`), `internal/master/master_test.go` | DNP3-004/005 |
| Request/response qualifier allow-list (0x00 start/stop / 0x06 all-objects / 0x07 count8 / 0x17 count8+index8 / 0x27 count16 / 0x28 count+index) | `internal/master/qualifier_golden_test.go`, `internal/al/object_header_test.go`, `pkg/dnp3/master/class0_qualifier_test.go`, `internal/outstation/static_qualifier_test.go` | MEXT-016; VEC-01 corrective #9 |
| Link-layer handshake (reset link + link status) | `internal/dll/link/link_test.go`, `test/integration/tcp_test.go` | DNP3-006/007 |
| Link handshake external frame vectors + Connect both-exchanges requirement (MEXT-017) | `internal/master/link_handshake_vectors_test.go`, golden `active_work/testdata/link-*.hex` | MEXT-017 |
| App-layer SEQ stream (0-15 wrap) + CON confirm + response SEQ match (solicited path) | `internal/master/master_test.go`, `internal/master/app_confirm_test.go`, `internal/master/confirm_timeout_test.go`, `internal/master/fcb_test.go` | DNP3-008/009/010/055 |
| SEQ + CON solicited-path audit (end-to-end CON confirm+response, mismatch rejection, retry-SEQ characterization) (MEXT-018) | `internal/master/seq_con_audit_test.go` | MEXT-018 |
| IIN bit-map freeze vs IEEE 1815 (named critical masks + full 16-bit round-trip) (MEXT-019) | `internal/al/iin_freeze_test.go`, `internal/al/application.go` (IIN FREEZE note) | MEXT-019 |
| VEC-01 external capture fixture format (`.vec` + PCAP sidecar) (MEXT-020) | `active_work/testdata/external/FORMAT.md`, `sample-vec01-placeholder.vec` | MEXT-020 / R4 |
| External MVP gate script (Tier 1 internal real-TCP + Tier 2 VEC-01 fail-closed) (MEXT-021) | `scripts/verify-external-mvp.sh`, `scripts/README.md` | MEXT-021 / R4 |
| Real-TCP full MVP path (Connect→IntegrityPoll→Operate→Disconnect over TCP, no sim transport) + outstation V0=default-variation (MEXT-022) | `test/integration/real_tcp_full_mvp_test.go`, `internal/outstation/static_variation_test.go`, `internal/outstation/outstation.go` (normalizeStaticVariation) | MEXT-022 |
| Workbench external-smoke (programmatic public-API E2E over real TCP; reproducible from README) (MEXT-023) | `scripts/run-workbench-smoke.sh`, `scripts/workbench_e2e.go`, `scripts/README.md` | MEXT-023 |
| Operate success/fail matrix on TCP (success / not-supported / timeout-on-drop) + bridge non-success-status fix (MEXT-024) | `test/integration/operate_status_matrix_test.go`, `pkg/dnp3/outstation/bridge_status_matrix_test.go`, `pkg/dnp3/outstation/server.go` (WriteBinaryOutput/WriteAnalogOutput non-success → error) | MEXT-024 |
| Reconnect + DeviceRestart IIN on TCP (no stuck state after peer drop; DeviceRestart not raisable on v0 outstation) (MEXT-025) | `test/integration/reconnect_restart_tcp_test.go`, `pkg/dnp3/master/disconnect_errno_test.go`, `internal/master/master.go` (IsDisconnectError errno match), `pkg/dnp3/master/client.go` (markTransportDisconnected) | MEXT-025 |
| Negative: bad CRC / wrong address must not hang (no deadlock) (MEXT-026) | `test/integration/rogue_peer_negative_test.go` (real-TCP rogue peer), `pkg/dnp3/master/crc_reject_test.go` (TestReadBadCRCNoHangDeadline) | MEXT-026 |
| External-use master example (Connect→IntegrityPoll→Operate→Close over TCP) (MEXT-031) | `examples/master/main.go`, `examples/README.md` | MEXT-031 |
| Master error taxonomy consumer notes (ClassifyError + ErrorCode table; NACK → Sequence) (MEXT-032) | `docs/error-taxonomy.md`, `README.md` | MEXT-032 |
| External-MVP full gate (real-TCP MVP + operate matrix + multi-header; exit 0; VEC-01 advisory) (MEXT-033) | `scripts/verify-external-mvp.sh`, `scripts/README.md` | MEXT-033 |
| README external claim (v0 verified via real-TCP gate; NOT third-party interop; NOT full IEEE 1815) (MEXT-034) | `README.md` | MEXT-034 |
| External MVP acceptance record (both verify scripts exit 0; EXTERNAL MVP COMPLETE) (MEXT-035) | `active_work/handoff.md`, `active_work/supported-profile.md` | MEXT-035 |
| `Connect` / `Disconnect` / `Close` lifecycle + context cancellation | `pkg/dnp3/master/client_test.go`, `test/integration/close_reuse_test.go` | DNP3-022/024/050 |
| Public `Read` / `IntegrityPoll` (Class-0 G1/G30/G20) | `pkg/dnp3/master/client_test.go`, `pkg/dnp3/master/integrity_poll_test.go`, `test/integration/mvp_loopback_test.go` | DNP3-036/037/045 |
| `IntegrityPoll` single multi-header exchange + per-group fallback (MEXT-015) | `pkg/dnp3/master/integrity_poll_test.go` (`TestIntegrityPollSingleMultiHeaderExchange`, `TestIntegrityPollFallbackPerGroup`) | MEXT-015 |
| Public `Operate` (Direct Operate G12V1 CROB) + command status | `test/integration/mvp_loopback_test.go` | DNP3-021/045 |
| `Operate` real-TCP success + classified failure (MEXT-013) | `test/integration/operate_real_tcp_test.go` (`TestOperateRealTCPSuccess`, `TestOperateRealTCPBlockedStatus`) | MEXT-013 |
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

The v0 path now has scoped external evidence from a single independently
implemented peer (OpenDNP3 3.1.2) in an isolated lab. This is **not** a general
external-interoperability, conformance, or production-readiness claim: it covers
only the tested v0 subset (Class-0 G1V1/G30V1/G20V1 read and one G12V1
Direct-Operate) against that one peer/version, and independent evidence for any
other peer, vendor, version, or profile is still required. A known unresolved
defect remains: the master mis-decodes **G1V2** (Binary Input With Flags) by
consuming the flags octet as the point value — outside the locked v0 G1V1
profile, not fixed and not claimed supported (see the G1V2 caveat in the
verification document §52.5).

### MVP Acceptance Gate Record (DNP3-056)

The Master MVP acceptance gate was run on this commit via
`./scripts/verify-mvp.sh` and passed:

- `go build ./...` — OK
- `go vet` (MVP packages, excluding the pre-existing out-of-scope
  `internal/outstation` unreachable-code note) — OK
- `go test -count=1` (MVP unit + integration) — all OK
- `go test -race -count=1` (MVP race-relevant packages) — all OK
- `verify-mvp.sh` exit code: 0

**MVP COMPLETE** (internal verification; scoped external VEC-01 v0 evidence was
separately accepted on 2026-11-11 — see the Current Verification State note and
R4 above — while general external interoperability beyond the tested v0 path
remains unproven). The gate is reproducible: any session may re-run
`./scripts/verify-mvp.sh` to confirm.

## External Residuals

These residuals describe boundaries of external use of the v0 profile. Each has
a resolving MEXT task; the list is authoritative so agents need not read git
history. R4 now carries a scoped, documentation-only resolution for the tested
v0 path (see its row); the others remain as stated.

| ID | Residual | Impact | Resolving task(s) |
|----|----------|--------|-------------------|
| R1 | Operate against a real outstation often `ControlTimeout` because real outstations may omit the G12V1 control-status echo on a valid Direct-Operate success. **Resolved by MEXT-012/013** — MEXT-012: IIN-only response with clear IIN now treated as success (error IIN never success; truncated G12V1 never success); MEXT-013: proven on real TCP — DirectOperate against the in-repo outstation (IIN-only response, no G12V1 echo) returns ControlSuccess, not ControlTimeout. | Resolved. | MEXT-012 ✅, MEXT-013 ✅ |
| R2 | CROB control-code values were a 1..8 enum, then (MEXT-011) an incorrect one-hot bitfield. **Resolved** by the owner-authorized coherent correction (constants now the IEEE 1815 A.8.1 low-nibble enum 0x00/0x01/0x02/0x03/0x04; master encoder, outstation decode, public bridge and goldens updated). | Resolved. Real IEDs receive spec-correct codes. | MEXT-010/011; coherent CROB correction ✅ |
| R3 | Multi-object-header Class-0 parse can lose points; `IntegrityPoll` used a per-group workaround. Fragile versus real multi-header responses. **Resolved by MEXT-014/015** — `skipGroupData` is qualifier-aware (G1+G20+G30 in one APDU parsed without point loss), and `IntegrityPoll` now uses a single Class-0 multi-group read as the primary path (one exchange returns the full set), with a per-group fallback for peers that reject the multi-group exchange. | Resolved. | MEXT-014 ✅, MEXT-015 ✅ |
| R4 | **Historical blocker:** no VEC-01 / independent PCAP or third-party stack capture proof existed; repository self-encoded frames were not acceptable evidence. **Scoped resolution (2026-11-11):** the owner accepted the isolated-lab VEC-01 campaign at `work` checkpoint `6c297df0fb7c702f66818e506008e00e9d716ea4` — one Go Master ↔ independently implemented OpenDNP3 3.1.2 outstation for the v0 subset only (Class-0 G1V1/G30V1/G20V1 read `BI[0]=true`, `AI[0]=2300` ONLINE, `CT[0]=42` ONLINE; one G12V1 Direct-Operate SUCCESS with `BO[0]=true`; negative payload-block CRC rejection). **Documented disposition: `R4 RESOLVED — VEC-01 v0 independent-lab evidence accepted`** (documentation-only, scoped). | Scoped external interop is evidenced for the tested v0 path with that single peer/version in the isolated lab; it is **not** general conformance, production readiness, or interoperability with any other peer. See `docs/verification/VEC-01-LAB-EXEC-001.md` §§50–55 and `testdata/external/vec01-20261111/`. | MEXT-020, MEXT-022, MEXT-033; VEC-01 post-corrective campaign ✅ (scoped) |
| R5 | Marketing/docs risk of over-claiming; consumers may assume production-ready. | Consumers assume production-ready. | MEXT-005, MEXT-034, MEXT-035 |
| R6 | Qualifier `0x00` was mislabelled `QualIndex8` (count + 2-octet index) instead of IEEE 1815 `UINT8_START_STOP` (8-bit start + 8-bit stop); the indexed-count code `0x17` was absent. **Resolved** by the owner-authorized VEC-01 corrective #9: `0x00` now decodes/encodes as a 5-octet start/stop header, `0x17` as count8 with per-point 8-bit index prefixes, and the outstation static builders emit `0x17`/`0x07` accordingly. | Resolved. Conformant peers' start/stop Class-0 responses decode. | VEC-01 corrective #9 |

## Object Wire-Field Inventory

| Field | Required v0 encoding | Current implementation status |
|-------|----------------------|-------------------------------|
| Object group, variation, qualifier, count | One-octet fields | Present; verified by `internal/al/object_header_test.go` (DNP3-002/003). |
| Point index | Two-octet LSB-first unsigned integer | LSB-first; verified by `pkg/dnp3/master/object_vector_test.go` and `internal/master/master_test.go` (DNP3-001). |
| Binary Input Variation 1 value | One packed state bit | Present; verified by the MVP loopback in `test/integration/mvp_loopback_test.go` (DNP3-045). |
| Analog Input Variation 1 value | Four-octet signed integer with one quality octet, LSB-first value order | LSB-first int32 + quality; verified by `pkg/dnp3/master/object_vector_test.go` and the MVP loopback (DNP3-001/045). |
| Counter Variation 1 value | Four-octet unsigned integer, LSB-first octet order | LSB-first uint32; verified by `pkg/dnp3/master/object_vector_test.go` and the MVP loopback (DNP3-001/045). |
| Quality flags | One octet | Present; surfaced through the MVP loopback assertions (DNP3-045). |
| CROB operation type | One octet, IEEE 1815 A.8.1 low-nibble enum (0x00 NUL, 0x01 Pulse On, 0x02 Pulse Off, 0x03 Latch On, 0x04 Latch Off); upper nibble (trip/close, clear) zero in v0 | 1815 low-nibble enum. Goldens in `internal/master/control_vector_test.go`; outstation decode in `internal/outstation/outstation.go`; public bridge in `pkg/dnp3/outstation/server.go`. |
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
   **Satisfied (scoped) 2026-11-11:** the accepted VEC-01 campaign used original
   lab-only PCAPs from an independently implemented OpenDNP3 3.1.2 peer, with
   packet-derived `.vec` vectors and a SHA-256 manifest under
   `testdata/external/vec01-20261111/`. Independent captures for other peers,
   versions, or profiles would still be required for any broader claim.
