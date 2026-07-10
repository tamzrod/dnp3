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
	AppFIRBit = 0x80 // First fragment
	AppFINBit = 0x40 // Final fragment
	AppCONBit = 0x20 // Confirmation required
	AppUNSBit = 0x10 // Unsolicited response
	AppSeqMask = 0x0F // Sequence number mask (4 bits)
)

// Sequence number limits
const (
	AppSeqMax = 15            // Maximum application sequence number
	AppSeqMod = 16            // Sequence number modulus
)

// Function code ranges
const (
	FuncCodeMinMaster  = 1  // Minimum master-to-outstation function code
	FuncCodeMaxMaster  = 64 // Maximum master-to-outstation function code
	FuncCodeMinMfr     = 64 // Minimum manufacturer-specific code
	FuncCodeMaxMfr     = 127 // Maximum manufacturer-specific code
	FuncCodeResponse   = 0  // Response function code
	FuncCodeUnsolicited = 1 // Unsolicited response
	FuncCodeNoAck      = 127 // No acknowledgment required
)

// Application control field
type AppControl struct {
	FIR bool // First fragment
	FIN bool // Final fragment
	CON bool // Confirmation required
	UNS bool // Unsolicited response
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
	Control AppControl
	FuncCode uint8
	Data    []byte
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
type IIN struct {
	// IIN.1 flags (first byte)
	AllStop       bool // Device stopped
	ByteOver      bool // Buffer overflow
	Limit64K      bool // At 64K data limit
	Limit16K      bool // At 16K data limit
	MemUnavail    bool // Requested memory unavailable
	CheckFail     bool // Checksum/test failure
	Busy          bool // Device is busy
	ParamUnavail  bool // Parameter/block unavailable
	
	// IIN.2 flags (second byte)
	TranAborted   bool // File transfer aborted
	AOBInProgress  bool // Analog output block transfer in progress
	DataLogAvail   bool // Data log available
	ConfigError    bool // Configuration error
	MemUnavailable bool // Internal memory unavailable
	NeedsTimeSync  bool // Clock needs synchronization
	GeneralEnableOff bool // General enable off
	IINMissing     bool // Internal indicator block missing
}

// IIN returns the 2-byte IIN representation.
func (i *IIN) Bytes() [2]byte {
	var result [2]byte
	if i.AllStop       { result[0] |= 0x80 }
	if i.ByteOver      { result[0] |= 0x40 }
	if i.Limit64K      { result[0] |= 0x20 }
	if i.Limit16K      { result[0] |= 0x10 }
	if i.MemUnavail    { result[0] |= 0x08 }
	if i.CheckFail     { result[0] |= 0x04 }
	if i.Busy          { result[0] |= 0x02 }
	if i.ParamUnavail  { result[0] |= 0x01 }
	
	if i.TranAborted   { result[1] |= 0x80 }
	if i.AOBInProgress { result[1] |= 0x40 }
	if i.DataLogAvail  { result[1] |= 0x20 }
	if i.ConfigError   { result[1] |= 0x10 }
	if i.MemUnavailable { result[1] |= 0x08 }
	if i.NeedsTimeSync { result[1] |= 0x04 }
	if i.GeneralEnableOff { result[1] |= 0x02 }
	if i.IINMissing    { result[1] |= 0x01 }
	
	return result
}

// SetIIN decodes a 2-byte IIN.
func (i *IIN) SetIIN(b [2]byte) {
	i.AllStop = (b[0] & 0x80) != 0
	i.ByteOver = (b[0] & 0x40) != 0
	i.Limit64K = (b[0] & 0x20) != 0
	i.Limit16K = (b[0] & 0x10) != 0
	i.MemUnavail = (b[0] & 0x08) != 0
	i.CheckFail = (b[0] & 0x04) != 0
	i.Busy = (b[0] & 0x02) != 0
	i.ParamUnavail = (b[0] & 0x01) != 0
	
	i.TranAborted = (b[1] & 0x80) != 0
	i.AOBInProgress = (b[1] & 0x40) != 0
	i.DataLogAvail = (b[1] & 0x20) != 0
	i.ConfigError = (b[1] & 0x10) != 0
	i.MemUnavailable = (b[1] & 0x08) != 0
	i.NeedsTimeSync = (b[1] & 0x04) != 0
	i.GeneralEnableOff = (b[1] & 0x02) != 0
	i.IINMissing = (b[1] & 0x01) != 0
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
