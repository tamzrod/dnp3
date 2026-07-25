# What Must Never Be Modified in a Seed

**Seed ID**: SEED-001
**Version**: 1.0.0
**Status**: FROZEN

---

## Immutability Rule

**Once a Seed is frozen, it shall NEVER be modified.**

This document defines what must never be changed within a frozen Seed.

---

## Absolute Prohibitions

### Never Modify

| Element | Reason |
|---------|--------|
| **Core Principles** | Define fundamental AI behavior |
| **Scientific Loop** | Defines knowledge evolution cycle |
| **Evidence Model** | Defines what counts as evidence |
| **Knowledge Model** | Defines what constitutes knowledge |
| **Confidence Model** | Defines how confidence is assigned |
| **Ambiguity Handling** | Defines how uncertainty is processed |
| **Seed Manifest** | Documents frozen state |
| **Seed Status** | Frozen status is permanent |

---

## Rationale

### Why Seeds Are Immutable

1. **Historical Validity**
   - Old experiments remain valid in their original context
   - Reasoning that produced results is preserved

2. **Reproducibility**
   - Future experiments can reference frozen reasoning
   - No risk of retroactive changes

3. **Provenance**
   - Clear lineage from reasoning to experiment to knowledge
   - Every result traces to its reasoning generation

4. **Trust**
   - Users can rely on Seed content never changing
   - No surprise modifications

---

## What Happens If You Need to Change?

### If Core Reasoning Changes

```
DO NOT MODIFY SEED-001

Instead:
1. Create SEED-002
2. Inherit from SEED-001
3. Apply changes
4. Freeze SEED-002
```

### If Methodology Changes

```
DO NOT MODIFY SEEDS

Instead:
1. Update the Engine
2. Create new Engine version
3. Update engine compatibility list
```

### If Process Changes

```
DO NOT MODIFY SEEDS

Instead:
1. Update Laboratory
2. Update relevant subsystem
```

---

## Migration Path

### When Seed-Level Change Needed

```
┌─────────────────────────────────────────────────────────┐
│                   CHANGE DECISION                        │
└─────────────────────────────────────────────────────────┘

         Is this a FUNDAMENTAL REASONING change?

                    ┌─────────┴─────────┐
                    │                   │
                   YES                  NO
                    │                   │
                    ▼                   ▼
           ┌───────────────┐   ┌───────────────────┐
           │ CREATE NEW    │   │ What type of      │
           │    SEED       │   │     change?       │
           └───────────────┘   └─────────┬─────────┘
                                          │
                              ┌───────────┼───────────┐
                              │           │           │
                          Engine      Laboratory    Tooling
                              │           │           │
                              ▼           ▼           ▼
                        Update      Update      Update
                        Engine      Lab         Tool
```

---

## Forbidden Actions

### Within a Frozen Seed

**NEVER**:

- [ ] Modify principles
- [ ] Change loop structure
- [ ] Update evidence model
- [ ] Modify knowledge definition
- [ ] Alter confidence methodology
- [ ] Change ambiguity handling
- [ ] Edit seed.yaml status
- [ ] Remove seed contents
- [ ] Add new reasoning elements
- [ ] Clarify ambiguous language
- [ ] Fix errors
- [ ] Update examples
- [ ] Improve documentation
- [ ] Update formatting

### Rationale

Even "improvements" are forbidden because:
- They change historical context
- They invalidate old provenance
- They break reproducibility guarantees

---

## What CAN Be Done

### Outside Frozen Seeds

- [x] Create new Seed
- [x] Update Engine
- [x] Update Laboratory
- [x] Update Knowledge
- [x] Update Governance
- [x] Update Research
- [x] Update any subsystem

### Within Frozen Seeds

Only:
- [x] Read
- [x] Reference
- [x] Copy (to new Seed)

---

## Enforcement

### Technical

- Seed status marked FROZEN
- No write permissions on seed files
- Version control protection

### Process

- Governance approval required for any exception
- Exception means new Seed creation

### Culture

- Seeds are sacred
- Immutable reasoning enables trust
- Evolution through new Seeds, not modification

---

## Exceptions

### There Are No Exceptions

The only "exception" to immutability is:
- Creating a new Seed
- Never modifying an existing one

---

## Related Documents

- [WHAT-IS-A-SEED.md](./WHAT-IS-A-SEED.md) — Seed definition
- [WHEN-TO-CREATE.md](./WHEN-TO-CREATE.md) — Creation guidelines
- [seed.yaml](./seed.yaml) — Seed manifest

---

**Status**: FROZEN
**Modifiable**: NO
**Exception**: NONE
