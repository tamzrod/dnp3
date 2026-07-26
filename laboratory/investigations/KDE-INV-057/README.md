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

# Investigation: Watchdog for Bootstrap Integrity and Rogue AI Detection

**Investigation ID**: KDE-INV-057  
**Title**: Watchdog: Bootstrap Integrity and Rogue AI Detection  
**Status**: IN_PROGRESS  
**Engine**: KDE-ENGINE-002 (Beta)  
**Date**: 2026-07-26

---

## 1. Executive Summary

### 1.1 Overview

This investigation explores the design and implementation of a watchdog system for the KDE Runtime that:
1. Monitors bootstrap directory integrity
2. Detects rogue or runaway AI behavior
3. Triggers appropriate responses

### 1.2 Key Findings

| Finding | Status | Evidence |
|---------|--------|----------|
| Bootstrap integrity can be monitored via checksums | ✅ Implemented | `.kde/bootstrap/status.py` |
| Watch mode provides continuous monitoring | ✅ Implemented | `make kde-watch` |
| File change detection is feasible | ✅ Implemented | `BootstrapWatchdog` class |
| Rogue AI patterns can be categorized | 📋 Documented | Section 4.3 |

---

## 2. Problem Context

### 2.1 The Challenge

In AI agent sandboxed environments (like OpenHands):

1. **No Direct Human Oversight**: Agents operate autonomously
2. **Bootstrap Integrity Risk**: A malfunctioning agent could corrupt `.kde/`
3. **Runaway Behavior Risk**: Infinite loops or memory spirals could consume resources
4. **Silent Failures**: Issues could go unnoticed for extended periods

### 2.2 Use Cases

| Scenario | Without Watchdog | With Watchdog |
|----------|------------------|---------------|
| Agent corrupts `state.json` | Silent failure | Detected immediately |
| Infinite loop in agent | Resource exhaustion | Detected and logged |
| Unauthorized file deletion | Lost governance | Alert triggered |
| Memory leak | System crash | Early warning |

---

## 3. Watchdog Architecture

### 3.1 Component Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        WATCHDOG SYSTEM                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Bootstrap Integrity Monitor                                     │
│  ├── File checksum computation                                  │
│  ├── Change detection                                           │
│  └── Integrity reporting                                        │
│                                                                 │
│  Process Behavior Analyzer                                       │
│  ├── Action pattern tracking                                    │
│  ├── Infinite loop detection                                    │
│  └── Unauthorized action detection                             │
│                                                                 │
│  Resource Monitor                                               │
│  ├── CPU/Memory tracking                                        │
│  ├── Disk I/O monitoring                                        │
│  └── Threshold alerting                                         │
│                                                                 │
│  Alert Engine                                                   │
│  ├── Log only (info)                                           │
│  ├── Notify (warning)                                          │
│  └── Trigger (critical - shutdown)                              │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 3.2 Bootstrap Integrity Monitor

**Implementation**: `.kde/bootstrap/status.py`

**Features**:
```python
class BootstrapStatusChecker:
    def get_status(self) -> BootstrapStatus:
        """Get complete bootstrap status"""
        
    def verify_module(self, module_name: str) -> ModuleStatus:
        """Verify a single module"""
        
    def compute_checksum(self, path: Path) -> str:
        """Compute SHA256 checksum"""
```

**Usage**:
```bash
# Check status
python3 .kde/bootstrap/status.py

# Watch mode (continuous)
python3 .kde/bootstrap/status.py --watch

# JSON output for automation
python3 .kde/bootstrap/status.py --json
```

### 3.3 BootstrapWatchdog Class

**Implementation**: `.kde/bootstrap/status.py`

**Features**:
```python
class BootstrapWatchdog:
    def compute_baseline(self):
        """Compute baseline checksums for integrity monitoring"""
        
    def check_integrity(self) -> Dict:
        """Check for file changes"""
        
    def watch(self, duration: float = None):
        """Watch mode - continuously monitor"""
```

---

## 4. Rogue AI Detection

### 4.1 Detection Patterns

| Pattern | Description | Detection Method | Severity |
|---------|-------------|-----------------|----------|
| **Infinite Loop** | Same action repeated | Track action history, detect repetition | CRITICAL |
| **Memory Spiral** | Unbounded memory growth | Monitor RSS, detect monotonic increase | HIGH |
| **File Spam** | Rapid file creation/deletion | Count files per time window | HIGH |
| **Unauthorized Action** | Action outside scope | Maintain allowed action list | CRITICAL |
| **Stall** | No progress for extended time | Track last action timestamp | MEDIUM |
| **Resource Exhaustion** | CPU at 100% for long time | Monitor CPU usage | HIGH |

### 4.2 Pattern Detection Implementation

```python
class ProcessBehaviorAnalyzer:
    def detect_infinite_loop(self, action_history: List[str]) -> bool:
        """Detect if same action repeated N times"""
        
    def detect_memory_spiral(self, memory_samples: List[int]) -> bool:
        """Detect if memory grows without bound"""
        
    def detect_file_spam(self, file_operations: List[FileOp]) -> bool:
        """Detect rapid file creation/deletion"""
        
    def detect_unauthorized_action(self, action: str) -> bool:
        """Check if action is in allowed list"""
```

### 4.3 Response Actions

| Severity | Response | Action |
|----------|----------|--------|
| INFO | Log only | Record event, continue |
| WARNING | Notify | Alert, log, continue |
| HIGH | Alert + Log | Full context, continue with warning |
| CRITICAL | Trigger | Log, notify, optional shutdown |

---

## 5. Implementation Status

### 5.1 Phase 1: Bootstrap Integrity ✅ DONE

**Status**: Implemented in `.kde/bootstrap/status.py`

**Features**:
- [x] Module existence verification
- [x] File checksum computation
- [x] Integrity status reporting
- [x] Watch mode for continuous monitoring
- [x] JSON output for automation

**Evidence**:
```bash
$ python3 .kde/bootstrap/status.py
======================================================================
KDE BOOTSTRAP STATUS
======================================================================
Timestamp: 2026-07-26T22:06:00.000000
Project:   DNP3 Library
State:     ready
Integrity: ✅ OK

--- Modules ---
  [✅] engines
  [✅] experts
  [✅] knowledge
  ...
```

### 5.2 Phase 2: Process Monitor ⏳ FUTURE

**Status**: Design documented, implementation deferred

**Features to implement**:
- [ ] Action history tracking
- [ ] Pattern matching for rogue behaviors
- [ ] Alert generation

### 5.3 Phase 3: Resource Monitor ⏳ FUTURE

**Status**: Design documented, implementation deferred

**Features to implement**:
- [ ] CPU/Memory sampling
- [ ] Threshold configuration
- [ ] Alert triggers

---

## 6. Integration with KDE Runtime

### 6.1 Current Makefile Integration

```makefile
## kde-status - Show bootstrap status (no environment checks)
kde-status:
	@python3 .kde/bootstrap/status.py

## kde-watch - Watch bootstrap status continuously
kde-watch:
	@python3 .kde/bootstrap/status.py --watch --interval 5
```

### 6.2 Future Integration

```makefile
## kde-watchdog - Run watchdog mode
kde-watchdog:
	@python3 .kde/bootstrap/watchdog.py --mode bootstrap,process,resource
```

---

## 7. Recommendations

### REC-1: Implement Process Monitor ✅ START

**Priority**: HIGH  
**Action**: Implement `ProcessBehaviorAnalyzer` class

**Implementation**:
```python
class ProcessBehaviorAnalyzer:
    def __init__(self):
        self.action_history = []
        self.memory_samples = []
        
    def track_action(self, action: str):
        self.action_history.append(action)
        if len(self.action_history) > 1000:
            self.action_history.pop(0)
            
    def detect_anomalies(self) -> List[Anomaly]:
        anomalies = []
        if self.detect_infinite_loop():
            anomalies.append(Anomaly(
                type='INFINITE_LOOP',
                severity='CRITICAL'
            ))
        # ... more detections
        return anomalies
```

### REC-2: Add Resource Monitoring ✅ START

**Priority**: MEDIUM  
**Action**: Add `ResourceMonitor` class

### REC-3: Create Watchdog Dashboard ✅ START

**Priority**: LOW  
**Action**: Create web-based status dashboard

---

## 8. Conclusion

### 8.1 Summary

The watchdog concept is feasible and has been partially implemented:

1. **Bootstrap Integrity**: ✅ Implemented via `BootstrapStatusChecker`
2. **Watch Mode**: ✅ Implemented via `BootstrapWatchdog`
3. **Rogue AI Patterns**: 📋 Documented, implementation deferred

### 8.2 Next Steps

| Step | Action | Priority |
|------|--------|----------|
| 1 | Implement `ProcessBehaviorAnalyzer` | HIGH |
| 2 | Implement `ResourceMonitor` | MEDIUM |
| 3 | Add alerting integration | MEDIUM |
| 4 | Create watchdog dashboard | LOW |

---

## 9. Related Artifacts

| Artifact | Type | Relationship |
|----------|------|--------------|
| KDE-INV-056 | Investigation | Startup optimization |
| `.kde/bootstrap/status.py` | Implementation | Bootstrap integrity |
| `.kde/bootstrap/gates.py` | Implementation | Preflight checks |
| Makefile | Integration | Command wiring |

---

*Investigation Status*: IN_PROGRESS  
*Next Step*: Implement ProcessBehaviorAnalyzer  
*Human Review Required*: Yes
