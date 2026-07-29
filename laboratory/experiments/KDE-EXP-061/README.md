# Experiment KDE-EXP-061

**Experiment ID**: KDE-EXP-061
**Investigation**: KDE-INV-061
**Title**: Master-Outstation Data Parsing Bug
**Engine**: KDE-ENGINE-004 (Delta)
**Status**: IN_PROGRESS
**Date**: 2026-07-29T07:15:00Z

---

## Executive Summary

**Hypothesis**: The master's data parsing functions have incorrect data size calculations for DNP3 object variations.

**Evidence from prior runs**:
- Outstation sends correct data (verified by debug output)
- Master receives IIN=[0,0] - no errors
- Master parses 9 binary inputs instead of 2 expected
- Master parses 0 analog inputs instead of 2 expected
- Master parses 0 counters instead of 2 expected

**Root Cause Hypothesis**: The `parseBinaryInputs`, `parseAnalogInputs`, and `parseCounters` functions in `pkg/dnp3/master/client.go` have incorrect data size calculations.

---

## Experimental Design

### Setup

| Component | Version |
|-----------|---------|
| Go | 1.21+ |
| DNP3 Library | Current |
| Test | Integration test with master-outstation |

### Test Data

**Outstation DefaultDataHandler provides**:
- Binary Inputs: 2 items (index 0,1)
- Analog Inputs: 2 items (index 0,1)  
- Counters: 2 items (index 0,1)

**Outstation builds data for Group 30, Variation 1**:
- Expected: 4-byte float + 1-byte quality per point
- Expected total per point: 6 bytes (2 index + 4 value + 1 quality, minus 1 for overlap)

---

## Expected vs Actual

| Metric | Expected | Actual |
|--------|----------|--------|
| Binary Inputs | 2 | 9 |
| Analog Inputs | 2 | 0 |
| Counters | 2 | 0 |

### Data Flow Analysis

```
Raw Response Data (46 bytes):
010100020000010001011e010002000042c900000100014348400001140100020000000003e8010001000007d001

Parsing at offset 10 (Group 30):
Offset 10-13: 1e 01 00 02  (Group=30, Var=1, Qual=0, Count=2)
Offset 14-19: 00 00 42 c9 00 00  (First AI: index 0, float 100.5)
Offset 20-25: 01 00 01 43 48 40  (Second AI: index 256, float ???)

ISSUE: After parsing 2 AI points (12 bytes), offset = 26
But next header at offset 26 appears to be "01 00" - this is parsed as
group=1, var=0, count=67 (which is garbage)
```

**Hypothesis**: The master's parser for Group 30 Variation 1 expects 7 bytes per point, but the data is only 6 bytes per point.

---

## Run 1

**Date**: 2026-07-29T07:15:00Z
**Status**: IN_PROGRESS

### Evidence

[E1] Outstation Debug Output:
```
DEBUG buildBinaryInputData: result len=10, hex=01010002000001000101
DEBUG buildAnalogInputData: result len=18, hex=1e010002000042c900000100014348400001
DEBUG buildCounterData: result len=18, hex=140100020000000003e8010001000007d001
```

[E2] Master Debug Output:
```
DEBUG BI: offset=10, group=30, var=1, qual=0, count=2
```

[E3] Parsing Calculation:
```
For Variation 1 (32-bit float with flags):
- Index: 2 bytes
- Value: 4 bytes
- Quality: 1 byte
- Total per point: 7 bytes

For 2 points: 4 header + 14 data = 18 bytes
Actual: 18 bytes ✓ (matches)
```

### Analysis

The outstation data looks correct. The master's parser needs verification.

---

## Findings

### Finding 1: Data Alignment

**Classification**: NEEDS_VERIFICATION
**Evidence**: E1, E2, E3

The outstation data encoding appears correct. Need to verify master parsing logic.

### Finding 2: Public DefaultDataHandler Returns 2 Items

**Classification**: CONFIRMED
**Evidence**: Prior debug shows `internalDataHandler.GetBinaryInputs: count=2`

The public `DefaultDataHandler` returns only 2 items, not 5 as the internal one does.

---

## Next Steps

1. Verify master parser data size calculations
2. Check if variation numbering differs between outstation and master
3. Run controlled test with known input data

---

**Experiment Status**: IN_PROGRESS
**Engine**: KDE-ENGINE-004 (Delta)
**Agent**: OpenHands
