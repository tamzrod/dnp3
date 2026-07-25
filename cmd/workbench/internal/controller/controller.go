// Package controller handles application events and state management.
package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"dnp3/cmd/workbench/internal/logger"
	"dnp3/cmd/workbench/internal/protocol"
	"dnp3/cmd/workbench/internal/session"
	"dnp3/pkg/dnp3/types"
)

// AppState represents the current application state.
type AppState struct {
	Mode            session.SessionMode
	Connection      session.ConnectionState
	Address         string
	Port            int
	LastResponse    *session.Response
	DecodedFrame    *protocol.DecodedFrame
	ConnectionError string
}

// Controller handles application events and coordinates between UI and session.
type Controller struct {
	mu         sync.RWMutex
	appState   *AppState
	session    session.Session
	logger     *logger.Logger
	decoder    *protocol.Decoder
	onStateChange func(*AppState)
	onLogEntry   func(*logger.Entry)
}

// Config holds controller configuration.
type Config struct {
	OnStateChange func(*AppState)
	OnLogEntry    func(*logger.Entry)
}

// New creates a new controller.
func New(cfg *Config) *Controller {
	c := &Controller{
		appState: &AppState{
			Mode:       session.ModeMaster,
			Connection: session.StateDisconnected,
			Address:    "localhost",
			Port:       20000,
		},
		logger:  logger.New(),
		decoder: protocol.NewDecoder(),
	}

	if cfg != nil {
		c.onStateChange = cfg.OnStateChange
		c.onLogEntry = cfg.OnLogEntry
	}

	// Set up log entry callback
	c.logger.SetCallback(func(entry *logger.Entry) {
		if c.onLogEntry != nil {
			c.onLogEntry(entry)
		}
	})

	return c
}

// Start initializes the controller.
func (c *Controller) Start() error {
	c.logger.Info("Controller starting")
	c.notifyStateChange()
	return nil
}

// Stop shuts down the controller.
func (c *Controller) Stop() error {
	c.logger.Info("Controller stopping")
	if c.session != nil {
		c.session.Close()
	}
	return nil
}

// Connect establishes a connection to the outstation.
func (c *Controller) Connect(address string, port int) error {
	c.mu.Lock()
	
	if c.session != nil && c.session.State() == session.StateConnected {
		c.mu.Unlock()
		return nil
	}

	c.appState.Connection = session.StateConnecting
	c.appState.Address = address
	c.appState.Port = port
	c.mu.Unlock()

	c.notifyStateChange()
	c.logger.Info("Connecting to %s:%d", address, port)

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
			c.appState.Connection = session.StateError
			c.appState.ConnectionError = err.Error()
			c.mu.Unlock()
			c.notifyStateChange()
			return
		}

		c.mu.Lock()
		c.session = masterSession
		c.appState.Connection = session.StateConnected
		c.appState.ConnectionError = ""
		c.mu.Unlock()
		c.logger.Info("Connected successfully")
		c.notifyStateChange()

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

	c.logger.Info("Disconnecting")
	c.appState.Connection = session.StateDisconnected
	c.notifyStateChange()

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
	state := c.appState.Connection
	c.mu.RUnlock()

	if state != session.StateConnected || s == nil {
		return fmt.Errorf("not connected")
	}

	var groups []types.GroupRequest
	switch class {
	case 0:
		groups = []types.GroupRequest{{Group: 60, Variation: 1}} // All static data
		c.logger.Info("Sending READ Class 0 (Integrity Poll)")
	case 1:
		groups = []types.GroupRequest{{Group: 60, Variation: 2}} // Class 1 events
		c.logger.Info("Sending READ Class 1")
	case 2:
		groups = []types.GroupRequest{{Group: 60, Variation: 3}} // Class 2 events
		c.logger.Info("Sending READ Class 2")
	case 3:
		groups = []types.GroupRequest{{Group: 60, Variation: 4}} // Class 3 events
		c.logger.Info("Sending READ Class 3")
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

// Operate sends a control operation.
func (c *Controller) Operate(point uint16, value bool) error {
	c.mu.RLock()
	s := c.session
	state := c.appState.Connection
	c.mu.RUnlock()

	if state != session.StateConnected || s == nil {
		return fmt.Errorf("not connected")
	}

	cmd := &session.OperateCommand{
		Group:             12, // Binary Output
		Variation:         1,
		Index:             point,
		Value:             value,
		SelectThenOperate: true,
	}

	c.logger.Info("Sending OPERATE: Index=%d Value=%v", point, value)

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

// ReadObjects sends a read request for specific object groups.
func (c *Controller) ReadObjects(groups []types.GroupRequest) error {
	c.mu.RLock()
	s := c.session
	state := c.appState.Connection
	c.mu.RUnlock()

	if state != session.StateConnected || s == nil {
		return fmt.Errorf("not connected")
	}

	cmd := &session.ReadCommand{Groups: groups}
	c.logger.Info("Sending READ for %d object groups", len(groups))

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := s.SendCommand(ctx, cmd)
		if err != nil {
			c.handleError("Read Objects failed: %v", err)
			return
		}

		c.handleResponse(resp)
	}()

	return nil
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
		case "error":
			if err, ok := event.Data.(error); ok {
				c.handleError("Session error: %v", err)
			}
		case "connected":
			c.logger.Info("Session connected")
		case "disconnected":
			c.logger.Info("Session disconnected")
			c.mu.Lock()
			c.appState.Connection = session.StateDisconnected
			c.mu.Unlock()
			c.notifyStateChange()
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
	c.appState.LastResponse = resp
	c.mu.Unlock()

	c.logger.Info("Response received: %d binary, %d analog, %d counters",
		len(resp.BinaryInputs), len(resp.AnalogInputs), len(resp.Counters))

	c.notifyStateChange()
}

// handleError logs and displays an error.
func (c *Controller) handleError(format string, args ...interface{}) {
	errMsg := fmt.Sprintf(format, args...)
	c.logger.Error(errMsg)
	
	c.mu.Lock()
	c.appState.ConnectionError = errMsg
	c.mu.Unlock()
	
	c.notifyStateChange()
}

// notifyStateChange notifies listeners of state changes.
func (c *Controller) notifyStateChange() {
	if c.onStateChange != nil {
		c.mu.RLock()
		state := *c.appState
		c.mu.RUnlock()
		c.onStateChange(&state)
	}
}

// State returns the current application state.
func (c *Controller) State() *AppState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	state := *c.appState
	return &state
}

// Logger returns the controller logger.
func (c *Controller) Logger() *logger.Logger {
	return c.logger
}

// Decoder returns the protocol decoder.
func (c *Controller) Decoder() *protocol.Decoder {
	return c.decoder
}

// Session returns the current session.
func (c *Controller) Session() session.Session {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.session
}
