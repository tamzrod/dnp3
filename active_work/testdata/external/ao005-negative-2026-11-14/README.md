# AO-005 negative-interoperability evidence bundle (2026-11-14)

Owner-authorized bounded AO-005 **negative interoperability** iteration: malformed
Group 41 Direct-Operate request objects (independent raw injector) and after-send
fault injection (fault proxy + opted-in Go master). Isolated lab only. Group 41
remains **UNSUPPORTED**.

This directory holds original raw packet captures plus the runner/endpoint logs
and the independent decode output. It complements (does not replace) the earlier
`ao005-pcap-2026-11-14` positive-path bundle.

## Layout

```
raw/     ao005-neg-A.pcap, ao005-neg-B.pcap + SHA256SUMS
logs/    runner/outstation/peer/proxy/tcpdump/routes logs + SHA256SUMS
decoded/ decoded-A.txt, decoded-B.txt + SHA256SUMS
```

## Matrix A — malformed G41 requests (independent injector -> Go outstation)

`lab/ao004/injector` builds DNP3 link frames from IEEE 1815 first principles (no
project encoding code) and sends one crafted DirectOperate per fresh TCP
connection over the isolated two-namespace veth. Each malformed object is
rejected with a solicited RESPONSE carrying IIN FuncUnknown (`0x0100`) and **no**
Group 41 echo; the outstation accepted **zero** handler writes for all malformed
cases (only the four positive controls wrote). Positive controls and the
out-of-range index behave exactly as in the positive bundle. A link-layer
header-CRC-corrupt frame is silently discarded at the link layer (no APDU).

`decoded/decoded-A.txt` records the literal request/response hex per connection.

## Matrix B — after-send fault injection (Go master -> pinned OpenDNP3 3.1.2 peer)

`lab/ao004/faultproxy` relays `10.77.0.1:20000 -> 10.77.0.2:20000` and, once a
DirectOperate is on the wire, either withholds the peer response
(`drop-response`) or tears the path down immediately (`close-after-request`).
`lab/ao004/neg-master` drives the existing public Go master client.

| Case | Outcome |
|---|---|
| `presend-notconn` | Operate before Connect -> definite pre-send failure, not ambiguous, no quarantine |
| `pass` | proxy transparent: control applied once, `status=success` |
| `presend-cancel` | cancelled ctx before send -> definite pre-send failure; follow-up Operate admitted (not quarantined) |
| `after-send-drop-response` | response withheld -> `control outcome unknown`, quarantine; second Operate refused |
| `after-send-close-after-request` | disconnect after send -> `control outcome unknown`, quarantine; second Operate refused |

The peer accepted **exactly four** writes (the four `pass`/`presend-cancel`
follow-ups); the after-send cases wrote exactly once each and were never blindly
retried. `decoded/decoded-B.txt` records the wire frames per connection
(`drop-response`/`close-after-request` connections have no response frame).

## Provenance

| Item | Value |
|---|---|
| `raw/ao005-neg-A.pcap` | `sha256:e6826eb9590c12e81637b837a94ca1e787abca1cef5c0323e11f6fd9ba8e6373` |
| `raw/ao005-neg-B.pcap` | `sha256:5540a10551896071156b5acd50f592d429aace90f897503d6494381b850dab80` |
| Harness image | `ao004-g41-peer:3.1.2` @ `sha256:f0ab3ced08fb9e783e168d79358e7eecb12ca160de478b2001bbd3a366383b68` |
| OpenDNP3 source | tag `3.1.2`, archive `sha256:183cc29222c3cb58099a9753c43defed5cb8677fda985c3f88e38b5c19a36ff2` (unmodified `lab/vec01/acquire-pinned.sh`) |
| Go toolchain | `1.22.12` |
| `lab/ao004/injector/main.go` | `d768350122a56bf0c9d887de1d46748a68f65aa05f15016464371469a61903fd` |
| `lab/ao004/faultproxy/main.go` | `8b6ab2a257f7ee6231bd6439a489bf56d9606d2d86a884af3c0f00b0716bc2d9` |
| `lab/ao004/neg-master/main.go` | `ddd33fca764ab358b323ee346fd88d33a7887e708723b1b3ec5bd4feaf01b8c4` |
| `lab/ao004/ao005-negative-runner.sh` | `57f402f99500bfbc9a20e037e648cd6ce87bdd789f5e5a0828c53fb31caea3e9` |
| `lab/ao004/ao005-negative-decode.py` | `6de99c5bd6c58407c2103f672a1d17e4b1a6a9a68cbe58531be6dc89ce026439` |

## Reproducibility

Build the three Go harness binaries and copy them plus the runner/decoder into a
staging dir, then run inside the `--network none` lab container with that dir
mounted read-only at `/opt/ao005`:

```sh
go build -o staging/ ./lab/ao004/injector ./lab/ao004/faultproxy ./lab/ao004/neg-master
cp lab/ao004/ao005-negative-runner.sh lab/ao004/ao005-negative-decode.py staging/
docker run --rm --network none --cap-add=NET_ADMIN --cap-add=SYS_ADMIN --cap-add=NET_RAW \
  -v "$PWD/staging":/opt/ao005:ro -v "$PWD/out":/out ao004-g41-peer:3.1.2 \
  -c 'MATRIX=A OUT=/out /opt/ao005/ao005-negative-runner.sh'   # then MATRIX=B
python3 lab/ao004/ao005-negative-decode.py out/ao005-neg-A.pcap matrix=A
```

The two namespaces and the container are `--rm`; no host baseline change
persists. No field, NGCP, SCADA or production endpoint is ever contacted; every
harness fails closed on any non-lab address.
