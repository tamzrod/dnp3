# KDSE Audit Report Template

## Purpose

This template defines the required structure for all KDSE audit reports. Every KDSE audit, whether foundation audit or compliance audit, must follow this template.

## When to Use This Template

Use this template for:
- Foundation Audits of KDSE
- Compliance Audits of KDSE-adopting repositories
- Any formal KDSE evaluation

## Template Structure

```
[Audit Report Title]
[Audit Metadata]
[Executive Summary]
[Audit Details]
  [Dimension 1]
  [Dimension 2]
  ...
[Findings]
[Gap Analysis]
[Recommendations]
[Final Verdict]
[Appendices]
```

---

# AUDIT REPORT

## Audit Metadata

```
Standard: KDSE Audit Standard
Standard Version: 1.0
Audit Type: [Foundation | Compliance]
Repository: [Name or "KDSE"]
Repository URL: [URL or "N/A for KDSE self-audit"]
Repository Version: [Commit hash, branch, or date]
Audit Date: [YYYY-MM-DD]
Auditor: [Name and role]
Audit Context: [Background information]
Previous Audit: [Date and reference, if applicable]
```

### Metadata Definitions

**Standard Version:**
The version of the KDSE Audit Standard used for this audit. Current version is 1.0.

**Audit Type:**
- **Foundation**: Audit of the KDSE methodology itself
- **Compliance**: Audit of a repository claiming KDSE compliance

**Repository:**
For foundation audits, use "KDSE". For compliance audits, use the repository name.

**Repository Version:**
For foundation audits, use the git commit hash or version tag. For compliance audits, use the specific commit or version audited.

**Audit Date:**
Date the audit was completed in ISO format (YYYY-MM-DD).

**Auditor:**
Name and role of the auditor. For self-audits, indicate "Internal Audit" or similar.

**Audit Context:**
Brief description of why this audit is being conducted.

**Previous Audit:**
Reference to any previous audit of the same target, if applicable.

---

## Executive Summary

Provide a high-level overview of the audit results. This section should be readable by stakeholders who may not read the full report.

### Required Elements

1. **Verdict Statement**: Clear statement of overall finding
2. **Overall Score**: Prominently displayed score with scale
3. **Key Findings**: 3-5 bullet points summarizing main findings
4. **Critical Gaps**: Any issues requiring immediate attention
5. **Next Steps**: Recommended immediate actions

### Template

```
**Verdict: [Overall finding statement]**

**Overall Score: X.X / 10**

### Key Findings

- [Finding 1]
- [Finding 2]
- [Finding 3]

### Critical Gaps

- [Gap 1, if any]
- [Gap 2, if any]

### Recommended Next Steps

1. [Action 1]
2. [Action 2]
3. [Action 3]
```

---

## Audit Details

This section contains the detailed evaluation of each audit dimension.

### Dimension Template

For each dimension, include:

```
## [Dimension Name]

**Score: X/10**

### Assessment

[Brief assessment of this dimension]

### Evidence

[Specific evidence supporting the assessment]
[List of artifacts examined]
[Key observations]

### Strengths

- [Strength 1]
- [Strength 2]

### Weaknesses

- [Weakness 1]
- [Weakness 2]

### Score Justification

[Explanation of why this score was assigned]
[Comparison to defined criteria]
[Evidence supporting this level]

### Recommendations for This Dimension

- [Recommendation 1]
- [Recommendation 2]
```

---

## Foundation Audit Dimensions

For foundation audits, evaluate these dimensions:

| Dimension | Description | Weight |
|-----------|-------------|--------|
| Identity | Clarity of methodology identity | High |
| Vision | Clarity of long-term direction | Medium |
| Repository Structure | Organization of documentation | Medium |
| Body of Knowledge | Completeness of content | High |
| Engineering Philosophy | Internal consistency | High |
| Terminology | Consistency of definitions | High |
| Traceability | Traceability framework | High |
| Practicality | Ease of application | Medium |
| Scalability | Applicability at scale | Medium |
| Independence | Technology/vendor neutrality | Medium |

### Score Summary Table

```
| Dimension | Score | Previous | Change |
|-----------|-------|----------|--------|
| Identity | X/10 | X/10 | +/-X |
| Vision | X/10 | X/10 | +/-X |
| Repository Structure | X/10 | X/10 | +/-X |
| Body of Knowledge | X/10 | X/10 | +/-X |
| Engineering Philosophy | X/10 | X/10 | +/-X |
| Terminology | X/10 | X/10 | +/-X |
| Traceability | X/10 | X/10 | +/-X |
| Practicality | X/10 | X/10 | +/-X |
| Scalability | X/10 | X/10 | +/-X |
| Independence | X/10 | X/10 | +/-X |
| **Overall** | **X.X/10** | **X.X/10** | **+/-X.X** |
```

---

## Compliance Audit Dimensions

For compliance audits, evaluate these dimensions:

| Dimension | Description | Weight |
|-----------|-------------|--------|
| Knowledge Artifacts | Quality of knowledge artifacts | High |
| Architecture Artifacts | Quality of architecture artifacts | High |
| Implementation Artifacts | Quality of implementation | High |
| Verification Practices | Verification implementation | High |
| Traceability | Traceability maintenance | High |
| Authority Hierarchy | Authority implementation | High |
| Governance | Governance practices | Medium |

### Score Summary Table

```
| Dimension | Score |
|-----------|-------|
| Knowledge Artifacts | X/10 |
| Architecture Artifacts | X/10 |
| Implementation Artifacts | X/10 |
| Verification Practices | X/10 |
| Traceability | X/10 |
| Authority Hierarchy | X/10 |
| Governance | X/10 |
| **Overall** | **X.X/10** |
```

---

## Findings

### Critical Findings

Findings that must be addressed:

```
### Critical Finding [Number]

**Severity:** Critical
**Dimension:** [Affected dimension]
**Description:** [Clear description]

**Evidence:**
[Specific evidence]

**Impact:**
[What this means]

**Required Action:**
[What must be done]
```

### Major Findings

Findings that should be addressed:

```
### Major Finding [Number]

**Severity:** Major
**Dimension:** [Affected dimension]
**Description:** [Clear description]

**Evidence:**
[Specific evidence]

**Impact:**
[What this means]

**Recommended Action:**
[What should be done]
```

### Minor Findings

Findings that are opportunities for improvement:

```
### Minor Finding [Number]

**Severity:** Minor
**Dimension:** [Affected dimension]
**Description:** [Clear description]

**Evidence:**
[Specific evidence]

**Suggested Action:**
[What could be done]
```

---

## Gap Analysis

### Gap Summary

| Gap | Severity | Impact | Priority |
|-----|----------|--------|----------|
| [Gap 1] | Critical/Major/Minor | High/Medium/Low | High/Medium/Low |
| [Gap 2] | Critical/Major/Minor | High/Medium/Low | High/Medium/Low |

### Detailed Gap Analysis

For each significant gap:

```
### Gap: [Gap Name]

**Source:** [Where the gap was identified]
**Evidence:** [What demonstrates the gap]
**Affected Dimensions:** [List of affected areas]
**Impact Analysis:** [What this prevents or causes]
**Severity Assessment:** [Why this severity is appropriate]
```

---

## Recommendations

### Priority 1: Critical (Must Address)

```
### Recommendation [Number]

**What:** [What should be done]
**Why:** [Why this is critical]
**How:** [How to accomplish this]
**Timeline:** [When this should be complete]
**Success Criteria:** [How to know when complete]
```

### Priority 2: High (Should Address)

```
### Recommendation [Number]

**What:** [What should be done]
**Why:** [Why this matters]
**How:** [How to accomplish this]
**Timeline:** [When this should be complete]
**Success Criteria:** [How to know when complete]
```

### Priority 3: Medium (Consider Addressing)

```
### Recommendation [Number]

**What:** [What could be done]
**Why:** [Why this would help]
**How:** [How to accomplish this]
**Effort:** [Estimated effort]
**Benefit:** [Expected benefit]
```

---

## Final Verdict

### Verdict Statement

Clear statement of the overall audit finding:

```
**VERDICT: [VERDICT TEXT]**

[Detailed explanation of verdict]
```

### Compliance Levels (For Compliance Audits)

| Level | Score Range | Meaning |
|-------|-------------|---------|
| Compliant | 8-10 | Fully compliant with KDSE |
| Substantially Compliant | 6-8 | Compliant with minor gaps |
| Partially Compliant | 4-6 | Significant gaps exist |
| Non-Compliant | 0-4 | Does not meet KDSE standards |

### Readiness Assessment (For Foundation Audits)

| Level | Score Range | Meaning |
|-------|-------------|---------|
| Ready for Release | 8-10 | Can be released |
| Ready with Notes | 6-8 | Can be released with documented gaps |
| Needs Work | 4-6 | Significant work required |
| Not Ready | 0-4 | Major issues must be addressed |

---

## Appendices

### Appendix A: Evidence Index

Complete list of all evidence examined:

```
| Evidence ID | Description | Source | Date |
|------------|-------------|--------|------|
| E001 | [Description] | [File/URL] | [Date] |
| E002 | [Description] | [File/URL] | [Date] |
```

### Appendix B: Terminology

Any terminology used in the audit that may be unclear:

```
| Term | Definition |
|------|-------------|
| [Term] | [Definition] |
```

### Appendix C: Methodology

Description of how the audit was conducted:

```
**Audit Method:** [How audit was performed]
**Evidence Collection:** [How evidence was gathered]
**Scoring Method:** [How scores were determined]
**Review Process:** [How findings were validated]
```

### Appendix D: Previous Audit Comparison

For audits with previous versions:

```
| Dimension | Previous | Current | Change |
|-----------|----------|---------|--------|
| [Dimension] | X/10 | X/10 | +/-X |

**Notable Changes:**
- [Change 1]
- [Change 2]
```

---

## Checklist

Before submitting an audit report, verify:

- [ ] All metadata fields are complete
- [ ] Executive summary is present and accurate
- [ ] All dimensions are scored with justification
- [ ] All findings have supporting evidence
- [ ] Recommendations are prioritized
- [ ] Final verdict is clearly stated
- [ ] Appendices contain required materials
- [ ] Scores are internally consistent
- [ ] Report follows this template structure

---

## Version History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-07-10 | KDSE | Initial template |

---

*This template is part of the KDSE Audit Standard. See [README.md](README.md) for audit system overview.*
