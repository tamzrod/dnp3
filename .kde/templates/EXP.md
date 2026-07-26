# Experiment Template

**File**: EXP.md
**Version**: 1.0.0
**Date**: 2026-07-26
**Source**: KDE-INV-052

---

## Purpose

This template provides the standard format for Experiments.

## Usage

Copy this template to create a new Experiment:

```bash
mkdir -p laboratory/experiments/PROJECT-EXP-XXX
cp .kde/templates/EXP.md laboratory/experiments/PROJECT-EXP-XXX/README.md
cp .kde/templates/INV-SPEC.md laboratory/experiments/PROJECT-EXP-XXX/SPEC.md
cp .kde/templates/INV-CONCLUSION.md laboratory/experiments/PROJECT-EXP-XXX/CONCLUSION.md
```

## Template

### README.md

```markdown
---
id: PROJECT-EXP-XXX
type: experiment
title: "[Experiment Title]"
authority: "KDE Runtime (DNP3 Library)"
status: IN_PROGRESS
date: "YYYY-MM-DD"
execution_agent: "OpenHands Agent"
---

# [Experiment Title]

**Experiment ID**: PROJECT-EXP-XXX
**Status**: IN_PROGRESS
**Date**: YYYY-MM-DD
**Execution Agent**: OpenHands Agent

---

## Problem Statement

Original problem or hypothesis being tested.

## Hypotheses

| ID | Hypothesis | Status |
|----|------------|--------|
| H1 | [Hypothesis] | TESTING |

## Methodology

1. [Step 1]
2. [Step 2]
3. [Step 3]

## Evidence Collected

### Evidence E1: [Title]

**Type**: Direct | Document | Calculation
**Source**: [Source]

```
[Evidence content]
```

## Findings

### Finding F1: [Title]

**Classification**: [Type]
**Evidence**: E1
**Status**: CONFIRMED | REJECTED | INCONCLUSIVE

Description of finding.

## Validation Status

| Validation | Status | Evidence |
|------------|--------|----------|
| [Validation] | [Status] | [Evidence] |

## Lessons Learned

### Lesson L1: [Title]

Description of lesson learned.

## Recommendations

| Recommendation | Priority | Owner |
|----------------|----------|-------|
| REC-1 | HIGH | Agent |

---

**Experiment Status**: IN_PROGRESS
**Human Review Required**: Yes
```

---

*Per KDE-INV-052*
