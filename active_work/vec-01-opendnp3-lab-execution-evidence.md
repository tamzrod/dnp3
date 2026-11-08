# VEC-01 OpenDNP3 lab execution — evidence index (TASK_BLOCKED)

Concise index for the isolated master ↔ OpenDNP3 3.1.2 interoperability
campaign (CURRENT `VEC-01-LAB-EXEC-001`). The campaign is **BLOCKED at Gate B**;
nothing was executed on the wire. Full detail and hashes:
`docs/verification/VEC-01-LAB-EXEC-001.md`.

**VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY INTEROPERABILITY NOT PROVEN.**

Durable lab artifact store (outside Git, outside `/tmp`):
`/workspace/vec01-lab-artifacts`.

## Campaign matrix

| Gate | Step | Status | Evidence |
|------|------|--------|----------|
| Startup | markers, `cwal-check.sh`, remote ancestry | PASS | `origin/main=02a58f7` ancestor of `origin/work=f44d28b`; `CWAL CHECK OK` |
| A | retrieve/reproduce pinned peer + deps, durable retention | PASS | tarball `183cc292…`; tag peeled `26b4c01e…`; all dep SHA-256/SHA-1 match; hermetic rebuild byte-identical `8ae8c841…` / `e2926474…` |
| B | named independent reviewer isolation/output sign-off | **BLOCKED** | no named independent reviewer available; no autonomous self-approval permitted |
| C | positive matrix (TCP connect, link reset/status, Class-0 G60V1, G1/G30/G20 decode, G12V1 direct-operate, close) | NOT REACHED | Gate B is a hard precondition |
| D | negatives (bad CRC, wrong address, drop, command reject, incomplete capture) | NOT REACHED | Gate B is a hard precondition |
| E | raw PCAP, faithful `.vec`, SHA-256 manifest, reviewer verdict | NOT REACHED | no capture started; no `.vec` fabricated |
| Go gates | `gofmt`, `git diff --check`, `go vet`, `go build`, `go test`, `-race`, `verify-mvp.sh`, `verify-external-mvp.sh` | `git diff --check` PASS; others NOT_RUN | Go toolchain absent; real-TCP gates also prohibited pre-approval |

## Peer provenance (re-verified from retrieved bytes)

| Item | Value |
|------|-------|
| Upstream | `https://github.com/dnp3/opendnp3` (archived; Apache-2.0) |
| Tag `3.1.2` peeled commit | `26b4c01e4839bbbda8866655e086471c4917ee53` |
| Source tarball SHA-256 | `183cc29222c3cb58099a9753c43defed5cb8677fda985c3f88e38b5c19a36ff2` |
| `outstation-demo` SHA-256 | `8ae8c84125c4920352b6daef4145e2a91de9e02cd6e2c1b94bf6995ebeb093bb` |
| `libopendnp3.a` SHA-256 | `e2926474288fb8d2c6c7962192bad2ff0f724f63c07d2209fe5b30c26fbe0677` |
| Dependencies | asio `asio-1-16-0`, exe4cpp `fb878a4…`, ser4cpp `3c44973…`, catch `v2.11.3` (unused) — all SHA-256/SHA-1 match |

## Honest scoped / non-claims

- The pinned peer was **reproduced**, not executed; this is build evidence only.
- No listener, no TCP session, no DNP3 packet, no capture, no control occurred.
- No interoperability, conformance, certification, or production-readiness claim.
- The blocker is the mandatory **independent reviewer** authority requirement,
  not a technical protocol failure.
- Closing R4, changing `supported-profile.md`, production use, and feature
  expansion each require separate owner approval — none was sought or given.
