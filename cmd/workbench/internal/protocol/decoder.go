// Package protocol provides DNP3 protocol decoding and formatting.
package protocol

import (
	"fmt"
	"strings"
)

// DecodedFrame represents a fully decoded DNP3 frame.
type DecodedFrame struct {
	DLL    *DecodedDLL
	TL     *DecodedTL
	AL     *DecodedAL
	RawHex string
}

// DecodedDLL represents decoded Data Link Layer.
type DecodedDLL struct {
	Direction   string // "Master→Outstation" or "Outstation→Master"
	Primary      bool   // Primary (request) or Secondary (response)
	FrameType    string // e.g., "Reset Link", "Confirmed User Data"
	FCB         bool   // Frame count bit
	FCV         bool   // Frame count valid
	FunctionCode uint8
	FunctionName string
	Destination uint16
	Source      uint16
	Length      uint16
	CRC         uint16
	RawHex      string
}

// DecodedTL represents decoded Transport Layer.
type DecodedTL struct {
	FIR bool // First fragment
	FIN bool // Final fragment
	CON bool // Confirmation required
	UNS bool // Unsolicited
	Seq  uint8
	RawHex string
}

// DecodedAL represents decoded Application Layer.
type DecodedAL struct {
	FunctionCode   uint8
	FunctionName   string
	InternalInd1   uint8
	InternalInd2   uint8
	IINBits        []string
	Objects        []ObjectHeader
	RawHex         string
}

// ObjectHeader represents an object header in AL.
type ObjectHeader struct {
	Group      uint8
	Variation  uint8
	Qualifier  uint8
	Range      string
	Count      uint16
}

// Decoder decodes DNP3 protocol frames.
type Decoder struct{}

// NewDecoder creates a new decoder.
func NewDecoder() *Decoder {
	return &Decoder{}
}

// Decode decodes raw bytes into a DecodedFrame.
func (d *Decoder) Decode(data []byte) *DecodedFrame {
	if len(data) < 10 {
		return &DecodedFrame{
			RawHex: FormatHex(data),
		}
	}

	frame := &DecodedFrame{
		RawHex: FormatHex(data),
	}

	// Decode DLL (first 10 bytes + CRC)
	if len(data) >= 10 {
		frame.DLL = d.decodeDLL(data[:10])
		frame.DLL.RawHex = FormatHex(data[:10])
	}

	// Decode TL (byte 10, if present)
	if len(data) > 10 {
		frame.TL = d.decodeTL(data[10])
	}

	// Decode AL (remaining bytes after TL)
	if len(data) > 11 {
		frame.AL = d.decodeAL(data[11:])
	}

	return frame
}

// decodeDLL decodes Data Link Layer.
func (d *Decoder) decodeDLL(data []byte) *DecodedDLL {
	if len(data) < 10 {
		return nil
	}

	dll := &DecodedDLL{}

	// Start byte (0x0564)
	start := data[0]
	_ = start

	// Length
	dll.Length = uint16(data[1])

	// Control byte
	ctrl := data[2]
	dll.Direction = "Master→Outstation"
	if ctrl&0x40 != 0 { // DIR bit
		dll.Direction = "Outstation→Master"
	}
	dll.Primary = ctrl&0x80 == 0 // PRM=0 means secondary

	dll.FCB = ctrl&0x20 != 0
	dll.FCV = ctrl&0x10 != 0

	funcCode := ctrl & 0x0F
	dll.FunctionCode = funcCode
	dll.FunctionName = d.dllFunctionName(funcCode, dll.Primary)

	// Destination
	dll.Destination = uint16(data[3]) | (uint16(data[4]) << 8)

	// Source
	dll.Source = uint16(data[5]) | (uint16(data[6]) << 8)

	// CRC
	dll.CRC = uint16(data[8]) | (uint16(data[9]) << 8)

	return dll
}

// decodeTL decodes Transport Layer.
func (d *Decoder) decodeTL(b byte) *DecodedTL {
	tl := &DecodedTL{}

	tl.FIR = b&0x80 != 0
	tl.FIN = b&0x40 != 0
	tl.CON = b&0x20 != 0
	tl.UNS = b&0x10 != 0
	tl.Seq = b & 0x0F

	return tl
}

// decodeAL decodes Application Layer.
func (d *Decoder) decodeAL(data []byte) *DecodedAL {
	if len(data) < 2 {
		return nil
	}

	al := &DecodedAL{}

	// Application control
	ac := data[0]
	_ = ac

	// Function code
	al.FunctionCode = data[1]
	al.FunctionName = d.alFunctionName(al.FunctionCode)

	// Internal indication (if response)
	if len(data) >= 4 && (al.FunctionCode == 0x81 || al.FunctionCode == 0x82 || al.FunctionCode == 0x83) {
		al.InternalInd1 = data[2]
		al.InternalInd2 = data[3]
		al.IINBits = d.parseIIN(al.InternalInd1, al.InternalInd2)
	}

	return al
}

// dllFunctionName returns the DLL function name.
func (d *Decoder) dllFunctionName(code uint8, primary bool) string {
	if primary {
		// Primary functions
		switch code {
		case 0x00:
			return "Reset Link"
		case 0x01:
			return "Reset Station"
		case 0x02:
			return "Start Application"
		case 0x03:
			return "Stop Application"
		case 0x04:
			return "Delay Measure"
		case 0x09:
			return "Link Status"
		case 0x0A:
			return "Reset FCS"
		case 0x0B:
			return "Unsolicited Status"
		default:
			if code >= 0x05 && code <= 0x08 {
				return "Request" // Class 1-4 poll
			}
			return fmt.Sprintf("Unknown(0x%02X)", code)
		}
	}
	// Secondary functions
	switch code {
	case 0x00:
		return "ACK"
	case 0x01:
		return "NACK"
	case 0x0B:
		return "Link Status"
	case 0x0F:
		return "Not Supported"
	default:
		return fmt.Sprintf("Unknown(0x%02X)", code)
	}
}

// alFunctionName returns the AL function name.
func (d *Decoder) alFunctionName(code uint8) string {
	switch code {
	case 0x01:
		return "READ"
	case 0x02:
		return "WRITE"
	case 0x03:
		return "SELECT"
	case 0x04:
		return "OPERATE"
	case 0x05:
		return "DIRECT OPERATE"
	case 0x06:
		return "DIRECT OPERATE NO ACK"
	case 0x07:
		return "FREEZE"
	case 0x08:
		return "FREEZE NO ACK"
	case 0x09:
		return "FREEZE CLEAR"
	case 0x0A:
		return "FREEZE CLEAR NO ACK"
	case 0x0B:
		return "FREEZE AT TIME"
	case 0x0C:
		return "FREEZE AT TIME NO ACK"
	case 0x14:
		return "ENABLE UNSOLICITED"
	case 0x15:
		return "DISABLE UNSOLICITED"
	case 0x19:
		return "ASSIGN CLASS"
	case 0x20:
		return "CLOSE APPLICATION"
	case 0x21:
		return "OPEN APPLICATION"
	case 0x28:
		return "TIME SYNC"
	case 0x81:
		return "RESPONSE"
	case 0x82:
		return "UNSOLICITED RESPONSE"
	case 0x83:
		return "AUTHENTICATION"
	default:
		return fmt.Sprintf("Unknown(0x%02X)", code)
	}
}

// parseIIN parses Internal Indication bits.
func (d *Decoder) parseIIN(b1, b2 uint8) []string {
	var bits []string

	// Byte 1 bits
	if b1&0x01 != 0 {
		bits = append(bits, "ALL_STATIONS")
	}
	if b1&0x04 != 0 {
		bits = append(bits, "CLASS1_EVENTS")
	}
	if b1&0x08 != 0 {
		bits = append(bits, "CLASS2_EVENTS")
	}
	if b1&0x10 != 0 {
		bits = append(bits, "CLASS3_EVENTS")
	}
	if b1&0x20 != 0 {
		bits = append(bits, "NEED_TIME")
	}
	if b1&0x40 != 0 {
		bits = append(bits, "LOCAL_CONTROL")
	}
	if b1&0x80 != 0 {
		bits = append(bits, "DEVICE_TROUBLE")
	}

	// Byte 2 bits
	if b2&0x01 != 0 {
		bits = append(bits, "DEVICE_RESTART")
	}
	if b2&0x04 != 0 {
		bits = append(bits, "COUNTER_ROLLOVER")
	}
	if b2&0x08 != 0 {
		bits = append(bits, "MEMORY_ERROR")
	}
	if b2&0x10 != 0 {
		bits = append(bits, "CONFIGURATION_ERROR")
	}
	if b2&0x20 != 0 {
		bits = append(bits, "EXTERNAL_ERROR")
	}
	if b2&0x40 != 0 {
		bits = append(bits, "BUFFER_OVERFLOW")
	}

	if len(bits) == 0 {
		bits = append(bits, "NO_ISSUES")
	}

	return bits
}

// FormatFrame formats a decoded frame for display.
func FormatFrame(f *DecodedFrame) string {
	var lines []string

	lines = append(lines, "=== DNP3 Frame ===")
	lines = append(lines, fmt.Sprintf("Raw: %s", f.RawHex))

	if f.DLL != nil {
		lines = append(lines, "")
		lines = append(lines, "[Data Link Layer]")
		lines = append(lines, fmt.Sprintf("  Direction:  %s", f.DLL.Direction))
		lines = append(lines, fmt.Sprintf("  Type:       %s", f.DLL.FunctionName))
		lines = append(lines, fmt.Sprintf("  Dest:       %d (0x%04X)", f.DLL.Destination, f.DLL.Destination))
		lines = append(lines, fmt.Sprintf("  Src:        %d (0x%04X)", f.DLL.Source, f.DLL.Source))
		lines = append(lines, fmt.Sprintf("  Length:     %d bytes", f.DLL.Length))
		if f.DLL.FCB {
			lines = append(lines, fmt.Sprintf("  FCB:        %d", boolToInt(f.DLL.FCB)))
		}
	}

	if f.TL != nil {
		lines = append(lines, "")
		lines = append(lines, "[Transport Layer]")
		lines = append(lines, fmt.Sprintf("  FIR:        %v", f.TL.FIR))
		lines = append(lines, fmt.Sprintf("  FIN:        %v", f.TL.FIN))
		lines = append(lines, fmt.Sprintf("  CON:        %v", f.TL.CON))
		lines = append(lines, fmt.Sprintf("  UNS:        %v", f.TL.UNS))
		lines = append(lines, fmt.Sprintf("  Sequence:   %d", f.TL.Seq))
	}

	if f.AL != nil {
		lines = append(lines, "")
		lines = append(lines, "[Application Layer]")
		lines = append(lines, fmt.Sprintf("  Function:   %s (0x%02X)", f.AL.FunctionName, f.AL.FunctionCode))
		if len(f.AL.IINBits) > 0 {
			lines = append(lines, fmt.Sprintf("  IIN:        %s", strings.Join(f.AL.IINBits, ", ")))
		}
	}

	return strings.Join(lines, "\n")
}

// FormatHex formats bytes as hex string.
func FormatHex(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	var parts []string
	for _, b := range data {
		parts = append(parts, fmt.Sprintf("%02X", b))
	}
	return strings.Join(parts, " ")
}

// FormatHexSpaced formats bytes as spaced hex string.
func FormatHexSpaced(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	var parts []string
	for _, b := range data {
		parts = append(parts, fmt.Sprintf("%02X", b))
	}
	return strings.Join(parts, " ")
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
