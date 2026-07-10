---
title: "Repository Structure"
layer: 4-project
---

# Repository Structure

## Purpose

This document defines the organization of this repository. It explains where each type of content belongs and why.

## The Four-Layer Structure

```
docs/
├── protocol/          # Layer 1: Protocol Knowledge
├── architecture/      # Layer 2: Architecture
├── adr/              # Layer 3: Architecture Decisions
└── project/          # Layer 4: Project Governance

src/                   # Layer 4: Implementation
├── pkg/              # Public packages
├── internal/          # Private packages
└── test/             # Test utilities
```

## Layer 1: Protocol Knowledge

### Location

```
docs/protocol/
```

### Purpose

Contains vendor-neutral, implementation-independent protocol documentation.

### Contents

- Protocol specifications
- Behavioral descriptions
- Invariant documentation
- Edge case analysis

### Characteristics

- **Language neutral**: Written without reference to any programming language
- **Implementation neutral**: Applicable to any implementation
- **Vendor neutral**: Based on official specifications
- **Timeless**: Valid regardless of current technology

### Example Documents

```
docs/protocol/
├── dnp3/
│   ├── 000-introduction.md
│   ├── 010-history.md
│   ├── 060-link-layer.md
│   └── ...
└── README.md
```

### Ownership

These documents define the **what**. They cannot be contradicted by any other layer.

## Layer 2: Architecture

### Location

```
docs/architecture/
```

### Purpose

Contains system design documentation for this specific implementation.

### Contents

- System architecture
- Component design
- Interface specifications
- Concurrency models
- Memory management

### Characteristics

- **Implementation specific**: Describes this implementation
- **Architecture derived**: Based on protocol knowledge
- **Design focused**: Defines structure and organization
- **Maintained**: Updated with architecture evolution

### Example Documents

```
docs/architecture/
├── 000-philosophy.md
├── 001-goals.md
├── 004-package-architecture.md
├── 008-concurrency-model.md
└── README.md
```

### Ownership

These documents define the **how** for this implementation. They must not contradict protocol knowledge.

## Layer 3: Architecture Decision Records

### Location

```
docs/adr/
```

### Purpose

Contains documentation of architectural decisions and their rationale.

### Contents

- Decision records
- Alternatives considered
- Consequences documented
- Supersession records

### Characteristics

- **Rationale focused**: Explains why, not just what
- **Historical**: Documents decision evolution
- **Referenced**: Linked from architecture and implementation
- **Maintained**: Updated when decisions change

### Example Documents

```
docs/adr/
├── README.md
├── 001-package-structure.md
├── 002-error-handling.md
└── ...
```

### Ownership

These documents explain the **why** behind architectural decisions. They provide context for current design.

## Layer 4: Project Governance

### Location

```
docs/project/
```

### Purpose

Contains project management and methodology documentation.

### Contents

- Engineering philosophy
- Development workflow
- Contribution guidelines
- Chain of authority

### Characteristics

- **Governance focused**: Defines how work is organized
- **Methodology defined**: Establishes processes
- **Authority clarified**: Defines decision hierarchy
- **Permanent**: Governs all work regardless of technology

### Example Documents

```
docs/project/
├── README.md
├── PROTOCOL_ENGINEERING_PHILOSOPHY.md
├── KNOWLEDGE_FIRST_ENGINEERING.md
├── CHAIN_OF_AUTHORITY.md
├── REPOSITORY_STRUCTURE.md
└── DEVELOPMENT_WORKFLOW.md
```

### Ownership

These documents govern the project. They apply to all layers below.

## Implementation Layer

### Location

```
pkg/           # Public packages
internal/      # Private packages
cmd/           # Command-line tools
test/          # Test utilities
```

### Purpose

Contains the actual implementation.

### Contents

- Source code
- Tests
- Build configuration
- Examples

### Characteristics

- **Realization focused**: Implements architecture
- **Language specific**: Written in Go
- **Tested**: Has test coverage
- **Traceable**: References architecture

### Ownership

Implementation is governed by all layers above. It cannot contradict any layer.

## Directory Summary

| Directory | Layer | Contents |
|-----------|-------|---------|
| `docs/protocol/` | 1 | Protocol knowledge |
| `docs/architecture/` | 2 | System architecture |
| `docs/adr/` | 3 | Decision records |
| `docs/project/` | 4 | Project governance |
| `pkg/`, `internal/` | Implementation | Source code |
| `test/` | Implementation | Test code |

## Supporting Directories

### `.github/`

```
.github/
├── workflows/     # CI/CD configuration
├── ISSUE_TEMPLATE/  # Issue templates
└── PULL_REQUEST_TEMPLATE/  # PR templates
```

Contains configuration for GitHub features.

### `scripts/`

```
scripts/
├── build.sh
├── test.sh
└── ...
```

Contains utility scripts for development tasks.

### `examples/`

```
examples/
├── basic/
├── advanced/
└── README.md
```

Contains example code demonstrating usage.

## Root Documents

### `README.md`

Entry point for the repository. Links to all layers.

### `LICENSE`

Apache 2.0 license.

### `CONTRIBUTING.md`

Contribution guidelines.

### `CODE_OF_CONDUCT.md`

Community conduct rules.

### `SECURITY.md`

Security reporting process.

### `Makefile`

Build and development tasks.

## Layer Relationship Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         README.md                            │
│                    (Entry point, links to layers)            │
└─────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        ↓                     ↓                     ↓
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│   PROTOCOL    │   │  ARCHITECTURE │   │    PROJECT    │
│   (Layer 1)  │   │   (Layer 2)   │   │   (Layer 4)   │
│               │   │               │   │               │
│ What the      │   │ How we        │   │ How we        │
│ protocol      │   │ implement     │   │ govern        │
│ requires      │   │               │   │               │
└───────────────┘   └───────────────┘   └───────────────┘
        │                     │                     │
        └─────────────────────┼─────────────────────┘
                              ↓
                    ┌───────────────┐
                    │      ADR      │
                    │   (Layer 3)   │
                    │               │
                    │ Why we made   │
                    │ decisions     │
                    └───────────────┘
                              │
                              ↓
                    ┌───────────────┐
                    │IMPLEMENTATION │
                    │  (Code)       │
                    │               │
                    │ The actual    │
                    │ realization   │
                    └───────────────┘
```

## Cross-References

Documents must reference other layers:

| Document Type | Must Reference |
|--------------|----------------|
| Architecture | Protocol documents |
| ADR | Architecture documents |
| Implementation | Architecture documents |
| Tests | Architecture documents |

## Summary

The repository structure reflects the knowledge hierarchy:

1. Protocol knowledge at the foundation
2. Architecture built upon it
3. Decisions documented separately
4. Implementation at the top

This structure makes the knowledge hierarchy visible and enforceable.

---

## See Also

- [KNOWLEDGE_FIRST_ENGINEERING.md](KNOWLEDGE_FIRST_ENGINEERING.md)
- [CHAIN_OF_AUTHORITY.md](CHAIN_OF_AUTHORITY.md)
- [DEVELOPMENT_WORKFLOW.md](DEVELOPMENT_WORKFLOW.md)
