# KDE Unified State Vocabulary

**Document ID**: KDE-GOVERNANCE-STATE-001
**Version**: 1.0.0
**Status**: APPROVED
**Authority**: INV-AUDIT-REVIEW-001 (Priority 6)
**Effective Date**: 2026-07-27
**Source**: State Machine Standardization

---

## Purpose

KDE uses multiple state machines across different subsystems. Each subsystem developed independently, leading to inconsistent terminology.

This document establishes a **unified state vocabulary** that standardizes terminology across all KDE subsystems. The goal is clarity, not uniformity—subsystems may have different states, but the same terms should mean the same thing everywhere.

---

## The Problem

| Subsystem | Active State | Complete State | Frozen State |
|-----------|-------------|----------------|--------------|
| **Investigation** | ACTIVE | COMPLETE | — |
| **Knowledge** | — | PROMOTED | VALIDATED |
| **Expert** | ACTIVE | REGISTERED | SYNTHESIZED |
| **Document** | DRAFT | APPROVED | VALIDATED |
| **Seed** | ACTIVE | — | FROZEN |

**Issues**:
- "ACTIVE" means different things in different contexts
- "VALIDATED" has different meanings
- No consistent "terminal state" terminology

---

## Unified State Vocabulary

### Core States

These states have consistent meaning across all subsystems:

| State | Meaning | Terminal? | AI Can Set? |
|-------|---------|----------|--------------|
| **DRAFT** | Work in progress | No | Yes |
| **IN_REVIEW** | Submitted for review | No | Yes |
| **ACTIVE** | In production use | No | No (Human only) |
| **COMPLETE** | Work finished successfully | Yes | Yes |
| **REJECTED** | Work rejected | Yes | No (Human only) |
| **DEPRECATED** | Superseded or obsolete | Yes | No (Human only) |

### Domain-Specific States

These states have specialized meanings within their domains:

| State | Domain | Meaning | Terminal? |
|-------|--------|---------|-----------|
| **VALIDATED** | Knowledge | Passed validation tests | No (leads to PROMOTED) |
| **PROMOTED** | Knowledge | Moved to /knowledge/ | Yes |
| **REGISTERED** | Expert | Added to expert registry | No (leads to ACTIVE) |
| **SYNTHESIZED** | Expert | Initial synthesis complete | No (leads to CANDIDATE) |
| **CANDIDATE** | Expert/Knowledge | Ready for validation | No |
| **FROZEN** | Seed | Immutable, production-ready | Yes |

---

## State Mapping by Subsystem

### Investigation States

| Current State | Unified State | Notes |
|--------------|---------------|-------|
| DRAFT | DRAFT | Work in progress |
| PROPOSED | IN_REVIEW | Submitted for approval |
| ACTIVE | ACTIVE | Investigation in progress |
| COMPLETE | COMPLETE | Investigation finished |
| PROMOTED | DEPRECATED | Superseded |

### Knowledge States

| Current State | Unified State | Notes |
|--------------|---------------|-------|
| DRAFT | DRAFT | Work in progress |
| CANDIDATE | IN_REVIEW | Ready for validation |
| VALIDATED | VALIDATED | Domain-specific state |
| PROMOTED | PROMOTED | Domain-specific terminal state |
| DEPRECATED | DEPRECATED | Superseded |

### Expert States

| Current State | Unified State | Notes |
|--------------|---------------|-------|
| SYNTHESIZED | SYNTHESIZED | Domain-specific initial state |
| CANDIDATE | IN_REVIEW | Ready for validation |
| VALIDATED | VALIDATED | Domain-specific state |
| REGISTERED | REGISTERED | Domain-specific state |
| ACTIVE | ACTIVE | In production use |

### Seed States

| Current State | Unified State | Notes |
|--------------|---------------|-------|
| PROPOSED | IN_REVIEW | Under consideration |
| ACTIVE | ACTIVE | In production use |
| FROZEN | FROZEN | Immutable |

### Document States

| Current State | Unified State | Notes |
|--------------|---------------|-------|
| DRAFT | DRAFT | Work in progress |
| REVIEW | IN_REVIEW | Submitted for review |
| APPROVED | ACTIVE | Human approved |
| REVISION_REQUIRED | DRAFT | Returned for changes |
| VALIDATED | VALIDATED | Domain-specific state |
| PROMOTED | PROMOTED | Domain-specific state |
| REJECTED | REJECTED | Rejected |

---

## Transition Rules

### Universal Rules

1. **No Skip**: States must be traversed in order (no jumping from DRAFT to COMPLETE)
2. **No Backward Skip**: Cannot skip backward (no COMPLETE to DRAFT)
3. **Terminal States**: Once COMPLETE, REJECTED, or DEPRECATED, state is final
4. **Human Authority**: ACTIVE, REJECTED, and DEPRECATED require human action

### State Transition Matrix

| From | To | Authority | Evidence Required |
|------|-----|-----------|------------------|
| (new) | DRAFT | Any | None |
| DRAFT | IN_REVIEW | Author | Work complete |
| IN_REVIEW | ACTIVE | Human | Approval granted |
| IN_REVIEW | DRAFT | Human | Revision requested |
| IN_REVIEW | REJECTED | Human | Rejection reason |
| ACTIVE | COMPLETE | Author | Work finished |
| ACTIVE | DEPRECATED | Human | Deprecation reason |
| COMPLETE | DEPRECATED | Human | Deprecation reason |

### Domain-Specific Transitions

#### Knowledge

| From | To | Authority | Evidence Required |
|------|-----|-----------|------------------|
| DRAFT | IN_REVIEW | Author | Document complete |
| IN_REVIEW | VALIDATED | Validator | Validation passed |
| VALIDATED | PROMOTED | Human | Promotion approved |

#### Expert

| From | To | Authority | Evidence Required |
|------|-----|-----------|------------------|
| SYNTHESIZED | IN_REVIEW | Expert author | Synthesis complete |
| IN_REVIEW | VALIDATED | Validator | Validation passed |
| VALIDATED | REGISTERED | Governance | Registration approved |
| REGISTERED | ACTIVE | Governance | Activation approved |

#### Seed

| From | To | Authority | Evidence Required |
|------|-----|-----------|------------------|
| PROPOSED | ACTIVE | Human | Approval granted |
| ACTIVE | FROZEN | Human | Freeze approved |

---

## Naming Conventions

### State Field Names

Use consistent field names across all documents:

| Use | Not |
|-----|-----|
| `**Status**` | `**State**`, `**Document Status**`, `**Investigation Status**` |

### State Values

Use consistent formatting:

| Use | Not |
|-----|-----|
| `ACTIVE` | `Active`, `active`, `IN PROGRESS` |
| `IN_REVIEW` | `IN_REVIEW`, `In_Review`, `IN REVIEW` |
| `COMPLETE` | `COMPLETE`, `Complete`, `completed` |

---

## Migration Guide

### For New Documents

Use the unified vocabulary in all new documents.

### For Existing Documents

No immediate migration required. This is a vocabulary standard, not a mandate for retroactive changes. Update documents during their next revision cycle.

### Checklist

- [ ] Uses `**Status**` field name consistently
- [ ] Uses uppercase state values
- [ ] Follows transition rules
- [ ] Documents state transitions in commit messages

---

## Enforcement

This vocabulary is enforced through:

1. **Template Updates**: All templates updated to use unified vocabulary
2. **Documentation**: This document establishes the standard
3. **Review**: Governance reviews ensure compliance

---

## References

| Document | Relationship |
|----------|--------------|
| `governance/STATE-MACHINE.md` | Document state machine (updated to match) |
| `knowledge/KDE-KNOWLEDGE-LIFECYCLE.md` | Knowledge lifecycle |
| `experts/_lifecycle.md` | Expert lifecycle |
| `laboratory/templates/investigation-template.md` | Investigation template |

---

## Version History

| Version | Date | Changes | Authority |
|---------|------|---------|-----------|
| 1.0.0 | 2026-07-27 | Initial unified vocabulary | INV-AUDIT-REVIEW-001 |

---

**Document Status**: APPROVED
**Authority**: INV-AUDIT-REVIEW-001
**Compliance**: RECOMMENDED (vocabulary standard)
