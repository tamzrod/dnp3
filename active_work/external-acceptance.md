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
- [ ] CROB control codes match IEEE 1815 bitfield golden vectors (MEXT-011)
- [ ] Operate does not ControlTimeout on valid success APDU from in-repo outstation over TCP (MEXT-012/013)
- [ ] Class-0 multi-object-header response returns all G1/G20/G30 points (MEXT-014/015)
- [ ] README external claims match tests (MEXT-034/035)

## Record (fill at MEXT-035)

| Field | Value |
|-------|-------|
| Date | |
| Commit | |
| verify-mvp.sh | |
| verify-external-mvp.sh | |
| Notes | |
