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
	"math"
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

// HasFlag checks if an IIN flag is set.
func (o *Outstation) HasFlag(bit byte) bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return (o.IIN[0]&bit) != 0 || (o.IIN[1]&bit) != 0
}

// Master errors
var (
	ErrNotConnected       = errors.New("master not connected")
	ErrOutstationNotFound = errors.New("outstation not found")
	ErrTimeout            = errors.New("operation timeout")
	ErrMaxRetries         = errors.New("maximum retries exceeded")
	ErrInvalidResponse    = errors.New("invalid response")
	ErrConfirmTimeout     = errors.New("confirmation timeout")
)

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
	m.onError = h
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

	// Perform DNP3 link-layer session establishment
	if err := m.performLinkHandshake(); err != nil {
		m.SetState(StateError)
		return fmt.Errorf("link handshake failed: %w", err)
	}

	m.SetState(StateConnected)
	m.SetState(StateActive)
	return nil
}

// performLinkHandshake performs the DNP3 link-layer handshake sequence.
// This establishes the data link connection before application layer communication.
func (m *Master) performLinkHandshake() error {
	// Step 1: Send Reset Link Stations (Function Code 0)
	// This tells the outstation to reset its link layer state
	if err := m.sendResetLink(); err != nil {
		return fmt.Errorf("reset link failed: %w", err)
	}

	// Wait for acknowledgment from outstation
	// In a full implementation, we would receive and validate the ACK here

	// Small delay to allow outstation to process
	time.Sleep(100 * time.Millisecond)

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
		// layer. Consume that frame here so the next application request reads
		// an application response rather than a stale handshake ACK.
		m.transport.SetTimeout(m.config.Timeout)
		if _, err := m.transport.Receive(); err != nil {
			return fmt.Errorf("failed to receive reset acknowledgment: %w", err)
		}
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
	var lastErr error

	for attempt := 0; attempt < m.config.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(m.config.RetryDelay) * time.Millisecond)
		}

		// Send request with DNP3 protocol layers
		data := req.Encode()

		// Transport layer fragmentation
		fragments := m.fragmenter.Fragmentize(data)

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
				lastErr = err
				continue
			}

		}
		// If CON bit is set, wait for confirmation first
		if req.Control.CON {
			if err := m.waitForConfirmation(req.Control.Seq); err != nil {
				lastErr = err
				continue
			}
		}

		// Wait for response
		resp, err := m.waitForResponse()
		if err != nil {
			lastErr = err
			continue
		}

		// Process response
		_, err = m.processResponse(resp, outstationID)
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
	var lastErr error

	for attempt := 0; attempt < m.config.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(m.config.RetryDelay) * time.Millisecond)
		}

		// Send request with DNP3 protocol layers
		data := req.Encode()

		// Transport layer fragmentation
		fragments := m.fragmenter.Fragmentize(data)

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
				lastErr = err
				continue
			}

		}
		// If CON bit is set, wait for confirmation first
		if req.Control.CON {
			if err := m.waitForConfirmation(req.Control.Seq); err != nil {
				lastErr = err
				continue
			}
		}

		// Wait for response
		resp, err := m.waitForResponse()
		if err != nil {
			lastErr = err
			continue
		}

		// Process the response and get application layer data
		appData, err := m.processResponse(resp, outstationID)
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

// waitForConfirmation waits for an application layer confirmation.
func (m *Master) waitForConfirmation(expectedSeq uint8) error {
	m.transport.SetTimeout(m.config.Timeout)

	for {
		data, err := m.transport.Receive()
		if err != nil {
			return fmt.Errorf("%w: %v", ErrConfirmTimeout, err)
		}

		// Process received bytes
		appData, err := m.processReceivedBytes(data)
		if err != nil {
			continue
		}

		// Decode APDU
		apdu, err := al.Decode(appData)
		if err != nil {
			continue
		}

		// Check if this is a confirmation
		// A confirmation has FuncCode=0 and no IIN (empty data)
		if apdu.FuncCode == al.FuncResponse && len(apdu.Data) == 0 {
			// Sequence should match (confirms the right request)
			if apdu.Control.Seq == expectedSeq {
				return nil
			}

		}

		// If we got a full response before confirmation, that's also acceptable
		// In some implementations, the response IS the confirmation
		if apdu.FuncCode == al.FuncResponse && len(apdu.Data) >= 2 {
			return nil
		}

	}

}

// waitForResponse waits for a response with timeout.
func (m *Master) waitForResponse() ([]byte, error) {
	m.transport.SetTimeout(m.config.Timeout)
	return m.transport.Receive()
}

// processResponse processes a received response and returns the application layer data.
func (m *Master) processResponse(data []byte, outstationID uint16) ([]byte, error) {
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

	// Update outstation IIN
	o, ok := m.GetOutstation(outstationID)
	if ok {
		o.UpdateIIN(resp.IIN.Bytes())
	}

	return appData, nil
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

// buildPollRequest creates a poll request based on poll type.
func buildPollRequest(pollType PollType) []byte {
	// Object headers for different poll types
	switch pollType {
	case PollIntegrity:
		// Group 60, Variation 1 = All static data (Class 0)
		return []byte{60, 1, 0x07, 0x00} // All data with prefix
	case PollClass0:
		// Group 60, Variation 1
		return []byte{60, 1, 0x07, 0x00}

	case PollClass1:
		// Group 2, Variation 1 = Binary events
		return []byte{2, 1, 0x07, 0x00}

	case PollClass2:
		// Group 32, Variation 1 = Analog events
		return []byte{32, 1, 0x07, 0x00}

	case PollClass3:
		// Group 22, Variation 1 = Counter events
		return []byte{22, 1, 0x07, 0x00}

	case PollEvent:
		// All event classes
		return []byte{2, 1, 0x07, 0x00, 32, 1, 0x07, 0x00, 22, 1, 0x07, 0x00}

	default:
		return nil
	}

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
		// Format: code(1), count(1), onTime(4), offTime(4), status(1)
		// Outstation CROB codes: 2=CLOSE(on), 3=OPEN(off), 7=LATCH_ON, 8=LATCH_OFF
		var code uint8
		var count uint8 = 1
		var onTime uint32 = 0
		var offTime uint32 = 0
		var status uint8 = 0

		switch v := value.(type) {
		case bool:
			if v {
				code = 7 // LATCH_ON → Value = true
			} else {
				code = 8 // LATCH_OFF → Value = false
			}
		case uint8:
			code = v
		case uint16:
			code = uint8(v)
		default:
			return nil // Unsupported type
		}

		valueBytes = []byte{
			code,                                                                  // Control code
			count,                                                                 // Count
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
	data := []byte{1, 1, 0x07, 0x00, byte(start), byte(start >> 8), byte(stop), byte(stop >> 8)}
	req := buildRequest(0, al.FuncRead, data)

	return m.sendWithRetry(req, outstationID)
}

// ReadDoubleBinaryInputs reads double-bit binary inputs from an outstation.
func (m *Master) ReadDoubleBinaryInputs(outstationID uint16, start, stop uint16) error {
	if m.State() < StateInitialized {
		return ErrNotConnected
	}

	// Group 3, Variation 1 with range qualifier
	data := []byte{3, 1, 0x07, 0x00, byte(start), byte(start >> 8), byte(stop), byte(stop >> 8)}
	req := buildRequest(0, al.FuncRead, data)

	return m.sendWithRetry(req, outstationID)
}

// ReadAnalogInputs reads analog inputs from an outstation.
func (m *Master) ReadAnalogInputs(outstationID uint16, start, stop uint16) error {
	if m.State() < StateInitialized {
		return ErrNotConnected
	}

	// Group 30, Variation 1 with range qualifier
	data := []byte{30, 1, 0x07, 0x00, byte(start), byte(start >> 8), byte(stop), byte(stop >> 8)}
	req := buildRequest(0, al.FuncRead, data)

	return m.sendWithRetry(req, outstationID)
}

// ReadCounters reads counters from an outstation.
func (m *Master) ReadCounters(outstationID uint16, start, stop uint16) error {
	if m.State() < StateInitialized {
		return ErrNotConnected
	}

	// Group 20, Variation 1 with range qualifier
	data := []byte{20, 1, 0x07, 0x00, byte(start), byte(start >> 8), byte(stop), byte(stop >> 8)}
	req := buildRequest(0, al.FuncRead, data)

	return m.sendWithRetry(req, outstationID)
}

// ReadFrozenCounters reads frozen counters from an outstation.
func (m *Master) ReadFrozenCounters(outstationID uint16, start, stop uint16) error {
	if m.State() < StateInitialized {
		return ErrNotConnected
	}

	// Group 21, Variation 1 with range qualifier
	data := []byte{21, 1, 0x07, 0x00, byte(start), byte(start >> 8), byte(stop), byte(stop >> 8)}
	req := buildRequest(0, al.FuncRead, data)

	return m.sendWithRetry(req, outstationID)
}

// ReadBinaryOutputStatus reads binary output status from an outstation.
func (m *Master) ReadBinaryOutputStatus(outstationID uint16, start, stop uint16) error {
	if m.State() < StateInitialized {
		return ErrNotConnected
	}

	// Group 10, Variation 1 with range qualifier
	data := []byte{10, 1, 0x07, 0x00, byte(start), byte(start >> 8), byte(stop), byte(stop >> 8)}
	req := buildRequest(0, al.FuncRead, data)

	return m.sendWithRetry(req, outstationID)
}

// ReadAnalogOutputStatus reads analog output status from an outstation.
func (m *Master) ReadAnalogOutputStatus(outstationID uint16, start, stop uint16) error {
	if m.State() < StateInitialized {
		return ErrNotConnected
	}

	// Group 40, Variation 1 with range qualifier
	data := []byte{40, 1, 0x07, 0x00, byte(start), byte(start >> 8), byte(stop), byte(stop >> 8)}
	req := buildRequest(0, al.FuncRead, data)

	return m.sendWithRetry(req, outstationID)
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
