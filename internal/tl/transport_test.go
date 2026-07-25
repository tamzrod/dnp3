package tl

import (
	"bytes"
	"testing"
)

func TestFragmentHeader(t *testing.T) {
	tests := []struct {
		name     string
		fir      bool
		fin      bool
		seq      uint8
		expected byte
	}{
		{"Single Fragment", true, true, 0, 0xC0},  // FIR + FIN
		{"First of Multi", true, false, 0, 0x80},
		{"Middle Fragment", false, false, 5, 0x05},
		{"Final Fragment", false, true, 10, 0x4A},
		{"Max Seq", true, true, 63, 0xFF},         // 0xC0 | 0x3F
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Fragment{
				FIR:  tt.fir,
				FIN:  tt.fin,
				Seq:  tt.seq,
				Data: []byte{0x01, 0x02, 0x03},
			}

			if got := f.Header(); got != tt.expected {
				t.Errorf("Header() = 0x%02X, want 0x%02X", got, tt.expected)
			}
		})
	}
}

func TestSetHeader(t *testing.T) {
	tests := []struct {
		name        string
		header      byte
		wantFIR     bool
		wantFIN     bool
		wantSeq     uint8
	}{
		{"Single Fragment", 0xC0, true, true, 0},
		{"First of Multi", 0x80, true, false, 0},
		{"Middle", 0x05, false, false, 5},
		{"Final", 0x4A, false, true, 10},
		{"Max Seq", 0xFF, true, true, 63},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f Fragment
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

func TestFragmenterSingleFragment(t *testing.T) {
	f := NewFragmenter()
	data := []byte{0x01, 0x02, 0x03}

	frags := f.Fragmentize(data)

	if len(frags) != 1 {
		t.Fatalf("Expected 1 fragment, got %d", len(frags))
	}

	if !frags[0].FIR || !frags[0].FIN {
		t.Error("Single fragment should have FIR and FIN set")
	}

	if frags[0].Seq != 0 {
		t.Errorf("First fragment seq = %d, want 0", frags[0].Seq)
	}

	if !bytes.Equal(frags[0].Data, data) {
		t.Errorf("Data mismatch: got %v, want %v", frags[0].Data, data)
	}
}

func TestFragmenterMultiFragment(t *testing.T) {
	f := NewFragmenter()
	// Create data larger than one fragment (291 bytes)
	data := make([]byte, 600)
	for i := range data {
		data[i] = byte(i)
	}

	frags := f.Fragmentize(data)

	// 600 bytes / 291 bytes per fragment = 3 fragments
	expectedFrags := 3
	if len(frags) != expectedFrags {
		t.Fatalf("Expected %d fragments, got %d", expectedFrags, len(frags))
	}

	// Check first fragment
	if !frags[0].FIR || frags[0].FIN {
		t.Error("First fragment should have FIR set, FIN clear")
	}
	if frags[0].Seq != 0 {
		t.Errorf("First fragment seq = %d, want 0", frags[0].Seq)
	}
	if len(frags[0].Data) != MaxFragmentData {
		t.Errorf("First fragment data len = %d, want %d", len(frags[0].Data), MaxFragmentData)
	}

	// Check middle fragment
	if frags[1].FIR || frags[1].FIN {
		t.Error("Middle fragment should have FIR and FIN clear")
	}
	if frags[1].Seq != 1 {
		t.Errorf("Middle fragment seq = %d, want 1", frags[1].Seq)
	}

	// Check last fragment
	if frags[2].FIR || !frags[2].FIN {
		t.Error("Last fragment should have FIN set, FIR clear")
	}
	if frags[2].Seq != 2 {
		t.Errorf("Last fragment seq = %d, want 2", frags[2].Seq)
	}

	// Verify data concatenation
	totalLen := len(frags[0].Data) + len(frags[1].Data) + len(frags[2].Data)
	if totalLen != len(data) {
		t.Errorf("Total data length = %d, want %d", totalLen, len(data))
	}
}

func TestFragmenterSequenceNumbers(t *testing.T) {
	f := NewFragmenter()

	// Send multiple messages
	for msg := 0; msg < 3; msg++ {
		data := []byte{byte(msg)}
		frags := f.Fragmentize(data)

		// Each message starts with next sequence number
		expectedStartSeq := uint8(msg)
		if frags[0].Seq != expectedStartSeq {
			t.Errorf("Message %d: first seq = %d, want %d", msg, frags[0].Seq, expectedStartSeq)
		}
	}
}

func TestFragmenterSeqWrap(t *testing.T) {
	f := NewFragmenter()

	// Send 64 single-fragment messages to wrap sequence
	for i := 0; i < 64; i++ {
		data := []byte{byte(i)}
		frags := f.Fragmentize(data)
		if len(frags) != 1 {
			t.Fatalf("Message %d: expected 1 fragment, got %d", i, len(frags))
		}
	}

	// Next message should wrap back to 0
	data := []byte{0xFF}
	frags := f.Fragmentize(data)
	if frags[0].Seq != 0 {
		t.Errorf("After wrap: seq = %d, want 0", frags[0].Seq)
	}
}

func TestReassemblerSingleFragment(t *testing.T) {
	r := NewReassembler()
	data := []byte{0x01, 0x02, 0x03}

	frag := Fragment{
		FIR:  true,
		FIN:  true,
		Seq:  0,
		Data: data,
	}

	result, err := r.Push(frag)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if result == nil {
		t.Fatal("Expected complete message, got nil")
	}

	if !bytes.Equal(result, data) {
		t.Errorf("Reassembled data = %v, want %v", result, data)
	}
}

func TestReassemblerMultiFragment(t *testing.T) {
	r := NewReassembler()
	data := make([]byte, 600)
	for i := range data {
		data[i] = byte(i)
	}

	// Simulate receiving fragments
	frag1 := Fragment{FIR: true, FIN: false, Seq: 0, Data: data[0:291]}
	frag2 := Fragment{FIR: false, FIN: false, Seq: 1, Data: data[291:582]}
	frag3 := Fragment{FIR: false, FIN: true, Seq: 2, Data: data[582:]}

	// Receive first fragment
	result, err := r.Push(frag1)
	if err != nil {
		t.Fatalf("Push(frag1) error = %v", err)
	}
	if result != nil {
		t.Error("First fragment should not complete message")
	}

	// Receive second fragment
	result, err = r.Push(frag2)
	if err != nil {
		t.Fatalf("Push(frag2) error = %v", err)
	}
	if result != nil {
		t.Error("Second fragment should not complete message")
	}

	// Receive third fragment
	result, err = r.Push(frag3)
	if err != nil {
		t.Fatalf("Push(frag3) error = %v", err)
	}
	if result == nil {
		t.Fatal("Third fragment should complete message")
	}

	if !bytes.Equal(result, data) {
		t.Errorf("Reassembled data mismatch")
	}
}

func TestReassemblerSequenceMismatch(t *testing.T) {
	r := NewReassembler()

	frag1 := Fragment{FIR: true, FIN: false, Seq: 0, Data: []byte{0x01}}
	_, err := r.Push(frag1)
	if err != nil {
		t.Fatalf("Push(frag1) error = %v", err)
	}

	// Wrong sequence number
	frag2 := Fragment{FIR: false, FIN: true, Seq: 5, Data: []byte{0x02}}
	_, err = r.Push(frag2)
	if err == nil {
		t.Error("Expected error for sequence mismatch")
	}
}

func TestReassemblerMissingFIR(t *testing.T) {
	r := NewReassembler()

	// Try to push non-FIR fragment without prior FIR
	frag := Fragment{FIR: false, FIN: true, Seq: 0, Data: []byte{0x01}}
	_, err := r.Push(frag)
	if err != ErrMissingFirstFragment {
		t.Errorf("Expected ErrMissingFirstFragment, got %v", err)
	}
}

func TestReassemblerReset(t *testing.T) {
	r := NewReassembler()

	frag := Fragment{FIR: true, FIN: false, Seq: 0, Data: []byte{0x01}}
	_, err := r.Push(frag)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	r.Reset()

	// Should be able to start fresh
	newFrag := Fragment{FIR: true, FIN: true, Seq: 0, Data: []byte{0x02}}
	result, err := r.Push(newFrag)
	if err != nil {
		t.Fatalf("Push() after reset error = %v", err)
	}
	if result == nil {
		t.Error("Expected complete message after reset")
	}
}

func TestDecodeEncodeFragment(t *testing.T) {
	orig := Fragment{
		FIR:  true,
		FIN:  false,
		Seq:  15,
		Data: []byte{0xAA, 0xBB, 0xCC},
	}

	encoded := EncodeFragment(orig)
	decoded, err := DecodeFragment(encoded)
	if err != nil {
		t.Fatalf("DecodeFragment() error = %v", err)
	}

	if decoded.FIR != orig.FIR {
		t.Errorf("FIR = %v, want %v", decoded.FIR, orig.FIR)
	}
	if decoded.FIN != orig.FIN {
		t.Errorf("FIN = %v, want %v", decoded.FIN, orig.FIN)
	}
	if decoded.Seq != orig.Seq {
		t.Errorf("Seq = %d, want %d", decoded.Seq, orig.Seq)
	}
	if !bytes.Equal(decoded.Data, orig.Data) {
		t.Errorf("Data = %v, want %v", decoded.Data, orig.Data)
	}
}

func TestDecodeFragmentInvalid(t *testing.T) {
	_, err := DecodeFragment([]byte{})
	if err != ErrInvalidHeader {
		t.Errorf("Expected ErrInvalidHeader for empty data, got %v", err)
	}
}

func TestFragmentCount(t *testing.T) {
	// MaxFragmentData is now 249 (was 291)
	tests := []struct {
		dataLen  int
		expected int
	}{
		{0, 1},
		{1, 1},
		{249, 1},
		{250, 2},
		{498, 2},
		{747, 3},
	}

	for _, tt := range tests {
		got := FragmentCount(tt.dataLen)
		if got != tt.expected {
			t.Errorf("FragmentCount(%d) = %d, want %d", tt.dataLen, got, tt.expected)
		}
	}
}

func TestIsMultiFragment(t *testing.T) {
	// MaxFragmentData is now 249 (was 291)
	tests := []struct {
		dataLen  int
		expected bool
	}{
		{0, false},
		{1, false},
		{249, false},
		{250, true},
		{1000, true},
	}

	for _, tt := range tests {
		got := IsMultiFragment(tt.dataLen)
		if got != tt.expected {
			t.Errorf("IsMultiFragment(%d) = %v, want %v", tt.dataLen, got, tt.expected)
		}
	}
}

func TestValidateFragment(t *testing.T) {
	validFrag := Fragment{FIR: true, FIN: true, Seq: 0, Data: []byte{0x01}}
	if err := ValidateFragment(validFrag); err != nil {
		t.Errorf("ValidateFragment() error = %v, want nil", err)
	}

	invalidFrag := Fragment{FIR: true, FIN: true, Seq: 64, Data: []byte{0x01}}
	if err := ValidateFragment(invalidFrag); err == nil {
		t.Error("ValidateFragment() should error for invalid sequence")
	}
}

func TestEncodedFragments(t *testing.T) {
	f := NewFragmenter()
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}

	encoded := f.EncodedFragments(data)

	if len(encoded) != 1 {
		t.Fatalf("Expected 1 encoded fragment, got %d", len(encoded))
	}

	// First byte should be header (FIR | FIN | Seq)
	if encoded[0][0] != 0xC0 {
		t.Errorf("Header = 0x%02X, want 0xC0", encoded[0][0])
	}

	// Data should follow header
	if !bytes.Equal(encoded[0][1:], data) {
		t.Errorf("Data = %v, want %v", encoded[0][1:], data)
	}
}
