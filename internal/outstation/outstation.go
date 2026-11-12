// Package outstation implements the DNP3 Outstation role.
//
// The Outstation responds to requests from the Master, provides data
// (binary inputs, analog inputs, counters), executes control commands,
// and reports events via unsolicited responses.
//
// Reference: IEEE 1815-2012 Section 8
package outstation

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/tl"
)

// debugControls enables debug logging for control operations (group 12, 41-44).
// Set DNP3_DEBUG=1 environment variable to enable.
var debugControls = os.Getenv("DNP3_DEBUG") == "1"

func debugLog(format string, v ...interface{}) {
	if debugControls {
		fmt.Printf(format+"\n", v...)
	}
}

// DNP3Epoch is the DNP3 time epoch (2000-01-01 00:00:00 UTC)
var DNP3Epoch = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

// State represents the outstation's operational state.
type State int

const (
	StateDown State = iota
	StateInitialized
	StateOperational
	StateSuspended
	StateError
)

// String returns the state name.
func (s State) String() string {
	names := []string{
		"Down",
		"Initialized",
		"Operational",
		"Suspended",
		"Error",
	}
	if s >= 0 && int(s) < len(names) {
		return names[s]
	}
	return "Unknown"
}

// Outstation errors
var (
	ErrNotInitialized  = errors.New("outstation not initialized")
	ErrNotOperational  = errors.New("outstation not operational")
	ErrInvalidRequest  = errors.New("invalid request")
	ErrBufferFull      = errors.New("event buffer full")
	ErrTransportNotSet = errors.New("transport not configured")
	// ErrNotSupported: the request cannot be executed because the configured
	// data handler does not support the required operation. Fail-closed: a
	// non-success response is returned rather than a false success (AO-002 F1).
	ErrNotSupported = errors.New("operation not supported by data handler")
	// SBO errors
	ErrSelectTimeout   = errors.New("select timeout - no operate received")
	ErrNoSelectPending = errors.New("no select pending for this point")
	ErrSelectMismatch  = errors.New("operate does not match select")
)

// ControlStatusError carries a non-success IEEE 1815 command status (the CTRL-01
// per-point status octet) alongside an error. A DataWriter returns it to signal
// that a control was refused for a specific, wire-encodable reason (for example
// OUT_OF_RANGE) so the outstation can echo the object with that status instead of
// a generic error response. The wrapped error preserves the existing fail-closed
// classification; Status is the authoritative wire value.
type ControlStatusError struct {
	Status uint8
	Err    error
}

func (e ControlStatusError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("control refused with status %d", e.Status)
	}
	return e.Err.Error()
}

func (e ControlStatusError) Unwrap() error { return e.Err }

// NewControlStatusError builds a ControlStatusError. status is the IEEE 1815
// command-status octet value (for example 12 = OUT_OF_RANGE).
func NewControlStatusError(status uint8, err error) ControlStatusError {
	return ControlStatusError{Status: status, Err: err}
}

// AsControlStatusError reports whether err (or anything it wraps) is a
// ControlStatusError, returning the carried status.
func AsControlStatusError(err error) (ControlStatusError, bool) {
	var cse ControlStatusError
	if errors.As(err, &cse) {
		return cse, true
	}
	return ControlStatusError{}, false
}

// Freeze errors
var (
	ErrFreezeNotSupported = errors.New("freeze not supported for this point")
)

// DataHandler defines the interface for accessing data points.
type DataHandler interface {
	GetBinaryInputs() []BinaryInput
	GetAnalogInputs() []AnalogInput
	GetCounters() []Counter
	// Output status for Class 0 reads
	GetBinaryOutputs() []BinaryOutput
	GetAnalogOutputs() []AnalogOutput
	// Freeze support (optional methods - implementations may return nil)
	GetFrozenCounters() []Counter
	FreezeCounters(clear bool) error // If clear=true, also zero the counters
}

// Config holds outstation configuration.
type Config struct {
	OutstationAddress uint16 // This outstation's address
	MasterAddress     uint16 // Expected master's address
	Timeout           int    // Response timeout in milliseconds
	MaxEventBuffers   int    // Maximum event buffer size
	Unsolicited       bool   // Enable unsolicited responses
	SBOTimeout        int    // Select-before-operate timeout in milliseconds (default 10000)
}

// DefaultConfig returns default outstation configuration.
func DefaultConfig() *Config {
	return &Config{
		OutstationAddress: 1024, // Common default RTU address
		MasterAddress:     0xFFFF,
		Timeout:           5000,
		MaxEventBuffers:   1000,
		Unsolicited:       false,
		SBOTimeout:        10000, // 10 seconds default for SBO timeout
	}
}

// PendingSelect represents a pending SELECT operation for SBO.
type PendingSelect struct {
	Group     uint8     // Object group
	Variation uint8     // Object variation
	Index     uint16    // Point index
	Value     []byte    // Command value
	Timestamp time.Time // When select was received
}

// BinaryInput represents a binary input data point.
type BinaryInput struct {
	Value   bool   // Current value (true=on, false=off)
	Quality uint8  // Quality flags
	Time    uint64 // Timestamp in DNP3 time format (milliseconds since 2000-01-01)
}

// Quality flags for binary inputs
const (
	BinaryQualityOnline        = 0x01 // Point online
	BinaryQualityRestart       = 0x40 // Device restarted
	BinaryQualityCommLost      = 0x80 // Communication lost
	BinaryQualityDiscontinuous = 0x20 // Data discontinuous
)

// AnalogInput represents an analog input data point.
type AnalogInput struct {
	Value   float64 // Current value
	Quality uint8   // Quality flags
	Time    uint64  // Timestamp in DNP3 time format (milliseconds since 2000-01-01)
}

// AnalogQuality flags
const (
	AnalogQualityOnline    = 0x01 // Point online
	AnalogQualityRestart   = 0x40
	AnalogQualityCommLost  = 0x80
	AnalogQualityOverRange = 0x04
	AnalogQualityRefErr    = 0x02
)

// Counter represents a counter data point.
type Counter struct {
	Value   uint32 // Current count
	Quality uint8  // Quality flags
	Time    uint64 // Timestamp in DNP3 time format (milliseconds since 2000-01-01)
}

// CounterQuality flags
const (
	CounterQualityOnline   = 0x01
	CounterQualityRestart  = 0x40
	CounterQualityCommLost = 0x80
	CounterQualityRollOver = 0x08
)

// BinaryOutput represents a binary output status.
type BinaryOutput struct {
	Value   bool  // Current output value
	Quality uint8 // Quality flags
}

// AnalogOutput represents an analog output status.
type AnalogOutput struct {
	Value   float64 // Current output value
	Quality uint8   // Quality flags
}

// EventClass represents the priority class of an event.
type EventClass int

const (
	Class1 EventClass = iota // Highest priority
	Class2
	Class3
)

// Event represents a DNP3 event.
type Event struct {
	Class   EventClass
	Group   uint8
	Index   uint16
	Value   []byte
	Quality uint8
	Time    time.Time
}

// EventQueue manages event buffers for outstation.
type EventQueue struct {
	mu      sync.Mutex
	buffers map[EventClass][]Event
	maxSize int
}

// NewEventQueue creates a new event queue with the specified maximum size.
func NewEventQueue(maxSize int) *EventQueue {
	return &EventQueue{
		buffers: map[EventClass][]Event{
			Class1: make([]Event, 0, maxSize/3),
			Class2: make([]Event, 0, maxSize/3),
			Class3: make([]Event, 0, maxSize/3),
		},
		maxSize: maxSize,
	}
}

// Add adds an event to the queue.
func (q *EventQueue) Add(event Event) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	classEvents := q.buffers[event.Class]
	if len(classEvents) >= q.maxSize/3 {
		return false // Buffer full for this class
	}

	q.buffers[event.Class] = append(classEvents, event)
	return true
}

// Clear removes all events.
func (q *EventQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	for class := range q.buffers {
		q.buffers[class] = q.buffers[class][:0]
	}
}

// ClearClass removes all events of a specific class.
func (q *EventQueue) ClearClass(class EventClass) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.buffers[class] = q.buffers[class][:0]
}

// GetAll returns all events.
func (q *EventQueue) GetAll() []Event {
	q.mu.Lock()
	defer q.mu.Unlock()

	var all []Event
	for _, events := range q.buffers {
		all = append(all, events...)
	}
	return all
}

// Count returns the total number of events.
func (q *EventQueue) Count() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	count := 0
	for _, events := range q.buffers {
		count += len(events)
	}
	return count
}

// IsFull returns true if any event class buffer is full.
func (q *EventQueue) IsFull() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, events := range q.buffers {
		if len(events) >= q.maxSize/3 {
			return true
		}
	}
	return false
}

// DefaultDataHandler provides default mock data for testing.
type DefaultDataHandler struct {
	mu             sync.RWMutex
	binaryInputs   []BinaryInput
	analogInputs   []AnalogInput
	counters       []Counter
	frozenCounters []Counter
	binaryOutputs  []BinaryOutput
	analogOutputs  []AnalogOutput
}

// NewDefaultDataHandler creates a data handler with mock data.
func NewDefaultDataHandler() *DefaultDataHandler {
	return &DefaultDataHandler{
		binaryInputs: []BinaryInput{
			{Value: true, Quality: BinaryQualityOnline},
			{Value: false, Quality: BinaryQualityOnline},
			{Value: true, Quality: BinaryQualityOnline},
			{Value: false, Quality: BinaryQualityOnline},
			{Value: true, Quality: BinaryQualityOnline},
		},
		analogInputs: []AnalogInput{
			{Value: 100.5, Quality: AnalogQualityOnline},
			{Value: 200.25, Quality: AnalogQualityOnline},
			{Value: 300.75, Quality: AnalogQualityOnline},
			{Value: 400.0, Quality: AnalogQualityOnline},
			{Value: 500.125, Quality: AnalogQualityOnline},
		},
		counters: []Counter{
			{Value: 1000, Quality: CounterQualityOnline},
			{Value: 2000, Quality: CounterQualityOnline},
			{Value: 3000, Quality: CounterQualityOnline},
			{Value: 4000, Quality: CounterQualityOnline},
			{Value: 5000, Quality: CounterQualityOnline},
		},
		frozenCounters: make([]Counter, 5),
		binaryOutputs:  make([]BinaryOutput, 5),
		analogOutputs:  make([]AnalogOutput, 5),
	}
}

func (d *DefaultDataHandler) GetBinaryInputs() []BinaryInput {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.binaryInputs
}

func (d *DefaultDataHandler) GetAnalogInputs() []AnalogInput {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.analogInputs
}

func (d *DefaultDataHandler) GetCounters() []Counter {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.counters
}

// GetBinaryOutputs returns the binary output status points.
func (d *DefaultDataHandler) GetBinaryOutputs() []BinaryOutput {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.binaryOutputs
}

// GetAnalogOutputs returns the analog output status points.
func (d *DefaultDataHandler) GetAnalogOutputs() []AnalogOutput {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.analogOutputs
}

// GetFrozenCounters returns the frozen counters.
func (d *DefaultDataHandler) GetFrozenCounters() []Counter {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.frozenCounters
}

// FreezeCounters freezes the current counter values into the frozen counters.
func (d *DefaultDataHandler) FreezeCounters(clear bool) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Copy current counters to frozen counters
	for i, c := range d.counters {
		d.frozenCounters[i] = Counter{
			Value:   c.Value,
			Quality: c.Quality,
		}
	}

	// If clear is true, zero the counters after freezing
	if clear {
		for i := range d.counters {
			d.counters[i] = Counter{
				Value:   0,
				Quality: CounterQualityOnline,
			}
		}
	}

	return nil
}

// WriteBinaryOutput handles writing a binary output (CROB).
func (d *DefaultDataHandler) WriteBinaryOutput(index uint16, crob *CROB) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Ensure binary outputs slice is large enough
	if int(index) >= len(d.binaryOutputs) {
		// Expand the slice
		newOutputs := make([]BinaryOutput, index+1)
		copy(newOutputs, d.binaryOutputs)
		d.binaryOutputs = newOutputs
	}

	// Update the binary output based on the CROB operation type (IEEE 1815
	// A.8.1 low-nibble enumeration).
	switch crob.Code {
	case 0x00: // NUL (no operation)
		// No action
	case 0x01: // Pulse On — energize
		d.binaryOutputs[index] = BinaryOutput{Value: true, Quality: BinaryQualityOnline}
	case 0x02: // Pulse Off — de-energize
		d.binaryOutputs[index] = BinaryOutput{Value: false, Quality: BinaryQualityOnline}
	case 0x03: // Latch On
		d.binaryOutputs[index] = BinaryOutput{Value: true, Quality: BinaryQualityOnline}
	case 0x04: // Latch Off
		d.binaryOutputs[index] = BinaryOutput{Value: false, Quality: BinaryQualityOnline}
	default:
		return fmt.Errorf("unsupported CROB code: 0x%02X", crob.Code)
	}

	return nil
}

// WriteAnalogOutput handles writing an analog output.
func (d *DefaultDataHandler) WriteAnalogOutput(index uint16, value interface{}, variation uint8) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Ensure analog outputs slice is large enough
	if int(index) >= len(d.analogOutputs) {
		// Expand the slice
		newOutputs := make([]AnalogOutput, index+1)
		copy(newOutputs, d.analogOutputs)
		d.analogOutputs = newOutputs
	}

	// Convert value to float64 and store
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
		return fmt.Errorf("unsupported analog output type: %T", value)
	}

	d.analogOutputs[index] = AnalogOutput{
		Value:   floatValue,
		Quality: AnalogQualityOnline,
	}

	return nil
}

// TransportHandler defines the interface for sending/receiving data.
type TransportHandler interface {
	Send(data []byte) error
	Receive() ([]byte, error)
	SetTimeout(ms int)
}

// RequestHandler defines the callback for handling requests.
type RequestHandler func(req *al.APDU) (*al.APDU, error)

// Outstation represents a DNP3 Outstation.
type Outstation struct {
	config      *Config
	state       State
	data        DataHandler
	mu          sync.RWMutex
	transport   TransportHandler
	onRequest   RequestHandler
	appSeq      uint8  // Application sequence number
	unsolicited bool   // Unsolicited responses enabled
	iin         al.IIN // Internal Indication
	fragmenter  *tl.Fragmenter
	reassembler *tl.Reassembler
	// SBO state tracking
	pendingSelects map[uint16]*PendingSelect // Index -> PendingSelect
	selectMu       sync.Mutex                // Separate mutex for SBO operations
	// Event queue
	eventQueue *EventQueue
}

// NewOutstation creates a new Outstation.
func NewOutstation(config *Config) *Outstation {
	if config == nil {
		config = DefaultConfig()
	}
	return &Outstation{
		config:         config,
		state:          StateDown,
		data:           NewDefaultDataHandler(),
		appSeq:         0,
		fragmenter:     tl.NewFragmenter(),
		reassembler:    tl.NewReassembler(),
		pendingSelects: make(map[uint16]*PendingSelect),
		eventQueue:     NewEventQueue(config.MaxEventBuffers),
	}
}

// SetDataHandler sets the data handler.
func (o *Outstation) SetDataHandler(dh DataHandler) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.data = dh
}

// SetTransport sets the transport handler.
func (o *Outstation) SetTransport(t TransportHandler) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.transport = t
}

// SetRequestHandler sets the request handler callback.
func (o *Outstation) SetRequestHandler(rh RequestHandler) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.onRequest = rh
}

// State returns the current outstation state.
func (o *Outstation) State() State {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.state
}

// SetState updates the outstation state.
func (o *Outstation) SetState(state State) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.state = state
}

// IIN returns the current Internal Indication.
func (o *Outstation) IIN() al.IIN {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.iin
}

// Initialize performs outstation initialization.
func (o *Outstation) Initialize() error {
	if o.State() != StateDown {
		return fmt.Errorf("already initialized")
	}
	o.SetState(StateInitialized)
	return nil
}

// Start begins outstation operation.
func (o *Outstation) Start() error {
	if o.State() < StateInitialized {
		return ErrNotInitialized
	}
	o.SetState(StateOperational)
	return nil
}

// Stop halts outstation operation.
func (o *Outstation) Stop() error {
	o.SetState(StateDown)
	// Clear all pending selects on shutdown
	o.selectMu.Lock()
	o.pendingSelects = make(map[uint16]*PendingSelect)
	o.selectMu.Unlock()
	return nil
}

// ClearPendingSelects clears all pending select operations.
// This should be called when transitioning to a new master connection.
func (o *Outstation) ClearPendingSelects() {
	o.selectMu.Lock()
	o.pendingSelects = make(map[uint16]*PendingSelect)
	o.selectMu.Unlock()
}

// PendingSelectCount returns the number of pending select operations.
func (o *Outstation) PendingSelectCount() int {
	o.selectMu.Lock()
	defer o.selectMu.Unlock()
	return len(o.pendingSelects)
}

// ResponseCacheEntry represents a cached response for duplicate detection.
type ResponseCacheEntry struct {
	Response  []byte
	Timestamp time.Time
}

// ResponseCache manages cached responses for duplicate detection.
type ResponseCache struct {
	mu      sync.Mutex
	entries map[uint8]*ResponseCacheEntry // Keyed by sequence number
	maxAge  time.Duration
}

// NewResponseCache creates a new response cache.
func NewResponseCache(maxAge time.Duration) *ResponseCache {
	return &ResponseCache{
		entries: make(map[uint8]*ResponseCacheEntry),
		maxAge:  maxAge,
	}
}

// Add adds a response to the cache.
func (c *ResponseCache) Add(seq uint8, response []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[seq] = &ResponseCacheEntry{
		Response:  response,
		Timestamp: time.Now(),
	}
}

// Get retrieves a cached response if it exists and is not expired.
func (c *ResponseCache) Get(seq uint8) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.entries[seq]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Since(entry.Timestamp) > c.maxAge {
		delete(c.entries, seq)
		return nil, false
	}

	return entry.Response, true
}

// Cleanup removes expired entries.
func (c *ResponseCache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for seq, entry := range c.entries {
		if now.Sub(entry.Timestamp) > c.maxAge {
			delete(c.entries, seq)
		}
	}
}

// Clear removes all entries.
func (c *ResponseCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[uint8]*ResponseCacheEntry)
}

// GenerateEvent adds a new event to the event queue.
// Returns false if the event was not added due to buffer overflow.
func (o *Outstation) GenerateEvent(class EventClass, group uint8, index uint16, value []byte, quality uint8) bool {
	event := Event{
		Class:   class,
		Group:   group,
		Index:   index,
		Value:   value,
		Quality: quality,
		Time:    time.Now(),
	}

	added := o.eventQueue.Add(event)

	// Update IIN if buffer is full
	if o.eventQueue.IsFull() {
		o.mu.Lock()
		o.iin.BufferOverflow = true
		o.mu.Unlock()
	}

	return added
}

// EventCount returns the current number of events in the queue.
func (o *Outstation) EventCount() int {
	return o.eventQueue.Count()
}

// ClearEvents removes all events from the queue.
func (o *Outstation) ClearEvents() {
	o.eventQueue.Clear()
	o.mu.Lock()
	o.iin.BufferOverflow = false
	o.mu.Unlock()
}

// HasEvents returns true if there are events in the queue.
func (o *Outstation) HasEvents() bool {
	return o.eventQueue.Count() > 0
}

// GetEvents returns all events in the queue.
func (o *Outstation) GetEvents() []Event {
	return o.eventQueue.GetAll()
}

// Run starts the outstation's main loop.
// This is a convenience method that uses context.Background().
func (o *Outstation) Run() error {
	return o.RunWithContext(context.Background())
}

// RunWithContext starts the outstation's main loop.
// This version supports context cancellation for graceful shutdown.
func (o *Outstation) RunWithContext(ctx context.Context) error {
	o.mu.RLock()
	transport := o.transport
	o.mu.RUnlock()

	if transport == nil {
		return ErrTransportNotSet
	}

	for {
		select {
		case <-ctx.Done():
			// Context cancelled - perform cleanup
			o.cleanup()
			return ctx.Err()
		default:
		}

		// Check if we've been stopped
		if o.State() == StateDown {
			o.cleanup()
			return nil
		}

		// Receive data from transport
		data, err := transport.Receive()
		if err != nil {
			continue
		}

		// Reassemble transport fragments
		msg, err := o.reassembleMessage(data)
		if err != nil {
			o.sendErrorResponse(0, err)
			continue
		}
		if msg == nil {
			continue
		}

		// Decode APDU
		req, err := al.Decode(msg)
		if err != nil {
			o.sendErrorResponse(0, err)
			continue
		}

		// Check if confirmation is required (CON bit set)
		needsConfirm := req.Control.CON

		// Process request
		resp, err := o.ProcessRequest(req)
		if err != nil {
			o.sendErrorResponse(req.Control.Seq, err)
			// Still send confirmation even on error (unless it's a critical error)
			if needsConfirm {
				o.sendConfirmation(req.Control.Seq)
			}
			continue
		}

		// Send response with DLL framing
		if resp != nil {
			respData := o.buildTransportFragments(resp)
			for _, frag := range respData {
				// Wrap in DLL frame. A solicited application response is user
				// data and must be a primary frame (PRM=1) in both directions;
				// unconfirmed user data carries no FCB/FCV.
				dllFrame := &frame.Frame{
					Control: frame.Control{
						DIR:      false, // Outstation-to-Master
						PRM:      true,  // Primary: carries user data
						FuncCode: frame.FuncUnconfirmedUserData,
					},
					DestAddr: o.config.MasterAddress,
					SrcAddr:  o.config.OutstationAddress,
					Data:     frag,
				}
				dllEncoded, err := frame.Encode(dllFrame)
				if err != nil {
					continue
				}
				if err := transport.Send(dllEncoded); err != nil {
					continue
				}
			}
		}

		// Send confirmation if required
		if needsConfirm {
			o.sendConfirmation(req.Control.Seq)
		}
	}
}

// sendConfirmation sends an application layer confirmation.
func (o *Outstation) sendConfirmation(seq uint8) {
	o.mu.RLock()
	transport := o.transport
	o.mu.RUnlock()

	if transport == nil {
		return
	}

	// Create confirmation APDU
	confirm := al.NewConfirm(seq)
	fragments := o.buildTransportFragments(confirm)

	for _, frag := range fragments {
		// An application confirmation is user data; frame it primary too.
		dllFrame := &frame.Frame{
			Control: frame.Control{
				DIR:      false, // Outstation-to-Master
				PRM:      true,  // Primary: carries user data
				FuncCode: frame.FuncUnconfirmedUserData,
			},
			DestAddr: o.config.MasterAddress,
			SrcAddr:  o.config.OutstationAddress,
			Data:     frag,
		}
		dllEncoded, err := frame.Encode(dllFrame)
		if err != nil {
			continue
		}
		_ = transport.Send(dllEncoded)
	}
}

// reassembleMessage reassembles transport fragments into a complete message.
// It first decodes DLL frames, then extracts and reassembles TL fragments.
func (o *Outstation) reassembleMessage(data []byte) ([]byte, error) {
	o.reassembler.Reset()

	offset := 0
	for offset < len(data) {
		if offset >= len(data) {
			break
		}
		remaining := data[offset:]
		if len(remaining) < 1 {
			break
		}

		// Decode DLL frame first
		dllFrame, err := frame.Decode(remaining)
		if err != nil {
			offset++
			continue
		}

		// Move past this complete link frame, including the header CRC and each
		// 16-octet payload CRC block.
		offset += frame.EncodedSize(len(dllFrame.Data))

		// Skip frames not directed to us
		if dllFrame.DestAddr != o.config.OutstationAddress {
			continue
		}

		// Handle link-layer control frames (not user data)
		// These require acknowledgment at the link layer
		if dllFrame.Control.PRM && dllFrame.Control.FuncCode == frame.FuncResetLinkStations {
			// Reset Link Stations - respond with ACK
			o.sendLinkAck(dllFrame.SrcAddr)
			continue
		}

		if dllFrame.Control.PRM && dllFrame.Control.FuncCode == frame.FuncResetLinkStatus {
			// Link Status Request - respond with Link Status
			o.sendLinkStatus(dllFrame.SrcAddr)
			continue
		}

		// Skip non-user-data frames. Only primary user-data frames carry an
		// application fragment: confirmed (wire 3) and unconfirmed (wire 4) user
		// data are both valid from a master; all other primary function codes are
		// link-control frames already handled above or unsupported.
		if !dllFrame.Control.PRM ||
			(dllFrame.Control.FuncCode != frame.FuncConfirmedUserData &&
				dllFrame.Control.FuncCode != frame.FuncUnconfirmedUserData) {
			continue
		}

		// Extract TL data from DLL frame
		tlData := dllFrame.Data
		if len(tlData) < 1 {
			continue
		}

		// Decode TL fragment
		frag, err := tl.DecodeFragment(tlData)
		if err != nil {
			continue
		}

		msg, err := o.reassembler.Push(frag)
		if err != nil {
			continue
		}

		if msg != nil {
			return msg, nil
		}
	}

	return nil, nil
}

// buildTransportFragments builds transport fragments from an APDU.
func (o *Outstation) buildTransportFragments(apdu *al.APDU) [][]byte {
	fragmenter := tl.NewFragmenter()
	apduData := apdu.Encode()
	fragments := fragmenter.Fragmentize(apduData)

	result := make([][]byte, len(fragments))
	for i, frag := range fragments {
		result[i] = tl.EncodeFragment(frag)
	}
	return result
}

// ProcessRequest processes an incoming request and returns a response.
func (o *Outstation) ProcessRequest(req *al.APDU) (*al.APDU, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Increment sequence number
	o.appSeq = (o.appSeq + 1) % 16

	// Check for custom request handler
	if o.onRequest != nil {
		return o.onRequest(req)
	}

	// Default request processing
	var resp *al.APDU
	var err error

	switch req.FuncCode {
	case al.FuncRead:
		resp, err = o.handleRead(req)
	case al.FuncWrite:
		resp, err = o.handleWrite(req)
	case al.FuncSelect:
		resp, err = o.handleSelect(req)
	case al.FuncOperate:
		resp, err = o.handleOperate(req)
	case al.FuncDirectOperate:
		resp, err = o.handleDirectOperate(req)
	case al.FuncDirectOperateNoResp:
		resp, err = o.handleDirectOperateNoResp(req)
	case al.FuncEnableUnsolicited:
		resp, err = o.handleEnableUnsolicited(req)
	case al.FuncDisableUnsolicited:
		resp, err = o.handleDisableUnsolicited(req)
	case al.FuncFreeze:
		resp, err = o.handleFreeze(req)
	case al.FuncFreezeClear:
		resp, err = o.handleFreezeClear(req)
	case al.FuncTimeSync:
		resp, err = o.handleTimeSync(req)
	case al.FuncDelayMeasurement:
		resp, err = o.handleDelayMeasurement(req)
	case al.FuncRecordCurrentTime:
		resp, err = o.handleRecordCurrentTime(req)
	default:
		// DNP3-080: unsupported function codes are reported back to the master
		// as a clear failure. The run loop maps this error to a FuncResponse
		// carrying IIN.FuncUnknown (function code cannot be processed). Do NOT
		// mutate the persistent o.iin here — FuncUnknown is per-request and is
		// applied by sendErrorResponse, so it does not leak into later replies.
		err = fmt.Errorf("unsupported function code: %d (0x%02X)", req.FuncCode, req.FuncCode)
	}

	// If we have a response, prepend IIN bytes
	if resp != nil && err == nil {
		iinBytes := o.iin.Bytes()
		// Prepend IIN to data
		newData := make([]byte, 2+len(resp.Data))
		newData[0] = iinBytes[0]
		newData[1] = iinBytes[1]
		if len(resp.Data) > 0 {
			copy(newData[2:], resp.Data)
		}
		resp.Data = newData

		// Per IEEE 1815, a solicited response echoes the application sequence
		// number of the request it answers (DNP3-010). Override the
		// per-handler counter value with the request's sequence.
		if !resp.Control.UNS {
			resp.Control.Seq = req.Control.Seq
		}
	}

	return resp, err
}

// handleRead handles READ requests.
func (o *Outstation) handleRead(req *al.APDU) (*al.APDU, error) {
	resp := &al.APDU{
		Control: al.AppControl{
			FIR: true,
			FIN: true,
			CON: false,
			UNS: false,
			Seq: o.appSeq,
		},
		FuncCode: al.FuncResponse,
		Data:     o.buildReadResponse(req.Data),
	}
	return resp, nil
}

// normalizeStaticVariation implements IEEE 1815 "variation 0 = default
// variation" for the static (non-event) data groups. The v0 outstation's
// default static variation is Variation 1 for every supported group, so a
// request for G1V0 / G10V0 / G20V0 / G30V0 / G40V0 is served as V1. Event
// groups (2/21/31) are not normalized here; their empty-buffer stub behavior is
// unchanged. This lets the master's MEXT-015 multi-group integrity read (which
// requests the default variation per group) succeed against the real
// outstation, not just the simulator.
func normalizeStaticVariation(group, variation uint8) uint8 {
	if variation == 0 {
		switch group {
		case 1, 10, 20, 30, 40:
			return 1
		}
	}
	return variation
}

// requestObjectHeaderSize returns the wire size in octets of the request
// object header beginning at offset, determined by its qualifier. Returns 0 if
// the qualifier octet is not yet available.
func requestObjectHeaderSize(data []byte, offset int) int {
	if offset+3 > len(data) {
		return 0
	}
	switch data[offset+2] {
	case 0x06: // all-objects: no range field
		return 3
	case 0x07, 0x17: // count8 / index8: 1-octet count
		return 4
	case 0x00: // start/stop: 1-octet start + 1-octet stop
		return 5
	case 0x27: // count16: 2-octet count
		return 5
	case 0x28: // range16: 2-octet start + 2-octet stop
		return 7
	default:
		return 4
	}
}

// buildReadResponse builds data for a READ response.
func (o *Outstation) buildReadResponse(requestData []byte) []byte {
	var result []byte

	// If no specific data requested, return all static data
	if len(requestData) < 4 {
		result = o.buildAllStaticData()
		return result
	}

	// classEventsRequested is true when the master asked for Class 1/2/3 event
	// data (G60 V2/V3/V4). The v0 outstation has an empty event buffer stub
	// (DNP3-088): such a poll deterministically returns no event objects. It
	// must NOT fall back to "all static data" — an empty event-class response
	// is the correct, protocol-valid reply.
	classEventsRequested := false

	// Parse object headers. Each header's size depends on its qualifier
	// (IEEE 1815): all-objects (0x06) is 3 octets; start/stop (0x00) is 5 octets
	// (start(1)+stop(1)); count8/index8 (0x07/0x17) carry a 1-octet count (4);
	// count16 (0x27) a 2-octet count (5); range16 (0x28) a 2-octet start/stop (7).
	offset := 0
	for offset+3 <= len(requestData) {
		group := requestData[offset]
		variation := requestData[offset+1]

		switch group {
		case 0: // Null qualifier - all data
			result = append(result, o.buildAllStaticData()...)
		case 1: // Binary Input (static - no timestamp)
			result = append(result, o.buildBinaryInputData(normalizeStaticVariation(group, variation))...)
		case 2: // Binary Input Event (with timestamp)
			result = append(result, o.buildBinaryInputEventData(variation)...)
		case 10: // Binary Output (static - no timestamp)
			result = append(result, o.buildBinaryOutputData(normalizeStaticVariation(group, variation))...)
		case 20: // Counter (static - no timestamp)
			result = append(result, o.buildCounterData(normalizeStaticVariation(group, variation))...)
		case 21: // Counter Event (with timestamp)
			result = append(result, o.buildCounterEventData(variation)...)
		case 30: // Analog Input (static - no timestamp)
			result = append(result, o.buildAnalogInputData(normalizeStaticVariation(group, variation))...)
		case 31: // Analog Input Event (with timestamp)
			result = append(result, o.buildAnalogInputEventData(variation)...)
		case 40: // Analog Output (static - no timestamp)
			result = append(result, o.buildAnalogOutputData(normalizeStaticVariation(group, variation))...)
		case 60: // Class data
			switch variation {
			case 1: // Class 0 (integrity) — all static data
				result = append(result, o.buildAllStaticData()...)
			case 2, 3, 4: // Class 1/2/3 events — empty stub (DNP3-088)
				classEventsRequested = true
			default:
				// Unknown class variation — treat as Class 0.
				result = append(result, o.buildAllStaticData()...)
			}
		}

		size := requestObjectHeaderSize(requestData, offset)
		if size == 0 || offset+size > len(requestData) {
			break
		}
		offset += size
	}

	// Only fall back to "all static data" when nothing was recognized AND the
	// master did not explicitly request an event class. An explicit event-class
	// poll with an empty buffer returns an empty (object-less) response.
	if len(result) == 0 && !classEventsRequested {
		result = o.buildAllStaticData()
	}

	return result
}

// buildAllStaticData builds all static data for Class 0.
func (o *Outstation) buildAllStaticData() []byte {
	var result []byte
	result = append(result, o.buildBinaryInputData(1)...)
	result = append(result, o.buildAnalogInputData(1)...)
	result = append(result, o.buildCounterData(1)...)
	result = append(result, o.buildBinaryOutputData(1)...)
	result = append(result, o.buildAnalogOutputData(1)...)
	return result
}

// buildBinaryInputData builds binary input data.
func (o *Outstation) buildBinaryInputData(variation uint8) []byte {
	var result []byte
	data := o.data.GetBinaryInputs()

	if len(data) == 0 {
		return result
	}

	// Object header: group, variation, qualifier, count - OUTSIDE loop
	result = append(result, 1)         // Group 1
	result = append(result, variation) // Variation
	if variation == 1 {
		// Variation 1 is packed binary state; qualifier 0x07 is an 8-bit
		// point count and the values follow as packed bits.
		result = append(result, 0x07, byte(len(data)))
		packed := make([]byte, (len(data)+7)/8)
		for i, bi := range data {
			if bi.Value {
				packed[i/8] |= 1 << uint(i%8)
			}
		}
		return append(result, packed...)
	}
	result = append(result, 0x17)            // Indexed qualifier (UINT8_CNT_UINT8_INDEX)
	result = append(result, byte(len(data))) // Count

	for i, bi := range data {
		// Index (1 byte, UINT8_CNT_UINT8_INDEX prefix)
		result = append(result, byte(i&0xFF))

		// Value based on variation
		switch variation {
		case 1: // Binary Input with flags (no timestamp for static)
			val := byte(0)
			if bi.Value {
				val = 0x80
			}
			result = append(result, val|bi.Quality)
		case 2: // Binary Input without flags (no timestamp for static)
			if bi.Value {
				result = append(result, 0x01)
			} else {
				result = append(result, 0x00)
			}
		}
	}

	return result
}

// buildBinaryInputEventData builds binary input event data with timestamps.
func (o *Outstation) buildBinaryInputEventData(variation uint8) []byte {
	var result []byte
	data := o.data.GetBinaryInputs()

	if len(data) == 0 {
		return result
	}

	// Object header for event data
	result = append(result, 2)               // Group 2 (Binary Input Event)
	result = append(result, variation)       // Variation
	result = append(result, 0x00)            // Qualifier: index
	result = append(result, byte(len(data))) // Count

	for i, bi := range data {
		// Index (2 bytes, big-endian)
		result = append(result, byte(i>>8), byte(i&0xFF))

		switch variation {
		case 1: // Binary Input Event with time (8 bytes)
			val := byte(0)
			if bi.Value {
				val = 0x80
			}
			result = append(result, val|bi.Quality)
			// Time - 8 bytes (big-endian uint64)
			timeValue := bi.Time
			if timeValue == 0 {
				// Use current time if not set
				timeValue = uint64(time.Now().Sub(DNP3Epoch).Milliseconds())
			}
			result = append(result,
				byte(timeValue>>56), byte(timeValue>>48),
				byte(timeValue>>40), byte(timeValue>>32),
				byte(timeValue>>24), byte(timeValue>>16),
				byte(timeValue>>8), byte(timeValue))
		}
	}

	return result
}

// buildAnalogInputData builds analog input data.
func (o *Outstation) buildAnalogInputData(variation uint8) []byte {
	var result []byte
	data := o.data.GetAnalogInputs()

	if len(data) == 0 {
		return result
	}

	// Object header - OUTSIDE loop
	result = append(result, 30)        // Group 30
	result = append(result, variation) // Variation
	if variation == 1 {
		// Variation 1 is a signed 32-bit value with flags. Qualifier 0x07
		// carries an 8-bit count and sequential points without indexes. The
		// flags octet is first, then the little-endian value (IEEE 1815;
		// opendnp3 3.1.2 Group30Var1 = (flags, value32)).
		result = append(result, 0x07, byte(len(data)))
		for _, ai := range data {
			value := int32(ai.Value)
			result = append(result, ai.Quality, byte(value), byte(value>>8), byte(value>>16), byte(value>>24))
		}
		return result
	}
	result = append(result, 0x17)            // Indexed qualifier (UINT8_CNT_UINT8_INDEX)
	result = append(result, byte(len(data))) // Count

	for i, ai := range data {
		// Index (1 byte, UINT8_CNT_UINT8_INDEX prefix)
		result = append(result, byte(i&0xFF))

		// Value based on variation
		switch variation {
		case 1: // 32-bit float with flags
			bits := float64ToUint32Bits(ai.Value)
			result = append(result,
				byte(bits>>24), byte(bits>>16),
				byte(bits>>8), byte(bits))
			result = append(result, ai.Quality)
		case 2: // 16-bit integer with flags
			val := int16(ai.Value)
			result = append(result, byte(val>>8), byte(val))
			result = append(result, ai.Quality)
		case 3: // 32-bit integer with flags
			val := int32(ai.Value)
			result = append(result,
				byte(val>>24), byte(val>>16),
				byte(val>>8), byte(val))
			result = append(result, ai.Quality)
		case 5: // 32-bit float without flags
			bits := float64ToUint32Bits(ai.Value)
			result = append(result,
				byte(bits>>24), byte(bits>>16),
				byte(bits>>8), byte(bits))
		}
	}

	return result
}

// buildCounterData builds counter data.
func (o *Outstation) buildCounterData(variation uint8) []byte {
	var result []byte
	data := o.data.GetCounters()

	if len(data) == 0 {
		return result
	}

	// Object header - OUTSIDE loop
	result = append(result, 20)        // Group 20
	result = append(result, variation) // Variation
	if variation == 1 {
		// Variation 1 is an unsigned 32-bit counter with flags. Qualifier
		// 0x07 carries an 8-bit count and sequential points without indexes.
		// The flags octet is first, then the little-endian value (IEEE 1815;
		// opendnp3 3.1.2 Group20Var1 = (flags, count32)).
		result = append(result, 0x07, byte(len(data)))
		for _, c := range data {
			result = append(result, c.Quality, byte(c.Value), byte(c.Value>>8), byte(c.Value>>16), byte(c.Value>>24))
		}
		return result
	}
	result = append(result, 0x17)            // Indexed qualifier (UINT8_CNT_UINT8_INDEX)
	result = append(result, byte(len(data))) // Count

	for i, c := range data {
		// Index (1 byte, UINT8_CNT_UINT8_INDEX prefix)
		result = append(result, byte(i&0xFF))

		// Value based on variation
		switch variation {
		case 1: // 32-bit counter with flags
			result = append(result,
				byte(c.Value>>24), byte(c.Value>>16),
				byte(c.Value>>8), byte(c.Value))
			result = append(result, c.Quality)
		case 5: // 16-bit counter with flags
			val := uint16(c.Value)
			result = append(result, byte(val>>8), byte(val))
			result = append(result, c.Quality)
		case 6: // 32-bit counter without flags
			result = append(result,
				byte(c.Value>>24), byte(c.Value>>16),
				byte(c.Value>>8), byte(c.Value))
		}
	}

	return result
}

// buildAnalogInputEventData builds analog input event data with timestamps.
func (o *Outstation) buildAnalogInputEventData(variation uint8) []byte {
	var result []byte
	data := o.data.GetAnalogInputs()

	if len(data) == 0 {
		return result
	}

	// Object header for event data
	result = append(result, 31)              // Group 31 (Analog Input Event)
	result = append(result, variation)       // Variation
	result = append(result, 0x00)            // Qualifier: index
	result = append(result, byte(len(data))) // Count

	for i, ai := range data {
		// Index (2 bytes, big-endian)
		result = append(result, byte(i>>8), byte(i&0xFF))

		switch variation {
		case 1: // Analog Input Event with time - 32-bit float
			bits := float64ToUint32Bits(ai.Value)
			result = append(result,
				byte(bits>>24), byte(bits>>16),
				byte(bits>>8), byte(bits))
			result = append(result, ai.Quality)
			// Time - 8 bytes (big-endian uint64)
			timeValue := ai.Time
			if timeValue == 0 {
				timeValue = uint64(time.Now().Sub(DNP3Epoch).Milliseconds())
			}
			result = append(result,
				byte(timeValue>>56), byte(timeValue>>48),
				byte(timeValue>>40), byte(timeValue>>32),
				byte(timeValue>>24), byte(timeValue>>16),
				byte(timeValue>>8), byte(timeValue))
		}
	}

	return result
}

// buildCounterEventData builds counter event data with timestamps.
func (o *Outstation) buildCounterEventData(variation uint8) []byte {
	var result []byte
	data := o.data.GetCounters()

	if len(data) == 0 {
		return result
	}

	// Object header for event data
	result = append(result, 21)              // Group 21 (Counter Event)
	result = append(result, variation)       // Variation
	result = append(result, 0x00)            // Qualifier: index
	result = append(result, byte(len(data))) // Count

	for i, c := range data {
		// Index (2 bytes, big-endian)
		result = append(result, byte(i>>8), byte(i&0xFF))

		switch variation {
		case 1: // Counter Event with time - 32-bit counter
			result = append(result,
				byte(c.Value>>24), byte(c.Value>>16),
				byte(c.Value>>8), byte(c.Value))
			result = append(result, c.Quality)
			// Time - 8 bytes (big-endian uint64)
			timeValue := c.Time
			if timeValue == 0 {
				timeValue = uint64(time.Now().Sub(DNP3Epoch).Milliseconds())
			}
			result = append(result,
				byte(timeValue>>56), byte(timeValue>>48),
				byte(timeValue>>40), byte(timeValue>>32),
				byte(timeValue>>24), byte(timeValue>>16),
				byte(timeValue>>8), byte(timeValue))
		}
	}

	return result
}

// buildBinaryOutputData builds binary output status data (Group 10).
func (o *Outstation) buildBinaryOutputData(variation uint8) []byte {
	var result []byte
	data := o.data.GetBinaryOutputs()

	if len(data) == 0 {
		return result
	}

	// Object header - OUTSIDE loop
	result = append(result, 10)              // Group 10
	result = append(result, variation)       // Variation
	result = append(result, 0x07)            // Qualifier: UINT8_CNT (sequential)
	result = append(result, byte(len(data))) // Count

	for _, bo := range data {
		// Value based on variation
		switch variation {
		case 1: // Binary Output with flags
			val := byte(0)
			if bo.Value {
				val = 0x80
			}
			result = append(result, val|bo.Quality)
		case 2: // Binary Output without flags
			if bo.Value {
				result = append(result, 0x01)
			} else {
				result = append(result, 0x00)
			}
		}
	}

	return result
}

// buildAnalogOutputData builds analog output status data (Group 40).
func (o *Outstation) buildAnalogOutputData(variation uint8) []byte {
	var result []byte
	data := o.data.GetAnalogOutputs()

	if len(data) == 0 {
		return result
	}

	// Object header - OUTSIDE loop
	result = append(result, 40)              // Group 40
	result = append(result, variation)       // Variation
	result = append(result, 0x07)            // Qualifier: UINT8_CNT (sequential)
	result = append(result, byte(len(data))) // Count

	for _, ao := range data {
		// Value based on variation
		switch variation {
		case 1: // 32-bit float with flags
			bits := float64ToUint32Bits(ao.Value)
			result = append(result,
				byte(bits>>24), byte(bits>>16),
				byte(bits>>8), byte(bits))
			result = append(result, ao.Quality)
		case 2: // 16-bit integer with flags
			val := int16(ao.Value)
			result = append(result, byte(val>>8), byte(val))
			result = append(result, ao.Quality)
		}
	}

	return result
}

// float64ToUint32Bits converts float64 to uint32 bits.
func float64ToUint32Bits(f float64) uint32 {
	// Clamp to float32 range and convert
	if f > math.MaxFloat32 {
		f = math.MaxFloat32
	}
	if f < -math.MaxFloat32 {
		f = -math.MaxFloat32
	}
	return math.Float32bits(float32(f))
}

// parseControlObjectHeader parses the leading object header of a control or
// write request and returns the group, variation, object count, the first
// object's index (LSB-first) and the offset at which the first object's value
// begins. Supported v0 qualifiers: 0x28 (16-bit count + 16-bit index, used by
// the conformant control path) and 0x00 (8-bit count + 16-bit index).
func parseControlObjectHeader(data []byte) (group, variation uint8, count int, index uint16, valueOffset int, ok bool) {
	if len(data) < 4 {
		return 0, 0, 0, 0, 0, false
	}
	group = data[0]
	variation = data[1]
	switch data[2] {
	case 0x28: // 16-bit count (LSB) + 16-bit index (LSB)
		if len(data) < 7 {
			return 0, 0, 0, 0, 0, false
		}
		count = int(data[3]) | int(data[4])<<8
		index = uint16(data[5]) | uint16(data[6])<<8
		valueOffset = 7
	case 0x00: // 8-bit count + 16-bit index (LSB)
		if len(data) < 6 {
			return 0, 0, 0, 0, 0, false
		}
		count = int(data[3])
		index = uint16(data[4]) | uint16(data[5])<<8
		valueOffset = 6
	default:
		return 0, 0, 0, 0, 0, false
	}
	return group, variation, count, index, valueOffset, true
}

// handleWrite handles WRITE requests.
func (o *Outstation) handleWrite(req *al.APDU) (*al.APDU, error) {
	// Parse write request header (group, qualifier, count, index).
	group, _, count, firstIndex, dataIdx, ok := parseControlObjectHeader(req.Data)
	if !ok {
		o.iin.ParameterError = true
		return nil, ErrInvalidRequest
	}

	// Process each item. The first object's index is carried in the header
	// (0x28) or immediately after it (0x00); each subsequent object is
	// prefixed with its own 2-octet LSB index.
	for i := 0; i < count; i++ {
		var index uint16
		if i == 0 {
			index = firstIndex
		} else {
			if dataIdx+2 > len(req.Data) {
				o.iin.ParameterError = true
				return nil, ErrInvalidRequest
			}
			index = uint16(req.Data[dataIdx]) | uint16(req.Data[dataIdx+1])<<8
			dataIdx += 2
		}

		// Parse value based on group/variation
		switch group {
		case 12: // CROB (Control Relay Output Block)
			// CROB: Code(1) + Count(1) + OnTime(4) + OffTime(4) + Status(1) = 11 bytes
			if dataIdx+11 > len(req.Data) {
				o.iin.ParameterError = true
				return nil, ErrInvalidRequest
			}
			crob := CROB{
				Code:    req.Data[dataIdx],
				Count:   req.Data[dataIdx+1],
				OnTime:  uint32(req.Data[dataIdx+2]) | uint32(req.Data[dataIdx+3])<<8 | uint32(req.Data[dataIdx+4])<<16 | uint32(req.Data[dataIdx+5])<<24,
				OffTime: uint32(req.Data[dataIdx+6]) | uint32(req.Data[dataIdx+7])<<8 | uint32(req.Data[dataIdx+8])<<16 | uint32(req.Data[dataIdx+9])<<24,
				Status:  req.Data[dataIdx+10],
			}
			// Call the data handler to process the CROB
			if dh, ok := o.data.(DataWriter); ok {
				if err := dh.WriteBinaryOutput(index, &crob); err != nil {
					o.iin.ParameterError = true
					return nil, err
				}
			}
			dataIdx += 11

		case 41: // Analog Output (Group 41)
			// Fail closed (AO-004 pre-PR corrective F2): the approved G41 command
			// profile permits only FuncDirectOperate WITH RESPONSE. An application
			// WRITE carrying any G41 object must not invoke a host analog command
			// handler or mutate state, and must not be acknowledged as success.
			// The G41 value is never decoded here; the request is rejected before
			// any side effect regardless of variation or record shape.
			o.iin.ParameterError = true
			return nil, fmt.Errorf("G41 analog output is not supported via WRITE; use DirectOperate with response")

		default:
			o.iin.ParameterError = true
			return nil, fmt.Errorf("unsupported write group: %d", group)
		}
	}

	resp := &al.APDU{
		Control: al.AppControl{
			FIR: true,
			FIN: true,
			CON: false,
			UNS: false,
			Seq: o.appSeq,
		},
		FuncCode: al.FuncResponse,
		Data:     nil,
	}
	return resp, nil
}

// CROB represents a Control Relay Output Block (Group 12, Variation 1).
type CROB struct {
	Code    uint8  // Control code (0=NOT_USED, 1=OPEN, 2=CLOSE, 3=PULSE_ON, etc.)
	Count   uint8  // Number of operation cycles
	OnTime  uint32 // Time in ms to energize
	OffTime uint32 // Time in ms to de-energize
	Status  uint8  // Qualifier code (0x00=STANDARD)
}

// DataWriter interface for handling write operations.
type DataWriter interface {
	WriteBinaryOutput(index uint16, crob *CROB) error
	WriteAnalogOutput(index uint16, value interface{}, variation uint8) error
}

// handleSelect handles SELECT requests.
// Implements SBO: Stores the select parameters for later validation in handleOperate.
func (o *Outstation) handleSelect(req *al.APDU) (*al.APDU, error) {
	// Parse the object header from request data
	// Format: Group(1) + Variation(1) + Qualifier(1) + Count(1) + [Index(2) + Value(n)]...
	group, variation, _, index, dataIdx, ok := parseControlObjectHeader(req.Data)
	if !ok {
		return nil, ErrInvalidRequest
	}
	data := make([]byte, len(req.Data)-dataIdx)
	copy(data, req.Data[dataIdx:])

	debugLog("handleSelect: group=%d, variation=%d, index=%d, valueLen=%d", group, variation, index, len(data))

	// Store the pending select
	o.selectMu.Lock()
	o.pendingSelects[index] = &PendingSelect{
		Group:     group,
		Variation: variation,
		Index:     index,
		Value:     data,
		Timestamp: time.Now(),
	}
	o.selectMu.Unlock()

	resp := &al.APDU{
		Control: al.AppControl{
			FIR: true,
			FIN: true,
			CON: false,
			UNS: false,
			Seq: o.appSeq,
		},
		FuncCode: al.FuncResponse,
		Data:     nil, // Success response with no data
	}
	return resp, nil
}

// handleOperate handles OPERATE requests.
// Implements SBO: Validates that a prior SELECT was received for this point.
func (o *Outstation) handleOperate(req *al.APDU) (*al.APDU, error) {
	debugLog("handleOperate: received")

	// Parse the object header from request data
	// Format: Group(1) + Variation(1) + Qualifier(1) + Count(1) + [Index(2) + Value(n)]...
	group, variation, _, index, dataIdx, ok := parseControlObjectHeader(req.Data)
	if !ok {
		return nil, ErrInvalidRequest
	}
	value := make([]byte, len(req.Data)-dataIdx)
	copy(value, req.Data[dataIdx:])

	debugLog("handleOperate: group=%d, variation=%d, index=%d", group, variation, index)

	// Validate against pending select
	o.selectMu.Lock()
	pending, exists := o.pendingSelects[index]

	if !exists {
		o.selectMu.Unlock()
		debugLog("handleOperate: no pending select for index=%d", index)
		return nil, ErrNoSelectPending
	}

	// Check timeout
	timeoutDuration := time.Duration(o.config.SBOTimeout) * time.Millisecond
	if time.Since(pending.Timestamp) > timeoutDuration {
		delete(o.pendingSelects, index)
		o.selectMu.Unlock()
		debugLog("handleOperate: select timeout for index=%d", index)
		return nil, ErrSelectTimeout
	}

	// Validate that operate matches select
	if pending.Group != group || pending.Variation != variation {
		delete(o.pendingSelects, index)
		o.selectMu.Unlock()
		debugLog("handleOperate: select mismatch for index=%d", index)
		return nil, ErrSelectMismatch
	}

	// Clear the pending select after successful validation
	delete(o.pendingSelects, index)
	o.selectMu.Unlock()

	// Execute the actual control operation using the stored select value
	// Call the DataWriter if the data handler implements it
	if dh, ok := o.data.(DataWriter); ok {
		debugLog("handleOperate: calling executeControl for index=%d", index)
		err := o.executeControl(dh, group, variation, index, pending.Value)
		if err != nil {
			o.iin.ParameterError = true
			debugLog("handleOperate: executeControl error: %v", err)
			return nil, err
		}
	} else {
		debugLog("handleOperate: data handler does not implement DataWriter!")
	}

	resp := &al.APDU{
		Control: al.AppControl{
			FIR: true,
			FIN: true,
			CON: false,
			UNS: false,
			Seq: o.appSeq,
		},
		FuncCode: al.FuncResponse,
		Data:     nil,
	}
	return resp, nil
}

// handleDirectOperate handles DIRECT OPERATE requests.
func (o *Outstation) handleDirectOperate(req *al.APDU) (*al.APDU, error) {
	debugLog("handleDirectOperate: received")

	group, variation, count, index, dataIdx, ok := parseControlObjectHeader(req.Data)
	if !ok {
		return nil, ErrInvalidRequest
	}
	value := make([]byte, len(req.Data)-dataIdx)
	copy(value, req.Data[dataIdx:])

	debugLog("handleDirectOperate: group=%d, variation=%d, index=%d, count=%d, valueLen=%d", group, variation, index, count, len(value))

	// AO-002 F3 / AO-003: this slice supports exactly ONE G41 point per Direct
	// Operate. Validate the variation, count and record length before any side
	// effect so a batch, empty, wrong-variation or malformed request is rejected
	// without calling the output handler and without echoing SUCCESS.
	// parseControlObjectHeader discards the count for G12V1, whose single-point
	// behavior is unchanged and covered by the existing CROB suite.
	if group == 41 {
		width := g41ValueWidth(variation)
		if width == 0 || count != 1 || len(value) != width+1 {
			debugLog("handleDirectOperate: rejecting G41 variation=%d count=%d valueLen=%d", variation, count, len(value))
			o.iin.ParameterError = true
			return nil, ErrInvalidRequest
		}
	}

	// Execute the control operation directly.
	dh, ok := o.data.(DataWriter)
	if !ok {
		// AO-002 F1: fail closed. Without a DataWriter the control cannot be
		// executed, so a success response must never be constructed. For G41V1
		// the master reads the echoed object's command-status octet, so emitting
		// a SUCCESS echo here would be reported as a false success despite no
		// execution. No handler is called and no output state is mutated.
		debugLog("handleDirectOperate: data handler does not implement DataWriter; failing closed")
		o.iin.ParameterError = true
		return nil, ErrNotSupported
	}
	debugLog("handleDirectOperate: calling executeControl")
	if err := o.executeControl(dh, group, variation, index, value); err != nil {
		// AO-004 Appendix T: when the handler refuses with a specific,
		// wire-encodable command status (for example OUT_OF_RANGE), echo the
		// object with that status so the master reads a definite per-point
		// non-success instead of a generic error response. A malformed,
		// unsupported or otherwise unclassifiable request still fails closed
		// with an error response (no echo) — refusal is never turned into
		// success.
		if group == 41 && g41ValueWidth(variation) != 0 && len(value) == g41ValueWidth(variation)+1 {
			if cse, ok := AsControlStatusError(err); ok {
				debugLog("handleDirectOperate: echoing G41 object with command status %d", cse.Status)
				return &al.APDU{
					Control: al.AppControl{
						FIR: true,
						FIN: true,
						CON: false,
						UNS: false,
						Seq: o.appSeq,
					},
					FuncCode: al.FuncResponse,
					Data:     buildG41EchoObject(variation, index, value, cse.Status),
				}, nil
			}
		}
		o.iin.ParameterError = true
		debugLog("handleDirectOperate: executeControl error: %v", err)
		return nil, err
	}

	resp := &al.APDU{
		Control: al.AppControl{
			FIR: true,
			FIN: true,
			CON: false,
			UNS: false,
			Seq: o.appSeq,
		},
		FuncCode: al.FuncResponse,
		Data:     nil,
	}
	// G41 (AO-002 V1 / AO-003 V2): echo the object with the index, the numeric
	// value and a command-status octet (SUCCESS here — the handler accepted the
	// value). A handler refusal that carries a wire-encodable status is echoed
	// above with that status; any other refusal yields an error response the
	// master classifies as a failure (never success), not an echo.
	if group == 41 && g41ValueWidth(variation) != 0 && len(value) == g41ValueWidth(variation)+1 {
		resp.Data = buildG41EchoObject(variation, index, value, 0x00)
	}
	return resp, nil
}

// buildG41EchoObject builds the G41 response object echoed for a Direct Operate:
// group(0x29) variation qualifier(0x28) count(01 00), then the 16-bit index (LE),
// the little-endian numeric value (taken from the request record) and a
// command-status octet. The value width follows the variation (V1 int32 = 4
// octets, V2 int16 = 2 octets, V3 float32 = 4 octets, V4 float64 = 8 octets).
// status is the IEEE 1815 command-status octet (0x00 = SUCCESS; 0x0C =
// OUT_OF_RANGE), supplied by the caller so a handler refusal is echoed with its
// actual status rather than a fabricated success.
func buildG41EchoObject(variation uint8, index uint16, requestValue []byte, status uint8) []byte {
	width := g41ValueWidth(variation)
	out := []byte{41, variation, 0x28, 0x01, 0x00, byte(index), byte(index >> 8)}
	out = append(out, requestValue[:width]...)
	return append(out, status)
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

// executeControl executes a control operation by parsing the value and calling the DataWriter.
func (o *Outstation) executeControl(dh DataWriter, group uint8, variation uint8, index uint16, value []byte) error {
	debugLog("executeControl: group=%d, variation=%d, index=%d, valueLen=%d", group, variation, index, len(value))

	switch group {
	case 12: // CROB (Control Relay Output Block)
		if len(value) < 11 {
			debugLog("executeControl: CROB value too short, need 11 bytes got %d", len(value))
			return ErrInvalidRequest
		}
		crob := CROB{
			Code:    value[0],
			Count:   value[1],
			OnTime:  uint32(value[2]) | uint32(value[3])<<8 | uint32(value[4])<<16 | uint32(value[5])<<24,
			OffTime: uint32(value[6]) | uint32(value[7])<<8 | uint32(value[8])<<16 | uint32(value[9])<<24,
			Status:  value[10],
		}
		debugLog("executeControl: calling WriteBinaryOutput index=%d crob=%+v", index, crob)
		return dh.WriteBinaryOutput(index, &crob)

	case 41: // Analog Output — AO-002 V1 (int32), AO-003 V2 (int16)/V3 (float32)/V4 (float64), little-endian
		// G41 object: signed little-endian value then one command-status octet.
		// Reject any other variation and any wrong length rather than falling back
		// to a guessed decode.
		width := g41ValueWidth(variation)
		if width == 0 {
			debugLog("executeControl: unsupported G41 variation %d", variation)
			return ErrInvalidRequest
		}
		if len(value) != width+1 {
			debugLog("executeControl: G41 variation %d value must be %d bytes, got %d", variation, width+1, len(value))
			return ErrInvalidRequest
		}
		switch variation {
		case 1:
			intVal := int32(uint32(value[0]) | uint32(value[1])<<8 | uint32(value[2])<<16 | uint32(value[3])<<24)
			debugLog("executeControl: calling WriteAnalogOutput index=%d value=%v", index, intVal)
			return dh.WriteAnalogOutput(index, intVal, variation)
		case 2:
			intVal := int16(uint16(value[0]) | uint16(value[1])<<8)
			debugLog("executeControl: calling WriteAnalogOutput index=%d value=%v", index, intVal)
			return dh.WriteAnalogOutput(index, intVal, variation)
		case 3: // float32
			// Reject an inbound NaN/±Inf wire value before any handler side effect
			// (owner-approved AO-003 float policy); it must never be executed or
			// echoed as SUCCESS.
			f := math.Float32frombits(uint32(value[0]) | uint32(value[1])<<8 | uint32(value[2])<<16 | uint32(value[3])<<24)
			if math.IsNaN(float64(f)) || math.IsInf(float64(f), 0) {
				debugLog("executeControl: rejecting non-finite G41V3 value")
				return ErrInvalidRequest
			}
			debugLog("executeControl: calling WriteAnalogOutput index=%d value=%v", index, f)
			return dh.WriteAnalogOutput(index, f, variation)
		default: // 4: float64
			// Reject an inbound NaN/±Inf wire value before any handler side effect
			// (owner-approved AO-003 float policy); it must never be executed or
			// echoed as SUCCESS.
			f := math.Float64frombits(uint64(value[0]) | uint64(value[1])<<8 | uint64(value[2])<<16 | uint64(value[3])<<24 |
				uint64(value[4])<<32 | uint64(value[5])<<40 | uint64(value[6])<<48 | uint64(value[7])<<56)
			if math.IsNaN(f) || math.IsInf(f, 0) {
				debugLog("executeControl: rejecting non-finite G41V4 value")
				return ErrInvalidRequest
			}
			debugLog("executeControl: calling WriteAnalogOutput index=%d value=%v", index, f)
			return dh.WriteAnalogOutput(index, f, variation)
		}

	default:
		debugLog("executeControl: unknown group %d", group)
		return ErrInvalidRequest
	}
}

// handleDirectOperateNoResp handles DIRECT OPERATE NO RESPONSE requests.
func (o *Outstation) handleDirectOperateNoResp(req *al.APDU) (*al.APDU, error) {
	// No response for this function code
	return nil, nil
}

// handleEnableUnsolicited handles ENABLE UNSOLICITED requests.
// Note: Called while holding lock from ProcessRequest.
func (o *Outstation) handleEnableUnsolicited(req *al.APDU) (*al.APDU, error) {
	o.unsolicited = true

	resp := &al.APDU{
		Control: al.AppControl{
			FIR: true,
			FIN: true,
			CON: false,
			UNS: false,
			Seq: o.appSeq,
		},
		FuncCode: al.FuncResponse,
		Data:     nil,
	}
	return resp, nil
}

// handleDisableUnsolicited handles DISABLE UNSOLICITED requests.
// Note: Called while holding lock from ProcessRequest.
func (o *Outstation) handleDisableUnsolicited(req *al.APDU) (*al.APDU, error) {
	o.unsolicited = false

	resp := &al.APDU{
		Control: al.AppControl{
			FIR: true,
			FIN: true,
			CON: false,
			UNS: false,
			Seq: o.appSeq,
		},
		FuncCode: al.FuncResponse,
		Data:     nil,
	}
	return resp, nil
}

// handleFreeze handles FREEZE requests.
func (o *Outstation) handleFreeze(req *al.APDU) (*al.APDU, error) {
	// Freeze counters without clearing
	if err := o.data.FreezeCounters(false); err != nil {
		o.iin.ParameterError = true
		return nil, err
	}

	resp := &al.APDU{
		Control: al.AppControl{
			FIR: true,
			FIN: true,
			CON: false,
			UNS: false,
			Seq: o.appSeq,
		},
		FuncCode: al.FuncResponse,
		Data:     nil,
	}
	return resp, nil
}

// handleFreezeClear handles FREEZE CLEAR requests.
func (o *Outstation) handleFreezeClear(req *al.APDU) (*al.APDU, error) {
	// Freeze counters and clear
	if err := o.data.FreezeCounters(true); err != nil {
		o.iin.ParameterError = true
		return nil, err
	}

	resp := &al.APDU{
		Control: al.AppControl{
			FIR: true,
			FIN: true,
			CON: false,
			UNS: false,
			Seq: o.appSeq,
		},
		FuncCode: al.FuncResponse,
		Data:     nil,
	}
	return resp, nil
}

// handleTimeSync handles TIME SYNCHRONIZATION requests.
func (o *Outstation) handleTimeSync(req *al.APDU) (*al.APDU, error) {
	// Time sync is typically done via broadcast; no response expected
	return nil, nil
}

// handleDelayMeasurement handles DELAY MEASUREMENT requests.
func (o *Outstation) handleDelayMeasurement(req *al.APDU) (*al.APDU, error) {
	resp := &al.APDU{
		Control: al.AppControl{
			FIR: true,
			FIN: true,
			CON: false,
			UNS: false,
			Seq: o.appSeq,
		},
		FuncCode: al.FuncResponse,
		Data:     []byte{0, 0}, // Delay in milliseconds
	}
	return resp, nil
}

// handleRecordCurrentTime handles RECORD CURRENT TIME requests.
func (o *Outstation) handleRecordCurrentTime(req *al.APDU) (*al.APDU, error) {
	// Record time and don't respond
	return nil, nil
}

// sendLinkAck sends a link-layer ACK response.
// This is used to acknowledge Reset Link Stations frames.
func (o *Outstation) sendLinkAck(masterAddr uint16) {
	o.mu.RLock()
	transport := o.transport
	o.mu.RUnlock()

	if transport == nil {
		return
	}

	// ACK frame: DIR=0, PRM=0, FuncCode=0 (ACK)
	dllFrame := &frame.Frame{
		Control: frame.Control{
			DIR:      false,         // Outstation-to-Master
			PRM:      false,         // Secondary station
			FuncCode: frame.FuncAck, // ACK
		},
		DestAddr: masterAddr,
		SrcAddr:  o.config.OutstationAddress,
		Data:     nil, // ACK has no data
	}

	dllEncoded, err := frame.Encode(dllFrame)
	if err != nil {
		return
	}

	_ = transport.Send(dllEncoded)
}

// sendLinkStatus sends a link-layer status response.
// This is used to respond to Link Status Request frames.
func (o *Outstation) sendLinkStatus(masterAddr uint16) {
	o.mu.RLock()
	transport := o.transport
	o.mu.RUnlock()

	if transport == nil {
		return
	}

	// Link Status frame: DIR=0, PRM=0, FuncCode=2 (Link Status)
	dllFrame := &frame.Frame{
		Control: frame.Control{
			DIR:      false,                // Outstation-to-Master
			PRM:      false,                // Secondary station
			FuncCode: frame.FuncLinkStatus, // Link Status
			DFC:      false,                // Data link not busy
		},
		DestAddr: masterAddr,
		SrcAddr:  o.config.OutstationAddress,
		Data:     nil, // Link Status has no data
	}

	dllEncoded, err := frame.Encode(dllFrame)
	if err != nil {
		return
	}

	_ = transport.Send(dllEncoded)
}

// sendErrorResponse sends an error response with DLL framing.
func (o *Outstation) sendErrorResponse(seq uint8, err error) {
	o.mu.RLock()
	transport := o.transport
	o.mu.RUnlock()

	if transport == nil {
		return
	}

	iin := &al.IIN{
		FuncUnknown: true,
	}

	resp := &al.APDU{
		Control: al.AppControl{
			FIR: true,
			FIN: true,
			CON: false,
			UNS: false,
			Seq: seq,
		},
		FuncCode: al.FuncResponse,
		Data:     al.EncodeIIN(iin),
	}

	fragments := o.buildTransportFragments(resp)
	for _, frag := range fragments {
		// Wrap in DLL frame. An error response is user data; frame it primary.
		dllFrame := &frame.Frame{
			Control: frame.Control{
				DIR:      false,
				PRM:      true,
				FuncCode: frame.FuncUnconfirmedUserData,
			},
			DestAddr: o.config.MasterAddress,
			SrcAddr:  o.config.OutstationAddress,
			Data:     frag,
		}
		dllEncoded, err := frame.Encode(dllFrame)
		if err != nil {
			continue
		}
		_ = transport.Send(dllEncoded)
	}
}

// cleanup performs resource cleanup when stopping.
func (o *Outstation) cleanup() {
	// Clear pending selects
	o.selectMu.Lock()
	o.pendingSelects = make(map[uint16]*PendingSelect)
	o.selectMu.Unlock()
	// Clear event queue
	o.eventQueue.Clear()
	// Clear IIN
	o.mu.Lock()
	o.iin = al.IIN{}
	o.mu.Unlock()
}
