# Verification — VEC-01-LAB-DEPLOY-001

Package: VEC-01-LAB-DEPLOY-001
Branch: work
Task: VEC-01-LAB-DEPLOY-001 — Isolated OpenDNP3 peer acquisition, build and static smoke gates
Execution commit SHA: `5766bd1c4c3577d67b455201dbbb102026d48a11`
Promotion baseline: `f21f32cf5d0631c8f50e6b3f398bffe975bbb6ab` (merged PR #20)
Lab workspace (outside the Git checkout): `/tmp/vec01-lab`

## Purpose

Prepare a reproducible, independently sourced OpenDNP3 3.1.2 C++ outstation
build artifact and a reviewable isolated-laboratory configuration, with
dependency provenance/licensing and static smoke validation. No DNP3 listener
was started, no master/outstation session was opened, no protocol frames were
transmitted, no traffic was captured and no control was issued. **VEC-01 NOT
EXECUTED / R4 OPEN / THIRD-PARTY INTEROPERABILITY NOT PROVEN.**

## Deliverables

- `active_work/vec-01-opendnp3-lab-deployment-readiness.md` (new) — static
  source/dependency/isolation/build evidence, explicitly NO PROTOCOL EXECUTION.
- `docs/verification/VEC-01-LAB-DEPLOY-001.md` (this file) — executed commands,
  toolchain, SHA, exit codes, hashes, artifact locations and limitations.
- `archive/active-work/VEC-01-LAB-DEPLOY-001.md` (new) — original task body
  verbatim under an archive preamble.
- `workflow/active_work/README.md`, `handoff.md` — matching closeout markers.
- `workflow/active_work/VEC-01-LAB-DEPLOY-001.md` removed after archival.

No Go source/tests/fixtures, `go.mod`/`go.sum`, CI, scripts, Compose/Containerfile,
vendor tree, profiles, governance, roadmap, README claims or historical evidence
were changed. No external source archive, binary, capture or credential was
committed; the repo checkout stayed clean.

## Startup / branch readiness

- `git fetch origin`; local `work` fast-forwarded to `origin/work` =
  `5766bd1c4c3577d67b455201dbbb102026d48a11` (no force, no rebase, no merge
  commit, no reset; promotion commits preserved).
- Markers in `workflow/active_work/README.md` and `handoff.md` match:
  Repository `tamzrod/dnp3`, Package `VEC-01-LAB-DEPLOY-001`, Branch `work`,
  CURRENT `VEC-01-LAB-DEPLOY-001`, QUEUED `NONE`, Status `TASK_READY`.
- `origin/main` = `95ac10289ead46c09264a9da2bb877f0e0507410`; merge base of
  `work` and `main` = `95ac1028…` (main is an ancestor of `work`).
- `bash scripts/cwal-check.sh` on the clean, remote-synced checkout →
  `CWAL CHECK OK: CURRENT=VEC-01-LAB-DEPLOY-001 branch=work`, exit 0.
- `git status --porcelain` clean before editing.

## Toolchain provisioning

- Go: private `$HOME/sdk/go` (go1.22.12 linux/amd64), official archive SHA-256
  `4fa4f869b0f7fc6bb1eb2660e74657fbf04cdd290b5aef905585c86051b34d43` verified.
- CMake: **not present** initially; installed `cmake` 3.31.6-2 via
  `apt-get` (global, passwordless sudo available). Satisfies the ≥ 3.11 minimum.
- `file` was absent; installed via `apt-get` for binary identification.
- C++ toolchain: g++ 14.2.0 (already present).

## External preparation — executed commands and results

Lab workspace `/tmp/vec01-lab` (disposable, outside the Git checkout):
`downloads/`, `depsrc/`, `src/`, `build-offline/`, `artifacts/`, `config/`, `logs/`.

### A. Source + dependency provenance

| Command | Result | Exit |
|---------|--------|------|
| `git ls-remote https://github.com/dnp3/opendnp3.git 'refs/tags/3.1.2*'` | `refs/tags/3.1.2` = `523f7d2d…`; `refs/tags/3.1.2^{}` = **`26b4c01e4839bbbda8866655e086471c4917ee53`** (matches pin) | 0 |
| `curl -sSL https://codeload.github.com/dnp3/opendnp3/tar.gz/refs/tags/3.1.2` | acquired tarball (610113 bytes) | 0 |
| `sha256sum opendnp3-3.1.2.tar.gz` | `183cc29222c3cb58099a9753c43defed5cb8677fda985c3f88e38b5c19a36ff2` | 0 |
| download asio/exe4cpp/ser4cpp/catch archives | 4 archives acquired | 0 |
| `sha256sum` + `sha1sum` on each | SHA-256 recorded; SHA-1 cross-checks **match** upstream declared `URL_HASH` exactly | 0 |

SHA-256 of actually obtained bytes (also in `/tmp/vec01-lab/artifacts/sha256-manifest.txt`):

- opendnp3-3.1.2.tar.gz — `183cc29222c3cb58099a9753c43defed5cb8677fda985c3f88e38b5c19a36ff2`
- asio-1-16-0.zip — `18883c350558eadf91218ae5ce9a8f8f85099c2f036c1c8579fb639092e928c5`
- exe4cpp-fb878a4.zip — `2a8a2c7d9982f54858b92c03cff91ef433907472039f8e61ff7c9f9b3d48db6c`
- ser4cpp-3c44973.zip — `5a78f83de2d70bdf4be2ab607bc4416c6866ce64a257a8f19948cf5746762f6e`
- catch-2.11.3.hpp — `197640c056d6d01e5a56737fbdfe671d2b872c8c7533e01cc2de4be19ab7a028`

Licenses: OpenDNP3 Apache-2.0; exe4cpp BSD-3-Clause; ser4cpp BSD-3-Clause; asio
Boost-1.0; catch Boost-1.0 (test-only, unused). Full inventory in the readiness
document.

### B. Hermetic offline build

| Command | Result | Exit |
|---------|--------|------|
| `sudo unshare -n` network probe | `NO NETWORK`, no DNS (isolation confirmed) | 0 |
| `cmake -S src/opendnp3-3.1.2 -B build-offline … -DFETCHCONTENT_FULLY_DISCONNECTED=ON -DFETCHCONTENT_SOURCE_DIR_{ASIO,EXE4CPP,SER4CPP}=…` (inside netns) | `Configuring done`, `Generating done` | 0 |
| `cmake --build build-offline --target outstation-demo -j4` (inside netns) | `Built target opendnp3`; `Built target outstation-demo` | 0 |

Features: `DNP3_TLS=OFF DNP3_TESTS=OFF DNP3_JAVA=OFF DNP3_FUZZING=OFF
DNP3_EXAMPLES=ON DNP3_STATIC_LIBS=ON`. No network fetch occurred during
configure/compile (fully disconnected + local sources). A second fully clean
offline build reproduced a **byte-identical** binary.

### C. Static-only smoke gate

| Command | Result | Exit |
|---------|--------|------|
| `file outstation-demo` | ELF 64-bit LSB pie executable, x86-64, dynamically linked, not stripped | 0 |
| `readelf -h` | ELF64, Type DYN (PIE), Machine x86-64 | 0 |
| `ldd outstation-demo` | only `libstdc++/libm/libgcc_s/libc` + vdso + ld-linux | 0 |
| `readelf -d` | NEEDED = the four standard libs; no RPATH/RUNPATH | 0 |
| `sha256sum outstation-demo` | `8ae8c84125c4920352b6daef4145e2a91de9e02cd6e2c1b94bf6995ebeb093bb` | 0 |
| `sha256sum libopendnp3.a` | `e2926474288fb8d2c6c7962192bad2ff0f724f63c07d2209fe5b30c26fbe0677` | 0 |
| `/proc/net/tcp` scan for port 20000 (0x4E20) | **no listener** | 0 |
| `pgrep -af outstation-demo` | not running | 1 (no match, expected) |

**The demo was NOT run** — its `main()` binds `0.0.0.0:20000`
(`cpp/examples/outstation/main.cpp` @ `26b4c01`). Static inspection only.

### D. Lab configuration (static)

Prepared `/tmp/vec01-lab/config/{lab-network-design.md,outstation-lab-config.notes}`;
static inspection of interfaces (`lo`, `eth0`) and absence of a port-20000
listener. No interface configured, no production address probed.

## Mandatory executable regression gates — all executed, all PASS

Environment: Debian 13 x86_64, Go 1.22.12, execution SHA
`5766bd1c4c3577d67b455201dbbb102026d48a11`.

| # | Command | Result | Exit |
|---|---------|--------|------|
| 1 | `bash scripts/cwal-check.sh` | PASS on clean startup; re-run PASS post-commit | 0 |
| 2 | `test -z "$(gofmt -l .)"` | PASS — empty output | 0 |
| 3 | `git diff --check` | PASS | 0 |
| 4 | `go version` | `go1.22.12 linux/amd64` | 0 |
| 5 | `go vet ./...` | PASS — no findings | 0 |
| 6 | `go build ./...` | PASS | 0 |
| 7 | `go test ./... -count=1` | PASS — all listed packages `ok` | 0 |
| 8 | `go test -race ./internal/outstation/... -count=1` | PASS — `ok dnp3/internal/outstation` | 0 |
| 9 | `./scripts/verify-mvp.sh` | PASS | 0 |
| 10 | `./scripts/verify-external-mvp.sh` | PASS — Tiers 1–3; VEC-01 advisory only | 0 |

Note: `scripts/cwal-check.sh` requires a clean tracked worktree, so it fails
mid-edit while the new deliverables are uncommitted; that is not an authorized
gate. It passed on clean startup and is re-run after the checkpoint commit.

## Acceptance mapping

- Source pinned to tag `3.1.2` at commit `26b4c01e…` (verified); acquired
  archive SHA-256 computed from actual bytes; no invented hashes.
- All transitive dependencies acquired, hashed (SHA-256) and SHA-1 cross-checked
  against upstream declared values; complete license inventory; no GPL/AGPL.
- Build is hermetic (no-network namespace, fully disconnected FetchContent,
  local dependency sources) and reproducible (byte-identical rebuild).
- Static smoke gate passed; the socket-binding demo was not executed; no
  listener on port 20000; no protocol activity.
- Isolated lab configuration prepared as static inspection only; simulated
  endpoints only; no session initiated.
- All mandatory Go regression gates actually PASS with new evidence; no NOT_RUN.
- Repo checkout stayed clean; only allowlisted Markdown paths change on closeout.

## Limitations / non-claims

- **VEC-01 NOT EXECUTED; R4 OPEN; third-party interoperability NOT PROVEN.**
- The static build artifact is not protocol interoperability evidence.
- No IEEE 1815 certification, full-conformance or production-readiness claim.
- Upstream archived; dependencies pinned to 2022-era revisions (documented risk).
- Lab artifacts remain outside the repository and are not committed.
- `verify-external-mvp.sh` PASS covers in-repo Tiers 1–3 only.
- Any wire exchange requires a separately owner-promoted VEC-01-LAB-EXEC-001.
