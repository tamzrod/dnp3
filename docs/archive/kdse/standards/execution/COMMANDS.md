# KDSE Runtime Command Interface

**Document Version:** 1.0  
**Type:** Informative Reference Implementation  
**Effective Date:** 2026-07-10

---

## Purpose

This document defines the KDSE Runtime command interface. Commands provide a technology-neutral way to interact with the Runtime, regardless of implementation (human, AI assistant, CLI, CI pipeline, etc.).

---

## Command Overview

### Command Categories

| Category | Commands | Description |
|----------|----------|-------------|
| Session Management | Run KDSE, Continue KDSE, Close KDSE | Start, resume, end sessions |
| Session Control | Pause KDSE, Resume KDSE | Pause and resume sessions |
| Session Information | KDSE Status, KDSE Report, KDSE Progress | View session state |
| Query | KDSE Scores, KDSE Findings, KDSE History | Query session data |
| Decision | Approve, Reject, Defer | Operator decisions |

### Command Summary Table

| Command | Category | Required | Participants |
|---------|----------|----------|--------------|
| Run KDSE | Session | Yes | All |
| Continue KDSE | Session | Yes | All |
| Close KDSE | Session | Yes | All |
| Pause KDSE | Control | No | Human only |
| Resume KDSE | Control | No | All |
| KDSE Status | Information | Yes | All |
| KDSE Report | Information | Yes | All |
| KDSE Progress | Information | Yes | All |
| KDSE Scores | Query | No | All |
| KDSE Findings | Query | No | All |
| KDSE History | Query | No | All |
| Approve | Decision | Yes | Human only |
| Reject | Decision | Yes | Human only |
| Defer | Decision | Yes | Human only |

---

## Session Management Commands

### Run KDSE

Starts a new KDSE session.

**Purpose:** Initialize and begin a new KDSE session against a target repository.

**Inputs:**

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| repository | Yes | String | Target repository path or URL |
| target | Yes | Number | Target compliance score (0-10) |
| operator | Yes | String | Human responsible for decisions |
| scope | No | Array | Specific dimensions to focus on |
| constraints | No | Object | Resource limits, deadlines |

**Outputs:**

| Output | Type | Description |
|--------|------|-------------|
| session_id | String | Unique session identifier |
| state | String | Initial session state |
| baseline_score | Number | Initial compliance score |

**Preconditions:**
- Runtime is in Idle state
- KDSE Standard is accessible
- Target repository is accessible

**Postconditions:**
- Runtime is in Loading state
- Session ID generated
- Standard documents loaded

**State Transition:**
```
Idle → Loading → Verifying → Assessing → Reporting → Pending Approval
```

**Example Usage:**

Minimal:
```
Run KDSE

Repository: /workspace/project/myapp
Target: 7.0
Operator: Jane Developer
```

With scope:
```
Run KDSE

Repository: /workspace/project/myapp
Target: 7.0
Operator: Jane Developer
Scope: [Knowledge, Verification]
```

---

### Continue KDSE

Resumes an existing or paused session.

**Purpose:** Resume a previously paused or initialized session.

**Inputs:**

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| session_id | Yes | String | Session to resume |

**Outputs:**

| Output | Type | Description |
|--------|------|-------------|
| session_id | String | Resumed session ID |
| state | String | Current session state |
| current_report | Object | Latest Runtime Report |

**Preconditions:**
- Session exists and is not Closed
- Session is in Paused, Pending Approval, or Initializing state

**Postconditions:**
- Session state restored
- Runtime continues from saved state

**State Transition:**
```
Current State → (Continue) → Last Valid State
```

**Example Usage:**
```
Continue KDSE

Session ID: KDSE-RT-2026-07-10-001
```

---

### Close KDSE

Ends the current session.

**Purpose:** Formally close a KDSE session and record final metrics.

**Inputs:**

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| session_id | Yes | String | Session to close |
| reason | No | String | Reason for closing |

**Outputs:**

| Output | Type | Description |
|--------|------|-------------|
| session_id | String | Closed session ID |
| final_score | Number | Final compliance score |
| improvement | Number | Total score improvement |
| summary | Object | Session summary metrics |

**Preconditions:**
- Session exists and is not Closed

**Postconditions:**
- Session state is Closed
- Final metrics recorded
- Session artifacts archived

**State Transition:**
```
Any State → Closed
```

**Example Usage:**
```
Close KDSE

Session ID: KDSE-RT-2026-07-10-001
Reason: Target maturity reached
```

---

## Session Control Commands

### Pause KDSE

Pauses the current session.

**Purpose:** Temporarily suspend a session while maintaining state.

**Inputs:**

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| session_id | Yes | String | Session to pause |
| reason | No | String | Reason for pausing |

**Outputs:**

| Output | Type | Description |
|--------|------|-------------|
| session_id | String | Paused session ID |
| state | String | Paused state |
| resume_point | String | Where to resume |

**Preconditions:**
- Session exists and is active
- Session is in Verifying, Assessing, Reporting, or Pending Approval state

**Postconditions:**
- Session state is Paused
- All session data preserved
- Can be resumed with Continue KDSE

**State Transition:**
```
(Verifying | Assessing | Reporting | Pending Approval) → Paused
```

**Example Usage:**
```
Pause KDSE

Session ID: KDSE-RT-2026-07-10-001
Reason: Waiting for architecture decision
```

---

### Resume KDSE

Resumes a paused session.

**Purpose:** Continue a previously paused session.

**Inputs:**

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| session_id | Yes | String | Session to resume |

**Outputs:**

| Output | Type | Description |
|--------|------|-------------|
| session_id | String | Resumed session ID |
| state | String | Current state before pause |
| resume_point | String | Where session continues |

**Preconditions:**
- Session exists and is in Paused state
- Resume point is valid

**Postconditions:**
- Session continues from pause point
- Runtime state restored

**State Transition:**
```
Paused → (Resume Point State)
```

**Example Usage:**
```
Resume KDSE

Session ID: KDSE-RT-2026-07-10-001
```

---

## Session Information Commands

### KDSE Status

Displays the current session state.

**Purpose:** View the current state of a session without changing it.

**Inputs:**

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| session_id | Yes | String | Session to query |

**Outputs:**

| Output | Type | Description |
|--------|------|-------------|
| session_id | String | Session identifier |
| state | String | Current state |
| phase | String | Current phase |
| iteration | Number | Current iteration count |
| duration | String | Elapsed time |
| baseline_score | Number | Initial score |
| current_score | Number | Current score |
| target_score | Number | Target score |

**Preconditions:**
- Session exists

**Postconditions:**
- None (read-only)

**Example Usage:**
```
KDSE Status

Session ID: KDSE-RT-2026-07-10-001
```

**Example Output:**
```
Session ID: KDSE-RT-2026-07-10-001
State: Pending Approval
Phase: Iteration 2
Duration: 1 hour 23 minutes
Baseline: 4.8/10
Current: 5.8/10
Target: 7.0/10
```

---

### KDSE Report

Generates or retrieves the current Runtime Report.

**Purpose:** Access the Runtime Report for the current session.

**Inputs:**

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| session_id | Yes | String | Session to query |
| format | No | String | full, summary, delta |

**Outputs:**

| Output | Type | Description |
|--------|------|-------------|
| report_id | String | Report identifier |
| report | Object | Runtime Report content |

**Preconditions:**
- Session exists
- At least one assessment has completed

**Postconditions:**
- None (read-only)

**Example Usage:**
```
KDSE Report

Session ID: KDSE-RT-2026-07-10-001
Format: full
```

---

### KDSE Progress

Displays progress toward target maturity.

**Purpose:** View how much progress has been made toward the target.

**Inputs:**

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| session_id | Yes | String | Session to query |

**Outputs:**

| Output | Type | Description |
|--------|------|-------------|
| baseline | Number | Starting score |
| current | Number | Current score |
| target | Number | Target score |
| delta | Number | Improvement |
| remaining | Number | Points to target |
| progress_percent | Number | Percentage complete |

**Preconditions:**
- Session exists

**Postconditions:**
- None (read-only)

**Example Usage:**
```
KDSE Progress

Session ID: KDSE-RT-2026-07-10-001
```

**Example Output:**
```
Baseline:    4.8/10
Current:     5.8/10
Target:      7.0/10
───────────────────────
Improvement: +1.0 points
Remaining:   1.2 points
Progress:    45% complete
```

---

## Query Commands

### KDSE Scores

Displays compliance scores for the session.

**Purpose:** View detailed dimension scores.

**Inputs:**

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| session_id | Yes | String | Session to query |
| compare_to | No | String | Previous session ID for comparison |

**Outputs:**

| Output | Type | Description |
|--------|------|-------------|
| dimensions | Array | Score for each dimension |
| overall | Number | Overall compliance score |
| comparison | Object | Delta from previous session |

**Preconditions:**
- Session exists
- At least one assessment completed

**Postconditions:**
- None (read-only)

**Example Usage:**
```
KDSE Scores

Session ID: KDSE-RT-2026-07-10-001
Compare to: KDSE-RT-2026-07-09-001
```

---

### KDSE Findings

Displays audit findings for the session.

**Purpose:** View findings from compliance audits.

**Inputs:**

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| session_id | Yes | String | Session to query |
| filter | No | String | all, critical, high, medium, low |
| dimension | No | String | Filter by dimension |

**Outputs:**

| Output | Type | Description |
|--------|------|-------------|
| findings | Array | List of findings |
| count | Object | Counts by severity |
| categories | Object | Counts by dimension |

**Preconditions:**
- Session exists
- At least one assessment completed

**Postconditions:**
- None (read-only)

**Example Usage:**
```
KDSE Findings

Session ID: KDSE-RT-2026-07-10-001
Filter: high
Dimension: Knowledge
```

---

### KDSE History

Displays session execution history.

**Purpose:** View the sequence of events in a session.

**Inputs:**

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| session_id | Yes | String | Session to query |

**Outputs:**

| Output | Type | Description |
|--------|------|-------------|
| events | Array | Chronological event list |
| iterations | Array | Iteration summaries |
| decisions | Array | Operator decisions |

**Preconditions:**
- Session exists

**Postconditions:**
- None (read-only)

**Example Usage:**
```
KDSE History

Session ID: KDSE-RT-2026-07-10-001
```

---

## Decision Commands

### Approve

Approves the current recommendation.

**Purpose:** Authorize implementation of the recommended action.

**Inputs:**

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| recommendation_id | Yes | String | Recommendation to approve |
| reason | No | String | Rationale for approval |

**Outputs:**

| Output | Type | Description |
|--------|------|-------------|
| decision | String | APPROVED |
| timestamp | String | Decision timestamp |
| next_state | String | Runtime proceeds to Implementation |

**Preconditions:**
- Session is in Pending Approval state
- A recommendation is pending

**Postconditions:**
- Decision recorded
- Runtime proceeds to Implementation state
- Implementation can begin

**State Transition:**
```
Pending Approval → (Approve) → Implementing
```

**Example Usage:**
```
Approve

Recommendation: KDSE-ACT-001
Reason: Aligns with Q3 priorities
```

---

### Reject

Rejects the current recommendation.

**Purpose:** Decline the current recommendation and request an alternative.

**Inputs:**

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| recommendation_id | Yes | String | Recommendation to reject |
| reason | Yes | String | Rationale for rejection |

**Outputs:**

| Output | Type | Description |
|--------|------|-------------|
| decision | String | REJECTED |
| timestamp | String | Decision timestamp |
| next_state | String | Runtime generates alternative |

**Preconditions:**
- Session is in Pending Approval state
- A recommendation is pending

**Postconditions:**
- Decision recorded
- Runtime generates next highest-value recommendation
- Runtime returns to Pending Approval with new recommendation

**State Transition:**
```
Pending Approval → (Reject) → Reporting → Pending Approval
```

**Example Usage:**
```
Reject

Recommendation: KDSE-ACT-001
Reason: Priority shifted to security work this quarter
```

---

### Defer

Defers the current recommendation.

**Purpose:** Postpone the decision to a later session.

**Inputs:**

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| recommendation_id | Yes | String | Recommendation to defer |
| reason | Yes | String | Rationale for deferral |
| resume_after | No | String | Condition for resumption |

**Outputs:**

| Output | Type | Description |
|--------|------|-------------|
| decision | String | DEFERRED |
| timestamp | String | Decision timestamp |
| session_id | String | Session to resume later |

**Preconditions:**
- Session is in Pending Approval state
- A recommendation is pending

**Postconditions:**
- Decision recorded
- Session enters Paused state
- Session can be resumed with Continue KDSE

**State Transition:**
```
Pending Approval → (Defer) → Paused
```

**Example Usage:**
```
Defer

Recommendation: KDSE-ACT-001
Reason: Waiting for architecture review
Resume After: Architecture decision meeting (July 20)
```

---

## Command Reference Summary

| Command | Required | State Change | Human Only |
|---------|----------|--------------|------------|
| Run KDSE | Yes | Idle → Loading | No |
| Continue KDSE | Yes | Paused → (Resume) | No |
| Close KDSE | Yes | Any → Closed | No |
| Pause KDSE | No | Active → Paused | Yes |
| Resume KDSE | No | Paused → (Resume) | No |
| KDSE Status | Yes | None | No |
| KDSE Report | Yes | None | No |
| KDSE Progress | Yes | None | No |
| KDSE Scores | No | None | No |
| KDSE Findings | No | None | No |
| KDSE History | No | None | No |
| Approve | Yes | Pending → Implementing | Yes |
| Reject | Yes | Pending → Reporting | Yes |
| Defer | Yes | Pending → Paused | Yes |

---

## Command Interface Principles

### 1. Technology Neutral

Commands are abstract and can be implemented by:
- Humans following workflow
- CLI tools
- AI assistants
- CI/CD pipelines
- IDE extensions

### 2. State Machine Integration

All commands result in defined state transitions:
- Valid transitions are enforced
- Invalid commands return errors
- State is always consistent

### 3. Human Authorization

Certain commands require human-only execution:
- Approve
- Reject
- Defer
- Pause

This ensures humans remain in control of implementation decisions.

### 4. Audit Trail

All commands are recorded:
- Commands executed
- Timestamps
- Decisions made
- State transitions

---

## Versioning

Command interface is stable per [VERSIONING.md](VERSIONING.md):
- Required commands do not change in minor versions
- New optional commands may be added
- Breaking changes require major version increment

---

## Related Documents

| Document | Relationship |
|----------|-------------|
| [ARCHITECTURE.md](ARCHITECTURE.md) | Command interface context |
| [SESSION_PROTOCOL.md](SESSION_PROTOCOL.md) | State transitions |
| [EXECUTION_MODEL.md](EXECUTION_MODEL.md) | Command execution |
| [CONFORMANCE.md](CONFORMANCE.md) | Required commands |

---

## Document Relationships

```
COMMANDS.md (this document)
    │
    ├── Defines: Command interface
    │
    ├── Referenced by:
    │   ├── ARCHITECTURE.md
    │   ├── SESSION_PROTOCOL.md
    │   └── CONFORMANCE.md
    │
    └── Related to:
        ├── PROMPTS.md (command templates)
        └── WORKFLOW.md (command sequences)
```

---

*This document is an informative reference implementation. It defines the Runtime command interface, not KDSE requirements.*
