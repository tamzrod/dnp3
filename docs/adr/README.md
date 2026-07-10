# Architecture Decision Records (ADRs)

This directory contains Architecture Decision Records for go-dnp3.

## Position in the Layer Hierarchy

```
Layer 1: Protocol Knowledge (docs/protocol/)  ← Higher Authority
Layer 2: Architecture (docs/architecture/)      ← ADRs document decisions
Layer 3: ADRs (THIS DIRECTORY)                  ← Documents WHY
Layer 4: Project (docs/project/)                  ← Governance
```

ADRs document decisions made during architecture and implementation. They explain the **why** behind choices.

## What is an ADR?

An Architecture Decision Record (ADR) is a document that captures an important architectural decision made along with its context and consequences.

## Purpose

ADRs serve multiple purposes:

1. **Traceability**: Links decisions to architecture and protocol
2. **Context**: Provides reasoning for future maintainers
3. **History**: Records how understanding evolved
4. **Governance**: Ensures decisions are reviewed

## Relationship to Other Layers

```
Protocol Knowledge (Layer 1)        ← Highest Authority
        ↓
Architecture (Layer 2)             ← ADRs document decisions
        ↓
This ADR Layer (Layer 3)           ← Documents WHY decisions were made
        ↓
Implementation (Code)               ← Must follow ADRs
```

**Rule**: ADRs must not contradict architecture. Architecture must not contradict protocol knowledge.

## Why ADRs?

ADRs help us:

- Document decisions for future reference
- Track the evolution of the architecture
- Share reasoning with new contributors
- Provide context for changes
- Make decision-making transparent

## When to Write an ADR

Write an ADR when making:

- A significant architectural choice
- A decision between alternatives
- A constraint that affects design
- A technology selection
- A non-obvious trade-off

## ADR Format

```markdown
# ADR-XXX: Title

## Status
[Proposed | Accepted | Deprecated | Superseded by ADR-YYY]

## Context
[What is the issue that we're seeing that is motivating this decision?]

## Decision
[What is the change that we're proposing and/or doing?]

## Consequences
[What becomes easier or more difficult because of this change?]

## Traceability
[References to Protocol Knowledge and Architecture documents]
```

## ADR Lifecycle

### 1. Proposed

ADR is draft, under discussion.

### 2. Accepted

ADR is approved, decisions made.

### 3. Deprecated

ADR is superseded or no longer relevant.

### 4. Superseded

ADR is replaced by another ADR.

## Naming Convention

ADRs are numbered sequentially:

```
ADR-001-package-structure.md
ADR-002-error-handling.md
ADR-003-concurrency-model.md
...
```

## Traceability Requirements

Every ADR must reference:

1. **Architecture documents**: Which architecture document(s) does this affect?
2. **Protocol knowledge**: Does this relate to any protocol behavior?

```markdown
## Traceability

- Architecture: [Link to docs/architecture/*]
- Protocol: [Link to docs/protocol/*] (if applicable)
```

## Processing New ADRs

### Creating an ADR

1. Create a new file in this directory
2. Follow the ADR format
3. Include traceability references
4. Set status to "Proposed"
5. Submit for review

### Review Process

1. Open a pull request with the ADR
2. Team reviews and discusses
3. Verify traceability
4. Address feedback
5. Merge when approved
6. Update status to "Accepted"

### Updating ADRs

To update an accepted ADR:

1. Create a new ADR that supersedes it
2. Update the original ADR status
3. Reference the superseding ADR

## Index

| Number | Title | Status |
|--------|-------|--------|
| ADR-001 | Package Structure | Accepted |
| ADR-002 | Error Handling Strategy | Accepted |
| ADR-003 | Concurrency Model | Accepted |
| ADR-004 | Memory Model | Accepted |
| ADR-005 | Testing Strategy | Accepted |

## References

- [Documenting Architecture Decisions - Reginald Braithwaite](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
- [ADRs on ThoughtWorks Tech Radar](https://www.thoughtworks.com/radar/techniques/lightweight-architecture-decision-records)
- [Chain of Authority](../project/CHAIN_OF_AUTHORITY.md)
- [Repository Structure](../project/REPOSITORY_STRUCTURE.md)

## Contributing

Before writing code, check if there's an ADR that affects your work.

If your work requires a new architectural decision, write an ADR first.

## Questions?

Open an issue to discuss architectural decisions.
