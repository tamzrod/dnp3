# Traceability

## Purpose

This document defines what traceability means in KDSE, why it exists, what can be traced, and the engineering value it provides.

## What Traceability Is

Traceability is the ability to follow relationships between artifacts. Traceability enables understanding of how artifacts relate to each other and why they exist.

### Traceability Is Not Documentation

Documentation records information. Traceability records relationships.

A document may state that "architecture was derived from requirements." Traceability confirms this by showing which specific requirements informed which specific architectural decisions.

Documentation can be incomplete or inaccurate. Traceability requires explicit relationships that can be verified.

### Traceability Is Not Lineage

Lineage tracks where things came from. Traceability tracks why things exist.

Lineage might show that code came from a specific branch. Traceability shows that code implements a specific architectural decision that was made for specific reasons based on specific knowledge.

Lineage is about origin. Traceability is about justification.

### Traceability Is Not Dependency

Dependency tracks what things require. Traceability tracks why things exist.

A component might depend on another component. The dependency exists because the first component uses services provided by the second.

Traceability might show that the decision to separate these components was made because of specific knowledge about team structure and deployment requirements.

Dependency is structural. Traceability is intentional.

## Why Traceability Exists

Traceability exists because engineering decisions require justification and because understanding requires context.

### Reason 1: Justification

Every decision requires justification. Traceability provides justification by showing the chain of reasoning.

When an engineer asks "why is the system structured this way?", traceability provides the answer by showing:

- What knowledge authorized the structure
- What decisions were made based on that knowledge
- What alternatives were considered

Without traceability, this question cannot be answered. Without justification, decisions cannot be evaluated.

### Reason 2: Impact Analysis

Changes have impacts. Traceability enables impact analysis by showing what depends on what.

When knowledge changes, traceability shows:

- Which architectural decisions depend on this knowledge
- Which implementation artifacts implement these decisions
- Which verification criteria relate to this knowledge

Without traceability, impact analysis is guesswork. Engineers cannot know what will break when requirements change.

### Reason 3: Audit

Engineering work requires audit. Traceability enables audit by providing a complete record.

Auditors can verify:

- That decisions were properly authorized
- That implementation follows architecture
- That architecture traces to knowledge

Without traceability, audit requires reconstructing relationships from artifacts. This reconstruction may be incomplete or incorrect.

### Reason 4: Maintenance

Systems require maintenance. Traceability enables maintenance by providing context.

Maintenance engineers can:

- Understand why the system is structured as it is
- Identify what changes might affect specific components
- Verify that proposed changes are consistent with original intent

Without traceability, maintenance is archaeology. Engineers must infer intent from implementation, often incorrectly.

### Reason 5: Communication

Engineering requires communication. Traceability enables communication by providing shared understanding.

Engineers can:

- Reference specific artifacts rather than vague concepts
- Understand how their work relates to others' work
- Discuss decisions with shared context

Without traceability, communication relies on undocumented context that may be lost or misunderstood.

## What Can Be Traced

In KDSE, specific relationships can and should be traced.

### Traceability Matrix

| From | To | Relationship | Required |
|------|-----|-------------|----------|
| Knowledge | Architecture | Authorizes | Yes |
| Knowledge | Verification Criteria | Derives | Yes |
| Architecture | ADR | Documents | Yes |
| Architecture | Implementation | Guides | Yes |
| ADR | Implementation | Justifies | Yes |
| Implementation | Verification | Subject to | Yes |
| Verification | Knowledge | Confirms alignment | Yes |

### Forward Traceability

Forward traceability follows relationships from higher-authority to lower-authority artifacts.

```
Knowledge
    │
    ├──► Architecture (authorizes)
    │
    ├──► Verification Criteria (derives)
    │
    ▼
Architecture
    │
    ├──► ADR (documents)
    │
    └──► Implementation (guides)
    │
    ▼
Implementation
    │
    └──► Verification (subject to)
```

Forward traceability confirms that lower artifacts derive from higher artifacts.

### Backward Traceability

Backward traceability follows relationships from lower-authority to higher-authority artifacts.

```
Implementation
    │
    └──◄ Architecture (guides)
    │
◄──┘
Architecture
    │
    ├──◄ Knowledge (authorizes)
    │
    ├──◄ ADR (documents)
    │
    └──◄ Verification Criteria (derives)
    │
◄──┘
Knowledge
```

Backward traceability confirms that higher artifacts justify lower artifacts.

### Traceability Requirements

Each artifact must maintain:

1. **Upward links**: What artifacts this derives from
2. **Downward links**: What artifacts derive from this
3. **Basis documentation**: Why these relationships exist

## Traceability Depth

Not all relationships require the same traceability depth.

### Deep Traceability

Deep traceability traces every artifact to every relevant artifact.

**Characteristics**:

- Every component traces to architectural decisions
- Every decision traces to specific knowledge
- Every verification criterion traces to requirements

**When Required**:

- Regulated industries
- Safety-critical systems
- Systems with high change rates
- Systems with complex dependencies

### Moderate Traceability

Moderate traceability traces key artifacts while accepting lighter traceability for details.

**Characteristics**:

- Major architectural decisions trace to knowledge
- Critical components trace to architecture
- Routine implementation follows patterns without explicit traceability

**When Required**:

- Typical business systems
- Systems with moderate complexity
- Systems with established patterns

### Light Traceability

Light traceability traces only when necessary for understanding.

**Characteristics**:

- Decisions are documented but links are implicit
- Architecture traces to major knowledge
- Implementation follows architecture without explicit links

**When Required**:

- Small projects
- Rapid development
- Early-stage exploration

## Traceability Challenges

Maintaining traceability presents challenges.

### Challenge 1: Traceability Overhead

Creating and maintaining traceability requires effort.

**Resolution**: Focus traceability on decisions that matter. Not every line of code requires traceability to requirements. Focus on architectural decisions, significant implementation choices, and non-functional requirements.

### Challenge 2: Traceability Staleness

Artifacts change. Traceability links may become stale.

**Resolution**: Include traceability maintenance in artifact change processes. When knowledge changes, review affected architecture. When architecture changes, review affected implementation.

### Challenge 3: Traceability Granularity

Determining appropriate granularity is difficult.

**Resolution**: Start with coarse traceability. Trace at artifact boundaries rather than element boundaries. Refine as needed when coarser traceability proves insufficient.

### Challenge 4: Traceability Tooling

Maintaining traceability without tooling is burdensome.

**Resolution**: Use tools appropriate to scale. Small projects may use manual tracking. Large projects require tool support. Choose tools that integrate with existing workflows.

## Traceability Verification

Traceability must be verified to be meaningful.

### Verification Activities

1. **Completeness check**: Confirm that all required links exist
2. **Correctness check**: Confirm that links represent actual relationships
3. **Consistency check**: Confirm that links are consistent across artifacts
4. **Currency check**: Confirm that links reflect current artifact state

### Verification Frequency

Traceability should be verified:

- When artifacts are created
- When artifacts change
- Before release
- During audit

### Verification Responsibilities

Traceability verification is everyone's responsibility:

- **Knowledge stewards**: Verify that knowledge authorizes appropriate architecture
- **Architecture stewards**: Verify that architecture traces to knowledge
- **Implementation stewards**: Verify that implementation traces to architecture
- **Verification stewards**: Verify that verification criteria trace to knowledge

## Traceability and Authority

Traceability is the mechanism by which authority is maintained.

### Authority Without Traceability

Authority without traceability is assertion, not justification.

Claiming that "architecture is authoritative" is meaningless if the architecture's basis cannot be demonstrated.

### Traceability Without Authority

Traceability without authority is documentation, not justification.

Showing that "architecture traces to requirements" is insufficient if the requirements are themselves unjustified.

### Traceability Enables Authority

Traceability enables authority by demonstrating justification.

When architecture traces to validated knowledge, architecture carries knowledge's authority. When implementation traces to architecture, implementation carries architecture's authority.

```
Knowledge (authority)
    │
    │ Traceability demonstrates
    ▼
Architecture (inherits authority)
    │
    │ Traceability demonstrates
    ▼
Implementation (inherits authority)
```

## Traceability Value Summary

Traceability provides value at each lifecycle stage.

### During Development

- Decisions are justified
- Architecture is consistent
- Implementation follows direction
- Changes are understood

### During Maintenance

- Impact analysis is possible
- Context is available
- Decisions can be evaluated
- Knowledge can be validated

### During Audit

- Compliance can be demonstrated
- Decisions can be reviewed
- Coverage can be verified
- Issues can be identified

## What Traceability Is Not

Traceability is often misunderstood.

### Traceability Is Not Completeness

Traceability confirms relationships exist. Traceability does not confirm that relationships are complete.

A system may have complete traceability yet still have missing requirements or unaddressed concerns.

### Traceability Is Not Correctness

Traceability confirms that artifacts relate as claimed. Traceability does not confirm that artifacts are correct.

Architecture may trace to knowledge yet still be wrong. Implementation may trace to architecture yet still be flawed.

### Traceability Is Not Quality

Traceability confirms that relationships are documented. Traceability does not confirm that relationships are good.

Well-traced decisions may still be poor decisions. Poor decisions with good traceability are still poor decisions.

## Glossary Additions

This document introduces the following terms for glossary inclusion:

### Traceability

The ability to follow relationships between artifacts. Traceability enables understanding of how artifacts relate to each other and why they exist.

### Forward Traceability

Traceability that follows relationships from higher-authority to lower-authority artifacts. Forward traceability confirms that lower artifacts derive from higher artifacts.

### Backward Traceability

Traceability that follows relationships from lower-authority to higher-authority artifacts. Backward traceability confirms that higher artifacts justify lower artifacts.

### Traceability Link

An explicit relationship between two artifacts. Traceability links form the chains that traceability follows.

### Traceability Verification

The process of confirming that traceability links are complete, correct, consistent, and current.
