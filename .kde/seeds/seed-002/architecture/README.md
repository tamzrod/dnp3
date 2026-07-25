# Architecture: KDE System Structure

**Seed ID**: SEED-002
**Section**: Architecture

---

## System Overview

KDE is an evidence-based knowledge discovery system. It discovers knowledge through experiments, validates through evidence, and evolves through lessons learned.

---

## Component Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         K D E                                      │
└─────────────────────────────────────────────────────────────────┘

                          ┌─────────────────────┐
                          │        Seed           │
                          │                     │
                          │ Reasoning DNA       │
                          │ Immutable           │
                          │ Versioned           │
                          │                     │
                          │ SEED-002: Evolution │
                          │ [FROZEN]           │
                          └──────────┬──────────┘
                                     │
                                     │ consumed by
                                     ▼
                          ┌─────────────────────┐
                          │       Engine          │
                          │                     │
                          │ Methodology          │
                          │ Evolvable            │
                          │ Boundary-compliant   │
                          │                     │
                          │ KDE-ENGINE-XXX     │
                          └──────────┬──────────┘
                                     │
                                     │ executes under
                                     ▼
                          ┌─────────────────────┐
                          │     Laboratory       │
                          │                     │
                          │ Experiment          │
                          │ Execution           │
                          │ Standards-compliant │
                          │                     │
                          │ LAB-XXX            │
                          └──────────┬──────────┘
                                     │
                                     │ validates
                                     ▼
                          ┌─────────────────────┐
                          │      Knowledge       │
                          │                     │
                          │ Validated           │
                          │ Definitions         │
                          │ Boundary-compliant   │
                          │                     │
                          │ KNOW-XXX           │
                          └─────────────────────┘
```

---

## Component Responsibilities

### Seed
**Owner**: KDE (immutable)
**Purpose**: Contains reasoning DNA
**Changes**: NEVER

| Contains | Does NOT Contain |
|----------|------------------|
| Principles | Methodology |
| Models | Engine rules |
| Boundaries | Laboratory procedures |
| Evolution process | Knowledge content |

### Engine
**Owner**: KDE (evolvable)
**Purpose**: Implements methodology
**Changes**: When methodology improves

| Contains | Does NOT Contain |
|----------|------------------|
| Methodology | Core principles |
| Pipeline | Reasoning models |
| Rules | Knowledge |
| Seed compatibility | Laboratory data |

### Laboratory
**Owner**: KDE (active)
**Purpose**: Executes experiments
**Changes**: When execution needs improvement

| Contains | Does NOT Contain |
|----------|------------------|
| Experiments | Reasoning |
| Runs | Methodology definition |
| Evidence | Knowledge definitions |
| Templates | Governance decisions |

### Knowledge
**Owner**: KDE (passive)
**Purpose**: Stores validated knowledge
**Changes**: When new knowledge discovered

| Contains | Does NOT Contain |
|----------|------------------|
| Definitions | Reasoning DNA |
| Contexts | Methodology |
| Boundaries | Experiments |
| Versions | Governance |

---

## Communication Flow

### Seed → Engine
```
Seed declares compatible engines
Engine references seed
Engine implements seed reasoning
```

### Engine → Laboratory
```
Engine provides methodology
Laboratory executes under engine
Laboratory reports to engine
```

### Laboratory → Knowledge
```
Laboratory validates knowledge
Laboratory generates evidence
Knowledge stores validated definitions
```

### Knowledge → Governance
```
Knowledge reports status
Governance reviews knowledge
Governance approves changes
```

### Governance → Research
```
Governance identifies gaps
Governance directs research
Research proposes knowledge
```

---

## Directory Structure

The canonical KDE architecture follows a five-directory structure:

```
kde/
├── seeds/                    # Immutable reasoning DNA
│   ├── seed-001/           # Genesis (FROZEN)
│   └── seed-002/           # Evolution (FROZEN)
│
├── engines/                  # Evolvable methodology implementations
│   ├── alpha/              # KDE-ENGINE-001
│   ├── beta/               # KDE-ENGINE-002
│   └── gamma/              # KDE-ENGINE-003
│
├── laboratory/              # Scientific workflow
│   ├── questions/          # Research questions
│   ├── experiments/        # LAB-XXX
│   └── evidence/           # Collected evidence
│
├── knowledge/               # Validated knowledge definitions
│
└── governance/              # Repository rules and standards
```

---

## Boundary Compliance

### Seed Boundaries
- Seed contains only reasoning DNA
- Seed references compatible engines
- Seed never modified after freeze

### Engine Boundaries
- Engine implements methodology
- Engine consumes seed
- Engine evolves independently

### Laboratory Boundaries
- Laboratory executes experiments
- Laboratory follows engine methodology
- Laboratory produces evidence

### Knowledge Boundaries
- Knowledge stores validated definitions
- Knowledge does not create knowledge
- Knowledge evolves through validation

---

## Evolution Path

```
SEED-001 (Genesis)
    │
    │ Lessons learned
    ▼
SEED-002 (Evolution)
    │
    │ When lessons learned
    ▼
SEED-003 (Future)
    │
    │ etc.
    ▼
```

---

## Related Documents

- [boundaries/README.md](../boundaries/README.md) — Explicit ownership
- [evolution/README.md](../evolution/README.md) — Change process

---

**Inherited from**: SEED-001
**Changes**: Added boundary compliance requirements
