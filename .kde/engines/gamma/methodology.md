# KDE-ENGINE-003 (Gamma) Methodology

**Engine ID**: KDE-ENGINE-003
**Version**: 0.1.0
**Codename**: Gamma
**Name**: Causal Knowledge Discovery Engine

---

## Overview

This document describes the Gamma methodology for causal knowledge discovery. Gamma extends contextual knowledge discovery to include causal inference, mechanism identification, and intervention prediction.

---

## Core Principles

Gamma implements the 5 Core Principles plus additional principles for causal knowledge:

### Base Principles (from Alpha/Beta)

1. **Evidence Over Intuition** — Decisions must be grounded in verifiable evidence
2. **Experiment Before Deployment** — Validate knowledge before operational use
3. **Preserve Ambiguity** — Do not prematurely resolve uncertainty
4. **Traceability Always** — Every conclusion traces to evidence
5. **Reproducibility Required** — All experiments must be reproducible

### Beta-Extended Principles

6. **Context is Required** — Knowledge without context is incomplete
7. **Boundaries Define Knowledge** — Understanding limits is as important as applicability
8. **Confidence Must Be Statistical** — Confidence is derived from statistical analysis
9. **Conditions Matter** — "When" is as important as "what"
10. **No Universal Claims** — Every knowledge claim must include applicability conditions

### Gamma-Specific Principles

11. **Causation Requires Mechanism** — Correlation is not causation; identify the causal pathway
12. **Confounders Must Be Addressed** — Third variables may explain observed relationships
13. **Intervention Changes Mechanisms** — Causal claims enable intervention prediction
14. **Causal Models Are Hypotheses** — Causal diagrams represent testable hypotheses
15. **Temporal Precedence Matters** — Cause must precede effect in time

---

## The Gamma Discovery Question

### Alpha Question
```
Does X correlate with Y?
```
**Answer:** Yes/No/Unknown

### Beta Question
```
Under what conditions does X correlate with Y?
```
**Answer:** Yes, when [context conditions] / No, except when [exceptions]

### Gamma Question
```
What is the causal mechanism by which X leads to Y?
How would intervening on X change Y?
```
**Answer:**
- Mechanism: [causal pathway]
- Effect Size: [magnitude]
- Confounders: [identified confounders]
- Intervention: [predicted outcome]
- Confidence: [statistical measure]

---

## Causal Discovery vs Correlation Discovery

### Correlation Discovery (Alpha/Beta)

**Focus:** Associations between variables

**Questions:**
- Is X associated with Y?
- Under what conditions is X associated with Y?

**Outputs:**
- Correlation coefficients
- Context conditions
- Confidence intervals

**Limitations:**
- Does not identify causal direction
- Does not account for confounding
- Cannot predict intervention effects

### Causal Discovery (Gamma)

**Focus:** Causal mechanisms and interventions

**Questions:**
- What causes what?
- How does X causally lead to Y?
- What would happen if we changed X?

**Outputs:**
- Causal diagrams
- Effect sizes (average treatment effect)
- Confounder adjustments
- Intervention predictions

**Capabilities:**
- Identifies causal direction
- Accounts for confounding
- Predicts intervention effects

---

## Causal Concepts

### 1. Causal Mechanism

A causal mechanism is the process by which a cause produces an effect.

```
Cause (X) → Mechanism (M) → Effect (Y)
```

**Example:**
```
Vaccination (X) → Immune Response (M) → Disease Prevention (Y)
```

### 2. Confounding Variable

A confounder is a variable that causes both X and Y, creating a spurious association.

```
Confounder (C)
    │
    ├──→ X
    │
    └──→ Y
```

**Example:**
```
Socioeconomic Status (C) → Education Level (X) → Income (Y)
```

### 3. Average Treatment Effect (ATE)

The average causal effect of X on Y across the population.

```
ATE = E[Y | do(X=1)] - E[Y | do(X=0)]
```

### 4. Intervention (do-operator)

An intervention fixes X at a specific value, cutting its incoming causal arrows.

```
do(X = x) → [Causal Model] → Y
```

### 5. Counterfactual

What would have happened to a specific unit if X had been different?

```
Y_i(1) - Y_i(0) = Individual Treatment Effect
```

---

## Causal Discovery Methods

### Method 1: Temporal Precedence

If X precedes Y in time, X may cause Y (but not necessarily).

**Rule:** Cause must occur before effect.

### Method 2: Intervention Analysis

If intervening to change X changes Y, X causes Y.

**Rule:** do(X) → Y implies causation.

### Method 3: Mechanism Identification

If we can identify the mechanism M linking X to Y, causation is supported.

**Rule:** X → M → Y provides causal pathway.

### Method 4: Confounder Adjustment

If the relationship persists after adjusting for confounders, causation is supported.

**Rule:** X → Y after controlling for C.

### Method 5: Counterfactual Reasoning

If Y would have been different had X been different, X causes Y.

**Rule:** Y_i(1) ≠ Y_i(0) implies causation.

---

## Pipeline Processing

Gamma extends Beta's 6-stage pipeline to 8 stages:

```
Stage 1: Evidence Ingestion
        │
        ▼
Stage 2: Observation Extraction
        │
        ▼
Stage 3: Pattern Detection
        │
        ▼
Stage 4: Statistical Validation
        │
        ▼
Stage 5: Context Detection
        │
        ▼
Stage 6: Boundary Detection
        │
        ▼
Stage 7: Causal Discovery ← NEW
        │
        ▼
Stage 8: Causal Knowledge Generation ← NEW
```

---

## Stage 7: Causal Discovery

### Input
- Validated patterns with statistics
- Context conditions
- Boundary information

### Output
- Causal hypotheses
- Causal diagrams
- Confounder assessments
- Effect size estimates

### Process

1. **Temporal Analysis**
   - Identify temporal ordering of variables
   - Apply temporal precedence rule

2. **Confounder Detection**
   - Search for common causes
   - Identify potential confounders
   - Assess confounder strength

3. **Mechanism Identification**
   - Search for mediating variables
   - Identify causal pathways
   - Document mechanism steps

4. **Causal Hypothesis Generation**
   - Generate causal hypotheses
   - Assess causal evidence strength
   - Assign causal confidence

### Causal Hypothesis Format

```yaml
causal_hypothesis:
  id: CH-XXX
  hypothesis: "[X causes Y]"
  mechanism: [mechanism description or null]
  confounders:
    - name: [confounder name]
      adjustment_method: [method]
      effect_on_x: [description]
      effect_on_y: [description]
  evidence:
    temporal_precedence: [strong/moderate/weak/none]
    mechanism_identified: [yes/partial/no]
    confounder_adjusted: [yes/partial/no]
    intervention_studied: [yes/no]
  causal_strength: [strong/moderate/weak]
  status: [hypothesis/validated/rejected]
```

---

## Stage 8: Causal Knowledge Generation

### Input
- Causal hypotheses
- Causal diagrams
- Effect size estimates
- Intervention predictions

### Output
- Complete causal knowledge object

### Process

1. **Effect Size Estimation**
   - Calculate average treatment effect
   - Estimate confidence intervals
   - Assess effect magnitude

2. **Intervention Prediction**
   - Predict outcomes of do(X) interventions
   - Assess intervention effects
   - Identify intervention side effects

3. **Causal Knowledge Assembly**
   - Assemble complete causal knowledge
   - Include all causal elements
   - Validate completeness

### Causal Knowledge Format

```yaml
causal_knowledge:
  id: CAUSAL-XXX
  version: "1.0"
  
  # Causal Statement
  statement: "[X causes Y via mechanism M]"
  
  # Causal Mechanism
  mechanism:
    description: [mechanism description]
    steps:
      - step: 1
        description: "[X occurs]"
      - step: 2
        description: "[M is activated]"
      - step: 3
        description: "[Y results]"
    confidence: [high/medium/low]
  
  # Effect Estimation
  effect:
    type: [ATE/ITE/LATE]
    value: [number]
    unit: [unit]
    confidence_interval: [lower, upper]
    confidence: [number]
  
  # Confounders
  confounders:
    - name: [name]
      adjusted: [true/false]
      method: [method]
      residual_effect: [description]
  
  # Intervention
  intervention:
    do_operation: "[do(X=x)]"
    predicted_outcome:
      value: [predicted value]
      confidence_interval: [lower, upper]
    side_effects:
      - effect: [description]
        probability: [number]
  
  # Evidence
  evidence:
    supporting:
      - evidence_id: [link]
        contribution: [type]
    contradicting:
      - evidence_id: [link]
        contribution: [type]
    total_count: [number]
  
  # Statistics
  statistics:
    sample_size: [number]
    p_value: [number]
    effect_size: [number]
    confidence_interval: [lower, upper]
  
  # Confidence
  confidence:
    value: [0-100]
    level: [high/medium/low]
    basis: [statistical/causal_theory/mixed]
    factors:
      - [factor 1]
      - [factor 2]
  
  # Applicability
  context:
    dimensions: [applicable dimensions]
    conditions: [specific conditions]
    applicability: [universal/conditional/limited]
  
  # Limits
  boundary:
    type: [boundary type]
    conditions: [failure conditions]
    severity: [critical/major/minor]
  
  # Metadata
  assumptions:
    - [assumption 1]
    - [assumption 2]
  
  reproducibility:
    verified: [true/false]
    replications: [number]
  
  provenance:
    engine_id: KDE-ENGINE-003
    engine_version: 0.1.0
    created: [timestamp]
  
  # Status
  status: [draft/validated/approved/published]
```

---

## Causal Assessment

Gamma produces knowledge with explicit causal assessments:

### Causal Confidence Levels

| Level | Description | Requirements |
|-------|-------------|--------------|
| **Established** | Causal relationship well-supported | Strong mechanism + confounder adjustment + intervention evidence |
| **Supported** | Causal relationship probable | Moderate mechanism + some adjustment |
| **Possible** | Causal relationship uncertain | Temporal precedence only |
| **Unlikely** | Causal relationship improbable | Confounding more likely |

### Causal Evidence Hierarchy

| Evidence Type | Strength |
|---------------|----------|
| Randomized Controlled Trial | Strongest |
| Natural Experiment | Strong |
| Instrumental Variable | Moderate |
| Longitudinal Observation | Moderate |
| Cross-sectional Observation | Weak |
| Temporal Precedence Only | Weakest |

---

## Validation Process

### Stage Gate Model

```
Evidence → Observation → Pattern → Validation → Context → Boundary → Causal Discovery → Causal Knowledge
    │          │           │         │           │          │              │               │
    ▼          ▼           ▼         ▼           ▼          ▼              ▼               ▼
  Gate 1     Gate 2      Gate 3    Gate 4      Gate 5     Gate 6        Gate 7         Gate 8
```

### Gate Requirements

| Gate | Requirement | Failure Action |
|------|-------------|----------------|
| Gate 1 | Evidence exists | Stop, require evidence |
| Gate 2 | Observations extractable | Adjust extraction |
| Gate 3 | Pattern candidate identified | Continue observation |
| Gate 4 | Statistical validation passes | Reject or modify pattern |
| Gate 5 | At least one context found | Flag as context-free |
| Gate 6 | Boundaries tested | Document as untested |
| Gate 7 | Causal evidence identified | Mark as correlation-only |
| Gate 8 | All causal fields complete | Return for completion |

---

## Scientific Learning Loop

Gamma extends the scientific learning loop for causal knowledge:

```
Research → Evidence → Observation → Causal Model → Prediction → Experiment → Knowledge
                            │                    │
                            ▼                    ▼
                     Causal Discovery        Intervention
                            │                    │
                            └────────────────────┘
                                      │
                                   Feedback
```

---

## Limitations

### Known Limitations

1. **Observational Data Limitations**
   - Cannot definitively prove causation from observation alone
   - Requires assumptions (no unmeasured confounding)
   - Confounder adjustment may be incomplete

2. **Complexity Limitations**
   - Cannot handle arbitrarily complex causal structures
   - Limited to specified causal relationships
   - Non-linear causality may not be captured

3. **Temporal Limitations**
   - Requires temporal data for temporal precedence
   - Feedback loops are difficult to model
   - Dynamic causal effects may change over time

### Assumptions Required

1. **Causal Sufficiency**
   - All common causes are measured
   - No important confounders omitted

2. **Temporal Validity**
   - Timestamps are accurate
   - Temporal ordering is correct

3. **No Selection Bias**
   - Sample is representative
   - No systematic missing data

---

## Related Documents

- [specification.md](./specification.md) — Engine identity and scope
- [pipeline.md](./pipeline.md) — Detailed pipeline documentation
- [knowledge-model.md](./knowledge-model.md) — Knowledge object specification
- [provenance.md](./provenance.md) — Engine history
- [changes.md](./changes.md) — Version history

---

**Document Status**: EXPERIMENTAL
