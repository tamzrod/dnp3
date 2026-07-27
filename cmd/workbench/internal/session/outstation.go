// Package session provides DNP3 session management for the engineering workbench.
package session

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"dnp3/pkg/dnp3/types"
)

// OutstationSession represents a DNP3 Outstation session (server).
type OutstationSession struct {
	mu        sync.RWMutex
	address   string
	port      int
	state     ConnectionState
	listener  net.Listener
	conn      net.Conn
	events    chan SessionEvent
	log       Logger

	// Data points
	BinaryInputs   []*types.BinaryInput
	AnalogInputs   []*types.AnalogInput
	Counters       []*types.Counter
	BinaryOutputs  []*types.BinaryOutput
	AnalogOutputs  []*types.AnalogOutput

	// Connection info
	clientAddress string
}

// NewOutstationSession creates a new Outstation session.
func NewOutstationSession(log Logger) (*OutstationSession, error) {
	return &OutstationSession{
		state:          StateDisconnected,
		events:         make(chan SessionEvent, 100),
		log:            log,
		BinaryInputs:   make([]*types.BinaryInput, 0),
		AnalogInputs:   make([]*types.AnalogInput, 0),
		Counters:       make([]*types.Counter, 0),
		BinaryOutputs:  make([]*types.BinaryOutput, 0),
		AnalogOutputs:  make([]*types.AnalogOutput, 0),
	}, nil
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

// Start begins listening for connections.
func (s *OutstationSession) Start(address string, port int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == StateConnected {
		return nil
	}

	s.address = address
	s.port = port
	s.state = StateConnecting
	s.log.Info("Starting outstation on %s:%d", address, port)

	// Create TCP listener
	addr := fmt.Sprintf("%s:%d", address, port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		s.state = StateError
		s.log.Error("Failed to start listener: %v", err)
		return fmt.Errorf("listen: %w", err)
	}

	s.listener = listener
	s.state = StateConnected
	s.log.Info("Outstation listening on %s:%d", address, port)

	// Emit started event
	s.emitEvent("started", map[string]interface{}{
		"address": address,
		"port":    port,
	})

	// Accept connections in background
	go s.acceptConnections()

	return nil
}

// Stop stops the outstation.
func (s *OutstationSession) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.log.Info("Stopping outstation")

	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}

	if s.listener != nil {
		s.listener.Close()
		s.listener = nil
	}

	s.state = StateDisconnected
	s.emitEvent("stopped", nil)

	return nil
}

// acceptConnections accepts incoming connections.
func (s *OutstationSession) acceptConnections() {
	s.mu.RLock()
	listener := s.listener
	s.mu.RUnlock()

	if listener == nil {
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			s.mu.RLock()
			state := s.state
			s.mu.RUnlock()
			if state == StateDisconnected {
				return // Graceful shutdown
			}
			s.log.Error("Accept error: %v", err)
			continue
		}

		s.mu.Lock()
		if s.conn != nil {
			// Already have a connection, close new one
			conn.Close()
			s.mu.Unlock()
			continue
		}
		s.conn = conn
		s.clientAddress = conn.RemoteAddr().String()
		s.mu.Unlock()

		s.log.Info("Master connected from %s", s.clientAddress)
		s.emitEvent("master_connected", map[string]interface{}{
			"address": s.clientAddress,
		})

		// Handle connection in background
		go s.handleConnection(conn)
	}
}

// handleConnection handles a master connection.
func (s *OutstationSession) handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 4096)
	for {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		n, err := conn.Read(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// Timeout, continue
				continue
			}
			s.log.Info("Master disconnected: %v", err)
			s.emitEvent("master_disconnected", nil)
			return
		}

		if n > 0 {
			s.log.Debug("Received %d bytes from master", n)
			// For MVP, just log the raw data
			// Full implementation would parse DNP3 frames
			s.emitEvent("data_received", map[string]interface{}{
				"data": buf[:n],
				"size": n,
			})

			// For MVP, send mock response
			s.sendMockResponse(conn)
		}
	}
}

// sendMockResponse sends a mock DNP3 response.
func (s *OutstationSession) sendMockResponse(conn net.Conn) {
	// For MVP, send minimal response
	// Full implementation would build proper DNP3 response
	response := []byte{0x05, 0x64, 0x0C, 0x01, 0x00}
	conn.Write(response)
}

// AddBinaryInput adds a binary input point.
func (s *OutstationSession) AddBinaryInput(index uint16, value bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bi := &types.BinaryInput{
		Index:   index,
		Value:   value,
		Quality: types.QualityOnline,
		Time:    types.Timestamp{}.Now(),
	}
	s.BinaryInputs = append(s.BinaryInputs, bi)
}

// SetBinaryInput sets a binary input value.
func (s *OutstationSession) SetBinaryInput(index uint16, value bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, bi := range s.BinaryInputs {
		if bi.Index == index {
			bi.Value = value
			bi.Time = types.Timestamp{}.Now()
			return
		}
	}
}

// AddAnalogInput adds an analog input point.
func (s *OutstationSession) AddAnalogInput(index uint16, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ai := &types.AnalogInput{
		Index:   index,
		Value:   value,
		Quality: types.QualityOnline,
		Time:    types.Timestamp{}.Now(),
	}
	s.AnalogInputs = append(s.AnalogInputs, ai)
}

// SetAnalogInput sets an analog input value.
func (s *OutstationSession) SetAnalogInput(index uint16, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, ai := range s.AnalogInputs {
		if ai.Index == index {
			ai.Value = value
			ai.Time = types.Timestamp{}.Now()
			return
		}
	}
}

// AddCounter adds a counter point.
func (s *OutstationSession) AddCounter(index uint16, value uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c := &types.Counter{
		Index:   index,
		Value:   value,
		Quality: types.QualityOnline,
		Time:    types.Timestamp{}.Now(),
	}
	s.Counters = append(s.Counters, c)
}

// IncrementCounter increments a counter.
func (s *OutstationSession) IncrementCounter(index uint16, delta uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, c := range s.Counters {
		if c.Index == index {
			c.Value += delta
			c.Time = types.Timestamp{}.Now()
			return
		}
	}
}

// SetBinaryOutput sets a binary output value.
func (s *OutstationSession) SetBinaryOutput(index uint16, value bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, bo := range s.BinaryOutputs {
		if bo.Index == index {
			bo.Value = value
			return
		}
	}
	// Add new output
	bo := &types.BinaryOutput{
		Index:  index,
		Value:  value,
		Status: types.BinaryOutputVerified,
	}
	s.BinaryOutputs = append(s.BinaryOutputs, bo)
}

// SetAnalogOutput sets an analog output value.
func (s *OutstationSession) SetAnalogOutput(index uint16, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, ao := range s.AnalogOutputs {
		if ao.Index == index {
			ao.Value = value
			return
		}
	}
	// Add new output
	ao := &types.AnalogOutput{
		Index:  index,
		Value:  value,
		Status: types.AnalogOutputVerified,
	}
	s.AnalogOutputs = append(s.AnalogOutputs, ao)
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

// Transport returns nil for outstation (no transport until connected).
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

// SendCommand handles commands for outstation (unsupported in MVP).
func (s *OutstationSession) SendCommand(ctx context.Context, cmd Command) (*Response, error) {
	return nil, fmt.Errorf("outstation does not support SendCommand")
}

// Required interface compliance
func (s *OutstationSession) SendCommandCompat(ctx context.Context, cmd Command) (*Response, error) {
	return s.SendCommand(ctx, cmd)
}
