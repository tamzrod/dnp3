# What is a Seed?

**Seed ID**: SEED-002
**Version**: 1.0.0
**Parent**: SEED-001

---

## Definition

A **Seed** is the immutable, foundational layer of KDE reasoning. It contains the core DNA that defines how KDE discovers, validates, and evolves knowledge.

A Seed is NOT a repository snapshot. It is the accumulated reasoning DNA produced after surviving an evolutionary cycle.

---

## Seed Lifecycle

```
┌─────────────────────────────────────────────────────────────┐
│                    SEED LIFECYCLE                              │
└─────────────────────────────────────────────────────────────┘

         Seed
            │
            ▼
    Development
    (KDE builds with Seed)
            │
            ▼
      Experiments
    (Testing, validating)
            │
            ▼
        Evidence
    (Accumulated results)
            │
            ▼
    Lessons Learned
    (What worked, what didn't)
            │
            ▼
       New Seed
    (Evolution from lessons)
```

**If no lessons were learned, no new Seed should exist.**

---

## What Makes Seed-002 Different from Seed-001

Seed-002 is NOT a clone of Seed-001. It exists because:

1. **10 lessons were learned** from actual KDE development
2. **Evidence exists** for each lesson
3. **Improvements are traceable** to real problems

### Evidence

Every modification in Seed-002 traces to:
- Specific lesson from KDE development
- Observable impact on KDE operation
- Clear root cause identification
- Defined resolution approach

### Prohibited

- No speculative improvements
- No architectural changes based on opinion
- No modifications without lesson documentation

---

## What a Seed IS

| Property | Description |
|----------|-------------|
| **Immutable** | Never modified after freezing |
| **Versioned** | Unique version for each generation |
| **Reproducible** | Enables experiment reproducibility |
| **Complete** | Contains all reasoning DNA |
| **Lineaged** | References parent Seed |

---

## What a Seed IS NOT

| Not This | Reason |
|----------|--------|
| **An Engine** | Engines implement methodology, not reasoning |
| **A Laboratory** | Laboratory executes experiments |
| **A Dataset** | Seeds don't contain data |
| **Knowledge** | Knowledge is discovered |
| **Experiment Results** | Results come from experiments |

---

## Seed Hierarchy

```
┌─────────────────────────────────────────────────────────────┐
│                         K D E                                  │
└─────────────────────────────────────────────────────────────┘

                          │
                          ▼
               ┌─────────────────────┐
               │        Seed           │
               │                     │
               │  Reasoning DNA       │
               │  Immutable          │
               │  Versioned         │
               │                     │
               │  SEED-002: Evolution│
               │  [FROZEN]          │
               └──────────┬──────────┘
                          │
                          │ consumes
                          ▼
               ┌─────────────────────┐
               │       Engine        │
               │                     │
               │  Methodology        │
               │  Evolvable         │
               │                     │
               │  [Evolving]        │
               └──────────┬──────────┘
                          │
                          │ executes under
                          ▼
               ┌─────────────────────┐
               │     Laboratory      │
               │                     │
               │  Experiment         │
               │  Execution         │
               │                     │
               │  [Active]          │
               └─────────────────────┘
```

---

## Seed Contents

### From Parent (Inherited)

Seeds inherit valid components from parent:
- Core principles that remain sound
- Models that work as intended
- Foundations that don't need change

### New (Evolved)

Seeds add lessons-learned improvements:
- Solutions to identified problems
- Enhanced clarity
- New requirements discovered through experience

---

## Why Seeds Exist

### Problem Seeds Solve

| Problem | Solution |
|---------|----------|
| Reasoning embedded in Engine | Separate immutable Seed |
| Historical context lost | Seeds preserve lineage |
| Can't compare generations | Each Seed documents changes |
| Evolution loses history | New Seeds inherit from previous |

### Benefits

| Benefit | Description |
|---------|-------------|
| **Reproducibility** | Experiments valid in original context |
| **Provenance** | Clear lineage from reasoning to results |
| **Evolution** | New Seeds inherit without modifying |
| **Immutability** | Historical reasoning never changed |

---

## Seed vs Engine vs Laboratory

| Aspect | Seed | Engine | Laboratory |
|--------|------|--------|------------|
| **Contains** | Reasoning DNA | Methodology | Experiments |
| **Changes** | NEVER | When improved | When needed |
| **Status** | FROZEN | EVOLVING | ACTIVE |
| **Versioning** | Semantic (1.0.0) | Semantic (0.1.0) | By experiment |
| **Immutable** | YES | NO | NO |

---

## Creating New Seeds

### When to Create

A new Seed should be created when:
1. Lessons were learned from previous Seed
2. Fundamental reasoning changes
3. Evidence supports architectural change

### When NOT to Create

Do NOT create new Seed when:
- No lessons were learned
- Only methodology changed (use Engine version)
- Only process changed (update Laboratory)

### Process

```
1. Document lessons from previous Seed
2. Create design objectives
3. Implement only what lessons require
4. Reference parent Seed
5. Freeze new Seed
```

---

## Lineage Documentation

Every Seed must document:
- What was inherited from parent
- What changed and why
- What compatibility is preserved

This enables:
- Understanding why decisions were made
- Tracing evolution of reasoning
- Comparing generations

---

## Related Documents

- [seed.yaml](./seed.yaml) — Machine-readable manifest
- [LESSONS-LEARNED.md](./LESSONS-LEARNED.md) — Lessons from SEED-001
- [DESIGN-OBJECTIVES.md](./DESIGN-OBJECTIVES.md) — Objectives for SEED-002

---

**Status**: DEFINITION
**Version**: 1.0.0
