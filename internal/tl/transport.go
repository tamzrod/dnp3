// Package tl implements DNP3 Transport Layer functions.
//
// The Transport Layer segments Application PDUs into multiple Data Link frames
// and reassembles received fragments back into complete messages.
//
// Transport Header (1 byte):
//   - Bit 7 (FIR): First fragment indicator
//   - Bit 6 (FIN): Final fragment indicator
//   - Bits 5-0: Sequence number (0-63)
//
// Reference: IEEE 1815-2012 Section 6
package tl

import (
	"errors"
	"fmt"
)

// Maximum transport fragment size (Data Link payload - transport header)
const MaxFragmentData = 292 - 1 // 291 bytes

// Sequence number limits
const (
	SeqMax     = 63              // Maximum sequence number
	SeqMod     = 64              // Sequence number modulus
	SeqMask    = 0x3F           // Sequence number mask
)

// Transport header bit masks
const (
	FIRBit = 0x80 // First fragment
	FINBit = 0x40 // Final fragment
)

// Transport errors
var (
	ErrMissingFirstFragment  = errors.New("received non-FIR fragment without prior FIR")
	ErrSequenceMismatch      = errors.New("sequence number mismatch")
	ErrDuplicateFragment     = errors.New("duplicate fragment received")
	ErrIncompleteMessage    = errors.New("incomplete message on close")
	ErrBufferOverflow       = errors.New("reassembly buffer overflow")
	ErrInvalidHeader        = errors.New("invalid transport header")
)

// Fragment represents a single transport fragment.
type Fragment struct {
	FIR bool // First fragment
	FIN bool // Final fragment
	Seq uint8 // Sequence number (0-63)
	Data []byte // Fragment data
}

// Header returns the transport header byte for this fragment.
func (f *Fragment) Header() byte {
	var h byte
	if f.FIR {
		h |= FIRBit
	}
	if f.FIN {
		h |= FINBit
	}
	h |= f.Seq & SeqMask
	return h
}

// SetHeader decodes a transport header byte into fragment flags.
func (f *Fragment) SetHeader(h byte) {
	f.FIR = (h & FIRBit) != 0
	f.FIN = (h & FINBit) != 0
	f.Seq = h & SeqMask
}

// Reassembler handles reassembly of transport fragments into complete messages.
type Reassembler struct {
	buf      []byte      // Reassembly buffer
	expectedSeq uint8   // Expected sequence number
	started  bool        // Have we received FIR?
	complete bool        // Is message complete?
}

// NewReassembler creates a new fragment reassembler.
func NewReassembler() *Reassembler {
	return &Reassembler{
		buf:        make([]byte, 0, 4096),
		expectedSeq: 0,
		started:    false,
		complete:   false,
	}
}

// Reset clears the reassembler state for a new message.
func (r *Reassembler) Reset() {
	r.buf = r.buf[:0]
	r.expectedSeq = 0
	r.started = false
	r.complete = false
}

// Push adds a fragment to the reassembly buffer.
// Returns the completed message when FIN is received, or nil if more fragments needed.
func (r *Reassembler) Push(f Fragment) ([]byte, error) {
	// Check if this is the first fragment
	if f.FIR {
		if r.started {
			// New message started before previous completed
			r.Reset()
		}
		r.expectedSeq = f.Seq
		r.started = true
	} else {
		// Not FIR - must have received FIR first
		if !r.started {
			return nil, ErrMissingFirstFragment
		}
	}

	// Verify sequence number
	if f.Seq != r.expectedSeq {
		return nil, fmt.Errorf("%w: expected %d, got %d", ErrSequenceMismatch, r.expectedSeq, f.Seq)
	}

	// Append data to buffer
	if len(r.buf)+len(f.Data) > cap(r.buf) {
		r.Reset()
		return nil, ErrBufferOverflow
	}
	r.buf = append(r.buf, f.Data...)

	// Increment expected sequence (mod 64)
	r.expectedSeq = (r.expectedSeq + 1) % SeqMod

	// Check if complete
	if f.FIN {
		r.complete = true
		result := make([]byte, len(r.buf))
		copy(result, r.buf)
		return result, nil
	}

	return nil, nil
}

// IsComplete returns true if a complete message has been reassembled.
func (r *Reassembler) IsComplete() bool {
	return r.complete
}

// BufferLen returns the current buffer length.
func (r *Reassembler) BufferLen() int {
	return len(r.buf)
}

// Fragmenter handles segmentation of application PDUs into transport fragments.
type Fragmenter struct {
	seq uint8 // Current sequence number
}

// NewFragmenter creates a new fragmenter.
func NewFragmenter() *Fragmenter {
	return &Fragmenter{
		seq: 0,
	}
}

// Reset resets the fragmenter sequence number.
func (f *Fragmenter) Reset() {
	f.seq = 0
}

// Fragmentize segments an application PDU into transport fragments.
func (f *Fragmenter) Fragmentize(data []byte) []Fragment {
	var fragments []Fragment

	// Calculate number of fragments needed
	fragCount := (len(data) + MaxFragmentData - 1) / MaxFragmentData

	if fragCount == 0 {
		// Empty message - single fragment with no data
		fragments = append(fragments, Fragment{
			FIR: true,
			FIN: true,
			Seq: f.seq,
			Data: nil,
		})
		f.seq = (f.seq + 1) % SeqMod
		return fragments
	}

	for i := 0; i < fragCount; i++ {
		isFirst := i == 0
		isLast := i == fragCount-1

		start := i * MaxFragmentData
		end := start + MaxFragmentData
		if end > len(data) {
			end = len(data)
		}

		frag := Fragment{
			FIR: isFirst,
			FIN: isLast,
			Seq: f.seq,
			Data: data[start:end],
		}
		fragments = append(fragments, frag)

		// Don't increment seq for last fragment yet
		if !isLast {
			f.seq = (f.seq + 1) % SeqMod
		}
	}

	// Increment after last fragment
	f.seq = (f.seq + 1) % SeqMod
	return fragments
}

// EncodedFragments returns fragments as encoded byte slices (header + data).
func (f *Fragmenter) EncodedFragments(data []byte) [][]byte {
	fragments := f.Fragmentize(data)
	result := make([][]byte, len(fragments))

	for i, frag := range fragments {
		encoded := make([]byte, 1+len(frag.Data))
		encoded[0] = frag.Header()
		copy(encoded[1:], frag.Data)
		result[i] = encoded
	}

	return result
}

// DecodeFragment decodes a transport fragment from a byte slice.
func DecodeFragment(data []byte) (Fragment, error) {
	if len(data) < 1 {
		return Fragment{}, ErrInvalidHeader
	}

	f := Fragment{}
	f.SetHeader(data[0])
	if len(data) > 1 {
		f.Data = make([]byte, len(data)-1)
		copy(f.Data, data[1:])
	}
	return f, nil
}

// EncodeFragment encodes a fragment into a byte slice.
func EncodeFragment(f Fragment) []byte {
	result := make([]byte, 1+len(f.Data))
	result[0] = f.Header()
	copy(result[1:], f.Data)
	return result
}

// ValidateFragment performs basic validation on a fragment.
func ValidateFragment(f Fragment) error {
	if f.Seq > SeqMax {
		return fmt.Errorf("sequence number %d exceeds maximum %d", f.Seq, SeqMax)
	}
	return nil
}

// FragmentCount calculates how many fragments a message will require.
func FragmentCount(dataLen int) int {
	if dataLen == 0 {
		return 1
	}
	return (dataLen + MaxFragmentData - 1) / MaxFragmentData
}

// IsMultiFragment returns true if the data requires fragmentation.
func IsMultiFragment(dataLen int) bool {
	return dataLen > MaxFragmentData
}
