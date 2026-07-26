---
id: KDE-INV-055
type: investigation
title: "Engine Configuration Update (ECU): Automatic vs Manual Selection"
status: completed
authority: "KDE Runtime (DNP3 Library)"
created: "2026-07-26T04:35:00Z"
execution_agent: "OpenHands Agent"
---

# Investigation Conclusion: Engine Configuration Update (ECU)

**Investigation ID**: KDE-INV-055  
**Title**: Engine Configuration Update (ECU): Automatic vs Manual Selection  
**Status**: COMPLETED  
**Date**: 2026-07-26  
**Authority**: KDE Runtime (DNP3 Library)  
**Execution Agent**: OpenHands Agent  

---

## 1. Summary

### 1.1 Investigation Outcome

**✅ COMPLETED** - Engine configuration analyzed, recommendations made.

### 1.2 Key Metrics

| Metric | Before | After |
|--------|--------|-------|
| Engine clarity | Mixed (main + specialized) | To be categorized |
| Seed default | SEED-003 active, not configured | REC: Always load SEED-003 |
| Engine selection | Manual only | REC: ECU-based suggestion |

---

## 2. Recommendations

### REC-001: Adopt Hybrid ECU Approach

**Decision**: APPROVED  
**Action**: Implement Option C - SEED-003 default + ECU engine selection

**Implementation Steps**:
1. Configure runtime to always load SEED-003 at startup
2. Implement ECU keyword-based engine selection logic
3. Document override mechanism for explicit selection

### REC-002: Categorize Engines

**Decision**: APPROVED  
**Action**: Separate main KDE engines from specialized engines

**Implementation Steps**:
1. Create `.kde/engines/CATEGORIES.md`
2. Move specialized engines to `.kde/engines/specialized/`
3. Update `current.md` to reference only main engines

### REC-003: Deprecate Alpha (Optional)

**Decision**: DEFERRED  
**Action**: Mark Alpha (KDE-ENGINE-001) as Deprecated

**Rationale**: Low priority, Alpha is already Historical

---

## 3. Decision Record

| Decision | Rationale |
|----------|-----------|
| SEED-003 as default | Ensures bootstrap verification always runs |
| ECU for engine selection | Provides intelligent defaults without full automation |
| Categorize engines | Clarifies purpose, reduces confusion |
| Human override | Maintains flexibility for complex investigations |
| Specialized = Parallel | These engines ARE useful, just different workflow |

---

## 4. Evidence

### 4.1 Engine Inventory Complete

| Engine ID | Codename | Status | Default | Use Case |
|-----------|----------|--------|---------|----------|
| KDE-ENGINE-001 | Alpha | Historical | NO | Legacy (use Beta instead) |
| KDE-ENGINE-002 | Beta | Active | **YES** | Contextual knowledge |
| KDE-ENGINE-003 | Gamma | Active | NO | Causal discovery |
| KDE-ENGINE-004 | Delta | Active | NO | Bootstrap + Context |
| KDE-ENGINE-005 | Epsilon | Candidate | NO | Gap Analysis |

### 4.2 Seed Inventory Complete

| Seed ID | Codename | Status | Purpose |
|---------|----------|--------|---------|
| SEED-001 | Genesis | FROZEN | Initial reasoning |
| SEED-002 | Evolution | FROZEN | Evolution principles |
| SEED-003 | Bootstrap | ACTIVE | Bootstrap validation |

### 4.3 Specialized Engines (PARALLEL CAPABILITIES)

| Engine | Purpose | When to Use |
|--------|---------|-------------|
| ADVERSARIAL-EVAL-001 | Security testing | After protocol vulnerabilities found |
| CONSENSUS-ADVERSARIAL-001 | Consensus testing | Testing consensus implementations |
| CONSENSUS-SYNTH-001 | Consensus synthesis | Building distributed systems |
| PROTOCOL-SYNTH-001 | Protocol synthesis | Designing new protocols |

**Note**: These engines CAN be utilized - they're just **PARALLEL CAPABILITIES** not part of the main investigation workflow.

---

## 5. ECU Logic Summary

### 5.1 Decision Tree

```
Session Start:
  → Load SEED-003 (Bootstrap)
  → Run Bootstrap Gates (B1, B2, B3)

Engine Selection (ECU):
  → Keywords: "why/cause/mechanism" → Gamma
  → Keywords: "bootstrap/setup/init" → Delta
  → Keywords: "gap/weakness/improvement" → Epsilon (if implemented)
  → Default → Beta

Human Override:
  → session_override: engine_id
```

### 5.2 ECU Keywords Mapping

| Keywords | Engine | Rationale |
|----------|--------|-----------|
| why, cause, mechanism, intervention, root cause | Gamma | Causal discovery |
| bootstrap, setup, reproduce, initialization | Delta | Bootstrap enforcement |
| gap, weakness, improvement, missing, opportunity | Epsilon | Gap analysis |
| (default) | Beta | General contextual knowledge |

---

## 6. Next Steps

### 6.1 Immediate Actions

| Action | Owner | Status |
|--------|-------|--------|
| Create ENGINE-CATEGORIES.md | Runtime | TODO |
| Move specialized engines | Runtime | TODO |
| Update current.md | Runtime | TODO |

### 6.2 Future Actions

| Action | Priority | Status |
|--------|----------|--------|
| Implement ECU selection logic | HIGH | PENDING |
| Document override mechanism | MEDIUM | PENDING |
| Deprecate Alpha | LOW | DEFERRED |

---

## 7. Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| ECU keywords insufficient | MEDIUM | LOW | Human override always available |
| SEED-003 overhead | LOW | LOW | Minimal, runs once at startup |
| Engine categorization confusion | LOW | LOW | Clear documentation |

**Overall Risk**: LOW  
**Risk Assessment**: Acceptable for implementation

---

## 8. Approval

| Role | Name | Date | Signature |
|------|------|------|-----------|
| Execution Agent | OpenHands Agent | 2026-07-26 | OpenHands |
| Human Approver | [APPROVED] | 2026-07-26 | ✅ |

**Approved Recommendations**:
- ✅ REC-001: Adopt Hybrid ECU Approach
- ✅ REC-002: Categorize Engines (Parallel Capabilities)
- ⏸️ REC-003: Deprecate Alpha (DEFERRED)

---

*Investigation concluded: 2026-07-26*  
*Classification: ENGINE CONFIGURATION*  
*Status: COMPLETED - APPROVED*
