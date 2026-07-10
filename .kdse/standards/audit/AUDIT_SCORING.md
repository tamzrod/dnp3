# KDSE Audit Scoring Model

## Overview

The KDSE Audit Scoring Model provides a standardized framework for evaluating methodology maturity and repository compliance. The model uses a 0-10 scale divided into six maturity levels.

## Scoring Scale

### Score Ranges and Maturity Levels

| Score | Level | Name | Engineering Meaning |
|-------|-------|------|---------------------|
| 0-2 | 1 | Concept | Early stage development |
| 2-4 | 2 | Defined | Documentation exists |
| 4-6 | 3 | Structured | Processes formalized |
| 6-8 | 4 | Usable | Applied in practice |
| 8-9 | 5 | Validated | Benefits demonstrated |
| 9-10 | 6 | Proven | Repeated success |

### Detailed Level Definitions

#### Level 1: Concept (0-2)

**Characteristics:**
- Basic ideas exist but are not documented
- No formal structure
- No consistent practice
- Initial understanding only

**Engineering Indicators:**
- "We have some documents"
- "There's a wiki with ideas"
- "A few people understand it"
- No standardized processes

**What This Score Is Not:**
- Not a complete methodology
- Not ready for adoption by others
- Not suitable for compliance evaluation

#### Level 2: Defined (2-4)

**Characteristics:**
- Core concepts are documented
- Artifact types are identified
- Basic roles are understood
- Processes exist but may be informal

**Engineering Indicators:**
- Documentation exists for main concepts
- Artifact types are named and described
- Someone is responsible for each area
- Traceability is attempted but inconsistent

**What This Score Is Not:**
- Not necessarily consistent in practice
- Not validated through application
- Not suitable for regulated environments

#### Level 3: Structured (4-6)

**Characteristics:**
- Formal processes are defined
- Practices are documented
- Review workflows exist
- Traceability is maintained

**Engineering Indicators:**
- Documents follow consistent structure
- Reviews are scheduled and performed
- Traceability links are established
- Change processes are defined

**What This Score Is Not:**
- Not necessarily producing good outcomes
- Not validated by measurement
- May lack practical guidance

#### Level 4: Usable (6-8)

**Characteristics:**
- Processes are applied consistently
- Teams produce expected artifacts
- Quality standards are met
- Practices are repeatable

**Engineering Indicators:**
- Teams follow documented processes
- Artifacts meet quality criteria
- Traceability is verified regularly
- New team members can onboard

**What This Score Is Not:**
- Not necessarily optimized
- Not necessarily validated through measurement
- May not have proven benefits

#### Level 5: Validated (8-9)

**Characteristics:**
- Outcomes are measured
- Benefits are demonstrated
- Gaps are identified through evidence
- Improvements are implemented

**Engineering Indicators:**
- Metrics are collected and analyzed
- Success is demonstrated with data
- Failure modes are understood
- Continuous improvement occurs

**What This Score Is Not:**
- Not necessarily repeatable across contexts
- Not necessarily optimized
- May still have edge cases

#### Level 6: Proven (9-10)

**Characteristics:**
- Repeated success demonstrated
- Multiple contexts validated
- Methodology is refined through evidence
- Community validation achieved

**Engineering Indicators:**
- Multiple successful applications
- Diverse contexts demonstrate success
- Methodology evolved based on evidence
- External recognition of quality

**What This Score Is Not:**
- Not claiming perfection
- Not static - continues to evolve
- Not mandatory for all applications

## Scoring Principles

### Evidence-Based Scoring

Every score must be supported by evidence:

1. **Direct Observation**: Score based on actual artifacts, not claims
2. **Multiple Examples**: Verify patterns, not exceptions
3. **Consistent Standards**: Same criteria applied to all targets
4. **Clear Justification**: Document why each score was assigned

### Avoiding Arbitrary Scores

**Do Not:**
- Give high scores to incomplete work
- Score based on potential rather than reality
- Allow familiarity to inflate scores
- Score based on effort rather than outcomes

**Do:**
- Score based on demonstrated characteristics
- Require evidence for high scores
- Document the basis for each score
- Be willing to give low scores to inadequate work

### Score Interpretation

**Low Scores (0-4):**
- Indicate significant gaps
- Require substantial work
- Should not block honest assessment

**Mid Scores (4-6):**
- Indicate foundation exists
- Show progress toward maturity
- Identify areas needing attention

**High Scores (6-10):**
- Indicate mature practices
- Show validated results
- May still have improvement opportunities

## Scoring Process

### Step 1: Define Criteria

For each dimension, define specific criteria that indicate each maturity level.

### Step 2: Gather Evidence

Collect evidence through direct examination of artifacts and processes.

### Step 3: Evaluate Against Criteria

Compare evidence to defined criteria for each level.

### Step 4: Assign Score

Assign the score that best matches the evidence. When evidence spans multiple levels, document the rationale.

### Step 5: Document Justification

Record why each score was assigned, including specific evidence.

## Overall Score Calculation

### Dimension Scores

Each audit evaluates multiple dimensions. Each dimension receives its own score.

**Example Dimensions for Foundation Audit:**
- Identity
- Vision
- Repository Structure
- Body of Knowledge
- Engineering Philosophy
- Terminology
- Traceability
- Practicality
- Scalability
- Independence

**Example Dimensions for Compliance Audit:**
- Knowledge Artifacts
- Architecture Artifacts
- Implementation Artifacts
- Verification Practices
- Traceability Implementation
- Authority Hierarchy
- Governance

### Aggregating Scores

**Overall Score Options:**

1. **Simple Average**: Sum all dimension scores, divide by count
2. **Weighted Average**: Apply weights based on dimension importance
3. **Minimum Score**: Use the lowest dimension score
4. **Critical Path**: Use the score of the most critical dimension

**Recommended Approach:**

For most KDSE audits, use the **simple average** of dimension scores. This:
- Avoids arbitrary weighting
- Ensures no dimension is ignored
- Provides balanced assessment

For specific purposes, audits may document alternative aggregation methods.

### Score Presentation

Present scores clearly:

```
Overall Score: 6.8 / 10

| Dimension | Score |
|-----------|-------|
| Identity | 8/10 |
| Vision | 7/10 |
| ... | ... |
```

Include:
- Overall score prominently displayed
- Dimension scores in table format
- Score changes from previous audits (if applicable)
- Trend indicators (improving, stable, declining)

## Score Interpretation Guide

### What Scores Mean

| Score | Interpretation | Action Required |
|-------|---------------|-----------------|
| 0-2 | Concept only | Complete foundation work |
| 2-4 | Basic structure | Formalize processes |
| 4-6 | Functional | Improve consistency |
| 6-8 | Mature | Validate outcomes |
| 8-9 | Advanced | Share learnings |
| 9-10 | Exemplary | Maintain and evolve |

### What Scores Do Not Mean

**Scores Do Not:**
- Guarantee success
- Predict future performance
- Replace judgment
- Define compliance

**Scores Indicate:**
- Current maturity level
- Strengths and weaknesses
- Areas for improvement
- Progress over time

## Score Quality Assurance

### Self-Consistency Check

Verify that scores are internally consistent:

- High "Identity" score should correlate with clear documentation
- High "Traceability" score should correlate with documented links
- Low scores should have documented justification

### External Consistency Check

Where possible, verify scores against external evidence:

- Previous audit scores
- Independent assessments
- User feedback
- Outcomes data

### Score Revision

Scores may be revised if:
- New evidence emerges
- Errors are discovered
- Criteria are clarified

Revisions should be documented with rationale.

## Scoring Examples

### Low Score Example (2/10 - Defined)

**Dimension: Body of Knowledge**

**Evidence Found:**
- Two documents exist
- Documents have titles but minimal content
- No consistent structure
- No cross-references between documents

**Score Justification:**
"The methodology has basic documentation naming some concepts. However, documents lack substantive content, consistent structure, or cross-references. The approach barely exceeds concept level but has minimal formalization."

### High Score Example (8/10 - Validated)

**Dimension: Traceability**

**Evidence Found:**
- All artifacts have explicit traceability links
- Traceability is verified during review
- Metrics show 95% traceability completeness
- Gaps are tracked and addressed

**Score Justification:**
"Traceability is consistently implemented and verified. Metrics demonstrate high compliance rates. The team actively monitors and improves traceability. Evidence supports validated maturity level."

## Glossary Additions

This document introduces the following terms:

### Maturity Level

A category indicating the development state of a methodology or practice. KDSE defines six maturity levels: Concept, Defined, Structured, Usable, Validated, and Proven.

### Dimension Score

The score assigned to a specific evaluation area within an audit. Multiple dimension scores are combined to produce an overall score.

### Evidence-Based Scoring

A scoring approach where scores are assigned based on direct observation of artifacts and processes, not claims or assumptions.

### Score Aggregation

The method used to combine dimension scores into an overall score. Common methods include simple average, weighted average, minimum, and critical path.
