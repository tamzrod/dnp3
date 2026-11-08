# VEC-01-LAB-EXEC-001 — Isolated independent OpenDNP3 interoperability execution and capture review

Status: CURRENT — HUMAN_PROMOTED
Package: VEC-01-LAB-EXEC-001
Branch: work
Human authorization: owner expressly promoted this bounded laboratory execution campaign on 2026-11-08
Promotion baseline: `02a58f76cfd8b82cbea1bcddcb36f61ce2738b21` (merged PR #21, main fast-forwarded into permanent work)
Authority boundary: **non-production isolated simulated test only**, gated by independent isolation/output review before first protocol packet. Promotion does NOT certify interoperability or authorize a production connection.

## Objective and support lock
Execute a bounded, reproducible, **real-wire TCP** campaign between `tamzrod/dnp3` native Go **master** and independently implemented C++ **OpenDNP3 3.1.2 outstation**, derived from pinned upstream commit `26b4c01e4839bbbda8866655e086471c4917ee53`. Capture immutable original traffic, valid trace-derived reviewable `.vec` records and logs. Obtain a reviewer-independent `ACCEPT`, `REJECT` or `INCOMPLETE` disposition without overclaiming.

Strict v0 profile only: TCP one master ↔ one outstation, Class-0 G60V1, static G1V1 Binary Inputs, G30V1 Analog Inputs and G20V1 Counters, G12V1 Direct-Operate CROB **to simulated memory-backed output only**, link reset/status, lifecycle; no SBO, unsolicited, event classes, DNP3 serial, TLS, secure authentication, file transfer, time sync, freeze, restart control, new variations or standards-certification claims. Reference `active_work/supported-profile.md` as authority.

**At promotion: VEC-01 NOT EXECUTED; R4 OPEN; third-party interoperability NOT PROVEN.** A successful test is evidence *only* for this peer version/configuration; independent reviewer must accept evidence before any scoped interoperability statement. Closing R4, changing supported profile, production use, and feature expansion require **separate** owner approval.

## CWAL startup and main synchronization (readiness, not standalone work)
1. Read `AGENTS.md`, `operation cwal.md`, `BLACK_SHEEP_WALL.md` and only the top six authority markers in `workflow/active_work/README.md` and root `handoff.md`. Both must match exactly: repository, this package, branch `work`, CURRENT `VEC-01-LAB-EXEC-001`, QUEUED NONE, Status TASK_READY.
2. `git fetch origin`; verify the clean/safely preservable local worktree, remote `work` and `main`, exact CURRENT task and ancestry. **Synchronization is part of readiness**: only fast-forward where it preserves the promoted CURRENT and all commits; if main advances incompatibly with promoted CURRENT, stop on divergence and obtain owner resolution. No reset, forced checkout, rebase, merge commit, force-push, task branch, promotion replacement or `main` push.
3. `bash scripts/cwal-check.sh` must report CWAL CHECK OK on a clean synced CURRENT tree before any activity. Record exact remote SHAs and local execution SHA. Use BLACK SHEEP WALL: read just CURRENT and named VEC-01 plan/capture/profile/format/evidence plus minimal test APIs and peer source configuration.
4. Verify post-PR #21 integration baseline and latest peer/master provenance. An undocumented change of master code, dependency, binary, configuration or peer version is TASK_BLOCKED.

## Gate A — retain/recover or reproducibly rebuild actual peer
- Existing disposable staging workspace `/tmp/vec01-lab` may have vanished. Before any network listener, locate/read-only inventory the original source/dependency archives, `outstation-demo`, `libopendnp3.a`, config and logs; compute hashes of actual bytes. Never assume an earlier SHA implies existing bytes.
- Verify OpenDNP3 3.1.2 peeled tag commit `26b4c01e4839bbbda8866655e086471c4917ee53` and historical source tarball SHA256 `183cc29222c3cb58099a9753c43defed5cb8677fda985c3f88e38b5c19a36ff2`; dependency pins/licenses/actual archive SHA256 from `active_work/vec-01-opendnp3-lab-deployment-readiness.md` and `docs/verification/VEC-01-LAB-DEPLOY-001.md`. Verify current retrieved bytes, not merely copy manifests.
- Historical peer hashes: `outstation-demo` `8ae8c84125c4920352b6daef4145e2a91de9e02cd6e2c1b94bf6995ebeb093bb`; `libopendnp3.a` `e2926474288fb8d2c6c7962192bad2ff0f724f63c07d2209fe5b30c26fbe0677`. If unavailable, controlled re-acquisition on non-OT staging followed by offline/no-network rebuild from verified pinned source/deps is allowed within isolated lab only. Document toolchain, actual new hashes and explain differences; unexpected mismatches of supposedly identical inputs are BLOCKED until understood.
- Move/copy verified artifacts and required licenses/manifests/logs into a **durable access-controlled lab artifact store outside Git and outside disposable `/tmp`**. Verify with fresh SHA-256 after transfer, record destination and retention/access plan. Do not commit C++ binaries, tarballs, PCAP, secrets or unverified external source.
- Inspect the exact peer release's static G1V1/G30V1/G20V1 and G12V1 CROB capabilities. The upstream demo defaults to `0.0.0.0:20000` and link addresses 10/1; **do not run with those defaults**. Before activation, derive and review lab-bound interface/port/link-address 4/3, simulated BI[0]=true/BI[1]=false, AI[0]=2300 (or recorded known distinctive signed value + quality), CT[0]=42 and dedicated simulated BO[0]. Unsupported exact variation, qualifier or control status must be recorded, not silently made compliant by modifying `tamzrod/dnp3`.

## Gate B — independent safety sign-off BEFORE first TCP/DNP3 packet
- Establish a **dedicated offline VM/host or internal-only network namespace**, one logical master ↔ one OpenDNP3 outstation, non-routed illustrative subnet `10.77.0.0/24`, dedicated capture interface; no route to internet or any OT/SCADA/field network. No host networking, privileged network bridge, forwarding/NAT, published listener ports, field hardware or production credentials.
- Record actual host/interface IPs, routes, listening-socket configuration, firewall/security controls, link master address 3 and outstation address 4, TCP port 20000 confined to the isolated segment, binary SHA, simulated point values and mapping, and capture placement. Explicitly verify the OpenDNP3 listener binds **only** its isolated lab address (not `0.0.0.0` if reachable from outside). Preflight review should be static or namespace-local before first packet; do not probe production routes/addresses.
- **A named independent reviewer**, distinct from executor, must inspect and approve the isolation checklist, physical-output absence, peer configuration and single simulated CROB target **before** any listener activation or wire exchange. Record identity, decision, date, reviewed config hashes and exact scope. **No independent approval → TASK_BLOCKED and STOP.** There is no autonomous self-approval.
- Define bounded wall-clock deadlines for each step and entire campaign, watchdog/kill mechanism, and stop conditions before opening a socket. Disable any blind control auto-retry; use a single bounded Direct-Operate attempt whose outcome is never guessed. On isolation loss, unknown routing, unexpected peer, unintended physical effect, hang or ambiguous control outcome: stop the run immediately, preserve evidence, disable listener and report BLOCKED/ABORTED.

## Gate C — execute positive interoperability matrix only after Gate A+B PASS
Record each run ID, config hashes, master commit SHA, OpenDNP3 binary SHA, precise expected vs actual values, timing/deadline, raw frame references and result. Capture begins **before** TCP connect and ends after close.
1. TCP connect to exactly one isolated peer (no TLS).
2. DNP3 link reset and status: correct source/destination addresses and expected response/control.
3. G60V1 Class-0 all-objects integrity poll (record qualifier); independently decode G1V1 binary, G30V1 analog, G20V1 counter data and quality against fixed simulated ground truth. A response alone does not pass; actual values and indices must match.
4. Multiple object headers in one APDU: verify no missing/duplicate objects **if peer emits multi-header**; if it doesn't, record NOT_OBSERVED and require an explicit evidence limitation or separate independent test—do not falsely PASS.
5. One G12V1 Direct-Operate CROB to the independently reviewed BO[0] memory-only simulated endpoint; verify exact qualifier/control encoding, response status/IIN, and simulated state/result. Never treat TCP ACK as protocol command success. If outcome ambiguous, mark AMBIGUOUS and **never blind retry**.
6. Clean disconnect/close without hang/resource leak. A failed prerequisite stops subsequent positive steps.

## Gate D — isolated deterministic negative tests
Execute individually with fresh capture/run IDs and reproducible bounded deadlines; preserve all failures. Negative test mechanisms cannot reach real OT addresses. No mutation of supported library code or undocumented retry policy to make tests pass.
- Bad link payload/header CRC: controlled corrupted lab-only frame via deterministic fault emitter/replay; reject with bounded failure, no false success or hang.
- Wrong link source/destination address: bounded classified rejection/timeout, no hang.
- TCP drop mid-exchange: bounded classified disconnect/timeout, no deadlock.
- G12V1 command rejected by peer: explicit status/IIN failure; never report ControlSuccess. Keep simulator only; **no automatic/manual retry if prior command outcome is unknown**.
- Incomplete capture: intentionally or deterministically truncate separate capture and classify INCOMPLETE/REJECT AS EVIDENCE; never convert to an accepted positive.
- A test failure, hang, unexpected retry or safety violation is a real finding; do not silently rerun to erase failure. Clearly separate incomplete evidence from proven negative behavior.

## Gate E — raw evidence, faithful .vec and independent reviewer
- Preserve each original untouched `.pcap` or `.pcapng` as **source of truth** in durable external artifact store; include capture start/end, operator, master SHA, Go version, independent peer version/tag/build SHA, topology, listener binding, config hash, capture tool/decoder versions, clock offset and uncertainty, monotonic correlation with application logs, test ID and verdict. Compute SHA256 on retained original bytes, derived artifacts and logs.
- Reassemble TCP stream correctly (including retransmission handling) and extract observed *original* link frames; correlate each with direction, timestamps and PCAP packet/span offsets. Validate DNP3 `05 64` framing, link length, link header CRC and every 16-octet payload-block CRC (including trailing partial), ≤250-byte link payload, little-endian fields, DNP3 addresses, application sequence/confirmation and object qualifiers/status. Store observed point values and raw response.
- Generate reviewable `.vec` **from real capture bytes only**, conforming to `active_work/testdata/external/FORMAT.md`: matching file stem `id`, title, source, capture, role, master-addr, outstation-addr, optional pcap, and ordered `@ M->O link` / `@ O->M link` records; app-layer records only if correctly reassembled and labeled. Do not use `sample-vec01-placeholder.vec` as external evidence. No synthetic/fabricated/hand-edited wire bytes.
- Preserve separate redacted `.vec` sidecar with full list of redactions; **do not alter original DNP3 protocol bytes in a way that masquerades as actual capture**. If any protocol byte would need modification, label a derivative instead; never offer it as original proof.
- Store raw PCAP, derived `.vec`, optional sidecar, capture and master logs, SHA256 manifest and review decision **outside Git**. Do not commit binaries, packet traces, external source or potentially sensitive network capture in this package. Provide durable absolute artifact paths and access-control/retention notes in tracked verification *without secrets*.
- **Independent reviewer (not executor)** examines originals plus derived `.vec` and assigns one verdict with reasons for each run and campaign: `ACCEPT` only for independently eligible peer and complete valid positive + bounded negative outcomes + hashes/provenance + no false control success; `REJECT` for technical failure/ineligible peer/unsafe condition; `INCOMPLETE` for insufficient/absent captures, missing provenance or interrupted campaign, while keeping all known technical failures explicitly recorded. Missing independent review → INCOMPLETE/TASK_BLOCKED, **not** ACCEPT.
- Evidence review may support only a *scoped* statement that the tested v0 path interoperated with OpenDNP3 3.1.2 under specified conditions. **Do not close R4, amend supported-profile.md or claim IEEE 1815 full conformance/certification/production readiness within this CURRENT.**

## Exact tracked Git edit allowlist for CURRENT execution
1. **NEW** `docs/verification/VEC-01-LAB-EXEC-001.md` — actual executed steps, deadlines, per-test results (PASS/FAIL/NOT_OBSERVED/AMBIGUOUS/ABORTED/INCOMPLETE), full external evidence manifest SHA256/locations, isolation and peer-output independent preflight signoff, reviewer decision and limitations.
2. **NEW** `active_work/vec-01-opendnp3-lab-execution-evidence.md` — concise indexed campaign matrix, evidence references and honest scoped/non-claims; do not embed PCAP/binary.
3. **NEW** `archive/active-work/VEC-01-LAB-EXEC-001.md` — original human-promoted CURRENT task preserved verbatim beneath clearly separated archive preamble.
4. `workflow/active_work/README.md` and root `handoff.md` — only matching authority marker lifecycle and concise execution closeout.
5. `workflow/active_work/VEC-01-LAB-EXEC-001.md` — remove **only after** complete acceptance and verbatim archival.
**Promotion-only paths**: new CURRENT task and two authority marker files. External lab artifacts, including real raw PCAP and generated .vec, stay **outside Git**. No edits to Go code/test/fixtures, `active_work/testdata/external/*`, supported-profile, existing deployment/readiness/historical verification, task archives from earlier packages, `go.mod`/`go.sum`, CI/scripts, dependency files, `README.md`, KDE/governance or roadmap. If code fixes, fixture ingestion, additional tracked files or variation support become necessary, record blocker and ask for a separate human-promoted corrective task. The workflow must not commit live captures by implication.

## Mandatory executable Go regression gates
Actually run and record exact command, checkout SHA, Go version, exit code and relevant outputs on the same checkout. No previous PASS substitutes, no NOT_RUN counted as successful completion:
```sh
bash scripts/cwal-check.sh  # clean startup, rerun at clean post-commit checkpoint
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
If system Go missing, privately provision checksum-verified official Go 1.22.12 or compatible pinned toolchain. Record every PASS/FAIL with actual measured evidence. Note the internal external-MVP Tiers 1–3 are not independent third-party proof.

## Completion, negative decisions and stop conditions
- PASS readiness only with verified artifacts, durable retention, **independent preflight approval**, safe isolated listener configuration, exact independent peer, and simulated points.
- COMPLETE campaign only when all required positive and negative cases executed with independently reviewed complete raw PCAP/derived .vec, auditable SHA manifest and **independent campaign verdict ACCEPT**, plus every mandatory Go gate PASS. If peer cannot emit multi-header or exact variations, record NOT_OBSERVED/UNSUPPORTED with explicit scope; no silent success. If these materially prevent accepted v0 coverage, keep CURRENT blocked and require corrective authorization.
- On test incompatibility, crash, timeout, ambiguous control, missing artifact/reviewer or safety breach, **fail closed**: preserve evidence, stop safely, keep CURRENT active with truthful TASK_BLOCKED/ABORTED evidence; never erase failure or skip to QUEUED. Reviewer verdict REJECT/INCOMPLETE is recorded, not rounded up to acceptance.
- On full acceptance: record independent decision and actual hashes/results, archive original task verbatim, close both six-field markers to `CURRENT: NONE`, `QUEUED: NONE`, `Status: WORK_PACKAGE_READY`, commit/push **only** permanent `work`, read back exact remote SHA and files, report `TASK_CHECKPOINTED`. **R4 status decision remains separate** even after campaign acceptance.
- No autonomous PR, merge, field deployment, production/SCADA/OT device connection, code scope expansion, unsafe retry, unsupported feature implementation, certification or full conformance claim.
