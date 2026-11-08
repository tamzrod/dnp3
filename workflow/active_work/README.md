# DNP3 — CURRENT Active Work

Repository: tamzrod/dnp3
Package: VEC-01-LAB-EXEC-001
Branch: work
CURRENT: VEC-01-LAB-EXEC-001
QUEUED: NONE
Status: TASK_BLOCKED








## Current Gate B authority — owner correction (2026-11-08)
The owner has amended the existing CURRENT with a later, explicit precedence rule: **objective fail-closed safety checks replace mandatory separate reviewer identity/signature/approval**. Earlier reviewer language below and in previously recorded evidence is **historical**, not a current execution gate. The independent peer remains OpenDNP3 3.1.2 and interoperability/compliance must be demonstrated by actual raw PCAP + faithfully derived .vec + IEEE 1815 **v0 subset** checks; this does not establish general certification.

**Still TASK_BLOCKED on actual technical evidence:** the inspected pod has internet egress/forwarding and does not prove isolated OT-free routes; stock demo binds `0.0.0.0:20000` with 10/1 addresses, offers G1V2/G30V5 rather than verified target G1V1/G30V1, and lacks a demonstrated memory-only BO[0] CROB target. Missing isolation, listener, v0 variation, control safety or bounded deadlines must report FAIL/NOT_TESTED, never PASS. No network/test/control authorization or protocol execution occurred during the correction. `CURRENT` remains unchanged and `TASK_BLOCKED` until objective Gate B PASS, followed by valid independent-peer interoperability evidence, negative tests and mandatory Go gates. No PR/merge/R4 closure. JR invocation remains `operation cwal`.

### Gate B substep 1 result — isolation construction and proof (2026-11-08)
Owner-authorized "isolation construction and proof only" substep executed at SHA
`675183a64179c084ece77728bb57f779c679af4e`. A dedicated isolated lab was built
from two private netns (`vec01-master`, `vec01-outstation`) joined by an internal
veth pair only (`10.77.0.0/24`; addrs `10.77.0.1`/`10.77.0.2`), forwarding
disabled, deny-all inside the lab; the host/pod routes, firewall and cluster
policy were **not** modified. Objective boundary checks all **PASS**: no external
IPv4/IPv6 route, no route to cluster API/gateway, no DNS, no NAT/bridge/published
listener, lab-local-only reachability, no IPv6 globals, no physical I/O, host
boundary unchanged. **Result: `NETWORK_ISOLATION_PASS` (isolation only).** Gate B
overall remains `TASK_BLOCKED`; still separately authorized-only: B2 listener
safety, B3 peer v0 variations + verifiable memory-backed BO[0], B4 bounds/retry,
then Gates C/D/E and mandatory Go gates. Evidence (outside Git, SHA-256 manifest):
`/workspace/vec01-lab-artifacts/gate-b-isolation-proof/`. No listener, DNP3/TCP
packet, control, CROB, PCAP or networked Go test ran. See
`docs/verification/VEC-01-LAB-EXEC-001.md` §11.

### Gate B substep 2 result — static listener/link-address proof (2026-11-08)
Owner-authorized "static listener and link-address configuration only" substep
executed at SHA `bcec52ede1e7fe59c6a457426116ddc6efb5ad8b`. Its §2 precondition —
a **live** reverification of the substep-1 namespaces before any B2 work — did
**not** hold, so B2 stopped as mandated and the lab was **not** silently
recreated. Actual live reverification: no named netns (`/run/netns`,
`/var/run/netns` absent); self netns inode `net:[4026532553]` ≠ recorded host
`net:[4026532371]` (different pod `runtime-aozicfatuefownhn-86999879b-cmc2d`);
only `eth0`/`lo` (no veth, no `10.77.0.0/24`); default route `via 169.254.1.1`
present; `ip_forward=1`; nft empty and iptables policy ACCEPT; `ip` MISSING;
serial/USB/mem/port absent (PASS); pinned peer bytes and prior durable artifacts
ABSENT; Go MISSING. **Result: `NETWORK_ISOLATION_NOT_TESTED`;
`LISTENER_STATIC_CONFIG_NOT_TESTED`.** Gate B overall remains `TASK_BLOCKED`; B2
still separately authorized-only. Evidence (outside Git, SHA-256 manifest):
`/workspace/vec01-lab-artifacts/gate-b-substep2-reverify/`. No listener,
DNP3/TCP packet, control, CROB, PCAP, namespace recreation or networked Go test
ran. See `docs/verification/VEC-01-LAB-EXEC-001.md` §12.

### Gate B B1+B2 single-run recovery — ephemeral runner (2026-11-08)
Owner-authorized "B1+B2 single-run recovery — ephemeral runner correction"
addendum executed at SHA `ad48def4c57b22ce6c6c163e5a426234236d8fc8`. Because
OpenHands runners are ephemeral, a **fresh** lab was built in this execution and
its actual properties were proven. Pinned inputs were re-verified (tarball
`183cc292…`; asio/exe4cpp/ser4cpp SHA-256 match + SHA-1 match; tag peel
`26b4c01e…`), and a hermetic offline stock build reproduced `outstation-demo`
`8ae8c841…` and `libopendnp3.a` `e2926474…` **exactly**. A two-namespace lab
(`vec01-master` `10.77.0.1`, `vec01-outstation` `10.77.0.2`, single internal veth,
`10.77.0.0/24`) proved no route/DNS/egress/NAT/bridge, forwarding off, deny-all
policy, lab-local-only reachability and no physical I/O → **`NETWORK_ISOLATION_PASS`**.
An address-only peer patch (bind `10.77.0.2:20000`, `LocalAddr=4`, `RemoteAddr=3`,
rebuild `12f96e21…`) with matching prospective Go master settings (local 3, remote
4, target `10.77.0.2:20000`) passed static checks → **`LISTENER_STATIC_CONFIG_PASS`**.
Overall Gate B and the CURRENT remain **`TASK_BLOCKED`**; no DNP3 listener/frame,
control, CROB, PCAP or networked Go test ran, **but** a temporary non-DNP3 TCP
listener with a lab-local TCP connection and two external TCP probes (both
blocked) occurred **outside authorization**; Go gates **NOT_RUN** (no Go
toolchain on runner). Evidence (outside Git, SHA-256 manifest):
`/workspace/vec01-lab-artifacts/gate-b-b1b2-recovery/`; durability across runners
**NOT verified**. See `docs/verification/VEC-01-LAB-EXEC-001.md` §13 and the
owner-authorized **TCP scope correction §14**. B3/B4 and Gates C/D/E remain
separately authorized-only.

## CURRENT — VEC-01-LAB-EXEC-001 (human-promoted 2026-11-08)
Only the human-promoted `workflow/active_work/VEC-01-LAB-EXEC-001.md` is CURRENT, on permanent `work` after a safe non-force fast-forward to merged PR #21 main `02a58f76cfd8b82cbea1bcddcb36f61ce2738b21`. This promotion authorizes only the bounded independent **isolated simulated** master ↔ OpenDNP3 3.1.2 lab interoperability campaign specified there, conditional on verified external artifacts and durable retention, an independently reviewed network-isolation and simulated-output sign-off **before any listener or packet**, bounded positive and negative tests, raw PCAP and trace-derived .vec integrity, never-blind-control-retry, full fresh Go gates and independently accepted evidence. No other Git files beyond the task's exact allowlist, no production/OT connectivity, unsupported features, R4 closure, PR, merge or release. **No test execution has occurred during promotion. VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY INTEROPERABILITY NOT PROVEN.**

## Package closure — DNP3-061 (2026-11-08)
The VEC-01-LAB-DEPLOY-001 startup provenance correction is complete. DNP3-061
replaced only the startup `origin/main`/merge-base paragraph in
`docs/verification/VEC-01-LAB-DEPLOY-001.md`: the independently verified
promotion/integration baseline is `main` at
`f21f32cf5d0631c8f50e6b3f398bffe975bbb6ab` (merged PR #20, also the merge base of
`work` and `main`), while the earlier `95ac1028…` assertion is identified as PR
#17's historical baseline and the original local `origin/main` tracking-ref
observation is recorded as **UNVERIFIED** (not invented). All other bytes,
build hashes, recorded test exits, archive and readiness plan are unchanged.
All mandatory Go gates PASSED on the exact `work` tree at
`1fe78501281b6ffd73c2f48d687e67745d7d0809` with provisioned Go 1.22.12.
Evidence: `docs/verification/DNP3-061.md`; archive:
`archive/active-work/DNP3-061.md`. **VEC-01 NOT EXECUTED / R4 OPEN /
THIRD-PARTY INTEROPERABILITY NOT PROVEN.**

## Package closure — VEC-01-LAB-DEPLOY-001 (2026-11-08)
The isolated OpenDNP3 peer acquisition/build/static-smoke package is complete
**without protocol execution**. It added
`active_work/vec-01-opendnp3-lab-deployment-readiness.md` and lab artifacts
outside the repo: OpenDNP3 3.1.2 pinned at commit
`26b4c01e4839bbbda8866655e086471c4917ee53` (tag verified), acquired-archive
SHA-256 `183cc292…`; dependencies asio-1-16-0, exe4cpp, ser4cpp and (unused)
catch acquired with SHA-256 recorded and SHA-1 cross-checked against upstream
declarations; licenses OpenDNP3 Apache-2.0, exe4cpp/ser4cpp BSD-3-Clause, asio
Boost-1.0. Hermetic no-network build (`unshare -n`, FetchContent fully
disconnected) produced a byte-identical `outstation-demo`
(SHA-256 `8ae8c841…`); static ELF/ldd/hash checks passed; the socket-binding
demo was NOT run and no listener/port-20000/protocol activity occurred. All
mandatory Go regression gates PASSED on the exact `work` tree at
`5766bd1c4c3577d67b455201dbbb102026d48a11` with provisioned Go 1.22.12.
Evidence: `docs/verification/VEC-01-LAB-DEPLOY-001.md`; archive:
`archive/active-work/VEC-01-LAB-DEPLOY-001.md`. **VEC-01 NOT EXECUTED / R4 OPEN
/ THIRD-PARTY INTEROPERABILITY NOT PROVEN.**

## Package closure — VEC-01-LAB-PREP-001 (2026-11-08)
The OpenDNP3 laboratory preparation package is complete as **documentation
only**. It added `active_work/vec-01-opendnp3-lab-preparation.md`, a planning
placeholder covering peer provenance (independently implemented OpenDNP3 C++,
Apache-2.0, archived upstream, pinned tag 3.1.2), reproducible preparation
design, isolated non-production topology, the v0 configuration matrix, a future
test matrix (not execution), capture/artifact plan, and authorization gates. No
peer was cloned, built, installed or run; no DNP3 packets, capture, field
control or production interaction occurred. **VEC-01 remains NOT EXECUTED, R4
remains OPEN, and third-party interoperability is NOT PROVEN.** All mandatory
regression gates PASSED on the exact `work` tree at
`08d3882a3b63010e21be6d4021593dda0c5cc4e3` with a provisioned official Go
1.22.12 toolchain: `gofmt -l .` clean, `git diff --check`, `go vet ./...`,
`go build ./...`, `go test ./... -count=1`,
`go test -race ./internal/outstation/... -count=1`, `./scripts/verify-mvp.sh`,
and `./scripts/verify-external-mvp.sh` (Tiers 1-3). Evidence:
`docs/verification/VEC-01-LAB-PREP-001.md`; archive:
`archive/active-work/VEC-01-LAB-PREP-001.md`.

## Package closure — DNP3-060 (2026-11-08)
The MEXT-042 provenance correction is complete. DNP3-060 corrected the two
inaccurate PR #18 base-SHA labels in `docs/verification/MEXT-042.md` and
`archive/active-work/MEXT-042.md` from `95ac1028…` (which is actually merged
PR #17) to the correct merged PR #18 SHA
`b334c8cd3d8d1f84842c952c84ced4bef42efdd7` — exactly one line changed per file.
Historical MEXT-042 execution SHA, test exit results and the archived task body
are unchanged. All mandatory checks PASSED on the exact `work` tree at
`da00a24e42d93e0025033d32037421864cc6cf3a` with a provisioned official Go
1.22.12 toolchain: `gofmt -l .` clean, `git diff --check`, `go vet ./...`,
`go build ./...`, `go test ./... -count=1`,
`go test -race ./internal/outstation/... -count=1`, `./scripts/verify-mvp.sh`,
and `./scripts/verify-external-mvp.sh` (Tiers 1-3). Evidence:
`docs/verification/DNP3-060.md`; archive: `archive/active-work/DNP3-060.md`.
R4/VEC-01 remains OPEN.

## Package closure — MEXT-042 (2026-11-08)
The independent third-party capture procedure package is complete as
**documentation only**. MEXT-042 added `active_work/third-party-capture-procedure.md`,
a planning placeholder defining peer provenance, isolated topology, the v0-only
operation matrix, capture/artifact format, wire inspection checks, negative
cases, safe-stop and reviewer acceptance criteria. No capture was executed, no
harness or code was added, and no supported behavior changed. **VEC-01 remains
NOT EXECUTED, R4 remains OPEN, and third-party interoperability is NOT PROVEN.**
All mandatory checks PASSED on the exact `work` tree at
`4b6eecc81a7581677f91655c69d1d0050cfd8228` with a provisioned official Go
1.22.12 toolchain: `gofmt -l .` clean, `git diff --check`, `go vet ./...`,
`go build ./...`, `go test ./... -count=1`,
`go test -race ./internal/outstation/... -count=1`, `./scripts/verify-mvp.sh`,
and `./scripts/verify-external-mvp.sh` (Tiers 1-3). Evidence:
`docs/verification/MEXT-042.md`; archive: `archive/active-work/MEXT-042.md`.

## Execution stop
CURRENT is NONE and QUEUED is NONE. No implementation, protocol, profile or
governance task may run without a new human-promoted bounded CWAL microtask with
matching markers in this file and root `handoff.md`. Any actual VEC-01 campaign
requires a separately owner-approved task and safety gates. PR creation, merge,
release and field-device tests each require separate human authorization.

## VEC-01-LAB-EXEC-001 execution attempt — TASK_BLOCKED (2026-11-08)
The promoted isolated interoperability campaign was attempted and is **BLOCKED
at Gate B**, so it did not execute. Startup readiness PASSED (markers match,
`cwal-check.sh` OK, `origin/main` `02a58f76…` is the ancestor/merge-base of
`origin/work`). Gate A PASSED: the disposable `/tmp/vec01-lab` staging had
vanished, so the pinned peer was independently **re-produced** — OpenDNP3 3.1.2
peeled tag `26b4c01e4839bbbda8866655e086471c4917ee53`, source tarball SHA-256
`183cc292…`, all four dependencies re-hashed with matching SHA-256/SHA-1 — and a
hermetic no-network build (`unshare -n`, FetchContent fully disconnected)
reproduced `outstation-demo` `8ae8c841…` and `libopendnp3.a` `e2926474…`
**byte-identically** (a first build at a different absolute path differed only by
embedded source-path strings; rebuilding at the historical layout reproduced the
pinned hashes exactly). Verified bytes are retained in the durable store
`/workspace/vec01-lab-artifacts` (outside Git, outside `/tmp`) with a SHA-256
manifest. Gate B then BLOCKED: the task mandates a **named independent reviewer**
(distinct from the executor) to approve isolation / no-OT-reachability /
simulator-only CROB targets **before any listener or packet**, and states *"No
independent approval → TASK_BLOCKED and STOP. There is no autonomous
self-approval."* No such reviewer exists in this environment (single session,
no approver authority, no reviewer mechanism in governance), and the executor
may not self-approve. Consequently **no listener, TCP session, DNP3 packet,
capture or control occurred**; Gates C–E were not reached; the Go regression
gates are NOT_RUN (Go toolchain absent, and the real-TCP gates are also
prohibited pre-approval). CURRENT is retained with truthful `TASK_BLOCKED`
evidence (no archive, no marker closure). Evidence:
`docs/verification/VEC-01-LAB-EXEC-001.md`; index:
`active_work/vec-01-opendnp3-lab-execution-evidence.md`. **VEC-01 NOT EXECUTED /
R4 OPEN / THIRD-PARTY INTEROPERABILITY NOT PROVEN.** R4 status is unchanged and
requires separate owner approval.

### Gate B evidence-preparation substep (addendum, 2026-11-08) — still TASK_BLOCKED
Under the owner-authorized Gate B pushback addendum (EVIDENCE PREPARATION ONLY),
passive/static evidence was gathered with **no listener, no TCP/DNP3 packet, no
capture, no control and no probing** (execution SHA
`9d6833c914608c517748f9c26a8a5fc06701170f`). Findings: the executor environment is
a shared, routed Kubernetes pod with internet egress, `ip_forward=1`, no firewall
policy and no `ip`/iproute2, so isolation / no-OT-reachability is **UNVERIFIED and
not demonstrable here**; the peer's as-shipped Class-0 binary inputs default to
**G1V2** (not v0 **G1V1**), its demo binds `0.0.0.0:20000` with link addrs 10/1,
and it has **no verifiable simulator-only BO[0] CROB target** (operate returns a
status echo only). An **UNAPPROVED** reviewer packet (with a blank independent
checklist) is retained outside Git at
`/workspace/vec01-lab-artifacts/gate-b-reviewer-packet/`. Resumption requires an
authenticated named independent reviewer sign-off plus a demonstrably isolated
topology, v0-compliant peer variations, a verifiable simulator-only BO[0], and
owner authorization to proceed. Until then CURRENT stays `TASK_BLOCKED`; no
archive, closure, PR, merge or R4 change. See
`docs/verification/VEC-01-LAB-EXEC-001.md` §9.

## Persistent working branch policy
The maintainer designates `work` as the sole persistent execution branch for the remaining active DNP3 roadmap. New CWAL tasks are promoted and closed on `work`; no task-specific branch is created unless the maintainer explicitly reverses this direction. Every new CURRENT still requires explicit human promotion and bounded scope. `main` is the PR integration baseline, not the CWAL execution branch.
