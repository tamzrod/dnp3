// Package master provides Master-specific functionality for the DNP3 workbench.
package master

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"dnp3/cmd/workbench/internal/logger"
	"dnp3/cmd/workbench/internal/session"
	"dnp3/pkg/dnp3/types"
)

// State represents the current state of the Master controller.
type State struct {
	Connection       session.ConnectionState
	Address          string
	Port             int
	Error            string
	LastResponse     *session.Response
	AutoPollEnabled  bool
	AutoWriteEnabled bool
	SimulationMode   bool
}

// Controller handles Master-specific operations.
type Controller struct {
	mu          sync.RWMutex
	session     *session.MasterSession
	logger      *logger.Logger
	state       *State
	autoPollCh  chan bool
	autoWriteCh chan bool
	autoPollCtx context.Context
	autoPollCancel context.CancelFunc
}

// NewController creates a new Master controller.
func NewController(log *logger.Logger) *Controller {
	ctx, cancel := context.WithCancel(context.Background())
	return &Controller{
		logger:        log,
		state: &State{
			Connection:        session.StateDisconnected,
			Address:           "127.0.0.1",
			Port:              20000,
			AutoPollEnabled:   false,
			AutoWriteEnabled:  false,
			SimulationMode:    false,
		},
		autoPollCh:    make(chan bool, 1),
		autoWriteCh:   make(chan bool, 1),
		autoPollCtx:   ctx,
		autoPollCancel: cancel,
	}
}

// Start initializes the controller.
func (c *Controller) Start() error {
	c.logger.Info("Master controller started")
	// Start auto-poll handler
	go c.handleAutoPoll()
	// Start auto-write handler
	go c.handleAutoWrite()
	return nil
}

// Stop shuts down the controller.
func (c *Controller) Stop() error {
	c.logger.Info("Master controller stopping")
	// Stop auto-poll
	c.setAutoPollEnabled(false)
	c.setAutoWriteEnabled(false)
	c.autoPollCancel()
	if c.session != nil {
		c.session.Close()
	}
	return nil
}

// EnableAutoPoll enables or disables auto-poll (1 second interval).
func (c *Controller) EnableAutoPoll(enabled bool) {
	c.mu.Lock()
	wasEnabled := c.state.AutoPollEnabled
	c.state.AutoPollEnabled = enabled
	if !enabled {
		c.state.SimulationMode = false
	}
	c.mu.Unlock()
	
	if enabled && !wasEnabled {
		c.logger.Info("Auto-poll ENABLED (1 second interval)")
	} else if !enabled && wasEnabled {
		c.logger.Info("Auto-poll DISABLED")
	}
	
	// Signal to auto-poll handler
	select {
	case c.autoPollCh <- enabled:
	default:
	}
}

// setAutoPollEnabled sets auto-poll state without logging (internal use).
func (c *Controller) setAutoPollEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.state.AutoPollEnabled = enabled
}

// IsAutoPollEnabled returns whether auto-poll is enabled.
func (c *Controller) IsAutoPollEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state.AutoPollEnabled
}

// EnableAutoWrite enables or disables auto-write (random operate).
func (c *Controller) EnableAutoWrite(enabled bool) {
	c.mu.Lock()
	wasEnabled := c.state.AutoWriteEnabled
	c.state.AutoWriteEnabled = enabled
	if !enabled {
		c.state.SimulationMode = false
	}
	c.mu.Unlock()
	
	if enabled && !wasEnabled {
		c.logger.Info("Auto-write ENABLED (random operate)")
	} else if !enabled && wasEnabled {
		c.logger.Info("Auto-write DISABLED")
	}
	
	// Signal to auto-write handler
	// Use non-blocking send with buffered channel; if full, handler already has latest state
	select {
	case c.autoWriteCh <- enabled:
	default:
		// Channel full, handler already has latest state - this is fine
	}
}

// setAutoWriteEnabled sets auto-write state without logging (internal use).
func (c *Controller) setAutoWriteEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.state.AutoWriteEnabled = enabled
}

// IsAutoWriteEnabled returns whether auto-write is enabled.
func (c *Controller) IsAutoWriteEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state.AutoWriteEnabled
}

// EnableSimulationMode enables or disables simulation mode (both auto-read and auto-write).
func (c *Controller) EnableSimulationMode(enabled bool) {
	c.mu.Lock()
	c.state.SimulationMode = enabled
	c.state.AutoPollEnabled = enabled
	c.state.AutoWriteEnabled = enabled
	c.mu.Unlock()
	
	if enabled {
		c.logger.Info("Simulation mode ENABLED")
	} else {
		c.logger.Info("Simulation mode DISABLED")
	}
	
	// Signal to handlers
	select {
	case c.autoPollCh <- enabled:
	default:
	}
	select {
	case c.autoWriteCh <- enabled:
	default:
	}
}

// IsSimulationModeEnabled returns whether simulation mode is enabled.
func (c *Controller) IsSimulationModeEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state.SimulationMode
}

// handleAutoPoll manages the auto-poll goroutine.
func (c *Controller) handleAutoPoll() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-c.autoPollCtx.Done():
			return
		case enabled := <-c.autoPollCh:
			if !enabled {
				// Drain the ticker channel if disabling
				for {
					select {
					case <-ticker.C:
					default:
						goto tickerReset
					}
				}
			tickerReset:
				ticker.Stop()
				ticker = time.NewTicker(1 * time.Second)
			}
		case <-ticker.C:
			c.mu.RLock()
			enabled := c.state.AutoPollEnabled
			s := c.session
			c.mu.RUnlock()
			
			if enabled && s != nil {
				c.doAutoRead()
			}
		}
	}
}

// doAutoRead performs a single auto-read of Class 0.
func (c *Controller) doAutoRead() {
	c.mu.RLock()
	s := c.session
	c.mu.RUnlock()
	
	if s == nil {
		return
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	cmd := &session.ReadCommand{
		Groups: []types.GroupRequest{
			{Group: 60, Variation: 1}, // All static data (Class 0)
		},
	}
	
	c.logger.Info("Auto-read: sending READ Class 0")
	resp, err := s.SendCommand(ctx, cmd)
	if err != nil {
		c.logger.Error("Auto-read failed: %v", err)
		return
	}
	
	c.handleResponse(resp)
	c.logger.Info("Auto-read: %d BI, %d AI, %d CTR, %d BO, %d AO",
		len(resp.BinaryInputs), len(resp.AnalogInputs), len(resp.Counters),
		len(resp.BinaryOutputs), len(resp.AnalogOutputs))
}

// handleAutoWrite manages the auto-write goroutine.
func (c *Controller) handleAutoWrite() {
	for {
		select {
		case <-c.autoPollCtx.Done():
			return
		case enabled := <-c.autoWriteCh:
			if !enabled {
				continue // Wait for enable signal
			}
			// Random interval between 1-3 seconds
			interval := time.Duration(1+rand.Intn(3)) * time.Second
			ticker := time.NewTicker(interval)
			
			autoWriteLoop:
			for {
				select {
				case <-c.autoPollCtx.Done():
					ticker.Stop()
					return
				case enabled := <-c.autoWriteCh:
					ticker.Stop()
					if !enabled {
						// Exit to outer loop to wait for next enable
						break autoWriteLoop
					}
					// Restart with new random interval
					interval = time.Duration(1+rand.Intn(3)) * time.Second
					ticker = time.NewTicker(interval)
					continue
				case <-ticker.C:
					c.mu.RLock()
					enabled := c.state.AutoWriteEnabled
					s := c.session
					c.mu.RUnlock()
					
					if enabled && s != nil {
						c.doRandomOperate()
					}
				}
			}
		}
	}
}

// doRandomOperate performs a random operate command.
func (c *Controller) doRandomOperate() {
	c.mu.RLock()
	s := c.session
	resp := c.state.LastResponse
	c.mu.RUnlock()
	
	if s == nil {
		return
	}
	
	// Determine available BO/AO indices from last response
	boCount := 2  // Default count
	aoCount := 2  // Default count
	
	if resp != nil {
		// BO indices come from BinaryOutputs if available
		if len(resp.BinaryOutputs) > 0 {
			boCount = len(resp.BinaryOutputs)
		}
		if len(resp.AnalogOutputs) > 0 {
			aoCount = len(resp.AnalogOutputs)
		}
	}
	
	// Randomly choose: 70% BO operate, 30% AO operate
	if rand.Float64() < 0.7 {
		// Binary Output operate
		index := uint16(rand.Intn(boCount))
		value := rand.Float64() < 0.5
		
		c.mu.RLock()
		c.logger.Info("Auto-write: BO%d = %v", index, value)
		c.mu.RUnlock()
		
		cmd := &session.OperateCommand{
			Group:             12, // Binary Output
			Variation:         1,
			Index:             index,
			Value:             value,
			SelectThenOperate: false, // Use DirectOperate for reliability
		}
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		opResp, err := s.SendCommand(ctx, cmd)
		if err != nil {
			c.handleError("Auto-write failed: %v", err)
			return
		}
		
		c.handleResponse(opResp)
	} else {
		// Analog Output operate
		index := uint16(rand.Intn(aoCount))
		// Random value in a reasonable range (0-100)
		value := float64(rand.Intn(101))
		
		c.mu.RLock()
		c.logger.Info("Auto-write: AO%d = %.2f", index, value)
		c.mu.RUnlock()
		
		cmd := &session.OperateCommand{
			Group:             41, // Analog Output
			Variation:         9,  // Variation 9 = 32-bit float (matches float64 encoding)
			Index:             index,
			Value:             value,
			SelectThenOperate: false, // Use DirectOperate for reliability
		}
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		opResp, err := s.SendCommand(ctx, cmd)
		if err != nil {
			c.handleError("Auto-write failed: %v", err)
			return
		}
		
		c.handleResponse(opResp)
	}
}

// Connect establishes a connection to an outstation.
func (c *Controller) Connect(address string, port int) error {
	c.mu.Lock()

	if c.session != nil && c.state.Connection == session.StateConnected {
		c.mu.Unlock()
		return nil
	}

	c.state.Connection = session.StateConnecting
	c.state.Address = address
	c.state.Port = port
	c.mu.Unlock()

	c.logger.Info("Master connecting to %s:%d", address, port)

	// Create session in goroutine
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Create master session
		masterSession, err := session.NewMasterSession(c.logger)
		if err != nil {
			c.handleError("Failed to create session: %v", err)
			return
		}

		// Connect
		if err := masterSession.Connect(ctx, address, port); err != nil {
			c.handleError("Connection failed: %v", err)
			c.mu.Lock()
			c.state.Connection = session.StateError
			c.mu.Unlock()
			return
		}

		c.mu.Lock()
		c.session = masterSession
		c.state.Connection = session.StateConnected
		c.state.Error = ""
		c.mu.Unlock()

		c.logger.Info("Master connected successfully")

		// Pump events
		c.pumpEvents()
	}()

	return nil
}

// Disconnect closes the current connection.
func (c *Controller) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Always disable auto-poll and auto-write on disconnect
	c.state.AutoPollEnabled = false
	c.state.AutoWriteEnabled = false
	c.state.SimulationMode = false

	if c.session == nil {
		return nil
	}

	c.logger.Info("Master disconnecting")
	c.state.Connection = session.StateDisconnected

	go func() {
		c.session.Close()
		c.mu.Lock()
		c.session = nil
		c.mu.Unlock()
	}()

	return nil
}

// ReadClass sends a read request for the specified class.
func (c *Controller) ReadClass(class int) error {
	c.mu.RLock()
	s := c.session
	state := c.state.Connection
	c.mu.RUnlock()

	if state != session.StateConnected || s == nil {
		return fmt.Errorf("not connected")
	}

	var groups []types.GroupRequest
	switch class {
	case 0:
		groups = []types.GroupRequest{{Group: 60, Variation: 1}} // All static
		c.logger.Info("Master sending READ Class 0")
	case 1:
		groups = []types.GroupRequest{{Group: 60, Variation: 2}}
		c.logger.Info("Master sending READ Class 1")
	case 2:
		groups = []types.GroupRequest{{Group: 60, Variation: 3}}
		c.logger.Info("Master sending READ Class 2")
	case 3:
		groups = []types.GroupRequest{{Group: 60, Variation: 4}}
		c.logger.Info("Master sending READ Class 3")
	default:
		return fmt.Errorf("invalid class: %d", class)
	}

	cmd := &session.ReadCommand{Groups: groups}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := s.SendCommand(ctx, cmd)
		if err != nil {
			c.handleError("Read Class %d failed: %v", class, err)
			return
		}

		c.handleResponse(resp)
	}()

	return nil
}

// Operate sends a binary control operation.
func (c *Controller) Operate(index uint16, value bool) error {
	c.mu.RLock()
	s := c.session
	state := c.state.Connection
	c.mu.RUnlock()

	if state != session.StateConnected || s == nil {
		return fmt.Errorf("not connected")
	}

	cmd := &session.OperateCommand{
		Group:             12, // Binary Output
		Variation:         1,
		Index:             index,
		Value:             value,
		SelectThenOperate: true,
	}

	c.logger.Info("Master sending OPERATE: Index=%d Value=%v", index, value)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := s.SendCommand(ctx, cmd)
		if err != nil {
			c.handleError("Operate failed: %v", err)
			return
		}

		c.handleResponse(resp)
	}()

	return nil
}

// OperateAnalog sends an analog control operation.
func (c *Controller) OperateAnalog(index uint16, value float64) error {
	c.mu.RLock()
	s := c.session
	state := c.state.Connection
	c.mu.RUnlock()

	if state != session.StateConnected || s == nil {
		return fmt.Errorf("not connected")
	}

	cmd := &session.OperateCommand{
		Group:             41, // Analog Output
		Variation:         1,
		Index:             index,
		Value:             value,
		SelectThenOperate: true,
	}

	c.logger.Info("Master sending ANALOG OPERATE: Index=%d Value=%v", index, value)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := s.SendCommand(ctx, cmd)
		if err != nil {
			c.handleError("Analog Operate failed: %v", err)
			return
		}

		c.handleResponse(resp)
	}()

	return nil
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
		case "response":
			if resp, ok := event.Data.(*session.Response); ok {
				c.handleResponse(resp)
			}
		case "connected":
			c.logger.Info("Master session connected")
		case "disconnected":
			c.logger.Info("Master session disconnected")
			c.mu.Lock()
			c.state.Connection = session.StateDisconnected
			c.mu.Unlock()
			return
		}
	}
}

// handleResponse processes a response from the session.
func (c *Controller) handleResponse(resp *session.Response) {
	if resp == nil {
		return
	}

	c.mu.Lock()
	c.state.LastResponse = resp
	c.mu.Unlock()

	c.logger.Info("Master received: %d binary, %d analog, %d counters",
		len(resp.BinaryInputs), len(resp.AnalogInputs), len(resp.Counters))
}

// handleError logs and stores an error.
func (c *Controller) handleError(format string, args ...interface{}) {
	errMsg := fmt.Sprintf(format, args...)
	c.logger.Error("Master error: %s", errMsg)

	c.mu.Lock()
	c.state.Error = errMsg
	c.mu.Unlock()
}
