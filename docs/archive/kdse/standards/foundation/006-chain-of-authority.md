# Chain of Authority

## Purpose

This document establishes the hierarchy of authority in KDSE and defines the rules governing how authority flows through the engineering process.

## Fundamental Rule

**No lower layer may contradict a higher layer.**

This rule is fundamental to KDSE. Violations of this rule represent methodology non-compliance.

## Authority Hierarchy

```
Highest Authority
        │
        │ Knowledge
        │ (Authorizes all subsequent decisions)
        ▼
    Architecture
    (Authorizes implementation)
        │
        │ ADR
        │ (Documents specific decisions)
        ▼
    Implementation
    (Realizes architecture)
        │
        │ Verification
        │ (Confirms alignment)
        ▼
    Verification
    (Reports alignment status)
        │
Lowest Authority
```

## Layer Definitions

### Knowledge (Highest Authority)

Knowledge holds the highest authority in KDSE. All other artifacts derive their authority from knowledge.

**Knowledge may not be contradicted by any other artifact type.**

If architecture, implementation, or verification appears to contradict knowledge, the contradiction must be resolved by revisiting the knowledge artifact for clarification or correction.

### Architecture

Architecture derives authority from knowledge. Architecture authorizes implementation.

**Architecture may not be contradicted by implementation or verification.**

If implementation appears to contradict architecture, the implementation must be corrected to align with architecture. If verification identifies such contradictions, they must be reported and corrected.

**Architecture may not contradict knowledge.**

If architecture cannot be derived from knowledge without contradicting other architectural decisions, the knowledge artifact may need clarification or the architecture may need reconsideration.

### Architecture Decision Record (ADR)

ADRs document specific decisions made within the architectural context. ADRs derive authority from architecture and must align with both architecture and knowledge.

**ADRs may not contradict architecture or knowledge.**

### Implementation

Implementation derives authority from architecture. Implementation must realize architecture.

**Implementation may not contradict architecture or knowledge.**

Implementation that cannot be derived from architecture without contradiction indicates either:
- Implementation error requiring correction
- Architecture gap requiring documentation
- Architecture error requiring reconsideration

### Verification

Verification does not create new authority. Verification applies authority to confirm alignment.

**Verification must trace to knowledge and architecture.**

Verification criteria derive from knowledge. Verification execution confirms alignment between implementation and architecture, and between architecture and knowledge.

**Verification may not contradict the authority hierarchy.**

Verification that identifies contradictions must report them. Verification does not create exceptions or justifications for non-compliance.

## Authority Flow Rules

### Rule 1: Authority Flows Downward

Higher authority grants permission for lower authority actions.

```
Knowledge ──grants permission──▶ Architecture ──grants permission──▶ Implementation
```

### Rule 2: Information Flows Upward

Lower layers provide information to higher layers for decision-making.

```
Implementation ──reports──▶ Architecture ──reports──▶ Knowledge
```

### Rule 3: Corrections Flow Both Directions

When contradictions are identified, corrections may flow in either direction.

```
Knowledge ◀──clarification──│──correction──▶ Architecture ◀──correction──│──correction──▶ Implementation
```

### Rule 4: Authority Cannot Be Delegated Downward

Lower layers cannot grant authority that they do not possess.

```
Architecture cannot authorize exceptions to Knowledge
Implementation cannot authorize exceptions to Architecture
```

### Rule 5: Authority Can Be Clarified Upward

Lower layers can request clarification that may result in higher layer changes.

```
Implementation requests clarification from Architecture
Architecture requests clarification from Knowledge
```

## Practical Implications

### For Architecture Decisions

Architectural decisions must cite the knowledge artifacts that authorize them. An architectural decision without knowledge basis is not compliant with KDSE.

### For Implementation Decisions

Implementation decisions must cite the architectural artifacts that authorize them. Implementation must trace to architecture.

### For Verification Results

Verification results must cite the knowledge and architectural artifacts against which verification was performed. Verification without authority citation is not meaningful in KDSE.

### For Exceptions

There are no exceptions to the authority hierarchy. If a lower layer cannot comply with a higher layer's authority, the situation must be resolved through clarification or correction, not through exception-granting.

## Violation Examples

### Example 1: Implementation Contradicts Architecture

**Situation**: Implementation includes a feature that architecture does not authorize.

**Violation**: Implementation is contradicting Architecture.

**Resolution**: Remove the unauthorized feature, or obtain architectural authorization through proper channels.

### Example 2: Architecture Contradicts Knowledge

**Situation**: Architecture specifies a constraint that knowledge does not authorize.

**Violation**: Architecture is contradicting Knowledge.

**Resolution**: Revise architecture to align with knowledge, or clarify knowledge to authorize the constraint.

### Example 3: Verification Without Authority

**Situation**: Verification tests are written without reference to knowledge or architecture artifacts.

**Violation**: Verification lacks authority basis.

**Resolution**: Trace verification criteria to knowledge artifacts before execution.

## Governance Integration

The authority hierarchy operates within the governance framework established by Governance artifacts. Governance defines:

- Who may authorize changes to each layer
- What review processes apply
- How contradictions are resolved
- How exceptions (if any) are handled

Governance artifacts do not supersede the authority hierarchy; they operationalize it.
