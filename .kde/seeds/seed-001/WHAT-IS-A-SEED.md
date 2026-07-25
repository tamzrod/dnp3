# What is a Seed?

**Seed ID**: SEED-001
**Version**: 1.0.0

---

## Definition

A **Seed** is the immutable, foundational layer of KDE reasoning. It contains the core DNA that defines how KDE discovers, validates, and evolves knowledge.

### Seed Properties

| Property | Description |
|----------|-------------|
| **Immutable** | Never modified after creation |
| **Versioned** | Unique version for each reasoning generation |
| **Reproducible** | Enables experiment reproducibility |
| **Complete** | Contains all fundamental reasoning elements |
| **Frozen** | Locked once created |

---

## What a Seed IS

### A Seed IS:

- **The Reasoning DNA**: Core principles that define KDE's thinking
- **Immutable**: Frozen at creation, never changed
- **Versioned**: Each generation has a unique version
- **Complete**: Contains all fundamental reasoning elements
- **Foundational**: The base layer upon which engines build

### A Seed IS NOT:

| Not This | Reason |
|----------|--------|
| **An Engine** | Engines implement methodology, not reasoning DNA |
| **A Laboratory** | Laboratory executes experiments |
| **A Dataset** | Seeds don't contain data |
| **Knowledge** | Knowledge is discovered, not foundational |
| **Experiment Results** | Results come from experiments |

---

## Seed Hierarchy

```
┌─────────────────────────────────────────────────────────┐
│                      K D E                                │
└─────────────────────────────────────────────────────────┘

                          │
                          ▼
               ┌─────────────────────┐
               │        Seed         │
               │                     │
               │  Reasoning DNA      │
               │  Core Principles   │
               │  Foundational      │
               │  Models            │
               │                     │
               │  STATUS: FROZEN    │
               └──────────┬──────────┘
                          │
                          │ consumed by
                          ▼
               ┌─────────────────────┐
               │       Engine        │
               │                     │
               │  Methodology        │
               │  Implementation     │
               │  Pipeline           │
               │                     │
               │  STATUS: Evolving   │
               └──────────┬──────────┘
                          │
                          │ executes under
                          ▼
               ┌─────────────────────┐
               │     Laboratory      │
               │                     │
               │  Experiment        │
               │  Execution         │
               │  Observation       │
               │                     │
               │  STATUS: Active     │
               └─────────────────────┘
```

---

## Seed Contents

A Seed contains:

### 1. Core Principles

The fundamental rules governing AI behavior:
- No Auto-Continuation
- No Self-Approval
- No Self-Promotion
- Evidence-Inference-Hypothesis distinction
- Evidence-Based Changes

### 2. Scientific Learning Loop

The continuous improvement cycle:
- Research → Knowledge → Laboratory → Evidence → Governance → Research

### 3. Foundational Models

- Evidence Model: What counts as evidence
- Knowledge Model: What constitutes knowledge
- Confidence Model: How confidence is assigned
- Ambiguity Model: How uncertainty is handled

---

## Why Seeds Exist

### Problem Seeds Solve

1. **Methodology vs Reasoning Confusion**
   - Engines define methodology (evolving)
   - Seeds define reasoning (immutable)

2. **Historical Validity**
   - Old experiments remain valid
   - Original reasoning preserved

3. **Evolution Without Modification**
   - New reasoning → new Seed
   - Never modify historical reasoning

### Benefits

| Benefit | Description |
|---------|-------------|
| **Reproducibility** | Experiments valid in original context |
| **Provenance** | Clear lineage from reasoning to results |
| **Evolution** | New Seeds for new reasoning |
| **Immutability** | Historical reasoning never changed |

---

## Seed Lifecycle

```
┌─────────────────────────────────────────────────────────┐
│                    SEED LIFECYCLE                        │
└─────────────────────────────────────────────────────────┘

         ┌─────────────────────────────────────┐
         │           SEED CREATION             │
         │                                       │
         │  1. Identify reasoning DNA            │
         │  2. Migrate foundational elements    │
         │  3. Create seed manifest             │
         │  4. Document seed contents           │
         └──────────────┬────────────────────────┘
                        │
                        ▼
         ┌─────────────────────────────────────┐
         │           SEED ACTIVE               │
         │                                       │
         │  - Engines consume seed               │
         │  - Experiments execute under seed    │
         │  - Knowledge discovered              │
         └──────────────┬────────────────────────┘
                        │
                        ▼
         ┌─────────────────────────────────────┐
         │           SEED FROZEN               │
         │                                       │
         │  - Seed becomes immutable             │
         │  - No modifications allowed          │
         │  - Future seeds inherit from this     │
         └──────────────┬────────────────────────┘
                        │
                        ▼
              (If reasoning changes)
                        │
                        ▼
              ┌─────────────────────┐
              │   NEW SEED CREATED  │
              │                     │
              │  SEED-002 → etc.   │
              └─────────────────────┘
```

---

## Seed vs Engine Comparison

| Aspect | Seed | Engine |
|--------|------|--------|
| **Purpose** | Reasoning DNA | Methodology |
| **Changes** | NEVER | When improved |
| **Status** | FROZEN | EVOLVING |
| **Contains** | Principles, Models | Pipeline, Rules |
| **Version** | Semantic (1.0.0) | Semantic (0.1.0) |
| **Immutable** | YES | NO |

---

## Related Documents

- [WHEN-TO-CREATE.md](./WHEN-TO-CREATE.md) — When to create new Seeds
- [NEVER-MODIFY.md](./NEVER-MODIFY.md) — Immutability rules
- [seed.yaml](./seed.yaml) — Seed manifest

---

**Status**: FOUNDATIONAL DEFINITION
