// Package types provides common data types for DNP3.
//
// This package contains the data types used by both Master clients
// and Outstation servers for exchanging data over the DNP3 protocol.
//
// The types in this package are part of the public API and are
// guaranteed to be stable across versions.
package types

import "time"

// QualityFlags represents data quality indicators for DNP3 data points.
// These flags indicate the health and validity of a data sample.
type QualityFlags uint8

// Quality flag constants
const (
	// QualityOnline indicates the point is online and working correctly
	QualityOnline QualityFlags = 0x01
	// QualityRestart indicates the device recently restarted
	QualityRestart QualityFlags = 0x40
	// QualityCommLost indicates communication with the device is lost
	QualityCommLost QualityFlags = 0x80
	// QualityDiscontinuous indicates data is discontinuous
	QualityDiscontinuous QualityFlags = 0x20
	// QualityBlocked indicates the value is blocked
	QualityBlocked QualityFlags = 0x08
	// QualityConfigured indicates the value is configured
	QualityConfigured QualityFlags = 0x04
	// QualityOverRange indicates the value is out of range
	QualityOverRange QualityFlags = 0x04
	// QualityRefError indicates a reference error
	QualityRefError QualityFlags = 0x02
	// QualityRollOver indicates counter rollover
	QualityRollOver QualityFlags = 0x08
)

// IsGood returns true if the quality indicates good data
func (q QualityFlags) IsGood() bool {
	return q&QualityOnline != 0 && q&QualityCommLost == 0
}

// String returns a human-readable description of the quality flags
func (q QualityFlags) String() string {
	var flags []string
	if q&QualityOnline != 0 {
		flags = append(flags, "ONLINE")
	}
	if q&QualityRestart != 0 {
		flags = append(flags, "RESTART")
	}
	if q&QualityCommLost != 0 {
		flags = append(flags, "COMM_LOST")
	}
	if q&QualityDiscontinuous != 0 {
		flags = append(flags, "DISCONTINUOUS")
	}
	if len(flags) == 0 {
		return "BAD"
	}
	return joinStrings(flags, "|")
}

// BinaryInput represents a binary (digital) input data point.
// Binary inputs are single-bit values representing on/off or true/false states.
type BinaryInput struct {
	// Index is the point index within the outstation
	Index uint16
	// Value is the binary state (true = on/closed, false = off/open)
	Value bool
	// Quality indicates the data quality
	Quality QualityFlags
	// Time is the timestamp of the sample (may be nil if time not available)
	Time *Timestamp
}

// NewBinaryInput creates a new BinaryInput with the given values
func NewBinaryInput(index uint16, value bool, quality QualityFlags) *BinaryInput {
	return &BinaryInput{
		Index:   index,
		Value:   value,
		Quality: quality,
	}
}

// BinaryOutput represents a binary (digital) output data point.
// Binary outputs are single-bit values that can be commanded.
type BinaryOutput struct {
	// Index is the point index within the outstation
	Index uint16
	// Value is the command value (true = operate, false = de-energize)
	Value bool
	// Status is the current status of the output
	Status BinaryOutputStatus
}

// BinaryOutputStatus represents the status of a binary output
type BinaryOutputStatus uint8

const (
	// BinaryOutputQueued indicates the command is queued
	BinaryOutputQueued BinaryOutputStatus = 0x01
	// BinaryOutputVerified indicates the command was verified
	BinaryOutputVerified BinaryOutputStatus = 0x02
	// BinaryOutputSynchError indicates a synchronization error
	BinaryOutputSynchError BinaryOutputStatus = 0x04
	// BinaryOutputBlocked indicates the output is blocked
	BinaryOutputBlocked BinaryOutputStatus = 0x08
	// BinaryOutputLocal indicates the output is in local mode
	BinaryOutputLocal BinaryOutputStatus = 0x10
)

// AnalogInput represents an analog input data point.
// Analog inputs are floating-point values representing measurements.
type AnalogInput struct {
	// Index is the point index within the outstation
	Index uint16
	// Value is the analog measurement value
	Value float64
	// Quality indicates the data quality
	Quality QualityFlags
	// Time is the timestamp of the sample (may be nil if time not available)
	Time *Timestamp
}

// NewAnalogInput creates a new AnalogInput with the given values
func NewAnalogInput(index uint16, value float64, quality QualityFlags) *AnalogInput {
	return &AnalogInput{
		Index:   index,
		Value:   value,
		Quality: quality,
	}
}

// AnalogOutput represents an analog output data point.
// Analog outputs are floating-point values that can be commanded.
type AnalogOutput struct {
	// Index is the point index within the outstation
	Index uint16
	// Value is the command value
	Value float64
	// Status is the current status of the output
	Status AnalogOutputStatus
}

// AnalogOutputStatus represents the status of an analog output
type AnalogOutputStatus uint8

const (
	// AnalogOutputQueued indicates the command is queued
	AnalogOutputQueued AnalogOutputStatus = 0x01
	// AnalogOutputVerified indicates the command was verified
	AnalogOutputVerified AnalogOutputStatus = 0x02
	// AnalogOutputSynchError indicates a synchronization error
	AnalogOutputSynchError AnalogOutputStatus = 0x04
	// AnalogOutputBlocked indicates the output is blocked
	AnalogOutputBlocked AnalogOutputStatus = 0x08
	// AnalogOutputLocal indicates the output is in local mode
	AnalogOutputLocal AnalogOutputStatus = 0x10
	// AnalogOutputPastLimit indicates the value is past limit
	AnalogOutputPastLimit AnalogOutputStatus = 0x20
)

// Counter represents a counter data point.
// Counters are unsigned 32-bit integers that increment over time.
type Counter struct {
	// Index is the point index within the outstation
	Index uint16
	// Value is the current count
	Value uint32
	// Quality indicates the data quality
	Quality QualityFlags
	// Time is the timestamp of the sample (may be nil if time not available)
	Time *Timestamp
}

// NewCounter creates a new Counter with the given values
func NewCounter(index uint16, value uint32, quality QualityFlags) *Counter {
	return &Counter{
		Index:   index,
		Value:   value,
		Quality: quality,
	}
}

// FrozenCounter represents a frozen counter data point.
// Frozen counters are counter values that have been frozen (captured) at a specific time.
type FrozenCounter struct {
	// Index is the point index within the outstation
	Index uint16
	// Value is the frozen count value
	Value uint32
	// Quality indicates the data quality
	Quality QualityFlags
	// Time is the timestamp when the value was frozen
	Time *Timestamp
}

// BinaryOutputEvent represents a change event for a binary output.
type BinaryOutputEvent struct {
	// Index is the point index
	Index uint16
	// Value is the new value
	Value bool
	// Status is the status of the output
	Status BinaryOutputStatus
	// Time is the timestamp of the event
	Time *Timestamp
}

// AnalogOutputEvent represents a change event for an analog output.
type AnalogOutputEvent struct {
	// Index is the point index
	Index uint16
	// Value is the new value
	Value float64
	// Status is the status of the output
	Status AnalogOutputStatus
	// Time is the timestamp of the event
	Time *Timestamp
}

// Timestamp represents a DNP3 timestamp.
// DNP3 uses milliseconds since 2000-01-01 00:00:00 UTC as its time base.
type Timestamp struct {
	// Value is the timestamp in milliseconds since the DNP3 epoch (2000-01-01)
	Value uint64
}

// DNP3Epoch is the DNP3 time epoch (2000-01-01 00:00:00 UTC)
var DNP3Epoch = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

// Time returns the timestamp as a time.Time
func (t *Timestamp) Time() time.Time {
	return DNP3Epoch.Add(time.Duration(t.Value) * time.Millisecond)
}

// Now creates a timestamp from the current time
func (t *Timestamp) Now() *Timestamp {
	return &Timestamp{
		Value: uint64(time.Now().Sub(DNP3Epoch).Milliseconds()),
	}
}

// IsNull returns true if the timestamp is null (not set)
func (t *Timestamp) IsNull() bool {
	return t.Value == 0xFFFFFFFFFFFFFFFF || t.Value == 0
}

// String returns a human-readable string representation
func (t *Timestamp) String() string {
	if t.IsNull() {
		return "null"
	}
	return t.Time().Format(time.RFC3339)
}

// joinStrings is a helper function to join strings
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for _, s := range strs[1:] {
		result += sep + s
	}
	return result
}
