# Document State Machine

This document defines the states research documents pass through and the transitions between them.

## States

| State | Description |
|-------|-------------|
| **DRAFT** | Initial state; work in progress |
| **REVIEW** | Submitted for human review |
| **APPROVED** | Human has approved the work |
| **REVISION_REQUIRED** | Human has requested changes |
| **VALIDATED** | Definition has passed validation tests |
| **PROMOTED** | Definition has been moved to /knowledge/ |
| **REJECTED** | Work has been rejected |

## State Diagram

```
                    ┌─────────────┐
                    │   DRAFT     │
                    └──────┬──────┘
                           │
                    Submit for Review
                           │
                           ▼
                    ┌─────────────┐
                    │   REVIEW    │◄────┐
                    └──────┬──────┘     │
                           │            │Revision
                    Human  │            │Requested
                    Review │            │
                    ▼      ▼            │
              ┌──────────┐  ┌─────────────┴───┐
              │ APPROVED │  │ REVISION_REQUIRED│
              └────┬─────┘  └─────────────────┘
                   │
                   ▼
            ┌─────────────┐
            │ VALIDATED  │
            └──────┬──────┘
                   │
                   ▼
            ┌─────────────┐
            │  PROMOTED   │
            └─────────────┘
```

## State Transitions

| From | To | Trigger |
|------|-----|---------|
| DRAFT | REVIEW | Submit for review |
| REVIEW | APPROVED | Human approves |
| REVIEW | REVISION_REQUIRED | Human requests changes |
| REVISION_REQUIRED | REVIEW | AI resubmits |
| APPROVED | VALIDATED | Validation tests pass |
| VALIDATED | PROMOTED | Human approves promotion |

## Rejection Path

```
REVIEW → REJECTED (if work is fundamentally flawed)
```

Rejected documents should not be continued without significant revision.

## Key Principles

1. **No Self-Approval** — AI cannot move REVIEW to APPROVED
2. **No Self-Promotion** — AI cannot move VALIDATED to PROMOTED
3. **No Auto-Continuation** — AI cannot move from one state to next without authorization

## Implementation

Document state is tracked in the document header:

```markdown
**State**: DRAFT
```

State transitions are recorded in git commit messages.

## Validation Before Promotion

Before a document can be PROMOTED, it must:

1. Have state = VALIDATED
2. Have passed all validation tests in the Validation Plan section
3. Have human approval confirming validation

## Exception Handling

If a document cannot proceed through normal states:

1. Document the reason in the Knowledge Promotion section
2. State what is needed to proceed
3. Wait for human authorization

No document should be in an undefined or inconsistent state.
