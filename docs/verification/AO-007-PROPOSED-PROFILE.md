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
