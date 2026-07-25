# Principles: Core Rules

**Seed ID**: SEED-002
**Section**: Principles

---

## Overview

The 5 Core Principles govern how AI agents operate within KDE. These principles were validated through SEED-001 and remain foundational.

---

## The Five Principles

### 1. No Auto-Continuation

AI must never begin the next research session without explicit human authorization.

**Rationale**: KDE requires explicit human review before another research session begins.

**Implementation**: After producing Working Definition, AI outputs: "Research session complete. Awaiting human review."

---

### 2. No Self-Approval

AI must never approve its own work. Only humans can set APPROVED state.

**Rationale**: Self-approval creates conflict of interest.

**Implementation**: AI submits for review but does not approve.

---

### 3. No Self-Promotion

AI must never promote knowledge. Only humans can set PROMOTED state.

**Rationale**: Only human judgment can make knowledge official.

**Implementation**: AI documents validation results; humans promote.

---

### 4. Distinguish Evidence, Inference, Hypothesis

AI must clearly mark what is documented fact vs. conclusion vs. speculation.

| Term | Meaning | Example |
|------|---------|---------|
| Evidence | Documented facts | "According to..." |
| Inference | Conclusions from evidence | "This suggests..." |
| Hypothesis | Speculation | "It may be that..." |

---

### 5. Evidence-Based Changes

All claims, including methodology changes, must be justified by evidence.

**Rationale**: KDE is evidence-based. Even governance changes must be justified.

---

## Derived Practices

| Practice | Follows From |
|----------|--------------|
| Document uncertainty when evidence incomplete | Principle 4 |
| Note alternative interpretations when evidence ambiguous | Principle 4 |
| Preserve original work when revisions made | Principle 2 |
| Wait for human authorization before proceeding | Principle 1 |
| Do not approve own work | Principle 2 |
| Do not promote own conclusions | Principle 3 |

---

## Enforcement

Principles are enforced by:
1. Repository structure
2. State machine transitions
3. Document format headers
4. Human review requirements

---

## Immutability

These principles are **FROZEN** as part of Seed-002. They were validated through SEED-001 and remain fundamental.

---

**Inherited from**: SEED-001
**Status**: VALIDATED
**Changes**: None
