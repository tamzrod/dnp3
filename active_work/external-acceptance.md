# External Master Acceptance Checklist

**Series:** MEXT  
**Gate task:** MEXT-035  
**In-repo v0 external-MVP gate completed at MEXT-035.** This is not proof of independent third-party interoperability, full IEEE 1815 conformance, or production readiness. R4/VEC-01 remains OPEN.

## Commands

```bash
./scripts/verify-mvp.sh
./scripts/verify-external-mvp.sh   # required from MEXT-033 onward
```

## Checklist

- [x] `verify-mvp.sh` exit 0 (no internal regression) — MEXT-035 handoff; re-executed PASS under DNP3-058
- [x] `verify-external-mvp.sh` exit 0 (in-repo real TCP master↔outstation path, Tiers 1–3) — MEXT-035 and DNP3-058 PASS; VEC-01 advisory remains missing
- [x] CROB control codes match IEEE 1815 bitfield golden vectors (MEXT-011) ✅
- [x] Operate does not ControlTimeout on valid success APDU from in-repo outstation over TCP (MEXT-012/013) — parse-side fix done in MEXT-012; proven on real TCP in MEXT-013 (`TestOperateRealTCPSuccess`)
- [x] Class-0 multi-object-header response returns all G1/G20/G30 points (MEXT-014/015; MEXT-033 Tier 3 in-repo gate)
- [x] README external claims match the tested in-repo v0 profile, excluding independent third-party interoperability and full IEEE 1815 conformance (MEXT-034/035)

## Evidence reconciliation — 2026-11-08

| Field | Verified historical record |
|-------|----------------------------|
| MEXT-035 gate | COMPLETE in `active_work/handoff.md` (original gate SHA not independently established in this reconciliation) |
| `verify-mvp.sh` | PASS at MEXT-035; re-executed PASS in DNP3-058 on `2afbeb5acf0fce9126edb122290b29ba3f532151` |
| `verify-external-mvp.sh` | PASS Tiers 1–3 at MEXT-035 and DNP3-058, **not** third-party VEC-01 proof |
| Additional Go checks | DNP3-058 executed Go 1.22.12 build, vet, unit, race, CI-equivalent coverage gates with PASS evidence in `docs/verification/DNP3-058.md` |
| MEXT-040 | Documentation-only SBO deferral completed; feature **NOT SUPPORTED** |
| MEXT-041 | Documentation-only event/unsolicited deferral completed; feature **NOT SUPPORTED** |
| PR #17 | Merged into `main` at `95ac10289ead46c09264a9da2bb877f0e0507410`; latest main CI success |
| R4 / VEC-01 | **OPEN** — no independent PCAP/third-party-stack interoperability proof |

This synchronization is a documentation reconciliation, **not** an assertion that the old MEXT-035 checks were newly rerun on this documentation commit. See historical handoff and DNP3-058 evidence for precise execution contexts.
