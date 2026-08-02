// Package outstation provides Outstation-specific functionality for the DNP3 workbench.
package outstation

import (
	"fmt"
	"sync"
	"time"

	"dnp3/cmd/workbench/internal/logger"
	"dnp3/cmd/workbench/internal/session"
	"dnp3/cmd/workbench/internal/simulation"
	"dnp3/pkg/dnp3/types"
)

// State represents the current state of the Outstation controller.
type State struct {
	Running           bool
	ListenAddress     string
	ListenPort        int
	ConnectedMasters  int
	SimulationEnabled bool
	UpdateRate        time.Duration
	LastError         string
}

// Controller handles Outstation-specific operations.
type Controller struct {
	mu        sync.RWMutex
	session   *session.OutstationSession
	simulator *simulation.Simulator
	logger    *logger.Logger
	state     *State
}

// NewController creates a new Outstation controller.
func NewController(log *logger.Logger) *Controller {
	// Create simulator with default configuration
	sim := simulation.NewSimulator(nil)
	sim.AddDefaultPoints()

	return &Controller{
		logger: log,
		state: &State{
			ListenAddress:     "0.0.0.0",
			ListenPort:        20000,
			SimulationEnabled:  true,
			UpdateRate:        time.Second,
		},
		simulator: sim,
	}
}

// Start initializes the controller (starts simulation if enabled).
func (c *Controller) Start() error {
	c.logger.Info("Outstation controller started")
	if c.state.SimulationEnabled {
		c.simulator.Start()
	}
	return nil
}

// Stop shuts down the controller.
func (c *Controller) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.Info("Outstation controller stopping")

	// Stop session if running
	if c.session != nil {
		c.session.Close()
		c.session = nil
	}

	c.state.Running = false
	return nil
}

// StartServer begins listening for master connections.
func (c *Controller) StartServer(address string, port int) error {
	c.mu.Lock()

	if c.state.Running {
		c.mu.Unlock()
		return nil
	}

	c.state.ListenAddress = address
	c.state.ListenPort = port
	c.state.Running = true
	c.mu.Unlock()

	c.logger.Info("Outstation starting server on %s:%d", address, port)

	// Create outstation session
	outstationSession, err := session.NewOutstationSession(c.logger)
	if err != nil {
		c.handleError("Failed to create session: %v", err)
		c.mu.Lock()
		c.state.Running = false
		c.mu.Unlock()
		return fmt.Errorf("create session: %w", err)
	}

	// Set the simulator on the session
	outstationSession.SetSimulator(c.simulator)

	// Start session
	if err := outstationSession.Start(address, port); err != nil {
		c.handleError("Failed to start server: %v", err)
		c.mu.Lock()
		c.state.Running = false
		c.state.LastError = err.Error()
		c.mu.Unlock()
		return fmt.Errorf("start server: %w", err)
	}

	c.mu.Lock()
	c.session = outstationSession
	c.state.LastError = ""
	c.mu.Unlock()

	// Start simulation if enabled
	if c.state.SimulationEnabled {
		c.simulator.Start()
		c.logger.Info("Outstation simulation started")
	}

	c.logger.Info("Outstation listening on %s:%d", address, port)

	// Pump events in background
	go c.pumpEvents()

	return nil
}

// SetSimulationEnabled enables or disables simulation.
func (c *Controller) SetSimulationEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.state.SimulationEnabled = enabled

	if enabled && c.state.Running {
		c.simulator.Start()
		c.logger.Info("Simulation enabled")
	} else if !enabled && c.simulator != nil {
		c.simulator.Stop()
		c.logger.Info("Simulation disabled")
	}
}

// SetUpdateRate sets the simulation update rate.
func (c *Controller) SetUpdateRate(rate time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.state.UpdateRate = rate
	c.logger.Info("Simulation update rate set to %v", rate)
}

// GetBinaryInputs returns current binary inputs.
func (c *Controller) GetBinaryInputs() []*types.BinaryInput {
	return c.simulator.GetBinaryInputs()
}

// GetAnalogInputs returns current analog inputs.
func (c *Controller) GetAnalogInputs() []*types.AnalogInput {
	return c.simulator.GetAnalogInputs()
}

// GetCounters returns current counters.
func (c *Controller) GetCounters() []*types.Counter {
	return c.simulator.GetCounters()
}

// GetBinaryOutputs returns current binary outputs.
func (c *Controller) GetBinaryOutputs() []*types.BinaryOutput {
	return c.simulator.GetBinaryOutputs()
}

// GetAnalogOutputs returns current analog outputs.
func (c *Controller) GetAnalogOutputs() []*types.AnalogOutput {
	return c.simulator.GetAnalogOutputs()
}

// SetBinaryInput manually sets a binary input.
func (c *Controller) SetBinaryInput(index uint16, value bool) {
	c.simulator.SetBinaryInput(index, value)
}

// SetAnalogInput manually sets an analog input.
func (c *Controller) SetAnalogInput(index uint16, value float64) {
	c.simulator.SetAnalogInput(index, value)
}

// State returns the current controller state.
func (c *Controller) State() *State {
	c.mu.RLock()
	defer c.mu.RUnlock()
	state := *c.state
	return &state
}

// Logger returns the controller logger.
func (c *Controller) Logger() *logger.Logger {
	return c.logger
}

// pumpEvents handles session events.
func (c *Controller) pumpEvents() {
	c.mu.RLock()
	s := c.session
	c.mu.RUnlock()

	if s == nil {
		return
	}

	for event := range s.Events() {
		switch event.Type {
		case "started":
			c.logger.Info("Outstation server started")
		case "stopped":
			c.logger.Info("Outstation server stopped")
			c.mu.Lock()
			c.state.Running = false
			c.mu.Unlock()
			return
		case "master_connected":
			c.mu.Lock()
			c.state.ConnectedMasters++
			c.mu.Unlock()
			c.logger.Info("Master connected, total: %d", c.state.ConnectedMasters)
		case "master_disconnected":
			c.mu.Lock()
			if c.state.ConnectedMasters > 0 {
				c.state.ConnectedMasters--
			}
			c.mu.Unlock()
			c.logger.Info("Master disconnected, remaining: %d", c.state.ConnectedMasters)
		}
	}
}

// handleError logs and stores an error.
func (c *Controller) handleError(format string, args ...interface{}) {
	errMsg := fmt.Sprintf(format, args...)
	c.logger.Error("Outstation error: %s", errMsg)

	c.mu.Lock()
	c.state.LastError = errMsg
	c.mu.Unlock()
}
