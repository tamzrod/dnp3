---
id: KDE-INV-057
type: investigation
title: "Watchdog: Bootstrap Integrity and Rogue AI Detection"
status: IN_PROGRESS
authority: "KDE Runtime (DNP3 Library)"
created: "2026-07-26T22:06:00Z"
execution_agent: "OpenHands Agent"
engine: KDE-ENGINE-002 (Beta)
---

# Investigation Specification: Watchdog for Bootstrap Integrity and Rogue AI Detection

**Investigation ID**: KDE-INV-057  
**Title**: Watchdog: Bootstrap Integrity and Rogue AI Detection  
**Status**: IN_PROGRESS  
**Engine**: KDE-ENGINE-002 (Beta)  
**Date**: 2026-07-26

---

## 1. Problem Statement

### 1.1 Context

In sandboxed AI agent environments (like OpenHands), the KDE Runtime operates without direct human oversight. This raises two critical concerns:

1. **Bootstrap Integrity**: Can we detect if the `.kde/` directory has been modified by a rogue or malfunctioning agent?
2. **Rogue AI Detection**: Can we detect runaway AI behavior (infinite loops, memory leaks, unauthorized actions)?

### 1.2 Questions

1. How can we monitor bootstrap directory integrity?
2. How can we detect rogue or runaway AI behavior?
3. What actions should be taken when issues are detected?

---

## 2. Investigation Plan

### 2.1 Objectives

1. Define watchdog scope and responsibilities
2. Design bootstrap integrity monitoring
3. Design rogue AI detection mechanisms
4. Specify intervention strategies
5. Implement proof-of-concept

### 2.2 Success Criteria

| Criterion | Target |
|-----------|--------|
| Bootstrap integrity monitored | File changes detected within 5s |
| Rogue AI patterns defined | At least 5 patterns cataloged |
| Detection mechanisms designed | Concrete implementation plan |
| Intervention strategy defined | Response actions specified |

---

## 3. Scope

### 3.1 In Scope

- Bootstrap directory monitoring (`.kde/`)
- Process behavior analysis
- Resource usage monitoring
- File integrity verification
- Alert mechanisms

### 3.2 Out of Scope

- Network traffic monitoring
- Full system intrusion detection
- Human interaction during execution
- Recovery mechanisms (future work)

---

## 4. Watchdog Design

### 4.1 Component Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        WATCHDOG SYSTEM                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐        │
│  │  Bootstrap  │    │   Process  │    │  Resource  │        │
│  │  Integrity  │    │  Behavior  │    │   Monitor  │        │
│  │   Monitor   │    │   Analyzer │    │            │        │
│  └──────┬──────┘    └──────┬──────┘    └──────┬──────┘        │
│         │                   │                   │               │
│         └───────────────────┼───────────────────┘               │
│                             │                                   │
│                    ┌────────▼────────┐                         │
│                    │   Alert Engine  │                         │
│                    └────────┬────────┘                         │
│                             │                                   │
│         ┌───────────────────┼───────────────────┐              │
│         │                   │                   │              │
│  ┌──────▼──────┐    ┌──────▼──────┐    ┌──────▼──────┐       │
│  │    Log      │    │   Notify    │    │   Trigger   │       │
│  │   Only     │    │   (Warning) │    │ (Shutdown)  │       │
│  └─────────────┘    └─────────────┘    └─────────────┘       │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 4.2 Bootstrap Integrity Monitor

**Responsibilities**:
- Watch `.kde/` directory for file changes
- Compute and store checksums of critical files
- Detect unauthorized modifications
- Maintain audit trail

**Monitored Files**:
| File/Directory | Reason |
|----------------|--------|
| `.kde/runtime/state.json` | Runtime state integrity |
| `.kde/bootstrap/config.yaml` | Configuration integrity |
| `.kde/engines/` | Engine code integrity |
| `.kde/governance/` | Policy integrity |
| `.kde/seeds/` | Seed methodology integrity |

### 4.3 Process Behavior Analyzer

**Rogue AI Patterns**:

| Pattern | Detection | Severity |
|---------|-----------|----------|
| Infinite Loop | Repeated same action, no progress | CRITICAL |
| Memory Spiral | Increasing memory without bound | HIGH |
| File Spam | Rapid file creation/deletion | HIGH |
| Unauthorized Action | Action outside allowed scope | CRITICAL |
| Stall | No output for extended period | MEDIUM |

### 4.4 Resource Monitor

**Metrics Tracked**:
| Metric | Threshold | Action |
|--------|-----------|--------|
| CPU Usage | >95% for 60s | Warning |
| Memory Usage | >80% for 30s | Warning |
| Disk I/O | >1GB/s sustained | Warning |
| File Count | >1000 new files/min | Alert |

---

## 5. Implementation Plan

### 5.1 Phase 1: Bootstrap Integrity (DONE)

Status: ✅ Implemented in `.kde/bootstrap/status.py`

Features:
- Module existence verification
- File checksum computation
- Integrity status reporting
- Watch mode for continuous monitoring

### 5.2 Phase 2: Process Monitor (FUTURE)

Status: TODO

Features:
- Subprocess tracking
- Action pattern analysis
- Infinite loop detection

### 5.3 Phase 3: Resource Monitor (FUTURE)

Status: TODO

Features:
- CPU/Memory tracking
- Disk usage monitoring
- Alert thresholds

---

## 6. Evidence Requirements

### 6.1 Required Evidence

- Bootstrap status checker implementation
- Watch mode demonstration
- Pattern catalog documentation

---

*Specification created: 2026-07-26*
