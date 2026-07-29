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

1. **Laboratory Investigation**: Every prompt must be treated as a Laboratory Investigation when the KDE Engine is running.

2. **File Protection**: No files outside the `/laboratory/` directory may be edited without explicit human approval. Exception: runtime operation files in `/runtime/` may be modified during engine operations. This rule is also governed by engine running state - protections end when the engine stops.

## Investigation Protocol

When conducting investigations:

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
from runtime.preflight import run_preflight_check, format_report
report = run_preflight_check()
print(format_report(report))
"
```
