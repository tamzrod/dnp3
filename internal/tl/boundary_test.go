package tl

import (
	"bytes"
	"testing"
)

// TestFragmentSizeBoundaries locks the DNP3-059 transport-layer fragment-size
// boundaries (exact 0 / 249 / 250 bytes, plus the adjacent 248 and 498/499
// edges) so there is no off-by-one in fragmentation or reassembly. Each case
// asserts the fragment count, every fragment's FIR/FIN flags and data length,
// the per-fragment sequence numbering, and a full fragmentize → reassemble
// round-trip reproducing the original payload byte-for-byte.
func TestFragmentSizeBoundaries(t *testing.T) {
	cases := []struct {
		name    string
		dataLen int
		// expected[i] = (fragmentDataLen, isFIR, isFIN)
		expected []struct {
			len int
			fir bool
			fin bool
		}
	}{
		{
			name:    "empty 0 bytes",
			dataLen: 0,
			expected: []struct {
				len int
				fir bool
				fin bool
			}{{0, true, true}},
		},
		{
			name:    "just under max 248 bytes",
			dataLen: 248,
			expected: []struct {
				len int
				fir bool
				fin bool
			}{{248, true, true}},
		},
		{
			name:    "exact max single 249 bytes",
			dataLen: 249,
			expected: []struct {
				len int
				fir bool
				fin bool
			}{{249, true, true}},
		},
		{
			name:    "one over max 250 bytes",
			dataLen: 250,
			expected: []struct {
				len int
				fir bool
				fin bool
			}{{249, true, false}, {1, false, true}},
		},
		{
			name:    "exact two-fragment 498 bytes",
			dataLen: 498,
			expected: []struct {
				len int
				fir bool
				fin bool
			}{{249, true, false}, {249, false, true}},
		},
		{
			name:    "two-fragment plus one 499 bytes",
			dataLen: 499,
			expected: []struct {
				len int
				fir bool
				fin bool
			}{{249, true, false}, {249, false, false}, {1, false, true}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := make([]byte, tc.dataLen)
			for i := range data {
				data[i] = byte(i % 251) // distinguishable fill (251 is prime)
			}

			f := NewFragmenter()
			frags := f.Fragmentize(data)
			if len(frags) != len(tc.expected) {
				t.Fatalf("fragment count = %d, want %d", len(frags), len(tc.expected))
			}

			// Validate each fragment's flags, length, and sequence number.
			for i, want := range tc.expected {
				got := frags[i]
				if got.FIR != want.fir {
					t.Errorf("frag[%d].FIR = %v, want %v", i, got.FIR, want.fir)
				}
				if got.FIN != want.fin {
					t.Errorf("frag[%d].FIN = %v, want %v", i, got.FIN, want.fin)
				}
				if len(got.Data) != want.len {
					t.Errorf("frag[%d] data len = %d, want %d", i, len(got.Data), want.len)
				}
				if got.Seq != byte(i%SeqMod) {
					t.Errorf("frag[%d].Seq = %d, want %d", i, got.Seq, i%SeqMod)
				}
			}

			// Full round-trip: encode → decode → reassemble must reproduce data.
			r := NewReassembler()
			var reassembled []byte
			for _, frag := range frags {
				encoded := EncodeFragment(frag)
				decoded, err := DecodeFragment(encoded)
				if err != nil {
					t.Fatalf("DecodeFragment(%d): %v", len(encoded), err)
				}
				out, err := r.Push(decoded)
				if err != nil {
					t.Fatalf("Reassembler.Push(%d): %v", len(decoded.Data), err)
				}
				if out != nil {
					reassembled = out
				}
			}
			if !r.IsComplete() {
				t.Fatalf("reassembler not complete after all fragments")
			}
			if !bytes.Equal(reassembled, data) {
				t.Fatalf("reassembled payload mismatch: got %d bytes, want %d", len(reassembled), len(data))
			}
		})
	}
}

// TestFragmentSizeBoundaryEncodedRoundTrip exercises the wire path
// (EncodedFragments → DecodeFragment → Reassembler) for the exact 249/250/0
// boundary sizes, asserting the reassembled length matches and (for non-empty)
// the content is preserved. This complements TestFragmentSizeBoundaries by
// using the batch EncodedFragments entry point the master actually uses.
func TestFragmentSizeBoundaryEncodedRoundTrip(t *testing.T) {
	for _, dataLen := range []int{0, 249, 250} {
		data := make([]byte, dataLen)
		for i := range data {
			data[i] = byte(0xA0 ^ byte(i))
		}
		f := NewFragmenter()
		encoded := f.EncodedFragments(data)

		r := NewReassembler()
		var got []byte
		for _, e := range encoded {
			frag, err := DecodeFragment(e)
			if err != nil {
				t.Fatalf("dataLen=%d: DecodeFragment: %v", dataLen, err)
			}
			out, err := r.Push(frag)
			if err != nil {
				t.Fatalf("dataLen=%d: Push: %v", dataLen, err)
			}
			if out != nil {
				got = out
			}
		}
		if len(got) != dataLen {
			t.Errorf("dataLen=%d: reassembled len = %d, want %d", dataLen, len(got), dataLen)
		}
		if dataLen > 0 && !bytes.Equal(got, data) {
			t.Errorf("dataLen=%d: reassembled content mismatch", dataLen)
		}
	}
}
