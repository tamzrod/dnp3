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
