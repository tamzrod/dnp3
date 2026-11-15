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
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"dnp3/internal/outstation"
	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/types"
	"dnp3/pkg/transport"
)

// sessionTCP turns a dead TCP read into ErrSessionClosed so the outstation
// loop releases the single connection slot. Timeouts and resync errors are
// returned unchanged.
type sessionTCP struct {
	*transport.TCPTransport
}

func (s *sessionTCP) Receive() ([]byte, error) {
	data, err := s.TCPTransport.Receive()
	if err == nil || !sessionDead(err) {
		return data, err
	}
	return nil, fmt.Errorf("%w: %v", outstation.ErrSessionClosed, err)
}

func sessionDead(err error) bool {
	if transport.IsDisconnect(err) || errors.Is(err, transport.ErrNotConnected) {
		return true
	}
	if errors.Is(err, transport.ErrTimeout) || errors.Is(err, transport.ErrInvalidFrame) {
		return false
	}
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "TCP receive failed") || strings.Contains(msg, "connection reset") || strings.Contains(msg, "broken pipe")
}

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

	// SetUnsolicitedHandler sets the handler for enabling/disabling unsolicited.
	// Supported-profile: Reject by default — no handler is registered unless the
	// explicit, non-default WithExperimentalEventDelivery switch is set.
	SetUnsolicitedHandler(handler UnsolicitedHandler)

	// PublishBinaryEvent queues a binary input event. The host does not build
	// a DNP3 frame. quality is the Group 2 flag octet without the state bit.
	// A zero time selects the variation without a timestamp.
	// Supported-profile: Reject by default — returns ErrUnsupportedOption unless
	// the explicit, non-default WithExperimentalEventDelivery switch is set.
	PublishBinaryEvent(index uint16, value bool, quality uint8, class EventClass, at time.Time) error

	// PublishAnalogEvent queues an analog input event as Group 32 signed int32.
	// Supported-profile: Reject by default (see PublishBinaryEvent).
	PublishAnalogEvent(index uint16, value int32, quality uint8, class EventClass, at time.Time) error

	// PublishCounterEvent queues a counter event as Group 22 unsigned int32.
	// Supported-profile: Reject by default (see PublishBinaryEvent).
	PublishCounterEvent(index uint16, value uint32, quality uint8, class EventClass, at time.Time) error

	// Close gracefully shuts down the server
	Close() error
}

// EventClass is the DNP3 event class assigned to a published event.
type EventClass int

const (
	// EventClass1 is DNP3 Class 1.
	EventClass1 EventClass = iota
	// EventClass2 is DNP3 Class 2.
	EventClass2
	// EventClass3 is DNP3 Class 3.
	EventClass3
)

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
	// UnsolicitedMode enables sending unsolicited responses.
	// Supported-profile: Reject by default — enabling it requires the explicit,
	// non-default WithExperimentalEventDelivery switch or Validate rejects it.
	UnsolicitedMode bool
	// MaxEventBuffers is the maximum event buffer size
	MaxEventBuffers int
	// ExperimentalEventDelivery opts into the EVT-001 event and unsolicited
	// path (PublishBinaryEvent/PublishAnalogEvent/PublishCounterEvent,
	// SetUnsolicitedHandler, WithUnsolicitedMode). The published v0 supported
	// profile does not advertise events or unsolicited delivery, so the default
	// is false: publishing returns ErrUnsupportedOption, no handler is
	// registered, and WithUnsolicitedMode(true) fails validation. This is an
	// explicit, non-default experimental switch, not a profile promotion.
	// Supported-profile: Reject by default (experimental, off by default).
	ExperimentalEventDelivery bool
	// ExperimentalAnalogControl opts into the Group 41 analog-output command
	// path on the outstation bridge (the host CommandHandler.HandleAnalogCommand
	// call for a G41 Direct Operate). The published v0 supported profile rejects
	// "Command handler analog controls" (Group 41 and related variations), so the
	// default is false: a G41 Direct Operate fails closed before any host analog
	// handler is invoked and is never acknowledged as SUCCESS. This is an
	// explicit, non-default experimental switch, not a profile promotion. Enable
	// it only for the isolated G41 lab path.
	// Supported-profile: Reject by default (experimental, off by default).
	ExperimentalAnalogControl bool
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
		OutstationAddress: 1024,
		MasterAddress:     0xFFFF,
		TransportType:     dnp3.TCP,
		Address:           "", // all interfaces
		Port:              20000,
		MaxFragmentSize:   2048,
		KeepAliveInterval: 30 * time.Second,
		MaxEventBuffers:   1000,
		MaxConnections:    1, // DNP3-084: MVP single-master profile
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
// Supported-profile: Reject — no unsolicited delivery path in v0. Enabling it
// requires the explicit, non-default experimental event delivery switch;
// otherwise NewServer rejects the configuration.
func WithUnsolicitedMode(enabled bool) ConfigOption {
	return func(c *Config) {
		c.UnsolicitedMode = enabled
	}
}

// WithExperimentalEventDelivery opts into the EVT-001 event and unsolicited
// path. It is an explicit, non-default experimental switch: the published v0
// supported profile does not advertise events or unsolicited delivery, so
// without it event publishing returns ErrUnsupportedOption, no unsolicited
// handler is registered, and WithUnsolicitedMode(true) fails validation. It is
// not a profile promotion; it only opens the isolated EVT-001 lab path.
// Supported-profile: Reject by default (experimental, off by default).
func WithExperimentalEventDelivery() ConfigOption {
	return func(c *Config) {
		c.ExperimentalEventDelivery = true
	}
}

// WithExperimentalAnalogControl opts into the Group 41 analog-output command
// path on the outstation bridge (the host CommandHandler.HandleAnalogCommand
// call for a G41 Direct Operate). It is an explicit, non-default experimental
// switch: the published v0 supported profile rejects "Command handler analog
// controls" (Group 41 and related variations), so without it a G41 Direct
// Operate fails closed before any host analog handler is invoked and is never
// acknowledged as SUCCESS. It is not a profile promotion; it only opens the
// isolated G41 lab path.
// Supported-profile: Reject by default (experimental, off by default).
func WithExperimentalAnalogControl() ConfigOption {
	return func(c *Config) {
		c.ExperimentalAnalogControl = true
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
	if c.UnsolicitedMode && !c.ExperimentalEventDelivery {
		// DNP3-087: the v0 outstation has no unsolicited delivery path.
		// Reject enabling unsolicited responses with a clear error unless the
		// explicit, non-default experimental event delivery switch is set.
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
	publicHandler  DataHandler
	commandHandler CommandHandler
	// analogControlEnabled mirrors Config.ExperimentalAnalogControl. When false
	// (the default) WriteAnalogOutput fails closed before invoking the host
	// analog handler, so a Group 41 Direct Operate is never executed and never
	// acknowledged as SUCCESS through the ordinary public API.
	analogControlEnabled bool
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

	// Convert the CROB operation type (IEEE 1815 A.8.1 low-nibble enum) to a
	// binary value for the public command handler. Unsupported operation
	// types (including any trip/close or clear bits) are rejected explicitly
	// rather than silently mapped.
	var value bool
	switch crob.Code {
	case 0x01, 0x03: // Pulse On, Latch On
		value = true
	case 0x00, 0x02, 0x04: // NUL, Pulse Off, Latch Off
		value = false
	default:
		return fmt.Errorf("unsupported CROB operation type: 0x%02X", crob.Code)
	}

	debugLog("WriteBinaryOutput: index=%d, value=%v, code=%d", index, value, crob.Code)

	cmd := &types.ControlOutput{
		Group:       12, // Binary Output
		Variation:   1,  // CROB
		Index:       index,
		Value:       &types.BinaryCommandValue{Value: value},
		CommandType: types.SelectThenOperate,
		Count:       crob.Count,
		OnTime:      crob.OnTime,
		OffTime:     crob.OffTime,
	}

	status, err := h.commandHandler.HandleBinaryCommand(cmd)
	if err != nil {
		debugLog("WriteBinaryOutput ERROR: %v", err)
		return err
	}
	if status != nil && *status != types.ControlSuccess {
		// MEXT-024: a handler that signals a non-success status with no error
		// must NOT surface as ControlSuccess. The bridge returns an error so the
		// outstation emits an error response (a complete APDU, not a silent
		// timeout) carrying a failure IIN, which the master resolves to a
		// failure status (never success). Without this, a not-supported /
		// blocked command reported via status-only would be a false success.
		debugLog("WriteBinaryOutput: non-success status %d -> failure", *status)
		return fmt.Errorf("command rejected by handler: status %d", *status)
	}
	if status != nil {
		debugLog("WriteBinaryOutput result: status=%d", *status)
	}
	return nil
}

// wireCommandStatus maps a public ControlStatus to its IEEE 1815 command-status
// octet (the CTRL-01 per-point status byte). Only statuses with a verified,
// unambiguous wire encoding are mapped: public values 0..10 correspond 1:1 to
// IEEE 1815 command-status codes 0..10 (the internal master parses the wire byte
// directly into that enumeration), and ControlOutOfRange maps to wire 12 (0x0C,
// OUT_OF_RANGE in the pinned OpenDNP3 3.1.2 reference). The public synch values
// 11/12 are deliberately not mapped (they would collide with the distinct wire
// meanings 11 IN_SYNCH / 12 OUT_OF_RANGE); ok is false for them so the bridge
// fails closed with a generic error rather than guessing a wire status.
func wireCommandStatus(s types.ControlStatus) (uint8, bool) {
	switch s {
	case types.ControlSuccess:
		return 0, true
	case types.ControlTimeout:
		return 1, true
	case types.ControlNoSelect:
		return 2, true
	case types.ControlBadFormat:
		return 3, true
	case types.ControlNotSupported:
		return 4, true
	case types.ControlAlreadyActive:
		return 5, true
	case types.ControlHardwareError:
		return 6, true
	case types.ControlLocal:
		return 7, true
	case types.ControlTooMany:
		return 8, true
	case types.ControlNotAuthorized:
		return 9, true
	case types.ControlAutonomous:
		return 10, true
	case types.ControlOutOfRange:
		return 12, true
	default:
		return 0, false
	}
}

// WriteAnalogOutput handles writing an analog output from the master.
// Converts internal value to public ControlOutput and delegates to commandHandler.
// AO-002: Group 41 Variation 1 (int32) is carried as a typed
// AnalogInt32CommandValue with DirectOperate semantics (the supported mode);
// SelectThenOperate is not falsely presented. AO-003 adds the typed G41V2
// (int16), G41V3 (float32) and G41V4 (float64) carriers under the same
// DirectOperate-with-response semantics and the owner-approved float policy.
func (h *internalDataHandler) WriteAnalogOutput(index uint16, value interface{}, variation uint8) error {
	if !h.analogControlEnabled {
		// AO-005: the published v0 supported profile rejects "Command handler
		// analog controls" (Group 41 and related variations). Fail closed before
		// invoking the host analog handler or mutating any state, and never
		// acknowledge the command as SUCCESS. The G41 path is reachable only via
		// the explicit, non-default WithExperimentalAnalogControl switch.
		debugLog("WriteAnalogOutput: refusing Group 41 analog control (variation=%d); v0 profile rejects analog controls", variation)
		return fmt.Errorf("analog output (Group 41) is rejected by the v0 supported profile; enable WithExperimentalAnalogControl for the isolated lab path")
	}
	if h.commandHandler == nil {
		debugLog("ERROR: WriteAnalogOutput called but commandHandler is nil (index=%d)", index)
		return fmt.Errorf("no command handler registered")
	}

	cmd := &types.ControlOutput{
		Group:       41, // Analog Output
		Variation:   variation,
		Index:       index,
		CommandType: types.DirectOperate,
	}

	if variation == 1 {
		// G41V1: signed int32. The value arrives as int32 from the outstation
		// G41V1 parser; int16 is also accepted for direct bridge callers.
		var intVal int32
		switch v := value.(type) {
		case int32:
			intVal = v
		case int16:
			intVal = int32(v)
		default:
			return fmt.Errorf("unsupported G41V1 analog output type: %T", value)
		}
		cmd.Value = &types.AnalogInt32CommandValue{Value: intVal}
	} else if variation == 2 {
		// G41V2 (AO-003): signed int16. The value arrives as int16 from the
		// outstation G41V2 parser; an in-range int32 is also accepted for direct
		// bridge callers (an out-of-range int32 is rejected, not narrowed).
		var intVal int16
		switch v := value.(type) {
		case int16:
			intVal = v
		case int32:
			if v < math.MinInt16 || v > math.MaxInt16 {
				return fmt.Errorf("G41V2 analog output value %d out of int16 range", v)
			}
			intVal = int16(v)
		default:
			return fmt.Errorf("unsupported G41V2 analog output type: %T", value)
		}
		cmd.Value = &types.AnalogInt16CommandValue{Value: intVal}
	} else if variation == 3 {
		// G41V3 (AO-003): single-precision IEEE-754 float. The value arrives as
		// float32 from the outstation G41V3 parser (exact bits preserved). A float64
		// is accepted only when exactly representable as float32. NaN and ±Inf are
		// rejected (owner-approved AO-003 float policy).
		var f float32
		switch v := value.(type) {
		case float32:
			f = v
		case float64:
			c := float32(v)
			if float64(c) != v || math.IsInf(float64(c), 0) {
				return fmt.Errorf("G41V3 analog output value %v is not exactly representable as float32", v)
			}
			f = c
		default:
			return fmt.Errorf("unsupported G41V3 analog output type: %T", value)
		}
		if math.IsNaN(float64(f)) || math.IsInf(float64(f), 0) {
			return fmt.Errorf("G41V3 analog output value must be finite")
		}
		cmd.Value = &types.AnalogFloat32CommandValue{Value: f}
	} else if variation == 4 {
		// G41V4 (AO-003): double-precision IEEE-754 float. The value arrives as
		// float64 from the outstation G41V4 parser (exact bits preserved). Only a
		// float64 source is accepted; NaN and ±Inf are rejected (owner-approved
		// AO-003 float policy).
		f, ok := value.(float64)
		if !ok {
			return fmt.Errorf("unsupported G41V4 analog output type: %T", value)
		}
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return fmt.Errorf("G41V4 analog output value must be finite")
		}
		cmd.Value = &types.AnalogFloat64CommandValue{Value: f}
	} else {
		// Out of AO-002 scope (only Variation 1 is implemented). Retain the
		// generic float64 carrier for API continuity.
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
		cmd.Value = &types.AnalogCommandValue{Value: floatValue}
	}

	debugLog("WriteAnalogOutput: index=%d, variation=%d", index, variation)

	status, err := h.commandHandler.HandleAnalogCommand(cmd)
	// AO-004 Appendix T: fail closed, but preserve a handler's definite non-success
	// command status so the outstation can echo the G41 object with the actual
	// status octet (for example OUT_OF_RANGE) rather than a generic error response.
	// A nil status or a refusal without a wire-encodable status still returns an
	// error (an error response, never a SUCCESS echo). The returned error always
	// wraps the handler error (or a synthetic one) so existing fail-closed
	// classification is unchanged; the carried status is authoritative on the wire.
	if err != nil {
		if status != nil && *status != types.ControlSuccess {
			if ws, ok := wireCommandStatus(*status); ok {
				debugLog("WriteAnalogOutput ERROR: %v (wire status %d)", err, ws)
				return outstation.NewControlStatusError(ws, err)
			}
		}
		debugLog("WriteAnalogOutput ERROR: %v", err)
		return err
	}
	// AO-002 F2: fail closed. Only an explicit ControlSuccess is success. A nil
	// status (handler returned (nil, nil)) or any non-success status must not be
	// coerced into success — otherwise the outstation echoes a SUCCESS command
	// status for a command the handler did not actually accept.
	if status == nil {
		debugLog("WriteAnalogOutput: nil status -> failure (not success)")
		return fmt.Errorf("command handler returned no status")
	}
	if *status != types.ControlSuccess {
		// MEXT-024: a handler that signals a non-success status with no error
		// must NOT surface as ControlSuccess. AO-004 Appendix T: carry the status
		// to the outstation so the refusal is echoed with the actual wire status
		// when one is defined; otherwise fail closed with a generic error.
		debugLog("WriteAnalogOutput: non-success status %d -> failure", *status)
		if ws, ok := wireCommandStatus(*status); ok {
			return outstation.NewControlStatusError(ws, fmt.Errorf("command rejected by handler: status %d", *status))
		}
		return fmt.Errorf("command rejected by handler: status %d", *status)
	}
	debugLog("WriteAnalogOutput result: status=%d", *status)
	return nil
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
	config             *Config
	state              ServerState
	mu                 sync.RWMutex
	dataHandler        DataHandler
	commandHandler     CommandHandler
	unsolicitedHandler UnsolicitedHandler

	// Listener for accepting connections
	listener net.Listener

	// Track active connections
	connections   map[uint64]*outstationInstance
	connCounter   uint64
	connectionsMu sync.RWMutex
	eventQueue    *outstation.EventQueue

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
	id         uint64
	masterAddr net.Addr
	outstation *outstation.Outstation
	transport  transport.Handler
	done       chan struct{}
}

// NewServer creates a new Outstation server with the given configuration
func NewServer(config *Config) (Server, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	maxEvents := config.MaxEventBuffers
	if maxEvents < 3 {
		maxEvents = 1000
	}
	return &server{
		config:             config,
		state:              ServerStateDown,
		dataHandler:        &DefaultDataHandler{},
		commandHandler:     nil,
		unsolicitedHandler: nil,
		connections:        make(map[uint64]*outstationInstance),
		eventQueue:         outstation.NewEventQueue(maxEvents),
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
		Address:        "",
		Port:           0,
		ConnectTimeout: 5000,
		ReceiveTimeout: 5000,
		KeepAlive:      30000,
	}
	t := transport.NewTCPTransport(tcpConfig)
	// Set the connection directly
	t.SetConn(conn)
	link := &sessionTCP{TCPTransport: t}

	// Create new outstation for this connection
	internalConfig := &outstation.Config{
		OutstationAddress: s.config.OutstationAddress,
		MasterAddress:     s.config.MasterAddress,
		Timeout:           5000,
		MaxEventBuffers:   s.config.MaxEventBuffers,
		Unsolicited:       s.config.UnsolicitedMode,
	}
	inst := outstation.NewOutstation(internalConfig)
	inst.SetTransport(link)
	inst.UseEventQueue(s.eventQueue)
	inst.Initialize()
	inst.SetDataHandler(&internalDataHandler{
		publicHandler:        s.dataHandler,
		commandHandler:       s.commandHandler,
		analogControlEnabled: s.config.ExperimentalAnalogControl,
	})
	inst.Start()

	instance := &outstationInstance{
		id:         connID,
		masterAddr: conn.RemoteAddr(),
		outstation: inst,
		transport:  t,
		done:       make(chan struct{}),
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
			publicHandler:        handler,
			commandHandler:       s.commandHandler,
			analogControlEnabled: s.config.ExperimentalAnalogControl,
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
// Supported-profile: Reject by default — the v0 outstation has no unsolicited
// delivery path, so no handler is registered unless the explicit, non-default
// experimental event delivery switch is on. The method has no error return, and
// a prohibited capability must never be silently activated, so it records the
// refusal instead of registering the handler.
func (s *server) SetUnsolicitedHandler(handler UnsolicitedHandler) {
	if !s.config.ExperimentalEventDelivery {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.unsolicitedHandler = handler
}

func (s *server) publish(ev outstation.Event) error {
	if !s.config.ExperimentalEventDelivery {
		return fmt.Errorf("%w: event publishing; the v0 supported profile does not deliver events", dnp3.ErrUnsupportedOption)
	}
	if err := outstation.ValidateEvent(ev); err != nil {
		return err
	}
	s.connectionsMu.RLock()
	var target *outstation.Outstation
	for _, inst := range s.connections {
		target = inst.outstation
		break
	}
	s.connectionsMu.RUnlock()
	if target != nil {
		return target.PublishEvent(ev)
	}
	if !s.eventQueue.Add(ev) {
		return outstation.ErrEventBufferFull
	}
	return nil
}

func (s *server) newEvent(class EventClass, index uint16, quality uint8, at time.Time) (outstation.Event, error) {
	if class != EventClass1 && class != EventClass2 && class != EventClass3 {
		return outstation.Event{}, outstation.ErrEventClass
	}
	ev := outstation.Event{
		Class:   outstation.EventClass(class),
		Index:   index,
		Quality: quality,
	}
	if !at.IsZero() {
		ev.HasTime = true
		ev.Time = at
	}
	return ev, nil
}

// PublishBinaryEvent implements Server.PublishBinaryEvent.
func (s *server) PublishBinaryEvent(index uint16, value bool, quality uint8, class EventClass, at time.Time) error {
	ev, err := s.newEvent(class, index, quality, at)
	if err != nil {
		return err
	}
	ev.Kind = outstation.EventKindBinary
	ev.Group = 2
	ev.BinaryValue = value
	return s.publish(ev)
}

// PublishAnalogEvent implements Server.PublishAnalogEvent.
func (s *server) PublishAnalogEvent(index uint16, value int32, quality uint8, class EventClass, at time.Time) error {
	ev, err := s.newEvent(class, index, quality, at)
	if err != nil {
		return err
	}
	ev.Kind = outstation.EventKindAnalog
	ev.Group = 32
	ev.AnalogValue = value
	return s.publish(ev)
}

// PublishCounterEvent implements Server.PublishCounterEvent.
func (s *server) PublishCounterEvent(index uint16, value uint32, quality uint8, class EventClass, at time.Time) error {
	ev, err := s.newEvent(class, index, quality, at)
	if err != nil {
		return err
	}
	ev.Kind = outstation.EventKindCounter
	ev.Group = 22
	ev.CounterValue = value
	return s.publish(ev)
}

// Close implements Server.Close.
// Supported-profile: Target — one TCP listener lifecycle.
func (s *server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.Stop(ctx)
}
