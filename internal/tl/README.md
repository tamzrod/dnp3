# DNP3 Transport Layer

**Package:** `internal/tl`  
**Protocol Reference:** IEEE 1815-2012 Section 6  
**Steward:** Transport Layer Team

---

## Overview

The Transport Layer implements DNP3 Transport Functions as defined in IEEE 1815-2012 Section 6. It handles segmentation of Application PDUs into multiple Data Link frames and reassembles received fragments back into complete messages.

## Purpose

The Transport Layer solves the problem of sending application messages larger than one Data Link frame can carry:

- **Data Link Limit:** ~292 bytes of user data per frame
- **Application Needs:** Messages can be kilobytes
- **Solution:** Fragment into multiple frames with transport headers

## Key Concepts

### Transport Header (1 byte)

```
┌─────────────────────────────────────────┐
│  1 bit   │  1 bit   │   6 bits        │
│   FIR    │   FIN    │  Sequence Num   │
└─────────────────────────────────────────┘
```

| Field | Bit | Description |
|-------|-----|-------------|
| FIR | 7 (0x80) | First Fragment Indicator |
| FIN | 6 (0x40) | Final Fragment Indicator |
| Seq | 5-0 (0x3F) | Sequence Number (0-63) |

### Fragment Types

| FIR | FIN | Meaning |
|-----|-----|---------|
| 1 | 1 | Single fragment message |
| 1 | 0 | First of multi-fragment message |
| 0 | 0 | Middle fragment |
| 0 | 1 | Final fragment |

### Sequence Numbers

- Range: 0-63 (6 bits)
- Increment: +1 for each fragment (mod 64)
- Scope: Per-session, not per-message
- Purpose: Detect missing fragments

## Usage

### Fragmenting (Sending)

```go
f := NewFragmenter()
data := []byte{ /* application PDU */ }

// Fragmentize returns []Fragment
fragments := f.Fragmentize(data)

// Or get encoded bytes directly
encoded := f.EncodedFragments(data)

// Send each fragment via Data Link layer
for _, frag := range fragments {
    header := frag.Header()
    // ... send header + frag.Data
}
```

### Reassembling (Receiving)

```go
r := NewReassembler()

// For each received fragment
for _, data := range receivedData {
    frag, _ := DecodeFragment(data)
    
    msg, err := r.Push(frag)
    if err != nil {
        // Handle error (sequence mismatch, etc.)
    }
    
    if msg != nil {
        // Complete message received
        processMessage(msg)
    }
}
```

### Key Limits

| Constant | Value | Description |
|----------|-------|-------------|
| `MaxFragmentData` | 291 bytes | Max data per fragment |
| `SeqMax` | 63 | Maximum sequence number |
| `SeqMod` | 64 | Sequence number modulus |

## Architecture

### Components

1. **Fragment** - Represents a single transport fragment
2. **Fragmenter** - Segments Application PDUs into fragments
3. **Reassembler** - Reassembles fragments into complete messages

### State Machine (Reassembly)

```
┌──────────────┐
│ WaitingFirst │
└──────┬───────┘
       │ FIR received
       ▼
┌──────────────┐
│Reassembling  │◀──────────────────┐
└──────┬───────┘                   │
       │ FIN received               │ Timeout
       ▼                            │
┌──────────────┐                   │
│  Complete    │───────────────────┘
└──────────────┘
```

## Testing

Run unit tests:

```bash
go test ./internal/tl/... -v -cover
```

## Traceability

| Code | Architecture | Protocol |
|------|--------------|----------|
| Fragment header | ADR-001 Package Structure | 070-transport-layer.md |
| Fragmenter | ADR-001 | 070-transport-layer.md Section 6.2 |
| Reassembler | ADR-001, ADR-003 Concurrency | 070-transport-layer.md Section 6.3 |

## References

- IEEE 1815-2012 Section 6: Transport Functions
- IEEE 1815-2012 Section 6.2: Segmentation
- IEEE 1815-2012 Section 6.3: Reassembly
- Protocol Spec: `docs/protocol/dnp3/070-transport-layer.md`

---

*This package was implemented as part of Phase 4 of the go-dnp3 project.*
