// Package simulation provides random data simulation for DNP3 outstation testing.
package simulation

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"dnp3/pkg/dnp3/types"
)

// Config holds simulation configuration
type Config struct {
	// BinaryInputUpdateRate is the average flips per second for binary inputs
	BinaryInputUpdateRate float64
	// AnalogInputVariance is the max change per tick for analog inputs
	AnalogInputVariance float64
	// CounterIncrementRate is the probability of counter increment per tick
	CounterIncrementRate float64
	// CounterIncrementAmount is the average increment amount
	CounterIncrementAmount uint32
	// TickInterval is how often data is updated
	TickInterval time.Duration
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() *Config {
	return &Config{
		BinaryInputUpdateRate:  0.3,  // 0.3 flips per second (slower)
		AnalogInputVariance:    5.0,  // ±5 units per tick
		CounterIncrementRate:   1.0,  // 100% chance per tick (always increment)
		CounterIncrementAmount:  1,    // Increment by 1
		TickInterval:           500 * time.Millisecond,
	}
}

// SimulatedPoint represents a point being simulated
type SimulatedPoint interface {
	Update(cfg *Config)
}

// BinaryInputSimulation simulates a binary input point
type BinaryInputSimulation struct {
	Index       uint16
	Value       bool
	Quality     types.QualityFlags
	Time        *types.Timestamp
	FlipTimer   float64 // Time until next flip
	UpdateRate  float64 // Flips per second
}

// NewBinaryInputSimulation creates a new binary input simulation
func NewBinaryInputSimulation(index uint16, initialValue bool, updateRate float64) *BinaryInputSimulation {
	return &BinaryInputSimulation{
		Index:      index,
		Value:      initialValue,
		Quality:    types.QualityOnline,
		Time:       (&types.Timestamp{}).Now(),
		FlipTimer:  rand.Float64() / updateRate, // Random initial timer
		UpdateRate: updateRate,
	}
}

// Update updates the binary input state
func (b *BinaryInputSimulation) Update(cfg *Config) {
	b.FlipTimer -= cfg.TickInterval.Seconds()
	if b.FlipTimer <= 0 {
		b.Value = !b.Value
		b.Time = (&types.Timestamp{}).Now()
		// Set timer for next flip (average 2 seconds per flip)
		b.FlipTimer = rand.ExpFloat64() / b.UpdateRate
	}
}

// ToBinaryInput converts to a BinaryInput type
func (b *BinaryInputSimulation) ToBinaryInput() *types.BinaryInput {
	return &types.BinaryInput{
		Index:   b.Index,
		Value:   b.Value,
		Quality: b.Quality,
		Time:    b.Time,
	}
}

// AnalogInputSimulation simulates an analog input point
type AnalogInputSimulation struct {
	Index     uint16
	Value     float64
	MinValue  float64
	MaxValue  float64
	Variance  float64
	Quality   types.QualityFlags
	Time      *types.Timestamp
}

// NewAnalogInputSimulation creates a new analog input simulation
func NewAnalogInputSimulation(index uint16, initialValue, minValue, maxValue float64) *AnalogInputSimulation {
	return &AnalogInputSimulation{
		Index:    index,
		Value:    initialValue,
		MinValue: minValue,
		MaxValue: maxValue,
		Variance: 0, // Set variance from config
		Quality:  types.QualityOnline,
		Time:     (&types.Timestamp{}).Now(),
	}
}

// Update updates the analog input state with random drift
func (a *AnalogInputSimulation) Update(cfg *Config) {
	// Random drift within variance
	drift := (rand.Float64()*2 - 1) * cfg.AnalogInputVariance
	newValue := a.Value + drift
	
	// Clamp to range
	newValue = math.Max(a.MinValue, math.Min(a.MaxValue, newValue))
	
	// Always update with small variation
	a.Value = math.Round(newValue*100) / 100 // Round to 2 decimal places
	a.Time = (&types.Timestamp{}).Now()
}

// ToAnalogInput converts to an AnalogInput type
func (a *AnalogInputSimulation) ToAnalogInput() *types.AnalogInput {
	return &types.AnalogInput{
		Index:   a.Index,
		Value:   a.Value,
		Quality: a.Quality,
		Time:    a.Time,
	}
}

// CounterSimulation simulates a counter point
type CounterSimulation struct {
	Index    uint16
	Value    uint32
	Quality  types.QualityFlags
	Time     *types.Timestamp
}

// NewCounterSimulation creates a new counter simulation
func NewCounterSimulation(index uint16, initialValue uint32) *CounterSimulation {
	return &CounterSimulation{
		Index:   index,
		Value:   initialValue,
		Quality: types.QualityOnline,
		Time:    (&types.Timestamp{}).Now(),
	}
}

// Update increments the counter
func (c *CounterSimulation) Update(cfg *Config) {
	// Always increment the counter
	c.Value += cfg.CounterIncrementAmount
	c.Time = (&types.Timestamp{}).Now()
}

// ToCounter converts to a Counter type
func (c *CounterSimulation) ToCounter() *types.Counter {
	return &types.Counter{
		Index:   c.Index,
		Value:   c.Value,
		Quality: c.Quality,
		Time:    c.Time,
	}
}

// Simulator manages all simulated data points
type Simulator struct {
	mu      sync.RWMutex
	config  *Config
	stopCh  chan struct{}
	running bool

	BinaryInputs []*BinaryInputSimulation
	AnalogInputs []*AnalogInputSimulation
	Counters     []*CounterSimulation
}

// NewSimulator creates a new simulator with default configuration
func NewSimulator(cfg *Config) *Simulator {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Simulator{
		config:       cfg,
		stopCh:       make(chan struct{}),
		BinaryInputs: make([]*BinaryInputSimulation, 0),
		AnalogInputs: make([]*AnalogInputSimulation, 0),
		Counters:    make([]*CounterSimulation, 0),
	}
}

// AddDefaultPoints adds a set of default simulated points
func (s *Simulator) AddDefaultPoints() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Add 8 binary inputs
	for i := 0; i < 8; i++ {
		s.BinaryInputs = append(s.BinaryInputs, NewBinaryInputSimulation(
			uint16(i),
			i%2 == 0, // Alternate initial values
			s.config.BinaryInputUpdateRate,
		))
	}

	// Add 4 analog inputs
	analogConfigs := []struct {
		initial, min, max float64
	}{
		{100.0, 0.0, 200.0},    // Temperature-like
		{50.0, -50.0, 150.0},    // Pressure-like
		{25.0, 0.0, 100.0},     // Flow rate-like
		{0.0, -100.0, 100.0},    // Bidirectional
	}
	for i, cfg := range analogConfigs {
		sim := NewAnalogInputSimulation(uint16(i), cfg.initial, cfg.min, cfg.max)
		sim.Variance = s.config.AnalogInputVariance
		s.AnalogInputs = append(s.AnalogInputs, sim)
	}

	// Add 4 counters starting at 0
	for i := 0; i < 4; i++ {
		s.Counters = append(s.Counters, NewCounterSimulation(uint16(i), 0))
	}
}

// Start begins the simulation loop
func (s *Simulator) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.stopCh = make(chan struct{})
	s.running = true
	s.mu.Unlock()
	go s.runLoop()
}

// Stop ends the simulation loop
func (s *Simulator) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !s.running {
		return
	}
	s.running = false
	close(s.stopCh)
}

func (s *Simulator) runLoop() {
	ticker := time.NewTicker(s.config.TickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.tick()
		}
	}
}

func (s *Simulator) tick() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Update all simulated points
	for _, bi := range s.BinaryInputs {
		bi.Update(s.config)
	}
	for _, ai := range s.AnalogInputs {
		ai.Update(s.config)
	}
	for _, c := range s.Counters {
		c.Update(s.config)
	}
}

// GetBinaryInputs returns current binary inputs
func (s *Simulator) GetBinaryInputs() []*types.BinaryInput {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*types.BinaryInput, len(s.BinaryInputs))
	for i, bi := range s.BinaryInputs {
		result[i] = bi.ToBinaryInput()
	}
	return result
}

// GetAnalogInputs returns current analog inputs
func (s *Simulator) GetAnalogInputs() []*types.AnalogInput {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*types.AnalogInput, len(s.AnalogInputs))
	for i, ai := range s.AnalogInputs {
		result[i] = ai.ToAnalogInput()
	}
	return result
}

// GetCounters returns current counters
func (s *Simulator) GetCounters() []*types.Counter {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*types.Counter, len(s.Counters))
	for i, c := range s.Counters {
		result[i] = c.ToCounter()
	}
	return result
}

// GetFrozenCounters returns empty frozen counters
func (s *Simulator) GetFrozenCounters() []*types.Counter {
	return []*types.Counter{}
}

// FreezeCounters handles counter freeze (no-op for simulation)
func (s *Simulator) FreezeCounters(clear bool) error {
	return nil
}

// SetBinaryInput manually sets a binary input value
func (s *Simulator) SetBinaryInput(index uint16, value bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, bi := range s.BinaryInputs {
		if bi.Index == index {
			bi.Value = value
			bi.Time = (&types.Timestamp{}).Now()
			return
		}
	}
}

// SetAnalogInput manually sets an analog input value
func (s *Simulator) SetAnalogInput(index uint16, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, ai := range s.AnalogInputs {
		if ai.Index == index {
			ai.Value = value
			ai.Time = (&types.Timestamp{}).Now()
			return
		}
	}
}
