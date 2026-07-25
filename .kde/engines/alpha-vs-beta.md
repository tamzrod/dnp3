# Alpha vs Beta: Engine Comparison

**Document Version**: 1.0
**Date**: 2026-07-20

---

## Overview

This document provides a comprehensive comparison between KDE-ENGINE-001 (Alpha) and KDE-ENGINE-002 (Beta), the two official KDE Engines.

---

## Quick Comparison

| Aspect | Alpha | Beta |
|--------|-------|------|
| **Engine ID** | KDE-ENGINE-001 | KDE-ENGINE-002 |
| **Codename** | Alpha | Beta |
| **Status** | Historical | Active |
| **Discovery Type** | Pattern | Contextual Knowledge |
| **Question** | "Does X?" | "When does X?" |
| **Confidence** | Implicit | Statistical |
| **Context** | None | Automatic |
| **Boundaries** | None | Explicit |
| **Pipeline Modules** | 4 | 6 |
| **Use Case** | Legacy/Comparison | Production |

---

## Discovery Philosophy

### Alpha: Pattern Discovery

**Question:**
```
Does X correlate with Y?
```

**Answer Format:**
```
Pattern: X → Y
Confidence: Implicit
Context: None
Boundary: None
```

**Example:**
```
Pattern: Center Control → Winning
Confidence: Observed in 70% of games
Context: Not documented
Boundary: Not documented
```

### Beta: Contextual Knowledge Discovery

**Question:**
```
Under what conditions does X correlate with Y?
```

**Answer Format:**
```
Knowledge: Center Control → Winning (in classical openings)
Confidence: 97% (statistical)
Context: Classical openings, rating > 1400, time control ≥ 60min
Boundary: Not valid for hypermodern, flank, or endgame
```

---

## Pipeline Comparison

### Alpha Pipeline (4 Modules)

```
┌─────────────┐
│  Evidence   │
└──────┬──────┘
       │
       ▼
┌─────────────────┐
│   Observation   │
└──────┬─────────┘
       │
       ▼
┌─────────────────┐
│    Pattern      │
│    Detector     │
└──────┬─────────┘
       │
       ▼
┌─────────────────┐
│   Validation    │
└──────┬─────────┘
       │
       ▼
    Pattern
```

### Beta Pipeline (6 Modules)

```
┌─────────────┐
│  Evidence   │
└──────┬──────┘
       │
       ▼
┌─────────────────┐
│  Observation    │
│    Engine       │
└──────┬─────────┘
       │
       ▼
┌─────────────────┐
│    Pattern      │
│    Detector     │
└──────┬─────────┘
       │
       ▼
┌─────────────────┐
│   Statistical   │
│    Validator     │
└──────┬─────────┘
       │
       ▼
┌─────────────────┐
│    Context      │  ← NEW
│    Detector     │
└──────┬─────────┘
       │
       ▼
┌─────────────────┐
│    Boundary     │  ← NEW
│    Detector     │
└──────┬─────────┘
       │
       ▼
┌─────────────────┐
│    Knowledge    │
│    Generator    │
└──────┬─────────┘
       │
       ▼
   Knowledge
```

---

## Feature Comparison

### Core Features

| Feature | Alpha | Beta |
|---------|-------|------|
| Evidence Collection | ✅ | ✅ |
| Observation Extraction | ✅ | ✅ |
| Pattern Detection | ✅ | ✅ |
| Basic Validation | ✅ | ✅ |
| Statistical Validation | ❌ | ✅ |
| Context Detection | ❌ | ✅ |
| Boundary Detection | ❌ | ✅ |
| Knowledge Generation | ❌ | ✅ |

### Statistical Capabilities

| Statistic | Alpha | Beta |
|-----------|-------|------|
| Sample Size | Count only | Full analysis |
| P-value | ❌ | ✅ |
| Correlation | Basic | Full |
| Confidence Interval | ❌ | ✅ |
| Chi-square | ❌ | ✅ |
| Effect Size | ❌ | ✅ |

### Confidence Assessment

| Aspect | Alpha | Beta |
|--------|-------|------|
| **Basis** | Implicit/Subjective | Statistical |
| **Calculation** | Observation-based | Mathematically derived |
| **Confidence Interval** | ❌ | ✅ |
| **Factors Documented** | ❌ | ✅ |

### Context Capabilities

| Aspect | Alpha | Beta |
|--------|-------|------|
| Context Tracking | ❌ | ✅ |
| Dimension Search | ❌ | ✅ |
| Applicability Classification | ❌ | ✅ |
| Context Strength Quantification | ❌ | ✅ |

### Boundary Capabilities

| Aspect | Alpha | Beta |
|--------|-------|------|
| Boundary Tracking | ❌ | ✅ |
| Contradiction Detection | ❌ | ✅ |
| Exception Identification | ❌ | ✅ |
| Edge Case Discovery | ❌ | ✅ |
| Reverse Correlation Detection | ❌ | ✅ |

---

## Knowledge Object Comparison

### Alpha: Pattern Object

```yaml
pattern:
  id: PAT-001
  relationship: "Center Control → Winning"
  evidence:
    - EV-001
    - EV-002
  observation_count: 70
  confidence: high
  notes: "Observed correlation in games"
```

### Beta: Knowledge Object

```yaml
knowledge:
  id: KNOW-001
  version: "1.0"
  
  statement: |
    In classical chess openings, center control correlates
    with winning outcomes for players rated 1400+.
  
  evidence:
    supporting:
      - evidence_id: EV-001
        description: Match analysis
      - evidence_id: EV-002
        description: Statistical study
    total_count: 52
  
  statistics:
    sample_size: 70
    p_value: 0.003
    correlation: 0.72
    confidence_interval: [0.58, 0.82]
    effect_size: 0.89
  
  confidence:
    value: 97
    level: high
    basis: statistical
    factors:
      - Large sample size (n=70)
      - High correlation (r=0.72)
      - Very low p-value (p<0.01)
  
  context:
    dimensions:
      - name: opening_type
        values: [classical]
        confidence: high
      - name: player_rating
        values: [1400+]
        confidence: medium
    applicability: conditional
  
  boundary:
    type: exception
    description: Not valid for hypermodern or flank openings
    severity: minor
```

---

## When to Use Each Engine

### Use Alpha (KDE-ENGINE-001) When:

1. **Legacy Compatibility Required**
   - Existing experiments use Alpha
   - Need to compare with historical results
   - Compliance with historical methodology

2. **Simple Pattern Discovery Sufficient**
   - Basic correlation detection needed
   - Context not important
   - Quick initial exploration

3. **Historical Comparison Needed**
   - Validating against Alpha results
   - Demonstrating Beta improvements
   - Academic comparisons

### Use Beta (KDE-ENGINE-002) When:

1. **Context Detection Required**
   - Need to know when patterns apply
   - Applicability conditions matter
   - Domain understanding needed

2. **Boundary Definition Important**
   - Understanding pattern limits critical
   - Exception handling required
   - Safety implications exist

3. **Statistical Validation Required**
   - Formal significance needed
   - Confidence intervals required
   - Scientific publication planned

4. **Complete Knowledge Objects Needed**
   - Knowledge management system integration
   - Full provenance tracking
   - Comprehensive documentation

---

## Experiment Selection Matrix

| Requirement | Alpha | Beta |
|-------------|-------|------|
| Pattern discovery only | ✅ | ✅ |
| Context-aware results | ❌ | ✅ |
| Statistical significance | ❌ | ✅ |
| Explicit boundaries | ❌ | ✅ |
| Historical comparison | ✅ | ✅ |
| Simple/fast analysis | ✅ | ❌ |
| Complete knowledge model | ❌ | ✅ |
| Production use | ❌ | ✅ |

---

## Migration Path

### From Alpha to Beta

Existing experiments under Alpha **do not need migration**.

However, if you wish to re-run an Alpha experiment under Beta:

1. **Document the Original Results**
   - Preserve Alpha findings
   - Note any differences

2. **Re-run Under Beta**
   - Execute experiment with Beta engine
   - Compare outputs

3. **Analyze Differences**
   - Note new context discovered
   - Note boundaries identified
   - Assess statistical validation

4. **Update Knowledge**
   - Create Beta knowledge objects
   - Preserve Alpha for comparison
   - Document migration

---

## Performance Comparison

| Metric | Alpha | Beta |
|--------|-------|------|
| Processing Time | Faster | More thorough |
| Output Completeness | Basic | Comprehensive |
| Statistical Rigor | Minimal | Full |
| Knowledge Quality | Pattern | Contextual Knowledge |

---

## Future Engines

Both Alpha and Beta establish the foundation for future engines:

| Engine | Expected | Features |
|--------|----------|----------|
| Gamma | TBD | Causal inference |
| Delta | TBD | Temporal tracking |

See: [Future Engine Extension Guide](./future-engines.md)

---

## Summary

| Engine | Best For | Key Innovation |
|--------|----------|---------------|
| **Alpha** | Legacy, comparison | First pattern discovery |
| **Beta** | Production, science | Context & boundaries |

**Recommendation:** Use Beta for all new experiments. Reserve Alpha for historical comparison.

---

**Document Status**: APPROVED
