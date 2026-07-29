# Session Template with Engine Selection

**Document ID**: SESSION-TEMPLATE
**Date**: 2026-07-24
**Source**: LAB-BOOTSTRAP-ENGINE-AUDIT-001 REC-002
**Status**: APPROVED

---

## Purpose

This template provides a standardized session header that includes proper engine selection for KDE experiments.

---

## Standard Session Header

```yaml
# KDE Session Configuration
session:
  # Experiment metadata
  experiment_id: EXP-XXX
  title: Experiment Title
  created: 2026-07-24T00:00:00Z
  
  # Engine selection
  # Options: auto, KDE-ENGINE-002 (Beta), KDE-ENGINE-003 (Gamma), KDE-ENGINE-004 (Delta)
  
  # AUTO SELECTION (recommended for most experiments)
  # The runtime will automatically select based on keywords
  problem_statement: "Your experiment description here"
  # The following engines will be considered:
  # - Beta: context, validate, check, analyze, pattern, discover
  # - Gamma: why, cause, mechanism, root cause
  # - Delta: bootstrap, reproduce, validate, consistency
  
  # EXPLICIT OVERRIDE (use when specific engine required)
  # session_override:
  #   engine: KDE-ENGINE-003  # Use Gamma for causal analysis
  #   reason: "Causal analysis required"
```

---

## Engine Selection Examples

### Example 1: Pattern Discovery (Auto)

```yaml
session:
  experiment_id: EXP-PATTERN-001
  title: Pattern Discovery in BTC Data
  created: 2026-07-24T00:00:00Z
  problem_statement: "Discover recurring patterns in BTCUSDT 1-minute data"
  # Auto-selects Beta (pattern keywords detected)
```

### Example 2: Causal Analysis (Override)

```yaml
session:
  experiment_id: EXP-CAUSAL-001
  title: Why Do Mechanisms Degrade?
  created: 2026-07-24T00:00:00Z
  session_override:
    engine: KDE-ENGINE-003  # Gamma
    reason: "Causal analysis required to understand mechanism degradation"
  problem_statement: "Why do market mechanisms fail over time?"
```

### Example 3: Validation (Override)

```yaml
session:
  experiment_id: EXP-VALIDATE-001
  title: Validate Mechanisms on Holdout Data
  created: 2026-07-24T00:00:00Z
  session_override:
    engine: KDE-ENGINE-004  # Delta
    reason: "Bootstrap validation required for reproducibility"
  problem_statement: "Validate previously discovered mechanisms on unseen historical data"
```

### Example 4: Multi-Engine Collaboration

```yaml
session:
  experiment_id: EXP-COMPLEX-001
  title: Comprehensive Market Analysis
  created: 2026-07-24T00:00:00Z
  problem_statement: "Discover patterns, understand causes, and validate findings"
  parallel_execution:
    enabled: true
    mode: collaborative
    engines:
      - KDE-ENGINE-002  # Beta: Pattern discovery
      - KDE-ENGINE-003  # Gamma: Causal analysis
    synthesis: automatic
```

---

## Engine Selection Quick Reference

| Task | Recommended Engine | Keywords |
|------|-----------------|----------|
| Pattern discovery | Beta (auto) | pattern, discover, find |
| Statistical analysis | Beta (auto) | analyze, validate, check |
| Causal analysis | Gamma (override) | why, cause, mechanism |
| Root cause finding | Gamma (override) | root cause, resulted from |
| Bootstrap/Reproduce | Delta (override) | bootstrap, reproduce |
| Session validation | Delta (override) | validate, verify |
| Complex/multi-dimensional | Multi-engine | Multiple keyword types |

---

## Session Override Authority

**Human Authority**: Any user may override automatic selection.

**Override Syntax**:
```yaml
session_override:
  engine: <ENGINE-ID>
  reason: <HUMAN-READABLE-JUSTIFICATION>
```

**Required Fields**:
- `engine`: Must be valid KDE-ENGINE-ID
- `reason`: Must explain why override was necessary

---

## REC-002 Implementation Status

| Action | Status |
|--------|--------|
| Create template document | ✅ COMPLETE |
| Add to experiments template directory | 📋 PLANNED |
| Update laboratory template | 📋 PLANNED |

---

**Status**: APPROVED
**Authority**: Human Authority
**Source**: LAB-BOOTSTRAP-ENGINE-AUDIT-001 REC-002
