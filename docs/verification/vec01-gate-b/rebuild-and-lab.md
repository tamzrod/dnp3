# Offline rebuild recipe and isolated-lab construction

Everything here is reproducible from the pinned inputs in `provenance.md`. No step in this
packet was executed as protocol traffic; the recipe is provided so a fresh runner/reviewer
can reconstruct the lab.

## 1. Tools (staged outside the Git repo)

`iproute2`, `iputils-ping`, `cmake` (3.31.x used), `g++` (14.2.0 used), `make`, `file`,
`unzip`, `sha256sum`. The observed toolchain was Debian 13 (trixie), x86_64, kernel 6.8.x.

## 2. Retrieve and verify pinned inputs

Retrieve the five inputs from the URLs in `provenance.md` and verify the expected SHA-256
values **before use**. Retry transient `503`/error pages until the bytes match. Then:

```sh
sha256sum opendnp3-3.1.2.tar.gz asio-1-16-0.zip exe4cpp-*.zip ser4cpp-*.zip catch.hpp
# compare to provenance.md
git ls-remote https://github.com/dnp3/opendnp3 'refs/tags/3.1.2*'   # tag/peel check
```

Extract the OpenDNP3 source and unpack asio/exe4cpp/ser4cpp into a local `depsrc/`.

## 3. Hermetic offline build (stock, then lab)

Build **inside a network namespace with no network** (`unshare -n`). Verify the
namespace is offline by **passive local inspection only** — inspect the namespace's own
routes and interfaces, e.g.:

```sh
unshare -n sh -c 'ip -brief addr; ip route; ip -6 route'
```

Expect no default route and no non-loopback external interface. Do **not** probe any
external destination. CMake must run fully disconnected with the
dependency sources supplied locally so nothing is downloaded:

```sh
cmake -S . -B build-offline -G "Unix Makefiles" -DCMAKE_BUILD_TYPE=Release \
  -DDNP3_TLS=OFF -DDNP3_TESTS=OFF -DDNP3_JAVA=OFF -DDNP3_FUZZING=OFF \
  -DDNP3_EXAMPLES=ON -DDNP3_STATIC_LIBS=ON \
  -DFETCHCONTENT_FULLY_DISCONNECTED=ON \
  -DFETCHCONTENT_SOURCE_DIR_ASIO=<asio> \
  -DFETCHCONTENT_SOURCE_DIR_EXE4CPP=<exe4cpp> \
  -DFETCHCONTENT_SOURCE_DIR_SER4CPP=<ser4cpp>

cmake --build build-offline --target outstation-demo -j4
```

Stock expectations: `outstation-demo` `8ae8c841…093bb`, `libopendnp3.a` `e2926474…e0677`.

Then copy `src/opendnp3-3.1.2` → a lab copy, apply the single-file patch in
`lab-config.md`, and rebuild offline. Lab expectations: `outstation-demo` `6657ec80…0da3e7`,
`libopendnp3.a` `25f83aaa…e03b0`. Rebuilds at a different absolute path differ **only** by
the embedded source-path string (see `provenance.md`); do not claim equality with stock.

## 4. Isolated lab topology and objective isolation checks

Two private network namespaces joined by a **single internal veth** only; no host
networking, no bridge, no NAT, no host route, no published port, forwarding disabled,
deny-all lab firewall allowing only loopback and the internal veth.

| Element | Value |
|---------|-------|
| Subnet (non-routed, illustrative) | `10.77.0.0/24` |
| Master namespace | `vec01-master`, `veth-m` = `10.77.0.1/24` |
| Outstation namespace | `vec01-outstation`, `veth-o` = `10.77.0.2/24` |
| Link addresses | master local `3` (source) → outstation `4`; outstation local `4`, remote `3` |
| Prospective outstation bind | `10.77.0.2:20000` (never `0.0.0.0`) |

Objective checks (all must PASS on the **live** lab at sign-off; use passive inspection or
**strictly lab-local non-TCP** probes — no external destination probes):

1. Only the `10.77.0.0/24` route; no external IPv4 route.
2. No IPv6 global route (only `fe80::/64`).
3. `net.ipv4.ip_forward = 0` and `conf.all.forwarding = 0`.
4. Deny-all lab policy (INPUT/FORWARD/OUTPUT DROP) with only lo/veth allows.
5. No NAT/masquerade; no bridge; no published host port.
6. No lab listening sockets before activation.
7. No physical I/O (no `ttyS*`/`ttyUSB*`/`ttyACM*`/`mem`/`port`/GPIO).
8. Lab-local reachability between `10.77.0.1` ↔ `10.77.0.2` only.
9. Host/pod boundary unchanged (inode/routes/firewall).

Any unknown route, unintended egress, forwarding, missing tool or binding uncertainty is
**FAIL / NOT_TESTED**, not a presumed pass.

## 5. Bounded deadlines, watchdog and stop rules (planned; not executed)

- Per-step and whole-campaign **wall-clock deadlines**; a watchdog kills the run on expiry.
- **Single** bounded G12V1 Direct-Operate attempt to the memory-only BO[0]; **no blind
  auto-retry**; if the control outcome is ambiguous, mark `AMBIGUOUS` and never retry.
- Stop immediately, preserve evidence, disable the listener and report `BLOCKED`/`ABORTED`
  on: isolation loss, unknown routing, unexpected peer, unintended physical effect, hang, or
  ambiguous control outcome.
- Capture (raw PCAP) begins **before** TCP connect and ends after close — a **future**,
  separately authorized Gate C/D/E step; **no capture exists in this packet**.

## 6. Reproduction summary (fresh runner)

1. Stage tools; retrieve and verify the five pinned inputs.
2. Hermetic offline stock build (`unshare -n`, `FETCHCONTENT_FULLY_DISCONNECTED=ON`).
3. Copy source → lab copy; apply the `lab-config.md` patch; rebuild offline.
4. Build the two-namespace lab and run the §4 isolation checks (non-TCP only).
5. Apply the static checks (bind `10.77.0.2:20000`, link 4/3, variations, points, handler).
6. Hand the live lab + hashes to the named reviewer for `reviewer-checklist.md`.
