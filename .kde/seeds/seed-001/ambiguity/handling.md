# Ambiguity Handling

**Seed ID**: SEED-001
**Source**: Migrated from /knowledge/003-what-is-ambiguity.md
**Status**: FOUNDATIONAL

---

## Overview

Ambiguity is uncertainty in meaning, applicability, or interpretation. This document defines how KDE handles ambiguity in knowledge claims.

---

## What is Ambiguity?

Ambiguity is **unclear or uncertain meaning** that requires resolution or explicit acknowledgment.

### Ambiguity Types

| Type | Description | Example |
|------|-------------|---------|
| **Semantic** | Word or phrase meaning unclear | "system" could mean multiple things |
| **Contextual** | Depends on circumstances | Applies "when X is present" |
| **Interpretive** | Multiple valid interpretations | Different conclusions from same evidence |
| **Boundary** | Unclear limits of applicability | "small" vs "large" unclear |

---

## Ambiguity vs Uncertainty

| Concept | Definition | KDE Treatment |
|---------|------------|---------------|
| **Ambiguity** | Unclear meaning | Requires resolution or acknowledgment |
| **Uncertainty** | Unknown outcome | Express as confidence |
| **Ignorance** | Unknown unknowns | Document as limitation |

---

## Handling Principles

### Principle 1: Acknowledge Ambiguity

**Never hide ambiguity. Always acknowledge it.**

```
When ambiguity exists:
1. Document the ambiguity explicitly
2. Note the impact on knowledge
3. State what is unclear
4. Explain why ambiguity exists
```

### Principle 2: Resolve When Possible

**Attempt to resolve ambiguity before publishing.**

```
Resolution steps:
1. Identify source of ambiguity
2. Gather additional evidence
3. Consult relevant sources
4. Make explicit choice
5. Document the choice rationale
```

### Principle 3: Preserve When Necessary

**When resolution is not possible, preserve ambiguity explicitly.**

```
When preserving:
1. Define scope of ambiguity
2. State conditions where applicable
3. Note alternative interpretations
4. Set boundary markers
```

---

## Ambiguity Documentation

### Required Fields

When ambiguity exists:

| Field | Description |
|-------|-------------|
| **Type** | Semantic, contextual, interpretive, boundary |
| **Source** | What creates the ambiguity |
| **Impact** | How it affects the knowledge |
| **Resolution** | How handled (resolved or preserved) |

### Example

```yaml
ambiguity:
  type: "Contextual"
  source: "Term 'system' has multiple valid definitions"
  impact: "Knowledge applicability unclear without context"
  resolution:
    status: "Preserved"
    note: "Context required to determine which definition applies"
    alternatives:
      - "Definition A: when context includes X"
      - "Definition B: when context includes Y"
```

---

## Ambiguity in Knowledge

### Knowledge Must Address Ambiguity

Every knowledge item must:

| Requirement | Description |
|-------------|-------------|
| **Identify** | Note any ambiguity present |
| **Assess** | Evaluate impact on validity |
| **Handle** | Resolve or preserve appropriately |
| **Document** | Record ambiguity treatment |

### Prohibited

| Prohibition | Reason |
|-------------|--------|
| Ignored ambiguity | Misleads users |
| Hidden ambiguity | Violates transparency |
| Unresolved without note | Leaves users misinformed |

---

## Context and Ambiguity

### Context Resolves Contextual Ambiguity

```
Without context:
"X applies to systems" → Ambiguous

With context:
"X applies to systems [where Y is true]" → Resolved
```

### Context Documentation

| Element | Description |
|---------|-------------|
| **Conditions** | When knowledge applies |
| **Boundaries** | When knowledge does not apply |
| **Context markers** | Signals for context application |

---

## Ambiguity Thresholds

### Action Thresholds

| Ambiguity Level | Action Required |
|-----------------|-----------------|
| **Minor** | Note and continue |
| **Moderate** | Resolve or document alternatives |
| **Severe** | Do not publish; requires resolution |

### Publication Requirements

| Threshold | Can Publish? |
|-----------|---------------|
| **Minor** | YES with note |
| **Moderate** | YES with documented alternatives |
| **Severe** | NO until resolved |

---

## Ambiguity in Evidence

### Evidence Ambiguity

Evidence may be ambiguous:
- Unclear source attribution
- Conflicting interpretations
- Incomplete documentation

### Handling Evidence Ambiguity

```
1. Note the ambiguity in evidence
2. Assess impact on confidence
3. Adjust confidence appropriately
4. Document the ambiguity
```

---

## Interpretive Ambiguity

### When Multiple Interpretations Exist

| Situation | Action |
|-----------|--------|
| Evidence supports multiple conclusions | Document all, note primary |
| Different valid frameworks apply | Document both, note context |
| Expert disagreement exists | Document positions, note consensus |

### Example

```yaml
interpretation:
  alternatives:
    - view: "Interpretation A"
      evidence: "Source 1, Source 2"
      conclusion: "X leads to Y"
    - view: "Interpretation B"
      evidence: "Source 3, Source 4"
      conclusion: "X does not lead to Y"
  resolution:
    primary: "Interpretation A"
    rationale: "More consistent with majority of evidence"
    note: "Interpretation B remains valid in alternative context"
```

---

## Preserving Ambiguity

### When to Preserve

Preserve ambiguity when:
- Resolution requires unavailable information
- Multiple interpretations are equally valid
- Context determines applicability

### How to Preserve

```
1. Define the ambiguity scope
2. State what is uncertain
3. Note the conditions
4. Set clear boundaries
5. Avoid false precision
```

---

## Immutability Note

This Ambiguity Handling Model is **FROZEN** as part of Seed-001. It shall never be modified.

If the fundamental ambiguity handling approach must change, a new Seed shall be created.

---

**Source**: /knowledge/003-what-is-ambiguity.md
**Seed**: SEED-001
**Status**: FOUNDATIONAL
**Modifiable**: NO
