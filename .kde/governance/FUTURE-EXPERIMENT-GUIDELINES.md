# Future Experiment Guidelines

**Document ID**: FUTURE-EXPERIMENT-GUIDELINES
**Date**: 2026-07-24
**Source**: LAB-BOOTSTRAP-ENGINE-AUDIT-001 REC-004
**Status**: APPROVED

---

## Purpose

These guidelines ensure future KDE experiments properly utilize the multi-engine ecosystem and avoid the engine monopolization issue identified in LAB-BOOTSTRAP-ENGINE-AUDIT-001.

---

## Core Principle

**Use the right engine for the task.**

Engine selection should be based on:
1. Task requirements
2. Engine capabilities
3. Evidence-based matching

NOT on:
1. Default bias
2. Historical habit
3. Convenience

---

## Engine Selection Guidelines

### When to Use Beta (KDE-ENGINE-002)

| Task Type | Example | Why Beta |
|-----------|---------|---------|
| Pattern Discovery | "Find patterns in price data" | Primary pattern capability |
| Statistical Analysis | "Calculate return distribution" | Statistical analysis |
| Context Detection | "When does X occur?" | Context analysis |
| Boundary Detection | "What are the limits?" | Boundary detection |
| Validation | "Check if X is valid" | Validation keywords |

### When to Use Gamma (KDE-ENGINE-003)

| Task Type | Example | Why Gamma |
|-----------|---------|----------|
| Causal Analysis | "Why does X happen?" | Causal discovery |
| Root Cause Analysis | "What caused X?" | Root cause capability |
| Mechanism Understanding | "How does X cause Y?" | Mechanism analysis |
| Intervention Prediction | "What if we change X?" | Intervention prediction |

### When to Use Delta (KDE-ENGINE-004)

| Task Type | Example | Why Delta |
|-----------|---------|----------|
| Bootstrap | "Initialize new experiment" | Bootstrap capability |
| Reproducibility | "Reproduce previous finding" | Reproducibility |
| Consistency Checking | "Verify consistency" | Consistency checking |
| Session Validation | "Validate session integrity" | Validation |

### When to Use Multi-Engine

| Scenario | Engines | Why |
|----------|--------|-----|
| Pattern + Causal | Beta → Gamma | Discover patterns, then explain causes |
| Causal + Validate | Gamma → Delta | Find causes, then verify |
| Bootstrap + Analyze | Delta → Beta | Initialize, then analyze |
| Complex Investigation | Beta + Gamma + Delta | Multi-dimensional analysis |

---

## Experiment Design Checklist

Before starting an experiment, verify:

### Task Analysis

- [ ] I have clearly defined the task objective
- [ ] I have identified the key question(s) to answer
- [ ] I have determined if task requires pattern discovery, causal analysis, or validation

### Engine Selection

- [ ] I have selected the appropriate engine based on task requirements
- [ ] I have documented why this engine was selected
- [ ] I have considered if multi-engine collaboration would improve results
- [ ] If using override, I have documented the reason

### Experiment Design

- [ ] Experiment is designed for the selected engine's capabilities
- [ ] Methodology is appropriate for the engine
- [ ] Expected outcomes align with engine's strengths

---

## Common Mistakes to Avoid

### Mistake 1: Default to Beta Always

**Problem**: Beta was used exclusively in LAB-BOOTSTRAP-ENGINE-AUDIT-001, causing:
- Causal analysis not performed
- Bootstrap validation not used
- Underutilization of available capabilities

**Solution**: Evaluate task requirements and select appropriate engine.

### Mistake 2: Ignoring Gamma/Delta Capabilities

**Problem**: Gamma and Delta were registered but never invoked.

**Solution**: When task involves "why", "how does", or "cause", explicitly select Gamma.
When task involves reproducibility or validation, explicitly select Delta.

### Mistake 3: Not Documenting Selection

**Problem**: Engine selection rationale not documented, making review difficult.

**Solution**: Always document:
- Selected engine
- Keywords detected
- Scores for all engines
- Reason for selection or override

---

## Session Override Examples

### Example 1: Causal Investigation

```yaml
session_override:
  engine: KDE-ENGINE-003  # Gamma
  reason: "Task explicitly asks 'why' - causal analysis required"
```

### Example 2: Validation Requirement

```yaml
session_override:
  engine: KDE-ENGINE-004  # Delta
  reason: "Findings must be reproducible on holdout data"
```

### Example 3: Multi-Engine Collaboration

```yaml
parallel_execution:
  enabled: true
  mode: collaborative
  engines:
    - KDE-ENGINE-002  # Beta: Pattern discovery
    - KDE-ENGINE-003  # Gamma: Causal analysis
  synthesis: automatic
  reason: "Task requires both pattern discovery and causal understanding"
```

---

## Success Criteria for Future Experiments

Future experiments are considered successful if:

1. **Engine Selection**
   - [ ] Appropriate engine(s) selected based on task
   - [ ] Selection rationale documented
   - [ ] No unnecessary default bias

2. **Multi-Engine Utilization**
   - [ ] Gamma used for causal questions
   - [ ] Delta used for validation tasks
   - [ ] Multi-engine considered for complex tasks

3. **Documentation**
   - [ ] Engine selection documented in header
   - [ ] Selection log included
   - [ ] Override reason documented (if applicable)

---

## Resource Requirements by Engine

| Engine | Compute | Memory | Time |
|--------|---------|--------|------|
| Beta | Medium | Medium | Standard |
| Gamma | High | High | Extended |
| Delta | Medium | Medium | Standard |
| Multi-Engine | High | High | Extended |

---

## Approval and Implementation

**REC-004 Status**: APPROVED
**Authority**: Human Authority
**Source**: LAB-BOOTSTRAP-ENGINE-AUDIT-001

**Implementation**:
- These guidelines are effective immediately
- All new experiments should follow these guidelines
- Existing experiments should be retroactively annotated where possible

---

**Status**: APPROVED
**Date**: 2026-07-24
