# Engineering Model

## Lifecycle Overview

KDSE defines a lifecycle in which artifacts progress through defined stages. Each stage produces artifacts that serve as input to subsequent stages.

```
Knowledge
    ↓
Architecture
    ↓
Implementation
    ↓
Verification
    ↓
Evolution
```

## Stage 1: Knowledge

### Purpose

Capture, structure, and authorize the knowledge necessary to address a problem space.

### Input

- Problem statements
- Constraints
- Domain knowledge
- Existing knowledge artifacts (for continuation)

### Output

- Knowledge artifacts establishing requirements
- Knowledge artifacts establishing non-functional requirements
- Knowledge artifacts establishing context and assumptions

### Activities

- Elicit knowledge from stakeholders
- Structure knowledge in authorized formats
- Review and validate knowledge artifacts
- Authorize knowledge artifacts

### Why This Stage Comes First

Knowledge must precede architecture because architecture serves knowledge. Architecture decisions made without knowledge basis are arbitrary. By establishing knowledge artifacts first, subsequent stages have authoritative grounding.

## Stage 2: Architecture

### Purpose

Translate knowledge into structural decisions that guide implementation.

### Input

- Authorized knowledge artifacts
- Previously established architecture (for continuation)

### Output

- Architecture artifacts defining structure
- Architecture artifacts defining constraints
- Architecture Decision Records documenting decisions

### Activities

- Derive architecture from knowledge
- Document architectural approaches
- Record decisions with rationale
- Review architecture against knowledge
- Authorize architecture artifacts

### Why This Stage Comes Second

Architecture bridges knowledge and implementation. Architecture translates abstract knowledge into concrete structural decisions. Implementation without architecture guidance produces inconsistent systems.

## Stage 3: Implementation

### Purpose

Realize architecture through code that traces to authorized artifacts.

### Input

- Authorized architecture artifacts
- Architecture Decision Records
- Previously established implementation (for continuation)

### Output

- Implementation artifacts (code, configuration, data)
- Implementation-specific documentation
- Traceability records linking implementation to architecture

### Activities

- Implement according to architecture
- Maintain traceability to architecture
- Resolve implementation questions against architecture
- Document implementation decisions not requiring architecture change

### Why This Stage Comes Third

Implementation realizes architecture. Implementation decisions must align with architectural direction. Implementation that contradicts architecture represents drift requiring correction.

## Stage 4: Verification

### Purpose

Confirm that implementation aligns with architecture and that architecture aligns with knowledge.

### Input

- Authorized knowledge artifacts
- Authorized architecture artifacts
- Implementation artifacts

### Output

- Verification artifacts confirming alignment
- Reports documenting verification results
- Identified gaps requiring correction

### Activities

- Derive verification criteria from knowledge
- Execute verification against implementation
- Document verification results
- Identify and report misalignments

### Why This Stage Comes Fourth

Verification confirms alignment. Without prior stages producing artifacts, verification has no basis. Verification cannot confirm alignment that does not exist.

## Stage 5: Evolution

### Purpose

Manage changes to artifacts over time while maintaining traceability and authority.

### Input

- All prior stage artifacts
- Change requests
- Feedback from operation

### Output

- Updated artifacts reflecting changes
- Records of artifact evolution
- Maintained traceability chains

### Activities

- Evaluate change requests against artifact hierarchy
- Propagate changes through stages as necessary
- Maintain artifact versioning
- Retire obsolete artifacts
- Archive knowledge for future reference

### Why This Stage Is Continuous

Software systems evolve. Artifacts must evolve with them. Evolution is not a terminal stage but a continuous process that may trigger re-entry to any prior stage.

## Stage Relationships

```
        ┌───────────────────────────────────────┐
        │                                       │
        │           Evolution                   │
        │                                       │
        └───────────────────────────────────────┘
                        │
        ┌───────────────┼───────────────┐
        │               │               │
        ▼               ▼               ▼
    Knowledge      Architecture   Implementation
        │               │               │
        │               │               │
        └───────────────┴───────────────┘
                        │
                        ▼
                Verification
```

## Derivation Principle

Each stage derives from its predecessor:

- Architecture derives from Knowledge
- Implementation derives from Architecture
- Verification derives from Implementation (and confirms alignment with Knowledge and Architecture)

Derivation is not optional. Artifacts produced in later stages must trace to authorized artifacts from earlier stages.

## Authority Principle

Authority flows downward:

- Knowledge authorizes Architecture
- Architecture authorizes Implementation
- Implementation is subject to Verification

Lower stages cannot contradict higher stages. Violations represent methodology non-compliance.
