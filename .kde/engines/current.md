# Current KDE Engine

**Last Updated**: 2026-07-24 (INV-EVOLUTION-001 REC-003/REC-004)

---

## Engine Selection

The Laboratory supports multiple engines. Select the appropriate engine for each experiment:

| Engine ID | Version | Codename | Status | Default | Use Case |
|-----------|---------|----------|--------|---------|----------|
| **KDE-ENGINE-002** | 0.1.0 | Beta | Active | **YES** | Contextual knowledge discovery (default) |
| KDE-ENGINE-001 | 0.1.0 | Alpha | Historical | NO | Pattern discovery (legacy) |
| **KDE-ENGINE-003** | 0.1.0 | Gamma | **Active** | NO | Causal discovery (INV-EVOLUTION-001) |
| **KDE-ENGINE-004** | 0.1.0 | Delta | **Active** | NO | Bootstrap + Context discovery (INV-EVOLUTION-001) |

---

## Understanding Status vs. Default

**IMPORTANT**: Engine "Status" and "Default" are independent concepts.

### Status (Scientific Lifecycle)

Status describes the scientific validation level of an Engine:

| Status | Description |
|--------|-------------|
| **Experimental** | New engine, methodology under development |
| **Candidate** | Under validation |
| **Candidate (Validated)** | Validation complete |
| **Active** | Scientifically validated and available |
| **Historical** | Former engine, preserved for reference |

### Default (Runtime Configuration)

Default describes the Runtime's operational default:

| Default | Description |
|---------|-------------|
| **YES** | Engine is loaded at Runtime startup |
| **NO** | Engine requires explicit selection |

### Key Distinction

| Scenario | Status | Default | Explanation |
|----------|--------|---------|-------------|
| Beta | Active | YES | Scientifically active AND Runtime default |
| Gamma | Active | NO | Causal discovery capability, not default |
| Delta | Active | NO | Bootstrap capability, not default |
| Future Engine | Active | NO | Could be scientifically superior but human hasn't selected as default |

**Principle**: Scientific validation does NOT automatically change Runtime defaults. Only human authority can change defaults.

---

## Default Engine Authority

**Runtime defaults are under exclusive human authority.**

See [/governance/runtime/defaults.yaml](/workspace/project/kde/governance/runtime/defaults.yaml) for the authoritative Runtime configuration.

See [/governance/runtime/RUNTIME-STARTUP.md](/workspace/project/kde/governance/runtime/RUNTIME-STARTUP.md) for the startup sequence.

---

## Active Engine

The **default active** KDE Engine is:

| Field | Value |
|-------|-------|
| **Engine ID** | KDE-ENGINE-002 |
| **Version** | 0.1.0 |
| **Codename** | Beta |
| **Name** | Contextual Knowledge Discovery Engine |
| **Status** | Active |

---

## Engine Details

### Beta (KDE-ENGINE-002) — Active

For detailed information about Beta, see:

- [beta/specification.md](./beta/specification.md) — Engine identity and scope
- [beta/methodology.md](./beta/methodology.md) — Detailed methodology
- [beta/pipeline.md](./beta/pipeline.md) — Processing pipeline
- [beta/knowledge-model.md](./beta/knowledge-model.md) — Knowledge object specification
- [beta/changes.md](./beta/changes.md) — Version history
- [beta/provenance.md](./beta/provenance.md) — Engine lineage

### Alpha (KDE-ENGINE-001) — Historical

For detailed information about Alpha, see:

- [alpha/specification.md](./alpha/specification.md) — Engine identity and scope
- [alpha/methodology.md](./alpha/methodology.md) — Detailed methodology
- [alpha/changes.md](./alpha/changes.md) — Version history
- [alpha/provenance.md](./alpha/provenance.md) — Engine lineage

### Delta (KDE-ENGINE-004) — Candidate

For detailed information about Delta, see:

- [delta/README.md](./delta/README.md) — Quick reference
- [delta/specification.md](./delta/specification.md) — Engine identity and scope
- [delta/methodology.md](./delta/methodology.md) — Detailed methodology
- [delta/pipeline.md](./delta/pipeline.md) — Processing pipeline
- [delta/knowledge-model.md](./delta/knowledge-model.md) — Knowledge object specification
- [delta/changes.md](./delta/changes.md) — Version history
- [delta/provenance.md](./delta/provenance.md) — Engine lineage

**Note**: Delta is a Candidate Engine (Validated). VAL-001 validation complete. Additional validation recommended before promotion.

---

## Engine Lineage

| Engine ID | Version | Codename | Status | Default | Effective | Purpose |
|-----------|---------|----------|--------|---------|-----------|---------|
| KDE-ENGINE-001 | 0.1.0 | Alpha | Historical | NO | 2026-07-19 | Pattern discovery |
| KDE-ENGINE-002 | 0.1.0 | Beta | Active | **YES** | 2026-07-20 | Contextual knowledge |
| KDE-ENGINE-003 | 0.1.0 | Gamma | **Active** | NO | 2026-07-20 | Causal discovery |
| KDE-ENGINE-004 | 0.1.0 | Delta | **Active** | NO | 2026-07-20 | Bootstrap + Context |

---

## Engine Selection

### When to Use Beta (KDE-ENGINE-002)

Use Beta for new experiments when:
- Context detection is required
- Boundary definition is important
- Statistical validation is mandatory
- Complete knowledge objects needed

### When to Use Alpha (KDE-ENGINE-001)

Use Alpha for experiments when:
- Legacy compatibility required
- Simple pattern discovery sufficient
- Historical comparison needed

### When to Use Gamma (KDE-ENGINE-003)

Use Gamma for experiments when:
- Root cause analysis required ("why does X cause Y?")
- Causal mechanism documentation needed
- Intervention prediction is required ("what if we change X?")
- Failure propagation analysis needed

**Gamma Selection Keywords**: why, cause, mechanism, intervention, what if, predict outcome, root cause, lead to

### When to Use Delta (KDE-ENGINE-004)

Use Delta for experiments when:
- Bootstrap enforcement required
- Reproducibility is critical
- Session initialization consistency needed

---

## Transition Rules

When a new engine becomes active:

1. Update this file to mark new engine as Active
2. Update the new engine's specification with status: Active
3. Update the old engine's specification with status: Historical
4. Document the transition in both engines' changes.md
5. Update the Engine Lineage table above

**Note**: Engine status transitions do NOT automatically change Runtime defaults. See Session Override for how to use non-default engines.

### Session Override

Experiments and Investigations may use non-default engines by specifying a session override:

```yaml
session_override:
  engine: KDE-ENGINE-004  # Use Delta instead of default (Beta)
```

See [/governance/runtime/SESSION-OVERRIDE.md](/workspace/project/kde/governance/runtime/SESSION-OVERRIDE.md) for details.

---

## Engine Details (Gamma)

### Gamma (KDE-ENGINE-003) — **Active**

For detailed information about Gamma, see:

- [gamma/README.md](./gamma/README.md) — Quick reference
- [gamma/specification.md](./gamma/specification.md) — Engine identity and scope
- [gamma/methodology.md](./gamma/methodology.md) — Detailed methodology
- [gamma/pipeline.md](./gamma/pipeline.md) — Processing pipeline
- [gamma/knowledge-model.md](./gamma/knowledge-model.md) — Knowledge object specification
- [gamma/changes.md](./gamma/changes.md) — Version history
- [gamma/provenance.md](./gamma/provenance.md) — Engine lineage

**Activation Evidence**: LAB-017, LAB-044, LAB-045, LAB-046 (Repeatability: 100%)
**Activation Date**: 2026-07-24 (INV-EVOLUTION-001 REC-003)

---

## Migration History

| Date | Transition | Rationale |
|------|------------|-----------|
| 2026-07-19 | Initial Engine | KDE-ENGINE-001 (Alpha) established |
| 2026-07-20 | Engine Evolution | KDE-ENGINE-002 (Beta) released, Alpha → Historical |
| 2026-07-20 | Engine Expansion | KDE-ENGINE-003 (Gamma) experimental release |
| 2026-07-20 | Bootstrap Engine | KDE-ENGINE-004 (Delta) created, Candidate |
| 2026-07-22 | Delta Promotion | Delta: Candidate → Experimental (LAB-DELTA-VALIDATION-001) |
| 2026-07-23 | Gamma Promotion | Gamma: Experimental → Candidate (LAB-045/046) |
| 2026-07-24 | **Gamma Activation** | Gamma: Candidate → Active (INV-EVOLUTION-001 REC-003) |
| 2026-07-24 | **Delta Activation** | Delta: Experimental → Active (INV-EVOLUTION-001 REC-004) |

---

## Related Documents

### Runtime Configuration

- [/governance/runtime/defaults.yaml](/workspace/project/kde/governance/runtime/defaults.yaml) — Runtime default configuration
- [/governance/runtime/RUNTIME-STARTUP.md](/workspace/project/kde/governance/runtime/RUNTIME-STARTUP.md) — Runtime startup sequence
- [/governance/runtime/SESSION-OVERRIDE.md](/workspace/project/kde/governance/runtime/SESSION-OVERRIDE.md) — Session override behavior

### Engine Framework

- [interface.md](./interface.md) — Engine interface specification
- [README.md](./README.md) — Engine framework overview

---

**Do not edit this file to add new engines. Follow the engine transition process defined in the Engine Framework.**

**Do not change Runtime defaults here. Edit [/governance/runtime/defaults.yaml](/workspace/project/kde/governance/runtime/defaults.yaml) to change defaults.**
