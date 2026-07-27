# KDE Laboratory Standard Operating Procedures

**Document ID**: SOP-001  
**Title**: Laboratory Standard Operating Procedures  
**Version**: 1.0.0  
**Status**: APPROVED  
**Effective Date**: 2026-07-21  
**Owner**: KDE Governance  
**Supersedes**: None  

---

## Purpose

This document defines the Standard Operating Procedures (SOP) for all investigations conducted within the KDE Laboratory. It serves as the authoritative specification describing how every investigation is conducted, independent of the active Engine.

---

## Scope

These procedures apply to:
- All investigations conducted within the KDE Laboratory
- All experiments performed during investigations
- All evidence collected and synthesized
- All knowledge recommendations generated
- All research recommendations produced

---

## Architecture Hierarchy

The KDE Laboratory follows this governance hierarchy:

```
Governance
    │
    ▼
Laboratory Standard Operating Procedures (SOP)
    │
    ▼
Runtime
    │
    ▼
Engine (Exactly One Active)
    │
    ▼
Investigation
    │
    ▼
Experiments
    │
    ▼
Evidence
    │
    ▼
Knowledge Recommendation
    │
    ▼
Governance
```

### Responsibilities by Layer

| Layer | Responsibility | Defines |
|-------|---------------|---------|
| **Governance** | Ownership, approval | What is approved |
| **SOP** | Procedures, standards | What shall happen |
| **Runtime** | Execution, support | How procedures execute |
| **Engine** | Reasoning | How to think |
| **Investigation** | Research | What is studied |

---

## SOP-001: Investigation Lifecycle

### Purpose

Define the complete lifecycle of an investigation from research question to closure.

### Lifecycle Stages

```
┌─────────────────────────────────────────────────────────────────┐
│                     INVESTIGATION LIFECYCLE                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐                                           │
│  │ Research        │                                           │
│  │ Question        │ ◄── Human or Governance defined            │
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │ Investigation   │                                           │
│  │ Creation        │ ◄── Assign ID, scope, objectives          │
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │ Planning        │ ◄── Define methodology, experiments       │
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │ Execution       │ ◄── Conduct experiments, collect evidence  │
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │ Analysis        │ ◄── Synthesize evidence, draw conclusions  │
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │ Documentation   │ ◄── Complete required documents            │
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │ Review          │ ◄── Self-review, peer review              │
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │ Closure         │ ◄── Governed approval required            │
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │ Recommendations │ ◄── Knowledge and research recs           │
│  └─────────────────┘                                           │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Stage Definitions

#### 1. Research Question
- **Input**: Question from human, governance, or derived from prior investigation
- **Output**: Documented research question with scope definition
- **Responsible**: Requestor or Governance

#### 2. Investigation Creation
- **Input**: Research question
- **Output**: Investigation ID assigned, scope documented
- **Responsible**: Runtime (creates directory structure)

#### 3. Planning
- **Input**: Research question
- **Output**: Investigation plan with methodology and experiment definitions
- **Responsible**: Engine

#### 4. Execution
- **Input**: Investigation plan
- **Output**: Raw evidence collected
- **Responsible**: Engine with Runtime support

#### 5. Analysis
- **Input**: Raw evidence
- **Output**: Synthesized findings
- **Responsible**: Engine

#### 6. Documentation
- **Input**: Synthesized findings
- **Output**: Complete investigation documents
- **Responsible**: Engine

#### 7. Review
- **Input**: Complete documentation
- **Output**: Reviewed and validated documentation
- **Responsible**: Engine (self-review), then Human (approval)

#### 8. Closure
- **Input**: Reviewed documentation with approval
- **Output**: Investigation closed, archived
- **Responsible**: Governance (approval)

#### 9. Recommendations
- **Input**: Closed investigation
- **Output**: Knowledge recommendations, research recommendations
- **Responsible**: Engine (recommends), Governance (approves)

### Completion Requirements

An investigation may be closed when ALL of the following are satisfied:

| Requirement | Verification |
|-------------|--------------|
| Research question answered | Conclusion documented with evidence |
| All planned experiments completed | Experiment records present |
| Evidence archived | Evidence artifacts in investigation directory |
| Self-review performed | Review documented |
| Lessons learned recorded | Lessons documented |
| Recommendations prepared | Knowledge and research recommendations documented |
| Human approval obtained | Governance approval recorded |

---

## SOP-002: Experiment Lifecycle

### Purpose

Define the lifecycle of individual experiments within an investigation.

### Relationship to Investigation

```
Investigation
    │
    ├── Experiment 1
    │       ├── Purpose
    │       ├── Scope
    │       ├── Inputs
    │       ├── Execution
    │       ├── Outputs
    │       └── Results
    │
    ├── Experiment 2
    │
    └── Experiment N
```

### Clarification: Experiments vs Investigations

| Aspect | Investigation | Experiment |
|--------|---------------|------------|
| **Purpose** | Answer research question | Test specific hypothesis |
| **Scope** | Broad, multi-objective | Narrow, focused |
| **Duration** | Days to weeks | Minutes to hours |
| **Output** | Conclusions, recommendations | Evidence, results |
| **Hypothesis** | Research question | Testable claim |

### Experiment Lifecycle

```
┌─────────────────────────────────────────────────────────────────┐
│                     EXPERIMENT LIFECYCLE                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐                                           │
│  │ Hypothesis      │ ◄── Testable claim to validate             │
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │ Design          │ ◄── Define methodology, controls           │
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │ Execution       │ ◄── Run experiment, collect data          │
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │ Analysis        │ ◄── Analyze results, compare to hypothesis│
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │ Documentation   │ ◄── Record all findings                    │
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │ Integration     │ ◄── Integrate with other experiments       │
│  └─────────────────┘                                           │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Experiment Definition

Every experiment SHALL include:

#### Purpose
Clear statement of what the experiment tests or measures.

#### Scope
Boundaries of the experiment—what is included and excluded.

#### Inputs
- Hypothesis being tested
- Data sources
- Tools and resources required
- Constraints and assumptions

#### Success Criteria
Objective conditions that, when met, indicate the experiment succeeded.

#### Failure Criteria
Objective conditions that, when met, indicate the experiment failed.

#### Completion Requirements
Minimum requirements for the experiment to be considered complete.

### Experiment Documentation

Every experiment SHALL produce:

| Document | Required | Description |
|----------|----------|-------------|
| Hypothesis | YES | Testable claim |
| Design | YES | Methodology |
| Execution Log | YES | What was actually done |
| Raw Data | YES | Unprocessed results |
| Analysis | YES | Interpretation of results |
| Conclusion | YES | Support/refute hypothesis |
| Integration Notes | YES | How results inform investigation |

### Success vs Failure

**Experiments do not "fail" in the negative sense.**

An experiment that fails to support its hypothesis is NOT a failure—it is a valid result. Experiments fail only when:

| Failure Mode | Definition |
|-------------|------------|
| Execution Failure | Experiment could not be completed |
| Data Quality Failure | Data collected is insufficient or invalid |
| Scope Failure | Experiment did not test what it intended |

---

## SOP-003: Documentation Standards

### Purpose

Define mandatory documentation for all investigation artifacts.

### Required Documents by Artifact Type

#### Investigation Documents

| Document | Purpose | Required |
|----------|---------|----------|
| Investigation Plan | Define scope and methodology | YES |
| Index | Quick reference to all artifacts | YES |
| Status Reports | Track progress | YES |
| Lessons Learned | Capture insights | YES |
| Conclusion | Final findings | YES |
| Investigation Report | Complete narrative | YES |

#### Experiment Documents

| Document | Purpose | Required |
|----------|---------|----------|
| Hypothesis | Testable claim | YES |
| Design | Methodology | YES |
| Execution Log | What was done | YES |
| Raw Data | Unprocessed results | YES |
| Analysis | Interpretation | YES |
| Conclusion | Hypothesis verdict | YES |

#### Evidence Documents

| Document | Purpose | Required |
|----------|---------|----------|
| Evidence Record | What was collected | YES |
| Evidence Chain | How it was obtained | YES |
| Evidence Analysis | What it means | YES |

#### Recommendation Documents

| Document | Purpose | Required |
|----------|---------|----------|
| Knowledge Recommendation | Promote evidence to knowledge | YES (if applicable) |
| Research Recommendation | Recommend future work | YES (if applicable) |

### Document Metadata

Every document SHALL include:

| Field | Description |
|-------|-------------|
| Document ID | Unique identifier |
| Title | Descriptive title |
| Version | Semantic version |
| Status | Draft, Review, Approved, Archived |
| Created | ISO-8601 timestamp |
| Modified | ISO-8601 timestamp |
| Author | Creator identity |
| Investigation ID | Parent investigation (if applicable) |

### Document Naming Convention

```
{PREFIX}-{NUMBER}-{SHORT-TITLE}.md
```

Examples:
- `INV-015-investigation.md`
- `INV-015-index.md`
- `EXP-001-hypothesis.md`
- `EVD-001-runtime-trace.md`

### Purpose Documentation

Every document SHALL state its purpose within the first section:

```markdown
## Purpose

This document [verb] [object].

Example:
## Purpose

This document defines the methodology for testing the hypothesis that 
knowledge retrieval improves investigation quality.
```

---

## SOP-004: Evidence Standards

### Purpose

Define acceptable forms of evidence and explicitly prohibit unsupported conclusions.

### Evidence Definition

Evidence is observable, verifiable information used to support or refute claims.

### Acceptable Evidence Types

| Type | Description | Examples |
|------|-------------|----------|
| **Runtime Traces** | Observable system behavior | Logs, metrics, measurements |
| **Experimental Results** | Controlled test outcomes | Measurements, comparisons |
| **Source References** | External documentation | Documentation, specifications |
| **Generated Artifacts** | Outputs from investigation | Code, diagrams, configurations |
| **Comparative Analysis** | Side-by-side evaluation | A/B results, benchmarks |
| **Statistical Validation** | Mathematically verified | p-values, confidence intervals |
| **Observable Behavior** | Directly witnessed actions | User interactions, system responses |

### Evidence Hierarchy

| Level | Type | Reliability |
|-------|------|-------------|
| 1 | Experimental | Highest reliability |
| 2 | Observational | High reliability |
| 3 | Documented | Medium reliability |
| 4 | Referenced | Requires verification |
| 5 | Inferred | Requires strong support |

### Evidence Requirements by Investigation Type

| Investigation Type | Minimum Evidence |
|-------------------|-----------------|
| Root Cause Analysis | Observable behavior + source trace |
| Architecture Design | Comparative analysis + tradeoffs |
| Validation | Experimental results |
| Recommendation | Evidence synthesis |

### Prohibited Conclusions

The following conclusions are NOT permitted without evidence:

| Prohibited Statement | Requirement |
|---------------------|-------------|
| "The system works correctly" | Requires test results |
| "Knowledge was used" | Requires retrieval trace |
| "Performance improved" | Requires measurements |
| "Quality is acceptable" | Requires defined criteria |
| "The approach is better" | Requires comparative evidence |

### Evidence Chain

Evidence SHALL maintain a chain of custody:

```
Source → Collection → Storage → Analysis → Presentation
```

| Stage | Requirement |
|-------|-------------|
| Source | Document origin |
| Collection | Timestamp, method |
| Storage | Location, access |
| Analysis | Process documented |
| Presentation | Context preserved |

### Evidence Citation Format

All evidence SHALL be cited using this format:

```
[EVD-{NUMBER}] {Type}: {Description}
Link: {Reference}
Retrieved: {Timestamp}
```

Example:
```
[EVD-001] Runtime Trace: Git status during session
Link: investigation/INV-015/logs/session.log:45
Retrieved: 2026-07-21T14:30:00Z
```

---

## SOP-005: Knowledge Retrieval Policy

### Purpose

Define the laboratory policy governing knowledge retrieval during investigations.

### Policy Statement

**The laboratory SHALL first determine whether previous knowledge is relevant before conducting investigations.**

The laboratory does NOT require knowledge retrieval for every task. Retrieval is context-dependent.

### Retrieval Decision Framework

```
┌─────────────────────────────────────────────────────────────────┐
│                 KNOWLEDGE RETRIEVAL DECISION                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Is this a continuation investigation?                          │
│           │                                                     │
│      YES ─┴─ NO                                                │
│       │         │                                               │
│       ▼         ▼                                               │
│  Retrieval   Is this similar to historical work?               │
│  REQUIRED        │                                             │
│                 YES ─┴─ NO                                      │
│                  │         │                                    │
│                  ▼         ▼                                    │
│            Retrieval   Is this novel research?                   │
│            RECOMMENDED        │                                  │
│                              │                                  │
│                         YES ─┴─ NO                              │
│                          │         │                            │
│                          ▼         ▼                            │
│                    Retrieval   Is this routine work?             │
│                    OPTIONAL          │                          │
│                                      │                          │
│                                 YES ─┴─ NO                      │
│                                  │         │                     │
│                                  ▼         ▼                     │
│                            Minimal      Full                     │
│                            Retrieval    Retrieval                │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Retrieval Context Matrix

| Context | Retrieval | Documentation |
|---------|-----------|---------------|
| **Continuation** | REQUIRED | List prior knowledge referenced |
| **Similar historical** | RECOMMENDED | List similar investigations reviewed |
| **Novel research** | OPTIONAL | Document if applicable |
| **Routine engineering** | MINIMAL | Document significant knowledge used |
| **Complex engineering** | FULL | Document all relevant patterns |

### Retrieval Documentation Requirements

Every investigation SHALL document:

| Field | Description |
|-------|-------------|
| Relevant knowledge | List of knowledge artifacts reviewed |
| Retrieval decision | Why retrieval was or was not performed |
| Applied knowledge | Specific knowledge applied |
| Ignored knowledge | Relevant knowledge not applied and why |

### Retrieval Execution

The Runtime SHALL:
1. Load knowledge directory at investigation start
2. Provide retrieval capability on demand
3. Log retrieval events for measurement
4. Execute retrieval policy defined here

The Engine SHALL:
1. Determine whether retrieval is relevant
2. Request knowledge when needed
3. Document retrieval decisions
4. Apply retrieved knowledge appropriately

### Retrieval Measurement

The following SHALL be measured:

| Metric | Purpose |
|--------|---------|
| Retrieval rate | Artifacts accessed / artifacts available |
| Application rate | Artifacts applied / artifacts accessed |
| Coverage rate | Relevant artifacts / total artifacts |
| Decision accuracy | Appropriate retrieval decisions / total decisions |

---

## SOP-006: Knowledge Promotion

### Purpose

Define how validated evidence becomes KDE Knowledge.

### Promotion Hierarchy

```
Evidence → Investigation → Knowledge Recommendation → Governance → Knowledge
```

### Clarification of Roles

| Artifact | Created By | Approved By |
|----------|------------|-------------|
| **Evidence** | Experiment | N/A |
| **Investigation** | Engine | Human |
| **Knowledge Recommendation** | Engine | Governance |
| **Knowledge** | Governance | N/A |

### Promotion Criteria

Evidence may be promoted to Knowledge when ALL criteria are met:

| Criterion | Requirement |
|-----------|--------------|
| **Validation** | Evidence has been experimentally validated |
| **Documentation** | Complete evidence chain documented |
| **Synthesis** | Evidence synthesized into generalizable insight |
| **Applicability** | Scope and conditions clearly defined |
| **Peer Review** | At least one peer has reviewed |
| **Human Approval** | Governance has approved |

### Promotion Process

```
┌─────────────────────────────────────────────────────────────────┐
│                   KNOWLEDGE PROMOTION PROCESS                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐                                           │
│  │ Evidence       │ ◄── Validated experimental results        │
│  │ Identification │                                           │
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │ Synthesis       │ ◄── Extract generalizable insight          │
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │ Recommendation  │ ◄── Prepare knowledge recommendation      │
│  │ Creation        │                                           │
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │ Peer Review     │ ◄── External review                       │
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │ Governance      │ ◄── Approval decision                     │
│  │ Approval       │                                           │
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │ Promotion      │ ◄── Move to knowledge/                    │
│  └─────────────────┘                                           │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Knowledge Recommendation Document

Every knowledge recommendation SHALL include:

| Section | Description |
|---------|-------------|
| **Statement** | Precise, testable claim |
| **Evidence** | Links to supporting evidence |
| **Applicability** | When this knowledge applies |
| **Limitations** | When this knowledge fails |
| **Confidence** | Assessment of reliability |
| **Alternatives** | Contrasting approaches |

### Role Clarification

**Experiments NEVER create knowledge.**

Experiments produce evidence. Evidence does not automatically become knowledge. Only through the investigation process and governance approval can evidence be promoted.

**Investigations recommend knowledge.**

Investigations synthesize evidence into insights and prepare knowledge recommendations. They do not have authority to promote knowledge.

**Governance approves knowledge.**

Governance reviews knowledge recommendations and makes final promotion decisions. Only Governance can move artifacts into the knowledge directory.

---

## SOP-007: Research Recommendations

### Purpose

Define how investigations and experiments may recommend future research.

### Research Recommendation Definition

A research recommendation is a proposal for future investigation, not a directive.

### Recommendation Authority

| Artifact | May Recommend Research | May Initiate Research |
|----------|----------------------|----------------------|
| Experiment | YES | NO |
| Investigation | YES | NO |
| Governance | YES | YES |

### Recommendation Contents

Every research recommendation SHALL include:

| Section | Description |
|---------|-------------|
| **Rationale** | Why this research is needed |
| **Scope** | What the research should cover |
| **Priority** | Urgency assessment |
| **Dependencies** | Prior work required |
| **Expected Outcome** | What success looks like |

### Approval Process

```
Research Recommendation
         │
         ▼
Governance Review
         │
         ├── Approved ──► Investigation Created
         │
         └── Rejected ──► Not Implemented
```

### Research Recommendation Document

Every research recommendation SHALL be documented using this format:

```markdown
# Research Recommendation

**Recommendation ID**: REC-{NUMBER}
**Source**: {Investigation or Experiment ID}
**Date**: {ISO-8601}

## Rationale

[Why this research is needed]

## Proposed Scope

[What the research should cover]

## Priority

[Critical / High / Medium / Low]

## Dependencies

[Any prerequisite investigations]

## Expected Outcome

[What success looks like]

## Governance Decision

**Status**: [PENDING / APPROVED / REJECTED]
**Approved By**: 
**Date**:
```

### Automatic Initiation Prohibition

Neither experiments nor investigations may automatically create new investigations. All new investigations require Governance approval.

---

## SOP-008: Investigation Completion

### Purpose

Define the conditions required before an investigation may be closed.

### Completion Requirements

An investigation is COMPLETE when ALL of the following requirements are satisfied:

#### Research Requirements

| Requirement | Verification | Evidence |
|-------------|--------------|----------|
| Research question answered | Conclusion present | Conclusion document |
| All hypotheses tested | Experiment records | Experiment logs |
| Evidence archived | All evidence stored | Evidence directory |
| Recommendations prepared | Documents present | Recommendation docs |

#### Documentation Requirements

| Requirement | Verification | Evidence |
|-------------|--------------|----------|
| Investigation complete | All sections present | Investigation doc |
| Index updated | All artifacts listed | Index document |
| Lessons learned | Insights captured | Lessons doc |
| Self-review performed | Review documented | Review doc |

#### Approval Requirements

| Requirement | Verification | Evidence |
|-------------|--------------|----------|
| Human review | Reviewer approval | Approval record |
| Governance approval | Official approval | Governance doc |
| Closure authorized | Closure permission | Closure authorization |

### Completion Checklist

```
INVESTIGATION COMPLETION CHECKLIST
══════════════════════════════════

RESEARCH COMPLETION
═══════════════════
☐ Research question answered
☐ All experiments completed
☐ All evidence collected
☐ All evidence archived
☐ Analysis complete
☐ Conclusion documented

DOCUMENTATION COMPLETION
════════════════════════
☐ Investigation document complete
☐ Index document complete
☐ Lessons learned documented
☐ All evidence documented
☐ Recommendations prepared

REVIEW AND APPROVAL
═══════════════════
☐ Self-review performed
☐ Self-review documented
☐ Human review performed
☐ Human review approved
☐ Governance approval obtained

FINALIZATION
════════════
☐ All files in correct locations
☐ Naming conventions followed
☐ Metadata complete
☐ Investigation closed in tracking
☐ Recommendations submitted
```

### Closure Process

```
┌─────────────────────────────────────────────────────────────────┐
│                    INVESTIGATION CLOSURE                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐                                           │
│  │ Completion     │ ◄── Verify all requirements met            │
│  │ Verification   │                                           │
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │ Documentation  │ ◄── Finalize all documents                 │
│  │ Finalization   │                                           │
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │ Archive         │ ◄── Move to permanent storage             │
│  └────────┬────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │ Recommendations │ ◄── Submit knowledge and research recs    │
│  │ Submission      │                                           │
│  └─────────────────┘                                           │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Post-Closure

After investigation closure:

| Action | Responsible | Timeline |
|--------|-------------|----------|
| Update index | Runtime | Immediately |
| Archive artifacts | Runtime | Within 24h |
| Process recommendations | Governance | Within 7 days |
| Notify stakeholders | Runtime | Within 24h |

---

## Implementation

### Runtime Responsibilities

The Runtime SHALL:

1. **Initialize investigations** according to SOP-001
2. **Execute experiments** according to SOP-002
3. **Enforce documentation standards** per SOP-003
4. **Collect evidence** according to SOP-004
5. **Execute retrieval policy** per SOP-005
6. **Support promotion process** per SOP-006
7. **Process recommendations** per SOP-007
8. **Verify completion** according to SOP-008

### Engine Responsibilities

The Engine SHALL:

1. **Follow investigation lifecycle** per SOP-001
2. **Design and execute experiments** per SOP-002
3. **Produce required documentation** per SOP-003
4. **Collect and document evidence** per SOP-004
5. **Determine retrieval relevance** per SOP-005
6. **Prepare recommendations** per SOP-006 and SOP-007
7. **Self-verify completion** per SOP-008

### Governance Responsibilities

Governance SHALL:

1. **Own this SOP** and approve changes
2. **Approve investigations** for closure
3. **Approve knowledge promotion**
4. **Approve research recommendations**
5. **Monitor SOP compliance**

---

## Exceptions

### Emergency Procedures

In time-critical situations, the following exceptions apply:

| Normal Requirement | Exception | Justification Required |
|--------------------|-----------|----------------------|
| Full documentation | Summary only | Post-incident completion |
| Complete review | Expedited review | Time constraint |
| All experiments | Critical experiments only | Prioritization |

All exceptions SHALL be documented and reviewed within 7 days.

### Waiver Process

Investigations MAY request waivers for specific requirements:

1. Submit waiver request with justification
2. Governance reviews within 48 hours
3. Approved waivers documented in investigation

---

## Compliance

### Compliance Monitoring

| Metric | Measurement |
|--------|-------------|
| SOP adherence | Investigation reviews pass/fail |
| Completion rate | Investigations closed on time |
| Documentation quality | Review scores |
| Evidence quality | Evidence chain completeness |

### Non-Compliance

Non-compliance SHALL be addressed through:

1. Documentation of deviation
2. Root cause analysis
3. Corrective action
4. Process improvement recommendation

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2026-07-21 | KDE Governance | Initial release |

---

## Related Documents

| Document | Relationship |
|----------|-------------|
| [ENGINE-VERSIONING.md](./ENGINE-VERSIONING.md) | Engine management |
| [STATE-MACHINE.md](./STATE-MACHINE.md) | Artifact lifecycle |
| [seeds/seed-001/README.md](../seeds/seed-001/README.md) | Seed information |

---

**Document Status**: APPROVED  
**Effective Date**: 2026-07-21  
**Review Date**: 2027-07-21  
**Owner**: KDE Governance
