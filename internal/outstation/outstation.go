// Package outstation implements the DNP3 Outstation role.
//
// The Outstation responds to requests from the Master, provides data
// (binary inputs, analog inputs, counters), executes control commands,
// and reports events via unsolicited responses.
//
// Reference: IEEE 1815-2012 Section 8
package outstation

import (
	"errors"
	"fmt"
	"math"
	"sync"

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
)

// DataHandler defines the interface for accessing data points.
type DataHandler interface {
	GetBinaryInputs() []BinaryInput
	GetAnalogInputs() []AnalogInput
	GetCounters() []Counter
}

// Config holds outstation configuration.
type Config struct {
	OutstationAddress uint16 // This outstation's address
	MasterAddress     uint16 // Expected master's address
	Timeout           int    // Response timeout in milliseconds
	MaxEventBuffers   int    // Maximum event buffer size
	Unsolicited       bool   // Enable unsolicited responses
}

// DefaultConfig returns default outstation configuration.
func DefaultConfig() *Config {
	return &Config{
		OutstationAddress: 1024, // Common default RTU address
		MasterAddress:     0xFFFF,
		Timeout:           5000,
		MaxEventBuffers:   1000,
		Unsolicited:       false,
	}
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

// DefaultDataHandler provides default mock data for testing.
type DefaultDataHandler struct {
	mu           sync.RWMutex
	binaryInputs []BinaryInput
	analogInputs []AnalogInput
	counters     []Counter
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
}

// NewOutstation creates a new Outstation.
func NewOutstation(config *Config) *Outstation {
	if config == nil {
		config = DefaultConfig()
	}
	return &Outstation{
		config:      config,
		state:       StateDown,
		data:        NewDefaultDataHandler(),
		appSeq:      0,
		fragmenter:   tl.NewFragmenter(),
		reassembler: tl.NewReassembler(),
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
	return nil
}

// Run starts the outstation's main loop.
func (o *Outstation) Run() error {
	o.mu.RLock()
	transport := o.transport
	o.mu.RUnlock()

	if transport == nil {
		return ErrTransportNotSet
	}

	for o.State() != StateDown {
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

		// Process request
		resp, err := o.ProcessRequest(req)
		if err != nil {
			o.sendErrorResponse(req.Control.Seq, err)
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
	}

	return nil
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
	switch req.FuncCode {
	case al.FuncRead:
		return o.handleRead(req)
	case al.FuncWrite:
		return o.handleWrite(req)
	case al.FuncSelect:
		return o.handleSelect(req)
	case al.FuncOperate:
		return o.handleOperate(req)
	case al.FuncDirectOperate:
		return o.handleDirectOperate(req)
	case al.FuncDirectOperateNoResp:
		return o.handleDirectOperateNoResp(req)
	case al.FuncEnableUnsolicited:
		return o.handleEnableUnsolicited(req)
	case al.FuncDisableUnsolicited:
		return o.handleDisableUnsolicited(req)
	case al.FuncFreeze:
		return o.handleFreeze(req)
	case al.FuncFreezeClear:
		return o.handleFreezeClear(req)
	case al.FuncTimeSync:
		return o.handleTimeSync(req)
	case al.FuncDelayMeasurement:
		return o.handleDelayMeasurement(req)
	case al.FuncRecordCurrentTime:
		return o.handleRecordCurrentTime(req)
	default:
		o.iin.ParamUnavail = true
		return nil, fmt.Errorf("unsupported function code: %d", req.FuncCode)
	}
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

	for i, bi := range data {
		// Object header: group, variation, qualifier (index), count
		result = append(result, 1)              // Group 1
		result = append(result, variation)     // Variation
		result = append(result, 0x00)          // Qualifier: index
		result = append(result, byte(len(data))) // Count

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

	for i, ai := range data {
		// Object header
		result = append(result, 30)             // Group 30
		result = append(result, variation)      // Variation
		result = append(result, 0x00)          // Qualifier: index
		result = append(result, byte(len(data))) // Count

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

	for i, c := range data {
		// Object header
		result = append(result, 20)             // Group 20
		result = append(result, variation)     // Variation
		result = append(result, 0x00)          // Qualifier: index
		result = append(result, byte(len(data))) // Count

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

// handleSelect handles SELECT requests.
func (o *Outstation) handleSelect(req *al.APDU) (*al.APDU, error) {
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

// handleOperate handles OPERATE requests.
func (o *Outstation) handleOperate(req *al.APDU) (*al.APDU, error) {
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
