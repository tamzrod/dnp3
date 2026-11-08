# VEC-01-LAB-EXEC-001 — Isolated OpenDNP3 interoperability execution (TASK_BLOCKED)

**Status: TASK_BLOCKED.** The campaign did **not** execute. No DNP3 listener was
activated, no TCP connection was opened, no DNP3 packet was sent or captured, and
no control was issued. The task was blocked at **Gate B** because the mandatory
**named independent reviewer** approval, required *before any listener or packet*,
could not be obtained. Per the task's own rule — *"No independent approval →
TASK_BLOCKED and STOP"* — execution stopped.

**VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY INTEROPERABILITY NOT PROVEN.**

- Task: VEC-01-LAB-EXEC-001 (human-promoted 2026-11-08)
- Branch: `work`
- Promotion baseline: `02a58f76cfd8b82cbea1bcddcb36f61ce2738b21` (merged PR #21, `main`)
- Local execution SHA at task start: `f44d28be4e862422cbc2e8233d3dfa9717e8471e`
- Executor: OpenHands agent (autonomous), acting under Operation CWAL
- Durable lab artifact store (outside Git): `/workspace/vec01-lab-artifacts`

---

## 1. Startup / readiness gate (PASS)

| Check | Command | Result |
|-------|---------|--------|
| Governance read | `AGENTS.md`, `operation cwal.md`, `BLACK_SHEEP_WALL.md`, both six-field markers, CURRENT task | Read in full |
| Marker agreement | `bash scripts/cwal-check.sh` | `CWAL CHECK OK: CURRENT=VEC-01-LAB-EXEC-001 branch=work` |
| Working tree | `git status --porcelain` | clean |
| Remote fetch | `git fetch origin` (+ `main:refs/remotes/origin/main`, `--unshallow`) | OK |
| `origin/main` | `git rev-parse origin/main` | `02a58f76cfd8b82cbea1bcddcb36f61ce2738b21` |
| `origin/work` | `git rev-parse origin/work` | `f44d28be4e862422cbc2e8233d3dfa9717e8471e` |
| Ancestry | `git merge-base --is-ancestor origin/main origin/work` | **main IS ancestor of work** (merge base `02a58f76…`); non-force fast-forward already in place, no reset/rebase needed |
| Six-field markers | README + `handoff.md` | Identical: repo `tamzrod/dnp3`, package/branch/CURRENT `VEC-01-LAB-EXEC-001`/`work`, QUEUED `NONE`, Status `TASK_READY` |

The checkout is a shallow clone; `git fetch --unshallow origin` was used to
establish full ancestry. `main` is the direct ancestor of `work` at the promoted
baseline, so no synchronization action was required and none was performed.

## 2. Gate A — peer artifact recovery / reproduction (PASS)

The disposable `/tmp/vec01-lab` staging workspace from VEC-01-LAB-DEPLOY-001 had
**vanished** (as the task anticipated). The peer was therefore **independently
re-produced** from pinned upstream sources, and every byte was re-hashed from
newly retrieved data (no copy manifest was trusted).

### 2.1 Source provenance (re-verified from retrieved bytes)

| Field | Expected (task) | Observed | Match |
|-------|-----------------|----------|-------|
| OpenDNP3 tag `3.1.2` (annotated) | — | `523f7d2d6c65aa671f594fc63c3adcd74e81407d` | — |
| OpenDNP3 `3.1.2^{}` (peeled commit) | `26b4c01e4839bbbda8866655e086471c4917ee53` | `26b4c01e4839bbbda8866655e086471c4917ee53` | ✅ |
| Source tarball SHA-256 | `183cc29222c3cb58099a9753c43defed5cb8677fda985c3f88e38b5c19a36ff2` | `183cc29222c3cb58099a9753c43defed5cb8677fda985c3f88e38b5c19a36ff2` | ✅ |

Acquisition URL: `https://codeload.github.com/dnp3/opendnp3/tar.gz/refs/tags/3.1.2`.
Tag refs read with `git ls-remote https://github.com/dnp3/opendnp3`.

### 2.2 Dependency provenance (re-verified)

| Dependency | Expected SHA-256 | Observed SHA-256 | SHA-1 (declared=observed) | Match |
|------------|------------------|------------------|---------------------------|-------|
| asio `asio-1-16-0` | `18883c350558eadf91218ae5ce9a8f8f85099c2f036c1c8579fb639092e928c5` | same | `6BDD…5402` = `6bdd…5402` | ✅ |
| exe4cpp `fb878a4de598ba9d6e4338afebf83f96e03af1b8` | `2a8a2c7d9982f54858b92c03cff91ef433907472039f8e61ff7c9f9b3d48db6c` | same | `18B1…01BD` = `18b1…01bd` | ✅ |
| ser4cpp `3c449734dc530a8f465eb0982de29165cc4e23d5` | `5a78f83de2d70bdf4be2ab607bc4416c6866ce64a257a8f19948cf5746762f6e` | same | `937B…A672` = `937b…a672` | ✅ |
| catch `v2.11.3` (test-only, unused) | `197640c056d6d01e5a56737fbdfe671d2b872c8c7533e01cc2de4be19ab7a028` | same | `C127…6C15` = `c127…6c15` | ✅ |

All dependency SHA-1s match the upstream-declared `URL_HASH` values in
`deps/*.cmake` at commit `26b4c01`.

### 2.3 Hermetic no-network build and byte-identical reproduction

Build executed inside a network namespace with **no network** (`unshare -n`;
verified in-namespace `curl https://github.com` → exit 6, no route). CMake ran
with `-DFETCHCONTENT_FULLY_DISCONNECTED=ON` and all three dependency sources
supplied locally via `FETCHCONTENT_SOURCE_DIR_*`, so no download occurred.

Toolchain: Debian 13 (trixie), x86_64, kernel 6.8.0-1055-gke, **CMake 3.31.6**,
**g++ 14.2.0**, GNU Make 4.4.1, ld 2.44.

Configure:

```
cmake -S . -B build-offline -G "Unix Makefiles" -DCMAKE_BUILD_TYPE=Release \
  -DDNP3_TLS=OFF -DDNP3_TESTS=OFF -DDNP3_JAVA=OFF -DDNP3_FUZZING=OFF \
  -DDNP3_EXAMPLES=ON -DDNP3_STATIC_LIBS=ON \
  -DFETCHCONTENT_FULLY_DISCONNECTED=ON \
  -DFETCHCONTENT_SOURCE_DIR_ASIO=<asio> \
  -DFETCHCONTENT_SOURCE_DIR_EXE4CPP=<exe4cpp> \
  -DFETCHCONTENT_SOURCE_DIR_SER4CPP=<ser4cpp>
```

Build: `cmake --build build-offline --target outstation-demo -j4` → exit 0.

| Artifact | Historical pinned SHA-256 | Reproduced SHA-256 (2 clean builds) | Match |
|----------|---------------------------|--------------------------------------|-------|
| `outstation-demo` | `8ae8c84125c4920352b6daef4145e2a91de9e02cd6e2c1b94bf6995ebeb093bb` | `8ae8c84125c4920352b6daef4145e2a91de9e02cd6e2c1b94bf6995ebeb093bb` | ✅ |
| `libopendnp3.a` | `e2926474288fb8d2c6c7962192bad2ff0f724f63c07d2209fe5b30c26fbe0677` | `e2926474288fb8d2c6c7962192bad2ff0f724f63c07d2209fe5b30c26fbe0677` | ✅ |

**Byte-order / path dependence finding (documented, understood, not a blocker):**
a first build performed at a *different* absolute source path
(`/tmp/vec01-build/opendnp3-3.1.2`) produced a **different** `outstation-demo`
(`18a19bb5…`) and `libopendnp3.a` (`2120d8e3…`). Inspection with `strings` showed
the release binary embeds the absolute source path in assertion/filename
strings (e.g. `…/cpp/lib/src/app/parsing/APDUParser.cpp(105)`). Rebuilding at the
historical layout (`/tmp/vec01-lab/src/opendnp3-3.1.2`) reproduced the pinned
hashes **exactly**, twice, confirming the mismatch was solely an embedded
absolute-path difference between otherwise identical inputs — not a source,
dependency, or toolchain discrepancy. This is the required "explain unexpected
mismatch before proceeding" step; it is resolved.

Static checks (durable copy): ELF64, type DYN (PIE), x86-64, mode `0755`; `ldd`
shows only `libstdc++/libm/libgcc_s/libc/linux-vdso` — no peer runtime libs.

### 2.4 Durable retention (outside Git, outside `/tmp`)

| Item | Durable path |
|------|--------------|
| Artifact store root | `/workspace/vec01-lab-artifacts` (persistent volume, outside the Git checkout) |
| Peer binary | `/workspace/vec01-lab-artifacts/artifacts/outstation-demo` |
| Peer static lib | `/workspace/vec01-lab-artifacts/artifacts/libopendnp3.a` |
| Licenses | `/workspace/vec01-lab-artifacts/artifacts/licenses/` (OpenDNP3 Apache-2.0; exe4cpp/ser4cpp BSD-3-Clause; asio BSL-1.0) |
| Source/dep archives | `…/source/`, `…/deps/` |
| Build logs | `…/logs/` |
| SHA-256 manifests | `…/manifests/sha256-manifest.txt`, `…/manifests/sha256-manifest-full.txt` |

All hashes above were recomputed from the durable copies after transfer. No C++
binary, tarball, PCAP or external source was committed to Git.

## 3. Gate B — independent isolation/output sign-off (BLOCKED — decisive)

The task requires, **before any listener activation or wire exchange**, that a
**named independent reviewer, distinct from the executor**, inspect and approve:
the isolation checklist, physical-output absence, peer configuration, and the
single simulated CROB target — recording identity, decision, date, reviewed
config hashes and exact scope. It states plainly: *"No independent approval →
TASK_BLOCKED and STOP. There is no autonomous self-approval."*

**Finding: no named independent reviewer is available in this execution
environment.** Verified by:

- This is the **only** conversation/session in the workspace
  (`/workspace/conversations/` contains exactly one conversation, this one); there
  is no separate reviewer session, agent, or human approver.
- No reviewer, approval, or sign-off authority is defined anywhere in the
  repository governance, `.kde/`, `active_work/`, or `.github/`; no reviewer
  identity, contact, or mechanism is recorded.
- The agent is the executor; per KDE core rule *"Never self-approve your own
  work"* and the task's explicit prohibition, it **cannot** self-approve the
  Gate B sign-off.

Consequently the Gate B precondition could not be satisfied. Opening a listener
or sending a DNP3 packet without it would violate both the task and the
repository safety invariants, so execution was stopped here. The required
"identity, decision, date, reviewed config hashes and exact scope" record
**cannot** be produced, and no autonomous substitute is permitted.

## 4. Gates C, D, E — NOT REACHED

Gates C (positive interoperability matrix), D (negative tests) and E (raw PCAP /
`.vec` / independent verdict) were **not executed**, because Gate B is a hard
precondition ("only after Gate A+B PASS"). No capture was started, so there is
no `.pcap`, no `.vec`, and no reviewer verdict. None is fabricated.

## 5. Mandatory Go regression gates — NOT RUN (with reasons)

| Gate | Status | Reason |
|------|--------|--------|
| `bash scripts/cwal-check.sh` | PASS | Clean startup, `CWAL CHECK OK` |
| `test -z "$(gofmt -l .)"` | NOT_RUN | Go toolchain not installed in this environment |
| `git diff --check` | PASS | Clean |
| `go version` | NOT_RUN | `go: command not found`; `/usr/local/go` absent |
| `go vet ./...` | NOT_RUN | Go toolchain absent |
| `go build ./...` | NOT_RUN | Go toolchain absent |
| `go test ./... -count=1` | NOT_RUN | Go toolchain absent |
| `go test -race ./internal/outstation/... -count=1` | NOT_RUN | Go toolchain absent |
| `./scripts/verify-mvp.sh` | NOT_RUN | Requires Go; also runs real-TCP loopback |
| `./scripts/verify-external-mvp.sh` | NOT_RUN | Requires Go; also runs real-TCP loopback |

Two independent reasons apply: (a) the Go toolchain is not present in this
environment; and (b) `verify-mvp.sh` / `verify-external-mvp.sh` exercise the
in-repo real-TCP master↔outstation loopback path, which **opens listeners and
sends DNP3 frames**. Running them before the Gate B independent approval would
breach the campaign's "no listener or packet before approval" constraint. The
task states no previous PASS may substitute and that NOT_RUN must not be counted
as success; they are therefore recorded truthfully as NOT_RUN, not PASS.

## 6. Scope compliance

Only Git paths touched by this task: this new verification report, the new
`active_work/vec-01-opendnp3-lab-execution-evidence.md` index, and the two
authority-marker files (`workflow/active_work/README.md`, `handoff.md`). No Go
code/test/fixture, `active_work/testdata/external/*`, `supported-profile.md`,
`go.mod`/`go.sum`, CI/scripts, governance, or roadmap file was modified. The
CURRENT task file was **not** archived and **not** removed (acceptance not
reached). No PR, merge, or field-device interaction occurred.

## 7. Non-claims and limitations

- **VEC-01 NOT EXECUTED; R4 remains OPEN; third-party interoperability NOT PROVEN.**
- No DNP3 listener, session, packet, capture, or control occurred.
- Gate A only proves the pinned peer builds reproducibly in isolation; it is
  **not** evidence of protocol interoperability.
- No IEEE 1815 certification, full-conformance, or production-readiness claim.
- The blocker is authority/independence, not a technical protocol failure.

## 8. Unblock path (for the maintainer)

To proceed, the owner must either (a) record a **named independent reviewer**
(distinct from the executor) who can inspect and sign the Gate B isolation /
physical-output / simulated-target checklist, or (b) supply explicit written
owner authorization that resolves the task's mandatory independent-review
requirement for an isolated, simulated, non-routable lab. Only after that
approval may Gates C–E and the full Go gates run. No corrective action was taken
autonomously.

---

## 9. Gate B evidence-preparation substep (human-authorized addendum, 2026-11-08)

Under the owner-authorized **Gate B pushback addendum (EVIDENCE PREPARATION
ONLY)**, the following **passive/static** work was performed while the task
remains `TASK_BLOCKED`. **No listener was started; no TCP/DNP3 connection was
made; no packet was sent or captured; no CROB was invoked; nothing was probed on
any operational network.** Execution SHA for this substep:
`9d6833c914608c517748f9c26a8a5fc06701170f` (branch `work`).

### 9.1 Durable location and hashes

All substep evidence is retained **outside Git** in the durable store
`/workspace/vec01-lab-artifacts`, under
`gate-b-reviewer-packet/` (UNAPPROVED reviewer packet; see its
`packet-manifest.txt` for SHA-256 of every file). Retained peer bytes were
re-hashed fresh and still match the pinned values:
`outstation-demo` `8ae8c841…`, `libopendnp3.a` `e2926474…`, source tarball
`183cc292…`, dependencies `18883c35…` / `2a8a2c7d…` / `5a78f83d…` /
`197640c0…`.

### 9.2 Static network / isolation inspection (passive)

| Observation | Actual value | Verdict |
|-------------|--------------|---------|
| Host | `runtime-omgblkbssxnzsmru-8967c77bc-q48q7` (Kubernetes pod) | not a dedicated offline host |
| Interfaces | only `lo` and `eth0` | no isolated segment |
| Address / route | `10.2.19.149/…`, live default route via `169.254.1.1` | routed, not non-routed |
| Internet egress | DNS `10.3.0.10`, k8s API `10.3.0.1:443` reachable | internet/control-plane reachable |
| `net.ipv4.ip_forward` | `1` (and `conf.all.forwarding=1`) | forwarding ON |
| Firewall | `iptables` chains empty (policy ACCEPT); `nft` ruleset empty | no isolation policy |
| `ip`/iproute2 | **absent** (also `bridge`, `veth`, `ss` absent) | cannot build `10.77.0.0/24` |
| Netns capability | `unshare` present; non-root `CapEff=0`, `sudo` `CapEff=0x1ffffffffff` | no evidence of an isolated, no-OT-reachability topology |

**Result:** the environment is a shared, routed, internet-connected pod.
Isolation and no-OT-reachability are **UNVERIFIED** and, on current evidence,
**not demonstrable here**. The addendum forbids claiming isolation from intended
design alone, so none is claimed.

### 9.3 Static peer configuration inspection

- As-shipped demo binds `IPEndpoint("0.0.0.0", 20000)` and uses link addresses
  `LocalAddr=10`, `RemoteAddr=1` — explicitly unsafe/out-of-spec, **not run**.
- Default static variations (`MeasurementInfo.cpp`): **Binary Input = G1V2**
  (`Group1Var2`), Analog Input = G30V1, Counter = G20V1. The v0 profile and this
  task require **G1V1** binary inputs → **real variation mismatch (recorded, not
  hidden)**. The demo also overrides `analog_input[0].svariation = G30Var5`.
- `SuccessCommandHandler` echoes the CROB and returns fixed `SUCCESS`; its base
  `DoOperate` writes nothing (no physical and no memory output).

### 9.4 Static simulator / output-mapping inspection

- Demo points are memory-backed (`struct State`), but the task's proposed exact
  ground-truth values (BI[0]=true, AI[0]=2300, CT[0]=42) are **not** preset and
  are **UNVERIFIED**.
- **No BO[0] target exists** in the demo; no BO/CROB `Update` is issued. The
  required "single verifiable simulated CROB target" is **UNVERIFIED / ABSENT**;
  a CROB "success" from this peer is a status echo, not a verified state change.
- No physical I/O path: no `gpio`/`/dev/tty`/`serial`/`modbus` references;
  `ldd` shows only standard system libraries. (Positive safety property.)

### 9.5 UNAPPROVED reviewer packet

Assembled at `/workspace/vec01-lab-artifacts/gate-b-reviewer-packet/` with:
`README.md`, `inspection/network-inventory.txt`,
`inspection/firewall-inventory.txt`, `inspection/peer-artifact-manifest.txt`,
`inspection/peer-configuration.txt`,
`inspection/simulator-and-output-mapping.txt`,
`config-hashes/config-file-sha256.txt`, `proposed-topology.md`,
`pending-deadlines-and-watchdog.md`, `risk-and-uncertainty.md`,
`independent-preflight-checklist.md` (**blank**), and `packet-manifest.txt`.

### 9.6 Outstanding UNVERIFIED fields and resumption prerequisites

- Reviewer identity, qualification, decision, date and authenticated sign-off:
  **not provided** (no reviewer or review channel supplied).
- Environment isolation / no-OT-reachability: **UNVERIFIED / not demonstrable**.
- Peer v0 variation compliance (G1V1): **UNVERIFIED / mismatched as shipped**.
- Verifiable simulated CROB BO[0] target: **UNVERIFIED / absent**.
- Exact Class-0 response qualifier, multi-header packing, sequence/confirm
  behaviour, CROB response bytes: **UNVERIFIED** (require execution).
- Go toolchain for the master side: **absent**.

**Resumption prerequisites (all required):** authenticated independent reviewer
sign-off on the actual environment and hashes; a demonstrably isolated,
non-routable, no-OT-reachability topology; a peer configuration with v0
variations (G1V1/G30V1/G20V1) and a verifiable simulator-only BO[0] target; and
owner authorization to proceed to execution. Until then the task stays
`TASK_BLOCKED`.

### 9.7 Non-networking checks run in this substep

| Check | Result |
|-------|--------|
| `bash scripts/cwal-check.sh` | PASS (clean, synced CURRENT tree) |
| `git diff --check` | PASS |
| Fresh `sha256sum` of retained peer bytes | PASS (match pinned) |
| Passive `/proc` network/namespace reads | PASS (evidence recorded) |
| `sudo -n iptables -L` / `nft list ruleset` (list-only) | PASS (empty policy observed) |
| `gofmt` / `go vet` / `go build` / `go test` / `verify-*.sh` | NOT_RUN (Go absent; real-TCP gates prohibited pre-approval) |

**No network listener, TCP/DNP3 packet, control, PCAP capture, Gate C/D/E work,
archive/closure, PR, merge, or R4 status change occurred in this substep.**
**VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY INTEROPERABILITY NOT PROVEN.**

---

## 10. Owner-authorized Gate B rule correction — 2026-11-08 (NO EXECUTION)

**Prospective authority change, not a rewrite of the checkpoint:** The original sections above truthfully report JR's historical stop under the then-effective independent-review requirement. The repository owner has now amended CURRENT `workflow/active_work/VEC-01-LAB-EXEC-001.md` to **remove mandatory separate-person review/signature** and replace it with objective, recorded, fail-closed isolation and simulator-only verification. The earlier requested unsigned reviewer packet is retained as historical, optional evidence; no signature or reviewer identity is a current blocking prerequisite. No past independent approval is claimed.

**Still BLOCKED for technical reasons, independently of paperwork:** The reported execution environment is a routed Kubernetes pod with internet egress, IP forwarding and no proven protected isolated network; default peer listener `0.0.0.0:20000` / addresses 10/1 are not approved; stock OpenDNP3 binary input G1V2 and analog G30V5 differ from required G1V1/G30V1, and the required memory-backed BO[0] CROB target has not been demonstrated. These are actual deficiencies or UNVERIFIED settings, not resolved by changing the reviewer rule. A peer config/adaptation and isolation must be demonstrated without unapproved changes to Go master or tracked code.

**Current evidence state:** Gate A remains recorded PASS (reproduction and historical hashes; rehash retained bytes when resuming). Gate B objective safety: FAIL / NOT_TESTED on the items above. Gates C–E: NOT_TESTED. Required Go gates: retain their exact earlier NOT_RUN outcomes; none retroactively PASS. No listener, TCP/DNP3 packet, physical output, CROB, raw PCAP or `.vec` has occurred under this correction.

**Prospective acceptance:** Automated, recorded isolation/routes/forwarding/firewall/listener checks and actual simulator-only CROB wiring must PASS before **any** protocol execution; finite deadlines, no blind retries; then actual external OpenDNP3 3.1.2 interoperable traffic, genuine raw PCAP and capture-derived `.vec`, IEEE 1815 v0 wire-field comparisons, mandatory positive/negative cases and fresh Go gates must PASS. Missing/failed/untested mandatory checks retain TASK_BLOCKED. A second-person sign-off or separate independent reviewer verdict is no longer necessary, but the peer remains an independently implemented C++ stack and evidence must be independently reproducible.

**Checkpoint scope:** Historical report text above preserved; this additive note is not a claim of execution, compliance, R4 closure, new PR or merge. **VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY INTEROPERABILITY NOT PROVEN.**

---

## 11. Gate B substep 1 — isolation construction and proof — 2026-11-08 (NETWORK_ISOLATION_PASS)

**Authority:** Owner-authorized "queued Gate B substep 1 — isolation construction
and proof only" addendum in CURRENT `workflow/active_work/VEC-01-LAB-EXEC-001.md`.
Scope is limited to resolving **only** the first objective Gate B blocker
(isolated, non-routed network with no internet/OT reachability). Gates B2–B5, C,
D, E were **not** authorized and were **not** performed. Execution SHA:
`675183a64179c084ece77728bb57f779c679af4e` (branch `work`).

**Verdict: `NETWORK_ISOLATION_PASS`** (isolation only). **Gate B overall remains
`TASK_BLOCKED`** because the listener-safety (B2), peer v0 capability/simulated
BO[0] (B3) and bounds/retry (B4) conditions are separate and not yet authorized
or passed.

### 11.1 What was built (non-Git, disposable, lab-only)

A dedicated isolated network was constructed using two private network
namespaces joined by an **internal veth pair only** — no host networking,
no privileged bridge, no NAT, no external port publishing, no production/OT
interface. The host/pod routes and firewall were **not** modified.

| Item | Value |
|------|-------|
| Master netns | `vec01-master`, netns inode `net:[4026532705]` |
| Outstation netns | `vec01-outstation`, netns inode `net:[4026532776]` |
| Host/pod netns (unchanged) | `net:[4026532371]` |
| Lab subnet | `10.77.0.0/24` (non-routed illustrative) |
| Master lab addr | `10.77.0.1/24` on `veth-m` (peer `link-netns vec01-outstation`) |
| Outstation lab addr | `10.77.0.2/24` on `veth-o` (peer `link-netns vec01-master`) |
| Link addrs (planned) | master 3, outstation 4 (not yet activated) |
| Tooling | iproute2 6.15.0, iptables 1.8.11 (nf_tables), util-linux unshare 2.41.5 |

### 11.2 Objective isolation checks (all PASS)

| # | Check | Method | Result |
|---|-------|--------|--------|
| C1 | No external IPv4 route | passive `ip route get 8.8.8.8` | exit 2 (unreachable), both ns |
| C2 | No external IPv6 route | passive `ip -6 route get 2001:4860:4860::8888` | exit 2 (unreachable), both ns |
| C3 | No route to cluster API `10.3.0.1` | passive `ip route get` | exit 2 (unreachable), both ns |
| C4 | No route to pod gateway `169.254.1.1` | passive `ip route get` | exit 2 (unreachable), both ns |
| C5 | Lab-local reachability | `ping` 10.77.0.2 / 10.77.0.1 | exit 0 both directions |
| C6 | No listeners inside lab | `ss -tln` | 0 LISTEN sockets, both ns |
| C7 | Forwarding disabled | `sysctl net.ipv4.ip_forward` | `0`, both ns |
| C8 | DNS unreachable | `getent hosts example.com` | exit 2, both ns |
| C9 | No NAT table entries | `iptables -t nat -L -n` | empty, both ns |
| C10 | No bridge device | `ip link show type bridge` | none |
| C11 | veth peers are lab-internal | `ip -d link show` | `link-netns` = the other lab ns |
| C12 | No IPv6 globals | `ip -6 addr show scope global` | 0, both ns |
| C13 | Deny-all firewall (lab only) | `iptables -L -n -v` | OUTPUT/INPUT DROP except `10.77.0.0/24` + `lo` |

### 11.3 Physical-I/O absence (static)

- `/dev/ttyS0`, `/dev/ttyS1`, `/dev/ttyUSB0`, `/dev/ttyACM0`, `/dev/mem`,
  `/dev/port`: **absent**.
- No GPIO controllers exposed (`/sys/class/gpio/gpiochip*` absent).
- No field-bus / gateway path present. (Also confirmed in §9.4 for the peer.)

### 11.4 Host/pod boundary unchanged (no scope creep)

- Host netns inode still `net:[4026532371]`; host default route still
  `default via 169.254.1.1 dev eth0`; host `iptables` still empty; host listeners
  unchanged (agent runtime ports 60000/60001 only). No host/production route,
  firewall, or Kubernetes network policy was modified.

### 11.5 Evidence location and hashes (outside Git)

Durable path: `/workspace/vec01-lab-artifacts/gate-b-isolation-proof/`.

| File | SHA-256 |
|------|---------|
| `isolation-summary.txt` | `1203f57f25ddb7176db870f22ceca40883ee8d18fe4014f4c6ca355511af6313` |
| `isolation-checks.txt` | `8debda109aa51d2351774dd6984e371e47d56cf0ee664c27de50c3a1f6f5e2d7` |
| `lab-namespace-inventory.txt` | `77d285930b010149ce4396570f7911e867bb41192871827da7ef2b38082bf352` |
| `physical-io-and-boundary.txt` | `3fb52549b782fccaff953380e8971392d4a61c0d9afb66633835188b4314bce5` |
| `build-isolated-lab.sh` (reproducible construction) | `78d302534e3b29b160b2cac07dff5d6666a80777ede9983322f22cc6c81acca4` |
| `isolation-manifest.txt` | self-describing manifest of the above |

### 11.6 Scope compliance and hard prohibitions respected

- **No** DNP3 listener activation, **no** TCP/DNP3 packet, **no** control/CROB,
  **no** PCAP, **no** networked Go test suite execution, **no** peer-variation or
  simulated-BO adaptation, **no** production probe, **no** new branch/PR/merge,
  **no** R4 change. Only allowlisted docs updated.
- Both top markers remain `Status: TASK_BLOCKED`; CURRENT not archived/closed.

### 11.7 Honest limitations and next gate

- A PASS of **network isolation only** is **not** a PASS of Gate B overall.
- Still BLOCKED and separately authorized-only: B2 listener safety (proposed
  bind limited to `10.77.0.2:20000`, addrs 3/4), B3 peer v0 capability
  (G1V1/G30V1/G20V1) with a verifiable memory-backed **BO[0]** CROB target, B4
  bounds/retry policy, then Gates C/D/E and the mandatory Go gates.
- The lab namespaces are disposable and were left constructed (no listener).
  Teardown command: `sudo ip netns del vec01-master && sudo ip netns del vec01-outstation`.

**No protocol execution occurred. VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY
INTEROPERABILITY NOT PROVEN.**

---

## 12. Gate B substep 2 — static listener / link-address proof — 2026-11-08 (NOT_TESTED)

**Authority:** Owner-authorized "Gate B substep 2 — static listener and
link-address configuration only" addendum in CURRENT
`workflow/active_work/VEC-01-LAB-EXEC-001.md`. Scope: reverify the surviving
isolated lab and statically prove a prospective listener binding of exactly
`10.77.0.2:20000` with outstation link address 4 and master link address 3,
**without** activating any listener, sending any TCP/DNP3 packet or invoking any
control. Preconditions §2 of that addendum require a **live** reverification of
the substep-1 namespaces before any B2 work; if they are missing, the addendum
mandates **STOP TASK_BLOCKED** with `NETWORK_ISOLATION_NOT_TESTED`/`FAIL` and
forbids silently recreating them under B2-only authority. Execution SHA:
`bcec52ede1e7fe59c6a457426116ddc6efb5ad8b` (branch `work`).

**Verdict: `NETWORK_ISOLATION_NOT_TESTED`; `LISTENER_STATIC_CONFIG_NOT_TESTED`.**
Overall Gate B and the CURRENT remain `TASK_BLOCKED`. No B2 artifact was produced
because the mandatory preconditions did not hold.

### 12.1 Live precondition reverification (all FAIL / NOT_TESTED)

The substep-1 lab does **not** survive, and the current execution environment is
not the pod recorded in substep 1. Actual measured results:

| # | Precondition (addendum §2) | Method | Actual result | Verdict |
|---|----------------------------|--------|---------------|---------|
| P1 | `vec01-master` / `vec01-outstation` namespaces exist | `ls -la /run/netns`, `/var/run/netns` | both directories **absent**; no named netns | **FAIL** |
| P2 | Namespace identities match substep 1 | `readlink /proc/self/ns/net` | self inode `net:[4026532553]` ≠ recorded host `net:[4026532371]`; different pod | **FAIL** |
| P3 | Internal veth pair + lab addresses present | `/sys/class/net`, `ip` | only `eth0`, `lo`; **no** `veth-m`/`veth-o`; no `10.77.0.0/24` | **FAIL** |
| P4 | No public/OT/cluster/default route (IPv4) | `/proc/net/route` | default route `via 169.254.1.1 dev eth0` **present** (internet egress) | **FAIL** |
| P5 | Forwarding disabled | `sysctl net.ipv4.ip_forward` | `1` (enabled) | **FAIL** |
| P6 | Deny-all lab firewall | `nft list ruleset`, `iptables -L` | nft **empty**; iptables INPUT/FORWARD/OUTPUT policy **ACCEPT**, no rules | **FAIL** |
| P7 | No externally reachable published listener | `/proc/net/tcp{,6}` | 2 LISTEN sockets on `0.0.0.0` (agent runtime 60000/60001); no lab listener | n/a (host pod) |
| P8 | iproute2 available to build/inspect the lab | `command -v ip` | `ip` **MISSING** (iproute2 not installed) | **NOT_TESTED** |
| P9 | No physical I/O | `ls /dev/ttyS*` etc. | ttyS0/S1/USB0/ACM0, `/dev/mem`, `/dev/port` **absent** | PASS |
| P10 | Host/pod identity | `hostname` | `runtime-aozicfatuefownhn-86999879b-cmc2d` ≠ substep-1 `runtime-omgblkbssxnzsmru-8967c77bc-q48q7` | different environment |

Because the isolated lab is gone and the present host is a routed, forwarding,
internet-egress pod, **no isolation state can be reverified and no B2 listener
proof can be made here**. Per addendum §2 the namespaces must **not** be silently
recreated under this B2-only authorization, and no listener/static bind test was
attempted.

### 12.2 Retained peer artifacts (Gate A precondition) — ABSENT

- Durable store `/workspace/vec01-lab-artifacts/` currently contains **only** this
  substep's reverification directory; the historical `gate-b-isolation-proof/`
  and `gate-b-reviewer-packet/` subdirectories are **ABSENT**.
- `outstation-demo` and `libopendnp3.a` are **absent** everywhere on the
  filesystem (searched to depth 6); `/tmp/vec01-lab` is **absent**.
- Therefore the pinned peer bytes cannot be rehashed and no peer
  configuration/binary can be inspected or prepared. The addendum §3 requires
  validating these before edits; without them B2 preparation is not possible.
- Go toolchain: **MISSING** (`go: command not found`), so no Go gate could run
  (consistent with prior NOT_RUN records; not converted to PASS).

### 12.3 Newly retained evidence (outside Git)

Durable path: `/workspace/vec01-lab-artifacts/gate-b-substep2-reverify/`.

| File | SHA-256 |
|------|---------|
| `live-netns-inventory.txt` | `e9e66f327a224df81535a1f6a6ec86f76a426700442c566d59fe8d81cb7ff04f` |
| `live-network-and-boundary.txt` | `86bd6b581d25a278a4f5659c35d6eab3d793a49a175bb991555c9d82dc216cd9` |
| `live-firewall-io-and-artifacts.txt` | `005c2e8e91d8620072431029122708a79e21487c311e3c105d60183a023ebfe6` |
| `SHA256SUMS.txt` | self-describing manifest of the above |

### 12.4 Scope compliance and hard prohibitions respected

- **No** listener activation, **no** TCP/DNP3 packet, **no** control/CROB,
  **no** PCAP, **no** networked Go test suite, **no** peer-variation or
  simulated-BO adaptation, **no** namespace recreation, **no** production probe,
  **no** new branch/PR/merge, **no** R4 change.
- Only allowlisted documentation was updated; both top markers remain
  `Status: TASK_BLOCKED`; CURRENT not archived/closed.

### 12.5 Honest limitation and next gate

- `LISTENER_STATIC_CONFIG_PASS` requires verified surviving isolation **and**
  static confirmation of all endpoint/address settings. Neither condition can be
  met in the present environment, so B2 is recorded **NOT_TESTED**, not FAIL of
  the design, and not a PASS.
- B2 remains separately authorized-only. Resumption prerequisites: a dedicated
  isolated, non-routed runtime (no public/OT/cluster route, forwarding disabled,
  deny-all lab policy) with iproute2 available; the pinned OpenDNP3 3.1.2 peer
  bytes restored to the durable store; then a fresh owner-authorized B2 step can
  prepare and statically prove `10.77.0.2:20000` with link addresses 3/4.
- B3 (peer v0 variations + memory-backed BO[0]) and B4 (bounds/retry), then Gates
  C/D/E and the mandatory Go gates, remain not authorized and not performed.

**No protocol execution occurred. VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY
INTEROPERABILITY NOT PROVEN.**
