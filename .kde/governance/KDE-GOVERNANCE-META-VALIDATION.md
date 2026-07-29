# KDE Meta-Validation Framework

**Document ID**: KDE-GOVERNANCE-META-001
**Version**: 1.0.0
**Status**: APPROVED
**Authority**: INV-AUDIT-REVIEW-001 (Priority 10)
**Effective Date**: 2026-07-27
**Source**: Meta-Validation Implementation
**Dependencies**: Knowledge Provenance (v2.0.0)

---

## Purpose

This document establishes a **Meta-Validation Framework** for KDE. As identified in INV-AUDIT-REVIEW-001:

> "Highest impact but furthest in future. Validates the validation methodology itself."

Meta-validation answers the question: **"How do we know our validations are valid?"**

This framework enables:
- Consistent validation standards across domains
- Traceable validation chains
- Self-improving validation methodology
- Quality assurance for validation processes

---

## The Meta-Validation Question

### Why Meta-Validation Matters

Standard validation asks: "Is this artifact correct?"
Meta-validation asks: "Is our definition of 'correct' correct?"

```
┌─────────────────────────────────────────────────────────────────┐
│                    VALIDATION HIERARCHY                         │
│                                                                  │
│    ┌─────────────────────────────────────────────────────────┐  │
│    │              META-VALIDATION                             │  │
│    │    "Is our validation methodology correct?"             │  │
│    └───────────────────────┬─────────────────────────────────┘  │
│                            │                                     │
│    ┌───────────────────────┴─────────────────────────────────┐  │
│    │              VALIDATION                                 │  │
│    │    "Is this artifact correct?"                          │  │
│    └───────────────────────┬─────────────────────────────────┘  │
│                            │                                     │
│    ┌───────────────────────┴─────────────────────────────────┐  │
│    │              EVIDENCE                                    │  │
│    │    "What evidence supports correctness?"                │  │
│    └─────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### The Problem

Without meta-validation:
1. Validation criteria may be inconsistent across domains
2. Validation processes may drift over time
3. There's no way to improve validation methodology
4. "Correct" is defined ad-hoc per artifact

---

## Validation Levels

### Level 1: Syntax Validation

**Question**: "Does the artifact follow the correct format?"

| Check | Description |
|-------|-------------|
| Schema compliance | Matches template/schema |
| Required fields | All required fields present |
| Format correctness | Dates, IDs, versions correct |
| Link validity | All references resolve |

**Who validates**: Automated tooling

### Level 2: Content Validation

**Question**: "Does the artifact content make sense?"

| Check | Description |
|-------|-------------|
| Internal consistency | No contradictory statements |
| Completeness | All required sections present |
| Clarity | Clear, unambiguous language |
| Relevance | Content matches stated purpose |

**Who validates**: Author + peer review

### Level 3: Evidence Validation

**Question**: "Is the evidence sufficient and credible?"

| Check | Description |
|-------|-------------|
| Evidence existence | Evidence present for claims |
| Evidence quality | Evidence from credible sources |
| Evidence relevance | Evidence supports claims |
| Evidence sufficiency | Enough evidence for confidence |

**Who validates**: Domain expert

### Level 4: Methodology Validation

**Question**: "Was the artifact produced correctly?"

| Check | Description |
|-------|-------------|
| Process compliance | Followed required process |
| Tool correctness | Used appropriate tools |
| Provenance trace | Source investigations linked |
| Version correctness | Used correct engine/seed versions |

**Who validates**: Process auditor

### Level 5: Meta-Validation

**Question**: "Is our validation methodology itself valid?"

| Check | Description |
|-------|-------------|
| Criteria consistency | Validation criteria applied uniformly |
| Criteria coverage | All important aspects validated |
| Criteria relevance | Criteria match real requirements |
| Criteria improvement | Criteria improve over time |

**Who validates**: Governance (meta-audits)

---

## Validation Criteria Catalog

### Evidence Quality Criteria

| Criterion | Definition | Assessment |
|-----------|------------|------------|
| **Credibility** | Source is trustworthy | 1-5 scale |
| **Relevance** | Evidence directly supports claim | Boolean |
| **Sufficiency** | Enough evidence for confidence | Boolean |
| **Consistency** | Evidence consistent with other evidence | Boolean |
| **Freshness** | Evidence is current | Age in days |

### Knowledge Quality Criteria

| Criterion | Definition | Assessment |
|-----------|------------|------------|
| **Accuracy** | Correct representation of facts | Binary |
| **Completeness** | All relevant aspects covered | Percentage |
| **Clarity** | Unambiguous to intended audience | 1-5 scale |
| **Actionability** | Enables informed action | Boolean |
| **Traceability** | Source investigations documented | Boolean |

### Process Quality Criteria

| Criterion | Definition | Assessment |
|-----------|------------|------------|
| **Compliance** | Followed required process | Boolean |
| **Documentation** | Process documented | Boolean |
| **Reproducibility** | Can be repeated | Boolean |
| **Efficiency** | Completed in reasonable time | Boolean |
| **Improvement** | Process improved over time | Boolean |

---

## Validation Chain

### What Is a Validation Chain?

A **validation chain** traces an artifact back through all validations to its source evidence.

```
┌─────────────────────────────────────────────────────────────────┐
│                    VALIDATION CHAIN EXAMPLE                      │
│                                                                  │
│    Knowledge Document (KDE-XXX)                                  │
│         │                                                        │
│         │ Validated by: Validator [Date]                         │
│         ▼                                                        │
│    Validation Record (validated.md)                              │
│         │                                                        │
│         │ Produced by: Investigation INV-YYY                     │
│         ▼                                                        │
│    Investigation (INV-YYY)                                       │
│         │                                                        │
│         │ Engine: KDE-ENGINE-003 (v0.1.0)                        │
│         │ Seed: SEED-001 (v1.0.0)                               │
│         ▼                                                        │
│    Source Evidence (EV-001, EV-002, ...)                        │
│         │                                                        │
│         │ Collected from: [Sources]                               │
│         ▼                                                        │
│    External Sources                                               │
└─────────────────────────────────────────────────────────────────┘
```

### Chain Integrity

For a validation chain to be valid:

1. **Completeness**: All steps present
2. **Correctness**: Each step follows required process
3. **Consistency**: Steps don't contradict
4. **Currentness**: Evidence is fresh enough

---

## Meta-Validation Process

### When to Conduct Meta-Validation

| Trigger | Frequency | Scope |
|---------|-----------|-------|
| New domain introduced | Per domain | Domain only |
| Validation criteria change | Per change | Affected domains |
| Quality issues discovered | As needed | Specific artifacts |
| Quarterly review | Quarterly | All domains |

### Meta-Validation Checklist

#### For New Validation Criteria

- [ ] Criteria defined clearly
- [ ] Criteria assess meaningful qualities
- [ ] Criteria can be applied consistently
- [ ] Criteria don't conflict with existing criteria
- [ ] Assessment method defined
- [ ] Training provided for validators

#### For Validation Process Review

- [ ] Process followed correctly
- [ ] Process appropriate for artifact type
- [ ] Process documented accurately
- [ ] Process produces consistent results
- [ ] Process improvements identified

#### For Validation Quality Assessment

- [ ] Validations produce consistent results
- [ ] Validators trained and calibrated
- [ ] Quality metrics tracked over time
- [ ] Quality improvements implemented

---

## Validation Standards by Domain

### Knowledge Validation

| Level | Standard | Evidence Required |
|-------|----------|-------------------|
| 1 | Syntax correct | Template compliance |
| 2 | Content valid | Internal consistency |
| 3 | Evidence sufficient | 3+ credible sources |
| 4 | Methodology sound | Investigation linked |
| 5 | Standards met | All prior levels |

### Investigation Validation

| Level | Standard | Evidence Required |
|-------|----------|-------------------|
| 1 | Structure correct | Template compliance |
| 2 | Methodology sound | Process documented |
| 3 | Evidence sufficient | 3+ pieces evidence |
| 4 | Conclusions valid | Evidence supports |
| 5 | Reproducible | Chain documented |

### Expert Validation

| Level | Standard | Evidence Required |
|-------|----------|-------------------|
| 1 | Interface correct | Schema compliance |
| 2 | Implementation sound | Code review |
| 3 | Tests pass | 90%+ coverage |
| 4 | Performance acceptable | Benchmarks pass |
| 5 | Production ready | All prior levels |

---

## Validation Calibration

### What Is Calibration?

**Calibration** ensures validators apply criteria consistently.

### Calibration Process

1. **Establish baseline**: Validator reviews sample artifacts
2. **Compare results**: Compare validator results to gold standard
3. **Identify drift**: Find systematic differences
4. **Correct drift**: Retrain or clarify criteria
5. **Verify correction**: Re-test on sample

### Calibration Schedule

| Validator Type | Calibration Frequency |
|---------------|---------------------|
| Human validators | Quarterly |
| Automated validators | Per release |
| Domain experts | Semi-annually |

---

## Quality Metrics

### Validation Quality Metrics

| Metric | Definition | Target |
|--------|------------|--------|
| **Consistency Rate** | % validations with consistent results | > 95% |
| **Appeal Rate** | % validations appealed | < 5% |
| **Appeal Success Rate** | % appeals that succeed | < 20% |
| **Turnaround Time** | Average validation duration | < 7 days |
| **Coverage** | % artifacts validated | 100% |

### Tracking Quality Over Time

```markdown
## Validation Quality Dashboard

| Metric | Q1 | Q2 | Q3 | Q4 | Target |
|--------|----|----|----|----|----|
| Consistency Rate | 92% | 94% | 96% | 97% | 95% |
| Appeal Rate | 8% | 6% | 5% | 4% | 5% |
| Turnaround (days) | 10 | 8 | 7 | 6 | 7 |

### Trend Analysis
- ✅ Consistency improving
- ✅ Appeal rate declining
- ✅ Turnaround improving
```

---

## Validation Improvement

### Identifying Improvement Opportunities

Review validation processes for:
- Inconsistent results across validators
- Appeals frequently succeeding
- Quality issues not caught by validation
- Time-consuming steps that add little value
- Gaps in coverage

### Improvement Process

1. **Identify**: Document quality issue
2. **Analyze**: Root cause analysis
3. **Propose**: Suggest criteria/process change
4. **Pilot**: Test on subset
5. **Evaluate**: Measure improvement
6. **Rollout**: Apply to all

### Criteria Evolution

Validation criteria should evolve:
- **Annually**: Full review of criteria relevance
- **Quarterly**: Review of criteria application
- **As needed**: Emergency changes for quality issues

---

## Validation Audit

### Conducting a Validation Audit

An **audit** examines whether validation processes are being followed correctly.

### Audit Checklist

- [ ] All artifacts have validation records
- [ ] Validation records are complete
- [ ] Validators are authorized
- [ ] Validation criteria are applied consistently
- [ ] Appeals are handled correctly
- [ ] Quality metrics are tracked
- [ ] Improvement actions are taken

### Audit Report Template

```markdown
## Validation Audit Report

**Audit Period**: YYYY-QX
**Auditor**: [Name]
**Date**: YYYY-MM-DD

### Scope
- Artifacts audited: [Count]
- Domains covered: [List]
- Validation periods: [Date range]

### Findings
| Finding | Severity | Count |
|---------|----------|-------|
| Critical | High | X |
| Major | Medium | Y |
| Minor | Low | Z |

### Recommendations
1. [Recommendation 1]
2. [Recommendation 2]

### Follow-up
- Actions required: [List]
- Deadline: [Date]
- Owner: [Name]
```

---

## Enforcement

### Validation Requirements

| Artifact Type | Validation Required | Validator |
|---------------|--------------------| -----------|
| Knowledge | Yes (Level 3+) | Domain expert |
| Investigation | Yes (Level 3+) | Process auditor |
| Expert | Yes (Level 4+) | Governance |
| Seed | Yes (Level 5) | Human only |

### Non-Compliance

| Violation | Consequence |
|-----------|-------------|
| Missing validation | Artifact not promoted |
| Invalid validation | Validation must be repeated |
| Unauthorized validator | Validation not recognized |
| False validation | Serious violation, review required |

---

## References

| Document | Relationship |
|----------|--------------|
| `knowledge/KDE-KNOWLEDGE-LIFECYCLE.md` | Knowledge validation |
| `governance/KDE-GOVERNANCE-DEPENDENCY-TRACKING.md` | Dependency tracking |
| `laboratory/validations/README.md` | Validation registry |

---

## Version History

| Version | Date | Changes | Authority |
|---------|------|---------|-----------|
| 1.0.0 | 2026-07-27 | Initial meta-validation framework | INV-AUDIT-REVIEW-001 |

---

**Document Status**: APPROVED
**Authority**: INV-AUDIT-REVIEW-001
**Compliance**: MANDATORY
**Review Date**: 2027-01-27
