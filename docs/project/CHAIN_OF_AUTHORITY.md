---
title: "Chain of Authority"
layer: 4-project
---

# Chain of Authority

## Purpose

This document establishes the **hierarchy of authority** that governs this repository. It defines which layer has authority over which, and establishes the rules for resolving conflicts.

## The Hierarchy

```
┌─────────────────────────────────────────────────────────────┐
│                    KNOWLEDGE LAYER                           │
│                          ↑                                  │
│                    (Highest Authority)                       │
└─────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────┐
│                    ARCHITECTURE LAYER                       │
│                          ↑                                   │
│                   (Authority over Code)                      │
└─────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────┐
│                    ADR LAYER                                │
│                          ↑                                   │
│              (Authority over Decisions)                      │
└─────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────┐
│                   IMPLEMENTATION LAYER                       │
│                          ↑                                  │
│                    (Lowest Authority)                       │
└─────────────────────────────────────────────────────────────┘
```

## Layer Definitions

### Layer 1: Knowledge (Highest)

**Authority**: Protocol specification

**Contains**: Protocol behavior, invariants, requirements

**Origin**: IEEE standards, RFCs, formal specifications

**Example Documents**:

- `docs/protocol/*` - Protocol knowledge base

**Authority Scope**:

- Defines what must be implemented
- Cannot be contradicted by lower layers
- Is the source of truth for behavior

### Layer 2: Architecture

**Authority**: Derived from Knowledge Layer

**Contains**: System design, component responsibilities, interfaces

**Origin**: Design decisions based on knowledge

**Example Documents**:

- `docs/architecture/*` - Architecture documents

**Authority Scope**:

- Defines how knowledge is realized
- Must not contradict Knowledge Layer
- Is authoritative for implementation

### Layer 3: Architecture Decision Records (ADRs)

**Authority**: Derived from Architecture Layer

**Contains**: Rationale for specific decisions, alternatives considered

**Origin**: Decisions made during architecture and implementation

**Example Documents**:

- `docs/adr/*` - Architecture decision records

**Authority Scope**:

- Documents why decisions were made
- Must not contradict Architecture Layer
- Provides context for implementation

### Layer 4: Implementation (Lowest)

**Authority**: Subject to all layers above

**Contains**: Source code, tests, build artifacts

**Origin**: Realization of architecture

**Example**:

- `pkg/*`, `internal/*`, `test/*`

**Authority Scope**:

- Must conform to all layers above
- Cannot contradict any layer
- Is the least authoritative layer

## The Authority Rule

### The Fundamental Rule

> **No lower layer may contradict a higher layer.**

### Application to Implementation

```
If implementation contradicts architecture
    → Implementation is WRONG
    → Architecture MUST be correct
    → Implementation MUST be fixed
```

```
If architecture contradicts knowledge
    → Architecture is WRONG
    → Knowledge MUST be correct
    → Architecture MUST be fixed
```

### The Implication

When a conflict exists:

1. The higher layer is presumed correct
2. The lower layer must change
3. Changes must be documented

## Conflict Resolution

### When Implementation Contradicts Architecture

**Detection**: Code review, testing, static analysis

**Resolution Process**:

```
1. Identify the contradiction
2. Determine which layer should govern
3. If Architecture governs:
   a. Architecture is correct
   b. Implementation is wrong
   c. Fix implementation
4. Document the correction
5. Update tests if needed
```

### When Architecture Contradicts Knowledge

**Detection**: Architecture review, knowledge validation

**Resolution Process**:

```
1. Identify the contradiction
2. Re-examine Knowledge Layer
3. If Knowledge Layer is correct:
   a. Architecture is wrong
   b. Fix architecture
   c. Update ADRs
   d. Update implementation
4. If Knowledge Layer needs correction:
   a. Update Knowledge Layer
   b. Re-validate architecture
   c. Continue resolution
```

### When ADR Contradicts Architecture

**Detection**: ADR review, consistency checks

**Resolution Process**:

```
1. Identify the contradiction
2. Determine if ADR is current
3. If ADR is current:
   a. Architecture may need updating
   b. Consider new ADR to supersede
4. Update as needed
```

## The Traceability Mandate

### Why Traceability Matters

Traceability enables:

- Understanding where requirements come from
- Identifying impact of changes
- Resolving conflicts
- Validating completeness

### Traceability Requirements

| From | To | Required |
|------|-----|----------|
| Implementation | Architecture | Always |
| Architecture | Knowledge | Always |
| ADR | Architecture | When applicable |
| Test | Architecture | Always |
| Test | Knowledge | Recommended |

### Traceability Format

```markdown
# Feature: Binary Input Event Generation

## Knowledge Reference
- Document: docs/protocol/dnp3/150-events.md
- Section: "Event Generation"
- Defines: When events are generated

## Architecture Reference
- Document: docs/architecture/004-package-architecture.md
- Section: "Event Handling"
- Defines: Component responsibilities

## ADR Reference
- Document: docs/adr/001-event-buffer-design.md
- Decision: Use ring buffer for events

## Implementation Reference
- Code: internal/events/buffer.go
- Tests: test/events/buffer_test.go
```

## Authority Boundaries

### What Each Layer Governs

| Layer | Governs |
|-------|---------|
| Knowledge | Protocol behavior, invariants |
| Architecture | System design, components, interfaces |
| ADRs | Design rationale, decisions |
| Implementation | Code, tests, builds |

### What Each Layer Does Not Govern

| Layer | Does Not Govern |
|-------|----------------|
| Knowledge | Implementation details, language choice |
| Architecture | Protocol behavior, specific algorithms |
| ADRs | Day-to-day implementation decisions |
| Implementation | System design, protocol behavior |

## The Review Process

### Before Implementation

1. **Verify Knowledge**: Confirm understanding of requirements
2. **Review Architecture**: Ensure design matches knowledge
3. **Check ADRs**: Understand relevant decisions
4. **Plan Implementation**: Map to architecture

### During Implementation

1. **Stay Within Architecture**: Follow documented design
2. **Document Deviations**: Note when architecture doesn't fit
3. **Request Clarification**: When architecture is unclear
4. **Propose Changes**: Through proper channels

### After Implementation

1. **Validate Against Architecture**: Ensure conformance
2. **Update Traceability**: Document references
3. **Review Tests**: Ensure coverage
4. **Document Lessons**: Capture insights

## Summary

The Chain of Authority establishes clear governance:

1. Knowledge is highest authority
2. Architecture derives from Knowledge
3. ADRs document Architecture decisions
4. Implementation realizes Architecture
5. Lower layers must not contradict higher layers

This hierarchy ensures:

- Consistent understanding
- Clear decision-making
- Effective conflict resolution
- Complete traceability

---

## See Also

- [KNOWLEDGE_FIRST_ENGINEERING.md](KNOWLEDGE_FIRST_ENGINEERING.md)
- [REPOSITORY_STRUCTURE.md](REPOSITORY_STRUCTURE.md)
- [DEVELOPMENT_WORKFLOW.md](DEVELOPMENT_WORKFLOW.md)
