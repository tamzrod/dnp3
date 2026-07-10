# KDSE Session Protocol

**Document Version:** 1.0  
**Type:** Informative Reference Implementation  
**Effective Date:** 2026-07-10

---

## Purpose

This document defines the standard lifecycle for a KDSE development session. A session is a bounded period of engineering work guided by the KDSE Standard.

The Session Protocol provides a consistent structure for:
- Starting sessions
- Executing work
- Making decisions
- Verifying progress
- Completing sessions

---

## Session Definition

A **KDSE Session** is:

1. **Bounded**: Has defined start and end points
2. **Guided**: Follows the KDSE Standard
3. **Documented**: Produces Runtime Reports
4. **Measured**: Tracks compliance progress
5. **Authorized**: Requires human approval for implementation

---

## Session Phases

```
┌─────────────────────────────────────────────────────────────┐
│                     Session Lifecycle                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌───────────┐    ┌───────────┐    ┌───────────┐         │
│  │   Start   │───▶│ Execute  │───▶│ Complete  │         │
│  └───────────┘    └─────┬─────┘    └───────────┘         │
│                         │                                  │
│                         ▼                                  │
│                   ┌───────────┐                           │
│                   │  Verify   │                           │
│                   └───────────┘                           │
│                         │                                  │
│                         ▼                                  │
│                   ┌───────────┐                           │
│                   │ Continue  │◀──┐                       │
│                   │     or    │    │                      │
│                   │   Close   │────┘                      │
│                   └───────────┘                           │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Phase 1: Session Start

### 1.1 Initialize

**Purpose:** Establish session context

**Actions:**
1. Receive session parameters (repository, target, operator)
2. Load KDSE Standard references
3. Establish baseline context
4. Initialize session record

**Session Parameters:**

| Parameter | Required | Description |
|-----------|----------|-------------|
| repository | Yes | Target repository path or URL |
| target_maturity | Yes | Desired compliance score (0-10) |
| operator | Yes | Human responsible for approvals |
| scope | No | Specific dimensions to focus on |
| constraints | No | Resource limits, deadlines |

### 1.2 Verify Standards

**Purpose:** Confirm KDSE Standard is accessible

**Actions:**
1. Verify Foundation documents available
2. Confirm Audit templates accessible
3. Validate scoring criteria present
4. Check glossary terminology

**References:**
- [docs/foundation/](../docs/foundation/)
- [docs/audit/](../docs/audit/)

### 1.3 Assess Repository

**Purpose:** Evaluate current state

**Actions:**
1. Inventory repository artifacts
2. Map artifact relationships
3. Execute Compliance Audit
4. Document findings

**References:**
- [COMPLIANCE_AUDIT.md - All Dimensions](../docs/audit/COMPLIANCE_AUDIT.md)

---

## Phase 2: Session Execution

### 2.1 Generate Report

**Purpose:** Summarize assessment for the operator

**Actions:**
1. Create Runtime Report per [REPORT_SPEC.md](REPORT_SPEC.md)
2. Identify highest-priority findings
3. Calculate expected impact
4. Prepare recommendation

**Runtime Report Sections:**
- Current Status
- Foundation Status
- Compliance Status
- Summary of Findings
- Highest Priority Recommendation
- Expected Impact
- Required Approval
- Session State

### 2.2 Present Recommendation

**Purpose:** Enable operator decision

**Actions:**
1. Present Runtime Report
2. Explain recommendation rationale
3. Show expected impact
4. Await operator decision

**Recommendation Structure:**
```
RECOMMENDED ACTION: [Clear description]

EXPECTED IMPACT: [+X.X compliance points]

RATIONALE: [Why this action has highest value]

ALTERNATIVES CONSIDERED: [What else was evaluated]
```

### 2.3 Await Approval

**Purpose:** Pause for operator authorization

**Actions:**
1. Present decision options
2. Await operator response
3. Record decision
4. Proceed or reassess

**Decision Options:**

| Decision | Meaning | Next Action |
|----------|---------|-------------|
| APPROVE | Proceed with recommendation | Implement |
| APPROVE WITH MODIFICATIONS | Proceed with changes | Implement with changes |
| REJECT | Decline recommendation | Present alternative |
| DEFER | Postpone decision | Pause session |
| CLOSE | End session | Complete |

---

## Phase 3: Decision Points

### 3.1 If Approved

Proceed to implementation.

### 3.2 If Rejected

1. Identify next highest-value recommendation
2. Present alternative recommendation
3. Await new decision

### 3.3 If Deferred

1. Preserve session state
2. Await resumption
3. Resume from deferred point

### 3.4 If Closed

Proceed to session completion.

---

## Phase 4: Implementation

### 4.1 Prepare Implementation

**Purpose:** Ready for execution

**Actions:**
1. Review implementation constraints
2. Confirm preconditions
3. Prepare implementation environment
4. Assign responsibilities

### 4.2 Execute Implementation

**Purpose:** Perform the approved action

**Actions:**
1. Follow KDSE Engineering Model
2. Maintain authority hierarchy
3. Document decisions
4. Create or update artifacts
5. Maintain traceability

**Constraints:**
- Must respect [Chain of Authority](../docs/foundation/006-chain-of-authority.md)
- Must follow [Engineering Model](../docs/foundation/004-engineering-model.md)
- Must maintain traceability

### 4.3 Document Changes

**Purpose:** Record work performed

**Actions:**
1. Record all modifications
2. Document deviations (if any)
3. Update artifact references
4. Maintain evidence trail

---

## Phase 5: Verification

### 5.1 Re-assess

**Purpose:** Confirm results through re-audit

**Actions:**
1. Re-run Compliance Audit
2. Compare to baseline
3. Document improvements
4. Identify new gaps (if any)

### 5.2 Measure Progress

**Purpose:** Quantify advancement toward target

**Actions:**
1. Calculate score improvement
2. Assess findings addressed
3. Evaluate progress to target
4. Assess diminishing returns

**Progress Calculation:**
```
Progress = (Current Score - Baseline Score) / (Target Score - Baseline Score) × 100%
```

### 5.3 Generate Updated Report

**Purpose:** Document current state

**Actions:**
1. Update Runtime Report
2. Include new recommendations (if applicable)
3. Record metrics

---

## Phase 6: Session Continuation

### 6.1 Evaluate Next Steps

**Purpose:** Determine whether to continue

**Decision Matrix:**

| Condition | Action |
|-----------|--------|
| Target maturity reached | Complete session |
| No actionable findings | Complete session |
| Diminishing returns threshold | Complete session |
| More high-value actions | Continue session |
| Operator requests close | Complete session |

### 6.2 If Continuing

1. Return to Phase 2.1 (Generate Report)
2. Present next recommendation
3. Await approval
4. Continue cycle

### 6.3 If Completing

Proceed to session completion.

---

## Phase 7: Session Completion

### 7.1 Finalize Report

**Purpose:** Close documentation

**Actions:**
1. Complete final Runtime Report
2. Record session metrics
3. Archive session artifacts
4. Update maturity records

### 7.2 Record Metrics

**Required Metrics:**

| Metric | Description |
|--------|-------------|
| session_duration | Total session time |
| iterations_completed | Number of assess-implement-verify cycles |
| actions_approved | Count of approved actions |
| actions_rejected | Count of rejected recommendations |
| score_start | Baseline compliance score |
| score_end | Final compliance score |
| improvement | score_end - score_start |
| target_reached | Boolean |

### 7.3 Close Session

**Purpose:** Formally end session

**Actions:**
1. Mark session complete
2. Notify operator
3. Provide completion summary
4. Preserve session for reference

---

## Session States Summary

| State | Description |
|-------|-------------|
| INITIALIZING | Loading standards and establishing context |
| VERIFYING | Running Foundation Verification |
| ASSESSING | Running Repository Assessment |
| REPORTING | Generating Runtime Report |
| PENDING_APPROVAL | Awaiting operator decision |
| IMPLEMENTING | Executing approved work |
| VERIFYING_RESULTS | Confirming implementation |
| COMPLETING | Finalizing session |
| CLOSED | Session ended |

---

## State Transitions

```
INITIALIZING → VERIFYING → ASSESSING → REPORTING → PENDING_APPROVAL
                                     ↓
                                     ├─ APPROVED → IMPLEMENTING → VERIFYING_RESULTS
                                     │                                        │
                                     │                                        ├─ Continue → REPORTING
                                     │                                        │
                                     │                                        └─ Complete → COMPLETING → CLOSED
                                     │
                                     ├─ REJECTED → REPORTING
                                     │
                                     ├─ DEFERRED → (await resume)
                                     │
                                     └─ CLOSED → COMPLETING → CLOSED
```

---

## Example Session Flow

### Example: Repository Maturation

**Context:** Empty repository, target maturity 6.0

```
[Session Start]
  → Initialize session
  → Verify standards
  → Assess repository (Baseline: 2.0/10)

[Iteration 1]
  → Generate Runtime Report
  → Recommend: Create knowledge artifacts
  → Operator: APPROVE
  → Implement knowledge artifacts
  → Verify results (Score: 3.2/10)

[Iteration 2]
  → Generate Runtime Report
  → Recommend: Define architecture approach
  → Operator: APPROVE
  → Implement architecture
  → Verify results (Score: 4.5/10)

[Iterations 3-6]
  → Continue cycle
  → Each iteration improves scores

[Iteration 7]
  → Generate Runtime Report
  → Verify results (Score: 6.1/10)
  → Target reached → COMPLETE
```

### Example: Targeted Improvement

**Context:** Repository at 6.8/10, gap in verification practices

```
[Session Start]
  → Initialize with scope: verification
  → Assess (focused on verification dimension)

[Single Iteration]
  → Generate Runtime Report
  → Recommend: Implement verification criteria
  → Operator: APPROVE
  → Implement verification criteria
  → Verify results (improvement confirmed)

[Session Complete]
  → Record improvement
  → Close session
```

---

## Principles

The Session Protocol follows KDSE principles:

| Principle | Protocol Application |
|-----------|---------------------|
| Evidence-Based | Recommendations trace to audit findings |
| Traceability | All decisions documented |
| Authority Hierarchy | Implementation follows hierarchy |
| Human Authorization | Approval gates at each step |
| Progress Measurement | Scores tracked throughout |

---

## References

- [EXECUTION_MODEL.md](EXECUTION_MODEL.md) - Runtime lifecycle
- [REPORT_SPEC.md](REPORT_SPEC.md) - Runtime Report format
- [COMPLIANCE_AUDIT.md](../docs/audit/COMPLIANCE_AUDIT.md) - Audit standard
- [FOUNDATION_AUDIT.md](../docs/audit/FOUNDATION_AUDIT.md) - Foundation verification
- [ENGINEERING_MODEL.md](../docs/foundation/004-engineering-model.md) - Engineering stages

---

*This document is an informative reference implementation. It provides a template for KDSE sessions, not a requirement.*
