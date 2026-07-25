# When to Create a New Seed

**Seed ID**: SEED-001
**Version**: 1.0.0

---

## Overview

A new Seed should be created when KDE's fundamental reasoning DNA changes in a way that cannot be represented as an evolution of an existing Seed.

---

## Creation Criteria

### Create New Seed When

| Criterion | Description | Example |
|-----------|-------------|---------|
| **Reasoning Changes** | Core thinking process fundamentally different | New evidence model |
| **Principle Changes** | Core principles modified | New AI behavior rule |
| **Model Changes** | Foundational models replaced | New confidence approach |
| **Loop Changes** | Scientific loop restructured | New subsystem added |
| **Breaking Change** | Old reasoning incompatible | Complete methodology overhaul |

### Do NOT Create New Seed When

| Scenario | Action | Reason |
|----------|--------|--------|
| Engine improvement | Update Engine | Methodology evolves |
| New experiment | Run in Laboratory | Execution context |
| New knowledge | Add to Knowledge | Discovered, not foundational |
| Process improvement | Update Laboratory | Execution layer |
| Tool update | Update tooling | Implementation detail |

---

## Decision Tree

```
┌─────────────────────────────────────────────────────────┐
│              SEED CREATION DECISION TREE                  │
└─────────────────────────────────────────────────────────┘

              Is this a fundamental change to
              how KDE thinks about knowledge?

                        │
           ┌────────────┴────────────┐
           │                         │
          YES                        NO
           │                         │
           ▼                         ▼
    ┌─────────────┐          ┌─────────────────┐
    │ CREATE NEW  │          │ Is this a change│
    │    SEED     │          │ to methodology? │
    └─────────────┘          └────────┬────────┘
                                      │
                            ┌─────────┴─────────┐
                            │                   │
                           YES                  NO
                            │                   │
                            ▼                   ▼
                     ┌───────────┐      ┌────────────┐
                     │ UPDATE    │      │ Is this a  │
                     │  ENGINE   │      │ process    │
                     └───────────┘      │ change?    │
                                        └─────┬──────┘
                                              │
                                    ┌─────────┴─────────┐
                                    │                   │
                                   YES                  NO
                                    │                   │
                                    ▼                   ▼
                             ┌───────────┐      ┌────────────┐
                             │ UPDATE    │      │ NO ACTION  │
                             │LABORATORY │      │  NEEDED    │
                             └───────────┘      └────────────┘
```

---

## Seed Creation Process

### Step 1: Identify Reasoning Change

Document what fundamental reasoning is changing:
- Which principles are affected?
- Which models are modified?
- Which concepts are added/removed?

### Step 2: Assess Scope

Determine if change is:
- **Seed-level**: Fundamental reasoning change
- **Engine-level**: Methodology improvement
- **Laboratory-level**: Execution improvement

### Step 3: Create New Seed

If Seed-level change:

```
1. Create seeds/seed-XXX/
2. Copy Seed-001 contents
3. Apply reasoning changes
4. Update seed.yaml
5. Update SEED-XXX README
6. Document migration
7. Freeze new seed
```

### Step 4: Update Engine Compatibility

Update engines to declare compatibility:
```yaml
compatible_seeds:
  - SEED-001
  - SEED-002  # New seed
```

---

## Examples

### Example 1: New Evidence Model

**Change**: KDE introduces quantitative evidence standards

**Decision**: CREATE NEW SEED

**Reason**: Evidence model is foundational reasoning DNA

### Example 2: New Validation Pipeline

**Change**: Add new validation stage to experiment pipeline

**Decision**: UPDATE ENGINE

**Reason**: Pipeline is methodology, not reasoning DNA

### Example 3: New Experiment Type

**Change**: Add A/B testing capability

**Decision**: UPDATE LABORATORY

**Reason**: Execution context, not fundamental reasoning

### Example 4: New Knowledge Format

**Change**: Knowledge objects include new fields

**Decision**: UPDATE ENGINE

**Reason**: Knowledge model evolution, not core reasoning

---

## Seed Evolution

### Inherit, Don't Modify

Future Seeds should:
- **Inherit** from previous Seeds
- **Extend** rather than replace
- **Document** what changed from previous

### Example: SEED-002

```
SEED-001 (Genesis)
    │
    │ Inherits from
    │
    ▼
SEED-002
    │
    ├── Core Principles (inherited from SEED-001)
    ├── New Principle A (added)
    ├── Model X (modified)
    └── Model Y (new)
```

---

## Anti-Patterns

### Do NOT Create New Seed When

1. **Engine wants to evolve** → Update Engine
2. **Laboratory needs new features** → Update Laboratory
3. **Knowledge needs new format** → Update Engine knowledge model
4. **Process needs improvement** → Update respective subsystem
5. **Tooling needs update** → Update tooling

### Seed Creation is Rare

Creating a new Seed should be **exceptional**, not common.

Most changes are:
- Engine improvements (Engine version bump)
- Process improvements (Laboratory update)
- Tool updates (Tooling changes)

---

## Related Documents

- [WHAT-IS-A-SEED.md](./WHAT-IS-A-SEED.md) — Seed definition
- [NEVER-MODIFY.md](./NEVER-MODIFY.md) — Immutability rules
- [seed.yaml](./seed.yaml) — Seed manifest

---

**Status**: FOUNDATIONAL DEFINITION
