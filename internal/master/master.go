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

	// ErrTransportDisconnected indicates the transport reported the peer closed
	// the connection (or the transport was closed). The master transitions to
	// StateError when this is observed (DNP3-031).
	ErrTransportDisconnected = errors.New("transport disconnected")
)

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
}

// DefaultConfig returns default master configuration.
func DefaultConfig() *Config {
	return &Config{
		MasterAddress: 0xFFFF, // Master uses broadcast or specific address
		Timeout:       DefaultTimeout,
		MaxRetries:    MaxRetries,
		RetryDelay:    100,
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
	// sequence is the application-layer sequence number (0-15). It advances
	// by one (mod 16) for each successfully sent request (DNP3-008).
	sequence uint8

	// reqMu serializes the request path (send → receive → reassembly) so that
	// concurrent requests on the same master do not race on the shared
	// reassembler or interleave fragments on a single link (DNP3-025). A DNP3
	// link is request/response; concurrent requests are safely queued.
	reqMu sync.Mutex

	// DNP3-013 reaction hooks. Optional; invoked from processResponse when
	// the corresponding IIN bit is set. These are stubs/logs — full time
	// objects and integrity polling are later roadmap items.
	onDeviceRestart func(outstationID uint16)
	onNeedTimeSync  func(outstationID uint16)
}

// TransportHandler defines the interface for sending/receiving data.
type TransportHandler interface {
	Send(data []byte) error
	Receive() ([]byte, error)
	SetTimeout(ms int)
}

// ErrorHandler defines the callback for error notifications.
type ErrorHandler func(err error, context string)

// NewMaster creates a new Master Station.
func NewMaster(config *Config) *Master {
	if config == nil {
		config = DefaultConfig()
	}

	return &Master{
		config:      config,
		state:       StateDisconnected,
		outstations: make(map[uint16]*Outstation),
		fragmenter:  tl.NewFragmenter(),
		reassembler: tl.NewReassembler(),
	}

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

// SetState updates the master state.
func (m *Master) SetState(state State) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state = state
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

	m.SetState(StateConnecting)

	// Clear TL/sequence state left over from any previous session so a reconnect
	// after a mid-session drop re-handshakes from a clean slate (DNP3-032). This
	// discards stale reassembly buffers and resets the fragment sequence.
	m.resetForReconnect()

	// Perform DNP3 link-layer session establishment
	if err := m.performLinkHandshake(); err != nil {
		m.SetState(StateError)
		return fmt.Errorf("link handshake failed: %w", err)
	}

	m.SetState(StateConnected)
	m.SetState(StateActive)
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

		if err := m.transport.Send(encoded); err != nil {
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
			return fmt.Errorf("invalid reset link ACK: %w", err)
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
func (m *Master) Disconnect() error {
	m.SetState(StateDisconnected)
	m.outstations = make(map[uint16]*Outstation)
	return nil
}

// Initialize performs master initialization sequence.
func (m *Master) Initialize() error {
	if m.State() != StateConnected && m.State() != StateActive {
		return ErrNotConnected
	}

	m.SetState(StateInitialized)
	return nil
}

// Enable unsolicits on an outstation.
func (m *Master) EnableUnsolicited(outstationID uint16) error {
	if m.State() < StateConnected {
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
	if m.State() < StateConnected {
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
	if m.State() < StateInitialized {
		return ErrNotConnected
	}

	data := buildPollRequest(pollType)
	req := buildRequest(0, al.FuncRead, data)

	return m.sendWithRetry(req, outstationID)
}

// Operate issues a control operation.
func (m *Master) Operate(outstationID uint16, selectThenOperate bool, group, variation uint8, index uint16, value interface{}) error {
	if m.State() < StateInitialized {
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
	if m.State() < StateInitialized {
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
	if m.State() < StateInitialized {
		return ErrNotConnected
	}

	// Build time sync request with current time
	now := time.Now()
	data := encodeDNP3Time(now)
	req := buildRequest(0, al.FuncTimeSync, data)

	return m.sendWithRetry(req, outstationID)
}

// sendWithRetry sends a request with retry logic.
func (m *Master) sendWithRetry(req *al.APDU, outstationID uint16) error {
	// Serialize the request path so concurrent requests do not race on the
	// shared reassembler or interleave fragments on a single link (DNP3-025).
	m.reqMu.Lock()
	defer m.reqMu.Unlock()

	var lastErr error

	for attempt := 0; attempt < m.config.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(m.config.RetryDelay) * time.Millisecond)
		}

		// Allocate the application-layer sequence for this request attempt.
		// Retries reuse the same sequence value within one logical request;
		// the sequence advances only after a successful send (DNP3-008).
		seq := m.nextSequence()
		req.Control.Seq = seq

		// Send request with DNP3 protocol layers
		data := req.Encode()

		// Transport layer fragmentation
		fragments := m.fragmenter.Fragmentize(data)

		sentOK := true
		for _, frag := range fragments {
			// Transport layer encode
			tlEncoded := tl.EncodeFragment(frag)

			// Data link layer frame
			dllFrame := &frame.Frame{
				Control: frame.Control{
					DIR:      true, // Master-to-Outstation
					PRM:      true, // Primary station
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
			// (DNP3-031).
			if isDisconnectError(lastErr) {
				return lastErr
			}
			continue
		}
		// The request was sent successfully; advance the sequence so the
		// next request uses the next value in the 0-15 stream (DNP3-008).
		m.advanceSequence()
		// If CON bit is set, wait for confirmation first
		if req.Control.CON {
			if err := m.waitForConfirmation(seq); err != nil {
				lastErr, _ = m.markDisconnected(err)
				if isDisconnectError(lastErr) {
					return lastErr
				}
				continue
			}
		}

		// Wait for response
		resp, err := m.waitForResponse()
		if err != nil {
			lastErr, _ = m.markDisconnected(err)
			if isDisconnectError(lastErr) {
				return lastErr
			}
			continue
		}

		// Process response; the response SEQ must match the request (DNP3-010).
		_, err = m.processResponse(resp, outstationID, seq)
		if err != nil {
			lastErr = err
			continue
		}

		return nil // Success
	}

	return fmt.Errorf("%w: %v", ErrMaxRetries, lastErr)
}

// sendWithRetryAndGetResponse sends a request with retry logic and returns the processed application layer data.
func (m *Master) sendWithRetryAndGetResponse(req *al.APDU, outstationID uint16) ([]byte, error) {
	// Serialize the request path so concurrent requests do not race on the
	// shared reassembler or interleave fragments on a single link (DNP3-025).
	m.reqMu.Lock()
	defer m.reqMu.Unlock()

	var lastErr error

	for attempt := 0; attempt < m.config.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(m.config.RetryDelay) * time.Millisecond)
		}

		// Allocate the application-layer sequence for this request attempt
		// (DNP3-008). Retries reuse the same value; it advances only on a
		// successful send.
		seq := m.nextSequence()
		req.Control.Seq = seq

		// Send request with DNP3 protocol layers
		data := req.Encode()

		// Transport layer fragmentation
		fragments := m.fragmenter.Fragmentize(data)

		sentOK := true
		for _, frag := range fragments {
			// Transport layer encode
			tlEncoded := tl.EncodeFragment(frag)

			// Data link layer frame
			dllFrame := &frame.Frame{
				Control: frame.Control{
					DIR:      true, // Master-to-Outstation
					PRM:      true, // Primary station
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
			// (DNP3-031).
			if isDisconnectError(lastErr) {
				return nil, lastErr
			}
			continue
		}
		// Successful send; advance the sequence (DNP3-008).
		m.advanceSequence()
		// If CON bit is set, wait for confirmation first
		if req.Control.CON {
			if err := m.waitForConfirmation(seq); err != nil {
				lastErr, _ = m.markDisconnected(err)
				if isDisconnectError(lastErr) {
					return nil, lastErr
				}
				continue
			}
		}

		// Wait for response
		resp, err := m.waitForResponse()
		if err != nil {
			lastErr, _ = m.markDisconnected(err)
			if isDisconnectError(lastErr) {
				return nil, lastErr
			}
			continue
		}

		// Process the response and get application layer data; the response
		// SEQ must match the request (DNP3-010).
		appData, err := m.processResponse(resp, outstationID, seq)
		if err != nil {
			lastErr = err
			continue
		}

		// Return the processed application layer data for the caller to decode
		return appData, nil
	}

	return nil, fmt.Errorf("%w: %v", ErrMaxRetries, lastErr)
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
	m.transport.SetTimeout(m.config.Timeout)

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
func (m *Master) waitForResponse() ([]byte, error) {
	m.transport.SetTimeout(m.config.Timeout)
	return m.transport.Receive()
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
	m.SetState(StateError)
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
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	// Application layer decode
	resp, err := al.DecodeResponse(appData)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
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

	// DNP3-013: react to critical IIN bits.
	m.reactToIIN(outstationID, resp.IIN, o)

	return appData, nil
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

			return nil, fmt.Errorf("DLL decode error at offset %d: %w", offset, err)
		}

		// Move past this complete link frame, including the header CRC and each
		// 16-octet payload CRC block.
		offset += frame.EncodedSize(len(dllFrame.Data))

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

// nextSequence returns the current application-layer sequence number without
// advancing it. The caller uses this value as the request's AppControl.Seq
// and calls advanceSequence() only after the request is sent successfully
// (DNP3-008: increment only on successful send).
func (m *Master) nextSequence() uint8 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sequence
}

// advanceSequence advances the sequence number by one (mod 16). It is called
// after a request has been sent successfully so the next request uses the
// next sequence value in the 0-15 stream.
func (m *Master) advanceSequence() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sequence = (m.sequence + 1) % 16
}

// currentSequence returns the current sequence number (alias for nextSequence
// in read-only contexts).
func (m *Master) currentSequence() uint8 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sequence
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
	if m.State() < StateInitialized {
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
	if m.State() < StateInitialized {
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
	if m.State() < StateInitialized {
		return ErrNotConnected
	}

	// Group 1, Variation 1 with range qualifier
	data := buildReadRangeRequest(1, 1, start, stop)
	req := buildRequest(0, al.FuncRead, data)

	return m.sendWithRetry(req, outstationID)
}

// ReadDoubleBinaryInputs reads double-bit binary inputs from an outstation.
func (m *Master) ReadDoubleBinaryInputs(outstationID uint16, start, stop uint16) error {
	if m.State() < StateInitialized {
		return ErrNotConnected
	}

	// Group 3, Variation 1 with range qualifier
	data := buildReadRangeRequest(3, 1, start, stop)
	req := buildRequest(0, al.FuncRead, data)

	return m.sendWithRetry(req, outstationID)
}

// ReadAnalogInputs reads analog inputs from an outstation.
func (m *Master) ReadAnalogInputs(outstationID uint16, start, stop uint16) error {
	if m.State() < StateInitialized {
		return ErrNotConnected
	}

	// Group 30, Variation 1 with range qualifier
	data := buildReadRangeRequest(30, 1, start, stop)
	req := buildRequest(0, al.FuncRead, data)

	return m.sendWithRetry(req, outstationID)
}

// ReadCounters reads counters from an outstation.
func (m *Master) ReadCounters(outstationID uint16, start, stop uint16) error {
	if m.State() < StateInitialized {
		return ErrNotConnected
	}

	// Group 20, Variation 1 with range qualifier
	data := buildReadRangeRequest(20, 1, start, stop)
	req := buildRequest(0, al.FuncRead, data)

	return m.sendWithRetry(req, outstationID)
}

// ReadFrozenCounters reads frozen counters from an outstation.
func (m *Master) ReadFrozenCounters(outstationID uint16, start, stop uint16) error {
	if m.State() < StateInitialized {
		return ErrNotConnected
	}

	// Group 21, Variation 1 with range qualifier
	data := buildReadRangeRequest(21, 1, start, stop)
	req := buildRequest(0, al.FuncRead, data)

	return m.sendWithRetry(req, outstationID)
}

// ReadBinaryOutputStatus reads binary output status from an outstation.
func (m *Master) ReadBinaryOutputStatus(outstationID uint16, start, stop uint16) error {
	if m.State() < StateInitialized {
		return ErrNotConnected
	}

	// Group 10, Variation 1 with range qualifier
	data := buildReadRangeRequest(10, 1, start, stop)
	req := buildRequest(0, al.FuncRead, data)

	return m.sendWithRetry(req, outstationID)
}

// ReadAnalogOutputStatus reads analog output status from an outstation.
func (m *Master) ReadAnalogOutputStatus(outstationID uint16, start, stop uint16) error {
	if m.State() < StateInitialized {
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
