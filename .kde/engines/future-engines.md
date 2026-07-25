# Future Engine Extension Guide

**Document Version**: 1.0
**Date**: 2026-07-20

---

## Overview

This guide describes how to create future KDE Engines. The Engine Framework is designed to support unlimited engine versions while maintaining compatibility with existing experiments.

---

## Engine Lineage

```
KDE-ENGINE-001 (Alpha) ──→ KDE-ENGINE-002 (Beta) ──→ KDE-ENGINE-003 (Gamma) ──→ KDE-ENGINE-004 (Delta) ──→ [Future Engines]
     Pattern Discovery         Contextual Knowledge      Causal Discovery        Bootstrap + Context         [Epsilon]
```

### Current Engines

| Engine ID | Version | Codename | Status | Purpose |
|-----------|---------|----------|--------|---------|
| KDE-ENGINE-001 | 0.1.0 | Alpha | Historical | Pattern discovery |
| KDE-ENGINE-002 | 0.1.0 | Beta | Active (Default) | Contextual knowledge |
| KDE-ENGINE-003 | 0.1.0 | Gamma | Active | Causal discovery |
| KDE-ENGINE-004 | 0.1.0 | Delta | Active | Bootstrap + Context |

### Future Potential Engines

| Engine ID | Codename | Expected Features | Status |
|-----------|----------|------------------|--------|
| KDE-ENGINE-005 | Epsilon | Formal Verification | Gap Identified (P3) | |

---

## Engine Creation Checklist

When creating a new engine, complete all items:

- [ ] Create engine directory under `engine/[codename]/`
- [ ] Create `specification.md`
- [ ] Create `methodology.md`
- [ ] Create `pipeline.md` (if pipeline differs)
- [ ] Create `knowledge-model.md` (if model differs)
- [ ] Create `changes.md`
- [ ] Create `provenance.md`
- [ ] Implement standard interface
- [ ] Update `engine/current.md`
- [ ] Update `engine/README.md`
- [ ] Document Alpha vs new engine
- [ ] Update this guide

---

## Directory Structure

```
engine/
├── README.md              # Framework overview
├── interface.md           # Standard interface
├── current.md            # Active engine pointer
├── alpha-vs-beta.md      # Engine comparison
├── future-engines.md     # This guide
│
├── alpha/                # KDE-ENGINE-001
│   ├── specification.md
│   ├── methodology.md
│   ├── changes.md
│   └── provenance.md
│
├── beta/                 # KDE-ENGINE-002
│   ├── specification.md
│   ├── methodology.md
│   ├── pipeline.md
│   ├── knowledge-model.md
│   ├── changes.md
│   └── provenance.md
│
└── [new-engine]/         # KDE-ENGINE-XXX
    ├── specification.md
    ├── methodology.md
    ├── pipeline.md
    ├── knowledge-model.md
    ├── changes.md
    └── provenance.md
```

---

## Required Documents

### 1. specification.md

**Purpose:** Define engine identity and scope.

**Required Sections:**
- Engine identity (ID, version, codename, status)
- Purpose statement
- Scope definition
- Relationship to previous engines
- Engine lifecycle

**Template:**
```markdown
# KDE-ENGINE-XXX Specification

**Engine ID**: KDE-ENGINE-XXX
**Version**: 0.1.0
**Codename**: [Codename]
**Status**: [Active|Historical]
**Effective Date**: YYYY-MM-DD

---

## Engine Identity

[Table with all identity fields]

## Purpose

[What this engine does differently]

## Scope

[What it covers]
[What it doesn't cover]

## Relationship to Previous Engines

[How it relates to KDE-ENGINE-001, 002, etc.]

## Engine Lifecycle

[Status definitions and transitions]
```

### 2. methodology.md

**Purpose:** Document the methodology implemented by the engine.

**Required Sections:**
- Core principles
- Discovery question (what it asks)
- Pipeline stages
- Validation process
- Knowledge assessment

**Template:**
```markdown
# KDE-ENGINE-XXX Methodology

## Core Principles

[List principles]

## The Discovery Question

[What question does this engine ask?]

## Pipeline Processing

[Stage-by-stage methodology]

## Validation Process

[How patterns become knowledge]

## Knowledge Assessment

[How confidence is calculated]
```

### 3. pipeline.md

**Purpose:** Document the processing pipeline (if different from existing engines).

**Required Sections:**
- Pipeline overview diagram
- Module specifications
- Data transformations
- Error handling

**Template:**
```markdown
# KDE-ENGINE-XXX Pipeline

## Pipeline Overview

```
[ASCII diagram]
```

## Module Specifications

### Module 1: [Name]

[Responsibilities]
[Input]
[Output]

[Repeat for each module]
```

### 4. knowledge-model.md

**Purpose:** Document the knowledge object structure (if different).

**Required Sections:**
- Knowledge object fields
- Field specifications
- Complete example
- Validation checklist

### 5. changes.md

**Purpose:** Document version history and changes.

**Required Sections:**
- Version history
- Changes from previous engine
- Breaking changes
- Migration guidance

**Template:**
```markdown
# KDE-ENGINE-XXX Changes

## Version History

### v0.1.0 (YYYY-MM-DD) — Initial Release

#### Changes from [Previous Engine]

[What changed]

#### Breaking Changes

[Any breaking changes]

#### Migration

[How to migrate from previous engine]
```

### 6. provenance.md

**Purpose:** Document engine lineage and history.

**Required Sections:**
- Engine lineage
- Relationship to other engines
- Evolution history
- Architecture decisions

---

## Engine Interface Implementation

All engines MUST implement the standard interface:

```yaml
interface:
  Initialize():         # Setup engine
  Analyze(evidence):   # Process evidence
  Validate(knowledge):  # Verify knowledge
  GenerateKnowledge(): # Create knowledge
  GenerateReport():    # Format output
  Capabilities():      # Return capabilities
  Version():          # Return version
  Metadata():         # Return metadata
```

See: [`interface.md](./interface.md)

---

## Engine Metadata Requirements

Each engine must provide:

```yaml
metadata:
  id: KDE-ENGINE-XXX
  name: "[Full Name]"
  codename: "[Codename]"
  version: "X.Y.Z"
  status: [active|historical|deprecated]
  effective_date: YYYY-MM-DD
  documentation:
    specification: path
    methodology: path
    pipeline: path (if applicable)
    knowledge_model: path (if applicable)
  parent_engine: [parent ID or null]
  child_engine: [child ID or null]
```

---

## Version Numbering

### Semantic Versioning

Engines use semantic versioning (MAJOR.MINOR.PATCH):

| Component | Description | When It Changes |
|-----------|-------------|-----------------|
| **MAJOR** | Breaking changes | Fundamental methodology changes |
| **MINOR** | Additive improvements | New modules or features |
| **PATCH** | Clarifications | Non-substantive improvements |

### Codename Conventions

| Codename | Meaning |
|----------|---------|
| **Alpha** | Initial development |
| **Beta** | Stable, actively developed |
| **Gamma** | Feature complete |
| **Delta** | Additional features |
| **Release** | Production-ready |

### Engine ID Format

```
KDE-ENGINE-[NNN]
```

Where NNN is a sequential number:
- KDE-ENGINE-001 = Alpha
- KDE-ENGINE-002 = Beta
- KDE-ENGINE-003 = Gamma
- etc.

---

## Status Transitions

```
┌─────────────┐
│   Active    │ ← Current engine for new experiments
└──────┬──────┘
       │ (when successor released)
       ▼
┌─────────────┐
│ Historical  │ ← Preserved for reference
└─────────────┘
       │ (when deprecated)
       ▼
┌─────────────┐
│ Deprecated  │ ← Not recommended
└─────────────┘
```

---

## Documenting a New Engine

### Before Creating

1. **Identify the Need**
   - What limitation does the new engine address?
   - What new capability is needed?
   - How does it improve on existing engines?

2. **Define the Scope**
   - What will the engine do?
   - What will it not do?
   - How does it relate to existing engines?

3. **Plan the Pipeline**
   - What modules are needed?
   - How does data flow?
   - What new stages are required?

### During Creation

1. **Create Directory Structure**
2. **Write specification.md**
3. **Write methodology.md**
4. **Write pipeline.md** (if needed)
5. **Write knowledge-model.md** (if needed)
6. **Write changes.md**
7. **Write provenance.md**
8. **Implement Interface**

### After Creation

1. **Update current.md**
   - Add new engine to table
   - Update default if appropriate
   - Update migration history

2. **Update README.md**
   - Add to directory structure
   - Update engine list

3. **Update alpha-vs-beta.md**
   - Add comparison with new engine

4. **Update this guide**
   - Add new engine to lineage
   - Document planned engines

---

## Planned Future Engines

### KDE-ENGINE-005 (Epsilon) — Formal Verification

**Expected Features:**
- Mathematical proof of correctness
- Invariant preservation verification
- Formal boundary validation
- Audit-ready verification

**Status:** Gap Identified (P3) — See [/engines/epsilon/SPEC.md](./epsilon/SPEC.md)

---

## Engine Selection Guide

When creating a new experiment, select the engine that best matches your needs:

| Need | Engine |
|------|--------|
| Pattern discovery only | Alpha (Historical) |
| Context-aware knowledge | Beta (Active, Default) |
| Causal inference | Gamma (Active) |
| Bootstrap enforcement | Delta (Active) |
| Formal verification | Epsilon (Future, P3) |

---

## Common Patterns

### Pattern 1: Extending Beta

To extend Beta with new features:

1. Create new engine (e.g., Gamma)
2. Inherit Beta's pipeline
3. Add new module(s)
4. Document changes

### Pattern 2: Replacing Beta

To create a complete replacement:

1. Create new engine with new ID
2. Document all differences
3. Update migration guide
4. Mark Beta as Historical

### Pattern 3: Specialized Engine

To create a domain-specific engine:

1. Create new engine
2. Extend pipeline with domain modules
3. Add domain-specific context dimensions
4. Document domain applicability

---

## Validation Checklist

Before releasing a new engine:

- [ ] All required documents created
- [ ] specification.md complete
- [ ] methodology.md complete
- [ ] pipeline.md complete (if applicable)
- [ ] knowledge-model.md complete (if applicable)
- [ ] changes.md documents evolution
- [ ] provenance.md documents lineage
- [ ] Interface implemented
- [ ] current.md updated
- [ ] README.md updated
- [ ] Engine comparison documented
- [ ] This guide updated

---

## Support Resources

- [Engine Framework README](./README.md)
- [Engine Interface](./interface.md)
- [Alpha vs Beta](./alpha-vs-beta.md)
- [Beta Specification](./beta/specification.md)
- [Beta Methodology](./beta/methodology.md)
- [Beta Pipeline](./beta/pipeline.md)

---

**Document Status**: APPROVED
**Last Updated**: 2026-07-24 (INV-EVOLUTION-001 REC-007)

**Next Review**: Upon evidence of need for KDE-ENGINE-005 (Epsilon)
