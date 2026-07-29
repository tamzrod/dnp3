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
	"sync"
	"time"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/tl"
)

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
	ErrNotInitialized     = errors.New("outstation not initialized")
	ErrNotOperational    = errors.New("outstation not operational")
	ErrInvalidRequest     = errors.New("invalid request")
	ErrBufferFull        = errors.New("event buffer full")
	ErrTransportNotSet   = errors.New("transport not configured")
	// SBO errors
	ErrSelectTimeout      = errors.New("select timeout - no operate received")
	ErrNoSelectPending    = errors.New("no select pending for this point")
	ErrSelectMismatch    = errors.New("operate does not match select")
)

// Freeze errors
var (
	ErrFreezeNotSupported = errors.New("freeze not supported for this point")
)

// DataHandler defines the interface for accessing data points.
type DataHandler interface {
	GetBinaryInputs() []BinaryInput
	GetAnalogInputs() []AnalogInput
	GetCounters() []Counter
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
	Group     uint8  // Object group
	Variation uint8  // Object variation
	Index     uint16 // Point index
	Value     []byte // Command value
	Timestamp time.Time // When select was received
}

// BinaryInput represents a binary input data point.
type BinaryInput struct {
	Value   bool  // Current value (true=on, false=off)
	Quality uint8 // Quality flags
}

// Quality flags for binary inputs
const (
	BinaryQualityOnline    = 0x01 // Point online
	BinaryQualityRestart   = 0x40 // Device restarted
	BinaryQualityCommLost  = 0x80 // Communication lost
	BinaryQualityDiscontinuous = 0x20 // Data discontinuous
)

// AnalogInput represents an analog input data point.
type AnalogInput struct {
	Value   float64 // Current value
	Quality uint8   // Quality flags
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

	// Update the binary output based on CROB command
	// In a real implementation, this would trigger actual control action
	switch crob.Code {
	case 1: // NUL (no operation)
		// No action
	case 2: // CLOSE (turn ON)
		d.binaryOutputs[index] = BinaryOutput{Value: true, Quality: BinaryQualityOnline}
	case 3: // OPEN (turn OFF)
		d.binaryOutputs[index] = BinaryOutput{Value: false, Quality: BinaryQualityOnline}
	case 4: // TRIP/PERMIT (toggle or pulse)
		// Pulse on/off
		d.binaryOutputs[index] = BinaryOutput{Value: true, Quality: BinaryQualityOnline}
	case 5: // PULSE_ON
		// Would typically pulse high then low
		d.binaryOutputs[index] = BinaryOutput{Value: true, Quality: BinaryQualityOnline}
	case 6: // PULSE_OFF
		// Would typically pulse low then high
		d.binaryOutputs[index] = BinaryOutput{Value: false, Quality: BinaryQualityOnline}
	case 7: // LATCH_ON
		d.binaryOutputs[index] = BinaryOutput{Value: true, Quality: BinaryQualityOnline}
	case 8: // LATCH_OFF
		d.binaryOutputs[index] = BinaryOutput{Value: false, Quality: BinaryQualityOnline}
	default:
		return fmt.Errorf("unsupported CROB code: %d", crob.Code)
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
	config     *Config
	state      State
	data       DataHandler
	mu         sync.RWMutex
	transport  TransportHandler
	onRequest  RequestHandler
	appSeq     uint8         // Application sequence number
	unsolicited bool         // Unsolicited responses enabled
	iin        al.IIN        // Internal Indication
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
		fragmenter:      tl.NewFragmenter(),
		reassembler:    tl.NewReassembler(),
		pendingSelects: make(map[uint16]*PendingSelect),
		eventQueue:    NewEventQueue(config.MaxEventBuffers),
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
		o.iin.ByteOver = true
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
	o.iin.ByteOver = false
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
				// Wrap in DLL frame
				dllFrame := &frame.Frame{
					Control: frame.Control{
						DIR:      false, // Outstation-to-Master
						PRM:      false, // Secondary station
						FuncCode: frame.FuncConfirmedUserData,
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

	return nil
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
		dllFrame := &frame.Frame{
			Control: frame.Control{
				DIR:      false, // Outstation-to-Master
				PRM:      false, // Secondary station
				FuncCode: frame.FuncConfirmedUserData,
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

		// Move past this DLL frame
		// Header: sync(2) + length(1) + control(1) + dest(2) + src(2) = 8 bytes
		// Then data + CRC
		headerSize := 8
		crcSize := ((len(dllFrame.Data) + 1) / 2) * 2
		offset += headerSize + len(dllFrame.Data) + crcSize

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

		// Skip non-user-data frames (primary station function codes)
		// For Master→Outstation, only FuncConfirmedUserData (4) carries user data
		if !dllFrame.Control.PRM || dllFrame.Control.FuncCode != frame.FuncConfirmedUserData {
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
		o.iin.ParamUnavail = true
		err = fmt.Errorf("unsupported function code: %d", req.FuncCode)
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

// buildReadResponse builds data for a READ response.
func (o *Outstation) buildReadResponse(requestData []byte) []byte {
	var result []byte

	// If no specific data requested, return all static data
	if len(requestData) < 4 {
		result = o.buildAllStaticData()
		return result
	}

	// Parse object headers
	offset := 0
	for offset < len(requestData) {
		if offset+4 > len(requestData) {
			break
		}

		group := requestData[offset]
		variation := requestData[offset+1]
		_ = requestData[offset+2] // qualifier
		_ = requestData[offset+3] // range

		switch group {
		case 0: // Null qualifier - all data
			result = append(result, o.buildAllStaticData()...)
		case 1: // Binary Input
			result = append(result, o.buildBinaryInputData(variation)...)
		case 2: // Binary Input Event
			result = append(result, o.buildBinaryInputData(variation)...)
		case 20: // Counter
			result = append(result, o.buildCounterData(variation)...)
		case 21: // Counter Event
			result = append(result, o.buildCounterData(variation)...)
		case 30: // Analog Input
			result = append(result, o.buildAnalogInputData(variation)...)
		case 31: // Analog Input Event
			result = append(result, o.buildAnalogInputData(variation)...)
		case 60: // Class data
			result = append(result, o.buildAllStaticData()...)
		}

		offset += 4
	}

	if len(result) == 0 {
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
	return result
}

// buildBinaryInputData builds binary input data.
func (o *Outstation) buildBinaryInputData(variation uint8) []byte {
	var result []byte
	data := o.data.GetBinaryInputs()

	if len(data) == 0 {
		return result
	}

	// Object header: group, variation, qualifier (index), count - OUTSIDE loop
	result = append(result, 1)              // Group 1
	result = append(result, variation)     // Variation
	result = append(result, 0x00)         // Qualifier: index
	result = append(result, byte(len(data))) // Count

	for i, bi := range data {
		// Index (2 bytes, big-endian)
		result = append(result, byte(i>>8), byte(i&0xFF))

		// Value based on variation
		switch variation {
		case 1: // Binary Input with flags
			val := byte(0)
			if bi.Value {
				val = 0x01
			}
			result = append(result, val|bi.Quality)
		case 2: // Binary Input without flags
			if bi.Value {
				result = append(result, 0x01)
			} else {
				result = append(result, 0x00)
			}
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
	result = append(result, 30)             // Group 30
	result = append(result, variation)      // Variation
	result = append(result, 0x00)          // Qualifier: index
	result = append(result, byte(len(data))) // Count

	for i, ai := range data {
		// Index (2 bytes, big-endian)
		result = append(result, byte(i>>8), byte(i&0xFF))

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
	result = append(result, 20)             // Group 20
	result = append(result, variation)     // Variation
	result = append(result, 0x00)          // Qualifier: index
	result = append(result, byte(len(data))) // Count

	for i, c := range data {
		// Index (2 bytes, big-endian)
		result = append(result, byte(i>>8), byte(i&0xFF))

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

// handleWrite handles WRITE requests.
func (o *Outstation) handleWrite(req *al.APDU) (*al.APDU, error) {
	// Parse write request
	// Format: Group(1) + Variation(1) + Qualifier(1) + Count(1) + [Index(2) + Value(n)]...
	if len(req.Data) < 5 {
		o.iin.ParamUnavail = true
		return nil, ErrInvalidRequest
	}

	group := req.Data[0]
	variation := req.Data[1]
	count := int(req.Data[3])

	// Process each item
	dataIdx := 4
	for i := 0; i < count; i++ {
		if dataIdx+2 > len(req.Data) {
			o.iin.ParamUnavail = true
			return nil, ErrInvalidRequest
		}

		index := uint16(req.Data[dataIdx])<<8 | uint16(req.Data[dataIdx+1])
		dataIdx += 2

		// Parse value based on group/variation
		switch group {
		case 12: // CROB (Control Relay Output Block)
			// CROB: Code(1) + Count(1) + OnTime(4) + OffTime(4) + Status(1) = 11 bytes
			if dataIdx+11 > len(req.Data) {
				o.iin.ParamUnavail = true
				return nil, ErrInvalidRequest
			}
			crob := CROB{
				Code:   req.Data[dataIdx],
				Count:  req.Data[dataIdx+1],
				OnTime: uint32(req.Data[dataIdx+2])<<24 | uint32(req.Data[dataIdx+3])<<16 | uint32(req.Data[dataIdx+4])<<8 | uint32(req.Data[dataIdx+5]),
				OffTime: uint32(req.Data[dataIdx+6])<<24 | uint32(req.Data[dataIdx+7])<<16 | uint32(req.Data[dataIdx+8])<<8 | uint32(req.Data[dataIdx+9]),
				Status: req.Data[dataIdx+10],
			}
			// Call the data handler to process the CROB
			if dh, ok := o.data.(DataWriter); ok {
				if err := dh.WriteBinaryOutput(index, &crob); err != nil {
					o.iin.ParamUnavail = true
					return nil, err
				}
			}
			dataIdx += 11

		case 41, 42, 43, 44: // Analog outputs
			// Parse analog value based on variation
			var value interface{}
			switch variation {
			case 1, 5: // 16-bit signed
				if dataIdx+2 > len(req.Data) {
					o.iin.ParamUnavail = true
					return nil, ErrInvalidRequest
				}
				value = int16(req.Data[dataIdx])<<8 | int16(req.Data[dataIdx+1])
				dataIdx += 2
			case 2, 6: // 16-bit unsigned
				if dataIdx+2 > len(req.Data) {
					o.iin.ParamUnavail = true
					return nil, ErrInvalidRequest
				}
				value = uint16(req.Data[dataIdx])<<8 | uint16(req.Data[dataIdx+1])
				dataIdx += 2
			case 3, 7: // 32-bit signed
				if dataIdx+4 > len(req.Data) {
					o.iin.ParamUnavail = true
					return nil, ErrInvalidRequest
				}
				value = int32(req.Data[dataIdx])<<24 | int32(req.Data[dataIdx+1])<<16 | int32(req.Data[dataIdx+2])<<8 | int32(req.Data[dataIdx+3])
				dataIdx += 4
			case 4, 8: // 32-bit unsigned
				if dataIdx+4 > len(req.Data) {
					o.iin.ParamUnavail = true
					return nil, ErrInvalidRequest
				}
				value = uint32(req.Data[dataIdx])<<24 | uint32(req.Data[dataIdx+1])<<16 | uint32(req.Data[dataIdx+2])<<8 | uint32(req.Data[dataIdx+3])
				dataIdx += 4
			case 9, 10: // 32-bit float
				if dataIdx+4 > len(req.Data) {
					o.iin.ParamUnavail = true
					return nil, ErrInvalidRequest
				}
				bits := uint32(req.Data[dataIdx])<<24 | uint32(req.Data[dataIdx+1])<<16 | uint32(req.Data[dataIdx+2])<<8 | uint32(req.Data[dataIdx+3])
				value = math.Float32frombits(bits)
				dataIdx += 4
			case 13, 14: // 64-bit double
				if dataIdx+8 > len(req.Data) {
					o.iin.ParamUnavail = true
					return nil, ErrInvalidRequest
				}
				bits := uint64(req.Data[dataIdx])<<56 | uint64(req.Data[dataIdx+1])<<48 | uint64(req.Data[dataIdx+2])<<40 | uint64(req.Data[dataIdx+3])<<32 |
					uint64(req.Data[dataIdx+4])<<24 | uint64(req.Data[dataIdx+5])<<16 | uint64(req.Data[dataIdx+6])<<8 | uint64(req.Data[dataIdx+7])
				value = math.Float64frombits(bits)
				dataIdx += 8
			default:
				o.iin.ParamUnavail = true
				return nil, fmt.Errorf("unsupported analog variation: %d", variation)
			}

			// Call the data handler to process the analog output
			if dh, ok := o.data.(DataWriter); ok {
				if err := dh.WriteAnalogOutput(index, value, variation); err != nil {
					o.iin.ParamUnavail = true
					return nil, err
				}
			}
		default:
			o.iin.ParamUnavail = true
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
	if len(req.Data) < 5 {
		return nil, ErrInvalidRequest
	}

	group := req.Data[0]
	variation := req.Data[1]
	// qualifier := req.Data[2] // Not used but part of structure
	// count := req.Data[3]     // Not used but part of structure
	
	// Extract index and value from data (after 4-byte header)
	dataIdx := 4
	if dataIdx+2 > len(req.Data) {
		return nil, ErrInvalidRequest
	}
	index := uint16(req.Data[dataIdx])<<8 | uint16(req.Data[dataIdx+1])
	value := make([]byte, len(req.Data)-dataIdx-2)
	copy(value, req.Data[dataIdx+2:])

	// Store the pending select
	o.selectMu.Lock()
	o.pendingSelects[index] = &PendingSelect{
		Group:     group,
		Variation: variation,
		Index:     index,
		Value:     value,
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
	// Parse the object header from request data
	// Format: Group(1) + Variation(1) + Qualifier(1) + Count(1) + [Index(2) + Value(n)]...
	if len(req.Data) < 5 {
		return nil, ErrInvalidRequest
	}

	group := req.Data[0]
	variation := req.Data[1]
	
	// Extract index and value from data (after 4-byte header)
	dataIdx := 4
	if dataIdx+2 > len(req.Data) {
		return nil, ErrInvalidRequest
	}
	index := uint16(req.Data[dataIdx])<<8 | uint16(req.Data[dataIdx+1])
	value := make([]byte, len(req.Data)-dataIdx-2)
	copy(value, req.Data[dataIdx+2:])

	// Validate against pending select
	o.selectMu.Lock()
	pending, exists := o.pendingSelects[index]
	
	if !exists {
		o.selectMu.Unlock()
		return nil, ErrNoSelectPending
	}

	// Check timeout
	timeoutDuration := time.Duration(o.config.SBOTimeout) * time.Millisecond
	if time.Since(pending.Timestamp) > timeoutDuration {
		delete(o.pendingSelects, index)
		o.selectMu.Unlock()
		return nil, ErrSelectTimeout
	}

	// Validate that operate matches select
	if pending.Group != group || pending.Variation != variation {
		delete(o.pendingSelects, index)
		o.selectMu.Unlock()
		return nil, ErrSelectMismatch
	}

	// Clear the pending select after successful validation
	delete(o.pendingSelects, index)
	o.selectMu.Unlock()

	// Execute the actual control operation
	// In a real implementation, this would call the data handler to perform the action
	// For now, we just return success

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
		o.iin.ParamUnavail = true
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
		o.iin.ParamUnavail = true
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
			DIR:      false,            // Outstation-to-Master
			PRM:      false,            // Secondary station
			FuncCode: frame.FuncAck,   // ACK
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
			DIR:      false,              // Outstation-to-Master
			PRM:      false,              // Secondary station
			FuncCode: frame.FuncLinkStatus, // Link Status
			DFC:      false,              // Data link not busy
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
		ParamUnavail: true,
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
		// Wrap in DLL frame
		dllFrame := &frame.Frame{
			Control: frame.Control{
				DIR:      false,
				PRM:      false,
				FuncCode: frame.FuncConfirmedUserData,
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
