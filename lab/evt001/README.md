# EVT-001 isolated OpenDNP3 3.1.2 event harness

This directory builds two small OpenDNP3 3.1.2 programs and two Go programs.
They exchange memory-only binary, analog, and counter events on 127.0.0.1.
Nothing here talks to field devices, physical I/O, NGCP, or SCADA.

The OpenDNP3 sources are not vendored. The recorded run used tag 3.1.2,
commit `26b4c01e4839bbbda8866655e086471c4917ee53`, tarball sha256
`183cc29222c3cb58099a9753c43defed5cb8677fda985c3f88e38b5c19a36ff2`.
Offline dependencies, also hash-checked before the build:

- asio 1.16.0, sha256 `18883c350558eadf91218ae5ce9a8f8f85099c2f036c1c8579fb639092e928c5`
- exe4cpp `fb878a4de598ba9d6e4338afebf83f96e03af1b8`, sha256 `2a8a2c7d9982f54858b92c03cff91ef433907472039f8e61ff7c9f9b3d48db6c`
- ser4cpp `3c449734dc530a8f465eb0982de29165cc4e23d5`, sha256 `5a78f83de2d70bdf4be2ab607bc4416c6866ce64a257a8f19948cf5746762f6e`

`CMakeLists.txt` sets `FETCHCONTENT_FULLY_DISCONNECTED` and refuses to
configure when those directories are missing. `image.Dockerfile` is the
context Dockerfile: `src/opendnp3-3.1.2`, `deps/`, `harness/` (this
directory), `derive_pcap.py`, and `run-isolated.sh`.

## Isolation

The recorded run used `docker run --network none` with `NET_RAW` and
`NET_ADMIN` only so tcpdump can capture loopback. No published ports and no
host network. `run-isolated.sh` exits 2 before any DNP3 exchange unless
`ip -4 route` has no default route and a TCP connect to `1.1.1.1:443` fails.

The library image used for that run was `evt001-opendnp3:3.1.2`, image id
`sha256:5ec7c3898e92e7d1c10247a5db62942aa3a67e48197b5f8f01583b53fef2ff92`.
The passing `evt_master` and `evt_outstation` were recompiled from this
directory inside that image and mounted for the run. The Go programs were
built on the host with `CGO_ENABLED=0` from the same tree and mounted at
`/opt/evt/bin`.

```bash
docker run --rm --network none --cap-add NET_RAW --cap-add NET_ADMIN \
  -e MASTER=/opt/evt/bin/evt_master \
  -e OUTSTATION=/opt/evt/bin/evt_outstation \
  -v "$OUT":/out \
  -v "$BIN":/opt/evt/bin:ro \
  evt001-opendnp3:3.1.2
```

Addresses are master 1 and outstation 1024. Ports are 20000 and 20001 on
127.0.0.1 only. Points are memory values. OpenDNP3 `DNPTime` is given the
IEEE millisecond count since 2000-01-01, because that library writes the
stored count as a raw 48-bit field.

## What the positive run checks

`run-isolated.sh` is the positive exchange. Direction A is the Go outstation
and the OpenDNP3 master: class 1/2/3 events, absolute timestamps, an empty
second read, unsolicited disabled, unsolicited enabled, confirm removal, and
a bounded overflow with class and overflow IIN bits.

Direction B is the OpenDNP3 outstation and the Go master: class read,
timestamp preservation, a second empty read, unsolicited disabled, then
enabled, and confirm removal.

OpenDNP3 confirms every fragment that asks for confirmation, and it does so
once. `run-isolated.sh` does not inject a second confirm or drop one. The
positive capture stays `docs/verification/evt001/evt001.pcap`.

Class reads use `AddClassScan` with a 30 second period. OpenDNP3 3.1.2 drops
a one-shot `ScanClasses` issued before the master link is online. The bound
scan starts when the link comes up and does not repeat during these runs.

## Negative run

`run-negative.sh` uses the same image, isolation gate, addresses, and
OpenDNP3 master. `confirm_proxy.py` is a loopback TCP proxy. The master
connects to the proxy. The proxy connects to the Go outstation.

Stock OpenDNP3 still creates each application CONFIRM. The proxy does not
build one.

Duplicate mode holds the first confirm whose application sequence matches
the unsolicited fragment just forwarded. When `/tmp/evt-proxy-release`
appears, it writes that same link frame to the outstation twice. Later
confirms are forwarded once. The second write is the injection.

Withhold mode does not write application CONFIRM frames toward the
outstation. Those frames stay on the master-to-proxy leg. After the
outstation has sent the original fragment and two identical retries, a new
OpenDNP3 master reads the retained event on the outstation port.

`check_negative.py` is the pass/fail read of those captures. It separates
the master-to-proxy leg from the proxy-to-outstation leg by TCP port.
`derive_pcap.py` counts both legs together and is not that decision.

Evidence is in `docs/verification/EVT-001-OPENDNP3.md`.
