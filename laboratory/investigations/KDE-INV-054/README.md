---
id: KDE-INV-054
type: investigation
title: "OpenHands Automatic Runtime Bootstrap"
status: completed
authority: "KDE Runtime (DNP3 Library)"
created: "2026-07-26T04:30:00Z"
execution_agent: "OpenHands Agent"
---

# Investigation: OpenHands Automatic Runtime Bootstrap

**Investigation ID**: KDE-INV-054  
**Title**: OpenHands Automatic Runtime Bootstrap  
**Status**: COMPLETED  
**Date**: 2026-07-26  
**Authority**: KDE Runtime (DNP3 Library)  
**Execution Agent**: OpenHands Agent  

---

## 1. Executive Summary

### 1.1 Overview

This investigation implemented automated dependency installation for the KDE Runtime using OpenHands' `.openhands/setup.sh` feature. The solution ensures that all required dependencies (PyYAML, Go toolchain) are automatically installed at the start of each OpenHands conversation.

### 1.2 Key Findings

| Finding | Evidence |
|---------|----------|
| `.openhands/setup.sh` feature exists | OpenHands documentation |
| PyYAML required for KDE Runtime | `.kde/bootstrap/requirements.json` |
| Go toolchain required for Go projects | `gates.py` Gate B3 |
| Bootstrap gates can pass automatically | Test run showed 8/8 passed |

### 1.3 Outcome

**SUCCESS** - All hypotheses validated. Bootstrap gates now pass automatically.

---

## 2. Investigation Details

### 2.1 Problem

Each OpenHands conversation required manual installation of dependencies:
```bash
pip install pyyaml
# Manual Go installation required
```

Result: Bootstrap gates failed until manual steps were taken.

### 2.2 Solution

Created `.openhands/setup.sh` that automatically:
1. Installs PyYAML if not present
2. Downloads and installs Go 1.22.5 if not present
3. Downloads Go module dependencies
4. Runs bootstrap gates verification

### 2.3 Implementation

```bash
#!/bin/bash
# KDE Runtime Bootstrap Setup

# Install PyYAML
pip install pyyaml --quiet

# Install Go toolchain
if ! command -v go &> /dev/null; then
    curl -sL https://go.dev/dl/go1.22.5.linux-amd64.tar.gz -o /tmp/go.tar.gz
    sudo tar -C /usr/local -xzf /tmp/go.tar.gz
fi

# Download Go dependencies
go mod download
```

---

## 3. Evidence

### 3.1 Setup Script Created

**Location**: `.openhands/setup.sh`  
**Permissions**: Executable (`-rwxr-xr-x`)  
**Content**: See implementation above

### 3.2 Bootstrap Gates Verification

```
======================================================================
KDE BOOTSTRAP GATE VERIFICATION
======================================================================

--- Gate B1 ---
  [✓] runtime_state: PASSED
  [✓] experiments_directory: PASSED
  [✓] laboratory_rules: PASSED

--- Gate B2 ---
  [✓] git_log_check: PASSED
  [✓] git_status_check: PASSED

--- Gate B3 ---
  [✓] python_runtime: PASSED
  [✓] go_toolchain: PASSED
  [✓] go_dependencies: PASSED

======================================================================
RESULT: PASSED
Summary: Bootstrap gates verified: 8/8 checks passed. Can proceed with investigation.
======================================================================
```

### 3.3 Files Created

| File | Purpose |
|------|---------|
| `.openhands/setup.sh` | Automatic setup script |
| `laboratory/investigations/KDE-INV-054/SPEC.md` | Investigation specification |
| `laboratory/investigations/KDE-INV-054/README.md` | This document |
| `laboratory/investigations/KDE-INV-054/CONCLUSION.md` | Investigation conclusion |

---

## 4. Hypothesis Results

### H1: Setup script will install dependencies automatically

**Status**: ✅ VALIDATED  
**Evidence**: Script executed successfully, all dependencies installed

### H2: Bootstrap gates will pass automatically

**Status**: ✅ VALIDATED  
**Evidence**: 8/8 checks passed after setup script execution

### H3: Go modules will be downloaded

**Status**: ✅ VALIDATED  
**Evidence**: `go mod download` completed successfully

---

## 5. Recommendations

### REC-001: Adopt Automatic Bootstrap

**Recommendation**: Implement `.openhands/setup.sh` as the standard bootstrap mechanism for all OpenHands conversations.

**Rationale**:
1. Eliminates manual setup steps
2. Ensures consistent environment
3. Validates dependencies automatically
4. Reduces friction for new sessions

### REC-002: Document Setup Script

**Recommendation**: Add documentation about `.openhands/setup.sh` to the KDE Runtime documentation.

**Rationale**:
1. Makes the solution discoverable
2. Enables customization for project needs
3. Provides troubleshooting guidance

---

## 6. Limitations

### 6.1 Current Limitations

- Go version is hardcoded (1.22.5)
- Assumes sudo access for Go installation
- Only handles Python and Go dependencies

### 6.2 Future Improvements

- Make Go version configurable
- Add error handling and retry logic
- Support additional project types
- Add dependency version checking

---

## 7. Related Documents

| Document | Relationship |
|----------|-------------|
| [.openhands/setup.sh](../../../../.openhands/setup.sh) | Implementation artifact |
| [.kde/bootstrap/gates.py](../../../../.kde/bootstrap/gates.py) | Bootstrap verification |
| [SPEC.md](./SPEC.md) | Investigation specification |

---

## 8. Appendix: Full Setup Script

```bash
#!/bin/bash
# KDE Runtime Bootstrap Setup
# This script runs automatically when an OpenHands conversation starts

set -e

echo "=========================================="
echo "KDE Runtime Bootstrap Setup"
echo "=========================================="

cd /workspace/project/dnp3

# Install PyYAML
if ! python3 -c "import yaml" 2>/dev/null; then
    echo "[1/3] Installing PyYAML..."
    pip install pyyaml --quiet
fi

# Install Go toolchain
if ! command -v go &> /dev/null; then
    echo "[2/3] Installing Go toolchain..."
    curl -sL https://go.dev/dl/go1.22.5.linux-amd64.tar.gz -o /tmp/go.tar.gz
    sudo tar -C /usr/local -xzf /tmp/go.tar.gz
fi

export PATH=$PATH:/usr/local/go/bin

# Download Go modules
if [ -f go.mod ]; then
    echo "[3/3] Downloading Go modules..."
    go mod download
fi

# Run bootstrap gates
python3 .kde/bootstrap/gates.py --project-type go
```

---

*Investigation completed: 2026-07-26*  
*Execution Agent: OpenHands Agent*  
*Classification: RUNTIME INFRASTRUCTURE*
