# External Master Acceptance Checklist

**Series:** MEXT  
**Gate task:** MEXT-035  
**MEXT-035 in-repo v0 gate recorded complete.** This is not an independent third-party interoperability or IEEE 1815 conformance certification. R4/VEC-01 remains open.

## Commands

```bash
./scripts/verify-mvp.sh
./scripts/verify-external-mvp.sh   # required from MEXT-033 onward
```

## Checklist

- [x] `verify-mvp.sh` exit 0 (no internal regression) — recorded in MEXT-035 handoff
- [x] `verify-external-mvp.sh` exit 0 (in-repo real-TCP Master↔Outstation; Tiers 1–3) — recorded in MEXT-035 handoff; VEC-01 advisory absent
- [x] CROB control codes match IEEE 1815 bitfield golden vectors (MEXT-011) ✅
- [x] Operate does not ControlTimeout on valid success APDU from in-repo outstation over TCP (MEXT-012/013) — parse-side fix done in MEXT-012; proven on real TCP in MEXT-013 (`TestOperateRealTCPSuccess`)
- [x] Class-0 multi-object-header response returns all G1/G20/G30 points (MEXT-014/015) — primary single-exchange read, per-group fallback; covered by MEXT-033 Tier 3
- [x] README external claims match the tested in-repo v0 gate and explicitly disclaim third-party interoperability/full IEEE 1815 (MEXT-034/035)

## MEXT-035 acceptance record (synchronized 2026-11-08)

| Field | Value |
|-------|-------|
| Historical gate | MEXT-035 completed, as documented in `active_work/handoff.md` and `active_work/supported-profile.md` |
| Exact MEXT-035 verification SHA | Not established by this synchronization; consult original gate history before claiming a pinned SHA |
| `verify-mvp.sh` | PASS recorded by historical MEXT-035 handoff |
| `verify-external-mvp.sh` | PASS recorded by historical MEXT-035 handoff (in-repo Tiers 1–3) |
| Additional checkpoint | PR #15 merged 2026-11-08 at `b38d7691aeb1adccf52fe665e368d4d77766977c`; merge CI completed successfully |
| Limitations | No third-party VEC-01/PCAP proof; not full IEEE 1815 and not production SCADA certification |

Do not treat the later PR #15 CI result as a substitute for the historical
MEXT-035 gate record. The acceptance above is a reconciliation of existing
recorded evidence, **not** a claim that those MEXT checks were re-run now.
