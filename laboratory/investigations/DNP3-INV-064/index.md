# Investigation Index: DNP3-INV-064

**Title**: KDE Runtime Auto-Start Failure
**Status**: CLOSED
**Type**: Root Cause Analysis
**Created**: 2026-07-29

## Summary

Identified that the KDE Investigation Framework skill's Quick Start command is missing required Python path setup (`sys.path.insert(0, '.kde')`), causing `ModuleNotFoundError` when attempting to auto-start the runtime.

## Experiments

No experiments required - root cause identified through diagnostic commands.

## Evidence Files

| File | Description |
|------|-------------|
| `investigation.md` | Full investigation document |
| `evidence/failed-quickstart.log` | Failed skill Quick Start output |
| `evidence/working-command.log` | Successful corrected command |
| `evidence/bootstrap-gates.log` | Bootstrap gate verification |

## Related Links

- Skill: `.agents/skills/kde-investigation-framework.md`
- Bootstrap Gates: `.kde/bootstrap/gates.py`
- Preflight: `.kde/runtime/preflight.py`
- Start Engine: `start-engine.md` (correct reference)

## Fix

Update Quick Start in `.agents/skills/kde-investigation-framework.md` to include:
```python
import sys
sys.path.insert(0, '.kde')
```

## Status History

| Date | Status | Note |
|------|--------|------|
| 2026-07-29 | ACTIVE | Investigation opened |
| 2026-07-29 | CLOSED | Root cause identified |
