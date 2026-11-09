---
name: kde-investigation-framework
type: repo
triggers:
  - investigation
  - experiment
  - start engine
  - preflight check
---

# KDE Investigation Framework

This repository operates under the Knowledge-Driven Engineering (KDE) Laboratory methodology.

## Core Rules

1. **Explicit investigation route only**: Use this methodology when a human requests a KDE investigation or experiment. Do not reinterpret `operation cwal`, `PLAN`, `CHECK`, governance maintenance, or ordinary implementation as laboratory investigations simply because a KDE runtime is installed or running.

2. **File protection and approval**: KDE investigation artifacts remain governed by applicable KDE protection rules. No file outside the human-approved route's exact edit allowlist may be modified. An explicitly human-promoted CWAL CURRENT is separate implementation authorization, never a general bypass of KDE policy. Stopping KDE does not waive repository safety or approval rules.

## Investigation Protocol

For explicitly selected KDE investigations only:

1. **Acknowledge the Five Core Principles** before any work
2. **Follow the Bootstrap Sequence** for each session
3. **Run Pre-Flight Check** to verify system readiness
4. **Document findings** with proper evidence classification
5. **Never auto-continue** without human authorization
6. **Never self-approve** your own work
7. **Never self-promote** knowledge without human approval

## Five Core Principles

| Rule | Enforcement |
|------|-------------|
| No Auto-Continuation | Checkpoints block unauthorized continuation |
| No Self-Approval | Blocks REVIEW → APPROVED for AI |
| No Self-Promotion | Blocks VALIDATED → PROMOTED for AI |
| Distinguish Evidence | Classifies content by evidence level |
| Evidence-Based Changes | Requires evidence citations |

## Startup Command

To initialize the KDE runtime, run:

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

## Route precedence (GOV-RECON-001)

- `operation cwal`: follow `AGENTS.md`, `operation cwal.md`, the six matching CURRENT markers and exactly the human-promoted CURRENT; do not run a KDE investigation or auto-bootstrap by default.
- `PLAN` / `CHECK`: read-only design / independent audit. No implementation authority.
- `start engine` / explicit KDE investigation: follow approved KDE governance and session permissions. Runtime readiness alone cannot authorize edits or installs.
- Governance maintenance: requires a separate human-approved scope, branch and file allowlist; never modifies an active CWAL CURRENT or its markers.
- On conflicting instructions, preserve the stricter safety prohibition and stop for human resolution. Never self-promote or auto-continue.
