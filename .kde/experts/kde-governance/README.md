# KDE Governance Expert

**Expert ID**: KDE-EXPERT-001  
**Domain**: KDE Runtime Governance  
**Version**: 1.0.0  
**Status**: Active  

---

## Overview

This expert contains domain knowledge for KDE Runtime governance, policies, and procedures.

## Domain Knowledge

### Governance Hierarchy

| Tier | Authority | Purpose |
|------|-----------|---------|
| Tier 1 | Governance Authority (External) | Defines policies |
| Tier 2 | Runtime Authority | Authorizes execution |
| Tier 3 | Execution Authority | Agents/Humans execute |

### Laboratory Rules

1. **Bootstrap-First**: Verify runtime state before investigation
2. **Experiment Entry**: Create entries before investigation
3. **Pre-Existence Check**: Verify issue exists before investigating
4. **Environment Verification**: Verify toolchain before promising tests
5. **Evidence Preservation**: Document all evidence

### Artifact Naming

| Type | Prefix | Directory |
|------|--------|-----------|
| Investigation (KDE) | KDE-INV- | investigations/ |
| Investigation (Project) | PROJECT-INV- | investigations/ |
| Experiment | PROJECT-EXP- | experiments/ |
| Decision | TDR- | decisions/ |

## Rules and Constraints

### Investigation Rules

1. All investigations must use Delta Engine (KDE-ENGINE-004)
2. Bootstrap gates B1, B2, B3 must be verified
3. All findings must trace to evidence
4. Human approval required for significant changes

### Policy Constraints

| Policy | Document | Required |
|--------|----------|----------|
| Dependencies | DEP-001 | Yes |
| Environment | ENV-001 | Yes |
| Naming | NAMING-CONVENTIONS.md | Yes |

## Best Practices

### Investigation Workflow

1. Run bootstrap gates (`gates.py`)
2. Create experiment entry
3. Document evidence
4. Apply Delta Engine pipeline
5. Generate conclusions
6. Request human review

### Policy Development

1. Identify gap or weakness
2. Draft policy document
3. Submit for approval
4. Implement after approval
5. Monitor compliance

## Reference Standards

- KDE-INV-051: Bootstrap Compliance Investigation
- KDE-INV-052: Gap Analysis Investigation
- DEP-001: Runtime Dependencies Policy
- ENV-001: Environment Verification Policy

## Related Artifacts

| Artifact | Purpose |
|----------|---------|
| .kde/governance/ | Governance policies |
| .kde/bootstrap/gates.py | Bootstrap gates |
| laboratory/investigations/ | Investigation artifacts |

---

**Expert Status**: ACTIVE  
**Last Updated**: 2026-07-26
