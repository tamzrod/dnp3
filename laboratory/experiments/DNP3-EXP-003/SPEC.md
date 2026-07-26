# DNP3-EXP-003: Fyne Dependency Resolution Failure

**Objective**: Determine why cmd/workbench resolves the legacy module `github.com/fyne-io/fyne` instead of `fyne.io/fyne/v2`.

**Status**: COMPLETED + FIXED

---

## Investigation Summary

### Root Cause
Stale indirect require entry for `github.com/fyne-io/fyne v1.4.2` that was not actually pulled in by any dependency in the module graph.

### Evidence
1. `cmd/workbench/go.mod` line 10: `github.com/fyne-io/fyne v1.4.2 // indirect`
2. Only `fyne.io/fyne/v2` imports exist in the codebase
3. `fyne.io/fyne/v2@v2.4.0` does not require `github.com/fyne-io/fyne`
4. No `replace` directive introducing legacy fyne
5. No go.work file affecting resolution

### Classification
**Cause Type**: Stale require entry (not replace, not cached module, not workspace)

---

## Resolution

### Fix Applied
Removed `github.com/fyne-io/fyne v1.4.2 // indirect` from `cmd/workbench/go.mod`

### Verification Status
| Check | Status | Notes |
|-------|--------|-------|
| go.mod updated | ✅ | Removed stale entry |
| Build verification | ⚠️ | Pre-existing go-text/types issue |
| Dependencies valid | ⚠️ | go-text/types pseudo-version inaccessible |

### Known Issue
Pre-existing issue: `github.com/go-text/types@v0.0.0-20240903004611-7a47d0d1c6ba` is inaccessible from this environment due to git ls-remote restrictions. This is a separate issue from the fyne problem.

---

## Changes Made

| File | Change |
|------|--------|
| `cmd/workbench/go.mod` | Removed `github.com/fyne-io/fyne v1.4.2 // indirect` |

---

**Investigator**: OpenHands Agent  
**Date**: 2026-07-26  
**Authority**: User-approved implementation
