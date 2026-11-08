# DNP3 — CURRENT CWAL Handoff

Repository: tamzrod/dnp3
Package: VEC-01-LAB-EXEC-001
Branch: work
CURRENT: VEC-01-LAB-EXEC-001
QUEUED: NONE
Status: TASK_READY








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

## Persistent branch rule (owner instruction, 2026-11-08)
Use `work` for the remaining active roadmap, including subsequent separately human-approved CWAL packages. Do not switch branches by package. A stable branch name does not bypass task authorization, PR audit against `main`, CI, or merge approval. The superseded `docs/sync-mext-post-pr17-20261108` ref is not the active CWAL branch; retire it after confirming the new `work` tip.
