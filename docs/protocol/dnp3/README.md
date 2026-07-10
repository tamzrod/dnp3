# DNP3 Protocol Knowledge Base

**Vendor-neutral, AI-optimized engineering reference for IEEE 1815 (DNP3)**

---

## Purpose

This directory contains the authoritative, vendor-neutral protocol knowledge base for DNP3 (Distributed Network Protocol 3) as defined in IEEE 1815. This knowledge base:

- **Teaches DNP3** as an engineering discipline, not as a library feature
- **Remains implementation-agnostic** - valid regardless of language choice
- **Serves as the single source of truth** for all future work
- **Enables deterministic protocol interpretation** without reference implementations

---

## What This Is

This is an **open engineering handbook** for DNP3. It explains:

- Why the protocol was designed as it was
- What problems each feature solves
- How protocol concepts interrelate
- Common misunderstandings and how to avoid them
- Engineering trade-offs in protocol design

---

## What This Is NOT

- Implementation documentation
- API reference
- Architecture decisions
- Package documentation
- Language-specific guides

---

## Document Organization

### Learning Path (Recommended Reading Order)

1. [000-introduction.md](000-introduction.md) - What is DNP3?
2. [010-history.md](010-history.md) - Protocol evolution
3. [020-design-goals.md](020-design-goals.md) - Design objectives
4. [030-core-concepts.md](030-core-concepts.md) - Fundamental concepts
5. [050-layer-model.md](050-layer-model.md) - Protocol architecture
6. [060-link-layer.md](060-link-layer.md) - Data Link Layer
7. [070-transport-layer.md](070-transport-layer.md) - Transport Layer
8. [080-application-layer.md](080-application-layer.md) - Application Layer
9. [090-object-model.md](090-object-model.md) - Object addressing
10. [100-measurements.md](100-measurements.md) - Binary and analog data
11. [110-controls.md](110-controls.md) - Control operations
12. [150-events.md](150-events.md) - Event generation and reporting
13. [170-unsolicited-responses.md](170-unsolicited-responses.md) - Unsolicted responses
14. [180-time-synchronization.md](180-time-synchronization.md) - Time synchronization
15. [210-sequence-numbers.md](210-sequence-numbers.md) - Message sequencing
16. [280-security.md](280-security.md) - Secure Authentication
17. [330-glossary.md](330-glossary.md) - Term definitions

---

## Document Categories

### Foundational (Start Here)
- Introduction, History, Design Goals, Core Concepts

### Protocol Layers
- Layer Model, Link Layer, Transport Layer, Application Layer

### Data Types
- Object Model, Variations, Measurements, Quality Flags, Deadbands

### Operations
- Controls, Events, Time Synchronization, Confirmations

### System Architecture
- Master, Outstation, Database, Class Polling

### Advanced Topics
- Fragmentation, Sequence Numbers, FCB, Security
- Interoperability, Conformance, Performance
- Common Misconceptions, FAQ

---

## Ownership Rule

Each protocol concept has exactly one owner document. Other documents must link to the owner rather than duplicating content.

**Example**: Sequence numbers are explained in detail in [210-sequence-numbers.md](210-sequence-numbers.md). Other documents reference it, not explain it.

---

## Quality Standards

Every document must:

1. **Answer exactly ONE question** - "What is X?" or "How does X work?"
2. **Include all required sections** - Purpose, Problem Solved, Concepts, Relationships, etc.
3. **Be self-contained** - Can be understood in isolation
4. **Be cross-linked** - Links to related concepts
5. **Avoid duplication** - Reference other documents, don't repeat
6. **Be deterministic** - One correct interpretation, no ambiguity

---

## Source References

This knowledge base is derived from:

- IEEE 1815-2012 (Primary Standard)
- IEEE 1815-2012 Technical Corrigendum 1
- Protocol analysis and behavioral specification
- Industry best practices

**Note**: This knowledge base does NOT reference implementation source code from OpenDNP3, Rust DNP3, or any other library. Protocol knowledge comes from protocol sources only.

---

## Maintenance

This knowledge base is the **permanent asset**. Implementations are disposable.

When updating:

1. Verify change reflects protocol specification
2. Ensure no duplication with existing documents
3. Update cross-links if necessary
4. Document rationale for change

---

## Contributing

When contributing protocol knowledge:

1. Verify information from primary sources
2. Follow document template and style guidelines
3. Ensure concept ownership is respected
4. Link to related documents
5. Include engineering notes for practical context

---

## Quick Reference

| Concept | Owner Document |
|---------|---------------|
| Link Layer | [060-link-layer.md](060-link-layer.md) |
| Transport Layer | [070-transport-layer.md](070-transport-layer.md) |
| Application Layer | [080-application-layer.md](080-application-layer.md) |
| Objects | [090-object-model.md](090-object-model.md) |
| Events | [150-events.md](150-events.md) |
| Security | [280-security.md](280-security.md) |
| Sequence Numbers | [210-sequence-numbers.md](210-sequence-numbers.md) |

---

## Status

This knowledge base is being developed to become the definitive, vendor-neutral reference for DNP3 protocol engineering.
