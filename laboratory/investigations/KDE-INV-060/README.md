# KDE-INV-060: Enhanced DNP3 Data Viewing Experience

**Investigation**: KDE-INV-060  
**Domain**: UI/UX, Workbench Application  
**Status**: open  
**Created**: 2026-07-27  
**Severity**: HIGH  
**Engine**: kde-engine-ux-001  
**Parent**: KDE-INV-059  

---

## Executive Summary

Design a better user experience for viewing and interacting with DNP3 data in the Engineering Workbench. This includes:

1. **Table-based data view** with checkboxes for selection
2. **Separate Master and Outstation modes** with appropriate UI for each
3. **Read/Write operations** for controlling outputs

---

## Current State Analysis

### Existing Data Types (from `pkg/dnp3/types/types.go`)

| Type | Description | Has Write Support |
|------|-------------|-------------------|
| `BinaryInput` | Single-bit status (DI) | No (read-only) |
| `BinaryOutput` | Single-bit control (DO) | Yes |
| `AnalogInput` | Float measurement (AI) | No (read-only) |
| `AnalogOutput` | Float control (AO) | Yes |
| `Counter` | 32-bit count | No (read-only) |
| `FrozenCounter` | Frozen count | No (read-only) |

### Current UI Structure

```
┌─────────────────────────────────────────────────────────────┐
│ CONNECTION                                                 │
│ IP Address: [_______________]                               │
│ TCP Port:   [_______________]                               │
│ [Connect] [Disconnect]                                    │
├─────────────────────────────────────────────────────────────┤
│ MODE SELECTION                                             │
│ ( ) Master Mode                                            │
│ ( ) Outstation Mode                                        │
│ MVP: Master Mode Only                                      │
├─────────────────────────────────────────────────────────────┤
│ POINT VALUES                                               │
│ Binary Inputs: (No data)                                  │
│ Analog Inputs: (No data)                                   │
│ Counters: (No data)                                       │
└─────────────────────────────────────────────────────────────┘
```

### Problems with Current Design

1. **Mode confusion**: Master/Outstation terminology doesn't match DNP3 roles
2. **No data table**: Simple labels don't scale for real data
3. **No write controls**: Can't operate outputs from UI
4. **No selection**: Can't select specific points to monitor

---

## Proposed Design

### Mode Separation

Rename modes to match DNP3 roles clearly:

| Old Name | New Name | Description | UI Action |
|----------|----------|-------------|-----------|
| Master Mode | **Poll Outstation** | Connect as master, poll data | Connect button |
| Outstation Mode | **Simulate Outstation** | Act as DNP3 server | Start Server button |

### Master Mode UI (Poll Outstation)

```
┌────────────────────────────────────────────────────────────────────┐
│ OUTSTATION CONNECTION                                              │
│ Address: [_______________] Port: [_____]  [Connect] [Disconnect]  │
├────────────────────────────────────────────────────────────────────┤
│ DATA MONITORING                    [Read Class 0] [Read Class 1]   │
├────────────────────────────────────────────────────────────────────┤
│ SELECT | INDEX |  DI  |  DO  |  AI  |  AO  | CNTR | VALUE | QUAL │
├────────┼────────┼------+------+------+------+------+-------+-------│
│  ☑     |   0    |  ☑   |  ☐   |  ☐   |  ☐   |  ☐   | ON    | GOOD │
│  ☐     |   1    |  ☑   |  ☐   |  ☐   |  ☐   |  ☐   | OFF   | GOOD │
│  ☑     |   2    |  ☐   |  ☐   |  ☑   |  ☐   |  ☐   | 42.5  | GOOD │
│  ☐     |   3    |  ☐   |  ☐   |  ☑   |  ☐   |  ☐   | 100.0 | GOOD │
│  ☑     |   4    |  ☐   |  ☐   |  ☐   |  ☐   |  ☑   | 1234  | GOOD │
├────────────────────────────────────────────────────────────────────┤
│ CONTROL PANEL (Selected: DI-0, DI-2, AI-2, CNTR-4)                │
│ [Operate DO0: ON] [Operate DO0: OFF] [Set AO0: ___] [Send]        │
├────────────────────────────────────────────────────────────────────┤
│ PROTOCOL LOG                                                      │
│ ...                                                               │
└────────────────────────────────────────────────────────────────────┘
```

### Outstation Mode UI (Simulate Outstation)

```
┌────────────────────────────────────────────────────────────────────┐
│ OUTSTATION SIMULATION                              [Stop Server]   │
│ Listening on: 0.0.0.0:20000                                        │
├────────────────────────────────────────────────────────────────────┤
│ DATA POINT CONFIGURATION                                           │
│ [+ Add Binary Input] [+ Add Analog Input] [+ Add Counter]           │
├────────────────────────────────────────────────────────────────────┤
│ POINT | TYPE  | INDEX | VALUE | QUALITY | EVENTS | STATUS         │
├────────┼--------+-------+-------+---------+--------+--------        │
│   ☐    | BI    |   0   |  ON   | GOOD    |   12   | [Edit]       │
│   ☐    | AI    |   0   | 42.5  | GOOD    |    5   | [Edit]       │
│   ☐    | CTR   |   0   | 1000  | GOOD    |    3   | [Edit]       │
├────────────────────────────────────────────────────────────────────┤
│ OPERATIONS LOG                                                     │
│ 10:23:45 - Master connected from 192.168.1.100                      │
│ 10:23:46 - Read Class 0 request                                    │
│ 10:23:47 - Operate DO0 = ON                                        │
└────────────────────────────────────────────────────────────────────┘
```

---

## Data Model

### DataPoint Interface

```go
type DataPoint interface {
    GetIndex() uint16
    GetType() PointType
    GetValue() interface{}
    GetQuality() types.QualityFlags
    GetTimestamp() *types.Timestamp
}
```

### Point Types

```go
type PointType string

const (
    PointTypeBinaryInput   PointType = "DI"
    PointTypeBinaryOutput  PointType = "DO"
    PointTypeAnalogInput   PointType = "AI"
    PointTypeAnalogOutput  PointType = "AO"
    PointTypeCounter       PointType = "CTR"
    PointTypeFrozenCounter PointType = "FRZ"
)
```

### Selection State

```go
type SelectionState struct {
    Selected map[PointType]map[uint16]bool  // Type -> Index -> Selected
}
```

---

## Component Design

### 1. DataTablePanel

**Features**:
- Virtual scrolling for large datasets
- Column sorting (by index, value, quality)
- Multi-select with checkboxes
- Real-time value updates
- Quality indicators with color coding

**Columns**:
| Column | Type | Width | Sortable |
|--------|------|-------|----------|
| Select | Checkbox | 40px | No |
| Index | Number | 60px | Yes |
| DI | Checkbox | 40px | No |
| DO | Checkbox | 40px | No |
| AI | Checkbox | 40px | No |
| AO | Checkbox | 40px | No |
| CTR | Checkbox | 40px | No |
| Value | String | 100px | Yes |
| Quality | Badge | 80px | Yes |
| Time | String | 140px | Yes |

### 2. ControlPanel

**For Master Mode**:
- Select output points (DO, AO)
- Enter control values
- Send operate commands
- SelectThenOperate toggle

**For Outstation Mode**:
- Edit point values
- Configure event generation
- Set quality flags

### 3. ModeSwitcher

**Behavior**:
- Clear separation between modes
- Appropriate controls for each mode
- State preservation when switching
- Confirmation if data would be lost

---

## Interaction Flows

### Master Mode: Read Data

```
1. User enters address/port
2. User clicks "Connect"
3. Connection established
4. User clicks "Read Class 0"
5. Response received
6. Data table updates with new values
7. Timestamps update
```

### Master Mode: Write Output

```
1. User selects DO points in table (checkboxes)
2. User clicks "Operate" button
3. Control panel shows selected outputs
4. User sets value (ON/OFF for DO, value for AO)
5. User clicks "Send"
6. Operate request sent
7. Response received
8. Table updates with new output status
```

### Outstation Mode: Simulate

```
1. User selects "Outstation Mode"
2. User clicks "Start Server"
3. Server listens on configured port
4. User adds/configures data points
5. Master connects
6. Master requests data
7. Server responds with configured values
8. Operations log shows requests
```

---

## Implementation Notes

### Fyne Widget Options

1. **widget.NewTree** - For hierarchical data
2. **widget.NewTable** - For flat tabular data (requires custom cell factory for checkboxes)
3. **Custom widget** - Extend widget to add checkbox column

### Recommended Approach

Use a combination of:
- `fyne.Container` with `fyne.Layout` for overall structure
- Custom `DataGrid` widget wrapping Fyne's Table
- `widget.NewCheck` for individual cells
- `binding` package for reactive updates

### Architecture Changes

| File | Changes |
|------|---------|
| `panels/data.go` | Rewrite as `DataTablePanel` |
| `panels/mode.go` | Rename to `ModePanel`, add Outstation support |
| `panels/control.go` | New file for control operations |
| `panels/outstation.go` | New file for outstation simulation |
| `session/session.go` | Add Outstation session |
| `controller/controller.go` | Handle both session types |

---

## Outstation Session (Required)

Currently only `MasterSession` exists. Need `OutstationSession`:

```go
// OutstationSession represents a DNP3 Outstation session.
type OutstationSession struct {
    mu        sync.RWMutex
    address   string
    port      int
    state     ConnectionState
    transport transport.Handler
    events    chan SessionEvent
    log       Logger
    
    // Data points
    binaryInputs   []*types.BinaryInput
    analogInputs   []*types.AnalogInput
    counters       []*types.Counter
    binaryOutputs  []*types.BinaryOutput
    analogOutputs  []*types.AnalogOutput
}

// NewOutstationSession creates a new Outstation session.
func NewOutstationSession(log Logger) (*OutstationSession, error)

// Start begins listening for connections
func (s *OutstationSession) Start(address string, port int) error

// Stop stops the outstation
func (s *OutstationSession) Stop() error
```

---

## Status History

| Date | Status | Notes |
|------|--------|-------|
| 2026-07-27 | open | Initial investigation |

---

## Related Investigations

- **KDE-INV-059**: Workbench GUI Usability Issues (parent)
