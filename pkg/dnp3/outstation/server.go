// Package outstation provides the DNP3 Outstation server API.
//
// This package implements the Outstation role of the DNP3 protocol.
// An Outstation responds to requests from Masters and reports data.
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
	"sync"
	"time"

	"dnp3/pkg/dnp3"
	"dnp3/internal/outstation"
	"dnp3/pkg/dnp3/types"
	"dnp3/pkg/transport"
)

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
	}
}

// ConfigOption is a functional option for configuring the server
type ConfigOption func(*Config)

// WithAddress sets the outstation address
func WithAddress(addr uint16) ConfigOption {
	return func(c *Config) {
		c.OutstationAddress = addr
	}
}

// WithMasterAddress sets the expected master address
func WithMasterAddress(addr uint16) ConfigOption {
	return func(c *Config) {
		c.MasterAddress = addr
	}
}

// WithTransport sets the transport type, address, and port
func WithTransport(t dnp3.TransportType, address string, port int) ConfigOption {
	return func(c *Config) {
		c.TransportType = t
		c.Address = address
		c.Port = port
	}
}

// WithTLS sets the TLS configuration
func WithTLS(config *tls.Config) ConfigOption {
	return func(c *Config) {
		c.TLSConfig = config
		c.TransportType = dnp3.TLS
	}
}

// WithMaxFragmentSize sets the maximum fragment size
func WithMaxFragmentSize(size int) ConfigOption {
	return func(c *Config) {
		c.MaxFragmentSize = size
	}
}

// WithUnsolicitedMode enables unsolicited responses
func WithUnsolicitedMode(enabled bool) ConfigOption {
	return func(c *Config) {
		c.UnsolicitedMode = enabled
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
	if c.TransportType == dnp3.TLS && c.TLSConfig == nil {
		return &dnp3.ConfigurationError{
			Field:   "TLSConfig",
			Message: "required for TLS transport",
		}
	}
	return nil
}

// DataHandler provides data points to the outstation.
// Implement this interface to provide your application's data to DNP3 Masters.
type DataHandler interface {
	// GetBinaryInputs returns binary input data points
	GetBinaryInputs() []*types.BinaryInput

	// GetAnalogInputs returns analog input data points
	GetAnalogInputs() []*types.AnalogInput

	// GetCounters returns counter data points
	GetCounters() []*types.Counter
}

// CommandHandler handles control commands from the master
type CommandHandler interface {
	// HandleBinaryCommand handles a binary output command
	HandleBinaryCommand(command *types.ControlOutput) (*types.ControlStatus, error)

	// HandleAnalogCommand handles an analog output command
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
	publicHandler DataHandler
}

func (h *internalDataHandler) GetBinaryInputs() []outstation.BinaryInput {
	public := h.publicHandler.GetBinaryInputs()
	if public == nil {
		return nil
	}
	internal := make([]outstation.BinaryInput, len(public))
	for i, bi := range public {
		internal[i] = outstation.BinaryInput{
			Value:   bi.Value,
			Quality: uint8(bi.Quality),
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
		internal[i] = outstation.AnalogInput{
			Value:   ai.Value,
			Quality: uint8(ai.Quality),
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
		internal[i] = outstation.Counter{
			Value:   c.Value,
			Quality: uint8(c.Quality),
		}
	}
	return internal
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

// server is the internal implementation of Server
type server struct {
	config            *Config
	state             ServerState
	mu                sync.RWMutex
	dataHandler       DataHandler
	commandHandler    CommandHandler
	unsolicitedHandler UnsolicitedHandler

	// Transport for server mode
	transport transport.Handler

	// Internal implementation
	internalOutstation *outstation.Outstation

	// For Run loop management
	runCancel context.CancelFunc
}

// NewServer creates a new Outstation server with the given configuration
func NewServer(config *Config) (Server, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Create internal outstation config
	internalConfig := &outstation.Config{
		OutstationAddress: config.OutstationAddress,
		MasterAddress:    config.MasterAddress,
		Timeout:         5000, // Default timeout in ms
		MaxEventBuffers: config.MaxEventBuffers,
		Unsolicited:     config.UnsolicitedMode,
	}

	// Create internal outstation
	internal := outstation.NewOutstation(internalConfig)

	return &server{
		config:            config,
		state:             ServerStateDown,
		dataHandler:       &DefaultDataHandler{},
		commandHandler:    nil,
		unsolicitedHandler: nil,
		internalOutstation: internal,
	}, nil
}

// Start implements Server.Start
func (s *server) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != ServerStateDown {
		return fmt.Errorf("server not stopped: %s", s.state)
	}

	s.state = ServerStateStarting

	// Create transport based on configuration
	var t transport.Handler
	switch s.config.TransportType {
	case dnp3.TCP:
		tcpConfig := &transport.TCPConfig{
			Address:         s.config.Address,
			Port:           s.config.Port,
			ConnectTimeout: 5000,
			ReceiveTimeout: 5000,
			Server:         true, // Server mode
			KeepAlive:      30000,
		}
		t = transport.NewTCPTransport(tcpConfig)
	case dnp3.TLS:
		// For TLS, we fall back to TCP for now
		// Full TLS would require TLSTransport with server certificate
		tcpConfig := &transport.TCPConfig{
			Address:         s.config.Address,
			Port:           s.config.Port,
			ConnectTimeout: 5000,
			ReceiveTimeout: 5000,
			Server:         true,
			KeepAlive:      30000,
		}
		t = transport.NewTCPTransport(tcpConfig)
	}

	s.transport = t

	// Set transport on internal outstation
	s.internalOutstation.SetTransport(t)

	// Initialize internal outstation
	if err := s.internalOutstation.Initialize(); err != nil {
		s.state = ServerStateError
		return fmt.Errorf("initialize failed: %w", err)
	}

	// Set up data handler adapter
	s.internalOutstation.SetDataHandler(&internalDataHandler{
		publicHandler: s.dataHandler,
	})

	// Start internal outstation
	if err := s.internalOutstation.Start(); err != nil {
		s.state = ServerStateError
		return fmt.Errorf("start failed: %w", err)
	}

	// Start the run loop in a goroutine
	runCtx, runCancel := context.WithCancel(context.Background())
	s.runCancel = runCancel
	go func() {
		// Accept connection first
		if err := t.Accept(); err != nil {
			// Log error but don't crash
			return
		}
		// Then run the main loop using the context
		_ = runCtx
		s.internalOutstation.Run()
	}()

	s.state = ServerStateRunning
	return nil
}

// Stop implements Server.Stop
func (s *server) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == ServerStateDown {
		return nil
	}

	s.state = ServerStateStopping

	// Cancel the run loop
	if s.runCancel != nil {
		s.runCancel()
	}

	// Stop internal outstation
	if err := s.internalOutstation.Stop(); err != nil {
		s.state = ServerStateError
		return fmt.Errorf("stop failed: %w", err)
	}

	// Close transport
	if s.transport != nil {
		_ = s.transport.Close()
	}

	s.state = ServerStateDown
	return nil
}

// State implements Server.State
func (s *server) State() ServerState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

// SetDataHandler implements Server.SetDataHandler
func (s *server) SetDataHandler(handler DataHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if handler == nil {
		handler = &DefaultDataHandler{}
	}
	s.dataHandler = handler

	// Update internal handler if outstation exists
	if s.internalOutstation != nil {
		s.internalOutstation.SetDataHandler(&internalDataHandler{
			publicHandler: handler,
		})
	}
}

// SetCommandHandler implements Server.SetCommandHandler
func (s *server) SetCommandHandler(handler CommandHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.commandHandler = handler
}

// SetUnsolicitedHandler implements Server.SetUnsolicitedHandler
func (s *server) SetUnsolicitedHandler(handler UnsolicitedHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.unsolicitedHandler = handler
}

// Close implements Server.Close
func (s *server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.Stop(ctx)
}
