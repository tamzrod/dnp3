# KDSE Foundation Audit Standard

## Purpose

The Foundation Audit evaluates the KDSE methodology itself. This audit type ensures that KDSE meets its own standards and continues to evolve through evidence.

## Scope

A Foundation Audit examines:

1. **Core Documents**: All documents in `docs/foundation/`
2. **Methodology Consistency**: Internal coherence of KDSE
3. **Conceptual Completeness**: All required concepts defined
4. **Terminology Alignment**: Consistent use of terms
5. **Evidence Basis**: All claims supported by evidence

## Normative References

- KDSE Audit Standard Version: 1.0
- KDSE Audit Scoring Model (AUDIT_SCORING.md)
- KDSE Audit Maturity Model (AUDIT_MATURITY.md)
- KDSE Audit Report Template (AUDIT_TEMPLATE.md)

## Prerequisites

Before conducting a Foundation Audit:

1. **Documentation Review**: Auditor has read all foundation documents
2. **Scope Definition**: Audit scope is clearly defined
3. **Standard Reference**: Auditor understands scoring criteria
4. **Evidence Standards**: Auditor understands evidence requirements

## Required Inputs

The auditor must have access to:

1. **All Foundation Documents**: Complete set in `docs/foundation/`
2. **Evolution Documents**: Historical evidence in `docs/evolution/`
3. **Audit History**: Previous audits in `docs/audit/`
4. **Repository Metadata**: Git history, version information

## Audit Dimensions

### 1. Identity

**Definition**: How clearly KDSE defines itself.

**What to Evaluate:**
- Canonical definition present and clear
- Problem statement articulated
- Scope boundaries defined
- Differentiation from related methodologies

**Scoring Criteria:**

| Score | Criteria |
|-------|----------|
| 0-2 | No clear definition exists |
| 2-4 | Basic definition exists but incomplete |
| 4-6 | Definition is clear but may lack specificity |
| 6-8 | Clear, specific definition with scope |
| 8-10 | Exemplary clarity with strong differentiation |

**Evidence Required:**
- Canonical definition statement
- Problem statement documentation
- Scope boundaries explicit
- Differentiation documented

### 2. Vision

**Definition**: How clearly KDSE defines its long-term direction.

**What to Evaluate:**
- Vision statement present
- Evolution principles defined
- Non-goals articulated
- Boundaries maintained over time

**Scoring Criteria:**

| Score | Criteria |
|-------|----------|
| 0-2 | No vision articulated |
| 2-4 | Basic vision exists but vague |
| 4-6 | Vision defined with some specifics |
| 6-8 | Clear vision with defined principles |
| 8-10 | Comprehensive vision with evolution path |

**Evidence Required:**
- Vision statement
- Evolution principles
- Non-claims documented
- Boundary maintenance evidence

### 3. Repository Structure

**Definition**: How well documentation is organized.

**What to Evaluate:**
- Logical document organization
- Clear navigation structure
- Appropriate file naming
- Cross-references defined

**Scoring Criteria:**

| Score | Criteria |
|-------|----------|
| 0-2 | Disorganized, no clear structure |
| 2-4 | Basic organization exists |
| 4-6 | Logical structure with some issues |
| 6-8 | Well-organized with clear navigation |
| 8-10 | Exemplary organization scalable to growth |

**Evidence Required:**
- Directory structure
- Document naming conventions
- Index or navigation present
- Cross-reference system

### 4. Body of Knowledge

**Definition**: Completeness of methodology content.

**What to Evaluate:**
- All artifact types defined
- All principles documented
- All processes specified
- Missing content identified

**Scoring Criteria:**

| Score | Criteria |
|-------|----------|
| 0-2 | Minimal content exists |
| 2-4 | Basic concepts documented |
| 4-6 | Core content complete |
| 6-8 | Comprehensive content with minor gaps |
| 8-10 | Complete content with extensions |

**Evidence Required:**
- Artifact type coverage
- Principle coverage
- Process coverage
- Gap analysis documented

### 5. Engineering Philosophy

**Definition**: Internal consistency of methodology.

**What to Evaluate:**
- Principles are internally consistent
- No contradictions between documents
- Principles support each other
- Core logic is sound

**Scoring Criteria:**

| Score | Criteria |
|-------|----------|
| 0-2 | Multiple contradictions exist |
| 2-4 | Some inconsistencies identified |
| 4-6 | Generally consistent with minor issues |
| 6-8 | Consistent with documented resolutions |
| 8-10 | Exemplary internal consistency |

**Evidence Required:**
- Principle consistency analysis
- Document cross-reference verification
- Contradiction log (if any)
- Resolution documentation

### 6. Terminology

**Definition**: Consistency of terminology across documents.

**What to Evaluate:**
- Glossary present and complete
- Terms used consistently
- No terminology drift
- Clear definitions for all key terms

**Scoring Criteria:**

| Score | Criteria |
|-------|----------|
| 0-2 | No glossary, inconsistent terms |
| 2-4 | Basic glossary exists |
| 4-6 | Glossary with some gaps |
| 6-8 | Comprehensive, consistent terminology |
| 8-10 | Exemplary terminology clarity |

**Evidence Required:**
- Glossary completeness
- Term usage analysis
- Cross-document consistency check
- Terminology drift evidence

### 7. Traceability

**Definition**: Traceability within methodology documentation.

**What to Evaluate:**
- Traceability framework defined
- Links between documents exist
- Dependencies documented
- Impact analysis possible

**Scoring Criteria:**

| Score | Criteria |
|-------|----------|
| 0-2 | No traceability defined |
| 2-4 | Basic traceability concept exists |
| 4-6 | Traceability framework defined |
| 6-8 | Comprehensive traceability implemented |
| 8-10 | Exemplary traceability throughout |

**Evidence Required:**
- Traceability framework documentation
- Document relationship mapping
- Dependency documentation
- Impact analysis capability

### 8. Practicality

**Definition**: Ease of applying the methodology.

**What to Evaluate:**
- Clear guidance for practitioners
- Adoption path defined
- Worked examples present
- Common challenges addressed

**Scoring Criteria:**

| Score | Criteria |
|-------|----------|
| 0-2 | Theoretical only, no guidance |
| 2-4 | Basic guidance exists |
| 4-6 | Practical guidance with gaps |
| 6-8 | Comprehensive practical guidance |
| 8-10 | Exemplary practical guidance |

**Evidence Required:**
- Adoption model presence
- Worked examples
- Common patterns documented
- Practitioner feedback (if available)

### 9. Scalability

**Definition**: Applicability across different scales.

**What to Evaluate:**
- Scale considerations addressed
- Different contexts covered
- Scaling patterns defined
- Size-appropriate guidance

**Scoring Criteria:**

| Score | Criteria |
|-------|----------|
| 0-2 | No scale considerations |
| 2-4 | Basic scale acknowledgment |
| 4-6 | Scale addressed for common cases |
| 6-8 | Comprehensive scale guidance |
| 8-10 | Exemplary multi-scale support |

**Evidence Required:**
- Scale range documentation
- Context-specific guidance
- Scaling patterns
- Size-appropriate processes

### 10. Independence

**Definition**: Technology and vendor neutrality.

**What to Evaluate:**
- No technology dependencies
- No vendor lock-in
- Platform-agnostic principles
- Tool-neutral guidance

**Scoring Criteria:**

| Score | Criteria |
|-------|----------|
| 0-2 | Strong technology dependencies |
| 2-4 | Some dependencies exist |
| 4-6 | Generally independent |
| 6-8 | Fully independent with documented choices |
| 8-10 | Exemplary independence |

**Evidence Required:**
- Technology independence verification
- Vendor neutrality check
- Platform coverage analysis
- Tool recommendation clarity

## Evaluation Process

### Phase 1: Planning

1. **Define Scope**: Confirm this is a Foundation Audit
2. **Gather Materials**: Obtain all foundation documents
3. **Review Standards**: Understand scoring criteria
4. **Plan Schedule**: Allocate time for thorough review

### Phase 2: Document Review

1. **Read All Documents**: Systematic review of each
2. **Take Notes**: Record observations per dimension
3. **Collect Evidence**: Identify supporting evidence
4. **Note Concerns**: Record questions and issues

### Phase 3: Dimension Evaluation

For each dimension:

1. **Examine Evidence**: Review collected evidence
2. **Compare to Criteria**: Match against scoring criteria
3. **Assign Score**: Determine score with justification
4. **Document Reasoning**: Record why score was assigned

### Phase 4: Cross-Document Analysis

1. **Consistency Check**: Verify cross-document consistency
2. **Terminology Audit**: Check for drift or inconsistency
3. **Dependency Mapping**: Verify document relationships
4. **Gap Identification**: Identify missing content

### Phase 5: Synthesis

1. **Calculate Overall Score**: Average of dimension scores
2. **Identify Themes**: Common issues across dimensions
3. **Prioritize Findings**: Rank by severity and impact
4. **Form Recommendations**: Actionable improvement suggestions

### Phase 6: Reporting

1. **Complete Template**: Follow AUDIT_TEMPLATE.md structure
2. **Document Evidence**: Cite specific evidence for findings
3. **Provide Recommendations**: Clear, actionable guidance
4. **Deliver Verdict**: Clear overall assessment

## Scoring Guidance

### Score Interpretation

| Score | Level | Interpretation |
|-------|-------|----------------|
| 0-2 | Concept | Early development, basic ideas |
| 2-4 | Defined | Documented but informal |
| 4-6 | Structured | Formalized processes |
| 6-8 | Usable | Applied in practice |
| 8-9 | Validated | Benefits demonstrated |
| 9-10 | Proven | Repeated success |

### Score Quality

**Good Scores:**
- Based on direct evidence
- Consistent across dimensions
- Documented justification
- Internally coherent

**Poor Scores:**
- Based on assumptions
- Inconsistently applied
- Lacking justification
- Contradicting evidence

## Deliverables

A Foundation Audit report must include:

1. **Metadata**: Standard audit metadata
2. **Executive Summary**: High-level findings
3. **Dimension Scores**: All 10 dimensions scored
4. **Evidence**: Supporting evidence for each finding
5. **Gap Analysis**: Identified gaps categorized
6. **Recommendations**: Prioritized improvement suggestions
7. **Final Verdict**: Overall assessment

## Review Process

### Self-Review

Before completing:

- [ ] All dimensions scored with justification
- [ ] Evidence cited for each finding
- [ ] Recommendations are actionable
- [ ] Scores are internally consistent
- [ ] Template structure followed

### Peer Review

If available:

- Second auditor reviews scoring
- Inconsistencies identified and resolved
- Additional evidence may be requested

## Recommendations

### Priority 1: Critical

Issues that prevent methodology from functioning:

- Missing core definitions
- Fundamental contradictions
- Broken cross-references
- Incomprehensible content

### Priority 2: High

Issues that significantly impact usability:

- Missing practical guidance
- Inconsistent terminology
- Incomplete processes
- Unclear scope

### Priority 3: Medium

Issues that affect quality:

- Minor inconsistencies
- Documentation improvements
- Additional examples
- Clarity enhancements

### Priority 4: Low

Nice-to-have improvements:

- Formatting improvements
- Navigation enhancements
- Cross-reference additions
- Glossary expansion

## Verification Evidence Requirements

### Purpose

KDSE distinguishes between test assets (plans, cases, documentation) and verification evidence (execution records, results). This section defines requirements for verification evidence in Foundation Audits.

### Verification Evidence States

During a Foundation Audit, the auditor must classify verification practices into one of four states:

| State | Definition | Score Impact |
|-------|------------|--------------|
| **Verified** | Tests executed with passing results | Full scoring range (0-10) |
| **Verified with Failures** | Tests executed with documented failures | Reduced scoring range (0-7) |
| **Not Verified** | Test assets exist but no execution evidence | Maximum score capped at 4/10 |
| **Not Assessed** | No verification artifacts | Maximum score capped at 2/10 |

### Critical Principle

> **The presence of test assets (test cases, test plans, test documentation) alone does NOT constitute verification evidence.**

This principle is enforced through:
1. **Evidence Classification**: Clear separation of test assets from execution evidence
2. **Score Caps**: Test assets-only verification is capped at 4/10
3. **Reporting Requirements**: Explicit status reporting for each verification category
4. **Risk Assessment**: Uncertainty reflected in risk levels

### Evidence Requirements for Verification Scoring

| Evidence Type | Required for Scoring? | Notes |
|--------------|---------------------|-------|
| Verification plans | No | Test Asset - planning only |
| Test cases | No | Test Asset - definition only |
| Test documentation | No | Test Asset - documentation only |
| Test execution records | **YES** | Execution Evidence - proves tests ran |
| Test results | **YES** | Execution Evidence - proves outcomes |
| CI/CD build logs | **YES** (can substitute) | Execution Evidence |

### Foundation Audit Verification Requirements

When auditing KDSE itself for verification practices:

1. **Check for verification artifacts**: Look for verification plans, test cases, test results
2. **Distinguish asset from evidence**: Separate test documentation from execution records
3. **Report explicit status**: State "Verified", "Not Verified", or "Not Assessed"
4. **Apply score caps**: Do not score above 4/10 if only test assets are present
5. **Document limitations**: Note if verification evidence was unavailable for review

### Verification Evidence Classification in Practice

**Example - Correct Classification:**

| Verification Category | Assets Exist | Execution Evidence | Status | Score Cap |
|---------------------|--------------|-------------------|--------|-----------|
| Unit Tests | Yes | Yes (CI logs + results) | Verified | 10/10 |
| Integration Tests | Yes | No | Not Verified | 4/10 |
| System Tests | No | No | Not Assessed | 2/10 |

**Example - Incorrect Classification (Before Fix):**

| Verification Category | Assets Exist | Execution Evidence | Status | Score Cap |
|---------------------|--------------|-------------------|--------|-----------|
| Unit Tests | Yes | No | PASS (incorrect) | 10/10 (incorrect) |

The incorrect example would violate KDSE's evidence-based principle by reporting "PASS" without execution evidence.

## Version

- **Document Version**: 1.1
- **Effective Date**: 2026-07-10
- **Standard Version**: KDSE Audit Standard 1.1
- **Change Note**: Added verification evidence classification requirements (KDSE-DEFECT-001)

---

*This standard is part of the KDSE Audit System. See [README.md](README.md) for audit system overview.*
