# VEC-01 OpenDNP3 lab deployment readiness

**NO PROTOCOL EXECUTION.** This document records the static-source, dependency,
isolation and build evidence for an isolated, independently sourced **OpenDNP3
3.1.2** C++ outstation build artifact. No DNP3 listener was started, no DNP3
master/outstation session was opened, no protocol frame was transmitted, no
traffic was captured, and no control was issued. **VEC-01 NOT EXECUTED / R4
OPEN / THIRD-PARTY INTEROPERABILITY NOT PROVEN.**

- Task: VEC-01-LAB-DEPLOY-001 (human-promoted 2026-11-08)
- Promotion baseline: `f21f32cf5d0631c8f50e6b3f398bffe975bbb6ab` (merged PR #20), branch `work`
- Lab workspace (outside the Git checkout, disposable): `/tmp/vec01-lab`
- Repo checkout `work` remained clean throughout; no external source, binary or
  artifact was committed to `tamzrod/dnp3`.

---

## 1. Independent source provenance

| Field | Verified value |
|-------|----------------|
| Upstream | `https://github.com/dnp3/opendnp3` (formerly `automatak/dnp3`) |
| License | Apache-2.0 (repo `LICENSE` = Apache License 2.0; `NOTICE` names Green Energy Corp + Automatak LLC) |
| Maintenance | archived upstream (last push 2022-05-18) |
| Pinned tag | `3.1.2` |
| Tag → commit | `refs/tags/3.1.2` = `523f7d2d6c65aa671f594fc63c3adcd74e81407d` (annotated); peeled `refs/tags/3.1.2^{}` = **`26b4c01e4839bbbda8866655e086471c4917ee53`** (matches the task's pinned commit) |
| Lineage | Independently implemented C++ stack, distinct from this native Go repository; not a fork of `tamzrod/dnp3` |
| Acquisition URL | `https://codeload.github.com/dnp3/opendnp3/tar.gz/refs/tags/3.1.2` |
| Acquired archive SHA-256 | `183cc29222c3cb58099a9753c43defed5cb8677fda985c3f88e38b5c19a36ff2` (computed from the bytes actually obtained) |

CMake confirms the source is `3.1.2` (`OPENDNP3_MAJOR/MINOR/MICRO = 3/1/2`,
`cmake_minimum_required(VERSION 3.11)`). No floating `release`/`main`/substitute
was used.

## 2. Dependency provenance and licensing

All four FetchContent dependencies declared in `deps/*.cmake` at commit
`26b4c01` were acquired and hashed. SHA-1 cross-checks match the upstream
declared `URL_HASH` exactly; SHA-256 was computed from the actual bytes.

| Dependency | Pinned version | Upstream URL | Declared SHA-1 | Actual SHA-1 (match) | Actual SHA-256 |
|------------|----------------|--------------|----------------|----------------------|----------------|
| asio | `asio-1-16-0` | `https://github.com/chriskohlhoff/asio/archive/asio-1-16-0.zip` | `6BDD33522D5B95B36445ABB2072A481F7CE15402` | `6bdd33522d5b95b36445abb2072a481f7ce15402` ✅ | `18883c350558eadf91218ae5ce9a8f8f85099c2f036c1c8579fb639092e928c5` |
| exe4cpp | `fb878a4de598ba9d6e4338afebf83f96e03af1b8` | `https://github.com/automatak/exe4cpp/archive/fb878a4de598ba9d6e4338afebf83f96e03af1b8.zip` | `18B141E8CF09DC8D28CC62DD5FA2920670D501BD` | `18b141e8cf09dc8d28cc62dd5fa2920670d501bd` ✅ | `2a8a2c7d9982f54858b92c03cff91ef433907472039f8e61ff7c9f9b3d48db6c` |
| ser4cpp | `3c449734dc530a8f465eb0982de29165cc4e23d5` | `https://github.com/automatak/ser4cpp/archive/3c449734dc530a8f465eb0982de29165cc4e23d5.zip` | `937B759B7CC80180DA26B47037E796B59798A672` | `937b759b7cc80180da26b47037e796b59798a672` ✅ | `5a78f83de2d70bdf4be2ab607bc4416c6866ce64a257a8f19948cf5746762f6e` |
| catch (test-only) | `v2.11.3` | `https://github.com/catchorg/Catch2/releases/download/v2.11.3/catch.hpp` | `C127EBB7A4F65C6CEFF7587C8EF18F84A74D6C15` | `c127ebb7a4f65c6ceff7587c8ef18f84a74d6c15` ✅ | `197640c056d6d01e5a56737fbdfe671d2b872c8c7533e01cc2de4be19ab7a028` |

**License inventory:**

| Component | License / SPDX | Evidence |
|-----------|----------------|----------|
| OpenDNP3 3.1.2 | Apache-2.0 | repo `LICENSE`, `NOTICE` |
| exe4cpp | BSD-3-Clause | `LICENSE` ("BSD 3-Clause License, Copyright (c) 2018, Automatak") |
| ser4cpp | BSD-3-Clause | `LICENSE` ("BSD 3-Clause License, Copyright (c) 2018, Automatak") |
| asio 1.16.0 | Boost Software License 1.0 | `asio/LICENSE_1_0.txt`, `asio/COPYING`; header comment in `asio.hpp` |
| catch v2.11.3 | Boost Software License 1.0 | test-only; **not used** (tests disabled) |

All licenses are permissive and compatible with isolated non-production testing.
No dependency is GPL/AGPL. `catch` is only pulled when `DNP3_TESTS`/`DNP3_FUZZING`
is enabled, which is **off** here.

**Risk note:** the upstream project is archived and dependencies are pinned to
2022-era revisions. This is a documented risk, not a blocker for a static build;
it means no upstream fixes or newer releases are available.

## 3. Hermetic / isolated build

Build performed in a disposable workspace `/tmp/vec01-lab` **outside** the Git
checkout, inside a network namespace with **no network** (`unshare -n`;
verified: no DNS, no route, `curl https://github.com` fails). CMake ran with
`-DFETCHCONTENT_FULLY_DISCONNECTED=ON` and the three dependency sources supplied
locally via `FETCHCONTENT_SOURCE_DIR_*`, so no download occurred during
configure or compile.

Toolchain (recorded, `artifacts/toolchain-versions.txt`): Debian 13 (trixie),
x86_64, kernel 6.8.0-1055-gke, **CMake 3.31.6** (≥ 3.11 required), **g++
14.2.0** (C++14, ≥ C++11), GNU Make 4.4.1, ld (binutils) 2.44, Go 1.22.12.

Configure command (features limited to the native outstation/library):

```
cmake -S src/opendnp3-3.1.2 -B build-offline -G "Unix Makefiles" \
  -DCMAKE_BUILD_TYPE=Release \
  -DDNP3_TLS=OFF -DDNP3_TESTS=OFF -DDNP3_JAVA=OFF -DDNP3_FUZZING=OFF \
  -DDNP3_EXAMPLES=ON -DDNP3_STATIC_LIBS=ON \
  -DFETCHCONTENT_FULLY_DISCONNECTED=ON \
  -DFETCHCONTENT_SOURCE_DIR_ASIO=... -DFETCHCONTENT_SOURCE_DIR_EXE4CPP=... -DFETCHCONTENT_SOURCE_DIR_SER4CPP=...
```

Build: `cmake --build build-offline --target outstation-demo -j4` →
`Built target opendnp3`, `Built target outstation-demo`, exit 0. TLS, tests,
Java, .NET, fuzzing and coverage are off; only the native `opendnp3` static
library and the `outstation-demo` example are built.

**Reproducibility:** a second, fully clean configure+build inside the no-network
namespace produced a **byte-identical** `outstation-demo`
(SHA-256 `8ae8c84125c4920352b6daef4145e2a91de9e02cd6e2c1b94bf6995ebeb093bb`).

## 4. Static-only smoke checks

| Check | Result |
|-------|--------|
| Binary identity (`file`) | ELF 64-bit LSB **pie** executable, x86-64, dynamically linked, not stripped |
| ELF header (`readelf -h`) | Class ELF64; Type DYN (PIE); Machine x86-64 |
| Dynamic dependencies (`ldd`) | `libstdc++.so.6`, `libm.so.6`, `libgcc_s.so.1`, `libc.so.6`, `linux-vdso.so.1`, `/lib64/ld-linux-x86-64.so.2` — standard system libs only; **no DNP3/peer runtime libraries** |
| `NEEDED`/`RPATH`/`RUNPATH` | only the four standard system libs; no RPATH/RUNPATH |
| File mode | `0755` (`-rwxr-xr-x`) |
| `outstation-demo` SHA-256 | `8ae8c84125c4920352b6daef4145e2a91de9e02cd6e2c1b94bf6995ebeb093bb` |
| `libopendnp3.a` SHA-256 | `e2926474288fb8d2c6c7962192bad2ff0f724f63c07d2209fe5b30c26fbe0677` |
| Listener on port 20000 | **none** (checked `/proc/net/tcp`; the only LISTEN sockets are the agent runtime on unrelated ports) |
| `outstation-demo` process | **not running** |

**The demo binary was intentionally NOT executed.** Its `main()`
(`cpp/examples/outstation/main.cpp` @ `26b4c01`) calls
`manager.AddTCPServer("server", ..., IPEndpoint("0.0.0.0", 20000), ...)`, i.e. it
binds a TCP listener. Per the task's static-only smoke gate, a process that
opens/binds a socket must not be started in this task. No version/help probe was
run either, because the upstream demo has no non-networking `--help` path that
is guaranteed not to construct the server.

## 5. Isolated laboratory configuration (STATIC — not activated)

Prepared under `/tmp/vec01-lab/config/` (outside the repo), inspection only:

- Topology: dedicated offline VM / internal-only network namespace.
- Hypothetical subnet `10.77.0.0/24`; TCP port **20000 unpublished**; master
  link address `3`, outstation link address `4`; simulated memory-backed points
  only; no bridging/forwarding/host networking; no route to any OT/SCADA network.
- Upstream demo defaults are `IPEndpoint("0.0.0.0", 20000)`, `LocalAddr = 10`,
  `RemoteAddr = 1`. Proposed lab-only adjustments (bind to the isolated
  interface, addresses 4/3) are recorded as `TO VERIFY` and were **not applied
  or executed**.
- Static inspection only: interfaces (`lo`, `eth0`), no listener on 20000, no
  `outstation-demo` process. No production address was probed.

## 6. Future VEC-01 readiness (NOT execution)

A later, separately owner-promoted task (**VEC-01-LAB-EXEC-001**) is required
before any wire exchange. It would need: master `tamzrod/dnp3` at a recorded
SHA, the outstation above started on the isolated segment only, a Class-0
G60V1 poll, G1V1/G30V1/G20V1 static reads, and a G12V1 Direct-Operate to a
**simulator-only** output. Variation/qualifier behavior, peer command status and
multi-header packing remain `TO VERIFY`. Evidence would follow
`active_work/testdata/external/FORMAT.md` and
`active_work/third-party-capture-procedure.md`. No PCAP/`.vec` is fabricated or
accepted as third-party proof now.

## 7. Non-claims and limitations

- **VEC-01 NOT EXECUTED; R4 OPEN; third-party interoperability NOT PROVEN.**
- No DNP3 listener, protocol packets, peer session, capture, or control occurred.
- The static build artifact proves only that the pinned source compiles in
  isolation; it is **not** evidence of protocol interoperability.
- No IEEE 1815 certification, full-conformance, or production-readiness claim.
- Upstream archived; dependencies pinned to 2022 revisions (documented risk).
- Lab artifacts live outside the repository and were not committed.
