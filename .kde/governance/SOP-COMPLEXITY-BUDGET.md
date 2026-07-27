# SOP: Governance Complexity Budget

**SOP ID**: SOP-COMPLEXITY-BUDGET
**Version**: 1.0.0
**Status**: APPROVED
**Authority**: INV-AUDIT-REVIEW-001 (Priority 7)
**Effective Date**: 2026-07-27
**Source**: Governance Complexity Prevention

---

## Purpose

This SOP establishes a **complexity budget** to prevent governance from growing beyond manageable scope. As identified in INV-AUDIT-001, governance complexity creates risk:

> "Governance complexity is a self-reinforcing cycle—each SOP creates new edge cases, which require new SOPs."

The complexity budget constrains growth by requiring removal when adding.

---

## The Problem

| Metric | Value | Trend |
|--------|-------|-------|
| LABORATORY-SOP.md size | 39KB+ | Growing |
| Number of SOPs | 15+ | Growing |
| Average SOP complexity | Increasing | Growing |

**Risk**: Eventually, governance becomes so complex that following it correctly is nearly impossible.

---

## Complexity Budget Definition

### Budget Components

| Component | Budget | Unit | Rationale |
|-----------|--------|------|-----------|
| **SOP Count** | 20 | Maximum | One page per SOP average |
| **SOP Size** | 20KB | Maximum per SOP | Readable in one sitting |
| **Total Governance** | 200KB | Maximum | 10-page governance document |
| **Cross-References** | 10 | Maximum per SOP | Beyond this, SOPs are tightly coupled |

### Current Usage

| Component | Current | Budget | Headroom |
|-----------|---------|--------|----------|
| SOP Count | 15 | 20 | 5 |
| Total Size | ~80KB | 200KB | 120KB |
| Average SOP Size | ~5KB | 20KB | 15KB |

---

## Budget Rules

### Rule 1: Addition Requires Removal

**Before adding a new SOP**, you MUST:

1. Identify the SOP to be removed or merged
2. Document why the new SOP is more important
3. Obtain approval for the trade-off
4. Execute removal in same commit as addition

### Rule 2: Size Limits

**SOPs exceeding size budget** MUST be split:

| Size | Action Required |
|------|-----------------|
| > 10KB | Consider splitting |
| > 15KB | Split recommended |
| > 20KB | Split required |

### Rule 3: Annual Simplification

**Every quarter**, governance SHALL:

1. Review all SOPs for obsolescence
2. Identify SOPs that can be removed or merged
3. Execute simplification
4. Report budget status

---

## Complexity Metrics

### How to Measure

```bash
# Count SOPs
find governance -name "SOP-*.md" | wc -l

# Measure total size
du -sh governance/

# Measure largest SOPs
du -sh governance/*.md | sort -rh | head -5

# Count cross-references
grep -r "see.*SOP\|refer.*SOP" governance/*.md | wc -l
```

### Current Metrics

| Metric | Value | Status |
|--------|-------|--------|
| SOP Count | 15 | ✅ Under budget |
| Total Size | ~80KB | ✅ Under budget |
| Largest SOP | LABORATORY-SOP.md (~40KB) | ⚠️ Over individual limit |
| Cross-References | ~20 | ✅ Under budget |

---

## SOP Review Checklist

Before adding a new SOP, complete this checklist:

- [ ] Is this SOP addressing an edge case or fundamental issue?
- [ ] Could existing SOPs be modified instead?
- [ ] What SOP will be removed to make room?
- [ ] Is this SOP likely to generate new edge cases?
- [ ] Have we measured current complexity budget?

---

## Exception Process

### When to Exceed Budget

Exceeding budget requires:

1. **Justification**: Document why the exception is necessary
2. **Time Limit**: Exception expires in 90 days
3. **Remediation Plan**: Document how budget will be restored
4. **Approval**: Human governance authority approval required

### Exception Template

```markdown
## Complexity Budget Exception Request

**Date**: YYYY-MM-DD
**Requested By**: [Name]
**SOP**: [SOP name]

### Justification
[Why budget must be exceeded]

### Remediation Plan
[How budget will be restored]

### Timeline
- Exception requested until: [Date + 90 days]
- Remediation deadline: [Date + 90 days]

### Approval
[ ] Approved
[ ] Rejected

**Approver**: [Name]
**Date**: YYYY-MM-DD
```

---

## Enforcement

### Automated Checking

The archive detector script (`.kde/scripts/archive-detector.py`) can be extended to track complexity:

```python
def check_complexity_budget():
    """Check if governance is within complexity budget."""
    sop_count = count_sops()
    total_size = measure_governance_size()
    
    if sop_count > SOP_COUNT_BUDGET:
        return False, f"SOP count {sop_count} exceeds budget {SOP_COUNT_BUDGET}"
    
    if total_size > TOTAL_SIZE_BUDGET:
        return False, f"Total size {total_size}KB exceeds budget {TOTAL_SIZE_BUDGET}KB"
    
    return True, "Within complexity budget"
```

### Manual Review

Governance review includes complexity assessment:
- Is governance within budget?
- Are exceptions justified?
- Is simplification on track?

---

## Simplification Opportunities

### Current Candidates for Removal/Merge

| SOP | Reason | Action |
|-----|--------|--------|
| SOP-ARCHIVE | Never followed | Merge with LABORATORY-SOP |
| SOP-CLOSURE | Overlap with investigation template | Simplify |

### Consolidation Plan

1. Merge SOP-ARCHIVE into LABORATORY-SOP (reduce by 1)
2. Simplify SOP-CLOSURE (reduce size)
3. Document exception for LABORATORY-SOP size (temporary)

---

## References

| Document | Relationship |
|----------|--------------|
| `governance/LABORATORY-SOP.md` | Main governance SOP |
| `governance/SOP-ARCHIVE.md` | Archive SOP (candidate for merge) |
| `.kde/scripts/archive-detector.py` | Archive detection script |

---

## Version History

| Version | Date | Changes | Authority |
|---------|------|---------|-----------|
| 1.0.0 | 2026-07-27 | Initial complexity budget SOP | INV-AUDIT-REVIEW-001 |

---

**SOP Status**: APPROVED
**Authority**: INV-AUDIT-REVIEW-001
**Compliance**: MANDATORY
**Review Date**: 2026-10-27 (Quarterly review)
