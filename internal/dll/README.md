# Data Link Layer (dll)

**Status:** Phase 4.1 - In Progress

## Overview

This package implements the DNP3 Data Link Layer per IEEE 1815-2012 Section 5.

## Package Structure

```
dll/
├── crc/       # CRC-16-DNP calculations
├── frame/     # Frame encoding/decoding
├── link/      # Link state machine
└── README.md  # This file
```

## Components

### CRC Package (`crc/`)

Implements CRC-16-DNP checksum calculation:
- Polynomial: `x^16 + x^13 + x^12 + x^11 + x^10 + x^8 + x^6 + x^5 + x^2 + 1`
- LSB-first (reflected) calculation
- Used for error detection in frames

Reference: IEEE 1815-2012 Section 5.5

### Frame Package (`frame/`)

Implements DNP3 Data Link Layer frame encoding and decoding:

**Frame Format:**
```
┌────────┬────────┬────────┬────────┬────────┬──────────┬────────┬────────┐
│  0x05  │ Length │  Ctrl  │ Dest   │ Source │  Data    │  CRC   │  CRC   │
│ (0x0564│ (1 B)  │  (1 B) │ (2 B)  │ (2 B)  │ (0-292 B)│ (1 B)  │ (1 B)  │
│ sync)  │        │        │ Addr   │ Addr   │          │ (Low)  │ (High) │
└────────┴────────┴────────┴────────┴────────┴──────────┴────────┴────────┘
```

**Features:**
- Frame encoding/decoding with CRC validation
- Control byte parsing and construction
- Address range validation (0x0000 - 0xFFFF)
- Broadcast address detection
- Round-trip consistency verification

Reference: IEEE 1815-2012 Section 5.2

### Link Package (`link/`)

Implements the Data Link Layer state machine:

**States:**
- `LinkDown` - Connection not established
- `LinkReset` - Resetting link
- `Operational` - Ready to send/receive
- `WaitingConfirm` - Awaiting confirmation
- `WaitingResponse` - Awaiting response
- `Error` - Error condition

**Roles:**
- `Primary` - Initiates communication
- `Secondary` - Responds to primary

**Modes:**
- **Unbalanced**: Master is always Primary, Outstation is always Secondary
- **Balanced**: Either device can be Primary

Reference: IEEE 1815-2012 Section 5.4

## Usage

### Encoding a Frame

```go
f := &frame.Frame{
    Control: frame.Control{
        DIR:      true,
        PRM:      true,
        FuncCode: frame.FuncConfirmedUserData,
    },
    DestAddr: 0xFFFF,
    SrcAddr:  0x0000,
    Data:     []byte{0x01, 0x02, 0x03},
}

encoded, err := frame.Encode(f)
if err != nil {
    // Handle error
}
```

### Decoding a Frame

```go
decoded, err := frame.Decode(data)
if err != nil {
    // Handle error
}

// Access fields
fmt.Printf("Dest: 0x%04X\n", decoded.DestAddr)
fmt.Printf("Func: %d\n", decoded.Control.FuncCode)
```

### Using the State Machine

```go
config := link.DefaultConfig(0x0001, 0xFFFF, link.RolePrimary)
sm := link.NewStateMachine(config)

ctx := context.Background()
sm.Start(ctx)

// Send reset
sm.ResetLinkStations(ctx)

// Send data
sm.SendData(ctx, []byte{0x01, 0x02}, false)
```

## Testing

Run tests with:

```bash
go test ./internal/dll/... -v
```

Run with coverage:

```bash
go test ./internal/dll/... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

Run race detector:

```bash
go test ./internal/dll/... -race
```

## Traceability

**Architecture Reference:**
- ADR-001: Package Structure
- ADR-002: Error Handling Strategy
- ADR-003: Concurrency Model
- ADR-004: Memory Model
- ADR-005: Testing Strategy

**Protocol Reference:**
- `docs/protocol/dnp3/060-link-layer.md`

## Roadmap

- [x] CRC-16-DNP implementation
- [x] Frame encoding/decoding
- [x] Link state machine
- [x] Unit tests
- [ ] Integration tests
- [ ] Fuzz tests

## Next

Transport Layer (`internal/tl/`): Segmentation and reassembly.
