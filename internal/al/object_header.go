package al

import "fmt"

// ErrUnsupportedQualifier is returned by Encode/DecodeObjectHeader when an
// object-header qualifier byte is outside the v0 allow-list (DNP3-028). The
// supported qualifiers are: 0x00 (8-bit start/stop), 0x06 (all-objects),
// 0x07 (8-bit count), 0x17 (8-bit count + 8-bit index), 0x28 (16-bit range),
// 0x27 (16-bit count).
var ErrUnsupportedQualifier = fmt.Errorf("al: unsupported object qualifier")

// Object-header qualifier codes supported by the v0 profile. The byte values
// follow the IEEE 1815 range/prefix codes. The earlier mapping of 0x00 to an
// 8-bit index qualifier was a wire defect: 0x00 is UINT8_START_STOP, while
// the indexed-count code is 0x17 (UINT8_CNT_UINT8_INDEX).
//
// DNP3 wire fields are LSB-first. The object header is always
// group(1) + variation(1) + qualifier(1) + range(1..N). The range field
// that follows the qualifier byte depends on the qualifier code.
const (
	// QualStartStop8 (0x00) is IEEE 1815 UINT8_START_STOP: an 8-bit start
	// index followed by an 8-bit stop index (5-octet header). The objects
	// that follow are sequential with no per-point index prefix.
	QualStartStop8 uint8 = 0x00
	// QualAllObjects (0x06) requests every point of the group/variation.
	// The range field is a single 0x00 octet.
	QualAllObjects uint8 = 0x06
	// QualCount8 (0x07) carries an 8-bit point count; points are sequential
	// starting at index 0 and carry no per-point index prefix.
	QualCount8 uint8 = 0x07
	// QualIndex8 (0x17) is IEEE 1815 UINT8_CNT_UINT8_INDEX: an 8-bit point
	// count followed by a per-point 8-bit index prefix for each object.
	QualIndex8 uint8 = 0x17
	// QualRange16 (0x28) carries a 16-bit start index and 16-bit stop index,
	// both LSB-first.
	QualRange16 uint8 = 0x28
	// QualCount16 (0x27) carries a 16-bit point count, LSB-first.
	QualCount16 uint8 = 0x27
)

// ObjectHeader models a DNP3 object header.
//
// Fields:
//   - Group:     object group (e.g. 1 = binary input)
//   - Variation: object variation (e.g. 1)
//   - Qualifier: range/prefix qualifier byte (see Qual* constants)
//   - Count:     point count for count/index qualifiers (ignored for QualAllObjects)
//   - Start:     first index for start/stop and range16 qualifiers (LSB-first)
//   - Stop:      last index for start/stop and range16 qualifiers (LSB-first)
type ObjectHeader struct {
	Group     uint8
	Variation uint8
	Qualifier uint8
	Count     uint16
	Start     uint16
	Stop      uint16
}

// EncodedSize returns the number of wire octets the header (including its
// range field) occupies. It does not include the per-point object data that
// may follow the header.
func (h ObjectHeader) EncodedSize() int {
	switch h.Qualifier {
	case QualAllObjects:
		// group, variation, qualifier (all-objects carries no range field)
		return 3
	case QualCount8, QualIndex8:
		// group, variation, qualifier, count(1)
		return 4
	case QualStartStop8:
		// group, variation, qualifier, start(1), stop(1)
		return 5
	case QualRange16:
		// group, variation, qualifier, start(2), stop(2)
		return 7
	case QualCount16:
		// group, variation, qualifier, count(2)
		return 5
	default:
		return 4
	}
}

// Encode appends the wire representation of the object header to dst and
// returns the resulting slice. Only the qualifier codes in the v0 profile are
// supported; unsupported qualifiers return an error.
func (h ObjectHeader) Encode(dst []byte) ([]byte, error) {
	switch h.Qualifier {
	case QualAllObjects:
		// group, variation, 0x06 (all-objects carries no range field)
		return append(dst, h.Group, h.Variation, QualAllObjects), nil
	case QualCount8:
		return append(dst, h.Group, h.Variation, QualCount8, byte(h.Count)), nil
	case QualIndex8:
		return append(dst, h.Group, h.Variation, QualIndex8, byte(h.Count)), nil
	case QualStartStop8:
		// group, variation, 0x00, start(1), stop(1)
		return append(dst, h.Group, h.Variation, QualStartStop8, byte(h.Start), byte(h.Stop)), nil
	case QualRange16:
		// group, variation, 0x28, start(LSB 2), stop(LSB 2)
		return append(dst,
			h.Group, h.Variation, QualRange16,
			byte(h.Start), byte(h.Start>>8),
			byte(h.Stop), byte(h.Stop>>8),
		), nil
	case QualCount16:
		return append(dst,
			h.Group, h.Variation, QualCount16,
			byte(h.Count), byte(h.Count>>8),
		), nil
	default:
		return dst, fmt.Errorf("%w: 0x%02X", ErrUnsupportedQualifier, h.Qualifier)
	}
}

// EncodeObjectHeaders encodes a sequence of object headers to dst.
func EncodeObjectHeaders(dst []byte, headers []ObjectHeader) ([]byte, error) {
	for _, h := range headers {
		var err error
		dst, err = h.Encode(dst)
		if err != nil {
			return dst, err
		}
	}
	return dst, nil
}

// DecodeObjectHeader reads a single object header (including its range field)
// from data starting at offset. It returns the decoded header and the number
// of bytes consumed, or an error for unsupported qualifiers, truncated input,
// or a count/index that exceeds its declared width.
//
// Only the qualifier codes in the v0 profile are accepted; all other
// qualifier bytes return an error.
func DecodeObjectHeader(data []byte, offset int) (ObjectHeader, int, error) {
	if offset < 0 {
		return ObjectHeader{}, 0, fmt.Errorf("al: negative offset %d", offset)
	}
	if offset+3 > len(data) {
		return ObjectHeader{}, 0, fmt.Errorf("al: object header truncated at offset %d: need 3 bytes, have %d", offset, len(data)-offset)
	}
	h := ObjectHeader{
		Group:     data[offset],
		Variation: data[offset+1],
		Qualifier: data[offset+2],
	}
	consumed := 3

	switch h.Qualifier {
	case QualAllObjects:
		// All-objects carries no range field (header is 3 octets).
	case QualCount8, QualIndex8:
		if offset+4 > len(data) {
			return ObjectHeader{}, 0, fmt.Errorf("al: object header truncated at offset %d: need 4 bytes, have %d", offset, len(data)-offset)
		}
		h.Count = uint16(data[offset+3])
		consumed = 4
	case QualStartStop8:
		if offset+5 > len(data) {
			return ObjectHeader{}, 0, fmt.Errorf("al: start/stop header truncated at offset %d: need 5 bytes, have %d", offset, len(data)-offset)
		}
		h.Start = uint16(data[offset+3])
		h.Stop = uint16(data[offset+4])
		if h.Stop < h.Start {
			return ObjectHeader{}, 0, fmt.Errorf("al: start/stop header at offset %d: stop %d < start %d", offset, h.Stop, h.Start)
		}
		consumed = 5
	case QualRange16:
		// 4th octet is unused; start(2 LSB) + stop(2 LSB) follow.
		if offset+7 > len(data) {
			return ObjectHeader{}, 0, fmt.Errorf("al: range16 header truncated at offset %d: need 7 bytes, have %d", offset, len(data)-offset)
		}
		h.Start = uint16(data[offset+3]) | uint16(data[offset+4])<<8
		h.Stop = uint16(data[offset+5]) | uint16(data[offset+6])<<8
		consumed = 7
	case QualCount16:
		// count(2 LSB) follow the qualifier.
		if offset+5 > len(data) {
			return ObjectHeader{}, 0, fmt.Errorf("al: count16 header truncated at offset %d: need 5 bytes, have %d", offset, len(data)-offset)
		}
		h.Count = uint16(data[offset+3]) | uint16(data[offset+4])<<8
		consumed = 5
	default:
		return ObjectHeader{}, 0, fmt.Errorf("%w: 0x%02X", ErrUnsupportedQualifier, h.Qualifier)
	}

	return h, consumed, nil
}

// ValidQualifier reports whether q is a qualifier code supported by the v0
// object-header model.
func ValidQualifier(q uint8) bool {
	switch q {
	case QualStartStop8, QualAllObjects, QualCount8, QualIndex8, QualRange16, QualCount16:
		return true
	default:
		return false
	}
}
