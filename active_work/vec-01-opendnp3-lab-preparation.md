# VEC-01 OpenDNP3 peer laboratory preparation plan

**PLANNING ONLY — VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY INTEROPERABILITY NOT PROVEN.**
This document specifies how an independent, explicitly pinned **OpenDNP3** C++
outstation *could later* be prepared for VEC-01 interoperability testing against
the `tamzrod/dnp3` Go master. It is **not** a deployment, an executed test, or a
capture. No peer was cloned, downloaded, built, installed or run to produce it.
It grants no authority to deploy a lab or to execute VEC-01; those require
separate human-promoted tasks (§7).

Baseline: merged PR #19 (`f165e8509436153dfa53ab68bb67c51be9a88b10`), permanent
branch `work`. Profile under test: `active_work/supported-profile.md` (v0,
TCP-only, static G1V1/G30V1/G20V1 + G12V1 Direct-Operate).

---

## 1. Independent peer provenance

**Selected candidate:** OpenDNP3.

| Field | Verified value | Source / status |
|-------|----------------|-----------------|
| Upstream URL | `https://github.com/dnp3/opendnp3` (formerly `automatak/dnp3`) | GitHub repository API (repo id 7551306) |
| Project description | "DNP3 (IEEE-1815) protocol stack. Modern C++ with bindings for .NET and Java." | Upstream repository description |
| Implementation lineage | Independently implemented **C++** stack, distinct from this repository's native **Go** implementation. Not a fork/derivative of `tamzrod/dnp3`. | Upstream language/description; no shared lineage |
| License | **Apache-2.0** (SPDX `Apache-2.0`) | GitHub license metadata for the repository |
| Maintenance status | **ARCHIVED** upstream (`archived: true`); last push `2022-05-18`; no open issues | GitHub repository API |
| Pinned release | Tag **3.1.2**, published `2022-04-22` | GitHub releases/tags API |
| Pinned source SHA | `26b4c01e4839bbbda8866655e086471c4917ee53` (commit for tag `3.1.2`) | GitHub tags API |
| Default branch | `release` | GitHub repository API |
| Project homepage | `https://dnp3.github.io` | Upstream repository metadata |

**Peer eligibility review (plan level):**

- Independence: OpenDNP3 is a separately authored C++ stack; it is not a copy of,
  nor derived from, this Go library. **ELIGIBLE (pending §7 license/dependency
  review of the exact pinned tree).**
- Archived status: the upstream repository is archived. This does **not**
  disqualify it as an independent peer — it remains a genuine, independently
  implemented third-party stack — but it means no upstream fixes and no
  maintained releases. Record this as a known limitation for any future claim.
- Version pinning: use exactly tag `3.1.2` at source SHA
  `26b4c01e4839bbbda8866655e086471c4917ee53`. Do not float to `release` HEAD.

**Third-party dependency provenance (to record at acquisition):** OpenDNP3's
build pulls additional components (for example its ASIO/CMake toolchain and, for
the .NET/Java bindings, additional runtimes). These are **not** enumerated here
because they have not been inspected. A future authorized preparation step must
produce a complete dependency list with licenses for the exact pinned tree before
any build.

**Artifact checksum:** the tag **commit SHA** above is verifiable via the GitHub
API without downloading. The release **tarball SHA-256** is **NOT recorded here**
because computing it would require a download, which this task prohibits.
Record it at acquisition time (future authorized step) and compare against the
official release.

**Do not infer** from this section that any specific v0 object variation,
multi-header behavior, or control feature is supported by this release; those are
marked `TO VERIFY` in §4.

---

## 2. Reproducible preparation design

All commands below are **NOT EXECUTED** by this task. They describe the intended,
reproducible preparation for a future authorized lab task.

**Host platform targets (propose, then pin one):**

| Item | Proposed value | Status |
|------|----------------|--------|
| Host OS | Debian/Ubuntu LTS x86-64 (server, non-desktop) | Proposed |
| Architectures | `linux/amd64` primary; `linux/arm64` optional | Proposed |
| Go toolchain (master side) | Go **1.22** (matches `go.mod` / CI `go-version: '1.22'`) | Required by this repo |
| C++ toolchain | C++11-capable GCC/Clang plus CMake (version to be recorded) | `TO VERIFY` |
| Build system | CMake per OpenDNP3 upstream build docs for the pinned tag | `TO VERIFY` |
| Peer artifacts | OpenDNP3 outstation demo binary built from the pinned tag | `TO VERIFY` |

**Planned command shape (illustrative only — NOT EXECUTED):**

```
# NOT EXECUTED — future authorized acquisition
git clone --depth 1 --branch 3.1.2 https://github.com/dnp3/opendnp3.git
git -C opendnp3 rev-parse HEAD   # expect 26b4c01e4839bbbda8866655e086471c4917ee53
# build per upstream CMake instructions; record compiler, CMake and OS versions
# record resulting binary SHA-256
```

**Expected artifacts (to capture later):** the pinned source tree at the recorded
commit, the built outstation demo binary with its SHA-256, the full compiler/CMake/
OS version record, and the resolved dependency/license list.

**Explicitly out of scope for this task:** no clone, download, compile,
installation, container build, or invocation of OpenDNP3 (or any other DNP3 peer)
is performed here.

---

## 3. Isolation design

The lab must be isolated from any production OT/SCADA network.

| Requirement | Design |
|-------------|--------|
| Topology | Dedicated **offline VM/host** *or* an **internal-only container network** (no external connectivity) |
| Sample private subnet | `10.77.0.0/24` (example private range) |
| Port exposure | No published/exposed DNP3 port; peers reachable only on the internal segment |
| Production reachability | **None** — no route to any production OT/SCADA network; no bridging to host LAN |
| Networking features | No routing/forwarding; no privileged container; no host network mode |
| Credentials | No production credentials of any kind |
| Data points | **Simulator-only, memory-backed** points; no physical I/O |
| Review | Isolation checklist must be reviewed before any connection is attempted |
| Permission | Any actual connection requires a **separate** human-promoted task (§7) |

Isolation checklist (to confirm before future execution):

1. The peer network cannot reach production or the public internet.
2. No DNP3 port is published outside the isolated segment.
3. No privileged/host networking; no routing/forwarding enabled.
4. All control targets are simulated memory-backed outputs.
5. A named reviewer signs off on isolation before the first packet.

---

## 4. Configuration matrix

Hypothetical v0 configuration for a future run. Items not demonstrable for the
pinned release are marked `TO VERIFY`.

| Setting | Proposed value | Status |
|---------|----------------|--------|
| Transport | TCP only | Required (v0) |
| TCP port | `20000` | Proposed (DNP3 default) |
| Master DNP3 link address | `3` | Proposed |
| Outstation DNP3 link address | `4` | Proposed |
| Static points served | G1V1 Binary Input, G30V1 Analog Input, G20V1 Counter | Required (v0) |
| Binary Input indices/values | e.g. BI[0]=true, BI[1]=false | Illustrative — `TO VERIFY` against peer config |
| Analog Input indices/values | e.g. AI[0]=2300, quality ONLINE | Illustrative — `TO VERIFY` |
| Counter indices/values | e.g. CT[0]=42 | Illustrative — `TO VERIFY` |
| Class-0 poll request | G60V1, qualifier 0x06 (all objects) | Required (v0) |
| Direct-Operate | G12V1 CROB, qualifier 0x00 index8, count 1, to a **simulated** output only | Required (v0) |
| Multi-header packing in one APDU | behavior of this peer release | **`TO VERIFY`** |
| Variation/qualifier acceptance | peer's accepted variations for G1/G20/G30 | **`TO VERIFY`** |

**Explicitly unsupported (must remain rejected):** select-before-operate,
direct-operate-no-response, unsolicited responses and event classes, serial
transport, TLS, secure authentication, time sync, file transfer, device
attributes, restart/delay/freeze, and any object group/variation outside the v0
subset. No unsupported feature may silently become supported.

---

## 5. Future test matrix — NOT execution

Each row is a planned future test. None is executed by this task. Prerequisites,
expected evidence and stop criteria are defined so a later authorized task can
run them reproducibly.

| # | Test | Prerequisite | Expected evidence | Stop criterion | Classification |
|---|------|--------------|-------------------|----------------|----------------|
| 1 | Link reset + link status | Isolated link up | Captured handshake frames; decoded control/addrs | No handshake within deadline | FAIL / ABORT |
| 2 | Class-0 read (G1/G30/G20) | Test 1 pass | Decoded static points matching config | Missing/incorrect points | FAIL |
| 3 | Multi-header check | Peer packs multiple headers | All G1/G20/G30 points decoded, no loss | Point loss | FAIL |
| 4 | G12V1 Direct-Operate result | Simulated output configured | Protocol command status / IIN-based result | Ambiguous outcome | AMBIGUOUS (never blind retry) |
| 5 | Clean close | Tests 1–4 | TCP close, no stuck state | Hang | FAIL |
| 6 | Bad CRC | Corrupted frame | Bounded CRC error, no hang | Hang / false success | FAIL |
| 7 | Mismatched address | Wrong peer address | Bounded address error, no hang | Hang | FAIL |
| 8 | Connection drop | Mid-exchange socket close | Classified disconnect/timeout within deadline | Deadlock | FAIL |
| 9 | Command rejection | Peer rejects control | Classified failure from status/IIN | ControlSuccess reported | FAIL (false success) |
| 10 | Incomplete capture | Capture stops mid-run | Artifact marked incomplete | — | INCOMPLETE (rejected as evidence) |

Rules: never blindly retry a control whose outcome is unknown; treat ambiguous
outcomes as ambiguous; a hang or deadlock is a finding, not a pass.

---

## 6. Capture and artifact plan

| Item | Plan |
|------|------|
| Capture tooling | Passive `tshark`/`tcpdump` on the isolated segment, plus master-side application logs |
| Raw capture | Unmodified `.pcap`/`.pcapng`, retained as-is |
| Derived fixture | `.vec` per `active_work/testdata/external/FORMAT.md` (metadata header + `@ <direction> <layer>` frame records with original wire bytes) |
| Redaction | Separate redacted sidecar `.vec`; list every redaction |
| Clock correlation | Shared monotonic reference between capture and application logs; record offset/uncertainty |
| Provenance fields | Peer stack/version + build SHA; peer source artifact SHA-256; file SHA-256 of each artifact; tool versions; timestamps; exact master execution SHA |
| Reviewer mapping | Independent reviewer assigns ACCEPT / REJECT / INCOMPLETE |
| Integrity | No fabricated captures; synthetic-from-spec fixtures do **not** count for R4; incomplete captures are rejected |

---

## 7. Authorization gates and risks

**Future authorization gates (each separate; none granted by this document):**

1. A human-promoted CWAL task to **deploy** the isolated lab (acquire/build the
   pinned OpenDNP3 peer).
2. A further human-promoted CWAL task to **execute** the VEC-01 test campaign.
3. Verify the source **license** of the pinned candidate and all transitive
   dependencies before building.
4. Confirm and review network isolation before the first connection.
5. Restrict all control targets to simulated outputs.
6. Independent review before any interoperability claim is made.

**Risks:**

| Risk | Mitigation |
|------|------------|
| Archived upstream (no fixes/releases) | Pin exact tag/SHA; record archived status as a limitation |
| Unverified dependency licenses | Full dependency/license inventory before build |
| Accidental production reachability | Enforced isolation checklist + reviewer sign-off |
| Over-claiming from a single capture | Scope any claim to the tested peer version only |
| Blind control retry on ambiguity | Explicit never-blind-retry policy (§5) |
| Fabricated/synthetic evidence | Integrity rules (§6); synthetic fixtures excluded from R4 |

**Non-claims:** this lab plan cannot close R4 and cannot establish IEEE 1815
conformance, certification, production readiness, or universal DNP3
compatibility. A single accepted capture would evidence only the tested v0 path
against the pinned OpenDNP3 version.
