# AO-004 Appendix W — isolated negative/timeout coverage evidence (2026-11-13)

Durable, non-sensitive evidence for the owner-authorized **bounded malformed G41 /
timeout / quarantine interop coverage** microtask (authorization `807644e`, after
the CHECKED Appendix V checkpoint `01a2d3b`). This is a **new** capture taken
under that authorization.

## Provenance

- **New capture**, taken 2026-11-13 in the proven isolated lab; never a recovered
  historical original (Appendix U §U.1: the historical AO-004 PCAPs remain
  `ORIGINAL_PCAP_UNAVAILABLE`).
- Peer/master: pinned OpenDNP3 **3.1.2** (`tag 3.1.2` =
  `523f7d2d6c65aa671f594fc63c3adcd74e81407d`; tarball SHA-256
  `183cc29222c3cb58099a9753c43defed5cb8677fda985c3f88e38b5c19a36ff2`) —
  the `g41_master` harness (independent C++ master).
- Outstation under test: the `work` Go outstation via the memory-only
  `lab/ao004/go-outstation` fixture built from commit `807644e` (Go 1.22.12;
  binary SHA-256 `a4a4b25f155af48dc975758c915fc2b52f9f38d1db2c75f31a75c9acfd101a9a`).
  Index 0 only; other indices return `ControlOutOfRange`.
- Lab: `network_mode: none` container (not privileged; caps `NET_ADMIN`,`NET_RAW`,
  `SYS_ADMIN` only; no ports/host mounts). Isolation proven **before any packet**:
  `lo` only and **0** routes; then `vec01-master=10.77.0.1/24` ↔
  `vec01-outstation=10.77.0.2/24` over one veth, **0** default routes,
  `ip_forward=0`, lab-local ping OK, external DNS/route unreachable. Listener
  bound only `10.77.0.2:20000`; DNP3 link master `3` / outstation `4`. Capture
  interface `veth-o`. All packets are observed wire bytes.
- Analysis: independent Python decoder (TCP reassembly, DNP3 CRC-16 validation
  poly `0x3D65`, TL/APDU and G41 record decode).

## Files

| File | Notes |
|------|-------|
| `ao004-negative-20261113-{v1i0,v2i0,v3i0,v4i0}.pcap` | positive controls, index 0 |
| `ao004-negative-20261113-{v1i1,v2i1,v3i1,v4i1}.pcap` | non-success controls, index 1 |
| `ao004-negative-20261113-refused.pcap` | peer unavailable before send (SYN + RST/ACK, no DNP3 payload) |
| `ao004-neg-20261113-*.vec` | packet-order link frames per case |
| `SHA256SUMS.txt` | SHA-256 of every file above |

## Verified wire results (independent-wire, pinned OpenDNP3 3.1.2 → Go outstation)

| Case | Master point result | Fixture writes | Wire echo |
|------|---------------------|----------------|-----------|
| G41V1 idx0 | `state=SUCCESS status=SUCCESS` | 1 (`int32:2300`) | `G41 V1 idx=0 val=fc080000 st=0x00`, IIN `0000` |
| G41V2 idx0 | `state=SUCCESS status=SUCCESS` | 1 (`int16:12345`) | `G41 V2 idx=0 val=3930 st=0x00`, IIN `0000` |
| G41V3 idx0 | `state=SUCCESS status=SUCCESS` | 1 (`float32:1.5`) | `G41 V3 idx=0 val=0000c03f st=0x00`, IIN `0000` |
| G41V4 idx0 | `state=SUCCESS status=SUCCESS` | 1 (`float64:1.5`) | `G41 V4 idx=0 val=000000000000f83f st=0x00`, IIN `0000` |
| G41V1 idx1 | `state=SUCCESS status=OUT_OF_RANGE` | 0 | `G41 V1 idx=1 val=fc080000 st=0x0C`, IIN `0000` |
| G41V2 idx1 | `state=SUCCESS status=OUT_OF_RANGE` | 0 | `G41 V2 idx=1 val=3930 st=0x0C`, IIN `0000` |
| G41V3 idx1 | `state=SUCCESS status=OUT_OF_RANGE` | 0 | `G41 V3 idx=1 val=0000c03f st=0x0C`, IIN `0000` |
| G41V4 idx1 | `state=SUCCESS status=OUT_OF_RANGE` | 0 | `G41 V4 idx=1 val=000000000000f83f st=0x0C`, IIN `0000` |
| peer-unavailable | `summary=FAILURE_NO_COMMS state=INIT status=UNDEFINED` | 0 | **zero** DNP3 link frames (SYN → RST/ACK only) |

- All committed `.vec` link frames: header and payload CRC **valid**.
- The refused case proves the pre-send failure is classified as a communication
  failure with **zero** command frames and **zero** handler writes — not a
  control-status echo and not a fabricated success.

## Malformed-G41 and timeout/quarantine: why independent-wire is NOT_RUN

The pinned OpenDNP3 3.1.2 public master API constructs only well-formed
`CommandSet` objects, so it cannot emit a truncated / wrong-width / wrong-variation
/ bad-qualifier / count≠1 / extra-object / invalid-index / non-finite request on
the wire; packet injection is disallowed by the authorization and no existing
public master API exists for it. Mid-command response-loss/disconnect requires a
controlled fault injector with an observable *single-shot* timing proof, which is
not available with the container's tooling (no `iptables`/`tc`-based deterministic
one-shot drop, and `tc` without netem/iptables cannot prove single-shot
ambiguity). Those cases are therefore recorded as independent-wire **`NOT_RUN`**
and are covered instead by the existing Go **unit** suites (UNIT_ONLY), which the
authorization permits citing without modification:

- malformed/truncated/wrong-width/wrong-variation/count 0 & >1/extra object/bad
  qualifier: `internal/outstation` `TestOutstationG41*` (incl.
  `TestOutstationG41V1F3CountValidation`, `TestOutstationG41RejectionsNoHandlerCall`,
  `TestOutstationG41V4WritePathRejectsNonFiniteAndTruncation`) — PASS.
- invalid index: `TestOutstationG41InvalidIndex*` (Go side) + independent-wire
  OUT_OF_RANGE above.
- non-finite V3/V4: `TestOutstationG41V3RejectsNonFinite`,
  `TestOutstationG41V4RejectsNonFinite` — PASS.
- timeout / disconnect / quarantine / no-blind-retry / fail-closed:
  `internal/master` `TestControl*` (incl. `TestControlSingleShotNoRetryOnMissingResponse`,
  `TestControlSingleShotNoRetryOnDisconnect`, `TestControlQuarantineBlocksFurtherControls`,
  `TestControlPreSendFailureDoesNotQuarantine`) — PASS.

## Limits

Isolated simulated-lab coverage only. Does **not** establish IEEE 1815
conformance, production SCADA readiness, or third-party interoperability. G41
remains **UNSUPPORTED**. No secrets, credentials, tokens, private-network or
field/OT data are included; the captures contain only the synthetic memory-only
lab exchange.
