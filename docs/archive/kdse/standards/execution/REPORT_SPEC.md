# KDSE Runtime Report Specification

**Document Version:** 1.0  
**Type:** Informative Reference Implementation  
**Effective Date:** 2026-07-10

---

## Purpose

The Runtime Report summarizes KDSE assessment results for the operator. It presents audit findings in an actionable format, enabling informed decisions about next steps.

**Important:** The Runtime Report summarizes audits. It does not replace them.

---

## Relationship to Audits

```
┌─────────────────────────────────────────────────────────────┐
│                 KDSE Audit Documents                        │
│                                                             │
│  • FOUNDATION_AUDIT.md - Foundation verification           │
│  • COMPLIANCE_AUDIT.md - Repository compliance              │
│  • AUDIT_SCORING.md - Score calculation                    │
│                                                             │
│  ⚠️ These are the authoritative sources.                    │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ Summarized by
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Runtime Report                            │
│                                                             │
│  • Current Status (summary)                                 │
│  • Compliance Scores (from audits)                         │
│  • Findings Summary (from audits)                          │
│  • Recommendations (derived from findings)                 │
│  • Expected Impact (calculated)                            │
│                                                             │
│  ℹ️ This report summarizes audits for operators.             │
└─────────────────────────────────────────────────────────────┘
```

---

## Report Structure

```
┌─────────────────────────────────────────────────────────────┐
│                   RUNTIME REPORT                             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. Header                                                  │
│     - Session ID                                            │
│     - Timestamp                                             │
│     - Repository                                            │
│     - Operator                                              │
│                                                             │
│  2. Current Status                                          │
│     - Executive Summary                                     │
│     - Quick Metrics                                        │
│                                                             │
│  3. Foundation Status                                       │
│     - Standards Verification                                │
│                                                             │
│  4. Compliance Status                                       │
│     - Dimension Scores Table                                │
│     - Overall Score                                         │
│                                                             │
│  5. Summary of Findings                                     │
│     - Critical Findings                                     │
│     - High Findings                                         │
│     - Medium/Low Findings                                   │
│                                                             │
│  6. Highest Priority Recommendation                         │
│     - Recommended Action                                    │
│     - Rationale                                            │
│     - Expected Impact                                       │
│                                                             │
│  7. Required Approval                                       │
│     - Decision Options                                      │
│     - Deadline (if any)                                     │
│                                                             │
│  8. Session State                                           │
│     - Current Phase                                         │
│     - Progress                                              │
│     - Next Steps                                            │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Section Specifications

### 1. Header

```markdown
# Runtime Report

| Field | Value |
|-------|-------|
| Report ID | KDSE-RT-{session_id} |
| Generated | {ISO 8601 timestamp} |
| Repository | {repository path or URL} |
| Operator | {operator name} |
| Target Maturity | {target}/10 |
| Report Version | 1.0 |
```

**Field Descriptions:**

| Field | Source | Required |
|-------|--------|----------|
| Report ID | Runtime generated | Yes |
| Generated | System timestamp | Yes |
| Repository | Session parameter | Yes |
| Operator | Session parameter | Yes |
| Target Maturity | Session parameter | Yes |
| Report Version | Specification (always 1.0) | Yes |

---

### 2. Current Status

```markdown
## Current Status

### Executive Summary

{3-5 sentence summary of:
- Current compliance state
- Key findings
- Recommended direction
}

### Quick Metrics

| Metric | Value | Change |
|--------|-------|--------|
| Overall Score | {X}/10 ({level}) | {+/-delta} |
| Highest Dimension | {name} ({X}/10) | — |
| Lowest Dimension | {name} ({X}/10) | — |
| Critical Findings | {count} | {change} |
| High Findings | {count} | {change} |
```

**Maturity Levels** (per [AUDIT_SCORING.md](../docs/audit/AUDIT_SCORING.md)):

| Score Range | Level |
|-------------|-------|
| 0-2 | Concept |
| 2-4 | Defined |
| 4-6 | Structured |
| 6-8 | Usable |
| 8-9 | Validated |
| 9-10 | Proven |

---

### 3. Foundation Status

```markdown
## Foundation Status

### Standards Verification

| Check | Status |
|-------|--------|
| Foundation Documents | ✅ Available |
| Audit Templates | ✅ Available |
| Scoring Criteria | ✅ Available |
| Glossary | ✅ Available |

### Standards Reference

For details, see:
- [KDSE Foundation](../docs/foundation/)
- [Audit System](../docs/audit/)
```

**Note:** This section confirms the KDSE Standard is accessible and consistent. It does not perform a full Foundation Audit.

---

### 4. Compliance Status

```markdown
## Compliance Status

### Dimension Scores

| Dimension | Score | Status |
|-----------|-------|--------|
| 1. Knowledge Artifacts | {X}/10 | {indicator} |
| 2. Architecture Artifacts | {X}/10 | {indicator} |
| 3. Implementation Artifacts | {X}/10 | {indicator} |
| 4. Verification Practices | {X}/10 | {indicator} |
| 5. Traceability | {X}/10 | {indicator} |
| 6. Authority Hierarchy | {X}/10 | {indicator} |
| 7. Governance | {X}/10 | {indicator} |
| **Overall Score** | **{X}/10** | **{indicator}** |

### Score Indicators

- ✅ Meeting target (within 1 point)
- ⚠️ Below target (1-2 points below)
- ❌ Significantly below target (>2 points below)
- 🎯 Target reached

### Compliance Level

**Current Level:** {level} (Score: {X}/10)

Per [COMPLIANCE_AUDIT.md](../docs/audit/COMPLIANCE_AUDIT.md):

| Level | Score Range |
|-------|-------------|
| Compliant | 8-10 |
| Substantially Compliant | 6-8 |
| Partially Compliant | 4-6 |
| Minimally Compliant | 2-4 |
| Non-Compliant | 0-2 |

### Score Comparison

{If previous report exists:}

```
Dimension          Previous   Current   Change
───────────────────────────────────────────────
Knowledge              5.0       5.5     +0.5
Architecture           4.0       4.5     +0.5
...
───────────────────────────────────────────────
Overall                4.7       5.1     +0.4
```

**Reference:** [COMPLIANCE_AUDIT.md - Dimensions](../docs/audit/COMPLIANCE_AUDIT.md)

---

### 5. Summary of Findings

```markdown
## Summary of Findings

### Critical Findings ({count})

{For each critical finding:}

**{Finding Title}**

- **Dimension:** {affected dimension}
- **Evidence:** {brief evidence description}
- **Impact:** {what this prevents or causes}
- **Recommendation:** {how to address}

---

### High Findings ({count})

{For each high finding:}

**{Finding Title}**

- **Dimension:** {affected dimension}
- **Evidence:** {brief evidence description}
- **Impact:** {what this affects}
- **Recommendation:** {how to address}

---

### Other Findings ({count})

| Finding | Dimension | Severity |
|---------|-----------|----------|
| {title} | {dimension} | Medium/Low |

---

**Reference:** [COMPLIANCE_AUDIT.md - Gap Analysis](../docs/audit/COMPLIANCE_AUDIT.md)
```

---

### 6. Highest Priority Recommendation

```markdown
## Highest Priority Recommendation

### Recommended Action

| Field | Value |
|-------|-------|
| Action ID | KDSE-ACT-{sequential_id} |
| Priority | {1-Critical | 2-High | 3-Medium} |
| Target Dimension | {dimension} |
| Expected Impact | +{X.Y} points |

### Action Description

**What:** {explicit description of action to take}

**Why:** {rationale connecting to findings and impact}

**How:** {implementation approach in steps}

### Rationale

{Explanation of why this action has highest value:
- Which finding(s) it addresses
- Why it enables downstream improvements
- What alternatives were considered
}

### Expected Impact

| Dimension | Current | Projected | Change |
|-----------|---------|-----------|--------|
| {dimension} | {X}/10 | {Y}/10 | +{Z} |
| ... | ... | ... | ... |
| **Overall** | **{X}/10** | **{Y}/10** | **+{Z}** |

### Progress to Target

```
Current:  {X}/10
Target:    {Y}/10
Progress:  {Z}% complete
Remaining: {Y-X} points
```

### Constraints

- **Authority Level:** {Knowledge | Architecture | Implementation}
- **Traceability Required:** {Yes | No}
- **Steward Approval:** {Yes | No}

### Preconditions

Before implementing:
1. {prerequisite 1}
2. {prerequisite 2}

### Postconditions

After implementing:
1. {expected outcome 1}
2. {expected outcome 2}

### Alternatives Considered

| Alternative | Why Not Chosen |
|-------------|----------------|
| {alternative 1} | {reason} |
| {alternative 2} | {reason} |

---

**Note:** This recommendation is derived from audit findings. The operator decides whether to approve.
```

---

### 7. Required Approval

```markdown
## Required Approval

### Approval Request

**To:** {operator name}
**From:** KDSE Runtime
**Request:** Approve implementation of recommended action

### Recommended Action

{Repeat recommended action summary}

### Expected Impact

{Repeat expected impact summary}

### Decision Options

| Option | Meaning |
|--------|---------|
| **APPROVE** | Proceed with recommended action |
| **APPROVE WITH MODIFICATIONS** | Proceed with specified changes |
| **REJECT** | Decline; alternative will be provided |
| **DEFER** | Postpone to later session |
| **CLOSE SESSION** | End session without implementation |

### Deadline

{If applicable: "Decision requested by {deadline}"}
{If not: "Decision at your discretion"}

### Approval Format

```
APPROVAL: [APPROVE | APPROVE WITH MODIFICATIONS | REJECT | DEFER | CLOSE]
REASON: {your rationale, required for non-approve}
MODIFICATIONS: {if APPROVE WITH MODIFICATIONS}
OPERATOR: {your name}
TIMESTAMP: {ISO 8601}
```

### Accountability

By approving, you acknowledge:
1. The action is derived from audit evidence
2. Expected impact has been assessed
3. You accept responsibility for this decision
```

---

### 8. Session State

```markdown
## Session State

### Current Phase

| Field | Value |
|-------|-------|
| Phase | {current phase} |
| Iteration | {N} of unlimited |
| Duration | {elapsed time} |

### Progress

```
[✓] Initialize
[✓] Verify Standards
[✓] Assess Repository
[✓] Generate Report
[→] Await Approval
[ ] Implement
[ ] Verify
[ ] Complete or Continue
```

### Session History

| Iteration | Action | Approved | Score Change |
|-----------|--------|----------|--------------|
| 1 | {action 1} | {yes/no} | +{X.Y} |
| 2 | {action 2} | {yes/no} | +{X.Y} |
| ... | ... | ... | ... |

### Next Steps

**If Approved:**
1. Implement recommended action
2. Re-run assessment
3. Generate updated report
4. Continue or complete

**If Rejected:**
1. Identify next recommendation
2. Present alternative
3. Await new approval

**If Closed:**
1. Finalize report
2. Record metrics
3. Complete session

### Human Decision Required

**Action Needed:** Approval on recommended action

**Deadline:** {if applicable}
```

---

## Appendix: Audit Summary

```markdown
## Appendix: Audit Summary

### Full Dimension Breakdown

{Detailed audit results per COMPLIANCE_AUDIT.md}

### Findings Index

| ID | Dimension | Severity | Title |
|----|-----------|----------|-------|
| F001 | {dimension} | Critical | {title} |
| F002 | {dimension} | High | {title} |
| ... | ... | ... | ... |

### Evidence Index

| Evidence ID | Source | Finding(s) |
|-------------|--------|------------|
| E001 | {path} | F001, F002 |
| E002 | {path} | F003 |
| ... | ... | ... |

### Audit Methodology

- **Audit Reference:** KDSE Compliance Audit 1.0
- **Audit Date:** {timestamp}
- **Runtime Version:** {version}
- **Repository Version:** {hash}

---

**References:**
- [COMPLIANCE_AUDIT.md](../docs/audit/COMPLIANCE_AUDIT.md)
- [AUDIT_SCORING.md](../docs/audit/AUDIT_SCORING.md)
```

---

## Report Completion Checklist

Before finalizing a Runtime Report:

- [ ] Header complete with all required fields
- [ ] Current status reflects assessment accurately
- [ ] Compliance scores match audit results
- [ ] Findings prioritized by severity
- [ ] Recommendation justified with evidence
- [ ] Expected impact calculated
- [ ] Approval request clear
- [ ] Session state accurate
- [ ] References to Standard documents included

---

## Example Runtime Report

```markdown
# Runtime Report

| Field | Value |
|-------|-------|
| Report ID | KDSE-RT-2026-07-10-001 |
| Generated | 2026-07-10T14:30:00Z |
| Repository | /example/project |
| Operator | Jane Developer |
| Target Maturity | 7.0/10 |
| Report Version | 1.0 |

## Current Status

**Overall Score:** 5.2/10 (Structured)

Critical findings: 1 | High findings: 3

## Compliance Status

| Dimension | Score | Status |
|-----------|-------|--------|
| Knowledge | 3.5/10 | ❌ |
| Architecture | 5.0/10 | ⚠️ |
| Implementation | 5.5/10 | ⚠️ |
| Verification | 4.0/10 | ❌ |
| Traceability | 5.5/10 | ⚠️ |
| Authority | 6.5/10 | ✅ |
| Governance | 5.5/10 | ⚠️ |
| **Overall** | **5.2/10** | **⚠️** |

## Highest Priority Recommendation

**Action:** Create structured knowledge artifacts for core requirements

**Expected Impact:** +1.2 points (5.2 → 6.4)

## Required Approval

APPROVAL: [APPROVE | REJECT | DEFER | CLOSE]
OPERATOR: Jane Developer
```

---

## Document Relationships

```
REPORT_SPEC.md (this document)
    │
    ├── Defines: Runtime Report structure
    │
    ├── References:
    │   ├── COMPLIANCE_AUDIT.md
    │   ├── AUDIT_SCORING.md
    │   └── FOUNDATION_AUDIT.md
    │
    └── Used by:
        ├── SESSION_PROTOCOL.md
        └── EXECUTION_MODEL.md
```

---

*This document is an informative reference implementation. It defines the Runtime Report format, not KDSE requirements.*
