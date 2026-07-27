# KDE Engine Versioning

**Document Version**: 1.0
**Effective**: 2026-07-20
**Engine Version**: 0.1.0

---

## Overview

This document defines the official KDE Engine Versioning specification for the Knowledge Discovery Engine methodology.

As KDE evolves, the discovery process itself improves through new laboratory rules, evidence requirements, validation processes, governance rules, and reproducibility requirements. This versioning system ensures every experiment permanently records which methodology version produced it.

---

## Purpose

The KDE Engine Versioning system serves three purposes:

1. **Provenance**: Documents which methodology version produced each experiment
2. **Traceability**: Enables understanding of methodology evolution
3. **Immutability**: Preserves historical experiments in their original context

---

## Scope

### What This Covers

- Laboratory rules and procedures
- Evidence collection requirements
- Validation processes
- Governance rules
- Reproducibility requirements
- Documentation standards

### What This Does NOT Cover

- Software implementation changes
- Documentation edits and typo fixes
- New experiments using existing methodology
- New knowledge definitions
- Tool or infrastructure changes

---

## Versioning Philosophy

### Semantic Versioning for Methodology

The KDE Engine uses semantic versioning (MAJOR.MINOR.PATCH):

| Component | Description | When It Changes |
|-----------|-------------|-----------------|
| **MAJOR** | Breaking changes to methodology | Fundamental process changes that invalidate historical comparisons |
| **MINOR** | Additive improvements | New rules, requirements, or processes that enhance methodology |
| **PATCH** | Clarifications and corrections | Non-substantive improvements to documentation |

### Principles

1. **Immutability**: Historical experiments never change their Engine Version
2. **Traceability**: Every experiment links to its producing methodology
3. **Minimal Disruption**: Version changes are infrequent and well-documented
4. **Backward Compatibility**: New versions don't invalidate old experiments

---

## Version Changes

### What Triggers a Version Change

The following changes SHALL increment the version:

| Change Type | Version Increment | Examples |
|-------------|-----------------|----------|
| Fundamental process change | MAJOR | New validation gates, removed steps |
| New methodology requirements | MINOR | New evidence types, reproducibility standards |
| Documentation clarification | PATCH | Clarified rules, fixed ambiguities |

### What Does NOT Trigger a Version Change

The following shall NOT change the Engine Version:

| Change Type | Reason |
|-------------|--------|
| New experiments | Using existing methodology |
| New knowledge definitions | Content, not process |
| Documentation edits | No methodology change |
| Typo fixes | No methodology change |
| Software implementation | Separate from methodology |
| Tool changes | Infrastructure, not process |

---

## Current Engine Version

### Version 0.1.0 — Initial Knowledge Discovery Engine

**Status**: ACTIVE
**Effective Date**: 2026-07-20
**Baseline**: This represents the original KDE methodology from the beginning through LAB-011

**Characteristics**:
- Research methodology v1.0
- Laboratory v2.0 architecture
- 5 Core Principles for AI behavior
- Document state machine (DRAFT → REVIEW → APPROVED → VALIDATED → PROMOTED)
- Scientific learning loop (Research → Knowledge → Laboratory → Evidence → Governance)
- Reproducibility requirements
- Evidence-derived confidence

---

## Historical Policy

### Experiments Are Immutable Historical Records

Once an experiment begins, its Engine Version SHALL NEVER change.

| Policy | Rationale |
|--------|-----------|
| No retroactive changes | Historical experiments are evidence of past methodology |
| New versions don't affect old experiments | Methodology evolution is forward-looking only |
| Conclusions remain valid | Results reflect the methodology used at the time |

### Version Comparison Rules

When comparing experiments with different Engine Versions:

1. Acknowledge methodology differences
2. Consider version age when interpreting results
3. Don't assume newer is always better
4. Historical experiments retain their original value

---

## Experiment Metadata Requirements

### Required Fields

Every experiment SHALL include the following metadata:

```yaml
Engine:
  Version: 0.1.0
  Name: Initial Knowledge Discovery Engine
```

### Placement

The Engine metadata SHALL appear in the experiment metadata section, after the standard experiment fields.

### Example

```markdown
## Metadata

| Field | Value |
|-------|-------|
| Experiment ID | LAB-001 |
| Created | 2026-07-19 |
| Engine Version | 0.1.0 |
| Engine Name | Initial Knowledge Discovery Engine |
```

---

## Migration

### Scope

All experiments from LAB-001 through the current experiment SHALL be updated with Engine Version metadata.

### Process

1. Identify current Engine Version (0.1.0)
2. Add metadata to each experiment
3. Verify metadata placement
4. Document migration in experiment impact files

### Constraints

- Do NOT modify evidence
- Do NOT alter run history
- Do NOT change conclusions
- Do NOT modify timestamps
- Only add provenance metadata

---

## Version Lifecycle

### Version States

| State | Description |
|-------|-------------|
| **ACTIVE** | Current methodology in use |
| **HISTORICAL** | Former methodology, experiments may reference it |
| **DEPRECATED** | Not recommended for new experiments |

### Version Transitions

```
CONCEPTUAL → INITIAL → EVOLVING → ...
     ↓
  0.1.0 (ACTIVE)
```

---

## Governance

### Version Change Authority

Engine Version changes require:

1. Proposal documented with rationale
2. Evidence supporting the change
3. Human approval through governance process
4. Documentation update

### Change Documentation

Each version change SHALL document:

1. What changed
2. Why it changed
3. Impact on future experiments
4. Migration guidance for new experiments

---

## Version History

### v0.1.0 (2026-07-20) — Initial Knowledge Discovery Engine

**Status**: ACTIVE

- Original KDE methodology established
- Research methodology v1.0
- Laboratory v2.0 architecture
- 5 Core Principles for AI behavior
- Document state machine
- Scientific learning loop
- Reproducibility requirements

---

## Compliance

This specification is enforced by:

1. **Repository structure** — Experiments include Engine metadata
2. **Registry** — Engine version tracked in experiment registry
3. **Governance** — Version changes follow governance process
4. **Documentation** — This document defines the standard

---

## References

- [RESEARCH-METHODOLOGY.md](./RESEARCH-METHODOLOGY.md)
- [PRINCIPLES.md](./PRINCIPLES.md)
- [STATE-MACHINE.md](./STATE-MACHINE.md)
- [LABORATORY/README.md](../laboratory/README.md)

---

**Document Status**: APPROVED
**Review Date**: Upon methodology change proposal
