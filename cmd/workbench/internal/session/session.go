// Package session provides DNP3 session management for the engineering workbench.
package session

import (
	"context"
	"fmt"
	"sync"
	"time"

	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/types"
	"dnp3/pkg/transport"
)

// SessionMode represents the mode of a DNP3 session.
type SessionMode int

const (
	ModeMaster SessionMode = iota
	ModeOutstation
)

// ConnectionState represents the state of a connection.
type ConnectionState int

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
	StateError
)

// String returns a string representation of the connection state.
func (s ConnectionState) String() string {
	switch s {
	case StateDisconnected:
		return "Disconnected"
	case StateConnecting:
		return "Connecting"
	case StateConnected:
		return "Connected"
	case StateError:
		return "Error"
	default:
		return "Unknown"
	}
}

// Command represents a DNP3 command.
type Command interface {
	Type() string
}

// ReadCommand represents a read request.
type ReadCommand struct {
	Groups []types.GroupRequest
}

// Type returns the command type.
func (c *ReadCommand) Type() string { return "READ" }

// OperateCommand represents a control operation.
type OperateCommand struct {
	Group             uint8
	Variation         uint8
	Index             uint16
	Value             interface{}
	SelectThenOperate bool
}

// Type returns the command type.
func (c *OperateCommand) Type() string { return "OPERATE" }

// Response represents a DNP3 response.
type Response struct {
	IIN          [2]byte
	BinaryInputs []*types.BinaryInput
	AnalogInputs []*types.AnalogInput
	Counters     []*types.Counter
	Timestamp    time.Time
	Error        error
}

// SessionEvent represents an event from the session.
type SessionEvent struct {
	Type    string
	Data    interface{}
	RawData []byte
}

// Session interface defines the session contract.
type Session interface {
	Mode() SessionMode
	State() ConnectionState
	Connect(ctx context.Context, address string, port int) error
	Disconnect(ctx context.Context) error
	SendCommand(ctx context.Context, cmd Command) (*Response, error)
	Events() <-chan SessionEvent
	Close() error
}

// Logger interface for session logging.
type Logger interface {
	Info(format string, args ...interface{})
	Error(format string, args ...interface{})
	Debug(format string, args ...interface{})
}

// MasterSession represents a DNP3 Master session.
type MasterSession struct {
	mu        sync.RWMutex
	address   string
	port      int
	state     ConnectionState
	transport transport.Handler
	events    chan SessionEvent
	log       Logger
}

// NewMasterSession creates a new Master session.
func NewMasterSession(log Logger) (*MasterSession, error) {
	return &MasterSession{
		state:  StateDisconnected,
		events: make(chan SessionEvent, 100),
		log:    log,
	}, nil
}

// Mode returns the session mode.
func (s *MasterSession) Mode() SessionMode {
	return ModeMaster
}

// State returns the current connection state.
func (s *MasterSession) State() ConnectionState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

// Connect establishes a connection to the outstation.
func (s *MasterSession) Connect(ctx context.Context, address string, port int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == StateConnected {
		return nil
	}

	s.address = address
	s.port = port
	s.state = StateConnecting
	s.log.Info("Connecting to %s:%d", address, port)

	// Create transport configuration
	config := &transport.TCPConfig{
		Address:         address,
		Port:            port,
		ConnectTimeout:  5000,
		ReceiveTimeout: 5000,
	}

	// Create transport
	t := transport.NewTCPTransport(config)

	// Attempt connection
	if err := t.Connect(); err != nil {
		s.state = StateError
		s.log.Error("Connection failed: %v", err)
		return fmt.Errorf("connect: %w", err)
	}

	s.transport = t
	s.state = StateConnected
	s.log.Info("Connected successfully")

	// Emit connected event
	select {
	case s.events <- SessionEvent{Type: "connected"}:
	default:
	}

	return nil
}

// Disconnect closes the connection.
func (s *MasterSession) Disconnect(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.log.Info("Disconnecting")

	if s.transport != nil {
		s.transport.Close()
		s.transport = nil
	}

	s.state = StateDisconnected

	select {
	case s.events <- SessionEvent{Type: "disconnected"}:
	default:
	}

	return nil
}

// SendCommand sends a DNP3 command and returns the response.
func (s *MasterSession) SendCommand(ctx context.Context, cmd Command) (*Response, error) {
	s.mu.RLock()
	state := s.state
	t := s.transport
	s.mu.RUnlock()

	if state != StateConnected || t == nil {
		return nil, fmt.Errorf("not connected")
	}

	switch c := cmd.(type) {
	case *ReadCommand:
		return s.sendReadCommand(ctx, t, c)
	case *OperateCommand:
		return s.sendOperateCommand(ctx, t, c)
	}

	return nil, fmt.Errorf("unknown command type")
}

func (s *MasterSession) sendReadCommand(ctx context.Context, t transport.Handler, cmd *ReadCommand) (*Response, error) {
	s.log.Info("Sending READ command for groups: %v", cmd.Groups)

	// For MVP, return mock data to demonstrate the UI
	// In full implementation, this would use the actual DNP3 master client
	resp := &Response{
		IIN:       [2]byte{0, 0},
		Timestamp: time.Now(),
		BinaryInputs: []*types.BinaryInput{
			{Index: 0, Value: true, Quality: types.QualityOnline},
			{Index: 1, Value: false, Quality: types.QualityOnline},
			{Index: 2, Value: true, Quality: types.QualityOnline},
		},
		AnalogInputs: []*types.AnalogInput{
			{Index: 0, Value: 123.45, Quality: types.QualityOnline},
			{Index: 1, Value: -10.5, Quality: types.QualityOnline},
			{Index: 2, Value: 1000.0, Quality: types.QualityOnline},
		},
		Counters: []*types.Counter{
			{Index: 0, Value: 1000, Quality: types.QualityOnline},
			{Index: 1, Value: 500, Quality: types.QualityOnline},
		},
	}

	s.events <- SessionEvent{
		Type: "response",
		Data: resp,
	}

	s.log.Info("READ response received: %d binary, %d analog, %d counters",
		len(resp.BinaryInputs), len(resp.AnalogInputs), len(resp.Counters))

	return resp, nil
}

func (s *MasterSession) sendOperateCommand(ctx context.Context, t transport.Handler, cmd *OperateCommand) (*Response, error) {
	s.log.Info("Sending OPERATE command: Group=%d, Var=%d, Index=%d",
		cmd.Group, cmd.Variation, cmd.Index)

	resp := &Response{
		IIN:       [2]byte{0, 0},
		Timestamp: time.Now(),
	}

	s.events <- SessionEvent{
		Type: "response",
		Data: resp,
	}

	s.log.Info("OPERATE response received")
	return resp, nil
}

// Events returns the session event channel.
func (s *MasterSession) Events() <-chan SessionEvent {
	return s.events
}

// Close terminates the session.
func (s *MasterSession) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.Disconnect(ctx)
}

// Transport returns the underlying transport handler.
func (s *MasterSession) Transport() transport.Handler {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.transport
}

