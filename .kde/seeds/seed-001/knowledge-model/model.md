# Knowledge Model

**Seed ID**: SEED-001
**Source**: Migrated from /knowledge/001-what-is-knowledge.md
**Status**: FOUNDATIONAL

---

## Overview

Knowledge is validated, evidence-backed understanding that has been approved for use within KDE. This model defines what constitutes knowledge.

---

## What is Knowledge?

Knowledge is **approved, evidence-backed understanding** that has passed through the KDE validation process.

### Knowledge Properties

| Property | Description |
|----------|-------------|
| **Approved** | Reviewed and accepted by Governance |
| **Evidence-Backed** | Supported by verifiable evidence |
| **Validated** | Tested through experiments |
| **Published** | Available in /knowledge/ |
| **Versioned** | Tracked through versions |

---

## Knowledge vs Other Concepts

| Concept | Definition | Knowledge Difference |
|---------|------------|---------------------|
| **Data** | Raw observations | Unprocessed |
| **Information** | Processed data | Organized but unvalidated |
| **Knowledge** | Validated understanding | Approved and backed |
| **Wisdom** | Applied knowledge | Contextual judgment |

---

## Knowledge States

### State Machine

```
┌─────────────────────────────────────────────────────────┐
│                  KNOWLEDGE STATE MACHINE                 │
└─────────────────────────────────────────────────────────┘

         ┌─────────────────────────────────────┐
         │           DRAFT                      │
         │                                      │
         │  Initial knowledge candidate         │
         │  Created by Research                │
         └──────────────┬──────────────────────┘
                        │ submits for review
                        ▼
         ┌─────────────────────────────────────┐
         │           REVIEW                     │
         │                                      │
         │  Under Governance review            │
         │  Evidence being validated           │
         └──────────────┬──────────────────────┘
                        │ approved
                        ▼
         ┌─────────────────────────────────────┐
         │           APPROVED                   │
         │                                      │
         │  Accepted by Governance            │
         │  Ready for experimentation          │
         └──────────────┬──────────────────────┘
                        │ validated by experiment
                        ▼
         ┌─────────────────────────────────────┐
         │          VALIDATED                  │
         │                                      │
         │  Tested through experiments        │
         │  Evidence accumulated               │
         └──────────────┬──────────────────────┘
                        │ promoted by human
                        ▼
         ┌─────────────────────────────────────┐
         │          PROMOTED                   │
         │                                      │
         │  Published to /knowledge/          │
         │  Official KDE knowledge             │
         │  Available for use                 │
         └─────────────────────────────────────┘
```

### State Definitions

| State | Description | Transitions |
|-------|-------------|-------------|
| **DRAFT** | Initial candidate | → REVIEW |
| **REVIEW** | Under evaluation | → APPROVED, → DRAFT |
| **APPROVED** | Accepted | → VALIDATED, → REVIEW |
| **VALIDATED** | Tested | → PROMOTED, → REVIEW |
| **PROMOTED** | Official | (terminal) |

---

## Knowledge Structure

### Required Fields

| Field | Description | Required |
|-------|-------------|----------|
| **ID** | Unique identifier | YES |
| **Statement** | Core knowledge claim | YES |
| **Evidence** | Supporting evidence | YES |
| **Confidence** | Confidence level | YES |
| **Context** | Applicability conditions | YES |
| **Boundary** | Limitations and exceptions | YES |
| **Version** | Current version number | YES |
| **Status** | Current state | YES |

### Optional Fields

| Field | Description | When Used |
|-------|-------------|-----------|
| **Examples** | Concrete illustrations | Helpful for understanding |
| **Alternatives** | Alternative interpretations | When relevant |
| **Related** | Linked knowledge items | When applicable |

---

## Knowledge Requirements

### For PROMOTED Status

| Requirement | Description |
|-------------|-------------|
| **Evidence** | At least one supporting evidence item |
| **Validation** | At least one successful experiment |
| **Review** | Governance approval documented |
| **Promotion** | Human authorization for promotion |
| **Completeness** | All required fields filled |

### Prohibited

| Prohibition | Reason |
|-------------|--------|
| Self-approval | Only humans can approve |
| Self-promotion | Only humans can promote |
| Unvalidated promotion | Requires experimental evidence |
| Unreviewed promotion | Requires governance review |

---

## Knowledge Confidence

### Confidence Levels

| Level | Description | Evidence Requirement |
|-------|-------------|---------------------|
| **HIGH** | Strong support | Multiple consistent sources |
| **MEDIUM** | Moderate support | Some evidence, minor gaps |
| **LOW** | Weak support | Preliminary evidence only |
| **UNKNOWN** | Insufficient data | No evidence yet |

### Confidence Assignment

Based on:
- Evidence quantity
- Evidence quality
- Validation results
- Reproducibility
- Expert consensus

---

## Knowledge Context

### Context Definition

Knowledge is **conditional**. Every knowledge item includes:

| Element | Description |
|---------|-------------|
| **Conditions** | When knowledge applies |
| **Boundaries** | When knowledge does not apply |
| **Limitations** | Known gaps or uncertainties |

### Context Example

```yaml
knowledge:
  statement: "X correlates with Y"
  context:
    conditions:
      - "When Z is present"
      - "In domain D"
    boundaries:
      - "Does not apply when Z is absent"
      - "Limited to domain D"
```

---

## Knowledge in the Loop

### Role in Scientific Loop

```
Research ──creates──► Knowledge ──tests──► Laboratory
                              ▲
                              │
                         approves
                              │
                         Governance
```

Knowledge is:
- Created by Research
- Approved by Governance
- Tested by Laboratory
- Promoted to official status

---

## Immutability Note

This Knowledge Model is **FROZEN** as part of Seed-001. It shall never be modified.

If the fundamental knowledge standards must change, a new Seed shall be created.

---

**Source**: /knowledge/001-what-is-knowledge.md
**Seed**: SEED-001
**Status**: FOUNDATIONAL
**Modifiable**: NO
