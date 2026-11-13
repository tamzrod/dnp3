// Package master provides the DNP3 Master client API.
//
// This package implements the Master Station role of the DNP3 protocol.
// A Master station initiates all communication with outstations.
//
// # v0 supported public surface (frozen)
//
// The v0 interoperability profile (see active_work/supported-profile.md)
// supports exactly one path: one TCP master opens one connection to one
// outstation, performs a Class-0 read of static Binary Input / Analog Input /
// Counter data, and sends one Direct-Operate CROB control. The public surface
// below is frozen for external use; "Target" methods are in the v0 path,
// "Reject" methods return a clear unsupported error until a later profile.
//
// Config (NewConfig options):
//
//   - WithMasterAddress          — Target (one master address)
//   - WithOutstationAddress      — Target (one outstation address)
//   - WithTransport(TCP, ...)    — Target (one TCP peer)
//   - WithTLS / WithTransport(TLS, ...) — Reject (no plaintext fallback / no TLS)
//   - WithTimeout, WithRetry     — Defer (present, not externally verified)
//
// Client methods:
//
//   - Connect / Disconnect / Close — Target (one connection lifecycle)
//   - State                        — Target
//   - Read                         — Target (Class-0 static G1.1, G30.1, G20.1 only)
//   - IntegrityPoll                — Target (convenience over the Target Read groups)
//   - Operate                      — Target for DirectOperate G12V1 CROB;
//     Reject for SelectThenOperate / DirectOperateNoResponse
//   - LastIIN                      — Target
//   - EnableUnsolicited / DisableUnsolicited / ReadEvents / WatchUnsolicited —
//     implemented for EVT-001. The published supported profile is unchanged
//     and does not advertise this path.
//
// Explicitly unsupported in v0 (rejected by the public API): TLS, serial
// transport, secure authentication,
// time synchronization and time objects, file transfer, device attributes,
// restart/delay-measurement/freeze operations, select-before-operate, and any
// object group/variation not listed above.
//
// # Usage
//
//	config := master.NewConfig(
//		master.WithMasterAddress(0xFFFF),
//		master.WithOutstationAddress(1024),
//		master.WithTransport(dnp3.TCP, "192.168.1.100", 20000),
//	)
//	client, err := master.NewClient(config)
//	if err != nil {
//		log.Fatal(err)
//	}
//	ctx := context.Background()
//	if err := client.Connect(ctx); err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
//	resp, err := client.IntegrityPoll(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//	// resp.BinaryInputs / resp.AnalogInputs / resp.Counters hold the Class-0 data.
//
//	_, err = client.Operate(ctx, types.NewBinaryControl(0, true, types.DirectOperate))
package master

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"dnp3/internal/al"
	"dnp3/internal/master"
	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/types"
	"dnp3/pkg/transport"
)

// Client represents a DNP3 Master client.
// The client connects to outstations and issues read/write/control commands.
type Client interface {
	// Connect establishes a connection to the outstation
	Connect(ctx context.Context) error

	// Disconnect closes the connection
	Disconnect(ctx context.Context) error

	// State returns the current connection state
	State() dnp3.ConnectionState

	// Read issues a read request to the outstation.
	// Supported-profile: Target — Class-0 static Groups 1.1, 30.1, and 20.1
	// only; other groups are rejected with ErrUnsupportedGroup (DNP3-029).
	Read(ctx context.Context, request *types.ReadRequest) (*ReadResponse, error)

	// IntegrityPoll issues a Class-0 integrity poll — a single Read of all
	// MVP-supported static groups (Binary Input G1, Counter G20, Analog Input
	// G30) in one multi-group request — and returns the same ReadResponse shape
	// as Read. It is a convenience around Read for the common "read everything"
	// case (DNP3-037). Since MEXT-015 the primary path is one multi-header
	// exchange, with a per-group fallback for peers that reject it.
	// Supported-profile: Target (convenience over Target Read groups).
	IntegrityPoll(ctx context.Context) (*ReadResponse, error)

	// LastIIN returns the most recent Internal Indications received from the
	// configured outstation. The same value is carried on each ReadResponse;
	// this accessor exposes the master's stored copy (DNP3-012).
	// Supported-profile: Target.
	LastIIN() [2]byte

	// Operate issues a control command to the outstation.
	// Supported-profile: Target for DirectOperate (Group 12 Variation 1 only);
	// Reject for SelectThenOperate and DirectOperateNoResponse (DNP3-030).
	Operate(ctx context.Context, command *types.ControlOutput) (*OperateResponse, error)

	// EnableUnsolicited asks the outstation to enable Class 1, 2, and 3
	// unsolicited reporting. The published supported profile is unchanged
	// and still does not advertise this behavior.
	EnableUnsolicited(ctx context.Context) error

	// DisableUnsolicited asks the outstation to disable Class 1, 2, and 3
	// unsolicited reporting.
	DisableUnsolicited(ctx context.Context) error

	// ReadEvents reads queued Class 1, 2, and 3 events with Group 60
	// Variations 2, 3, and 4. This is a solicited read. It is not a
	// substitute for unsolicited delivery.
	ReadEvents(ctx context.Context) (*UnsolicitedResponse, error)

	// WatchUnsolicited receives unsolicited fragments until ctx is canceled.
	// It must not be called concurrently with Read, ReadEvents, Operate, or
	// EnableUnsolicited on the same client. The handler must not call back
	// into the client.
	WatchUnsolicited(ctx context.Context) error

	// SetUnsolicitedHandler sets the handler for unsolicited responses.
	SetUnsolicitedHandler(handler UnsolicitedHandler)

	// Close gracefully shuts down the client.
	// Supported-profile: Target (one connection lifecycle).
	Close() error
}

// Config holds Master client configuration.
//
// Each field/option carries a supported-profile disposition (see
// active_work/supported-profile.md): Target (in the v0 path and externally
// verified later), Reject (returns a clear unsupported error until a later
// profile), or Defer (present but not yet externally verified).
type Config struct {
	// MasterAddress is this master's DNP3 address.
	// Supported-profile: Target — one configured master address.
	MasterAddress uint16
	// OutstationAddress is the target outstation's DNP3 address.
	// Supported-profile: Target — one configured outstation address.
	OutstationAddress uint16
	// TransportType specifies TCP or TLS.
	// Supported-profile: Target for TCP; Reject for TLS (WithTLS).
	TransportType dnp3.TransportType
	// Address is the network address (IP or hostname).
	// Supported-profile: Target (TCP peer address).
	Address string
	// Port is the network port.
	// Supported-profile: Target (TCP peer port).
	Port int
	// TLSConfig is the TLS configuration (required for TLS transport).
	// Supported-profile: Reject — no TLS listener in v0; WithTLS is rejected.
	TLSConfig *tls.Config
	// Timeout is the response timeout.
	// Supported-profile: Defer — requires lifecycle/timeout verification.
	Timeout time.Duration
	// RetryCount is the maximum number of retries.
	// Supported-profile: Defer — requires lifecycle/timeout verification.
	RetryCount int
	// RetryDelay is the delay between retries.
	// Supported-profile: Defer — requires lifecycle/timeout verification.
	RetryDelay time.Duration
	// KeepAliveInterval is the keep-alive interval.
	// Supported-profile: Defer — not externally verified in v0.
	KeepAliveInterval time.Duration
	// IdleTimeout is the optional keep-alive/idle threshold (DNP3-042). When
	// positive, a background monitor closes the session (state → Disconnected)
	// if no send/receive activity is observed for this long. Zero or negative
	// disables idle monitoring (the default).
	// Supported-profile: Defer — optional, not externally verified in v0.
	IdleTimeout time.Duration
	// Logger is the optional diagnostic sink for frame/sequence events
	// (DNP3-044). nil means silent (the default); a non-nil Logger receives
	// structured events from the master (send/receive/confirm/retry/state).
	// Supported-profile: Defer — observability hook, not externally verified in v0.
	Logger Logger
	// AutoIntegrityOnRestart enables the DNP3-053 behavior: when a response
	// carries the DeviceRestart IIN bit (IIN1.7), the master automatically
	// issues a Class-0 integrity poll after the triggering request completes,
	// so the outstation's post-restart static data is re-read without the
	// caller having to poll manually. Disabled by default (opt-in). The
	// auto-poll reuses the normal request path; it is suppressed while another
	// auto-integrity poll is already in flight to avoid unbounded recursion.
	// Supported-profile: Defer — optional, not externally verified in v0.
	AutoIntegrityOnRestart bool
	// ConnectTimeout is the whole-connect budget (VEC-01 B4 F2): dial plus link
	// reset/status and initialization. Zero uses the internal default
	// (DefaultConnectTimeout, 5 s). The caller's earlier context deadline always
	// tightens it. Supported-profile: Defer — requires lifecycle verification.
	ConnectTimeout time.Duration
	// OperateTimeout is the whole-control budget (VEC-01 B4 F2) for the single
	// supported G12V1 DirectOperate: send, wait and status interpretation. Zero
	// uses the internal default (DefaultOperateTimeout, 5 s). The caller's earlier
	// context deadline always tightens it.
	// Supported-profile: Defer — requires lifecycle verification.
	OperateTimeout time.Duration
	// ReadTimeout is the whole-request budget (VEC-01 B4 F2) for a Class-0 read,
	// including bounded retries and, for an integrity poll, the shared per-group
	// fallback. Zero uses the internal default (DefaultReadTimeout, 10 s). The
	// caller's earlier context deadline always tightens it.
	// Supported-profile: Defer — requires lifecycle verification.
	ReadTimeout time.Duration
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		MasterAddress:     0xFFFF,
		OutstationAddress: 1024,
		TransportType:     dnp3.TCP,
		Address:           "localhost",
		Port:              20000,
		Timeout:           5 * time.Second,
		RetryCount:        3,
		RetryDelay:        100 * time.Millisecond,
		KeepAliveInterval: 30 * time.Second,
	}
}

// ConfigOption is a functional option for configuring the client
type ConfigOption func(*Config)

// WithMasterAddress sets the master address.
// Supported-profile: Target — one configured master address.
func WithMasterAddress(addr uint16) ConfigOption {
	return func(c *Config) {
		c.MasterAddress = addr
	}
}

// WithOutstationAddress sets the outstation address.
// Supported-profile: Target — one configured outstation address.
func WithOutstationAddress(addr uint16) ConfigOption {
	return func(c *Config) {
		c.OutstationAddress = addr
	}
}

// WithTransport sets the transport type and network address.
// Supported-profile: Target for TCP (one TCP peer); Reject for TLS — use of
// dnp3.TLS here is rejected at NewClient until a later profile (see WithTLS).
func WithTransport(t dnp3.TransportType, address string, port int) ConfigOption {
	return func(c *Config) {
		c.TransportType = t
		c.Address = address
		c.Port = port
	}
}

// WithTLS sets the TLS configuration.
// Supported-profile: Reject — no TLS listener in v0; NewClient returns
// ErrUnsupportedOption for TLS until a later profile.
func WithTLS(config *tls.Config) ConfigOption {
	return func(c *Config) {
		c.TLSConfig = config
		c.TransportType = dnp3.TLS
	}
}

// WithTimeout sets the response timeout.
// Supported-profile: Defer — requires lifecycle/timeout verification.
func WithTimeout(timeout time.Duration) ConfigOption {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithRetry sets the retry count and delay.
// Supported-profile: Defer — requires lifecycle/timeout verification.
func WithRetry(count int, delay time.Duration) ConfigOption {
	return func(c *Config) {
		c.RetryCount = count
		c.RetryDelay = delay
	}
}

// WithIdleTimeout sets the optional idle/keep-alive threshold (DNP3-042). When
// positive, a background monitor closes the session (state → Disconnected) if
// no send/receive activity is observed for this long. Zero or negative disables
// idle monitoring (the default).
// Supported-profile: Defer — optional, not externally verified in v0.
func WithIdleTimeout(d time.Duration) ConfigOption {
	return func(c *Config) {
		c.IdleTimeout = d
	}
}

// WithLogger installs the optional diagnostic logger for frame/sequence events
// (DNP3-044). The default is silent (no-op); pass a [Logger] to observe
// send/receive/confirm/retry/state events from the master. The logger must be
// safe for concurrent use.
// Supported-profile: Defer — observability hook, not externally verified in v0.
func WithLogger(l Logger) ConfigOption {
	return func(c *Config) {
		c.Logger = l
	}
}

// WithAutoIntegrityOnRestart enables the DNP3-053 behavior: a response carrying
// the DeviceRestart IIN bit automatically triggers a Class-0 integrity poll
// after the triggering request completes. See Config.AutoIntegrityOnRestart.
// Supported-profile: Defer — optional, not externally verified in v0.
func WithAutoIntegrityOnRestart() ConfigOption {
	return func(c *Config) {
		c.AutoIntegrityOnRestart = true
	}
}

// WithConnectTimeout sets the whole-connect budget (VEC-01 B4 F2). A
// non-positive value restores the internal default (5 s).
func WithConnectTimeout(d time.Duration) ConfigOption {
	return func(c *Config) {
		c.ConnectTimeout = d
	}
}

// WithOperateTimeout sets the whole-control budget (VEC-01 B4 F2) for the
// single supported G12V1 DirectOperate. A non-positive value restores the
// internal default (5 s).
func WithOperateTimeout(d time.Duration) ConfigOption {
	return func(c *Config) {
		c.OperateTimeout = d
	}
}

// WithReadTimeout sets the whole-request budget (VEC-01 B4 F2) for a Class-0
// read, including bounded retries and the shared integrity-poll fallback. A
// non-positive value restores the internal default (10 s).
func WithReadTimeout(d time.Duration) ConfigOption {
	return func(c *Config) {
		c.ReadTimeout = d
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

// Validate validates the configuration.
//
// DNP3-049: link-layer addresses are validated for the v0 single-outstation
// master profile:
//   - MasterAddress (the master's source address) must not be the reserved
//     null address 0x0000. The broadcast address 0xFFFF is permitted as a
//     master source (it is the established default and convention in this
//     codebase).
//   - OutstationAddress (the single configured destination) must not be the
//     reserved null address 0x0000, and must not be the broadcast address
//     0xFFFF: the MVP profile targets exactly one outstation, so a broadcast
//     destination is inappropriate.
func (c *Config) Validate() error {
	if c.Timeout <= 0 {
		return &dnp3.ConfigurationError{
			Field:   "Timeout",
			Message: "must be positive",
		}
	}
	if c.RetryCount < 0 {
		return &dnp3.ConfigurationError{
			Field:   "RetryCount",
			Message: "must be non-negative",
		}
	}
	if c.RetryDelay < 0 {
		return &dnp3.ConfigurationError{
			Field:   "RetryDelay",
			Message: "must be non-negative",
		}
	}
	if c.TransportType == dnp3.TLS && c.TLSConfig == nil {
		return &dnp3.ConfigurationError{
			Field:   "TLSConfig",
			Message: "required for TLS transport",
		}
	}
	// DNP3-049: address validation.
	if c.MasterAddress == 0x0000 {
		return &dnp3.ConfigurationError{
			Field:   "MasterAddress",
			Message: "0x0000 is reserved and not a valid master source address",
		}
	}
	if c.OutstationAddress == 0x0000 {
		return &dnp3.ConfigurationError{
			Field:   "OutstationAddress",
			Message: "0x0000 is reserved and not a valid outstation address",
		}
	}
	if c.OutstationAddress == 0xFFFF {
		return &dnp3.ConfigurationError{
			Field:   "OutstationAddress",
			Message: "0xFFFF broadcast is not a valid single-outstation destination",
		}
	}
	return nil
}

// UnsolicitedHandler is called when an unsolicited response is received
type UnsolicitedHandler func(response *UnsolicitedResponse)

// ReadResponse contains the response data from a read request.
//
// DNP3-035: the v0 MVP profile supports only the Class-0 static groups
// Binary Input (G1), Analog Input (G30), and Counter (G20). Only those three
// slices are populated by Read. The output-status and frozen-counter fields
// are retained for forward compatibility with later profiles but are NOT
// populated by the v0 Read path; callers must not rely on them being non-nil.
type ReadResponse struct {
	// IIN is the Internal Indication from the outstation
	IIN [2]byte
	// BinaryInputs contains binary input data (Group 1, MVP-supported).
	BinaryInputs []*types.BinaryInput
	// AnalogInputs contains analog input data (Group 30, MVP-supported).
	AnalogInputs []*types.AnalogInput
	// Counters contains counter data (Group 20, MVP-supported).
	Counters []*types.Counter
	// BinaryOutputs contains binary output status data (Group 10). NOT
	// populated by the v0 MVP Read path (DNP3-035); deferred to a later
	// profile.
	BinaryOutputs []*types.BinaryOutput
	// AnalogOutputs contains analog output status data (Group 40). NOT
	// populated by the v0 MVP Read path (DNP3-035); deferred to a later
	// profile.
	AnalogOutputs []*types.AnalogOutput
	// FrozenCounters contains frozen counter data. NOT populated by the v0
	// MVP Read path (DNP3-035); deferred to a later profile.
	FrozenCounters []*types.FrozenCounter
	// Timestamp is when the response was received
	Timestamp time.Time
}

// OperateResponse contains the response from a control operation
type OperateResponse struct {
	// IIN is the Internal Indication from the outstation
	IIN [2]byte
	// Status is the command status
	Status types.ControlStatus
	// Timestamp is when the response was received
	Timestamp time.Time
}

// UnsolicitedResponse contains data from an unsolicited response
type UnsolicitedResponse struct {
	// IIN is the Internal Indication from the outstation
	IIN [2]byte
	// BinaryInputs contains binary input event data
	BinaryInputs []*types.BinaryInput
	// AnalogInputs contains analog input event data
	AnalogInputs []*types.AnalogInput
	// Counters contains counter event data
	Counters []*types.Counter
	// Timestamp is when the response was received
	Timestamp time.Time
}

// transportAdapter wraps the public transport.Handler to satisfy internal/master.TransportHandler
type transportAdapter struct {
	transport.Handler
}

// client is the internal implementation of Client
type client struct {
	config   *Config
	state    dnp3.ConnectionState
	mu       sync.RWMutex
	handlers []UnsolicitedHandler

	// Internal implementation
	internalMaster *master.Master
	transport      transport.Handler

	// For response parsing
	sequence uint8

	// logger is the optional diagnostic sink for frame/sequence events
	// (DNP3-044). nil means silent (the default).
	logger Logger

	// autoIntegrityActive guards the DNP3-053 auto-integrity-on-restart path
	// against re-entrancy: while an auto integrity poll is in flight, the
	// Reads it issues must not themselves trigger a nested auto-poll even if
	// their responses also carry the DeviceRestart bit. Guarded by c.mu.
	autoIntegrityActive bool

	// connectGen is bumped at each public Connect entry (VEC-01 B4 F2-R2). A
	// connect attempt is identified by the generation it captured at entry; a
	// superseded attempt must not mutate client state or place a late socket
	// into a replacement session. Guarded by c.mu.
	connectGen uint64

	// connectWorkerOutstanding is true while a public Connect worker goroutine
	// has not yet terminated (VEC-01 B4 F2-R3/F2-R4). It covers the whole
	// connect attempt INCLUDING its cleanup, so a fresh Connect is refused
	// fail-closed with ErrSessionUnresolved while an old worker may still hold
	// the shared transport/master; a stale worker's late handshake,
	// initialization or cleanup cannot affect a replacement session. Guarded by
	// c.mu.
	connectWorkerOutstanding bool

	// teardown is non-nil while an asynchronous session teardown is outstanding
	// (VEC-01 B4 F2-R4/F3-R4). The teardown (internal master Disconnect +
	// transport Close) may block on an uncooperative transport, so it runs on its
	// own goroutine and the public callers return promptly; while it is
	// outstanding the previous transport teardown and any admitted control send
	// may still be in flight, so a replacement Connect is refused fail-closed
	// (ErrSessionUnresolved) and a concurrent Disconnect joins the same teardown
	// instead of starting a second one. Cleared only when the teardown goroutine
	// has terminated. Guarded by c.mu.
	teardown *teardown
}

// sendAndReceive sends a request and receives a response, returning parsed data
func (c *client) sendAndReceive(data []byte) ([]byte, error) {
	// Send request
	if err := c.transport.Send(data); err != nil {
		return nil, fmt.Errorf("send failed: %w", err)
	}

	// Receive response with timeout
	c.transport.SetTimeout(int(c.config.Timeout.Milliseconds()))
	resp, err := c.transport.Receive()
	if err != nil {
		return nil, fmt.Errorf("receive failed: %w", err)
	}

	return resp, nil
}

// NewClient creates a new Master client with the given configuration
func NewClient(config *Config) (Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Create internal master
	internalConfig := &master.Config{
		MasterAddress:  config.MasterAddress,
		Timeout:        int(config.Timeout.Milliseconds()),
		MaxRetries:     config.RetryCount,
		RetryDelay:     int(config.RetryDelay.Milliseconds()),
		IdleTimeout:    int(config.IdleTimeout.Milliseconds()),
		ConnectTimeout: int(config.ConnectTimeout.Milliseconds()),
		OperateTimeout: int(config.OperateTimeout.Milliseconds()),
		ReadTimeout:    int(config.ReadTimeout.Milliseconds()),
	}
	internal := master.NewMaster(internalConfig)

	// Create transport based on type
	var t transport.Handler
	switch config.TransportType {
	case dnp3.TCP:
		tcpConfig := &transport.TCPConfig{
			Address:        config.Address,
			Port:           config.Port,
			ConnectTimeout: int(config.Timeout.Milliseconds()),
			ReceiveTimeout: int(config.Timeout.Milliseconds()),
			KeepAlive:      int(config.KeepAliveInterval.Milliseconds()),
		}
		t = transport.NewTCPTransport(tcpConfig)
	case dnp3.TLS:
		return nil, fmt.Errorf("%w: TLS transport", dnp3.ErrUnsupportedOption)
	}

	// Set up internal master with transport
	internal.SetTransport(&transportAdapter{Handler: t})
	// DNP3-044: wire the optional diagnostic logger to the internal master so
	// frame/sequence events flow through the public Logger. nil logger leaves
	// the master silent (the default).
	if config.Logger != nil {
		internal.SetDiagnosticHook(diagAdapter(config.Logger))
	}

	c := &client{
		config:         config,
		state:          dnp3.StateDisconnected,
		handlers:       make([]UnsolicitedHandler, 0),
		internalMaster: internal,
		transport:      t,
		logger:         config.Logger,
	}
	internal.SetUnsolicitedCallback(c.deliverUnsolicitedApp)
	return c, nil
}

// NewClientWithTransport creates a Master client that uses a caller-supplied
// transport.Handler instead of a TCP/TLS connection. It is intended for tests
// and laboratory use — e.g. driving the full public Connect → Read → Operate
// flow against an in-memory outstation simulator (DNP3-036) with no network
// I/O. The config's TransportType/Address/Port are ignored; Timeout/Retry
// still apply.
func NewClientWithTransport(config *Config, t transport.Handler) (Client, error) {
	if t == nil {
		return nil, fmt.Errorf("%w: transport is nil", dnp3.ErrInvalidRequest)
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}

	internalConfig := &master.Config{
		MasterAddress: config.MasterAddress,
		Timeout:       int(config.Timeout.Milliseconds()),
		MaxRetries:    config.RetryCount,
		RetryDelay:    int(config.RetryDelay.Milliseconds()),
		IdleTimeout:   int(config.IdleTimeout.Milliseconds()),
	}
	internal := master.NewMaster(internalConfig)
	internal.SetTransport(&transportAdapter{Handler: t})
	// DNP3-044: wire the optional diagnostic logger (see NewClient).
	if config.Logger != nil {
		internal.SetDiagnosticHook(diagAdapter(config.Logger))
	}

	c := &client{
		config:         config,
		state:          dnp3.StateDisconnected,
		handlers:       make([]UnsolicitedHandler, 0),
		internalMaster: internal,
		transport:      t,
		logger:         config.Logger,
	}
	internal.SetUnsolicitedCallback(c.deliverUnsolicitedApp)
	return c, nil
}

// Connect implements Client.Connect
//
// Proper order:
// 1. Add outstation (needed for handshake destinations)
// 2. Transport connect (TCP)
// 3. Internal master connect (handshake with registered outstations)
//
// The caller's context is honored throughout (DNP3-022): if ctx is cancelled
// before or during connect, Connect returns promptly with ErrContextCanceled
// and tears down any partially-established connection so none is left live.
func (c *client) Connect(ctx context.Context) error {
	// Already-cancelled context: fail fast before mutating state.
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("%w: %v", dnp3.ErrContextCanceled, err)
	}

	// VEC-01 B4 F2-R3: claim the single connect-worker slot atomically. A fresh
	// Connect is refused while a previous Connect worker has not yet terminated,
	// because that worker may still hold the shared transport/master (e.g. blocked
	// in an uncooperative dial); a stale worker's late handshake, initialization or
	// cleanup must not be able to affect a replacement session. A prior control
	// send from a torn-down session (still possibly on the wire) is refused here
	// before dialing; the internal master enforces the same fence again at
	// ConnectContext.
	c.mu.Lock()
	if c.state != dnp3.StateDisconnected {
		// Capture the state under the lock: reading c.state again in the error
		// message after unlocking would race a concurrent Disconnect's
		// publication (VEC-01 B4 F2-R5).
		st := c.state
		c.mu.Unlock()
		return fmt.Errorf("already connected: %s", st)
	}
	if c.connectWorkerOutstanding {
		c.mu.Unlock()
		return fmt.Errorf("%w: prior connect worker outstanding", master.ErrSessionUnresolved)
	}
	// VEC-01 B4 F2-R4/F3-R4: refuse a replacement while a previous session
	// teardown is still outstanding. The teardown (internal master Disconnect +
	// transport Close) may be blocked on an uncooperative transport, and an
	// admitted control send may still be in flight on the old transport; a fresh
	// dial must not start against that half-torn-down shared transport. The fence
	// clears only when the teardown goroutine has actually terminated.
	if c.teardown != nil {
		c.mu.Unlock()
		return fmt.Errorf("%w: prior session teardown outstanding", master.ErrSessionUnresolved)
	}
	if c.internalMaster != nil && c.internalMaster.HasUnresolvedSession() {
		c.mu.Unlock()
		return fmt.Errorf("%w: prior control worker outstanding", master.ErrSessionUnresolved)
	}
	c.state = dnp3.StateConnecting
	c.connectWorkerOutstanding = true
	// VEC-01 B4 F2-R2: identify this attempt so a superseded attempt cannot commit
	// success or leave a late socket in a replacement session.
	c.connectGen++
	gen := c.connectGen
	c.mu.Unlock()

	// VEC-01 B4 F2: establish ONE absolute whole-connect deadline at entry,
	// constrained by the caller's earlier deadline and the configured
	// ConnectTimeout. The dial, the two-stage link handshake, initialization and
	// bounded cleanup all share this single remaining budget; the internal master
	// must not start a fresh full budget after the dial.
	deadline := time.Now().Add(c.connectBudget())
	if dl, ok := ctx.Deadline(); ok && dl.Before(deadline) {
		deadline = dl
	}

	// Run the blocking connect steps on a goroutine so we can race them against
	// ctx.Done() AND the whole-connect deadline. The goroutine owns the
	// transport/master for the duration of the attempt. It clears the worker-
	// outstanding flag exactly when it terminates — the point at which it no
	// longer touches shared resources — so a replacement Connect is unblocked only
	// then (VEC-01 B4 F2-R3).
	type connectResult struct{ err error }
	done := make(chan connectResult, 1)
	go func() {
		// Clear the worker slot when the worker terminates (including on a panic),
		// i.e. once it no longer touches the shared transport/master. The slot is
		// held across connectBlocking AND its cleanup so a replacement Connect
		// cannot start against a half-torn-down transport (VEC-01 B4 F2-R3/F2-R4).
		defer func() {
			c.mu.Lock()
			c.connectWorkerOutstanding = false
			c.mu.Unlock()
		}()
		done <- connectResult{err: c.connectBlocking(ctx, gen, deadline)}
	}()

	// The attempt is bounded by the earlier of the caller's context and the
	// whole-connect deadline. Bounding on the deadline (not only ctx.Done())
	// makes Connect fail timely even for context.Background() and an
	// uncooperative transport whose dial blocks beyond budget. Teardown is
	// best-effort and the superseded worker stays fenced by its attempt
	// generation (and the worker-outstanding flag) so it cannot place a late
	// socket into a replacement session.
	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		c.abortConnect(gen)
		return fmt.Errorf("%w: %v", dnp3.ErrContextCanceled, ctx.Err())
	case <-timer.C:
		c.abortConnect(gen)
		return fmt.Errorf("%w: connect budget exhausted", dnp3.ErrTimeout)
	case res := <-done:
		c.mu.Lock()
		if gen != c.connectGen {
			// A newer attempt (or an explicit Disconnect) superseded this one;
			// do not commit its result.
			c.mu.Unlock()
			c.emitLog(LogWarn, "connect", "superseded connect attempt", nil)
			return fmt.Errorf("%w: connect attempt superseded", dnp3.ErrContextCanceled)
		}
		if res.err != nil {
			c.state = dnp3.StateError
		} else {
			c.state = dnp3.StateActive
			c.internalMaster.ResetUnsolicitedSession()
		}
		c.mu.Unlock()
		if res.err != nil {
			c.emitLog(LogError, "connect", "connect failed", res.err)
			return res.err
		}
		c.emitLog(LogInfo, "connect", "connected", nil)
		return nil
	}
}

// connectBlocking performs the synchronous connect sequence. It checks ctx
// between blocking steps so a cancellation that lands between the transport
// dial and the link handshake is observed without waiting for the next timeout.
// deadline is the single absolute whole-connect deadline established at the
// public Connect entry (VEC-01 B4 F2); every step shares its remaining budget.
//
// gen is the attempt generation captured at the public Connect entry (VEC-01 B4
// F2-R3). Every mutating step and the failure cleanup are guarded by it: once a
// newer attempt (or an explicit Disconnect) supersedes this one, the worker
// returns WITHOUT mutating the shared transport/master, so a stale worker's late
// handshake, initialization or cleanupConnect cannot affect a replacement
// session.
func (c *client) connectBlocking(ctx context.Context, gen uint64, deadline time.Time) error {
	// A superseded attempt must not touch shared state at all.
	if !c.attemptCurrent(gen) {
		return fmt.Errorf("%w: connect attempt superseded", dnp3.ErrContextCanceled)
	}

	// Step 1: Add outstation FIRST (needed for link handshake)
	c.internalMaster.AddOutstation(c.config.OutstationAddress, fmt.Sprintf("outstation-%d", c.config.OutstationAddress))

	// fail tears the attempt down only while it is still current; a superseded
	// worker must not run cleanupConnect (that would tear down the replacement).
	fail := func(err error) error {
		c.cleanupConnectIfCurrent(gen)
		return err
	}

	if err := ctx.Err(); err != nil {
		return fail(fmt.Errorf("%w: %v", dnp3.ErrContextCanceled, err))
	}
	if !time.Now().Before(deadline) {
		return fail(fmt.Errorf("%w: connect budget exhausted before dial", dnp3.ErrTimeout))
	}
	if !c.attemptCurrent(gen) {
		return fmt.Errorf("%w: connect attempt superseded", dnp3.ErrContextCanceled)
	}

	// Step 2: Connect transport
	if err := c.transport.Connect(); err != nil {
		return fail(fmt.Errorf("transport connect failed: %w", err))
	}

	if err := ctx.Err(); err != nil {
		return fail(fmt.Errorf("%w: %v", dnp3.ErrContextCanceled, err))
	}
	if !c.attemptCurrent(gen) {
		return fmt.Errorf("%w: connect attempt superseded", dnp3.ErrContextCanceled)
	}

	// Step 3: Connect internal master (performs link-layer handshake). VEC-01 B4
	// F2: pass the REMAINING whole-connect budget (not a fresh one) so the dial
	// and the link handshake share one finite budget. ConnectContext tightens it
	// further by the caller's earlier deadline.
	remaining := time.Until(deadline)
	if remaining <= 0 {
		return fail(fmt.Errorf("%w: connect budget exhausted before handshake", dnp3.ErrTimeout))
	}
	if err := c.internalMaster.ConnectContext(ctx, remaining); err != nil {
		return fail(fmt.Errorf("master connect failed: %w", err))
	}

	if err := ctx.Err(); err != nil {
		return fail(fmt.Errorf("%w: %v", dnp3.ErrContextCanceled, err))
	}
	if !c.attemptCurrent(gen) {
		return fmt.Errorf("%w: connect attempt superseded", dnp3.ErrContextCanceled)
	}

	// Step 4: Initialize master
	if err := c.internalMaster.Initialize(); err != nil {
		return fail(fmt.Errorf("master initialize failed: %w", err))
	}
	return nil
}

// attemptCurrent reports whether the connect attempt identified by gen is still
// the current one (VEC-01 B4 F2-R3). A superseded attempt must not mutate shared
// state. Guarded by c.mu.
func (c *client) attemptCurrent(gen uint64) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return gen == c.connectGen
}

// abortConnect is invoked when Connect observes ctx cancellation or the
// whole-connect budget being exhausted. It bumps the attempt generation so any
// in-flight worker is superseded (its late completion cannot commit), publishes
// Disconnected, and starts the session teardown asynchronously so the public
// Connect returns promptly even when an uncooperative transport's Close blocks
// indefinitely (VEC-01 B4 F2-R4).
//
// The action is guarded by the attempt generation. If a newer attempt (or an
// explicit Disconnect) already superseded this one, that actor now owns the
// state and the shared transport/master, so this stale attempt must neither
// clobber the state nor run a teardown that would sabotage the replacement
// session.
func (c *client) abortConnect(gen uint64) {
	c.mu.Lock()
	if gen != c.connectGen {
		c.mu.Unlock()
		return
	}
	c.connectGen++
	c.state = dnp3.StateDisconnected
	// VEC-01 B4 F2-R5: register the teardown fence in the SAME critical section
	// that publishes Disconnected. Otherwise a fresh Connect could observe
	// Disconnected, be admitted, and claim the shared transport/master before
	// this teardown was registered; the late cleanup would then close or mutate
	// the replacement session. Registering the fence atomically with the state
	// transition that admits a replacement closes that window. The physical
	// teardown runs on the goroutine started by registerTeardownLocked; this
	// caller does NOT wait for it, so the public Connect returns promptly even
	// when an uncooperative transport's Close blocks indefinitely (F2-R4).
	c.registerTeardownLocked()
	c.mu.Unlock()
}

// teardown tracks one asynchronous session teardown (VEC-01 B4 F2-R4/F3-R4).
type teardown struct {
	done chan struct{}
	err  error
}

// registerTeardownLocked publishes the shared teardown fence and returns its
// handle, starting the teardown goroutine only once. It MUST be called with
// c.mu held so the fence becomes visible in the SAME critical section as the
// state transition (abortConnect / Disconnect / markTransportDisconnected) that
// otherwise admits a replacement Connect (VEC-01 B4 F2-R5). If a teardown is
// already registered, the existing handle is returned and no second goroutine
// starts: there is exactly one teardown owner, so transport Close is never
// duplicated. The caller releases c.mu; the teardown completes on its own
// goroutine.
func (c *client) registerTeardownLocked() *teardown {
	if c.teardown != nil {
		return c.teardown
	}
	t := &teardown{done: make(chan struct{})}
	c.teardown = t
	go c.teardownLoop(t)
	return t
}

// teardownLoop performs the physical session teardown — internal master
// Disconnect followed by transport Close — and then clears the fence, but ONLY
// if it still owns it. A superseding teardown that replaced c.teardown must not
// be cleared by an older, slower goroutine. The fence is cleared before t.done
// is closed so a goroutine woken by <-t.done observes the cleared fence.
func (c *client) teardownLoop(t *teardown) {
	// The internal teardown first: it is bounded and bumps the session
	// generation under safetyMu, fencing any admitted control send. The
	// transport Close may block on an uncooperative transport.
	_ = c.internalMaster.Disconnect()
	err := c.transport.Close()
	c.mu.Lock()
	if c.teardown == t {
		c.teardown = nil
	}
	t.err = err
	c.mu.Unlock()
	close(t.done)
}

// cleanupConnectIfCurrent tears the attempt down only while it is still the
// current attempt (VEC-01 B4 F2-R3/F2-R4). A superseded worker must NOT run the
// teardown: its transport/master have already been handed to the replacement
// attempt, and tearing them down here would sabotage that session. The teardown
// is asynchronous so a blocked transport Close cannot wedge the worker; the
// teardown fence keeps a replacement Connect refused until it terminates.
func (c *client) cleanupConnectIfCurrent(gen uint64) {
	c.mu.Lock()
	if gen != c.connectGen {
		c.mu.Unlock()
		return
	}
	// Register the teardown fence atomically with the generation check, so a
	// superseding actor that owns the shared transport/master is never torn down
	// by this stale worker. The physical teardown runs on its own goroutine; this
	// worker does NOT wait for it, so a blocked Close cannot wedge the worker.
	c.registerTeardownLocked()
	c.mu.Unlock()
}

// Disconnect implements Client.Disconnect
//
// The caller's context is honored (DNP3-024): if ctx is cancelled before the
// teardown completes, Disconnect returns promptly with ErrContextCanceled. The
// client state is reset to Disconnected immediately (VEC-01 B4 F3-R4), but the
// session is NOT left reusable: while the asynchronous teardown and any admitted
// control send remain unresolved, a replacement Connect is refused fail-closed
// (the client teardown fence plus the internal-master unresolved-session fence)
// so no replacement dials or sends against a half-torn-down transport.
func (c *client) Disconnect(ctx context.Context) error {
	// VEC-01 B4 F2-R2: supersede any in-flight connect attempt so a late
	// completion cannot commit Active after an explicit teardown.
	c.mu.Lock()
	c.connectGen++
	c.state = dnp3.StateDisconnected
	c.sequence = 0 // DNP3-050: clean slate for a later reconnect.
	// VEC-01 B4 F2-R5: register the teardown fence in the SAME critical section
	// as the Disconnected publication, so a fresh Connect can never be admitted
	// into the window before the old session's teardown is registered.
	t := c.registerTeardownLocked()
	c.mu.Unlock()

	// The teardown is already registered; the physical cleanup runs outside the
	// lock. A pre-cancelled context still tears down (no live session survives)
	// and returns promptly without blocking on the potentially unbounded Close.
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("%w: %v", dnp3.ErrContextCanceled, err)
	}

	select {
	case <-ctx.Done():
		// Return promptly; the teardown continues in the background and its fence
		// keeps a replacement Connect refused until it terminates.
		return fmt.Errorf("%w: %v", dnp3.ErrContextCanceled, ctx.Err())
	case <-t.done:
		if t.err != nil {
			return fmt.Errorf("disconnect failed: %w", t.err)
		}
		return nil
	}
}

// State implements Client.State
func (c *client) State() dnp3.ConnectionState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

// connectBudget resolves the whole-connect budget (VEC-01 B4 F2): the configured
// ConnectTimeout when positive, else the master's documented default
// (master.DefaultConnectTimeout). This keeps the public Connect entry consistent
// with the internal master's default rather than treating a zero value as a
// zero-length budget.
func (c *client) connectBudget() time.Duration {
	if c.config != nil && c.config.ConnectTimeout > 0 {
		return c.config.ConnectTimeout
	}
	return time.Duration(master.DefaultConnectTimeout) * time.Millisecond
}

// markTransportDisconnected records that the underlying transport has dropped.
// It flips the public state to Disconnected and tears down the internal master
// + transport so a subsequent Connect re-dials a fresh socket rather than
// retrying the dead one (MEXT-025: no stuck state after a peer drop). The
// teardown is best-effort — the caller already holds a transport error — and
// runs without holding c.mu to avoid lock-ordering issues with the transport.
func (c *client) markTransportDisconnected() {
	c.mu.Lock()
	if c.state == dnp3.StateDisconnected {
		c.mu.Unlock()
		return
	}
	c.state = dnp3.StateDisconnected
	c.sequence = 0 // DNP3-050: clean slate so a reconnect starts at AC seq 0.
	// VEC-01 B4 F2-R5: register the teardown fence atomically with the
	// Disconnected publication so a fresh Connect cannot be admitted before the
	// dropped session's teardown is registered. The physical teardown runs on
	// the goroutine started by registerTeardownLocked; this caller does not wait
	// for it, so a transport whose Close blocks cannot wedge the caller.
	c.registerTeardownLocked()
	c.mu.Unlock()
}

// Read implements Client.Read
func (c *client) Read(ctx context.Context, request *types.ReadRequest) (*ReadResponse, error) {
	return c.readWithDeadline(ctx, request, time.Time{})
}

// readWithDeadline is the deadline-aware read implementation (VEC-01 B4 F2).
// When deadline is non-zero it bounds the whole request/response exchange,
// including bounded retries; a zero deadline defers to the master's configured
// read budget. The public Read entry point passes a zero deadline.
func (c *client) readWithDeadline(ctx context.Context, request *types.ReadRequest, deadline time.Time) (*ReadResponse, error) {
	if request == nil || len(request.Groups) == 0 {
		return nil, dnp3.ErrInvalidRequest
	}

	// v0 profile allow-list (DNP3-029): only Binary Input (G1), Counter (G20),
	// and Analog Input (G30) may be read. Variation 0 ("any") is accepted for
	// these groups; any other group/variation is rejected with a clear error
	// before any wire traffic is generated.
	for _, g := range request.Groups {
		if !isSupportedReadGroup(g.Group, g.Variation) {
			return nil, fmt.Errorf("%w: %s", dnp3.ErrUnsupportedGroup, al.DescribeObject(g.Group, g.Variation))
		}
	}

	c.mu.RLock()
	state := c.state
	c.mu.RUnlock()

	if state < dnp3.StateConnected {
		return nil, dnp3.ErrNotConnected
	}

	// VEC-01 B4 F2: the caller's earlier deadline, when present, tightens the
	// read budget so the whole exchange (all retries) cannot outlive it.
	if dl, ok := ctx.Deadline(); ok && (deadline.IsZero() || dl.Before(deadline)) {
		deadline = dl
	}

	// Wire to internal master for proper protocol handling
	outstationID := c.config.OutstationAddress

	// Build APDU. The public sequence counter is guarded by c.mu so concurrent
	// Reads do not race on it (DNP3-025). The internal master also assigns the
	// wire sequence for retry handling.
	c.mu.Lock()
	seq := c.sequence
	c.sequence = (c.sequence + 1) % 16
	c.mu.Unlock()
	apdu := buildReadRequest(seq, request)

	// Use internal master's sendWithRetryAndGetResponse which:
	// 1. Sends the request with retry logic
	// 2. Waits for and processes the response
	// 3. Returns the processed application layer data
	//
	// The caller's context is honored (DNP3-023): if ctx is cancelled while
	// the request is outstanding, Read returns promptly with ErrContextCanceled
	// and no partial points are surfaced to the caller.
	type readResult struct {
		data []byte
		err  error
	}
	done := make(chan readResult, 1)
	go func() {
		data, err := c.internalMaster.SendRequestWithRetryAndGetResponseDeadline(apdu, outstationID, deadline)
		done <- readResult{data: data, err: err}
	}()

	var respData []byte
	select {
	case <-ctx.Done():
		// Abandon the outstanding request; the goroutine's result is discarded.
		return nil, fmt.Errorf("%w: %v", dnp3.ErrContextCanceled, ctx.Err())
	case res := <-done:
		if res.err != nil {
			// If the transport disconnected, reflect that in the public state so
			// a subsequent Read returns ErrNotConnected rather than retrying a
			// dead link (DNP3-031). MEXT-025: also tear down the transport so a
			// later Connect re-dials a fresh socket instead of retrying the dead
			// one (no stuck state after a peer drop).
			if master.IsDisconnectError(res.err) {
				c.markTransportDisconnected()
			}
			// Attach the matching public error sentinel so callers can classify
			// the failure via dnp3.ClassifyError (DNP3-043). The internal error
			// chain is preserved.
			return nil, wrapInternalError("read failed", res.err)
		}
		respData = res.data
	}

	// Decode response (respData is already transport/data-link decoded)
	resp, err := al.DecodeResponse(respData)
	if err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	// Parse data into public types. DNP3-035: only the MVP-supported Class-0
	// static groups are surfaced (Binary Input G1, Analog Input G30, Counter
	// G20). Output-status (G10/G40) and frozen-counter parsing is intentionally
	// omitted from the v0 Read path; those response fields stay nil.
	binaryInputs := parseBinaryInputs(resp.Data)
	analogInputs := parseAnalogInputs(resp.Data)
	counters := parseCounters(resp.Data)

	// DNP3-053: if this response carried the DeviceRestart IIN bit and the
	// caller opted into auto-integrity-on-restart, automatically perform a
	// Class-0 integrity poll now (after the outstanding marker for this
	// request has been released by the internal master) so the outstation's
	// post-restart static data is re-read. The auto-poll's result is not
	// merged into this Read's response; it refreshes the master's stored
	// state (LastIIN / outstation IIN). Suppressed while an integrity poll is
	// already in flight to avoid recursion.
	c.maybeAutoIntegrityOnRestart(ctx)

	return &ReadResponse{
		IIN:          resp.IIN.Bytes(),
		Timestamp:    time.Now(),
		BinaryInputs: binaryInputs,
		AnalogInputs: analogInputs,
		Counters:     counters,
	}, nil
}

// IntegrityPoll implements Client.IntegrityPoll. It issues a Class-0 integrity
// poll — reading all MVP-supported static groups (Binary Input G1, Counter G20,
// Analog Input G30) — and returns a single ReadResponse carrying the combined
// points (DNP3-037).
//
// Since MEXT-015 the primary path is a single Class-0 multi-group Read: one
// request carrying G1/G20/G30 all-objects headers, answered by one
// multi-object-header response. The MEXT-014 qualifier-aware parsers populate
// every group from that single exchange without point loss. A per-group
// fallback is retained for peers that cannot satisfy a single multi-group
// Class-0 read (error on the primary exchange); see runIntegrityPoll.
func (c *client) IntegrityPoll(ctx context.Context) (*ReadResponse, error) {
	resp, err := c.runIntegrityPoll(ctx)
	if err != nil {
		return nil, fmt.Errorf("integrity poll: %w", err)
	}
	return resp, nil
}

// runIntegrityPoll is the shared Class-0 integrity implementation used by both
// IntegrityPoll and the DNP3-053 auto-integrity-on-restart path. It reads the
// MVP static groups (Binary Input G1, Counter G20, Analog Input G30) and
// returns a single ReadResponse carrying their combined points (DNP3-037).
//
// Primary path (MEXT-015): one Class-0 multi-group Read carrying G1/G20/G30
// all-objects headers. A conformant outstation answers with a single
// multi-object-header response; the MEXT-014 qualifier-aware parsers populate
// every group from that one exchange. This removes the prior per-group
// workaround as the primary path while keeping a documented fallback.
//
// Fallback: if the primary exchange errors (e.g. a peer that rejects a
// multi-group Class-0 read or returns a transport-level failure), the groups
// are re-read one at a time so a conformant-but-single-group outstation still
// yields the full set. The fallback's per-group error, if any, is surfaced.
func (c *client) runIntegrityPoll(ctx context.Context) (*ReadResponse, error) {
	// DNP3-053: mark an integrity poll in flight so the Reads it issues do not
	// themselves trigger a nested auto-integrity-on-restart poll (which would
	// recurse). This guard covers both user-initiated IntegrityPoll and the
	// auto-poll triggered by the DeviceRestart IIN bit.
	c.mu.Lock()
	c.autoIntegrityActive = true
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		c.autoIntegrityActive = false
		c.mu.Unlock()
	}()

	// VEC-01 B4 F2: the whole integrity poll — the primary multi-group read
	// AND any per-group fallback — shares ONE finite wall-clock budget. The
	// budget is derived once here so the fallback cannot silently reset it.
	deadline := time.Now().Add(c.internalMaster.ReadBudget())
	// Primary: single Class-0 multi-group read (MEXT-015).
	resp, err := c.readWithDeadline(ctx, types.NewReadRequest(
		types.GroupRequest{Group: 1, Variation: 0},
		types.GroupRequest{Group: 20, Variation: 0},
		types.GroupRequest{Group: 30, Variation: 0},
	), deadline)
	if err == nil {
		return resp, nil
	}
	// Fallback: per-group reads for peers that cannot satisfy a single
	// multi-group Class-0 exchange (MEXT-015).
	return c.integrityPollPerGroup(ctx, deadline)
}

// integrityPollPerGroup is the MEXT-015 fallback: each MVP static group is read
// in its own request and merged into one ReadResponse. It is used only when the
// primary single multi-group Class-0 read errors.
func (c *client) integrityPollPerGroup(ctx context.Context, deadline time.Time) (*ReadResponse, error) {
	type groupRead struct {
		group uint8
		apply func(*ReadResponse, *ReadResponse)
	}
	reads := []groupRead{
		{1, func(dst, src *ReadResponse) { dst.BinaryInputs = src.BinaryInputs }},
		{20, func(dst, src *ReadResponse) { dst.Counters = src.Counters }},
		{30, func(dst, src *ReadResponse) { dst.AnalogInputs = src.AnalogInputs }},
	}

	out := &ReadResponse{Timestamp: time.Now()}
	var lastIIN [2]byte
	for _, r := range reads {
		resp, err := c.readWithDeadline(ctx, types.NewReadRequest(types.GroupRequest{Group: r.group, Variation: 1}), deadline)
		if err != nil {
			return nil, fmt.Errorf("read group %d: %w", r.group, err)
		}
		r.apply(out, resp)
		lastIIN = resp.IIN
	}
	out.IIN = lastIIN
	return out, nil
}

// maybeAutoIntegrityOnRestart implements the DNP3-053 auto-integrity policy.
// When AutoIntegrityOnRestart is enabled and the outstation's response carried
// the DeviceRestart IIN bit (recorded by the internal master as NeedsIntegrity),
// it automatically issues a Class-0 integrity poll. It is a no-op when the
// policy is disabled, when no integrity poll is pending, or when an integrity
// poll is already in flight (preventing recursion). The auto-poll runs on the
// caller's context; its result is intentionally discarded — the side effect is
// refreshing the master's stored outstation state (LastIIN / IIN).
func (c *client) maybeAutoIntegrityOnRestart(ctx context.Context) {
	if !c.config.AutoIntegrityOnRestart {
		return
	}
	c.mu.RLock()
	active := c.autoIntegrityActive
	c.mu.RUnlock()
	if active {
		return
	}

	o, ok := c.internalMaster.GetOutstation(c.config.OutstationAddress)
	if !ok || !o.NeedsIntegrity() {
		return
	}
	// Clear the pending flag before the poll so a DeviceRestart bit carried by
	// the auto-poll's own responses (suppressed by autoIntegrityActive) does not
	// leave a stale requirement, and so a failed auto-poll is not retried
	// unboundedly by subsequent Reads.
	o.ClearNeedsIntegrity()
	c.emitLog(LogInfo, "auto-integrity", "DeviceRestart IIN → auto integrity poll", nil)
	// Best-effort: the auto-poll is a background correctness action; its error
	// is logged but never surfaced to the triggering Read caller.
	if _, err := c.runIntegrityPoll(ctx); err != nil {
		c.emitLog(LogError, "auto-integrity", "auto integrity poll failed", err)
	}
}

// LastIIN implements Client.LastIIN. It returns the master's stored copy of
// the most recent Internal Indications for the configured outstation
// (DNP3-012). The value is updated on every processed response.
func (c *client) LastIIN() [2]byte {
	outstation, ok := c.internalMaster.GetOutstation(c.config.OutstationAddress)
	if !ok {
		return [2]byte{}
	}
	return outstation.GetIIN()
}

// isSupportedReadGroup reports whether a (group, variation) pair is in the v0
// read profile (DNP3-029). The supported groups are Binary Input (G1), Counter
// (G20), and Analog Input (G30). Variation 0 means "any" and is accepted for
// these groups; the concrete v0 variation (1) is also accepted.
func isSupportedReadGroup(group, variation uint8) bool {
	switch group {
	case 1, 20, 30:
		return variation == 0 || variation == 1
	default:
		return false
	}
}

// buildReadRequest builds an APDU for a read request
func buildReadRequest(seq uint8, request *types.ReadRequest) *al.APDU {
	// Build object headers from request groups using the formal ObjectHeader
	// model. Each group is encoded with the 0x06 (all-objects) qualifier.
	var data []byte
	for _, group := range request.Groups {
		h := al.ObjectHeader{
			Group:     group.Group,
			Variation: group.Variation,
			Qualifier: al.QualAllObjects,
		}
		data, _ = h.Encode(data)
	}

	return &al.APDU{
		Control: al.AppControl{
			FIR: true,
			FIN: true,
			CON: false,
			UNS: false,
			Seq: seq,
		},
		FuncCode: al.FuncRead, // 0x02
		Data:     data,
	}
}

// packedBinaryRange returns the index base and point count for a packed
// Group 1 Variation 1 header. ok is false for qualifiers that do not carry a
// usable count/range (e.g. QualIndex8, which is not sequential). The packed
// format carries no per-point bytes, so the count comes from the qualifier's
// count field or, for start/stop and range16, from Start..Stop.
func packedBinaryRange(h al.ObjectHeader) (base int, n int, ok bool) {
	return sequentialRange(h)
}

// sequentialRange returns the index base and point count for a sequential
// (no per-point index prefix) object header. count8/count16 are sequential
// from index 0; start/stop and range16 are sequential from Start through Stop;
// all-objects yields zero points (a static-data response would normally use a
// count qualifier instead). ok is false for QualIndex8 (0x17), which carries a
// per-point index and is therefore not sequential.
func sequentialRange(h al.ObjectHeader) (base int, n int, ok bool) {
	switch h.Qualifier {
	case al.QualCount8, al.QualCount16:
		return 0, int(h.Count), true
	case al.QualAllObjects:
		return 0, 0, true
	case al.QualStartStop8, al.QualRange16:
		if h.Stop < h.Start {
			return 0, 0, false
		}
		return int(h.Start), int(h.Stop-h.Start) + 1, true
	default:
		return 0, 0, false
	}
}

// parseBinaryInputs parses binary input data from response (Group 1 and Group 2)
func parseBinaryInputs(data []byte) []*types.BinaryInput {
	var result []*types.BinaryInput
	offset := 0

	for offset < len(data) {
		h, consumed, err := al.DecodeObjectHeader(data, offset)
		if err != nil {
			break
		}
		offset += consumed
		group := h.Group
		variation := h.Variation

		// Group 1 = Binary Input (static), Group 2 = Binary Input Event (with timestamp)
		if group != 1 && group != 2 {
			offset = skipGroupData(offset, data, h)
			continue
		}

		// IEEE 1815: Group 1 Variation 1 is "Binary Input - Packed Format".
		// Points are packed 8 per byte (bit 0 of byte 0 = point 0), and no
		// per-point quality byte is carried; parsed points get ONLINE quality.
		// The qualifier only fixes the index base: count8/count16/all-objects
		// pack sequentially from index 0; range16 packs from Start..Stop.
		// Qualifier 0x00 (index8) is not valid for the packed format and falls
		// through to the per-point path below.
		if group == 1 && variation == 1 {
			if base, n, ok := packedBinaryRange(h); ok {
				packedBytes := (n + 7) / 8
				if offset+packedBytes > len(data) {
					break
				}
				for i := 0; i < n; i++ {
					result = append(result, &types.BinaryInput{
						Index:   uint16(base + i),
						Value:   data[offset+i/8]&(1<<uint(i%8)) != 0,
						Quality: types.QualityOnline,
					})
				}
				offset += packedBytes
				continue
			}
		}

		n, base, idxPrefix, ok := pointLayout(h)
		if !ok {
			break
		}
		for i := 0; i < n; i++ {
			var index uint16
			if idxPrefix > 0 {
				if offset+idxPrefix > len(data) {
					break
				}
				index = uint16(data[offset])
				offset += idxPrefix
			} else {
				index = uint16(base + i)
			}

			var value bool
			var quality types.QualityFlags
			var timestamp *types.Timestamp

			switch variation {
			case 1: // With flags
				if offset >= len(data) {
					break
				}
				val := data[offset]
				value = (val & 0x80) != 0
				quality = types.QualityFlags(val & 0x7F)
				offset++
			case 2: // Without flags
				if offset >= len(data) {
					break
				}
				value = data[offset] != 0
				quality = types.QualityOnline
				offset++
			default:
				if offset < len(data) {
					offset++
				}
			}

			// Group 2 (Binary Input Event) has timestamp after flags
			if group == 2 && offset+8 <= len(data) {
				var timeBytes [8]byte
				copy(timeBytes[:], data[offset:offset+8])
				offset += 8
				timestamp = types.NewTimestampFromDNP3(timeBytes)
			}

			result = append(result, &types.BinaryInput{
				Index:   index,
				Value:   value,
				Quality: quality,
				Time:    timestamp,
			})
		}
	}

	return result
}

// parseAnalogInputs parses analog input data from response (Group 30 and Group 31)
func parseAnalogInputs(data []byte) []*types.AnalogInput {
	var result []*types.AnalogInput
	offset := 0

	for offset < len(data) {
		h, consumed, err := al.DecodeObjectHeader(data, offset)
		if err != nil {
			break
		}
		offset += consumed
		group := h.Group
		variation := h.Variation

		// Group 30 = Analog Input (static), Group 31 = Analog Input Event (with timestamp)
		if group != 30 && group != 31 {
			offset = skipGroupData(offset, data, h)
			continue
		}

		// IEEE 1815: Group 30 Variation 1 is a signed 32-bit value with a
		// 1-octet flags byte (5 octets per point). The flags octet is first,
		// then the little-endian value, matching the pinned peer's own
		// generated serializer (opendnp3 3.1.2 Group30Var1 = (flags, value32)).
		// pointLayout fixes the index base and, for the indexed qualifier
		// (0x17), a per-point 1-octet index prefix. A declared count that does
		// not fit the remaining bytes is rejected without partial points.
		if group == 30 && variation == 1 {
			if n, base, idxPrefix, ok := pointLayout(h); ok {
				size := 5 + idxPrefix
				if offset+n*size > len(data) {
					break
				}
				for i := 0; i < n; i++ {
					var index uint16
					if idxPrefix > 0 {
						index = uint16(data[offset])
						offset++
					} else {
						index = uint16(base + i)
					}
					quality := types.QualityFlags(data[offset])
					value := int32(binary.LittleEndian.Uint32(data[offset+1 : offset+5]))
					offset += 5
					result = append(result, &types.AnalogInput{Index: index, Value: float64(value), Quality: quality})
				}
				continue
			}
		}

		// Need at least 7 bytes per point for variation 1 (32-bit float with flags)
		n, base, idxPrefix, ok := pointLayout(h)
		if !ok {
			break
		}
		for i := 0; i < n; i++ {
			var index uint16
			if idxPrefix > 0 {
				if offset+idxPrefix > len(data) {
					break
				}
				index = uint16(data[offset])
				offset += idxPrefix
			} else {
				index = uint16(base + i)
			}

			var value float64
			var quality types.QualityFlags
			var timestamp *types.Timestamp

			switch variation {
			case 1: // 32-bit float with flags
				if offset+5 > len(data) {
					break
				}
				bits := binary.LittleEndian.Uint32(data[offset : offset+4])
				value = float64(math.Float32frombits(bits))
				offset += 4
				quality = types.QualityFlags(data[offset])
				offset++
			case 2: // 16-bit integer with flags
				if offset+3 > len(data) {
					break
				}
				val := int16(binary.LittleEndian.Uint16(data[offset : offset+2]))
				value = float64(val)
				offset += 2
				quality = types.QualityFlags(data[offset])
				offset++
			case 3: // 32-bit integer with flags
				if offset+5 > len(data) {
					break
				}
				val := int32(binary.LittleEndian.Uint32(data[offset : offset+4]))
				value = float64(val)
				offset += 4
				quality = types.QualityFlags(data[offset])
				offset++
			case 5: // 32-bit float without flags
				if offset+4 > len(data) {
					break
				}
				bits := binary.LittleEndian.Uint32(data[offset : offset+4])
				value = float64(math.Float32frombits(bits))
				offset += 4
				quality = types.QualityOnline
			default:
				if offset < len(data) {
					offset++
				}
			}

			// Group 31 (Analog Input Event) has 8-byte timestamp after flags
			if group == 31 && offset+8 <= len(data) {
				var timeBytes [8]byte
				copy(timeBytes[:], data[offset:offset+8])
				offset += 8
				timestamp = types.NewTimestampFromDNP3(timeBytes)
			}

			result = append(result, &types.AnalogInput{
				Index:   index,
				Value:   value,
				Quality: quality,
				Time:    timestamp,
			})
		}
	}

	return result
}

// parseCounters parses counter data from response (Group 20 and Group 21)
func parseCounters(data []byte) []*types.Counter {
	var result []*types.Counter
	offset := 0

	for offset < len(data) {
		h, consumed, err := al.DecodeObjectHeader(data, offset)
		if err != nil {
			break
		}
		offset += consumed
		group := h.Group
		variation := h.Variation

		// Group 20 = Counter (static), Group 21 = Counter Event (with timestamp)
		if group != 20 && group != 21 {
			offset = skipGroupData(offset, data, h)
			continue
		}

		// IEEE 1815: Group 20 Variation 1 is an unsigned 32-bit counter with a
		// 1-octet flags byte (5 octets per point). The flags octet is first,
		// then the little-endian value, matching the pinned peer's own
		// generated serializer (opendnp3 3.1.2 Group20Var1 = (flags, count32)).
		// pointLayout fixes the index base and, for the indexed qualifier
		// (0x17), a per-point 1-octet index prefix. A declared count that does
		// not fit the remaining bytes is rejected without partial points.
		if group == 20 && variation == 1 {
			if n, base, idxPrefix, ok := pointLayout(h); ok {
				size := 5 + idxPrefix
				if offset+n*size > len(data) {
					break
				}
				for i := 0; i < n; i++ {
					var index uint16
					if idxPrefix > 0 {
						index = uint16(data[offset])
						offset++
					} else {
						index = uint16(base + i)
					}
					quality := types.QualityFlags(data[offset])
					value := binary.LittleEndian.Uint32(data[offset+1 : offset+5])
					offset += 5
					result = append(result, &types.Counter{Index: index, Value: value, Quality: quality})
				}
				continue
			}
		}

		// Need at least 7 bytes per point for variation 1 (32-bit counter with flags)
		n, base, idxPrefix, ok := pointLayout(h)
		if !ok {
			break
		}
		for i := 0; i < n; i++ {
			var index uint16
			if idxPrefix > 0 {
				if offset+idxPrefix > len(data) {
					break
				}
				index = uint16(data[offset])
				offset += idxPrefix
			} else {
				index = uint16(base + i)
			}

			var value uint32
			var quality types.QualityFlags
			var timestamp *types.Timestamp

			switch variation {
			case 1: // 32-bit with flags
				if offset+5 > len(data) {
					break
				}
				value = binary.LittleEndian.Uint32(data[offset : offset+4])
				offset += 4
				quality = types.QualityFlags(data[offset])
				offset++
			case 5: // 16-bit with flags
				if offset+3 > len(data) {
					break
				}
				value = uint32(binary.LittleEndian.Uint16(data[offset : offset+2]))
				offset += 2
				quality = types.QualityFlags(data[offset])
				offset++
			case 6: // 32-bit without flags
				if offset+4 > len(data) {
					break
				}
				value = binary.LittleEndian.Uint32(data[offset : offset+4])
				offset += 4
				quality = types.QualityOnline
			default:
				if offset < len(data) {
					offset++
				}
			}

			// Group 21 (Counter Event) has 8-byte timestamp after flags
			if group == 21 && offset+8 <= len(data) {
				var timeBytes [8]byte
				copy(timeBytes[:], data[offset:offset+8])
				offset += 8
				timestamp = types.NewTimestampFromDNP3(timeBytes)
			}

			result = append(result, &types.Counter{
				Index:   index,
				Value:   value,
				Quality: quality,
				Time:    timestamp,
			})
		}
	}

	return result
}

// parseBinaryOutputs parses binary output status data from response (Group 10)
func parseBinaryOutputs(data []byte) []*types.BinaryOutput {
	var result []*types.BinaryOutput
	offset := 0

	for offset < len(data) {
		h, consumed, err := al.DecodeObjectHeader(data, offset)
		if err != nil {
			break
		}
		offset += consumed
		group := h.Group
		variation := h.Variation

		if group != 10 { // Group 10 = Binary Output
			// Skip past this group's data
			offset = skipGroupData(offset, data, h)
			continue
		}

		n, base, idxPrefix, ok := pointLayout(h)
		if !ok {
			break
		}
		for i := 0; i < n; i++ {
			var index uint16
			if idxPrefix > 0 {
				if offset+idxPrefix > len(data) {
					break
				}
				index = uint16(data[offset])
				offset += idxPrefix
			} else {
				index = uint16(base + i)
			}

			if offset+1 > len(data) {
				break
			}

			var value bool
			var quality types.QualityFlags

			switch variation {
			case 1: // With flags
				val := data[offset]
				value = (val & 0x80) != 0
				quality = types.QualityFlags(val & 0x7F)
				offset++
			case 2: // Without flags
				value = data[offset] != 0
				quality = types.QualityOnline
				offset++
			default:
				offset++
			}

			result = append(result, &types.BinaryOutput{
				Index:   index,
				Value:   value,
				Quality: quality,
			})
		}
	}

	return result
}

// parseAnalogOutputs parses analog output status data from response (Group 40)
func parseAnalogOutputs(data []byte) []*types.AnalogOutput {
	var result []*types.AnalogOutput
	offset := 0

	for offset < len(data) {
		h, consumed, err := al.DecodeObjectHeader(data, offset)
		if err != nil {
			break
		}
		offset += consumed
		group := h.Group
		variation := h.Variation

		if group != 40 { // Group 40 = Analog Output
			// Skip past this group's data
			offset = skipGroupData(offset, data, h)
			continue
		}

		n, base, idxPrefix, ok := pointLayout(h)
		if !ok {
			break
		}
		for i := 0; i < n; i++ {
			var index uint16
			if idxPrefix > 0 {
				if offset+idxPrefix > len(data) {
					break
				}
				index = uint16(data[offset])
				offset += idxPrefix
			} else {
				index = uint16(base + i)
			}

			if offset+5 > len(data) {
				break
			}

			var value float64
			var quality types.QualityFlags

			switch variation {
			case 1: // 32-bit float with flags
				bits := binary.LittleEndian.Uint32(data[offset : offset+4])
				value = float64(math.Float32frombits(bits))
				offset += 4
				quality = types.QualityFlags(data[offset])
				offset++
			case 2: // 16-bit int with flags
				val := int16(binary.LittleEndian.Uint16(data[offset : offset+2]))
				value = float64(val)
				offset += 2
				quality = types.QualityFlags(data[offset])
				offset++
			default:
				offset += 5
			}

			result = append(result, &types.AnalogOutput{
				Index:   index,
				Value:   value,
				Quality: quality,
			})
		}
	}

	return result
}

// skipGroupData advances the offset past the data for a group.
// Returns the new offset position.
func skipGroupData(offset int, data []byte, h al.ObjectHeader) int {
	n := objectPointCount(h)
	if n == 0 {
		return offset
	}
	// G1V1 packed format: bits packed 8 per byte, no index/flags.
	if h.Group == 1 && h.Variation == 1 && isSequentialQualifier(h.Qualifier) {
		packedBytes := (n + 7) / 8
		newOffset := offset + packedBytes
		if newOffset > len(data) {
			return len(data)
		}
		return newOffset
	}
	valueBytes := objectValueBytes(h.Group, h.Variation)
	indexPrefix := objectIndexPrefixBytes(h.Qualifier)
	newOffset := offset + n*(indexPrefix+valueBytes)
	if newOffset > len(data) {
		return len(data)
	}
	return newOffset
}

// objectPointCount returns the number of points described by a header,
// independent of qualifier-specific encoding.
func objectPointCount(h al.ObjectHeader) int {
	switch h.Qualifier {
	case al.QualCount8, al.QualIndex8, al.QualCount16:
		return int(h.Count)
	case al.QualStartStop8, al.QualRange16:
		if h.Stop < h.Start {
			return 0
		}
		return int(h.Stop-h.Start) + 1
	case al.QualAllObjects:
		return 0
	default:
		return int(h.Count)
	}
}

// isSequentialQualifier reports whether the qualifier is a sequential
// (no per-point index) layout: count8, count16, range16, all-objects.
func isSequentialQualifier(q uint8) bool {
	switch q {
	case al.QualCount8, al.QualCount16, al.QualStartStop8, al.QualRange16, al.QualAllObjects:
		return true
	default:
		return false
	}
}

// objectIndexPrefixBytes returns the per-point index prefix size for a
// qualifier. Only index8 (0x17, UINT8_CNT_UINT8_INDEX) carries a per-point
// index in this stack (an 8-bit index); sequential qualifiers carry none.
func objectIndexPrefixBytes(q uint8) int {
	if q == al.QualIndex8 {
		return 1
	}
	return 0
}

// pointLayout returns the point count, index base and per-point index prefix
// size for an object header, and whether the qualifier is supported. Sequential
// qualifiers (count8/count16/start-stop/range16/all-objects) carry no per-point
// index; index8 (0x17) prefixes each point with its 8-bit index.
func pointLayout(h al.ObjectHeader) (n int, base int, indexPrefix int, ok bool) {
	switch h.Qualifier {
	case al.QualCount8, al.QualCount16:
		return int(h.Count), 0, 0, true
	case al.QualAllObjects:
		return 0, 0, 0, true
	case al.QualStartStop8, al.QualRange16:
		if h.Stop < h.Start {
			return 0, 0, 0, false
		}
		return int(h.Stop-h.Start) + 1, int(h.Start), 0, true
	case al.QualIndex8:
		return int(h.Count), 0, 1, true
	default:
		return 0, 0, 0, false
	}
}

// objectValueBytes returns the per-point value size (excluding any index
// prefix and any event timestamp) for the MVP-supported object variations.
func objectValueBytes(group, variation uint8) int {
	switch group {
	case 1: // Binary Input
		return 1
	case 20: // Counter
		switch variation {
		case 5: // uint16 + flags
			return 3
		case 6: // uint32 only
			return 4
		default: // uint32 + flags
			return 5
		}
	case 30: // Analog Input
		switch variation {
		case 2: // int16 + flags
			return 3
		case 5: // float32 only
			return 4
		default: // int32/float32 + flags
			return 5
		}
	case 10: // Binary Output
		return 1
	case 40: // Analog Output
		switch variation {
		case 2: // int16 + flags
			return 3
		default: // float32 + flags
			return 5
		}
	default:
		return 5
	}
}

// Operate implements Client.Operate
func (c *client) Operate(ctx context.Context, command *types.ControlOutput) (*OperateResponse, error) {
	if command == nil {
		return nil, dnp3.ErrInvalidRequest
	}

	// VEC-01 B4 F3: an already-cancelled context is a definite pre-send failure —
	// nothing was transmitted, so no control is dispatched and no quarantine
	// applies. Checked before spawning the dispatch worker so the outcome is
	// deterministic rather than racing the worker's own completion.
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", dnp3.ErrContextCanceled, err)
	}

	c.mu.RLock()
	state := c.state
	internal := c.internalMaster
	outstationID := c.config.OutstationAddress
	c.mu.RUnlock()

	// VEC-01 B4 F3: register the control worker (gate) BEFORE any other
	// admission so an UNKNOWN outcome or a concurrent control dominates every
	// other refusal — a quarantined outstation is reported as quarantined
	// regardless of connection state. Admission fails closed with
	// ErrControlQuarantined (prior outcome UNKNOWN) or ErrControlInProgress
	// (another worker active).
	gen, release, gerr := internal.AcquireControlGate(outstationID)
	if gerr != nil {
		return nil, wrapInternalError("operate refused", gerr)
	}

	if state < dnp3.StateConnected {
		release()
		return nil, dnp3.ErrNotConnected
	}

	// v0 profile gate (DNP3-030): only DirectOperate (with response) is
	// supported. Select-before-operate and direct-operate-no-response are
	// rejected with a clear error before any wire traffic.
	switch command.CommandType {
	case types.SelectThenOperate:
		release()
		return nil, fmt.Errorf("%w: select-before-operate", dnp3.ErrUnsupportedOption)
	case types.DirectOperateNoResponse:
		release()
		return nil, fmt.Errorf("%w: direct-operate-no-response", dnp3.ErrUnsupportedOption)
	case types.DirectOperate:
		// supported
	default:
		release()
		return nil, fmt.Errorf("%w: unknown command type %d", dnp3.ErrUnsupportedOption, command.CommandType)
	}

	// Fail-closed object gate (AO-002/AO-003): admit only the verified control
	// objects — Group 12 Variation 1 (CROB), Group 41 Variation 1 (int32 analog
	// output), Group 41 Variation 2 (int16 analog output), Group 41 Variation 3
	// (float32 analog output) and Group 41 Variation 4 (float64 analog output),
	// DirectOperate-with-response. Any other group/variation is rejected before
	// any wire traffic rather than encoded and sent.
	switch {
	case command.Group == 12 && command.Variation == 1:
		// CROB (existing behavior)
	case command.Group == 41 && (command.Variation == 1 || command.Variation == 2 || command.Variation == 3 || command.Variation == 4):
		// G41V1 int32 / G41V2 int16 / G41V3 float32 / G41V4 float64 analog output
	default:
		release()
		return nil, fmt.Errorf("%w: unsupported control object group %d variation %d", dnp3.ErrUnsupportedOption, command.Group, command.Variation)
	}

	// Translate command
	selectThenOperate := false // DirectOperate only in v0

	// Extract raw value from CommandValue interface
	var rawValue interface{}
	switch v := command.Value.(type) {
	case *types.BinaryCommandValue:
		rawValue = v.Value // bool
	case *types.AnalogInt32CommandValue:
		rawValue = v.Value // int32 (G41V1)
	case *types.AnalogInt16CommandValue:
		rawValue = v.Value // int16 (G41V2)
	case *types.AnalogFloat32CommandValue:
		rawValue = v.Value // float32 (G41V3)
	case *types.AnalogFloat64CommandValue:
		rawValue = v.Value // float64 (G41V4)
	case *types.AnalogCommandValue:
		rawValue = v.Value // float64
	default:
		rawValue = command.Value
	}

	// VEC-01 B4 F1/F3: a non-idempotent control is dispatched single-shot. The
	// worker (gate acquired above) runs on the caller's context with ONE
	// whole-control deadline, so cancellation and a session change are observed
	// at the transmission point. If the context ends while the worker is
	// outstanding, the outstation is quarantined (fail closed) because the
	// control may have been executed; the quarantine is persistent and is never
	// released automatically.
	deadline := time.Now().Add(internal.OperateBudget())

	type operateResult struct {
		status master.CommandStatus
		err    error
	}
	done := make(chan operateResult, 1)
	go func() {
		// The worker owns the gate for its whole lifetime: release runs only when
		// the worker terminates, so a still-blocked worker keeps the outstation
		// gated (no second control can be admitted behind it).
		defer release()
		status, err := internal.OperateWithStatusGated(ctx, gen, outstationID, selectThenOperate, command.Group, command.Variation, command.Index, rawValue, deadline)
		done <- operateResult{status: status, err: err}
	}()

	var cs master.CommandStatus
	select {
	case <-ctx.Done():
		// Fail closed: the control may already be on the wire, so quarantine the
		// outstation. This is never cleared by reconnect or a later control.
		if c.internalMaster != nil {
			c.internalMaster.QuarantineOutstation(outstationID)
		}
		return nil, fmt.Errorf("%w: %v", dnp3.ErrContextCanceled, ctx.Err())
	case res := <-done:
		if res.err != nil {
			// If the transport disconnected, reflect that in the public state so
			// a subsequent Operate returns ErrNotConnected rather than retrying a
			// dead link (DNP3-031). MEXT-025: also tear down the transport so a
			// later Connect re-dials a fresh socket instead of retrying the dead
			// one (no stuck state after a peer drop).
			if master.IsDisconnectError(res.err) {
				c.markTransportDisconnected()
			}
			if ctx.Err() != nil {
				// The caller's context ended; surface the cancellation sentinel
				// while preserving the internal ErrControlOutcomeUnknown chain.
				return nil, fmt.Errorf("%w: %w", dnp3.ErrContextCanceled, res.err)
			}
			// Attach the matching public error sentinel so callers can classify
			// the failure via dnp3.ClassifyError (DNP3-043). The internal error
			// chain is preserved.
			return nil, wrapInternalError("operate failed", res.err)
		}
		cs = res.status
	}
	// Get outstation state
	outstation, _ := internal.GetOutstation(outstationID)
	var iin [2]byte
	if outstation != nil {
		iin = outstation.IIN
	}

	return &OperateResponse{
		IIN:       iin,
		Status:    mapCommandStatus(cs),
		Timestamp: time.Now(),
	}, nil
}

// mapCommandStatus translates the internal master's CommandStatus to the public
// types.ControlStatus. Unknown/unparseable statuses surface as ControlTimeout
// (not ControlSuccess) so callers never mistake a missing status for success.
func mapCommandStatus(cs master.CommandStatus) types.ControlStatus {
	switch cs {
	case master.CommandStatusSuccess:
		return types.ControlSuccess
	case master.CommandStatusTimeout:
		return types.ControlTimeout
	case master.CommandStatusNoSelect:
		return types.ControlNoSelect
	case master.CommandStatusBadFormat:
		return types.ControlBadFormat
	case master.CommandStatusNotSupported:
		return types.ControlNotSupported
	case master.CommandStatusAlreadyActive:
		return types.ControlAlreadyActive
	case master.CommandStatusHardwareError:
		return types.ControlHardwareError
	case master.CommandStatusLocal:
		return types.ControlLocal
	case master.CommandStatusTooMany:
		return types.ControlTooMany
	case master.CommandStatusNotAuthorized:
		return types.ControlNotAuthorized
	case master.CommandStatusAutonomous:
		return types.ControlAutonomous
	case master.CommandStatusOutOfRange:
		return types.ControlOutOfRange
	default:
		// CommandStatusUnknown or any out-of-range value: never success.
		return types.ControlTimeout
	}
}

// EnableUnsolicited implements Client.EnableUnsolicited.
func (c *client) EnableUnsolicited(ctx context.Context) error {
	return c.sendClassFunction(ctx, al.FuncEnableUnsolicited)
}

// DisableUnsolicited implements Client.DisableUnsolicited.
func (c *client) DisableUnsolicited(ctx context.Context) error {
	return c.sendClassFunction(ctx, al.FuncDisableUnsolicited)
}

// SetUnsolicitedHandler implements Client.SetUnsolicitedHandler
func (c *client) SetUnsolicitedHandler(handler UnsolicitedHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers = append(c.handlers, handler)
}

// Close implements Client.Close
func (c *client) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.Disconnect(ctx)
}

// wrapInternalError attaches the matching public error sentinel to an error
// returned by the internal master so callers can classify it via dnp3.ClassifyError
// without importing internal packages (DNP3-043). The internal error is preserved
// in the chain (via %w) so existing errors.Is checks against internal sentinels
// (e.g. ErrTransportDisconnected) keep working. Disconnects are handled by the
// caller (state transition + ErrNotConnected); this helper only maps the
// non-disconnect protocol classes.
func wrapInternalError(prefix string, err error) error {
	if err == nil {
		return nil
	}
	switch {
	case master.IsDisconnectError(err):
		// Attach the public disconnect sentinel so callers can classify via
		// dnp3.ClassifyError; preserve the internal sentinel (e.g.
		// ErrTransportDisconnected) for existing errors.Is checks.
		return fmt.Errorf("%s: %w: %w", prefix, dnp3.ErrNotConnected, err)
	case errors.Is(err, master.ErrCRCError):
		return fmt.Errorf("%s: %w: %w", prefix, dnp3.ErrCRC, err)
	case errors.Is(err, master.ErrConfirmTimeout) || errors.Is(err, master.ErrTimeout):
		return fmt.Errorf("%s: %w: %w", prefix, dnp3.ErrTimeout, err)
	case errors.Is(err, master.ErrResponseSeqMismatch) || errors.Is(err, master.ErrConfirmSeqMismatch):
		return fmt.Errorf("%s: %w: %w", prefix, dnp3.ErrSequenceError, err)
	case errors.Is(err, master.ErrRequestOutstanding):
		return fmt.Errorf("%s: %w: %w", prefix, dnp3.ErrRequestOutstanding, err)
	default:
		return fmt.Errorf("%s: %w", prefix, err)
	}
}
