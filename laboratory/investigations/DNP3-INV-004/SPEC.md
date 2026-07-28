# DNP3-INV-004: Specification

## Overview

This specification defines the changes required to implement proper separation between Master and Outstation modes in the DNP3 Engineering Workbench.

## Goals

1. Create distinct, focused interfaces for Master and Outstation modes
2. Maintain backward compatibility with existing functionality
3. Improve user experience by reducing cognitive load
4. Enable simultaneous use of both modes (optional)

## Architecture Changes

### Current Architecture

```
cmd/workbench/
├── main.go                    # Single entry point
├── internal/
│   ├── controller/
│   │   └── controller.go      # Unified controller (Master only)
│   └── session/
│       ├── session.go        # Master session
│       └── outstation.go    # Outstation session
```

### Proposed Architecture

```
cmd/workbench/
├── main.go                    # Entry point with mode selection
├── master/
│   └── main.go               # Master-specific entry (optional)
├── outstation/
│   └── main.go               # Outstation-specific entry (optional)
├── internal/
│   ├── master/
│   │   ├── controller.go     # Master controller
│   │   └── window.go         # Master window
│   ├── outstation/
│   │   ├── controller.go     # Outstation controller
│   │   └── window.go         # Outstation window
│   └── shared/
│       ├── config/           # Shared configuration
│       ├── ui/               # Shared UI components
│       └── protocol/         # Shared protocol utilities
```

## Master Window Specification

### Purpose
Connect to remote DNP3 outstations and read/write data points.

### UI Elements

| Element | Type | Description |
|---------|------|-------------|
| Connection Panel | Container | Address, port, connect button |
| Commands Panel | Container | Read Class 0/1/2/3, enable unsol |
| Data Table | Table | Sortable list of data points |
| Control Panel | Container | Write binary/analog outputs |
| Log Panel | Text | Protocol traffic log |
| Status Bar | Container | Connection state, IIN |

### Layout

```
┌─────────────────────────────────────────────────────────────────┐
│  DNP3 Master - Connect to Outstation                              │
├─────────────────────────────────────────────────────────────────┤
│  Connection                    │  Data Table                     │
│  ┌──────────────────────────┐ │  ┌───────────────────────────┐  │
│  │ Address: [127.0.0.1    ] │ │  │ Index │ Type │ Value    │  │
│  │ Port:   [20000        ] │ │  ├───────┼───────┼──────────┤  │
│  │         [Connect      ] │ │  │ 0     │ BI    │ true     │  │
│  └──────────────────────────┘ │  │ 1     │ AI    │ 100.5    │  │
│                               │  │ ...   │ ...   │ ...      │  │
│  Commands                     │  └───────────────────────────┘  │
│  ┌──────────────────────────┐ │                                │
│  │ [Read 0] [Read 1]       │ │  Control                       │
│  │ [Read 2] [Read 3]       │ │  ┌───────────────────────────┐  │
│  │ [Enable Unsolicited]    │ │  │ Point: Binary Output #0  │  │
│  └──────────────────────────┘ │  │ Value: [true]            │  │
│                               │  │ [Operate]                │  │
│                               │  └───────────────────────────┘  │
├───────────────────────────────┴────────────────────────────────┤
│  Log                                                         │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │ [10:23:01] → SEND: Read Class 0                       │  │
│  │ [10:23:02] ← RECV: 10 binary, 5 analog, 3 counters    │  │
│  └─────────────────────────────────────────────────────────┘  │
├───────────────────────────────────────────────────────────────┤
│ ● Connected │ 127.0.0.1:20000 │ IIN: 0x0000                  │
└───────────────────────────────────────────────────────────────┘
```

## Outstation Window Specification

### Purpose
Run a local DNP3 outstation with simulated or configured data points.

### UI Elements

| Element | Type | Description |
|---------|------|-------------|
| Server Panel | Container | Listen port, start/stop button |
| Data Points Panel | Container | List of binary, analog, counter points |
| Simulation Panel | Container | Enable simulation, update rate |
| Log Panel | Text | Protocol traffic log |
| Status Bar | Container | Server state, connected masters |

### Layout

```
┌─────────────────────────────────────────────────────────────────┐
│  DNP3 Outstation - Simulate Data                                │
├─────────────────────────────────────────────────────────────────┤
│  Server                          │  Data Points                  │
│  ┌────────────────────────────┐ │  ┌─────────────────────────┐  │
│  │ Port: [20000             ] │ │  │ Binary Inputs (8)      │  │
│  │ Address: [0.0.0.0        ] │ │  │ ├─ BI0: true           │  │
│  │       [Start Server    ]   │ │  │ ├─ BI1: false          │  │
│  │       [Stop Server     ]   │ │  │ └─ ...                 │  │
│  └────────────────────────────┘ │  │                         │  │
│                                 │  │ Analog Inputs (4)     │  │
│  Simulation                     │  │ ├─ AI0: 100.5          │  │
│  ┌────────────────────────────┐ │  │ └─ ...                 │  │
│  │ [✓] Enable Simulation     │ │  │                         │  │
│  │ Update Rate: [1.0] sec    │ │  │ Counters (4)          │  │
│  │ Binary Rate: [0.5] Hz     │ │  │ ├─ C0: 1000            │  │
│  │ Analog Var: [±10.0]      │ │  │ └─ ...                 │  │
│  └────────────────────────────┘ │  └─────────────────────────┘  │
├───────────────────────────────┴────────────────────────────────┤
│  Log                                                         │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │ [10:23:01] ← INCOMING: Master connected from 127.0.0.1│  │
│  │ [10:23:02] → SENT: Response to Read Class 0            │  │
│  └─────────────────────────────────────────────────────────┘  │
├───────────────────────────────────────────────────────────────┤
│  ● Listening │ Port: 20000 │ Masters: 1                      │
└───────────────────────────────────────────────────────────────┘
```

## Mode Selection

### Command Line Options

```bash
# Show mode selection dialog
workbench

# Start as Master only
workbench --mode master

# Start as Outstation only
workbench --mode outstation
```

### Mode Selection Dialog

```
┌──────────────────────────────────────────────┐
│         DNP3 Engineering Workbench           │
│                                              │
│   Choose Operating Mode:                     │
│                                              │
│   ┌────────────────────────────────────────┐ │
│   │     [Master Mode]                      │ │
│   │     Connect to remote outstations      │ │
│   │     Read/write data points             │ │
│   └────────────────────────────────────────┘ │
│                                              │
│   ┌────────────────────────────────────────┐ │
│   │     [Outstation Mode]                   │ │
│   │     Run simulated DNP3 server          │ │
│   │     Generate random data               │ │
│   └────────────────────────────────────────┘ │
│                                              │
│              [Cancel]                        │
└──────────────────────────────────────────────┘
```

## Shared Components

### UI Components

| Component | Description |
|-----------|-------------|
| `LogPanel` | Reusable log display with filtering |
| `StatusBar` | Connection/mode indicator |
| `DataTable` | Sortable data display |

### Configuration

Shared YAML configuration file:

```yaml
master:
  default_address: "127.0.0.1"
  default_port: 20000

outstation:
  default_port: 20000
  simulation:
    enabled: true
    update_rate: 1s
    binary_rate: 0.5
    analog_variance: 10.0

window:
  width: 1200
  height: 800
```

## Backward Compatibility

- Default behavior unchanged (show mode selection)
- Existing keyboard shortcuts preserved
- Configuration file format maintained
- Single-instance mode still supported

## Testing Requirements

| Test | Description |
|------|-------------|
| Master connect | Verify master connects to outstation |
| Outstation start | Verify outstation accepts connections |
| Mode switch | Verify switching between modes works |
| Data sync | Verify data displays correctly in both modes |
| Command line | Verify --mode flag works correctly |

## Implementation Phases

### Phase 1: Refactor Controllers
- Create `master.Controller`
- Create `outstation.Controller`
- Extract shared types

### Phase 2: Create Windows
- Create `MasterWindow` in `master/window.go`
- Create `OutstationWindow` in `outstation/window.go`
- Extract shared UI components

### Phase 3: Mode Selection
- Add command-line flag parsing
- Create mode selection dialog
- Implement --mode option

### Phase 4: Testing & Polish
- Integration testing
- UI polish
- Documentation updates
