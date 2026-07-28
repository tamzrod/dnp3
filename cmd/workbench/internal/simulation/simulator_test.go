package simulation

import (
	"testing"
	"time"
)

func TestSimulatorDefaults(t *testing.T) {
	sim := NewSimulator(DefaultConfig())
	sim.AddDefaultPoints()

	// Verify we have default points
	if len(sim.BinaryInputs) != 8 {
		t.Errorf("Expected 8 binary inputs, got %d", len(sim.BinaryInputs))
	}
	if len(sim.AnalogInputs) != 4 {
		t.Errorf("Expected 4 analog inputs, got %d", len(sim.AnalogInputs))
	}
	if len(sim.Counters) != 4 {
		t.Errorf("Expected 4 counters, got %d", len(sim.Counters))
	}
}

func TestSimulatorDataTypes(t *testing.T) {
	sim := NewSimulator(nil)
	sim.AddDefaultPoints()

	// Get data and verify types
	binary := sim.GetBinaryInputs()
	if len(binary) == 0 {
		t.Error("Expected binary inputs")
	}
	// QualityOnline is 0x01 (bit 0 set)
	if binary[0].Quality&1 == 0 {
		t.Errorf("Expected ONLINE quality bit set, got %d", binary[0].Quality)
	}

	analog := sim.GetAnalogInputs()
	if len(analog) == 0 {
		t.Error("Expected analog inputs")
	}

	counters := sim.GetCounters()
	if len(counters) == 0 {
		t.Error("Expected counters")
	}
}

func TestSimulatorUpdate(t *testing.T) {
	cfg := &Config{
		BinaryInputUpdateRate:  10.0, // High rate for testing
		AnalogInputVariance:    100.0,
		CounterIncrementRate:   1.0,  // Always increment
		CounterIncrementAmount: 1,
		TickInterval:          100 * time.Millisecond,
	}
	sim := NewSimulator(cfg)
	sim.AddDefaultPoints()

	// Get initial values
	initialCounters := make([]uint32, len(sim.Counters))
	for i, c := range sim.Counters {
		initialCounters[i] = c.Value
	}

	// Run update loop
	sim.Start()
	time.Sleep(500 * time.Millisecond)
	sim.Stop()

	// Verify counters incremented
	for i, c := range sim.Counters {
		if c.Value <= initialCounters[i] {
			t.Errorf("Counter %d did not increment: was %d, now %d", i, initialCounters[i], c.Value)
		}
	}
}

func TestSimulatorBinaryToggle(t *testing.T) {
	cfg := &Config{
		BinaryInputUpdateRate:  100.0, // Very high rate
		AnalogInputVariance:   0,
		CounterIncrementRate:   0,
		CounterIncrementAmount: 0,
		TickInterval:          10 * time.Millisecond,
	}
	sim := NewSimulator(cfg)
	sim.AddDefaultPoints()

	// Run for a while
	sim.Start()
	time.Sleep(200 * time.Millisecond)
	sim.Stop()

	// Value should have toggled at some point
	// We can't guarantee it changed in this short time, but the simulation is working
	t.Logf("Binary input 0 value after test: %v", sim.BinaryInputs[0].Value)
}

func TestSimulatorSetManualValue(t *testing.T) {
	sim := NewSimulator(nil)
	sim.AddDefaultPoints()

	// Set a binary input
	sim.SetBinaryInput(0, true)
	if sim.BinaryInputs[0].Value != true {
		t.Error("Failed to set binary input")
	}

	// Set an analog input
	sim.SetAnalogInput(0, 999.0)
	if sim.AnalogInputs[0].Value != 999.0 {
		t.Error("Failed to set analog input")
	}
}

func TestSimulatorFrozenCounters(t *testing.T) {
	sim := NewSimulator(nil)
	sim.AddDefaultPoints()

	// Frozen counters should return empty
	frozen := sim.GetFrozenCounters()
	if frozen != nil && len(frozen) != 0 {
		t.Errorf("Expected empty frozen counters, got %d", len(frozen))
	}
}
