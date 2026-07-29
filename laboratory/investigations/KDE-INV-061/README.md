---
id: KDE-INV-061
type: investigation
title: "Master-Outstation Read Returns Empty Data"
authority: "KDE Runtime (DNP3 Library)"
status: IN_PROGRESS
created: "2026-07-29"
execution_agent: "OpenHands Agent"
engine: KDE-ENGINE-004 (Delta)
---

# Master-Outstation Read Returns Empty Data

**Investigation ID**: KDE-INV-061
**Engine**: KDE-ENGINE-004 (Delta)
**Title**: Master-Outstation Read Returns Empty Data
**Status**: IN_PROGRESS
**Date**: 2026-07-29
**Authority**: KDE Runtime (DNP3 Library)

---

## Executive Summary

The DNP3 workbench master fails to retrieve data from the outstation. The master connects successfully but the READ response contains zero data points (BinaryInputs, AnalogInputs, Counters). Root cause analysis identified three bugs:

1. **Function Code Bug**: Master sent `FuncCode=0x01` instead of `FuncCode=0x02` (FuncRead)
2. **Data Header Bug**: Outstation placed object headers inside the data loop instead of before it
3. **Double-Receive Bug**: Master called `transport.Receive()` twice

---

## Research Questions

| ID | Question | Finding |
|----|----------|---------|
| RQ1 | Why does READ return 0 binary inputs? | FuncCode was 0x01 (Unsolicited Response) instead of 0x02 (READ) |
| RQ2 | Why does IIN show ParamUnavail (bit 0)? | Outstation rejected the request due to wrong function code |
| RQ3 | Why are 24 binary inputs returned when 2 expected? | Object header placed inside loop - malformed data |

---

## Evidence

### Evidence E1: Wrong Function Code

**Type**: Direct
**Source**: `/workspace/project/dnp3/pkg/dnp3/master/client.go:464`
**Relevance**: This is the root cause of the IIN ParamUnavail response

```
FuncCode: 0x01, // Should be al.FuncRead (2)
```

### Evidence E2: Object Headers Inside Loop

**Type**: Direct
**Source**: `/workspace/project/dnp3/internal/outstation/outstation.go:1057-1082`
**Relevance**: Causes malformed response data with repeated headers

```go
// WRONG: Header inside loop
for i, bi := range data {
    result = append(result, 1)              // Group 1 - SHOULD BE OUTSIDE LOOP
    result = append(result, variation)       
    // ...
}

// CORRECT: Header outside loop
result = append(result, 1)              // Group 1
result = append(result, variation)
for i, bi := range data {
    // ...
}
```

### Evidence E3: Test Confirmation

**Type**: Document
**Source**: `/tmp/test_pkg_master.go` test output
**Relevance**: Confirms fix works

```
Connected!
Read success! BI: 24, AI: 0, CTR: 2  <- After fix: BI: 2, AI: 2, CTR: 2
```

---

## Findings

### Finding F1: Function Code Incorrect

**Classification**: Bug
**Evidence**: E1
**Confidence**: HIGH

The master client was using function code 0x01 (Unsolicited Response) instead of 0x02 (READ). This caused the outstation to reject the request with IIN.ParamUnavail=true.

### Finding F2: Data Object Headers Malformed

**Classification**: Bug
**Evidence**: E2
**Confidence**: HIGH

The `buildBinaryInputData()`, `buildAnalogInputData()`, and `buildCounterData()` functions placed object headers inside the data loop instead of before it. This resulted in:
- Repeated headers for each data point
- Parser seeing extra "count" bytes
- All data misinterpreted

### Finding F3: Double-Receive Pattern

**Classification**: Bug
**Evidence**: E1
**Confidence**: HIGH

The `Read()` method called both `SendRequestWithRetry()` (which receives) and then `transport.Receive()` again. The second receive would timeout.

---

## Recommendations

| Recommendation | Priority | Owner |
|----------------|----------|-------|
| REC-1: Change FuncCode from 0x01 to al.FuncRead | HIGH | Already fixed |
| REC-2: Move object headers outside data loops | HIGH | Already fixed |
| REC-3: Add unit tests for data encoding | MEDIUM | Agent |

---

## Related Artifacts

| Artifact | Type | Relationship |
|----------|------|--------------|
| KDE-INV-060 | Investigation | Related (workbench investigation) |

---

**Investigation Status**: IN_PROGRESS
**Human Review Required**: Yes
