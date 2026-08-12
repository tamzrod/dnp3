package al

import (
	"bytes"
	"testing"
)

// TestObjectHeaderEncodeAllObjects verifies the 0x06 (all-objects) header
// matches the golden Class-0 request layout: group, variation, 0x06, 0x00.
func TestObjectHeaderEncodeAllObjects(t *testing.T) {
	tests := []struct {
		name string
		h    ObjectHeader
		want []byte
	}{
		{"G1V1 all", ObjectHeader{Group: 1, Variation: 1, Qualifier: QualAllObjects}, []byte{0x01, 0x01, 0x06, 0x00}},
		{"G30V1 all", ObjectHeader{Group: 30, Variation: 1, Qualifier: QualAllObjects}, []byte{0x1E, 0x01, 0x06, 0x00}},
		{"G20V1 all", ObjectHeader{Group: 20, Variation: 1, Qualifier: QualAllObjects}, []byte{0x14, 0x01, 0x06, 0x00}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.Encode(nil)
			if err != nil {
				t.Fatalf("Encode error: %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("Encode = % X, want % X", got, tt.want)
			}
		})
	}
}

// TestObjectHeaderEncodeCount8 verifies the 0x07 (8-bit count) header matches
// the golden Class-0 response layout: group, variation, 0x07, count.
func TestObjectHeaderEncodeCount8(t *testing.T) {
	tests := []struct {
		name string
		h    ObjectHeader
		want []byte
	}{
		{"G1V1 count=1", ObjectHeader{Group: 1, Variation: 1, Qualifier: QualCount8, Count: 1}, []byte{0x01, 0x01, 0x07, 0x01}},
		{"G30V1 count=4", ObjectHeader{Group: 30, Variation: 1, Qualifier: QualCount8, Count: 4}, []byte{0x1E, 0x01, 0x07, 0x04}},
		{"G20V1 count=2", ObjectHeader{Group: 20, Variation: 1, Qualifier: QualCount8, Count: 2}, []byte{0x14, 0x01, 0x07, 0x02}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.Encode(nil)
			if err != nil {
				t.Fatalf("Encode error: %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("Encode = % X, want % X", got, tt.want)
			}
		})
	}
}

// TestObjectHeaderEncodeAppend verifies Encode appends to dst without
// overwriting existing content.
func TestObjectHeaderEncodeAppend(t *testing.T) {
	dst := []byte{0xAA, 0xBB}
	h := ObjectHeader{Group: 1, Variation: 1, Qualifier: QualAllObjects}
	got, err := h.Encode(dst)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	want := []byte{0xAA, 0xBB, 0x01, 0x01, 0x06, 0x00}
	if !bytes.Equal(got, want) {
		t.Fatalf("Encode append = % X, want % X", got, want)
	}
}

// TestEncodeObjectHeadersMultiple verifies a Class-0 request with three
// all-objects headers encodes to the expected concatenated bytes.
func TestEncodeObjectHeadersMultiple(t *testing.T) {
	headers := []ObjectHeader{
		{Group: 1, Variation: 1, Qualifier: QualAllObjects},
		{Group: 30, Variation: 1, Qualifier: QualAllObjects},
		{Group: 20, Variation: 1, Qualifier: QualAllObjects},
	}
	got, err := EncodeObjectHeaders(nil, headers)
	if err != nil {
		t.Fatalf("EncodeObjectHeaders error: %v", err)
	}
	want := []byte{
		0x01, 0x01, 0x06, 0x00,
		0x1E, 0x01, 0x06, 0x00,
		0x14, 0x01, 0x06, 0x00,
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("headers = % X, want % X", got, want)
	}
}

// TestObjectHeaderEncodeUnsupported verifies unsupported qualifiers error.
func TestObjectHeaderEncodeUnsupported(t *testing.T) {
	h := ObjectHeader{Group: 1, Variation: 1, Qualifier: 0xFF}
	if _, err := h.Encode(nil); err == nil {
		t.Fatalf("expected error for unsupported qualifier, got nil")
	}
}

// TestObjectHeaderEncodedSize verifies the EncodedSize matches Encode length
// for each supported v0 qualifier.
func TestObjectHeaderEncodedSize(t *testing.T) {
	tests := []struct {
		name string
		h    ObjectHeader
	}{
		{"all", ObjectHeader{Group: 1, Variation: 1, Qualifier: QualAllObjects}},
		{"count8", ObjectHeader{Group: 1, Variation: 1, Qualifier: QualCount8, Count: 10}},
		{"index8", ObjectHeader{Group: 12, Variation: 1, Qualifier: QualIndex8, Count: 1}},
		{"range16", ObjectHeader{Group: 1, Variation: 1, Qualifier: QualRange16, Start: 0, Stop: 9}},
		{"count16", ObjectHeader{Group: 30, Variation: 1, Qualifier: QualCount16, Count: 300}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc, err := tt.h.Encode(nil)
			if err != nil {
				t.Fatalf("Encode error: %v", err)
			}
			if got := tt.h.EncodedSize(); got != len(enc) {
				t.Fatalf("EncodedSize = %d, want %d (len of Encode)", got, len(enc))
			}
		})
	}
}

// TestObjectHeaderEncodeRange16LSB verifies the range qualifier encodes the
// 16-bit start/stop little-endian.
func TestObjectHeaderEncodeRange16LSB(t *testing.T) {
	h := ObjectHeader{Group: 1, Variation: 1, Qualifier: QualRange16, Start: 0x1234, Stop: 0x5678}
	got, err := h.Encode(nil)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	want := []byte{0x01, 0x01, 0x28, 0x34, 0x12, 0x78, 0x56}
	if !bytes.Equal(got, want) {
		t.Fatalf("range16 = % X, want % X", got, want)
	}
}

// TestDecodeObjectHeaderValid verifies decode of each supported v0 qualifier
// round-trips against Encode.
func TestDecodeObjectHeaderValid(t *testing.T) {
	tests := []struct {
		name string
		h    ObjectHeader
	}{
		{"all-objects", ObjectHeader{Group: 1, Variation: 1, Qualifier: QualAllObjects}},
		{"count8", ObjectHeader{Group: 30, Variation: 1, Qualifier: QualCount8, Count: 4}},
		{"index8", ObjectHeader{Group: 12, Variation: 1, Qualifier: QualIndex8, Count: 1}},
		{"range16", ObjectHeader{Group: 1, Variation: 1, Qualifier: QualRange16, Start: 0x1234, Stop: 0x5678}},
		{"count16", ObjectHeader{Group: 30, Variation: 1, Qualifier: QualCount16, Count: 300}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc, err := tt.h.Encode(nil)
			if err != nil {
				t.Fatalf("Encode error: %v", err)
			}
			got, n, err := DecodeObjectHeader(enc, 0)
			if err != nil {
				t.Fatalf("DecodeObjectHeader error: %v", err)
			}
			if n != len(enc) {
				t.Fatalf("consumed = %d, want %d", n, len(enc))
			}
			if got != tt.h {
				t.Fatalf("decoded = %+v, want %+v", got, tt.h)
			}
		})
	}
}

// TestDecodeObjectHeaderAtOffset verifies decode reads from a non-zero offset
// and returns the correct consumed count.
func TestDecodeObjectHeaderAtOffset(t *testing.T) {
	// Prefix bytes then a count8 header.
	prefix := []byte{0xAA, 0xBB}
	hdr := ObjectHeader{Group: 20, Variation: 1, Qualifier: QualCount8, Count: 2}
	enc, err := hdr.Encode(prefix)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	got, n, err := DecodeObjectHeader(enc, len(prefix))
	if err != nil {
		t.Fatalf("DecodeObjectHeader error: %v", err)
	}
	if n != 4 {
		t.Fatalf("consumed = %d, want 4", n)
	}
	if got != hdr {
		t.Fatalf("decoded = %+v, want %+v", got, hdr)
	}
}

// TestDecodeObjectHeaderTruncated verifies truncated input is rejected.
func TestDecodeObjectHeaderTruncated(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"3bytes", []byte{0x01, 0x01, 0x06}}, // need 4
		{"range16 short", []byte{0x01, 0x01, 0x28, 0x00, 0x00, 0x00}},     // need 7
		{"count16 short", []byte{0x1E, 0x01, 0x27, 0x2C}},                  // need 5
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, err := DecodeObjectHeader(tt.data, 0); err == nil {
				t.Fatalf("expected error for truncated input, got nil")
			}
		})
	}
}

// TestDecodeObjectHeaderUnsupportedQualifier verifies unsupported qualifier
// bytes return an error.
func TestDecodeObjectHeaderUnsupportedQualifier(t *testing.T) {
	data := []byte{0x01, 0x01, 0xFF, 0x00}
	if _, _, err := DecodeObjectHeader(data, 0); err == nil {
		t.Fatalf("expected error for unsupported qualifier 0xFF, got nil")
	}
}

// TestDecodeObjectHeaderCount8LSB verifies the count field is read
// little-endian for 0x07 and the value matches the wire byte.
func TestDecodeObjectHeaderCount8LSB(t *testing.T) {
	// Golden G30V1 vector: 1E 01 07 01 E8 03 00 00 01
	// Header is 1E 01 07 01 -> count=1.
	data := []byte{0x1E, 0x01, 0x07, 0x01, 0xE8, 0x03, 0x00, 0x00, 0x01}
	h, n, err := DecodeObjectHeader(data, 0)
	if err != nil {
		t.Fatalf("DecodeObjectHeader error: %v", err)
	}
	if n != 4 {
		t.Fatalf("consumed = %d, want 4", n)
	}
	if h.Group != 30 || h.Variation != 1 || h.Qualifier != QualCount8 || h.Count != 1 {
		t.Fatalf("decoded = %+v, want {30 1 0x07 count=1}", h)
	}
}

// TestDecodeObjectHeaderRange16LSB verifies start/stop are read LSB-first.
func TestDecodeObjectHeaderRange16LSB(t *testing.T) {
	// group=1 var=1 qualifier=0x28 start=0x1234(LSB 34 12) stop=0x5678(LSB 78 56)
	data := []byte{0x01, 0x01, 0x28, 0x34, 0x12, 0x78, 0x56}
	h, n, err := DecodeObjectHeader(data, 0)
	if err != nil {
		t.Fatalf("DecodeObjectHeader error: %v", err)
	}
	if n != 7 {
		t.Fatalf("consumed = %d, want 7", n)
	}
	if h.Start != 0x1234 || h.Stop != 0x5678 {
		t.Fatalf("start/stop = %X/%X, want 1234/5678", h.Start, h.Stop)
	}
}

// TestValidQualifier covers the ValidQualifier helper.
func TestValidQualifier(t *testing.T) {
	valid := []uint8{QualAllObjects, QualCount8, QualIndex8, QualRange16, QualCount16}
	for _, q := range valid {
		if !ValidQualifier(q) {
			t.Errorf("ValidQualifier(0x%02X) = false, want true", q)
		}
	}
	if ValidQualifier(0xFF) {
		t.Errorf("ValidQualifier(0xFF) = true, want false")
	}
}
