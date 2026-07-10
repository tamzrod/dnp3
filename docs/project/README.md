---
title: "Project Documentation"
layer: 4-project
---

# Project Layer

This directory contains the **governance** and **methodology** for this repository.

## Purpose

The Project layer defines:

- Engineering philosophy
- Development workflow
- Contribution guidelines
- Chain of authority
- Repository structure

## Documents

| Document | Purpose |
|----------|---------|
| [PROTOCOL_ENGINEERING_PHILOSOPHY.md](PROTOCOL_ENGINEERING_PHILOSOPHY.md) | Philosophy of protocol engineering |
| [KNOWLEDGE_FIRST_ENGINEERING.md](KNOWLEDGE_FIRST_ENGINEERING.md) | Knowledge-first methodology |
| [CHAIN_OF_AUTHORITY.md](CHAIN_OF_AUTHORITY.md) | Authority hierarchy |
| [REPOSITORY_STRUCTURE.md](REPOSITORY_STRUCTURE.md) | Repository organization |
| [DEVELOPMENT_WORKFLOW.md](DEVELOPMENT_WORKFLOW.md) | Development process |

## Layer Position

The Project layer is the **lowest layer** in the hierarchy:

```
Layer 1: Protocol Knowledge (Highest)
Layer 2: Architecture
Layer 3: Architecture Decision Records
Layer 4: Project (This layer)
Layer 5: Implementation (Lowest)
```

## Authority

The Project layer governs:

- How work is organized
- Who can make decisions
- How changes are approved
- What workflows must be followed

## Relationship to Other Layers

### Above: Architecture

Architecture must comply with Project governance.

### Below: Implementation

Implementation must comply with Project governance.

## Key Principle

> The Project layer exists to serve the other layers.
> It does not create knowledge.
> It does not design systems.
> It enables both.

## Status

This layer is **permanent**. Its principles govern all work on this repository regardless of technology, team, or timeframe.
