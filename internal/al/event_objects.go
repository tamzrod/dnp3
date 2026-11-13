package al

import (
	"fmt"
	"time"
)

// DNPEpoch is the DNP3 absolute-time origin: 2000-01-01 00:00:00 UTC.
var DNPEpoch = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

// DNPTimeMax is the largest millisecond count that fits in a 48-bit DNP time.
const DNPTimeMax = uint64(0xFFFFFFFFFFFF)

// Binary input event flag bits (Group 2). Bit 0x80 is the state.
const (
	BinEventState        uint8 = 0x80
	BinEventOnline       uint8 = 0x01
	BinEventRestart      uint8 = 0x02
	BinEventCommLost     uint8 = 0x04
	BinEventRemoteForced uint8 = 0x08
	BinEventLocalForced  uint8 = 0x10
	BinEventChatter      uint8 = 0x20
)

// Analog input event flag bits (Group 32).
const (
	AnalogEventOnline       uint8 = 0x01
	AnalogEventRestart      uint8 = 0x02
	AnalogEventCommLost     uint8 = 0x04
	AnalogEventRemoteForced uint8 = 0x08
	AnalogEventLocalForced  uint8 = 0x10
	AnalogEventOverRange    uint8 = 0x20
	AnalogEventReferenceErr uint8 = 0x40
)

// Counter event flag bits (Group 22).
const (
	CounterEventOnline        uint8 = 0x01
	CounterEventRestart       uint8 = 0x02
	CounterEventCommLost      uint8 = 0x04
	CounterEventRemoteForced  uint8 = 0x08
	CounterEventLocalForced   uint8 = 0x10
	CounterEventRollover      uint8 = 0x20
	CounterEventDiscontinuity uint8 = 0x40
)

// EventObject is one decoded Group 2, 22, or 32 event.
type EventObject struct {
	Group     uint8
	Variation uint8
	Index     uint16
	Flags     uint8
	Analog    int32
	Counter   uint32
	HasTime   bool
	Time      time.Time
}

// EncodeDNPTime encodes an absolute DNP time as 6 little-endian octets.
// A zero time, a time before the epoch, and a value that does not fit in
// 48 bits are errors.
func EncodeDNPTime(t time.Time) ([6]byte, error) {
	var out [6]byte
	if t.IsZero() {
		return out, fmt.Errorf("al: zero time has no absolute DNP timestamp")
	}
	if t.Before(DNPEpoch) {
		return out, fmt.Errorf("al: time before DNP epoch")
	}
	// Unix milliseconds avoid time.Duration, which overflows after ~290 years.
	ms := t.UnixMilli() - DNPEpoch.UnixMilli()
	if ms < 0 || uint64(ms) > DNPTimeMax {
		return out, fmt.Errorf("al: time does not fit in 48 bits")
	}
	u := uint64(ms)
	for i := 0; i < 6; i++ {
		out[i] = byte(u)
		u >>= 8
	}
	return out, nil
}

// DecodeDNPTime decodes 6 little-endian DNP time octets.
func DecodeDNPTime(b []byte) (time.Time, error) {
	if len(b) < 6 {
		return time.Time{}, fmt.Errorf("al: DNP time requires 6 octets")
	}
	var u uint64
	for i := 5; i >= 0; i-- {
		u = (u << 8) | uint64(b[i])
	}
	// time.Duration counts nanoseconds and overflows about 290 years after
	// the epoch, so a valid 48-bit count must not be added that way.
	// Six octets are at most 48 bits, which still fits in int64 milliseconds.
	return time.UnixMilli(DNPEpoch.UnixMilli() + int64(u)).UTC(), nil
}

// AppendIndexedEvents appends one qualifier-0x17 object header and its
// indexed objects. Each object is index (1) followed by the raw object octets.
// count must match len(objects). Indexes inside objects are the caller's.
func AppendIndexedEvents(dst []byte, group, variation uint8, objects [][]byte) ([]byte, error) {
	if len(objects) > 255 {
		return dst, fmt.Errorf("al: event count %d exceeds 8-bit qualifier", len(objects))
	}
	dst = append(dst, group, variation, QualIndex8, byte(len(objects)))
	for _, obj := range objects {
		dst = append(dst, obj...)
	}
	return dst, nil
}

// PutBinaryEventObject encodes one Group 2 object without the index prefix.
// withTime selects Variation 2; otherwise Variation 1.
func PutBinaryEventObject(flags uint8, withTime bool, t time.Time) ([]byte, error) {
	if !withTime {
		return []byte{flags}, nil
	}
	ts, err := EncodeDNPTime(t)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 1+6)
	out[0] = flags
	copy(out[1:], ts[:])
	return out, nil
}

// PutAnalogEventObject encodes one Group 32 Variation 1 or 3 object
// (flags, int32 little-endian, optional absolute time).
func PutAnalogEventObject(flags uint8, value int32, withTime bool, t time.Time) ([]byte, error) {
	u := uint32(value)
	out := []byte{flags, byte(u), byte(u >> 8), byte(u >> 16), byte(u >> 24)}
	if !withTime {
		return out, nil
	}
	ts, err := EncodeDNPTime(t)
	if err != nil {
		return nil, err
	}
	out = append(out, ts[:]...)
	return out, nil
}

// PutCounterEventObject encodes one Group 22 Variation 1 or 5 object
// (flags, uint32 little-endian, optional absolute time).
func PutCounterEventObject(flags uint8, value uint32, withTime bool, t time.Time) ([]byte, error) {
	out := []byte{flags, byte(value), byte(value >> 8), byte(value >> 16), byte(value >> 24)}
	if !withTime {
		return out, nil
	}
	ts, err := EncodeDNPTime(t)
	if err != nil {
		return nil, err
	}
	out = append(out, ts[:]...)
	return out, nil
}

func eventObjectSize(group, variation uint8) int {
	switch group {
	case 2:
		switch variation {
		case 1:
			return 1
		case 2:
			return 7
		}
	case 22:
		switch variation {
		case 1:
			return 5
		case 5:
			return 11
		}
	case 32:
		switch variation {
		case 1:
			return 5
		case 3:
			return 11
		}
	}
	return -1
}

// qualCountIndex16 is IEEE 1815 UINT16_CNT_UINT16_INDEX. OpenDNP3 3.1.2
// event writers emit this qualifier: a 16-bit object count, then a 16-bit
// index prefix on every object. The event decoder reads that layout directly.
// object_header.QualRange16 uses the same octet for a different header model
// and is not used here.
const qualCountIndex16 uint8 = 0x28

// DecodeEventObjects parses Group 2/22/32 events encoded with qualifier 0x17
// (8-bit count and index) or 0x28 (16-bit count and index). An empty buffer
// is a valid object-less class response.
func DecodeEventObjects(data []byte) ([]EventObject, error) {
	var out []EventObject
	for len(data) > 0 {
		if len(data) < 3 {
			return nil, fmt.Errorf("al: truncated event header")
		}
		group, variation, qual := data[0], data[1], data[2]
		var n, indexWidth, header int
		switch qual {
		case QualIndex8:
			if len(data) < 4 {
				return nil, fmt.Errorf("al: truncated event header")
			}
			n = int(data[3])
			indexWidth = 1
			header = 4
		case qualCountIndex16:
			if len(data) < 5 {
				return nil, fmt.Errorf("al: truncated event header")
			}
			n = int(data[3]) | int(data[4])<<8
			indexWidth = 2
			header = 5
		default:
			return nil, fmt.Errorf("al: event qualifier 0x%02X is not 0x17 or 0x28", qual)
		}
		size := eventObjectSize(group, variation)
		if size < 0 {
			return nil, fmt.Errorf("al: unsupported event g%d v%d", group, variation)
		}
		data = data[header:]
		for i := 0; i < n; i++ {
			if len(data) < indexWidth+size {
				return nil, fmt.Errorf("al: truncated event object")
			}
			var index uint16
			if indexWidth == 1 {
				index = uint16(data[0])
			} else {
				index = uint16(data[0]) | uint16(data[1])<<8
			}
			body := data[indexWidth : indexWidth+size]
			data = data[indexWidth+size:]
			ev := EventObject{Group: group, Variation: variation, Index: index, Flags: body[0]}
			switch group {
			case 32:
				ev.Analog = int32(uint32(body[1]) | uint32(body[2])<<8 | uint32(body[3])<<16 | uint32(body[4])<<24)
				if variation == 3 {
					tm, err := DecodeDNPTime(body[5:11])
					if err != nil {
						return nil, err
					}
					ev.HasTime = true
					ev.Time = tm
				}
			case 22:
				ev.Counter = uint32(body[1]) | uint32(body[2])<<8 | uint32(body[3])<<16 | uint32(body[4])<<24
				if variation == 5 {
					tm, err := DecodeDNPTime(body[5:11])
					if err != nil {
						return nil, err
					}
					ev.HasTime = true
					ev.Time = tm
				}
			case 2:
				if variation == 2 {
					tm, err := DecodeDNPTime(body[1:7])
					if err != nil {
						return nil, err
					}
					ev.HasTime = true
					ev.Time = tm
				}
			}
			out = append(out, ev)
		}
	}
	return out, nil
}
