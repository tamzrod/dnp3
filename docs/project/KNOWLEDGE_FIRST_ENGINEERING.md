---
title: "Knowledge-First Engineering"
layer: 4-project
---

# Knowledge-First Engineering

## Definition

**Knowledge-First Engineering** is a methodology where knowledge extraction, documentation, and preservation precede and govern all implementation activities.

## Core Principle

> Document what you know before you build what you design.
> Design what you will build before you build it.
> Build what you designed.

## The Knowledge Hierarchy

### Four Layers

```
┌─────────────────────────────────────────────────────────────┐
│                     LAYER 1: KNOWLEDGE                       │
│                   (Protocol Specification)                    │
│              Authority: This document defines reality          │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                    LAYER 2: ARCHITECTURE                    │
│                   (System Design)                           │
│            Authority: Derived from Knowledge Layer           │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                  LAYER 3: ARCHITECTURE                      │
│               (Design Decisions)                             │
│          Authority: Documents why decisions were made         │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                   LAYER 4: IMPLEMENTATION                    │
│                    (Code)                                  │
│         Authority: Must conform to all layers above          │
└─────────────────────────────────────────────────────────────┘
```

### Layer Definitions

| Layer | Name | Purpose | Governed By |
|-------|------|--------|------------|
| 1 | Knowledge | What the protocol requires | Protocol specification |
| 2 | Architecture | How we realize the protocol | Knowledge layer |
| 3 | ADRs | Why we made design choices | Architecture |
| 4 | Implementation | The actual code | All layers above |

## The Knowledge-First Workflow

### Phase 1: Knowledge Extraction

```
1. Read the protocol specification
2. Analyze protocol behavior
3. Document protocol invariants
4. Identify edge cases
5. Extract implementation requirements
```

**Output**: Protocol knowledge base

### Phase 2: Architecture Design

```
1. Design system architecture
2. Define component responsibilities
3. Design interfaces
4. Document design decisions
5. Review architecture against knowledge
```

**Output**: Architecture documents

### Phase 3: ADR Creation

```
1. Identify significant decisions
2. Document alternatives considered
3. Explain rationale
4. Record consequences
5. Review for completeness
```

**Output**: Architecture decision records

### Phase 4: Implementation

```
1. Implement per architecture
2. Follow documented patterns
3. Maintain traceability
4. Document deviations
5. Update as needed
```

**Output**: Source code

### Phase 5: Testing

```
1. Map tests to architecture
2. Map tests to protocol behavior
3. Document test coverage
4. Maintain traceability
5. Report gaps
```

**Output**: Test suite

## The Documentation Mandate

### Why Documentation First

Documentation:

- Forces understanding before implementation
- Creates reference for future work
- Enables review before investment
- Preserves knowledge independently
- Supports maintainability

### The Documentation Paradox

Many believe: "Working code is more important than documentation."

We believe: "Documentation enables working code."

Without documentation:

- Understanding is lost when people leave
- Design decisions are forgotten
- Edge cases are rediscovered repeatedly
- Consistency degrades over time

## Traceability Requirements

### What Must Be Traceable

Every implemented feature must have a trace:

```
Feature
    ↓
References Architecture
    ↓
References Knowledge
    ↓
References Protocol Specification
```

### Traceability Format

For each feature:

1. **Knowledge reference**: Which knowledge document defines the requirement?
2. **Architecture reference**: Which architecture document specifies the design?
3. **ADR reference**: Which ADR documents the decision?
4. **Test reference**: Which tests validate the implementation?

### Prohibited Patterns

```
Feature → Implementation → (nothing else)
```

Features may not skip layers. Every feature must be traceable to its origin.

## The Anti-Patterns

### Pattern 1: Code-First Documentation

```
Write code → Write tests → Write docs (if time)
```

**Problem**: Documentation describes implementation, not requirement.

**Correct Pattern**:

```
Document requirement → Design → Implement → Verify
```

### Pattern 2: Implementation as Specification

```
"Read the code to understand how it works"
```

**Problem**: Knowledge is embedded in code.

**Correct Pattern**:

```
"Read the knowledge base, then read the code"
```

### Pattern 3: Architecture as Afterthought

```
Build first → Refactor into architecture → Document
```

**Problem**: Architecture doesn't guide implementation.

**Correct Pattern**:

```
Design architecture → Implement per architecture → Document
```

### Pattern 4: Decision Amnesia

```
Make decision → Implement → Forget why
```

**Problem**: Future decisions may contradict past reasoning.

**Correct Pattern**:

```
Make decision → Document in ADR → Reference in code
```

## The AI Engineering Workflow

When an AI contributes:

### Step 1: Read Knowledge

```
Read docs/protocol/* first
```

An AI must understand the protocol before proceeding.

### Step 2: Read Architecture

```
Read docs/architecture/* second
```

An AI must understand the design before implementing.

### Step 3: Read ADRs

```
Read docs/adr/* third
```

An AI must understand decisions made.

### Step 4: Implement

```
Write code per architecture
```

Implementation must reference architecture.

### Step 5: Test

```
Write tests per architecture
```

Tests must validate implementation against architecture.

### The Forbidden Pattern

An AI must never:

1. Infer protocol behavior from source code
2. Use implementation as documentation
3. Skip layers in the hierarchy
4. Make architectural decisions without documentation

## Quality Standards

### For Knowledge Documents

- [ ] Self-contained: Understandable without other documents
- [ ] Complete: Covers all relevant aspects
- [ ] Accurate: Matches protocol specification
- [ ] Current: Updated when protocol changes
- [ ] Accessible: Organized for easy navigation

### For Architecture Documents

- [ ] Derived: Based on knowledge documents
- [ ] Complete: Covers all components
- [ ] Consistent: No contradictions within
- [ ] Traceable: Links to knowledge documents
- [ ] Maintained: Updated when requirements change

### For ADRs

- [ ] Justified: Explains why decision was made
- [ ] Complete: Documents alternatives considered
- [ ] Current: Updated if decision changes
- [ ] Referenced: Linked from implementation

## Summary

Knowledge-First Engineering ensures that:

1. Protocol knowledge is preserved independently
2. Architecture derives from knowledge
3. Decisions are documented
4. Implementation follows architecture
5. Testing validates implementation

The knowledge base is the foundation. Everything else is built upon it.

---

## See Also

- [CHAIN_OF_AUTHORITY.md](CHAIN_OF_AUTHORITY.md)
- [REPOSITORY_STRUCTURE.md](REPOSITORY_STRUCTURE.md)
- [DEVELOPMENT_WORKFLOW.md](DEVELOPMENT_WORKFLOW.md)
