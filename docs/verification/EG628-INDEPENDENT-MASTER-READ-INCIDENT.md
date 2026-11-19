# Independent-master DNP3 READ failure — EG628/Windows Tester

**Status: OPEN / ROOT CAUSE UNDETERMINED / NO FIX CLAIMED**
**Date:** 2026-11-19
**Owning library:** tamzrod/dnp3
**Independent peer:** Windows DNP3 Tester (exact build/version NOT CAPTURED)
**Hardware adapter:** EG628 (device commit / pinned DNP3 revision NOT CAPTURED)
**Investigation branch:** investigate/eg628-independent-master-read
**Working baseline:** work `1a1ac026ea51f676db0015e8dc4d9ca2b5f38f6c`
**CWAL isolation:** AO-008 is CURRENT on work; this investigation does not promote, replace, or modify it.

## Reported independent observation (owner report, not independently reproduced)

- TCP connection succeeds.
- Class 0+1+2+3 integrity READ fails with `malformed data: object header at offset 38: app: truncated object data`.
- Windows tester points table remains empty.
- IIN `PARAMETER_ERROR` also observed (not yet correlated to the same request/response sequence).
- **Original raw request, response, pcap, and tester logs: NOT PROVIDED.** Therefore neither faulty stack nor offset interpretation can be assigned. Do not blame EG628, dnp3, or the independent tester without bytes.

## Verified source inspection at baseline

- `internal/outstation/outstation.go`: `handleRead` builds a solicited response via `buildReadResponse`; Group 60 Variation 1 appends `buildAllStaticData()`; Variations 2/3/4 append `encodeClassRead` (class events).
- `buildReadResponse` iterates request object headers using `requestObjectHeaderSize` with qualifier-dependent widths. Current implementation breaks on a truncated request header rather than blindly advancing; verify exact response/IIN against captured malformed request rather than assuming a response-side error.
- `buildAllStaticData` concatenates G1V1 binary, G30V1 analog, G20V1 counter, G10V1 binary output, G40V1 analog output where populated. `buildBinaryInputData(1)` emits packed values with 0x07 count; G30V1 and G20V1 emit flags + little-endian 32-bit values with 0x07 count.
- `internal/al/object_header.go` contains qualifier-specific header decoding; transport/application fragmentation and event encoding also need verification if payload exceeds one fragment.
- **Suspicion only, not proven:** invalid count, qualifier, object payload width, unexpected event layout, or transport fragmentation at/near application offset 38. Also possible tester parse bug or adapter input mismatch. No fix is warranted before independent reproduction.

## Evidence-capture contract

Perform only **read-only DNP3 monitoring** on the authorized test network; do NOT send controls, modify GPIO/relays, or capture unrelated network traffic. Use a short capture limited to EG628 DNP3 TCP port and tester peer, or a dedicated isolated bench. Record UTC timestamps, tester version/build, remote/local endpoint, EG628 process binary SHA, dnp3 module build info and git revision, config/mapping sanitized for secrets, observed IIN, tester error and points table.

Capture full bidirectional TCP traffic as **pcapng/pcap** with full snaplen:
```sh
# On EG628 Linux; substitute the actual Windows Tester IP before execution.
sudo tcpdump -i any -s 0 -nn -U -w eg628-windows-tester-integrity.pcap 'tcp port 20000 and host <TESTER_IP>'
# Initiate exactly one Class 0+1+2+3 Integrity READ from tester, then stop tcpdump.
sha256sum eg628-windows-tester-integrity.pcap
go version -m /path/to/eg628-dnp3-binary
```
Keep the raw PCAP private until reviewed for endpoint/device metadata and sensitive payloads. Preserve the unedited original and SHA-256. Extract exact link + transport + application bytes **both directions**, with packet numbers, CRC check, FIR/FIN/CON, application sequence, function, IIN bytes, group/variation/qualifier/count/indices and zero-based application offsets. Distinguish TCP stream offsets from link frame, transport, and application body offsets. Include full request and response *even if the error mentions offset 38*; do not screenshot only.

## Independent no-hardware reproducer acceptance (NOT_RUN)

1. Checkout exact pinned `tamzrod/dnp3` revision used by the EG628 binary; configure a **fake in-memory** outstation whose BI/AI/counter/BO/AO, event buffers, class assignments and count/order match captured setup. No EG628 packages or hardware dependencies.
2. Replay the exact captured DNP3 Class 0+1+2+3 READ byte sequence into the generic outstation via loopback transport. Compare complete link frames and reassembled application response (IIN/object headers/object data), allowing only documented nondeterministic sequence/time values.
3. Parse the response through a **separate independent DNP3 implementation or an independently checked minimal packet decoder**, not only the library's own reader. Determine if the reported offset 38 points at malformed header, incorrect object width/count/qualifier, or incomplete application reassembly.
4. If defect is generic DNP3: write a minimal golden request/response wire regression covering the exact multi-class READ and golden field/CRC checks. Add negative cases for truncated headers, malformed count/length, unsupported qualifiers, invalid point payload, fragmentation/reassembly and IIN PARAMETER_ERROR. Fix only the generic library, not a Windows-tester-specific special case or EG628 workaround. Run Go build/vet/unit/race/integration/external-MVP and pinned isolated independent peer as relevant.
5. If defect is EG628 adapter: prove the generic loopback passes with identical valid points but adapter emits invalid data/uses unexpected DNP3 revision; fix in `tamzrod/eg628` only, pin corrected dnp3 revision and retest.
6. If independent tester is at fault: preserve exact standards-conformant wire response, multi-peer decode, and a minimal scrubbed failing trace; prepare a maintainers' issue rather than a library patch.
7. Pin/upgrade only the validated corrected library revision in EG628, with module checksum and binary/module verification. Run the **same original Windows Tester Class 0+1+2+3 scan** on actual EG628 and require populated point table and no erroneous PARAMETER_ERROR. Record tested build/version, wire PCAP and success criteria.

## Current evidence matrix

| Required proof | Status |
|---|---|
| Reported Windows Tester error | OWNER-REPORTED |
| Raw independent request/response frames | MISSING |
| Original pcap and SHA-256 | MISSING |
| Tester build, EG628 module pin | MISSING |
| Reproduction without hardware | NOT_RUN |
| Root cause attribution | UNDETERMINED |
| Golden wire regression / negatives | NOT_RUN |
| Go + external interop gates | NOT_RUN |
| EG628 module pin update | NOT_DONE |
| Actual Windows Tester retest | NOT_RUN |
| Defect resolved | **NOT CLAIMED** |

**Stop rule:** No statement of FIXED, interoperability PASS, production READY, or upstream PR until the original independent-master scan succeeds with captured, reproducible evidence. Work within an explicitly authorized implementation task/branch; this incident record alone does not override AO-008's CURRENT/GOV gates or authorize merging to `main`.
