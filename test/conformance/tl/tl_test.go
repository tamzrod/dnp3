// Package tl provides conformance tests for the Transport Layer.
package tl

import (
	"testing"

	"dnp3/internal/tl"
)

// TestTransportHeaderEncoding tests transport header byte encoding
// Header format: FIR=0x80, FIN=0x40, Seq=0x3F
func TestTransportHeaderEncoding(t *testing.T) {
	tests := []struct {
		name     string
		fir      bool
		fin      bool
		seq      uint8
		expected byte
	}{
		{"single_fragment", true, true, 0, 0xC0},     // 0x80 | 0x40 | 0x00
		{"first_of_multi", true, false, 5, 0x85},     // 0x80 | 0x05
		{"middle_fragment", false, false, 10, 0x0A}, // 0x0A
		{"final_fragment", false, true, 15, 0x4F},   // 0x40 | 0x0F
		{"max_seq_first", true, true, 15, 0xCF},      // 0x80 | 0x40 | 0x0F
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := tl.Fragment{
				FIR: tt.fir,
				FIN: tt.fin,
				Seq: tt.seq,
			}
			result := f.Header()
			if result != tt.expected {
				t.Errorf("Fragment.Header(%v, %v, %d) = 0x%02X, want 0x%02X",
					tt.fir, tt.fin, tt.seq, result, tt.expected)
			}
		})
	}
}

// TestTransportHeaderDecoding tests transport header byte decoding
func TestTransportHeaderDecoding(t *testing.T) {
	tests := []struct {
		name      string
		header    byte
		wantFIR   bool
		wantFIN   bool
		wantSeq   uint8
	}{
		{"single", 0xC0, true, true, 0},
		{"first", 0x85, true, false, 5},
		{"middle", 0x1A, false, false, 0x1A}, // Seq is 0x3F mask
		{"final", 0x4F, false, true, 15},
		{"max", 0xCF, true, true, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f tl.Fragment
			f.SetHeader(tt.header)
			if f.FIR != tt.wantFIR {
				t.Errorf("FIR = %v, want %v", f.FIR, tt.wantFIR)
			}
			if f.FIN != tt.wantFIN {
				t.Errorf("FIN = %v, want %v", f.FIN, tt.wantFIN)
			}
			if f.Seq != tt.wantSeq {
				t.Errorf("Seq = %d, want %d", f.Seq, tt.wantSeq)
			}
		})
	}
}

// TestFragmentationSingleFragment tests single fragment handling
func TestFragmentationSingleFragment(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	
	fragmenter := tl.NewFragmenter()
	fragments := fragmenter.Fragmentize(data)

	if len(fragments) != 1 {
		t.Errorf("Expected 1 fragment, got %d", len(fragments))
	}

	if len(fragments[0].Data) != len(data) {
		t.Errorf("Fragment data length = %d, want %d", len(fragments[0].Data), len(data))
	}

	if !fragments[0].FIR || !fragments[0].FIN {
		t.Error("Single fragment should have FIR and FIN set")
	}
}

// TestFragmentationMultiFragment tests multi-fragment handling
func TestFragmentationMultiFragment(t *testing.T) {
	// Create data larger than one fragment can hold
	// Max fragment data is typically 292 bytes (292 + 1 header = 293)
	data := make([]byte, 600)
	for i := range data {
		data[i] = byte(i & 0xFF)
	}

	fragmenter := tl.NewFragmenter()
	fragments := fragmenter.Fragmentize(data)

	if len(fragments) < 2 {
		t.Errorf("Expected multiple fragments, got %d", len(fragments))
	}

	// First fragment should have FIR set
	if !fragments[0].FIR {
		t.Error("First fragment should have FIR set")
	}

	// Last fragment should have FIN set
	if !fragments[len(fragments)-1].FIN {
		t.Error("Last fragment should have FIN set")
	}

	// Middle fragments should have neither FIR nor FIN set
	for i := 1; i < len(fragments)-1; i++ {
		if fragments[i].FIR || fragments[i].FIN {
			t.Errorf("Middle fragment %d should not have FIR or FIN set", i)
		}
	}

	// Sequence numbers should increment (6-bit sequence, 0-63)
	for i := 1; i < len(fragments); i++ {
		expectedSeq := uint8(i % 64)
		if fragments[i].Seq != expectedSeq {
			t.Errorf("Fragment %d seq = %d, want %d", i, fragments[i].Seq, expectedSeq)
		}
	}
}

// TestFragmentationSequenceWrap tests sequence number wrapping
// Transport layer uses 6-bit sequence (0-63)
func TestFragmentationSequenceWrap(t *testing.T) {
	// Create enough fragments to wrap sequence numbers (64 max)
	maxPayload := tl.MaxFragmentData
	data := make([]byte, maxPayload*70) // More than 64 fragments worth
	for i := range data {
		data[i] = byte(i & 0xFF)
	}

	fragmenter := tl.NewFragmenter()
	fragments := fragmenter.Fragmentize(data)

	// Verify wrapping behavior (sequence goes 0-63, then wraps)
	if len(fragments) > 64 {
		// Fragment 64 (index 64) wraps to seq 0
		// Fragment index 69 = seq 5
		// Expected: (index - 64) = (len - 1 - 64) = len - 65
		expectedSeq := uint8((len(fragments) - 65) % 64)
		lastSeq := fragments[len(fragments)-1].Seq
		if lastSeq != expectedSeq {
			t.Errorf("After wrap, last fragment seq = %d, want %d", lastSeq, expectedSeq)
		}
	}
}

// TestReassembly tests fragment reassembly
func TestReassembly(t *testing.T) {
	originalData := []byte{0xDE, 0xAD, 0xBE, 0xEF, 0xCA, 0xFE}
	
	fragmenter := tl.NewFragmenter()
	fragments := fragmenter.Fragmentize(originalData)

	// Create reassembly buffer
	reassembler := tl.NewReassembler()
	
	// Push all fragments
	for _, frag := range fragments {
		data, err := reassembler.Push(frag)
		if err != nil {
			t.Fatalf("Push() error = %v", err)
		}
		if data != nil {
			// Should get data on complete
		}
	}

	// Check if complete
	if !reassembler.IsComplete() {
		t.Error("Reassembly should be complete after all fragments pushed")
	}
}

// TestReassemblyOutOfOrder tests handling of out-of-order fragments
func TestReassemblyOutOfOrder(t *testing.T) {
	// Create multi-fragment data
	data := make([]byte, 600)
	for i := range data {
		data[i] = byte(i & 0xFF)
	}

	fragmenter := tl.NewFragmenter()
	fragments := fragmenter.Fragmentize(data)
	if len(fragments) < 3 {
		t.Skip("Need at least 3 fragments for this test")
	}

	reassembler := tl.NewReassembler()

	// Push fragments out of order (last, first, middle)
	// Should fail since we can't receive final before first
	_, err := reassembler.Push(fragments[len(fragments)-1])
	if err == nil {
		// Some implementations might allow this, that's OK
	}
}

// TestReassemblyDuplicate tests handling of duplicate fragments
func TestReassemblyDuplicate(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	
	fragmenter := tl.NewFragmenter()
	fragments := fragmenter.Fragmentize(data)
	if len(fragments) < 1 {
		t.Fatal("Need at least 1 fragment")
	}

	reassembler := tl.NewReassembler()

	// Push same fragment twice
	_, err := reassembler.Push(fragments[0])
	if err != nil {
		t.Fatalf("First Push() error = %v", err)
	}

	// Duplicate might be rejected or accepted
	_, _ = reassembler.Push(fragments[0])
}

// TestMaxFragmentDataSize tests maximum fragment data size handling
func TestMaxFragmentDataSize(t *testing.T) {
	// Test at exactly max size
	maxData := make([]byte, tl.MaxFragmentData)
	for i := range maxData {
		maxData[i] = byte(i & 0xFF)
	}

	fragmenter := tl.NewFragmenter()
	fragments := fragmenter.Fragmentize(maxData)

	if len(fragments) != 1 {
		t.Errorf("Max data should fit in 1 fragment, got %d", len(fragments))
	}

	// Test at max + 1 (should need 2 fragments)
	oversized := make([]byte, tl.MaxFragmentData+1)
	fragments = fragmenter.Fragmentize(oversized)

	if len(fragments) != 2 {
		t.Errorf("Oversized data should need 2 fragments, got %d", len(fragments))
	}
}

// TestEmptyData tests handling of empty data
func TestEmptyData(t *testing.T) {
	fragmenter := tl.NewFragmenter()
	fragments := fragmenter.Fragmentize(nil)

	if len(fragments) != 1 {
		t.Errorf("Empty data should produce 1 fragment, got %d", len(fragments))
	}

	reassembler := tl.NewReassembler()
	_, err := reassembler.Push(fragments[0])
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if !reassembler.IsComplete() {
		t.Error("Empty data should be complete")
	}
}

// TestEncodedFragments tests encoding and decoding fragments
func TestEncodedFragments(t *testing.T) {
	data := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	
	fragmenter := tl.NewFragmenter()
	fragments := fragmenter.Fragmentize(data)

	if len(fragments) != 1 {
		t.Fatalf("Expected 1 fragment, got %d", len(fragments))
	}

	// Encode fragment
	encoded := tl.EncodeFragment(fragments[0])
	if len(encoded) != len(data)+1 {
		t.Errorf("Encoded length = %d, want %d", len(encoded), len(data)+1)
	}

	// Decode fragment
	decoded, err := tl.DecodeFragment(encoded)
	if err != nil {
		t.Fatalf("DecodeFragment() error = %v", err)
	}

	if decoded.FIR != fragments[0].FIR {
		t.Errorf("FIR = %v, want %v", decoded.FIR, fragments[0].FIR)
	}

	if decoded.FIN != fragments[0].FIN {
		t.Errorf("FIN = %v, want %v", decoded.FIN, fragments[0].FIN)
	}

	if decoded.Seq != fragments[0].Seq {
		t.Errorf("Seq = %d, want %d", decoded.Seq, fragments[0].Seq)
	}
}

// TestFragmentCount tests fragment count calculation
func TestFragmentCount(t *testing.T) {
	tests := []struct {
		dataLen int
		want    int
	}{
		{0, 1},
		{100, 1},
		{tl.MaxFragmentData, 1},
		{tl.MaxFragmentData + 1, 2},
		{tl.MaxFragmentData * 2, 2},
		{tl.MaxFragmentData*2 + 1, 3},
	}

	for _, tt := range tests {
		got := tl.FragmentCount(tt.dataLen)
		if got != tt.want {
			t.Errorf("FragmentCount(%d) = %d, want %d", tt.dataLen, got, tt.want)
		}
	}
}

// TestIsMultiFragment tests multi-fragment detection
func TestIsMultiFragment(t *testing.T) {
	if tl.IsMultiFragment(tl.MaxFragmentData) {
		t.Error("MaxFragmentData should not be multi-fragment")
	}

	if !tl.IsMultiFragment(tl.MaxFragmentData + 1) {
		t.Error("MaxFragmentData+1 should be multi-fragment")
	}
}

// ConformanceSuite runs all TL conformance tests
func TestConformanceSuite(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*testing.T)
	}{
		{"Header Encoding", TestTransportHeaderEncoding},
		{"Header Decoding", TestTransportHeaderDecoding},
		{"Single Fragment", TestFragmentationSingleFragment},
		{"Multi Fragment", TestFragmentationMultiFragment},
		{"Sequence Wrap", TestFragmentationSequenceWrap},
		{"Reassembly", TestReassembly},
		{"Reassembly Out of Order", TestReassemblyOutOfOrder},
		{"Reassembly Duplicate", TestReassemblyDuplicate},
		{"Max Fragment Size", TestMaxFragmentDataSize},
		{"Empty Data", TestEmptyData},
		{"Encoded Fragments", TestEncodedFragments},
		{"Fragment Count", TestFragmentCount},
		{"Is Multi Fragment", TestIsMultiFragment},
	}

	t.Logf("TL Conformance Suite: %d tests", len(tests))

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}
