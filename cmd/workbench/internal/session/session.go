// Package session provides DNP3 session management for the engineering workbench.
package session

import (
	"context"
	"fmt"
	"sync"
	"time"

	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/master"
	"dnp3/pkg/dnp3/types"
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
	IIN            [2]byte
	BinaryInputs   []*types.BinaryInput
	AnalogInputs   []*types.AnalogInput
	Counters       []*types.Counter
	FrozenCounters []*types.FrozenCounter
	Timestamp      time.Time
	Error          error
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
	client    master.Client
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

	// Create real DNP3 master client
	config := master.NewConfig(
		master.WithOutstationAddress(1024),
		master.WithTransport(dnp3.TCP, address, port),
		master.WithTimeout(5*time.Second),
		master.WithRetry(3, 100*time.Millisecond),
	)

	client, err := master.NewClient(config)
	if err != nil {
		s.state = StateError
		s.log.Error("Failed to create DNP3 client: %v", err)
		return fmt.Errorf("create client: %w", err)
	}

	// Attempt connection
	if err := client.Connect(ctx); err != nil {
		s.state = StateError
		s.log.Error("Connection failed: %v", err)
		return fmt.Errorf("connect: %w", err)
	}

	s.client = client
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

	if s.client != nil {
		s.client.Close()
		s.client = nil
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
	client := s.client
	s.mu.RUnlock()

	if state != StateConnected || client == nil {
		return nil, fmt.Errorf("not connected")
	}

	switch c := cmd.(type) {
	case *ReadCommand:
		return s.sendReadCommand(ctx, client, c)
	case *OperateCommand:
		return s.sendOperateCommand(ctx, client, c)
	}

	return nil, fmt.Errorf("unknown command type")
}

func (s *MasterSession) sendReadCommand(ctx context.Context, client master.Client, cmd *ReadCommand) (*Response, error) {
	s.log.Info("Sending READ command for groups: %v", cmd.Groups)

	// Create read request from groups
	request := &types.ReadRequest{
		Groups: cmd.Groups,
	}

	// Use real DNP3 client
	readResp, err := client.Read(ctx, request)
	if err != nil {
		s.log.Error("READ command failed: %v", err)
		return nil, fmt.Errorf("read: %w", err)
	}

	// Convert to session response
	resp := &Response{
		IIN:           readResp.IIN,
		BinaryInputs:  readResp.BinaryInputs,
		AnalogInputs:  readResp.AnalogInputs,
		Counters:      readResp.Counters,
		FrozenCounters: readResp.FrozenCounters,
		Timestamp:     readResp.Timestamp,
	}

	s.events <- SessionEvent{
		Type: "response",
		Data: resp,
	}

	s.log.Info("READ response received: %d binary, %d analog, %d counters",
		len(resp.BinaryInputs), len(resp.AnalogInputs), len(resp.Counters))

	return resp, nil
}

func (s *MasterSession) sendOperateCommand(ctx context.Context, client master.Client, cmd *OperateCommand) (*Response, error) {
	s.log.Info("Sending OPERATE command: Group=%d, Var=%d, Index=%d, Value=%v",
		cmd.Group, cmd.Variation, cmd.Index, cmd.Value)

	// Determine command type based on SelectThenOperate flag
	cmdType := types.DirectOperate
	if cmd.SelectThenOperate {
		cmdType = types.SelectThenOperate
	}

	// Create control output based on command type
	var controlOutput *types.ControlOutput
	if cmd.Group == 12 {
		// Binary output command
		if v, ok := cmd.Value.(bool); ok {
			controlOutput = types.NewBinaryControl(cmd.Index, v, cmdType)
		}
	} else if cmd.Group == 41 {
		// Analog output command
		if v, ok := cmd.Value.(float64); ok {
			controlOutput = types.NewAnalogControl(cmd.Index, v, cmdType)
		}
	}

	if controlOutput == nil {
		return nil, fmt.Errorf("unsupported command type: group=%d", cmd.Group)
	}

	// Use real DNP3 client
	operateResp, err := client.Operate(ctx, controlOutput)
	if err != nil {
		s.log.Error("OPERATE command failed: %v", err)
		return nil, fmt.Errorf("operate: %w", err)
	}

	resp := &Response{
		IIN:       operateResp.IIN,
		Timestamp: operateResp.Timestamp,
	}

	s.events <- SessionEvent{
		Type: "response",
		Data: resp,
	}

	s.log.Info("OPERATE response received, status: %v", operateResp.Status)
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

// Transport returns nil for master session (client handles transport internally).
func (s *MasterSession) Transport() interface{} {
	return nil
}

