# Architecture Decision Records (ADRs)

This directory contains Architecture Decision Records for go-dnp3.

## What is an ADR?

An Architecture Decision Record (ADR) is a document that captures an 
important architectural decision made along with its context and 
consequences.

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

## Processing New ADRs

### Creating an ADR

1. Create a new file in this directory
2. Follow the ADR format
3. Set status to "Proposed"
4. Submit for review

### Review Process

1. Open a pull request with the ADR
2. Team reviews and discusses
3. Address feedback
4. Merge when approved
5. Update status to "Accepted"

### Updating ADRs

To update an accepted ADR:

1. Create a new ADR that supersedes it
2. Update the original ADR status
3. Reference the superseding ADR

## Index

| Number | Title | Status |
|--------|-------|--------|
| - | (No ADRs yet) | - |

## References

- [Documenting Architecture Decisions - Reginald Braithwaite](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
- [ADRs on ThoughtWorks Tech Radar](https://www.thoughtworks.com/radar/techniques/lightweight-architecture-decision-records)

## Contributing

Before writing code, check if there's an ADR that affects your work.

If your work requires a new architectural decision, write an ADR first.

## Questions?

Open an issue to discuss architectural decisions.
