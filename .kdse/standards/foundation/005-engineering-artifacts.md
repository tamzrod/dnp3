# Engineering Artifacts

## Overview

KDSE defines six canonical artifact types. Each artifact type has defined purpose, stewardship, lifetime, authority, dependencies, and deliverables.

## Artifact Lifecycle

Engineering artifacts evolve through defined lifecycle states. Lifecycle management ensures artifact quality, proper review, and appropriate authority levels.

### Lifecycle States

Artifacts progress through states that communicate readiness and authority:

| State | Meaning | Authority Level | Review Expectation |
|-------|---------|-----------------|-------------------|
| Proposed | Initial suggestion, not yet formal | None | Initial review |
| Experimental | Being explored, not committed | Low | Periodic review |
| Draft | Formal but incomplete | Medium | Active review |
| Reviewed | Reviewed, may need revision | Medium-High | Address feedback |
| Approved | Reviewed and authorized | High | Compliance check |
| Reference | Actively used as standard | Highest | Periodic validation |
| Canonical | Definitive version for domain | Highest | Change control |
| Superseded | Replaced, historical only | Archived | None |
| Deprecated | Not recommended | Archived | None |
| Archived | Retained for traceability | Historical | None |

### Lifecycle Characteristics

- **Authority by State**: States with higher authority grant more confidence. Lower-authority artifacts cannot contradict higher-authority artifacts.
- **State Transitions**: Transitions require appropriate review. Each artifact type defines its lifecycle states in its section below.
- **State vs. Version**: Artifact state is independent of repository version. A version-controlled artifact may be in Draft state or Approved state at different times.

### Why Lifecycle Management Is Necessary

Without explicit lifecycle states:

1. **Quality Gates Absent**: Teams cannot determine when artifacts are ready for use
2. **Review Requirements Unclear**: No formal trigger for review activities
3. **Authority Ambiguous**: Cannot determine artifact's current authority level
4. **Governance Impossible**: Without state, governance cannot enforce requirements
5. **Maintenance Neglected**: No systematic approach to artifact currency

Lifecycle management provides:

1. **Quality Gates**: Clear criteria for artifact progression
2. **Review Triggers**: Defined moments for review activities
3. **Authority Communication**: Immediate understanding of artifact status
4. **Governance Foundation**: State enables enforcement
5. **Maintenance Framework**: Systematic approach to artifact health

## Artifact Type 1: Knowledge

### Purpose

Capture and authorize the knowledge necessary to address a problem space.

### Description

Knowledge artifacts represent authoritative understanding about the problem domain, requirements, constraints, and context. Knowledge artifacts are the highest-authority artifact type in KDSE.

### Steward

Knowledge Steward, as designated by governance

### Lifetime

Persists until explicitly retired. Knowledge may be archived but is never deleted from traceability records.

### Authority

Highest authority in the artifact hierarchy. All other artifacts derive authority from knowledge.

### Dependencies

None (originates the artifact hierarchy)

### Deliverables

- Requirements knowledge
- Non-functional requirements knowledge
- Domain knowledge
- Context knowledge
- Assumption knowledge

### Characteristics

- Authoritative
- Traceable
- Versioned
- Reviewed and approved

### Lifecycle States

Knowledge artifacts follow: Proposed → Draft → Reviewed → Approved → Reference/Canonical → Superseded/Deprecated → Archived

## Artifact Type 2: Architecture

### Purpose

Translate knowledge into structural decisions that guide implementation.

### Description

Architecture artifacts represent the structural decisions derived from knowledge. Architecture artifacts define how the system is structured, what components exist, and how they interact.

### Steward

Architecture Steward, as designated by governance

### Lifetime

Persists for the system lifecycle. May evolve with version control.

### Authority

Derives authority from knowledge. Authorizes implementation.

### Dependencies

- Knowledge artifacts (required)

### Deliverables

- System architecture documentation
- Component architecture documentation
- Interface definitions
- Architecture Decision Records (ADRs)
- Architectural constraints

### Characteristics

- Derived from knowledge
- Authorizes implementation
- Traceable to knowledge
- Versioned

### Lifecycle States

Architecture artifacts follow: Proposed → Draft → Reviewed → Approved → Reference/Canonical → Superseded → Archived

## Artifact Type 3: Architecture Decision Record (ADR)

### Purpose

Document significant architectural decisions with rationale and consequences.

### Description

ADRs capture architectural decisions that merit documentation. An ADR captures what was decided, why it was decided, and what alternatives were considered.

### Steward

Architecture Steward, as designated by governance

### Lifetime

Persists for system lifecycle. Superseded ADRs are archived, not deleted.

### Authority

Derives authority from architecture, which derives authority from knowledge.

### Dependencies

- Architecture artifacts (required)
- Knowledge artifacts (indirect)

### Deliverables

- Decision title and identifier
- Status (proposed, accepted, deprecated, superseded)
- Context and constraints
- Decision
- Rationale
- Consequences (positive, negative, neutral)
- Related decisions

### Characteristics

- Immutable once accepted (superseded decisions are not modified)
- Traceable to knowledge and architecture
- Versioned
- Reviewed and approved

### Lifecycle States

ADR artifacts follow: Proposed → Accepted → Deprecated/Superseded → Archived

## Artifact Type 4: Implementation

### Purpose

Realize architecture through code and related artifacts.

### Description

Implementation artifacts represent the physical realization of the system. Implementation artifacts include code, configuration, scripts, and related files.

### Steward

Implementation Steward, as designated by governance

### Lifetime

Persists through system lifecycle. May be replaced through evolution.

### Authority

Derives authority from architecture.

### Dependencies

- Architecture artifacts (required)
- ADRs (when decisions are relevant)
- Knowledge artifacts (indirect, through architecture)

### Deliverables

- Source code
- Configuration files
- Build scripts
- Deployment artifacts
- Implementation-specific documentation

### Characteristics

- Traceable to architecture
- Subject to verification
- Version controlled
- Must not contradict architecture

### Lifecycle States

Implementation artifacts follow: Draft → Reviewed → Approved → Implemented → Verified → Superseded → Archived

## Artifact Type 5: Verification

### Purpose

Confirm that implementation aligns with architecture and that architecture aligns with knowledge.

### Description

Verification artifacts document the results of verification activities. Verification artifacts confirm that artifacts at lower levels of the hierarchy satisfy requirements at higher levels.

Verification is a first-class knowledge domain in KDSE. The verification domain encompasses not only verification artifacts but also the principles, processes, and criteria that guide verification activities.

### Steward

Verification Steward, as designated by governance

### Lifetime

Persists for audit purposes. Verification artifacts are archived, not deleted.

### Authority

Verification authority derives from knowledge and architecture. Verification does not create new authority but confirms existing authority alignment.

### Dependencies

- Knowledge artifacts (for verification criteria)
- Architecture artifacts (for verification criteria)
- Implementation artifacts (for what is verified)

### Deliverables

- Verification plans
- Test cases
- Test results
- Verification reports
- Gap analysis
- Non-conformance reports

### Characteristics

- Objective
- Traceable to knowledge and architecture
- Reproducible (where applicable)
- Documented and retained

### Lifecycle States

Verification artifacts follow: Draft → Reviewed → Approved → Archived

### Verification Knowledge Domain

Verification as a knowledge domain includes:

**Verification Goals:**
1. Prove implementation conforms to architecture
2. Prove architecture conforms to knowledge
3. Prove knowledge remains internally consistent
4. Identify and report non-conformances

**Verification Evidence:**
- Test results and execution records
- Review comments and sign-offs
- Analysis outputs
- Inspection findings
- Audit reports

**Verification Traceability:**
- Every verification activity traces to authorization
- Verification criteria trace to knowledge artifacts
- Verification results trace to specific requirements
- Non-conformances trace to specific artifacts

**Verification Criteria:**
1. **Completeness**: All requirements verified
2. **Correctness**: Verification produces accurate results
3. **Consistency**: Results are consistent across methods
4. **Reproducibility**: Verification can be repeated
5. **Independence**: Verification is independent of implementation

**Verification Authority:**
- Verification authority derives from knowledge
- Verification stewards authorize verification approaches
- Verification criteria require knowledge authorization
- Verification results report to authorized stakeholders

**Verification Lifecycle:**
Plan → Criteria Derivation → Execution → Documentation → Review → Reporting

## Artifact Type 6: Governance

### Purpose

Establish authority, stewardship, and process for all other artifact types.

### Description

Governance artifacts define the rules by which other artifacts are created, reviewed, approved, changed, and retired. Governance artifacts establish stewardship assignments, authority delegation, and process compliance requirements.

### Steward

Governance Steward, as designated by organizational governance

### Lifetime

Persists for the duration of the engineering initiative. May evolve through defined change processes.

### Authority

Governance artifacts derive authority from organizational governance. Governance authorizes the processes that produce all other artifacts.

### Dependencies

- Organizational governance (external)
- All other artifact types (governance applies to all)

### Deliverables

- Stewardship assignments
- Authority delegations
- Process definitions
- Review criteria
- Approval workflows
- Change management procedures
- Compliance requirements

### Characteristics

- Authoritative over process
- Stable but changeable through defined procedures
- Applied consistently across artifact types

### Lifecycle States

Governance artifacts follow: Draft → Reviewed → Approved → Superseded → Archived

## Artifact Dependency Graph

```
                    ┌─────────────┐
                    │ Governance  │
                    └──────┬──────┘
                           │
           ┌───────────────┼───────────────┐
           │               │               │
           ▼               ▼               ▼
      ┌─────────┐     ┌───────────┐   ┌─────────────┐
      │Knowledge│────▶│Architecture│──▶│Implementation│
      └─────────┘     └─────┬─────┘   └──────┬──────┘
                             │                │
                             ▼                ▼
                         ┌─────────┐   ┌─────────────┐
                         │   ADR   │   │ Verification│
                         └─────────┘   └─────────────┘
```

## Artifact Characteristics Summary

| Artifact | Purpose | Authority Level | Originates |
|----------|---------|-----------------|------------|
| Governance | Process authority | Meta | Organizational |
| Knowledge | Problem understanding | Highest | Stakeholders |
| Architecture | Structural decisions | High | Knowledge |
| ADR | Decision documentation | Medium | Architecture |
| Implementation | Physical realization | Low | Architecture |
| Verification | Alignment confirmation | Applied | All above |

## Engineering Stewardship

KDSE replaces ownership-oriented thinking with stewardship. Knowledge should not be "owned" but "stewarded." Stewardship reflects responsibility without dominion.

### Steward Roles

| Steward | Responsibility |
|---------|----------------|
| Knowledge Steward | Ensures knowledge artifact quality and evolution |
| Architecture Steward | Ensures architectural integrity |
| Implementation Steward | Ensures implementation alignment |
| Verification Steward | Ensures verification rigor |
| Governance Steward | Ensures methodology compliance |

### Stewardship Responsibilities

1. **Custody**: Ensuring artifact accessibility and preservation
2. **Maintenance**: Keeping artifacts current and valid
3. **Quality**: Ensuring artifact meets quality standards
4. **Evolution**: Managing artifact changes through lifecycle
5. **Transfer**: Properly handing off stewardship when needed

### Stewardship Transfer

Stewardship transfer occurs when:
- Primary steward changes roles or leaves
- Artifact scope changes requiring different expertise
- Organizational restructuring
- Strategic reassignment

Transfer requires:
1. Documentation of current state
2. Knowledge transfer to successor
3. Formal acknowledgment by successor
4. Notification to stakeholders

### Scaling Stewardship

**Individual Engineer:**
- Single steward for all artifact types
- Informally documented stewardship

**Small Team (2-10):**
- Designated stewards per artifact type
- Clear responsibility assignments
- Minimal formalization

**Large Organization (10-50+):**
- Multiple stewards with domains
- Formal stewardship agreements
- Escalation paths defined

**Open Source Community:**
- Stewardship by role, not person
- Multiple co-stewards allowed
- Community-based stewardship transfer
- Merit-based stewardship elevation
