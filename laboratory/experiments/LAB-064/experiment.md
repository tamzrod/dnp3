# LAB-064: Diagnose Master Not Receiving Data

**Engine**: KDE-ENGINE-001 (v1.0.0)  
**Seed**: SEED-001 (v1.0.0)  
**Timestamp**: 2026-07-29T08:36:00Z  
**Status**: 🔬 IN_PROGRESS

## Experiment Goal

Diagnose why master binary shows "Disconnected" and doesn't receive data.

## Evidence

### Finding 1: TCP No Listeners
```
$ ss -tlnp | grep 20004
→ No output (no listeners found)
```

### Finding 2: Outstation Running But Simulator Only
```
Outstation shows simulated data (BI, AI, CTR)
But no actual TCP server is listening
```

### Finding 3: TUI Bug - Connection Status Race Condition
In `main.go`:
```go
app.OnStart = func() {
    ctrl.Connect(address, port)  // Returns immediately!
    app.SetConnection("Connected")  // WRONG - connection is async
}
```

In `controller.go`:
```go
func (c *Controller) Connect(...) error {
    go func() {  // Connection happens in goroutine
        // ... actual connection ...
    }()
    return nil  // Returns immediately!
}
```

## Root Cause Hypothesis

**H1: Workbench TUI doesn't wait for async connection** (HIGH probability)
- Controller.Connect() starts goroutine and returns immediately
- TUI sets "Connected" status immediately without waiting
- Actual connection may fail silently

**H2: Outstation server not accepting connections** (MEDIUM)
- Need to verify TCP listener is actually created

## Next Action

Create experiment LAB-065 to test with properly wired controller (no TUI).
