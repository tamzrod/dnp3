# Engineering Knowledge

## Purpose

This document defines what constitutes engineering knowledge in KDSE and explains why structured knowledge serves as the authoritative foundation of the methodology.

## What Engineering Knowledge Is Not

Before defining engineering knowledge, it is necessary to distinguish it from related concepts.

### Engineering Knowledge Is Not Information

Information is data that has been processed to be meaningful. Information answers "what" questions.

Engineering knowledge answers "why" questions. Knowledge includes the reasoning, constraints, and trade-offs that inform decisions.

A requirement stating "the system shall respond within 100ms" is information. The knowledge behind this requirement includes why 100ms was chosen, what alternatives were considered, what constraints drove this decision, and what would need to change to alter this threshold.

### Engineering Knowledge Is Not Documentation

Documentation records information. Documentation may or may not capture the reasoning behind decisions.

Engineering knowledge captures understanding. Knowledge artifacts document not only what was decided but why.

A design document that states "we use a microservices architecture" is documentation. Knowledge includes why microservices were chosen over alternatives, what problems this architecture solves, what problems it introduces, and what conditions would make a different architecture preferable.

### Engineering Knowledge Is Not Source Code

Source code is executable specification. Source code describes how a system behaves.

Engineering knowledge describes why the system was built this way. Knowledge provides context that code cannot express.

Code may implement caching for performance reasons. Knowledge explains what performance problem necessitated caching, what alternatives were evaluated, why other approaches were rejected, and what would need to happen for this decision to be revisited.

### Engineering Knowledge Is Not Data

Data is recorded observations about the world. Data describes facts.

Engineering knowledge is validated understanding about the problem domain. Knowledge interprets facts and their implications.

Data might show that users abandon checkout at the shipping step 60% of the time. Knowledge interprets this data as evidence of a problem, proposes hypotheses for why this happens, and identifies what investigation would confirm or refute each hypothesis.

### Engineering Knowledge Is Not Notes

Notes capture thoughts without validation. Notes may be incomplete, inconsistent, or untested.

Engineering knowledge is validated understanding. Knowledge has been reviewed, tested against evidence, and found to be sound.

Meeting notes might record that "performance is a concern." Knowledge transforms this into a specific performance requirement with measurable criteria, documented constraints, and explicit trade-offs against other concerns.

### Engineering Knowledge Is Not Specifications

Specifications describe system behavior. Specifications answer "what" questions.

Engineering knowledge describes why specifications exist. Knowledge provides the foundation that specifications rest upon.

A specification might state "the API shall return XML or JSON based on the Accept header." Knowledge explains why this flexibility was provided, what client requirements drove this decision, and what would need to change for this behavior to be modified.

## What Engineering Knowledge Is

Engineering knowledge is validated understanding about a problem domain that informs engineering decisions. Engineering knowledge has four defining characteristics.

### Characteristic 1: Validated

Knowledge is not mere belief or assumption. Knowledge has been validated against evidence and found to be sound.

Validation methods include:

- Peer review by domain experts
- Testing against real-world scenarios
- Comparison with established theory
- Analysis of historical data
- Prototyping and experimentation

An assumption becomes knowledge when it has been reviewed and confirmed by those with relevant expertise and when evidence supports its validity.

### Characteristic 2: Contextual

Knowledge exists within context. The same statement may be knowledge in one context and assumption in another.

Context includes:

- Problem domain
- Stakeholder perspectives
- Constraints and trade-offs
- Historical decisions and their rationale
- Future considerations and evolution paths

A performance requirement of 100ms might be knowledge in the context of a real-time trading system but assumption in the context of a batch processing system where latency requirements differ.

### Characteristic 3: Traceable

Knowledge can be traced to its sources. The origin of knowledge is documented.

Traceability includes:

- Who validated the knowledge
- What evidence supports it
- What alternatives were considered
- What assumptions underlie it
- What conditions would invalidate it

This traceability allows future engineers to evaluate knowledge validity, understand its foundations, and identify when conditions have changed.

### Characteristic 4: Purposeful

Knowledge exists to inform decisions. Knowledge without purpose is merely information.

Purposefulness means:

- Each knowledge artifact serves a decision context
- Knowledge artifacts explain why decisions were made
- Knowledge artifacts identify what decisions depend on them
- Obsolete knowledge is explicitly marked as such

## What Makes Engineering Knowledge Structured

Not all knowledge qualifies for authority in KDSE. Only structured knowledge can become authoritative.

### Structure Defined

Structure in engineering knowledge refers to completeness, validation, and explicit relationships.

A structured knowledge artifact contains:

1. **The understanding itself**: The core knowledge being captured
2. **Validation evidence**: How this understanding was validated
3. **Source attribution**: Where this understanding came from
4. **Dependencies**: What other knowledge this understanding depends upon
5. **Dependents**: What decisions depend upon this understanding
6. **Lifetime**: When this understanding applies and when it becomes obsolete

### Why Structure Enables Authority

Authority requires consistency. Authority requires understanding. Authority requires accountability.

Unstructured knowledge cannot become authoritative because:

- **Consistency is impossible**: Without defined structure, knowledge artifacts may be incomplete, contradictory, or ambiguous
- **Understanding is uncertain**: Without validation and source attribution, the basis for knowledge is unclear
- **Accountability is absent**: Without explicit stewards and lifetimes, responsibility for knowledge is distributed and untraceable

Structured knowledge enables authority because:

- **Consistency is achievable**: Structure defines what must be present, enabling validation
- **Understanding is documented**: Validation evidence and sources are recorded
- **Accountability is clear**: Owners are assigned, lifetimes are defined, and dependencies are explicit

### Structured Knowledge Versus Documentation

Documentation can exist without structure. A requirements document may capture what users need without explaining why these needs exist or what constraints shaped them.

Structured knowledge requires:

- **Rationale**: Why does this understanding matter?
- **Validation**: How do we know this understanding is correct?
- **Dependencies**: What does this understanding depend upon?
- **Provenance**: Where did this understanding come from?

A requirements document might state that "users need to reset their passwords." Structured knowledge would explain why this requirement exists (user complaints, security audit findings), what constraints shaped it (security policy, user experience research), and what would need to change for this requirement to be revisited (changes to security policy, new user research findings).

## The Authority of Structured Knowledge

In KDSE, only structured knowledge can become authoritative. Authority derives from the properties that structure provides.

### Why Structured Knowledge Is Authoritative

Structured knowledge is authoritative because:

1. **It has been validated**: Authority requires confidence. Confidence requires validation. Structure ensures validation occurs.

2. **It is complete**: Authority requires shared understanding. Incomplete knowledge creates ambiguity. Structure ensures completeness criteria are met.

3. **It is traceable**: Authority requires accountability. Untraceable knowledge cannot be held accountable. Structure ensures traceability is maintained.

4. **It has clear stewardship**: Authority requires responsibility. Distributed responsibility creates diffusion. Structure ensures stewardship is assigned.

### Knowledge Authority and Artifact Hierarchy

Knowledge authority establishes the foundation for all other artifact authority.

```
Structured Knowledge
    │
    │ Authorizes
    ▼
Architecture
    │
    │ Derives from Knowledge
    │ Authorizes
    ▼
Implementation
    │
    │ Realizes
    ▼
Verification
```

Each layer derives its authority from the layer above. Architecture derives authority from knowledge. Implementation derives authority from architecture. Verification confirms that lower layers align with higher layers.

### Knowledge That Cannot Become Authoritative

Not all understanding qualifies as structured knowledge.

Understanding that cannot become authoritative includes:

- **Unvalidated assumptions**: Beliefs without evidence or review
- **Incomplete understanding**: Knowledge without context, dependencies, or validation
- **Unowned knowledge**: Understanding without assigned responsibility
- **Obsolete understanding**: Knowledge that has been superseded but not marked as such

Such understanding may inform decisions but does not carry authority. Decisions made based on non-authoritative understanding do not have the same status as decisions made based on validated, structured knowledge.

## Distinction Summary

| Concept | Primary Question | Authority Level |
|---------|------------------|-----------------|
| Data | What happened? | None |
| Information | What does it mean? | None |
| Notes | What was discussed? | None |
| Documentation | What was decided? | Low |
| Specifications | What should happen? | Medium |
| Source Code | How does it work? | Medium |
| Engineering Knowledge | Why is it this way? | High |
| Structured Knowledge | Why, validated? | Highest |

## Glossary Additions

This document introduces the following terms for glossary inclusion:

### Structured Knowledge

Engineering knowledge that has been validated, documented with dependencies and provenance, assigned stewardship, and defined with explicit lifetime boundaries. Only structured knowledge can carry authority in KDSE.

### Validation

The process of confirming that knowledge accurately represents understanding of the problem domain. Validation involves evidence review, expert consultation, and testing against scenarios.

### Provenance

The origin and history of knowledge. Provenance includes the sources of knowledge, the validation process, and the changes the knowledge has undergone.

### Knowledge Artifact

A structured knowledge unit that has been formally created, reviewed, approved, and assigned to a steward. Knowledge artifacts are the authoritative units of the Knowledge artifact type.
