# AO-008 fresh independent-lab evidence bundle (2026-11-19)

Fresh, this-invocation reproduction of the G41V1–V4 independent-wire evidence on
the pinned OpenDNP3 3.1.2 lab image, gathered by Operation CWAL during the
AO-008 Phase B implementation. Isolated `--network none` lab only. Group 41
remains **UNSUPPORTED** in the published profile.

This bundle is the Phase B re-verification of the AO-006/AO-007 evidence after
the stable `WithAnalogControl()` opt-in and the explicit legacy-carrier
rejection were added. It complements (does not replace) `ao005-*` and
`ao006-fresh-2026-11-18/`.

## Environment / provenance

| Item | Value |
|---|---|
| Repository | `tamzrod/dnp3`, branch `work` |
| Go toolchain | `go1.22.12 linux/amd64` (sha256-verified official tarball) |
| Lab image (peer+master) | `ao004-g41-peer:3.1.2` / `ao004-g41-master:3.1.2` |
| Image ID | `sha256:41e0406ef52e0ef33f510cc11d6dbb24e4277bebd2f29a9747d7e16bafadf2fa` |
| Independent peer | OpenDNP3 3.1.2 (sha256-pinned by `lab/vec01/acquire-pinned.sh`) |
| Docker server | 29.9.3 |
| Isolation | `docker run --network none` + private two-namespace veth, forwarding off, no default route |
| Decoders | `lab/ao004/ao005-decode-pcap.py`, `lab/ao004/ao005-negative-decode.py` (IEEE 1815 first-principles, no project code) |

Harness binaries were built this invocation from the unmodified lab sources
(`lab/ao004/{go-master,go-outstation,injector,faultproxy,neg-master}`,
`CGO_ENABLED=0 go build -trimpath`) and mounted read-only at `/opt/ao005`. No
harness source was modified.

## Layout

```
raw/     ao008-dirA.pcap, ao008-dirB.pcap, ao008-neg-A.pcap, ao008-neg-B.pcap + SHA256SUMS
decoded/ cases-dirA.txt, cases-dirB.txt, decoded-dirA.txt, decoded-dirB.txt,
         decoded-neg-A.txt, decoded-neg-B.txt + SHA256SUMS
logs/    runner-A.log, runner-B.log, runner-neg-A.log, runner-neg-B.log,
         float-edge.log + SHA256SUMS
```

## Results

- **Direction A** (C++ OpenDNP3 3.1.2 master → Go outstation): 8/8. V1–V4 index 0
  `status=0x00` SUCCESS, index 1 `status=0x0c` OUT_OF_RANGE; exact echoed bits
  (`V3 0x3fc00000`, `V4 0x3ff8000000000000`); IIN `0x0000`; 0 BAD header/block
  CRCs.
- **Direction B** (Go master → C++ OpenDNP3 3.1.2 outstation peer): 8/8; exact
  bits; `status=0x00`/`0x0c`; IIN `0x0090`; 0 BAD CRCs.
- **Matrix A** (malformed G41 → Go outstation): 4 positive controls SUCCESS with
  exactly 4 host writes; every malformed object rejected with IIN `0x0100` and
  zero host writes; idx1/idx2 OUT_OF_RANGE; the header-CRC-corrupt frame got only
  a link ACK and no application response (discarded).
- **Matrix B** (after-send fault injection → C++ peer): pre-send failures
  (`not connected`, `context canceled`) leave no quarantine; after-send
  `drop-response` and `close-after-request` yield `control outcome unknown` +
  quarantine and the second control is refused — never blindly retried. The peer
  was written exactly once per after-send case (4 accepted writes total).
- **Float edge** (`logs/float-edge.log`): V3/V4 `-0.0` and the smallest subnormal
  are normalized to `+0.0` (bits `0x0`) by the OpenDNP3 3.1.2 peer; the Go master
  fails closed (`no parseable command status`). `V3 +1.5` control succeeds with
  bits `0x3fc00000`. Classified **FAIL/DIVERGENCE** (peer-side serialization);
  Go does not accept the normalized echo.

## Limits

- Single independent stack/version (OpenDNP3 3.1.2). No other peer, no
  generalized IEEE 1815 conformance, no field/SCADA/production claim.
- Reverse-direction injected-fault negative remains **NOT_RUN** (unchanged from
  AO-006).
- The negative decoders stop at the deliberately truncated frame by design; the
  literal on-wire request hex is preserved for every case.
- Profile remains **G41 UNSUPPORTED**; the stable opt-in is implemented but not
  published as supported.
