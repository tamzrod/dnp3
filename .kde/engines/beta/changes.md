# KDE-ENGINE-002 (Beta) Changes

**Engine ID**: KDE-ENGINE-002
**Version**: 0.1.0
**Codename**: Beta

---

## Version History

### v0.1.0 (2026-07-20) — Initial Release

**Status**: Active

#### Changes from Alpha (KDE-ENGINE-001)

##### New Features

1. **Context Detection Module**
   - Discovers conditions where patterns exist
   - Searches all available dimensions
   - Quantifies context strength

2. **Boundary Detection Module**
   - Discovers where patterns fail
   - Identifies contradictions and exceptions
   - Documents applicability limits

3. **Statistical Validation Enhancement**
   - Formal statistical testing required
   - Confidence explicitly calculated
   - Sample size thresholds enforced

4. **Knowledge Model Extension**
   - Knowledge objects include context
   - Knowledge objects include boundaries
   - Confidence is statistical, not subjective

##### Pipeline Changes

| Stage | Alpha | Beta |
|-------|-------|------|
| 1 | Evidence | Evidence |
| 2 | Observation | Observation Engine |
| 3 | Pattern | Pattern Detector |
| 4 | Validation | Statistical Validator |
| 5 | — | Context Detector |
| 6 | — | Boundary Detector |
| 7 | — | Knowledge Generator |

##### Question Transformation

| Engine | Question | Output |
|--------|----------|--------|
| Alpha | "Does X correlate with Y?" | Pattern: Yes/No |
| Beta | "Under what conditions does X correlate with Y?" | Knowledge with Context |

#### What Changed

1. **Discovery Question** — From "Does X?" to "When does X?"
2. **Pipeline** — Added 2 new modules (Context, Boundary)
3. **Confidence** — Now statistical, not implicit
4. **Knowledge** — Now includes context and boundaries
5. **Validation** — Now requires statistical significance

#### Previous State (Alpha)

Alpha (KDE-ENGINE-001) provides:
- Pattern discovery
- Basic validation
- Implicit confidence
- No context tracking
- No boundary detection

#### Rationale

Beta extends Alpha to address limitations identified through experimentation:

1. **Context Missing** — Alpha patterns lack applicability conditions
2. **Boundaries Unknown** — When patterns fail is undocumented
3. **Confidence Subjective** — Alpha confidence is not statistically derived
4. **Knowledge Incomplete** — Patterns without context are incomplete knowledge

#### Breaking Changes

**None** — Beta is a new engine, not a modification to Alpha.

Existing experiments running under Alpha continue unchanged.

#### Migration

**No migration required.**

- Alpha experiments remain under Alpha
- New experiments may use Beta
- Laboratory supports both engines
- Engine selection through configuration

---

## Version Conventions

### Version Number Format

Format: MAJOR.MINOR.PATCH

| Component | Description | When It Changes |
|-----------|-------------|-----------------|
| **MAJOR** | Breaking changes | Fundamental methodology changes |
| **MINOR** | Additive improvements | New modules or features |
| **PATCH** | Clarifications | Non-substantive improvements |

### Codename Conventions

| Codename | Meaning |
|----------|---------|
| **Alpha** | Initial development stage |
| **Beta** | Stable, actively developed |
| **Gamma** | Feature complete |
| **Release** | Production-ready |
| **Legacy** | Historical, no longer developed |

---

## Future Change Proposals

When changes are needed, document them here with:

1. **What changed**: Specific methodology modifications
2. **Why it changed**: Rationale and evidence
3. **Impact**: Effect on future experiments
4. **Migration**: How to transition from previous version

---

## Deprecation Policy

An engine becomes deprecated when:

1. A successor engine exists
2. No new experiments are expected under this engine
3. Historical experiments remain valid

Deprecation does not invalidate existing experiments.

---

## Archive

Previous versions are archived here when superseded.

| Version | Codename | Status | Superseded By | Date |
|---------|----------|--------|---------------|------|
| (none) | | | | |

---

**Document Status**: APPROVED
