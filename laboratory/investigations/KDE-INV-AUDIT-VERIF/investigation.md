# KDE-INV-AUDIT-VERIF: External Audit Verification

**Engine**: KDE-ENGINE-001 (v1.0.0)  
**Seed**: SEED-001 (v1.0.0)  
**Timestamp**: 2026-07-29T10:45:00Z  
**Status**: 🔬 COMPLETE

## Investigation Objective

1. Verify claims from external Grok audit about DNP3 connection problems
2. Check if KDE experts were properly consulted during debugging

---

## External Audit Claims

### Audit Summary (Grok)

| Claim | Severity | Claim Summary |
|-------|----------|---------------|
| C1 | CRITICAL | TCPTransport uses custom 2-byte length prefix, not standard DNP3 |
| C2 | CRITICAL | Incomplete stack - DLL framing not used by public API path |
| C3 | HIGH | Outstation accept/run loop issues with single transport |
| C4 | HIGH | Master "Connected" means TCP only, not DNP3 session |
| C5 | MEDIUM | Tests don't prove real DNP3 connectivity |

---

## Verification

### C1: TCPTransport Length Prefix (CRITICAL)

**Audit Claim**: TCPTransport.Send/Receive uses 2-byte length prefix instead of standard DNP3 DLL framing.

**Code Evidence** (`pkg/transport/tcp.go:195-208`):
```go
// DNP3 over TCP uses length prefix
length := uint16(len(data))
header := make([]byte, 2)
binary.BigEndian.PutUint16(header, length)

_, err := t.conn.Write(header)
// ...
_, err = t.conn.Write(data)
```

**Proper DNP3 DLL Framing** (`internal/dll/frame/frame.go`):
```go
// Sync bytes indicating the start of a DNP3 frame
const (
    SyncByte1 = 0x05
    SyncByte2 = 0x64
)
// Frame format: 0x05 0x64, Length, Control, Dest, Src, Data, CRCs
```

**Verification Result**: ✅ **CONFIRMED**

The TCPTransport uses a custom 2-byte length prefix. This is NOT standard IEEE 1815 DNP3. The proper framing with sync bytes (0x05 0x64) exists in `internal/dll/frame/` but is NOT used by TCPTransport.

---

### C2: Incomplete Stack in Public Path (CRITICAL)

**Audit Claim**: Public Master/Outstation path does not reliably use AL→TL→DLL encode pipeline.

**Code Evidence** (`internal/master/master.go:426-453`):
```go
// sendWithRetry function DOES use full stack:
data := req.Encode()                        // AL
fragments := m.fragmenter.Fragmentize(data) // TL
for _, frag := range fragments {
    tlEncoded := tl.EncodeFragment(frag)   // TL
    // Data link layer frame
    dllFrame := &frame.Frame{
        // Proper DLL fields
    }
    dllEncoded, err := frame.Encode(dllFrame) // DLL
    m.transport.Send(dllEncoded)              // Sends to TCP
}
```

**BUT**: The TCPTransport then adds ANOTHER layer:
```go
// TCPTransport.Send adds:
length := uint16(len(data))    // Extra!
binary.BigEndian.PutUint16(header, length)
t.conn.Write(header)
t.conn.Write(data)             // DLL frame wrapped in length prefix
```

**Verification Result**: ⚠️ **PARTIALLY CONFIRMED**

The internal path correctly builds DLL frames, BUT:
1. TCPTransport wraps them in custom length prefix
2. This breaks interoperability with standard DNP3 devices
3. Internal master→outstation works (same custom framing)
4. External devices would fail

---

### C3: Outstation Accept Loop (HIGH)

**Audit Claim**: Single transport reused, multi-master broken, errors swallowed.

**Code Evidence** (`pkg/dnp3/outstation/server.go:474-498`):
```go
go func() {
    for {
        select {
        case <-runCtx.Done():
            return
        default:
            if err := t.Accept(); err != nil {
                if err == transport.ErrTimeout { continue }
                if err == transport.ErrClosed { return }
                continue  // other errors swallowed - NO LOGGING!
            }
            s.internalOutstation.Run()
        }
    }
}()
```

**TCPTransport Accept** (`pkg/transport/tcp.go:179`):
```go
t.conn = conn  // Single connection overwrites previous!
```

**Verification Result**: ✅ **CONFIRMED**

Issues found:
1. **Single transport per outstation**: `t.conn` is a single field, overwrites on new accept
2. **Multi-master broken**: Second master connection overwrites first
3. **Errors swallowed**: Line 492 just `continue`s with no logging
4. **Run() blocking**: `Run()` blocks indefinitely, accept loop doesn't handle properly

---

### C4: Master Connection State (HIGH)

**Audit Claim**: "Connected/Active" means TCP only, not DNP3 session established.

**Code Evidence** (`internal/master/master.go:270-279`):
```go
func (m *Master) Connect() error {
    if m.transport == nil {
        return errors.New("transport not configured")
    }
    m.SetState(StateConnecting)
    m.SetState(StateConnected)  // Just after TCP dial!
    return nil
}
```

**Verification Result**: ✅ **CONFIRMED**

`Connect()` only sets state to "Connected" after TCP dial. No DNP3 link-layer exchange is performed. True DNP3 session requires:
1. Link reset
2. Link status check
3. Application layer initialization

These are NOT performed in `Connect()`.

---

### C5: Test Coverage (MEDIUM)

**Audit Claim**: Tests don't prove real DNP3 connectivity.

**Evidence**:
- `TestTCPDirectCommunication` - exchanges raw APDUs through length-prefixed transport
- `TestTCPMasterOutstationRead` - TCP connect + state check
- No test against real DNP3 device or reference implementation

**Verification Result**: ✅ **CONFIRMED**

Tests verify the custom implementation works internally, but don't verify standard DNP3 compliance.

---

## Expert Consultation Check

### Available Experts

| Expert | Domain | Location |
|--------|--------|----------|
| DNP3-EXPERT-001 | DNP3 Protocol | `.kde/experts/dnp3-protocol/` |
| DNP3-SEC-EXPERT-001 | DNP3 Security | `.kde/experts/dnp3-security/` |

### Expert Consultation History

**Question**: Were these experts consulted during the master-outstation debugging?

| Evidence | Finding |
|----------|---------|
| Investigation artifacts | No mention of expert consultation |
| LAB-063 to LAB-067 | No evidence of expert invocation |
| Skill load | kde-investigation-framework loaded, not DNP3 experts |

**Gap**: The debugging focused on TCP connectivity and code changes, but did NOT consult the DNP3 Protocol Expert for domain-specific guidance.

---

## Root Cause Analysis

### Primary Root Cause

**TCPTransport Framing Incompatibility**

```
Master (internal)           Outstation (internal)
     |                            |
     | DLL Frame (0x05 0x64...)   |
     |                            |
     v                            v
TCPTransport.Send()    TCPTransport.Receive()
     |                            |
     | Length + DLL Frame        |
     |                            |
     +-------- TCP Wire --------+
                |
                v
     TCPTransport.Receive()  Length + DLL Frame
                |
                v
          DLL Decode() - FAILS!
```

The internal master correctly builds DNP3 DLL frames, but TCPTransport wraps them in a custom length prefix that standard receivers don't expect.

### Why Internal→Internal Works

Both master and outstation use the same TCPTransport with custom framing, so they understand each other. But any external DNP3 device would fail.

---

## Complete Audit Verification Summary

| Claim | Severity | Verification | Evidence |
|-------|----------|--------------|----------|
| C1: Length prefix framing | CRITICAL | ✅ CONFIRMED | tcp.go:195-208 |
| C2: Incomplete stack path | CRITICAL | ⚠️ PARTIAL | DLL built but wrapped |
| C3: Accept loop issues | HIGH | ✅ CONFIRMED | server.go:474-498 |
| C4: TCP-only connection | HIGH | ✅ CONFIRMED | master.go:270-279 |
| C5: Tests don't prove DNP3 | MEDIUM | ✅ CONFIRMED | No E2E tests |

---

## Expert Consultation Analysis

### Expert Gap

**Finding**: DNP3 Protocol Expert (`.kde/experts/dnp3-protocol/`) was NOT consulted.

**Evidence Required**: No mention in LAB-063 through LAB-067

**Missing**: Domain-specific guidance on IEEE 1815 compliance

---

## Recommendations from Audit

| Priority | Recommendation | Verification |
|----------|---------------|--------------|
| 1 | Remove 2-byte length prefix from TCPTransport | ✅ Confirmed |
| 2 | Wire full AL→TL→DLL pipeline in public path | ⚠️ Partially done |
| 3 | One transport per accepted connection | ✅ Confirmed broken |
| 4 | Connect semantics - link-layer exchange | ✅ Confirmed missing |
| 5 | Enable/complete E2E tests | ✅ Confirmed missing |

---

## Expert Consultation Gap

**Finding**: DNP3 Protocol Expert was NOT consulted during debugging.

**Evidence Required**:
- Expert invocation records
- Expert guidance in investigation artifacts

**Recommendation**: When debugging DNP3 protocol issues, consult `.kde/experts/dnp3-protocol/` for:
- IEEE 1815 compliance requirements
- Standard framing expectations
- Known interoperability patterns

---

## Investigation Metadata

| Field | Value |
|-------|-------|
| Pre-flight | ✅ PASSED |
| External Audit Source | Grok analysis |
| Verification Method | Code inspection |
| Expert Consultation | ❌ NOT VERIFIED |
| Audit Claims Verified | 4/5 CONFIRMED, 1 PARTIAL |
| Status | COMPLETE |

---

*Verification of external audit claims*
