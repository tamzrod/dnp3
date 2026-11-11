# VEC-01 post-corrective isolated OpenDNP3 3.1.2 evidence (2026-11-11)

Durable, non-sensitive, isolated-lab evidence for the owner-authorized
post-corrective re-test of the `work` Go master against pinned, independently
implemented C++ **OpenDNP3 3.1.2**. Full narrative: `docs/verification/VEC-01-LAB-EXEC-001.md` §50–§54.

## Provenance

- Peer: OpenDNP3 3.1.2, annotated tag `3.1.2` = `523f7d2d6c65aa671f594fc63c3adcd74e81407d`,
  peeled commit `26b4c01e4839bbbda8866655e086471c4917ee53`, tarball SHA-256
  `183cc29222c3cb58099a9753c43defed5cb8677fda985c3f88e38b5c19a36ff2`.
- Lab: `network_mode: none` Docker container; two private namespaces
  `vec01-master` (`10.77.0.1/24`) and `vec01-outstation` (`10.77.0.2/24`) joined by one lab-only
  veth; no default route, forwarding, DNS/OT path, NAT, bridge, published port, host ingress or
  physical I/O. DNP3 link addresses master `3` / outstation `4`. TCP port `20000`.
- Memory-only simulated `BO[0]` CROB handler; static variations `G1V1`/`G30V1`/`G20V1` at index 0;
  single-shot non-retrying control; bounded deadlines.

## Files

| File | Run | Notes |
|------|-----|-------|
| `vec01-20261111-positive-class0-crob.pcap` | Gate C | Original lab-only PCAP, `tcpdump` on `veth-o`, before connect through close |
| `vec01-20261111-positive-class0-crob.vec` | Gate C | Packet-derived chronological link frames (per `../FORMAT.md`) |
| `vec01-20261111-negative-payloadcrc.pcap` | Gate D | Original negative PCAP, `tcpdump` on master-ns `lo`, before connect through close |
| `vec01-20261111-negative-payloadcrc.vec` | Gate D | Packet-derived link frames; last payload-block CRC intentionally corrupted |
| `SHA256SUMS.txt` | — | SHA-256 of every file above |

## Verified results (summary)

- **Gate C PASS** — Class-0 G60V1 READ: `BI[0]=true`, `AI[0]=2300` ONLINE, `CT[0]=42` ONLINE; one
  G12V1 Direct-Operate CROB at `BO[0]` LATCH_ON with protocol status SUCCESS (`0`) and peer
  in-memory readback `state=true`. Independently decoded with tshark 4.4.19. No blind retry
  (exactly one READ, one control). Out-of-scope: peer also served `BI[1..9]` as G1V2, which the
  master misreads — recorded `G1V2_DECODER_DEFECT_CLASSIFIED_OUT_OF_SCOPE`, not fixed, not claimed.
- **Gate D PASS** — exactly one corrupted **payload-block** CRC on an O->M response (last block
  `f3b9`→`0c46`), link **header CRC valid**; master failed closed (`operation deadline exceeded`,
  rc=1, ~5.25 s), accepted no points, reported no false success, sent **no control**, and did not hang.

Independent CRC-16/DNP validation of the derived `.vec` bytes: positive = all header and payload
CRCs valid; negative = all valid except exactly one payload block (the last). All bytes are observed
wire bytes; none are synthetic or protocol-code-generated.

## Limits

- This is an isolated simulated-lab interoperability re-test. It does **not** establish IEEE 1815
  conformance or production SCADA readiness, and independent third-party interoperability remains
  **NOT PROVEN**. `Status: TASK_BLOCKED` and R4 remain open.
- No secrets, credentials, live/production/OT traffic, physical-I/O or third-party binary archives
  are included.
