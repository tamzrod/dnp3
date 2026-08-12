package al

import "fmt"

// GroupName returns the human-readable DNP3 object-group name for the given
// group number, or "unknown" when the group is not recognized. Used to make
// unsupported-group/variation error messages descriptive (DNP3-067).
//
// The table covers the groups the v0 profile and the most common DNP3 data
// objects reference; it is not exhaustive but every unrecognized group yields
// a stable "unknown" label so diagnostics remain clear.
func GroupName(group uint8) string {
	switch group {
	case 0:
		return "device attributes"
	case 1:
		return "binary input"
	case 2:
		return "binary input event"
	case 3:
		return "double-bit binary input"
	case 4:
		return "double-bit binary input event"
	case 10:
		return "binary output"
	case 11:
		return "binary output event"
	case 12:
		return "binary control block"
	case 13:
		return "control relay output block"
	case 20:
		return "counter"
	case 21:
		return "counter event"
	case 22:
		return "counter event"
	case 23:
		return "frozen counter"
	case 30:
		return "analog input"
	case 31:
		return "analog input event"
	case 32:
		return "analog input event"
	case 40:
		return "analog output status"
	case 41:
		return "analog output"
	case 42:
		return "analog output event"
	case 50:
		return "time and date"
	case 51:
		return "time and date CTO"
	case 52:
		return "time and date interval"
	case 60:
		return "class data objects (integrity scan)"
	case 70:
		return "file transport"
	case 80:
		return "internal indications (IIN)"
	case 81:
		return "storage module"
	case 82:
		return "device profile"
	case 83:
		return "non-volatile data set"
	case 90:
		return "demand side"
	default:
		return "unknown"
	}
}

// VariationName returns a short label for a known group/variation pair, or
// "" when the variation is not individually named. Callers should always
// include the numeric variation even when a name is returned, because DNP3
// variation numbering is group-relative (DNP3-067).
func VariationName(group, variation uint8) string {
	switch group {
	case 1: // binary input
		switch variation {
		case 0:
			return "any/default"
		case 1:
			return "single-bit binary"
		case 2:
			return "binary with status flags"
		case 3:
			return "binary with relative time"
		case 4:
			return "binary with relative time and sequence"
		}
	case 30: // analog input
		switch variation {
		case 0:
			return "any/default"
		case 1:
			return "32-bit signed"
		case 2:
			return "16-bit signed"
		case 3:
			return "32-bit floating"
		case 4:
			return "16-bit floating"
		case 5:
			return "32-bit double"
		case 6:
			return "32-bit signed with flags"
		case 7:
			return "16-bit signed with flags"
		case 8:
			return "32-bit floating with flags"
		case 9:
			return "16-bit floating with flags"
		}
	case 20: // counter
		switch variation {
		case 0:
			return "any/default"
		case 1:
			return "32-bit signed"
		case 2:
			return "16-bit signed"
		case 5:
			return "32-bit signed with flags"
		case 6:
			return "16-bit signed with flags"
		}
	case 13:
		switch variation {
		case 1:
			return "control relay output block"
		case 2:
			return "pattern control"
		case 3:
			return "pattern mask"
		case 4:
			return "pattern repeat count"
		}
	}
	return ""
}

// DescribeObject returns "group G (name) variation V [vlabel]" for diagnostic
// messages (DNP3-067). The variation label is omitted when unknown.
func DescribeObject(group, variation uint8) string {
	gname := GroupName(group)
	if vlabel := VariationName(group, variation); vlabel != "" {
		return fmt.Sprintf("group %d (%s), variation %d (%s)", group, gname, variation, vlabel)
	}
	return fmt.Sprintf("group %d (%s), variation %d", group, gname, variation)
}
