package master

import (
	"testing"

	"dnp3/internal/dll/frame"
	"dnp3/internal/tl"
)

// TestEncodedSizeMatchesEncodedLength is the DNP3-065 audit guard: for every
// payload length the receive loops encounter, frame.EncodedSize (used to advance
// the offset) must equal the actual number of wire octets frame.Encode emits.
// Any mismatch would mean the receive loop over-reads (skips into the next
// frame) or under-reads (re-parses trailing bytes) on concatenated frames.
func TestEncodedSizeMatchesEncodedLength(t *testing.T) {
	for _, dataLen := range []int{0, 1, 2, 15, 16, 17, 31, 32, 33, 100, 249} {
		payload := make([]byte, dataLen)
		for i := range payload {
			payload[i] = byte(i)
		}
		dll := &frame.Frame{
			Control:  frame.Control{DIR: false, PRM: false, FuncCode: frame.FuncConfirmedUserDataR},
			DestAddr: 1, SrcAddr: 2, Data: payload,
		}
		encoded, err := frame.Encode(dll)
		if err != nil {
			t.Fatalf("dataLen=%d: encode: %v", dataLen, err)
		}
		got := frame.EncodedSize(dataLen)
		if got != len(encoded) {
			t.Fatalf("dataLen=%d: EncodedSize=%d but len(encoded)=%d (over/under-read risk)", dataLen, got, len(encoded))
		}
	}
}

// buildDLLFrameWithTL wraps a TL fragment in a secondary confirmed-user-data
// DLL frame (the form the master receives from an outstation).
func buildDLLFrameWithTL(t *testing.T, frag tl.Fragment) []byte {
	t.Helper()
	dll := &frame.Frame{
		Control:  frame.Control{DIR: false, PRM: false, FuncCode: frame.FuncConfirmedUserDataR},
		DestAddr: 1, SrcAddr: 2, Data: tl.EncodeFragment(frag),
	}
	raw, err := frame.Encode(dll)
	if err != nil {
		t.Fatalf("encode DLL frame: %v", err)
	}
	return raw
}

// TestConcatenatedFramesConsumedExactly verifies DNP3-065: when a single receive
// buffer contains two concatenated DLL frames (a multi-fragment response:
// FIR+non-FIN then non-FIR+FIN), processReceivedBytes advances the offset by
// exactly each frame's EncodedSize and reassembles the full APDU with no
// over-read or under-read.
func TestConcatenatedFramesConsumedExactly(t *testing.T) {
	m := NewMaster(DefaultConfig())

	// Two TL fragments forming one application message: "AB" + "CD" → "ABCD".
	frag1 := tl.Fragment{FIR: true, FIN: false, Seq: 0, Data: []byte("AB")}
	frag2 := tl.Fragment{FIR: false, FIN: true, Seq: 1, Data: []byte("CD")}

	buf := append(buildDLLFrameWithTL(t, frag1), buildDLLFrameWithTL(t, frag2)...)

	// Sanity: the buffer is exactly two encoded frames (no trailing bytes),
	// so a correct offset advancement consumes it whole with no leftover.
	wantConsumed := len(buf)

	got, err := m.processReceivedBytes(buf)
	if err != nil {
		t.Fatalf("processReceivedBytes: %v", err)
	}
	if string(got) != "ABCD" {
		t.Fatalf("reassembled data = %q, want %q", string(got), "ABCD")
	}

	// Re-derive each frame's wire size via EncodedSize and confirm they sum to
	// the whole buffer — the precise no-over/under-read invariant.
	size1 := frame.EncodedSize(len(tl.EncodeFragment(frag1)))
	size2 := frame.EncodedSize(len(tl.EncodeFragment(frag2)))
	if size1+size2 != wantConsumed {
		t.Fatalf("EncodedSize sum=%d, buffer len=%d (over/under-read)", size1+size2, wantConsumed)
	}
}

// TestConcatenatedThreeFramesConsumedExactly verifies DNP3-065 with three
// concatenated DLL frames (FIR, mid, FIN) — exercises the loop iterating more
// than twice and confirms full consumption + correct reassembly order.
func TestConcatenatedThreeFramesConsumedExactly(t *testing.T) {
	m := NewMaster(DefaultConfig())

	frag1 := tl.Fragment{FIR: true, FIN: false, Seq: 0, Data: []byte("AA")}
	frag2 := tl.Fragment{FIR: false, FIN: false, Seq: 1, Data: []byte("BB")}
	frag3 := tl.Fragment{FIR: false, FIN: true, Seq: 2, Data: []byte("CC")}

	buf := buildDLLFrameWithTL(t, frag1)
	buf = append(buf, buildDLLFrameWithTL(t, frag2)...)
	buf = append(buf, buildDLLFrameWithTL(t, frag3)...)

	got, err := m.processReceivedBytes(buf)
	if err != nil {
		t.Fatalf("processReceivedBytes: %v", err)
	}
	if string(got) != "AABBCC" {
		t.Fatalf("reassembled data = %q, want %q", string(got), "AABBCC")
	}

	// EncodedSize of each frame must sum to the buffer length — no over/under-read.
	s1 := frame.EncodedSize(len(tl.EncodeFragment(frag1)))
	s2 := frame.EncodedSize(len(tl.EncodeFragment(frag2)))
	s3 := frame.EncodedSize(len(tl.EncodeFragment(frag3)))
	if s1+s2+s3 != len(buf) {
		t.Fatalf("EncodedSize sum=%d, buffer len=%d (over/under-read)", s1+s2+s3, len(buf))
	}
}
