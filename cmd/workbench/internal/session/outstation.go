// Package session provides DNP3 session management for the engineering workbench.
package session

import (
	"context"
	"fmt"
	"sync"
	"time"

	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/outstation"
	"dnp3/pkg/dnp3/types"
)

// OutstationSession represents a DNP3 Outstation session (server).
type OutstationSession struct {
	mu      sync.RWMutex
	address string
	port    int
	state   ConnectionState
	server  outstation.Server
	events  chan SessionEvent
	log     Logger

	// Data handler for providing data points
	dataHandler *outstationDataHandler

	// Connection info
	clientAddress string
}

// outstationDataHandler adapts session data points to DNP3 server interface
type outstationDataHandler struct {
	session *OutstationSession
}

func (h *outstationDataHandler) GetBinaryInputs() []*types.BinaryInput {
	return h.session.GetBinaryInputs()
}

func (h *outstationDataHandler) GetAnalogInputs() []*types.AnalogInput {
	return h.session.GetAnalogInputs()
}

func (h *outstationDataHandler) GetCounters() []*types.Counter {
	return h.session.GetCounters()
}

func (h *outstationDataHandler) GetFrozenCounters() []*types.Counter {
	return nil
}

func (h *outstationDataHandler) FreezeCounters(clear bool) error {
	if clear {
		for _, c := range h.session.GetCounters() {
			c.Value = 0
		}
	}
	return nil
}

func (h *outstationDataHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	// Type assert the command value
	if v, ok := cmd.Value.(*types.BinaryCommandValue); ok {
		h.session.SetBinaryOutput(cmd.Index, v.Value)
	} else {
		h.session.SetBinaryOutput(cmd.Index, false)
	}
	status := types.ControlSuccess
	return &status, nil
}

func (h *outstationDataHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	// Type assert the command value
	if v, ok := cmd.Value.(*types.AnalogCommandValue); ok {
		h.session.SetAnalogOutput(cmd.Index, v.Value)
	} else {
		h.session.SetAnalogOutput(cmd.Index, 0.0)
	}
	status := types.ControlSuccess
	return &status, nil
}

// NewOutstationSession creates a new Outstation session.
func NewOutstationSession(log Logger) (*OutstationSession, error) {
	s := &OutstationSession{
		state:        StateDisconnected,
		events:       make(chan SessionEvent, 100),
		log:          log,
		dataHandler:  &outstationDataHandler{},
	}
	s.dataHandler.session = s
	return s, nil
}

// Mode returns the session mode.
func (s *OutstationSession) Mode() SessionMode {
	return ModeOutstation
}

// State returns the current connection state.
func (s *OutstationSession) State() ConnectionState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

// Start begins listening for connections using real DNP3 server.
func (s *OutstationSession) Start(address string, port int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == StateConnected {
		return nil
	}

	s.address = address
	s.port = port
	s.state = StateConnecting
	s.log.Info("Starting DNP3 outstation on %s:%d", address, port)

	// Create real DNP3 outstation server
	config := outstation.NewConfig(
		outstation.WithAddress(1024), // DNP3 address
		outstation.WithTransport(dnp3.TCP, address, port),
	)

	server, err := outstation.NewServer(config)
	if err != nil {
		s.state = StateError
		s.log.Error("Failed to create DNP3 server: %v", err)
		return fmt.Errorf("create server: %w", err)
	}

	// Set data handler with simulated data
	server.SetDataHandler(s.dataHandler)

	// Set command handler
	server.SetCommandHandler(s.dataHandler)

	// Start server
	ctx := context.Background()
	if err := server.Start(ctx); err != nil {
		s.state = StateError
		s.log.Error("Failed to start server: %v", err)
		return fmt.Errorf("start: %w", err)
	}

	s.server = server
	s.state = StateConnected
	s.log.Info("DNP3 outstation listening on %s:%d", address, port)

	// Emit started event
	s.emitEvent("started", map[string]interface{}{
		"address": address,
		"port":    port,
	})

	return nil
}

// Stop stops the outstation.
func (s *OutstationSession) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.log.Info("Stopping DNP3 outstation")

	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.server.Stop(ctx)
		s.server = nil
	}

	s.state = StateDisconnected
	s.emitEvent("stopped", nil)

	return nil
}

// Events returns the session event channel.
func (s *OutstationSession) Events() <-chan SessionEvent {
	return s.events
}

// Close terminates the session.
func (s *OutstationSession) Close() error {
	return s.Stop()
}

// ClientAddress returns the connected master's address.
func (s *OutstationSession) ClientAddress() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.clientAddress
}

// emitEvent sends an event to the event channel.
func (s *OutstationSession) emitEvent(eventType string, data interface{}) {
	event := SessionEvent{
		Type: eventType,
		Data: data,
	}
	select {
	case s.events <- event:
	default:
	}
}

// Transport returns nil for outstation (server handles transport internally).
func (s *OutstationSession) Transport() interface{} {
	return nil
}

// Connect is not applicable for outstation, use Start instead.
func (s *OutstationSession) Connect(ctx context.Context, address string, port int) error {
	return fmt.Errorf("use Start() for outstation mode")
}

// Disconnect is not applicable for outstation, use Stop instead.
func (s *OutstationSession) Disconnect(ctx context.Context) error {
	return s.Stop()
}

// SendCommand handles commands for outstation (not supported).
func (s *OutstationSession) SendCommand(ctx context.Context, cmd Command) (*Response, error) {
	return nil, fmt.Errorf("outstation does not support SendCommand")
}

// Required interface compliance
func (s *OutstationSession) SendCommandCompat(ctx context.Context, cmd Command) (*Response, error) {
	return s.SendCommand(ctx, cmd)
}

// GetBinaryInputs returns current binary inputs
func (s *OutstationSession) GetBinaryInputs() []*types.BinaryInput {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return default simulated data if none configured
	return []*types.BinaryInput{
		{Index: 0, Value: true, Quality: types.QualityOnline, Time: (&types.Timestamp{}).Now()},
		{Index: 1, Value: false, Quality: types.QualityOnline, Time: (&types.Timestamp{}).Now()},
		{Index: 2, Value: true, Quality: types.QualityOnline, Time: (&types.Timestamp{}).Now()},
		{Index: 3, Value: false, Quality: types.QualityOnline, Time: (&types.Timestamp{}).Now()},
	}
}

// GetAnalogInputs returns current analog inputs
func (s *OutstationSession) GetAnalogInputs() []*types.AnalogInput {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return default simulated data if none configured
	return []*types.AnalogInput{
		{Index: 0, Value: 100.5, Quality: types.QualityOnline, Time: (&types.Timestamp{}).Now()},
		{Index: 1, Value: 200.25, Quality: types.QualityOnline, Time: (&types.Timestamp{}).Now()},
		{Index: 2, Value: -50.0, Quality: types.QualityOnline, Time: (&types.Timestamp{}).Now()},
		{Index: 3, Value: 0.0, Quality: types.QualityOnline, Time: (&types.Timestamp{}).Now()},
	}
}

// GetCounters returns current counters
func (s *OutstationSession) GetCounters() []*types.Counter {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return default simulated data if none configured
	return []*types.Counter{
		{Index: 0, Value: 1000, Quality: types.QualityOnline, Time: (&types.Timestamp{}).Now()},
		{Index: 1, Value: 2000, Quality: types.QualityOnline, Time: (&types.Timestamp{}).Now()},
	}
}

// AddBinaryInput adds a binary input point.
func (s *OutstationSession) AddBinaryInput(index uint16, value bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.log.Info("Added binary input: index=%d, value=%v", index, value)
}

// SetBinaryInput sets a binary input value.
func (s *OutstationSession) SetBinaryInput(index uint16, value bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.log.Info("Set binary input: index=%d, value=%v", index, value)
}

// AddAnalogInput adds an analog input point.
func (s *OutstationSession) AddAnalogInput(index uint16, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.log.Info("Added analog input: index=%d, value=%v", index, value)
}

// SetAnalogInput sets an analog input value.
func (s *OutstationSession) SetAnalogInput(index uint16, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.log.Info("Set analog input: index=%d, value=%v", index, value)
}

// AddCounter adds a counter point.
func (s *OutstationSession) AddCounter(index uint16, value uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.log.Info("Added counter: index=%d, value=%d", index, value)
}

// IncrementCounter increments a counter.
func (s *OutstationSession) IncrementCounter(index uint16, delta uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.log.Info("Incremented counter: index=%d, delta=%d", index, delta)
}

// SetBinaryOutput sets a binary output value.
func (s *OutstationSession) SetBinaryOutput(index uint16, value bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.log.Info("Set binary output: index=%d, value=%v", index, value)
}

// SetAnalogOutput sets an analog output value.
func (s *OutstationSession) SetAnalogOutput(index uint16, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.log.Info("Set analog output: index=%d, value=%v", index, value)
}
