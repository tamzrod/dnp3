# KDE Engine Framework

**Version**: 1.0
**Effective**: 2026-07-20

---

## Overview

The KDE Engine Framework establishes the **sole authoritative source** for laboratory methodology. This framework defines how the Knowledge Discovery Engine methodology evolves over time while preserving complete historical provenance for all experiments.

---

## Engine Authority

**The Engine is the single source of truth for laboratory methodology.**

| Rule | Description |
|------|-------------|
| **Rule 1** | The Laboratory SHALL NOT define its own methodology |
| **Rule 2** | Every laboratory experiment SHALL execute under a KDE Engine |
| **Rule 3** | The Engine directory SHALL be the single authoritative source for laboratory methodology |
| **Rule 4** | Laboratory documents may reference an Engine but SHALL NOT redefine Engine behavior |
| **Rule 5** | Future methodology improvements SHALL be implemented by creating a NEW Engine |
| **Rule 6** | Historical experiments SHALL permanently reference the Engine under which they were executed |
| **Rule 7** | Experiments discover knowledge. Engines discover better methodologies. |

---

## What is a KDE Engine?

A **KDE Engine** is the **sole authoritative source** for laboratory methodology. It defines:

- The processes for knowledge discovery
- The rules for evidence collection
- The validation procedures
- The governance requirements
- The laboratory workflow

Each engine has a unique identity, version, and lifecycle status.

---

## Engine vs Laboratory

| Aspect | Engine | Laboratory |
|--------|--------|-------------|
| **Defines** | Methodology | Experiments |
| **Authority** | Defines process | Executes process |
| **Evolves** | Methodology improvements | New experiments |
| **Responsibility** | How to do experiments | Running experiments |

The Engine defines methodology. The Laboratory executes it.

---

## Why Version Engines?

### Purpose

1. **Provenance Tracking**: Every experiment documents which engine produced it
2. **Methodology Evolution**: Enables understanding of how methodology changed over time
3. **Historical Preservation**: Experiments remain valid in their original methodology context
4. **Traceability**: Links experiments to their producing methodology
5. **Separation of Concerns**: Engines evolve methodology; experiments discover knowledge

### Benefits

- Clear lineage from methodology to experiment to knowledge
- No ambiguity about which rules applied to which experiments
- Foundation for evaluating methodology effectiveness over time
- Enables comparison across methodology versions
- Clear responsibility separation: Engines improve methodology; Laboratories run experiments

---

## Engine Version vs Software Version

| Aspect | Engine Version | Software Version |
|--------|----------------|------------------|
| **Subject** | Methodology | Implementation |
| **Changes when** | Process rules change | Code changes |
| **Affects** | Experiments | Application |
| **Examples** | New validation rules | Bug fixes, features |

### Key Distinction

- **Engine Version**: Governs how experiments are conducted
- **Software Version**: Governs how tools are implemented

A software update does NOT require an engine version change. An engine version change does NOT require a software update.

---

## How Experiments Reference Engines

Every experiment includes engine metadata:

```yaml
Engine:
  ID: KDE-ENGINE-001
  Version: 0.1.0
  Codename: Alpha
```

This metadata appears in the experiment metadata section.

---

## Engine Evolution Philosophy

### Principles

1. **Immutability**: Historical experiments never change their engine reference
2. **Traceability**: Every experiment links to its producing engine
3. **Minimal Disruption**: Engine changes are infrequent and well-documented
4. **Backward Compatibility**: New engines don't invalidate old experiments

### Version Numbering

Engines use semantic versioning (MAJOR.MINOR.PATCH):

| Component | Description | When It Changes |
|-----------|-------------|-----------------|
| **MAJOR** | Breaking changes | Fundamental methodology changes |
| **MINOR** | Additive improvements | New requirements or processes |
| **PATCH** | Clarifications | Non-substantive improvements |

---

## Engine Lifecycle

| State | Description |
|-------|-------------|
| **Active** | Current methodology in use |
| **Historical** | Former methodology, experiments may reference it |
| **Deprecated** | Not recommended for new experiments |

---

## Directory Structure

```
engine/
├── README.md           # This file
├── interface.md       # Engine interface specification
├── current.md         # Points to active engine
├── alpha/             # KDE-ENGINE-001 (Historical)
│   ├── specification.md
│   ├── methodology.md
│   ├── changes.md
│   └── provenance.md
├── beta/              # KDE-ENGINE-002 (Active)
│   ├── specification.md
│   ├── methodology.md
│   ├── pipeline.md
│   ├── knowledge-model.md
│   ├── changes.md
│   └── provenance.md
├── gamma/             # KDE-ENGINE-003 (Experimental)
│   ├── README.md
│   ├── specification.md
│   ├── methodology.md
│   ├── pipeline.md
│   ├── knowledge-model.md
│   ├── changes.md
│   └── provenance.md
└── [future engines]/
```

---

## Related Documents

- [governance/ENGINE-VERSIONING.md](../governance/ENGINE-VERSIONING.md) — Engine versioning specification
- [governance/PRINCIPLES.md](../governance/PRINCIPLES.md) — Core KDE principles
- [laboratory/RESEARCH-METHODOLOGY.md](../laboratory/RESEARCH-METHODOLOGY.md) — Research methodology
- [laboratory/README.md](../laboratory/README.md) — Laboratory process

---

## Governance

This framework is governed by:

1. **Engine Versioning Specification** — Defines versioning rules
2. **Engine Specification** — Defines each engine's scope
3. **Experiment Registry** — Tracks engine assignments

Changes to this framework follow the governance process defined in `/governance`.

---

**Document Status**: APPROVED
