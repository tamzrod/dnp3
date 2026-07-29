# REAUDIT-FIX-001: Re-Audit Fixes (P0 & P1)

**Source**: KDE-INV-REAUDIT-001  
**Date**: 2026-07-29  
**Status**: ✅ IMPLEMENTED

## Fixes Applied

### P0-1: Connect Order Bug (FIXED)

**Problem**: `AddOutstation` was called AFTER `internalMaster.Connect()`, so link handshake had no destinations.

**Before**:
```go
// 1. transport.Connect()
// 2. internalMaster.Connect()  ← handshake with empty outstations map!
// 3. AddOutstation()           ← too late
```

**After** (`pkg/dnp3/master/client.go`):
```go
// 1. AddOutstation()           ← FIRST - so handshake has targets
// 2. transport.Connect()
// 3. internalMaster.Connect()  ← handshake with registered outstations
// 4. Initialize()
```

---

### P0-2: Frame Length Calculation (FIXED)

**Problem**: `TCPTransport.Receive()` didn't account for CRC bytes, causing under-read.

**Before**:
```go
totalSize := 2 + 1 + frameLength  // sync + length + data only
```

**After** (`pkg/transport/tcp.go`):
```go
// Calculate data length from frame length
dataLen := frameLength - 5 // Control(1) + Dest(2) + Src(2)

// Calculate CRC bytes:
// - 3 header CRCs = 6 bytes
// - Data CRCs = ceil(dataLen/2) pairs * 2 bytes
numDataCRCPairs := (dataLen + 1) / 2
crcBytes := 6 + (numDataCRCPairs * 2)

// Total: sync(2) + length(1) + data + CRCs
totalSize := 2 + 1 + frameLength + crcBytes
```

---

### P1-1: Link-Layer Handlers (FIXED)

**Problem**: Outstation didn't respond to link-layer frames (Reset Link, Link Status).

**Added** (`internal/outstation/outstation.go`):

1. **Reset Link Stations handling** in `reassembleMessage()`:
```go
if dllFrame.Control.PRM && dllFrame.Control.FuncCode == frame.FuncResetLinkStations {
    o.sendLinkAck(dllFrame.SrcAddr)  // Send ACK
    continue
}
```

2. **Link Status Request handling**:
```go
if dllFrame.Control.PRM && dllFrame.Control.FuncCode == frame.FuncResetLinkStatus {
    o.sendLinkStatus(dllFrame.SrcAddr)  // Send Link Status
    continue
}
```

3. **New functions**:
   - `sendLinkAck(masterAddr)` - Sends DLL ACK frame
   - `sendLinkStatus(masterAddr)` - Sends DLL Link Status frame

---

## Files Modified

| File | Change |
|------|--------|
| `pkg/dnp3/master/client.go` | Fixed Connect order |
| `pkg/transport/tcp.go` | Fixed frame length calculation |
| `internal/outstation/outstation.go` | Added link-layer handlers |

---

## Build Status

✅ All packages compile successfully

---

## Remaining (P2)

- Update integration tests to verify wire format (0x05 0x64)
- Wire CommandHandler into per-connection instances

---

*From Grok re-audit findings*
