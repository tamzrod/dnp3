---
id: DNP3-INV-003
type: investigation
title: "DNP3 Workbench: Outstation & Master Integration"
status: IN_PROGRESS
authority: "DNP3 Library"
created: "2026-07-27T23:43:00Z"
execution_agent: "OpenHands Agent"
engine: KDE-ENGINE-004 (Delta)
---

# Investigation Specification: DNP3 Workbench - Outstation & Master Integration

**Investigation ID**: DNP3-INV-003
**Title**: DNP3 Workbench: Outstation & Master Integration
**Status**: IN_PROGRESS
**Engine**: KDE-ENGINE-004 (Delta)
**Date**: 2026-07-27

---

## 1. Background

### 1.1 Problem Statement

The DNP3 library needs a **working engineering workbench** - a tool to run both an outstation and master that can connect to each other and demonstrate interoperability. The main goal is:

> "The main goal of this repository is to have a working DNP3 library"

### 1.2 Current State Analysis

| Component | Status | Notes |
|-----------|--------|-------|
| Workbench App | Partial | Fyne UI skeleton exists |
| Master Session | Stub | Returns mock data, doesn't use real DNP3 |
| Outstation Session | Partial | TCP listener exists, no real DNP3 protocol |
| DNP3 Library | Unknown | Need to verify full protocol stack |

### 1.3 Requirements

| # | Requirement | Priority |
|---|-------------|----------|
| 1 | Two Windows executables (outstation + master) | Critical |
| 2 | Outstation provides random moving data | Critical |
| 3 | Master can read data from outstation | Critical |
| 4 | Master can write data to outstation | Critical |
| 5 | Any DNP3 master can connect to our outstation | Critical |
| 6 | Any DNP3 outstation can connect to our master | Critical |
| 7 | Cross-vendor interoperability | Goal |

---

## 2. Investigation Questions

### 2.1 Primary Questions

1. **Is the DNP3 protocol stack complete and functional?**
   - Can we send/receive valid DNP3 frames?
   - Is the transport layer (TCP) working?

2. **What is missing from the workbench implementation?**
   - Real DNP3 master client integration
   - Real DNP3 outstation server integration
   - Data point management
   - Random data simulation

3. **Can we build two standalone executables?**
   - workbench-master.exe
   - workbench-outstation.exe
   - Or one app with mode selection

4. **What DNP3 conformance is required?**
   - Level 1 vs Level 2 vs Level 3
   - Which function codes
   - Which object groups

### 2.2 Secondary Questions

1. How to handle data point configuration?
2. What are the default ports (20000/20001)?
3. Should we use TLS or just TCP?
4. What IIN (Internal Indications) should we support?

---

## 3. Evidence to Collect

### 3.1 Code Evidence

- [ ] DNP3 library implementation completeness
- [ ] Existing master/outstation session code
- [ ] Workbench build status
- [ ] Transport layer implementation

### 3.2 External Evidence

- [ ] DNP3 specification compliance
- [ ] Reference implementations for comparison
- [ ] Common outstation/master implementations

---

## 4. Success Criteria

| Criterion | Target | Verification |
|-----------|--------|--------------|
| Outstation runs | Windows executable | Manual test |
| Master runs | Windows executable | Manual test |
| Data exchange | Master reads random data | Automated test |
| Write support | Master writes to outstation | Manual test |
| Interoperability | Works with 3rd party tools | Optional goal |

---

## 5. Investigation Plan

### Phase 1: Gap Analysis (Current Session)
1. Analyze current DNP3 library implementation
2. Identify missing protocol components
3. Document what works vs what needs work

### Phase 2: Architecture Design
1. Design two-executable vs single-app architecture
2. Define data point model
3. Define random data generation strategy

### Phase 3: Implementation
1. Implement/verify DNP3 master client
2. Implement/verify DNP3 outstation server
3. Implement data simulation
4. Add workbench UI controls

### Phase 4: Testing & Validation
1. Build executables
2. Test master-to-outstation communication
3. Test interoperability (if possible)

---

## 6. Architecture Options

### Option A: Single Executable with Mode Selection
```
workbench.exe --mode master  # Act as DNP3 master
workbench.exe --mode outstation  # Act as DNP3 outstation
```

**Pros**: Single binary, easier distribution
**Cons**: More complex UI, larger executable

### Option B: Two Separate Executables
```
workbench-master.exe      # Master only
workbench-outstation.exe  # Outstation only
```

**Pros**: Simpler UI per app, smaller binaries
**Cons**: Two binaries to manage

### Option C: Single App with Runtime Mode Switch
```
workbench.exe  # GUI with mode selection at runtime
```

**Pros**: User-friendly, no command-line
**Cons**: More complex state management

---

## 7. Stakeholders

- **Author**: OpenHands Agent
- **Owner**: DNP3 Library maintainers
- **Users**: DNP3 developers, integrators, testers

---

## 8. Constraints

- Must produce Windows executables
- Must use existing DNP3 library (go-dnp3)
- Fyne UI framework is already chosen
- Target: Cross-vendor DNP3 compliance

---

## 9. Related Documents

| Document | Relationship |
|----------|--------------|
| DNP3-ENG-WORKBENCH-001.md | Original engineering plan |
| DNP3-INV-002 | Fyne API issues (resolved) |
| cmd/workbench/* | Implementation location |

---

## 10. Next Steps

1. Run gap analysis on DNP3 library
2. Select architecture (Option A/B/C)
3. Create experiments for each component
4. Implement missing pieces
