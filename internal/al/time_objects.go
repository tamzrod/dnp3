package al

import (
	"fmt"
	"time"
)

// This file defines the DNP3 time objects used by the logical-clock path:
// Group 50 Variation 1 (Absolute Time and Date) and Group 52 Variation 1
// (Time Delay). All multi-octet fields are LSB-first and the absolute time is
// the same 48-bit millisecond count since 1970-01-01 UTC used by the event
// objects, so the two encodings cannot drift apart.

const (
	// GroupTimeAndDate is Group 50 (time and date).
	GroupTimeAndDate uint8 = 50
	// GroupTimeDelay is Group 52 (time delay / delay measurement).
	GroupTimeDelay uint8 = 52
)

// BuildG50V1Object builds one Group 50 Variation 1 object for a TIME_SYNC,
// RECORD_CURRENT_TIME, or WRITE request/response: a count-1 object header
// followed by the 48-bit absolute time.
//
// Wire form: 32 01 07 01 <6-byte DNP time>.
func BuildG50V1Object(t time.Time) ([]byte, error) {
	enc, err := EncodeDNPTime(t)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 0, 4+len(enc))
	out = append(out, GroupTimeAndDate, 1, QualCount8, 1)
	out = append(out, enc[:]...)
	return out, nil
}

// ParseG50V1Object parses a single Group 50 Variation 1 absolute time from an
// application request body. It accepts the count-1 (0x07), index-1 (0x17),
// all-objects (0x06) and single-index range (0x28, start == stop) forms used by
// common masters, and requires exactly one time object. A missing, malformed,
// multi-point, or wrong-group/variation body is an error so a time sync cannot
// be acknowledged without a definite time value.
func ParseG50V1Object(data []byte) (time.Time, error) {
	if len(data) < 3 {
		return time.Time{}, fmt.Errorf("al: G50V1 request too short")
	}
	h, consumed, err := DecodeObjectHeader(data, 0)
	if err != nil {
		return time.Time{}, fmt.Errorf("al: G50V1 header: %w", err)
	}
	if h.Group != GroupTimeAndDate || h.Variation != 1 {
		return time.Time{}, fmt.Errorf("al: G50V1 expected group 50 var 1, got %d var %d", h.Group, h.Variation)
	}
	// Determine the point count and the offset of the first object.
	var count uint16
	objOffset := consumed
	switch h.Qualifier {
	case QualCount8, QualCount16, QualIndex8:
		count = h.Count
	case QualAllObjects:
		count = 1
	case QualRange16:
		if h.Start != h.Stop {
			return time.Time{}, fmt.Errorf("al: G50V1 range must cover exactly one point")
		}
		count = 1
	case QualStartStop8:
		if h.Start != h.Stop {
			return time.Time{}, fmt.Errorf("al: G50V1 range must cover exactly one point")
		}
		count = 1
	default:
		return time.Time{}, fmt.Errorf("al: G50V1 unsupported qualifier 0x%02X", h.Qualifier)
	}
	if count != 1 {
		return time.Time{}, fmt.Errorf("al: G50V1 requires exactly one point, got %d", count)
	}
	// An index prefix (0x17, 0x28) precedes the six-octet time value.
	switch h.Qualifier {
	case QualIndex8:
		objOffset++ // one-octet index
	case QualRange16:
		// DecodeObjectHeader already consumed start/stop; no per-object prefix.
	}
	if objOffset+6 > len(data) {
		return time.Time{}, fmt.Errorf("al: G50V1 time value truncated")
	}
	return DecodeDNPTime(data[objOffset : objOffset+6])
}

// BuildG52V1Object builds one Group 52 Variation 1 delay-measurement object.
// The 16-bit value is the measured round-trip delay in milliseconds.
// Wire form: 34 01 07 01 <delay-lo> <delay-hi>.
func BuildG52V1Object(delayMS uint16) []byte {
	out := make([]byte, 0, 6)
	out = append(out, GroupTimeDelay, 1, QualCount8, 1)
	out = append(out, byte(delayMS), byte(delayMS>>8))
	return out
}

// ParseG52V1Object parses a single Group 52 Variation 1 delay value from an
// application request body. A zero request value is the normal master probe.
func ParseG52V1Object(data []byte) (uint16, error) {
	if len(data) < 6 {
		return 0, fmt.Errorf("al: G52V1 request too short")
	}
	h, consumed, err := DecodeObjectHeader(data, 0)
	if err != nil {
		return 0, fmt.Errorf("al: G52V1 header: %w", err)
	}
	if h.Group != GroupTimeDelay || h.Variation != 1 || h.Count != 1 {
		return 0, fmt.Errorf("al: G52V1 expected one group 52 var 1 object")
	}
	if consumed+2 > len(data) {
		return 0, fmt.Errorf("al: G52V1 delay value truncated")
	}
	return uint16(data[consumed]) | uint16(data[consumed+1])<<8, nil
}
