# Investigation Specification: Investigation Artifact Authority Model

**Investigation ID**: KDE-INV-001  
**Title**: Investigation Artifact Authority Model  
**Status**: IN_PROGRESS  
**Date**: 2026-07-25  
**Author**: OpenHands Agent  

---

## 1. Background

The DNP3 Library uses the KDE (Knowledge Discovery Engine) Runtime for evidence-based engineering governance. KDE investigations are executed by AI agents under the authority of the KDE Runtime.

The current artifact model identifies the execution agent (e.g., "OpenHands Agent") as the "Author". This investigation examines whether this model is appropriate for KDE governance, where investigations are runtime-governed engineering activities rather than documents authored by an individual.

## 2. Objective

Determine the most appropriate authority, ownership, and provenance model for KDE-generated artifacts that:

1. Maintains governance clarity regardless of execution agent
2. Supports multi-agent collaboration
3. Enables proper auditability and provenance
4. Preserves runtime ownership
5. Remains agent-independent

## 3. Scope

### 3.1 Artifacts Under Review

| Artifact Type | Template Location | Current "Author" Field |
|--------------|------------------|----------------------|
| Investigation | `.kde/templates/` | Execution agent name |
| Experiment | `.kde/templates/` | Execution agent name |
| Decision (TDR) | `.kde/templates/` | Execution agent name |
| Implementation (IMP) | `.kde/templates/IMP.md` | Author field |
| Review | (Future) | TBD |

### 3.2 Related KDE Components

- `.kde/runtime/state.json` - Runtime state tracking
- `.kde/bootstrap/config.yaml` - Runtime configuration
- `.kde/governance/` - Governance policies
- `laboratory/investigations/` - Investigation storage

## 4. Research Questions

1. Should KDE artifacts identify an "Author"?
2. Is "Authority" a more appropriate governance concept than "Author"?
3. Should execution responsibility be separated from governance authority?
4. How should multi-agent investigations be represented?
5. How should investigation continuation across sessions be recorded?
6. How should human participation be represented?
7. How should artifact provenance be maintained?
8. What metadata is required for auditability?
9. What metadata is required for reproducibility?
10. Can the authority model remain agent-independent?

## 5. Candidate Models

### Model A: Author-Centric
- Artifacts identify an "Author" (typically the execution agent)
- Simple and familiar
- Problem: Ties artifacts to specific agents, not runtime

### Model B: Authority + Execution Agent
- Artifacts identify both "Authority" (KDE Runtime) and "Execution Agent" (who ran the investigation)
- Provides separation of concerns
- Problem: May be verbose

### Model C: Runtime + Session + Agent
- Artifacts reference Runtime, Session ID, and Execution Agent
- Comprehensive provenance
- Problem: Complex, requires session tracking

### Model D: Runtime-Owned Artifact with Execution History
- Artifacts are owned by the KDE Runtime
- Execution history is recorded separately
- Provenance via git history
- Problem: Less intuitive

## 6. Evaluation Criteria

| Criterion | Description |
|-----------|-------------|
| Governance Clarity | Clear understanding of who owns and authorizes artifacts |
| Agent Independence | Model remains stable across different execution agents |
| Multi-Agent Collaboration | Supports multiple agents working on same investigation |
| Auditability | Complete audit trail for decisions |
| Provenance | Clear chain of custody |
| Runtime Ownership | Artifacts owned by runtime, not agents |
| Simplicity | Easy to understand and implement |
| Long-term Maintainability | Sustainable over project lifetime |
| KDE Philosophy | Aligns with evidence-based, runtime-governed approach |

## 7. Deliverables

1. Current Model Assessment
2. Candidate Metadata Models
3. Comparative Analysis
4. Recommended Authority Model
5. Proposed Standard Investigation Metadata
6. Migration Impact
7. Backward Compatibility Considerations
8. Risks
9. Final Recommendation

## 8. Constraints

- Do not modify the KDE Runtime
- Do not modify existing investigations
- Produce recommendations only
- Await approval before changing KDE governance or artifact templates

---

*Investigation initiated: 2026-07-25*
