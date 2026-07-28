// Package master provides Master-specific functionality for the DNP3 workbench.
package master

import (
	"context"
	"fmt"
	"sync"
	"time"

	"dnp3/cmd/workbench/internal/logger"
	"dnp3/cmd/workbench/internal/session"
	"dnp3/pkg/dnp3/types"
)

// State represents the current state of the Master controller.
type State struct {
	Connection   session.ConnectionState
	Address      string
	Port         int
	Error        string
	LastResponse *session.Response
}

// Controller handles Master-specific operations.
type Controller struct {
	mu      sync.RWMutex
	session *session.MasterSession
	logger  *logger.Logger
	state   *State
}

// NewController creates a new Master controller.
func NewController(log *logger.Logger) *Controller {
	return &Controller{
		logger: log,
		state: &State{
			Connection: session.StateDisconnected,
			Address:    "127.0.0.1",
			Port:       20000,
		},
	}
}

// Start initializes the controller.
func (c *Controller) Start() error {
	c.logger.Info("Master controller started")
	return nil
}

// Stop shuts down the controller.
func (c *Controller) Stop() error {
	c.logger.Info("Master controller stopping")
	if c.session != nil {
		c.session.Close()
	}
	return nil
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
