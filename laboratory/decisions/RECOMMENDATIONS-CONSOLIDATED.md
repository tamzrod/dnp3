# KDE Runtime Recommendations - Consolidated

**Document ID**: REC-CONSOLIDATED-001  
**Date**: 2026-07-26  
**Source Investigations**: KDE-INV-056, KDE-INV-057  
**Status**: ACTIVE

---

## Executive Summary

This document consolidates all recommendations from recent KDE Runtime investigations related to startup optimization, bootstrap integrity, and watchdog capabilities.

### Quick Reference

| Category | Implemented | Deferred | Total |
|----------|-------------|----------|-------|
| Startup Optimization | 4 | 1 | 5 |
| Bootstrap Integrity | 2 | 0 | 2 |
| Watchdog System | 0 | 3 | 3 |
| **TOTAL** | **6** | **4** | **10** |

---

## Category 1: Startup Optimization (KDE-INV-056)

### ✅ REC-KDE-056-1: Dependency Auto-Check (IMPLEMENTED)

**Priority**: CRITICAL  
**Status**: ✅ IMPLEMENTED  
**Date**: 2026-07-26

**Action Taken**:
- Added `_try_user_install()` function to `.kde/bootstrap/gates.py`
- Auto-installs PyYAML using `--user` flag for sandbox compatibility
- Falls back to clear error message with manual install command

**Files Modified**:
- `.kde/bootstrap/gates.py` (lines 477-495)

**Verification**:
```bash
make kde-quick  # Auto-installs PyYAML if missing
```

---

### ✅ REC-KDE-056-2: Add --quick Flag (IMPLEMENTED)

**Priority**: HIGH  
**Status**: ✅ IMPLEMENTED  
**Date**: 2026-07-26

**Action Taken**:
- Modified `verify_all_gates()` to accept `quick` parameter
- Added `--quick` and `--full` CLI flags to gates.py
- Skip `go mod verify` in quick mode (saves 2.1 seconds)

**Performance Impact**:
| Mode | Time | Improvement |
|------|------|-------------|
| Full (`--full`) | 2.2s | Baseline |
| Quick (`--quick`) | 0.1s | **95% faster** |

**Files Modified**:
- `.kde/bootstrap/gates.py`

**Verification**:
```bash
python3 .kde/bootstrap/gates.py --quick  # ~0.1s
python3 .kde/bootstrap/gates.py --full   # ~2.2s
```

---

### ✅ REC-KDE-056-3: Add Makefile Targets (IMPLEMENTED)

**Priority**: MEDIUM  
**Status**: ✅ IMPLEMENTED  
**Date**: 2026-07-26

**Action Taken**:
- Added `kde-start`, `kde-check`, `kde-quick`, `kde-help` to Makefile
- Provides discoverable commands for KDE Runtime

**Commands Added**:
```makefile
make kde-start   # Start KDE runtime with quick preflight
make kde-check   # Run full preflight check
make kde-quick   # Run quick preflight check
make kde-help    # Show KDE-specific help
```

**Files Modified**:
- `Makefile` (lines 91-146)

---

### ✅ REC-KDE-056-4: Create requirements.txt (IMPLEMENTED)

**Priority**: MEDIUM  
**Status**: ✅ IMPLEMENTED  
**Date**: 2026-07-26

**Action Taken**:
- Created `.kde/requirements.txt` for Python dependencies
- Documents required packages with installation instructions

**Files Created**:
- `.kde/requirements.txt`

**Contents**:
```txt
# Install with: pip install --user -r .kde/requirements.txt
PyYAML>=6.0.0
```

---

### ⏳ REC-KDE-056-5: Create Unified kde-start Script (DEFERRED)

**Priority**: LOW  
**Status**: ⏳ DEFERRED  
**Reason**: Makefile targets provide similar functionality

**Proposed Action**:
- Create `scripts/kde-start` executable script
- Combine bootstrap check, preflight, and engine start
- Add to PATH for global access

**Future Work**:
```bash
# Proposed structure
scripts/kde-start
├── Check bootstrap status
├── Run preflight
├── Start engine
└── Handle errors gracefully
```

---

## Category 2: Bootstrap Integrity (KDE-INV-056, KDE-INV-057)

### ✅ REC-KDE-057-1: Bootstrap Status Checker (IMPLEMENTED)

**Priority**: HIGH  
**Status**: ✅ IMPLEMENTED  
**Date**: 2026-07-26

**Action Taken**:
- Created `BootstrapStatusChecker` class in `.kde/bootstrap/status.py`
- Verifies module existence and file integrity
- Provides JSON output for automation

**Features**:
- Module existence verification
- File checksum computation (SHA256)
- State.json validation
- Unexpected directory detection

**Files Created**:
- `.kde/bootstrap/status.py`

**Verification**:
```bash
make kde-status        # Human-readable status
python3 .kde/bootstrap/status.py --json  # JSON output
```

---

### ✅ REC-KDE-057-2: Bootstrap Watchdog (IMPLEMENTED)

**Priority**: HIGH  
**Status**: ✅ IMPLEMENTED  
**Date**: 2026-07-26

**Action Taken**:
- Created `BootstrapWatchdog` class in `.kde/bootstrap/status.py`
- Provides continuous monitoring mode
- Detects file changes from baseline

**Features**:
- Baseline checksum computation
- Change detection
- Continuous monitoring with configurable interval
- Interrupt handling

**Files Created**:
- `.kde/bootstrap/status.py`

**Verification**:
```bash
make kde-watch                  # Continuous monitoring (5s interval)
python3 .kde/bootstrap/status.py --watch --interval 10
```

---

## Category 3: Watchdog System (KDE-INV-057)

### ⏳ REC-KDE-057-3: ProcessBehaviorAnalyzer (DEFERRED)

**Priority**: HIGH  
**Status**: ⏳ DEFERRED  
**Reason**: Requires action tracking infrastructure

**Proposed Implementation**:
```python
class ProcessBehaviorAnalyzer:
    def detect_infinite_loop(self, history: List[str]) -> bool:
        """Detect if same action repeated N times"""
        
    def detect_memory_spiral(self, samples: List[int]) -> bool:
        """Detect if memory grows without bound"""
        
    def detect_file_spam(self, operations: List[FileOp]) -> bool:
        """Detect rapid file creation/deletion"""
```

**Rogue AI Patterns to Detect**:

| Pattern | Detection Method | Severity |
|---------|-----------------|----------|
| Infinite Loop | Action history repetition | CRITICAL |
| Memory Spiral | Monotonic memory growth | HIGH |
| File Spam | File count per time window | HIGH |
| Unauthorized Action | Allowed action list | CRITICAL |
| Stall | Timestamp delta | MEDIUM |
| Resource Exhaustion | CPU/memory thresholds | HIGH |

---

### ⏳ REC-KDE-057-4: ResourceMonitor (DEFERRED)

**Priority**: MEDIUM  
**Status**: ⏳ DEFERRED  
**Reason**: Requires additional dependencies (psutil)

**Proposed Implementation**:
```python
class ResourceMonitor:
    def get_memory_usage(self) -> int:
        import resource
        return resource.getrusage(resource.RUSAGE_SELF).ru_maxrss
        
    def get_cpu_usage(self) -> float:
        # Platform-specific implementation
        
    def check_thresholds(self, 
                        memory_pct: int = 80,
                        cpu_pct: int = 95) -> List[Alert]:
        """Check resource thresholds and return alerts"""
```

---

### ⏳ REC-KDE-057-5: Watchdog Dashboard (DEFERRED)

**Priority**: LOW  
**Status**: ⏳ DEFERRED  
**Reason**: Nice-to-have, not critical

**Proposed Features**:
- Web-based status dashboard
- Real-time monitoring visualization
- Alert history
- Bootstrap integrity graph

---

## Implementation Summary

### ✅ Implemented (6 Recommendations)

| ID | Recommendation | Files | Date |
|----|---------------|-------|------|
| REC-KDE-056-1 | Dependency Auto-Check | `.kde/bootstrap/gates.py` | 2026-07-26 |
| REC-KDE-056-2 | Add --quick Flag | `.kde/bootstrap/gates.py` | 2026-07-26 |
| REC-KDE-056-3 | Add Makefile Targets | `Makefile` | 2026-07-26 |
| REC-KDE-056-4 | Create requirements.txt | `.kde/requirements.txt` | 2026-07-26 |
| REC-KDE-057-1 | Bootstrap Status Checker | `.kde/bootstrap/status.py` | 2026-07-26 |
| REC-KDE-057-2 | Bootstrap Watchdog | `.kde/bootstrap/status.py` | 2026-07-26 |

### ⏳ Deferred (4 Recommendations)

| ID | Recommendation | Priority | Reason |
|----|---------------|----------|--------|
| REC-KDE-056-5 | Unified kde-start Script | LOW | Makefile sufficient |
| REC-KDE-057-3 | ProcessBehaviorAnalyzer | HIGH | Needs infrastructure |
| REC-KDE-057-4 | ResourceMonitor | MEDIUM | Needs psutil |
| REC-KDE-057-5 | Watchdog Dashboard | LOW | Nice-to-have |

---

## Quick Start Commands

```bash
# Bootstrap Status
make kde-status                    # Show status
make kde-watch                    # Watch continuously

# Preflight Checks  
make kde-quick                    # Quick check (~0.1s)
make kde-check                    # Full check (~2.2s)

# Start Engine
make kde-start                    # Start with verification

# Help
make kde-help                     # Show all commands
```

---

## Files Reference

| File | Purpose |
|------|---------|
| `.kde/bootstrap/gates.py` | Preflight checks, auto-install |
| `.kde/bootstrap/status.py` | Bootstrap status, watchdog |
| `.kde/requirements.txt` | Python dependencies |
| `Makefile` | Command aliases |
| `.kde/README.md` | Runtime documentation |

---

## Investigation Evidence

| Investigation | Topic | Recommendations |
|---------------|-------|----------------|
| KDE-INV-056 | Startup Optimization | 5 (4 impl, 1 defer) |
| KDE-INV-057 | Watchdog System | 5 (2 impl, 3 defer) |

---

*Document Status*: ACTIVE  
*Last Updated*: 2026-07-26  
*Next Review*: After implementing deferred recommendations
