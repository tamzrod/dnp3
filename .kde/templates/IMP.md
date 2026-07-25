# Implementation Specification Template

**File**: IMP.md
**Version**: 2.0.0
**Date**: 2026-07-25
**Source**: KDE-INV-043

---

## Purpose

This template provides the standard format for Implementation Specifications (IMP).

## Usage

Copy this template to create a new IMP:

```bash
cp .kde/templates/IMP.md laboratory/implementations/PROJECT-IMP-XXX/SPEC.md
```

## Template

```markdown
# Implementation Specification: PROJECT-IMP-XXX

**ID**: PROJECT-IMP-XXX
**Title**: [Implementation Title]
**Status**: DRAFT | APPROVED | COMPLETED
**Date**: YYYY-MM-DD
**Author**: [Author]
**Human Reviewer**: [Reviewer]

## Precondition Verification

| Component | Status | Evidence |
|-----------|--------|----------|
| Source Investigation | YES/NO | PROJECT-INV-XXX |
| Source Decision | YES/NO | TDR-XXX |
| Human Review | YES/NO | [Date] |

## 1. Overview

Brief description of what this implementation accomplishes.

## 2. Source Artifacts

### 2.1 Source Investigations

| Investigation | Relevance |
|--------------|-----------|
| PROJECT-INV-XXX | [Brief relevance] |

### 2.2 Source Experiments

| Experiment | Relevance |
|------------|-----------|
| PROJECT-EXP-XXX | [Brief relevance] |

### 2.3 Source Decisions

| Decision | Status |
|----------|--------|
| TDR-XXX | APPROVED |

## 3. Scope

### 3.1 In Scope

| Item | Description |
|------|-------------|
| 1 | [Description] |

### 3.2 Out of Scope

| Item | Reason |
|------|--------|
| 1 | [Reason] |

## 4. Acceptance Criteria

| # | Criterion | Verification Method |
|---|-----------|-------------------|
| 1 | [Criterion] | [Method] |

## 5. Dependencies

| Dependency | Status | Notes |
|------------|--------|-------|
| [Dependency] | Ready/Blocked | [Notes] |

## 6. Implementation Plan

| Phase | Task | Deliverable |
|-------|------|-------------|
| 1 | [Task] | [Deliverable] |

## 7. Verification Artifacts

| Artifact | Description |
|----------|-------------|
| [Artifact] | [Description] |

## 8. Related Commits

| Commit | Description |
|--------|-------------|
| [Hash] | [Description] |

## 9. Revision History

| Version | Date | Change |
|---------|------|--------|
| 1.0.0 | YYYY-MM-DD | Initial version |

**Status**: DRAFT
**Authority**: Human
```

*Per KDE-INV-043*
