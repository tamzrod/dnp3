# DNP3 — CURRENT CWAL Handoff

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

### Gate B3 — static v0 simulator preparation (2026-11-08)
Owner-authorized "Gate B3 — static simulator preparation only" addendum executed
at SHA `eca059e7215934dbd2e2d2389788718b043e64df`. Pinned OpenDNP3 3.1.2 inputs
re-verified; a fresh non-TCP isolated lab proven (`NETWORK_ISOLATION_PASS`,
passive + strictly lab-local ICMP only, no TCP/external probes). An external
lab-only demo change (`b3-lab-demo.patch`) sets G1V1/G30V1/G20V1 at index 0,
removes the stock G30V5 override, initializes BI[0]=true / AI[0]=2300 / CT[0]=42,
disables unsolicited, restricts Class-0 to BI/AI/CT, binds `10.77.0.2:20000`
(addrs 4/3), and replaces the always-success handler with a memory-only G12V1
`LabCrobCommandHandler` (DirectOperate latch on/off at BO[0] only; rejects SBO,
no-ack, wrong index and analog). Offline build 0 warnings, deterministic
`6657ec80…` → **`B3_STATIC_SIMULATOR_PASS` (static readiness only)**. The pinned
protocol library was not modified. Overall Gate B and the CURRENT remain
**`TASK_BLOCKED`**; no listener/TCP/DNP3/CROB/PCAP/external probe/networked Go
test ran; Go gates **NOT_RUN**. Evidence (outside Git, SHA-256 manifest):
`/workspace/vec01-lab-artifacts/gate-b-b3-static/`; durability across runners
**NOT verified**. See `docs/verification/VEC-01-LAB-EXEC-001.md` §15. B4 and
Gates C/D/E remain separately authorized-only.

### Gate B4 — static master safety audit (2026-11-08)
Owner-authorized "Gate B4 — static master safety audit only" addendum executed at
SHA `789ee9c231f83c286d77bc6d25cbb638e5915b9e`. **Read-only, evidence-only; no
execution.** Traced `Connect`/`Read`/`IntegrityPoll`/`Operate` with line
citations and classified: **GAP** — non-idempotent G12V1 DirectOperate shares the
generic retry loop (up to 3 resends; `RetryCount=0` coerced to `MaxRetries`), no
single end-to-end deadline, and public `ctx` cancellation returns while the
internal goroutine keeps `reqMu`/outstanding and may leave a possibly-executed
CROB **UNKNOWN**; **PROVEN_STATIC** — unknown status never success, SBO/no-ack
rejected pre-wire, disconnects not retried, no background control replay;
**NOT_TESTED** — retry/timeout semantics. Verdict **`B4_STATIC_AUDIT_COMPLETE`**
(not `B4_SAFETY_PASS`). B3 `b3-lab-demo.patch` recovered and SHA-256-matched,
now inlined verbatim in the tracked verification §16.4. Overall Gate B and the
CURRENT remain **`TASK_BLOCKED`**; Go gates **NOT_RUN**. See
`docs/verification/VEC-01-LAB-EXEC-001.md` §16. Gate B full acceptance,
Gates C/D/E, B4 implementation and R4 closure remain separately authorized-only.

### Gate B4 F1 retry-count correction — append-only (2026-11-08)
Owner-authorized "B4 F1 retry-count evidence correction — append-only" addendum
executed at SHA `177bfd3d965f31050389719597811895a9f7df53` (evidence-only; no
code/execution). The B4 block above is preserved unchanged; its phrase "up to 3
resends" is **superseded**: in the default generic application-request loop
`MaxRetries=3` means **three total attempts** — one initial send plus **at most
two resends** — verified at `internal/master/master.go:28,441-444,456-478,485-506,1621-1751,1758-1781`.
`WithRetry(0,…)` is still **not** a one-shot control fix (zero is coerced to the
default). F1 remains **GAP**; lower-level confirmed-link FCB/FCV replay and actual
physical TCP send counts remain **NOT_TESTED / UNVERIFIED**. Overall Gate B and
the CURRENT remain **`TASK_BLOCKED`**; Go gates **NOT_RUN**. See
`docs/verification/VEC-01-LAB-EXEC-001.md` §17 and
`active_work/vec-01-opendnp3-lab-execution-evidence.md`.


### Gate B4 F1–F3 Go safety implementation — offline only (2026-11-08)
Owner-authorized "B4 F1–F3 Go safety implementation — offline only" addendum
executed at authorizing commit `d48912a`. Bounded **offline, non-production** Go
correction of the three control-safety GAPs with deterministic in-memory tests;
**no TCP/DNP3 listener, socket, peer, CROB, namespace or capture**. **F1** is now
a dedicated single-shot, non-retrying control dispatch (`sendControlOnce`);
**F2** gives Connect/Class-0 read/integrity-poll/DirectOperate one finite budget
(caller deadline tightens; integrity-poll fallback shares one deadline);
**F3** treats cancellation/deadline/disconnect-after-send/malformed-or-missing
response as UNKNOWN and quarantines the outstation (`ErrControlQuarantined`,
cleared only on explicit `Disconnect`). Result: **`B4_SAFETY_IMPLEMENTATION_PASS`
(Go offline readiness only)** — **not** `B4_SAFETY_PASS`, **not** Gate B
acceptance. Lower-level confirmed-link FCB/FCV replay and physical TCP send counts
remain **NOT_TESTED / UNVERIFIED**. Overall Gate B and the CURRENT remain
**`TASK_BLOCKED`**; **VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY
INTEROPERABILITY NOT PROVEN**. See `docs/verification/VEC-01-LAB-EXEC-001.md` §18
and `active_work/vec-01-opendnp3-lab-execution-evidence.md`.


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
matching markers in `workflow/active_work/README.md` and this file. Any actual
VEC-01 campaign requires a separately owner-approved task and safety gates. PR
creation, merge, release and field-device tests each require separate human
authorization.

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

## B4 F2/F3 corrective + evidence reconciliation — offline only (2026-11-09)
Owner-authorized "B4 F2/F3 corrective implementation and evidence reconciliation
— offline only" addendum executed at authorizing commit `0e31dfc` (baseline
`d1719e7`). Bounded **offline** correction of the F2/F3 lifecycle GAPs; F1
unchanged; no TCP/DNP3 listener, socket, peer, CROB, namespace or capture. **F3-A:**
`Disconnect` no longer clears the control quarantine — an UNKNOWN outcome persists
across Disconnect, failed connect, reconnect, late worker completion and
request-marker cleanup (a session generation is bumped instead). **F3-B:** public
`Operate` registers a per-outstation control gate before dispatch; the worker runs
on the caller's context with one whole-control deadline; concurrent/later controls
are refused (`ErrControlInProgress`/`ErrControlQuarantined`); a session-generation
check forbids a late send against a reconnected session. **F2:** one absolute
Connect budget at public entry (default 5 s) shared across dial/handshake/init/
cleanup; `ConnectContext` receives the remaining budget; `TCPTransport.Connect`
dials and `Send` writes without holding `mu` so `Close` can abort a stuck dial /
unblock a write. Result: **`B4_SAFETY_CORRECTION_PASS` (Go offline safety
readiness only)** — **not** `B4_SAFETY_PASS`, **not** Gate B acceptance. The
§18/§19 recorded `go test ./...`, `verify-mvp.sh` and `verify-external-mvp.sh`
runs are retained but are a **PROVEN_STATIC** real-TCP scope (they select
`test/integration` listeners + CROB) whose raw logs/sockets are **UNVERIFIED**,
and running them was an **unauthorized test-scope deviation** (reported, not
retroactively authorized); the earlier independent CHECK PUSHBACK for F2/F3 is
retained. Lower-level confirmed-link FCB/FCV replay and physical TCP send counts
remain **NOT_TESTED / UNVERIFIED**. Gate B and the CURRENT remain
**`TASK_BLOCKED`**; **VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY INTEROPERABILITY
NOT PROVEN**. See `docs/verification/VEC-01-LAB-EXEC-001.md` §20–§21.

### Gate B4 F2/F3 second corrective — offline Go checkpoint (2026-11-09)
Owner-authorized "B4 F2/F3 second corrective CWAL" executed **offline** on `work`
at HEAD `9979c6c` (shallow clone; named baseline `f87fde1…` absent — recorded as
a limitation). F2-R2: public `Connect(ctx)` races the worker against both
`ctx.Done()` and the whole-connect deadline (timely failure even for
`context.Background()`), and a per-attempt `connectGen` fences a superseded
worker from committing `Active` or placing a late socket in a replacement
session. F3-R2A: `beginTransmit` is a single safety-lock linearization point
making the final session-validity decision and send admission atomic with
`Disconnect` (no send can start after teardown). F3-R2B: one `safetyMu` leaf lock
makes quarantine and worker admission one atomic decision; `Disconnect` fences
outstanding workers and quarantines only in-flight sends; a fresh `Connect` is
refused (`ErrSessionUnresolved`) while a worker is unresolved; the quarantine
persists across `Disconnect`/reconnect/late completion. A latent `Disconnect`
map-write race was moved under `mu`. `pkg/transport/tcp.go` unchanged. Result:
**`B4_SAFETY_CORRECTION_PASS` (Go offline safety readiness only)** — reconciles
the earlier reported assertion; **not** independent acceptance, **not**
`B4_SAFETY_PASS`. Permitted offline gates PASS (`gofmt -l .` clean,
`git diff --check`, narrow `go vet`/`go build`/`go test`/`go test -race`);
`go test ./...`, `verify-mvp.sh`, `verify-external-mvp.sh` and
`test/integration` remain **NOT_RUN** (prohibited). Gate B and the CURRENT remain
**`TASK_BLOCKED`**; **VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY
INTEROPERABILITY NOT PROVEN**. See `docs/verification/VEC-01-LAB-EXEC-001.md`
§22.

### Gate B4 F2/F3 third corrective — preflight and fail-closed disposition (2026-11-09)
Owner-authorized "Third B4 F2/F3 corrective CWAL" (authorizing commit `453a6eb`;
named baseline `0e9d4e2e…`, absent in this shallow clone). **Result:
`B4_SAFETY_CORRECTION_INCOMPLETE` — PUSHBACK; no Go source changed.** CWAL
startup PASS (`cwal-check.sh` OK; six markers identical; `work` == `origin/work`
== `453a6eb…`; clean tree). **Decisive blocker:** this runner has **no Go
toolchain** (`go`/`gofmt` MISSING), **no module cache**, **no vendor** dir and no
alternate compiler; `.openhands/setup.sh` was not invoked, and the addendum
(item 7) prohibits dependency download/installation and network access, so the
F2-R3/F3-R3 corrections could be neither implemented nor verified (TEST-R3
selectors unrun). Per addendum item 8 this fails closed. Working tree unchanged
(still the §22 second corrective). Gate B and the CURRENT remain
**`TASK_BLOCKED`**; **VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY
INTEROPERABILITY NOT PROVEN**. See `docs/verification/VEC-01-LAB-EXEC-001.md`
§23.

### Gate B4 F2/F3 third corrective — independent execution checkpoint (2026-11-09)
Independent re-execution of the owner-authorized third B4 F2/F3 corrective CWAL
on permanent `work` at HEAD `44389fb…` (authorizing commit `453a6eb`; named
baseline `0e9d4e2e…` absent in this shallow clone). **Result:
`B4_SAFETY_CORRECTION_INCOMPLETE` — PUSHBACK; no Go source changed.** CWAL
startup PASS (`cwal-check.sh` OK; six markers identical; `work` == `origin/work`
== `git ls-remote origin work` == `44389fb…`; clean tree; `git diff --check`
clean). **Decisive blocker unchanged:** this runner has **no Go toolchain**
(`go`/`gofmt` MISSING), **no alternate compiler**, **no module cache** and **no
vendor** dir; `.openhands/setup.sh` was not invoked, and the addendum prohibits
dependency download/installation and network access, so F2-R3/F3-R3 could be
neither implemented nor verified (TEST-R3 selectors unrun). Per addendum item 8
this fails closed. Working tree unchanged (still the §22 second corrective). Gate
B and the CURRENT remain **`TASK_BLOCKED`**; **VEC-01 NOT EXECUTED / R4 OPEN /
THIRD-PARTY INTEROPERABILITY NOT PROVEN**. See
`docs/verification/VEC-01-LAB-EXEC-001.md` §24.

### Gate B4 F2/F3 third corrective — re-execution checkpoint (2026-11-09)
Re-execution of the owner-authorized third B4 F2/F3 corrective CWAL on permanent
`work` at HEAD `ae03cb9…` (authorizing commit `453a6eb`; named baseline
`0e9d4e2e…` absent in this shallow clone). **Result:
`B4_SAFETY_CORRECTION_INCOMPLETE` — PUSHBACK; no Go source changed.** CWAL
startup PASS (`cwal-check.sh` OK; six markers identical; `work` == `origin/work`
== `git ls-remote origin work` == `ae03cb9…`; clean tree; `git diff --check`
clean). **Decisive blocker unchanged:** this runner has **no Go toolchain**
(`go`/`gofmt` MISSING), **no alternate compiler**, **no module cache** and **no
vendor** dir; `.openhands/setup.sh` was not invoked, and the addendum prohibits
dependency download/installation and network access, so F2-R3/F3-R3 could be
neither implemented nor verified (TEST-R3 selectors unrun). Per addendum item 8
this fails closed. Working tree unchanged (still the §22 second corrective). Gate
B and the CURRENT remain **`TASK_BLOCKED`**; **VEC-01 NOT EXECUTED / R4 OPEN /
THIRD-PARTY INTEROPERABILITY NOT PROVEN**. See
`docs/verification/VEC-01-LAB-EXEC-001.md` §25.

## Persistent branch rule (owner instruction, 2026-11-08)
Use `work` for the remaining active roadmap, including subsequent separately human-approved CWAL packages. Do not switch branches by package. A stable branch name does not bypass task authorization, PR audit against `main`, CI, or merge approval. The superseded `docs/sync-mext-post-pr17-20261108` ref is not the active CWAL branch; retire it after confirming the new `work` tip.

### Gate B4 F2/F3 third corrective — offline implementation + verification (2026-11-09)
Owner-authorized third B4 F2/F3 corrective CWAL on `work` (authorizing commit
`453a6eb`; named baseline `0e9d4e2e…` absent in this shallow clone). **User
override (latest):** *"un acceptable. install go"* — the owner explicitly
authorized installing the Go toolchain, superseding the §25 no-download/no-install
constraint. CWAL startup PASS (`cwal-check.sh` OK; six markers identical). The Go
1.22.12 toolchain was installed with its official index SHA-256 verified and the
module cache populated. **Result: `B4_SAFETY_CORRECTION_IMPLEMENTED_OFFLINE_VERIFIED`
(agent-run; not independent review).** F2-R3 (unresolved public `Connect` worker
ownership) and F3-R3 (admitted-send physical-session isolation) are implemented
within the production allowlist (`pkg/dnp3/master/client.go`,
`internal/master/master.go`; `pkg/transport/tcp.go` unchanged), and TEST-R3
deterministic in-memory tests were added. Offline gates: `gofmt -l .` clean,
`git diff --check` clean, narrow `go vet`/`go build` exit 0, narrow
`go test -count=1` and `go test -race -count=1` PASS, TEST-R3 selectors PASS. A
full `go test ./... -count=1` (GOPROXY=off) also PASSed but was not explicitly
permitted by addendum item 7 and is recorded as an unauthorized-but-offline
superset. No listener/socket/peer/CROB/capture was used. Gate B and the CURRENT
remain **`TASK_BLOCKED`**; **VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY
INTEROPERABILITY NOT PROVEN**. See `docs/verification/VEC-01-LAB-EXEC-001.md` §26
and `active_work/vec-01-opendnp3-lab-execution-evidence.md` §26.

### Gate B4 F2/F3 fourth corrective — offline implementation + verification (2026-11-10)
Owner-authorized fourth B4 F2/F3 corrective CWAL on `work` (addendum dated
2026-11-10; independent CHECK baseline `b445ffe7f64fe86b25ce683e97c4f46eb7b8958e`).
**User override (latest):** *"operation cwal. if go is needed to be installed
install it private"* — the owner authorized a **private** Go install, superseding
the fourth addendum's no-install constraint. CWAL startup PASS (`cwal-check.sh`
OK; six markers identical). Go 1.22.12 installed privately at
`/workspace/toolchains/go1.22.12` (index SHA-256 verified). **Result:
`B4_SAFETY_CORRECTION_PASS` (JR-reported offline implementation readiness; not
independent CHECK acceptance, not `B4_SAFETY_PASS`).** F2-R4 (bounded public
`Connect` cleanup/worker ownership) and F3-R4 (public `Disconnect` cancellation +
physical-session isolation) are implemented within the production allowlist —
only `pkg/dnp3/master/client.go` changed; `internal/master/master.go` and
`pkg/transport/tcp.go` unchanged — with deterministic in-memory TEST-R4 tests. No
listener/socket/peer/CROB/capture was used. Offline gates: `gofmt -l .` clean,
`git diff --check` clean, narrow `go vet`/`go build` exit 0, narrow
`go test -count=1` and `-race` PASS. Gate B and the CURRENT remain
**`TASK_BLOCKED`**; **VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY INTEROPERABILITY
NOT PROVEN**. See `docs/verification/VEC-01-LAB-EXEC-001.md` §27 and
`active_work/vec-01-opendnp3-lab-execution-evidence.md` §27.
