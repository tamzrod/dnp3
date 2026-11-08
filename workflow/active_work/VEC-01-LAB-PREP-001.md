# VEC-01-LAB-PREP-001 — Independent OpenDNP3 peer laboratory preparation plan

Status: CURRENT — HUMAN_PROMOTED
Package: VEC-01-LAB-PREP-001
Branch: work
Authorized: explicit human promotion on 2026-11-08
Approval class: documentation-only laboratory preparation; no runtime lab operation or DNP3 network traffic authorized
Promotion baseline: `f165e8509436153dfa53ab68bb67c51be9a88b10` (merged PR #19), permanent `work`

## Objective and strict task boundary
Produce **one standalone, reviewable planning document**, `active_work/vec-01-opendnp3-lab-preparation.md`, specifying how an independent, explicitly pinned **OpenDNP3** C++ outstation *could later* be prepared for VEC-01 interoperability testing with `tamzrod/dnp3` Go master. Describe peer provenance, reproducible configuration, isolated topology, dry-review validation gates, an authorized-future-execution matrix and evidence retention. This task is **plan-only**, not lab deployment. R4/VEC-01 remain OPEN/NOT EXECUTED/THIRD-PARTY INTEROP NOT PROVEN.

## Startup and authority
1. Read `AGENTS.md`, `operation cwal.md`, `BLACK_SHEEP_WALL.md`, and only the top current-state markers in `workflow/active_work/README.md` and root `handoff.md`.
2. Confirm both blocks' six fields match this CURRENT, permanent `work`, QUEUED NONE and TASK_READY. Fetch and safely synchronize remote `work`, verify clean worktree, then run `bash scripts/cwal-check.sh` on clean startup.
3. Read only CURRENT body, `active_work/third-party-capture-procedure.md`, `active_work/supported-profile.md`, `active_work/testdata/external/FORMAT.md`, and the relevant minimum source/fixture excerpts if needed. Do not read future task bodies.
4. If baseline or governance differs, STOP and report; no autonomous task widening.

## Deliverable requirements
Write `active_work/vec-01-opendnp3-lab-preparation.md`, with explicit headers and evidence fields for:
1. **Independent peer provenance:** OpenDNP3 upstream URL, independently implemented C++ lineage distinct from this Go repository, actual license and archived maintenance status; propose a pinned official version/tag/source SHA and artifact checksum only if verifiable. Do not invent hash, release, licensing terms, or exact feature support. Include peer eligibility review and third-party dependency provenance.
2. **Reproducible preparation design:** host OS/architectures, Go 1.22 compatibility, C++ compiler/CMake and OpenDNP3 build/demo requirements; document planned commands as `NOT EXECUTED` and expected artifacts without fetching, building, running or installing a peer.
3. **Isolation design:** either dedicated offline VM/host or internal-only container network (e.g., sample private test subnet `10.77.0.0/24`), no published DNP3 ports, no production OT/SCADA reachability, no routing/forwarding/privileged container/host networking, no production credentials, simulator-only memory-backed points, explicit review checklist and separate permission to perform any connection.
4. **Configuration matrix:** hypothetical TCP port 20000, master/outstation IDs 3/4, point indices and known G1V1/G30V1/G20V1 values, v0 G60V1 Class-0 poll, G12V1 Direct Operate to a simulated output only. Mark variations and multi-header behavior `TO VERIFY` where OpenDNP3 release/configuration has not been demonstrated. No unsupported SBO, unsolicited, serial, TLS, SA or event-class features.
5. **Future test matrix, NOT execution:** link reset/status, Class-0 reads, multi-header check, G12V1 CROB result/status, clean close, bad CRC, mismatched address, connection drop, command rejection and incomplete capture. Define prerequisites, expected evidence, stop criteria, failure and ambiguous-command classifications; never blindly retry uncertain control outcomes.
6. **Capture and artifact plan:** passive TShark/tcpdump and application logs, raw unmodified PCAP/PCAPNG, `.vec` matching the existing `FORMAT.md` including direction/layer and original wire bytes, separate redacted sidecars, clocks/timestamp correlation, stack/build SHA, source artifact SHA-256, file hashes, config and reviewer-independent ACCEPT/REJECT/INCOMPLETE mapping. No fabricated captures or synthetic evidence counted for R4.
7. **Authorization gates and risks:** identify future separate human-promoted CWAL for deployment and another for actual VEC-01 test execution, verify source licensing for candidate and transitive dependencies, isolate nonproduction network, restrict simulator outputs, separate review before claiming any interoperability. State lab plan cannot close R4 or establish IEEE 1815 conformance, certification or production readiness.

## Exact edit allowlist
**During this preparation/promotion:** `workflow/active_work/VEC-01-LAB-PREP-001.md`, `workflow/active_work/README.md`, root `handoff.md` only.

**When JR executes CURRENT:**
- New plan: `active_work/vec-01-opendnp3-lab-preparation.md`
- Completion evidence: `docs/verification/VEC-01-LAB-PREP-001.md`
- Verbatim archival: `archive/active-work/VEC-01-LAB-PREP-001.md`
- Closeout marker files: `workflow/active_work/README.md`, root `handoff.md`
- Remove `workflow/active_work/VEC-01-LAB-PREP-001.md` only after archive saved.

**Everything else is read-only.** No Go implementation, fixtures, `go.mod`/`go.sum`, CI/scripts, runnable Compose files, executable peer binaries, dependency locks, KDE governance, roadmap, supported profile, historical verification, production configuration or README claim changes.

## Absolute prohibitions
No clone, download, compile, installation or invocation of OpenDNP3 or another DNP3 peer as part of this task. No Docker Compose execution, network namespace deployment, port exposure, Wireshark capture, packet replay, live TCP handshake, DNP3 requests, loopback simulation test, control command, field-device access, SCADA/OT network probing, or third-party lab campaign. Existing Go tests and the existing repository MVP verification scripts are expressly permitted as **mandatory executable regression gates only**; they do not count as VEC-01 or authorize a new peer connection. Do not publish secrets or operational addresses.

## Mandatory executable regression verification — all must PASS
Run on JR's exact `work` checkout; record actual exit codes, Go version, execution commit/tree SHA and meaningful outputs. No NOT_RUN allowed:
```sh
bash scripts/cwal-check.sh
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
Run `cwal-check.sh` before edits on a clean tree and re-run after clean commit/checkpoint; report that mid-edit dirty tree is not an authorized PASS gate. If no Go is present, use verified official compatible Go globally where safe, then privately in JR's user directory, then an isolated Go runner with exact checkout. Verify checksum of downloaded official Go archive. Do not substitute older recorded results for new execution, and do not misrepresent in-repo external MVP Tiers 1–3 as third-party VEC-01 proof.

## Acceptance and closeout
- The standalone plan covers all seven deliverable requirements and explicitly labels **PLANNING ONLY / VEC-01 NOT EXECUTED / R4 OPEN / THIRD-PARTY INTEROPERABILITY NOT PROVEN**.
- OpenDNP3 is independently sourced and reviewed at the *plan* level; no unverified release/variation or working setup is claimed.
- Zero peer deployment, capture, field control or production network interaction; all changed tracked paths in allowlist, matching six-field CWAL authority markers.
- All actual mandatory regression gates pass with new evidence. Archive task body verbatim, add verification, close markers **CURRENT: NONE / QUEUED: NONE / Status: WORK_PACKAGE_READY**, commit/push only to `work`, and remote read-back exact head SHA and evidence.
- On any failure, unknown independent peer/license issue, inability to run mandatory gates, scope expansion, or safety concern: keep CURRENT, mark **TASK_BLOCKED** with truthful evidence, stop. No silent skipping to queued tasks.
- No PR, merge, direct `main` push, release, force-push, deployment, live tests, or DNP3 device commands authorized.
