// Package frame implements DNP3 Data Link Layer frame encoding and decoding.
//
// A DNP3 Data Link Layer frame consists of:
//   - Start bytes (0x05 0x64)
//   - Length (1 byte)
//   - Control byte (1 byte)
//   - Destination address (2 bytes, big-endian)
//   - Source address (2 bytes, big-endian)
//   - User data (0-292 bytes)
//   - CRC bytes (2 bytes per 16-bit quantity)
//
// Reference: IEEE 1815-2012 Section 5.2
package frame

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"dnp3/internal/dll/crc"
)

// Sync bytes indicating the start of a DNP3 frame
const (
	SyncByte1 = 0x05
	SyncByte2 = 0x64
)

// Maximum frame sizes per IEEE 1815-2012
const (
	// MaxFrameSize is the maximum total frame size (292 data + 10 header)
	MaxFrameSize = 302

	// MaxDataSize is the maximum user data size
	MaxDataSize = 292

	// HeaderSize is the size without data but with CRC
	HeaderSize = 10

	// MinFrameSize is the minimum valid frame size
	MinFrameSize = 10
)

// Function codes for primary station (PRM=1)
const (
	FuncResetLinkStations   = 0
	FuncResetLinkStatus     = 1
	FuncUnsolicitedTestFunc = 2
	FuncReturnLinkStatus    = 3
	FuncConfirmedUserData   = 4
)

// Function codes for secondary station (PRM=0)
const (
	FuncAck               = 0
	FuncNack              = 1
	FuncLinkStatus        = 2
	FuncNotSupported       = 3
	FuncConfirmedUserDataR = 4
)

// Special addresses
const (
	AddrBroadcast    = 0xFFFF
	AddrAllReset     = 0xFFFA
	AddrUnconfigured  = 0x0000
	AddrVirtualTerm   = 0xFFFB
	AddrSecChannel    = 0xFFFC
	AddrPriChannel    = 0xFFFD
	AddrReserved      = 0xFFFE
)

// Frame represents a DNP3 Data Link Layer frame.
type Frame struct {
	// Control contains the control byte fields
	Control Control

	// Destination address (big-endian)
	DestAddr uint16

	// Source address (big-endian)
	SrcAddr uint16

	// User data (application layer data)
	Data []byte
}

// Control represents the DNP3 control byte.
type Control struct {
	// DIR: Direction (1=Master-to-Outstation, 0=Outstation-to-Master)
	DIR bool

	// PRM: Primary (1=Primary station initiates, 0=Secondary responds)
	PRM bool

	// FCB: Frame Count Bit (used for confirmation in balanced mode)
	FCB bool

	// FCV: Frame Count Bit Valid (1=FCB is meaningful)
	FCV bool

	// DFC: Data Link Busy / Flow Control
	DFC bool

	// FuncCode: Function code (bits 5-3)
	FuncCode uint8
}

// ToByte converts the Control byte to a single byte.
//
// DNP3 Control byte format (IEEE 1815-2012 Section 5.2):
//   Bit 7: DIR (Direction)
//   Bit 6: PRM (Primary)
//   Bit 5: FCB (Frame Count Bit)
//   Bit 4: FCV (Frame Count Bit Valid)
//   Bits 3-0: Function Code (4 bits, primary functions 0-15)
func (c Control) ToByte() byte {
	var b byte

	if c.DIR {
		b |= 0x80
	}
	if c.PRM {
		b |= 0x40
	}
	if c.FCB {
		b |= 0x20
	}
	if c.FCV {
		b |= 0x10
	}

	// FuncCode occupies bits 3-0 (4 bits)
	b |= c.FuncCode & 0x0F

	return b
}

// FromByte parses a Control byte.
func (c *Control) FromByte(b byte) {
	c.DIR = (b & 0x80) != 0
	c.PRM = (b & 0x40) != 0
	c.FCB = (b & 0x20) != 0
	c.FCV = (b & 0x10) != 0
	// Function code is in bits 3-0 (4 bits)
	c.FuncCode = b & 0x0F
}

// IsResetLinkStations returns true if this is a Reset Link Stations frame.
func (c Control) IsResetLinkStations() bool {
	return c.PRM && c.FuncCode == FuncResetLinkStations && !c.FCV
}

// IsConfirmedUserData returns true if this is a Confirmed User Data frame.
func (c Control) IsConfirmedUserData() bool {
	return c.PRM && c.FuncCode == FuncConfirmedUserData && c.FCV
}

// IsAck returns true if this is an Acknowledgment frame.
func (c Control) IsAck() bool {
	return !c.PRM && c.FuncCode == FuncAck
}

// Encode encodes a Frame into bytes ready for transmission.
//
// The encoded frame format:
//   - 0x05 0x64 (sync bytes)
//   - Length (1 byte)
//   - Control (1 byte)
//   - Destination (2 bytes, big-endian)
//   - Source (2 bytes, big-endian)
//   - Data (0-292 bytes)
//   - CRC (2 bytes per 16-bit quantity, LSB first)
func Encode(f *Frame) ([]byte, error) {
	// Validate data size
	if len(f.Data) > MaxDataSize {
		return nil, fmt.Errorf("data size %d exceeds maximum %d", len(f.Data), MaxDataSize)
	}

	// Calculate total frame size
	// 2 (sync) + 1 (length) + 1 (control) + 2 (dest) + 2 (src) + data + CRCs
	numCRCGroups := 1 + 1 + 1 + (len(f.Data)+1)/2 // Length+Ctrl, Dest, Src, Data pairs
	frameSize := 2 + 1 + 1 + 2 + 2 + len(f.Data) + (numCRCGroups * 2)

	buf := make([]byte, 0, frameSize)

	// Write sync bytes
	buf = append(buf, SyncByte1, SyncByte2)

	// Calculate length: bytes from Control to end of Data
	length := byte(1 + 2 + 2 + len(f.Data)) // Control + Dest + Src + Data
	buf = append(buf, length)

	// Write control byte
	buf = append(buf, f.Control.ToByte())

	// Write destination address (big-endian)
	buf = binary.BigEndian.AppendUint16(buf, f.DestAddr)

	// Write source address (big-endian)
	buf = binary.BigEndian.AppendUint16(buf, f.SrcAddr)

	// Write data
	if len(f.Data) > 0 {
		buf = append(buf, f.Data...)
	}

	// Calculate and append CRCs
	// CRC 1: Length + Control
	buf = appendCRC16(buf, length, f.Control.ToByte())

	// CRC 2: Destination Address (both bytes)
	buf = appendCRC16(buf, byte(f.DestAddr>>8), byte(f.DestAddr))

	// CRC 3: Source Address (both bytes)
	buf = appendCRC16(buf, byte(f.SrcAddr>>8), byte(f.SrcAddr))

	// CRC for data (2 bytes at a time)
	for i := 0; i < len(f.Data); i += 2 {
		var b1, b2 byte
		b1 = f.Data[i]
		if i+1 < len(f.Data) {
			b2 = f.Data[i+1]
		}
		buf = appendCRC16(buf, b1, b2)
	}

	return buf, nil
}

// Decode decodes bytes into a Frame.
func Decode(data []byte) (*Frame, error) {
	if len(data) < MinFrameSize {
		return nil, fmt.Errorf("frame too short: %d bytes, minimum %d", len(data), MinFrameSize)
	}

	// Verify sync bytes
	if data[0] != SyncByte1 || data[1] != SyncByte2 {
		return nil, fmt.Errorf("invalid sync bytes: 0x%02x 0x%02x, expected 0x05 0x64", data[0], data[1])
	}

	offset := 2

	// Read length
	length := data[offset]
	offset++

	// Validate length
	expectedMinLen := byte(1 + 2 + 2) // Control + Dest + Src
	if length < expectedMinLen {
		return nil, fmt.Errorf("invalid length: %d, minimum %d", length, expectedMinLen)
	}

	// Calculate expected total frame size
	// CRCs cover pairs of bytes: 1 for Length+Ctrl, 1 for Dest, 1 for Src, and data pairs
	dataLen := int(length) - 1 - 2 - 2 // Length - Control - Dest - Src
	numCRCGroups := 3 + (dataLen+1)/2   // 3 header pairs + data pairs
	expectedSize := 2 + 1 + int(length) + (numCRCGroups * 2)

	if len(data) < expectedSize {
		return nil, fmt.Errorf("frame too short: %d bytes, expected %d", len(data), expectedSize)
	}

	// Read control byte
	control := Control{}
	control.FromByte(data[offset])
	offset++

	// Read destination address (big-endian)
	destAddr := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	// Read source address (big-endian)
	srcAddr := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	// Read data
	var frameData []byte
	if dataLen > 0 {
		frameData = make([]byte, dataLen)
		copy(frameData, data[offset:offset+dataLen])
		offset += dataLen
	}

	// Validate CRCs
	// CRCs are stored LSB-first at the end of the frame.
	// Calculate the starting offset for CRCs.
	// Frame structure: sync(2) + length(1) + control(1) + dest(2) + src(2) + data + crcs
	// CRC count = 3 (for header) + ceil(dataLen/2) (for data pairs)
	numCRCs := 3
	if dataLen > 0 {
		numCRCs += (dataLen + 1) / 2
	}
	crcStart := len(data) - (numCRCs * 2)

	// CRC 1: Length + Control (bytes 2-3)
	crcOffset := crcStart
	if err := validateCRCForRange(data, 2, 4, crcOffset); err != nil {
		return nil, fmt.Errorf("CRC validation failed for Length+Control: %v", err)
	}
	crcOffset += 2

	// CRC 2: Destination (bytes 4-5)
	if err := validateCRCForRange(data, 4, 6, crcOffset); err != nil {
		return nil, fmt.Errorf("CRC validation failed for Destination: %v", err)
	}
	crcOffset += 2

	// CRC 3: Source (bytes 6-7)
	if err := validateCRCForRange(data, 6, 8, crcOffset); err != nil {
		return nil, fmt.Errorf("CRC validation failed for Source: %v", err)
	}
	crcOffset += 2

	// CRC for data (pairs of bytes)
	for i := 0; i < len(frameData); i += 2 {
		dataEnd := i + 2
		var pair [2]byte
		pair[0] = frameData[i]
		if dataEnd < len(frameData) {
			pair[1] = frameData[i+1]
		} else {
			// Pad with 0x00 to make a complete pair, matching Encode behavior
			pair[1] = 0x00
		}
		calculatedCRC := crc.CRC16(pair[:])
		storedCRC := uint16(data[crcOffset]) | (uint16(data[crcOffset+1]) << 8)
		if calculatedCRC != storedCRC {
			return nil, fmt.Errorf("CRC validation failed for Data at offset %d", crcOffset)
		}
		crcOffset += 2
	}

	return &Frame{
		Control:  control,
		DestAddr: destAddr,
		SrcAddr:  srcAddr,
		Data:     frameData,
	}, nil
}

// appendCRC16 calculates and appends CRC-16-DNP for two bytes.
func appendCRC16(buf []byte, b1, b2 byte) []byte {
	crcVal := crc.CRC16([]byte{b1, b2})
	// CRC bytes are LSB first
	buf = append(buf, byte(crcVal&0xFF))
	buf = append(buf, byte(crcVal>>8))
	return buf
}

// validateCRCForRange validates the CRC for a range of bytes in the frame.
func validateCRCForRange(data []byte, start, end, crcOffset int) error {
	calculatedCRC := crc.CRC16(data[start:end])
	storedCRC := uint16(data[crcOffset]) | (uint16(data[crcOffset+1]) << 8)
	if calculatedCRC != storedCRC {
		return fmt.Errorf("expected 0x%04X, got 0x%04X", storedCRC, calculatedCRC)
	}
	return nil
}

// IsBroadcast returns true if the frame destination is broadcast.
// Per IEEE 1815-2012 Section 5.3, only 0xFFFF is the broadcast address.
// 0xFFFA (All-stations reset) is a special broadcast function but
// is NOT the general broadcast address.
func (f *Frame) IsBroadcast() bool {
	return f.DestAddr == AddrBroadcast
}

// String returns a human-readable representation of the frame.
func (f *Frame) String() string {
	return fmt.Sprintf("Frame{DIR=%t, PRM=%t, Func=%d, Dest=0x%04X, Src=0x%04X, DataLen=%d}",
		f.Control.DIR, f.Control.PRM, f.Control.FuncCode, f.DestAddr, f.SrcAddr, len(f.Data))
}

// Equal compares two frames for equality.
func (f *Frame) Equal(other *Frame) bool {
	if f.Control.ToByte() != other.Control.ToByte() {
		return false
	}
	if f.DestAddr != other.DestAddr {
		return false
	}
	if f.SrcAddr != other.SrcAddr {
		return false
	}
	return bytes.Equal(f.Data, other.Data)
}
