# KDE Dependency Tracking System

**Document ID**: KDE-GOVERNANCE-DEP-001
**Version**: 1.0.0
**Status**: APPROVED
**Authority**: INV-AUDIT-REVIEW-001 (Priority 8)
**Effective Date**: 2026-07-27
**Source**: Dependency Tracking Implementation

---

## Purpose

This document establishes a **dependency tracking system** for KDE. As identified in INV-AUDIT-DELTA-001, understanding dependencies is essential for predicting the impact of changes:

> "Understanding dependencies is essential for predicting the impact of changes."

This system enables:
- Impact analysis for proposed changes
- Understanding which artifacts depend on others
- Identifying critical paths and single points of failure
- Planning changes with full awareness of consequences

---

## Dependency Types

### Type 1: Engine → Seed Dependency

**Definition**: An engine requires a specific seed for operation.

```
Engines depend on Seeds:
- Beta (KDE-ENGINE-002) → SEED-001 (Genesis)
- Gamma (KDE-ENGINE-003) → SEED-001 (Genesis)
- Delta (KDE-ENGINE-004) → SEED-001 (Genesis)
```

**Impact**: Changing a seed affects all engines that depend on it.

### Type 2: Investigation → Knowledge Dependency

**Definition**: An investigation produces or references knowledge.

```
Investigations produce Knowledge:
- INV-AUDIT-001 → Knowledge Provenance
- INV-AUDIT-REVIEW-001 → Investigation Versioning
```

**Impact**: Deprecating knowledge may invalidate investigations.

### Type 3: Knowledge → Knowledge Dependency

**Definition**: A knowledge document builds on other knowledge.

```
Knowledge chains:
- KDE-001 (What is Knowledge) → KDE-002 (What is Evidence)
- KDE-001 → KDE-003 (What is Ambiguity)
```

**Impact**: Changing foundational knowledge affects dependent knowledge.

### Type 4: Governance → Component Dependency

**Definition**: Governance rules apply to specific components.

```
Governance rules:
- LABORATORY-SOP.md → All investigations
- SOP-COMPLEXITY-BUDGET.md → All SOPs
```

**Impact**: Changing governance affects all governed components.

### Type 5: Template → Schema Dependency

**Definition**: Templates enforce specific schemas.

```
Template chains:
- investigation-template.md → investigation.md
- knowledge-template.md → knowledge documents
```

**Impact**: Template changes affect all documents created from them.

---

## Dependency Registry

### Engine Dependencies

| Engine | Version | Depends On | Type |
|--------|---------|-----------|------|
| KDE-ENGINE-002 (Beta) | 0.1.0 | SEED-001 v1.0.0 | Engine→Seed |
| KDE-ENGINE-003 (Gamma) | 0.1.0 | SEED-001 v1.0.0 | Engine→Seed |
| KDE-ENGINE-004 (Delta) | 0.1.0 | SEED-001 v1.0.0 | Engine→Seed |

### Knowledge Dependencies

| Knowledge | Version | Dependencies |
|-----------|---------|--------------|
| KDE-001 | 1.0.0 | None (foundational) |
| KDE-002 | 1.0.0 | KDE-001 |
| KDE-003 | 1.0.0 | KDE-001 |
| KDE-KNOWLEDGE-LIFECYCLE | 1.0.0 | KDE-001, KDE-002 |
| KDE-KNOWLEDGE-TEMPLATES | 2.0.0 | KDE-KNOWLEDGE-LIFECYCLE |

### Governance Dependencies

| Governance | Governs |
|------------|---------|
| LABORATORY-SOP.md | All investigations, experiments |
| SOP-COMPLEXITY-BUDGET.md | All SOPs |
| KDE-GOVERNANCE-STATE-VOCABULARY.md | All documents with Status field |
| ARTIFACT-PROTECTION.md | All protected artifacts |

---

## Impact Analysis

### Before Making Changes

Before changing any artifact, complete this impact analysis:

```markdown
## Impact Analysis: [Change Description]

### Change Target
- Artifact: [Name]
- Current Version: [X.Y.Z]
- Proposed Change: [Description]

### Dependencies Identified
| Artifact | Dependency Type | Impact |
|---------|----------------|--------|
| [Name] | [Type] | [Impact description] |

### Impact Assessment
- **Breaking Changes**: [Yes/No]
- **Affected Artifacts**: [Count]
- **Rollback Complexity**: [Low/Medium/High]

### Mitigation Plan
1. [Step 1]
2. [Step 2]

### Approval Required
- [ ] Breaking changes require human approval
- [ ] Non-breaking changes require governance review
```

---

## Dependency Detection

### Automated Detection

Use `.kde/scripts/dependency-detector.py` to detect dependencies:

```bash
# Detect all dependencies
python .kde/scripts/dependency-detector.py

# Check specific artifact
python .kde/scripts/dependency-detector.py --artifact KDE-KNOWLEDGE-TEMPLATES

# Export dependency graph
python .kde/scripts/dependency-detector.py --export dependencies.json
```

### Manual Detection Checklist

- [ ] Search for references to artifact ID in all documents
- [ ] Check knowledge dependencies section
- [ ] Check governance applicability
- [ ] Check template usage
- [ ] Review git history for related changes

---

## Critical Path Analysis

### What Is a Critical Path?

A **critical path** is a chain of dependencies where changing one artifact requires changing all downstream artifacts.

### KDE Critical Paths

```
SEED-001 (Genesis)
    ↓
All Engines
    ↓
All Investigations
    ↓
Knowledge Production
    ↓
Governance (if knowledge affects governance)
```

**Risk**: SEED-001 is a single point of failure. Changes would cascade.

### Mitigation

1. **No Direct SEED-001 Changes**: SEED-001 is immutable
2. **Version Stamping**: Track which version was used
3. **Progressive Rollout**: Test changes on non-critical paths first

---

## Dependency Update Process

### When Dependencies Change

1. **Identify**: Use dependency detector to find all dependencies
2. **Assess**: Evaluate impact of change
3. **Notify**: Inform all dependent artifact owners
4. **Update**: Update dependent artifacts
5. **Verify**: Confirm all dependencies are satisfied
6. **Document**: Record dependency update in change log

### Dependency Update Checklist

- [ ] Identified all dependencies
- [ ] Assessed impact
- [ ] Notified affected parties
- [ ] Updated dependent artifacts
- [ ] Verified all dependencies satisfied
- [ ] Documented in change log

---

## Version Compatibility Matrix

### Engine ↔ Seed Compatibility

| Engine | SEED-001 v1.0.0 | SEED-002 v1.0.0 | SEED-003 v1.0.0 |
|--------|-----------------|-----------------|-----------------|
| KDE-ENGINE-002 v0.1.0 | ✅ Compatible | ⚠️ Partial | ⚠️ Partial |
| KDE-ENGINE-003 v0.1.0 | ✅ Compatible | ⚠️ Partial | ⚠️ Partial |
| KDE-ENGINE-004 v0.1.0 | ✅ Compatible | ⚠️ Partial | ⚠️ Partial |

### Knowledge ↔ Knowledge Compatibility

| Knowledge | Dependencies | Status |
|-----------|--------------|--------|
| KDE-KNOWLEDGE-LIFECYCLE | None | ✅ Stable |
| KDE-KNOWLEDGE-TEMPLATES v2.0.0 | KDE-KNOWLEDGE-LIFECYCLE | ✅ Compatible |

---

## Dependency Graph Visualization

### KDE Dependency Map

```
                    ┌─────────────────────────────────────┐
                    │           SEED-001 (Genesis)        │
                    │              (Immutable)             │
                    └───────────────────┬───────────────────┘
                                        │
                                        ▼
                    ┌─────────────────────────────────────┐
                    │         SEED-002 (Evolution)        │
                    └───────────────────┬───────────────────┘
                                        │
                                        ▼
                    ┌─────────────────────────────────────┐
                    │         SEED-003 (Bootstrap)        │
                    └───────────────────┬───────────────────┘
                                        │
        ┌───────────────────────────────┼───────────────────────────────┐
        │                               │                               │
        ▼                               ▼                               ▼
┌───────────────┐             ┌───────────────┐             ┌───────────────┐
│ Engine Beta   │             │ Engine Gamma  │             │ Engine Delta  │
│ (v0.1.0)      │             │ (v0.1.0)      │             │ (v0.1.0)      │
└───────┬───────┘             └───────┬───────┘             └───────┬───────┘
        │                               │                               │
        └───────────────────────────────┼───────────────────────────────┘
                                        │
                                        ▼
                    ┌─────────────────────────────────────┐
                    │           Investigations            │
                    │  (INV-001, INV-002, ..., INV-AUDIT) │
                    └───────────────────┬───────────────────┘
                                        │
                                        ▼
                    ┌─────────────────────────────────────┐
                    │             Knowledge                │
                    │    (KDE-001, KDE-002, ..., KDE-XX) │
                    └───────────────────┬───────────────────┘
                                        │
                                        ▼
                    ┌─────────────────────────────────────┐
                    │            Governance               │
                    │   (LABORATORY-SOP, SOP-*, ARCHIVE)  │
                    └─────────────────────────────────────┘
```

---

## Enforcement

### When Creating New Artifacts

Document dependencies in the artifact header:

```markdown
**Dependencies**:
  - [KDE-XXX: Description]
  - [INV-XXX: Description]
```

### When Modifying Existing Artifacts

1. Run dependency detector
2. Assess impact on all dependents
3. Update dependent artifacts if needed
4. Document changes in provenance section

---

## References

| Document | Relationship |
|----------|--------------|
| `.kde/scripts/dependency-detector.py` | Dependency detection script |
| `governance/KDE-GOVERNANCE-STATE-VOCABULARY.md` | Related governance |
| `laboratory/templates/investigation-template.md` | Investigation dependencies |

---

## Version History

| Version | Date | Changes | Authority |
|---------|------|---------|-----------|
| 1.0.0 | 2026-07-27 | Initial dependency tracking | INV-AUDIT-REVIEW-001 |

---

**Document Status**: APPROVED
**Authority**: INV-AUDIT-REVIEW-001
**Compliance**: RECOMMENDED
