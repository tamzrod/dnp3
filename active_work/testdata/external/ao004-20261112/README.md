# AO-004 Appendix U — isolated G41V1–V4 capture evidence (2026-11-12)

Durable, non-sensitive evidence for the owner-authorized **evidence-provenance
and closure** microtask (authorization `fce286e`, after Appendix T checkpoint
`800e6be`). This directory holds a **new** capture taken under that
authorization; it is **not** a recovered historical original.

## Provenance

- **This is a NEW capture**, never the original. The historical AO-004 capture
  artifacts referenced by Appendix H (`/tmp/ao004-docker4/cap*.pcap`,
  `driver*.out`) are **`ORIGINAL_PCAP_UNAVAILABLE`**: they lived only in the
  ephemeral OpenHands runner workspace and no longer exist. No original AO-004
  PCAP was ever committed; none is reconstructed, relabelled, or claimed.
- Peer: OpenDNP3 **3.1.2**, annotated tag `3.1.2` =
  `523f7d2d6c65aa671f594fc63c3adcd74e81407d`, peeled commit
  `26b4c01e4839bbbda8866655e086471c4917ee53`, tarball SHA-256
  `183cc29222c3cb58099a9753c43defed5cb8677fda985c3f88e38b5c19a36ff2`.
- Outstation under test: the `work` Go DNP3 outstation via the memory-only
  `lab/ao004/go-outstation` fixture (permits index 0 only; returns
  `ControlOutOfRange` otherwise). Static Go build SHA-256
  `f70ea941206648cebaa59b4ca800b12817f1817db586538eadde92c056b07e0d`.
- Lab: `network_mode: none` container (not privileged; caps
  `NET_ADMIN`,`NET_RAW`,`SYS_ADMIN` only; no host/ports/socket), two private
  namespaces `vec01-master` (`10.77.0.1/24`) and `vec01-outstation`
  (`10.77.0.2/24`) joined by one lab-only veth; **no** default route, DNS, NAT,
  forwarding or external access (proven before any packet). DNP3 link addresses
  master `3` / outstation `4`. TCP port `20000`. Capture interface `veth-o`.
- Capture: transient AF_PACKET helper writing classic-LE pcap, 2026-11-12.
  Analysis: independent Python decoder (TCP reassembly, DNP3 CRC-16 validation,
  TL/APDU and G41 record decode). All frames are observed wire bytes.

## Files

| File | Notes |
|------|-------|
| `ao004-20261112-isolated-g41v1-v4.pcap` | New lab-only PCAP, 128 packets / 8 TCP connections, only `10.77.0.1`↔`10.77.0.2:20000` |
| `ao004-u-20261112-g41v{1..4}-idx0-success.vec` | Packet-order link frames, index 0 (SUCCESS) |
| `ao004-u-20261112-g41v{1..4}-idx1-out_of_range.vec` | Packet-order link frames, index 1 (OUT_OF_RANGE) |
| `SHA256SUMS.txt` | SHA-256 of every file above |

## Verified results

Each `.vec` is one fresh-connection exchange: a G41 Direct Operate WITH RESPONSE
request, the echoed G41 response, then a bounded startup poll.

| Case | Request (M→O) | Response (O→M) | Master point result |
|------|---------------|----------------|---------------------|
| G41V1 idx0 | `G41 V1 q=0x28 cnt=1 idx=0 val=FC080000 st=00` | `G41 V1 idx=0 val=FC080000 st=00` IIN `0000` | `state=SUCCESS status=SUCCESS` |
| G41V2 idx0 | `G41 V2 idx=0 val=3930 st=00` | `G41 V2 idx=0 val=3930 st=00` IIN `0000` | `state=SUCCESS status=SUCCESS` |
| G41V3 idx0 | `G41 V3 idx=0 val=0000C03F st=00` | `G41 V3 idx=0 val=0000C03F st=00` IIN `0000` | `state=SUCCESS status=SUCCESS` |
| G41V4 idx0 | `G41 V4 idx=0 val=000000000000F83F st=00` | `G41 V4 idx=0 val=000000000000F83F st=00` IIN `0000` | `state=SUCCESS status=SUCCESS` |
| G41V1 idx1 | `G41 V1 idx=1 val=FC080000 st=00` | `G41 V1 idx=1 val=FC080000 **st=0x0C**` IIN `0000` | `state=SUCCESS status=OUT_OF_RANGE` |
| G41V2 idx1 | `G41 V2 idx=1 val=3930 st=00` | `G41 V2 idx=1 val=3930 **st=0x0C**` IIN `0000` | `state=SUCCESS status=OUT_OF_RANGE` |
| G41V3 idx1 | `G41 V3 idx=1 val=0000C03F st=00` | `G41 V3 idx=1 val=0000C03F **st=0x0C**` IIN `0000` | `state=SUCCESS status=OUT_OF_RANGE` |
| G41V4 idx1 | `G41 V4 idx=1 val=000000000000F83F st=00` | `G41 V4 idx=1 val=000000000000F83F **st=0x0C**` IIN `0000` | `state=SUCCESS status=OUT_OF_RANGE` |

- All 8 exchanges: every link frame header and payload CRC **valid**; application
  function `0x81` (Response) with clean IIN `0x0000`; echoed index/variation/value
  identical to the request (for idx1 the fixture accepted **0** writes).
- Index-0 positive control and index-1 `OUT_OF_RANGE` reproduce Appendix T.
- **Float divergence (Appendix H):** the Go-master → OpenDNP3-outstation V3/V4
  `-0.0`/subnormal characterization was **not** re-runnable — the only C++
  master harness (`lab/ao004/g41_master.cpp`) sends fixed finite values
  (`1.5f`/`1.5`) and the Go-master harness does not exist in-repo. The Appendix H
  finding therefore remains **evidence-only historic** (see Appendix U).

## Reproduction

Isolation gate (before any packet): container has only `lo` and **0** routes;
after topology each namespace has **0** default routes and `ip_forward=0`, a
lab-local ping succeeds, and external DNS/route are unreachable.

```
# per case V in 1..4, idx in 0 1
ip netns exec vec01-outstation /opt/ao004/go-outstation --listen &
ip netns exec vec01-master /opt/ao004/g41_master --run --variation V --index idx --timeout 15
# capture: ip netns exec vec01-outstation <afpacket-helper> -i veth-o -port 20000 -o cap.pcap
```

`.vec` derivation: reassemble each TCP direction, split DNP3 link frames, emit
in wire order per `../FORMAT.md`. Decode and CRC verification via the transient
Python decoder described in Appendix U.

## Limits

- Isolated simulated-lab interoperability re-test only. It does **not** establish
  IEEE 1815 conformance, production SCADA readiness, or third-party
  interoperability. G41 remains **UNSUPPORTED**; `INTEROP` for V3/V4 signed
  zero/subnormal remains **FAIL/DIVERGENCE** (peer-side, historic).
- No secrets, credentials, tokens, private-network or field/OT data are included;
  the capture contains only the synthetic memory-only lab exchange.
