# Evolution: How KDE Changes

**Seed ID**: SEED-002
**Section**: Evolution (NEW)
**Addresses**: LESSON-003, LESSON-004, LESSON-006

---

## Purpose

Evolution defines how KDE changes while preserving lineage and enabling reproducibility.

**Problems Addressed**: 
- LESSON-003: "No migration-first approach"
- LESSON-004: "Reasoning not versioned"
- LESSON-006: "Evolution overwrote previous architecture"

---

## Migration-First Principle

> "Every structural change follows a documented migration process."

---

## Migration Requirements

### When Migration Document Required

| Change Type | Requires MIGRATION-XXX.md |
|------------|---------------------------|
| Structural reorganization | YES |
| New component creation | YES |
| Boundary changes | YES |
| Seed creation | YES |
| Engine version change | YES |
| Laboratory structure change | YES |

### When Migration Document NOT Required

| Change Type | Requires Document |
|------------|-------------------|
| Bug fixes | NO |
| Documentation improvements | NO |
| Knowledge content updates | NO |
| Experiment runs | NO |

---

## Migration Document Structure

```markdown
# Migration-XXX: [Title]

## Overview
What is changing and why.

## Changes
What specifically is changing.

## Compatibility
What works with what.

## Rollback
How to revert if needed.

## Validation
How to verify success.
```

---

## Seed Evolution Process

### Creating New Seed from Lessons

```
1. Document lessons from previous Seed
         │
         ▼
2. Create design objectives
         │
         ▼
3. Implement only what lessons require
         │
         ▼
4. Document parent-child relationship
         │
         ▼
5. Freeze new Seed
```

### What Belongs in Migration Document

| Section | Content |
|---------|---------|
| Overview | What changed, why |
| Changes | Specific modifications |
| Compatibility | What still works |
| Rollback | How to revert |
| Validation | How to verify |

---

## Lineage Documentation

### Parent-Child Relationship

Every Seed must document:

| Element | Required |
|---------|----------|
| Parent Seed ID | YES |
| What was inherited | YES |
| What changed | YES |
| Why it changed | YES |
| What compatibility preserved | YES |

### Example

```markdown
## Lineage

### From SEED-001 (Genesis)

**Inherited**:
- 5 Core Principles (still valid)
- Scientific Learning Loop (sound architecture)
- Evidence Model (working as intended)

**Changed**:
- Boundary definitions (LESSON-002)
- Migration guidelines (LESSON-003)
- Lineage documentation (LESSON-004)

**Why Changed**:
- LESSON-002: Boundaries became blurred
- LESSON-003: No migration-first approach
- LESSON-004: Reasoning not versioned

**Preserved Compatibility**:
- All engines compatible with SEED-002
- All experiments remain valid
```

---

## Compatibility Matrix

### Seed Compatibility

| Seed | Engines Compatible | Experiments Valid |
|------|-------------------|-------------------|
| SEED-001 | KDE-001, KDE-002, KDE-003 | LAB-001 to LAB-019 |
| SEED-002 | KDE-001, KDE-002, KDE-003 | All future |

### Engine Compatibility

| Engine | Compatible Seeds |
|--------|-----------------|
| KDE-ENGINE-001 | SEED-001, SEED-002 |
| KDE-ENGINE-002 | SEED-001, SEED-002 |
| KDE-ENGINE-003 | SEED-001, SEED-002 |

---

## Rollback Path

Every migration must document rollback:

```markdown
## Rollback

To revert this migration:

1. [Step 1]
2. [Step 2]
3. [Step 3]

Verification: [How to confirm rollback]
```

---

## Changes from Seed-001

| Aspect | Seed-001 | Seed-002 |
|--------|----------|----------|
| Migration Process | Ad-hoc | **Standardized** |
| Lineage Documentation | Minimal | **Comprehensive** |
| Parent Reference | None | **Required** |
| Rollback Path | Not documented | **Required** |

---

## Lessons Addressed

- **LESSON-003**: "No migration-first approach"
- **LESSON-004**: "Reasoning not versioned"
- **LESSON-006**: "Evolution overwrote previous architecture"

---

**Status**: NEW IN SEED-002
**Evidence**: LESSON-003, LESSON-004, LESSON-006
