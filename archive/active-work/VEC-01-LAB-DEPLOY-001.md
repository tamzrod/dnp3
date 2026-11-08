<!-- Archived after verified acceptance. Original task body preserved verbatim below. -->

Archived: 2026-11-08
Execution commit SHA: 5766bd1c4c3577d67b455201dbbb102026d48a11
Promotion baseline: f21f32cf5d0631c8f50e6b3f398bffe975bbb6ab (merged PR #20)
Verification evidence: docs/verification/VEC-01-LAB-DEPLOY-001.md
Go toolchain: go1.22.12 linux/amd64 (private install $HOME/sdk/go)
Lab artifacts (outside repo): /tmp/vec01-lab (outstation-demo sha256 8ae8c84125c4920352b6daef4145e2a91de9e02cd6e2c1b94bf6995ebeb093bb)

# VEC-01-LAB-DEPLOY-001 — Isolated OpenDNP3 peer acquisition, build and static smoke gates

Status: CURRENT — HUMAN_PROMOTED
Package: VEC-01-LAB-DEPLOY-001
Branch: work
Human authorization: explicit owner promotion on 2026-11-08
Approval boundary: offline/isolated laboratory deployment preparation ONLY; not third-party protocol validation, not VEC-01 execution
Promotion baseline: `f21f32cf5d0631c8f50e6b3f398bffe975bbb6ab` on permanent `work` (merged PR #20)

## Objective
Prepare a **reproducible, independently sourced OpenDNP3 3.1.2 C++ outstation build artifact** and **reviewable isolated-laboratory configuration** for a future independently authorized VEC-01 campaign. Pinned upstream commit: `26b4c01e4839bbbda8866655e086471c4917ee53` (`dnp3/opendnp3`, tag `3.1.2`, archived, Apache-2.0). Validate artifact hashes, transitive build dependencies/licensing, containment and static smoke checks. Produce an auditable deployment-readiness report. **Do not start any DNP3 listener, connect a DNP3 master/outstation, transmit protocol frames, capture DNP3 traffic or issue any control.** VEC-01 remains **NOT EXECUTED**, residual R4 **OPEN**, third-party interoperability **NOT PROVEN**.

## Mandatory startup / branch readiness
1. Read `AGENTS.md`, `operation cwal.md`, `BLACK_SHEEP_WALL.md`; read only the top six authority fields in `workflow/active_work/README.md` and root `handoff.md`; confirm both identify exactly this CURRENT and permanent branch `work`, QUEUED NONE, TASK_READY. Read only this CURRENT body and named relevant source documents.
2. `git fetch origin`; verify local state is clean or safely preservable; verify remote `main`, remote `work`, and their ancestry. **Main synchronization belongs to the startup readiness gate, not a separate task**: if `work` is behind `main` and an exact non-force fast-forward of `work` is possible, synchronize **before editing CURRENT** and reverify markers/task presence. If promoted CURRENT already causes divergence, no force-push, rebase, merge commit, reset or overwrite; report TASK_BLOCKED and seek owner instruction. Preserve human promotion commits.
3. Record `git rev-parse HEAD`, `origin/work`, `origin/main` and merge base; no moving to a new task branch. Run `bash scripts/cwal-check.sh` **on the clean remote-synced checkout** before any edits; require `CWAL CHECK OK`. Stop on mismatched markers, unexpected worktree, or branch divergence.
4. Read only `active_work/vec-01-opendnp3-lab-preparation.md`, `active_work/third-party-capture-procedure.md`, `active_work/supported-profile.md`, and `active_work/testdata/external/FORMAT.md` as needed for preparation. Follow BLACK SHEEP WALL, no QUEUED bodies. KDE authority remains unchanged.

## Strictly bounded external preparation
External acquisition/compilation artifacts belong in a **separate, disposable, explicitly named lab workspace outside the Git checkout**. No external repositories may be installed as a new dependency of `tamzrod/dnp3`.

### A. Independent source + dependency provenance
- Verify that `https://github.com/dnp3/opendnp3` is independent of this Go library, upstream archived and Apache-2.0. Pin tag `3.1.2` and **verify exact commit** `26b4c01e4839bbbda8866655e086471c4917ee53`; no floating `release`, main, or substitution.
- Controlled acquisition on a non-OT staging host is allowed in this CURRENT. Record exact URLs, source archive hash (SHA-256 computed from bytes actually obtained), provenance and acquired tarball/checksum if applicable. A Git commit SHA is **not** an archive SHA-256; do not invent/check in unmeasured hashes.
- Inspect upstream tag's CMake/FetchContent dependency sources and pinned versions (including ASIO `asio-1-16-0`, exe4cpp `fb878a4de598ba9d6e4338afebf83f96e03af1b8`, ser4cpp `3c449734dc530a8f465eb0982de29165cc4e23d5`); verify upstream URL, archive bytes' SHA-256, integrity against declared upstream checksums, license/SPDX and full transitive inventory. No silent substitution, unreviewed network fetch or floating dependency resolution. A source-file SHA1 used by upstream is not sufficient as the sole artifact-integrity record.
- Do not proceed to compilation until all build inputs have verified provenance and license review. Archived/outdated dependencies are a documented risk; any unresolved dependency, license or toolchain issue is TASK_BLOCKED, not an authorization to improvise.

### B. Hermetic/isolated build and environment
- Stage reviewed source/dependencies into an isolated offline VM, disconnected host or equivalent container/build sandbox with no OT/SCADA/field network route and **no network fetch during compilation**. No production credentials, mounts, shared secrets, physical I/O, host-network mode, privileged container or uncontrolled published ports.
- Confirm reproducible OS/architecture, compiler, linker, CMake versions; pinned `3.1.2` `CMakeLists.txt` has `DNP3_EXAMPLES` (not `DNP3_DEMO`) and target `outstation-demo`. CMake minimum `3.11`. Build only the native C++ outstation/demo and essential library; keep TLS, tests, Java/.NET, fuzz and unnecessary optional features off. Explicitly disable/avoid auto-fetch, remote execution and auxiliary protocol traffic.
- Store generated C++ binaries, lock manifest, dependency archives, exact build command/output, configuration, SHA-256 and any logs **outside tracked repository paths**. If an isolated build runner is unavailable, stop rather than building on a host with production OT reachability.
- Prepare a lab-only network design and configuration **without activating a DNP3 listener**: dedicated offline VM or validated internal-only network; hypothetical `10.77.0.0/24`, master/outstation link addresses `3/4`, TCP port `20000` (unpublished); outstation simulated memory-backed points only; no bridging/forwarding/host networking, privileged access, or reachability to operational systems. Inspect static configs/routes/firewall/network namespace offline. Do not probe production addresses.

### C. Static-only smoke gate
**Allowed:** inspect source commit, pinned dependency hashes and license manifests; CMake configure; compile `outstation-demo`; verify binary file type, ELF metadata, linker/dynamic-library dependencies, file mode and hashes; examine proposed config files and network isolation settings; demonstrate offline build does not require internet; verify no published ports in static definitions; optionally check version/help **only after positively verifying it cannot bind a socket, open an interface, or send any traffic**.
**Not allowed:** starting `outstation-demo` if it opens/binds a socket; running a DNP3 server/client; TCP handshake, link reset/status, Class-0 reads, multi-header traffic, CROB, negative-frame injection, packet replay, protocol capture, or any real/field device connection. A process that unexpectedly opens a listener, reaches the network, or attempts control is an immediate safety stop and TASK_BLOCKED. No network-connected functional smoke test in this CURRENT.

### D. Future VEC-01 readiness plan, not execution
- Specify static configuration candidates for TCP, G60V1 Class-0 request, G1V1 binary, G30V1 analog, G20V1 counter and G12V1 Direct-Operate (simulator-only); use `TO VERIFY` for variation/qualifier behavior, peer command status, multi-header packing or anything not demonstrated. Do **not** claim these are proven.
- Document a separate human-reviewed handoff for later lab activation and **separately owner-promoted VEC-01-LAB-EXEC-001** before any wire exchange. Future evidence plan follows `active_work/testdata/external/FORMAT.md` and `active_work/third-party-capture-procedure.md`; no PCAP/`.vec` fabricated or accepted as third-party proof now.

## Exact tracked repository edit allowlist during CURRENT execution
1. **New** `active_work/vec-01-opendnp3-lab-deployment-readiness.md` — static-source, dependency, isolation and build evidence; clearly NO PROTOCOL EXECUTION.
2. **New** `docs/verification/VEC-01-LAB-DEPLOY-001.md` — actual executed regression/smoke commands, toolchain, SHA, exit codes, hashes, artifacts locations and limitations.
3. **New** `archive/active-work/VEC-01-LAB-DEPLOY-001.md` — original human-promoted task preserved verbatim beneath a clearly marked archive preamble.
4. `workflow/active_work/README.md` and root `handoff.md` — matching CURRENT closeout markers and concise completion notes only.
5. `workflow/active_work/VEC-01-LAB-DEPLOY-001.md` — removal **only after** successful archival and all gates pass.

**Promotion-only paths:** task file and the two matching authority marker files. **All other tracked files are read-only.** No Go source, tests, fixtures, `go.mod`/`go.sum`, CI, existing scripts, Compose/Containerfile, dependency vendor tree, profiles, governance/KDE, roadmap, README claims or historical evidence modifications. No commits of external source archives, binaries, packet captures or credentials. If extra paths are needed, stop and request new owner-approved CURRENT.

## Mandatory executable regression verification
On the exact synchronized `work` checkout (not inferred from past runs), provision official checksum-verified compatible Go if missing and record `go version`. Actually run all and record exit codes and meaningful outputs, no NOT_RUN accepted as successful completion:
```sh
bash scripts/cwal-check.sh  # clean startup and clean post-commit checkpoint
test -z "$(gofmt -l .)"
git diff --check
go version
go vet ./...
go build ./...
go test ./... -count=1
go test -race ./internal/outstation/... -count=1
./scripts/verify-mvp.sh
./scripts/verify-external-mvp.sh
```
Mandatory lab checks (report commands, outputs and exit codes): pinned Git commit/tag verification; checksums computed for **actually acquired** source/dependency artifacts; complete license inventory; offline CMake configuration/build output; `outstation-demo` binary identity/dependency inspection; offline network/isolation static inspection; absence of DNP3 protocol activity. If any is not run or fails, report TASK_BLOCKED, not fabricated PASS. The existing Go MVP tests are allowed regression checks only and **do not** satisfy independent VEC-01.

`scripts/cwal-check.sh` requires clean tracked worktree; execute before editing and again after clean checkpoint, not as a mid-edit claim.

## Safety stops / acceptance
- No uncontrolled outbound networking, runtime DNP3 listener, live protocol packet, field control, operational-network route or real hardware I/O. Stop on isolation breach or uncertain binary activation behavior.
- Source/tag/commit, all transitive dependency provenance/licensing and actual artifact SHA-256 are independently checkable; build is reproducible offline and artifact build logs truthful.
- A static laboratory configuration and manifest are reviewed with simulated endpoints only; no DNP3 master/outstation session was initiated.
- All mandatory executable regression and static build gates actually pass; tests/commands not run must be marked NOT_RUN and prevent completion.
- Upon complete acceptance: preserve original task body verbatim in archive, write verification evidence, close both markers to `CURRENT: NONE`, `QUEUED: NONE`, `Status: WORK_PACKAGE_READY`, commit and push only to permanent `work`, and independently fetch/read back remote SHA and evidence before reporting `TASK_CHECKPOINTED`.
- On blocker: keep CURRENT, record truthful TASK_BLOCKED and STOP. Do not start next task autonomously.
- **VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY INTEROPERABILITY NOT PROVEN**; no IEEE 1815 certification, full conformance, or production-readiness claim; no automatic PR/merge/release/deployment into OT/field environment or force push.
