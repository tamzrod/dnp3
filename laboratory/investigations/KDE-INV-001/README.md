---
id: KDE-INV-001
type: investigation
title: "Investigation Artifact Authority Model"
status: completed
authority: "KDE Runtime (DNP3 Library)"
created: "2026-07-25T10:46:04Z"
---
# Investigation Artifact Authority Model

**Investigation ID**: KDE-INV-001  
**Title**: Investigation Artifact Authority Model  
**Status**: COMPLETED  
**Date**: 2026-07-25  
**Author**: OpenHands Agent  
**Branch**: kde-bootstrap

---

## 1. Executive Summary

### 1.1 Overview

This investigation examined the authority, ownership, and provenance model for KDE-generated artifacts. The current model uses "Author" fields that identify execution agents (e.g., "OpenHands Agent"), which may conflict with KDE's runtime-governed philosophy where investigations are engineering activities under KDE authority rather than documents authored by individuals.

### 1.2 Key Findings

| Finding | Evidence |
|---------|----------|
| Current model uses "Author" field | IMP.md template: `**Author**: [Author]` |
| Execution agent is identified | KDE-INV-000 README: `**Author**: OpenHands Agent` |
| Runtime identity exists | `.kde/runtime/state.json`: `"project": "DNP3 Library"` |
| Bootstrap config tracks runtime | `.kde/bootstrap/config.yaml`: runtime configuration |
| Git history provides provenance | Existing investigations tracked via git commits |

### 1.3 Recommendation

**Adopt Model B: Authority + Execution Agent with Runtime Ownership**

Artifacts should identify:
1. **Authority**: KDE Runtime (project and runtime identity)
2. **Execution Agent**: Who executed the investigation
3. **Approver**: Human who approved the findings (if applicable)

### 1.4 Risk Assessment

**LOW RISK** - Recommendation is additive (adds fields without removing existing metadata).

---

## 2. Current Model Assessment

### 2.1 Current Metadata Structure

The current KDE artifact model uses the following metadata fields:

#### Investigation Header (from current templates)
```markdown
**Investigation ID**: KDE-INV-XXX  
**Title**: [Title]  
**Status**: DRAFT | IN_PROGRESS | COMPLETED  
**Date**: YYYY-MM-DD  
**Author**: [Author]  
```

#### IMP Template Header (from IMP.md)
```markdown
**ID**: PROJECT-IMP-XXX
**Title**: [Implementation Title]
**Status**: DRAFT | APPROVED | COMPLETED
**Date**: YYYY-MM-DD
**Author**: [Author]
**Human Reviewer**: [Reviewer]
```

### 2.2 Evidence from Existing Artifacts

| Artifact | Current Author | Evidence |
|----------|---------------|----------|
| KDE-INV-000 | OpenHands Agent | `**Author**: OpenHands Agent` |
| KDE-INV-001 (original) | OpenHands Agent | `**Author**: OpenHands Agent` |
| IMP.md template | N/A | `**Author**: [Author]` |

### 2.3 Problems with Current Model

1. **Agent Coupling**: Artifacts are tied to specific agent names
2. **No Runtime Identity**: No explicit link to KDE Runtime
3. **No Session Tracking**: No tracking of investigation sessions
4. **Inconsistent Fields**: IMP template has "Human Reviewer", investigations do not
5. **No Provenance Chain**: No explicit session or execution history

### 2.4 Current Provenance Mechanisms

The repository currently relies on:
- **Git History**: Commits track who made changes (via git config)
- **File Headers**: "Author" field in documents
- **Implicit Runtime**: No explicit runtime identity in artifacts

---

## 3. Candidate Metadata Models

### 3.1 Model A: Author-Centric

**Description**: Artifacts identify an "Author" field, typically the execution agent name.

**Example**:
```markdown
**Author**: OpenHands Agent
```

**Advantages**:
- Simple and familiar
- Matches conventional document authorship
- Easy to implement

**Disadvantages**:
- Ties artifacts to specific agents
- Does not reflect runtime-governed process
- Human reviewers not consistently represented
- No provenance beyond author name

**Score**: 3/9 criteria met

### 3.2 Model B: Authority + Execution Agent

**Description**: Artifacts identify both governance authority and execution agent.

**Example**:
```markdown
**Authority**: KDE Runtime (DNP3 Library)
**Execution Agent**: OpenHands Agent
**Human Approver**: [Name]
```

**Advantages**:
- Clear separation of governance vs. execution
- Runtime identity explicit
- Agent independence maintained
- Supports human oversight

**Disadvantages**:
- More verbose
- Requires template updates
- Backward compatibility considerations

**Score**: 7/9 criteria met

### 3.3 Model C: Runtime + Session + Agent

**Description**: Full provenance tracking with Runtime, Session ID, and Agent.

**Example**:
```markdown
**Runtime**: DNP3 Library KDE Runtime v1.0.0
**Session ID**: KDE-SESSION-20260725-001
**Execution Agent**: OpenHands Agent
**Git Commit**: abc1234
```

**Advantages**:
- Complete provenance chain
- Session-level tracking
- Full auditability
- Reproducibility support

**Disadvantages**:
- Complex to implement
- Requires session management system
- Verbose headers
- Session ID generation needed

**Score**: 6/9 criteria met

### 3.4 Model D: Runtime-Owned Artifact with Execution History

**Description**: Artifacts are owned by KDE Runtime. Execution history via git log.

**Example**:
```markdown
**Investigation ID**: KDE-INV-001
**Authority**: KDE Runtime
**Git History**: See git log
```

**Advantages**:
- Strong runtime ownership
- Git provides provenance
- Minimal header overhead
- Agent-agnostic

**Disadvantages**:
- Provenance requires git inspection
- Less intuitive for non-git users
- Git history may be squashed
- No explicit session tracking

**Score**: 5/9 criteria met

---

## 4. Comparative Analysis

### 4.1 Evaluation Matrix

| Criterion | Model A | Model B | Model C | Model D |
|-----------|---------|---------|---------|---------|
| Governance Clarity | ⚠️ | ✅ | ✅ | ✅ |
| Agent Independence | ❌ | ✅ | ✅ | ✅ |
| Multi-Agent Collaboration | ❌ | ✅ | ✅ | ⚠️ |
| Auditability | ⚠️ | ✅ | ✅ | ⚠️ |
| Provenance | ⚠️ | ✅ | ✅ | ⚠️ |
| Runtime Ownership | ❌ | ✅ | ✅ | ✅ |
| Simplicity | ✅ | ⚠️ | ❌ | ✅ |
| Long-term Maintainability | ⚠️ | ✅ | ⚠️ | ✅ |
| KDE Philosophy Alignment | ❌ | ✅ | ⚠️ | ⚠️ |

**Legend**: ✅ Strong | ⚠️ Partial | ❌ Weak

### 4.2 Detailed Analysis

#### Governance Clarity

| Model | Analysis |
|-------|----------|
| Model A | "Author" implies individual ownership, not runtime governance |
| Model B | "Authority" explicitly identifies KDE Runtime as governing body |
| Model C | Runtime identity present, but complex |
| Model D | "Authority" field, but provenance relies on git |

#### Agent Independence

| Model | Analysis |
|-------|----------|
| Model A | Directly ties artifacts to agent names |
| Model B | Separates runtime authority from execution agent |
| Model C | Explicit agent tracking, but separated from authority |
| Model D | No agent identity required in artifacts |

#### Multi-Agent Collaboration

| Model | Analysis |
|-------|----------|
| Model A | No mechanism for tracking multiple contributors |
| Model B | "Human Approver" field enables multi-agent workflows |
| Model C | Session tracking enables complex collaboration |
| Model D | Git history shows collaboration |

---

## 5. Recommended Authority Model

### 5.1 Proposed Model: Model B Enhanced

**Core Principle**: KDE artifacts are governed by the KDE Runtime and executed by agents. The model should clearly distinguish between:

1. **Governance Authority**: The KDE Runtime that governs the artifact
2. **Execution Responsibility**: The agent(s) who executed the work
3. **Human Oversight**: Human reviewers/approvers (when applicable)

### 5.2 Standard Investigation Metadata

```markdown
**Investigation ID**: KDE-INV-XXX
**Title**: [Title]
**Authority**: KDE Runtime ([Project Name])
**Status**: DRAFT | IN_PROGRESS | COMPLETED | APPROVED
**Date**: YYYY-MM-DD
**Execution Agent**: [Agent Name]
**Human Approver**: [Name] (if applicable)
**Branch**: [Git branch]
```

### 5.3 Standard Experiment Metadata

```markdown
**Experiment ID**: PROJECT-EXP-XXX
**Title**: [Title]
**Authority**: KDE Runtime ([Project Name])
**Status**: DRAFT | IN_PROGRESS | COMPLETED | VERIFIED
**Date**: YYYY-MM-DD
**Execution Agent**: [Agent Name]
**Human Reviewer**: [Name] (if applicable)
```

### 5.4 Standard Decision Metadata

```markdown
**Decision ID**: TDR-XXX
**Title**: [Title]
**Authority**: KDE Runtime ([Project Name])
**Status**: PROPOSED | APPROVED | REJECTED | SUPERSEDED
**Date**: YYYY-MM-DD
**Proposed By**: [Agent Name]
**Approved By**: [Human Name]
```

### 5.5 Standard IMP Metadata

```markdown
**ID**: PROJECT-IMP-XXX
**Title**: [Implementation Title]
**Authority**: KDE Runtime ([Project Name])
**Status**: DRAFT | APPROVED | COMPLETED | VERIFIED
**Date**: YYYY-MM-DD
**Author**: [Agent Name]
**Human Reviewer**: [Name]
**Source Investigation**: KDE-INV-XXX
**Source Decision**: TDR-XXX
```

---

## 6. Migration Impact

### 6.1 Template Changes Required

| Template | Current | Recommended |
|----------|---------|-------------|
| Investigation | `**Author**: [Author]` | `**Authority**: KDE Runtime ([Project])\n**Execution Agent**: [Agent]\n**Human Approver**: [Name]` |
| Experiment | (not defined) | Same as Investigation |
| Decision | (not defined) | `**Proposed By**: [Agent]\n**Approved By**: [Human]` |
| IMP | `**Author**: [Author]\n**Human Reviewer**: [Reviewer]` | `**Authority**: KDE Runtime ([Project])\n**Author**: [Agent]\n**Human Reviewer**: [Name]` |

### 6.2 Existing Artifact Migration

| Artifact | Action | Risk |
|----------|--------|------|
| KDE-INV-000 | Update metadata | LOW - Non-breaking |
| KDE-INV-001 (new) | Apply new model | N/A - Already follows |
| Future investigations | Use new model | N/A |

### 6.3 Governance Policy Updates

1. Update `.kde/governance/NAMING-CONVENTIONS.md` with authority model
2. Create new governance policy: `AUTHORITY-MODEL.md`
3. Update `.kde/templates/README.md` with metadata standards

---

## 7. Backward Compatibility Considerations

### 7.1 Non-Breaking Changes

The recommended model is **additive**:
- Existing "Author" fields can coexist with new fields
- No required removal of existing metadata
- Git history preserved

### 7.2 Gradual Migration Path

1. **Phase 1**: Update templates with new fields (additive)
2. **Phase 2**: Update new artifacts with new model
3. **Phase 3**: Update existing artifacts when modified
4. **Phase 4**: Deprecate old "Author" field (optional)

### 7.3 Template Backward Compatibility

```markdown
<!-- New template with backward compatibility -->
**Authority**: KDE Runtime ([Project Name])  <!-- NEW -->
**Author**: [Author]  <!-- DEPRECATED - Use Execution Agent -->
**Execution Agent**: [Agent]  <!-- NEW -->
**Human Approver**: [Name]  <!-- NEW -->
```

---

## 8. Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Template confusion | MEDIUM | LOW | Clear documentation of new fields |
| Migration effort | LOW | LOW | Gradual migration, no forced updates |
| Agent identity drift | LOW | LOW | Git history provides backup provenance |
| Human approver not tracked | MEDIUM | MEDIUM | Encourage but don't require |
| Runtime identity changes | LOW | MEDIUM | Runtime config is stable |

### 8.1 Risk Assessment Summary

**Overall Risk**: LOW - Model is additive and backward compatible.

---

## 9. Final Recommendation

### 9.1 Decision

**Adopt Model B: Authority + Execution Agent with Runtime Ownership**

### 9.2 Rationale

1. **Governance Clarity**: "Authority" field explicitly identifies KDE Runtime as governing body
2. **Agent Independence**: Runtime authority is separate from execution agent
3. **Multi-Agent Support**: "Human Approver" enables human oversight
4. **Auditability**: Git history + metadata provides complete audit trail
5. **Provenance**: Clear chain from Runtime → Agent → Artifact
6. **Runtime Ownership**: Artifacts governed by runtime, not individuals
7. **Simplicity**: Moderate complexity, well-scoped
8. **Long-term Maintainability**: Sustainable model
9. **KDE Philosophy**: Aligns with evidence-based, runtime-governed approach

### 9.3 Implementation Steps

1. **Update Templates**:
   - Add "Authority" field to all templates
   - Rename "Author" to "Execution Agent" where appropriate
   - Add "Human Approver" to investigations

2. **Update Governance**:
   - Add authority model to governance documentation
   - Document migration path

3. **Apply to New Artifacts**:
   - Use new model for KDE-INV-002 and beyond
   - Update existing artifacts when modified

4. **Monitor**:
   - Track adoption of new model
   - Adjust based on experience

---

## 10. Appendix: Evidence

### A. Current Template Evidence

```
.kde/templates/IMP.md:
**Author**: [Author]
**Human Reviewer**: [Reviewer]

laboratory/investigations/KDE-INV-000/README.md:
**Author**: OpenHands Agent
```

### B. Runtime Identity Evidence

```
.kde/runtime/state.json:
{
  "project": "DNP3 Library",
  "version": "1.0.0"
}

.kde/bootstrap/config.yaml:
runtime:
  name: "DNP3 Library KDE Runtime"
  project: "DNP3 Library"
```

### C. KDE Governance Principles Evidence

```
docs/kde/principles/ENGINEERING-PRINCIPLES.md:
### 7. Human as Approver
Humans review and authorize decisions.
```

---

*Investigation completed: 2026-07-25*  
*Author: OpenHands Agent*  
*Classification: GOVERNANCE MODEL RECOMMENDATION*  
*Recommendation: ADOPT MODEL B*
