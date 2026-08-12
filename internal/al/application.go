// Package al implements DNP3 Application Layer functions.
//
// The Application Layer defines data encoding, operations, and request/response
// structures for DNP3 protocol.
//
// Application PDU Structure:
//   - Application Control (1 byte)
//   - Function Code (1 byte)
//   - Objects (variable)
//
// Reference: IEEE 1815-2012 Section 7
package al

import (
	"fmt"
)

// Application control bit masks
const (
	AppFIRBit  = 0x80 // First fragment
	AppFINBit  = 0x40 // Final fragment
	AppCONBit  = 0x20 // Confirmation required
	AppUNSBit  = 0x10 // Unsolicited response
	AppSeqMask = 0x0F // Sequence number mask (4 bits)
)

// Sequence number limits
const (
	AppSeqMax = 15 // Maximum application sequence number
	AppSeqMod = 16 // Sequence number modulus
)

// Function code ranges
const (
	FuncCodeMinMaster   = 1   // Minimum master-to-outstation function code
	FuncCodeMaxMaster   = 64  // Maximum master-to-outstation function code
	FuncCodeMinMfr      = 64  // Minimum manufacturer-specific code
	FuncCodeMaxMfr      = 127 // Maximum manufacturer-specific code
	FuncCodeResponse    = 0   // Response function code
	FuncCodeUnsolicited = 1   // Unsolicited response
	FuncCodeNoAck       = 127 // No acknowledgment required
)

// Application control field
type AppControl struct {
	FIR bool  // First fragment
	FIN bool  // Final fragment
	CON bool  // Confirmation required
	UNS bool  // Unsolicited response
	Seq uint8 // Sequence number (0-15)
}

// Header returns the application control byte.
func (a *AppControl) Header() byte {
	var h byte
	if a.FIR {
		h |= AppFIRBit
	}
	if a.FIN {
		h |= AppFINBit
	}
	if a.CON {
		h |= AppCONBit
	}
	if a.UNS {
		h |= AppUNSBit
	}
	h |= a.Seq & AppSeqMask
	return h
}

// SetHeader decodes an application control byte.
func (a *AppControl) SetHeader(h byte) {
	a.FIR = (h & AppFIRBit) != 0
	a.FIN = (h & AppFINBit) != 0
	a.CON = (h & AppCONBit) != 0
	a.UNS = (h & AppUNSBit) != 0
	a.Seq = h & AppSeqMask
}

// Application PDU
type APDU struct {
	Control  AppControl
	FuncCode uint8
	Data     []byte
}

// NewRequest creates a new request APDU.
func NewRequest(seq uint8, funcCode uint8) *APDU {
	return &APDU{
		Control: AppControl{
			FIR: true,
			FIN: true,
			CON: false,
			UNS: false,
			Seq: seq % AppSeqMod,
		},
		FuncCode: funcCode,
		Data:     nil,
	}
}

// NewResponse creates a new response APDU.
func NewResponse(seq uint8, funcCode uint8, data []byte) *APDU {
	return &APDU{
		Control: AppControl{
			FIR: true,
			FIN: true,
			CON: false,
			UNS: funcCode == FuncCodeUnsolicited,
			Seq: seq % AppSeqMod,
		},
		FuncCode: funcCode,
		Data:     data,
	}
}

// NewUnsolicited creates an unsolicited response APDU.
func NewUnsolicited(seq uint8, data []byte) *APDU {
	return &APDU{
		Control: AppControl{
			FIR: true,
			FIN: true,
			CON: true,
			UNS: true,
			Seq: seq % AppSeqMod,
		},
		FuncCode: FuncCodeUnsolicited,
		Data:     data,
	}
}

// Encode serializes the APDU to bytes.
func (a *APDU) Encode() []byte {
	result := make([]byte, 2+len(a.Data))
	result[0] = a.Control.Header()
	result[1] = a.FuncCode
	if len(a.Data) > 0 {
		copy(result[2:], a.Data)
	}
	return result
}

// Decode parses an APDU from bytes.
func Decode(data []byte) (*APDU, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("APDU too short: %d bytes, expected at least 2", len(data))
	}

	a := &APDU{}
	a.Control.SetHeader(data[0])
	a.FuncCode = data[1]
	if len(data) > 2 {
		a.Data = make([]byte, len(data)-2)
		copy(a.Data, data[2:])
	}
	return a, nil
}

// IsRequest returns true if this is a master-to-outstation request.
func (a *APDU) IsRequest() bool {
	return a.FuncCode >= FuncCodeMinMaster
}

// IsResponse returns true if this is an outstation-to-master response.
func (a *APDU) IsResponse() bool {
	return a.FuncCode <= FuncCodeUnsolicited
}

// IsConfirmationRequired returns true if receiver must send confirm.
func (a *APDU) IsConfirmationRequired() bool {
	return a.Control.CON
}

// IsUnsolicited returns true if this is an unsolicited response.
func (a *APDU) IsUnsolicited() bool {
	return a.Control.UNS
}

// String returns a human-readable description of the APDU.
func (a *APDU) String() string {
	funcName := FunctionCodeName(a.FuncCode)
	return fmt.Sprintf("APDU{Control: 0x%02X, FuncCode: %d (%s), DataLen: %d}",
		a.Control.Header(), a.FuncCode, funcName, len(a.Data))
}

// Internal Indication (IIN) field - 2 bytes
// IIN holds the 2-byte Internal Indications field (IEEE 1815-2012,
// Section 10.5.1 "Field Type: Internal Indications"). The field is two
// octets transmitted IIN1 (first) then IIN2. Bits within each octet are
// transmitted MSB-first, so bit index 0 (IIN1.0) corresponds to the MSB
// (0x80) of octet 1, and bit index 7 (IIN1.7) corresponds to the LSB
// (0x01) of octet 1; likewise for IIN2.
//
// IIN1 (octet 1) — device status / event availability:
//
//	0x80 IIN1.0 AllStations   — broadcast/all-stations message received
//	0x40 IIN1.1 Class1Events  — Class 1 event data available
//	0x20 IIN1.2 Class2Events  — Class 2 event data available
//	0x10 IIN1.3 Class3Events  — Class 3 event data available
//	0x08 IIN1.4 NeedTime      — time synchronization required
//	0x04 IIN1.5 LocalControl  — local mode (points uncontrollable via DNP)
//	0x02 IIN1.6 DeviceTrouble — device trouble
//	0x01 IIN1.7 DeviceRestart — device restart
//
// IIN2 (octet 2) — application-level errors:
//
//	0x80 IIN2.0 FuncUnknown       — function code cannot be processed
//	0x40 IIN2.1 ObjectUnknown     — object group/variation cannot be processed
//	0x20 IIN2.2 ParameterError    — qualifier/range field is in error
//	0x10 IIN2.3 BufferOverflow    — event buffer overflowed, events lost
//	0x08 IIN2.4 AlreadyExecuting  — operation already executing
//	0x04 IIN2.5 BadConfig         — bad configuration
//	0x02 IIN2.6 Reserved2_6       — reserved, always 0
//	0x01 IIN2.7 Reserved2_7       — reserved, always 0
//
// MEXT-019 FREEZE: this bit map is frozen against IEEE 1815-2012 for the
// external v0 interop claim. The flag-to-octet/position mapping MUST NOT
// change without an explicit spec-continuity review. The frozen table is
// locked by internal/al/iin_freeze_test.go (named critical masks + a full
// 16-bit round-trip).
type IIN struct {
	// IIN1 (octet 1) flags
	AllStations   bool // IIN1.0 0x80 — broadcast/all-stations message received
	Class1Events  bool // IIN1.1 0x40 — Class 1 event data available
	Class2Events  bool // IIN1.2 0x20 — Class 2 event data available
	Class3Events  bool // IIN1.3 0x10 — Class 3 event data available
	NeedTime      bool // IIN1.4 0x08 — time synchronization required
	LocalControl  bool // IIN1.5 0x04 — local mode
	DeviceTrouble bool // IIN1.6 0x02 — device trouble
	DeviceRestart bool // IIN1.7 0x01 — device restart

	// IIN2 (octet 2) flags
	FuncUnknown      bool // IIN2.0 0x80 — function code unknown
	ObjectUnknown    bool // IIN2.1 0x40 — object unknown
	ParameterError   bool // IIN2.2 0x20 — parameter/qualifier/range error
	BufferOverflow   bool // IIN2.3 0x10 — event buffer overflow
	AlreadyExecuting bool // IIN2.4 0x08 — already executing
	BadConfig        bool // IIN2.5 0x04 — bad configuration
	Reserved2_6      bool // IIN2.6 0x02 — reserved (always 0)
	Reserved2_7      bool // IIN2.7 0x01 — reserved (always 0)
}

// Bytes returns the 2-byte IIN representation (IIN1, IIN2).
func (i *IIN) Bytes() [2]byte {
	var result [2]byte
	if i.AllStations {
		result[0] |= 0x80
	}
	if i.Class1Events {
		result[0] |= 0x40
	}
	if i.Class2Events {
		result[0] |= 0x20
	}
	if i.Class3Events {
		result[0] |= 0x10
	}
	if i.NeedTime {
		result[0] |= 0x08
	}
	if i.LocalControl {
		result[0] |= 0x04
	}
	if i.DeviceTrouble {
		result[0] |= 0x02
	}
	if i.DeviceRestart {
		result[0] |= 0x01
	}

	if i.FuncUnknown {
		result[1] |= 0x80
	}
	if i.ObjectUnknown {
		result[1] |= 0x40
	}
	if i.ParameterError {
		result[1] |= 0x20
	}
	if i.BufferOverflow {
		result[1] |= 0x10
	}
	if i.AlreadyExecuting {
		result[1] |= 0x08
	}
	if i.BadConfig {
		result[1] |= 0x04
	}
	if i.Reserved2_6 {
		result[1] |= 0x02
	}
	if i.Reserved2_7 {
		result[1] |= 0x01
	}

	return result
}

// SetIIN decodes a 2-byte IIN (IIN1, IIN2).
func (i *IIN) SetIIN(b [2]byte) {
	i.AllStations = (b[0] & 0x80) != 0
	i.Class1Events = (b[0] & 0x40) != 0
	i.Class2Events = (b[0] & 0x20) != 0
	i.Class3Events = (b[0] & 0x10) != 0
	i.NeedTime = (b[0] & 0x08) != 0
	i.LocalControl = (b[0] & 0x04) != 0
	i.DeviceTrouble = (b[0] & 0x02) != 0
	i.DeviceRestart = (b[0] & 0x01) != 0

	i.FuncUnknown = (b[1] & 0x80) != 0
	i.ObjectUnknown = (b[1] & 0x40) != 0
	i.ParameterError = (b[1] & 0x20) != 0
	i.BufferOverflow = (b[1] & 0x10) != 0
	i.AlreadyExecuting = (b[1] & 0x08) != 0
	i.BadConfig = (b[1] & 0x04) != 0
	i.Reserved2_6 = (b[1] & 0x02) != 0
	i.Reserved2_7 = (b[1] & 0x01) != 0
}

// EncodeIIN encodes the IIN as 2 bytes.
func EncodeIIN(i *IIN) []byte {
	bytes := i.Bytes()
	return bytes[:]
}

// DecodeIIN decodes a 2-byte IIN.
func DecodeIIN(data []byte) (*IIN, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("IIN too short: %d bytes, expected 2", len(data))
	}
	i := &IIN{}
	var b [2]byte
	copy(b[:], data[:2])
	i.SetIIN(b)
	return i, nil
}

// Response APDU with IIN
type Response struct {
	Header APDU
	IIN    IIN
	Data   []byte
}

// NewResponse creates a new response with IIN.
func NewAppResponse(seq uint8, iin IIN, data []byte) *Response {
	return &Response{
		Header: APDU{
			Control: AppControl{
				FIR: true,
				FIN: true,
				CON: false,
				UNS: false,
				Seq: seq % AppSeqMod,
			},
			FuncCode: FuncCodeResponse,
		},
		IIN:  iin,
		Data: data,
	}
}

// Encode serializes the response to bytes.
func (r *Response) Encode() []byte {
	iinBytes := r.IIN.Bytes()
	result := make([]byte, 2+2+len(r.Data))
	result[0] = r.Header.Control.Header()
	result[1] = r.Header.FuncCode
	result[2] = iinBytes[0]
	result[3] = iinBytes[1]
	if len(r.Data) > 0 {
		copy(result[4:], r.Data)
	}
	return result
}

// Decode parses a response from bytes.
func DecodeResponse(data []byte) (*Response, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("Response too short: %d bytes, expected at least 4", len(data))
	}

	r := &Response{}
	r.Header.Control.SetHeader(data[0])
	r.Header.FuncCode = data[1]
	r.Header.Data = data[4:]

	var iinBytes [2]byte
	iinBytes[0] = data[2]
	iinBytes[1] = data[3]
	r.IIN.SetIIN(iinBytes)

	if len(data) > 4 {
		r.Data = make([]byte, len(data)-4)
		copy(r.Data, data[4:])
	}

	return r, nil
}

// MinAPDUHeaderLen is the minimum APDU header length.
const MinAPDUHeaderLen = 2

// MinResponseLen is the minimum response length (header + IIN).
const MinResponseLen = 4

// NewConfirm creates a new confirmation APDU.
// In DNP3, a confirmation is a response with FuncCode=0 and no IIN.
func NewConfirm(seq uint8) *APDU {
	return &APDU{
		Control: AppControl{
			FIR: true,
			FIN: true,
			CON: false, // Confirm does not require further confirmation
			UNS: false,
			Seq: seq,
		},
		FuncCode: FuncResponse, // FuncCode 0 is used for both response and confirm
		Data:     nil,          // No IIN for confirm
	}
}

// NewConfirmWithIIN creates a confirmation with IIN (e.g., for solicited confirmation).
func NewConfirmWithIIN(seq uint8, iin *IIN) *APDU {
	data := make([]byte, 2)
	if iin != nil {
		bytes := iin.Bytes()
		data[0] = bytes[0]
		data[1] = bytes[1]
	}
	return &APDU{
		Control: AppControl{
			FIR: true,
			FIN: true,
			CON: false,
			UNS: false,
			Seq: seq,
		},
		FuncCode: FuncResponse,
		Data:     data,
	}
}
