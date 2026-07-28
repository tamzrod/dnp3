---
id: DNP3-INV-062
type: investigation
title: "Workbench Master/Outstation Instance Separation"
status: approved_with_addendum
authority: "KDE Runtime (DNP3 Library)"
created: "2026-07-27T23:45:00Z"
addendum: "File menu with window controls (Minimize, Maximize/Restore, Close)"
---

# DNP3-INV-062: Workbench Master/Outstation Instance Separation

**Investigation ID**: DNP3-INV-062
**Title**: Workbench Master/Outstation Instance Separation
**Authority**: KDE Runtime (DNP3 Library)
**Status**: IN PROGRESS
**Date**: 2026-07-27
**Execution Agent**: OpenHands Agent

---

## 1. Executive Summary

### 1.1 Investigation Request

User request: "dis approve. the master and outstation must be separated and not in one screen as you cannot do the same thing at the same time. if you want to do it run a seperate program. you miss undertood. the executable is the same. what i mean is you run a two instance of it."

### 1.2 Core Finding

**The user requires complete separation of Master and Outstation into dedicated instances, not tabs, panes, or mode switching within a single window.**

### 1.3 Key Requirements

| Requirement | Description |
|-------------|-------------|
| Single executable | Same binary for both modes |
| Dedicated instances | Each instance is either Master OR Outstation |
| No shared UI | No tabs, panes, or mode switching |
| Parallel operation | Two instances can run simultaneously |
| Startup mode selection | Command-line flag or dialog |

---

## 2. Current State Analysis

### 2.1 Current Architecture

```
cmd/workbench/
├── main.go                    # Single entry point
├── internal/
│   ├── controller/
│   │   └── controller.go      # Unified controller (Master only)
│   └── session/
│       ├── session.go        # Master session
│       └── outstation.go     # Outstation session
```

### 2.2 Current Session Implementation

| Session Type | File | Implementation | Status |
|-------------|------|----------------|--------|
| Master | `session/session.go` | `MasterSession` | ✅ Functional |
| Outstation | `session/outstation.go` | `OutstationSession` | ✅ Functional |

### 2.3 Current UI Architecture

```
┌────────────────────────────────────────────────────┐
│  Mode Panel (radio buttons)                       │
│  (•) Poll Outstation  ( ) Simulate Outstation     │
├────────────────────┬───────────────────────────────┤
│  Connection Panel  │  Data Table Panel             │
│  Commands Panel    │  Control Panel               │
└────────────────────┴───────────────────────────────┘
```

---

## 3. User Requirements Analysis

### 3.1 Explicit Requirements

| # | Requirement | Evidence |
|---|-------------|----------|
| 1 | Master and Outstation separated | "must be separated and not in one screen" |
| 2 | Cannot do both at same time | "you cannot do the same thing at the same time" |
| 3 | Run separate program | "run a separate program" |
| 4 | Same executable | "the executable is the same" |
| 5 | Two instances | "run a two instance of it" |

### 3.2 Derived Requirements

| # | Requirement | Justification |
|---|-------------|---------------|
| 6 | Mode selection at startup | "run a two instance" implies choosing mode |
| 7 | Dedicated windows | Each instance needs its own window |
| 8 | No mode switching | "not in one screen" |
| 9 | Independent configuration | Each mode has different config needs |

### 3.3 Anti-Requirements (What User Does NOT Want)

| # | Anti-Requirement | User Statement |
|---|------------------|---------------|
| A | No tabs | "not in one screen" |
| B | No dual panes | "must be separated" |
| C | No mode switching | "not in one screen as you cannot do the same thing" |
| D | No single-window mode toggle | "run a separate program" |

---

## 4. Proposed Architecture

### 4.1 Instance Separation Model

```
┌─────────────────────────────────────────────────────┐
│              SINGLE EXECUTABLE                      │
│         (workbench.exe / workbench)                │
└─────────────────────────────────────────────────────┘
        │                           │
        ▼                           ▼
┌───────────────┐           ┌───────────────┐
│  --mode master│           │--mode outstation│
└───────┬───────┘           └───────┬───────┘
        │                           │
        ▼                           ▼
┌───────────────┐           ┌───────────────┐
│ MASTER WINDOW │           │OUTSTATION WINDOW│
│               │           │                │
│ Connection    │           │ Server Config  │
│ Commands      │           │ Data Points    │
│ Data Table    │           │ Simulation     │
└───────────────┘           └───────────────┘
```

### 4.2 Command-Line Interface

```bash
# Run as Master
./workbench --mode master

# Run as Outstation
./workbench --mode outstation

# Show mode selection dialog (default)
./workbench

# Show help
./workbench --help
```

### 4.3 Mode Selection Dialog

When run without `--mode` flag, display a selection dialog:

```
┌──────────────────────────────────────────────┐
│         DNP3 Engineering Workbench           │
│                                              │
│   Choose Operating Mode:                    │
│                                              │
│   ┌────────────────────────────────────────┐ │
│   │  [Master Mode]                        │ │
│   │  Connect to remote outstations         │ │
│   │  Read/write data points               │ │
│   └────────────────────────────────────────┘ │
│                                              │
│   ┌────────────────────────────────────────┐ │
│   │  [Outstation Mode]                    │ │
│   │  Run simulated DNP3 server             │ │
│   │  Generate random data                 │ │
│   └────────────────────────────────────────┘ │
│                                              │
│              [Cancel]                        │
└──────────────────────────────────────────────┘
```

---

## 5. Investigation Artifacts

| Artifact | Description | Status |
|----------|-------------|--------|
| [SPEC.md](SPEC.md) | Investigation specification | TODO |
| [CONCLUSION.md](CONCLUSION.md) | Investigation conclusions | TODO |

---

## 6. Evidence Collection

### 6.1 Code Evidence

| Evidence | File | Finding |
|----------|------|---------|
| Current mode switching | `ui/window.go` | Radio button in sidebar |
| Controller implementation | `controller/controller.go` | Master-only implementation |
| Session implementations | `session/*.go` | Both exist but not decoupled |

### 6.2 User Feedback Evidence

| Statement | Interpretation |
|-----------|----------------|
| "must be separated and not in one screen" | No shared UI between modes |
| "you cannot do the same thing at the same time" | Mutual exclusion requirement |
| "run a separate program" | Dedicated execution context |
| "the executable is the same" | Single binary requirement |
| "run a two instance of it" | Parallel instance capability |

---

## 7. Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Breaking existing functionality | LOW | HIGH | Maintain backward compatibility |
| User confusion on startup | MEDIUM | LOW | Mode selection dialog |
| Configuration conflicts | LOW | MEDIUM | Separate config files |

---

## 8. Next Steps

1. [ ] Verify KDE Runtime is operational
2. [ ] Create SPEC.md with detailed implementation plan
3. [ ] Propose changes following laboratory rules
4. [ ] Create experiments for validation
5. [ ] Implement changes if approved

---

*Investigation initiated: 2026-07-27*
*Engineering Diagnosis: In Progress*
