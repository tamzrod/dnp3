# DNP3 Event Subsystem Architecture

**Date:** 2026-08-03
**Status:** Implemented

---

## 1. Overview

The Event Subsystem provides the DNP3 protocol layer for detecting point changes, generating timestamped events, and managing event queues. This enables the outstation to report changes to the master through the standard DNP3 event reporting mechanism.

### Key Design Principles

1. **Separation of Concerns**: Points (current state) and Events (change history) are separate responsibilities
2. **Configuration-Driven**: Event generation is configurable per point type
3. **Thread-Safe**: All operations are protected by mutexes
4. **Protocol-Aligned**: Events use standard DNP3 object groups and variations

---

## 2. Architecture

### Layer Structure

```
┌─────────────────────────────────────────────────────────────────┐
│                        Outstation                                │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │              Point Data (Current State)                   │    │
│  │  ├── Binary Inputs    ├── Analog Inputs                  │    │
│  │  ├── Counters        ├── Binary Outputs                 │    │
│  │  └── Analog Outputs                                        │    │
│  └─────────────────────────────────────────────────────────┘    │
│                              │                                    │
│                              ▼                                    │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                  Event Engine                            │    │
│  │  • Detects value changes                                │    │
│  │  • Generates timestamped events                         │    │
│  │  • Manages event configuration                          │    │
│  └─────────────────────────────────────────────────────────┘    │
│                              │                                    │
│                              ▼                                    │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    Event Queue                            │    │
│  │  • FIFO per event class (Class 1, 2, 3)                │    │
│  │  • Thread-safe operations                                │    │
│  │  • Overflow handling                                    │    │
│  └─────────────────────────────────────────────────────────┘    │
│                              │                                    │
│                              ▼                                    │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                  Event Builder                            │    │
│  │  • Encodes events to DNP3 object format                  │    │
│  │  • Group 2: Binary Input Events                          │    │
│  │  • Group 11: Binary Output Events                        │    │
│  │  • Group 22: Counter Events                             │    │
│  │  • Group 32: Analog Input Events                       │    │
│  │  • Group 42: Analog Output Events                       │    │
│  └─────────────────────────────────────────────────────────┘    │
│                              │                                    │
└──────────────────────────────┼──────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Protocol Layer                              │
│                                                                  │
│  Master ←── Read Response (with events) ── Outstation           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 3. Component Details

### 3.1 Event Types (`event.go`)

```go
// Event represents a DNP3 event
type Event struct {
    PointType PointType    // Type of point (Binary Input, Analog, etc.)
    Index     uint16        // Point index within its type
    Value     []byte        // Value at time of change
    Quality   uint8         // Quality flags at time of change
    Time      time.Time     // Timestamp when event was generated
    Class     EventClass    // Priority class (1, 2, or 3)
    Group     uint8         // DNP3 object group for this event
    Variation uint8         // DNP3 variation for this event
}

// PointType identifies the type of data point
type PointType uint8

const (
    PointTypeBinaryInput   PointType = 1
    PointTypeDoubleBinary  PointType = 2  // Not implemented
    PointTypeBinaryOutput PointType = 10
    PointTypeCounter      PointType = 20
    PointTypeFrozenCounter PointType = 21 // Not implemented
    PointTypeAnalogInput  PointType = 30
    PointTypeAnalogOutput PointType = 40
)

// EventClass represents the priority class
type EventClass uint8

const (
    Class1 EventClass = 1  // Highest priority
    Class2 EventClass = 2
    Class3 EventClass = 3  // Lowest priority
)
```

### 3.2 Event Configuration (`event.go`)

```go
// EventConfig enables/disables event generation per point type
type EventConfig struct {
    mu       sync.RWMutex
    enabled  map[PointType]bool
    classes  map[PointType]EventClass
}

// Default Configuration:
//   Binary Input Events:  Enabled, Class 1
//   Analog Input Events:  Enabled, Class 1
//   Counter Events:       Enabled, Class 1
//   Binary Output Events:  Enabled, Class 1
//   Analog Output Events: Enabled, Class 1
//   Frozen Counter Events: Disabled (not implemented)
```

### 3.3 Event Engine (`engine.go`)

The Event Engine is responsible for change detection and event generation.

```go
type EventEngine struct {
    mu       sync.RWMutex
    config   *EventConfig
    queue    *EventQueue
    prevValues map[string]PointValue  // Previous values for comparison
    onEventGenerated func(Event)       // Callback for event generation
}

// CheckAndGenerate checks if a point value has changed and generates an event
// Returns true if an event was generated
func (e *EventEngine) CheckAndGenerate(pointType PointType, index uint16, newValue []byte, quality uint8) bool
```

**Change Detection Logic:**
1. First value for a point → Store value, no event
2. Subsequent values → Compare with stored value
3. If value changed → Generate event
4. If quality changed → Generate event

### 3.4 Event Queue (`queue.go`)

The Event Queue manages event storage per class.

```go
type EventQueue struct {
    mu       sync.Mutex
    buffers  map[EventClass][]Event
    maxSize  int
}

// Operations:
// Push(event)     - Add event to queue (FIFO within class)
// Pop(maxCount)   - Remove and return up to maxCount events
// GetAll()        - Return all events without removing
// Clear()         - Remove all events
// Count()         - Return total event count
// IsFull()        - Check if any class is at capacity
```

**Queue Organization:**
- Each event class (1, 2, 3) has its own buffer
- Events are served in priority order (Class 1 first)
- Maximum size is divided equally among classes

### 3.5 Event Builder (`builder.go`)

The Event Builder encodes events into DNP3 object format.

```go
type EventBuilder struct{}

// BuildEventObjects encodes events into DNP3 APDU data format
func (b *EventBuilder) BuildEventObjects(events []Event) []byte
```

**Encoding by Point Type:**

| Point Type | DNP3 Group | Variation | Data Format |
|------------|------------|-----------|-------------|
| Binary Input | 2 | 1 | Index(2) + Flags(1) + Time(6) |
| Binary Output | 11 | 1 | Index(2) + Flags(1) + Time(6) |
| Counter | 22 | 1 | Index(2) + Value(4) + Flags(1) + Time(6) |
| Analog Input | 32 | 1 | Index(2) + Float32(4) + Flags(1) + Time(6) |
| Analog Output | 42 | 1 | Index(2) + Float32(4) + Flags(1) + Time(6) |

---

## 4. Event Lifecycle

### 4.1 Event Generation Flow

```
Point Value Changes
        │
        ▼
┌───────────────────┐
│ Event Engine      │
│ Check Enabled?    │──── No ────┐
└───────────────────┘            │
        │ Yes                     │
        ▼                        │
┌───────────────────┐            │
│ First Value?      │──── Yes ───┴─── Store, No Event
└───────────────────┘
        │ No
        ▼
┌───────────────────┐
│ Value Changed?    │──── No ──── Return
└───────────────────┘
        │ Yes
        ▼
┌───────────────────┐
│ Create Event      │
│ • Point Type      │
│ • Index           │
│ • Value           │
│ • Quality         │
│ • Timestamp       │
│ • Class           │
│ • Group/Variation │
└───────────────────┘
        │
        ▼
┌───────────────────┐
│ Push to Queue     │
│ (Thread-Safe)     │
└───────────────────┘
        │
        ▼
┌───────────────────┐
│ Callback (if set) │
└───────────────────┘
```

### 4.2 Event Retrieval Flow

```
Master Request (Read Class 1/2/3)
        │
        ▼
┌───────────────────┐
│ Outstation        │
│ buildAllEvents()  │
└───────────────────┘
        │
        ▼
┌───────────────────┐
│ Event Queue       │
│ Pop(maxCount)     │
└───────────────────┘
        │
        ▼
┌───────────────────┐
│ Event Builder     │
│ Encode to DNP3   │
│ Object Format     │
└───────────────────┘
        │
        ▼
┌───────────────────┐
│ Response APDU    │
│ (with events)     │
└───────────────────┘
        │
        ▼
┌───────────────────┐
│ Master Receives   │
│ and Parses        │
└───────────────────┘
```

---

## 5. Integration Points

### 5.1 Outstation Integration

The Event Subsystem integrates with the outstation through:

1. **EventEngine initialization** in `NewOutstation()`
2. **GenerateEvent()** method for explicit event generation
3. **buildAllEvents()** for including events in responses
4. **EventCount()**, **ClearEvents()** for management

```go
// In NewOutstation()
eventConfig := events.NewEventConfig()
eventEngine := events.NewEventEngine(eventConfig, config.MaxEventBuffers)

// In Outstation
func (o *Outstation) GenerateEvent(class events.EventClass, ...) bool
func (o *Outstation) buildAllEvents() []byte
```

### 5.2 Point Update Integration

The Event Engine can be called during point updates:

```go
// Example: Update a binary input with event generation
func UpdateBinaryInput(engine *events.EventEngine, index uint16, value bool, quality uint8) {
    valueBytes := []byte{0}
    if value {
        valueBytes[0] = 1
    }
    engine.CheckAndGenerate(events.PointTypeBinaryInput, index, valueBytes, quality)
}
```

### 5.3 Master Decoder Integration

The master already supports event parsing:

```go
// In pkg/dnp3/master/client.go
// parseBinaryInputs() handles Group 1 (static) and Group 2 (event)
// parseAnalogInputs() handles Group 30 (static) and Group 31 (event)
// parseCounters() handles Group 20 (static) and Group 21 (event)
```

Events in responses are distinguished by:
- **Group number**: 2, 11, 22, 32, 42 (events) vs 1, 10, 20, 30, 40 (static)
- **Timestamp presence**: Events include 6-byte DNP3 timestamp

---

## 6. Configuration

### 6.1 Default Configuration

```go
config := events.NewEventConfig()

// All point types enabled by default
config.IsEnabled(events.PointTypeBinaryInput)  // true
config.IsEnabled(events.PointTypeAnalogInput)  // true
config.IsEnabled(events.PointTypeCounter)      // true
config.IsEnabled(events.PointTypeBinaryOutput)  // true
config.IsEnabled(events.PointTypeAnalogOutput)  // true

// All events default to Class 1
config.GetClass(events.PointTypeBinaryInput)   // Class1
```

### 6.2 Custom Configuration

```go
// Disable events for a point type
config.SetEnabled(events.PointTypeBinaryOutput, false)

// Change event class
config.SetClass(events.PointTypeAnalogInput, events.Class2)

// Re-enable events
config.SetEnabled(events.PointTypeBinaryOutput, true)
```

### 6.3 Per-Outstation Configuration

```go
config := &outstation.Config{
    OutstationAddress: 1024,
    MaxEventBuffers:   1000,  // Buffer size per class
    Unsolicited:       false, // Not implemented yet
}
ost := outstation.NewOutstation(config)

// Access event engine for configuration
engine := ost.EventEngine()
engine.Config().SetEnabled(events.PointTypeBinaryInput, false)
```

---

## 7. DNP3 Object Groups

### 7.1 Supported Event Groups

| Group | Name | Variation | Format |
|-------|------|----------|--------|
| 2 | Binary Input Event | 1 | Index(2) + Flags(1) + Time(6) |
| 11 | Binary Output Event | 1 | Index(2) + Flags(1) + Time(6) |
| 22 | Counter Event | 1 | Index(2) + Value(4) + Flags(1) + Time(6) |
| 32 | Analog Input Event | 1 | Index(2) + Float32(4) + Flags(1) + Time(6) |
| 42 | Analog Output Event | 1 | Index(2) + Float32(4) + Flags(1) + Time(6) |

### 7.2 Timestamp Format

DNP3 timestamps are 6 bytes:
- **Bytes 0-4**: Absolute time (milliseconds since 2000-01-01)
- **Byte 5**: Relative time (milliseconds since last event)

---

## 8. Thread Safety

All Event Subsystem components are thread-safe:

| Component | Mutex Type | Protected Operations |
|-----------|------------|---------------------|
| EventConfig | RWMutex | IsEnabled, SetEnabled, GetClass, SetClass |
| EventQueue | Mutex | Push, Pop, GetAll, Clear, Count |
| EventEngine | RWMutex | CheckAndGenerate, GetEvents, PopEvents, ClearEvents |

The Event Engine releases its lock before calling the event callback to prevent deadlocks.

---

## 9. Known Limitations

The following are **intentionally not implemented** in this phase:

| Feature | Status | Reason |
|---------|--------|--------|
| Class 1/2/3 Retrieval | **TODO** | Requires master-side request support |
| Event Confirmation | **TODO** | Requires Class 1/2/3 response parsing |
| Unsolicited Responses | **TODO** | Requires background task and trigger logic |
| Time Synchronization | **TODO** | Separate feature |
| Double-bit Binary Events (Group 3) | **TODO** | Not implemented |
| Frozen Counter Events (Group 23) | **TODO** | Not implemented |
| Deadband Configuration | **TODO** | For analog event thresholding |
| Event Time-to-Live | **TODO** | For aging out old events |

---

## 10. Files

### Core Implementation
- `internal/outstation/events/event.go` - Event types and configuration
- `internal/outstation/events/engine.go` - Event engine (change detection)
- `internal/outstation/events/queue.go` - Event queue management
- `internal/outstation/events/builder.go` - DNP3 object encoding
- `internal/outstation/events/handler.go` - Event-generating data handler

### Tests
- `internal/outstation/events/events_test.go` - Unit tests

### Demo
- `test/integration/event_demo.go` - Demonstration script

### Documentation
- `docs/architecture/EVENT-SUBSYSTEM.md` - This document

---

## 11. Verification

### Build
```bash
$ go build ./...
# Success
```

### Tests
```bash
$ go test ./internal/outstation/events/... -v
=== RUN   TestEventConfig
--- PASS: TestEventConfig
=== RUN   TestEventQueue
--- PASS: TestEventQueue
=== RUN   TestEventQueueConcurrency
--- PASS: TestEventQueueConcurrency
=== RUN   TestEventEngine
--- PASS: TestEventEngine
=== RUN   TestEventEngineCallback
--- PASS: TestEventEngineCallback
=== RUN   TestEventBuilder
--- PASS: TestEventBuilder
PASS
```

### Demo Output
```
$ go run test/integration/event_demo.go
DNP3 Event Subsystem Demonstration
===========================================

1. Binary Input Event Generation
  Step 1: Initialize Binary Input 0 to TRUE
  Events in queue: 0
  Step 2: Set Binary Input 0 to FALSE (change)
  Events in queue: 1

2. Analog Input Event Generation
  ...

===========================================
Event Subsystem Demonstration Complete
===========================================
```

---

*Document Version: 1.0*
*Last Updated: 2026-08-03*
