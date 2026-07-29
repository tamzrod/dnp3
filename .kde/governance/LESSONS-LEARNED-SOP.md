# Lessons Learned Standard Operating Procedure

**Document ID**: SOP-LESSONS
**Title**: Lessons Learned Capture and Registry
**Version**: 1.0.0
**Status**: APPROVED (INV-EVOLUTION-001 REC-002)
**Effective Date**: 2026-07-24
**Authority**: Human Authority (Governance)
**Source**: INV-EVOLUTION-001 Section 5.1

---

## Purpose

This document establishes mandatory lessons-learned capture for all KDE experiments, addressing the finding that 85% of experiments (45/53) lack lessons-learned documentation. This SOP ensures KDE continuously improves through systematic learning capture.

---

## Scope

This SOP applies to:
- All experiments in `/laboratory/experiments/`
- All investigations in `/laboratory/investigations/`
- All new experiments and investigations created after 2026-07-24

---

## Mandatory Lessons-Learned Requirements

### REC-002 Section 1: Experiment Requirements

**All experiments MUST capture lessons-learned.**

| Experiment Duration | Lessons-Learned Required | Timeline |
|---------------------|---------------------------|----------|
| <1 day | RECOMMENDED | Before closure |
| 1-7 days | REQUIRED | Before closure |
| >7 days | REQUIRED + Mid-project | Ongoing + Before closure |

### REC-002 Section 2: Investigation Requirements

**Investigations lasting >1 day MUST capture lessons-learned.**

| Investigation Duration | Lessons-Learned Required |
|-----------------------|--------------------------|
| ≤1 day | OPTIONAL |
| >1 day | REQUIRED |

---

## Lessons-Learned Document Structure

### Required Sections

Every lessons-learned document MUST contain:

```markdown
# Lessons Learned: [LAB-XXX|INV-XXX]

**ID**: [LAB-XXX|INV-XXX]
**Duration**: [X days]
**Date**: YYYY-MM-DD
**Author**: [Name]

---

## What Worked

### Category 1: [e.g., Methodology]

| Item | Evidence | Impact |
|------|----------|--------|
| [Description] | [Link/Ref] | HIGH/MEDIUM/LOW |

### Category 2: [e.g., Process]

| Item | Evidence | Impact |
|------|----------|--------|
| [Description] | [Link/Ref] | HIGH/MEDIUM/LOW |

---

## What Didn't Work

### Category 1: [e.g., Methodology]

| Item | Evidence | Impact | Mitigation |
|------|----------|--------|------------|
| [Description] | [Link/Ref] | HIGH/MEDIUM/LOW | [Action taken] |

### Category 2: [e.g., Process]

| Item | Evidence | Impact | Mitigation |
|------|----------|--------|------------|
| [Description] | [Link/Ref] | HIGH/MEDIUM/LOW | [Action taken] |

---

## Future Improvements

| Improvement | Priority | Owner | Timeline |
|-------------|----------|-------|----------|
| [Description] | P0/P1/P2/P3 | [Role] | YYYY-MM-DD |

---

## Unexpected Findings

| Finding | Evidence | Value | Applied |
|---------|----------|-------|---------|
| [Description] | [Link/Ref] | HIGH/MEDIUM/LOW | YES/NO |

---

## Cross-Reference to Templates

For investigation-specific template, see:
- `/laboratory/templates/investigation-template.md`

For experiment-specific template, see:
- `/laboratory/templates/experiment-template.md`

---

## Lessons-Learned Registry

### REC-002 Section 3: Registry Maintenance

All lessons-learned documents MUST be registered in the central registry.

**Registry Location**: `/laboratory/lessons-registry.md`

### Registry Entry Format

```markdown
| ID | Duration | Date | What Worked | What Didn't | Cross-Applied |
|----|----------|------|-------------|-------------|---------------|
| LAB-XXX | 3 days | 2026-07-24 | X items | Y items | Z items |
```

### Registry Review Cadence

| Review | Frequency | Owner | Purpose |
|--------|-----------|-------|---------|
| Individual | Per closure | Investigator | Quality check |
| Quarterly | Every 3 months | Governance | Pattern identification |
| Annual | Yearly | Governance | System improvement |

---

## Quality Standards

### Lessons-Learned Quality Criteria

| Criterion | Description | Required |
|-----------|-------------|----------|
| **Specificity** | Concrete, not generic | YES |
| **Evidence** | Linked to specific artifacts | YES |
| **Impact** | Impact level assigned | YES |
| **Actionable** | Recommendations are actionable | YES |
| **Cross-Applied** | Noted if applied elsewhere | YES |

### Anti-Patterns to Avoid

| Anti-Pattern | Example | Avoid By |
|--------------|---------|----------|
| Generic statements | "Communication could be better" | Be specific |
| No evidence | "We learned X" without link | Cite artifacts |
| No impact | "X happened" | Assign HIGH/MEDIUM/LOW |
| Vague recommendations | "Improve things" | Be actionable |
| Isolated learning | No cross-reference | Link to similar |

---

## Enforcement Rules

### Rule 1: Closure Gate

**An experiment/investigation CANNOT transition to COMPLETE without:**

1. ✅ `lessons-learned.md` present (if duration >1 day)
2. ✅ Quality criteria met
3. ✅ Registry entry created
4. ✅ Cross-references documented

### Rule 2: Promotion Gate

**Experiments/investigations with missing lessons-learned CANNOT be promoted.**

Exception: Duration ≤1 day with documented rationale.

### Rule 3: Quality Review

**Lessons-learned undergo quality review before closure approval.**

| Review Aspect | Criteria |
|---------------|----------|
| Specificity | ≥3 specific items per category |
| Evidence | 100% items have links |
| Impact | All items have impact level |
| Actionability | All recommendations are specific |

---

## Pattern Identification

### REC-002 Section 4: Cross-Experiment Analysis

Quarterly, Governance SHALL identify patterns across lessons-learned.

**Pattern Categories:**

| Pattern Type | Detection Method | Action |
|--------------|-------------------|--------|
| Recurring success | Same lesson in 3+ experiments | Document as best practice |
| Recurring failure | Same lesson in 3+ experiments | Create improvement SOP |
| Cross-domain | Same lesson in different domains | Update templates |
| System-level | Affects multiple experiments | Governance intervention |

### Pattern Documentation

When patterns are identified:

```markdown
## Pattern: [Pattern Name]

**Identified**: YYYY-MM-DD
**Frequency**: X occurrences
**Evidence**: [Links to lessons-learned]

### Pattern Description

[Detailed description]

### Recommended Action

[Specific action with owner and timeline]

### Status

| Status | Date | Notes |
|--------|------|-------|
| IDENTIFIED | YYYY-MM-DD | Initial discovery |
| REVIEWED | YYYY-MM-DD | Governance review |
| ACTIONED | YYYY-MM-DD | Implementation complete |
```

---

## Integration with Other SOPs

### With INVESTIGATION-CLOSURE-SOP.md

Lessons-learned is REQUIRED for investigation closure:

```markdown
## Closure Checklist (from INVESTIGATION-CLOSURE-SOP.md)

☐ conclusion.md present
☐ lessons-learned.md present (if >1 day) ← ENFORCED HERE
☐ Human Review obtained
☐ Signatures complete
```

### With LABORATORY-SOP.md

Lessons-learned supports continuous improvement cycle:

```
Evidence → Analysis → Lessons Learned → Pattern → Improvement → Evidence
                              ↑
                        Captured here
```

---

## Metrics

### REC-002 Section 5: Capture Rate Tracking

| Metric | Target | Current | Trend |
|--------|--------|---------|-------|
| Experiment lessons capture rate | 100% (if >1 day) | 15% | Improving |
| Investigation lessons capture rate | 100% (if >1 day) | 15% | Improving |
| Cross-application rate | 50% | Unknown | TBD |
| Pattern identification rate | Quarterly | N/A | TBD |

---

## Related Documents

| Document | Relationship |
|----------|--------------|
| [INVESTIGATION-CLOSURE-SOP.md](./INVESTIGATION-CLOSURE-SOP.md) | Closure requirements |
| [LABORATORY-SOP.md](./LABORATORY-SOP.md) | Lifecycle procedures |
| [LABORATORY/registry.md](../laboratory/registry.md) | Experiment registry |
| [laboratory/templates/experiment-template.md](../laboratory/templates/experiment-template.md) | Experiment template |
| [laboratory/templates/investigation-template.md](../laboratory/templates/investigation-template.md) | Investigation template |

---

## Version History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2026-07-24 | INV-EVOLUTION-001 | Initial SOP (REC-002 implementation) |

---

**SOP Status**: APPROVED
**Authority**: Human Authority
**Enforcement**: MANDATORY (for >1 day)
**Source**: INV-EVOLUTION-001 REC-002
