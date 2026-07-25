# KDE Governance Authority Hierarchy

**Document ID**: GOV-HIERARCHY-001
**Version**: 1.0.0
**Date**: 2026-07-25
**Status**: APPROVED
**Authority**: KDE Governance (per KDE-INV-002)

---

## Purpose

This policy establishes the formal Governance Authority Hierarchy for KDE Runtime, defining the relationships between Governance Authority, Runtime Authority, Execution Authority, and Approval Authority.

This policy was created as a result of investigation **KDE-INV-002**, which validated the hypothesis that KDE requires a formal Governance Authority Hierarchy.

---

## Governance Authority Hierarchy

### Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│ TIER 1: GOVERNANCE AUTHORITY (External)                              │
│ - Defines KDE methodology and governance policies                    │
│ - Creates the KDE Runtime Framework                                 │
│ - Does not participate in day-to-day execution                      │
└─────────────────────────────────────────────────────────────────────┘
                              ↓ defines
┌─────────────────────────────────────────────────────────────────────┐
│ TIER 2: RUNTIME AUTHORITY (KDE Runtime Instance)                     │
│ - Executes governance according to approved policies                 │
│ - Authorizes execution agents                                        │
│ - Owns project artifacts                                            │
│ - Maintains runtime state                                           │
└─────────────────────────────────────────────────────────────────────┘
                              ↓ authorizes
┌─────────────────────────────────────────────────────────────────────┐
│ TIER 3: EXECUTION AUTHORITY                                         │
│ - Execution Agents: AI agents (OpenHands, Claude, etc.)             │
│ - Human Contributors: Humans who perform work                       │
│ - Performs investigations, experiments, implementations             │
│ - Produces artifacts under Runtime Authority                        │
└─────────────────────────────────────────────────────────────────────┘
                              ↓ oversight for governance matters
┌─────────────────────────────────────────────────────────────────────┐
│ APPROVAL AUTHORITY (Humans)                                          │
│ - Reviews and approves governance-affecting decisions               │
│ - Can accept, reject, or request modifications                      │
│ - Provides human oversight for accountability                       │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Authority Types

### 1. Governance Authority

**Definition**: The external entity (human or organization) that defines KDE methodology and governance policies.

**Characteristics**:
- Creates the KDE Runtime Framework
- Defines governance principles and policies
- Does not participate in day-to-day execution
- Authority is external to any specific project

**Relationship to Runtime**:
- Creates and defines the Runtime Framework
- Runtime executes governance as defined by Governance Authority

### 2. Runtime Authority

**Definition**: The KDE Runtime instance that executes governance for a project.

**Characteristics**:
- Instantiated from the KDE Runtime Framework
- Executes governance policies
- Authorizes execution agents
- Maintains runtime state
- Owns project artifacts

**Relationship to Governance Authority**:
- Operates under authority of Governance Authority
- Cannot modify its own governance structure

**Relationship to Execution Authority**:
- Authorizes execution agents
- Receives artifacts from execution agents

### 3. Execution Authority

**Definition**: The agents and humans who perform work under the authority of the KDE Runtime.

**Characteristics**:
- Performs investigations, experiments, implementations
- Produces artifacts under Runtime Authority
- May be AI agents or humans
- Authority is derived from Runtime

**Relationship to Runtime Authority**:
- Authorized by Runtime
- Operates within boundaries set by Runtime
- Produces artifacts owned by Runtime

### 4. Approval Authority

**Definition**: Humans who provide approval for governance-affecting decisions.

**Characteristics**:
- Reviews and approves governance-affecting artifacts
- Can accept, reject, or request modifications
- Authority is separate from Execution Authority
- Human oversight for accountability

**Relationship to Other Authorities**:
- Provides oversight for Governance Authority decisions
- Reviews Runtime Authority proposals (if any)
- Approves Execution Authority artifacts (when required)

---

## Decision Authority Matrix

| Decision Type | Authority | Approval Required |
|--------------|-----------|-------------------|
| Governance policies | Governance Authority | Yes (by Governance Authority) |
| Runtime configuration | Runtime Instance | No |
| Investigation execution | Execution Agent | No |
| Investigation conclusions | Execution Agent | Recommended |
| Governance-affecting decisions | Execution Agent | Yes (Human Approver) |
| Artifact acceptance | Execution Agent | Recommended |
| Governance change proposals | Any | Yes (Human Approver) |

---

## Responsibility Matrix (RACI)

| Activity | Governance Authority | Runtime Authority | Execution Authority | Approval Authority |
|----------|---------------------|-------------------|--------------------|--------------------|
| Define governance | R, A | I | I | I |
| Create Runtime | R | A | I | I |
| Execute investigation | I | I | R, A | I |
| Approve investigation | I | I | R | A |
| Modify governance | R | C | C | A |
| Resolve conflicts | A | R | R | C |

**Legend**:
- **R** = Responsible (performs the work)
- **A** = Accountable (final decision maker)
- **C** = Consulted (provides input)
- **I** = Informed (kept updated)

---

## Governance Evolution

### Process Overview

```
┌─────────────────────────────────────────────────────────────────┐
│ STATE 1: Initial Governance                                      │
│ - Governance Authority defines KDE Runtime Framework              │
│ - Runtime instantiated with initial policies                     │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│ STATE 2: Runtime Operation                                       │
│ - Runtime executes governance                                    │
│ - Execution Agents perform work                                  │
│ - Artifacts produced under Runtime Authority                     │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│ STATE 3: Evolution Request                                       │
│ - Execution Agent identifies governance need                      │
│ - Proposal created as investigation                             │
│ - Human Approver reviews                                         │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│ STATE 4: Governance Change (if approved)                        │
│ - Governance Authority definition updated                        │
│ - Runtime Framework updated                                       │
│ - New governance propagated to all Runtime instances             │
└─────────────────────────────────────────────────────────────────┘
```

### Governance Change Process

1. **Identification**: Execution Agent identifies need for governance change
2. **Investigation**: Conduct KDE investigation to analyze the need
3. **Proposal**: Document proposed governance change
4. **Review**: Human Approver reviews the proposal
5. **Decision**: Human Approver approves or rejects
6. **Implementation**: If approved, governance is updated
7. **Propagation**: Updated governance applies to all Runtime instances

---

## Artifact Metadata

### Standard Metadata Fields

Per KDE-INV-001, artifacts should include:

```markdown
**Authority**: KDE Runtime ([Project Name])
**Execution Agent**: [Agent Name]
**Human Approver**: [Name] (if applicable)
```

### Extended Metadata (for Governance Documents)

```markdown
**Governance Authority**: KDE Runtime Framework (External)
**Runtime Authority**: KDE Runtime ([Project Name])
**Execution Agent**: [Agent Name]
**Human Approver**: [Name] (if applicable)
```

---

## Related Policies

| Policy ID | Name | Relationship |
|-----------|------|--------------|
| GOV-HIERARCHY-001 | This document | Defines authority hierarchy |
| GOV-NAMING-001 | Laboratory Artifact Naming Conventions | Uses hierarchy for artifact naming |
| GOV-AUTH-001 | Authorization Requirements | (Future) Defines authorization rules |
| GOV-EVIDENCE-001 | Evidence Preservation Standards | Evidence supports hierarchy decisions |

---

## References

- KDE-INV-001: Investigation Artifact Authority Model
- KDE-INV-002: KDE Governance Authority Hierarchy

---

**Status**: ENFORCED
**Review Date**: Upon any governance incident or KDE-INV-002 follow-up

---

*Generated by KDE Governance - 2026-07-25*
*Per KDE-INV-002 Recommendation: ADOPT MODEL C - Hierarchical Governance*
