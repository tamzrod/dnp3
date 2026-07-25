# Knowledge Derivation

## Purpose

This document defines how structured knowledge becomes architecture in KDSE. It describes the conceptual transformation from knowledge artifacts to architectural decisions and architectural artifacts.

## The Derivation Concept

Derivation is the process by which higher-authority artifacts produce lower-authority artifacts. Derivation is not translation or transcription. Derivation involves reasoning, analysis, and decision-making.

### What Derivation Is Not

Derivation is not:

- **Direct mapping**: Knowledge does not map directly to architecture. Architecture requires interpretation and reasoning about knowledge.

- **Automated transformation**: Derivation involves judgment. Multiple architects deriving from the same knowledge may produce different architectures.

- **Deterministic process**: Given the same knowledge, derivation may produce different architectures based on different interpretations or priorities.

- **One-way conversion**: Derivation may reveal gaps in knowledge that require returning to the knowledge stage for clarification.

### What Derivation Is

Derivation is the application of engineering judgment to knowledge to produce architecture. Derivation transforms validated understanding into structural decisions.

Derivation involves:

1. **Interpreting knowledge**: Understanding what knowledge means for system structure
2. **Identifying implications**: Determining what knowledge implies for architecture
3. **Resolving tensions**: Balancing competing knowledge when it suggests different approaches
4. **Making decisions**: Choosing among alternatives based on knowledge and constraints
5. **Documenting rationale**: Recording why decisions were made as they were

## Derivation Stages

Knowledge-to-architecture derivation proceeds through five conceptual stages.

### Stage 1: Concept Analysis

The architect analyzes knowledge artifacts to identify concepts that have architectural implications.

Concept analysis involves:

1. **Identifying domain concepts**: Extracting key concepts from problem domain knowledge
2. **Mapping relationships**: Understanding how concepts relate to each other
3. **Identifying constraints**: Finding constraints that limit possible solutions
4. **Discovering trade-offs**: Recognizing tensions between competing concerns
5. **Gathering context**: Understanding the environment in which the system will operate

**Input**: Knowledge artifacts (requirements, constraints, context, assumptions)
**Output**: Analyzed concepts with architectural implications

**Example**: From knowledge stating "users need to authenticate" and "authentication must be secure" and "authentication must be fast," the architect identifies tensions between security and speed that require architectural resolution.

### Stage 2: Responsibility Identification

The architect identifies responsibilities that the system must fulfill.

Responsibility identification involves:

1. **Defining system boundaries**: What is inside the system versus outside
2. **Allocating concerns**: Which responsibilities belong to which parts
3. **Identifying interfaces**: How parts will communicate
4. **Separating concerns**: Ensuring single responsibilities per component
5. **Establishing stewardship**: Assigning stewards to responsibilities

**Input**: Analyzed concepts from Stage 1
**Output**: Responsibility model with identified system components

**Example**: From the authentication concept, the architect identifies responsibilities including identity verification, credential storage, session management, and access control. These become separate responsibilities with defined interfaces.

### Stage 3: Architectural Decision Making

The architect makes decisions that establish the system's structure.

Architectural decisions address:

1. **Structural decisions**: How components are organized
2. **Behavioral decisions**: How components interact
3. **Implementation decisions**: Technology and pattern choices
4. **Quality decisions**: How non-functional requirements are addressed

Each decision requires:

- **Basis**: What knowledge justifies this decision
- **Alternatives**: What alternatives were considered
- **Trade-offs**: What was weighed in making this decision
- **Consequences**: What implications this decision has

**Input**: Responsibility model from Stage 2
**Output**: Architectural decisions with documented rationale

**Example**: The architect decides to use OAuth 2.0 for authentication. This decision is based on security requirements knowledge, considers alternatives including SAML and local authentication, weighs complexity against security benefits, and has consequences for client compatibility and implementation approach.

### Stage 4: Architecture Documentation

The architect documents the resulting architecture as artifacts.

Architecture documentation includes:

1. **Architecture overview**: High-level structure and organization
2. **Component specifications**: Detailed descriptions of each component
3. **Interface definitions**: How components communicate
4. **Architecture Decision Records**: Individual decision documents
5. **Constraint documentation**: What constraints architecture must satisfy

**Input**: Architectural decisions from Stage 3
**Output**: Architecture artifacts (specifications, ADRs, diagrams)

**Example**: The architect creates an architecture overview showing authentication as a separate service, documents the OAuth 2.0 implementation approach, and records the ADR explaining why OAuth 2.0 was chosen over alternatives.

### Stage 5: Verification Against Knowledge

The architect verifies that the architecture satisfies knowledge artifacts.

Verification involves:

1. **Traceability confirmation**: Ensuring each architectural decision traces to knowledge
2. **Coverage analysis**: Confirming all knowledge has been addressed
3. **Consistency check**: Ensuring decisions do not contradict each other
4. **Gap identification**: Finding knowledge that has not been addressed

**Input**: Architecture artifacts, Knowledge artifacts
**Output**: Verified architecture with confirmed traceability

**Example**: The architect traces each architectural decision back to specific knowledge artifacts, confirms that all security requirements have been addressed, and identifies that the knowledge artifact on performance criteria needs clarification before the caching strategy can be finalized.

## Derivation Dependencies

Each stage depends on the previous stage producing adequate output.

```
Concept Analysis
      │
      ▼ Requires
Responsibility Identification
      │
      ▼ Requires
Architectural Decision Making
      │
      ▼ Requires
Architecture Documentation
      │
      ▼ Requires
Verification Against Knowledge
```

If any stage produces inadequate output, the derivation may need to return to earlier stages.

## Traceability Requirements

Every architectural artifact must trace to the knowledge that authorized it.

### Traceability Direction

Traceability flows upward through derivation:

```
Knowledge
    │
    │ Authorizes
    ▼
Architecture
    │
    │ Justifies
    ▼
Architectural Decisions
    │
    │ Realize
    ▼
Architecture Artifacts
```

### Traceability Depth

Each artifact must identify:

1. **Direct knowledge basis**: Which knowledge artifacts this derives from
2. **Decision basis**: Which architectural decisions this represents
3. **Dependency knowledge**: Which knowledge affects interpretation

### Traceability Example

An authentication service specification traces to:

- Knowledge: User authentication requirements
- Knowledge: Security constraints
- Knowledge: Performance criteria
- ADR: OAuth 2.0 adoption decision
- ADR: Microservices decomposition decision

## What Is Not Derivation

Certain activities are related to derivation but are not derivation itself.

### Derivation Is Not Design

Design produces detailed specifications. Derivation produces architectural decisions.

Design answers questions like "how is this component implemented?" Derivation answers questions like "what components exist and why?"

### Derivation Is Not Implementation Planning

Implementation planning determines how work is organized. Derivation determines what work exists.

Implementation planning answers questions like "in what order do we build components?" Derivation answers questions like "what components do we build?"

### Derivation Is Not Review

Review evaluates derivation output. Derivation produces that output.

Review confirms that architecture satisfies knowledge. Derivation creates architecture that should satisfy knowledge.

## Iterative Derivation

Derivation is not strictly linear. Iterations occur when earlier stages reveal issues.

### Iteration Triggers

Derivation may return to earlier stages when:

1. **Knowledge gaps**: Knowledge is insufficient to support architectural decisions
2. **Knowledge conflicts**: Different knowledge artifacts suggest contradictory approaches
3. **Architecture gaps**: Architecture does not fully address knowledge
4. **Feasibility issues**: Architecture reveals technical constraints not captured in knowledge

### Iteration Process

```
Derivation proceeds until blocked
      │
      ▼
Issue identified
      │
      ▼
Determine whether issue is:
      │
      ├── Knowledge gap ──────► Return to Knowledge stage
      │
      ├── Architecture gap ───► Return to earlier derivation stage
      │
      └── Feasibility issue ──► Return to Knowledge or Architecture
      │
      ▼
Resolve issue
      │
      ▼
Continue derivation
```

### Example Iteration

During architectural decision making, the architect realizes that knowledge about performance requirements is insufficient to determine whether a monolithic or microservices architecture is appropriate. The architect returns to the knowledge stage to clarify performance criteria.

## Derivation and Authority

Derivation is the mechanism by which knowledge maintains authority.

### Authority Through Derivation

When architecture derives from knowledge, architecture inherits knowledge's authority.

- Architecture that traces to validated knowledge carries authority
- Architecture that cannot trace to knowledge carries no authority
- Architecture that contradicts knowledge violates the authority hierarchy

### Authority Limitations

Derivation transfers but does not create authority.

- Derivation cannot add authority that knowledge does not possess
- Derivation cannot resolve conflicts between knowledge artifacts
- Derivation cannot make decisions where knowledge is silent

When knowledge is silent, architecture must make decisions without authoritative guidance. Such decisions carry less authority than those supported by clear knowledge.

## Glossary Additions

This document introduces the following terms for glossary inclusion:

### Derivation

The process by which higher-authority artifacts produce lower-authority artifacts through analysis, decision-making, and documentation. Derivation transforms validated understanding into structural decisions.

### Concept Analysis

The derivation stage in which knowledge artifacts are analyzed to identify concepts with architectural implications.

### Responsibility Identification

The derivation stage in which system responsibilities are identified and allocated to components.

### Architectural Decision

A decision that establishes the system's structure. Architectural decisions include the basis for the decision, alternatives considered, trade-offs weighed, and consequences anticipated.

### Traceability

The ability to follow relationships between artifacts. In derivation, traceability confirms that each artifact traces to the knowledge that authorized it.
