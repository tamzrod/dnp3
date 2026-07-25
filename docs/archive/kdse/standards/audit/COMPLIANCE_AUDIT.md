# KDSE Compliance Audit Standard

## Purpose

The Compliance Audit evaluates a repository that claims to follow KDSE. This audit type verifies that KDSE practices are properly implemented and provides evidence of compliance or gaps.

## Scope

A Compliance Audit examines:

1. **Knowledge Artifacts**: Presence and quality of knowledge artifacts
2. **Architecture Artifacts**: Presence and quality of architecture artifacts
3. **Implementation Artifacts**: Code and related artifacts
4. **Verification Practices**: Verification implementation
5. **Traceability**: Maintenance of traceability links
6. **Authority Hierarchy**: Implementation of authority structure
7. **Governance**: Governance practices for KDSE artifacts

## Normative References

- KDSE Audit Standard Version: 1.0
- KDSE Audit Scoring Model (AUDIT_SCORING.md)
- KDSE Audit Maturity Model (AUDIT_MATURITY.md)
- KDSE Audit Report Template (AUDIT_TEMPLATE.md)
- KDSE Foundation Documents (docs/foundation/)

## Required Inputs

The auditor must have access to:

1. **Repository Access**: Read access to repository
2. **Repository Documentation**: All documentation files
3. **Artifact Samples**: Representative sample of artifacts
4. **Traceability Records**: Traceability between artifacts
5. **Version Information**: Repository version being audited

## Repository Discovery

### Initial Discovery Process

Before detailed audit, discover the repository structure:

```
1. Identify Artifact Locations
   ├── Where are documents stored?
   ├── What file formats are used?
   └── How is documentation organized?

2. Identify Artifact Types
   ├── Knowledge artifacts present?
   ├── Architecture artifacts present?
   ├── Implementation artifacts present?
   └── Verification artifacts present?

3. Identify Relationships
   ├── How are artifacts linked?
   ├── Is traceability maintained?
   └── Is authority hierarchy evident?
```

### Discovery Checklist

- [ ] Repository URL and version identified
- [ ] Documentation location confirmed
- [ ] Artifact types cataloged
- [ ] Relationships identified
- [ ] Traceability links mapped
- [ ] Governance practices identified

## Audit Dimensions

### 1. Knowledge Artifacts

**Definition**: Quality and completeness of knowledge artifacts.

**What to Evaluate:**
- Knowledge artifacts exist and are structured
- Validation evidence present
- Provenance documented
- Dependencies identified
- Stewardship assigned

**Scoring Criteria:**

| Score | Criteria |
|-------|----------|
| 0-2 | No knowledge artifacts or completely unstructured |
| 2-4 | Basic documents exist without structure |
| 4-6 | Structured knowledge with some gaps |
| 6-8 | Complete knowledge artifacts with good structure |
| 8-10 | Exemplary knowledge artifacts |

**Evidence Required:**
- Sample knowledge artifacts
- Validation documentation
- Provenance records
- Dependency documentation

### 2. Architecture Artifacts

**Definition**: Quality and completeness of architecture artifacts.

**What to Evaluate:**
- Architecture artifacts exist
- Architecture derives from knowledge
- Architecture defines structure
- Decisions are documented (ADRs)
- Authority hierarchy maintained

**Scoring Criteria:**

| Score | Criteria |
|-------|----------|
| 0-2 | No architecture artifacts |
| 2-4 | Basic architecture without derivation |
| 4-6 | Architecture with some derivation evidence |
| 6-8 | Complete architecture with derivation |
| 8-10 | Exemplary architecture documentation |

**Evidence Required:**
- Architecture documentation
- Knowledge-architecture links
- Decision records
- Structure definitions

### 3. Implementation Artifacts

**Definition**: Quality of implementation relative to architecture.

**What to Evaluate:**
- Implementation traces to architecture
- Implementation follows architecture direction
- No unauthorized deviations
- Configuration managed appropriately
- Code quality standards met

**Scoring Criteria:**

| Score | Criteria |
|-------|----------|
| 0-2 | No clear implementation or chaotic |
| 2-4 | Implementation exists but undocumented |
| 4-6 | Implementation with some traceability |
| 6-8 | Well-traced implementation |
| 8-10 | Exemplary implementation practices |

**Evidence Required:**
- Code structure analysis
- Architecture-compliance evidence
- Traceability to architecture
- Deviation documentation (if any)

### 4. Verification Practices

**Definition**: Implementation of verification practices.

**What to Evaluate:**
- Verification criteria derived from knowledge
- Verification executed against implementation
- Verification results documented
- Non-conformances identified and tracked
- Verification authority established

**Verification Evidence Classification:**

KDSE distinguishes between four verification states. Auditors MUST classify verification status into one of these categories:

| State | Description | Evidence Required |
|-------|-------------|------------------|
| **Verified** | Tests executed and passed | Test execution records + Test results showing pass |
| **Verified with Failures** | Tests executed with documented failures | Test execution records + Test results showing failures |
| **Not Verified** | Test assets exist but no execution evidence | Test plans/cases only, NO execution records |
| **Not Assessed** | No verification artifacts found | Absence of any verification artifacts |

**Critical Principle:** The presence of test assets (test cases, test plans, test documentation) alone does NOT constitute verification evidence. Verification requires **execution evidence** - records that prove tests were actually run.

**Scoring Criteria:**

| Score | Criteria |
|-------|----------|
| 0-2 | No verification practices |
| 2-4 | Basic testing without knowledge basis |
| 4-6 | Verification with some criteria derivation |
| 6-8 | Comprehensive verification practices |
| 8-10 | Exemplary verification implementation |

**Evidence Required:**

Auditors MUST document the verification state for each artifact category:

| Evidence Type | Classification | Required for Verification? |
|--------------|----------------|---------------------------|
| Verification plans | Test Asset | No (planning only) |
| Test cases | Test Asset | No (definition only) |
| Test documentation | Test Asset | No (documentation only) |
| Test execution records | Execution Evidence | **YES - REQUIRED** |
| Test results | Execution Evidence | **YES - REQUIRED** |
| Non-conformance reports | Execution Evidence | **YES - REQUIRED (if failures exist)** |

**Evidence Classification Requirements:**

1. **Test Assets Present Only** (No Execution Evidence):
   - Score capped at 4/10
   - Report as "Not Verified"
   - Risk Assessment: HIGH uncertainty

2. **Execution Evidence Available** (Test Results + Records):
   - Score based on comprehensive criteria (4-10)
   - Report actual verification state (Verified, Verified with Failures)
   - Risk Assessment: Based on actual results

3. **No Verification Artifacts**:
   - Score 0/10
   - Report as "Not Assessed"
   - Risk Assessment: MAXIMUM uncertainty

**Reporting Requirements:**

Each verification artifact category MUST be reported with explicit status:

```
| Verification Category | Assets Exist | Execution Evidence | Status | Risk Level |
|---------------------|--------------|-------------------|--------|------------|
| [Category Name]      | Yes/No       | Yes/No            | Verified / Not Verified / Not Assessed | Low/Medium/High |
```

**Risk Assessment Guidance:**

- **Low Risk**: Verified with documented execution and passing results
- **Medium Risk**: Verified with failures, or Not Verified with partial execution
- **High Risk**: Not Verified (test assets exist but no execution), Not Assessed

Reports SHALL explicitly identify "Not Verified" when test assets exist but execution evidence is unavailable. Risk assessments SHALL reflect uncertainty instead of assuming correctness.

### 5. Traceability

**Definition**: Maintenance of traceability between artifacts.

**What to Evaluate:**
- Forward traceability (Knowledge → Architecture → Implementation)
- Backward traceability (Implementation → Architecture → Knowledge)
- Traceability links documented
- Traceability completeness
- Traceability verification

**Scoring Criteria:**

| Score | Criteria |
|-------|----------|
| 0-2 | No traceability maintained |
| 2-4 | Partial traceability, inconsistent |
| 4-6 | Basic traceability implemented |
| 6-8 | Comprehensive traceability |
| 8-10 | Exemplary traceability practices |

**Evidence Required:**
- Traceability matrix
- Link documentation
- Completeness metrics
- Verification evidence

### 6. Authority Hierarchy

**Definition**: Implementation of KDSE authority structure.

**What to Evaluate:**
- Knowledge as highest authority
- Architecture derives from knowledge
- Implementation follows architecture
- Authority violations identified
- Authority resolution practiced

**Scoring Criteria:**

| Score | Criteria |
|-------|----------|
| 0-2 | Authority hierarchy not implemented |
| 2-4 | Partial hierarchy, significant gaps |
| 4-6 | Basic hierarchy implemented |
| 6-8 | Comprehensive hierarchy maintained |
| 8-10 | Exemplary authority implementation |

**Evidence Required:**
- Authority structure documentation
- Derivation evidence
- Violation log (if any)
- Resolution documentation

### 7. Governance

**Definition**: Governance practices for KDSE artifacts.

**What to Evaluate:**
- Stewardship assigned
- Review processes defined
- Change management implemented
- Artifact lifecycle managed
- Compliance verification

**Scoring Criteria:**

| Score | Criteria |
|-------|----------|
| 0-2 | No governance practices |
| 2-4 | Basic governance, informal |
| 4-6 | Defined governance processes |
| 6-8 | Comprehensive governance |
| 8-10 | Exemplary governance practices |

**Evidence Required:**
- Stewardship assignments
- Review process documentation
- Change management records
- Lifecycle documentation

## Gap Analysis

### Gap Categories

**Critical Gaps:**
- Required artifact types missing
- No traceability maintained
- Authority hierarchy violated
- Fundamental KDSE principles ignored

**Major Gaps:**
- Incomplete artifact structure
- Partial traceability
- Some authority violations
- Governance gaps

**Minor Gaps:**
- Documentation quality issues
- Minor traceability gaps
- Process improvements possible
- Enhancement opportunities

### Gap Documentation

For each gap identified:

```
### Gap [ID]: [Gap Title]

**Dimension:** [Affected dimension]
**Severity:** [Critical | Major | Minor]
**Description:** [Clear description of gap]

**Evidence:**
[Specific evidence demonstrating the gap]

**Impact:**
[What this prevents or causes]

**Recommendation:**
[How to address this gap]
```

## Compliance Levels

### Level Definitions

| Level | Score Range | Description |
|-------|-------------|-------------|
| Compliant | 8-10 | Fully compliant with KDSE |
| Substantially Compliant | 6-8 | Compliant with minor gaps |
| Partially Compliant | 4-6 | Significant gaps exist |
| Minimally Compliant | 2-4 | Basic structure, major gaps |
| Non-Compliant | 0-2 | Does not meet KDSE standards |

### Level Requirements

**Compliant (8-10):**
- All artifact types properly implemented
- Complete traceability
- Authority hierarchy maintained
- Comprehensive governance
- Minor improvement opportunities only

**Substantially Compliant (6-8):**
- All artifact types present with good quality
- Mostly complete traceability
- Authority hierarchy generally maintained
- Governance mostly effective
- Some gaps that don't fundamentally affect KDSE

**Partially Compliant (4-6):**
- Most artifact types present
- Partial traceability
- Some authority issues
- Governance gaps
- Significant work needed

**Minimally Compliant (2-4):**
- Basic structure exists
- Minimal traceability
- Authority issues common
- Governance incomplete
- Major effort required

**Non-Compliant (0-2):**
- Does not demonstrate KDSE understanding
- Fundamental principles not implemented
- Not suitable for compliance claim

## Findings

### Finding Categories

**Compliance Finding:**
```
### Compliance Finding [ID]

**Type:** [Positive | Negative]
**Dimension:** [Affected dimension]
**Description:** [Clear description]

**Evidence:**
[Supporting evidence]

**Assessment:**
[Why this is a finding]
```

**Gap Finding:**
```
### Gap Finding [ID]

**Type:** Gap
**Dimension:** [Affected dimension]
**Severity:** [Critical | Major | Minor]
**Description:** [Clear description]

**Evidence:**
[Supporting evidence]

**Impact:**
[What this affects]

**Recommendation:**
[How to address]
```

### Finding Prioritization

**Priority 1 - Critical:**
- Fundamental KDSE principles not implemented
- Traceability completely absent
- Authority hierarchy violated
- No governance practices

**Priority 2 - High:**
- Major artifact types missing
- Significant traceability gaps
- Authority issues affecting outcomes
- Governance gaps

**Priority 3 - Medium:**
- Artifact quality issues
- Partial traceability
- Minor authority issues
- Process improvements needed

**Priority 4 - Low:**
- Documentation quality
- Enhancement opportunities
- Best practice suggestions
- Optimization possibilities

## Recommendations

### Recommendation Structure

```
### Recommendation [ID]

**Priority:** [1 | 2 | 3 | 4]
**Dimension:** [Affected dimension]
**Title:** [Clear, actionable title]

**What:** [What should be done]

**Why:** [Why this matters]

**How:** [How to accomplish this]

**Effort:** [Estimated effort: Low | Medium | High]

**Impact:** [Expected benefit]
```

### Recommendation Categories

**Must Do (Priority 1):**
- Required for compliance
- Fundamental to KDSE
- No alternatives

**Should Do (Priority 2):**
- Strongly recommended
- Significant improvement
- Clear benefit

**Consider (Priority 3):**
- Suggested improvements
- Moderate benefit
- Resource dependent

**Could Do (Priority 4):**
- Optional enhancements
- Minor improvements
- Nice to have

## Final Verdict

### Verdict Statement

The final verdict must include:

```
**VERDICT: [COMPLIANCE LEVEL]**

**Overall Score: X.X / 10**

**Summary:**
[Brief explanation of verdict]

**Key Strengths:**
- [Strength 1]
- [Strength 2]

**Key Gaps:**
- [Gap 1]
- [Gap 2]

**Required Actions:**
1. [Action 1]
2. [Action 2]

**Timeline for Compliance:**
[Timeframe if partial compliance]
```

### Verdict Confidence

Indicate confidence in the verdict:

- **High Confidence**: Thorough examination, clear evidence
- **Medium Confidence**: Good examination, some limitations
- **Low Confidence**: Limited examination, significant unknowns

## Required Deliverables

A Compliance Audit report must include:

1. **Repository Metadata**: URL, version, date
2. **Discovery Summary**: Initial findings
3. **Dimension Scores**: All 7 dimensions scored
4. **Evidence**: Supporting evidence for findings
5. **Gap Analysis**: All gaps categorized
6. **Recommendations**: Prioritized actions
7. **Final Verdict**: Clear compliance assessment
8. **Appendix**: Evidence index

## Repository Independence

This audit must remain repository-independent:

- No repository-specific recommendations
- Apply KDSE standards consistently
- Use same criteria for all repositories
- Document any repository-specific considerations separately

## Version

- **Document Version**: 1.1
- **Effective Date**: 2026-07-10
- **Standard Version**: KDSE Audit Standard 1.1
- **Change Note**: Added verification evidence classification requirements to address KDSE-DEFECT-001

---

*This standard is part of the KDSE Audit System. See [README.md](README.md) for audit system overview.*
