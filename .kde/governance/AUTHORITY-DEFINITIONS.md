# KDE Authority Definitions

**Document ID**: GOV-AUTHORITY-001
**Version**: 1.0.0
**Date**: 2026-07-25
**Status**: APPROVED
**Authority**: KDE Governance (per KDE-INV-002)

---

## Purpose

This document provides detailed definitions for each authority type in the KDE Governance Authority Hierarchy, as established by GOV-HIERARCHY-001.

---

## Authority Overview

| Authority Type | Tier | Source | Primary Role |
|---------------|------|--------|--------------|
| Governance Authority | 1 | External (Framework) | Defines methodology |
| Runtime Authority | 2 | Runtime Instance | Executes governance |
| Execution Authority | 3 | Agents/Humans | Performs work |
| Approval Authority | - | Humans | Provides oversight |

---

## 1. Governance Authority

### Definition

The external entity (human or organization) that defines KDE methodology and governance policies.

### Characteristics

| Attribute | Description |
|-----------|-------------|
| **Type** | External (outside any specific project) |
| **Scope** | KDE Runtime Framework and methodology |
| **Composition** | Human or organizational entity |
| **Tenure** | Long-term, stable |

### Responsibilities

1. Define KDE methodology and principles
2. Create governance policies
3. Establish the KDE Runtime Framework
4. Review and approve governance changes
5. Maintain KDE standards

### Boundaries

- **CAN**: Define, create, review, approve
- **CANNOT**: Execute day-to-day work, modify runtime state

### Relationship Diagram

```
Governance Authority
    │
    │ Creates
    ▼
KDE Runtime Framework
    │
    │ Instantiates
    ▼
KDE Runtime Instance
```

---

## 2. Runtime Authority

### Definition

The KDE Runtime instance that executes governance for a project.

### Characteristics

| Attribute | Description |
|-----------|-------------|
| **Type** | Runtime instance (per project) |
| **Scope** | Single project or repository |
| **Composition** | Runtime framework + configuration |
| **Tenure** | Project lifetime |

### Responsibilities

1. Execute governance policies
2. Authorize execution agents
3. Maintain runtime state
4. Own project artifacts
5. Track investigations and decisions

### Boundaries

- **CAN**: Execute, authorize, own artifacts, track state
- **CANNOT**: Modify governance framework, define methodology

### Relationship Diagram

```
Governance Authority
    │
    │ Defines framework
    ▼
KDE Runtime Framework
    │
    │ Instantiates
    ▼
Runtime Authority (Instance)
    │
    │ Authorizes
    ▼
Execution Agents
```

---

## 3. Execution Authority

### Definition

The agents and humans who perform work under the authority of the KDE Runtime.

### Characteristics

| Attribute | Description |
|-----------|-------------|
| **Type** | Agent or Human |
| **Scope** | Work assigned by Runtime |
| **Composition** | AI agents, human contributors |
| **Tenure** | Per-task or ongoing |

### Sub-Types

#### 3.1 Execution Agents (AI)

AI systems that perform KDE engineering work:

| Agent Type | Description | Examples |
|------------|-------------|----------|
| OpenHands | Primary execution agent | OpenHands Agent |
| Claude | Alternative agent | Claude Code |
| Copilot | Alternative agent | GitHub Copilot |
| Codex | Alternative agent | OpenAI Codex |

#### 3.2 Human Contributors

Humans who perform work under KDE:

| Contributor Type | Description |
|-----------------|-------------|
| Developer | Implements code changes |
| Reviewer | Reviews and approves work |
| Architect | Designs system architecture |
| Stakeholder | Provides requirements |

### Responsibilities

1. Perform investigations
2. Execute experiments
3. Implement approved changes
4. Produce artifacts under Runtime Authority
5. Follow KDE methodology

### Boundaries

- **CAN**: Investigate, experiment, implement, produce artifacts
- **CANNOT**: Modify governance, approve governance changes

### Relationship Diagram

```
Runtime Authority
    │
    │ Authorizes
    ▼
Execution Authority
    │
    ├──► AI Agents (OpenHands, Claude, etc.)
    │
    └──► Human Contributors (Developers, Reviewers, etc.)
```

---

## 4. Approval Authority

### Definition

Humans who provide approval for governance-affecting decisions.

### Characteristics

| Attribute | Description |
|-----------|-------------|
| **Type** | Human |
| **Scope** | Governance-affecting decisions |
| **Composition** | Designated human approvers |
| **Tenure** | Role-based, may change |

### Responsibilities

1. Review governance-affecting proposals
2. Approve or reject investigations
3. Authorize significant changes
4. Provide accountability oversight
5. Resolve conflicts

### When Approval is Required

| Decision Type | Approval Required |
|--------------|-------------------|
| Governance changes | Yes (mandatory) |
| Major architectural decisions | Yes (mandatory) |
| Investigation conclusions | Recommended |
| Artifact acceptance | Recommended |
| Policy modifications | Yes (mandatory) |

### Boundaries

- **CAN**: Approve, reject, request modifications
- **CANNOT**: Execute work, modify runtime

### Relationship Diagram

```
Execution Authority
    │
    │ Produces artifacts
    ▼
Artifacts
    │
    │ Reviewed by
    ▼
Approval Authority (Human)
    │
    │ Approves/Rejects
    ▼
Outcome
```

---

## Authority Matrix

### Decision Authority

| Decision | Governance | Runtime | Execution | Approval |
|----------|------------|---------|-----------|----------|
| Define methodology | **A** | I | I | I |
| Create Runtime | R | **A** | I | I |
| Execute investigation | I | I | **R, A** | I |
| Approve investigation | I | I | R | **A** |
| Modify governance | **A** | C | C | **A** |
| Resolve conflicts | **A** | R | R | C |

### Resource Authority

| Resource | Governance | Runtime | Execution | Approval |
|----------|------------|---------|-----------|----------|
| Runtime framework | Owns | Uses | Uses | I |
| Runtime instance | I | Owns | Uses | I |
| Artifacts | I | Owns | Creates | Reviews |
| Governance policies | Owns | Implements | Follows | Interprets |

---

## Relationship Summary

### Hierarchy Flow

```
Governance Authority (Tier 1)
    ↓ defines
Runtime Authority (Tier 2)
    ↓ authorizes
Execution Authority (Tier 3)
    ↓ under oversight of
Approval Authority
```

### Communication Flow

```
Governance ←→ Runtime: Framework definition
Runtime ←→ Execution: Authorization and reporting
Execution ←→ Approval: Review and approval
Runtime ←→ Approval: Oversight coordination
```

---

## Role Independence

### Key Principle

Each authority type maintains independence from others:

| Principle | Description |
|-----------|-------------|
| **Stable Hierarchy** | Governance Authority does not change with project |
| **Consistent Runtime** | Runtime Authority persists across agent changes |
| **Agent Agnostic** | Execution Authority can be any compatible agent |
| **Human Oversight** | Approval Authority provides independent review |

### Evidence

The authority hierarchy ensures:

1. **Agent Independence**: Changing execution agents does not affect runtime
2. **Runtime Stability**: Runtime persists regardless of project changes
3. **Governance Continuity**: Governance Authority provides long-term stability
4. **Accountability**: Human approval provides oversight

---

## Related Documents

| Document | Purpose |
|----------|---------|
| GOV-HIERARCHY-001 | Authority hierarchy policy |
| GOV-NAMING-001 | Artifact naming conventions |
| KDE-INV-001 | Artifact authority model |
| KDE-INV-002 | Governance authority hierarchy |

---

**Status**: ENFORCED
**Review Date**: Upon any authority conflict or KDE-INV-002 follow-up

---

*Generated by KDE Governance - 2026-07-25*
*Per KDE-INV-002 Recommendation*
