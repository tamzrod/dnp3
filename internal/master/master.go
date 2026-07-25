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
	MaxOutstations = 255       // Maximum outstations per master
	MaxRetries     = 3         // Default maximum retries
	DefaultTimeout = 5000       // Default timeout in milliseconds
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
	PollEvent                    // Read all pending events
	PollException                // Read changed data only
	PollClass0                   // Read Class 0 (static) data
	PollClass1                   // Read Class 1 events
	PollClass2                   // Read Class 2 events
	PollClass3                   // Read Class 3 events
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
	ID         uint16        // Outstation address
	Label      string        // Human-readable label
	LastSeen   time.Time     // Last communication time
	State      string        // Operational state
	Unsolicited bool         // Supports unsolicited responses
	IIN         [2]byte      // Last Internal Indication
	mu          sync.RWMutex
}

// NewOutstation creates a new outstation entry.
func NewOutstation(id uint16, label string) *Outstation {
	return &Outstation{
		ID:         id,
		Label:      label,
		LastSeen:   time.Now(),
		State:      "Unknown",
		Unsolicited: false,
		IIN:        [2]byte{0, 0},
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
	ErrNotConnected      = errors.New("master not connected")
	ErrOutstationNotFound = errors.New("outstation not found")
	ErrTimeout           = errors.New("operation timeout")
	ErrMaxRetries       = errors.New("maximum retries exceeded")
	ErrInvalidResponse   = errors.New("invalid response")
)

// Config holds master configuration.
type Config struct {
	MasterAddress uint16   // This master's address
	Timeout      int      // Response timeout in milliseconds
	MaxRetries   int      // Maximum retry attempts
	RetryDelay   int      // Delay between retries in milliseconds
}

// DefaultConfig returns default master configuration.
func DefaultConfig() *Config {
	return &Config{
		MasterAddress: 0xFFFF, // Master uses broadcast or specific address
		Timeout:      DefaultTimeout,
		MaxRetries:   MaxRetries,
		RetryDelay:   100,
	}
}

// Master represents a DNP3 Master Station.
type Master struct {
	config    *Config
	state     State
	outstations map[uint16]*Outstation
	mu        sync.RWMutex
	transport TransportHandler
	onError   ErrorHandler
        fragmenter  *tl.Fragmenter   // Transport layer fragmenter
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
		config:     config,
		state:      StateDisconnected,
		outstations: make(map[uint16]*Outstation),
		fragmenter:   tl.NewFragmenter(),
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

// Connect establishes connection to outstations.
func (m *Master) Connect() error {
	if m.transport == nil {
		return errors.New("transport not configured")
	}
	m.SetState(StateConnecting)
	m.SetState(StateConnected)
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
	if m.State() != StateConnected {
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
					DIR:      true,                  // Master-to-Outstation
					PRM:      true,                  // Primary station
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
		
		// Wait for response
		resp, err := m.waitForResponse()
		if err != nil {
			lastErr = err
			continue
		}
		
		// Process response
		if err := m.processResponse(resp, outstationID); err != nil {
			lastErr = err
			continue
		}
		
		return nil // Success
	}
	
	return fmt.Errorf("%w: %v", ErrMaxRetries, lastErr)
}

// waitForResponse waits for a response with timeout.
func (m *Master) waitForResponse() ([]byte, error) {
	m.transport.SetTimeout(m.config.Timeout)
	return m.transport.Receive()
}

// processResponse processes a received response.
func (m *Master) processResponse(data []byte, outstationID uint16) error {
	resp, err := al.DecodeResponse(data)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}
	
	// Update outstation IIN
	o, ok := m.GetOutstation(outstationID)
	if ok {
		o.UpdateIIN(resp.IIN.Bytes())
	}
	
	return nil
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
	
	// Index: 2 bytes big-endian
	indexBytes := []byte{byte(index >> 8), byte(index & 0xFF)}
	
	// Value encoding depends on type
	var valueBytes []byte
	switch v := value.(type) {
	case bool:
		if v {
			valueBytes = []byte{0x01}
		} else {
			valueBytes = []byte{0x00}
		}
	case uint8:
		valueBytes = []byte{v}
	case uint16:
		valueBytes = []byte{byte(v >> 8), byte(v & 0xFF)}
	case uint32:
		valueBytes = []byte{
			byte(v >> 24), byte(v >> 16),
			byte(v >> 8), byte(v & 0xFF),
		}
	case float32:
		bits := float32ToUint32Bits(v)
		valueBytes = []byte{
			byte(bits >> 24), byte(bits >> 16),
			byte(bits >> 8), byte(bits & 0xFF),
		}
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
