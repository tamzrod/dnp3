# AO-004 isolated OpenDNP3 3.1.2 G41 memory-only peer + master harness + Go outstation fixture

Bounded, owner-authorized (AO-004 authorizations `3b0afec`, `c7deb8a`, `ff3fb7d`) scaffolding for an
independent OpenDNP3 3.1.2 outstation **and master**, plus a memory-only **Go outstation fixture** built on
the existing public Go API. Together these can later exercise the Group 41 (Analog Output)
Direct-Operate-with-response subset: **G41V1 int32, G41V2 int16, G41V3 float32, G41V4 float64** — in both
directions (C++ master → Go outstation, and Go master → C++ peer).

This directory is a **test harness only** and starts **no** DNP3 traffic by default. The C++ executables
are compiled against the sha256-pinned OpenDNP3 3.1.2 in a build-only container and the Go fixture is built
with the pinned Go 1.22.12 toolchain (see Status below), but no protocol exchange has been run: readiness
is **not** interoperability acceptance.

## Scope and hard boundaries

- Source modification is limited to this independent OpenDNP3 **test harness** — never the Go DNP3
  implementation. Exact file allowlist (this directory only):
  `README.md`, `Dockerfile`, `compose.yaml`, `G41MemoryCommandHandler.h`, `G41MemoryCommandHandler.cpp`,
  `g41_peer.cpp`, `G41MasterCommandCallback.h`, `G41MasterCommandCallback.cpp`, `g41_master.cpp`,
  `CMakeLists.txt`, and the Go fixture `go-outstation/main.go`, `go-outstation/main_test.go`.
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
`opendnp3::levels::NORMAL`. The master harness uses the same pinned API: `DNP3Manager::AddTCPClient` →
`IChannel::AddMaster` with `MasterStackConfig{master, link}`, `IMasterOperations::DirectOperate(CommandSet&&,
callback, config)` with `CommandSet`/`WithIndex`, the four `AnalogOutput*` command types, and
`CommandStatus`/`CommandPointState`/`TaskCompletion`. Exact header/enum names were re-verified against the
sha256-pinned source and confirmed by a successful compile+link of both executables.

## C++ master harness (`g41_master`) — compile-only preparation

`g41_master` is the independently implemented OpenDNP3 3.1.2 **master** counterpart to the Go outstation.
It is intended to later, under a **separate** owner authorization, issue single-point G41 Direct Operate
**with response** commands (V1 int32, V2 int16, V3 float32, V4 float64) at index 0 to our Go outstation.

- Link addresses (master): `LocalAddr=3`, `RemoteAddr=4`; target `10.77.0.2:20000`, reachable only inside
  the future isolated `vec01-master` ↔ `vec01-outstation` topology. Only **Direct Operate with response** is
  used — no SelectAndOperate, no DirectOperateNoResponse, no implicit retry.
- Per-point observability comes from `CommandResultCallbackT`: the task `TaskCompletion` summary and every
  `CommandPointResult` (index/state/status) are recorded; the live path waits a bounded time for the
  callback.
- **Default (no arguments) is offline readiness only**: it constructs no channel, opens no socket, starts
  no master session, sends no command, and exits 0 after a bounded feature print. The live `--run` mode
  exists in source but is **NOT authorized** by the microtask that authored it; it refuses any host other
  than `10.77.0.2:20000` and any variation outside 1–4, and requires an explicit `--variation`.
- The Go outstation executable is **never** started by this harness.

## Go outstation fixture (`go-outstation/`) — memory-only preparation

`go-outstation/` is a minimal, memory-only DNP3 outstation built on the **existing public** Go API
(`dnp3/pkg/dnp3/outstation`). It is the reverse target for the C++ master harness: it accepts single-point
G41 Direct Operate **with response** at index 0 (V1 int32, V2 int16, V3 float32, V4 float64) and stores the
accepted command in memory only.

- **Feasibility (source-verified, not assumed):** the Go outstation **does** implement the G41V1–V4 Direct
  Operate path — `internal/outstation` `handleDirectOperate`/`executeControl` decode the typed value and
  call the data writer, and the public `WriteAnalogOutput` bridge maps it to
  `CommandHandler.HandleAnalogCommand` with the typed `AnalogInt32/Int16/Float32/Float64CommandValue`
  carriers. So **no** `GO_OUTSTATION_G41_VARIATION_UNSUPPORTED` gap exists; the fixture only wires the
  public API. No `internal/`/`pkg/` change was made or is needed.
- Link addresses: outstation = 4, master = 3; endpoint `10.77.0.2:20000`. Index 0 only; any other index
  returns the explicit non-success `ControlOutOfRange`. Binary (G12V1) control is rejected explicitly.
- **Default (no arguments) is offline readiness only**: it constructs no server, opens no socket, starts no
  session and sends no DNP3 traffic; it prints a bounded readiness summary and exits 0. The live `--listen`
  mode exists in source but is **NOT authorized**; it refuses any bind other than exactly `10.77.0.2:20000`
  (fail-closed, before constructing a server).
- Build/test are offline: `go build ./lab/ao004/go-outstation/`, `go test ./lab/ao004/go-outstation/`.
  Fixture unit tests cover the accept/reject paths (index policy, per-variation typed carriers, signed-zero
  bit preservation, non-finite rejection) with **no** TCP.

## Status (this microtask)

- **ISOLATION:** `ISOLATED_RUNNER_UNAVAILABLE` / `TOPOLOGY_NOT_RUN` — no owner-provided isolated Docker
  host/runner is reachable and this session runs on the routed JR Kubernetes pod (explicitly prohibited as
  a lab); `ip`/`nsenter` for netns are not available on the pod, so the two-namespace topology was **not**
  created and no configuration-only topology claim is made.
- **GO_FIXTURE_BUILD_PASS** — the Go fixture builds with the pinned Go 1.22.12 toolchain; its offline unit
  tests pass; `gofmt`/`go vet` clean; the no-argument readiness path exits 0 with no listening socket.
- **REVERSE_HARNESS_COMPILE_PASS** — both `g41_peer` and `g41_master` were (re)compiled and linked against
  the sha256-pinned OpenDNP3 3.1.2 in a build-only Docker container (Debian 12, g++ 12.2.0); the master's
  no-argument readiness mode was run under `--network none` with no sockets/ports and exits 0. The C++
  `--value` help discrepancy was corrected (fixed typed defaults; no `--value` option).
- **REVERSE_G41_INTEROP_NOT_RUN** — no protocol execution authorized; no success/error claim is made for
  G41 wire traffic, float signed-zero/subnormal handling, command statuses, connect/reconnect, or Go
  outstation runtime.

See `docs/verification/AO-004.md` for the full evidence record.

## Usage (only on an authorized isolated runner)

```sh
# Build (build-time network; sha256-gated acquisition) — AUTHORIZED RUNNER ONLY
docker compose -f lab/ao004/compose.yaml build

# Default: readiness echo only (no listener, no DNP3 traffic)
docker compose -f lab/ao004/compose.yaml run --rm lab

# Master readiness echo (no channel, no session, no command)
docker compose -f lab/ao004/compose.yaml run --rm master

# Later, separately authorized protocol step (NOT authorized now):
# docker compose -f lab/ao004/compose.yaml run --rm lab /opt/ao004/g41_peer
# docker compose -f lab/ao004/compose.yaml run --rm master \
#     /opt/ao004/g41_master --run --variation 1

# Go outstation fixture (offline; no listener, no DNP3 traffic)
go build ./lab/ao004/go-outstation/
go test  ./lab/ao004/go-outstation/
go run   ./lab/ao004/go-outstation/            # readiness echo only
# Later, separately authorized live step (NOT authorized now):
# go run ./lab/ao004/go-outstation/ --listen    # binds 10.77.0.2:20000 only
```

## Non-claims

Not AO-004 interoperability evidence, not R4 closure, not IEEE 1815 conformance, not third-party
interoperability proof, not production/OT readiness. G41 remains **UNSUPPORTED** in
`active_work/supported-profile.md`.
