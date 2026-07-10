# Documentation

This directory contains the complete documentation hierarchy for this repository, organized into four distinct layers.

## The Four Layers

```
┌─────────────────────────────────────────────────────────────┐
│                    LAYER 1: PROTOCOL                         │
│               Protocol Knowledge (Highest)                    │
│          docs/protocol/ - Language-neutral knowledge          │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                    LAYER 2: ARCHITECTURE                     │
│                System Design (go-dnp3 specific)             │
│              docs/architecture/ - Design documents          │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                    LAYER 3: ADRs                             │
│              Architecture Decision Records                    │
│                  docs/adr/ - Decision rationale               │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                    LAYER 4: PROJECT                         │
│                  Project Governance                          │
│                 docs/project/ - Methodology                  │
└─────────────────────────────────────────────────────────────┘
```

## Layer 1: Protocol Knowledge

**Location**: `docs/protocol/`

**Purpose**: Vendor-neutral, implementation-independent protocol documentation

**Contents**: Protocol specifications, behavioral descriptions, invariants

**Characteristics**:
- Language neutral
- Implementation neutral
- Based on official standards
- Valid regardless of technology choices

**Example**: `docs/protocol/dnp3/060-link-layer.md`

**Read this if**: You need to understand the protocol

---

## Layer 2: Architecture

**Location**: `docs/architecture/`

**Purpose**: System design for this specific implementation

**Contents**: Component design, interfaces, concurrency model, memory model

**Characteristics**:
- Implementation specific
- Derived from protocol knowledge
- Design focused
- Updated with architecture evolution

**Example**: `docs/architecture/004-package-architecture.md`

**Read this if**: You need to understand how go-dnp3 is designed

---

## Layer 3: Architecture Decision Records

**Location**: `docs/adr/`

**Purpose**: Documentation of decisions and rationale

**Contents**: Decision records, alternatives, consequences

**Characteristics**:
- Rationale focused
- Historical record
- Referenced from code
- Updated when decisions change

**Example**: `docs/adr/README.md`

**Read this if**: You want to understand why decisions were made

---

## Layer 4: Project Governance

**Location**: `docs/project/`

**Purpose**: Project methodology and governance

**Contents**: Engineering philosophy, workflow, contribution guidelines

**Characteristics**:
- Governance focused
- Methodology defined
- Authority clarified
- Permanent

**Documents**:
- [README.md](project/README.md) - Layer overview
- [PROTOCOL_ENGINEERING_PHILOSOPHY.md](project/PROTOCOL_ENGINEERING_PHILOSOPHY.md) - Why protocol engineering matters
- [KNOWLEDGE_FIRST_ENGINEERING.md](project/KNOWLEDGE_FIRST_ENGINEERING.md) - Methodology
- [CHAIN_OF_AUTHORITY.md](project/CHAIN_OF_AUTHORITY.md) - Hierarchy rules
- [REPOSITORY_STRUCTURE.md](project/REPOSITORY_STRUCTURE.md) - Organization
- [DEVELOPMENT_WORKFLOW.md](project/DEVELOPMENT_WORKFLOW.md) - Process

**Read this if**: You want to understand how to contribute or how the project is governed

---

## The Chain of Authority

```
Knowledge (Protocol)  →  Architecture  →  ADRs  →  Implementation
     ↑                        ↑              ↑            ↑
 Highest Authority      Derives from   Derives from   Subject to
                       Knowledge        Architecture   All Above
```

**Rule**: No lower layer may contradict a higher layer.

---

## Quick Navigation

| What you need | Where to look |
|--------------|---------------|
| Understand DNP3 protocol | `docs/protocol/` |
| Understand go-dnp3 design | `docs/architecture/` |
| Understand why decisions were made | `docs/adr/` |
| Understand how to contribute | `docs/project/` |

---

## Contributing to Documentation

### For Protocol Knowledge

1. Read the [protocol engineering philosophy](../project/PROTOCOL_ENGINEERING_PHILOSOPHY.md)
2. Verify against official specifications
3. Maintain language neutrality
4. Submit for review

### For Architecture

1. Read the [knowledge-first engineering](../project/KNOWLEDGE_FIRST_ENGINEERING.md)
2. Derive from protocol knowledge
3. Document decisions
4. Submit for review

### For ADRs

1. Follow the [ADR format](adr/README.md)
2. Document alternatives considered
3. Explain consequences
4. Submit for review

### For Project Governance

1. Follow the [chain of authority](../project/CHAIN_OF_AUTHORITY.md)
2. Propose changes with rationale
3. Ensure no layer violations
4. Submit for review

---

## Key Principles

1. **Knowledge is permanent** - The protocol knowledge layer is the foundation
2. **Architecture derives from knowledge** - Design must not contradict protocol
3. **Implementation follows architecture** - Code must not contradict design
4. **Traceability is mandatory** - Every feature traces to its origin
5. **Documentation precedes implementation** - Document before you build
