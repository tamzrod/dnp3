# Core Principles

## Principles

These principles are timeless. They do not change with technology, domain, or organizational context.

### Principle 1: Knowledge Precedes Architecture

Architecture derives from knowledge. Architecture decisions must be traceable to specific knowledge artifacts. Architecture that cannot be traced to knowledge is not compliant with KDSE.

**Rationale**: Architecture without knowledge basis is opinion. Opinion is insufficient for engineering decisions that affect system behavior, maintainability, and evolution.

### Principle 2: Architecture Precedes Implementation

Implementation derives from architecture. Implementation must conform to architecture. Implementation that contradicts architecture represents architecture drift.

**Rationale**: Implementation without architectural guidance leads to inconsistent systems. Each implementation decision must align with the architectural direction established by knowledge-derived architecture.

### Principle 3: Implementation Precedes Verification

Verification derives from implementation and the knowledge that authorized the implementation. Verification confirms that implementation satisfies knowledge-derived requirements.

**Rationale**: Verification without implementation has nothing to verify. Verification without knowledge basis cannot confirm alignment with intended outcomes.

### Principle 4: Knowledge Is the Longest-Lived Artifact

Knowledge outlives the systems it describes. When systems are replaced, knowledge persists as institutional memory.

**Rationale**: Systems are replaced. The knowledge that guided their creation informs future systems. Treating knowledge as ephemeral destroys institutional memory.

### Principle 5: Engineering Decisions Must Be Traceable

Every architectural decision, implementation choice, and verification criterion must trace to authorized knowledge artifacts. Traceability enables impact analysis, audit, and evolution.

**Rationale**: Untraceable decisions cannot be understood, evaluated, or changed systematically. Traceability is the mechanism by which knowledge maintains authority.

### Principle 6: Code Realizes Knowledge

Code is the physical manifestation of knowledge. Code that cannot be traced to knowledge does not represent authorized engineering decisions.

**Rationale**: Code exists to realize knowledge. Code without knowledge basis is orphaned artifact with unknown purpose and constraints.

### Principle 7: Knowledge Is Language-Independent

Knowledge captures intent, not implementation. The same knowledge may realize in multiple languages, platforms, or technologies.

**Rationale**: Technologies change. Languages become obsolete. Knowledge persists by transcending implementation specifics.

### Principle 8: Authority Flows Downward

Lower artifact types cannot contradict higher artifact types. Implementation cannot contradict architecture. Architecture cannot contradict knowledge.

**Rationale**: Authority hierarchy maintains system coherence. When lower artifacts contradict higher artifacts, traceability fails and the system degrades.

### Principle 9: Verification Confirms Alignment

Verification artifacts confirm that implementation aligns with architecture, and that architecture aligns with knowledge. Verification that cannot trace to knowledge has no authoritative basis.

**Rationale**: Verification without knowledge basis verifies nothing of engineering significance. Verification must confirm alignment with intended outcomes.

### Principle 10: Evolution Maintains Authority

Changes must flow through proper channels and maintain authority hierarchy. Changes to lower artifacts require understanding of higher artifacts. Changes originate from changed understanding, propagate through proper channels, and realize through authorized modification.

**Rationale**: Bottom-up change requests without proper channels produce drift. Authority remains with higher layers; lower layers request, higher layers decide. Evolution is planned change, not drift.

## Principles Are Not Practices

These principles are not practices. They do not prescribe:

- How to capture knowledge
- How to document architecture
- How to write code
- How to perform verification
- What tools to use

Practices are derived from principles for specific contexts. Principles guide practice selection; they do not replace it.
