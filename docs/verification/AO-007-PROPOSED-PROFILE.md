# AO-007 — PROPOSED G41 Supported-Profile Amendment (UNAPPROVED DRAFT)

**Status: DRAFT / REVIEW ONLY / NOT PROMOTED.** This is not `active_work/supported-profile.md`, does not alter the published v0 acceptance contract, does not authorize G41 by default, and is not a release or production interoperability claim.

**Owner task:** `workflow/active_work/AO-007.md` (2026-11-19).  
**Baseline supported-profile Git blob:** `3e592e77ad6e31d5492c609c5fdd62382505904b` (G41 Reject).  
**Historical evidence:** `docs/verification/AO-006.md`, merged focused PR #31 at `82414d672e9ab12aba6170ec7f3f0578f6e71958`.

## Proposed *future* supported subset — subject to human/KDE approval

| Field | Candidate specification |
|---|---|
| Transport and role | TCP Master and Outstation, existing one-master/one-outstation bounded topology |
| Function | DNP3 Direct Operate **with response** only |
| Group/Variations | G41V1 int32; G41V2 int16; G41V3 float32; G41V4 float64 |
| Cardinality | Exactly one indexed point per request, approved count/index qualifier only |
| API | Typed `NewAnalogInt32Control`, `NewAnalogInt16Control`, `NewAnalogFloat32Control`, `NewAnalogFloat64Control`; `Operate`; analog command handler |
| Command outcome | Per-point SUCCESS/non-SUCCESS, strict response index/group/variation/value/status validation; no false SUCCESS |
| Failure semantics | Malformed/bad CRC rejection, invalid indices/qualifier/count, nil handler non-success; after-send ambiguous outcome quarantined, no blind retry |
| Explicitly excluded | SBO; no-response; multi-point; G41 WRITE; legacy `NewAnalogControl(float64)`; NGCP point map/scaling/units; field/SCADA/production claims |
| External evidence | Pinned isolated OpenDNP3 **3.1.2 only**, 8/8 each direction for V1–V4 idx0 success / idx1 OUT_OF_RANGE |
| Float caveat | On V3/V4 `-0.0` and smallest-subnormal echoes OpenDNP3 3.1.2 normalizes to `+0.0`; strict Go echo verification fails closed: **FAIL/DIVERGENCE** |
| Not independently verified | Second peer/version; reverse-direction fault proxy negative; generalized IEEE 1815 conformance |

## Decision required: how does supported interact with experimental opt-in?

**Unresolved architectural/API compatibility decision.** `WithExperimentalAnalogControl()` currently activates G41 explicitly while default v0 rejects it. A mere documentation change declaring G41 supported while default public APIs still reject it would be misleading. Before publication, human/KDE approval must select and verify one coherent contract, including:

- Either (A) retain an explicit opt-in and describe the supported capability as opt-in with a name/configuration policy appropriate for a nonexperimental supported profile, **or** (B) make only the approved subset default-admitted, with negative regression tests for all other commands.
- Both choices must explicitly preserve legacy `NewAnalogControl(float64)` rejection, strict typed wire encoding, malformed fail-closed behavior and the signed-zero/subnormal limitation. Do not implement either choice before approval.
- Define an observable migration and rollback path, including compatibility of published v0 clients and outstations; document whether profile version increments.
- Update `pkg/dnp3/master/client.go` and `pkg/dnp3/types/commands.go` public API comments only when the selected semantics and tests actually match.
- The outstation's analog command handler must have an equally clear default/admission contract. No accidental v0 enablement.

## Exact proposed profile areas to amend *after* separate authorization

1. `## Included Object Subset`: add tightly scoped G41V1–V4 Direct Operate with response, with restrictions and evidence links; retain existing G12V1 CROB.
2. `## Explicitly Unsupported`: retain all unrelated exclusions, explicitly retain SBO, no-response, unsupported Group 41 variants and legacy ambiguity. Narrow the generic rejection language only as approved.
3. `## Public API Disposition`: revise only the `Operate`, `SetCommandHandler`, analog handler, and *typed* analog constructor rows, with precise opt-in/default contract; keep legacy `NewAnalogControl` Reject separately. Avoid implying general G41 via other command modes.
4. `## Current Verification State`: add dated AO-005/AO-006 scoped peer, positive/negative PCAP, SHA-256 and CI evidence and limitations; retain historical VEC-01 and MEXT claims without retroactively rewriting their dates.
5. `README.md` and exported API descriptions: update *after* profile decision and tests; do not advertise NGCP interoperability, certified IEEE 1815 compliance or unrestricted production support.

## Proposed separate-authorization acceptance gate

- KDE design decision on supported-vs-explicit-opt-in behavior, profile version, compatibility and float exclusion signed off by human owner. **BLOCKER** until resolved.
- Approved code/API and supported-profile edits verified as coherent, with regression tests showing the chosen default and opt-in semantics (including legacy excluded constructor). No false success or unsupported-wire path.
- Go build/vet/full unit/integration/race + `scripts/verify-external-mvp.sh`, `scripts/cwal-check.sh`, `git diff --check`, `gofmt -l`, all freshly recorded in executable environment; tests NOT_RUN in this GitHub-only draft.
- Scope-specific independent OpenDNP3 3.1.2 V1–V4 interop and negatives replicated as required for *new behavior*; raw hashes and peer provenance retained. Second peer needed before any wider-than-OpenDNP3 3.1.2 claim.
- Independent pre-PR audit and separate authorization to modify the published profile, create PR, merge, release or deploy.

**Recommendation for AO-007 preparation:** define the configuration/public-default decision first. Until that is resolved, report `PROMOTION_PREP_BLOCKED_ON_APPROVAL`, not “G41 SUPPORTED.” The current published profile must remain blob `3e592e77ad6e31d5492c609c5fdd62382505904b`.

## AO-007 source/API disposition audit — 2026-11-19 (read-only inspection)

Source blobs inspected on `work`: `pkg/dnp3/master/client.go` (`eb782335e721dd0ac0a0aee84039255986e747ac`), `pkg/dnp3/outstation/server.go` (`51af68aa4f0a4bfe6493055222fa661dd0e4f20c`), and `pkg/dnp3/types/commands.go` (`6a626ee9f8d734dda1954175863cb0fb840eb127`). No source changes were authorized or made during this audit.

| Existing public surface/location | Verified current contract | Approved future migration touchpoint |
|---|---|---|
| Master `Config.ExperimentalAnalogControl` (`client.go` ~248–256), `WithExperimentalAnalogControl()` (~403–413) | Opt-in off by default; documented lab-only | Decide compatible supported opt-in naming vs default admission; do not silently make old switch mean something else |
| Master `Client.Operate` comment (~119–127), implementation gate (~2115–2153) | Reject G41V1–V4 absent opt-in, preserve CROB G12V1, reject SBO/no-response/others before send | Both implementation gate and advertised API contract must match the approved support policy; preserve negative cases |
| Outstation `Config.ExperimentalAnalogControl`, `WithExperimentalAnalogControl()` (~209–218, ~311–323) | Off by default; explicit lab switch | Approval must specify a symmetric outstation admission policy; master-only promotion would misrepresent end-to-end support |
| Outstation `CommandHandler.HandleAnalogCommand` (~425–443) and `WriteAnalogOutput` (~673–805) | Host analog handler not invoked without opt-in; typed V1–V4 dispatch; nil handler returns error; explicit non-success status retained | Update contract/docs only together with approved code and tests; preserve no-side-effect-on-reject |
| `types.NewAnalogControl(float64)` (`commands.go` ~231–243) | Legacy G41V1 label with float64 carrier, not the typed int32 wire contract; excluded/fail-closed | **Never auto-promote**; keep explicit Reject and add compatibility regression on any profile change |
| `NewAnalogInt32Control` (~245–260), `NewAnalogInt16Control` (~262–277), `NewAnalogFloat32Control` (~279–297), `NewAnalogFloat64Control` (~299–317) | Typed G41V1–V4; comments still say Reject until AO-004 peer check/AO-005 profile change even though AO-006 has now completed | Amend these four constructor descriptions only after a real promotion; do not present them as default-supported before gate decision |
| Outstation bridge comments (~676–681) | `DirectOperate` is described as “supported mode” in the experimental path | Distinguish implementation-supported command mode from *published-profile* supported capability |
| `active_work/supported-profile.md` public disposition table | `Master Operate`=G12 only, `SetCommandHandler`=G12 only, analog handler=Reject, `NewAnalogControl`=Reject, G41 excluded | Split typed constructor rows from legacy constructor; update only after human/KDE approval and behavior checks |
| `README.md` / exported package comment in `client.go` ~1–48 | Existing v0/G12-only external claim | Update with peer-scoped limits and upgrade notes only after profile admission is verified |

**Existing regression sources inspected:** `pkg/dnp3/master/g41v1_public_test.go` ~86–118 tests rejection with zero wire transmissions by default; `test/integration/g41_default_reject_test.go` ~39–96 checks default outstation does not invoke host analog handler, even against an opted-in master. These are **valuable compatibility baselines**: any default-admitted promotion requires an explicitly approved change of expected semantics, not simply deleting failing tests. Typed status, malformed, echo, after-send and race tests in AO-006 remain the promotion prerequisites.

**Wire acceptance detail:** The recorded AO-006 case files show a single indexed object using G41V1–V4 with qualifier `0x28`, count `1`, and indexes `0` (SUCCESS) / `1` (OUT_OF_RANGE); status `0x00` / `0x0c`. Do not generalize tested qualifier/cardinality or claim any other independent peer. No second stack or reverse-direction negative captured.

**Evidence versus tests in this continuation:** Remote GitHub files and source references were inspected; source blobs and marker consistency were read back. The local runtime has Go 1.23.2 and git but no repository checkout or Docker and cannot resolve `github.com`; `git ls-remote` failed with `Could not resolve host`. Therefore `cwal-check.sh`, `git diff --check` in checkout, Go build/vet/tests/race, and isolated OpenDNP3 are **NOT_RUN_THIS_INVOCATION**. No code or profile behavior changed, so no behavior-changing gate is falsely reported as PASS.

**Preparation finding:** No new reproducible protocol defect established. The reviewable promotion proposal is complete enough to bring to a KDE/owner design decision, but the decision and fresh executable gates have not passed. **`PROMOTION_PREP_BLOCKED_ON_APPROVAL_AND_EXECUTION`**, not a profile promotion or a closure recommendation.


## Owner-approved design direction — 2026-11-19: EXPLICIT OPT-IN (not KDE approval)

**Decision:** The owner selected the **explicit-opt-in** *design direction* for future official G41V1–V4 Direct Operate with response. This rejects the alternate “default-admitted without an opt-in” proposal. This is **not** approval to edit the supported profile or any production code, and does **not** substitute for KDE design/governance approval.

### Proposed executable contract (future change only)

1. **Default remains closed on both roles.** With only `master.NewConfig(...)` or `outstation.NewConfig(...)`, the published v0 TCP/Class-0/G12 path is unchanged; G41 Direct Operate is rejected before the master sends G41 wire traffic and before the outstation invokes `HandleAnalogCommand`. The opt-in must be explicit **on each endpoint**; enabling the master alone never enables an outstation. No implicit peer negotiation or auto-detection.
2. **Supported and opt-in are distinct attributes.** When the supported-profile amendment is approved, say “**supported only with explicit local G41 opt-in on each participating role, within this exact subset**”, not “G41 enabled by default.” Use a **new stable, nonexperimental configuration name** for the supported opt-in subject to KDE/API review (candidate `WithAnalogControl()` in the master and outstation package). Preserve existing `WithExperimentalAnalogControl()` as a compatible experimental alias if maintained; document which names are considered public/stable. **No new API exists yet.**
3. **Typed constructors only.** Accept `types.NewAnalogInt32Control` (V1), `NewAnalogInt16Control` (V2), `NewAnalogFloat32Control` (V3), and `NewAnalogFloat64Control` (V4), only with `types.DirectOperate`. The legacy ambiguous `NewAnalogControl(uint16,float64,...)` stays excluded/fail-closed, including if the new stable opt-in is set. Reject V0/V5/other groups, unsupported qualifiers and command modes.
4. **One indexed control per request.** Evidence-supported DNP3 G41 with count `1`, qualifier `0x28`, explicit 16-bit point index, and response. Validate response function/sequence, group, variation, count, index, typed value bits and status. Successful transport does not imply successful operation. Preserve specific non-success outcomes (e.g. OUT_OF_RANGE `0x0c`) and no false SUCCESS. Accepted external cases use index 0 SUCCESS and index 1 OUT_OF_RANGE in each variation.
5. **Fail-closed lifecycle.** Reject NaN/±Inf, malformed/truncated objects, CRC/address failures, unsupported variation, nil handler, unsafe retry and unknown-after-send. Master must quarantine controls after ambiguous outcome and refuse the next control until safely reconciled. Preserve status/error distinguishability; do not weaken exact-bit comparison or quarantine.
6. **Float caveat remains explicit.** OpenDNP3 **3.1.2** may normalize `-0.0` and smallest subnormal echoes to `+0.0`. The typed Go master *fails closed* on such mismatch; this is a recorded **FAIL/DIVERGENCE**, **not** a supported-success interop case. A normalization-tolerant variant is **not** authorized or implied by this design approval.
7. **Peer and scope limitation.** The sole independent stack/version is OpenDNP3 3.1.2 in isolated lab, both directions 8/8 positive, plus scoped negative captures. Reverse-direction fault injection is NOT_RUN. No general IEEE 1815 certification, hardware, NGCP, field/SCADA mapping, unrestricted output scaling or claimed utility acceptance.
8. **No silent profile upgrade.** Published `active_work/supported-profile.md` remains blob `3e592e77ad6e31d5492c609c5fdd62382505904b`, G41 **UNSUPPORTED**; no consumer may rely on this DRAFT as supported capability.

### Future migration/rollback sequence (requires separate permission)

- **Stage 0 — current**: v0 unchanged; existing experimental opt-in only. Archive AO-006 and positive/negative PCAP/SHA provenance unchanged.
- **Stage 1 — KDE design approval**: approve the opt-in name, exact master/outstation admission semantics, API compatibility and profile version. **Recommended:** a new explicitly versioned supported-profile amendment (v1 scoped opt-in), while leaving the legacy v0 reject statement historically intact. This is a proposal, not a version decision.
- **Stage 2 — implementation authorization**: introduce the stable opt-in on both roles (if approved), preserve experimental alias, and ensure the typed-only gate excludes the ambiguous legacy constructor. New changes must not allow G41 by default, must not inadvertently enable unsupported variations or commands, and must preserve CROB and unrelated event gates.
- **Stage 3 — tests**: default master no-G41-wire for V1–V4, default outstation handler untouched even when master opted in, master-only and outstation-only opt-in negatives, dual-opt-in each V1–V4 success/non-success, legacy constructor rejected *even with opt-in*, invalid/malformed/NaN/Inf and echo mismatch, timeout/disconnect/unknown/quarantine, signed-zero/subnormal fail-closed, existing CROB/Class-0/event/regression and race. Replicate isolated 3.1.2 interop when runtime behavior changes; retain raw PCAPs and hashes.
- **Stage 4 — publication authorization**: only after fresh tests, independent audit and owner/KDE approvals, update the *authoritative* supported profile plus README/public comments with scoped opt-in language and observed limitations; separately authorize any PR/merge/release.
- **Rollback**: retain legacy default G41 Reject and experimental switch throughout preparation; if any new opt-in/migration fails tests, revert only the new stable option/profile amendment (with owner-approved rollback) while leaving G12 and the existing experimental lab path unchanged. Never silently downgrade float checks.

### Gate disposition for this session

| Gate | Status | Basis |
|---|---|---|
| Owner opt-in design direction | **APPROVED** | Explicit 2026-11-19 user authorization, preparation only |
| KDE architecture/governance approval | **PENDING** | No KDE approval artifact or approved design decision found or manufactured |
| Current remote CWAL marker check | **PASS** | GitHub readback of `work` at start |
| AO-006 evidence and PR #31 | **PREVIOUSLY_RECORDED** | Accepted prior campaign; not rerun |
| Master/outstation source mapping | **PASS (read-only)** | Previous AO-007 GitHub source audit and proposal |
| GitHub-connected local checkout | **UNAVAILABLE** | Container cannot resolve github.com; GitHub connector is document API, not shell |
| `git diff --check`, `scripts/cwal-check.sh` | **NOT_RUN** | No repository checkout |
| Go build/vet/test/race/external-MVP | **NOT_RUN** | No repository checkout |
| Isolated OpenDNP3 3.1.2 | **NOT_RUN** | No Docker, no lab checkout |
| Published profile promotion | **NOT_DONE** | Explicitly prohibited |

**Result:** `DESIGN_DIRECTION_ACCEPTED__KDE_AND_EXECUTION_PENDING`. The opt-in decision resolves the *choice of direction*, but NOT the required KDE approval, migration implementation, fresh executable verification or authorization to promote. Keep AO-007 CURRENT and G41 UNSUPPORTED.


## Fresh executable verification — 2026-11-19 (supersedes the NOT_RUN gate table above)

A later CWAL invocation ran in an executable environment (Go 1.22.12 official checksum-verified toolchain; Docker daemon available) and executed the preparation gates freshly on `work` `8b9eb9eac01e2d733a863db508a03c3fa0ff90f4` (`main` `82414d6`). **No source, API, or protocol behavior was changed; this document remains an UNAPPROVED DRAFT.**

| Gate | Fresh status | Basis |
|---|---|---|
| Marker / remote baseline readback | **PASS** | `git ls-remote`: work `8b9eb9e`, main `82414d6`; six markers identical |
| `git diff --check`, `scripts/cwal-check.sh` | **PASS** | exit 0 |
| `gofmt -l`, `go vet ./...`, `go build ./...` | **PASS** | gofmt empty; vet/build exit 0 |
| `go test ./... -count=1`, `go test -race ./... -count=1` | **PASS** | exit 0 |
| `scripts/verify-external-mvp.sh` | **PASS** | all tiers OK |
| Source/API disposition audit | **PASS (read-only)** | cited blobs and line ranges re-verified on `work` HEAD |
| AO-006 evidence review | **PASS** | 15/15 SHA-256 records verify; PCAPs re-decode 8/8 both directions, 0 BAD CRCs |
| Isolated OpenDNP3 3.1.2 direction A / B | **PASS** | 8/8 each; exact bits; status `0x00`/`0x0c` |
| Negative matrices A / B | **PASS** | malformed → IIN `0x0100`, zero writes; after-send → unknown + quarantine + refused second control |
| Float `-0.0`/subnormal | **FAIL/DIVERGENCE (retained)** | peer bits `0x0`; Go master fails closed; `+1.5` control succeeds |
| KDE/human profile-amendment approval | **PENDING** | not manufactured; separate governance gate |
| Published profile promotion | **NOT_DONE** | explicitly prohibited |

Fresh artifact SHA-256 (this invocation, not committed — allowlist permits only this draft doc): dirA `a1d54cfe…`, dirB `32758272…`, negA `7356132a…`, negB `81b1041c…`, float-edge log `39a18298…`; lab image `sha256:bc13c24e91c006242a46cd7be6ccd664c0fdd19e1f42b8baf067f01e386f10f9`.

**Result: `PROMOTION_PREP_READY`** — the preparation, proposal, migration/rollback plan and fresh verification are complete. *Actual* promotion remains blocked pending separate KDE/owner approval of the profile amendment. Published `active_work/supported-profile.md` stays blob `3e592e77ad6e31d5492c609c5fdd62382505904b`; G41 remains **UNSUPPORTED**.
