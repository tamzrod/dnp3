---
name: kde-investigation-framework
type: repo
triggers:
  - investigation
  - experiment
  - start engine
  - preflight check
---

# KDE Investigation Framework Skill

This repository operates under the Knowledge-Driven Engineering (KDE) Laboratory methodology.

## Purpose

This skill provides guidance for conducting investigations and experiments within the KDE Laboratory framework.

## Authoritative Sources

| Content | Location | Purpose |
|---------|----------|---------|
| Five Core Principles | `runtime/principles_enforcer.py` | Actual enforcement code |
| Laboratory Workflow | `laboratory/WORKFLOW.md` | Investigation lifecycle (9 stages) |
| Investigation Templates | `laboratory/templates/investigation-template.md` | Investigation artifact template |
| Experiment Templates | `laboratory/templates/experiment-template.md` | Experiment artifact template |
| ECU Runtime | `runtime/ecu/` | Execution control and governance |
| Pre-Flight Check | `runtime/preflight.py` | System readiness verification |

## Investigation Workflow

For explicitly selected KDE investigations only, follow the applicable, verified laboratory workflow and approval records. The previously referenced `laboratory/WORKFLOW.md` is absent in this repository; do not invent or silently substitute its authority. The nine conceptual stages are:

1. IDEA → 2. INVESTIGATION → 3. EVIDENCE COLLECTION → 4. OBSERVATION → 5. SYNTHESIS → 6. VALIDATION → 7. CANDIDATE KNOWLEDGE → 8. PROMOTION → 9. KNOWLEDGE REPOSITORY

## Required Artifacts

For each investigation, create artifacts in `/laboratory/investigations/INV-XXX/`:
- `investigation.md` - Research question and scope
- `index.md` - Experiment index
- `links/` - Links to experiments

For each experiment, create artifacts in `/laboratory/experiments/LAB-XXX/`:
- `experiment.md` - Experiment definition with hypothesis
- `runs/` - Run records
- `evidence/` - Evidence files

## Version Stamping

All investigation artifacts MUST include:
- **Engine**: KDE-ENGINE-XXX (vX.Y.Z)
- **Seed**: SEED-XXX (vX.Y.Z)
- Timestamps in ISO-8601 UTC format

## Quick Start

To initialize the KDE runtime and verify system readiness:

```bash
python3 -c "
import sys
sys.path.insert(0, '.kde')
from runtime.preflight import run_preflight_check, format_report
report = run_preflight_check()
print(format_report(report))
"
```

> **Note**: The `sys.path.insert(0, '.kde')` is required because the KDE runtime modules are located at `.kde/runtime/`, which is not in the default Python path.

Confirm the current authoritative laboratory procedure before execution; if absent or conflicting, stop and seek human clarification. This skill cannot promote a CWAL task.

## Route precedence (GOV-RECON-001)

- `operation cwal`: follow `AGENTS.md`, `operation cwal.md`, the six matching CURRENT markers and exactly the human-promoted CURRENT; do not run a KDE investigation or auto-bootstrap by default.
- `PLAN` / `CHECK`: read-only design / independent audit. No implementation authority.
- `start engine` / explicit KDE investigation: follow approved KDE governance and session permissions. Runtime readiness alone cannot authorize edits or installs.
- Governance maintenance: requires a separate human-approved scope, branch and file allowlist; never modifies an active CWAL CURRENT or its markers.
- On conflicting instructions, preserve the stricter safety prohibition and stop for human resolution. Never self-promote or auto-continue.
