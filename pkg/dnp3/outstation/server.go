// Package outstation provides the DNP3 Outstation server API.
//
// This package implements the Outstation role of the DNP3 protocol.
// An Outstation responds to requests from Masters and reports data.
//
// Multi-master support: Each accepted connection gets its own outstation instance.
//
// # Usage
//
// Create a new server:
//
//	config := outstation.NewConfig().
//		WithAddress(1024).
//		WithTransport(dnp3.TCP, "", 20000)
//
//	server, err := outstation.NewServer(config)
//
// Define a data handler:
//
//	handler := &myDataHandler{}
//
// Start the server:
//
//	ctx := context.Background()
//	if err := server.Start(ctx); err != nil {
//		log.Fatal(err)
//	}
//	defer server.Stop(ctx)
package outstation

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"dnp3/pkg/dnp3"
	"dnp3/internal/outstation"
	"dnp3/pkg/dnp3/types"
	"dnp3/pkg/transport"
)

// debugControls enables debug logging. Set DNP3_DEBUG=1 environment variable to enable.
var debugControls = os.Getenv("DNP3_DEBUG") == "1"

func debugLog(format string, v ...interface{}) {
	if debugControls {
		fmt.Printf(format+"\n", v...)
	}
}

// Server represents a DNP3 Outstation server.
// The server listens for connections from Masters and responds to their requests.
type Server interface {
	// Start begins listening for connections
	Start(ctx context.Context) error

	// Stop stops the server
	Stop(ctx context.Context) error

	// State returns the current server state
	State() ServerState

	// SetDataHandler sets the data provider
	SetDataHandler(handler DataHandler)

	// SetCommandHandler sets the command handler
	SetCommandHandler(handler CommandHandler)

	// SetUnsolicitedHandler sets the handler for enabling/disabling unsolicited
	SetUnsolicitedHandler(handler UnsolicitedHandler)

	// Close gracefully shuts down the server
	Close() error
}

// ServerState represents the state of an outstation server
type ServerState int

const (
	// ServerStateDown - server is not running
	ServerStateDown ServerState = iota
	// ServerStateStarting - server is starting
	ServerStateStarting
	// ServerStateRunning - server is running and accepting connections
	ServerStateRunning
	// ServerStateStopping - server is stopping
	ServerStateStopping
	// ServerStateError - server encountered an error
	ServerStateError
)

// String returns the server state name
func (s ServerState) String() string {
	switch s {
	case ServerStateDown:
		return "Down"
	case ServerStateStarting:
		return "Starting"
	case ServerStateRunning:
		return "Running"
	case ServerStateStopping:
		return "Stopping"
	case ServerStateError:
		return "Error"
	default:
		return "Unknown"
	}
}

// Config holds Outstation server configuration
type Config struct {
	// OutstationAddress is this outstation's DNP3 address
	OutstationAddress uint16
	// MasterAddress is the expected master's DNP3 address
	MasterAddress uint16
	// TransportType specifies TCP or TLS
	TransportType dnp3.TransportType
	// Address is the network address to bind to (empty = all interfaces)
	Address string
	// Port is the network port to listen on
	Port int
	// TLSConfig is the TLS configuration (required for TLS transport)
	TLSConfig *tls.Config
	// MaxFragmentSize is the maximum fragment size
	MaxFragmentSize int
	// KeepAliveInterval is the keep-alive interval
	KeepAliveInterval time.Duration
	// UnsolicitedMode enables sending unsolicited responses
	UnsolicitedMode bool
	// MaxEventBuffers is the maximum event buffer size
	MaxEventBuffers int
	// MaxConnections is the maximum number of simultaneous master
	// connections the outstation will serve. The v0 MVP profile is
	// single-master: a connection that arrives while MaxConnections are
	// already active is rejected (closed) by the server with a clear log.
	// DNP3-084. Default 1.
	MaxConnections int
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		OutstationAddress:  1024,
		MasterAddress:      0xFFFF,
		TransportType:      dnp3.TCP,
		Address:            "", // all interfaces
		Port:               20000,
		MaxFragmentSize:    2048,
		KeepAliveInterval:  30 * time.Second,
		MaxEventBuffers:    1000,
		MaxConnections:     1, // DNP3-084: MVP single-master profile
	}
}

// ConfigOption is a functional option for configuring the server
type ConfigOption func(*Config)

// WithAddress sets the outstation address.
// Supported-profile: Target — one configured outstation/master address pair.
func WithAddress(addr uint16) ConfigOption {
	return func(c *Config) {
		c.OutstationAddress = addr
	}
}

// WithMasterAddress sets the expected master address.
// Supported-profile: Target — one configured outstation/master address pair.
func WithMasterAddress(addr uint16) ConfigOption {
	return func(c *Config) {
		c.MasterAddress = addr
	}
}

// WithTransport sets the transport type, address, and port.
// Supported-profile: Target for TCP (TCP listener only); Reject for TLS.
func WithTransport(t dnp3.TransportType, address string, port int) ConfigOption {
	return func(c *Config) {
		c.TransportType = t
		c.Address = address
		c.Port = port
	}
}

// WithTLS sets the TLS configuration.
// Supported-profile: Reject — no TLS listener in v0.
func WithTLS(config *tls.Config) ConfigOption {
	return func(c *Config) {
		c.TLSConfig = config
		c.TransportType = dnp3.TLS
	}
}

// WithMaxFragmentSize sets the maximum fragment size.
// Supported-profile: Defer — must be reconciled with tested fragmentation limits.
func WithMaxFragmentSize(size int) ConfigOption {
	return func(c *Config) {
		c.MaxFragmentSize = size
	}
}

// WithUnsolicitedMode enables unsolicited responses.
// Supported-profile: Reject — no unsolicited delivery path in v0.
func WithUnsolicitedMode(enabled bool) ConfigOption {
	return func(c *Config) {
		c.UnsolicitedMode = enabled
	}
}

// WithMaxConnections sets the maximum simultaneous master connections.
// The v0 MVP profile is single-master (default 1). DNP3-084.
func WithMaxConnections(n int) ConfigOption {
	return func(c *Config) {
		c.MaxConnections = n
	}
}

// NewConfig creates a new configuration with options applied
func NewConfig(opts ...ConfigOption) *Config {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.OutstationAddress == 0 || c.OutstationAddress == 0xFFFF {
		return &dnp3.ConfigurationError{
			Field:   "OutstationAddress",
			Message: "must be a valid address (not 0 or broadcast)",
		}
	}
	if c.MaxFragmentSize <= 0 || c.MaxFragmentSize > 65535 {
		return &dnp3.ConfigurationError{
			Field:   "MaxFragmentSize",
			Message: "must be between 1 and 65535",
		}
	}
	if c.TransportType == dnp3.TLS {
		// DNP3-087: TLS is outside the v0 MVP profile (no TLS listener).
		// Reject with a clear error instead of silently falling back to TCP.
		return &dnp3.ConfigurationError{
			Field:   "TransportType",
			Message: "TLS transport is not supported in the v0 MVP profile (use TCP)",
		}
	}
	if c.UnsolicitedMode {
		// DNP3-087: the v0 outstation has no unsolicited delivery path.
		// Reject enabling unsolicited responses with a clear error.
		return &dnp3.ConfigurationError{
			Field:   "UnsolicitedMode",
			Message: "unsolicited responses are not supported in the v0 MVP profile",
		}
	}
	if c.MaxConnections < 1 {
		return &dnp3.ConfigurationError{
			Field:   "MaxConnections",
			Message: "must be at least 1 (MVP single-master profile)",
		}
	}
	return nil
}

// DataHandler provides data points to the outstation.
// Implement this interface to provide your application's data to DNP3 Masters.
//
// v0 MVP profile (DNP3-089): only the static input groups are read by the
// outstation — G1.1 (binary input), G30.1 (analog input), and G20.1 (counter).
// Those three getters are the MVP-required methods and must return real data.
//
// The remaining methods are outside the v0 read profile (see
// supported-profile.md: "Data handler binary/analog output status and
// frozen-counter methods — Reject / Not read by the v0 profile"). They exist on
// the interface for API continuity but are not read in v0; a minimal handler
// may implement them as no-ops / returning nil. The outstation's builders
// return an empty (object-less) payload for any group whose getter returns an
// empty slice, so a handler that provides only the MVP input groups still
// serves a valid Class-0 integrity response.
type DataHandler interface {
	// GetBinaryInputs returns binary input data points (MVP-required: G1.1).
	GetBinaryInputs() []*types.BinaryInput

	// GetAnalogInputs returns analog input data points (MVP-required: G30.1).
	GetAnalogInputs() []*types.AnalogInput

	// GetCounters returns counter data points (MVP-required: G20.1).
	GetCounters() []*types.Counter

	// GetBinaryOutputs returns binary output status data points. Not read by
	// the v0 profile; return nil for a minimal MVP handler.
	GetBinaryOutputs() []*types.BinaryOutput

	// GetAnalogOutputs returns analog output status data points. Not read by
	// the v0 profile; return nil for a minimal MVP handler.
	GetAnalogOutputs() []*types.AnalogOutput

	// GetFrozenCounters returns frozen counter data points. Not read by the
	// v0 profile; return nil for a minimal MVP handler.
	GetFrozenCounters() []*types.Counter

	// FreezeCounters freezes the counters. Not read by the v0 profile; a
	// minimal MVP handler may implement this as a no-op returning nil.
	FreezeCounters(clear bool) error
}

// CommandHandler handles control commands from the master.
//
// v0 MVP profile (DNP3-090): only Group 12 Variation 1 (CROB) direct binary
// control is in scope. HandleBinaryCommand is the MVP-required method and
// must return a ControlStatus for G12V1 commands.
//
// HandleAnalogCommand covers Group 41 and related analog-output variations,
// which are outside the v0 profile (supported-profile.md: "Command handler
// analog controls — Reject — Group 41 and related variations are out of
// scope"). A minimal MVP handler must therefore reject analog commands by
// returning ControlNotSupported together with a clear error rather than
// executing them. The interface method is retained for API continuity.
type CommandHandler interface {
	// HandleBinaryCommand handles a binary output command (MVP-required:
	// Group 12 Variation 1 CROB). Returns the resulting ControlStatus.
	HandleBinaryCommand(command *types.ControlOutput) (*types.ControlStatus, error)

	// HandleAnalogCommand handles an analog output command. Out of v0 scope;
	// a minimal MVP handler returns ControlNotSupported with a clear error.
	HandleAnalogCommand(command *types.ControlOutput) (*types.ControlStatus, error)
}

// UnsolicitedHandler handles unsolicited response configuration
type UnsolicitedHandler interface {
	// OnUnsolicitedEnabled is called when unsolicited is enabled
	OnUnsolicitedEnabled()

	// OnUnsolicitedDisabled is called when unsolicited is disabled
	OnUnsolicitedDisabled()
}

// internalDataHandler adapts public DataHandler to internal/outstation.DataHandler
type internalDataHandler struct {
	publicHandler   DataHandler
	commandHandler CommandHandler
}

func (h *internalDataHandler) GetBinaryInputs() []outstation.BinaryInput {
	public := h.publicHandler.GetBinaryInputs()
	if public == nil {
		return nil
	}
	internal := make([]outstation.BinaryInput, len(public))
	for i, bi := range public {
		var timeValue uint64
		if bi.Time != nil && !bi.Time.IsNull() {
			// Convert types.Timestamp to uint64 DNP3 time
			// DNP3 time is milliseconds since 2000-01-01
			timeValue = bi.Time.Value
		}
		internal[i] = outstation.BinaryInput{
			Value:   bi.Value,
			Quality: uint8(bi.Quality),
			Time:    timeValue,
		}
	}
	return internal
}

func (h *internalDataHandler) GetAnalogInputs() []outstation.AnalogInput {
	public := h.publicHandler.GetAnalogInputs()
	if public == nil {
		return nil
	}
	internal := make([]outstation.AnalogInput, len(public))
	for i, ai := range public {
		var timeValue uint64
		if ai.Time != nil && !ai.Time.IsNull() {
			timeValue = ai.Time.Value
		}
		internal[i] = outstation.AnalogInput{
			Value:   ai.Value,
			Quality: uint8(ai.Quality),
			Time:    timeValue,
		}
	}
	return internal
}

func (h *internalDataHandler) GetCounters() []outstation.Counter {
	public := h.publicHandler.GetCounters()
	if public == nil {
		return nil
	}
	internal := make([]outstation.Counter, len(public))
	for i, c := range public {
		var timeValue uint64
		if c.Time != nil && !c.Time.IsNull() {
			timeValue = c.Time.Value
		}
		internal[i] = outstation.Counter{
			Value:   c.Value,
			Quality: uint8(c.Quality),
			Time:    timeValue,
		}
	}
	return internal
}

func (h *internalDataHandler) GetBinaryOutputs() []outstation.BinaryOutput {
	public := h.publicHandler.GetBinaryOutputs()
	if public == nil {
		return nil
	}
	internal := make([]outstation.BinaryOutput, len(public))
	for i, bo := range public {
		internal[i] = outstation.BinaryOutput{
			Value:   bo.Value,
			Quality: uint8(bo.Quality),
		}
	}
	return internal
}

func (h *internalDataHandler) GetAnalogOutputs() []outstation.AnalogOutput {
	public := h.publicHandler.GetAnalogOutputs()
	if public == nil {
		return nil
	}
	internal := make([]outstation.AnalogOutput, len(public))
	for i, ao := range public {
		internal[i] = outstation.AnalogOutput{
			Value:   ao.Value,
			Quality: uint8(ao.Quality),
		}
	}
	return internal
}

func (h *internalDataHandler) GetFrozenCounters() []outstation.Counter {
	public := h.publicHandler.GetFrozenCounters()
	if public == nil {
		return nil
	}
	internal := make([]outstation.Counter, len(public))
	for i, c := range public {
		internal[i] = outstation.Counter{
			Value:   c.Value,
			Quality: uint8(c.Quality),
		}
	}
	return internal
}

func (h *internalDataHandler) FreezeCounters(clear bool) error {
	return h.publicHandler.FreezeCounters(clear)
}

// WriteBinaryOutput handles writing a binary output (CROB) from the master.
// Converts internal CROB to public ControlOutput and delegates to commandHandler.
func (h *internalDataHandler) WriteBinaryOutput(index uint16, crob *outstation.CROB) error {
	if h.commandHandler == nil {
		debugLog("ERROR: WriteBinaryOutput called but commandHandler is nil (index=%d, code=%d)", index, crob.Code)
		return fmt.Errorf("no command handler registered")
	}

	// Convert CROB to binary control output
	// CROB codes: 1=NUL, 2=CLOSE(ON), 3=OPEN(OFF), 4=TRIP, 5=PULSE_ON, 6=PULSE_OFF, 7=LATCH_ON, 8=LATCH_OFF
	var value bool
	switch crob.Code {
	case 2, 4, 5, 7: // CLOSE, TRIP, PULSE_ON, LATCH_ON
		value = true
	case 1, 3, 6, 8: // NUL, OPEN, PULSE_OFF, LATCH_OFF
		value = false
	default:
		value = false
	}

	debugLog("WriteBinaryOutput: index=%d, value=%v, code=%d", index, value, crob.Code)

	cmd := &types.ControlOutput{
		Group:      12, // Binary Output
		Variation:  1, // CROB
		Index:     index,
		Value:     &types.BinaryCommandValue{Value: value},
		CommandType: types.SelectThenOperate,
		Count:     crob.Count,
		OnTime:    crob.OnTime,
		OffTime:   crob.OffTime,
	}

	status, err := h.commandHandler.HandleBinaryCommand(cmd)
	if err != nil {
		debugLog("WriteBinaryOutput ERROR: %v", err)
	} else if status != nil {
		debugLog("WriteBinaryOutput result: status=%d", *status)
	}
	return err
}

// WriteAnalogOutput handles writing an analog output from the master.
// Converts internal value to public ControlOutput and delegates to commandHandler.
func (h *internalDataHandler) WriteAnalogOutput(index uint16, value interface{}, variation uint8) error {
	if h.commandHandler == nil {
		debugLog("ERROR: WriteAnalogOutput called but commandHandler is nil (index=%d)", index)
		return fmt.Errorf("no command handler registered")
	}

	// Convert interface{} to float64
	var floatValue float64
	switch v := value.(type) {
	case int16:
		floatValue = float64(v)
	case uint16:
		floatValue = float64(v)
	case int32:
		floatValue = float64(v)
	case uint32:
		floatValue = float64(v)
	case float32:
		floatValue = float64(v)
	case float64:
		floatValue = v
	default:
		floatValue = 0
	}

	debugLog("WriteAnalogOutput: index=%d, value=%v, variation=%d", index, floatValue, variation)

	cmd := &types.ControlOutput{
		Group:        41, // Analog Output
		Variation:   variation,
		Index:       index,
		Value:       &types.AnalogCommandValue{Value: floatValue},
		CommandType: types.SelectThenOperate,
	}

	status, err := h.commandHandler.HandleAnalogCommand(cmd)
	if err != nil {
		debugLog("WriteAnalogOutput ERROR: %v", err)
	} else if status != nil {
		debugLog("WriteAnalogOutput result: status=%d", *status)
	}
	return err
}

// DefaultDataHandler is a simple data handler that returns empty data
type DefaultDataHandler struct{}

// GetBinaryInputs returns mock data
func (d *DefaultDataHandler) GetBinaryInputs() []*types.BinaryInput {
	return []*types.BinaryInput{
		{Index: 0, Value: true, Quality: types.QualityOnline},
		{Index: 1, Value: false, Quality: types.QualityOnline},
	}
}

// GetAnalogInputs returns mock data
func (d *DefaultDataHandler) GetAnalogInputs() []*types.AnalogInput {
	return []*types.AnalogInput{
		{Index: 0, Value: 100.5, Quality: types.QualityOnline},
		{Index: 1, Value: 200.25, Quality: types.QualityOnline},
	}
}

// GetCounters returns mock data
func (d *DefaultDataHandler) GetCounters() []*types.Counter {
	return []*types.Counter{
		{Index: 0, Value: 1000, Quality: types.QualityOnline},
		{Index: 1, Value: 2000, Quality: types.QualityOnline},
	}
}

// GetBinaryOutputs returns binary output status data points
func (d *DefaultDataHandler) GetBinaryOutputs() []*types.BinaryOutput {
	return []*types.BinaryOutput{
		{Index: 0, Value: true, Quality: types.QualityOnline},
		{Index: 1, Value: false, Quality: types.QualityOnline},
	}
}

// GetAnalogOutputs returns analog output status data points
func (d *DefaultDataHandler) GetAnalogOutputs() []*types.AnalogOutput {
	return []*types.AnalogOutput{
		{Index: 0, Value: 50.0, Quality: types.QualityOnline},
		{Index: 1, Value: 25.0, Quality: types.QualityOnline},
	}
}

// GetFrozenCounters returns empty frozen counters
func (d *DefaultDataHandler) GetFrozenCounters() []*types.Counter {
	return []*types.Counter{}
}

// FreezeCounters is a no-op for DefaultDataHandler
func (d *DefaultDataHandler) FreezeCounters(clear bool) error {
	return nil
}

// server is the internal implementation of Server
type server struct {
	config            *Config
	state             ServerState
	mu                sync.RWMutex
	dataHandler       DataHandler
	commandHandler    CommandHandler
	unsolicitedHandler UnsolicitedHandler

	// Listener for accepting connections
	listener net.Listener

	// Track active connections
	connections map[uint64]*outstationInstance
	connCounter uint64
	connectionsMu sync.RWMutex

	// Context for graceful shutdown. DNP3-085: derived from the caller's
	// Start(ctx) context so cancelling that ctx produces a clean stop.
	runCtx    context.Context
	runCancel context.CancelFunc
	// acceptDone is closed when the accept loop has fully exited and
	// shutdown completed. Allows Stop to wait for a clean stop.
	acceptDone chan struct{}
}

// outstationInstance represents a single master connection
type outstationInstance struct {
	id        uint64
	masterAddr net.Addr
	outstation *outstation.Outstation
	transport  transport.Handler
	done      chan struct{}
}

// NewServer creates a new Outstation server with the given configuration
func NewServer(config *Config) (Server, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &server{
		config:            config,
		state:             ServerStateDown,
		dataHandler:       &DefaultDataHandler{},
		commandHandler:    nil,
		unsolicitedHandler: nil,
		connections:      make(map[uint64]*outstationInstance),
	}, nil
}

// Start implements Server.Start.
// Supported-profile: Target — one TCP listener lifecycle.
// DNP3-085: the server's run context is derived from the caller's ctx, so
// cancelling that ctx produces a clean stop (listener closed, connections
// dropped, state -> Down).
func (s *server) Start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("start context already cancelled: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != ServerStateDown {
		return fmt.Errorf("server not stopped: %s", s.state)
	}

	s.state = ServerStateStarting

	// Create TCP listener directly for better control
	addr := fmt.Sprintf("%s:%d", s.config.Address, s.config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		s.state = ServerStateError
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = listener

	debugLog("DNP3 Outstation listening on %s", addr)

	// DNP3-085: derive the run context from the caller's ctx so cancelling
	// the caller's ctx cancels the accept loop and triggers shutdown.
	s.runCtx, s.runCancel = context.WithCancel(ctx)
	s.acceptDone = make(chan struct{})

	// Start accept loop in a separate goroutine
	go s.acceptLoop()

	s.state = ServerStateRunning
	return nil
}

// acceptLoop accepts incoming connections and creates outstation instances.
// DNP3-085: it exits when runCtx is cancelled (including via the caller's
// Start ctx) and performs an idempotent shutdown before returning.
func (s *server) acceptLoop() {
	defer func() {
		s.shutdown()
		s.mu.Lock()
		ch := s.acceptDone
		s.mu.Unlock()
		if ch != nil {
			close(ch)
		}
	}()

	for {
		select {
		case <-s.runCtx.Done():
			debugLog("Accept loop shutting down (context cancelled)")
			return
		default:
			// Set accept deadline
			s.listener.(*net.TCPListener).SetDeadline(time.Now().Add(1 * time.Second))

			conn, err := s.listener.Accept()
			if err != nil {
				// Check if it's a timeout
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				// Check if context was cancelled
				if s.runCtx.Err() != nil {
					return
				}
				// Log other errors but continue
				debugLog("Accept error: %v", err)
				continue
			}

			// Handle new connection
			go s.handleConnection(conn)
		}
	}
}

// shutdown is idempotent: it cancels the run context, closes the listener and
// all active connections, and transitions state to Down. Safe to call from the
// accept loop (on context cancellation) or from Stop.
func (s *server) shutdown() {
	s.mu.Lock()
	if s.state == ServerStateDown {
		s.mu.Unlock()
		return
	}
	s.state = ServerStateStopping
	s.mu.Unlock()

	// Cancel the run context so the accept loop (if still running) exits.
	if s.runCancel != nil {
		s.runCancel()
	}

	// Close all active connections
	s.connectionsMu.Lock()
	for id, inst := range s.connections {
		debugLog("Closing connection %d", id)
		inst.outstation.Stop()
		inst.transport.Close()
	}
	s.connections = make(map[uint64]*outstationInstance)
	s.connectionsMu.Unlock()

	// Close listener
	s.mu.Lock()
	if s.listener != nil {
		_ = s.listener.Close()
		s.listener = nil
	}
	s.state = ServerStateDown
	s.mu.Unlock()
}

// handleConnection handles a single master connection
func (s *server) handleConnection(conn net.Conn) {
	s.connectionsMu.Lock()
	// DNP3-084: MVP single-master profile. Reject a connection that arrives
	// while MaxConnections are already active: close it immediately with a
	// clear log rather than spawning a competing outstation instance.
	active := len(s.connections)
	if active >= s.config.MaxConnections {
		max := s.config.MaxConnections
		s.connectionsMu.Unlock()
		debugLog("Connection from %s rejected: %d/%d connections active (MVP single-master)",
			conn.RemoteAddr(), active, max)
		_ = conn.Close()
		return
	}
	s.connCounter++
	connID := s.connCounter

	// Create transport for this connection
	tcpConfig := &transport.TCPConfig{
		Address:         "",
		Port:           0,
		ConnectTimeout: 5000,
		ReceiveTimeout: 5000,
		KeepAlive:      30000,
	}
	t := transport.NewTCPTransport(tcpConfig)
	// Set the connection directly
	t.SetConn(conn)

	// Create new outstation for this connection
	internalConfig := &outstation.Config{
		OutstationAddress: s.config.OutstationAddress,
		MasterAddress:     s.config.MasterAddress,
		Timeout:         5000,
		MaxEventBuffers: s.config.MaxEventBuffers,
		Unsolicited:     s.config.UnsolicitedMode,
	}
	inst := outstation.NewOutstation(internalConfig)
	inst.SetTransport(t)
	inst.Initialize()
	inst.SetDataHandler(&internalDataHandler{
		publicHandler:   s.dataHandler,
		commandHandler: s.commandHandler,
	})
	inst.Start()

	instance := &outstationInstance{
		id:        connID,
		masterAddr: conn.RemoteAddr(),
		outstation: inst,
		transport:  t,
		done:      make(chan struct{}),
	}
	s.connections[connID] = instance
	s.connectionsMu.Unlock()

	debugLog("Connection %d accepted from %s", connID, conn.RemoteAddr())

	// Run the outstation (this blocks until connection closes)
	inst.Run()

	// Clean up
	s.connectionsMu.Lock()
	delete(s.connections, connID)
	s.connectionsMu.Unlock()

	conn.Close()
	debugLog("Connection %d closed", connID)
}

// Stop implements Server.Stop.
// Supported-profile: Target — one TCP listener lifecycle.
// DNP3-085: triggers an idempotent shutdown and waits (up to ctx deadline) for
// the accept loop to fully exit.
func (s *server) Stop(ctx context.Context) error {
	s.mu.Lock()
	if s.state == ServerStateDown {
		s.mu.Unlock()
		return nil
	}
	ch := s.acceptDone
	s.mu.Unlock()

	s.shutdown()

	if ch != nil {
		select {
		case <-ch:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// State implements Server.State
func (s *server) State() ServerState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

// SetDataHandler implements Server.SetDataHandler.
// Supported-profile: Target — static Groups 1.1, 30.1, and 20.1 only.
func (s *server) SetDataHandler(handler DataHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if handler == nil {
		handler = &DefaultDataHandler{}
	}
	s.dataHandler = handler

	// Update all active connections with new data handler
	s.connectionsMu.Lock()
	for _, inst := range s.connections {
		inst.outstation.SetDataHandler(&internalDataHandler{
			publicHandler:   handler,
			commandHandler: s.commandHandler,
		})
	}
	s.connectionsMu.Unlock()
}

// ActiveConnections returns the count of active connections
func (s *server) ActiveConnections() int {
	s.connectionsMu.RLock()
	defer s.connectionsMu.RUnlock()
	return len(s.connections)
}

// listenerAddr returns the bound listener address (nil before Start). Used by
// in-package tests to dial an ephemeral-port listener.
func (s *server) listenerAddr() net.Addr {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.listener == nil {
		return nil
	}
	return s.listener.Addr()
}

// SetCommandHandler implements Server.SetCommandHandler.
// Supported-profile: Target — Direct Group 12 Variation 1 control only.
func (s *server) SetCommandHandler(handler CommandHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.commandHandler = handler
}

// SetUnsolicitedHandler implements Server.SetUnsolicitedHandler.
// Supported-profile: Reject — no unsolicited delivery path in v0.
func (s *server) SetUnsolicitedHandler(handler UnsolicitedHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.unsolicitedHandler = handler
}

// Close implements Server.Close.
// Supported-profile: Target — one TCP listener lifecycle.
func (s *server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.Stop(ctx)
}
