# Runtime Startup Sequence

**Document ID**: RUNTIME-STARTUP
**Version**: 1.1.0
**Date**: 2026-07-24
**Authority**: Human Authority
**Status**: PRODUCTION
**Source**: INV-AUTO-ENGINE-SELECTION (Human Approved)

---

## Overview

This document describes the Runtime startup sequence for KDE. The Runtime initializes deterministically using human-configured defaults and Automatic Engine Selection.

**Update v1.1.0**: Added Automatic Engine Selection (Step 4) based on INV-AUTO-ENGINE-SELECTION.

---

## Startup Sequence

```
┌─────────────────────────────────────────────────────────────┐
│                    RUNTIME STARTUP                          │
└─────────────────────────────────────────────────────────────┘

                    ┌─────────────────┐
                    │  1. Bootstrap   │
                    │  Read BOOTSTRAP │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │  2. Load Config │
                    │ Read Runtime    │
                    │ Defaults        │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │  3. Check       │
                    │ Session Override │
                    └────────┬────────┘
                             │
                    ┌────────┴────────┐
                    │ Override Present? │
                    └────────┬────────┘
                             │
              ┌──────────────┴──────────────┐
              │ YES                            │ NO
              ▼                               ▼
    ┌─────────────────┐            ┌─────────────────┐
    │  4a. Use Session│            │  4b. Automatic  │
    │  Override       │            │  Engine Selection│
    └────────┬────────┘            └────────┬────────┘
             │                               │
             └───────────────┬───────────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │  5. Load Engine │
                    │  (selected)     │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │  6. Load Seed  │
                    │  (from config) │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │  7. Verify      │
                    │  Compatibility  │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │  8. Initialize  │
                    │  Runtime State  │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │  9. Transfer    │
                    │  Authority      │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │  10. Engine    │
                    │  Control       │
                    └─────────────────┘
```

---

## Step Details

### Step 1: Bootstrap

**Action**: Read BOOTSTRAP.md entry point

**Purpose**: Establish canonical session entry

**Output**: Bootstrap acknowledged

---

### Step 2: Load Runtime Configuration

**Action**: Read Runtime default configuration

**Source**: `/governance/runtime/defaults.yaml`

**Configuration Loaded**:
```yaml
default_engine: KDE-ENGINE-002
default_seed: SEED-001
```

**Output**: Runtime defaults loaded into memory

---

### Step 3: Check Session Override

**Action**: Check for session-specific configuration

**Source**: Experiment or Investigation configuration

**Check**:
- Is `session_override.engine` specified?
- Is `session_override.seed` specified?

**Output**: Override status determined

---

### Step 4a: Apply Session Override (if present)

**Action**: Use session-specified Engine and/or Seed

**Rules**:
1. Override applies only to this session
2. Override does not modify Runtime defaults
3. Partial override allowed (can override just Engine or just Seed)

**Example**:
```yaml
session_override:
  engine: KDE-ENGINE-004  # Override engine
  # seed: not specified, use default
```

**Output**: Session configuration determined

**Logging**: Override is logged with reason for audit trail.

---

### Step 4b: Automatic Engine Selection (if no override)

**Action**: Select Engine based on problem characteristics

**Source**: `/governance/runtime/ENGINE-SELECTION.md`

**Algorithm**:
1. Parse problem statement (if provided in session)
2. Extract keywords
3. Score engines by keyword matches
4. Apply conflict resolution if ambiguous
5. Calculate confidence level
6. Log selection

**Keyword-to-Engine Mapping**:

| Keywords Detected | Engine Selected | Confidence |
|------------------|-----------------|-------------|
| why, cause, mechanism | Gamma (KDE-ENGINE-003) | HIGH |
| what if, prevent, intervene | Gamma (KDE-ENGINE-003) | HIGH |
| bootstrap, reproduce | Delta (KDE-ENGINE-004) | HIGH |
| context, validate, check | Beta (KDE-ENGINE-002) | HIGH |
| No specific keywords | Beta (KDE-ENGINE-002) | MEDIUM |

**Sequential Patterns**:
| Pattern | Sequence |
|---------|----------|
| Causal + Bootstrap | Gamma → Delta |
| Bootstrap + Analysis | Delta → Beta |
| Context + Causal | Beta → Gamma |

**Fallback Behavior**:
| Confidence | Action |
|------------|--------|
| ≥50% | Proceed with selection |
| 30-49% | Log warning, proceed |
| <30% | Use Beta default, flag |

**Output**: 
- Selected Engine
- Confidence level
- Selection rationale
- Alternatives considered

**Logging**: All selections logged per ENGINE-SELECTION.md specification.

---

### Step 4c: Session Configuration Summary

**Configuration**:
```yaml
# After Step 4a or 4b:
selected_engine: KDE-ENGINE-002  # Or KDE-ENGINE-003 or KDE-ENGINE-004
engine_selection_confidence: 85  # Percentage
selection_rationale: "Context keywords detected (validate, check)"
```

**Output**: Session configuration determined

---

### Step 5: Load Engine

**Action**: Load the selected Engine

**Source**: `/engines/{engine-id}/`

**Verification**:
- Engine files exist
- Engine version matches
- Engine is compatible with Seed

**Output**: Engine loaded into memory

---

### Step 6: Load Seed

**Action**: Load the selected Seed

**Source**: `/seeds/{seed-id}/`

**Verification**:
- Seed files exist
- Seed is compatible with Engine
- Seed status verified (FROZEN or active)

**Output**: Seed loaded into memory

---

### Step 7: Verify Compatibility

**Action**: Confirm Engine-Seed compatibility

**Check**: 
- Engine is in Seed's compatible engines list
- Versions are compatible
- All prerequisites met

**Output**: Compatibility verified

---

### Step 8: Initialize Runtime State

**Action**: Set Runtime state to INITIALIZING

**State Transition**:
```
UNINITIALIZED → INITIALIZING → READY
```

**Verification**:
- All engine modules loaded
- Engine configuration verified
- Prerequisites satisfied

**Output**: Runtime state: INITIALIZING

---

### Step 9: Transfer Authority

**Action**: Transfer execution authority to Engine

**Transfer**:
- From: AI Substrate
- To: Engine

**Post-Transfer**:
- AI operates as execution substrate
- Engine determines subsequent actions
- AI native planning suspended

**Output**: Authority transferred

---

### Step 10: Engine Control

**Action**: Runtime enters READY state

**State**: READY

**Control**: Engine now controls execution

**Await**: Engine directive for session task

---

## Session Termination

When a session completes:

```
Session End
     │
     ▼
Restore Runtime Defaults
     │
     ▼
Clear Session Override
     │
     ▼
Runtime Returns to Defaults
```

**Important**: Session-specific overrides are **not persisted**. Each new session starts with Runtime defaults.

---

## Error Handling

### Error: Configuration Not Found

**Condition**: `/governance/runtime/defaults.yaml` missing

**Action**: STOP and report missing artifact

**Recovery**: Human must restore configuration

---

### Error: Engine Not Found

**Condition**: Specified Engine does not exist

**Action**: STOP and report missing Engine

**Recovery**: Human must specify valid Engine

---

### Error: Seed Not Found

**Condition**: Specified Seed does not exist

**Action**: STOP and report missing Seed

**Recovery**: Human must specify valid Seed

---

### Error: Incompatible Engine-Seed

**Condition**: Engine and Seed are not compatible

**Action**: STOP and report incompatibility

**Recovery**: Human must specify compatible pair

---

## Key Principles

### Principle 1: Deterministic Startup

Runtime always starts with the same configuration unless explicitly overridden.

### Principle 2: Human Authority

Defaults are set by humans, not determined by KDE.

### Principle 3: Explicit Override

Sessions must explicitly specify overrides; absence of override means use defaults.

### Principle 4: Session Isolation

Overrides apply only to the current session; they do not modify Runtime configuration.

### Principle 5: Clean Termination

Sessions end with Runtime configuration restored to defaults.

---

## Related Documents

| Document | Purpose |
|----------|---------|
| [defaults.yaml](./defaults.yaml) | Runtime default configuration |
| [SESSION-OVERRIDE.md](./SESSION-OVERRIDE.md) | Session override behavior |
| [ENGINE-SELECTION.md](./ENGINE-SELECTION.md) | Automatic Engine Selection |
| [/laboratory/LABORATORY-RULES.md](/workspace/project/kde/laboratory/LABORATORY-RULES.md) | Laboratory Rules |

---

## Version History

| Version | Date | Changes | Authority |
|---------|------|---------|-----------|
| 1.0.0 | 2026-07-20 | Initial production release | SEED-001 |
| 1.1.0 | 2026-07-24 | Added Automatic Engine Selection (Step 4b) | INV-AUTO-ENGINE-SELECTION (Human Approved) |

---

**Status**: PRODUCTION
**Authority**: Human Authority
**Source**: INV-AUTO-ENGINE-SELECTION (Human Approved)
