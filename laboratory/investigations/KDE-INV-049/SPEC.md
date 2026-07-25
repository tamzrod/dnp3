# Subject-Based Laboratory Architecture Investigation

**Investigation ID**: KDE-INV-049
**Title**: Subject-Based Laboratory Architecture Investigation
**Status**: COMPLETED
**Date**: 2026-07-25
**Authority**: KDE Runtime (DNP3 Library)
**Bootstrap**: SUCCESS

---

## Executive Summary

**Primary Question**: Should the Runtime ECU determine laboratory ownership before classifying the operation?

**Answer**: **YES** ✅

The hypothesis is validated. Evidence supports implementing a **Subject-First Routing Model** where:
1. Subject (KDE, Project, or Laboratory) is determined first
2. Laboratory is selected based on subject
3. Operation is classified within the selected laboratory

---

## 1. Evidence Summary

### 1.1 Subject Emergence

| Subject | Evidence | Current Location |
|---------|----------|------------------|
| **KDE** | 10 investigations (KDE-INV-*) | Both locations |
| **Project** | 1 investigation (DNP3-INV-*) | laboratory/ |
| **Laboratory** | 7 artifacts (LAB-*) - non-standard | laboratory/ |

### 1.2 Current Architecture Issues

1. **Flat structure**: All artifacts in operation-type directories
2. **Subject in ID only**: Subject encoded in prefix, not path
3. **Naming violations**: LAB-* violates naming convention
4. **Implicit dual labs**: .kde/laboratory/ already exists

---

## 2. Deliverables

### 2.1 Subject Classification Specification

| Subject | Keywords | Laboratory |
|---------|----------|------------|
| **KDE** | runtime, engine, seed, bootstrap, governance, kde | .kde/laboratory/ |
| **Project** | implement, feature, protocol, library, bug | laboratory/ |
| **Laboratory** | assessment, self, lab, architecture | Context-dependent |

### 2.2 Laboratory Ownership Rules

1. **KDE owns** `.kde/laboratory/`
2. **Project owns** `laboratory/`
3. **Laboratory artifacts** go to `.kde/laboratory/` unless project-scoped

### 2.3 Runtime Routing Model (Recommended: Model C - Hybrid)

```
Bootstrap
    ↓
Is Subject = KDE?
    ├─ YES → .kde/laboratory/
    └─ NO  → laboratory/
              ↓
           Classify Operation
              ↓
           Assign ID
              ↓
           Execute
```

### 2.4 Boundary Specification

| Subject | Boundary Criteria |
|---------|-------------------|
| **KDE** | Primary subject is KDE Runtime Framework |
| **Project** | Primary subject is repository content |
| **Laboratory** | Primary subject is laboratory architecture |

### 2.5 Governance Rules

| Violation Type | Response | Severity |
|---------------|----------|----------|
| incorrect_laboratory | MOVE + WARN | Medium |
| cross_contamination | REJECT | High |
| ambiguous_subject | QUARANTINE | High |

### 2.6 Migration Strategy

| Phase | Actions | Risk |
|-------|---------|------|
| 1 | Establish structure | Low |
| 2 | Reclassify LAB-* | Medium |
| 3 | Update routing | Medium |
| 4 | Validate | Low |

### 2.7 Runtime Responsibility Matrix

| Responsibility | Bootstrap | Runtime ECU | Agent |
|---------------|-----------|-------------|-------|
| Subject determination | Init | Primary | Provide description |
| Laboratory selection | Init | Primary | N/A |
| Operation classification | Init | Primary | N/A |
| Artifact placement | Support | Primary | N/A |
| Violation detection | Init | Primary | N/A |

---

## 3. Conclusion

### 3.1 Hypothesis Validated

**The Runtime ECU SHOULD determine laboratory ownership before classifying the operation.**

### 3.2 Recommended Architecture

**Model C (Hybrid)** provides best balance:
- High determinism
- High auditability
- Medium simplicity
- High backward compatibility

### 3.3 Implementation Notes

Per constraints, **DO NOT IMPLEMENT**. Recommended next steps:
1. Update GOV-NAMING-001 to include Laboratory subject
2. Implement subject classifier in Runtime ECU
3. Reclassify LAB-* artifacts
4. Update LAB-GOV-001 lifecycle to include subject classification

---

*Investigation completed by KDE Runtime ECU*
*Evidence-based analysis complete*
*Recommendation: Implement Model C (Subject-First Hybrid)*
