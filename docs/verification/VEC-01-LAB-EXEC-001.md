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
