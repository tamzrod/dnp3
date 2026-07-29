# Engine Selection Specification

**Document ID**: ENGINE-SELECTION
**Version**: 1.1.0
**Date**: 2026-07-24
**Authority**: Human Authority
**Status**: PRODUCTION
**Source**: INV-AUTO-ENGINE-SELECTION (Human Approved)
**Update**: LAB-BOOTSTRAP-ENGINE-AUDIT-001 REC-001 (Human Approved)

---

## Purpose

This document defines the **Automatic Engine Selection** mechanism for KDE Runtime. It specifies how the Runtime automatically selects the most appropriate Engine based on the problem statement.

**Source**: INV-AUTO-ENGINE-SELECTION investigation (LAB-047 evidence: 100% task classification accuracy)

---

## REC-001 Approval

**From**: LAB-BOOTSTRAP-ENGINE-AUDIT-001
**Recommendation**: Implement capability-based engine selection
**Approval**: APPROVED by Human Authority
**Date**: 2026-07-24

### Approved Actions

1. ✅ Enhance keyword matching with capability weighting
2. ✅ Implement multi-engine selection for complex tasks
3. ✅ Add parallel execution capability
4. ✅ Document selection rationale in all experiment headers

### Implementation Status

| Component | Status | Evidence |
|-----------|--------|----------|
| Capability keywords | ✅ IMPLEMENTED | Section below |
| Multi-engine selection | ✅ IMPLEMENTED | Section 5 |
| Parallel execution | 🔄 IN PROGRESS | Section 5.2 |
| Selection documentation | ✅ IMPLEMENTED | Section 7 |

---

## Overview

### Selection Philosophy

Engine selection is based on **problem characteristics** rather than manual assignment:

| Principle | Description | Evidence |
|-----------|-------------|----------|
| **Deterministic** | Same input → Same output | LAB-047 Phase 4 |
| **Evidence-Based** | Selection based on proven criteria | LAB-047 >95% accuracy |
| **Explainable** | Selection rationale always available | Section 7 |
| **Reversible** | Manual override always available | Section 6 |
| **Logged** | Every selection is recorded | Section 8 |

### Engine Selection Keywords

The selection mechanism uses keyword matching to determine the appropriate Engine:

| Engine | Primary Keywords | Confidence |
|--------|-----------------|-------------|
| **Gamma** | why, cause, mechanism, leads to, resulted from | HIGH (90%) |
| **Gamma** | what if, prevent, intervene, how does | HIGH (85%) |
| **Delta** | bootstrap, reproduce, consistent, deterministic | HIGH (90%) |
| **Delta** | initialize, start, session | HIGH (90%) |
| **Beta** | context, when, where, condition, situation | HIGH (85%) |
| **Beta** | boundary, limit, exception, fail | HIGH (85%) |
| **Beta** | validate, check, verify | HIGH (80%) |
| **Beta** | find, detect, identify, pattern | MEDIUM (75%) |

**Note**: Alpha is Historical and not available for automatic selection.

---

## Selection Algorithm

### Algorithm: select_engine

```
INPUTS:
  - problem_statement: string
  - objective: string  
  - reasoning_type: enum (optional)
  - evidence: list (optional)

OUTPUT:
  - selected_engine: Engine ID
  - confidence: percentage
  - justification: string
  - alternatives: list
  - sequence: list (if applicable)

ALGORITHM:

  // Step 1: Keyword extraction
  keywords = extract_keywords(lowercase(problem_statement + " " + objective))
  
  // Step 2: Keyword scoring
  scores = {
    "gamma": count_matches(keywords, GAMMA_KEYWORDS),
    "delta": count_matches(keywords, DELTA_KEYWORDS),
    "beta": count_matches(keywords, BETA_KEYWORDS),
    "alpha": 0  // Historical, not auto-selected
  }
  
  // Step 3: Reasoning type override
  IF reasoning_type is specified:
    RETURN map_reasoning_to_engine(reasoning_type)
  
  // Step 4: Confidence calculation
  max_score = max(scores.values())
  total_keywords = len(keywords)
  
  IF total_keywords > 0:
    confidence = min(100, (max_score / total_keywords) * 100 + base_confidence)
  ELSE:
    confidence = 50  // Default confidence
  
  // Step 5: Conflict resolution
  IF is_ambiguous(scores):
    RETURN resolve_conflict(scores)
  
  // Step 6: Selection
  selected = argmax(scores)
  
  // Step 7: Sequential pattern detection
  IF has_sequential_pattern(scores):
    RETURN (primary, secondary, sequence_recommendation)
  
  RETURN (selected, confidence, justification)
```

### Keyword Definitions

```
GAMMA_KEYWORDS = [
    "why", "cause", "caused", "causing",
    "mechanism", "leads to", "resulted from", "results in",
    "what if", "prevent", "prevention", "intervene", "intervention",
    "how does", "how did", "reason", "root cause"
]

DELTA_KEYWORDS = [
    "bootstrap", "reproduce", "reproducible", "reproduction",
    "consistent", "consistency", "deterministic", "determinism",
    "initialize", "initialization", "start", "starting",
    "session", "authority", "transfer"
]

BETA_KEYWORDS = [
    "context", "when", "where", "condition", "conditions", "situation",
    "boundary", "boundaries", "limit", "limits", "exception", "fail", "fails", "failure",
    "validate", "validation", "check", "checking", "verify", "verification",
    "find", "found", "detect", "detected", "identify", "identified", "pattern", "patterns",
    "correlation", "relationship", "association", "confidence", "significance"
]
```

---

## Decision Tree

```
START: Problem Statement
         │
         ▼
┌─────────────────────────────────────────────────────┐
│ Is "why/cause/mechanism" present?                   │
│ (why, cause, mechanism, leads to, resulted from)  │
└─────────────────────────────────────────────────────┘
         │
    ┌────┴────┐
    │           │
   YES          NO
    │           │
    ▼           ▼
┌───────────────────────┐    ┌─────────────────────────────────────┐
│ Is "bootstrap/reproduce"│    │ Is "bootstrap/reproduce" present?   │
│ present?               │    │ (bootstrap, reproduce, consistent)  │
└───────────────────────┘    └─────────────────────────────────────┘
    │                               │
┌───┴───┐                       ┌───┴───┐
│        │                       │        │
YES      NO                      YES      NO
 │        │                       │        │
 ▼        ▼                        ▼        ▼
Gamma →  │                  Delta    ┌─────────────────────────────┐
Delta    │                                │ Is "context/validate/check"│
         │                                │ present?                  │
Sequential                          └─────────────────────────────┘
(Causal →                              │
Reproducible)                               │
                                            ├─────────┴─────────┐
                                           YES                     NO
                                            │                       │
                                            ▼                       ▼
                                        Beta                    Beta
                                        (HIGH confidence)       (MEDIUM confidence)
                                                              (Default)

LEGEND:
──────
 →  : Selection path
 ┌──┐ : Decision point
```

---

## Tie-Breaking Rules

When multiple engines have equal scores:

| Scenario | Rule | Example |
|----------|------|---------|
| Equal Gamma + Delta | Gamma primary, Delta secondary | "why reproduce" → Gamma→Delta |
| Equal Beta + Gamma | More specific wins (causal is more specific) | "why context" → Gamma |
| Equal Beta + Delta | More specific wins | "reproduce context" → Delta |
| All equal | Default to Beta | → Beta |

**Specificity Order**: Gamma > Delta > Beta > Alpha

---

## Confidence Thresholds

| Confidence Level | Range | Behavior |
|-----------------|-------|----------|
| **HIGH** | 80-100% | Proceed with selection |
| **MEDIUM** | 50-79% | Proceed, log info |
| **LOW** | 30-49% | Proceed, log warning |
| **VERY LOW** | <30% | Use Beta default, require confirmation |

### Fallback Behavior

| Confidence | Action | Authority |
|------------|---------|-----------|
| ≥50% | Proceed with selection | Automatic |
| 30-49% | Log warning, proceed | Automatic |
| <30% | Use Beta default, flag for review | Automatic + Log |

---

## Sequential Execution

### Identified Sequential Patterns

| Pattern | Primary | Secondary | Value | Example |
|---------|---------|-----------|-------|---------|
| **Causal → Reproducible** | Gamma | Delta | Root cause then ensure consistency | "Why did X fail? Then ensure reproducible" |
| **Bootstrap → Analyze** | Delta | Beta | Initialize then standard analysis | "Bootstrap then analyze problem" |
| **Context → Causal** | Beta | Gamma | Context analysis then causal | "When does X fail? Why does X fail?" |
| **Analyze → Causal → Validate** | Beta | Gamma | Multi-stage improvement | "Check, then analyze, then verify" |

### Sequential Selection Algorithm

```
FUNCTION detect_sequential(scores):
    IF scores.gamma > 0 AND scores.delta > 0:
        RETURN (GAMMA, DELTA, "Causal analysis followed by reproducible output")
    IF scores.delta > 0 AND scores.beta > 0:
        RETURN (DELTA, BETA, "Bootstrap initialization followed by analysis")
    IF scores.beta > 0 AND scores.gamma > 0:
        RETURN (BETA, GAMMA, "Context analysis followed by causal reasoning")
    RETURN None
```

---

## Manual Override

### Override Authority

**Human Authority**: Any user may override automatic selection by specifying `engine_override` in session configuration.

### Override Syntax

```yaml
session_override:
  engine: KDE-ENGINE-003  # Override to Gamma
  reason: "Causal reasoning required for root cause analysis"
```

### Override Logging

Overrides are logged with:
- Original automatic selection
- Override reason
- User/system identifier
- Timestamp

---

## Selection Logging

### Log Entry Format

Every Engine selection creates a log entry:

```yaml
engine_selection_log:
  timestamp: 2026-07-24T12:00:00Z
  problem_statement: "Why did the session fail?"
  selected_engine: KDE-ENGINE-003
  engine_name: Gamma
  confidence: 90
  keywords_detected:
    - why
    - cause
  scores:
    gamma: 2
    delta: 0
    beta: 0
    alpha: 0
  justification: "Causal keywords detected (why, cause)"
  alternatives:
    - KDE-ENGINE-002 (Beta)
  sequential: null
  override: false
```

### Log Storage

Selection logs are stored in:
- `/laboratory/logs/engine-selection/YYYY-MM-DD.yaml`
- Rotated daily, retained for 90 days

---

## Integration with Runtime

### Startup Sequence Integration

The Engine Selection occurs during Runtime initialization:

```
RUNTIME STARTUP
     │
     ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 1: Bootstrap (Read BOOTSTRAP.md)                      │
└─────────────────────────────────────────────────────────────┘
     │
     ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 2: Load Configuration (defaults.yaml)                   │
└─────────────────────────────────────────────────────────────┘
     │
     ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 3: Check Session Override                               │
│          IF override present: Use override                   │
│          IF override absent: Continue to selection            │
└─────────────────────────────────────────────────────────────┘
     │
     ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 4: Automatic Engine Selection                           │
│          - Parse problem statement                          │
│          - Extract keywords                                 │
│          - Score engines                                    │
│          - Select engine                                    │
│          - Calculate confidence                             │
│          - Log selection                                    │
└─────────────────────────────────────────────────────────────┘
     │
     ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 5: Load Selected Engine                                 │
└─────────────────────────────────────────────────────────────┘
     │
     ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 6: Load Seed                                          │
└─────────────────────────────────────────────────────────────┘
     │
     ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 7: Initialize Runtime State → READY                    │
└─────────────────────────────────────────────────────────────┘
```

### Session Configuration Integration

```yaml
# Session with problem statement (triggers auto-selection)
session:
  problem_statement: "Why did the runtime initialization fail?"
  # engine: will be auto-selected based on keywords

# Session with explicit override
session:
  problem_statement: "Analyze the pattern"
  session_override:
    engine: KDE-ENGINE-002  # Beta
    reason: "Standard analysis requested"
```

---

## Engine Capability Reference

| Engine | ID | Primary Capability | Selection Trigger |
|--------|-----|-------------------|------------------|
| **Beta** | KDE-ENGINE-002 | Context Discovery | Default, validation keywords |
| **Gamma** | KDE-ENGINE-003 | Causal Discovery | why, cause, mechanism |
| **Delta** | KDE-ENGINE-004 | Bootstrap + Context | bootstrap, reproduce |
| **Alpha** | KDE-ENGINE-001 | Pattern Discovery | Historical (not auto-selected) |

---

## Multi-Engine Selection (REC-001 Enhancement)

### When to Use Multi-Engine

Multi-engine selection is appropriate when:

| Scenario | Condition | Engines |
|----------|-----------|---------|
| Causal + Reproducible | Task requires both | Gamma → Delta |
| Pattern + Causal | Pattern discovery + cause | Beta → Gamma |
| Bootstrap + Analysis | Initialize + investigate | Delta → Beta |
| Complex Investigation | Multiple dimensions | Beta + Gamma + Delta |

### Multi-Engine Selection Algorithm

```
FUNCTION detect_multi_engine(scores, problem):
    engine_count = count(scores > threshold)
    
    IF engine_count == 1:
        RETURN single_engine(scores)
    
    IF engine_count > 1:
        // Check for known patterns
        IF has_causal_keywords(problem) AND has_validate_keywords(problem):
            RETURN [Gamma, Beta]
        IF has_bootstrap_keywords(problem) AND has_analyze_keywords(problem):
            RETURN [Delta, Beta]
        IF has_pattern_keywords(problem) AND has_cause_keywords(problem):
            RETURN [Beta, Gamma]
        
        // Default: prioritize by specificity
        RETURN prioritize_by_specificity(scores)
    
    RETURN single_engine(scores)
```

### Multi-Engine Execution

| Execution Mode | Description | Use Case |
|---------------|-------------|----------|
| Sequential | Execute engines one after another | Causal → Reproducible |
| Parallel | Execute engines simultaneously | Independent analyses |
| Collaborative | Engines work together on same problem | Complex investigations |

---

## Parallel Execution (REC-001 Enhancement)

### When to Use Parallel Execution

Parallel execution is appropriate when:

| Condition | Example | Benefit |
|-----------|---------|---------|
| Independent sub-problems | Pattern analysis + Volume analysis | 2x speed |
| Redundant analysis | Verify with 2 engines | Higher confidence |
| Competitive analysis | Compare Beta vs Gamma findings | Better synthesis |

### Parallel Execution Configuration

```yaml
parallel_execution:
  enabled: true
  mode: collaborative  # or "competitive"
  engines:
    - KDE-ENGINE-002  # Beta
    - KDE-ENGINE-003  # Gamma
  synthesis: automatic  # or "manual"
```

### Parallel Results Synthesis

| Mode | Behavior |
|------|----------|
| collaborative | Engines share findings, synthesize together |
| competitive | Engines produce independent results, select best |
| automatic | Runtime decides based on task characteristics |

---

## Evidence Base

| Evidence | Source | Value |
|----------|--------|-------|
| Task classification accuracy | LAB-047 | 100% (15/15) |
| Keyword-to-engine mapping | LAB-047 | >95% reliability |
| Sequential patterns | LAB-044, LAB-047 | 3 patterns identified |
| Conflict resolution | LAB-047 Phase 5 | 4 rules defined |
| Beta as default | LAB-031 | Proven (9.1s, 100% correct) |
| Multi-engine necessity | LAB-BOOTSTRAP-ENGINE-AUDIT-001 | REC-001 approved |

---

## Version History

| Version | Date | Changes | Authority |
|---------|------|---------|-----------|
| 1.0.0 | 2026-07-24 | Initial specification | Human (INV-AUTO-ENGINE-SELECTION approved) |
| 1.1.0 | 2026-07-24 | Added multi-engine and parallel execution | Human (LAB-BOOTSTRAP-ENGINE-AUDIT-001 REC-001 approved) |

---

## Related Documents

| Document | Purpose |
|----------|---------|
| [RUNTIME-STARTUP.md](./RUNTIME-STARTUP.md) | Runtime startup sequence |
| [SESSION-OVERRIDE.md](./SESSION-OVERRIDE.md) | Session override behavior |
| [defaults.yaml](./defaults.yaml) | Runtime default configuration |
| [BOOTSTRAP.md](../../laboratory/BOOTSTRAP.md) | KDE entry point |
| [INV-AUTO-ENGINE-SELECTION](../../laboratory/investigations/INV-AUTO-ENGINE-SELECTION/) | Source investigation |

---

**Status**: PRODUCTION
**Authority**: Human Authority
**Source**: INV-AUTO-ENGINE-SELECTION (Human Approved)
**Review Date**: Upon evidence of need
