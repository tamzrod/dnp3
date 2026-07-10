# DNP3 Application Layer

**Package:** `internal/al`  
**Protocol Reference:** IEEE 1815-2012 Section 7  
**Steward:** Application Layer Team

---

## Overview

The Application Layer implements DNP3 Application Layer functions as defined in IEEE 1815-2012 Section 7. It defines data encoding, operations, and request/response structures for DNP3 protocol.

## Purpose

The Application Layer provides:

1. **Data encoding** - Object groups, variations, and qualifiers
2. **Operations** - READ, WRITE, OPERATE, etc.
3. **Request/Response** - Structured communication patterns
4. **Confirmation** - Application-level acknowledgment

## Key Concepts

### Application PDU (APDU)

```
┌─────────────────────────────────────────┐
│  Application Control (1 B)              │
│  ┌─────┬─────┬─────┬─────┬───────────┐ │
│  │ FIR │ FIN │ CON │ UNS │   Seq     │ │
│  └─────┴─────┴─────┴─────┴───────────┘ │
├─────────────────────────────────────────┤
│  Function Code (1 B)                    │
├─────────────────────────────────────────┤
│  Objects (Variable)                     │
└─────────────────────────────────────────┘
```

### Application Control Field

| Bit | Name | Description |
|-----|------|-------------|
| 7 | FIR | First Fragment |
| 6 | FIN | Final Fragment |
| 5 | CON | Confirmation Required |
| 4 | UNS | Unsolicited Response |
| 3-0 | Seq | Sequence Number (0-15) |

### Internal Indication (IIN)

Every response includes a 2-byte IIN indicating device status:

| Byte | Flags |
|------|-------|
| IIN.1 | ALL_STOP, BYTE_OVER, 64K_LIM, 16K_LIM, MEM_UNAVAIL, CHECK_FAIL, BUSY, BRSV |
| IIN.2 | ABT_TRAN, ATB, DLT_AVAIL, CFG_ERR, MEM_UNAVAILABLE, SYN, ENA, EIB |

## Usage

### Creating Requests

```go
// Create a READ request
req := al.NewRequest(seq, al.FuncRead)

// Create a WRITE request
req := al.NewRequest(seq, al.FuncWrite)

// Create with data
req := &al.APDU{
    Control: al.AppControl{FIR: true, FIN: true, Seq: seq},
    FuncCode: al.FuncRead,
    Data: objectData,
}

// Encode to bytes
bytes := req.Encode()
```

### Creating Responses

```go
// Create a response with IIN
resp := al.NewAppResponse(seq, al.IIN{Busy: true}, data)

// Encode to bytes
bytes := resp.Encode()
```

### Decoding

```go
// Decode APDU
apdu, err := al.Decode(data)

// Check if request or response
if apdu.IsRequest() {
    // Handle request
}
if apdu.IsConfirmationRequired() {
    // Send confirmation
}
```

## Function Codes

### Master → Outstation

| Code | Name | Description |
|------|------|-------------|
| 2 | READ | Read data |
| 3 | WRITE | Write data |
| 4 | SELECT | Select for operate |
| 5 | OPERATE | Execute operation |
| 6 | DIRECT_OPERATE | Direct operate |
| 7 | DIRECT_OPERATE_NO_RESPONSE | No response needed |
| 41 | ENABLE_UNSOLICITED | Enable unsolicited |
| 42 | DISABLE_UNSOLICITED | Disable unsolicited |

### Outstation → Master

| Code | Name | Description |
|------|------|-------------|
| 0 | RESPONSE | Response with data |
| 1 | UNSOLICITED_RESPONSE | Unsolicited response |
| 127 | NO_ACK | No acknowledgment |

## Architecture

### Components

1. **AppControl** - Application control byte (FIR, FIN, CON, UNS, Seq)
2. **APDU** - Application Protocol Data Unit
3. **IIN** - Internal Indication field
4. **Response** - Response with IIN

## Testing

Run unit tests:

```bash
go test ./internal/al/... -v -cover
```

## Traceability

| Code | Architecture | Protocol |
|------|--------------|----------|
| APDU structure | ADR-001 Package Structure | 080-application-layer.md |
| Application control | ADR-001 | 080-application-layer.md Section 7.1 |
| Function codes | ADR-001 | 080-application-layer.md Section 7.2 |
| IIN field | ADR-001 | 080-application-layer.md Section 7.3 |

## References

- IEEE 1815-2012 Section 7: Application Layer
- IEEE 1815-2012 Section 7.1: Application Layer Function
- IEEE 1815-2012 Section 7.2: Function Codes
- IEEE 1815-2012 Section 7.3: Internal Indication
- Protocol Spec: `docs/protocol/dnp3/080-application-layer.md`

---

*This package was implemented as part of Phase 4 of the go-dnp3 KDSE case study.*
