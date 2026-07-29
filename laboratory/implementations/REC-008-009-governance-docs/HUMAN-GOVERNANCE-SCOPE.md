# Human Governance Scope

**Engine**: KDE-ENGINE-001 (v1.0.0)  
**Seed**: SEED-001 (v1.0.0)  
**Generated**: 2026-07-29T10:25:00Z

---

## Overview

This document defines the scope of human governance in the KDE Laboratory. Human oversight is a core design principle, but its scope has evolved over time.

## Governance Layers

```
┌─────────────────────────────────────────────────────────────┐
│                    HUMAN GOVERNANCE                          │
├─────────────────────────────────────────────────────────────┤
│  Layer 1: File Boundary Violations (ACTIVE)                 │
│  └── Human approval required for non-laboratory writes       │
├─────────────────────────────────────────────────────────────┤
│  Layer 2: Investigation Workflow (PARTIAL)                  │
│  └── Human catches violations retroactively                  │
│  └── REC-001 to REC-007 add proactive enforcement          │
├─────────────────────────────────────────────────────────────┤
│  Layer 3: Experiment Creation (PARTIAL)                     │
│  └── Synthesis required (REC-003)                           │
│  └── Run documentation required (REC-004/005)               │
├─────────────────────────────────────────────────────────────┤
│  Layer 4: Knowledge Promotion (ACTIVE)                       │
│  └── Human sets PROMOTED status                             │
│  └── AI can only set DRAFT/REVIEW                          │
└─────────────────────────────────────────────────────────────┘
```

## Human Approval Requirements

| Action | Human Required | Evidence |
|--------|---------------|----------|
| File write outside /laboratory/ | ✅ YES | FileBoundaryGuard |
| Knowledge promotion to VALIDATED | ✅ YES | State machine |
| Override investigation gate | ✅ YES | Human authorization |
| Experiment creation (after synthesis) | ❌ NO | Automatic |
| Investigation artifact creation | ❌ NO | Automatic |
| Run documentation | ❌ NO | Automatic |

## AI-Only Actions (Self-Service)

The following actions do NOT require human approval:

| Action | Authority | Evidence |
|--------|-----------|----------|
| Create investigation artifact | AI | Laboratory rules |
| Create experiment artifact | AI | Laboratory rules |
| Create run documentation | AI | REC-004 |
| Set investigation status to IN_PROGRESS | AI | State machine |
| Set experiment status to IN_PROGRESS | AI | State machine |
| Set knowledge status to DRAFT | AI | State machine |
| Set knowledge status to REVIEW | AI | State machine |

## AI-Prohibited Actions

The following actions are PROHIBITED for AI:

| Action | Prohibition | Evidence |
|--------|-------------|----------|
| AI cannot approve violations | SEED-001 Principle | ViolationHandler |
| AI cannot self-approve | SEED-001 Principle | ViolationHandler |
| AI cannot promote to VALIDATED | State machine | knowledge/ |
| AI cannot modify SEED files | File boundary | FileBoundaryGuard |
| AI cannot override governance | Architecture C | Rules |

## Governance Evolution

### Before REC-001 to REC-007

```
Human oversight: Retroactive
- Human catches violations after they occur
- Human intervenes to redirect non-compliant behavior
- Pattern: Bypass → Human catch → Redirect
```

### After REC-001 to REC-007

```
Human oversight: Proactive + Proportional
- Runtime enforces investigation artifacts first
- Synthesis required before new experiments
- Run documentation required for closure
- Human approval only for:
  1. File boundary violations
  2. Knowledge promotion
  3. Override of automated gates
```

## Implementation Coverage

| REC | Governance Type | Status |
|-----|----------------|--------|
| REC-001 | InvestigationArtifactGuard | ✅ IMPLEMENTED |
| REC-002 | PreToolCheck | ✅ IMPLEMENTED |
| REC-003 | SynthesisCheckpoint | ✅ IMPLEMENTED |
| REC-004/005 | ExperimentDocsGate | ✅ IMPLEMENTED |
| REC-006/007 | SkillEnforcement | ✅ IMPLEMENTED |

## Recommendations

1. **Extend human approval scope** when:
   - Automated gates miss critical violations
   - Novel scenarios not covered by existing rules

2. **Reduce human approval scope** when:
   - Automated gates prove reliable
   - Process matures and violation rate decreases

3. **Document new governance patterns** when:
   - New enforcement mechanisms are added
   - Scope boundaries change

## References

- SEED-001: Five Core Principles
- FileBoundaryGuard: `.kde/runtime/file_boundary_guard.py`
- InvestigationArtifactGuard: `.kde/runtime/investigation_guard.py`
- ViolationHandler: `.kde/runtime/violation_handler.py`
- State Machine: `laboratory/state-machine.md`

---

*Documented by REC-008 implementation*
*From KDE-INV-GOV-SYN recommendation REC-008*
