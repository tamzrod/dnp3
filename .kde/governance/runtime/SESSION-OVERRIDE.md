# Session Override Behavior

**Document ID**: SESSION-OVERRIDE
**Version**: 1.0.0
**Date**: 2026-07-20
**Authority**: Human Authority
**Status**: PRODUCTION

---

## Overview

This document describes how Experiments and Investigations can override Runtime defaults for a specific session.

---

## Purpose

Session override allows:
- Testing an Engine before it becomes default
- Comparing multiple Engines in the same context
- Validating candidate Engines
- Running experiments with specific configurations

---

## Override Syntax

### In Experiment Configuration

```yaml
experiment:
  id: EXP-XXX
  title: Experiment Title
  
session_override:
  engine: KDE-ENGINE-004
  seed: SEED-001
```

### In Investigation Configuration

```yaml
investigation:
  id: INV-XXX
  title: Investigation Title
  
session_override:
  engine: KDE-ENGINE-004
  seed: SEED-001
```

---

## Override Options

### Full Override

Both Engine and Seed specified:

```yaml
session_override:
  engine: KDE-ENGINE-004
  seed: SEED-001
```

**Effect**: Use specified Engine and Seed for entire session

---

### Partial Override: Engine Only

```yaml
session_override:
  engine: KDE-ENGINE-004
```

**Effect**: Use specified Engine, use default Seed (SEED-001)

---

### Partial Override: Seed Only

```yaml
session_override:
  seed: SEED-002
```

**Effect**: Use default Engine (KDE-ENGINE-002), use specified Seed

---

## Override Rules

### Rule 1: Explicit

Override must be explicitly specified in session configuration.

```yaml
# Override present → use override
session_override:
  engine: KDE-ENGINE-004

# Override absent → use Runtime defaults
```

**Default behavior**: If `session_override` is not present, Runtime uses defaults.

---

### Rule 2: Session-Scoped

Override applies only to the current session.

**During session**:
- Override Engine is active
- Override Seed is active

**After session**:
- Override is discarded
- Runtime returns to defaults
- Next session starts with defaults

---

### Rule 3: Non-Persistent

Override does **not** modify Runtime configuration.

```yaml
# This experiment uses Delta
session_override:
  engine: KDE-ENGINE-004

# But Runtime defaults.yaml still says:
# default_engine: KDE-ENGINE-002
```

**After experiment**:
- Delta is no longer active
- Runtime defaults unchanged
- Beta is again the default

---

### Rule 4: Human Authorized

Override must be authorized by human configuration.

KDE shall not:
- Override defaults automatically
- Infer a better Engine
- Promote itself to default

Only explicit human configuration enables override.

---

## Override Scope

### What Override Affects

| Scope | Affected by Override |
|-------|---------------------|
| Current session Engine | Yes |
| Current session Seed | Yes |
| Runtime default Engine | No |
| Runtime default Seed | No |
| Other sessions | No |
| Engine status | No |

### What Override Does NOT Affect

| Scope | Affected by Override |
|-------|---------------------|
| Runtime defaults.yaml | No |
| Engine registry | No |
| Engine status | No |
| Seed registry | No |
| Seed status | No |
| Other sessions | No |

---

## Override Lifecycle

```
Session Start
     │
     ▼
Check Override Present?
     │
    ┌─┴────────────────┐
    │ YES               │ NO
    ▼                   ▼
┌─────────┐    ┌─────────────────┐
│ Override │    │ Use Runtime     │
│ Config   │    │ Defaults        │
└────┬────┘    └────────┬────────┘
     │                   │
     └─────────┬─────────┘
               ▼
       ┌───────────────┐
       │ Session Active │
       │ (Override or   │
       │  Defaults)     │
       └───────┬───────┘
               ▼
        Session Ends
               │
               ▼
     ┌─────────────────┐
     │ Discard Override│
     │ Restore Defaults│
     └─────────────────┘
```

---

## Examples

### Example 1: Validating Candidate Engine

**Context**: VAL-001 wants to test Delta against Beta

**Configuration**:
```yaml
validation:
  id: VAL-001
  title: Delta Validation
  
session_override:
  engine: KDE-ENGINE-004
```

**Behavior**:
1. Session starts
2. Override applies: Delta is loaded
3. Validation runs with Delta
4. Session ends
5. Override discarded
6. Runtime returns to Beta

---

### Example 2: No Override

**Context**: Regular experiment without override

**Configuration**:
```yaml
experiment:
  id: EXP-001
  title: Regular Experiment
```

**Behavior**:
1. Session starts
2. No override present
3. Runtime defaults apply: Beta + SEED-001
4. Experiment runs with defaults
5. Session ends
6. Runtime unchanged

---

### Example 3: Partial Override

**Context**: Test new Seed with current Engine

**Configuration**:
```yaml
experiment:
  id: EXP-002
  title: Seed Test
  
session_override:
  seed: SEED-002
```

**Behavior**:
1. Session starts
2. Partial override: only Seed specified
3. Engine uses default (Beta)
4. Seed uses override (SEED-002)
5. Session runs with Beta + SEED-002
6. Session ends
7. Runtime returns to Beta + SEED-001

---

## Override in Validation Context

### Validation Run 1: Baseline

```yaml
validation:
  id: VAL-001
  
session_override:
  engine: KDE-ENGINE-002  # Baseline (Beta)
```

### Validation Run 2: Candidate

```yaml
validation:
  id: VAL-001
  
session_override:
  engine: KDE-ENGINE-004  # Candidate (Delta)
```

**Note**: Both runs explicitly specify Engine. Neither run changes Runtime defaults.

---

## Override in Investigation Context

### Investigation Using Candidate Engine

```yaml
investigation:
  id: INV-XXX
  title: Investigation with Delta
  
session_override:
  engine: KDE-ENGINE-004
  seed: SEED-001
```

### Investigation Using Default Engine

```yaml
investigation:
  id: INV-YYY
  title: Investigation with Beta
  
# No override - uses defaults
```

---

## Error Cases

### Error 1: Invalid Engine

```yaml
session_override:
  engine: KDE-ENGINE-INVALID
```

**Result**: STOP, report invalid Engine

---

### Error 2: Invalid Seed

```yaml
session_override:
  seed: SEED-INVALID
```

**Result**: STOP, report invalid Seed

---

### Error 3: Incompatible Pair

```yaml
session_override:
  engine: KDE-ENGINE-003
  seed: SEED-002
```

**Result**: STOP if Engine and Seed are incompatible

---

## Best Practices

### Do

- Explicitly specify Engine and Seed in validation configurations
- Document why override is needed
- Return to defaults after testing
- Keep overrides session-scoped

### Don't

- Persist override beyond session
- Modify Runtime defaults for experiments
- Use override to bypass governance
- Infer "better" Engines without human direction

---

## Related Documents

| Document | Purpose |
|----------|---------|
| [defaults.yaml](./defaults.yaml) | Runtime default configuration |
| [RUNTIME-STARTUP.md](./RUNTIME-STARTUP.md) | Runtime startup sequence |
| [/laboratory/LABORATORY-RULES.md](/workspace/project/kde/laboratory/LABORATORY-RULES.md) | Laboratory Rules |

---

**Status**: PRODUCTION
**Authority**: Human Authority
