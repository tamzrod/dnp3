# Boundaries: Explicit Ownership

**Seed ID**: SEED-002
**Section**: Boundaries (NEW)
**Addresses**: LESSON-002, LESSON-007

---

## Purpose

Boundaries define explicit ownership. Each KDE component has one clear owner and defined responsibilities.

**Problem Addressed**: LESSON-002 - "Architectural boundaries became blurred"

---

## Boundary Principle

> "Define what belongs where BEFORE implementation."

---

## KDE Ownership Matrix

| Component | Owner | Contains | Does NOT Contain |
|-----------|-------|---------|------------------|
| **Seed** | KDE (immutable) | Reasoning DNA, Principles, Models | Methodology, Engine-specific rules |
| **Engine** | KDE (evolvable) | Methodology, Pipeline, Rules | Reasoning DNA (consumed from Seed) |
| **Laboratory** | KDE (active) | Questions, Experiments, Evidence, Execution | Reasoning, Knowledge definition |
| **Knowledge** | KDE (passive) | Validated definitions | Reasoning, Methodology |
| **Governance** | KDE (oversight) | Decisions, Approvals, Standards | Execution, Implementation |

---

## What Belongs to Each Component

### Seed (Immutable)

```
Seed/
├── principles/          # ✅ Core principles
├── reasoning/           # ✅ How KDE reasons
├── boundaries/           # ✅ Ownership definitions
├── evolution/           # ✅ Change process
└── validation/         # ✅ Validation requirements

Seed does NOT contain:
├── engine-specific rules    # ❌ Belongs to Engine
├── experiment procedures     # ❌ Belongs to Laboratory
├── knowledge content         # ❌ Belongs to Knowledge
```

### Engine (Evolving)

```
Engine/
├── methodology/          # ✅ How experiments run
├── pipeline/            # ✅ Processing stages
├── rules/               # ✅ Experiment rules
└── compatibility/       # ✅ Seed compatibility

Engine does NOT contain:
├── core principles           # ❌ Belongs to Seed
├── fundamental models        # ❌ Belongs to Seed
├── experiment results         # ❌ Belongs to Laboratory
```

### Laboratory (Active)

```
Laboratory/
├── questions/           # ✅ Research questions
├── experiments/         # ✅ Experiment execution
├── evidence/            # ✅ Evidence collection
├── templates/          # ✅ Standard templates
└── runs/                # ✅ Run management

Laboratory does NOT contain:
├── reasoning principles      # ❌ Belongs to Seed
├── methodology definition    # ❌ Belongs to Engine
├── knowledge content         # ❌ Belongs to Knowledge
```

### Knowledge (Passive)

```
Knowledge/
├── definitions/         # ✅ Validated knowledge
├── contexts/           # ✅ Applicability conditions
├── boundaries/         # ✅ Limitations
└── versions/          # ✅ Evolution tracking

Knowledge does NOT contain:
├── reasoning DNA           # ❌ Belongs to Seed
├── experiment methodology   # ❌ Belongs to Engine
```

### Governance (Oversight)

```
Governance/
├── decisions/          # ✅ Approval decisions
├── reviews/            # ✅ Review processes
├── policies/           # ✅ Governance rules
└── oversight/          # ✅ Compliance

Governance does NOT contain:
├── experiment execution    # ❌ Belongs to Laboratory
├── knowledge creation      # ❌ Belongs to Laboratory
```

---

## Interface Rules

Components communicate through defined interfaces:

| From | To | Interface |
|------|-----|-----------|
| Seed | Engine | Compatibility declaration |
| Engine | Laboratory | Experiment methodology |
| Laboratory | Knowledge | Validation results |
| Knowledge | Governance | Knowledge status |
| Governance | Laboratory | Gap identification |

---

## Cross-Boundary Prevention

### Seed Cannot Reference
- Engine-specific implementation
- Laboratory procedures
- Knowledge content

### Engine Cannot Reference
- Seed internal structure
- Laboratory experiment data
- Knowledge definitions

### Laboratory Cannot Reference
- Seed reasoning implementation
- Engine specific rules
- Governance decisions

---

## Boundary Enforcement

Boundaries are enforced by:

1. **Document Structure**: Each component has explicit directory ownership
2. **Cross-Reference Prohibition**: Documents must not import from other components
3. **Interface Documentation**: Communication through documented interfaces only

---

## Changes from Seed-001

| Aspect | Seed-001 | Seed-002 |
|--------|----------|----------|
| Boundary Definition | Implicit | **Explicit** |
| Ownership Matrix | Missing | **Defined** |
| Interface Rules | Missing | **Documented** |
| Cross-Reference Rules | Missing | **Defined** |

---

## Lessons Addressed

- **LESSON-002**: "Architectural boundaries became blurred"
- **LESSON-007**: "Architecture became coupled by growth"

---

**Status**: NEW IN SEED-002
**Evidence**: LESSON-002, LESSON-007
