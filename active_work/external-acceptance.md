# External Master Acceptance Checklist

**Series:** MEXT  
**Gate task:** MEXT-035  
**Do not claim external Master readiness until every item is checked on one commit.**

## Commands

```bash
./scripts/verify-mvp.sh
./scripts/verify-external-mvp.sh   # required from MEXT-033 onward
```

## Checklist

- [ ] `verify-mvp.sh` exit 0 (no internal regression)
- [ ] `verify-external-mvp.sh` exit 0 (real TCP master↔outstation path)
- [x] CROB control codes match IEEE 1815 bitfield golden vectors (MEXT-011) ✅
- [x] Operate does not ControlTimeout on valid success APDU from in-repo outstation over TCP (MEXT-012/013) — parse-side fix done in MEXT-012; proven on real TCP in MEXT-013 (`TestOperateRealTCPSuccess`)
- [ ] Class-0 multi-object-header response returns all G1/G20/G30 points (MEXT-014/015) — parse-side fixed in MEXT-014; IntegrityPoll now uses a single Class-0 multi-group read as primary path (one exchange returns the full set) with a per-group fallback (MEXT-015); real-TCP proof pending MEXT-033
- [ ] README external claims match tests (MEXT-034/035)

## Record (fill at MEXT-035)

| Field | Value |
|-------|-------|
| Date | |
| Commit | |
| verify-mvp.sh | |
| verify-external-mvp.sh | |
| Notes | |
