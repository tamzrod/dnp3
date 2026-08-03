// Package events implements the DNP3 Event Subsystem.
// See event.go for package documentation.
package events

import (
	"encoding/binary"
	"time"
)

// EventBuilder encodes events into DNP3 object format.
type EventBuilder struct{}

// NewEventBuilder creates a new EventBuilder.
func NewEventBuilder() *EventBuilder {
	return &EventBuilder{}
}

// BuildEventObjects encodes events into DNP3 object data.
// Returns the encoded bytes suitable for inclusion in an APDU.
func (b *EventBuilder) BuildEventObjects(events []Event) []byte {
	if len(events) == 0 {
		return nil
	}

	// Group events by their object group
	eventGroups := make(map[uint8][]Event)
	for _, event := range events {
		eventGroups[event.Group] = append(eventGroups[event.Group], event)
	}

	var result []byte

	// Process each group
	for group, groupEvents := range eventGroups {
		if len(groupEvents) == 0 {
			continue
		}

		// Build object header
		// Group, Variation, Qualifier (0x17=prefix count-8), Count
		variation := groupEvents[0].Variation
		header := []byte{group, variation, 0x17, byte(len(groupEvents))}
		result = append(result, header...)

		// Build event data for this group
		for _, event := range groupEvents {
			eventData := b.encodeEvent(event)
			result = append(result, eventData...)
		}
	}

	return result
}

// encodeEvent encodes a single event into DNP3 format.
func (b *EventBuilder) encodeEvent(event Event) []byte {
	// Index (2 bytes big-endian)
	indexBytes := []byte{byte(event.Index >> 8), byte(event.Index & 0xFF)}

	// Value encoding depends on point type
	var valueBytes []byte
	switch event.PointType {
	case PointTypeBinaryInput, PointTypeBinaryOutput:
		// Binary: 1 byte (bit 7 = value, bits 0-6 = quality)
		val := uint8(0)
		if len(event.Value) > 0 && event.Value[0] != 0 {
			val = 0x80 // Set bit 7 for true
		}
		val |= event.Quality & 0x7F
		valueBytes = []byte{val}

	case PointTypeCounter:
		// Counter: 4 bytes big-endian + 1 byte quality + 6 bytes time
		valueBytes = make([]byte, 0, 11)
		if len(event.Value) >= 4 {
			valueBytes = append(valueBytes, event.Value[:4]...)
		} else {
			// Pad with zeros
			valueBytes = append(valueBytes, 0, 0, 0, 0)
			if len(event.Value) > 0 {
				copy(valueBytes, event.Value)
			}
		}
		valueBytes = append(valueBytes, event.Quality)

	case PointTypeAnalogInput, PointTypeAnalogOutput:
		// Analog: 4 bytes IEEE 754 float32 + 1 byte quality + 6 bytes time
		valueBytes = make([]byte, 0, 11)
		if len(event.Value) >= 4 {
			valueBytes = append(valueBytes, event.Value[:4]...)
		} else if len(event.Value) == 8 {
			// 64-bit float - convert to 32-bit
			f64 := binary.BigEndian.Uint64(event.Value)
			f32 := float32FromBits(uint32(f64))
			f32Bytes := make([]byte, 4)
			binary.BigEndian.PutUint32(f32Bytes, f32)
			valueBytes = append(valueBytes, f32Bytes...)
		} else {
			// Pad with zeros
			valueBytes = append(valueBytes, 0, 0, 0, 0)
		}
		valueBytes = append(valueBytes, event.Quality)

	default:
		// Default encoding
		valueBytes = event.Value
		if len(valueBytes) == 0 {
			valueBytes = []byte{0}
		}
	}

	// Time (6 bytes DNP3 time format)
	timeBytes := encodeDNPTimestamp(event.Time)

	// Combine: index + value + time
	return append(append(indexBytes, valueBytes...), timeBytes...)
}

// encodeDNPTimestamp encodes a time.Time into DNP3 timestamp format.
// DNP3 uses milliseconds since 2000-01-01 00:00:00 UTC
func encodeDNPTimestamp(t time.Time) []byte {
	// DNP3 epoch
	epoch := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	// Calculate milliseconds since epoch
	ms := t.Sub(epoch).Milliseconds()

	// DNP3 time is 2 x 4-byte longs: first 4 bytes = relative time, last 4 bytes = absolute time
	// For simplicity, we use absolute time (milliseconds since epoch)
	// Format: uint48 absolute time (ms) + uint16 relative time (ms)
	result := make([]byte, 6)

	// Absolute time (40 bits, milliseconds since epoch)
	result[0] = byte(ms >> 32) // Upper 8 bits of ms
	result[1] = byte(ms >> 24) // Next 8 bits
	result[2] = byte(ms >> 16) // Next 8 bits
	result[3] = byte(ms >> 8)  // Next 8 bits
	result[4] = byte(ms)       // Lower 8 bits

	// Relative time (8 bits, ms since last event - simplified)
	result[5] = 0

	return result
}

// float32FromBits converts uint32 bits to float32.
func float32FromBits(bits uint32) uint32 {
	return bits
}
