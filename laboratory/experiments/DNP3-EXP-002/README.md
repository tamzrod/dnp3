# DNP3-EXP-002 Experiment Report

**Experiment ID**: DNP3-EXP-002  
**Title**: Public API Wiring to Internal Implementations  
**Status**: COMPLETED  
**Date**: 2026-07-26  
**Authority**: KDE Runtime (DNP3 Library)  
**Execution Agent**: OpenHands Agent  
**Engine**: KDE-ENGINE-002 (Beta)

---

## 1. Executive Summary

### 1.1 Hypothesis Tested

| H# | Hypothesis | Result |
|----|-----------|--------|
| H1 | Public API is not fully wired to internal implementations | **PARTIALLY TRUE** |
| H2 | Missing constructor/wiring code | **FALSE** |

### 1.2 Key Finding

**The wiring EXISTS but is INCOMPLETE.**

| Component | Status | Details |
|-----------|--------|---------|
| Constructor wiring | ✅ COMPLETE | `NewClient()` and `NewServer()` create internal implementations |
| Lifecycle wiring | ✅ COMPLETE | `Connect()`, `Initialize()`, `Start()`, `Stop()` delegate to internal |
| Operation wiring | ⚠️ PARTIAL | `Read()` and `Operate()` bypass internal master |

---

## 2. Detailed Analysis

### 2.1 Master Client Wiring Analysis

| Operation | Internal Master Used | Assessment |
|-----------|---------------------|------------|
| `NewClient()` | ✅ `master.NewMaster()` created | COMPLETE |
| `Connect()` | ✅ `internalMaster.Connect()` | COMPLETE |
| `Initialize()` | ✅ `internalMaster.Initialize()` | COMPLETE |
| `Read()` | ❌ Direct APDU build | **GAP** |
| `Operate()` | ❌ Direct APDU build | **GAP** |
| `EnableUnsolicited()` | ❌ Direct APDU build | **GAP** |
| `DisableUnsolicited()` | ❌ Direct APDU build | **GAP** |

**GAP Location**: `pkg/dnp3/master/client.go` lines 397-438 (Read), 641-677 (Operate)

### 2.2 Outstation Server Wiring Analysis

| Operation | Internal Outstation Used | Assessment |
|-----------|------------------------|------------|
| `NewServer()` | ✅ `outstation.NewOutstation()` created | COMPLETE |
| `Start()` | ✅ `internalOutstation.Initialize()`, `Start()` | COMPLETE |
| `Stop()` | ✅ `internalOutstation.Stop()` | COMPLETE |
| Data Handler | ✅ `internalOutstation.SetDataHandler()` | COMPLETE |
| Request Handling | ✅ `internalOutstation.Run()` in loop | COMPLETE |

**Status**: ✅ COMPLETE

---

## 3. Gap Analysis

### 3.1 Master Read() Gap

**Current Implementation** (`pkg/dnp3/master/client.go:397-438`):
```go
func (c *client) Read(ctx context.Context, request *types.ReadRequest) (*ReadResponse, error) {
    // Builds APDU directly
    apdu := buildReadRequest(c.sequence, request)
    data := apdu.Encode()
    respData, err := c.sendAndReceive(data)
    // ...
}
```

**Issue**: Bypasses `internalMaster.Poll()` which has proper retry/timeout logic.

**Recommended Fix**: Use `internalMaster.Poll()` instead:
```go
func (c *client) Read(ctx context.Context, request *types.ReadRequest) (*ReadResponse, error) {
    // Use internal master for proper protocol handling
    pollType := mapRequestToPollType(request) // Custom mapping
    if err := c.internalMaster.Poll(c.config.OutstationAddress, pollType); err != nil {
        return nil, err
    }
    // ... parse response
}
```

### 3.2 Master Operate() Gap

**Current Implementation** (`pkg/dnp3/master/client.go:641-677`):
```go
func (c *client) Operate(ctx context.Context, command *types.ControlOutput) (*OperateResponse, error) {
    // Builds APDU directly
    apdu := buildOperateRequest(c.sequence, command)
    data := apdu.Encode()
    // ...
}
```

**Issue**: Bypasses `internalMaster.Operate()` which has proper select-then-operate logic.

---

## 4. Verification Results

### 4.1 Build Status

```bash
$ go build ./...
# SUCCESS - No errors
```

### 4.2 Test Status

```
ok  dnp3/benchmarks       0.003s
ok  dnp3/internal/al      0.003s
ok  dnp3/internal/dll/crc        0.003s
ok  dnp3/internal/dll/frame      0.006s
ok  dnp3/internal/dll/link       0.004s
ok  dnp3/internal/master 0.004s
ok  dnp3/internal/outstation     0.259s
ok  dnp3/pkg/dnp3/master 0.005s
ok  dnp3/pkg/dnp3/outstation    0.006s
ok  dnp3/test/integration       0.930s
```

**All tests pass ✅**

---

## 5. Root Cause Analysis

### 5.1 Root Cause: Inconsistent Delegation Pattern

| Layer | Pattern | Status |
|-------|---------|--------|
| Lifecycle operations | Delegate to internal | ✅ |
| Protocol operations | Direct implementation | ❌ |

**Why**: The `Read()` and `Operate()` functions were likely implemented before the internal master was fully integrated.

### 5.2 Impact Assessment

| Impact | Severity | Description |
|--------|----------|-------------|
| Protocol correctness | MEDIUM | Missing retry/timeout logic from internal master |
| Maintainability | MEDIUM | Duplicated APDU building code |
| Feature parity | LOW | Internal master features (select-then-operate) unavailable |

---

## 6. Recommendations

### REC-1: Wire Master Read() to Internal Master

**Priority**: MEDIUM  
**Effort**: LOW  
**Action**: Replace direct APDU building with `internalMaster.Poll()`

### REC-2: Wire Master Operate() to Internal Master

**Priority**: MEDIUM  
**Effort**: MEDIUM  
**Action**: Replace direct APDU building with `internalMaster.Operate()`

### REC-3: Add Integration Tests

**Priority**: HIGH  
**Effort**: LOW  
**Action**: Add tests that verify the wiring actually works end-to-end

---

## 7. Evidence

### 7.1 Files Analyzed

| File | Lines | Purpose |
|------|-------|---------|
| `pkg/dnp3/master/client.go` | 723 | Public Master API |
| `pkg/dnp3/outstation/server.go` | 585 | Public Outstation API |
| `internal/master/master.go` | 575+ | Internal Master implementation |
| `internal/outstation/outstation.go` | 912+ | Internal Outstation implementation |

### 7.2 Commands Run

```bash
# Build verification
go build ./...

# Test verification  
go test ./...

# Code analysis
grep -n "internalMaster" pkg/dnp3/master/client.go
grep -n "internalOutstation" pkg/dnp3/outstation/server.go
```

---

## 8. Conclusion

**Answer**: Public API wiring to internal implementations is **PARTIALLY COMPLETE**.

- ✅ Constructor wiring: Complete
- ✅ Lifecycle wiring: Complete  
- ⚠️ Operation wiring: Incomplete (Master Read/Operate bypass internal)

**Next Steps**:
1. Wire `Read()` to use `internalMaster.Poll()`
2. Wire `Operate()` to use `internalMaster.Operate()`
3. Add integration tests

---

*Experiment completed: 2026-07-26*  
*Engine: KDE-ENGINE-002 (Beta)*  
*Classification: PUBLIC API WIRING*
