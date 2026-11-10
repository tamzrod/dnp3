# VEC-01 isolated DNP3 interoperability lab (readiness scaffolding)

Bounded lab scaffolding for the owner-authorized VEC-01 campaign: native Go
master (link address 3) <-> independently implemented **OpenDNP3 3.1.2**
outstation (link address 4) over the v0 subset, inside one isolated container.

This directory is **lab scaffolding and readiness only**. It starts no DNP3
protocol traffic and issues no control by default. It does not change library
code, the supported profile, tests, CI, or governance. Readiness is **not**
interoperability acceptance.

## Isolation model

- `compose.yaml` runs **one** service with `network_mode: none` — no `ports`, no
  host networking, no default bridge.
- `lab-netns.sh` builds, **inside the container**, the objective Gate B topology:
  two private network namespaces joined by exactly one internal veth pair.
  - `vec01-master` `10.77.0.1/24`
  - `vec01-outstation` `10.77.0.2/24` (listener restricted to `10.77.0.2:20000`)
  - no default route, forwarding forced off (fail-closed), no DNS/egress.
- `preflight.sh` proves that topology objectively and fails closed otherwise.

## Privileges (least-required)

| Capability | Why |
|------------|-----|
| `NET_ADMIN` | create veth, assign addresses, manage routes |
| `SYS_ADMIN` | `ip netns` needs the `/run/netns` mount (`unshare`/`mount`) |
| `NET_RAW` | `tcpdump` capture on the internal veth |

Not `privileged: true`; no host PID namespace, no host devices, no Docker
daemon socket. `preflight.sh` verifies no `/var/run/docker.sock` is present.

## Pinned inputs (build time, sha256-verified)

`acquire-pinned.sh` verifies and installs, under `/opt`:

- Go `1.22.12` (sha256 `4fa4f869…b34d43`)
- OpenDNP3 `3.1.2` source (sha256 `183cc292…a36ff2`; commit
  `26b4c01e4839bbbda8866655e086471c4917ee53`)
- asio `asio-1-16-0` (`18883c35…e928c5`), exe4cpp `fb878a4…` (`2a8a2c7d…db6c`),
  ser4cpp `3c44973…` (`5a78f83d…762f6e`)

The runtime container is offline; all downloads happen at build time only.

## Usage

```sh
# Build the lab image (build-time network; sha256-gated acquisition).
docker compose -f lab/vec01/compose.yaml build

# Gate B readiness preflight (default entrypoint; no DNP3 traffic).
docker compose -f lab/vec01/compose.yaml run --rm lab

# Persistent capture directory (survives container recreation):
export VEC01_CAPTURE_DIR=/path/to/persistent/store   # -> /capture in the lab
```

A later, separately authorized `operation cwal` may invoke explicit protocol
steps as `docker compose -f lab/vec01/compose.yaml run --rm lab <command>`.

## Non-claims

Not Gate B/C/D/E acceptance, not R4 closure, not IEEE 1815 conformance or
third-party interoperability proof, not production/OT readiness. See
`docs/verification/VEC-01-LAB-EXEC-001.md`.
