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
| Five Core Principles | `.kde/runtime/principles_enforcer.py` | Actual enforcement code |
| Laboratory Workflow | `.kde/governance/STATE-MACHINE.md` | Verified document-state transitions; no nine-stage workflow file exists at the old path |
| Investigation Templates | Unavailable in this repository | Former investigation template path is absent; use only human-approved existing artifacts |
| Experiment Templates | Unavailable in this repository | Former experiment template path is absent; do not invent a replacement |
| ECU Runtime | `.kde/runtime/ecu/` | Execution control and governance |
| Pre-Flight Check | `.kde/runtime/preflight.py` | System readiness verification |

## Investigation Workflow

For explicitly selected KDE investigations only, follow the applicable, verified laboratory workflow and approval records. The previously referenced `laboratory/WORKFLOW.md` and two template paths are absent in this repository. `.kde/governance/STATE-MACHINE.md` supports document-state transitions, **not** a substitute nine-stage investigation procedure; do not invent or silently substitute workflow authority. The nine conceptual stages are:

1. IDEA → 2. INVESTIGATION → 3. EVIDENCE COLLECTION → 4. OBSERVATION → 5. SYNTHESIS → 6. VALIDATION → 7. CANDIDATE KNOWLEDGE → 8. PROMOTION → 9. KNOWLEDGE REPOSITORY

## Required Artifacts

For each explicitly human-approved investigation, use the verified repository-relative `laboratory/investigations/<approved-ID>/` path (not a filesystem-root `/laboratory/` path):
- `investigation.md` - Research question and scope
- `index.md` - Experiment index
- `links/` - Links to experiments

For an explicitly authorized experiment only, use `laboratory/experiments/<approved-ID>/` if that exact path and artifact schema are approved:
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
