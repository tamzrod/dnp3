# Migration Notes: Browser-Style Default Management

**Document ID**: MIGRATION-DEFAULT-MANAGEMENT
**Version**: 1.0.0
**Date**: 2026-07-20
**Authority**: Human Authority
**Status**: PRODUCTION

---

## Overview

This document describes the migration from implicit default selection to explicit browser-style default management.

---

## Problem Statement

### Before: Implicit Default Selection

Previously, KDE had no explicit distinction between:
- Scientific validation level (Engine Status)
- Operational configuration (which Engine is actually used)

This led to confusion:
- "Active" Engine was assumed to be the default
- Scientific validation seemed to imply operational selection
- No clear separation between "scientific state" and "runtime behavior"

### Example of the Problem

```
KDE-ENGINE-003 (Gamma) becomes "Experimental"
    ↓
Someone assumes Gamma will be used by default
    ↓
But Beta remains the actual Runtime default
    ↓
Confusion about which Engine is "active"
```

---

## Solution: Browser-Style Default Management

### Core Principle

Scientific lifecycle (Engine Status) and operational configuration (Runtime Defaults) are **completely separate**.

| Concept | Governed By | Changes When |
|---------|-------------|--------------|
| Engine Status | Scientific validation | Engine undergoes validation |
| Runtime Default | Human configuration | Human edits defaults.yaml |

### Browser Metaphor

| Browser Behavior | KDE Runtime Behavior |
|-----------------|---------------------|
| User sets home page | Human sets default Engine |
| User can change anytime | Human can change anytime |
| Extensions can override temporarily | Sessions can override temporarily |
| User always has final authority | Human always has final authority |

---

## Key Changes

### 1. New Configuration File

**Created**: `/governance/runtime/defaults.yaml`

This file is the **single source of truth** for Runtime defaults:

```yaml
default_engine: KDE-ENGINE-002
default_seed: SEED-001
```

### 2. Explicit Default Column in Registry

**Changed**: `/engines/current.md`

Engine registry now shows both Status and Default:

| Engine ID | Status | Default |
|-----------|--------|---------|
| KDE-ENGINE-002 | Active | **YES** |
| KDE-ENGINE-003 | Experimental | NO |
| KDE-ENGINE-004 | Candidate (Validated) | NO |

### 3. Session Override Mechanism

**Created**: `/governance/runtime/SESSION-OVERRIDE.md`

Sessions can override defaults without changing them:

```yaml
session_override:
  engine: KDE-ENGINE-004  # Use Delta instead of default
```

### 4. Governance Directive

**Added**: Human Authority Over Runtime Defaults

Laboratory Rules now explicitly state:
- Only humans can change defaults
- KDE shall never change defaults automatically
- Scientific validation does not imply default selection

---

## What Changed (and What Didn't)

### What Changed

| Before | After |
|--------|-------|
| No explicit default configuration | `defaults.yaml` is single source of truth |
| "Active" implied default | Status and Default are independent |
| No session override | Session override mechanism exists |
| Implicit default selection | Explicit Runtime configuration |
| No governance directive | Human authority directive added |

### What Did NOT Change

| Aspect | Status Quo |
|--------|-----------|
| Current default Engine | Still KDE-ENGINE-002 (Beta) |
| Current default Seed | Still SEED-001 |
| Runtime behavior | Still loads configured Engine and Seed |
| Engine implementations | Unchanged |
| Scientific lifecycle | Unchanged |

---

## Migration Path

### For Existing Sessions

1. Session starts
2. Runtime reads `defaults.yaml`
3. Session uses configured defaults
4. **No visible change** to existing workflows

### For New Experiments

1. Create experiment configuration
2. Optionally add `session_override` if needed
3. Runtime respects override (if present)
4. **No required changes** to experiment format

### For Validation Experiments

1. Validation specifies `session_override`
2. Runtime loads specified Engine
3. Validation runs with specified Engine
4. Session ends, defaults unchanged

---

## Common Questions

### Q: Does Delta becoming "Validated" mean it becomes the default?

**A**: No. Status and Default are independent. Delta remains "Candidate (Validated)" with Default="NO" until a human explicitly changes `defaults.yaml`.

### Q: How do I use Delta for an experiment?

**A**: Add a session override to your experiment configuration:

```yaml
session_override:
  engine: KDE-ENGINE-004
```

### Q: Who can change the default Engine?

**A**: Only humans can modify `/governance/runtime/defaults.yaml`. KDE shall never change this automatically.

### Q: What happens if an experiment doesn't specify an override?

**A**: Runtime uses the defaults from `defaults.yaml` (currently Beta + SEED-001).

### Q: Does this affect how experiments are validated?

**A**: No. Validations should specify `session_override` to use the Engine under test, but the validation results are the same.

---

## Breaking Changes

**None**. This migration is purely additive and maintains backward compatibility.

- Existing experiments continue to work
- Default Engine and Seed remain the same
- Runtime behavior is unchanged

---

## Rollback Plan

If issues arise:

1. **Restore previous behavior**: Remove `defaults.yaml` (Runtime will need alternative config source)
2. **Revert registry changes**: Remove "Default" column from Engine registry
3. **Remove governance directive**: Remove "Human Authority" section from Laboratory Rules

**Not recommended**: Changes are additive and low-risk.

---

## Related Documents

| Document | Purpose |
|----------|---------|
| [defaults.yaml](./defaults.yaml) | Runtime default configuration |
| [RUNTIME-STARTUP.md](./RUNTIME-STARTUP.md) | Runtime startup sequence |
| [SESSION-OVERRIDE.md](./SESSION-OVERRIDE.md) | Session override behavior |
| [/laboratory/LABORATORY-RULES.md](/workspace/project/kde/laboratory/LABORATORY-RULES.md) | Laboratory Rules (updated) |
| [/engines/current.md](/workspace/project/kde/engines/current.md) | Engine registry (updated) |

---

**Migration Date**: 2026-07-20
**Status**: Complete
**Authority**: Human Authority
