# KDE-ENGINE-003 (Gamma) Knowledge Model

**Engine ID**: KDE-ENGINE-003
**Version**: 0.1.0
**Codename**: Gamma

---

## Overview

This document defines the knowledge object structure for Gamma causal knowledge discovery. Gamma extends Beta's knowledge model with causal-specific fields for mechanisms, confounders, and intervention predictions.

---

## Knowledge Object Structure

### Complete Causal Knowledge Object

```yaml
causal_knowledge:
  # Identity
  id: CAUSAL-XXX
  version: "1.0"
  
  # Core Statement
  statement: <string>
  
  # Causal Mechanism
  mechanism:
    description: <string>
    identified: [yes|partial|no]
    steps:
      - step: <integer>
        from: <string>
        to: <string>
        description: <string>
    confidence: <0-100>
  
  # Causal Direction
  causal_direction:
    cause: <string>
    effect: <string>
    temporal_precedence: [strong|moderate|weak|none]
    rationale: <string>
  
  # Effect Estimation
  effect:
    type: [ATE|ITE|LATE]
    value: <number>
    unit: <string>
    confidence_interval: [lower, upper]
    confidence: <0-100>
    magnitude: [large|medium|small|negligible]
  
  # Confounders
  confounders:
    - name: <string>
      description: <string>
      effect_on_cause: <string>
      effect_on_effect: <string>
      adjustment_method: [none|stratification|regression|instrumental_variable|matching]
      adjusted: [yes|partial|no]
      residual_confounding: <string>
  
  # Mechanism
  mediators:
    - name: <string>
      description: <string>
      type: [complete|partial]
      proportion_mediated: <0-1>
  
  # Moderators
  moderators:
    - name: <string>
      description: <string>
      effect_modification: [strengthens|weakens|reverses|no_effect]
      conditions: <string>
  
  # Intervention
  intervention:
    do_operation: <string>
    predicted_outcome:
      value: <number>
      confidence_interval: [lower, upper]
      confidence: <0-100>
    mechanism_interruption:
      - step: <integer>
        interrupted: [yes|no|partial]
        effect: <string>
    side_effects:
      - effect: <string>
        probability: <0-1>
        severity: [critical|major|minor]
  
  # Evidence
  evidence:
    supporting:
      - evidence_id: <string>
        type: [temporal|mechanism|confounder_adjustment|intervention]
        strength: [strong|moderate|weak]
    contradicting:
      - evidence_id: <string>
        type: <string>
        strength: [strong|moderate|weak]
    total_count: <integer>
  
  # Causal Statistics
  statistics:
    sample_size: <integer>
    p_value: <number>
    effect_size: <number>
    correlation: <number>
    confidence_interval: [lower, upper]
    chi_square: <number>
    degrees_of_freedom: <integer>
  
  # Confidence Assessment
  confidence:
    value: <0-100>
    level: [high|medium|low]
    basis: [statistical|causal_theory|mixed]
    factors:
      - factor: <string>
        contribution: [positive|negative|neutral]
        magnitude: [high|medium|low]
    limitations:
      - <limitation 1>
      - <limitation 2>
  
  # Applicability
  context:
    dimensions:
      - name: <string>
        values: [<value1>, <value2>]
        confidence: <0-100>
    conditions:
      - condition: <string>
        required: [true|false]
    applicability: [universal|conditional|limited]
  
  # Boundaries
  boundary:
    type: [none|contradiction|exception|edge_case|moderation|feedback]
    description: <string>
    conditions: <string>
    severity: [critical|major|minor]
    occurrence_rate: <0-1>
    evidence:
      - observation_id: <string>
        description: <string>
  
  # Assumptions
  assumptions:
    - assumption: <string>
      type: [temporal|confounding|mechanism|measurement]
      testable: [true|false]
      tested: [true|false]
      violation_consequence: <string>
  
  # Reproducibility
  reproducibility:
    verified: [true|false]
    verification_count: <integer>
    replications:
      - replication_id: <string>
        result: [supported|contradicted]
        sample_size: <integer>
        effect_size: <number>
  
  # Provenance
  provenance:
    engine_id: KDE-ENGINE-003
    engine_version: 0.1.0
    created: <ISO8601 timestamp>
    created_by: <string>
    modified: <ISO8601 timestamp>
    modified_by: <string>
    review_history:
      - reviewer: <string>
        date: <ISO8601>
        decision: <string>
        comments: <string>
  
  # Status
  status: [draft|validated|approved|published]
  review_date: <ISO8601 date>
  expiration_date: <ISO8601 date or null>
```

---

## Field Specifications

### Identity Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| id | String | Yes | Unique knowledge identifier |
| version | String | Yes | Schema version |

### Statement Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| statement | String | Yes | Precise causal claim |

### Causal Mechanism Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| mechanism.description | String | Yes | Mechanism description |
| mechanism.identified | Enum | Yes | Has mechanism been identified? |
| mechanism.steps | Array | No | Mechanism step breakdown |
| mechanism.confidence | Number | Yes | Confidence in mechanism |

### Causal Direction Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| causal_direction.cause | String | Yes | Causal variable |
| causal_direction.effect | String | Yes | Effect variable |
| causal_direction.temporal_precedence | Enum | Yes | Cause precedes effect? |
| causal_direction.rationale | String | No | Reasoning for direction |

### Effect Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| effect.type | Enum | Yes | Effect type (ATE/ITE/LATE) |
| effect.value | Number | Yes | Effect size value |
| effect.unit | String | No | Effect unit |
| effect.confidence_interval | Array | No | 95% CI |
| effect.confidence | Number | Yes | Effect confidence |
| effect.magnitude | Enum | No | Effect magnitude |

### Confounder Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| confounders[].name | String | Yes | Confounder name |
| confounders[].adjusted | Boolean | Yes | Has been adjusted? |
| confounders[].adjustment_method | Enum | No | Method used |
| confounders[].residual_confounding | String | No | Residual effect |

### Intervention Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| intervention.do_operation | String | Yes | do(X) specification |
| intervention.predicted_outcome | Object | Yes | Predicted result |
| intervention.side_effects | Array | No | Known side effects |

### Evidence Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| evidence.supporting | Array | Yes | Supporting evidence |
| evidence.contradicting | Array | No | Contradicting evidence |
| evidence.total_count | Number | Yes | Total evidence items |

### Statistics Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| statistics.sample_size | Number | Yes | Total observations |
| statistics.p_value | Number | No | Statistical significance |
| statistics.effect_size | Number | No | Cohen's d or similar |
| statistics.confidence_interval | Array | No | 95% CI |
| statistics.correlation | Number | No | Correlation coefficient |

### Confidence Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| confidence.value | Number | Yes | 0-100 scale |
| confidence.level | Enum | Yes | High/Medium/Low |
| confidence.basis | Enum | Yes | Statistical/Causal/Mixed |
| confidence.factors | Array | No | Contributing factors |

### Context Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| context.dimensions | Array | No | Applicable dimensions |
| context.conditions | Array | No | Required conditions |
| context.applicability | Enum | Yes | Universal/Conditional/Limited |

### Boundary Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| boundary.type | Enum | No | Boundary type |
| boundary.description | String | No | Boundary description |
| boundary.severity | Enum | No | Critical/Major/Minor |

### Provenance Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| provenance.engine_id | String | Yes | KDE-ENGINE-003 |
| provenance.engine_version | String | Yes | 0.1.0 |
| provenance.created | Timestamp | Yes | Creation time |

---

## Complete Example

```yaml
causal_knowledge:
  id: CAUSAL-001
  version: "1.0"
  
  statement: |
    Vaccination causes reduced disease incidence through 
    immune system activation.
  
  mechanism:
    description: |
      Vaccination introduces weakened pathogens, triggering 
      immune response that produces antibodies, providing 
      future protection against infection.
    identified: yes
    steps:
      - step: 1
        from: Vaccination
        to: Immune Detection
        description: Immune system recognizes foreign antigen
      - step: 2
        from: Immune Detection
        to: Antibody Production
        description: B-cells produce targeted antibodies
      - step: 3
        from: Antibody Production
        to: Disease Resistance
        description: Antibodies neutralize future pathogens
    confidence: 92
  
  causal_direction:
    cause: Vaccination
    effect: Reduced Disease Incidence
    temporal_precedence: strong
    rationale: Vaccination precedes disease exposure chronologically
  
  effect:
    type: ATE
    value: 0.85
    unit: proportion
    confidence_interval: [0.78, 0.91]
    confidence: 95
    magnitude: large
  
  confounders:
    - name: Healthcare Access
      description: Access to medical care
      effect_on_cause: Higher access increases vaccination likelihood
      effect_on_effect: Higher access reduces disease severity
      adjustment_method: stratification
      adjusted: yes
      residual_confounding: Minimal
  
  intervention:
    do_operation: do(vaccination = true)
    predicted_outcome:
      value: 0.85
      confidence_interval: [0.78, 0.91]
      confidence: 95
    mechanism_interruption:
      - step: 2
        interrupted: partial
        effect: Weakened immune response
    side_effects:
      - effect: Mild fever
        probability: 0.15
        severity: minor
  
  evidence:
    supporting:
      - evidence_id: EV-001
        type: intervention
        strength: strong
      - evidence_id: EV-002
        type: temporal
        strength: strong
      - evidence_id: EV-003
        type: mechanism
        strength: moderate
    contradicting: []
    total_count: 12
  
  statistics:
    sample_size: 50000
    p_value: 0.0001
    effect_size: 1.2
    correlation: 0.72
    confidence_interval: [0.68, 0.76]
  
  confidence:
    value: 94
    level: high
    basis: mixed
    factors:
      - factor: Large sample size
        contribution: positive
        magnitude: high
      - factor: Multiple study types
        contribution: positive
        magnitude: high
      - factor: Mechanism identified
        contribution: positive
        magnitude: medium
    limitations:
      - Confounder adjustment may be incomplete
  
  context:
    dimensions:
      - name: age_group
        values: [all ages]
        confidence: 85
      - name: vaccine_type
        values: [approved vaccines]
        confidence: 90
    conditions:
      - condition: Approved vaccine administered
        required: true
      - condition: Standard dosage followed
        required: true
    applicability: conditional
  
  boundary:
    type: none
    description: No significant boundaries identified
    severity: none
    occurrence_rate: 0
  
  assumptions:
    - assumption: No unmeasured confounding
      type: confounding
      testable: true
      tested: true
      violation_consequence: Effect estimate would be biased
    - assumption: Correct temporal ordering
      type: temporal
      testable: true
      tested: true
      violation_consequence: Causal direction would be reversed
  
  reproducibility:
    verified: true
    verification_count: 5
    replications:
      - replication_id: REP-001
        result: supported
        sample_size: 10000
        effect_size: 0.83
  
  provenance:
    engine_id: KDE-ENGINE-003
    engine_version: 0.1.0
    created: "2026-07-20T12:00:00Z"
    created_by: OpenHands Agent
  
  status: validated
  review_date: "2026-08-20"
```

---

## Validation Checklist

### Required Fields

- [ ] id: Unique identifier
- [ ] version: Schema version
- [ ] statement: Causal claim
- [ ] mechanism: Mechanism details
- [ ] causal_direction: Cause and effect
- [ ] effect: Effect size
- [ ] evidence: Supporting evidence
- [ ] statistics: Statistical support
- [ ] confidence: Confidence assessment
- [ ] context: Applicability
- [ ] provenance: Engine information
- [ ] status: Knowledge status

### Recommended Fields

- [ ] confounders: Confounder analysis
- [ ] mediators: Mechanism mediators
- [ ] moderators: Effect moderators
- [ ] intervention: Intervention predictions
- [ ] boundary: Known limitations
- [ ] assumptions: Explicit assumptions
- [ ] reproducibility: Replication information

### Validation Rules

1. **Statement Validity**
   - Must be a causal claim (uses causal language)
   - Must identify cause and effect
   - Must specify mechanism or lack thereof

2. **Evidence Validity**
   - At least one supporting evidence item
   - Evidence must be linked to evidence IDs
   - Conflicting evidence must be documented

3. **Confidence Validity**
   - Value must be 0-100
   - Level must match value (High ≥ 80, Medium ≥ 50, Low < 50)
   - Basis must be documented

4. **Provenance Validity**
   - Engine ID must be KDE-ENGINE-003
   - Version must be specified
   - Created timestamp required

---

## Related Documents

- [methodology.md](./methodology.md) — Full methodology
- [pipeline.md](./pipeline.md) — Pipeline processing

---

**Document Status**: EXPERIMENTAL
