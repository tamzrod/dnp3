# AO-008 — KDE Architecture Design Approval Record (G41 Explicit Opt-In)

**Record ID:** AO-008-G41-DESIGN-APPROVAL
**Date:** 2026-11-19
**Repository:** tamzrod/dnp3
**Authority:** Repository owner acting as human KDE Approval Authority (owner authorization in current conversation, 2026-11-19)
**Decision:** **APPROVED — DESIGN CONTRACT ONLY**, contingent on independent governance CHECK and authorized integration of this record
**Execution authority:** **NONE** (this record does not authorize Phase B, Phase C, PR, merge, release, or deployment)
**Governance isolation:** `governance/ao-008-g41-optin-approval`, based on `work` `2867779df55f2c1dc9008ed2b27bfec3827f4d6b`. Only this approval record may be changed on this branch.
**CWAL task:** `workflow/active_work/AO-008.md` (CURRENT remains AO-008/TASK_IN_PROGRESS on `work`)

## Review provenance and authority

Human owner explicitly authorized the formal KDE design approval process and approval of the opt-in direction **only after** verifying the stable master/outstation API, legacy compatibility, typed controls, strict float exclusions, versioning, and rollback. The bounded decision below is the resulting proposed approval record; it is **not a substitute for independent CHECK or authorized integration** into `work`.

Reviewed:
- Approved governance hierarchy `.kde/governance/GOVERNANCE-HIERARCHY.md`: human approval required for governance-affecting architecture; agent execution not self-approval.
- `operation cwal.md` §GOV-RECON-001: parallel governance maintenance needs separate authorization, isolated branch, strict allowlist, independent CHECK, and separate integration approval.
- Accepted `archive/active-work/AO-007.md`, `docs/verification/AO-007-PROPOSED-PROFILE.md`, `docs/verification/AO-008-DESIGN-DECISION-DRAFT.md` with Phase A source mapping and `DESIGN_PENDING_KDE_APPROVAL` at `2867779`.
- GitHub source blobs on `work`: `pkg/dnp3/master/client.go` `eb782335e721dd0ac0a0aee84039255986e747ac`; `pkg/dnp3/outstation/server.go` `51af68aa4f0a4bfe6493055222fa661dd0e4f20c`; `pkg/dnp3/types/commands.go` `6a626ee9f8d734dda1954175863cb0fb840eb127`.
- Published v0 supported-profile blob `3e592e77ad6e31d5492c609c5fdd62382505904b`; **Group 41 remains UNSUPPORTED**.

## Exact decisions approved for subsequent *separate* implementation authorization

| Dimension | Approved narrow design |
|---|---|
| Control function | DNP3 Group 41 **Direct Operate with response** only |
| Variations | V1 signed int32; V2 signed int16; V3 IEEE-754 float32; V4 IEEE-754 float64 |
| Wire/cardinality | Exactly **one indexed point**, qualified count 1, verified qualifier **0x28**, explicit command response/status; no inferred multi-point support |
| Master stable opt-in | `master.WithAnalogControl()` returning existing `ConfigOption`; OFF by default |
| Outstation stable opt-in | `outstation.WithAnalogControl()` returning existing `ConfigOption`; OFF by default |
| Existing API compatibility | Keep `WithExperimentalAnalogControl()` on BOTH roles as backward-compatible opt-in alias; preserve its existing callers and default-off semantics; do not remove or silently deprecate for this task |
| Master/outstation symmetry | Each role must explicitly opt in independently; master-only cannot execute via default outstation; outstation-only cannot bypass default master no-wire reject |
| Typed-only values | Admit only `NewAnalogInt32Control`, `NewAnalogInt16Control`, `NewAnalogFloat32Control`, `NewAnalogFloat64Control` with correct value/variation; no implicit float narrowing/coercion |
| Legacy exclusion | `types.NewAnalogControl(index, float64, cmdType)` remains **REJECTED even when opted in**; require explicit negative test and prefer an explicit carrier-type validation before any wire send |
| Failure safety | Preserve exact-bit typed echoed value/group/variation/index/status checks, fail-closed rejects of malformed/CRC/count/qualifier/NaN/Inf/nil handler, and unknown-after-send quarantine without blind retries |
| Float compatibility | For OpenDNP3 3.1.2 G41V3/V4 `-0.0`/smallest subnormal → `+0.0` normalization, classify **FAIL/DIVERGENCE** and keep Go **fail closed**; no permissive normalization policy |
| Version approach | **Proposed scoped supported interoperability profile v1 (explicit-opt-in G41 only)**, while preserving historical v0 as unchanged/rejecting G41; `v1` is a support-contract profile designation, NOT a Go module major version change and NOT a release authorization |
| Rollback | Revert new stable alias/admission and profile-v1 draft if implementation fails; preserve old v0 default reject, G12 CROB, existing experimental opt-in functionality and archived wire evidence |
| Interoperability claim | Limited to isolated **OpenDNP3 3.1.2** single independent stack/version and tested 8/8 positive each direction/negative cases; uncommitted AO-007 fresh captures NOT independently reproducible from Git; reverse-direction injected-fault negative NOT_RUN |

## Explicit exclusions and stop conditions

**Unsupported** until separately approved: SBO, no-response, Group 41 WRITE, batched/multi-point commands, ambiguous legacy float64 constructor, other G41 variations/qualifiers, NGCP utility point indices, scaling, units, real I/O, field/SCADA/production commands, unrestricted IEEE 1815 conformance or certification claims.

**Phase B MUST NOT START solely because this file exists.** Required before implementation:
1. Independent read-only CHECK of this *isolated governance branch* and the decision against `work` at the verified baseline.
2. Separate **owner authorization to integrate** this governance decision into the execution branch through the approved non-force, non-PR governance process; no auto-merge.
3. Separate **owner authorization for AO-008 Phase B implementation** (explicit new edit-path allowlist and test matrix on `work`).
4. Preserve published G41 **UNSUPPORTED** until an additional, separate Phase C profile-promotion approval.

On future Phase B, run source-specific RED→GREEN tests for default-rejected V1–V4, asymmetric and dual opt-in, legacy explicit reject, strict response/negative cases, all Go build/vet/unit/race/external-MVP gates, and isolated OpenDNP3 3.1.2 wire/negative/float-characterization with verifiable capture provenance. Documentation-only approval does not claim any of those tests were freshly run.

## Decision disposition

**DESIGN_APPROVED_FOR_REVIEW_ONLY / GOVERNANCE_INTEGRATION_PENDING.**

Owner-approved direction has been bounded into an exact API and version contract, **but governance integration is not yet authorized**; the effective CWAL state on `work` is still `CURRENT: AO-008 / QUEUED: NONE / TASK_IN_PROGRESS / KDE_APPROVAL_PENDING` until independent CHECK and human authorization for integration. No Go source, profile, `work` marker, PR, merge, release, or deployment changes are included here.
