# Investigation Closure Standard Operating Procedure

**Document ID**: SOP-INV-CLOSURE
**Title**: Investigation Closure and Enforcement
**Version**: 1.0.0
**Status**: APPROVED (INV-EVOLUTION-001 REC-001)
**Effective Date**: 2026-07-24
**Authority**: Human Authority (Governance)
**Source**: INV-EVOLUTION-001 Section 5.1

---

## Purpose

This document establishes mandatory closure requirements for all KDE investigations, addressing the finding that 90% of investigations (46/51) lack proper closure. This SOP enforces investigation completion standards established in [LABORATORY-SOP.md](./LABORATORY-SOP.md).

---

## Scope

This SOP applies to:
- All investigations in `/laboratory/investigations/`
- All new investigations created after 2026-07-24
- Investigation promotions to knowledge or research recommendations

---

## Mandatory Closure Requirements

### REC-001 Section 1: Required Documents

All investigations MUST contain the following documents for COMPLETE status:

| Document | Required | Rationale |
|----------|----------|-----------|
| `investigation.md` | YES | Research question and scope |
| `conclusion.md` | YES | Final findings and recommendations |
| `lessons-learned.md` | YES (>1 day) | Captured learning |

### REC-001 Section 2: Investigation Status Progression

```
┌─────────────────────────────────────────────────────────────────┐
│                    INVESTIGATION STATUS PROGRESSION               │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ACTIVE ────────────────────────────────────────────────► CLOSED │
│    │                                                          │
│    ├── Research Question Defined                               │
│    ├── Scope Established                                      │
│    └── Planning Complete                                       │
│           │                                                   │
│           ▼                                                   │
│    IN_PROGRESS                                               │
│           │                                                   │
│           ├── Evidence Collection                              │
│           ├── Analysis Conducted                               │
│           └── Conclusions Drawn                               │
│           │                                                   │
│           ▼                                                   │
│    PENDING_CLOSURE ─────────────────────────────────────────────►│
│           │                                                   │
│           ├── conclusion.md present                            │
│           ├── lessons-learned.md present (if >1 day)           │
│           ├── Human Review Approved                           │
│           └── Closure Document Signed                          │
│           │                                                   │
│           ▼                                                   │
│    COMPLETE                                                   │
│           │                                                   │
│           └── Investigation Archived (if eligible)              │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Enforcement Rules

### Rule 1: Closure Gate

**An investigation CANNOT transition to COMPLETE without:**

1. ✅ `investigation.md` with status: ACTIVE → COMPLETE
2. ✅ `conclusion.md` present with:
   - Final conclusion statement
   - Evidence summary
   - Recommendations (if any)
3. ✅ `lessons-learned.md` present if investigation duration > 1 day

### Rule 2: Promotion Gate

**An investigation CANNOT be promoted without COMPLETE status.**

Promotions include:
- Knowledge promotion (to `/knowledge/`)
- Research recommendation (to Governance)
- Pattern documentation (to repository)

### Rule 3: Closure Review

**All investigations require Human Review before closure.**

| Review Type | Required For | Authority |
|-------------|--------------|-----------|
| Standard Review | All investigations | Human Authority |
| Enhanced Review | P0/P1 recommendations | Governance Board |

---

## Closure Document Template

### Required: conclusion.md

```markdown
# Conclusion: INV-XXX

**Investigation**: INV-XXX
**Date**: YYYY-MM-DD
**Status**: COMPLETE
**Confidence**: HIGH|MEDIUM|LOW
**Human Reviewer**: [Name]
**Review Date**: YYYY-MM-DD

---

## Final Conclusion

[State the conclusion based on evidence]

## Evidence Summary

| Evidence | Source | Support Level |
|----------|--------|---------------|
| [Evidence 1] | [Source] | HIGH |
| [Evidence 2] | [Source] | MEDIUM |

## Recommendations

| ID | Recommendation | Priority | Owner |
|----|-----------------|----------|-------|
| REC-XXX | [Text] | P0/P1/P2/P3 | [Role] |

## Signatures

| Role | Name | Date |
|------|------|------|
| Investigator | [Name] | YYYY-MM-DD |
| Human Reviewer | [Name] | YYYY-MM-DD |

---

**Closure Status**: APPROVED
**Investigation Status**: COMPLETE
```

### Required (if >1 day): lessons-learned.md

```markdown
# Lessons Learned: INV-XXX

**Investigation**: INV-XXX
**Duration**: [X days]
**Date**: YYYY-MM-DD

---

## What Worked

-

## What Didn't Work

-

## Future Improvements

-

## Unexpected Findings

-

## Applied to Repository

| Lesson | Applied Where | Evidence |
|--------|--------------|----------|
| [Lesson] | [Location] | [Evidence] |

---

**Lessons Captured**: YES
**Applied**: [YES/PENDING/NO]
```

---

## Closure Checklist

Before marking an investigation COMPLETE, verify:

| Check | Item | Verified |
|-------|------|----------|
| 1 | `investigation.md` has status: ACTIVE | ☐ |
| 2 | Research question answered | ☐ |
| 3 | All planned experiments executed | ☐ |
| 4 | `conclusion.md` present | ☐ |
| 5 | `lessons-learned.md` present (if >1 day) | ☐ |
| 6 | Evidence documented | ☐ |
| 7 | Human Review obtained | ☐ |
| 8 | Signatures complete | ☐ |

---

## Legacy Investigation Audit

### REC-001 Section 3: Incomplete Investigations

Per INV-EVOLUTION-001 findings, 46/51 investigations are incomplete.

**Audit Required:**

| Status | Count | Action |
|--------|-------|--------|
| Question-only (no investigation.md) | ~30 | Archive or complete |
| Missing conclusion.md | ~46 | Add or archive |
| Missing lessons-learned | ~45 | Add (if >1 day) or document |

### Recommended Actions for Legacy Investigations

1. **Complete**: Add required documents within 30 days
2. **Archive**: Move to `/laboratory/investigations/archive/` if no longer relevant
3. **Deprecate**: Mark as DEPRECATED if superseded

---

## Archive Criteria

Investigations MAY be archived when:

| Criterion | Description |
|-----------|-------------|
| **Age** | >90 days since last update |
| **Completion** | COMPLETE status achieved |
| **Relevance** | Not actively referenced |
| **Replacement** | Superseded by newer investigation |

### Archive Process

1. Mark investigation.md status: COMPLETE
2. Add to `/laboratory/investigations/archive/`
3. Update registry.md with archive reference
4. Maintain historical links

---

## Enforcement Metrics

Track closure rate to measure SOP effectiveness:

| Metric | Target | Current |
|--------|--------|---------|
| Investigation completion rate | 100% | 10% |
| Lessons-learned capture rate | 100% (if >1 day) | 15% |
| Average time to closure | <14 days | N/A |

---

## Related Documents

| Document | Relationship |
|----------|--------------|
| [LABORATORY-SOP.md](./LABORATORY-SOP.md) | Investigation lifecycle |
| [ARTIFACT-PROTECTION.md](./ARTIFACT-PROTECTION.md) | Artifact protection levels |
| [LABORATORY/registry.md](../laboratory/registry.md) | Investigation registry |
| [/laboratory/investigations/](../laboratory/investigations/) | Investigation directory |

---

## Version History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2026-07-24 | INV-EVOLUTION-001 | Initial SOP (REC-001 implementation) |

---

**SOP Status**: APPROVED
**Authority**: Human Authority
**Enforcement**: MANDATORY
**Source**: INV-EVOLUTION-001 REC-001
