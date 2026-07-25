# KDE-ENGINE-003 (Gamma) Changes

**Engine ID**: KDE-ENGINE-003
**Version**: 0.1.0
**Codename**: Gamma

---

## Version History

### v0.1.0 (2026-07-20) — Initial Release

#### Changes from Beta (KDE-ENGINE-002)

Gamma introduces causal discovery capabilities, extending Beta's contextual knowledge discovery with causal inference.

| Category | Change | Description |
|----------|--------|-------------|
| **Pipeline** | Extended | Added Stages 7 and 8 for causal discovery |
| **Methodology** | New | Causal discovery methodology |
| **Knowledge Model** | Extended | Added causal-specific fields |
| **Discovery Question** | Changed | From "When?" to "How?" |
| **Confidence** | Extended | Added causal confidence basis |

---

## Detailed Changes

### Pipeline Changes

#### Before (Beta 6-Stage Pipeline)

```
Evidence → Observation → Pattern → Validation → Context → Knowledge
```

#### After (Gamma 8-Stage Pipeline)

```
Evidence → Observation → Pattern → Validation → Context → Boundary → Causal Discovery → Causal Knowledge
```

#### New Modules

| Module | Purpose |
|--------|---------|
| **CausalDiscovery** | Identifies causal mechanisms and hypotheses |
| **CausalKnowledgeGeneration** | Assembles complete causal knowledge objects |

### Methodology Changes

#### Discovery Question Evolution

| Engine | Question | Focus |
|--------|----------|-------|
| Alpha | "Does X correlate with Y?" | Correlation |
| Beta | "Under what conditions does X correlate with Y?" | Context |
| Gamma | "How does X causally lead to Y?" | Causation |

#### New Principles Added

11. **Causation Requires Mechanism** — Correlation is not causation
12. **Confounders Must Be Addressed** — Third variables matter
13. **Intervention Changes Mechanisms** — do(X) implies causation
14. **Causal Models Are Hypotheses** — Testable causal diagrams
15. **Temporal Precedence Matters** — Cause precedes effect

### Knowledge Model Changes

#### New Fields Added

| Field Category | New Fields |
|----------------|------------|
| **Mechanism** | mechanism.description, mechanism.identified, mechanism.steps |
| **Causal Direction** | causal_direction.cause, causal_direction.effect, causal_direction.temporal_precedence |
| **Effect** | effect.type (ATE/ITE/LATE), effect.magnitude |
| **Confounders** | confounders[], confounders[].adjustment_method |
| **Intervention** | intervention.do_operation, intervention.predicted_outcome, intervention.side_effects |
| **Mediators** | mediators[] |
| **Moderators** | moderators[] |

### Confidence Assessment Changes

#### Before (Beta)

```yaml
confidence:
  value: <0-100>
  level: [high|medium|low]
  basis: [statistical|expert|heuristic]
```

#### After (Gamma)

```yaml
confidence:
  value: <0-100>
  level: [high|medium|low]
  basis: [statistical|causal_theory|mixed]
  factors:
    - factor: <string>
      contribution: [positive|negative|neutral]
      magnitude: [high|medium|low]
  limitations:
    - <limitation>
```

---

## Breaking Changes

### Knowledge Object Structure

Gamma knowledge objects are NOT compatible with Beta knowledge objects.

| Aspect | Beta | Gamma |
|--------|------|-------|
| ID prefix | KNOW-XXX | CAUSAL-XXX |
| Required fields | Knowledge fields | Causal fields |
| Confidence basis | statistical/expert/heuristic | statistical/causal_theory/mixed |

### Migration

To migrate from Beta to Gamma:

1. Re-run analysis under Gamma engine
2. Gamma will produce causal knowledge objects
3. Beta knowledge objects remain unchanged
4. Do NOT convert Beta objects to Gamma format (incompatible structures)

---

## Experimental Status

Gamma is released as **Experimental** for the following reasons:

1. **Methodology Validation**
   - Causal discovery methodology has not been empirically validated
   - Laboratory experiments required to assess methodology effectiveness

2. **Scope Limitations**
   - Limited to specified causal structures
   - Cannot handle all causal complexity

3. **Known Limitations**
   - Cannot prove causation from correlation alone
   - Requires assumptions about confounding
   - Causal models are hypotheses, not facts

### Validation Requirements

Before Gamma can be marked as Active:

| Requirement | Status |
|-------------|--------|
| Laboratory experiments conducted | Pending |
| Empirical validation complete | Pending |
| Methodology peer reviewed | Pending |
| Production readiness assessed | Pending |

---

## Compatibility Matrix

| Aspect | Alpha | Beta | Gamma |
|--------|-------|------|-------|
| **Interface** | v1.0 | v1.0 | v1.0 |
| **Evidence** | Compatible | Compatible | Compatible |
| **Observations** | Compatible | Compatible | Compatible |
| **Patterns** | Compatible | Compatible | Compatible |
| **Knowledge** | N/A | Partial | Full |
| **Causal Knowledge** | No | No | Yes |

---

## Migration Guide

### From Beta to Gamma

If you have an experiment under Beta and want to re-run under Gamma:

1. **Document Original Results**
   ```bash
   # Preserve Beta findings
   cp experiment/knowledge/ previous_beta_knowledge/
   ```

2. **Update Experiment Metadata**
   ```yaml
   Engine:
     ID: KDE-ENGINE-003
     Version: 0.1.0
     Codename: Gamma
   ```

3. **Re-run Under Gamma**
   ```bash
   # Execute experiment with Gamma engine
   laboratory run --engine=gamma experiment.md
   ```

4. **Compare Results**
   - Note new causal knowledge generated
   - Compare findings with Beta results
   - Document differences

5. **Update Knowledge**
   - Create Gamma causal knowledge objects
   - Preserve Beta objects for comparison
   - Document migration

### When NOT to Migrate

Do NOT migrate to Gamma if:

- [ ] You need strict compatibility with Beta
- [ ] Your analysis does not benefit from causal inference
- [ ] Causal assumptions cannot be justified
- [ ] You need to compare with historical Beta experiments

---

## Related Documents

- [specification.md](./specification.md) — Engine identity
- [methodology.md](./methodology.md) — Full methodology
- [pipeline.md](./pipeline.md) — Pipeline details
- [knowledge-model.md](./knowledge-model.md) — Knowledge structure
- [provenance.md](./provenance.md) — Engine lineage

---

**Document Status**: EXPERIMENTAL
