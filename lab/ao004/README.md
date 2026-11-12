# AO-004 isolated OpenDNP3 3.1.2 G41 memory-only peer harness (scaffolding)

Bounded, owner-authorized (AO-004 authorization `3b0afec`) **readiness-construction** scaffolding for
an independent OpenDNP3 3.1.2 outstation that can later serve the Go master's Group 41 (Analog Output)
Direct-Operate-with-response subset: **G41V1 int32, G41V2 int16, G41V3 float32, G41V4 float64**.

This directory is **static scaffolding only**. It is **NOT_BUILT** and **NOT_RUN** in the environment
where it was authored (no Docker daemon, no pinned peer source, no CMake). It starts **no** listener, no
DNP3 traffic and no control. Readiness is **not** interoperability acceptance.

## Scope and hard boundaries

- Source modification is limited to this independent OpenDNP3 **test harness** — never the Go DNP3
  implementation. Exact file allowlist (this directory only):
  `README.md`, `Dockerfile`, `compose.yaml`, `G41MemoryCommandHandler.h`, `G41MemoryCommandHandler.cpp`,
  `g41_peer.cpp`, `CMakeLists.txt`.
- Handler is **memory-only**: explicit fixed index 0, no real output, no OS I/O, no NGCP/SCADA mapping,
  no dynamic points. Only **Direct Operate with response** is enabled; SELECT/OPERATE/SBO and
  DirectOperateNoResponse are rejected.
- Non-finite floats (NaN/±Inf) are rejected **before** any state mutation, matching the Go finite-only
  policy. Exact IEEE-754 bits (±0, subnormals) are preserved and recorded.
- No protocol exchange, no peer connection and no command is authorized until a later distinct
  owner-authorized protocol-execution microtask (see `docs/verification/AO-004.md`).

## Isolation model (mirrors the approved `lab/vec01/` reference)

`compose.yaml` runs **one** service with `network_mode: none` — no `ports`, no host networking, no
default bridge. Least-required capabilities only (`NET_ADMIN`, `SYS_ADMIN`, `NET_RAW`) and **not**
`privileged`. No host devices, no Docker socket, no host network mounts. The default `CMD` is a
**readiness echo only**; the peer listener starts only on an explicit, separately authorized invocation.

Two distinct things, not one:

- `network_mode: none` is the safe container **baseline**: only loopback exists. It does **not** create
  the lab address `10.77.0.2`, so `g41_peer --listen` cannot bind until a topology exists.
- The functional **two-namespace topology** (veth pair `10.77.0.1` ↔ `10.77.0.2`, forwarding forced off,
  no default route) is built by the existing, **unmodified** `lab/vec01/lab-netns.sh` and checked by
  `lab/vec01/preflight.sh`. Building it is a later, explicitly authorized, isolated-runner-only step and
  is **not** run by the default command. Static YAML alone proves **no** isolation; `network_mode: none`
  is not a `NETWORK_ISOLATION_PASS` claim.

## Pinned inputs (build time only, sha256-verified)

`Dockerfile` reuses the existing verified `lab/vec01/acquire-pinned.sh` **unmodified** (installs under
`/opt/vec01`): Go `1.22.12`, OpenDNP3 `3.1.2` (commit `26b4c01e4839bbbda8866655e086471c4917ee53`), asio
`asio-1-16-0`, exe4cpp `fb878a4d…`, ser4cpp `3c449734…`. The runtime container is offline; all downloads
happen at image-build time and are hash-gated. `CMakeLists.txt` binds `FETCHCONTENT_SOURCE_DIR_{ASIO,
EXE4CPP,SER4CPP}` to those verified extracted directories with
`FETCHCONTENT_FULLY_DISCONNECTED=ON` and fails closed if any is missing — no implicit download.

`BASE_IMAGE_UNPINNED`: the `debian:12-slim` base tag and APT packages are not version-pinned, so
full-image reproducibility is **not** proved by the sha256-pinned OpenDNP3 input alone. This remains an
unresolved limitation (no digest invented).

## Pinned API note

The handler/peer were corrected against the pinned OpenDNP3 3.1.2 public API: `ICommandHandler` declares
ten pure virtual hooks (CROB + four analog types), all implemented here; `CommandStatus` lives in
`namespace opendnp3`; the endpoint header is `opendnp3/channel/IPEndpoint.h`; and `levels::NORMAL` is
`opendnp3::levels::NORMAL`. Exact header/enum names are still to be re-verified against the pinned
source on an authorized build runner before any compile.

## Status (this microtask)

- **ISOLATION:** `ISOLATED_RUNNER_UNAVAILABLE` — no owner-provided isolated Docker host/runner is
  reachable; the routed JR Kubernetes pod is explicitly prohibited as a lab.
- **HARNESS_BUILD:** `NOT_BUILT` — no CMake and no pinned peer source in the authoring environment.
- **INTEROP:** `NOT_RUN` — no protocol execution authorized.

See `docs/verification/AO-004.md` for the full evidence record.

## Usage (only on an authorized isolated runner)

```sh
# Build (build-time network; sha256-gated acquisition) — AUTHORIZED RUNNER ONLY
docker compose -f lab/ao004/compose.yaml build

# Default: readiness echo only (no listener, no DNP3 traffic)
docker compose -f lab/ao004/compose.yaml run --rm lab

# Later, separately authorized protocol step (NOT authorized now):
# docker compose -f lab/ao004/compose.yaml run --rm lab /opt/ao004/g41_peer
```

## Non-claims

Not AO-004 interoperability evidence, not R4 closure, not IEEE 1815 conformance, not third-party
interoperability proof, not production/OT readiness. G41 remains **UNSUPPORTED** in
`active_work/supported-profile.md`.
