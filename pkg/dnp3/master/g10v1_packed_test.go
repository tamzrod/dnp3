package master

import (
	"testing"

	"dnp3/internal/al"
)

// TestParseBinaryOutputsPackedG10V1 locks the IEEE 1815 packed encoding of
// Group 10 Variation 1 (one bit per point). The captured Windows-master
// response decoded G10V1 as one bit per point; emitting four flags octets
// caused "object header at offset 38: app: truncated object data".
func TestParseBinaryOutputsPackedG10V1(t *testing.T) {
	h := al.ObjectHeader{
		Group: 10, Variation: 1, Qualifier: al.QualCount8, Count: 4,
	}
	hdr, err := al.EncodeObjectHeaders(nil, []al.ObjectHeader{h})
	if err != nil {
		t.Fatalf("encode header: %v", err)
	}
	data := append(hdr, 0x04) // DO3 (index 2) on; all others off

	got := parseBinaryOutputs(data)
	if len(got) != 4 {
		t.Fatalf("parsed %d outputs, want 4", len(got))
	}
	want := []bool{false, false, true, false}
	for i, bo := range got {
		if bo.Index != uint16(i) {
			t.Errorf("point %d index = %d, want %d", i, bo.Index, i)
		}
		if bo.Value != want[i] {
			t.Errorf("point %d value = %v, want %v", i, bo.Value, want[i])
		}
	}
}

// TestParseBinaryOutputsLegacyIndexedG10V1 keeps compatibility with peers
// that still send one flags octet per point under an indexed qualifier.
func TestParseBinaryOutputsLegacyIndexedG10V1(t *testing.T) {
	data := []byte{
		0x0A, 0x01, 0x17, 0x02, // G10V1, count8+index8, count 2
		0x00, 0x01, // index 0, off
		0x01, 0x81, // index 1, on + online
	}
	got := parseBinaryOutputs(data)
	if len(got) != 2 {
		t.Fatalf("parsed %d outputs, want 2", len(got))
	}
	if got[0].Value || !got[1].Value {
		t.Fatalf("values = %v, %v; want false, true", got[0].Value, got[1].Value)
	}
}
