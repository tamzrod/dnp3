# Start Engine

**Aliases**: `start engine`, `start-runtime`, `initialize kde`, `init kde`, `run`, `bootstrap`

---

## Purpose

This file provides the canonical procedure for starting the KDE Runtime engine in this project (go-dnp3). It serves as the AI guide for fresh sandbox and fresh AI initialization.

---

## Quick Start (Copy-Paste)

For a fresh AI/sandbox session, run these commands in order:

```bash
# 1. Verify Python available
python3 --version

# 2. Run preflight check (auto-installs PyYAML if needed)
python3 .kde/bootstrap/gates.py --quick

# 3. Start the KDE Runtime
python3 -c "import sys; sys.path.insert(0, '.kde'); from runtime.runtime import demo; demo()"

# 4. Verify runtime state
cat .kde/runtime/state.json
```

---

## Bootstrap Sequence

When you say "start engine", "start runtime", or similar commands, follow this procedure:

### Step 1: Acknowledge Laboratory Rules

Before any work, acknowledge the Five Core Principles from `.kde/seeds/seed-001/`:

| Rule | Description |
|------|-------------|
| **No Auto-Continuation** | Never begin next session without human authorization |
| **No Self-Approval** | Never approve your own work |
| **No Self-Promotion** | Never promote knowledge to production |
| **Distinguish Evidence** | Mark fact vs. conclusion vs. speculation |
| **Evidence-Based Changes** | All claims must be justified |

### Step 2: Initialize Dependencies

Dependencies are automatically installed by the bootstrap gates:

```bash
# Auto-installs PyYAML if missing
python3 .kde/bootstrap/gates.py --quick
```

If you need to install manually:

```bash
pip install pyyaml
```

### Step 3: Run Preflight Check

Run the comprehensive pre-flight check to verify readiness:

```bash
python3 .kde/runtime/preflight.py
```

Expected output:
```
==============================================================================
PRE-FLIGHT CHECK - KDE RUNTIME
==============================================================================
■ RUNTIME HEALTH         ✅ [status]
■ ECU COMPONENT STATUS   ✅ [components]
■ GOVERNANCE STATUS      ✅ [rules]
■ MISSION READINESS      [status]
==============================================================================
```

### Step 4: Initialize KDE Runtime ECU

```python
import sys
sys.path.insert(0, '.kde')

from runtime.ecu import create_ecu

# Get project root (directory containing .kde/)
import os
project_root = os.path.dirname(os.path.abspath('.kde'))
ecu = create_ecu(project_root)
```

### Step 5: Start Runtime Demo (Optional)

To run the demonstration routine:

```bash
python3 -c "import sys; sys.path.insert(0, '.kde'); from runtime.runtime import demo; demo()"
```

---

## Detailed Command Reference

### Bootstrap Gates

| Command | Purpose | Time |
|---------|---------|------|
| `python3 .kde/bootstrap/gates.py --quick` | Quick preflight (~0.1s) | Fast iteration |
| `python3 .kde/bootstrap/gates.py --full` | Full preflight (~2.2s) | CI/CD validation |

### Make Targets

If Make is available:

```bash
make kde-quick      # Quick preflight check
make kde-check      # Full preflight check  
make kde-start      # Start runtime with quick preflight
make kde-status     # Show bootstrap status
```

### Direct Runtime Access

```bash
# Run demo
python3 -c "import sys; sys.path.insert(0, '.kde'); from runtime.runtime import demo; demo()"

# Check state
cat .kde/runtime/state.json

# Run preflight
python3 .kde/runtime/preflight.py

# Bootstrap status
python3 .kde/bootstrap/status.py
```

---

## Project Structure Context

This project uses a non-standard KDE structure:

```
project/
├── .kde/                    # KDE Runtime root (NOT /kde/)
│   ├── runtime/             # Core runtime (inside .kde/)
│   │   ├── ecu/             # ECU components
│   │   ├── preflight.py     # Pre-flight checks
│   │   └── ...
│   ├── engines/              # Investigation engines
│   ├── seeds/                # Seed knowledge
│   └── ...
└── laboratory/              # Engineering artifacts (NOT in .kde/)
    ├── investigations/
    ├── experiments/
    └── evidence/
```

**Key Difference**: The `runtime/` directory is at `.kde/runtime/`, not at the project root.

---

## Import Path Patterns

When writing code that imports KDE modules:

```python
# Correct - insert .kde into sys.path
import sys
sys.path.insert(0, '.kde')
from runtime.runtime import demo

# Correct - direct import after path setup
from runtime.ecu import create_ecu
from runtime.retrieval import RetrievalEngine

# Incorrect - this will fail
from runtime import demo  # No .kde in path!
```

---

## Canonical Commands

| Your Command | Maps To | Description |
|-------------|---------|-------------|
| `start engine` | gates.py --quick | Initialize runtime |
| `preflight` | preflight.py | Run system check |
| `run demo` | runtime.demo() | Execute demo |
| `check state` | cat state.json | Read runtime state |
| `bootstrap` | gates.py --full | Full initialization |
| `health` | status.py | Show bootstrap status |

---

## Verification Checklist

After starting the engine, verify:

- [ ] Preflight passes: `python3 .kde/bootstrap/gates.py --quick` shows PASSED
- [ ] State is ready: `cat .kde/runtime/state.json` shows `"status": "initialized"`
- [ ] Modules loaded: State shows 11 modules
- [ ] Python available: `python3 --version` shows 3.10+
- [ ] PyYAML available: `python3 -c "import yaml"` succeeds

---

## Troubleshooting

### "Module not found: runtime"

The `.kde/` directory is not in sys.path. Add:

```python
import sys
sys.path.insert(0, '.kde')
```

### "No module named yaml"

PyYAML not installed:

```bash
pip install pyyaml
```

### "ImportError: attempted relative import with no known parent package"

This happens when importing from `.kde/runtime/` directly. Use:

```bash
python3 .kde/runtime/preflight.py  # Run as script
# NOT: python3 -m .kde.runtime.preflight  # This fails
```

### State shows "uninitialized"

Run bootstrap:

```bash
python3 .kde/bootstrap/gates.py --quick
```

---

## Related Files

| File | Purpose |
|------|---------|
| `.kde/README.md` | Full KDE Runtime documentation |
| `.kde/bootstrap/gates.py` | Bootstrap gate verification |
| `.kde/runtime/preflight.py` | Pre-flight health checks |
| `.kde/runtime/state.json` | Runtime state |
| `.kde/seeds/seed-001/` | Core seed principles |

---

## Active Configuration

| Component | ID | Version | Status |
|-----------|-----|---------|--------|
| **Engine** | KDE-ENGINE-004 (Delta) | - | Active |
| **Seed** | SEED-001 (Genesis) | - | FROZEN |
| **Runtime** | KDE 1.1.0 | 1.1.0 | Initialized |
| **Project** | go-dnp3 | - | DNP3 Library |

---

**Document Status**: APPROVED
**For**: Fresh AI/Sandbox Initialization
**Updated**: 2026-07-27
