# AO-006 fresh independent-lab evidence bundle (2026-11-18)

Fresh, this-invocation reproduction of the G41V1–V4 independent-wire evidence on
the pinned OpenDNP3 3.1.2 lab image, gathered by Operation CWAL under the
`workflow/active_work/AO-006.md` owner authorization. Isolated `--network none`
lab only. Group 41 remains **UNSUPPORTED** in the published profile.

This bundle complements (does not replace) the earlier AO-005 bundles
(`ao005-pcap-2026-11-14/`, `ao005-negative-2026-11-14/`).

## Layout

```
raw/     ao006-dirA.pcap, ao006-dirB.pcap, ao006-neg-A.pcap, ao006-neg-B.pcap + SHA256SUMS
decoded/ cases-dirA.txt, cases-dirB.txt, decoded-dirA.txt, decoded-dirB.txt,
         decoded-neg-A.txt, decoded-neg-B.txt + SHA256SUMS
logs/    runner-A.log, runner-B.log, runner-neg-A.log, runner-neg-B.log,
         float-edge.log + SHA256SUMS
```

## Results

- Direction A (C++ OpenDNP3 3.1.2 master -> Go outstation): 8/8 (V1–V4 index 0
  SUCCESS, index 1 OUT_OF_RANGE); exact bits; status `0x00`/`0x0c`; IIN `0x0000`;
  0 BAD CRCs.
- Direction B (Go master -> C++ OpenDNP3 3.1.2 outstation peer): 8/8; exact bits;
  status `0x00`/`0x0c`; IIN `0x0090`; 0 BAD CRCs.
- Matrix A (malformed G41 -> Go outstation): 4 positive controls SUCCESS with
  exactly 4 writes; all malformed objects rejected with IIN `0x0100` and zero
  writes; idx1/idx2 OUT_OF_RANGE; header-CRC-corrupt frame discarded.
- Matrix B (after-send fault injection -> C++ peer): pre-send failures no
  quarantine; after-send loss/disconnect -> `control outcome unknown` +
  quarantine + refused second control; peer written exactly once per after-send
  case; never blindly retried.
- Float edge (`logs/float-edge.log`): V3/V4 `-0.0` and smallest subnormal are
  normalized to `+0.0` (bits `0x0`) by the peer; the Go master fails closed.
  Tracked FAIL/DIVERGENCE (peer-side serialization).

## Provenance

| Item | Value |
|---|---|
| Lab image | `ao004-g41-peer:3.1.2` @ `sha256:33ac02d9883133c415eb80bee3282aec329ee5f30c2193c2713ba974d05ffe79` |
| OpenDNP3 source | tag `3.1.2`, archive `sha256:183cc29222c3cb58099a9753c43defed5cb8677fda985c3f88e38b5c19a36ff2` (unmodified `lab/vec01/acquire-pinned.sh`) |
| Go toolchain | `1.22.12` |
| `raw/ao006-dirA.pcap` | `sha256:fa088eb87dd34571202af34d34c017242fcec5fa001efad422f05d95e7dc514a` |
| `raw/ao006-dirB.pcap` | `sha256:b83f01d87068a2a59bccd4f3b769b0ddd7e3d0fff836fc557638a4cddf4b8a21` |
| `raw/ao006-neg-A.pcap` | `sha256:eac2dac28421167741c534759665c118d9dd33370a0f328d97f117244031a5e1` |
| `raw/ao006-neg-B.pcap` | `sha256:fff0abfd44322d25cc79fa3818f298636dfc64586c3e5cfe60ef7003b3957271` |

See `docs/verification/AO-006.md` for the full command/result record and the
promotion-coverage assessment.

## Reproducibility

Build the image from `lab/ao004/Dockerfile`; build the Go harness binaries into a
staging dir (`go build -o staging/ ./lab/ao004/go-master ./lab/ao004/go-outstation
./lab/ao004/injector ./lab/ao004/faultproxy ./lab/ao004/neg-master`); copy the
runners/decoders in; mount the staging dir read-only at `/opt/ao005`; run inside
the `--network none` container with `--cap-add=NET_ADMIN --cap-add=SYS_ADMIN
--cap-add=NET_RAW --entrypoint bash`:

```sh
DIR=A OUT=/out /opt/ao005/ao005-pcap-runner.sh     # then DIR=B
MATRIX=A OUT=/out /opt/ao005/ao005-negative-runner.sh   # then MATRIX=B
python3 /opt/ao005/ao005-decode-pcap.py /out/ao005-dirA.pcap dir=A cases
```

The container and both namespaces are `--rm`; no host baseline change persists.
No field, NGCP, SCADA, or production endpoint is ever contacted; every harness
fails closed on any non-lab address.
