---
id: KDE-INV-055
type: investigation
title: "Engine Configuration Update (ECU): Automatic vs Manual Selection"
status: completed
authority: "KDE Runtime (DNP3 Library)"
created: "2026-07-26T04:35:00Z"
execution_agent: "OpenHands Agent"
---

# Investigation Specification: Engine Configuration Update (ECU)

**Investigation ID**: KDE-INV-055  
**Title**: Engine Configuration Update (ECU): Automatic vs Manual Selection  
**Status**: COMPLETED  
**Date**: 2026-07-26  
**Authority**: KDE Runtime (DNP3 Library)  
**Execution Agent**: OpenHands Agent  

---

## 1. Problem Statement

### 1.1 Issue Description

The KDE Runtime has multiple engines (Alpha, Beta, Gamma, Delta, Epsilon) and seeds (SEED-001, SEED-002, SEED-003). Currently:
- Only Beta is set as the default engine
- Seed selection is not clearly defined
- No automatic engine/seed selection mechanism exists

### 1.2 Question

**Should the runtime automatically select engines/seeds, or should we default to SEED-003 then let ECU (Engine Configuration Update) decide?**

### 1.3 Current State

| Component | Current Default | Configuration |
|-----------|-----------------|---------------|
| Engine | Beta (KDE-ENGINE-002) | Manual only |
| Seed | SEED-003 (Bootstrap) | Active but not default |

---

## 2. Investigation Plan

### 2.1 Objectives

1. Catalog all available engines and their status
2. Catalog all available seeds and their status
3. Analyze options for engine/seed selection
4. Make recommendation on ECU approach

### 2.2 Success Criteria

| Criterion | Target |
|-----------|--------|
| All engines cataloged | Complete list with status |
| All seeds cataloged | Complete list with status |
| Options analyzed | At least 3 approaches |
| Recommendation made | Clear decision |

---

## 3. Scope

### 3.1 In Scope

- Main KDE engines (Alpha, Beta, Gamma, Delta, Epsilon)
- Bootstrap seeds (SEED-001, SEED-002, SEED-003)
- Engine selection mechanisms
- Seed selection mechanisms

### 3.2 Out of Scope

- Additional engines (adversarial-eval, consensus-synth, etc.)
- Non-standard engines
- Runtime implementation changes

---

## 4. Evidence Requirements

### 4.1 Required Evidence

- Complete engine inventory
- Complete seed inventory
- Current configuration analysis
- Option comparison matrix

---

*Specification created: 2026-07-26*
