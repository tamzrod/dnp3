# DNP3-INV-062 Specification: Workbench Master/Outstation Instance Separation

**Document ID**: DNP3-INV-062
**Version**: 1.0.0
**Date**: 2026-07-27
**Status**: DRAFT
**Authority**: KDE Runtime (DNP3 Library)
**Investigation**: DNP3-INV-062

---

## 1. Executive Summary

This specification defines the implementation plan for separating Master and Outstation modes into dedicated instances of the same executable.

### 1.1 Design Decision

**Decision**: Implement `--mode` flag for startup mode selection.

**Rationale**:
1. Single executable requirement satisfied
2. Complete separation between modes
3. Parallel execution capability
4. User-friendly mode selection dialog

---

## 2. Implementation Specification

### 2.1 Command-Line Interface

#### 2.1.1 Flag Specification

| Flag | Values | Default | Description |
|------|--------|---------|-------------|
| `--mode` | `master`, `outstation` | `select` | Operating mode |
| `--help` | - | - | Show help message |

#### 2.1.2 CLI Implementation

```go
// main.go
func main() {
    mode := flag.String("mode", "select", 
        "Operating mode: master, outstation, or select (default)")
    
    flag.Parse()
    
    switch *mode {
    case "master":
        runMaster()
    case "outstation":
        runOutstation()
    case "select":
        showModeSelection()
    default:
        log.Fatalf("Invalid mode: %s", *mode)
    }
}
```

### 2.2 Mode Selection

#### 2.2.1 Dialog Specification

When `--mode select` or no flag is provided, display a mode selection dialog:

```
┌──────────────────────────────────────────────┐
│         DNP3 Engineering Workbench           │
│                                              │
│   Choose Operating Mode:                    │
│                                              │
│   ┌────────────────────────────────────────┐ │
│   │  [Master Mode]                        │ │
│   │  Connect to remote outstations         │ │
│   │  Read/write data points               │ │
│   └────────────────────────────────────────┘ │
│                                              │
│   ┌────────────────────────────────────────┐ │
│   │  [Outstation Mode]                    │ │
│   │  Run simulated DNP3 server             │ │
│   │  Generate random data                 │ │
│   └────────────────────────────────────────┘ │
│                                              │
│              [Cancel]                        │
└──────────────────────────────────────────────┘
```

### 2.3 Master Window Specification

#### 2.3.1 Window Properties

| Property | Value |
|----------|-------|
| Title | "DNP3 Master - Connect to Outstation" |
| Default Size | 1200 x 800 |
| Minimum Size | 800 x 600 |

#### 2.3.2 UI Layout

```
┌─────────────────────────────────────────────────────────────────┐
│  DNP3 Master - Connect to Outstation                    [_][□][X]│
├─────────────────────────────────────────────────────────────────┤
│  File  Edit  View  Session  Help                                  │
├─────────────────────────────┬─────────────────────────────────────┤
│                             │                                      │
│  Connection                 │  Data Table                         │
│  ┌───────────────────────┐ │  ┌─────────────────────────────┐   │
│  │ Address: [127.0.0.1] │ │  │ Index │ Type │ Value    │   │
│  │ Port:   [20000     ] │ │  ├───────┼───────┼──────────┤   │
│  │       [Connect     ] │ │  │ 0     │ BI    │ true     │   │
│  │       [Disconnect  ] │ │  │ 1     │ AI    │ 100.5    │   │
│  └───────────────────────┘ │  │ ...   │ ...   │ ...      │   │
│                             │  └─────────────────────────────┘   │
│  Commands                   │                                      │
│  ┌───────────────────────┐ │  Control                             │
│  │ [Read 0] [Read 1]    │ │  ┌─────────────────────────────┐   │
│  │ [Read 2] [Read 3]    │ │  │ Selected: Binary Output #0 │   │
│  │ [Enable Unsolicited] │ │  │ [ON] [OFF]                  │   │
│  └───────────────────────┘ │  └─────────────────────────────┘   │
│                             │                                      │
├─────────────────────────────┴─────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ Log: [10:23:01] → SEND Read Class 0                       │   │
│  │       [10:23:02] ← RECV 10 binary, 5 analog              │   │
│  └─────────────────────────────────────────────────────────────┘   │
├───────────────────────────────────────────────────────────────────┤
│ ● Connected │ 127.0.0.1:20000 │ IIN: 0x0000                      │
└───────────────────────────────────────────────────────────────────┘
```

### 2.4 Outstation Window Specification

#### 2.4.1 Window Properties

| Property | Value |
|----------|-------|
| Title | "DNP3 Outstation - Simulate Data" |
| Default Size | 1200 x 800 |
| Minimum Size | 800 x 600 |

#### 2.4.2 UI Layout

```
┌─────────────────────────────────────────────────────────────────┐
│  DNP3 Outstation - Simulate Data                        [_][□][X]│
├─────────────────────────────────────────────────────────────────┤
│  File  Edit  View  Session  Help                                  │
├─────────────────────────────┬─────────────────────────────────────┤
│                             │                                      │
│  Server                    │  Data Points                         │
│  ┌───────────────────────┐ │  ┌─────────────────────────────┐   │
│  │ Port: [20000        ] │ │  │ Binary Inputs (8)          │   │
│  │       [Start Server ] │ │  │ ├─ BI0: true               │   │
│  │       [Stop Server  ] │ │  │ ├─ BI1: false              │   │
│  └───────────────────────┘ │  │ └─ ...                     │   │
│                             │  │                             │   │
│  Simulation                 │  │ Analog Inputs (4)          │   │
│  ┌───────────────────────┐ │  │ ├─ AI0: 100.5              │   │
│  │ [✓] Enable Simulation│ │  │ └─ ...                     │   │
│  │ Update Rate: [1.0]s  │ │  │                             │   │
│  │ Binary Rate: [0.5]Hz │ │  │ Counters (4)              │   │
│  │ Analog Var: [±10.0] │ │  │ ├─ C0: 1000                │   │
│  └───────────────────────┘ │  │ └─ ...                     │   │
│                             │  └─────────────────────────────┘   │
├─────────────────────────────┴─────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ Log: [10:23:01] ← INCOMING: Master connected from 127.0.0.1│   │
│  │       [10:23:02] → SENT: Response to Read Class 0          │   │
│  └─────────────────────────────────────────────────────────────┘   │
├───────────────────────────────────────────────────────────────────┤
│  ● Listening │ Port: 20000 │ Masters: 0                           │
└───────────────────────────────────────────────────────────────────┘
```

---

## 3. File Changes

### 3.1 New Files

| File | Description |
|------|-------------|
| `cmd/workbench/main.go` | Updated entry point with mode selection |
| `cmd/workbench/internal/master/controller.go` | Master-specific controller |
| `cmd/workbench/internal/outstation/controller.go` | Outstation-specific controller |

### 3.2 Modified Files

| File | Change |
|------|--------|
| `cmd/workbench/main.go` | Add flag parsing, mode selection |
| `cmd/workbench/internal/controller/controller.go` | Refactor to support both modes |

### 3.3 File Structure

```
cmd/workbench/
├── main.go                    # Entry point with --mode flag
├── internal/
│   ├── controller/
│   │   └── controller.go     # Unified controller (refactored)
│   ├── master/
│   │   └── controller.go     # Master-specific operations
│   ├── outstation/
│   │   └── controller.go     # Outstation-specific operations
│   └── session/
│       ├── session.go        # Master session
│       └── outstation.go     # Outstation session
```

---

## 4. Implementation Tasks

### 4.1 Task List

| # | Task | Priority | Status |
|---|------|----------|--------|
| 1 | Add command-line flag parsing | HIGH | TODO |
| 2 | Create mode selection dialog | HIGH | TODO |
| 3 | Create Master controller | HIGH | TODO |
| 4 | Create Outstation controller | HIGH | TODO |
| 5 | Refactor main entry point | HIGH | TODO |
| 6 | Update menu structure per mode | MEDIUM | TODO |
| 7 | Update status bar per mode | MEDIUM | TODO |
| 8 | Test Master instance | HIGH | TODO |
| 9 | Test Outstation instance | HIGH | TODO |
| 10 | Test parallel execution | MEDIUM | TODO |

### 4.2 Implementation Sequence

```
1. Add flag parsing to main.go
2. Create mode selection dialog
3. Create master.Controller
4. Create outstation.Controller
5. Refactor main() to dispatch based on mode
6. Add mode-specific window layouts
7. Test both modes independently
8. Test parallel execution
```

---

## 5. Testing Specification

### 5.1 Unit Tests

| Test | Description |
|------|-------------|
| `TestModeFlag_Master` | Verify --mode master launches Master |
| `TestModeFlag_Outstation` | Verify --mode outstation launches Outstation |
| `TestModeFlag_Select` | Verify no flag shows selection dialog |
| `TestModeFlag_Invalid` | Verify invalid mode shows error |

### 5.2 Integration Tests

| Test | Description |
|------|-------------|
| `TestMasterConnect` | Master connects to outstation |
| `TestOutstationListen` | Outstation accepts connections |
| `TestParallelExecution` | Both instances run simultaneously |

### 5.3 Manual Test Plan

```bash
# Test 1: Master mode
./workbench --mode master
# Expected: Master window opens

# Test 2: Outstation mode
./workbench --mode outstation
# Expected: Outstation window opens

# Test 3: Parallel execution
./workbench --mode master &
./workbench --mode outstation &
# Expected: Both windows open

# Test 4: Default mode
./workbench
# Expected: Mode selection dialog opens
```

---

## 6. Backward Compatibility

### 6.1 Breaking Changes

None. The default behavior (mode selection) maintains existing workflow.

### 6.2 Migration Path

| Old Behavior | New Behavior | Compatibility |
|--------------|--------------|---------------|
| Default to Master | Default to mode selection | Enhanced |
| No flag support | --mode flag | Added |
| Single window | Dedicated windows per mode | Enhanced |

---

## 7. Specification Summary

### 7.1 Core Changes

1. **Single executable** with `--mode` flag
2. **Dedicated instances** - Master OR Outstation, not both
3. **Mode selection dialog** for default behavior
4. **Separate window layouts** per mode

### 7.2 User Impact

| Aspect | Impact |
|--------|--------|
| Startup | New mode selection dialog |
| Operation | Dedicated window per mode |
| Parallel use | Supported via two instances |

---

*Specification created by DNP3-INV-062 Investigation*
*Status: PENDING APPROVAL*
