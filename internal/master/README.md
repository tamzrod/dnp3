# DNP3 Master Station

**Package:** `internal/master`  
**Protocol Reference:** IEEE 1815-2012 Section 8  
**Steward:** Master Implementation Team

---

## Overview

The Master Station is the central control system in a DNP3 SCADA network. It initiates all communication with outstations, polls for data, issues commands, and manages the overall system.

## Purpose

The Master fulfills these SCADA system roles:

1. **Initiate Communication** - Outstations don't initiate (except unsolicited)
2. **Coordinate Data Collection** - Poll multiple outstations efficiently
3. **Issue Controls** - Execute operations across the system
4. **Manage State** - Track system-wide status

## Key Concepts

### Master States

| State | Description |
|-------|-------------|
| Disconnected | No connection to network |
| Connecting | Connection in progress |
| Connected | Basic connection established |
| Initialized | Master fully initialized |
| Active | Normal operation |
| Error | Error condition |

### Polling Patterns

| Poll Type | Purpose | Frequency |
|-----------|---------|-----------|
| Integrity Poll | All Class 0 (static) data | Hourly/on startup |
| Event Poll | All pending events | Seconds to minutes |
| Exception Poll | Changed data only | Frequent |
| Class 0/1/2/3 | Specific class data | As needed |

### Control Operations

DNP3 supports safe control operations:

1. **SELECT → OPERATE**: Two-step control (select then operate)
2. **DIRECT OPERATE**: Single-step control
3. **DIRECT OPERATE NO RESPONSE**: Fire-and-forget

## Usage

### Creating a Master

```go
config := &master.Config{
    MasterAddress: 0xFFFF,
    Timeout:      5000,  // ms
    MaxRetries:   3,
}

m := master.NewMaster(config)
m.SetTransport(myTransport)
m.SetErrorHandler(func(err error, ctx string) {
    log.Printf("Error in %s: %v", ctx, err)
})
```

### Managing Outstations

```go
// Add an outstation
o := m.AddOutstation(100, "RTU-1")

// Get outstation
o, ok := m.GetOutstation(100)

// Remove outstation
m.RemoveOutstation(100)
```

### Connecting

```go
if err := m.Connect(); err != nil {
    log.Fatal(err)
}
if err := m.Initialize(); err != nil {
    log.Fatal(err)
}
```

### Polling

```go
// Integrity poll (all static data)
m.Poll(100, master.PollIntegrity)

// Event poll (all events)
m.Poll(100, master.PollEvent)

// Specific class
m.Poll(100, master.PollClass1)
```

### Control Operations

```go
// Direct operate (single step)
m.Operate(100, false, group, variation, index, value)

// Select then operate (two step, safer)
m.Operate(100, true, group, variation, index, value)
```

### Time Synchronization

```go
if err := m.TimeSync(100); err != nil {
    log.Printf("Time sync failed: %v", err)
}
```

### Unsolicited Responses

```go
// Enable unsolicited on an outstation
m.EnableUnsolicited(100)

// Disable unsolicited
m.DisableUnsolicited(100)

// Handle received unsolicited data
m.HandleUnsolicited(data)
```

## Architecture

### Components

1. **Master** - Main master station implementation
2. **Outstation** - Outstation state tracking
3. **Config** - Master configuration
4. **Poll Types** - Different polling strategies

### State Machine

```
┌────────────────┐
│ Disconnected   │
└───────┬────────┘
        │ Connect
        ▼
┌────────────────┐
│  Connecting    │
└───────┬────────┘
        │ Success
        ▼
┌────────────────┐
│  Connected     │◀───────────┐
└───────┬────────┘            │
        │ Initialize          │
        ▼                     │
┌────────────────┐            │
│ Initialized    │            │
└───────┬────────┘            │
        │ Start operations    │
        ▼                     │
┌────────────────┐            │
│    Active      │────────────┘
└───────┬────────┘ Disconnect
        │ Error
        ▼
┌────────────────┐
│     Error     │
└───────────────┘
```

## Testing

Run unit tests:

```bash
go test ./internal/master/... -v -cover
```

## Traceability

| Code | Architecture | Protocol |
|------|--------------|----------|
| Master states | ADR-001 Package Structure | 120-master.md |
| Polling | ADR-001 | 120-master.md Section 8.3 |
| Control operations | ADR-001 | 120-master.md Section 8.4 |
| Time sync | ADR-001 | 120-master.md Section 8.5 |

## References

- IEEE 1815-2012 Section 8: Master Station Functions
- Protocol Spec: `docs/protocol/dnp3/120-master.md`

---

*This package was implemented as part of Phase 4 of the go-dnp3 project.*
