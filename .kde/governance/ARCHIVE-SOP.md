# Archive Standard Operating Procedure

**Document ID**: SOP-ARCHIVE
**Title**: Archive Management SOP
**Version**: 1.0.0
**Status**: APPROVED (INV-EVOLUTION-001 REC-005)
**Effective Date**: 2026-07-24
**Authority**: Human Authority (Governance)
**Source**: INV-EVOLUTION-001 Section 5.3

---

## Purpose

This document establishes the Archive Standard Operating Procedure for KDE investigations and experiments, addressing the finding that 0 experiments have been archived despite many being complete. This SOP ensures historical reference maintenance while keeping the active repository clean.

---

## Scope

This SOP applies to:
- All investigations in `/laboratory/investigations/`
- All experiments in `/laboratory/experiments/`
- Archive-eligible artifacts as defined by criteria below

---

## Archive Criteria

### REC-005 Section 1: Archive Eligibility

An investigation or experiment MAY be archived when ALL of the following criteria are met:

| Criterion | Description | Required |
|-----------|-------------|----------|
| **Completion** | Investigation/Experiment has COMPLETE status | YES |
| **Age** | >90 days since last update | YES |
| **Relevance** | Not actively referenced in current work | YES |
| **Replacement** | Superseded by newer investigation/experiment OR | YES |
| | No pending recommendations requiring action | |

### Archive Decision Matrix

| Completion | Age | Relevance | Replacement | Archive? |
|------------|-----|-----------|-------------|----------|
| COMPLETE | >90 days | Not referenced | YES | ✅ Archive |
| COMPLETE | >90 days | Not referenced | N/A | ✅ Archive |
| COMPLETE | >90 days | Referenced | YES | ⚠️ Review |
| COMPLETE | <90 days | Any | Any | ❌ No |
| IN_PROGRESS | Any | Any | Any | ❌ No |
| ACTIVE | Any | Any | Any | ❌ No |

---

## Archive Categories

### REC-005 Section 2: Archive Categories

Investigations and experiments are archived into appropriate categories:

| Category | Description | Directory |
|----------|-------------|-----------|
| **Historical** | Complete, aged, not referenced | `/laboratory/investigations/archive/` |
| **Superseded** | Replaced by newer work | `/laboratory/investigations/archive/superseded/` |
| **Reference** | Kept for historical reference | `/laboratory/investigations/archive/reference/` |
| **Incomplete** | Question-only, never progressed | `/laboratory/investigations/archive/incomplete/` |

---

## Archive Process

### REC-005 Section 3: Archive Procedure

```
┌─────────────────────────────────────────────────────────────────┐
│                       ARCHIVE DECISION FLOW                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. Identify candidate (age >90 days, COMPLETE status)          │
│           │                                                     │
│           ▼                                                     │
│  2. Check for active references                                  │
│           │                                                     │
│           ├──► Has active references?                            │
│           │        │                                            │
│           │        ├── YES → Keep active or mark REFERENCE      │
│           │        │                                            │
│           │        └── NO → Continue                             │
│           ▼                                                     │
│  3. Check for replacement                                        │
│           │                                                     │
│           ├──► Has newer version?                               │
│           │        │                                            │
│           │        ├── YES → Archive as SUPERSEDED              │
│           │        │                                            │
│           │        └── NO → Continue                             │
│           ▼                                                     │
│  4. Check completion status                                      │
│           │                                                     │
│           ├──► COMPLETE?                                        │
│           │        │                                            │
│           │        ├── YES → Archive as HISTORICAL              │
│           │        │                                            │
│           │        └── NO → Review for ARCHIVE-INCOMPLETE        │
│           │                                                     │
│           └──► Not COMPLETE?                                    │
│                    │                                            │
│                    └── Archive as INCOMPLETE (if >180 days)      │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Archive Steps

| Step | Action | Owner |
|------|--------|-------|
| 1 | Identify archive candidate | Governance |
| 2 | Verify completion status | Investigator |
| 3 | Check for active references | Governance |
| 4 | Determine archive category | Governance |
| 5 | Move files to archive directory | Governance |
| 6 | Update registry (if applicable) | Governance |
| 7 | Document archive decision | Governance |

---

## Archive Directory Structure

### Standard Archive Structure

```
laboratory/
├── investigations/
│   ├── [active investigations]
│   └── archive/
│       ├── HISTORICAL/
│       │   ├── INV-XXX/
│       │   │   ├── investigation.md
│       │   │   ├── conclusion.md
│       │   │   └── archive-note.md
│       │   └── ...
│       ├── SUPERSEDED/
│       │   ├── INV-XXX/
│       │   └── ...
│       ├── REFERENCE/
│       │   ├── INV-XXX/
│       │   └── ...
│       └── INCOMPLETE/
│           ├── INV-XXX/
│           └── ...
└── experiments/
    ├── [active experiments]
    └── archive/
        ├── HISTORICAL/
        ├── SUPERSEDED/
        └── INCOMPLETE/
```

### Archive Note Format

Each archived artifact MUST include an archive note:

```markdown
# Archive Note: [INV-XXX|LAB-XXX]

**Original ID**: [INV-XXX|LAB-XXX]
**Archive Date**: YYYY-MM-DD
**Archive Category**: [HISTORICAL|SUPERSEDED|REFERENCE|INCOMPLETE]
**Original Status**: [COMPLETE|INCOMPLETE|...]
**Archived By**: [Name]
**Archive Reason**: [Brief explanation]

## Why Archived

[Explanation of why this was archived]

## Historical Reference

[How to reference this archived work]

## Unarchive Criteria

[Conditions under which this could be restored]
```

---

## Archive Metrics

### REC-005 Section 4: Archive Tracking

| Metric | Target | Current |
|--------|--------|---------|
| Investigations archived | All eligible | 0 |
| Experiments archived | All eligible | 0 |
| Archive review frequency | Quarterly | N/A |
| Archive compliance | 100% eligible archived | 0% |

### Quarterly Archive Review

| Quarter | Reviews Conducted | Items Archived |
|---------|------------------|---------------|
| 2026-Q3 | (Initial) | TBD |

---

## Reference Archives

### Reference Category Criteria

Investigations/experiments may be kept in REFERENCE archive when:

| Criterion | Description |
|-----------|-------------|
| **Historical Value** | Contains patterns or learnings still relevant |
| **Template Value** | Useful as reference for future work |
| **Educational Value** | Demonstrates methodology or process |

Reference archives are:
- NOT actively maintained
- NOT updated with new information
- KEPT for lookup purposes only

---

## Incomplete Archives

### Incomplete Category Criteria

Investigations/experiments are archived as INCOMPLETE when:

| Criterion | Description |
|-----------|-------------|
| **Never Progressed** | Question-only, no investigation.md |
| **Abandoned** | Started but never completed |
| **Duration >180 days** | Old and incomplete |

Incomplete archives document:
- Why the work was never completed
- What would be needed to complete it
- Whether completion is still valuable

---

## Legacy Audit

### REC-005 Section 5: Initial Archive Audit

Per INV-EVOLUTION-001 findings, 46/51 investigations lack closure.

**Initial Audit Required:**

| Category | Count | Action |
|----------|-------|--------|
| Question-only (never started) | ~30 | Archive as INCOMPLETE |
| Started, incomplete | ~16 | Review for closure or archive |
| Complete | ~5 | Archive as HISTORICAL |

### Recommended Actions

1. **Immediate (30 days)**:
   - Archive all question-only investigations as INCOMPLETE
   - Archive INV-EVOLUTION-001 findings (46 incomplete)

2. **Short-term (90 days)**:
   - Complete or archive remaining incomplete investigations
   - Update registry with archive references

---

## Enforcement Rules

### Rule 1: Quarterly Review

Governance SHALL conduct quarterly archive reviews to:
1. Identify new eligible archives
2. Verify archived items still meet criteria
3. Process unarchive requests if any
4. Update archive metrics

### Rule 2: Reference Maintenance

Archived REFERENCE items SHALL:
1. Be clearly labeled as historical
2. Include date of archival
3. Document why kept for reference
4. NOT be modified after archival

### Rule 3: Unarchive Process

Archived items MAY be unarchived if:
1. New investigation requires the work
2. Original work has significant value
3. Human Authority approves unarchive

Unarchive Process:
```markdown
1. Request unarchive with justification
2. Governance review
3. Human Authority approval
4. Move from archive to active
5. Update registry
6. Document unarchive reason
```

---

## Related Documents

| Document | Relationship |
|----------|--------------|
| [INVESTIGATION-CLOSURE-SOP.md](./INVESTIGATION-CLOSURE-SOP.md) | Closure requirements |
| [LESSONS-LEARNED-SOP.md](./LESSONS-LEARNED-SOP.md) | Lessons capture |
| [LABORATORY/registry.md](../laboratory/registry.md) | Investigation registry |
| [laboratory/investigations/archive/](../laboratory/investigations/archive/) | Archive directory |

---

## Version History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2026-07-24 | INV-EVOLUTION-001 | Initial SOP (REC-005 implementation) |

---

**SOP Status**: APPROVED
**Authority**: Human Authority
**Enforcement**: MANDATORY
**Review Cadence**: Quarterly
**Source**: INV-EVOLUTION-001 REC-005
