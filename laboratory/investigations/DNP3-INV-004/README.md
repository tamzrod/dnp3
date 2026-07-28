# DNP3-INV-004: Master/Outstation Separation Analysis

**Investigation**: DNP3-INV-004
**Title**: Improve Master/Outstation Mode Separation in Workbench
**Date**: 2026-07-27
**Status**: IN PROGRESS

---

## 1. Problem Statement

The current DNP3 Workbench uses a single-window approach with mode switching between "Poll Outstation" (Master) and "Simulate Outstation" (Outstation). This approach has several issues:

1. **Confusing UI**: Users must understand both Master and Outstation concepts in one interface
2. **Hidden Functionality**: Outstation mode has limited UI exposure
3. **Inconsistent State**: Controller only implements Master mode fully
4. **Poor UX**: Mode switching is buried in sidebar rather than being a primary choice

## 2. Current State Analysis

### 2.1 Session Types

| Session Type | Implementation | UI Panel | Status |
|-------------|----------------|----------|--------|
| Master | `session.NewMasterSession()` | Connection, Commands, Data Table | ✅ Complete |
| Outstation | `session.NewOutstationSession()` | Mode Panel (minimal) | ⚠️ Partial |

### 2.2 Controller Analysis

```go
// Current controller only supports Master mode
func (c *Controller) Connect(address string, port int) error {
    // Always creates MasterSession
    masterSession, err := session.NewMasterSession(c.logger)
    masterSession.Connect(ctx, address, port)
}
```

### 2.3 Current Layout

```
┌────────────────────────────────────────────┐
│  Mode Panel (radio buttons)                │
│  (•) Poll Outstation  ( ) Simulate       │
├────────────────────────────────────────────┤
│  Connection Panel  │  Data Table           │
│  Commands Panel    │  Control Panel        │
└────────────────────────────────────────────┘
```

## 3. Proposed Improvements

### 3.1 Option A: Dual Window Approach (Recommended)

Separate windows for Master and Outstation modes:

```
┌──────────────────────┐     ┌──────────────────────┐
│   MASTER WINDOW      │     │  OUTSTATION WINDOW   │
├──────────────────────┤     ├──────────────────────┤
│ Connection Config    │     │ Server Config        │
│ Address:Port        │     │ Listen Port: 20000   │
│ [Connect] [Disconnect]    │ [Start] [Stop]       │
├──────────────────────┤     ├──────────────────────┤
│ Read Commands       │     │ Data Points          │
│ [Class 0] [Class 1]│     │ Binary Inputs       │
│ [Class 2] [Class 3]│     │ Analog Inputs       │
│                      │     │ Counters            │
├──────────────────────┤     ├──────────────────────┤
│ Data Table          │     │ Simulation Config    │
│ (read-only data)    │     │ Update Rate: 1s     │
│                      │     │ Random Data: [x]    │
└──────────────────────┘     └──────────────────────┘
```

**Pros:**
- Clear separation of concerns
- Each mode has dedicated UI
- Easier to understand and use
- Natural workflow (use one or the other)

**Cons:**
- Requires window management
- More complex to implement
- Can't run both simultaneously in same app

### 3.2 Option B: Tabbed Interface

Single window with tabs for Master/Outstation:

```
┌────────────────────────────────────────────┐
│  [Master] [Outstation]                     │ ← Tab bar
├────────────────────────────────────────────┤
│  Master Tab Active                         │
│  Connection Config | Data Table            │
└────────────────────────────────────────────┘
```

**Pros:**
- Single window
- Easy to switch modes
- Common elements can be shared

**Cons:**
- Can be confusing which mode is active
- Still mixes concepts in single view

### 3.3 Option C: Side-by-Side Panel Toggle

Expand current mode panel to show/hide entire mode-specific sections:

```
┌────────────────────────────────────────────┐
│  MASTER MODE                    [Switch]    │ ← Mode header
├────────────────────────────────────────────┤
│  Connection Config                          │
│  Commands (Master-specific)                 │
│  Data Table (read-only)                    │
└────────────────────────────────────────────┘
              ↓ Switch ↓
┌────────────────────────────────────────────┐
│  OUTSTATION MODE               [Switch]    │
├────────────────────────────────────────────┤
│  Server Config                              │
│  Data Points (configurable)                │
│  Simulation Controls                       │
└────────────────────────────────────────────┘
```

**Pros:**
- Single window
- Clear visual separation
- Easy to implement
- Maintains single-instance workflow

**Cons:**
- Both modes can't be visible at once
- Some state lost on switch

### 3.4 Option D: Split View (Dual Pane)

Both modes visible simultaneously:

```
┌────────────────────────┬────────────────────────┐
│     MASTER PANE        │   OUTSTATION PANE      │
├────────────────────────┼────────────────────────┤
│ Connection Config     │ Server Config          │
│ [Connect]             │ [Start]                │
├────────────────────────┼────────────────────────┤
│ Data Table (read)     │ Data Points (config)   │
│                       │ Simulation Controls    │
└────────────────────────┴────────────────────────┘
```

**Pros:**
- Both modes visible
- Can run simultaneously
- Compare master reads vs outstation data

**Cons:**
- Complex UI
- Requires more screen space
- Overwhelming for new users

## 4. Recommended Solution

**Recommendation: Option A - Dual Window Approach**

Rationale:
1. **Clarity**: Each window has exactly one purpose
2. **Simplicity**: Users only see relevant controls
3. **Usability**: Matches mental model (Master vs Outstation)
4. **Implementation**: Can start app with mode selection or use command-line flag

### Implementation Strategy

```go
// main.go
func main() {
    mode := flag.String("mode", "select", "master|outstation|select")
    
    switch *mode {
    case "master":
        runMasterOnly()
    case "outstation":
        runOutstationOnly()
    case "select":
        showModeSelection()
    }
}
```

### New Window Structure

```
cmd/workbench/
├── internal/
│   ├── master/
│   │   ├── controller.go     # Master-specific controller
│   │   └── window.go        # Master window
│   ├── outstation/
│   │   ├── controller.go     # Outstation-specific controller
│   │   └── window.go        # Outstation window
│   └── shared/
│       ├── components/      # Shared UI components
│       └── types.go        # Shared types
```

## 5. Implementation Tasks

| Task | Description | Priority |
|------|-------------|----------|
| 5.1 | Create separate Master controller | High |
| 5.2 | Create separate Outstation controller | High |
| 5.3 | Create Master window/layout | High |
| 5.4 | Create Outstation window/layout | High |
| 5.5 | Add command-line mode selection | Medium |
| 5.6 | Add mode selection splash screen | Medium |
| 5.7 | Update menu structure | Low |
| 5.8 | Update documentation | Low |

## 6. Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Breaking existing functionality | High | Maintain backwards compatibility via flag |
| Increased code complexity | Medium | Use shared components |
| Window state management | Low | Fyne handles most window state |

## 7. Next Steps

1. **Decision**: Confirm recommended approach (Option A)
2. **Planning**: Break down implementation into PRs
3. **Phase 1**: Create separate controllers ✅ COMPLETED
4. **Phase 2**: Create separate window layouts
5. **Phase 3**: Add mode selection
6. **Phase 4**: Testing and documentation

## 8. Implementation Status

### Phase 1: Controllers ✅ COMPLETED

| File | Status |
|------|--------|
| `cmd/workbench/internal/master/controller.go` | ✅ Created |
| `cmd/workbench/internal/outstation/controller.go` | ✅ Created |
| `session.OutstationSession.SetSimulator()` | ✅ Added |

### Phase 2: Windows - PENDING

| File | Status |
|------|--------|
| `cmd/workbench/internal/master/window.go` | TODO |
| `cmd/workbench/internal/outstation/window.go` | TODO |
| `cmd/workbench/internal/shared/ui/` | TODO |

### Phase 3: Mode Selection - PENDING

| File | Status |
|------|--------|
| Mode selection dialog | TODO |
| Command-line flag | TODO |
| Main entry update | TODO |

---

*Investigation initiated: 2026-07-27*
*Last updated: 2026-07-27*
