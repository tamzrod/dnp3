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

# 2. Read-only preflight (does not install dependencies)
python3 .kde/runtime/preflight.py

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

### Step 2: Inspect dependencies without installing

The KDE bootstrap gates are observation-only by default: missing PyYAML is reported as blocked, never automatically installed. A separate human-approved provisioning workflow is required for any installation.

```bash
python3 --version
python3 -c 'import yaml; print("PyYAML present")'
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

### Bootstrap Gates (observation-only; no dependency installation)

| Command | Purpose | Time |
|---------|---------|------|
| `python3 .kde/bootstrap/gates.py --quick` | Non-installing bootstrap checks | Available for explicitly selected KDE work |
| `python3 .kde/bootstrap/gates.py --full` | Non-installing full checks | Available for explicitly selected KDE work |

### Make Targets

If Make is available:

```bash
make kde-quick      # Non-installing KDE gate checks
make kde-check      # Non-installing full KDE gate checks
make kde-start      # Start KDE demo after non-installing checks; explicit KDE intent only
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
| `start engine` | preflight.py then explicit KDE runtime start | Does not authorize downloads or CWAL implementation |
| `preflight` | preflight.py | Run system check |
| `run demo` | runtime.demo() | Execute demo |
| `check state` | cat state.json | Read runtime state |
| `bootstrap` | gates.py --full | Non-installing environment checks; no task promotion |
| `health` | status.py | Show bootstrap status |

---

## Verification Checklist

After starting the engine, verify:

- [ ] Read-only preflight: `python3 .kde/runtime/preflight.py` result recorded; missing dependencies do not trigger installs
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

PyYAML not installed. Report the missing dependency and stop. Do not auto-install it during CWAL, PLAN, CHECK or implicit OpenHands startup. Provisioning requires separate human approval.

### "ImportError: attempted relative import with no known parent package"

This happens when importing from `.kde/runtime/` directly. Use:

```bash
python3 .kde/runtime/preflight.py  # Run as script
# NOT: python3 -m .kde.runtime.preflight  # This fails
```

### State shows "uninitialized"

Inspect runtime state and report the missing initialization. Bootstrap performs non-installing checks and cannot self-authorize provisioning or task execution.

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
| **Engine** | KDE-ENGINE-002 (Beta) | - | Default (see defaults.yaml) |
| **Seed** | SEED-003 (Bootstrap Validation) | - | Default (see defaults.yaml) |
| **Runtime** | KDE 1.1.0 | 1.1.0 | Initialized |
| **Project** | go-dnp3 | - | DNP3 Library |

---

**Document Status**: APPROVED
**For**: Fresh AI/Sandbox Initialization
**Updated**: 2026-07-27

## Separation from Operation CWAL (GOV-RECON-001)

- `start engine` initializes KDE within human-approved environment permissions; it does **not** select or promote any CWAL task.
- `operation cwal` uses the exact human-promoted CURRENT and matching markers; it does not automatically initialize KDE, install tools or create a laboratory investigation.
- `PLAN` and `CHECK` remain read-only; governance maintenance requires separate human authorization.
- The approved `.kde/governance/runtime/defaults.yaml` is authoritative for engine/seed defaults. This file reports those settings but never changes them.
- KDE bootstrap and Make's KDE checks are non-installing; no invocation grants authority to install missing dependencies or exceed the selected task's safety restrictions.
