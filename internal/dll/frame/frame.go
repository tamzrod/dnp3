// Package frame implements DNP3 Data Link Layer frame encoding and decoding.
//
// A DNP3 Data Link Layer frame consists of:
//   - Start bytes (0x05 0x64)
//   - Length (1 byte)
//   - Control byte (1 byte)
//   - Destination address (2 bytes, little-endian)
//   - Source address (2 bytes, little-endian)
//   - User data (0-292 bytes)
//   - CRC bytes (one header CRC and one CRC per 16-byte data block)
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
	// MaxFrameSize is the maximum total frame size (250 data + 10 header + 32 CRC = 292).
	// Per IEEE 1815-2012, the length field is 1 byte (max 255)
	// Length = Control(1) + Dest(2) + Src(2) + Data
	// Max Data = 255 - 1 - 2 - 2 = 250 bytes
	MaxFrameSize = 292

	// MaxDataSize is the maximum user data size (250 bytes)
	// This fits within the 1-byte length field: 1+2+2+250 = 255
	MaxDataSize = 250

	// HeaderSize is the size without data but with CRC
	HeaderSize = 10

	// MinFrameSize is the minimum valid frame size
	MinFrameSize = 10
)

// Function codes for primary station (PRM=1)
const (
	FuncResetLinkStations   = 0
	FuncTestLinkStates      = 2
	FuncConfirmedUserData   = 3
	FuncUnconfirmedUserData = 4
	FuncRequestLinkStatus   = 9

	// Deprecated aliases retained while callers are migrated.
	FuncResetLinkStatus  = FuncRequestLinkStatus
	FuncReturnLinkStatus = FuncRequestLinkStatus
)

// Function codes for secondary station (PRM=0)
const (
	FuncAck                = 0
	FuncNack               = 1
	FuncLinkStatus         = 2
	FuncNotSupported       = 3
	FuncConfirmedUserDataR = 4
)

// Special addresses
const (
	AddrBroadcast    = 0xFFFF
	AddrAllReset     = 0xFFFA
	AddrUnconfigured = 0x0000
	AddrVirtualTerm  = 0xFFFB
	AddrSecChannel   = 0xFFFC
	AddrPriChannel   = 0xFFFD
	AddrReserved     = 0xFFFE
)

// Frame represents a DNP3 Data Link Layer frame.
type Frame struct {
	// Control contains the control byte fields
	Control Control

	// Destination address (little-endian)
	DestAddr uint16

	// Source address (little-endian)
	SrcAddr uint16

	// User data (transport-layer data)
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
//
//	Bit 7: DIR (Direction)
//	Bit 6: PRM (Primary)
//	Bit 5: FCB (Frame Count Bit)
//	Bit 4: FCV (Frame Count Bit Valid)
//	Bits 3-0: Function Code (4 bits, primary functions 0-15)
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
//   - Destination (2 bytes, little-endian)
//   - Source (2 bytes, little-endian)
//   - Data (0-250 bytes)
//   - CRC (one for the header and one per 16-byte data block, LSB first)
func Encode(f *Frame) ([]byte, error) {
	// Validate data size
	if len(f.Data) > MaxDataSize {
		return nil, fmt.Errorf("data size %d exceeds maximum %d", len(f.Data), MaxDataSize)
	}

	buf := make([]byte, 0, EncodedSize(len(f.Data)))

	// Write sync bytes
	buf = append(buf, SyncByte1, SyncByte2)

	// Calculate length: bytes from Control to end of Data
	length := byte(1 + 2 + 2 + len(f.Data)) // Control + Dest + Src + Data
	buf = append(buf, length)

	// Write control byte
	buf = append(buf, f.Control.ToByte())

	// DNP3 multi-octet link addresses are transmitted LSB first.
	buf = binary.LittleEndian.AppendUint16(buf, f.DestAddr)

	// DNP3 multi-octet link addresses are transmitted LSB first.
	buf = binary.LittleEndian.AppendUint16(buf, f.SrcAddr)

	// The link header CRC covers the complete header prefix, including sync,
	// length, control, destination, and source.
	buf = appendCRC16(buf, buf[0:8])

	// Data CRCs are interleaved after each 16-octet data block.
	for i := 0; i < len(f.Data); i += 16 {
		end := i + 16
		if end > len(f.Data) {
			end = len(f.Data)
		}
		buf = append(buf, f.Data[i:end]...)
		buf = appendCRC16(buf, f.Data[i:end])
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

	// Calculate expected total frame size.
	dataLen := int(length) - 1 - 2 - 2 // Length - Control - Dest - Src

	// Reject oversize frames (DNP3-027): the data link layer caps user data at
	// MaxDataSize (250). The 1-byte length field makes larger claims impossible,
	// but guard explicitly so an oversize claim is a clear error rather than
	// silently accepted.
	if dataLen > MaxDataSize {
		return nil, fmt.Errorf("frame too large: data length %d exceeds maximum %d", dataLen, MaxDataSize)
	}

	expectedSize := EncodedSize(dataLen)

	if len(data) < expectedSize {
		return nil, fmt.Errorf("frame too short: %d bytes, expected %d", len(data), expectedSize)
	}

	// Read control byte
	control := Control{}
	control.FromByte(data[offset])
	offset++

	// Read destination address (LSB first).
	destAddr := binary.LittleEndian.Uint16(data[offset : offset+2])
	offset += 2

	// Read source address (LSB first).
	srcAddr := binary.LittleEndian.Uint16(data[offset : offset+2])
	offset += 2

	if err := validateCRCForRange(data, 0, 8, offset); err != nil {
		return nil, fmt.Errorf("header CRC validation failed: %w", err)
	}
	offset += 2

	frameData := make([]byte, 0, dataLen)
	for len(frameData) < dataLen {
		blockLen := 16
		if remaining := dataLen - len(frameData); remaining < blockLen {
			blockLen = remaining
		}
		block := data[offset : offset+blockLen]
		if err := validateCRCForRange(data, offset, offset+blockLen, offset+blockLen); err != nil {
			return nil, fmt.Errorf("data CRC validation failed at payload offset %d: %w", len(frameData), err)
		}
		frameData = append(frameData, block...)
		offset += blockLen + 2
	}

	return &Frame{
		Control:  control,
		DestAddr: destAddr,
		SrcAddr:  srcAddr,
		Data:     frameData,
	}, nil
}

// EncodedSize returns the number of octets in a complete DNP3 link frame.
func EncodedSize(dataLen int) int {
	return HeaderSize + dataLen + ((dataLen+15)/16)*2
}

// appendCRC16 calculates and appends CRC-16-DNP for one link-layer block.
func appendCRC16(buf []byte, block []byte) []byte {
	crcVal := crc.CRC16(block)
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
