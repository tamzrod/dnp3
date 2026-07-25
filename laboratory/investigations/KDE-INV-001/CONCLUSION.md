---
id: KDE-INV-001
type: investigation
title: "Investigation Artifact Authority Model"
status: completed
authority: "KDE Runtime (DNP3 Library)"
created: "2026-07-25T10:46:04Z"
---
# Investigation Artifact Authority Model - Conclusion

**Investigation ID**: KDE-INV-001  
**Status**: COMPLETED  
**Date**: 2026-07-25

---

## Summary

This investigation assessed the authority, ownership, and provenance model for KDE-generated artifacts and developed recommendations for an improved metadata model.

## Research Questions Addressed

| Question | Finding |
|----------|---------|
| Should KDE artifacts identify an "Author"? | Replace with "Authority" and "Execution Agent" |
| Is "Authority" more appropriate than "Author"? | Yes - reflects runtime governance |
| Should execution be separated from governance? | Yes - Model B provides separation |
| How to handle multi-agent investigations? | "Human Approver" field enables oversight |
| How to track session continuation? | Git history + metadata fields |
| How to represent human participation? | "Human Approver" / "Human Reviewer" fields |
| How to maintain provenance? | Runtime + Agent + Git history |
| What metadata for auditability? | Authority, Agent, Approver, Status, Date |
| What metadata for reproducibility? | Authority, Branch, Git Commit |
| Can model be agent-independent? | Yes - Runtime owns artifacts, agents execute |

## Classification

| Classification | Result |
|---------------|--------|
| **Current Model** | Author-Centric (Model A) |
| **Recommended Model** | Authority + Execution Agent (Model B) |
| **Risk Level** | LOW |
| **Migration Impact** | LOW (additive changes) |

## Candidate Model Comparison

| Criterion | Model A (Author) | Model B (Recommended) | Model C (Runtime+Session) | Model D (Runtime-Owned) |
|-----------|-----------------|---------------------|---------------------------|------------------------|
| Governance Clarity | ⚠️ | ✅ | ✅ | ✅ |
| Agent Independence | ❌ | ✅ | ✅ | ✅ |
| Multi-Agent Support | ❌ | ✅ | ✅ | ⚠️ |
| Auditability | ⚠️ | ✅ | ✅ | ⚠️ |
| Provenance | ⚠️ | ✅ | ✅ | ⚠️ |
| Runtime Ownership | ❌ | ✅ | ✅ | ✅ |
| Simplicity | ✅ | ⚠️ | ❌ | ✅ |
| **Total** | 3/9 | 7/9 | 6/9 | 5/9 |

## Recommended Metadata Fields

### Investigation Metadata
```markdown
**Investigation ID**: KDE-INV-XXX
**Title**: [Title]
**Authority**: KDE Runtime ([Project Name])
**Status**: DRAFT | IN_PROGRESS | COMPLETED | APPROVED
**Date**: YYYY-MM-DD
**Execution Agent**: [Agent Name]
**Human Approver**: [Name] (if applicable)
```

### Experiment Metadata
```markdown
**Experiment ID**: PROJECT-EXP-XXX
**Title**: [Title]
**Authority**: KDE Runtime ([Project Name])
**Status**: DRAFT | IN_PROGRESS | COMPLETED | VERIFIED
**Date**: YYYY-MM-DD
**Execution Agent**: [Agent Name]
**Human Reviewer**: [Name] (if applicable)
```

### Decision Metadata
```markdown
**Decision ID**: TDR-XXX
**Title**: [Title]
**Authority**: KDE Runtime ([Project Name])
**Status**: PROPOSED | APPROVED | REJECTED | SUPERSEDED
**Date**: YYYY-MM-DD
**Proposed By**: [Agent Name]
**Approved By**: [Human Name]
```

## Key Findings

### Problems with Current Model (Model A)

1. **Agent Coupling**: Artifacts tied to specific agent names
2. **No Runtime Identity**: No explicit link to KDE Runtime
3. **Inconsistent Fields**: Different templates have different fields
4. **Weak Provenance**: No explicit session or execution history

### Benefits of Recommended Model (Model B)

1. **Clear Governance**: "Authority" identifies KDE Runtime
2. **Agent Independence**: Execution separated from governance
3. **Multi-Agent Ready**: Supports human oversight
4. **Complete Provenance**: Git + metadata provides full audit trail
5. **Backward Compatible**: Additive changes, no forced removal

## Implementation Requirements

### Templates to Update

1. `.kde/templates/INV.md` (new)
2. `.kde/templates/EXP.md` (new)
3. `.kde/templates/TDR.md` (new)
4. `.kde/templates/IMP.md` (existing - update)

### Governance to Update

1. `.kde/governance/AUTHORITY-MODEL.md` (new policy)
2. `.kde/governance/NAMING-CONVENTIONS.md` (update)

## Risk Assessment

| Risk | Level | Mitigation |
|------|-------|------------|
| Template confusion | LOW | Clear documentation |
| Migration effort | LOW | Gradual migration |
| Agent identity drift | LOW | Git history backup |
| Human approver missing | MEDIUM | Encourage but don't require |

**Overall Risk**: LOW

## Final Decision

**RECOMMENDATION: ADOPT MODEL B**

The investigation recommends adopting Model B (Authority + Execution Agent) because:

1. It provides clear governance authority identification
2. It maintains agent independence
3. It supports multi-agent collaboration
4. It aligns with KDE philosophy of runtime-governed artifacts
5. It is backward compatible and low-risk
6. It provides complete provenance

## Next Steps

| Priority | Action | Owner |
|----------|--------|-------|
| HIGH | Update templates with new metadata fields | KDE Governance |
| HIGH | Create AUTHORITY-MODEL.md governance policy | KDE Governance |
| MEDIUM | Apply new model to new investigations | Agent |
| LOW | Update existing artifacts when modified | Agent |
| LOW | Deprecate old "Author" field (optional) | KDE Governance |

---

*Investigation completed: 2026-07-25*  
*Author: OpenHands Agent*  
*Recommendation: ADOPT MODEL B - Authority + Execution Agent with Runtime Ownership*  
*Status: AWAITING APPROVAL*
