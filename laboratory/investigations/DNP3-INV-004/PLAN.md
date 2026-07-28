# Implementation Plan: DNP3-INV-004

## Phase 1: Refactor Controllers

### 1.1 Create Master Controller

**File**: `cmd/workbench/internal/master/controller.go`

```go
package master

type Controller struct {
    session *session.MasterSession
    logger *logger.Logger
    state *State
}

func NewController(log *logger.Logger) *Controller
func (c *Controller) Connect(address string, port int) error
func (c *Controller) Disconnect() error
func (c *Controller) ReadClass(class int) error
func (c *Controller) Operate(index uint16, value interface{}) error
func (c *Controller) State() *State
```

### 1.2 Create Outstation Controller

**File**: `cmd/workbench/internal/outstation/controller.go`

```go
package outstation

type Controller struct {
    session *session.OutstationSession
    simulator *simulation.Simulator
    logger *logger.Logger
    state *State
}

func NewController(log *logger.Logger) *Controller
func (c *Controller) Start(address string, port int) error
func (c *Controller) Stop() error
func (c *Controller) SetSimulationEnabled(enabled bool)
func (c *Controller) SetUpdateRate(rate time.Duration)
func (c *Controller) State() *State
```

### 1.3 Extract Shared Types

**File**: `cmd/workbench/internal/shared/types.go`

```go
package shared

type State struct {
    Mode Mode
    Connection ConnectionState
    // ... common fields
}

type Mode string
const (
    ModeMaster Mode = "master"
    ModeOutstation Mode = "outstation"
)
```

## Phase 2: Create Windows

### 2.1 Master Window

**File**: `cmd/workbench/internal/master/window.go`

- ConnectionPanel
- CommandsPanel
- DataTablePanel
- ControlPanel
- LogPanel
- StatusBar

### 2.2 Outstation Window

**File**: `cmd/workbench/internal/outstation/window.go`

- ServerPanel
- DataPointsPanel
- SimulationPanel
- LogPanel
- StatusBar

### 2.3 Shared UI Components

**File**: `cmd/workbench/internal/shared/ui/`

- `log_panel.go` - Reusable log component
- `status_bar.go` - Shared status bar
- `data_table.go` - Shared table component

## Phase 3: Mode Selection

### 3.1 Update Main Entry Point

**File**: `cmd/workbench/main.go`

```go
func main() {
    mode := flag.String("mode", "select", 
        "Select operating mode: master, outstation, or select")
    
    switch *mode {
    case "master":
        runMaster()
    case "outstation":
        runOutstation()
    case "select":
        showModeSelection()
    }
}
```

### 3.2 Create Mode Selection Dialog

**File**: `cmd/workbench/internal/shared/ui/mode_select.go`

## Phase 4: Testing

### 4.1 Unit Tests
- Master controller tests
- Outstation controller tests

### 4.2 Integration Tests
- Master → Outstation communication
- Mode switching

---

## Tasks

| # | Task | Status |
|---|------|--------|
| 1 | Create master controller | TODO |
| 2 | Create outstation controller | TODO |
| 3 | Extract shared types | TODO |
| 4 | Create master window | TODO |
| 5 | Create outstation window | TODO |
| 6 | Create shared UI components | TODO |
| 7 | Update main entry point | TODO |
| 8 | Create mode selection dialog | TODO |
| 9 | Add command-line flag | TODO |
| 10 | Test integration | TODO |
