---
id: KDE-INV-056
type: investigation
title: "Optimize Start Engine and Preflight Check Performance"
status: COMPLETED
authority: "KDE Runtime (DNP3 Library)"
created: "2026-07-26T21:55:00Z"
execution_agent: "OpenHands Agent"
engine: KDE-ENGINE-002 (Beta)
---

# Investigation: Optimize Start Engine and Preflight Check Performance

**Investigation ID**: KDE-INV-056  
**Title**: Optimize Start Engine and Preflight Check Performance  
**Status**: IN_PROGRESS  
**Engine**: KDE-ENGINE-002 (Beta)  
**Date**: 2026-07-26  
**Authority**: KDE Runtime (DNP3 Library)

---

## 1. Executive Summary

### 1.1 Overview

This investigation analyzes the KDE Runtime startup process and preflight check mechanism, identifying bottlenecks and proposing optimizations to minimize manual intervention and improve user experience.

### 1.2 Key Findings

| Finding | Evidence | Impact |
|---------|----------|--------|
| PyYAML not bundled | runtime.py fails with ImportError | BLOCKING |
| Go toolchain not pre-installed | gates.py fails B3 check | BLOCKING |
| `go mod verify` takes 2.2s | Measured via profiling | SLOW |
| No Makefile targets | Missing kde-start, kde-check | UX |
| Demo requires sys.path hacks | Manual path manipulation | UX |

### 1.3 Recommendation

**Implement a unified startup script** (`kde-start`) that:
1. Auto-installs Python dependencies
2. Auto-installs Go toolchain if missing
3. Caches dependency check results
4. Provides clear error messages

---

## 2. Evidence

### 2.1 Evidence E1: Startup Failure Logs

**Type**: Direct Observation  
**Source**: Terminal output  
**Relevance**: Documents blocking issues

```
# First attempt - Missing PyYAML
$ python3 .kde/runtime/runtime.py
ImportError: No module named 'yaml'

# Second attempt - Missing Go
$ python3 .kde/bootstrap/gates.py
RESULT: FAILED
[✗] go_available: FAILED: Go toolchain not available
```

### 2.2 Evidence E2: Bootstrap Gate Timing Profile

**Type**: Calculation  
**Source**: Profiling script  
**Relevance**: Identifies bottleneck

| Component | Time | Percentage |
|-----------|------|------------|
| check_go_dependencies() | 2201ms | 98% |
| check_python_runtime() | 40ms | 2% |
| check_go_toolchain() | 6ms | <1% |
| Gate B1 checks | <1ms | <1% |
| Gate B2 git checks | 12ms | <1% |

**Bottleneck**: `go mod verify` command takes 2.2 seconds.

### 2.3 Evidence E3: Makefile Analysis

**Type**: Document Analysis  
**Source**: Makefile  
**Relevance**: No KDE-specific targets exist

Current targets:
- `help`, `docs`, `test`, `build`, `bootstrap`
- Missing: `kde-start`, `kde-check`, `kde-gates`

---

## 3. Current State Analysis

### 3.1 Startup Flow Diagram

```
User Command
     │
     ▼
┌─────────────────────────────────────────────────────────────┐
│ 1. python3 .kde/runtime/runtime.py                         │
│    ├── Import yaml → FAILS (not installed)                 │
│    └── Needs: pip install pyyaml                          │
└─────────────────────────────────────────────────────────────┘
     │
     ▼ (after pip install)
┌─────────────────────────────────────────────────────────────┐
│ 2. python3 .kde/bootstrap/gates.py                        │
│    ├── B1: Runtime state → OK (11ms)                      │
│    ├── B2: Git log/status → OK (12ms)                     │
│    ├── B3: Python check → OK (40ms)                       │
│    ├── B3: Go toolchain → FAILS (not installed)           │
│    └── B3: Go deps → SKIPPED                              │
└─────────────────────────────────────────────────────────────┘
     │
     ▼ (after Go install)
┌─────────────────────────────────────────────────────────────┐
│ 3. python3 .kde/bootstrap/gates.py                        │
│    ├── B3: go mod verify → 2201ms (slow!)                │
│    └── RESULT: PASSED but slow                           │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 Issues Catalog

| ID | Issue | Severity | Current State | Root Cause |
|----|-------|----------|---------------|------------|
| ISSUE-1 | Missing PyYAML | CRITICAL | Manual install | Not bundled |
| ISSUE-2 | Missing Go toolchain | CRITICAL | Manual install | Not bundled |
| ISSUE-3 | Slow go mod verify | HIGH | Every run | `go mod verify` is slow |
| ISSUE-4 | No startup script | MEDIUM | Manual commands | Missing entry point |
| ISSUE-5 | No Make targets | LOW | Not discoverable | Missing integration |

---

## 4. Optimization Opportunities

### 4.1 OPT-1: Bundle PyYAML (CRITICAL)

**Problem**: Runtime fails without PyYAML installed.

**Solution Options**:

| Option | Pros | Cons | Feasibility |
|--------|------|------|-------------|
| A: Add requirements.txt | Simple | May not match system Python | HIGH |
| B: Inline YAML parser | No deps | Adds code complexity | MEDIUM |
| C: Check + install | Auto-fix | Modifies environment | HIGH |

**Recommendation**: Option C - Add dependency check with auto-install.

### 4.2 OPT-2: Auto-install Go Toolchain (CRITICAL)

**Problem**: Go toolchain not available in all environments.

**Solution Options**:

| Option | Pros | Cons | Feasibility |
|--------|------|------|-------------|
| A: Document requirement | Simple | Poor UX | HIGH |
| B: Detect + download | Auto-fix | Complex logic | MEDIUM |
| C: Skip if not Go project | Graceful | May miss issues | HIGH |

**Recommendation**: Option B - Detect missing Go, provide install instructions.

### 4.3 OPT-3: Cache Dependency Check Results (HIGH)

**Problem**: `go mod verify` takes 2.2 seconds every run.

**Solution Options**:

| Option | Pros | Cons | Feasibility |
|--------|------|------|-------------|
| A: Cache results in file | Fast subsequent runs | Stale cache risk | HIGH |
| B: Use go list instead | Faster | Less thorough | MEDIUM |
| C: Skip if go.mod unchanged | Smart caching | Complex | HIGH |
| D: Make it optional | Fast when skipped | May miss issues | HIGH |

**Recommendation**: Option D - Make slow checks optional with `--full` flag.

### 4.4 OPT-4: Create Unified Startup Script (MEDIUM)

**Problem**: Multiple manual steps required.

**Solution**: Create `kde-start` script that:
1. Checks Python deps (auto-install if needed)
2. Checks Go toolchain (skip if not Go project)
3. Runs bootstrap gates
4. Starts runtime demo

### 4.5 OPT-5: Add Makefile Targets (LOW)

**Problem**: KDE commands not discoverable.

**Solution**: Add to Makefile:
```makefile
## kde-start - Start KDE runtime
kde-start:
	@python3 .kde/bootstrap/gates.py && python3 .kde/runtime/runtime.py

## kde-check - Run preflight check
kde-check:
	@python3 .kde/bootstrap/gates.py
```

---

## 5. Recommendations

### REC-1: Implement Dependency Auto-Check (HIGH PRIORITY) ✅ IMPLEMENTED

**Action**: Modified bootstrap gates to:
1. Check PyYAML availability
2. Auto-install using `--user` flag (sandbox compatible)
3. Provide clear error with installation instructions if auto-install fails

**Effort**: LOW (1-2 hours)  
**Impact**: CRITICAL (blocks startup) → Now auto-resolves!

### REC-2: Make Slow Checks Optional (HIGH PRIORITY)

**Action**: Add `--quick` flag to skip slow checks:
```bash
python3 .kde/bootstrap/gates.py --quick  # Skip go mod verify
python3 .kde/bootstrap/gates.py --full   # Run all checks
```

**Effort**: LOW (1 hour)  
**Impact**: HIGH (2.2s → 0.2s for quick mode)

### REC-3: Create kde-start Script (MEDIUM PRIORITY)

**Action**: Create executable script `scripts/kde-start`:
```bash
#!/bin/bash
set -e
cd "$(dirname "$0")/.."

echo "=== KDE Runtime Startup ==="

# Quick preflight
echo "Running preflight check..."
python3 .kde/bootstrap/gates.py --quick || {
    echo "Warning: Preflight issues detected"
}

# Start runtime
echo "Starting KDE Engine..."
python3 -c "
import sys
sys.path.insert(0, '.kde')
from runtime.runtime import demo
demo()
"
```

**Effort**: MEDIUM (2-3 hours)  
**Impact**: HIGH (improves UX)

### REC-4: Add Makefile Targets (LOW PRIORITY)

**Action**: Add to Makefile:
```makefile
## kde-start - Start KDE runtime
kde-start:
	@echo "Starting KDE Runtime..."
	@python3 .kde/bootstrap/gates.py --quick
	@python3 -c "import sys; sys.path.insert(0, '.kde'); from runtime.runtime import demo; demo()"

## kde-check - Run full preflight check
kde-check:
	@python3 .kde/bootstrap/gates.py

## kde-quick - Run quick preflight check
kde-quick:
	@python3 .kde/bootstrap/gates.py --quick
```

**Effort**: LOW (30 minutes)  
**Impact**: MEDIUM (discovers KDE commands)

---

## 6. Implementation Plan

### Phase 1: Quick Wins ✅ COMPLETED
1. ✅ Add PyYAML error message improvement
2. ✅ Add `--quick` flag to gates.py
3. ✅ Add Makefile targets
4. ✅ Auto-install PyYAML with `--user` flag

### Phase 2: Script Creation (Next Session)
1. Create `scripts/kde-start`
2. Make it executable
3. Add to PATH or document usage

### Phase 3: Future Improvements
1. Interactive setup wizard
2. Go toolchain auto-download (if feasible)
3. Requirements.txt integration

---

## 7. Related Artifacts

| Artifact | Type | Relationship |
|----------|------|--------------|
| KDE-INV-055 | Investigation | ECU configuration |
| KDE-INV-051 | Investigation | Bootstrap gates |
| .kde/bootstrap/gates.py | Source | Gate implementation |
| .kde/runtime/runtime.py | Source | Runtime implementation |

---

## 8. Conclusion

The KDE Runtime startup process can be significantly optimized by:

1. **Immediate**: Adding `--quick` flag to skip slow `go mod verify`
2. **Immediate**: Adding Makefile targets for discoverability
3. **Short-term**: Creating unified `kde-start` script
4. **Long-term**: Implementing auto-dependency resolution

These changes will reduce startup friction and improve user experience.

---

*Investigation Status*: IN_PROGRESS  
*Next Step*: Implement REC-1 and REC-2  
*Human Review Required*: Yes
