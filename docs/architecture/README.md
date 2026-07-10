# Architecture Documents

This directory contains architecture documents that define the system design for go-dnp3.

## Position in the Layer Hierarchy

```
Layer 1: Protocol Knowledge (docs/protocol/)  ← Higher Authority
Layer 2: Architecture (THIS DIRECTORY)          ← Architecture derives from Protocol
Layer 3: ADRs (docs/adr/)                        ← Documents decisions
Layer 4: Project (docs/project/)                  ← Governance
```

**Important**: Architecture must not contradict protocol knowledge. The protocol knowledge base (`docs/protocol/`) is authoritative for protocol behavior.

## Documents

The architecture documents are numbered to indicate reading order:

| Document | Purpose |
|----------|---------|
| [000-philosophy.md](000-philosophy.md) | Native Go, protocol-first design philosophy |
| [001-goals.md](001-goals.md) | Project goals and success criteria |
| [002-non-goals.md](002-non-goals.md) | Explicit non-goals and exclusions |
| [003-guiding-principles.md](003-guiding-principles.md) | Design and engineering principles |
| [004-package-architecture.md](004-package-architecture.md) | Proposed package structure |
| [005-development-methodology.md](005-development-methodology.md) | Architecture-first workflow |
| [006-testing-strategy.md](006-testing-strategy.md) | Testing approach and layers |
| [007-performance-goals.md](007-performance-goals.md) | Performance targets |
| [008-concurrency-model.md](008-concurrency-model.md) | Go concurrency patterns |
| [009-memory-model.md](009-memory-model.md) | Memory management approach |
| [010-roadmap.md](010-roadmap.md) | 7-phase roadmap |

## Purpose

These documents define the **architecture** of go-dnp3. They explain how the protocol is realized in code.

## Relationship to Protocol Knowledge

Architecture **derives from** protocol knowledge:

```
docs/protocol/dnp3/050-layer-model.md  →  docs/architecture/004-package-architecture.md
       (What the protocol requires)              (How we implement it)
```

Architecture documents **must not redefine** protocol behavior. If the architecture document describes protocol behavior, it should link to the protocol knowledge base, not duplicate it.

## Key Principles

1. **Architecture derives from Protocol** - Every design decision traces to protocol requirements
2. **Architecture defines the How** - ADRs explain the Why
3. **Implementation follows Architecture** - Code must not contradict architecture
4. **Traceability is mandatory** - Every architectural concept references protocol knowledge

## Reading Order

Start with 000-philosophy.md and proceed in order.

## Contributing

When contributing architecture changes:

1. Read the [project documentation](../project/README.md) first
2. Follow the [chain of authority](../project/CHAIN_OF_AUTHORITY.md)
3. Ensure architecture does not contradict protocol knowledge
4. Follow the existing document structure
5. Use the numbered document format
6. Include diagrams where helpful
7. Update related ADRs
