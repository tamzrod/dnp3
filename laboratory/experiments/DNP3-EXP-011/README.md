# DNP3-EXP-011: Random Data Simulation for Outstation

**Experiment**: DNP3-EXP-011
**Title**: Random Data Simulation
**Investigation**: DNP3-INV-003
**Status**: COMPLETED
**Date**: 2026-07-27

---

## 1. Purpose

Implement random data simulation for the DNP3 outstation to provide moving data points that can be read by a master.

## 2. Implementation

### New Files
- `cmd/workbench/internal/simulation/simulator.go` - Random data simulation module

### Simulated Data Types

#### Binary Inputs (8 points)
- Toggle randomly at configurable rate (default: 0.5 Hz)
- Exponential distribution for realistic timing
- Quality flags always set to ONLINE

#### Analog Inputs (4 points)
| Index | Initial | Range | Description |
|-------|---------|-------|-------------|
| 0 | 100.0 | 0-200 | Temperature-like |
| 1 | 50.0 | -50-150 | Pressure-like |
| 2 | 25.0 | 0-100 | Flow rate-like |
| 3 | 0.0 | -100-100 | Bidirectional |

- Random drift within configurable variance (default: ±10 per tick)
- Clamped to min/max range
- Quality flags always set to ONLINE

#### Counters (4 points)
- Increment with configurable probability (default: 10% per tick)
- Increment amount configurable (default: 1)
- Quality flags always set to ONLINE

### Configuration

```go
type Config struct {
    BinaryInputUpdateRate  float64     // Flips per second (0.5)
    AnalogInputVariance    float64     // Max change per tick (±10)
    CounterIncrementRate   float64     // Probability per tick (0.1)
    CounterIncrementAmount uint32      // Increment amount (1)
    TickInterval           time.Duration // Update frequency (1s)
}
```

## 3. Usage

```go
// Create simulator with default config
sim := simulation.NewSimulator(nil)

// Add default points
sim.AddDefaultPoints()

// Start simulation loop
sim.Start()

// Get current values
binaryInputs := sim.GetBinaryInputs()
analogInputs := sim.GetAnalogInputs()
counters := sim.GetCounters()

// Stop when done
sim.Stop()
```

## 4. Build Verification

```
$ go build ./cmd/workbench/internal/simulation/...
# Success - no errors
```

## 5. Evidence

- [x] Simulation module created
- [x] Binary input simulation implemented
- [x] Analog input simulation implemented
- [x] Counter simulation implemented
- [x] Configurable parameters
- [x] Build verified

## 6. Related Experiments

- DNP3-EXP-009: Master Real DNP3 Integration (COMPLETED)
- DNP3-EXP-010: Outstation Real DNP3 Integration (COMPLETED)
- DNP3-EXP-012: Integrate Simulation into Workbench (PENDING)
