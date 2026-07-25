# KDE-ENGINE-003 (Gamma) README

**Engine ID**: KDE-ENGINE-003
**Version**: 0.1.0
**Codename**: Gamma
**Name**: Causal Knowledge Discovery Engine
**Status**: Experimental

---

## Quick Reference

| Aspect | Value |
|--------|-------|
| **Engine ID** | KDE-ENGINE-003 |
| **Discovery Type** | Causal Knowledge Discovery |
| **Question** | "How does X causally lead to Y?" |
| **Confidence** | Statistical + Causal (X% ± Y%) |
| **Status** | Experimental |

---

## Overview

Gamma is the third KDE reasoning engine, extending contextual knowledge discovery (Beta) with causal inference capabilities. Gamma identifies causal mechanisms, accounts for confounding, and predicts intervention effects.

---

## Engine Comparison

| Engine | Discovery | Question | Confidence |
|--------|-----------|----------|------------|
| **Alpha** | Pattern | "Does X correlate with Y?" | Qualitative |
| **Beta** | Context | "When does X correlate with Y?" | Statistical |
| **Gamma** | Causal | "How does X causally lead to Y?" | Causal |

---

## Key Features

### What's New in Gamma

1. **Causal Discovery** — Identifies causal relationships, not just correlations
2. **Mechanism Identification** — Documents causal pathways and steps
3. **Confounder Analysis** — Detects and adjusts for third variables
4. **Effect Estimation** — Quantifies average treatment effects
5. **Intervention Prediction** — Predicts outcomes of do(X) operations

### Pipeline

Gamma implements an 8-stage pipeline:

```
1. Evidence Ingestion
2. Observation Extraction
3. Pattern Detection
4. Statistical Validation
5. Context Detection
6. Boundary Detection
7. Causal Discovery ← NEW
8. Causal Knowledge Generation ← NEW
```

---

## Discovery Question

### Alpha
```
Does X correlate with Y?
```

### Beta
```
Under what conditions does X correlate with Y?
```

### Gamma
```
What is the causal mechanism by which X leads to Y?
How would intervening on X change Y?
```

---

## Knowledge Output

### Beta Knowledge (KNOW-XXX)

```yaml
knowledge:
  statement: "X correlates with Y under conditions C"
  confidence: 92%
  context: [conditions]
  boundary: [limitations]
```

### Gamma Causal Knowledge (CAUSAL-XXX)

```yaml
causal_knowledge:
  statement: "X causes Y via mechanism M"
  confidence: 94%
  mechanism:
    steps:
      - X activates M
      - M produces Y
  effect:
    type: ATE
    value: 0.85
  confounders:
    - C identified and adjusted
  intervention:
    do(X): Y predicted to increase by 0.85
```

---

## When to Use Gamma

### Use Gamma When:

1. **Understanding Causes**
   - Need to know WHY X affects Y
   - Mechanism matters for decisions

2. **Intervention Planning**
   - Want to predict outcome of changing X
   - Need to estimate effect sizes

3. **Confounder Awareness**
   - Third variables may explain relationships
   - Need to adjust for confounders

### Don't Use Gamma When:

1. **Correlation Discovery**
   - Only need to know if X is associated with Y
   - Use Beta instead

2. **Quick Pattern Detection**
   - Initial exploration without causal claims
   - Use Alpha instead

3. **Causal Assumptions Cannot Be Met**
   - Cannot justify causal assumptions
   - Temporal precedence unclear
   - Use Beta or Alpha instead

---

## Status and Validation

### Current Status: EXPERIMENTAL

Gamma requires laboratory validation before production use.

### Validation Checklist

| Requirement | Status |
|-------------|--------|
| Methodology documented | ✅ Complete |
| Pipeline defined | ✅ Complete |
| Knowledge model specified | ✅ Complete |
| Laboratory experiments | ⏳ Pending |
| Empirical validation | ⏳ Pending |
| Production readiness | ⏳ Pending |

---

## Documentation Index

| Document | Purpose |
|----------|---------|
| [specification.md](./specification.md) | Engine identity and scope |
| [methodology.md](./methodology.md) | Full causal discovery methodology |
| [pipeline.md](./pipeline.md) | 8-stage processing pipeline |
| [knowledge-model.md](./knowledge-model.md) | Causal knowledge object structure |
| [changes.md](./changes.md) | Changes from Beta |
| [provenance.md](./provenance.md) | Engine lineage and history |

---

## Quick Start

### Using Gamma in Experiments

1. **Set Engine in Experiment**
   ```yaml
   Engine:
     ID: KDE-ENGINE-003
     Version: 0.1.0
     Codename: Gamma
   ```

2. **Provide Causal Evidence**
   - Temporal precedence evidence
   - Mechanism evidence
   - Confounder information
   - Intervention studies (if available)

3. **Interpret Causal Output**
   - Review causal hypotheses
   - Assess confounder adjustments
   - Examine intervention predictions

### Expected Output

- CAUSAL-XXX knowledge objects
- Causal diagrams
- Effect size estimates
- Intervention predictions
- Assumption documentation

---

## Limitations

### Known Limitations

1. **Observational Data**
   - Cannot definitively prove causation
   - Requires assumptions

2. **Complexity**
   - Limited to specified structures
   - May miss non-linear causality

3. **Temporal Requirements**
   - Needs temporal data
   - Cannot handle feedback loops well

### Required Assumptions

1. No unmeasured confounding
2. Correct temporal ordering
3. No selection bias

---

## Engine Lineage

```
Alpha (001) → Beta (002) → Gamma (003) → Delta (004)
 Pattern         Context         Causal        Temporal
 Discovery     Discovery      Discovery     (planned)
```

---

## Contact and Support

For questions about Gamma:

1. Review [methodology.md](./methodology.md) for causal discovery approach
2. Review [pipeline.md](./pipeline.md) for processing details
3. Review [knowledge-model.md](./knowledge-model.md) for output structure

---

## Related Documentation

- [Engine Framework](../README.md) — Overall engine architecture
- [Engine Interface](../interface.md) — Standard interface specification
- [Alpha Specification](../alpha/specification.md) — Alpha engine
- [Beta Specification](../beta/specification.md) — Beta engine
- [Future Engines](../future-engines.md) — Engine roadmap

---

**Status**: EXPERIMENTAL
**Review Date**: Upon completion of validation experiments
