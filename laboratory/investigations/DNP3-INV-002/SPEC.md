---
id: DNP3-INV-002
type: investigation
title: "Fyne API Compatibility: Recurring Build Failures in Workbench UI"
status: IN_PROGRESS
authority: "DNP3 Library (Workbench)"
created: "2026-07-27T10:35:00Z"
execution_agent: "OpenHands Agent"
engine: KDE-ENGINE-002 (Beta)
---

# Investigation Specification: Fyne API Compatibility Issues

**Investigation ID**: DNP3-INV-002
**Title**: Fyne API Compatibility: Recurring Build Failures in Workbench UI
**Status**: IN_PROGRESS
**Engine**: KDE-ENGINE-002 (Beta)
**Date**: 2026-07-27

---

## 1. Background

### 1.1 Problem Statement

The DNP3 Workbench UI (cmd/workbench) has experienced **recurring build failures** due to Fyne API compatibility issues. Each attempt to implement UX improvements introduces new API calls that don't exist in the current Fyne version.

### 1.2 Recurring Pattern

| Date | Issue | Root Cause |
|------|-------|------------|
| 2026-07-27 | SetTooltip undefined | Button lacks SetTooltip |
| 2026-07-27 | SetMinSize undefined | Window lacks SetMinSize |
| 2026-07-27 | Geometry/Position undefined | Window lacks these methods |
| 2026-07-27 | TextStyle.Color undefined | TextStyle has no Color field |
| 2026-07-27 | ProgressBar.Start/Stop undefined | ProgressBar lacks animation methods |
| 2026-07-27 | ViewBottomSheetIcon undefined | Theme lacks this icon |
| 2026-07-27 | widget.TextStyle | Should be fyne.TextStyle |

### 1.3 Impact

- **Build Time**: Each iteration requires manual fix and re-push
- **Trust Erosion**: Code changes appear untested
- **Efficiency Loss**: Repeated cycles of fix-push-test-fail

---

## 2. Investigation Questions

### 2.1 Primary Questions

1. **What Fyne version is the project using?**
2. **What API methods actually exist in the Fyne version?**
3. **Why are API calls being made that don't exist?**
4. **How can we prevent future compatibility issues?**

### 2.2 Secondary Questions

1. Is there a Fyne API reference available?
2. Can we add automated API validation?
3. What is the process for checking Fyne documentation before coding?

---

## 3. Evidence to Collect

### 3.1 Code Evidence

- [ ] go.mod dependency versions
- [ ] Fyne widget.Button API surface
- [ ] Fyne Window API surface
- [ ] Fyne theme icons available

### 3.2 Process Evidence

- [ ] IDE/editor used for coding
- [ ] Fyne documentation access
- [ ] Testing process before commit

### 3.3 Historical Evidence

- [ ] Previous Fyne issues (if documented)
- [ ] API version pinning history

---

## 4. Success Criteria

1. **Identified**: Root cause of recurring API issues
2. **Documented**: Valid Fyne API methods for Workbench
3. **Process**: New coding process to prevent issues
4. **Validated**: Code compiles without Fyne API errors

---

## 5. Investigation Plan

### Phase 1: Fyne Version Analysis
1. Check go.mod for Fyne version
2. Research Fyne API documentation
3. Document valid API surface

### Phase 2: Code Audit
1. Audit current toolbar.go
2. Audit current window.go
3. Audit current statusbar.go
4. Audit all UI panel files

### Phase 3: Process Improvement
1. Create Fyne API cheat sheet
2. Add pre-commit validation (if possible)
3. Document lessons learned

---

## 6. Stakeholders

- **Author**: OpenHands Agent
- **Owner**: DNP3 Library maintainers
- **Reviewer**: Human maintainer (before merge)

---

## 7. Constraints

- Cannot change Fyne version without dependency analysis
- Must maintain UX feature functionality
- Must work within existing architecture
