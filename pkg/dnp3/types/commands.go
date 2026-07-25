package types

// CommandType represents the type of control command
type CommandType int

const (
	// SelectThenOperate uses the select-before-operate pattern
	SelectThenOperate CommandType = iota
	// DirectOperate sends the command directly without select
	DirectOperate
	// DirectOperateNoResponse sends the command with no response expected
	DirectOperateNoResponse
	// SelectAndOperateOnce same as select then operate but only operates once
	SelectAndOperateOnce
	// SelectAndOperateMany selects and operates multiple times
	SelectAndOperateMany
)

// ControlOutput represents a DNP3 control output command
type ControlOutput struct {
	// Group is the object group for the command
	Group uint8
	// Variation is the object variation for the command
	Variation uint8
	// Index is the point index to control
	Index uint16
	// Value is the command value
	Value CommandValue
	// CommandType specifies the command pattern
	CommandType CommandType
	// Count is the number of times to repeat (for SelectAndOperateMany)
	Count uint8
	// OnTime is the on-time in milliseconds (for pulse commands)
	OnTime uint32
	// OffTime is the off-time in milliseconds (for pulse commands)
	OffTime uint32
	// Status is the command status (set by outstation)
	Status ControlStatus
}

// CommandValue represents the value to send with a control command
type CommandValue interface {
	// isCommandValue is unexported to ensure only this package can create values
	isCommandValue()
}

// BinaryCommandValue represents a binary (on/off) command value
type BinaryCommandValue struct {
	Value bool
}

func (b *BinaryCommandValue) isCommandValue() {}

// AnalogCommandValue represents an analog command value
type AnalogCommandValue struct {
	Value float64
}

func (a *AnalogCommandValue) isCommandValue() {}

// ControlStatus represents the status of a control command
type ControlStatus uint8

const (
	// ControlSuccess indicates the command was successful
	ControlSuccess ControlStatus = 0
	// ControlTimeout indicates the command timed out
	ControlTimeout ControlStatus = 1
	// ControlNoSelect indicates select was not received
	ControlNoSelect ControlStatus = 2
	// ControlBadFormat indicates bad format in the command
	ControlBadFormat ControlStatus = 3
	// ControlNotSupported indicates the command is not supported
	ControlNotSupported ControlStatus = 4
	// ControlAlreadyActive indicates the output is already active
	ControlAlreadyActive ControlStatus = 5
	// ControlBlocked indicates the output is blocked
	ControlBlocked ControlStatus = 6
	// ControlLocal indicates the output is in local mode
	ControlLocal ControlStatus = 7
	// ControlTooMany indicates too many commands are queued
	ControlTooMany ControlStatus = 8
	// ControlNotAuthorized indicates the command is not authorized
	ControlNotAuthorized ControlStatus = 9
	// ControlAutonomous indicates autonomous operation
	ControlAutonomous ControlStatus = 10
	// ControlSynchTimeout indicates synchronization timed out
	ControlSynchTimeout ControlStatus = 11
	// ControlSynchBlocked indicates synchronization is blocked
	ControlSynchBlocked ControlStatus = 12
)

// String returns a human-readable description of the control status
func (s ControlStatus) String() string {
	switch s {
	case ControlSuccess:
		return "success"
	case ControlTimeout:
		return "timeout"
	case ControlNoSelect:
		return "no_select"
	case ControlBadFormat:
		return "bad_format"
	case ControlNotSupported:
		return "not_supported"
	case ControlAlreadyActive:
		return "already_active"
	case ControlBlocked:
		return "blocked"
	case ControlLocal:
		return "local"
	case ControlTooMany:
		return "too_many"
	case ControlNotAuthorized:
		return "not_authorized"
	case ControlAutonomous:
		return "autonomous"
	case ControlSynchTimeout:
		return "synch_timeout"
	case ControlSynchBlocked:
		return "synch_blocked"
	default:
		return "unknown"
	}
}

// GroupRequest represents a request for a specific object group
type GroupRequest struct {
	// Group is the object group number
	Group uint8
	// Variation is the object variation number (0 = any)
	Variation uint8
	// Qualifier specifies the index qualifier
	Qualifier QualifierCode
}

// QualifierCode represents DNP3 qualifier codes
type QualifierCode uint8

const (
	// QualifierAll indicates all points
	QualifierAll QualifierCode = 0x00
	// QualifierIndex indicates index is specified
	QualifierIndex QualifierCode = 0x01
	// QualifierCount8 indicates 8-bit count follows
	QualifierCount8 QualifierCode = 0x07
	// QualifierCount16 indicates 16-bit count follows
	QualifierCount16 QualifierCode = 0x08
	// QualifierRange8 indicates 8-bit range
	QualifierRange8 QualifierCode = 0x17
	// QualifierRange16 indicates 16-bit range
	QualifierRange16 QualifierCode = 0x18
)

// NewBinaryControl creates a new binary control command
func NewBinaryControl(index uint16, value bool, cmdType CommandType) *ControlOutput {
	return &ControlOutput{
		Group:       12, // Group 12 = Binary Output
		Variation:   1,  // Variation 1 = Binary Output (with status)
		Index:       index,
		Value:       &BinaryCommandValue{Value: value},
		CommandType: cmdType,
	}
}

// NewAnalogControl creates a new analog control command
func NewAnalogControl(index uint16, value float64, cmdType CommandType) *ControlOutput {
	return &ControlOutput{
		Group:       41, // Group 41 = Analog Output
		Variation:   1,  // Variation 1 = Analog Output (32-bit)
		Index:       index,
		Value:       &AnalogCommandValue{Value: value},
		CommandType: cmdType,
	}
}

// NewPulseControl creates a new pulse control command
func NewPulseControl(index uint16, onTime, offTime uint32, cmdType CommandType) *ControlOutput {
	return &ControlOutput{
		Group:       12, // Group 12 = Binary Output
		Variation:   3,  // Variation 3 = Latch On (pulse)
		Index:       index,
		Value:       &BinaryCommandValue{Value: true},
		CommandType: cmdType,
		OnTime:      onTime,
		OffTime:     offTime,
	}
}

// ReadRequest represents a request to read data from an outstation
type ReadRequest struct {
	// Groups specifies which object groups to read
	Groups []GroupRequest
}

// NewReadRequest creates a new read request for specified groups
func NewReadRequest(groups ...GroupRequest) *ReadRequest {
	return &ReadRequest{
		Groups: groups,
	}
}

// Common group requests for convenience
var (
	// ReadAllStatic reads all static data (Class 0)
	ReadAllStatic = []GroupRequest{
		{Group: 1, Variation: 0},  // Binary Inputs
		{Group: 30, Variation: 0}, // Analog Inputs
		{Group: 20, Variation: 0}, // Counters
	}
	// ReadAllEvents reads all event data
	ReadAllEvents = []GroupRequest{
		{Group: 2, Variation: 0},  // Binary Input Events
		{Group: 32, Variation: 0}, // Analog Input Events
		{Group: 22, Variation: 0}, // Counter Events
	}
	// ReadBinaryInputs reads binary inputs
	ReadBinaryInputs = []GroupRequest{
		{Group: 1, Variation: 1}, // Binary Input with flags
	}
	// ReadAnalogInputs reads analog inputs
	ReadAnalogInputs = []GroupRequest{
		{Group: 30, Variation: 1}, // Analog Input with flags (32-bit float)
	}
	// ReadCounters reads counters
	ReadCounters = []GroupRequest{
		{Group: 20, Variation: 1}, // 32-bit Counter with flags
	}
)

// WriteRequest represents a request to write data to an outstation
type WriteRequest struct {
	// BinaryOutputs contains binary output values to write
	BinaryOutputs []BinaryOutput
	// AnalogOutputs contains analog output values to write
	AnalogOutputs []AnalogOutput
}
