# KDE-ENGINE-003 (Gamma) Provenance

**Engine ID**: KDE-ENGINE-003
**Version**: 0.1.0
**Codename**: Gamma

---

## Engine Lineage

```
KDE-ENGINE-001 (Alpha)
        │
        │ Pattern Discovery
        │
        ▼
KDE-ENGINE-002 (Beta)
        │
        │ Context Discovery + Statistical Validation
        │
        ▼
KDE-ENGINE-003 (Gamma) ← CURRENT
        │
        │ Causal Discovery
        │
        ▼
[KDE-ENGINE-004 (Delta)] — Planned
```

---

## Engine Evolution History

### KDE-ENGINE-001 (Alpha) — Pattern Discovery

| Attribute | Value |
|-----------|-------|
| **Launch Date** | 2026-07-20 |
| **Approach** | Pattern Detection |
| **Key Innovation** | First structured knowledge discovery |
| **Limitations** | No context, no boundaries, qualitative confidence |

### KDE-ENGINE-002 (Beta) — Context Discovery

| Attribute | Value |
|-----------|-------|
| **Launch Date** | 2026-07-20 |
| **Parent** | KDE-ENGINE-001 (Alpha) |
| **Approach** | Contextual Knowledge Discovery |
| **Key Innovation** | Context detection, boundary detection, statistical confidence |
| **Improvements** | Quantitative confidence, applicability conditions |

### KDE-ENGINE-003 (Gamma) — Causal Discovery

| Attribute | Value |
|-----------|-------|
| **Launch Date** | 2026-07-20 |
| **Parent** | KDE-ENGINE-002 (Beta) |
| **Approach** | Causal Knowledge Discovery |
| **Key Innovation** | Causal inference, mechanism identification, intervention prediction |
| **Improvements** | Causal models, confounder adjustment, effect estimation |

---

## Architectural Decisions

### Decision 1: Extend Beta Pipeline

**Decision**: Add causal discovery stages rather than creating new pipeline.

**Rationale**:
- Reuse validated Beta components
- Maintain compatibility where possible
- Focus new development on causal modules

**Alternatives Considered**:
- Create completely new pipeline → Rejected (too much divergence)
- Modify Beta stages → Rejected (would break Beta compatibility)

### Decision 2: Causal Knowledge as Separate Object

**Decision**: Create CAUSAL-XXX knowledge objects distinct from KNOW-XXX objects.

**Rationale**:
- Causal knowledge has fundamentally different structure
- Prevents accidental mixing of correlation and causation
- Clear separation enables comparison

**Alternatives Considered**:
- Extend Beta KNOW objects → Rejected (too much complexity)
- Create entirely new knowledge schema → Accepted (causal specificity)

### Decision 3: Experimental Status

**Decision**: Release Gamma as Experimental, not Active.

**Rationale**:
- Causal discovery methodology requires validation
- Laboratory experiments needed before production use
- Prevents premature reliance on unvalidated methodology

**Alternatives Considered**:
- Release as Active → Rejected (insufficient validation)
- Delay release → Rejected (experiments can validate in use)

### Decision 4: Causal Principles as Extensions

**Decision**: Add Gamma-specific principles to existing 10 principles.

**Rationale**:
- Gamma builds on Beta, which builds on Alpha
- Principles should accumulate, not replace
- Maintains lineage awareness

**Alternatives Considered**:
- Replace with new principles → Rejected (breaks lineage)
- Start fresh at 5 principles → Rejected (loses context)

---

## Design Philosophy

### Principle of Cumulative Development

Each engine builds on its predecessors:

| Engine | Adds to Previous |
|--------|-----------------|
| Alpha | Pattern discovery foundation |
| Beta | + Context, boundaries, statistics |
| Gamma | + Causal inference, mechanisms |
| Delta | + Temporal patterns (planned) |

### Principle of Incremental Innovation

New engines extend, not replace:

```
Alpha ──→ Beta ──→ Gamma ──→ Delta
  │        │        │
  └──10────┴──15────┘──20 (principles accumulate)
```

### Principle of Experimental Validation

New engines require validation:

| Engine | Status | Validation |
|--------|--------|------------|
| Alpha | Historical | Complete |
| Beta | Active | Substantial |
| Gamma | Experimental | Pending |

---

## Engine Relationships

### Parent-Child Relationships

```
Alpha (001)
    │
    └── Beta (002)
            │
            └── Gamma (003)
```

### Engine Comparison

| Aspect | Alpha | Beta | Gamma |
|--------|-------|------|-------|
| **Discovery Type** | Pattern | Context | Causal |
| **Question** | "Does X?" | "When X?" | "How X?" |
| **Confidence** | Qualitative | Statistical | Statistical + Causal |
| **Context** | None | Required | Required |
| **Boundaries** | None | Required | Required |
| **Mechanism** | No | No | Yes |
| **Confounders** | No | No | Yes |
| **Intervention** | No | No | Yes |

---

## Experiment Lineage

### Experiments Using Each Engine

| Engine | Experiments | Status |
|--------|-------------|--------|
| Alpha | LAB-001 to LAB-011 | Historical |
| Beta | LAB-012 to LAB-015 | Active |
| Gamma | [Pending validation] | Experimental |

### Historical Experiments

| Experiment | Engine | Purpose |
|------------|--------|---------|
| LAB-001 to LAB-011 | Alpha | Pattern discovery era |
| LAB-012 to LAB-015 | Beta | Contextual knowledge era |
| LAB-016+ | Gamma | Causal discovery era (planned) |

---

## Version Information

### Current Version

| Field | Value |
|-------|-------|
| **Engine ID** | KDE-ENGINE-003 |
| **Version** | 0.1.0 |
| **Codename** | Gamma |
| **Status** | Experimental |
| **Created** | 2026-07-20 |

### Version Components

| Component | Value | Description |
|-----------|-------|-------------|
| **Major** | 0 | Pre-release |
| **Minor** | 1 | First minor version |
| **Patch** | 0 | Initial release |

### Upgrade Path

```
0.1.0 (Experimental)
    │
    ├── 0.2.0 (Alpha validation complete)
    │
    ├── 0.3.0 (Beta validation complete)
    │
    └── 1.0.0 (Production ready)
```

---

## Related Documents

- [specification.md](./specification.md) — Engine identity
- [methodology.md](./methodology.md) — Full methodology
- [pipeline.md](./pipeline.md) — Pipeline details
- [knowledge-model.md](./knowledge-model.md) — Knowledge structure
- [changes.md](./changes.md) — Version history

---

**Document Status**: EXPERIMENTAL
