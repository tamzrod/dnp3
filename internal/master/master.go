// Package master implements the DNP3 Master Station role.
//
// The Master Station is the central control system that initiates all communication.
// It polls outstations for data, issues commands, and manages the SCADA system.
//
// Reference: IEEE 1815-2012 Section 8
package master

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"strings"
	"sync"
	"syscall"
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

	// DefaultConnectTimeout is the default whole-connect budget in
	// milliseconds (VEC-01 B4 F2): dial + link reset/status handshake +
	// initialization must complete within one finite wall-clock budget.
	DefaultConnectTimeout = 5000
	// DefaultOperateTimeout is the default whole-control budget in
	// milliseconds (VEC-01 B4 F2): the single non-idempotent DirectOperate
	// attempt (send + confirm/response wait + status interpretation) must
	// complete within one finite wall-clock budget.
	DefaultOperateTimeout = 5000
	// DefaultReadTimeout is the default whole-request budget in
	// milliseconds (VEC-01 B4 F2) applied to the public Read path: the
	// primary Class-0 read and any per-group fallback share this single
	// finite budget.
	DefaultReadTimeout = 10000
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
		StateDisconnected: true,
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
	// link. Toggles after each confirmed-data transaction; set to true by a
	// Reset Link Stations exchange, so the first confirmed frame sent after a
	// reset carries FCB=1 (IEEE 1815 / OpenDNP3 ResetReadFCB).
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

// resetFCB sets the primary-side FCB to its post-reset value (true), called
// after a Reset Link Stations exchange (DNP3-057). Per IEEE 1815 the first
// confirmed user-data frame sent after a reset carries FCB=1 (OpenDNP3
// ResetReadFCB sets nextReadFCB=true); subsequent frames toggle from there.
func (o *Outstation) resetFCB() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.fcb = true
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

	// errIncompleteResponse reports that the bytes received so far are a valid
	// prefix of a transport message that has not finished arriving. A solicited
	// response may span multiple transport fragments delivered across separate
	// Receive() calls, so the response wait keeps reading rather than treating a
	// partial message as a failure.
	errIncompleteResponse = errors.New("incomplete response")

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

	// ErrOperationDeadlineExceeded indicates an operation did not complete
	// within its single finite wall-clock budget (VEC-01 B4 F2). It is
	// distinct from ErrTimeout (a per-step receive timeout): it means the
	// whole operation's budget was exhausted, so no further retry or
	// fallback may extend it.
	ErrOperationDeadlineExceeded = errors.New("operation deadline exceeded")

	// ErrControlQuarantined indicates a new control was refused because a
	// previous control's outcome is UNKNOWN/indeterminate and the outstation
	// is quarantined (VEC-01 B4 F3). A non-idempotent control must not be
	// issued while a prior possibly-sent control's outcome is unresolved.
	// The quarantine is PERSISTENT: it survives Disconnect, reconnect, a failed
	// connect, late worker completion and request-marker cleanup, and is never
	// released automatically (VEC-01 B4 F2/F3 corrective addendum).
	ErrControlQuarantined = errors.New("control quarantined: prior control outcome unknown")

	// ErrControlInProgress indicates a new control was refused because a
	// previous control worker is still active for the outstation (VEC-01 B4
	// F3). Only one non-idempotent control may be in flight at a time; a
	// concurrent or later control is rejected rather than queued so it can
	// never race a possibly-outstanding control on the same link.
	ErrControlInProgress = errors.New("control already in progress for outstation")

	// ErrControlOutcomeUnknown indicates a control was sent (or a send was
	// attempted) but its outcome could not be positively established — e.g. the
	// response was missing, malformed, or arrived after a disconnect (VEC-01 B4
	// F3). It is distinct from a definite pre-send failure and from an explicit
	// remote rejection (which is a valid response carrying a failure status).
	ErrControlOutcomeUnknown = errors.New("control outcome unknown")

	// ErrSessionUnresolved indicates the outstation has an outstanding control
	// worker (e.g. blocked in an uncooperative Send) whose session was torn
	// down. A fresh Connect or control admission for that outstation is refused
	// fail-closed until the worker terminates, so a late send cannot contaminate
	// a replacement session (VEC-01 B4 F3 second corrective addendum).
	ErrSessionUnresolved = errors.New("session unresolved: prior control worker outstanding")
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
	// MEXT-025: a write/read that fails because the peer dropped the TCP
	// connection surfaces as a syscall errno (EPIPE on write after a
	// closed/receiver-gone peer, ECONNRESET/ECONNABORTED on a reset). These are
	// disconnects too — without recognizing them the master stays stuck in a
	// connected state after the peer is gone, so a subsequent Connect refuses
	// ("already connected") and the client can never recover. Match the
	// underlying errno so the public state flips to Disconnected (DNP3-031).
	var se syscall.Errno
	if errors.As(err, &se) {
		switch se {
		case syscall.ECONNRESET, syscall.ECONNABORTED, syscall.EPIPE, syscall.ETIMEDOUT:
			return true
		}
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
	// ConnectTimeout is the whole-connect budget in milliseconds (VEC-01 B4
	// F2): the dial plus link reset/status handshake plus initialization must
	// complete within this single finite wall-clock budget. When
	// non-positive, DefaultConnectTimeout is used.
	ConnectTimeout int
	// OperateTimeout is the whole-control budget in milliseconds (VEC-01 B4
	// F2) for the single non-idempotent DirectOperate attempt. When
	// non-positive, DefaultOperateTimeout is used.
	OperateTimeout int
	// ReadTimeout is the whole-request budget in milliseconds (VEC-01 B4 F2)
	// applied to the public Read path (primary Class-0 read plus any
	// per-group fallback sharing one budget). When non-positive,
	// DefaultReadTimeout is used.
	ReadTimeout int
}

// DefaultConfirmTimeout is the default application-layer confirm timeout in
// milliseconds (DNP3-066). It is intentionally shorter than DefaultTimeout:
// a confirm is a small dedicated reply, not a full object response.
const DefaultConfirmTimeout = 2000

// DefaultConfig returns default master configuration.
func DefaultConfig() *Config {
	return &Config{
		MasterAddress:  0xFFFF, // Master uses broadcast or specific address
		Timeout:        DefaultTimeout,
		MaxRetries:     MaxRetries,
		RetryDelay:     100,
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

	// Unsolicited delivery state. The sequence and callback are guarded by
	// unsolMu and must not call back into a method that takes reqMu.
	// The peer address is a parameter of each wait, so concurrent requests
	// do not share one slot.
	unsolMu       sync.Mutex
	unsolInit     bool
	unsolLast     uint8
	onUnsolicited func([]byte)

	// outstanding tracks per-outstation in-flight requests so a concurrent
	// request to the SAME outstation is rejected with ErrRequestOutstanding
	// instead of blocking behind reqMu indefinitely (DNP3-040). Different
	// outstations still queue via reqMu. Guarded by outstandingMu.
	outstanding   map[uint16]struct{}
	outstandingMu sync.Mutex

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

	// connectTimeout is the whole-connect budget (VEC-01 B4 F2): dial plus
	// link reset/status handshake plus initialization share one finite
	// wall-clock budget, resolved from Config.ConnectTimeout.
	connectTimeout time.Duration
	// operateTimeout is the whole-control budget (VEC-01 B4 F2) for the single
	// non-idempotent DirectOperate attempt, resolved from Config.OperateTimeout.
	operateTimeout time.Duration
	// readTimeout is the whole-request budget (VEC-01 B4 F2) applied to the
	// public Read path, resolved from Config.ReadTimeout.
	readTimeout time.Duration

	// safetyMu guards the per-outstation control-safety state (VEC-01 B4
	// F2/F3 second corrective addendum): quarantine, control-worker gates,
	// the session generation and the unresolved-worker fence. A single lock
	// makes UNKNOWN/quarantine state, worker admission and the session-validity
	// check ONE atomic decision, so a quarantine established by an
	// ambiguous/cancelled control cannot race with a fresh admission and a
	// worker admitted in one session cannot transmit after that session's
	// teardown has linearized. safetyMu is a leaf lock: it is never held across
	// a possibly-blocking transport Send, nor while acquiring
	// reqMu/outstandingMu/mu, so admission cannot deadlock the request path.
	safetyMu sync.Mutex
	// quarantine tracks per-outstation control-safety quarantine state. An
	// outstation is quarantined when a control's outcome is
	// UNKNOWN/indeterminate (cancellation, deadline, disconnect after a
	// possible send, or a malformed/missing response). While quarantined, no
	// new control may be issued and the quarantine is PERSISTENT: it is never
	// released automatically by Disconnect, reconnect, a failed connect, late
	// worker completion or request-marker cleanup. Without a separately
	// owner-approved reconciliation protocol an UNKNOWN outcome stays
	// quarantined indefinitely.
	quarantine map[uint16]bool
	// controlGates tracks per-outstation control-worker ownership. A
	// non-idempotent control is admitted only when no worker is active for the
	// outstation; the worker records the session generation it began in so a
	// stale worker from a torn-down session cannot send through a reconnected
	// one. A gate is removed from the map once it is fully idle (no active
	// worker, no in-flight send) so the map cannot grow unbounded.
	controlGates map[uint16]*controlGate
	// sessionGen is bumped on every Disconnect so a control worker started in a
	// previous session can detect that its session is gone and refuse to send.
	sessionGen uint64
	// sessionUnresolved records outstation IDs for which a control worker
	// remains outstanding (e.g. blocked in an uncooperative Send) after its
	// session was torn down. Until that worker terminates, a fresh
	// Connect/admission for the outstation is refused fail-closed so a late
	// send cannot contaminate a replacement session. Guarded by safetyMu.
	sessionUnresolved map[uint16]bool
}

// controlGate records the control-worker ownership of a single outstation.
// active is true while a control worker holds the gate; gen is the session
// generation in which that worker began; sending is true while an admitted
// transmission is in flight (so teardown can fence a possible send and the
// gate is retained until the send completes). Guarded by safetyMu.
type controlGate struct {
	active  bool
	sending bool
	gen     uint64
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
		config:            config,
		state:             StateDisconnected,
		outstations:       make(map[uint16]*Outstation),
		fragmenter:        tl.NewFragmenter(),
		reassembler:       tl.NewReassembler(),
		retryPolicy:       DefaultRetryPolicy(config),
		outstanding:       make(map[uint16]struct{}),
		idleTimeout:       time.Duration(config.IdleTimeout) * time.Millisecond,
		confirmTimeout:    confirmTimeoutFromConfig(config),
		connectTimeout:    connectTimeoutFromConfig(config),
		operateTimeout:    operateTimeoutFromConfig(config),
		readTimeout:       readTimeoutFromConfig(config),
		quarantine:        make(map[uint16]bool),
		controlGates:      make(map[uint16]*controlGate),
		sessionUnresolved: make(map[uint16]bool),
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

// connectTimeoutFromConfig resolves the whole-connect budget (VEC-01 B4 F2):
// Config.ConnectTimeout (ms) when positive, else DefaultConnectTimeout.
func connectTimeoutFromConfig(config *Config) time.Duration {
	if config != nil && config.ConnectTimeout > 0 {
		return time.Duration(config.ConnectTimeout) * time.Millisecond
	}
	return time.Duration(DefaultConnectTimeout) * time.Millisecond
}

// operateTimeoutFromConfig resolves the whole-control budget (VEC-01 B4 F2):
// Config.OperateTimeout (ms) when positive, else DefaultOperateTimeout.
func operateTimeoutFromConfig(config *Config) time.Duration {
	if config != nil && config.OperateTimeout > 0 {
		return time.Duration(config.OperateTimeout) * time.Millisecond
	}
	return time.Duration(DefaultOperateTimeout) * time.Millisecond
}

// readTimeoutFromConfig resolves the whole-request budget (VEC-01 B4 F2):
// Config.ReadTimeout (ms) when positive, else DefaultReadTimeout.
func readTimeoutFromConfig(config *Config) time.Duration {
	if config != nil && config.ReadTimeout > 0 {
		return time.Duration(config.ReadTimeout) * time.Millisecond
	}
	return time.Duration(DefaultReadTimeout) * time.Millisecond
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

// quarantineOutstationLocked marks outstationID as control-quarantined under
// safetyMu. The caller must hold safetyMu.
func (m *Master) quarantineOutstationLocked(outstationID uint16) {
	m.quarantine[outstationID] = true
}

// quarantineOutstation marks outstationID as control-quarantined because a
// control's outcome is UNKNOWN/indeterminate (VEC-01 B4 F3).
func (m *Master) quarantineOutstation(outstationID uint16) {
	m.safetyMu.Lock()
	m.quarantineOutstationLocked(outstationID)
	m.safetyMu.Unlock()
}

// clearQuarantine releases the control-safety quarantine for outstationID. It
// is intentionally NOT called from Disconnect or any automatic teardown
// (VEC-01 B4 F2/F3 corrective addendum): releasing an UNKNOWN outcome without a
// separately owner-approved reconciliation protocol would silently discard a
// possibly-executed control. It exists only for a future explicitly approved
// reconciliation path and for tests that need a clean slate.
func (m *Master) clearQuarantine(outstationID uint16) {
	m.safetyMu.Lock()
	delete(m.quarantine, outstationID)
	m.safetyMu.Unlock()
}

// ClearQuarantine releases the control-safety quarantine for outstationID. It
// is exposed so a future, separately owner-approved explicit reconciliation
// protocol can release a quarantine. It is deliberately NOT wired to
// Disconnect/reconnect: an UNKNOWN outcome persists until such a protocol
// exists.
func (m *Master) ClearQuarantine(outstationID uint16) {
	m.clearQuarantine(outstationID)
}

// QuarantineOutstation marks outstationID control-quarantined (VEC-01 B4 F3).
// Exposed so a supervising caller (the public client) can fail closed when it
// observes cancellation of an in-flight non-idempotent control whose outcome it
// cannot positively establish. The quarantine is persistent: it survives
// Disconnect, reconnect, late worker completion and request-marker cleanup.
func (m *Master) QuarantineOutstation(outstationID uint16) {
	m.quarantineOutstation(outstationID)
}

// IsQuarantined reports whether outstationID is currently control-quarantined
// (VEC-01 B4 F3). Exposed for tests and diagnostics.
func (m *Master) IsQuarantined(outstationID uint16) bool {
	m.safetyMu.Lock()
	defer m.safetyMu.Unlock()
	return m.quarantine[outstationID]
}

// acquireControlGate admits a single non-idempotent control for outstationID
// (VEC-01 B4 F3 second corrective addendum). Under one safetyMu critical
// section it atomically (a) refuses admission when the outstation is
// quarantined (UNKNOWN outcome outstanding), when another worker holds the gate,
// or when the outstation has an unresolved worker from a torn-down session, and
// (b) records the current session generation so a later session change can be
// detected. It returns the generation plus a release function the worker must
// call when it terminates. The release function clears the active flag but
// NEVER releases a quarantine — an UNKNOWN outcome persists independently.
func (m *Master) acquireControlGate(outstationID uint16) (uint64, func(), error) {
	m.safetyMu.Lock()
	defer m.safetyMu.Unlock()
	if m.quarantine[outstationID] {
		return 0, nil, fmt.Errorf("%w: 0x%04X", ErrControlQuarantined, outstationID)
	}
	if m.sessionUnresolved[outstationID] {
		return 0, nil, fmt.Errorf("%w: 0x%04X", ErrSessionUnresolved, outstationID)
	}
	g, ok := m.controlGates[outstationID]
	if !ok {
		g = &controlGate{}
		m.controlGates[outstationID] = g
	}
	if g.active {
		return 0, nil, fmt.Errorf("%w: 0x%04X", ErrControlInProgress, outstationID)
	}
	g.active = true
	g.gen = m.sessionGen
	gen := g.gen
	release := func() {
		m.safetyMu.Lock()
		if cg, ok := m.controlGates[outstationID]; ok {
			cg.active = false
			if !cg.sending {
				delete(m.controlGates, outstationID)
			}
		}
		// The worker has terminated, so it is no longer an outstanding worker
		// for a torn-down session: lift the unresolved fence (but never a
		// quarantine).
		delete(m.sessionUnresolved, outstationID)
		m.safetyMu.Unlock()
	}
	return gen, release, nil
}

// beginTransmit is the single linearization point for "this worker may
// transmit through session gen" (VEC-01 B4 F3-R2A). Under safetyMu it verifies
// that the outstation is not quarantined and that the session is still current,
// then marks the gate as sending. The check and the marking are atomic with
// respect to Disconnect (which bumps the session generation under the same
// lock), so a worker cannot START a transmission after its session's teardown
// has linearized. A nil error means the send is admitted; callers must pair it
// with endTransmit. A non-nil error is a definite NOT_SENT refusal.
func (m *Master) beginTransmit(gen uint64, outstationID uint16) error {
	m.safetyMu.Lock()
	defer m.safetyMu.Unlock()
	if m.quarantine[outstationID] {
		return fmt.Errorf("%w: 0x%04X", ErrControlQuarantined, outstationID)
	}
	if gen != m.sessionGen {
		return fmt.Errorf("%w: session changed before control send", ErrOperationDeadlineExceeded)
	}
	g, ok := m.controlGates[outstationID]
	if !ok {
		g = &controlGate{}
		m.controlGates[outstationID] = g
	}
	g.sending = true
	return nil
}

// endTransmit clears the in-flight send marker under safetyMu and drops the
// gate if the worker has also released it (so the gate map does not grow
// unbounded). Safe to call only after a beginTransmit that returned true.
func (m *Master) endTransmit(outstationID uint16) {
	m.safetyMu.Lock()
	if g, ok := m.controlGates[outstationID]; ok {
		g.sending = false
		if !g.active {
			delete(m.controlGates, outstationID)
		}
	}
	m.safetyMu.Unlock()
}

// anySessionUnresolved reports whether any outstation has an unresolved control
// worker from a torn-down session (VEC-01 B4 F3-R2B). A fresh Connect is
// refused while this is true.
func (m *Master) anySessionUnresolved() bool {
	m.safetyMu.Lock()
	defer m.safetyMu.Unlock()
	return len(m.sessionUnresolved) > 0
}

// HasUnresolvedSession reports whether any outstation still has an unresolved
// control worker from a torn-down session (VEC-01 B4 F3-R3). It is exposed so
// the public client can fail a fresh Connect fast, before dialing, while an old
// control send could still affect the shared transport. The internal master
// enforces the same fence again at ConnectContext; this is the pre-dial check.
func (m *Master) HasUnresolvedSession() bool {
	return m.anySessionUnresolved()
}

// sessionUnresolvedFor reports whether outstationID has an unresolved control
// worker from a torn-down session (VEC-01 B4 F3-R2B). Exposed for
// tests/diagnostics.
func (m *Master) sessionUnresolvedFor(outstationID uint16) bool {
	m.safetyMu.Lock()
	defer m.safetyMu.Unlock()
	return m.sessionUnresolved[outstationID]
}

// HasActiveControl reports whether a control worker currently holds the gate
// for outstationID (VEC-01 B4 F3). Exposed for tests and diagnostics so a
// supervising caller can observe that a worker is still active (e.g. blocked in
// an uncooperative transport) and that the fail-closed gate remains held.
func (m *Master) HasActiveControl(outstationID uint16) bool {
	m.safetyMu.Lock()
	defer m.safetyMu.Unlock()
	g, ok := m.controlGates[outstationID]
	return ok && (g.active || g.sending)
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
	if err := m.performLinkHandshake(time.Time{}); err != nil {
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

// ConnectContext performs the connect sequence under a single finite
// wall-clock budget (VEC-01 B4 F2). The dial, link reset/status handshake and
// initialization all share this one budget; the caller's context, when it
// carries an earlier deadline, tightens the effective budget but never
// extends it. A nil context is treated as context.Background(). If the budget
// is exhausted the connect fails closed with ErrOperationDeadlineExceeded and
// the master is left in StateError (no partial session is reported active).
func (m *Master) ConnectContext(ctx context.Context, budget time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if budget <= 0 {
		budget = m.connectTimeout
	}
	// VEC-01 B4 F3-R2B: refuse a fresh connect while a control worker from a
	// torn-down session remains outstanding for any registered outstation. A
	// late send from that worker must not be able to contaminate the new
	// session; the caller must wait for it to terminate (or restart) first.
	if m.anySessionUnresolved() {
		return fmt.Errorf("%w: connect refused while a prior control worker is outstanding", ErrSessionUnresolved)
	}
	connectDeadline := time.Now().Add(budget)
	effective := connectDeadline
	if dl, ok := ctx.Deadline(); ok && dl.Before(effective) {
		effective = dl
	}

	if m.transport == nil {
		return errors.New("transport not configured")
	}

	// DNP3-039: Connect may only begin from Disconnected or Error (reconnect).
	cur := m.State()
	if cur != StateDisconnected && cur != StateError {
		return fmt.Errorf("%w: %s -> Connecting", ErrIllegalStateTransition, cur)
	}
	if err := m.transitionTo(StateConnecting); err != nil {
		return err
	}

	m.resetForReconnect()

	if err := ctx.Err(); err != nil {
		_ = m.transitionTo(StateError)
		return fmt.Errorf("%w: %v", ErrOperationDeadlineExceeded, err)
	}
	if m.deadlineExpired(effective) {
		_ = m.transitionTo(StateError)
		return ErrOperationDeadlineExceeded
	}

	if err := m.performLinkHandshake(effective); err != nil {
		_ = m.transitionTo(StateError)
		if m.deadlineExpired(effective) {
			return ErrOperationDeadlineExceeded
		}
		return fmt.Errorf("link handshake failed: %w", err)
	}

	if m.deadlineExpired(effective) {
		_ = m.transitionTo(StateError)
		return ErrOperationDeadlineExceeded
	}

	if err := m.transitionTo(StateConnected); err != nil {
		return err
	}
	if err := m.transitionTo(StateActive); err != nil {
		return err
	}
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

// maxHandshakePeerFrames bounds how many outstation-primary user-data frames
// are set aside while one secondary handshake response is outstanding.
// OpenDNP3 3.1.2 transmits a null unsolicited response as soon as the TCP
// session is up, and that primary frame can arrive before the secondary ACK.
const maxHandshakePeerFrames = 4

// maxResponseBytes bounds the reassembly of one solicited response. The
// transport fragments a single application message into 249-byte payloads, and
// the APDU cap for this stack is far below this bound; the limit only stops an
// endless stream of fragments from growing the buffer without limit.
const maxResponseBytes = 64 * 1024

// peerPrimaryUserData reports a primary user-data frame addressed from the
// outstation to this master. It is not a secondary handshake response.
// A mis-addressed frame is not matched, so it still fails the handshake.
func peerPrimaryUserData(f *frame.Frame, outstationID, masterAddr uint16) bool {
	if f == nil || f.Control.DIR || !f.Control.PRM {
		return false
	}
	if f.SrcAddr != outstationID || f.DestAddr != masterAddr {
		return false
	}
	switch f.Control.FuncCode {
	case frame.FuncUnconfirmedUserData, frame.FuncConfirmedUserData:
		return true
	default:
		return false
	}
}

// awaitSecondaryHandshake reads until the next frame is not an addressed
// outstation-primary user-data frame. Those frames are copied into pending
// and do not consume a Reset Link retry. The returned buffer is validated
// by the caller as the secondary ACK or Link Status response.
func (m *Master) awaitSecondaryHandshake(deadline time.Time, outstationID uint16, pending *[][]byte) ([]byte, error) {
	nominal := time.Duration(m.config.Timeout) * time.Millisecond
	skipped := 0
	for {
		wait := constrainTimeout(m, nominal, deadline)
		m.transport.SetTimeout(int(wait / time.Millisecond))
		t0 := time.Now()
		raw, err := m.transport.Receive()
		if err != nil {
			if m.waitHitBudget(deadline, nominal, wait, t0) {
				return nil, ErrOperationDeadlineExceeded
			}
			return nil, err
		}
		f, derr := frame.Decode(raw)
		if derr == nil && peerPrimaryUserData(f, outstationID, m.config.MasterAddress) {
			skipped++
			if skipped > maxHandshakePeerFrames {
				return nil, fmt.Errorf("ignored %d outstation primary frames waiting for a secondary handshake response", skipped)
			}
			if pending != nil {
				*pending = append(*pending, append([]byte(nil), raw...))
			}
			if m.deadlineExpired(deadline) {
				return nil, ErrOperationDeadlineExceeded
			}
			continue
		}
		return raw, nil
	}
}

// confirmHandshakeUnsolicited confirms outstation-primary fragments that
// arrived before the link handshake finished. OpenDNP3 will not answer a
// class read while a null unsolicited response is still awaiting confirm.
func (m *Master) confirmHandshakeUnsolicited(pending [][]byte) {
	for _, raw := range pending {
		f, err := frame.Decode(raw)
		if err != nil {
			continue
		}
		app, err := m.processReceivedBytes(raw)
		if err != nil {
			continue
		}
		m.consumeUnsolicited(app, f.SrcAddr)
	}
}

// performLinkHandshake performs the DNP3 link-layer handshake sequence.
// This establishes the data link connection before application layer communication.
// The sequence is:
//  1. Reset Link Stations → secondary ACK
//  2. Request Link Status → secondary Link Status response
//
// Connect succeeds only after both exchanges validate (DNP3-006/DNP3-007).
// Outstation-primary user data seen while waiting is confirmed afterward.
func (m *Master) performLinkHandshake(deadline time.Time) error {
	var pending [][]byte
	// Step 1: Send Reset Link Stations and validate the secondary ACK.
	if err := m.sendResetLinkDeadline(deadline, &pending); err != nil {
		return fmt.Errorf("reset link failed: %w", err)
	}

	// Step 2: Request Link Status and validate the secondary Link Status
	// response. This confirms the link is fully established in both
	// directions before application communication begins.
	if err := m.sendLinkStatusRequestDeadline(deadline, &pending); err != nil {
		return fmt.Errorf("link status request failed: %w", err)
	}
	m.confirmHandshakeUnsolicited(pending)
	return nil
}

// sendResetLink sends a Reset Link Stations frame to initialize the link layer.
//
// DNP3-058: a secondary NACK in response to Reset Link Stations is treated as
// a retryable failure (the link is alive but rejected the frame). The
// handshake is retried up to the NACK retry budget from the master's
// RetryPolicy; non-NACK failures fail immediately.
func (m *Master) sendResetLink() error {
	return m.sendResetLinkDeadline(time.Time{}, nil)
}

// sendResetLinkDeadline issues Reset Link Stations with each blocking receive
// capped to the operation's remaining wall-clock budget (VEC-01 B4 F2). A zero
// deadline means no cap (the legacy per-step timeout applies unchanged).
func (m *Master) sendResetLinkDeadline(deadline time.Time, pending *[][]byte) error {
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
			// layer. A primary user-data frame that arrives first is not that
			// ACK; keep reading until the secondary frame or the skip bound.
			raw, err := m.awaitSecondaryHandshake(deadline, o.ID, pending)
			if err != nil {
				if errors.Is(err, ErrOperationDeadlineExceeded) {
					return ErrOperationDeadlineExceeded
				}
				return fmt.Errorf("failed to receive reset acknowledgment: %w", err)
			}
			if err := validateResetLinkACK(raw, o.ID, m.config.MasterAddress); err != nil {
				lastErr = fmt.Errorf("invalid reset link ACK: %w", err)
				// DNP3-058: only a link-layer NACK is retryable for the handshake.
				if !m.retryAgain(policy, &lastErr, attempt) {
					return lastErr
				}
				if m.deadlineExpired(deadline) {
					return ErrOperationDeadlineExceeded
				}
				continue
			}
			// DNP3-057: a Reset Link Stations exchange resets the primary-side FCB
			// to its post-reset value (true) for this link.
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
	return m.sendLinkStatusRequestDeadline(time.Time{}, nil)
}

// sendLinkStatusRequestDeadline issues Request Link Status with each blocking
// receive capped to the operation's remaining wall-clock budget (VEC-01 B4 F2).
// A zero deadline means no cap (the legacy per-step timeout applies unchanged).
func (m *Master) sendLinkStatusRequestDeadline(deadline time.Time, pending *[][]byte) error {
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

		raw, err := m.awaitSecondaryHandshake(deadline, o.ID, pending)
		if err != nil {
			if errors.Is(err, ErrOperationDeadlineExceeded) {
				return ErrOperationDeadlineExceeded
			}
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
	// Registered outstations are dropped on teardown; assign under mu so a
	// concurrent teardown (e.g. a bounded connect's cleanup racing the caller)
	// cannot race this map write.
	m.outstations = make(map[uint16]*Outstation)
	m.mu.Unlock()
	// VEC-01 B4 F3-R2A/R2B: bump the session generation under safetyMu so the
	// final session-validity check and admission to transmit (beginTransmit)
	// share one linearization point with this teardown. Any worker still active
	// or with a send in flight for an outstation is fenced as unresolved: a
	// fresh Connect/admission for that outstation is refused fail-closed until
	// the worker terminates, so a late send cannot contaminate a replacement
	// session. The control-safety quarantine is deliberately NOT cleared here:
	// an UNKNOWN outcome of a possibly-executed control persists across
	// Disconnect, reconnect and late worker completion. Releasing it requires a
	// separately owner-approved explicit reconciliation protocol.
	m.safetyMu.Lock()
	m.sessionGen++
	for id, g := range m.controlGates {
		if g.active || g.sending {
			// Fence any outstanding worker so it cannot transmit into a
			// replacement session; quarantine only when a send was actually in
			// flight (an idle, not-yet-sent worker is a definite NOT_SENT).
			m.sessionUnresolved[id] = true
			if g.sending {
				m.quarantine[id] = true
			}
		}
	}
	m.safetyMu.Unlock()
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
	_, err := m.OperateWithStatus(outstationID, selectThenOperate, group, variation, index, value)
	return err
}

// OperateWithStatus performs a control operation and returns the per-point
// command status parsed from the outstation's response (DNP3-020). Unlike
// Operate, it reads the response and never reports success when the response
// status byte indicates a failure. A response with no parseable G12V1 status
// yields CommandStatusUnknown (not success).
func (m *Master) OperateWithStatus(outstationID uint16, selectThenOperate bool, group, variation uint8, index uint16, value interface{}) (CommandStatus, error) {
	// VEC-01 B4 F1/F2/F3: the legacy entry point now performs a single-shot,
	// deadline-bounded control with fail-closed cancellation semantics. The
	// public client drives the cancelable OperateWithStatusContext variant; this
	// wrapper uses the master's configured whole-control budget and a background
	// context so direct internal callers get the same safety behavior.
	return m.OperateWithStatusContext(context.Background(), outstationID, selectThenOperate, group, variation, index, value, time.Now().Add(m.operateTimeout))
}

// OperateWithStatusContext performs exactly one logical G12V1 DirectOperate
// request under a single finite wall-clock budget (VEC-01 B4 F1/F2/F3):
//
//   - F1: the control is sent AT MOST ONCE. There is no hidden retry on
//     timeout, NACK, CRC, invalid response, missing confirmation or disconnect.
//     A non-idempotent control must never be resent autonomously.
//   - F2: send + confirm/response wait + status interpretation share the one
//     budget in deadline; the caller's ctx deadline, when earlier, tightens it.
//   - F3: when cancellation, the deadline, a disconnect after a possible send,
//     or a missing/malformed response leaves the outcome UNKNOWN, the
//     outstation is quarantined (no new control may be issued until explicit
//     teardown) and ErrControlOutcomeUnknown is returned. A valid response
//     carrying an explicit failure status is NOT unknown and is returned as-is.
//
// An already-cancelled context returns without sending anything (definitely
// NOT_SENT). A failure positively established before any send also does not
// quarantine. Returns (status, nil) only for a positively classified outcome.
func (m *Master) OperateWithStatusContext(ctx context.Context, outstationID uint16, selectThenOperate bool, group, variation uint8, index uint16, value interface{}, deadline time.Time) (CommandStatus, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if dl, ok := ctx.Deadline(); ok && (deadline.IsZero() || dl.Before(deadline)) {
		deadline = dl
	}
	if deadline.IsZero() {
		deadline = time.Now().Add(m.operateTimeout)
	}

	if !m.isOperational() {
		return CommandStatusUnknown, ErrNotConnected
	}

	// F3: register the control worker BEFORE any dispatch. Admission fails
	// closed if a prior control's outcome is UNKNOWN (quarantine) or if another
	// worker already holds the gate. The release returned here clears the gate
	// when this call returns; it never releases a quarantine.
	gen, release, err := m.acquireControlGate(outstationID)
	if err != nil {
		return CommandStatusUnknown, err
	}
	defer release()

	return m.operateGated(ctx, gen, outstationID, selectThenOperate, group, variation, index, value, deadline)
}

// operateGated performs the single-shot control while the caller holds the
// control gate for outstationID in session generation gen (VEC-01 B4 F3). The
// gate must already be held; the session generation is threaded to the
// transmission point so a worker from a torn-down session cannot send.
func (m *Master) operateGated(ctx context.Context, gen uint64, outstationID uint16, selectThenOperate bool, group, variation uint8, index uint16, value interface{}, deadline time.Time) (CommandStatus, error) {
	// Build the request(s) up front. A build failure is a definite pre-send
	// failure (NOT_SENT): no control was transmitted and no quarantine applies.
	var selectReq, operateReq *al.APDU
	if selectThenOperate {
		selectReq = m.buildControlRequest(al.FuncSelect, group, variation, index, value)
		if selectReq == nil {
			return CommandStatusUnknown, fmt.Errorf("unsupported control value for group %d variation %d", group, variation)
		}
		operateReq = m.buildControlRequest(al.FuncOperate, group, variation, index, value)
	} else {
		operateReq = m.buildControlRequest(al.FuncDirectOperate, group, variation, index, value)
	}
	if operateReq == nil {
		return CommandStatusUnknown, fmt.Errorf("unsupported control value for group %d variation %d", group, variation)
	}

	if err := ctx.Err(); err != nil {
		return CommandStatusUnknown, fmt.Errorf("%w: %w", ErrOperationDeadlineExceeded, err)
	}

	if selectThenOperate {
		// Each stage is itself single-shot and shares the same budget.
		if _, err := m.sendControlOnce(ctx, gen, selectReq, outstationID, deadline); err != nil {
			return CommandStatusUnknown, err
		}
	}
	return m.sendControlOnce(ctx, gen, operateReq, outstationID, deadline)
}

// AcquireControlGate admits a single non-idempotent control for outstationID
// (VEC-01 B4 F3). It is exposed so the public client can register the control
// worker before dispatch. It fails closed with ErrControlQuarantined when the
// outstation's prior control outcome is UNKNOWN, and with ErrControlInProgress
// when a worker already holds the gate. On success the caller must invoke the
// returned release exactly once when the worker terminates.
func (m *Master) AcquireControlGate(outstationID uint16) (gen uint64, release func(), err error) {
	return m.acquireControlGate(outstationID)
}

// OperateWithStatusGated performs a single-shot control while the caller holds
// the control gate acquired via AcquireControlGate for session generation gen
// (VEC-01 B4 F3). The public client uses this to run the control worker on the
// caller's context with a single deadline, so cancellation and session
// replacement are observed at the transmission point.
func (m *Master) OperateWithStatusGated(ctx context.Context, gen uint64, outstationID uint16, selectThenOperate bool, group, variation uint8, index uint16, value interface{}, deadline time.Time) (CommandStatus, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if dl, ok := ctx.Deadline(); ok && (deadline.IsZero() || dl.Before(deadline)) {
		deadline = dl
	}
	if deadline.IsZero() {
		deadline = time.Now().Add(m.operateTimeout)
	}
	if !m.isOperational() {
		return CommandStatusUnknown, ErrNotConnected
	}
	return m.operateGated(ctx, gen, outstationID, selectThenOperate, group, variation, index, value, deadline)
}

// OperateBudget returns the configured whole-control budget (VEC-01 B4 F2) so
// the public client can derive one absolute deadline for a single control.
func (m *Master) OperateBudget() time.Duration {
	return m.operateTimeout
}

// sendControlOnce transmits one control request AT MOST ONCE and classifies its
// outcome (VEC-01 B4 F1/F3). ctx cancellation observed after a send is treated
// as an UNKNOWN outcome (quarantine + fail closed); cancellation observed before
// the send is a definite NOT_SENT failure with no quarantine. It reads the response and returns the per-point
// command status. Any ambiguous outcome quarantines the outstation and returns
// ErrControlOutcomeUnknown wrapped with the underlying cause; a definite
// pre-send failure returns without quarantining; a valid response (success or
// explicit failure status) is returned as classified and never quarantines.
//
// gen is the session generation in which the owning worker was admitted; a
// mismatch immediately before transmission means the session was torn down
// and/or replaced, so this worker must NOT send (definite NOT_SENT).
func (m *Master) sendControlOnce(ctx context.Context, gen uint64, req *al.APDU, outstationID uint16, deadline time.Time) (CommandStatus, error) {
	// Fail closed before any send if the outstation is already quarantined
	// (VEC-01 B4 F3): a prior control's outcome is unresolved.
	if m.IsQuarantined(outstationID) {
		return CommandStatusUnknown, fmt.Errorf("%w: 0x%04X", ErrControlQuarantined, outstationID)
	}

	// DNP3-040: reject a concurrent request to the SAME outstation.
	if err := m.beginRequest(outstationID); err != nil {
		return CommandStatusUnknown, err
	}
	defer m.endRequest(outstationID)

	m.reqMu.Lock()
	defer m.reqMu.Unlock()

	// F3: refuse to transmit if the caller has already cancelled. This is a
	// definite pre-send (NOT_SENT) condition: nothing has been transmitted.
	if err := ctx.Err(); err != nil {
		return CommandStatusUnknown, fmt.Errorf("%w: %w", ErrOperationDeadlineExceeded, err)
	}
	if m.deadlineExpired(deadline) {
		// Definite pre-send failure: nothing was transmitted.
		return CommandStatusUnknown, ErrOperationDeadlineExceeded
	}

	// F3-R2A: the final session-validity decision and admission to transmit is
	// ONE linearization point shared with Disconnect (beginTransmit). A worker
	// whose session was torn down, or an outstation now quarantined, is refused
	// here BEFORE any send. The in-flight marker stays set for the whole
	// single-shot transmission so teardown can fence a possible send.
	if err := m.beginTransmit(gen, outstationID); err != nil {
		return CommandStatusUnknown, err
	}
	defer m.endTransmit(outstationID)

	seq := m.nextSequence(outstationID)
	req.Control.Seq = seq
	data := req.Encode()
	fragments := m.fragmenter.Fragmentize(data)
	fcb := m.outstationFCB(outstationID)

	// Send the request exactly once (all fragments of the single APDU).
	for _, frag := range fragments {
		tlEncoded := tl.EncodeFragment(frag)
		dllFrame := &frame.Frame{
			Control: frame.Control{
				DIR:      true,
				PRM:      true,
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
			// Encode failure is pre-send: nothing was transmitted.
			return CommandStatusUnknown, err
		}
		if err := m.transport.Send(dllEncoded); err != nil {
			// A send failure means a send was attempted: the outcome is
			// indeterminate. Fail closed and quarantine.
			wrapped, _ := m.markDisconnected(err)
			m.quarantineOutstation(outstationID)
			m.diag(DiagError, "control", seq, "control send failed; outcome unknown", err)
			return CommandStatusUnknown, fmt.Errorf("%w: %w", ErrControlOutcomeUnknown, wrapped)
		}
	}
	m.advanceSequence(outstationID)
	m.recordActivity()
	m.diag(DiagInfo, "control", seq, fmt.Sprintf("sent single control to outstation %d", outstationID), nil)

	// Optional application confirmation (single wait, single attempt).
	if req.Control.CON {
		if err := m.waitForConfirmationTimeout(constrainTimeout(m, m.confirmTimeout, deadline), seq); err != nil {
			wrapped, _ := m.markDisconnected(err)
			m.quarantineOutstation(outstationID)
			m.diag(DiagError, "control", seq, "control confirmation failed; outcome unknown", err)
			return CommandStatusUnknown, fmt.Errorf("%w: %w", ErrControlOutcomeUnknown, wrapped)
		}
	}

	// Wait for the response (single wait, single attempt).
	resp, err := m.waitForResponseTimeout(constrainTimeout(m, time.Duration(m.config.Timeout)*time.Millisecond, deadline), outstationID)
	if err != nil {
		wrapped, _ := m.markDisconnected(err)
		m.quarantineOutstation(outstationID)
		m.diag(DiagError, "control", seq, "control response missing; outcome unknown", err)
		return CommandStatusUnknown, fmt.Errorf("%w: %w", ErrControlOutcomeUnknown, wrapped)
	}

	// Process the response (single attempt). A malformed or mismatched response
	// leaves the outcome indeterminate.
	appData, err := m.processResponse(resp, outstationID, seq)
	if err != nil {
		m.quarantineOutstation(outstationID)
		m.diag(DiagError, "control", seq, "control response malformed; outcome unknown", err)
		return CommandStatusUnknown, fmt.Errorf("%w: %w", ErrControlOutcomeUnknown, err)
	}

	m.recordActivity()
	m.advanceOutstationFCB(outstationID)

	apdu, err := al.DecodeResponse(appData)
	if err != nil {
		m.quarantineOutstation(outstationID)
		return CommandStatusUnknown, fmt.Errorf("%w: %w", ErrControlOutcomeUnknown, err)
	}
	status := resolveOperateStatus(req, apdu)
	if status == CommandStatusUnknown {
		// The response was a valid APDU but carried no parseable command
		// status: the remote outcome cannot be positively established.
		m.quarantineOutstation(outstationID)
		m.diag(DiagError, "control", seq, "control status unparseable; outcome unknown", nil)
		return CommandStatusUnknown, fmt.Errorf("%w: no parseable command status", ErrControlOutcomeUnknown)
	}
	// A positively classified outcome (success or explicit failure status).
	m.diag(DiagInfo, "control", seq, fmt.Sprintf("control outcome %d", status), nil)
	return status, nil
}

// resolveOperateStatus determines the per-point command status for a
// DirectOperate/Operate response (MEXT-012, fixing R1). A response is a
// success when EITHER:
//
//	(a) it carries a G12V1 object whose per-point status byte is 0
//	(CommandStatusSuccess), or
//	(b) it carries no G12V1 object (an IIN-only echo) and the IIN has no
//	error bits — real outstations may omit the G12V1 status echo on a valid
//	Direct-Operate success.
//
// A non-zero status byte is surfaced as-is (never success). An IIN-only
// response with error IIN is a failure (never success). A truncated G12V1
// object is a failure (CommandStatusUnknown), never success.
func resolveOperateStatus(req *al.APDU, resp *al.Response) CommandStatus {
	// G41 (AO-002 V1, AO-003 V2): the outstation echoes the object with index, the
	// numeric value and a command-status octet. The status is accepted only when the
	// echoed object matches the request's group/variation/index exactly; a missing,
	// truncated, malformed or mismatched object is never success. There is
	// deliberately no IIN-only success fallback for G41: a conformant response
	// carries the object.
	if group, variation, index, value, ok := controlRequestKey(req); ok && group == 41 && g41ValueWidth(variation) != 0 {
		return g41CommandStatus(resp.Data, variation, index, value, resp.IIN)
	}

	status := parseCommandStatus(resp.Data)
	if status != CommandStatusUnknown {
		return status
	}
	// Unknown: either no G12V1 object (IIN-only) or a truncated G12V1 object.
	if containsG12V1Header(resp.Data) {
		// A G12V1 header was present but the status byte could not be read
		// (truncated) — treat as a failure, never success.
		return CommandStatusUnknown
	}
	// IIN-only response (no object): fall back to the IIN. A clear IIN means
	// the outstation accepted the command (success); error IIN means it was
	// rejected (failure, never success).
	return commandStatusFromIIN(resp.IIN)
}

// controlRequestKey extracts the group, variation, first index and the G41
// numeric value from a control request APDU's object data (the request header
// built by buildControlRequest). Only the single-point qualifier 0x28 this slice
// emits is recognized; any other qualifier yields ok=false, so it is never
// treated as a matching G41 request. The value is the raw little-endian payload
// normalized into a uint64 and is meaningful only for the G41 variations this
// slice emits (four octets for V1/V3, two octets for V2, eight octets for V4);
// the comparison width is derived from the variation, not from this field alone.
func controlRequestKey(req *al.APDU) (group, variation uint8, index uint16, value uint64, ok bool) {
	if req == nil || len(req.Data) < 7 {
		return 0, 0, 0, 0, false
	}
	group = req.Data[0]
	variation = req.Data[1]
	// Qualifier 0x28 (16-bit count + 16-bit index): header(5) + index(2 LE).
	if req.Data[2] != 0x28 {
		return 0, 0, 0, 0, false
	}
	index = uint16(req.Data[5]) | uint16(req.Data[6])<<8
	if group == 41 {
		switch g41ValueWidth(variation) {
		case 8:
			if len(req.Data) < 15 {
				return 0, 0, 0, 0, false
			}
			for i := 0; i < 8; i++ {
				value |= uint64(req.Data[7+i]) << (8 * uint(i))
			}
		case 4:
			if len(req.Data) < 11 {
				return 0, 0, 0, 0, false
			}
			value = uint64(uint32(req.Data[7]) | uint32(req.Data[8])<<8 | uint32(req.Data[9])<<16 | uint32(req.Data[10])<<24)
		case 2:
			if len(req.Data) < 9 {
				return 0, 0, 0, 0, false
			}
			value = uint64(uint16(req.Data[7]) | uint16(req.Data[8])<<8)
		default:
			return 0, 0, 0, 0, false
		}
	}
	return group, variation, index, value, true
}

// g41ValueWidth returns the numeric value width in octets for a supported Group
// 41 command variation, or 0 if the variation is not supported. AO-002 covers V1
// (int32, 4 octets); AO-003 adds V2 (int16, 2 octets), V3 (float32, 4 octets)
// and V4 (float64, 8 octets).
func g41ValueWidth(variation uint8) int {
	switch variation {
	case 1:
		return 4 // int32
	case 2:
		return 2 // int16
	case 3:
		return 4 // float32
	case 4:
		return 8 // float64
	default:
		return 0
	}
}

// g41CommandStatus returns the per-point command status from a G41 response
// object. The response is accepted only when it is structurally exactly one G41
// record of the requested variation using qualifier 0x28 (16-bit count = 1,
// 16-bit index) with no surplus bytes or extra records, and the echoed index AND
// numeric value both match the request. Any deviation yields
// CommandStatusUnknown (never success).
//
// AO-002 F6: when the echoed per-point status is SUCCESS, it is reconciled
// against the response IIN before SUCCESS is returned. An error-indicating IIN
// contradicts a SUCCESS echo, so the IIN's implied failure status wins
// (FuncUnknown/ObjectUnknown -> NotSupported, ParameterError/BadConfig ->
// BadFormat, AlreadyExecuting -> AlreadyActive, LocalControl -> Local,
// BufferOverflow -> Unknown). A clean or non-error IIN leaves SUCCESS. An
// explicit non-SUCCESS per-point status is never overwritten.
func g41CommandStatus(data []byte, variation uint8, wantIndex uint16, wantValue uint64, iin al.IIN) CommandStatus {
	idx, value, status, ok := parseG41Record(data, 0, variation)
	if !ok || idx != wantIndex || value != wantValue {
		return CommandStatusUnknown
	}
	if status == CommandStatusSuccess {
		return commandStatusFromIIN(iin)
	}
	return status
}

// parseG41Record reads a single G41 record of the given variation at the given
// header offset. The only supported qualifier is 0x28 (16-bit count + 16-bit
// index) with a count of exactly 1; the record is exactly header(5) + index(2 LE)
// + value(LE, g41ValueWidth octets) + status(1) with no surplus bytes. The legacy
// 0x00 branch is not accepted as a valid G41 control record. ok is false for any
// other shape or an unsupported variation.
func parseG41Record(data []byte, offset int, variation uint8) (index uint16, value uint64, status CommandStatus, ok bool) {
	width := g41ValueWidth(variation)
	if width == 0 {
		return 0, 0, 0, false
	}
	recordLen := 5 + 2 + width + 1
	if offset < 0 || offset+recordLen != len(data) {
		return 0, 0, 0, false
	}
	if data[offset] != 41 || data[offset+1] != variation || data[offset+2] != 0x28 {
		return 0, 0, 0, false
	}
	count := int(data[offset+3]) | int(data[offset+4])<<8
	if count != 1 {
		return 0, 0, 0, false
	}
	index = uint16(data[offset+5]) | uint16(data[offset+6])<<8
	switch width {
	case 8:
		for i := 0; i < 8; i++ {
			value |= uint64(data[offset+7+i]) << (8 * uint(i))
		}
	case 4:
		value = uint64(uint32(data[offset+7]) | uint32(data[offset+8])<<8 | uint32(data[offset+9])<<16 | uint32(data[offset+10])<<24)
	case 2:
		value = uint64(uint16(data[offset+7]) | uint16(data[offset+8])<<8)
	}
	status = CommandStatus(data[offset+7+width])
	return index, value, status, true
}

// containsG12V1Header reports whether the object data contains a Group 12
// Variation 1 object header. It distinguishes an IIN-only response (no
// object at all) from a truncated G12V1 response (header present but status
// byte missing).
func containsG12V1Header(data []byte) bool {
	offset := 0
	for offset+4 <= len(data) {
		if isControlObject(data[offset], data[offset+1]) {
			return true
		}
		next, ok := skipObject(data, offset, data[offset], data[offset+1], data[offset+2])
		if !ok {
			return false
		}
		offset = next
	}
	return false
}

// commandStatusFromIIN maps an IIN-only DirectOperate response to a command
// status. A clear IIN (no error bits) is CommandStatusSuccess — the
// outstation accepted the command without echoing a G12V1 status object.
// Error IIN bits map to the corresponding failure status and are never success.
func commandStatusFromIIN(iin al.IIN) CommandStatus {
	switch {
	case iin.FuncUnknown, iin.ObjectUnknown:
		return CommandStatusNotSupported
	case iin.ParameterError, iin.BadConfig:
		return CommandStatusBadFormat
	case iin.AlreadyExecuting:
		return CommandStatusAlreadyActive
	case iin.LocalControl:
		return CommandStatusLocal
	case iin.BufferOverflow:
		return CommandStatusUnknown
	default:
		// Clear IIN (or only non-error indications such as event/class flags):
		// the command was accepted.
		return CommandStatusSuccess
	}
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
	// VEC-01 B4 F2: the whole request shares one finite wall-clock budget.
	deadline := time.Now().Add(m.readTimeout)

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
					DIR:      true, // Master-to-Outstation
					PRM:      true, // Primary station
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
			m.diag(DiagWarn, "send", seq, "send failed", lastErr)
			if !m.retryAgain(policy, &lastErr, attempt) {
				return lastErr
			}
			if m.deadlineExpired(deadline) {
				return ErrOperationDeadlineExceeded
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
			if err := m.waitForConfirmationTimeout(constrainTimeout(m, m.confirmTimeout, deadline), seq); err != nil {
				lastErr, _ = m.markDisconnected(err)
				m.diag(DiagWarn, "confirm", seq, "confirmation failed", err)
				if !m.retryAgain(policy, &lastErr, attempt) {
					return lastErr
				}
				if m.deadlineExpired(deadline) {
					return ErrOperationDeadlineExceeded
				}
				continue
			}
		}

		// Wait for response
		resp, err := m.waitForResponseTimeout(constrainTimeout(m, time.Duration(m.config.Timeout)*time.Millisecond, deadline), outstationID)
		if err != nil {
			lastErr, _ = m.markDisconnected(err)
			m.diag(DiagWarn, "receive", seq, "response failed", err)
			if !m.retryAgain(policy, &lastErr, attempt) {
				return lastErr
			}
			if m.deadlineExpired(deadline) {
				return ErrOperationDeadlineExceeded
			}
			continue
		}

		// Process response; the response SEQ must match the request (DNP3-010).
		_, err = m.processResponse(resp, outstationID, seq)
		if err != nil {
			lastErr = err
			m.diag(DiagWarn, "retry", seq, "response error", err)
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

// sendWithRetryAndGetResponse sends a request with retry logic and returns the
// processed application layer data.
//
// DNP3-034: see sendWithRetry — failures are classified and the retry decision
// is taken from the master's RetryPolicy.
//
// VEC-01 B4 F2: the whole read request (all retries, confirm wait and response
// wait) shares ONE finite wall-clock budget resolved from ReadTimeout. Retries
// remain bounded and interruptible; a retry that would start after the budget
// is exhausted is refused with ErrOperationDeadlineExceeded instead of silently
// extending the budget.
func (m *Master) sendWithRetryAndGetResponse(req *al.APDU, outstationID uint16) ([]byte, error) {
	return m.performReadRequest(req, outstationID, time.Now().Add(m.readTimeout))
}

// performReadRequest is the retry-enabled read/response path with a single
// finite deadline (VEC-01 B4 F2). allowRetry==true preserves the historical
// read retry policy, bounded and deadline-aware.
func (m *Master) performReadRequest(req *al.APDU, outstationID uint16, deadline time.Time) ([]byte, error) {
	return m.performRequest(req, outstationID, deadline)
}

// performRequest is the single, deadline-aware request engine used by both the
// read path (allowRetry==true) and the single-shot control path
// (allowRetry==false). It returns the decoded application response data.
//
// VEC-01 B4 F1: when allowRetry==false the request is attempted EXACTLY ONCE.
// There is no hidden retry on timeout, NACK, CRC, invalid response, missing
// confirmation or disconnect. Because the v0 DirectOperate is non-idempotent,
// a second send could double-execute the control, so the dedicated control path
// must never resend on its own.
//
// VEC-01 B4 F2: every send/receive/confirm wait and retry delay respects the
// remaining budget; the parent budget is never reset for a retry. Blocking
// receive waits are capped to the remaining budget via the transport timeout so
// an operation cannot overshoot its budget by a full step timeout.
func (m *Master) performRequest(req *al.APDU, outstationID uint16, deadline time.Time) ([]byte, error) {
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

		if m.deadlineExpired(deadline) {
			return nil, ErrOperationDeadlineExceeded
		}

		sentOK := true
		for _, frag := range fragments {
			// Transport layer encode
			tlEncoded := tl.EncodeFragment(frag)

			// Data link layer frame (DNP3-057: set FCB + FCV on confirmed data)
			dllFrame := &frame.Frame{
				Control: frame.Control{
					DIR:      true, // Master-to-Outstation
					PRM:      true, // Primary station
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
			if m.deadlineExpired(deadline) {
				return nil, ErrOperationDeadlineExceeded
			}
			continue
		}
		// The request was sent successfully; advance the sequence (DNP3-008).
		m.advanceSequence(outstationID)
		m.recordActivity()
		m.diag(DiagInfo, "send", seq, fmt.Sprintf("sent request to outstation %d", outstationID), nil)

		// If CON bit is set, wait for confirmation first.
		if req.Control.CON {
			if err := m.waitForConfirmationTimeout(constrainTimeout(m, m.confirmTimeout, deadline), seq); err != nil {
				lastErr, _ = m.markDisconnected(err)
				m.diag(DiagWarn, "confirm", seq, "confirmation failed", err)
				if deadlineErr := m.deadlineOrRetry(policy, &lastErr, attempt, deadline); deadlineErr != nil {
					return nil, deadlineErr
				}
				continue
			}
		}

		// Wait for response.
		resp, err := m.waitForResponseTimeout(constrainTimeout(m, time.Duration(m.config.Timeout)*time.Millisecond, deadline), outstationID)
		if err != nil {
			lastErr, _ = m.markDisconnected(err)
			m.diag(DiagWarn, "receive", seq, "response failed", err)
			if deadlineErr := m.deadlineOrRetry(policy, &lastErr, attempt, deadline); deadlineErr != nil {
				return nil, deadlineErr
			}
			continue
		}

		// Process the response and get application layer data; the response
		// SEQ must match the request (DNP3-010).
		appData, err := m.processResponse(resp, outstationID, seq)
		if err != nil {
			lastErr = err
			m.diag(DiagWarn, "retry", seq, "response error", err)
			if deadlineErr := m.deadlineOrRetry(policy, &lastErr, attempt, deadline); deadlineErr != nil {
				return nil, deadlineErr
			}
			continue
		}

		// Successful response counts as link activity for the idle monitor (DNP3-042).
		m.recordActivity()
		// DNP3-057: the confirmed-data transaction committed (send + optional
		// confirmation + response all succeeded); toggle the FCB for the next
		// transaction. Retries of a failed attempt keep the same FCB.
		m.advanceOutstationFCB(outstationID)
		m.diag(DiagInfo, "receive", seq, "response received", nil)
		return appData, nil
	}
}

// performance helpers

// deadlineExpired reports whether deadline has passed (a zero deadline never
// expires).
func (m *Master) deadlineExpired(deadline time.Time) bool {
	return !deadline.IsZero() && !time.Now().Before(deadline)
}

// waitHitBudget reports whether a blocking transport wait that returned an
// error consumed essentially the whole budget because the deadline capped the
// nominal timeout (VEC-01 B4 F2). When the deadline shortened the wait and the
// wait then elapsed for most of that shortened window, the failure is an
// operation-deadline overrun rather than a peer error. A genuine early failure
// (the wait returns almost immediately) leaves the peer error to surface.
func (m *Master) waitHitBudget(deadline time.Time, nominal, wait time.Duration, started time.Time) bool {
	if deadline.IsZero() {
		return false
	}
	if m.deadlineExpired(deadline) {
		return true
	}
	if wait >= nominal {
		// The deadline did not shorten this wait, so a failure is not a
		// deadline overrun.
		return false
	}
	return time.Since(started) >= wait-wait/8
}

// constrainTimeout returns the per-step timeout capped to the remaining
// wall-clock budget, so a single blocking receive cannot overshoot the
// operation deadline (VEC-01 B4 F2).
func constrainTimeout(m *Master, step time.Duration, deadline time.Time) time.Duration {
	if deadline.IsZero() {
		return step
	}
	rem := time.Until(deadline)
	if rem <= 0 {
		return time.Millisecond
	}
	if step <= 0 || rem < step {
		return rem
	}
	return step
}

// deadlineOrRetry applies the single-shot/retry decision for a mid-request
// failure and enforces the operation deadline (VEC-01 B4 F1/F2). It returns a
// non-nil error to surface immediately (single-shot path, retry budget
// exhausted, or deadline exceeded) or nil when the caller should retry.
func (m *Master) deadlineOrRetry(policy *RetryPolicy, lastErr *error, attempt int, deadline time.Time) error {
	if !m.retryAgain(policy, lastErr, attempt) {
		return *lastErr
	}
	if m.deadlineExpired(deadline) {
		return ErrOperationDeadlineExceeded
	}
	return nil
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

// SendRequestWithRetryAndGetResponseDeadline is the deadline-aware read dispatch
// (VEC-01 B4 F2). The caller supplies an absolute deadline that bounds the whole
// read/response exchange (all bounded retries, confirm wait and response wait).
// A zero deadline falls back to the master's configured read budget. This lets a
// multi-group integrity poll share ONE budget across its primary and per-group
// fallback exchanges instead of resetting the budget per group.
func (m *Master) SendRequestWithRetryAndGetResponseDeadline(req *al.APDU, outstationID uint16, deadline time.Time) ([]byte, error) {
	if deadline.IsZero() {
		deadline = time.Now().Add(m.readTimeout)
	}
	return m.performReadRequest(req, outstationID, deadline)
}

// ReadBudget returns the configured whole-request budget (VEC-01 B4 F2) so the
// public client can derive a shared absolute deadline for a multi-exchange
// operation such as an integrity poll.
func (m *Master) ReadBudget() time.Duration {
	return m.readTimeout
}

// waitForConfirmation waits for an application-layer confirmation for a request
// sent with the CON bit set (DNP3-009). A confirmation is an APDU with
// FuncCode=FuncConfirm (0x00). The confirmation's sequence number must match the
// request's sequence; a mismatch is a protocol error (ErrConfirmSeqMismatch).
// A transport receive error (timeout) is reported as ErrConfirmTimeout, which
// the caller retries.
//
// In this stack a full response (FuncResponse carrying IIN+data) may arrive in
// lieu of a dedicated confirm; such a response is accepted as the confirmation.
// Strict response-sequence matching is the responsibility of DNP3-010.
func (m *Master) waitForConfirmation(expectedSeq uint8) error {
	return m.waitForConfirmationTimeout(m.confirmTimeout, expectedSeq)
}

// waitForConfirmationTimeout waits for an application-layer confirmation with
// stepTimeout already capped to the operation's remaining budget (VEC-01 B4 F2).
func (m *Master) waitForConfirmationTimeout(stepTimeout time.Duration, expectedSeq uint8) error {
	// DNP3-066: the confirm timeout is distinct from the response timeout. A
	// confirm is a small dedicated reply, so it typically uses a shorter
	// budget than a full object response. VEC-01 B4 F2: the caller passes this
	// timeout already capped to the operation's remaining wall-clock budget.
	m.transport.SetTimeout(int(stepTimeout / time.Millisecond))

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

		// A confirmation is function code 0x00; a full solicited response may also
		// arrive in lieu of a dedicated confirm (FuncResponse).
		if apdu.FuncCode != al.FuncConfirm && apdu.FuncCode != al.FuncResponse {
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
	return m.waitForResponseTimeout(time.Duration(m.config.Timeout)*time.Millisecond, 0)
}

// waitForResponseTimeout waits for a response with stepTimeout already capped to
// the operation's remaining budget (VEC-01 B4 F2). peer is the outstation the
// solicited request was sent to, and is used only to confirm an unsolicited
// fragment that arrives during the wait.
func (m *Master) waitForResponseTimeout(stepTimeout time.Duration, peer uint16) ([]byte, error) {
	// VEC-01 B4 F2: stepTimeout is the response timeout capped to the
	// operation's remaining wall-clock budget by the caller. The overall wait is
	// bounded by this single deadline so that a peer emitting repeated bare
	// link-layer ACKs cannot extend the wait indefinitely.
	deadline := time.Now().Add(stepTimeout)
	// A solicited response may span several transport fragments, and the
	// transport delivers one data-link frame per Receive(). Accumulate the raw
	// frames of the in-progress message and re-run the DLL+TL path over the whole
	// buffer each time; errIncompleteResponse means more fragments are needed.
	var acc []byte
	// appConfirmed counts the non-final application fragments already confirmed on
	// this response, so a re-parse of the growing buffer does not confirm twice.
	appConfirmed := 0
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil, fmt.Errorf("%w: no response within %v", ErrTimeout, stepTimeout)
		}
		// Round up so the first iteration preserves the exact configured
		// timeout while later iterations still shrink toward the deadline.
		m.transport.SetTimeout(int((remaining + time.Millisecond - 1) / time.Millisecond))
		data, err := m.transport.Receive()
		if err != nil {
			if isDisconnectError(err) {
				return nil, err
			}
			return nil, fmt.Errorf("%w: %v", ErrTimeout, err)
		}
		// A conformant outstation ACKs confirmed user data at the link layer
		// (a bare secondary ACK carrying no transport fragment) before sending
		// its solicited application response. Skip it only when the whole receive
		// is exactly that one bare frame, so a segment that also carries a
		// transport fragment is not dropped. NACKs are surfaced earlier by
		// processReceivedBytes, so only ACK is skipped here.
		if len(data) == frame.EncodedSize(0) {
			if f, derr := frame.Decode(data); derr == nil && !f.Control.PRM &&
				f.Control.FuncCode == frame.FuncAck && len(f.Data) == 0 {
				continue
			}
		}
		acc = append(acc, data...)
		if len(acc) > maxResponseBytes {
			return nil, fmt.Errorf("%w: response exceeds %d bytes", ErrInvalidResponse, maxResponseBytes)
		}
		// An unsolicited fragment has its own application sequence. Confirm it
		// when it is in order and keep waiting for the solicited response.
		app, aerr := m.processReceivedBytes(acc)
		if errors.Is(aerr, errIncompleteResponse) {
			// A multi-fragment solicited response requires an application-layer
			// confirmation per fragment: OpenDNP3's outstation sets CON on every
			// fragment and will not send the next until the current one is
			// confirmed. Confirm each complete fragment once.
			m.confirmPendingAppFragments(acc, peer, &appConfirmed)
			continue
		}
		if aerr == nil && m.consumeUnsolicited(app, peer) {
			acc = acc[:0]
			appConfirmed = 0
			continue
		}
		// The solicited response is complete. Confirm every CON fragment not yet
		// confirmed, including the final one: OpenDNP3 sets CON on the last
		// fragment of a multi-fragment solicited response too and stays in the
		// confirm-wait state until it is confirmed.
		m.confirmPendingAppFragments(acc, peer, &appConfirmed)
		return acc, nil
	}
}

// confirmPendingAppFragments sends a solicited application confirmation for
// every complete application fragment in data that carries CON and has not yet
// been confirmed. done counts the fragments already handled; it advances past
// each fragment so a re-parse of the growing buffer does not confirm twice.
//
// Each fragment is confirmed with its own sequence: OpenDNP3 advances the
// solicited sequence per fragment and requires the matching confirmation before
// sending the next.
func (m *Master) confirmPendingAppFragments(data []byte, peer uint16, done *int) {
	frags, err := m.parseAppFragments(data)
	if err != nil || len(frags) == 0 {
		return
	}
	for i := *done; i < len(frags); i++ {
		if frag, derr := al.Decode(frags[i]); derr == nil && frag.Control.CON {
			_ = m.sendApplicationConfirm(frag.Control.Seq, peer, false)
		}
		*done = i + 1
	}
}

// SetUnsolicitedCallback registers the handler for an accepted unsolicited
// application fragment. The callback runs without reqMu and must not start
// another request on this master.
func (m *Master) SetUnsolicitedCallback(fn func([]byte)) {
	m.unsolMu.Lock()
	m.onUnsolicited = fn
	m.unsolMu.Unlock()
}

// ResetUnsolicitedSession accepts the next unsolicited sequence as the first
// sequence of a new connection.
func (m *Master) ResetUnsolicitedSession() {
	m.unsolMu.Lock()
	m.unsolInit = false
	m.unsolMu.Unlock()
}

// consumeUnsolicited reports whether app is an unsolicited fragment.
// In-order fragments are delivered once and confirmed. A duplicate of the
// last accepted sequence is confirmed and not delivered. Any other sequence
// is neither confirmed nor delivered.
func (m *Master) consumeUnsolicited(app []byte, peer uint16) bool {
	resp, err := al.DecodeResponse(app)
	if err != nil {
		return false
	}
	if resp.Header.FuncCode != al.FuncUnsolicitedResponse && !resp.Header.Control.UNS {
		return false
	}
	seq := resp.Header.Control.Seq
	m.unsolMu.Lock()
	deliver := false
	confirm := false
	switch {
	case !m.unsolInit:
		m.unsolInit = true
		m.unsolLast = seq
		deliver = true
		confirm = true
	case seq == m.unsolLast:
		confirm = true
	case seq == (m.unsolLast+1)%16:
		m.unsolLast = seq
		deliver = true
		confirm = true
	}
	cb := m.onUnsolicited
	m.unsolMu.Unlock()
	if deliver && cb != nil {
		cb(app)
	}
	if confirm && resp.Header.Control.CON {
		_ = m.sendApplicationConfirm(seq, peer, true)
	}
	return true
}

// PollUnsolicited waits for one unsolicited fragment. It holds the request
// lock only for this receive, so it does not run beside a solicited read.
func (m *Master) PollUnsolicited(peer uint16, wait time.Duration) error {
	if !m.isLinkUp() {
		return ErrNotConnected
	}
	if wait <= 0 {
		wait = 200 * time.Millisecond
	}
	m.reqMu.Lock()
	defer m.reqMu.Unlock()
	m.transport.SetTimeout(int((wait + time.Millisecond - 1) / time.Millisecond))
	defer m.transport.SetTimeout(m.config.Timeout)

	deadline := time.Now().Add(wait)
	var acc []byte
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return fmt.Errorf("%w: no unsolicited response within %v", ErrTimeout, wait)
		}
		m.transport.SetTimeout(int((remaining + time.Millisecond - 1) / time.Millisecond))
		data, err := m.transport.Receive()
		if err != nil {
			return err
		}
		if f, derr := frame.Decode(data); derr == nil && !f.Control.PRM &&
			f.Control.FuncCode == frame.FuncAck && len(f.Data) == 0 {
			continue
		}
		// A single unsolicited response may span several transport fragments, one
		// data-link frame per Receive(). Accumulate raw frames and retry until the
		// message is complete.
		acc = append(acc, data...)
		if len(acc) > maxResponseBytes {
			return fmt.Errorf("%w: unsolicited response exceeds %d bytes", ErrInvalidResponse, maxResponseBytes)
		}
		app, err := m.processReceivedBytes(acc)
		if errors.Is(err, errIncompleteResponse) {
			continue
		}
		if err != nil {
			return err
		}
		if !m.consumeUnsolicited(app, peer) {
			return fmt.Errorf("%w: received a solicited frame while waiting for unsolicited data", ErrInvalidResponse)
		}
		return nil
	}
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

	// DNP3-054: an application-layer confirmation (CON=1) is sent from the
	// response receive path (waitForResponseTimeout), where each application
	// fragment's own sequence is available. A multi-fragment response carries a
	// distinct sequence per fragment, so confirming from the joined APDU (which
	// keeps the first fragment's sequence for the DNP3-010 check) would send the
	// wrong sequence.

	// DNP3-013: react to critical IIN bits.
	m.reactToIIN(outstationID, resp.IIN, o)

	return appData, nil
}

// sendApplicationConfirm sends a DNP3 application-layer confirmation (DNP3-054)
// to outstationID for the response carrying seq. The confirm is a single
// FIR|FIN APDU with FuncCode 0 (Confirm) and CON=0, sent as unconfirmed user
// data (no link-layer ACK expected).
func (m *Master) sendApplicationConfirm(seq uint8, outstationID uint16, unsolicited bool) error {
	confirm := al.NewConfirm(seq)
	// IEEE 1815: the confirmation of an unsolicited fragment carries UNS.
	// A solicited confirmation leaves UNS clear. OpenDNP3 ignores a null
	// unsolicited confirmation that does not set the bit.
	if unsolicited {
		confirm.Control.UNS = true
	}
	data := confirm.Encode()
	fragments := m.fragmenter.Fragmentize(data)
	for _, frag := range fragments {
		tlEncoded := tl.EncodeFragment(frag)
		dllFrame := &frame.Frame{
			Control: frame.Control{
				DIR:      true, // Master-to-Outstation
				PRM:      true, // Primary station
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
//   - DeviceRestart (octet 1 mask 0x80): clear local state and flag re-integrity.
//   - NeedTime (octet 1 mask 0x10): record that a time sync is required (stub).
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
	frags, err := m.parseAppFragments(data)
	if err != nil {
		return nil, err
	}
	if len(frags) == 0 {
		return nil, errIncompleteResponse
	}
	// The run is complete only once the final application fragment (FIN=1) has
	// arrived; until then the caller keeps reading.
	if last, lerr := al.Decode(frags[len(frags)-1]); lerr == nil && !last.Control.FIN {
		return nil, errIncompleteResponse
	}
	if len(frags) == 1 {
		return frags[0], nil
	}
	return joinAppFragments(frags), nil
}

// parseAppFragments decodes all complete transport messages in data, in order.
// Each FIN-terminated transport message is one application fragment; the
// reassembler is reset between messages so consecutive application fragments of
// one response are not conflated. It surfaces the same link-layer error classes
// as the receive path (ErrCRCError, ErrLinkNACK).
func (m *Master) parseAppFragments(data []byte) ([][]byte, error) {
	m.reassembler.Reset()

	var frags [][]byte
	offset := 0
	for offset < len(data) {
		// Decode DLL frame
		dllFrame, err := frame.Decode(data[offset:])
		if err != nil {
			// If fragments were already parsed, return them; trailing bytes may be a
			// partial frame that a later read completes.
			if len(frags) > 0 {
				return frags, nil
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

		// Only user-data frames carry a transport fragment. A confirmed or
		// unconfirmed user-data frame is identified by function code 3 or 4
		// regardless of PRM: a conformant outstation acts as the primary station
		// (DIR=0, PRM=1) when transmitting its solicited responses, so a PRM
		// check must not discard them. Non-data link frames (ACK/NACK/link
		// status/reset) carry no fragment and are skipped.
		if dllFrame.Control.FuncCode != frame.FuncConfirmedUserData &&
			dllFrame.Control.FuncCode != frame.FuncConfirmedUserDataR {
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

		// Push fragment to reassembler - returns one complete application fragment
		// when FIN is received.
		result, err := m.reassembler.Push(tlFrag)
		if err != nil {
			return nil, fmt.Errorf("reassembly error: %w", err)
		}

		if result != nil {
			frags = append(frags, result)
			// A FIN ends this transport message; the next fragment (if any) starts a
			// new application fragment with its own FIR.
			m.reassembler.Reset()
		}

	}

	return frags, nil
}

// joinAppFragments concatenates the object data of consecutive application
// fragments into one APDU. The first fragment's control byte, function code, and
// IIN are authoritative; each fragment repeats them, so only the object data
// (octets after the 4-octet header) of every later fragment is appended.
func joinAppFragments(frags [][]byte) []byte {
	first, err := al.Decode(frags[0])
	if err != nil {
		return frags[0]
	}
	out := make([]byte, 0, 4096)
	out = append(out, first.Control.Header(), first.FuncCode)
	out = append(out, first.Data...)
	for _, f := range frags[1:] {
		frag, derr := al.Decode(f)
		if derr != nil || len(frag.Data) < 2 {
			continue
		}
		out = append(out, frag.Data[2:]...)
	}
	return out
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
	// Object prefix: Group, Variation, qualifier 0x28 (16-bit count + 16-bit
	// index, UINT16_CNT_UINT16_INDEX), count = 1 (LSB-first). A bare index-only
	// qualifier is rejected by conformant peers (OpenDNP3 IsAllowed requires a
	// count+index qualifier for control objects).
	prefix := []byte{group, variation, 0x28, 0x01, 0x00}

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

		// Reject raw operation types the v0 profile does not support rather
		// than emitting a non-conformant code octet.
		if !isSupportedCROBOpType(code) {
			return nil
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

	case 41:
		// Group 41 (Analog Output). AO-002 implements Variation 1 (32-bit signed
		// integer); AO-003 adds Variation 2 (16-bit signed integer) and Variation 3
		// (single-precision IEEE-754 float). Per the verified independent OpenDNP3
		// 3.1.2 reference (docs/verification/AO-001.md), a G41 object is a value
		// serialized little-endian followed by one command-status octet. Any other
		// Group 41 variation, a wrong value type, a non-integral float, an
		// out-of-range integer, or a NaN/±Inf float is rejected before sending.
		switch variation {
		case 1:
			intVal, ok := controlInt32Value(value)
			if !ok {
				return nil
			}
			valueBytes = []byte{
				byte(intVal), byte(intVal >> 8), byte(intVal >> 16), byte(intVal >> 24),
				0x00, // request object command status (SUCCESS)
			}
		case 2:
			intVal, ok := controlInt16Value(value)
			if !ok {
				return nil
			}
			valueBytes = []byte{
				byte(intVal), byte(intVal >> 8),
				0x00, // request object command status (SUCCESS)
			}
		case 3:
			// V3 float32: reject NaN/±Inf before transmission; preserve the exact
			// IEEE-754 bits (including signed zero and subnormals) via Float32bits.
			f, ok := controlFloat32Value(value)
			if !ok {
				return nil
			}
			bits := math.Float32bits(f)
			valueBytes = []byte{
				byte(bits), byte(bits >> 8), byte(bits >> 16), byte(bits >> 24),
				0x00, // request object command status (SUCCESS)
			}
		case 4:
			// V4 float64: reject NaN/±Inf before transmission; preserve the exact
			// IEEE-754 bits (including signed zero and subnormals) via Float64bits.
			f, ok := controlFloat64Value(value)
			if !ok {
				return nil
			}
			bits := math.Float64bits(f)
			valueBytes = []byte{
				byte(bits), byte(bits >> 8), byte(bits >> 16), byte(bits >> 24),
				byte(bits >> 32), byte(bits >> 40), byte(bits >> 48), byte(bits >> 56),
				0x00, // request object command status (SUCCESS)
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
	Code    uint8  // Operation type (IEEE 1815 A.8.1 low-nibble enum; see CROBCode* constants)
	Count   uint8  // Number of operation cycles
	OnTime  uint32 // Time in ms to energize
	OffTime uint32 // Time in ms to de-energize
	Status  uint8  // Qualifier code (0x00=STANDARD)
}

// CROB operation-type values for the G12V1 control-code octet. IEEE 1815
// (A.8.1) and OpenDNP3 define the operation type as a sequential enumeration in
// the LOW NIBBLE of the code octet, not independent one-hot flags. The upper
// nibble carries the trip/close code (bits 6-7) and clear (bit 5); the v0
// DirectOperate profile sends neither, so those bits must be zero. See
// active_work/crob-code-audit.md.
const (
	CROBCodeNUL      uint8 = 0x00 // no operation
	CROBCodePulseOn  uint8 = 0x01 // energize for OnTime
	CROBCodePulseOff uint8 = 0x02 // de-energize for OffTime
	CROBCodeLatchOn  uint8 = 0x03 // set output ON  (Value = true)
	CROBCodeLatchOff uint8 = 0x04 // set output OFF (Value = false)
)

// isSupportedCROBOpType reports whether code is a v0 G12V1 CROB operation type
// (IEEE 1815 A.8.1 low-nibble enumeration). The upper nibble (trip/close, clear)
// must be zero because the v0 profile does not use it.
func isSupportedCROBOpType(code uint8) bool {
	switch code {
	case CROBCodeNUL, CROBCodePulseOn, CROBCodePulseOff, CROBCodeLatchOn, CROBCodeLatchOff:
		return true
	default:
		return false
	}
}

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
	// CommandStatusHardwareError indicates something is stopping the command,
	// often a local/remote interlock (IEEE 1815 command status 6). The pinned
	// OpenDNP3 3.1.2 reference names code 6 HARDWARE_ERROR.
	CommandStatusHardwareError CommandStatus = 6
	// CommandStatusBlocked is a deprecated alias for CommandStatusHardwareError.
	// Code 6 is HARDWARE_ERROR in the reference; the wire value is unchanged.
	CommandStatusBlocked = CommandStatusHardwareError
	// CommandStatusLocal indicates the output is in local control mode.
	CommandStatusLocal CommandStatus = 7
	// CommandStatusTooMany indicates too many commands are queued.
	CommandStatusTooMany CommandStatus = 8
	// CommandStatusNotAuthorized indicates the command was not authorized.
	CommandStatusNotAuthorized CommandStatus = 9
	// CommandStatusAutonomous indicates autonomous operation.
	CommandStatusAutonomous CommandStatus = 10
	// CommandStatusOutOfRange indicates the requested point/index is outside the
	// outstation's supported range (IEEE 1815 command status 12). The pinned
	// OpenDNP3 3.1.2 reference names code 12 OUT_OF_RANGE.
	CommandStatusOutOfRange CommandStatus = 12
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

		if isControlObject(group, variation) {
			return commandStatusForQualifier(data, offset, group, variation, qualifier)
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

// isControlObject reports whether group/variation is a control object whose
// response echoes a per-point command status. AO-002 covers Group 12 Variation 1
// (CROB) and Group 41 Variation 1 (32-bit signed analog output); AO-003 adds
// Group 41 Variation 2 (16-bit signed analog output).
func isControlObject(group, variation uint8) bool {
	return (group == 12 && variation == 1) || (group == 41 && g41ValueWidth(variation) != 0)
}

// controlPointValueSize returns the per-point value size in octets (value plus
// the trailing command-status octet) for a control object, or 0 if unknown.
func controlPointValueSize(group, variation uint8) int {
	switch {
	case group == 12 && variation == 1:
		return 11 // code(1)+count(1)+onTime(4)+offTime(4)+status(1)
	case group == 41:
		if w := g41ValueWidth(variation); w != 0 {
			return w + 1 // numeric value(w)+status(1)
		}
	}
	return 0
}

// commandStatusForQualifier reads the per-point command status byte from a
// control object (G12V1 CROB or G41V1 analog output) at the given header offset
// for the given qualifier. The response object echoes the request: for CROB
// code(1), count(1), onTime(4), offTime(4), status(1); for G41V1 an int32
// value(4) then status(1) — the last octet is the command status. For the
// index-only qualifier (0x00) and the 16-bit count+index qualifier (0x28) the
// per-point index (2 octets) precedes the value.
func commandStatusForQualifier(data []byte, offset int, group, variation, qualifier uint8) CommandStatus {
	pointSize := controlPointValueSize(group, variation)
	if pointSize == 0 {
		return CommandStatusUnknown
	}
	statusInPoint := pointSize - 1
	switch qualifier {
	case 0x00: // index-only: header(4) + index(2) + point
		statusIdx := offset + 4 + 2 + statusInPoint
		if statusIdx < len(data) {
			return CommandStatus(data[statusIdx])
		}
	case 0x28: // 16-bit count + 16-bit index: header(5) + index(2) + point
		statusIdx := offset + 5 + 2 + statusInPoint
		if statusIdx < len(data) {
			return CommandStatus(data[statusIdx])
		}
	case 0x07, 0x27: // count8/count16: header(4) [+count16(2)] + point (no index prefix)
		hdrLen := 4
		if qualifier == 0x27 {
			hdrLen = 5
		}
		statusIdx := offset + hdrLen + statusInPoint
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
	case 0x28:
		if isControlObject(group, variation) {
			// 16-bit count + 16-bit index (control): header(5) then
			// count * (index(2) + point payload).
			if offset+5 > len(data) {
				return 0, false
			}
			count := int(data[offset+3]) | int(data[offset+4])<<8
			return offset + 5 + count*(2+pointSize), true
		}
		// 16-bit start/stop (reads): header(7) then n * point payload.
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
		return 11, qualifier == 0x00 || qualifier == 0x28, 2 // CROB; count+index uses a 2-octet index
	case group == 41:
		if w := g41ValueWidth(variation); w != 0 {
			return w + 1, qualifier == 0x00 || qualifier == 0x28, 2 // G41 numeric value + status
		}
	}
	return 0, false, 0
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

// controlInt32Value converts a control value to a signed 32-bit integer for the
// Group 41 Variation 1 analog-output path. It rejects (ok=false) any value that
// is not an exactly-representable, in-range int32: a float (even an integral
// one) is rejected rather than silently truncated or rounded, and an int/int64
// outside the int32 range is rejected rather than wrapped. The caller treats a
// false result as a definite pre-send rejection.
func controlInt32Value(value interface{}) (int32, bool) {
	switch v := value.(type) {
	case int32:
		return v, true
	case int:
		if v < math.MinInt32 || v > math.MaxInt32 {
			return 0, false
		}
		return int32(v), true
	case int64:
		if v < math.MinInt32 || v > math.MaxInt32 {
			return 0, false
		}
		return int32(v), true
	default:
		return 0, false
	}
}

// controlInt16Value returns the exact int16 carried by a G41V2 value, rejecting
// any source that is not an exactly-representable, in-range integer (an out-of-
// range value is rejected rather than wrapped/truncated). The caller treats a
// false result as a definite pre-send rejection.
func controlInt16Value(value interface{}) (int16, bool) {
	switch v := value.(type) {
	case int16:
		return v, true
	case int32:
		if v < math.MinInt16 || v > math.MaxInt16 {
			return 0, false
		}
		return int16(v), true
	case int:
		if v < math.MinInt16 || v > math.MaxInt16 {
			return 0, false
		}
		return int16(v), true
	case int64:
		if v < math.MinInt16 || v > math.MaxInt16 {
			return 0, false
		}
		return int16(v), true
	default:
		return 0, false
	}
}

// controlFloat32Value returns the exact float32 carried by a G41V3 value. NaN
// and ±Inf are rejected before transmission (the owner-approved AO-003 float
// policy), as is any source type that is not exactly representable as a float32:
// a float64 is accepted only when converting to float32 and back is bit-exact
// (rejecting precision loss, overflow to ±Inf and subnormal underflow to zero).
// Signed zero is preserved because the value is compared by bits, not by ==. The
// caller treats a false result as a definite pre-send rejection.
func controlFloat32Value(value interface{}) (float32, bool) {
	var f float32
	switch v := value.(type) {
	case float32:
		f = v
	case float64:
		c := float32(v)
		if float64(c) != v || math.IsInf(float64(c), 0) {
			return 0, false
		}
		f = c
	default:
		return 0, false
	}
	if math.IsNaN(float64(f)) || math.IsInf(float64(f), 0) {
		return 0, false
	}
	return f, true
}

// controlFloat64Value returns the exact float64 carried by a G41V4 value. NaN
// and ±Inf are rejected before transmission (the owner-approved AO-003 float
// policy). Only a float64 source is accepted: unlike V3 there is no narrower
// carrier to promote, and a float32 source is deliberately rejected so an
// accidental width mismatch at the API boundary is not silently widened. Signed
// zero is preserved because the value is compared by bits, not by ==. The caller
// treats a false result as a definite pre-send rejection.
func controlFloat64Value(value interface{}) (float64, bool) {
	f, ok := value.(float64)
	if !ok {
		return 0, false
	}
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, false
	}
	return f, true
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
