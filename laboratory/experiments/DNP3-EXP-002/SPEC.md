---
id: DNP3-EXP-002
type: experiment
title: "Public API Wiring to Internal Implementations"
status: completed
authority: "KDE Runtime (DNP3 Library)"
created: "2026-07-26T05:30:00Z"
execution_agent: "OpenHands Agent"
engine:
  id: KDE-ENGINE-002
  version: "0.1.0"
  codename: "Beta"
seed:
  id: SEED-003
  codename: "Bootstrap"
---

# DNP3-EXP-002 Specification

**Experiment ID**: DNP3-EXP-002  
**Title**: Public API Wiring to Internal Implementations  
**Status**: DRAFT  
**Date**: 2026-07-26  
**Authority**: KDE Runtime (DNP3 Library)  
**Execution Agent**: OpenHands Agent  
**Engine**: KDE-ENGINE-002 (Beta) v0.1.0  
**Seed**: SEED-003 (Bootstrap)

---

## 1. Problem Statement

### 1.1 Issue Description

From README.md (Current Status):
> ⚠️ Public API wiring to internal implementations

The DNP3 library has internal implementations (internal/master, internal/outstation, internal/al, internal/tl, internal/dll) but the public API in `pkg/dnp3` may not be fully wired to these implementations.

### 1.2 Evidence

From IMP-STATUS-001:
| Component | Lines | Status |
|-----------|-------|--------|
| Types | dnp3.go | ✅ Complete |
| Master Client Interface | client.go | ⚠️ Interface Defined, Wiring Unverified |
| Outstation Server Interface | server.go | ⚠️ Interface Defined, Wiring Unverified |

### 1.3 Impact

| Impact | Severity | Description |
|--------|----------|-------------|
| Library usability | HIGH | Users cannot use the library |
| Library completeness | HIGH | Implementation exists but inaccessible |
| Integration | MEDIUM | Layers work but not exposed |

---

## 2. Hypothesis

### H1: Public API is not fully wired to internal implementations

**Statement**: The public API in `pkg/dnp3/master/client.go` and `pkg/dnp3/outstation/server.go` exists but does not properly instantiate or call internal implementations.

### H2: Missing constructor/wiring code

**Statement**: The public API lacks proper constructors that create and wire together:
- `internal/master.Master` with `internal/tl.Fragmenter`, `internal/tl.Reassembler`
- `internal/outstation.Outstation` with `internal/tl.Fragmenter`, `internal/dll/frame`

---

## 3. Investigation Plan

### 3.1 Objectives

| ID | Objective | Success Criteria |
|----|-----------|------------------|
| O1 | Analyze public API structure | All public functions identified |
| O2 | Analyze internal implementations | All internal types identified |
| O3 | Identify wiring gaps | Gap list with severity |
| O4 | Propose solution | Implementation plan documented |

### 3.2 Success Criteria

| Criterion | Metric | Target |
|-----------|--------|--------|
| Analysis complete | Functions analyzed | 100% |
| Gaps identified | Gaps documented | Complete list |
| Solution proposed | Implementation plan | Documented |

---

## 4. Evidence Requirements

### 4.1 Required Evidence

- [ ] Public API function inventory
- [ ] Internal implementation inventory  
- [ ] Wiring gap analysis
- [ ] Proposed solution with code examples

### 4.2 Verification Commands

```bash
# List public API
grep -n "func " pkg/dnp3/master/client.go
grep -n "func " pkg/dnp3/outstation/server.go

# List internal types
grep -rn "type.*struct" internal/master/
grep -rn "type.*struct" internal/outstation/

# Check imports
grep -n "import" pkg/dnp3/master/client.go
```

---

## 5. Scope

### 5.1 In Scope

- `pkg/dnp3/master/client.go` - Master client API
- `pkg/dnp3/outstation/server.go` - Outstation server API
- `internal/master/master.go` - Master implementation
- `internal/outstation/outstation.go` - Outstation implementation
- Transport layer integration (internal/tl)
- Data link layer integration (internal/dll)

### 5.2 Out of Scope

- TLS transport implementation
- Serial transport
- Example code
- CLI tools
- Documentation updates

---

## 6. Dependencies

| Dependency | Status |
|------------|--------|
| Go toolchain | ✅ AVAILABLE |
| Python | ✅ AVAILABLE |
| PyYAML | ✅ AVAILABLE |

---

## 7. Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Complex wiring required | MEDIUM | HIGH | Document thoroughly |
| Breaking changes | LOW | MEDIUM | Preserve existing interfaces |
| Test failures | MEDIUM | MEDIUM | Run tests after changes |

---

## 8. Related Documents

| Document | Relationship |
|----------|---------------|
| [docs/IMPLEMENTATION-STATUS.md](../../docs/IMPLEMENTATION-STATUS.md) | Current implementation status |
| [README.md](../../README.md) | Project status |
| [DNP3-EXP-001](../DNP3-EXP-001/CONCLUSION.md) | Previous experiment |

---

*Specification created: 2026-07-26*
*Engine: KDE-ENGINE-002 (Beta)*
*Seed: SEED-003 (Bootstrap)*
