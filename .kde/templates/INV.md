# Investigation Template

**File**: INV.md
**Version**: 1.0.0
**Date**: 2026-07-26
**Source**: KDE-INV-052

---

## Purpose

This template provides the standard format for Investigations.

## Usage

Copy this template to create a new Investigation:

```bash
mkdir -p laboratory/investigations/KDE-INV-XXX
cp .kde/templates/INV.md laboratory/investigations/KDE-INV-XXX/README.md
cp .kde/templates/INV-SPEC.md laboratory/investigations/KDE-INV-XXX/SPEC.md
cp .kde/templates/INV-CONCLUSION.md laboratory/investigations/KDE-INV-XXX/CONCLUSION.md
```

## Template

### README.md

```markdown
---
id: KDE-INV-XXX
type: investigation
title: "[Investigation Title]"
authority: "KDE Runtime (DNP3 Library)"
status: IN_PROGRESS
created: "YYYY-MM-DD"
execution_agent: "OpenHands Agent"
engine: KDE-ENGINE-004 (Delta)
---

# [Investigation Title]

**Investigation ID**: KDE-INV-XXX
**Engine**: KDE-ENGINE-004 (Delta)
**Title**: [Investigation Title]
**Status**: IN_PROGRESS
**Date**: YYYY-MM-DD
**Authority**: KDE Runtime (DNP3 Library)

---

## Executive Summary

Brief summary of the investigation and key findings.

---

## Research Questions

| ID | Question | Finding |
|----|----------|---------|
| RQ1 | [Question] | [Finding] |

---

## Evidence

### Evidence E1: [Title]

**Type**: Direct | Document | Calculation
**Source**: [Source file or location]
**Relevance**: [Why this evidence matters]

```
[Evidence content or quote]
```

---

## Findings

### Finding F1: [Title]

**Classification**: [Type]
**Evidence**: E1, E2
**Confidence**: HIGH | MEDIUM | LOW

Description of the finding.

---

## Recommendations

| Recommendation | Priority | Owner |
|----------------|----------|-------|
| REC-1 | HIGH | Agent |

---

## Related Artifacts

| Artifact | Type | Relationship |
|----------|------|--------------|
| KDE-INV-XXX | Investigation | Parent |

---

**Investigation Status**: IN_PROGRESS
**Human Review Required**: Yes
```

### SPEC.md

```markdown
# Investigation Specification: KDE-INV-XXX

**Investigation ID**: KDE-INV-XXX
**Title**: [Investigation Title]
**Engine**: KDE-ENGINE-004 (Delta)
**Status**: IN_PROGRESS

---

## Investigation Scope

### In Scope
- [Item 1]
- [Item 2]

### Out of Scope
- [Item 1]

---

## Objectives

| ID | Objective | Status |
|----|-----------|--------|
| O1 | [Objective] | IN_PROGRESS |

---

## Evidence Sources

| Source | Type | Relevance |
|--------|------|-----------|
| [Source] | [Type] | [Relevance] |

---

## Methodology

This investigation applies the Delta Engine (KDE-ENGINE-004) pipeline.

---

## Success Criteria

| Criterion | Evidence | Status |
|-----------|----------|--------|
| [Criterion] | [Evidence] | [Status] |

---

**Spec Status**: IN_PROGRESS
**Created**: YYYY-MM-DD
**Engine**: KDE-ENGINE-004 (Delta)
```

### CONCLUSION.md

```markdown
# Investigation Conclusion: KDE-INV-XXX

**Investigation ID**: KDE-INV-XXX
**Engine**: KDE-ENGINE-004 (Delta)
**Status**: COMPLETED
**Date**: YYYY-MM-DD

---

## Summary

Summary of the investigation.

---

## Key Findings

### Finding 1: [Title]

**Classification**: [Type]
**Evidence**: [Evidence IDs]
**Confidence**: HIGH | MEDIUM | LOW

Description of finding.

---

## Recommendations

### REC-1: [Title]

**Priority**: HIGH | MEDIUM | LOW

Description of recommendation.

---

## Impact Assessment

| Aspect | Impact | Details |
|--------|--------|---------|
| [Aspect] | [Impact] | [Details] |

---

## Investigation Quality Assessment

| Criterion | Score | Evidence |
|-----------|-------|----------|
| Evidence Collection | X/10 | [Evidence] |
| Observation Extraction | X/10 | [Evidence] |

**Overall**: X.X/10

---

## Next Steps

| Step | Action | Owner |
|------|--------|-------|
| 1 | [Action] | [Owner] |

---

**Conclusion Status**: READY FOR REVIEW
**Human Approval Required**: Yes
```

---

*Per KDE-INV-052*
