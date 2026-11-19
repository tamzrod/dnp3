# AO-008 — KDE Design Decision Request (UNAPPROVED)

**Status:** PROPOSAL ONLY — NO KDE SIGN-OFF / NO PROFILE PROMOTION.
**Current contract:** `workflow/active_work/AO-008.md` (owner 2026-11-19)
**Accepted preparation:** `archive/active-work/AO-007.md`
**Source proposal:** `docs/verification/AO-007-PROPOSED-PROFILE.md`
**Published profile remains:** G41 UNSUPPORTED in `active_work/supported-profile.md` blob `3e592e77ad6e31d5492c609c5fdd62382505904b`.

## Required KDE/human architecture decision

Approve or revise the precise following contract **before** any production Go API behavior or support declaration changes:

1. **Capability:** G41V1 int32, V2 int16, V3 float32, V4 float64, one indexed control per request, evidence-backed qualifier 0x28/count 1, Direct Operate **with response**, explicit response and status verification. No SBO/no-response/WRITE/multi-point.
2. **Activation:** Master and outstation separately **explicit opt-in**; default v0 operation and the default-outstation host handler path remain fail-closed. No silent default G41 admission, auto-negotiation or privilege escalation.
3. **Stable public config name:** proposal `master.WithAnalogControl()` and `outstation.WithAnalogControl()` (names not implemented). Decide aliases and migration for existing `WithExperimentalAnalogControl()` on both roles. Existing experimental explicit lab behavior must not be silently dropped/broken.
4. **Typed-only controls:** preserve `NewAnalogInt32Control`, `NewAnalogInt16Control`, `NewAnalogFloat32Control`, `NewAnalogFloat64Control`. Exclude legacy ambiguous `NewAnalogControl(float64)` even with opt-in; verify via regression. Reject mismatched typed value/variation.
5. **Float policy:** strict exact-bit command echo remains; OpenDNP3 3.1.2 signed-zero/subnormal normalization is **FAIL/DIVERGENCE** and Go must fail closed. Document this as excluded successful-interoperability edge case. Do not add tolerance without separate design evidence.
6. **Safety:** exact echo/status validation, invalid syntax/CRC/index/count/qualifier rejects, NaN/Inf fail closed, nil handler no success, after-send ambiguous outcome quarantine, no blind retry. No generic production readiness claims.
7. **Versioning and compatibility:** recommended **scoped versioned opt-in profile** that does not rewrite historical v0 Reject. KDE decides exact profile version, semantic version implications for public API, upgrade/rollback, consumer communication and whether an experimental compatibility alias stays permanent or deprecated.
8. **Proof envelope:** OpenDNP3 3.1.2 is the single independent peer/version; positive A/B 8/8 each and negative evidence recorded in AO-006/AO-007. Reverse-direction fault injection **NOT_RUN**; fresh AO-007 PCAP/log/image artifacts not committed. No general IEEE 1815 certification or field/NGCP integration claim.

## Proposed future controlled implementation and verification

After **KDE sign-off AND a separate owner go-ahead for Phase B**, permitted changes should be narrowly reviewed in `pkg/dnp3/master/client.go`, `pkg/dnp3/outstation/server.go`, `pkg/dnp3/types/commands.go`, focused typed/legacy and default reject tests, public docs, and lab fixtures only as separately authorized. The authoritative `active_work/supported-profile.md` stays untouched until a further separate owner promotion authorization.

**Minimum test matrix:** master default rejection without wire send; outstation default no handler call; asymmetric single-role opt-in; both opted-in four typed variations success/out-of-range; experimental alias backwards compatibility; legacy constructor reject even when opted-in; G41V0/V5 and unsupported modes; malformed/CRC/index/count/qualifier; NaN/Inf, wrong echo/unknown status; disconnect-after-send/unknown/quarantine/no blind retry; signed-zero/subnormal mismatch fail-closed; G12 CROB and Class-0 regressions. Full Go vet/build/unit/race + isolated pinned OpenDNP3 3.1.2 rerun with original PCAP/SHA provenance on approved changes.

## Requested explicit KDE decisions

- **Approve / modify / reject** stable master+outstation opt-in naming and alias semantics.
- **Approve / modify / reject** scoped profile-version approach, old v0 compatibility and rollback.
- **Approve / modify / reject** float exact-bit exclusion and peer-limited support wording.
- Confirm whether phase-B implementation may proceed, with a separate owner authorization. KDE approval **alone does not** authorize profile publication.

**Until recorded KDE decision:** `DESIGN_PENDING_KDE_APPROVAL`; CWAL may checkpoint preparation but not implement admitted/public supported behavior, modify published profile or auto-promote subsequent work.


## Phase A reconciliation against actual source — 2026-11-19 (read-only verification)

Verified on `work` `8779556f455a0f66125dde0078b97167aa5bac85` (`main` `82414d6`). Every claim below was checked against live source in this invocation; nothing was changed.

### Exact source touchpoints (current line numbers)

| Surface | File:line | Verified current behavior |
|---|---|---|
| Master opt-in field | `pkg/dnp3/master/client.go:256` | `ExperimentalAnalogControl bool`, default `false` |
| Master opt-in option | `pkg/dnp3/master/client.go:410` | `WithExperimentalAnalogControl()` sets the field; documented non-default/experimental |
| Master object gate | `pkg/dnp3/master/client.go:2132–2153` | CROB G12V1 admitted unconditionally; G41V1–V4 admitted **only** when `ExperimentalAnalogControl`; all else `ErrUnsupportedOption` before wire |
| Outstation opt-in field | `pkg/dnp3/outstation/server.go:218` | `ExperimentalAnalogControl bool`, default `false` |
| Outstation opt-in option | `pkg/dnp3/outstation/server.go:320` | `WithExperimentalAnalogControl()` |
| Outstation gate mirror | `pkg/dnp3/outstation/server.go:464`, `1113`, `1189` | `analogControlEnabled` mirrors config into the data handler |
| Outstation fail-closed | `pkg/dnp3/outstation/server.go:681–692` | Without opt-in: refuses before host handler, no state mutation, never SUCCESS |
| Host handler call | `pkg/dnp3/outstation/server.go:793` | `HandleAnalogCommand` invoked only after the opt-in gate |
| Legacy constructor | `pkg/dnp3/types/commands.go:235` | `NewAnalogControl(uint16,float64,CommandType)` → Group 41 / **Variation 1** with `*AnalogCommandValue` (float64 carrier) |
| Typed constructors | `pkg/dnp3/types/commands.go:252,269,289,309` | V1 int32, V2 int16, V3 float32, V4 float64 |

### Verified finding — legacy exclusion is *implicit*, not explicit

`NewAnalogControl(float64)` produces Group 41 / Variation 1 carrying `*AnalogCommandValue` (float64). Under opt-in the master gate admits G41V1 by group/variation, then `controlInt32Value` (`internal/master/master.go:4082`) rejects the float64 carrier (only `int32`/`int`/`int64` accepted), so `encodeG41Value` returns `nil` and the control is rejected **before any wire traffic**. Therefore the legacy constructor already fails closed under opt-in — but the exclusion depends on the **value-type exactness rule**, not on an explicit "legacy constructor" type check.

**Risk to record for KDE:** this makes the exclusion easy to regress. If any future change adds float→int32 coercion for V1, the legacy ambiguous constructor would silently begin transmitting a truncated value. Phase B should therefore add an explicit regression test asserting `NewAnalogControl` (float64 carrier) is rejected **even with opt-in on both roles**, and consider an explicit carrier-type check rather than relying only on coercion strictness. This is a design observation, not a currently reproducible defect: current behavior is correct and fail-closed.

### Symmetry check

Master and outstation opt-ins are independent today (separate `Config` fields, no negotiation). A master-only opt-in sends G41 wire that a default outstation refuses at `server.go:681` without invoking the host handler — so no control executes. A outstation-only opt-in never receives G41 because the default master refuses at `client.go:2145`. Both asymmetric cases are already safe; Phase B must add explicit tests pinning them.

### Existing regression baselines that a promotion must not delete

- `pkg/dnp3/master/g41v1_public_test.go:92` `TestPublicOperateG41RejectedByDefault` — default master rejects with **zero wire transmissions**.
- `test/integration/g41_default_reject_test.go:39` `TestOperateG41DefaultOutstationRejectsEndToEnd` — default outstation never invokes the host analog handler even when the master is opted in.
- Typed positive/non-success tests: `g41v1_public_test.go:136,153`; `g41v2/v3/v4_public_test.go`; real-TCP `test/integration/g41v2/v3/v4_real_tcp_test.go`; outstation `analog_status_failclosed_test.go`.

### KDE approval status

Searched `.kde/` (governance, verification, engines, knowledge, laboratory) for any G41 / analog-control / opt-in approval artifact: **none exists**. The only approval record is `.kde/governance/REC-001-APPROVAL.md`, which concerns bootstrap engine selection and is unrelated. Therefore the honest state is **`KDE_APPROVAL_PENDING`**; no sign-off is manufactured and `.kde/` was not modified.

### Phase A gate results (this invocation)

| Gate | Status | Basis |
|---|---|---|
| Six authority markers identical | **PASS** | `AO-008 / work / CURRENT AO-008 / QUEUED NONE / TASK_IN_PROGRESS` in both handoffs |
| `bash scripts/cwal-check.sh` | **PASS** | `CWAL CHECK OK: CURRENT=AO-008 branch=work` |
| `git diff --check` | **PASS** | exit 0, clean tree at checkpoint start |
| Published profile blob | **UNCHANGED** | `3e592e77ad6e31d5492c609c5fdd62382505904b`, G41 UNSUPPORTED |
| Source reconciliation | **PASS (read-only)** | all touchpoints above verified |
| Go build/vet/unit/race/external MVP | **NOT_RUN (by design)** | Phase A changes no code; these become mandatory in Phase B when behavior changes |
| Isolated OpenDNP3 3.1.2 rerun | **NOT_RUN (by design)** | no behavior change in Phase A; AO-006/AO-007 evidence remains historical |
| KDE architecture approval | **PENDING** | no artifact found; not manufactured |

**Disposition: `DESIGN_PENDING_KDE_APPROVAL`.** Phase A preparation is complete and reconciled against live source. Phase B implementation is **not** authorized and requires KDE sign-off **and** a separate owner implementation authorization. Phase C published-profile promotion requires a further separate owner authorization. G41 remains **UNSUPPORTED**.
