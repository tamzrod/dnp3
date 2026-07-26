---
id: KDE-INV-055
type: investigation
title: "Engine Configuration Update (ECU): Automatic vs Manual Selection"
status: completed
authority: "KDE Runtime (DNP3 Library)"
created: "2026-07-26T04:35:00Z"
execution_agent: "OpenHands Agent"
---

# Investigation: Engine Configuration Update (ECU)

**Investigation ID**: KDE-INV-055  
**Title**: Engine Configuration Update (ECU): Automatic vs Manual Selection  
**Status**: COMPLETED  
**Date**: 2026-07-26  
**Authority**: KDE Runtime (DNP3 Library)  
**Execution Agent**: OpenHands Agent  

---

## 1. Executive Summary

### 1.1 Overview

This investigation catalogs all available KDE engines and seeds, analyzes options for automatic vs manual selection, and recommends an ECU (Engine Configuration Update) approach.

### 1.2 Key Findings

| Finding | Evidence |
|---------|----------|
| 5 main engines exist | Alpha, Beta, Gamma, Delta, Epsilon |
| 3 seeds exist | SEED-001, SEED-002, SEED-003 |
| Only Beta is default engine | current.md |
| SEED-003 is active but not configured as default | seeds/README.md |

### 1.3 Recommendation

**RECOMMEND**: Hybrid Approach - Default to SEED-003, then use ECU (runtime decision) for engine selection based on investigation type.

---

## 2. Engine Inventory

### 2.1 Main KDE Engines

| Engine ID | Codename | Status | Default | Purpose |
|-----------|----------|--------|---------|---------|
| KDE-ENGINE-001 | Alpha | Historical | NO | Pattern discovery (legacy) |
| KDE-ENGINE-002 | Beta | Active | **YES** | Contextual knowledge (default) |
| KDE-ENGINE-003 | Gamma | Active | NO | Causal discovery |
| KDE-ENGINE-004 | Delta | Active | NO | Bootstrap + Context |
| KDE-ENGINE-005 | Epsilon | Candidate | NO | Gap Analysis (not implemented) |

### 2.2 Specialized Engines (Domain-Specific)

| Engine ID | Status | Purpose | Use Case |
|-----------|--------|---------|----------|
| ADVERSARIAL-EVAL-001 | Active | Security testing | Protocol security audits |
| CONSENSUS-ADVERSARIAL-001 | Active | Consensus testing | Distributed system validation |
| CONSENSUS-SYNTH-001 | Active | Consensus synthesis | Novel consensus protocol discovery |
| PROTOCOL-SYNTH-001 | Active | Protocol synthesis | Novel protocol architecture generation |

### 2.3 Engine Assessment

**Why Specialized Engines Are Separated from Main KDE Engines:**

| Reason | Explanation |
|--------|-------------|
| **Different Purpose** | These engines are for *synthesizing* or *evaluating* protocols, not for *investigating* issues |
| **Different Context** | Main KDE engines answer "what is true?" and "why?" - specialized engines answer "what can we build?" |
| **Different Workflow** | Main engines run investigations - specialized engines run synthesis/evaluation pipelines |
| **Different Triggers** | Specialized engines require explicit invocation with specific parameters |

**They CAN be utilized**, but they serve a different role in the KDE ecosystem:

```
┌─────────────────────────────────────────────────────────────┐
│                    KDE Runtime                              │
├─────────────────────────────────────────────────────────────┤
│  Investigation Flow:                                       │
│  1. SEED-003 (Bootstrap) → Always runs first              │
│  2. ECU selects main engine (Beta/Gamma/Delta)           │
│  3. Investigation produces findings                        │
│                                                             │
│  Specialized Engines (Parallel Capabilities):               │
│  • ADVERSARIAL-EVAL → Security audit findings             │
│  • PROTOCOL-SYNTH → Generate new protocols                │
│  • CONSENSUS-SYNTH → Generate consensus algorithms         │
└─────────────────────────────────────────────────────────────┘
```

**When to Use Specialized Engines:**

| Engine | When to Use |
|--------|-------------|
| ADVERSARIAL-EVAL-001 | After investigation reveals protocol vulnerabilities |
| CONSENSUS-SYNTH-001 | When building distributed systems, need novel consensus |
| PROTOCOL-SYNTH-001 | When designing new secure communication protocols |
| CONSENSUS-ADVERSARIAL-001 | When testing existing consensus implementations |

**Updated Recommendation**: Categorize as **PARALLEL CAPABILITIES** rather than "not helping" - they're just not part of the main investigation workflow.

---

## 3. Seed Inventory

### 3.1 Available Seeds

| Seed ID | Codename | Status | Parent | Purpose |
|---------|----------|--------|--------|---------|
| SEED-001 | Genesis | FROZEN | — | Initial reasoning DNA |
| SEED-002 | Evolution | FROZEN | SEED-001 | Evolution principles |
| SEED-003 | Bootstrap | ACTIVE | SEED-002 | Bootstrap validation |

### 3.2 Seed Selection Analysis

**Current State:**
- SEED-003 (Bootstrap) is ACTIVE
- SEED-003 adds Bootstrap gates (B1, B2, B3)
- SEED-003 is NOT configured as default

**SEED-003 Core Principles:**
1. Bootstrap-First Verification (B1)
2. Pre-Existence Validation (B2)
3. Capability-Aware Commitment (B3)

---

## 4. Engine/Seed Selection Options

### 4.1 Option A: Fully Automatic Selection

**Description**: Runtime automatically selects engine and seed based on investigation type.

**Pros:**
- Zero configuration required
- Consistent behavior
- Fast initialization

**Cons:**
- May select wrong engine for complex investigations
- No human override mechanism
- Difficult to implement correctly

**Score**: 3/10

### 4.2 Option B: Manual Only (Current State)

**Description**: User must explicitly select engine and seed.

**Pros:**
- Full human control
- Clear selection
- Simple to implement

**Cons:**
- Requires user knowledge
- Easy to select wrong engine
- No guidance

**Score**: 5/10

### 4.3 Option C: SEED-003 Default + ECU Decision (RECOMMENDED)

**Description**: 
- Always default to SEED-003 (Bootstrap)
- Runtime uses ECU (Engine Configuration Update) logic to suggest/select engine based on:
  - Investigation type
  - Keywords detected
  - User override capability

**Pros:**
- Bootstrap gates always run (quality)
- ECU provides intelligent defaults
- Human can override when needed
- Clear, documented logic

**Cons:**
- Requires ECU logic implementation
- May need tuning

**Score**: 8/10

---

## 5. ECU Logic Design

### 5.1 Recommended ECU Decision Tree

```
1. Always load SEED-003 (Bootstrap)
2. Run Bootstrap Gates (B1, B2, B3)

3. ECU Engine Selection:
   ├── Investigation contains "why", "cause", "mechanism" 
   │   → Suggest Gamma (KDE-ENGINE-003)
   │
   ├── Investigation contains "bootstrap", "setup", "init"
   │   → Suggest Delta (KDE-ENGINE-004)
   │
   ├── Investigation contains "gap", "improvement", "weakness"
   │   → Suggest Epsilon (KDE-ENGINE-005) if implemented
   │
   ├── Default / No match
   │   → Use Beta (KDE-ENGINE-002)

4. Human Override:
   └── Allow explicit engine selection via session_override
```

### 5.2 ECU Keywords Mapping

| Keywords | Suggested Engine | Rationale |
|----------|------------------|-----------|
| why, cause, mechanism, what if | Gamma | Causal discovery |
| bootstrap, setup, init, reproduce | Delta | Bootstrap enforcement |
| gap, weakness, improvement, missing | Epsilon | Gap analysis |
| (default) | Beta | Contextual knowledge |

---

## 6. Parallel Capabilities Analysis

### 6.1 Why Specialized Engines Are Separate

These engines **CAN be utilized** - they serve a different purpose:

| Engine | Purpose | Workflow | Trigger |
|--------|---------|----------|---------|
| ADVERSARIAL-EVAL-001 | Security testing | Post-investigation | Explicit invocation |
| CONSENSUS-ADVERSARIAL-001 | Consensus testing | Validation | Explicit invocation |
| CONSENSUS-SYNTH-001 | Consensus synthesis | Build | Explicit invocation |
| PROTOCOL-SYNTH-001 | Protocol synthesis | Build | Explicit invocation |

**Key Difference**: Main engines answer "what is true?" / "why?" - Specialized engines answer "what can we build?" / "is it secure?"

### 6.2 Recommended Actions

1. **Create ENGINE-CATEGORIES.md** to clarify:
   - **Main KDE Engines** → Investigation workflow (Beta/Gamma/Delta)
   - **Parallel Capabilities** → Build/Evaluation workflow (synth/eval engines)

2. **Update current.md** to reference only main engines

3. **Document parallel capabilities** in separate section

---

## 7. Recommendations

### REC-001: Adopt Hybrid ECU Approach

**Decision**: ✅ APPROVED  
**Action**: Implement Option C - SEED-003 default + ECU engine selection

**Implementation**:
1. Configure runtime to always load SEED-003
2. Implement ECU keyword-based engine selection
3. Document override mechanism

### REC-002: Categorize Engines (PARALLEL CAPABILITIES)

**Decision**: ✅ APPROVED  
**Action**: Clarify engine categories - main vs parallel

**Implementation**:
1. Create `.kde/engines/CATEGORIES.md`
2. Update `current.md` to reference only main engines
3. Document parallel capabilities separately
4. **Note**: Specialized engines ARE useful - just different workflow

### REC-003: Deprecate Alpha

**Decision**: DEFERRED  
**Action**: Mark Alpha (KDE-ENGINE-001) as Deprecated

**Rationale**:
- Alpha is Historical status
- Beta supersedes Alpha
- Low priority - can be addressed later

---

## 8. Evidence

### 8.1 Files Analyzed

| File | Purpose |
|------|---------|
| `.kde/engines/current.md` | Engine status and defaults |
| `.kde/engines/alpha/specification.md` | Alpha engine details |
| `.kde/engines/beta/specification.md` | Beta engine details |
| `.kde/engines/gamma/specification.md` | Gamma engine details |
| `.kde/engines/delta/specification.md` | Delta engine details |
| `.kde/engines/epsilon/SPEC.md` | Epsilon engine details |
| `.kde/engines/adversarial-eval/manifest.yaml` | Adversarial engine |
| `.kde/engines/consensus-synth/manifest.yaml` | Consensus engine |
| `.kde/engines/protocol-synth/manifest.yaml` | Protocol engine |
| `.kde/seeds/README.md` | Seed inventory |
| `.kde/seeds/seed-003/README.md` | SEED-003 details |

### 8.2 Commands Run

```bash
# Runtime check
cat .kde/runtime/state.json

# Bootstrap gates
python3 .kde/bootstrap/gates.py --project-type go

# Engine listing
ls -la .kde/engines/
ls -la .kde/seeds/
```

---

## 9. Conclusion

**Answer to Question**: "Do we need automatic engine and seed selection or default to SEED-003 then ECU decide?"

**Answer**: **Default to SEED-003, then use ECU for engine selection.**

**Rationale**:
1. SEED-003 ensures bootstrap verification (quality)
2. ECU provides intelligent, keyword-based engine suggestions
3. Human override maintains flexibility
4. Simple to understand and implement

**Additional Finding**: 4 specialized engines (adversarial-eval, consensus-adversarial, consensus-synth, protocol-synth) are NOT helping general investigations and should be categorized separately.

---

*Investigation completed: 2026-07-26*  
*Execution Agent: OpenHands Agent*  
*Classification: ENGINE CONFIGURATION*
