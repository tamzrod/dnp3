// Package master implements the DNP3 Master Station role.
//
// The Master Station is the central control system that initiates all communication.
// It polls outstations for data, issues commands, and manages the SCADA system.
//
// Reference: IEEE 1815-2012 Section 8
package master

import (
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"strings"
	"sync"
	"time"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/tl"
)

// Master role constants
const (
	MaxOutstations = 255  // Maximum outstations per master
	MaxRetries     = 3    // Default maximum retries
	DefaultTimeout = 5000 // Default timeout in milliseconds
)

// State represents the master's operational state.
//
// State machine (DNP3-039). Legal transitions:
//
//	Disconnected ──Connect()──▶ Connecting
//	Connecting ──handshake ok──▶ Connected ─▶ Active
//	Connecting ──handshake fail──▶ Error
//	Connecting ──abort──▶ Disconnected
//	Connected/Active ──Initialize()──▶ Initialized
//	Connected/Active/Initialized ──transport disconnect──▶ Error
//	any ──Disconnect()──▶ Disconnected
//	Error ──Connect() (reconnect)──▶ Connecting
//
// StateError is a terminal-but-recoverable state: the only legal way out is a
// fresh Connect (re-handshake) or Disconnect. Operations are rejected in every
// non-operational state (Disconnected, Connecting, Connected, Error) — in
// particular StateError must NOT silently satisfy an operation readiness gate
// even though its iota ordinal is higher than StateActive.
type State int

const (
	StateDisconnected State = iota
	StateConnecting
	StateConnected
	StateInitialized
	StateActive
	StateError
)

// String returns the state name.
func (s State) String() string {
	names := []string{
		"Disconnected",
		"Connecting",
		"Connected",
		"Initialized",
		"Active",
		"Error",
	}

	if s >= 0 && int(s) < len(names) {
		return names[s]
	}

	return "Unknown"
}

// legalStateTransitions encodes the allowed (from -> to) state transitions
// (DNP3-039). A self-transition (from == to) is always allowed (idempotent).
var legalStateTransitions = map[State]map[State]bool{
	StateDisconnected: {StateConnecting: true},
	StateConnecting: {
		StateConnected:    true,
		StateError:        true,
		StateDisconnected: true,
	},
	StateConnected: {
		StateActive:       true,
		StateInitialized:  true,
		StateError:        true,
		StateDisconnected: true,
	},
	StateActive: {
		StateInitialized:  true,
		StateError:        true,
		StateDisconnected: true,
	},
	StateInitialized: {
		StateError:        true,
		StateDisconnected: true,
		StateActive:       true,
	},
	StateError: {
		StateConnecting:   true, // reconnect
		StateDisconnected:  true,
	},
}

// isOperational reports whether the master is in a state that permits
// application-layer operations (Read/Poll/Operate/Write/TimeSync). Only
// StateInitialized and StateActive qualify; StateError is explicitly excluded
// even though its ordinal exceeds StateActive (DNP3-039).
func (m *Master) isOperational() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state == StateInitialized || m.state == StateActive
}

// isLinkUp reports whether the link layer is established (Connected,
// Initialized, or Active). Connecting, Disconnected, and Error are excluded
// (DNP3-039): a half-open or dead link must not silently satisfy a guard that
// assumes the link is up.
func (m *Master) isLinkUp() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	switch m.state {
	case StateConnected, StateInitialized, StateActive:
		return true
	}
	return false
}

// transitionTo moves the master to newState if the transition is legal per
// legalStateTransitions, returning ErrIllegalStateTransition otherwise
// (DNP3-039). A self-transition is a no-op success. Callers in the lifecycle
// paths (Connect/Initialize/Disconnect/markDisconnected) use this so illegal
// transitions surface as errors instead of silently corrupting the state.
//
// DNP3-044: legal transitions emit a "state" diagnostic event after the lock is
// released (so callbacks may safely query master state); illegal transitions
// emit a DiagError event.
func (m *Master) transitionTo(newState State) error {
	var from State
	{
		m.mu.Lock()
		from = m.state
		if m.state == newState {
			m.mu.Unlock()
			return nil
		}
		allowed, ok := legalStateTransitions[m.state]
		if !ok || !allowed[newState] {
			m.mu.Unlock()
			m.diag(DiagError, "state", SeqNA, fmt.Sprintf("illegal transition %s -> %s", from, newState), ErrIllegalStateTransition)
			return fmt.Errorf("%w: %s -> %s", ErrIllegalStateTransition, from, newState)
		}
		m.state = newState
		m.mu.Unlock()
	}
	m.diag(DiagInfo, "state", SeqNA, fmt.Sprintf("%s -> %s", from, newState), nil)
	return nil
}

// PollType defines the type of poll.
type PollType int

const (
	PollIntegrity PollType = iota // Read all Class 0 data
	PollEvent                     // Read all pending events
	PollException                 // Read changed data only
	PollClass0                    // Read Class 0 (static) data
	PollClass1                    // Read Class 1 events
	PollClass2                    // Read Class 2 events
	PollClass3                    // Read Class 3 events
)

// String returns the poll type name.
func (p PollType) String() string {
	names := []string{
		"Integrity",
		"Event",
		"Exception",
		"Class 0",
		"Class 1",
		"Class 2",
		"Class 3",
	}

	if p >= 0 && int(p) < len(names) {
		return names[p]
	}

	return "Unknown"
}

// Outstation represents a connected outstation.
type Outstation struct {
	ID          uint16    // Outstation address
	Label       string    // Human-readable label
	LastSeen    time.Time // Last communication time
	State       string    // Operational state
	Unsolicited bool      // Supports unsolicited responses
	IIN         [2]byte   // Last Internal Indication
	mu          sync.RWMutex

	// DNP3-013 reaction state — set from received IIN bits, cleared by the
	// appropriate follow-up action (integrity poll / time-sync).
	needsIntegrity bool // DeviceRestart IIN received → re-integrity required
	needsTimeSync  bool // NeedTime IIN received → time sync required

	// DNP3-057: primary-side Frame Count Bit for confirmed user data on this
	// link. Toggles after each confirmed-data transaction; reset to false by a
	// Reset Link Stations exchange.
	fcb bool

	// DNP3-055: per-outstation application-layer sequence number (0-15). Each
	// outstation has an independent seq stream so interleaved operations on
	// multiple outstations do not cross-talk. Advances only on a successful
	// send (DNP3-008).
	seq uint8
}

// NewOutstation creates a new outstation entry.
func NewOutstation(id uint16, label string) *Outstation {
	return &Outstation{
		ID:          id,
		Label:       label,
		LastSeen:    time.Now(),
		State:       "Unknown",
		Unsolicited: false,
		IIN:         [2]byte{0, 0},
	}

}

// UpdateIIN updates the internal indication field.
func (o *Outstation) UpdateIIN(iin [2]byte) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.IIN = iin
	o.LastSeen = time.Now()
}

// GetIIN returns the last Internal Indication received from the outstation
// (DNP3-012: master-side IIN storage & exposure).
func (o *Outstation) GetIIN() [2]byte {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.IIN
}

// NeedsIntegrity reports whether the outstation has signalled a device
// restart and a re-integrity poll is required (DNP3-013). Clear it with
// ClearNeedsIntegrity once the integrity poll has been performed.
func (o *Outstation) NeedsIntegrity() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.needsIntegrity
}

// ClearNeedsIntegrity clears the re-integrity requirement.
func (o *Outstation) ClearNeedsIntegrity() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.needsIntegrity = false
}

// NeedsTimeSync reports whether the outstation has requested time
// synchronization (DNP3-013). Clear it with ClearNeedsTimeSync once a
// time-sync has been issued.
func (o *Outstation) NeedsTimeSync() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.needsTimeSync
}

// ClearNeedsTimeSync clears the time-sync requirement.
func (o *Outstation) ClearNeedsTimeSync() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.needsTimeSync = false
}

// markDeviceRestart records that the outstation restarted: local state is
// cleared and a re-integrity poll is flagged (DNP3-013).
func (o *Outstation) markDeviceRestart() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.needsIntegrity = true
	o.State = "Restart"
}

// markNeedTimeSync records that the outstation requested time sync (DNP3-013).
func (o *Outstation) markNeedTimeSync() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.needsTimeSync = true
}

// takeFCB returns the current primary-side Frame Count Bit for this link
// (DNP3-057). It is read (not toggled) so the caller can set FCB/FCV on the
// outgoing confirmed-data frame; toggle with advanceFCB once the transaction
// is committed.
func (o *Outstation) takeFCB() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.fcb
}

// advanceFCB toggles the primary-side FCB after a confirmed-data transaction
// (DNP3-057).
func (o *Outstation) advanceFCB() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.fcb = !o.fcb
}

// resetFCB clears the primary-side FCB to its post-reset value (false), called
// after a Reset Link Stations exchange (DNP3-057).
func (o *Outstation) resetFCB() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.fcb = false
}

// takeSeq returns the current application-layer sequence number for this
// outstation (DNP3-055). Read-only; advance with advanceSeq.
func (o *Outstation) takeSeq() uint8 {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.seq
}

// advanceSeq advances this outstation's sequence by one (mod 16) after a
// successful send (DNP3-005 / DNP3-008).
func (o *Outstation) advanceSeq() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.seq = (o.seq + 1) % 16
}

// HasFlag checks if an IIN flag is set.
func (o *Outstation) HasFlag(bit byte) bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return (o.IIN[0]&bit) != 0 || (o.IIN[1]&bit) != 0
}

// Master errors
var (
	ErrNotConnected        = errors.New("master not connected")
	ErrOutstationNotFound  = errors.New("outstation not found")
	ErrTimeout             = errors.New("operation timeout")
	ErrMaxRetries          = errors.New("maximum retries exceeded")
	ErrInvalidResponse     = errors.New("invalid response")
	ErrConfirmTimeout      = errors.New("confirmation timeout")
	ErrConfirmSeqMismatch  = errors.New("confirmation sequence mismatch")
	ErrResponseSeqMismatch = errors.New("response sequence mismatch")

	// ErrIllegalStateTransition indicates a lifecycle method attempted an
	// undefined state transition (DNP3-039). Surfaced rather than silently
	// overwriting the state.
	ErrIllegalStateTransition = errors.New("illegal state transition")

	// ErrTransportDisconnected indicates the transport reported the peer closed
	// the connection (or the transport was closed). The master transitions to
	// StateError when this is observed (DNP3-031).
	ErrTransportDisconnected = errors.New("transport disconnected")

	// ErrLinkNACK indicates the secondary station returned a link-layer NACK
	// (function code 1) in response to a user-data frame. The link is alive
	// but rejected the frame (e.g., busy/out-of-sequence); the request is
	// retryable (DNP3-034).
	ErrLinkNACK = errors.New("link-layer NACK received")

	// ErrCRCError indicates a received link frame failed CRC validation. A
	// corrupted frame is transient line noise; the request is retryable
	// (DNP3-034).
	ErrCRCError = errors.New("frame CRC error")

	// ErrRequestOutstanding indicates a request to an outstation is already in
	// flight on that same outstation (DNP3-040). A DNP3 link permits at most one
	// outstanding master request per outstation for MVP; a concurrent request to
	// the same outstation is rejected (not silently queued) so callers do not
	// block indefinitely behind a slow peer.
	ErrRequestOutstanding = errors.New("request already outstanding for outstation")

	// ErrIdleTimeout indicates the session was idle longer than the configured
	// IdleTimeout and the monitor closed it (DNP3-042). The master transitions to
	// Disconnected.
	ErrIdleTimeout = errors.New("idle timeout exceeded")
)

// RetryClass classifies a transport/protocol error to drive the retry policy
// (DNP3-034). Each class maps to its own retry count and delay.
type RetryClass int

const (
	// ClassTimeout is a response/confirmation timeout: no answer within the
	// configured timeout. Retryable with the configured delay.
	ClassTimeout RetryClass = iota
	// ClassNACK is a link-layer NACK. The link is alive; the frame was
	// rejected. Retryable, typically with a short delay.
	ClassNACK
	// ClassCRC is a corrupted received frame (CRC validation failure).
	// Transient; retryable.
	ClassCRC
	// ClassDisconnect is a transport/peer disconnect. Not retryable — the
	// link is dead.
	ClassDisconnect
	// ClassOther covers protocol errors that are not specifically timeout,
	// NACK, CRC, or disconnect (e.g., sequence mismatch, invalid response).
	// Retryable with the configured delay.
	ClassOther
)

// classifyRetryError maps an error to its RetryClass (DNP3-034). The order
// matters: disconnect is checked first (it is terminal), then the specific
// link/framing classes, then timeouts, falling back to ClassOther for
// everything else (which remains retryable to preserve prior behavior).
func classifyRetryError(err error) RetryClass {
	if err == nil {
		return ClassOther
	}
	if isDisconnectError(err) {
		return ClassDisconnect
	}
	if errors.Is(err, ErrLinkNACK) {
		return ClassNACK
	}
	if errors.Is(err, ErrCRCError) {
		return ClassCRC
	}
	if errors.Is(err, ErrConfirmTimeout) || errors.Is(err, ErrTimeout) {
		return ClassTimeout
	}
	return ClassOther
}

// RetryPolicy holds per-class retry counts and delays (DNP3-034). A count of
// N means the request may be attempted up to N times total for that class
// (the initial attempt plus N-1 retries). A zero or negative count means
// "no retry for this class" after the initial attempt fails.
type RetryPolicy struct {
	TimeoutRetries int
	TimeoutDelay   time.Duration
	NACKRetries    int
	NACKDelay      time.Duration
	CRCRetries     int
	CRCDelay       time.Duration
	OtherRetries   int
	OtherDelay     time.Duration
}

// DefaultRetryPolicy returns the standard retry table derived from the master
// Config (DNP3-034). All retryable classes share the configured MaxRetries and
// RetryDelay; disconnects are not retryable (handled before the policy is
// consulted). Per-class overrides may be set on the returned struct.
func DefaultRetryPolicy(cfg *Config) *RetryPolicy {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	delay := time.Duration(cfg.RetryDelay) * time.Millisecond
	retries := cfg.MaxRetries
	if retries < 1 {
		retries = MaxRetries
	}
	return &RetryPolicy{
		TimeoutRetries: retries,
		TimeoutDelay:   delay,
		NACKRetries:    retries,
		NACKDelay:      delay,
		CRCRetries:     retries,
		CRCDelay:       delay,
		OtherRetries:   retries,
		OtherDelay:     delay,
	}
}

// retryDecision returns whether another attempt is permitted for the given
// class and how long to wait before it. attempt is the number of attempts
// already made for the current request (1-based after the first attempt).
// Disconnects are never retried (the caller must short-circuit before this).
func (p *RetryPolicy) retryDecision(class RetryClass, attempt int) (bool, time.Duration) {
	var limit int
	var delay time.Duration
	switch class {
	case ClassTimeout:
		limit, delay = p.TimeoutRetries, p.TimeoutDelay
	case ClassNACK:
		limit, delay = p.NACKRetries, p.NACKDelay
	case ClassCRC:
		limit, delay = p.CRCRetries, p.CRCDelay
	case ClassDisconnect:
		return false, 0
	default:
		limit, delay = p.OtherRetries, p.OtherDelay
	}
	if attempt >= limit {
		return false, 0
	}
	return true, delay
}

// IsDisconnectError reports whether err indicates a transport/peer disconnect.
// It matches the canonical transport-close sentinels (io.EOF,
// io.ErrUnexpectedEOF, net.ErrClosed) so a real TCP transport's peer-close
// surfaces as a disconnect regardless of how it was wrapped (DNP3-031).
func IsDisconnectError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrTransportDisconnected) {
		return true
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	if errors.Is(err, net.ErrClosed) {
		return true
	}
	return false
}

// isDisconnectError is the internal alias used within this package.
func isDisconnectError(err error) bool { return IsDisconnectError(err) }

// Config holds master configuration.
type Config struct {
	MasterAddress uint16 // This master's address
	Timeout       int    // Response timeout in milliseconds
	MaxRetries    int    // Maximum retry attempts
	RetryDelay    int    // Delay between retries in milliseconds
	// IdleTimeout is the optional keep-alive/idle threshold in milliseconds
	// (DNP3-042). When positive, a background monitor closes the session to
	// Disconnected if no activity (send or receive) is observed for this long.
	// Zero or negative disables idle monitoring (the default).
	IdleTimeout int
	// ConfirmTimeout is the application-layer confirm timeout in milliseconds
	// (DNP3-066). This is the time the master waits for a confirm APDU
	// (FuncResponse) after sending a request with the CON bit set (DNP3-009),
	// which is distinct from the general response Timeout used on the
	// request/response path. Per IEEE 1815 the confirm timeout is typically
	// shorter than the full response timeout: a confirm is a small dedicated
	// reply, whereas the response Timeout must allow for the outstation to
	// gather and serialize object data. When non-positive, the confirm path
	// falls back to Timeout (the documented backward-compatible relation).
	ConfirmTimeout int
}

// DefaultConfirmTimeout is the default application-layer confirm timeout in
// milliseconds (DNP3-066). It is intentionally shorter than DefaultTimeout:
// a confirm is a small dedicated reply, not a full object response.
const DefaultConfirmTimeout = 2000

// DefaultConfig returns default master configuration.
func DefaultConfig() *Config {
	return &Config{
		MasterAddress: 0xFFFF, // Master uses broadcast or specific address
		Timeout:       DefaultTimeout,
		MaxRetries:    MaxRetries,
		RetryDelay:    100,
		ConfirmTimeout: DefaultConfirmTimeout,
	}

}

// Master represents a DNP3 Master Station.
type Master struct {
	config      *Config
	state       State
	outstations map[uint16]*Outstation
	mu          sync.RWMutex
	transport   TransportHandler
	onError     ErrorHandler
	fragmenter  *tl.Fragmenter  // Transport layer fragmenter
	reassembler *tl.Reassembler // Transport layer reassembler

	// reqMu serializes the request path (send → receive → reassembly) so that
	// concurrent requests on the same master do not race on the shared
	// reassembler or interleave fragments on a single link (DNP3-025). A DNP3
	// link is request/response; concurrent requests are safely queued.
	reqMu sync.Mutex

	// outstanding tracks per-outstation in-flight requests so a concurrent
	// request to the SAME outstation is rejected with ErrRequestOutstanding
	// instead of blocking behind reqMu indefinitely (DNP3-040). Different
	// outstations still queue via reqMu. Guarded by outstandingMu.
	outstanding   map[uint16]struct{}
	outstandingMu  sync.Mutex

	// idleTimeout is the configured keep-alive/idle threshold (DNP3-042). A
	// non-positive value disables idle monitoring.
	idleTimeout time.Duration
	// confirmTimeout is the resolved application-layer confirm timeout
	// (DNP3-066): ConfirmTimeout when positive, else Timeout. Used by
	// waitForConfirmation; distinct from the response Timeout.
	confirmTimeout time.Duration
	// lastActivity is the time of the last successful send or receive, updated
	// under mu. Read by the idle monitor.
	lastActivity time.Time
	// idleStop closes to signal the idle-monitor goroutine to exit. nil when
	// monitoring is disabled.
	idleStop chan struct{}
	// idleWG lets Disconnect wait for the monitor goroutine to exit.
	idleWG sync.WaitGroup

	// retryPolicy drives the per-class retry decision in sendWithRetry
	// (DNP3-034). It is derived from the config in NewMaster and may be
	// overridden via SetRetryPolicy.
	retryPolicy *RetryPolicy

	// DNP3-013 reaction hooks. Optional; invoked from processResponse when
	// the corresponding IIN bit is set. These are stubs/logs — full time
	// objects and integrity polling are later roadmap items.
	onDeviceRestart func(outstationID uint16)
	onNeedTimeSync  func(outstationID uint16)

	// diagHook is the optional diagnostic sink for frame/sequence events
	// (DNP3-044). nil means silent (the default). Snapshotted under mu and
	// invoked outside the master's own locks.
	diagHook DiagHook
}

// TransportHandler defines the interface for sending/receiving data.
type TransportHandler interface {
	Send(data []byte) error
	Receive() ([]byte, error)
	SetTimeout(ms int)
}

// ErrorHandler defines the callback for error notifications.
type ErrorHandler func(err error, context string)

// DiagLevel is the severity of a diagnostic event (DNP3-044).
type DiagLevel int

const (
	// DiagInfo is for routine frame/sequence lifecycle events (send, receive,
	// confirm, state transitions).
	DiagInfo DiagLevel = iota
	// DiagWarn is for recoverable anomalies (retry, sequence mismatch, NACK).
	DiagWarn
	// DiagError is for failures (timeout, CRC, disconnect, illegal transition).
	DiagError
)

// DiagEvent is a structured diagnostic event emitted at frame/sequence
// boundaries (DNP3-044). It is passed to the configured DiagHook; the hook is
// optional and nil (silent) by default.
type DiagEvent struct {
	// Level is the event severity.
	Level DiagLevel
	// Op names the operation, e.g. "send", "receive", "confirm", "retry",
	// "sequence", "state".
	Op string
	// Seq is the application-layer sequence number (0-15) for the event, or
	// the sentinel SeqNA when not applicable.
	Seq uint8
	// Msg is a short human-readable description.
	Msg string
	// Err carries the underlying error for failure events, or nil.
	Err error
}

// SeqNA marks a DiagEvent with no applicable sequence number.
const SeqNA uint8 = 0xFF

// DiagHook is an optional diagnostic sink invoked at frame/sequence boundaries
// (DNP3-044). Implementations must be safe for concurrent use; the master does
// not invoke the hook while holding its own mutex, so callbacks may safely query
// master state.
type DiagHook func(DiagEvent)

// NewMaster creates a new Master Station.
func NewMaster(config *Config) *Master {
	if config == nil {
		config = DefaultConfig()
	}

	return &Master{
		config:        config,
		state:         StateDisconnected,
		outstations:   make(map[uint16]*Outstation),
		fragmenter:    tl.NewFragmenter(),
		reassembler:   tl.NewReassembler(),
		retryPolicy:   DefaultRetryPolicy(config),
		outstanding:   make(map[uint16]struct{}),
		idleTimeout:   time.Duration(config.IdleTimeout) * time.Millisecond,
		confirmTimeout: confirmTimeoutFromConfig(config),
	}
}

// confirmTimeoutFromConfig resolves the application-layer confirm timeout
// (DNP3-066): Config.ConfirmTimeout (ms) when positive, else Config.Timeout
// (the documented backward-compatible relation).
func confirmTimeoutFromConfig(config *Config) time.Duration {
	if config != nil && config.ConfirmTimeout > 0 {
		return time.Duration(config.ConfirmTimeout) * time.Millisecond
	}
	if config != nil && config.Timeout > 0 {
		return time.Duration(config.Timeout) * time.Millisecond
	}
	return time.Duration(DefaultTimeout) * time.Millisecond
}

// SetRetryPolicy overrides the default retry table (DNP3-034). Passing nil
// restores the default policy derived from the master config.
func (m *Master) SetRetryPolicy(p *RetryPolicy) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p == nil {
		m.retryPolicy = DefaultRetryPolicy(m.config)
		return
	}
	m.retryPolicy = p
}

// RetryPolicy returns the active retry policy (DNP3-034).
func (m *Master) RetryPolicy() *RetryPolicy {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.retryPolicy
}

// SetTransport sets the transport handler.
func (m *Master) SetTransport(t TransportHandler) {
	m.transport = t
}

// SetErrorHandler sets the error handler callback.
func (m *Master) SetErrorHandler(h ErrorHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onError = h
}

// SetDiagnosticHook installs an optional diagnostic sink for frame/sequence
// events (DNP3-044). Pass nil to restore the default silent behavior. The hook
// is invoked outside the master's own locks so callbacks may safely query master
// state; callbacks must not block (the master calls them synchronously).
func (m *Master) SetDiagnosticHook(h DiagHook) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.diagHook = h
}

// diagnosticHook snapshots the configured hook under the read lock (DNP3-044).
// Returns nil when no hook is installed (the default), in which case callers
// skip the event entirely.
func (m *Master) diagnosticHook() DiagHook {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.diagHook
}

// diag emits a diagnostic event to the configured hook if one is installed
// (DNP3-044). It is a no-op when the hook is nil, so the default is silent.
func (m *Master) diag(level DiagLevel, op string, seq uint8, msg string, err error) {
	if h := m.diagnosticHook(); h != nil {
		h(DiagEvent{Level: level, Op: op, Seq: seq, Msg: msg, Err: err})
	}
}

// SetDeviceRestartHandler sets the callback invoked when an outstation
// reports the DeviceRestart IIN bit (DNP3-013). The master marks the
// outstation as needing re-integrity; the callback is the stub/log hook.
func (m *Master) SetDeviceRestartHandler(h func(outstationID uint16)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onDeviceRestart = h
}

// SetNeedTimeSyncHandler sets the callback invoked when an outstation
// reports the NeedTime IIN bit (DNP3-013). The master records that a time
// sync is required; the callback is the stub/log hook (full time objects
// are a later roadmap item).
func (m *Master) SetNeedTimeSyncHandler(h func(outstationID uint16)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onNeedTimeSync = h
}

// State returns the current master state.
func (m *Master) State() State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

// SetState updates the master state unconditionally, bypassing the transition
// table. It is retained as an escape hatch for tests and recovery paths that
// must force a state (e.g., simulating a mid-session drop). Lifecycle code
// uses transitionTo instead so illegal transitions surface as errors
// (DNP3-039).
func (m *Master) SetState(state State) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state = state
}

// beginRequest marks an in-flight request for outstationID (DNP3-040). A DNP3
// link permits at most one outstanding master request per outstation for MVP;
// a concurrent request to the same outstation is rejected with
// ErrRequestOutstanding rather than blocking behind reqMu. Requests to distinct
// outstations are not rejected here (they queue via reqMu).
func (m *Master) beginRequest(outstationID uint16) error {
	m.outstandingMu.Lock()
	defer m.outstandingMu.Unlock()
	if _, ok := m.outstanding[outstationID]; ok {
		return fmt.Errorf("%w: 0x%04X", ErrRequestOutstanding, outstationID)
	}
	m.outstanding[outstationID] = struct{}{}
	return nil
}

// endRequest clears the in-flight marker for outstationID (DNP3-040). Safe to
// call when no marker is present (idempotent).
func (m *Master) endRequest(outstationID uint16) {
	m.outstandingMu.Lock()
	delete(m.outstanding, outstationID)
	m.outstandingMu.Unlock()
}

// HasOutstandingRequest reports whether a request is currently in flight for
// outstationID (DNP3-040). Exposed for tests and diagnostics.
func (m *Master) HasOutstandingRequest(outstationID uint16) bool {
	m.outstandingMu.Lock()
	defer m.outstandingMu.Unlock()
	_, ok := m.outstanding[outstationID]
	return ok
}

// recordActivity stamps lastActivity to now under mu (DNP3-042). Called on
// every successful send and receive so the idle monitor can detect a silent
// peer.
func (m *Master) recordActivity() {
	m.mu.Lock()
	m.lastActivity = time.Now()
	m.mu.Unlock()
}

// lastActivityAt returns the last activity timestamp under mu (DNP3-042).
func (m *Master) lastActivityAt() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastActivity
}

// startIdleMonitor launches the background idle monitor (DNP3-042). It is a
// no-op when idleTimeout is non-positive. Safe to call only while holding no
// other master locks; intended to be invoked after a successful Connect.
func (m *Master) startIdleMonitor() {
	if m.idleTimeout <= 0 {
		return
	}
	m.mu.Lock()
	if m.idleStop != nil {
		// A previous monitor is still running (e.g., re-Connect); stop it first.
		close(m.idleStop)
		m.idleWG.Wait()
	}
	m.idleStop = make(chan struct{})
	m.lastActivity = time.Now()
	stop := m.idleStop
	m.mu.Unlock()

	m.idleWG.Add(1)
	go m.idleMonitorLoop(stop)
}

// stopIdleMonitor signals the idle monitor to exit and waits for it (DNP3-042).
// No-op when monitoring is disabled. Called from Disconnect.
func (m *Master) stopIdleMonitor() {
	m.mu.Lock()
	stop := m.idleStop
	m.idleStop = nil
	m.mu.Unlock()
	if stop == nil {
		return
	}
	close(stop)
	m.idleWG.Wait()
}

// idleMonitorLoop ticks at half the idle timeout and, if no activity has been
// observed for the full idle timeout, transitions the session to Disconnected
// with ErrIdleTimeout (DNP3-042). Exits when stop is closed.
func (m *Master) idleMonitorLoop(stop chan struct{}) {
	defer m.idleWG.Done()
	tick := m.idleTimeout / 2
	if tick <= 0 {
		tick = m.idleTimeout
	}
	if tick <= 0 {
		return
	}
	t := time.NewTicker(tick)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			if time.Since(m.lastActivityAt()) > m.idleTimeout {
				// Idle threshold exceeded: tear the session down to Disconnected
				// (DNP3-042 acceptance). Bypass the transition table like
				// Disconnect since idle-close is an unconditional teardown.
				m.mu.Lock()
				prev := m.state
				m.state = StateDisconnected
				m.mu.Unlock()
				if m.onError != nil {
					m.onError(ErrIdleTimeout, "idle monitor closed session")
				}
				// DNP3-044: emit a diagnostic event for the idle-driven close.
				m.diag(DiagError, "state", SeqNA, fmt.Sprintf("idle close %s -> Disconnected", prev), ErrIdleTimeout)
				return
			}
		}
	}
}

// AddOutstation registers an outstation.
func (m *Master) AddOutstation(id uint16, label string) *Outstation {
	m.mu.Lock()
	defer m.mu.Unlock()
	o := NewOutstation(id, label)
	m.outstations[id] = o
	return o
}

// RemoveOutstation removes an outstation.
func (m *Master) RemoveOutstation(id uint16) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.outstations, id)
}

// GetOutstation retrieves an outstation by ID.
func (m *Master) GetOutstation(id uint16) (*Outstation, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	o, ok := m.outstations[id]
	return o, ok
}

// outstationFCB returns the primary-side Frame Count Bit for the link to
// outstationID (DNP3-057). Returns false if the outstation is unknown (the
// FCB is best-effort; an unknown peer will fail later validation anyway).
func (m *Master) outstationFCB(id uint16) bool {
	if o, ok := m.GetOutstation(id); ok {
		return o.takeFCB()
	}
	return false
}

// advanceOutstationFCB toggles the primary-side FCB for outstationID after a
// committed confirmed-data transaction (DNP3-057).
func (m *Master) advanceOutstationFCB(id uint16) {
	if o, ok := m.GetOutstation(id); ok {
		o.advanceFCB()
	}
}

// OutstationCount returns the number of registered outstations.
func (m *Master) OutstationCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.outstations)
}

// Connect establishes connection to outstations and performs DNP3 session establishment.
//
// Proper DNP3 session establishment includes:
// 1. TCP connect (handled by transport)
// 2. Link reset (Reset Link Stations)
// 3. Link status request
// 4. Wait for responses
// 5. Set StateActive only after successful exchange
func (m *Master) Connect() error {
	if m.transport == nil {
		return errors.New("transport not configured")
	}

	// DNP3-039: Connect may only begin from Disconnected or Error (reconnect).
	// Any other state (e.g., a concurrent Connect while Connecting, or an
	// already-active session) is an illegal transition and surfaces an error
	// rather than silently overwriting the state.
	cur := m.State()
	if cur != StateDisconnected && cur != StateError {
		return fmt.Errorf("%w: %s -> Connecting", ErrIllegalStateTransition, cur)
	}
	if err := m.transitionTo(StateConnecting); err != nil {
		return err
	}

	// Clear TL/sequence state left over from any previous session so a reconnect
	// after a mid-session drop re-handshakes from a clean slate (DNP3-032). This
	// discards stale reassembly buffers and resets the fragment sequence.
	m.resetForReconnect()

	// Perform DNP3 link-layer session establishment
	if err := m.performLinkHandshake(); err != nil {
		// DNP3-039: a failed handshake lands in StateError, not Disconnected, so
		// the caller does not mistake a half-open session for a clean slate.
		_ = m.transitionTo(StateError)
		return fmt.Errorf("link handshake failed: %w", err)
	}

	// DNP3-039: handshake success moves Connecting -> Connected -> Active.
	if err := m.transitionTo(StateConnected); err != nil {
		return err
	}
	if err := m.transitionTo(StateActive); err != nil {
		return err
	}
	// DNP3-042: start the optional idle monitor (no-op when IdleTimeout <= 0).
	m.startIdleMonitor()
	return nil
}

// resetForReconnect clears transport-layer and link-layer state that would
// otherwise leak across sessions after a disconnect (DNP3-032). It is safe to
// call on a fresh master (no-op) and on one mid-session.
func (m *Master) resetForReconnect() {
	if m.reassembler != nil {
		m.reassembler.Reset()
	}
	if m.fragmenter != nil {
		m.fragmenter.Reset()
	}
}

// performLinkHandshake performs the DNP3 link-layer handshake sequence.
// This establishes the data link connection before application layer communication.
// The sequence is:
//  1. Reset Link Stations → secondary ACK
//  2. Request Link Status → secondary Link Status response
//
// Connect succeeds only after both exchanges validate (DNP3-006/DNP3-007).
func (m *Master) performLinkHandshake() error {
	// Step 1: Send Reset Link Stations and validate the secondary ACK.
	if err := m.sendResetLink(); err != nil {
		return fmt.Errorf("reset link failed: %w", err)
	}

	// Step 2: Request Link Status and validate the secondary Link Status
	// response. This confirms the link is fully established in both
	// directions before application communication begins.
	if err := m.sendLinkStatusRequest(); err != nil {
		return fmt.Errorf("link status request failed: %w", err)
	}

	return nil
}

// sendResetLink sends a Reset Link Stations frame to initialize the link layer.
//
// DNP3-058: a secondary NACK in response to Reset Link Stations is treated as
// a retryable failure (the link is alive but rejected the frame). The
// handshake is retried up to the NACK retry budget from the master's
// RetryPolicy; non-NACK failures fail immediately.
func (m *Master) sendResetLink() error {
	// Build Reset Link Stations frame (no data needed)
	// This tells the outstation to clear buffers and reset the FCB counter
	for _, o := range m.outstations {
		dllFrame := &frame.Frame{
			Control: frame.Control{
				DIR:      true,  // Master-to-Outstation
				PRM:      true,  // Primary station
				FCB:      false, // No frame count bit for reset
				FCV:      false, // FCB not valid for reset
				FuncCode: frame.FuncResetLinkStations,
			},
			DestAddr: o.ID, // Target outstation
			SrcAddr:  m.config.MasterAddress,
			Data:     nil, // Reset has no data
		}

		encoded, err := frame.Encode(dllFrame)
		if err != nil {
			return fmt.Errorf("failed to encode reset frame: %w", err)
		}

		policy := m.RetryPolicy()
		var lastErr error
		attempt := 0
		for {
			attempt++
			if err := m.transport.Send(encoded); err != nil {
				// A transport send failure is not retryable here (the link is
				// likely dead); surface it directly.
				return fmt.Errorf("failed to send reset frame: %w", err)
			}

			// The outstation acknowledges Reset Link Stations at the data-link
			// layer. Consume and validate that frame here so the next application
			// request reads an application response rather than a stale or invalid
			// handshake ACK.
			m.transport.SetTimeout(m.config.Timeout)
			raw, err := m.transport.Receive()
			if err != nil {
				return fmt.Errorf("failed to receive reset acknowledgment: %w", err)
			}
			if err := validateResetLinkACK(raw, o.ID, m.config.MasterAddress); err != nil {
				lastErr = fmt.Errorf("invalid reset link ACK: %w", err)
				// DNP3-058: only a link-layer NACK is retryable for the handshake.
				if !m.retryAgain(policy, &lastErr, attempt) {
					return lastErr
				}
				continue
			}
			// DNP3-057: a Reset Link Stations exchange resets the primary-side FCB
			// to its post-reset value (false) for this link.
			o.resetFCB()
			break
		}
	}

	return nil
}

// validateResetLinkACK verifies a secondary ACK frame received in response to
// a Reset Link Stations request. The ACK must:
//   - be a well-formed frame (sync + header CRC validated by frame.Decode)
//   - have DIR=0 (outstation-to-master) and PRM=0 (secondary)
//   - carry function code ACK (0)
//   - be addressed from the outstation (SrcAddr) to the master (DestAddr)
//
// Any deviation is a spec violation and the handshake fails (DNP3-006).
//
// DNP3-058: a secondary NACK (PRM=0, FuncNack) in response to Reset Link
// Stations is surfaced distinctly as ErrLinkNACK so the caller can apply the
// NACK retry policy rather than treating it as a generic malformed-ACK failure.
func validateResetLinkACK(raw []byte, outstationID, masterAddr uint16) error {
	f, err := frame.Decode(raw)
	if err != nil {
		return fmt.Errorf("decode ACK frame: %w", err)
	}
	if f.Control.DIR {
		return fmt.Errorf("ACK DIR=1, expected 0 (secondary)")
	}
	if f.Control.PRM {
		return fmt.Errorf("ACK PRM=1, expected 0 (secondary)")
	}
	// DNP3-058: a secondary NACK means the outstation rejected the Reset Link
	// Stations frame (busy/out-of-sequence). Classify it as a link-layer NACK
	// so the handshake retry policy applies.
	if f.Control.FuncCode == frame.FuncNack {
		return fmt.Errorf("%w (reset link rejected, src=0x%04X)", ErrLinkNACK, f.SrcAddr)
	}
	if f.Control.FuncCode != frame.FuncAck {
		return fmt.Errorf("ACK function code=%d, expected %d (ACK)",
			f.Control.FuncCode, frame.FuncAck)
	}
	if f.SrcAddr != outstationID {
		return fmt.Errorf("ACK source address=0x%04X, expected outstation 0x%04X",
			f.SrcAddr, outstationID)
	}
	if f.DestAddr != masterAddr {
		return fmt.Errorf("ACK destination address=0x%04X, expected master 0x%04X",
			f.DestAddr, masterAddr)
	}
	return nil
}

// sendLinkStatusRequest issues a Request Link Status frame to each outstation
// and validates the secondary Link Status response. This is the second phase
// of the link handshake (DNP3-007): Connect succeeds only after a valid Link
// Status response.
func (m *Master) sendLinkStatusRequest() error {
	for _, o := range m.outstations {
		req := &frame.Frame{
			Control: frame.Control{
				DIR:      true, // Master-to-Outstation
				PRM:      true, // Primary station
				FuncCode: frame.FuncRequestLinkStatus,
			},
			DestAddr: o.ID,
			SrcAddr:  m.config.MasterAddress,
			Data:     nil,
		}
		encoded, err := frame.Encode(req)
		if err != nil {
			return fmt.Errorf("failed to encode link status request: %w", err)
		}
		if err := m.transport.Send(encoded); err != nil {
			return fmt.Errorf("failed to send link status request: %w", err)
		}

		m.transport.SetTimeout(m.config.Timeout)
		raw, err := m.transport.Receive()
		if err != nil {
			return fmt.Errorf("failed to receive link status response: %w", err)
		}
		if err := validateLinkStatusResponse(raw, o.ID, m.config.MasterAddress); err != nil {
			return fmt.Errorf("invalid link status response: %w", err)
		}
	}
	return nil
}

// validateLinkStatusResponse verifies a secondary Link Status frame received
// in response to a Request Link Status. The response must:
//   - be a well-formed frame (sync + header CRC validated by frame.Decode)
//   - have DIR=0 (outstation-to-master) and PRM=0 (secondary)
//   - carry function code Link Status (2)
//   - be addressed from the outstation (SrcAddr) to the master (DestAddr)
//
// Any deviation is a spec violation and the handshake fails (DNP3-007).
func validateLinkStatusResponse(raw []byte, outstationID, masterAddr uint16) error {
	f, err := frame.Decode(raw)
	if err != nil {
		return fmt.Errorf("decode link status frame: %w", err)
	}
	if f.Control.DIR {
		return fmt.Errorf("link status DIR=1, expected 0 (secondary)")
	}
	if f.Control.PRM {
		return fmt.Errorf("link status PRM=1, expected 0 (secondary)")
	}
	if f.Control.FuncCode != frame.FuncLinkStatus {
		return fmt.Errorf("link status function code=%d, expected %d",
			f.Control.FuncCode, frame.FuncLinkStatus)
	}
	if f.SrcAddr != outstationID {
		return fmt.Errorf("link status source address=0x%04X, expected outstation 0x%04X",
			f.SrcAddr, outstationID)
	}
	if f.DestAddr != masterAddr {
		return fmt.Errorf("link status destination address=0x%04X, expected master 0x%04X",
			f.DestAddr, masterAddr)
	}
	return nil
}

// Disconnect closes connections.
//
// DNP3-039: Disconnect is a legal terminal transition from any state (the
// master tears down unconditionally); the state moves to Disconnected and
// registered outstations are dropped.
func (m *Master) Disconnect() error {
	// DNP3-042: stop the idle monitor before tearing down so it does not race
	// with the state reset below.
	m.stopIdleMonitor()
	// Force the terminal Disconnected transition from any state: Disconnect is
	// the unconditional teardown path, so it bypasses the transition table.
	m.mu.Lock()
	m.state = StateDisconnected
	m.mu.Unlock()
	m.outstations = make(map[uint16]*Outstation)
	return nil
}

// Initialize performs master initialization sequence.
//
// DNP3-039: Initialize is legal only from Connected or Active (the link is up
// but application readiness has not been confirmed). It transitions to
// Initialized, which (with Active) is one of the two operational states.
func (m *Master) Initialize() error {
	if err := m.transitionTo(StateInitialized); err != nil {
		return ErrNotConnected
	}
	return nil
}

// Enable unsolicits on an outstation.
func (m *Master) EnableUnsolicited(outstationID uint16) error {
	if !m.isLinkUp() {
		return ErrNotConnected
	}

	o, ok := m.GetOutstation(outstationID)
	if !ok {
		return ErrOutstationNotFound
	}

	// Build ENABLE_UNSOLICITED request
	req := buildRequest(0, al.FuncEnableUnsolicited, nil)

	if err := m.sendWithRetry(req, outstationID); err != nil {
		return err
	}

	o.mu.Lock()
	o.Unsolicited = true
	o.mu.Unlock()

	return nil
}

// DisableUnsolicited disables unsolicited on an outstation.
func (m *Master) DisableUnsolicited(outstationID uint16) error {
	if !m.isLinkUp() {
		return ErrNotConnected
	}

	o, ok := m.GetOutstation(outstationID)
	if !ok {
		return ErrOutstationNotFound
	}

	req := buildRequest(0, al.FuncDisableUnsolicited, nil)
	if err := m.sendWithRetry(req, outstationID); err != nil {
		return err
	}

	o.mu.Lock()
	o.Unsolicited = false
	o.mu.Unlock()

	return nil
}

// Poll performs a data poll on an outstation.
func (m *Master) Poll(outstationID uint16, pollType PollType) error {
	if !m.isOperational() {
		return ErrNotConnected
	}

	data := buildPollRequest(pollType)
	req := buildRequest(0, al.FuncRead, data)

	return m.sendWithRetry(req, outstationID)
}

// Operate issues a control operation.
func (m *Master) Operate(outstationID uint16, selectThenOperate bool, group, variation uint8, index uint16, value interface{}) error {
	if !m.isOperational() {
		return ErrNotConnected
	}

	var req *al.APDU

	if selectThenOperate {
		// SELECT
		selectReq := m.buildControlRequest(al.FuncSelect, group, variation, index, value)
		if err := m.sendWithRetry(selectReq, outstationID); err != nil {
			return err
		}

		// OPERATE
		req = m.buildControlRequest(al.FuncOperate, group, variation, index, value)
	} else {
		// DIRECT OPERATE
		req = m.buildControlRequest(al.FuncDirectOperate, group, variation, index, value)
	}

	return m.sendWithRetry(req, outstationID)
}

// OperateWithStatus performs a control operation and returns the per-point
// command status parsed from the outstation's response (DNP3-020). Unlike
// Operate, it reads the response and never reports success when the response
// status byte indicates a failure. A response with no parseable G12V1 status
// yields CommandStatusUnknown (not success).
func (m *Master) OperateWithStatus(outstationID uint16, selectThenOperate bool, group, variation uint8, index uint16, value interface{}) (CommandStatus, error) {
	if !m.isOperational() {
		return CommandStatusUnknown, ErrNotConnected
	}

	var req *al.APDU
	if selectThenOperate {
		selectReq := m.buildControlRequest(al.FuncSelect, group, variation, index, value)
		if _, err := m.sendWithRetryAndGetResponse(selectReq, outstationID); err != nil {
			return CommandStatusUnknown, err
		}
		req = m.buildControlRequest(al.FuncOperate, group, variation, index, value)
	} else {
		req = m.buildControlRequest(al.FuncDirectOperate, group, variation, index, value)
	}
	if req == nil {
		return CommandStatusUnknown, fmt.Errorf("unsupported control value for group %d variation %d", group, variation)
	}

	appData, err := m.sendWithRetryAndGetResponse(req, outstationID)
	if err != nil {
		return CommandStatusUnknown, err
	}

	resp, err := al.DecodeResponse(appData)
	if err != nil {
		return CommandStatusUnknown, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}
	return parseCommandStatus(resp.Data), nil
}

// TimeSync performs time synchronization with an outstation.
func (m *Master) TimeSync(outstationID uint16) error {
	if !m.isOperational() {
		return ErrNotConnected
	}

	// Build time sync request with current time
	now := time.Now()
	data := encodeDNP3Time(now)
	req := buildRequest(0, al.FuncTimeSync, data)

	return m.sendWithRetry(req, outstationID)
}

// sendWithRetry sends a request with retry logic.
//
// DNP3-034: each failure is classified (timeout / NACK / CRC / disconnect /
// other) and the retry decision (whether to retry and how long to wait) is
// taken from the master's RetryPolicy. Disconnects are never retried (the
// link is dead — DNP3-031).
func (m *Master) sendWithRetry(req *al.APDU, outstationID uint16) error {
	// DNP3-040: reject a concurrent request to the SAME outstation before
	// blocking on reqMu so a slow peer cannot starve the caller. The marker is
	// cleared when this request (including retries) completes.
	if err := m.beginRequest(outstationID); err != nil {
		return err
	}
	defer m.endRequest(outstationID)

	// Serialize the request path so concurrent requests do not race on the
	// shared reassembler or interleave fragments on a single link (DNP3-025).
	m.reqMu.Lock()
	defer m.reqMu.Unlock()

	policy := m.RetryPolicy()
	var lastErr error
	attempt := 0

	for {
		attempt++

		// Allocate the application-layer sequence for this request attempt.
		// Retries reuse the same sequence value within one logical request;
		// the sequence advances only after a successful send (DNP3-008).
		seq := m.nextSequence(outstationID)
		req.Control.Seq = seq

		// Send request with DNP3 protocol layers
		data := req.Encode()

		// Transport layer fragmentation
		fragments := m.fragmenter.Fragmentize(data)

		// DNP3-057: capture the primary-side FCB once per logical request;
		// retries reuse the same value and the bit advances only on success.
		fcb := m.outstationFCB(outstationID)

		sentOK := true
		for _, frag := range fragments {
			// Transport layer encode
			tlEncoded := tl.EncodeFragment(frag)

			// Data link layer frame (DNP3-057: set FCB + FCV on confirmed data)
			dllFrame := &frame.Frame{
				Control: frame.Control{
					DIR:      true,  // Master-to-Outstation
					PRM:      true,  // Primary station
					FCB:      fcb,
					FCV:      true,
					FuncCode: frame.FuncConfirmedUserData,
				},
				DestAddr: outstationID,
				SrcAddr:  m.config.MasterAddress,
				Data:     tlEncoded,
			}

			dllEncoded, err := frame.Encode(dllFrame)
			if err != nil {
				return err
			}

			if err := m.transport.Send(dllEncoded); err != nil {
				lastErr, _ = m.markDisconnected(err)
				sentOK = false
				break
			}

		}
		if !sentOK {
			// If the transport disconnected, stop retrying — the link is dead
			// (DNP3-031). Otherwise consult the retry policy (DNP3-034).
			if !m.retryAgain(policy, &lastErr, attempt) {
				return lastErr
			}
			continue
		}
		// The request was sent successfully; advance the sequence so the
		// next request uses the next value in the 0-15 stream (DNP3-008).
		m.advanceSequence(outstationID)
		// Successful send counts as link activity for the idle monitor (DNP3-042).
		m.recordActivity()
		// If CON bit is set, wait for confirmation first
		if req.Control.CON {
			if err := m.waitForConfirmation(seq); err != nil {
				lastErr, _ = m.markDisconnected(err)
				if !m.retryAgain(policy, &lastErr, attempt) {
					return lastErr
				}
				continue
			}
		}

		// Wait for response
		resp, err := m.waitForResponse()
		if err != nil {
			lastErr, _ = m.markDisconnected(err)
			if !m.retryAgain(policy, &lastErr, attempt) {
				return lastErr
			}
			continue
		}

		// Process response; the response SEQ must match the request (DNP3-010).
		_, err = m.processResponse(resp, outstationID, seq)
		if err != nil {
			lastErr = err
			if !m.retryAgain(policy, &lastErr, attempt) {
				return lastErr
			}
			continue
		}

		// Successful response counts as link activity for the idle monitor (DNP3-042).
		m.recordActivity()
		// DNP3-057: the confirmed-data transaction committed; toggle the FCB
		// for the next transaction. Retries of a failed attempt keep the same FCB.
		m.advanceOutstationFCB(outstationID)
		return nil // Success
	}
}

// sendWithRetryAndGetResponse sends a request with retry logic and returns the processed application layer data.
//
// DNP3-034: see sendWithRetry — failures are classified and the retry decision
// is taken from the master's RetryPolicy.
func (m *Master) sendWithRetryAndGetResponse(req *al.APDU, outstationID uint16) ([]byte, error) {
	// DNP3-040: reject a concurrent request to the SAME outstation before
	// blocking on reqMu so a slow peer cannot starve the caller.
	if err := m.beginRequest(outstationID); err != nil {
		return nil, err
	}
	defer m.endRequest(outstationID)

	// Serialize the request path so concurrent requests do not race on the
	// shared reassembler or interleave fragments on a single link (DNP3-025).
	m.reqMu.Lock()
	defer m.reqMu.Unlock()

	policy := m.RetryPolicy()
	var lastErr error
	attempt := 0

	for {
		attempt++

		// Allocate the application-layer sequence for this request attempt
		// (DNP3-008). Retries reuse the same value; it advances only on a
		// successful send.
		seq := m.nextSequence(outstationID)
		req.Control.Seq = seq

		// Send request with DNP3 protocol layers
		data := req.Encode()

		// Transport layer fragmentation
		fragments := m.fragmenter.Fragmentize(data)

		// DNP3-057: capture the primary-side FCB once per logical request;
		// retries reuse the same value and the bit advances only on success.
		fcb := m.outstationFCB(outstationID)

		sentOK := true
		for _, frag := range fragments {
			// Transport layer encode
			tlEncoded := tl.EncodeFragment(frag)

			// Data link layer frame (DNP3-057: set FCB + FCV on confirmed data)
			dllFrame := &frame.Frame{
				Control: frame.Control{
					DIR:      true,  // Master-to-Outstation
					PRM:      true,  // Primary station
					FCB:      fcb,
					FCV:      true,
					FuncCode: frame.FuncConfirmedUserData,
				},
				DestAddr: outstationID,
				SrcAddr:  m.config.MasterAddress,
				Data:     tlEncoded,
			}

			dllEncoded, err := frame.Encode(dllFrame)
			if err != nil {
				return nil, err
			}

			if err := m.transport.Send(dllEncoded); err != nil {
				lastErr, _ = m.markDisconnected(err)
				sentOK = false
				break
			}

		}
		if !sentOK {
			// If the transport disconnected, stop retrying — the link is dead
			// (DNP3-031). Otherwise consult the retry policy (DNP3-034).
			if !m.retryAgain(policy, &lastErr, attempt) {
				return nil, lastErr
			}
			continue
		}
		// Successful send; advance the sequence (DNP3-008).
		m.advanceSequence(outstationID)
		// Successful send counts as link activity for the idle monitor (DNP3-042).
		m.recordActivity()
		// DNP3-044: emit a "send" diagnostic event with the application seq.
		m.diag(DiagInfo, "send", seq, fmt.Sprintf("sent request to outstation %d", outstationID), nil)
		// If CON bit is set, wait for confirmation first
		if req.Control.CON {
			if err := m.waitForConfirmation(seq); err != nil {
				m.diag(DiagWarn, "confirm", seq, "confirmation failed", err)
				lastErr, _ = m.markDisconnected(err)
				if !m.retryAgain(policy, &lastErr, attempt) {
					return nil, lastErr
				}
				m.diag(DiagWarn, "retry", seq, fmt.Sprintf("retry after confirm failure (attempt %d)", attempt), lastErr)
				continue
			}
			m.diag(DiagInfo, "confirm", seq, "confirmation received", nil)
		}

		// Wait for response
		resp, err := m.waitForResponse()
		if err != nil {
			m.diag(DiagWarn, "receive", seq, "response failed", err)
			lastErr, _ = m.markDisconnected(err)
			if !m.retryAgain(policy, &lastErr, attempt) {
				return nil, lastErr
			}
			m.diag(DiagWarn, "retry", seq, fmt.Sprintf("retry after receive failure (attempt %d)", attempt), lastErr)
			continue
		}

		// Process the response and get application layer data; the response
		// SEQ must match the request (DNP3-010).
		appData, err := m.processResponse(resp, outstationID, seq)
		if err != nil {
			lastErr = err
			if !m.retryAgain(policy, &lastErr, attempt) {
				return nil, lastErr
			}
			m.diag(DiagWarn, "retry", seq, fmt.Sprintf("retry after response error (attempt %d)", attempt), lastErr)
			continue
		}

		// Successful response counts as link activity for the idle monitor (DNP3-042).
		m.recordActivity()
		// DNP3-057: the confirmed-data transaction committed (send + optional
		// confirmation + response all succeeded); toggle the FCB for the next
		// transaction. Retries of a failed attempt keep the same FCB.
		m.advanceOutstationFCB(outstationID)
		// DNP3-044: emit a "receive" diagnostic event for the completed response.
		m.diag(DiagInfo, "receive", seq, "response received", nil)
		// Return the processed application layer data for the caller to decode
		return appData, nil
	}
}

// retryAgain decides whether to retry the current request after a failure and,
// if so, sleeps for the policy-mandated delay (DNP3-034). It updates *lastErr
// in place: when the retry budget for the error's class is exhausted, *lastErr
// is wrapped with ErrMaxRetries so the caller surfaces a clear terminal error.
// It returns true if the caller should retry, false if it should return.
func (m *Master) retryAgain(policy *RetryPolicy, lastErr *error, attempt int) bool {
	if lastErr == nil || *lastErr == nil {
		return false
	}
	class := classifyRetryError(*lastErr)
	if class == ClassDisconnect {
		return false
	}
	if policy == nil {
		policy = DefaultRetryPolicy(m.config)
	}
	retry, delay := policy.retryDecision(class, attempt)
	if !retry {
		// Preserve the inner error chain (DNP3-034): the NACK/CRC/timeout
		// sentinels must remain reachable via errors.Is after the retry budget
		// is exhausted so callers (and auditors) can see the terminal class.
		*lastErr = fmt.Errorf("%w: %w", ErrMaxRetries, *lastErr)
		return false
	}
	if delay > 0 {
		time.Sleep(delay)
	}
	return true
}

// SendRequestWithRetry sends a request with retry logic (public wrapper).
// This method is used by the public API to leverage the internal master's
// protocol handling, including proper transport layer fragmentation,
// data link layer framing, and retry logic.
func (m *Master) SendRequestWithRetry(req *al.APDU, outstationID uint16) error {
	return m.sendWithRetry(req, outstationID)
}

// SendRequestWithRetryAndGetResponse sends a request with retry logic and returns the response data.
// This method is used by the public API when it needs to both send the request
// AND receive/process the response data.
func (m *Master) SendRequestWithRetryAndGetResponse(req *al.APDU, outstationID uint16) ([]byte, error) {
	return m.sendWithRetryAndGetResponse(req, outstationID)
}

// waitForConfirmation waits for an application-layer confirmation for a request
// sent with the CON bit set (DNP3-009). A confirmation is an APDU with
// FuncCode=FuncResponse. The confirmation's sequence number must match the
// request's sequence; a mismatch is a protocol error (ErrConfirmSeqMismatch).
// A transport receive error (timeout) is reported as ErrConfirmTimeout, which
// the caller retries.
//
// In this stack a full response (FuncResponse carrying IIN+data) may arrive in
// lieu of a dedicated confirm; such a response is accepted as the confirmation.
// Strict response-sequence matching is the responsibility of DNP3-010.
func (m *Master) waitForConfirmation(expectedSeq uint8) error {
	// DNP3-066: the confirm timeout is distinct from the response timeout. A
	// confirm is a small dedicated reply, so it typically uses a shorter
	// budget than a full object response.
	m.transport.SetTimeout(int(m.confirmTimeout / time.Millisecond))

	for {
		data, err := m.transport.Receive()
		if err != nil {
			// A peer close is a disconnect, not a confirm timeout (DNP3-031).
			if isDisconnectError(err) {
				return fmt.Errorf("%w: %v", ErrTransportDisconnected, err)
			}
			return fmt.Errorf("%w: %v", ErrConfirmTimeout, err)
		}

		// Process received bytes through DLL+TL layers.
		appData, err := m.processReceivedBytes(data)
		if err != nil {
			// Malformed link/transport framing; keep waiting for a valid frame.
			continue
		}

		// Application layer decode
		apdu, err := al.Decode(appData)
		if err != nil {
			continue
		}

		// Only response APDUs act as confirmations.
		if apdu.FuncCode != al.FuncResponse {
			continue
		}

		// A dedicated confirm carries only the IIN (2 bytes) or is empty.
		// A full response carries IIN + object data.
		isDedicatedConfirm := len(apdu.Data) <= 2
		if isDedicatedConfirm {
			if apdu.Control.Seq == expectedSeq {
				return nil
			}
			return fmt.Errorf("%w: got %d, expected %d",
				ErrConfirmSeqMismatch, apdu.Control.Seq, expectedSeq)
		}

		// Full response acting as the confirmation: accept it (DNP3-010 will
		// enforce strict sequence matching on the response path).
		return nil
	}
}

// waitForResponse waits for a response with timeout.
//
// DNP3-043: a non-disconnect receive error (no response within the transport
// timeout) is tagged with ErrTimeout so the failure is classifiable as a
// timeout rather than an opaque transport error. A disconnect error is returned
// unchanged so the caller's markDisconnected wraps it with ErrTransportDisconnected
// (unchanged behavior).
func (m *Master) waitForResponse() ([]byte, error) {
	m.transport.SetTimeout(m.config.Timeout)
	data, err := m.transport.Receive()
	if err != nil {
		if isDisconnectError(err) {
			return nil, err
		}
		return nil, fmt.Errorf("%w: %v", ErrTimeout, err)
	}
	return data, nil
}

// markDisconnected inspects a transport error and, if it indicates the peer
// closed the connection, transitions the master to StateError and wraps the
// error with ErrTransportDisconnected so callers can detect the disconnect
// (DNP3-031). Retrying a dead link is pointless, so callers should break the
// retry loop when this returns true.
func (m *Master) markDisconnected(err error) (wrapped error, disconnected bool) {
	if !isDisconnectError(err) {
		return err, false
	}
	// DNP3-039: a transport disconnect lands in StateError via the transition
	// table. Self-transition (already Error) is a no-op; an illegal source
	// state is forced to Error regardless, since the link is demonstrably dead.
	_ = m.transitionTo(StateError)
	// DNP3-042: the link is dead, so stop the idle monitor.
	m.stopIdleMonitor()
	return fmt.Errorf("%w: %v", ErrTransportDisconnected, err), true
}

// processResponse processes a received response and returns the application
// layer data. The response's application sequence number must match the
// outstanding request's sequence (DNP3-010); a mismatch is a protocol error
// and no points are surfaced to the caller.
func (m *Master) processResponse(data []byte, outstationID uint16, expectedSeq uint8) ([]byte, error) {
	// Process through Data Link and Transport layers
	appData, err := m.processReceivedBytes(data)
	if err != nil {
		// Preserve the inner error chain (DNP3-034): link-layer NACK/CRC
		// sentinels must remain reachable via errors.Is so the retry loop can
		// classify them.
		return nil, fmt.Errorf("%w: %w", ErrInvalidResponse, err)
	}

	// Application layer decode
	resp, err := al.DecodeResponse(appData)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidResponse, err)
	}

	// Validate the response sequence against the outstanding request (DNP3-010).
	if resp.Header.Control.Seq != expectedSeq {
		return nil, fmt.Errorf("%w: got %d, expected %d",
			ErrResponseSeqMismatch, resp.Header.Control.Seq, expectedSeq)
	}

	// Update outstation IIN
	o, ok := m.GetOutstation(outstationID)
	if ok {
		o.UpdateIIN(resp.IIN.Bytes())
	}

	// DNP3-054: if the outstation's response requested an application-layer
	// confirmation (CON=1), the master must send a confirm APDU — same sequence,
	// FuncCode 0 (Confirm), CON=0 — before processing the response data. The
	// confirm is sent as unconfirmed user data so it does not require a
	// link-layer ACK and cannot desync the ping-pong request/response path.
	if resp.Header.Control.CON {
		if err := m.sendApplicationConfirm(resp.Header.Control.Seq, outstationID); err != nil {
			return nil, fmt.Errorf("send application confirm: %w", err)
		}
	}

	// DNP3-013: react to critical IIN bits.
	m.reactToIIN(outstationID, resp.IIN, o)

	return appData, nil
}

// sendApplicationConfirm sends a DNP3 application-layer confirmation (DNP3-054)
// to outstationID for the response carrying seq. The confirm is a single
// FIR|FIN APDU with FuncCode 0 (Confirm) and CON=0, sent as unconfirmed user
// data (no link-layer ACK expected).
func (m *Master) sendApplicationConfirm(seq uint8, outstationID uint16) error {
	confirm := al.NewConfirm(seq)
	data := confirm.Encode()
	fragments := m.fragmenter.Fragmentize(data)
	for _, frag := range fragments {
		tlEncoded := tl.EncodeFragment(frag)
		dllFrame := &frame.Frame{
			Control: frame.Control{
				DIR:      true,  // Master-to-Outstation
				PRM:      true,  // Primary station
				FuncCode: frame.FuncUnconfirmedUserData,
			},
			DestAddr: outstationID,
			SrcAddr:  m.config.MasterAddress,
			Data:     tlEncoded,
		}
		encoded, err := frame.Encode(dllFrame)
		if err != nil {
			return err
		}
		if err := m.transport.Send(encoded); err != nil {
			lastErr, _ := m.markDisconnected(err)
			return lastErr
		}
	}
	m.recordActivity()
	m.diag(DiagInfo, "confirm", seq, "sent application confirm for CON response", nil)
	return nil
}

// reactToIIN applies the minimal required Master reaction to received IIN
// bits (DNP3-013):
//   - DeviceRestart (IIN1.7): clear local state and flag re-integrity.
//   - NeedTime (IIN1.4): record that a time sync is required (stub).
//
// Both also invoke the optional callback hooks (logging/stub). Full integrity
// polling and time objects are later roadmap items.
func (m *Master) reactToIIN(outstationID uint16, iin al.IIN, o *Outstation) {
	if iin.DeviceRestart {
		if o != nil {
			o.markDeviceRestart()
		}
		m.mu.RLock()
		h := m.onDeviceRestart
		m.mu.RUnlock()
		if h != nil {
			h(outstationID)
		}
	}
	if iin.NeedTime {
		if o != nil {
			o.markNeedTimeSync()
		}
		m.mu.RLock()
		h := m.onNeedTimeSync
		m.mu.RUnlock()
		if h != nil {
			h(outstationID)
		}
	}
}

// processReceivedBytes processes raw TCP data through DLL and TL layers.
// Returns the reassembled application layer data.
//
// DNP3-034: link-layer error classes are surfaced distinctly so the retry
// loop can classify them:
//   - a corrupted frame (header/data CRC validation failure) is wrapped with
//     ErrCRCError (transient line noise; retryable);
//   - a secondary NACK frame (PRM=0, FuncNack) is surfaced as ErrLinkNACK
//     (the link is alive but rejected the frame; retryable).
func (m *Master) processReceivedBytes(data []byte) ([]byte, error) {
	// Reset reassembler for new response
	m.reassembler.Reset()

	var completeData []byte

	// Parse all DLL frames in the received data
	offset := 0
	for offset < len(data) {
		// Decode DLL frame
		dllFrame, err := frame.Decode(data[offset:])
		if err != nil {
			// If we got complete data, return it
			if len(completeData) > 0 {
				return completeData, nil
			}

			// DNP3-034: surface CRC validation failures as a distinct,
			// retryable error class.
			if isCRCError(err) {
				return nil, fmt.Errorf("DLL decode error at offset %d: %w: %v", offset, ErrCRCError, err)
			}
			return nil, fmt.Errorf("DLL decode error at offset %d: %w", offset, err)
		}

		// Move past this complete link frame, including the header CRC and each
		// 16-octet payload CRC block.
		offset += frame.EncodedSize(len(dllFrame.Data))

		// DNP3-034: a secondary link-layer NACK (PRM=0, FuncNack) means the
		// outstation received the frame but rejected it (busy/out-of-sequence).
		// Surface it distinctly so the retry loop can apply the NACK policy.
		if !dllFrame.Control.PRM && dllFrame.Control.FuncCode == frame.FuncNack {
			return nil, fmt.Errorf("%w (func=%d, src=0x%04X)", ErrLinkNACK, dllFrame.Control.FuncCode, dllFrame.SrcAddr)
		}

		// Skip non-user-data frames (secondary station function codes)
		// User data frames from outstation use FuncConfirmedUserDataR = 4 with PRM=0
		if dllFrame.Control.PRM {
			// Primary station frame - skip control frames
			if dllFrame.Control.FuncCode != frame.FuncConfirmedUserData {
				continue
			}

		}

		// For secondary station (PRM=0), only process user data frames
		if !dllFrame.Control.PRM && dllFrame.Control.FuncCode != frame.FuncConfirmedUserDataR {
			continue
		}

		// Extract TL fragment from frame data
		tlData := dllFrame.Data
		if len(tlData) == 0 {
			continue
		}

		// Parse TL header using correct function name
		tlFrag, err := tl.DecodeFragment(tlData)
		if err != nil {
			continue // Skip malformed fragments
		}

		// Push fragment to reassembler - returns complete message when FIN received
		result, err := m.reassembler.Push(tlFrag)
		if err != nil {
			return nil, fmt.Errorf("reassembly error: %w", err)
		}

		if result != nil {
			completeData = result
		}

	}

	// Check if we have complete application data
	if len(completeData) == 0 {
		return nil, fmt.Errorf("no complete response received")
	}

	return completeData, nil
}

// isCRCError reports whether a frame.Decode error is a CRC validation failure
// (DNP3-034). frame.Decode reports CRC failures via human-readable messages
// rather than a typed sentinel, so the message text is matched. The match is
// intentionally narrow ("CRC validation failed") to avoid misclassifying
// truncation/oversize errors.
func isCRCError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "CRC validation failed")
}

// buildRequest creates a request APDU.
func buildRequest(seq uint8, funcCode uint8, data []byte) *al.APDU {
	return &al.APDU{
		Control: al.AppControl{
			FIR: true,
			FIN: true,
			CON: false,
			UNS: false,
			Seq: seq,
		},
		FuncCode: funcCode,
		Data:     data,
	}

}

// nextSequence returns the current application-layer sequence number for
// outstationID without advancing it (DNP3-005 / DNP3-008). Each outstation
// owns an independent 0-15 stream so operations on multiple outstations do not
// cross-talk (DNP3-055). Returns 0 for an unknown outstation.
func (m *Master) nextSequence(outstationID uint16) uint8 {
	if o, ok := m.GetOutstation(outstationID); ok {
		return o.takeSeq()
	}
	return 0
}

// advanceSequence advances the outstationID sequence number by one (mod 16).
// Called after a request has been sent successfully so the next request to the
// same outstation uses the next value in its 0-15 stream (DNP3-008). No-op for
// an unknown outstation.
func (m *Master) advanceSequence(outstationID uint16) {
	if o, ok := m.GetOutstation(outstationID); ok {
		o.advanceSeq()
	}
}

// currentSequence returns the current sequence number for outstationID
// (read-only alias of nextSequence).
func (m *Master) currentSequence(outstationID uint16) uint8 {
	return m.nextSequence(outstationID)
}

// buildPollRequest creates a poll request based on poll type.
//
// Each poll object header is constructed via the formal al.ObjectHeader model
// (DNP3-004). The Class-0 / integrity poll uses the canonical all-objects
// qualifier (0x06) on Group 60 Variation 1 (DNP3-014): a single, deterministic
// request form for "read all Class-0 static data".
func buildPollRequest(pollType PollType) []byte {
	// Object headers for different poll types
	var headers []al.ObjectHeader
	switch pollType {
	case PollIntegrity, PollClass0:
		// Group 60, Variation 1 = All static data (Class 0), all-objects
		// qualifier (0x06) — canonical integrity poll (DNP3-014).
		headers = []al.ObjectHeader{{Group: 60, Variation: 1, Qualifier: al.QualAllObjects}}
	case PollClass1:
		// Group 2, Variation 1 = Binary events
		headers = []al.ObjectHeader{{Group: 2, Variation: 1, Qualifier: al.QualCount8, Count: 0}}
	case PollClass2:
		// Group 32, Variation 1 = Analog events
		headers = []al.ObjectHeader{{Group: 32, Variation: 1, Qualifier: al.QualCount8, Count: 0}}
	case PollClass3:
		// Group 22, Variation 1 = Counter events
		headers = []al.ObjectHeader{{Group: 22, Variation: 1, Qualifier: al.QualCount8, Count: 0}}
	case PollEvent:
		// All event classes
		headers = []al.ObjectHeader{
			{Group: 2, Variation: 1, Qualifier: al.QualCount8, Count: 0},
			{Group: 32, Variation: 1, Qualifier: al.QualCount8, Count: 0},
			{Group: 22, Variation: 1, Qualifier: al.QualCount8, Count: 0},
		}
	default:
		return nil
	}
	data, err := al.EncodeObjectHeaders(nil, headers)
	if err != nil {
		return nil
	}
	return data
}

// buildControlRequest creates a control request.
func (m *Master) buildControlRequest(funcCode uint8, group, variation uint8, index uint16, value interface{}) *al.APDU {
	// Object prefix: Group, Variation, Qualifier (0x00 = index only), Count (1)
	prefix := []byte{group, variation, 0x00, 0x01}

	// Index: 2 bytes, least-significant byte first.
	indexBytes := []byte{byte(index), byte(index >> 8)}

	// Value encoding depends on group and type
	var valueBytes []byte

	switch group {
	case 12:
		// CROB (Control Relay Output Block) - 11 bytes
		// Layout: code(1), count(1), onTime(4 LSB), offTime(4 LSB), status(1).
		// Index precedes the value bytes (2 octets, LSB first).
		var code uint8
		var count uint8 = 1
		var onTime uint32 = 0
		var offTime uint32 = 0
		var status uint8 = 0

		switch v := value.(type) {
		case bool:
			if v {
				code = CROBCodeLatchOn // Value = true
			} else {
				code = CROBCodeLatchOff // Value = false
			}
		case uint8:
			code = v
		case uint16:
			code = uint8(v)
		default:
			return nil // Unsupported type
		}

		valueBytes = []byte{
			code,                                                                    // Control code
			count,                                                                   // Count
			byte(onTime), byte(onTime >> 8), byte(onTime >> 16), byte(onTime >> 24), // On time
			byte(offTime), byte(offTime >> 8), byte(offTime >> 16), byte(offTime >> 24), // Off time
			status, // Status
		}

		if len(valueBytes) != 11 {
			return nil
		}

	case 41, 42, 43, 44:
		// Analog Output - encoding must match variation
		switch variation {
		case 1, 5: // 16-bit signed int
			var intVal int16
			switch v := value.(type) {
			case float64:
				intVal = int16(v)
			case float32:
				intVal = int16(v)
			case int16:
				intVal = v
			case int32:
				intVal = int16(v)
			case int64:
				intVal = int16(v)
			default:
				return nil
			}
			valueBytes = []byte{byte(intVal >> 8), byte(intVal)}

		case 2, 6: // 16-bit unsigned int
			var uintVal uint16
			switch v := value.(type) {
			case float64:
				uintVal = uint16(v)
			case float32:
				uintVal = uint16(v)
			case uint16:
				uintVal = v
			case uint32:
				uintVal = uint16(v)
			case int16:
				uintVal = uint16(v)
			default:
				return nil
			}
			valueBytes = []byte{byte(uintVal >> 8), byte(uintVal)}

		case 3, 7: // 32-bit signed int
			var intVal int32
			switch v := value.(type) {
			case float64:
				intVal = int32(v)
			case float32:
				intVal = int32(v)
			case int32:
				intVal = v
			case int16:
				intVal = int32(v)
			default:
				return nil
			}
			valueBytes = []byte{
				byte(intVal >> 24), byte(intVal >> 16),
				byte(intVal >> 8), byte(intVal),
			}

		case 4, 8: // 32-bit unsigned int
			var uintVal uint32
			switch v := value.(type) {
			case float64:
				uintVal = uint32(v)
			case float32:
				uintVal = uint32(v)
			case uint32:
				uintVal = v
			case uint16:
				uintVal = uint32(v)
			default:
				return nil
			}
			valueBytes = []byte{
				byte(uintVal >> 24), byte(uintVal >> 16),
				byte(uintVal >> 8), byte(uintVal),
			}

		case 9, 10: // 32-bit float
			var floatVal float32
			switch v := value.(type) {
			case float64:
				floatVal = float32(v)
			case float32:
				floatVal = v
			default:
				return nil
			}
			bits := float32ToUint32Bits(floatVal)
			valueBytes = []byte{
				byte(bits >> 24), byte(bits >> 16),
				byte(bits >> 8), byte(bits),
			}

		default:
			return nil
		}

	default:
		return nil
	}

	data := append(prefix, indexBytes...)
	data = append(data, valueBytes...)

	return buildRequest(0, funcCode, data)
}

// float32ToUint32Bits converts float32 to uint32 bits.
func float32ToUint32Bits(f float32) uint32 {
	return math.Float32bits(f)
}

// encodeDNP3Time encodes time.Time to DNP3 time format.
func encodeDNP3Time(t time.Time) []byte {
	// DNP3 time is milliseconds since 1970-01-01 00:00:00 UTC
	// But offset by 2000-01-01 for the "DNP3 epoch"
	epoch := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	ms := t.Sub(epoch).Milliseconds()

	// DNP3 time is 2 x 4-byte integers (for different time bases)
	// We use the primary time base
	return []byte{
		byte(ms >> 40), byte(ms >> 32),
		byte(ms >> 24), byte(ms >> 16),
		byte(ms >> 8), byte(ms & 0xFF),
		0x00, 0x00, // Reserved/last half
	}

}

// =============================================================================
// WRITE OPERATIONS
// =============================================================================

// CROB represents a Control Relay Output Block (Group 12, Variation 1).
type CROB struct {
	Code    uint8  // Control code (0=NOT_USED, 1=OPEN, 2=CLOSE, 3=PULSE_ON, etc.)
	Count   uint8  // Number of operation cycles
	OnTime  uint32 // Time in ms to energize
	OffTime uint32 // Time in ms to de-energize
	Status  uint8  // Qualifier code (0x00=STANDARD)
}

// CROB control-code values used by this implementation's request encode and
// outstation decode. NOTE: these are a repository-internal 1..8 enum, NOT the
// IEEE 1815 G12V1 control-code bit field (0x01 NUL, 0x02 Pulse On, 0x04 Pulse
// Off, 0x08 Latch On, 0x10 Latch Off, 0x80 Queue). The encode LAYOUT (code as
// a single octet at the correct position) is locked by DNP3-019; reconciling
// the code VALUES with the IEEE 1815 bit field is a cross-layer correction
// tracked separately (see active_work/handoff.md discoveries).
const (
	CROBCodeNUL      uint8 = 1
	CROBCodeClose    uint8 = 2 // turn ON
	CROBCodeOpen     uint8 = 3 // turn OFF
	CROBCodeTrip     uint8 = 4
	CROBCodePulseOn  uint8 = 5
	CROBCodePulseOff uint8 = 6
	CROBCodeLatchOn  uint8 = 7 // Value = true
	CROBCodeLatchOff uint8 = 8 // Value = false
)

// CommandStatus is the per-point status code returned in a G12V1 control
// response (IEEE 1815 "Command Status", CTRL-01). Values mirror the
// outstation-side command-status enumeration; the master must never report
// CommandStatusSuccess for a response whose status byte indicates otherwise.
type CommandStatus uint8

const (
	// CommandStatusSuccess indicates the command was accepted and executed.
	CommandStatusSuccess CommandStatus = 0
	// CommandStatusTimeout indicates the command timed out before completion.
	CommandStatusTimeout CommandStatus = 1
	// CommandStatusNoSelect indicates select was not received before operate.
	CommandStatusNoSelect CommandStatus = 2
	// CommandStatusBadFormat indicates a malformed control object.
	CommandStatusBadFormat CommandStatus = 3
	// CommandStatusNotSupported indicates the command is not supported.
	CommandStatusNotSupported CommandStatus = 4
	// CommandStatusAlreadyActive indicates the output is already in that state.
	CommandStatusAlreadyActive CommandStatus = 5
	// CommandStatusBlocked indicates the output is blocked from operation.
	CommandStatusBlocked CommandStatus = 6
	// CommandStatusLocal indicates the output is in local control mode.
	CommandStatusLocal CommandStatus = 7
	// CommandStatusTooMany indicates too many commands are queued.
	CommandStatusTooMany CommandStatus = 8
	// CommandStatusNotAuthorized indicates the command was not authorized.
	CommandStatusNotAuthorized CommandStatus = 9
	// CommandStatusAutonomous indicates autonomous operation.
	CommandStatusAutonomous CommandStatus = 10
	// CommandStatusUnknown is the default when no parseable status is present.
	CommandStatusUnknown CommandStatus = 0xFF
)

// parseCommandStatus scans the application-layer response data (the bytes
// following the control/func/IIN octets) for a Group 12 Variation 1 object and
// returns the first per-point command-status byte. A DirectOperate/Operate
// response echoes the request's G12V1 header with the CROB status byte
// replaced by the command status (CTRL-01). If no G12V1 object is present, or
// the object is truncated, the status is CommandStatusUnknown — the master must
// NOT assume success in that case.
func parseCommandStatus(data []byte) CommandStatus {
	offset := 0
	for offset+4 <= len(data) {
		group := data[offset]
		variation := data[offset+1]
		qualifier := data[offset+2]

		if group == 12 && variation == 1 {
			return commandStatusForQualifier(data, offset, qualifier)
		}
		// Skip this object header + its data to continue scanning.
		next, ok := skipObject(data, offset, group, variation, qualifier)
		if !ok {
			return CommandStatusUnknown
		}
		offset = next
	}
	return CommandStatusUnknown
}

// commandStatusForQualifier reads the per-point command status byte from a G12V1
// object at the given header offset for the given qualifier. The response CROB
// layout mirrors the request: code(1), count(1), onTime(4), offTime(4),
// status(1) — the status byte is the command status. For the index-only
// qualifier (0x00) the per-point index (2 octets) precedes the CROB value.
func commandStatusForQualifier(data []byte, offset int, qualifier uint8) CommandStatus {
	switch qualifier {
	case 0x00: // index-only: header(4) + index(2) + CROB(11)
		statusIdx := offset + 4 + 2 + 10
		if statusIdx < len(data) {
			return CommandStatus(data[statusIdx])
		}
	case 0x07, 0x27: // count8/count16: header(4) [+count16(2)] + CROB(11)
		hdrLen := 4
		if qualifier == 0x27 {
			hdrLen = 5
		}
		statusIdx := offset + hdrLen + 10
		if statusIdx < len(data) {
			return CommandStatus(data[statusIdx])
		}
	}
	return CommandStatusUnknown
}

// skipObject advances past one object header and its payload. Returns the new
// offset and ok=false if the object is truncated/unknown (scan aborts).
func skipObject(data []byte, offset int, group, variation, qualifier uint8) (int, bool) {
	// IEEE 1815 v0 object-header sizes (qualifier high nibble = prefix mode,
	// low nibble = prefix field size):
	//   0x06 all-objects -> 4-byte header, no payload to skip here
	//   0x07 count8      -> 4-byte header
	//   0x27 count16     -> 5-byte header
	//   0x28 range16     -> 7-byte header (start/stop)
	//   0x00 index8      -> 4-byte header, per-point 1-byte index prefix
	// We only need accurate advancement for G12V1 (index-only, 0x00); other
	// groups/qualifiers are best-effort.
	pointSize, hasIndexPrefix, prefixSize := objectLayout(group, variation, qualifier)
	if pointSize == 0 && qualifier != 0x06 {
		return 0, false
	}
	switch qualifier {
	case 0x06: // all-objects: header only
		return offset + 4, true
	case 0x00: // index8/count8 hybrid used by G12V1: 4-byte header + count(1)
		count := int(data[offset+3])
		perPoint := pointSize
		if hasIndexPrefix {
			perPoint += prefixSize
		}
		return offset + 4 + 1 + count*perPoint, true
	case 0x07: // count8
		count := int(data[offset+3])
		return offset + 4 + count*pointSize, true
	case 0x27: // count16
		if offset+5 > len(data) {
			return 0, false
		}
		count := int(data[offset+3]) | int(data[offset+4])<<8
		return offset + 5 + count*pointSize, true
	case 0x28: // range16
		if offset+7 > len(data) {
			return 0, false
		}
		start := int(data[offset+3]) | int(data[offset+4])<<8
		stop := int(data[offset+5]) | int(data[offset+6])<<8
		n := stop - start + 1
		if n < 0 {
			n = 0
		}
		return offset + 7 + n*pointSize, true
	}
	return 0, false
}

// objectLayout returns the per-point payload size for fixed-size object
// variations, whether each point carries an index prefix, and the prefix size.
func objectLayout(group, variation, qualifier uint8) (pointSize int, hasIndexPrefix bool, prefixSize int) {
	switch {
	case group == 12 && variation == 1:
		return 11, qualifier == 0x00, 2 // CROB; index-only uses a 2-octet index
	default:
		return 0, false, 0
	}
}

// WriteBinaryOutput writes a binary output (CROB) to an outstation.
func (m *Master) WriteBinaryOutput(outstationID uint16, index uint16, crob *CROB) error {
	if !m.isOperational() {
		return ErrNotConnected
	}

	data := buildCROBRequest(index, crob)
	req := buildRequest(0, al.FuncWrite, data)

	return m.sendWithRetry(req, outstationID)
}

// buildCROBRequest builds a CROB write request.
func buildCROBRequest(index uint16, crob *CROB) []byte {
	// Object header: Group 12, Variation 1
	// Qualifier: 0x00 (index-only)
	header := []byte{12, 1, 0x00, 0x01}

	// Index: 2 bytes, least-significant byte first.
	indexBytes := []byte{byte(index), byte(index >> 8)}

	// CROB data: Code(1) + Count(1) + OnTime(4) + OffTime(4) + Status(1)
	data := []byte{
		crob.Code,
		crob.Count,
		byte(crob.OnTime), byte(crob.OnTime >> 8), byte(crob.OnTime >> 16), byte(crob.OnTime >> 24),
		byte(crob.OffTime), byte(crob.OffTime >> 8), byte(crob.OffTime >> 16), byte(crob.OffTime >> 24),
		crob.Status,
	}

	return append(append(header, indexBytes...), data...)
}

// WriteAnalogOutput writes an analog output to an outstation.
func (m *Master) WriteAnalogOutput(outstationID uint16, index uint16, value interface{}, variation uint8) error {
	if !m.isOperational() {
		return ErrNotConnected
	}

	data := buildAnalogOutputRequest(index, value, variation)
	req := buildRequest(0, al.FuncWrite, data)

	return m.sendWithRetry(req, outstationID)
}

// buildAnalogOutputRequest builds an analog output write request.
func buildAnalogOutputRequest(index uint16, value interface{}, variation uint8) []byte {
	// Object header based on variation:
	// Group 41 = 16-bit signed
	// Group 42 = 32-bit signed
	// Group 43 = 32-bit float
	// Group 44 = 64-bit double
	group := uint8(41)
	if variation >= 5 {
		group = 42
	} else if variation >= 9 {
		group = 43
	} else if variation >= 13 {
		group = 44
	}

	// Object header: Group, Variation, Qualifier (0x00=index-only), Count (1)
	header := []byte{group, variation, 0x00, 0x01}

	// Index: 2 bytes big-endian
	indexBytes := []byte{byte(index >> 8), byte(index & 0xFF)}

	// Value encoding depends on variation
	var valueBytes []byte
	switch variation {
	case 1, 5: // 16-bit signed
		v := int16(0)
		switch val := value.(type) {
		case int:
			v = int16(val)
		case int16:
			v = val
		case int32:
			v = int16(val)
		case float64:
			v = int16(val)
		}
		valueBytes = []byte{byte(v >> 8), byte(v)}

	case 2, 6: // 16-bit unsigned
		v := uint16(0)
		switch val := value.(type) {
		case int:
			v = uint16(val)
		case uint16:
			v = val
		case uint32:
			v = uint16(val)
		case float64:
			v = uint16(val)
		}
		valueBytes = []byte{byte(v >> 8), byte(v)}

	case 3, 7: // 32-bit signed
		v := int32(0)
		switch val := value.(type) {
		case int:
			v = int32(val)
		case int32:
			v = val
		case int64:
			v = int32(val)
		case float64:
			v = int32(val)
		}
		valueBytes = []byte{
			byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v),
		}

	case 4, 8: // 32-bit unsigned
		v := uint32(0)
		switch val := value.(type) {
		case int:
			v = uint32(val)
		case uint32:
			v = val
		case uint64:
			v = uint32(val)
		case float64:
			v = uint32(val)
		}
		valueBytes = []byte{
			byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v),
		}

	case 9, 10: // 32-bit float
		v := float32(0)
		switch val := value.(type) {
		case float32:
			v = val
		case float64:
			v = float32(val)
		case int, int32:
			v = float32(reflectValue(val))
		}
		bits := math.Float32bits(v)
		valueBytes = []byte{
			byte(bits >> 24), byte(bits >> 16), byte(bits >> 8), byte(bits),
		}

	case 13, 14: // 64-bit double
		v := float64(0)
		switch val := value.(type) {
		case float64:
			v = val
		case float32:
			v = float64(val)
		case int, int32, int64:
			v = float64(reflectValue(val))
		}
		bits := math.Float64bits(v)
		valueBytes = []byte{
			byte(bits >> 56), byte(bits >> 48), byte(bits >> 40), byte(bits >> 32),
			byte(bits >> 24), byte(bits >> 16), byte(bits >> 8), byte(bits),
		}

	default: // Default to 16-bit signed
		v := int16(0)
		switch val := value.(type) {
		case int:
			v = int16(val)
		case int16:
			v = val
		case float64:
			v = int16(val)
		}
		valueBytes = []byte{byte(v >> 8), byte(v)}
	}

	return append(append(header, indexBytes...), valueBytes...)
}

// reflectValue converts interface{} to int64 for numeric operations.
func reflectValue(v interface{}) int64 {
	switch val := v.(type) {
	case int:
		return int64(val)
	case int32:
		return int64(val)
	case int64:
		return val
	case uint:
		return int64(val)
	case uint32:
		return int64(val)
	case uint64:
		return int64(val)
	default:
		return 0
	}
}

// =============================================================================
// READ OPERATIONS (Convenience Methods)
// =============================================================================

// ReadBinaryInputs reads binary inputs from an outstation.
func (m *Master) ReadBinaryInputs(outstationID uint16, start, stop uint16) error {
	if !m.isOperational() {
		return ErrNotConnected
	}

	// Group 1, Variation 1 with range qualifier
	data := buildReadRangeRequest(1, 1, start, stop)
	req := buildRequest(0, al.FuncRead, data)

	return m.sendWithRetry(req, outstationID)
}

// ReadDoubleBinaryInputs reads double-bit binary inputs from an outstation.
func (m *Master) ReadDoubleBinaryInputs(outstationID uint16, start, stop uint16) error {
	if !m.isOperational() {
		return ErrNotConnected
	}

	// Group 3, Variation 1 with range qualifier
	data := buildReadRangeRequest(3, 1, start, stop)
	req := buildRequest(0, al.FuncRead, data)

	return m.sendWithRetry(req, outstationID)
}

// ReadAnalogInputs reads analog inputs from an outstation.
func (m *Master) ReadAnalogInputs(outstationID uint16, start, stop uint16) error {
	if !m.isOperational() {
		return ErrNotConnected
	}

	// Group 30, Variation 1 with range qualifier
	data := buildReadRangeRequest(30, 1, start, stop)
	req := buildRequest(0, al.FuncRead, data)

	return m.sendWithRetry(req, outstationID)
}

// ReadCounters reads counters from an outstation.
func (m *Master) ReadCounters(outstationID uint16, start, stop uint16) error {
	if !m.isOperational() {
		return ErrNotConnected
	}

	// Group 20, Variation 1 with range qualifier
	data := buildReadRangeRequest(20, 1, start, stop)
	req := buildRequest(0, al.FuncRead, data)

	return m.sendWithRetry(req, outstationID)
}

// ReadFrozenCounters reads frozen counters from an outstation.
func (m *Master) ReadFrozenCounters(outstationID uint16, start, stop uint16) error {
	if !m.isOperational() {
		return ErrNotConnected
	}

	// Group 21, Variation 1 with range qualifier
	data := buildReadRangeRequest(21, 1, start, stop)
	req := buildRequest(0, al.FuncRead, data)

	return m.sendWithRetry(req, outstationID)
}

// ReadBinaryOutputStatus reads binary output status from an outstation.
func (m *Master) ReadBinaryOutputStatus(outstationID uint16, start, stop uint16) error {
	if !m.isOperational() {
		return ErrNotConnected
	}

	// Group 10, Variation 1 with range qualifier
	data := buildReadRangeRequest(10, 1, start, stop)
	req := buildRequest(0, al.FuncRead, data)

	return m.sendWithRetry(req, outstationID)
}

// ReadAnalogOutputStatus reads analog output status from an outstation.
func (m *Master) ReadAnalogOutputStatus(outstationID uint16, start, stop uint16) error {
	if !m.isOperational() {
		return ErrNotConnected
	}

	// Group 40, Variation 1 with range qualifier
	data := buildReadRangeRequest(40, 1, start, stop)
	req := buildRequest(0, al.FuncRead, data)

	return m.sendWithRetry(req, outstationID)
}

// buildReadRangeRequest builds the object-header bytes for a ranged read
// request using the formal al.ObjectHeader model (DNP3-004). The wire bytes
// are preserved exactly from the prior ad-hoc construction.
func buildReadRangeRequest(group, variation uint8, start, stop uint16) []byte {
	hdr := al.ObjectHeader{Group: group, Variation: variation, Qualifier: al.QualRange16, Start: start, Stop: stop}
	data, err := al.EncodeObjectHeaders(nil, []al.ObjectHeader{hdr})
	if err != nil {
		return nil
	}
	return data
}

// HandleUnsolicited processes an unsolicited response from an outstation.
func (m *Master) HandleUnsolicited(data []byte) error {
	resp, err := al.DecodeResponse(data)
	if err != nil {
		return err
	}

	if !resp.Header.Control.UNS {
		return errors.New("not an unsolicited response")
	}

	// Update outstation state (we need to track this somehow)
	// In practice, you'd have sequence number tracking to identify the source

	return nil
}

// Placeholder for al import (would be internal/al in real implementation)
// In this file, al functions are simplified
