package master

import (
	"testing"

	"dnp3/internal/al"
)

// v0RequestQualifiers is the request-side qualifier allow-list (MEXT-016).
// Only these IEEE 1815 qualifier codes may appear in v0 request object
// headers:
//   - 0x06 all-objects  — Class-0 / integrity reads (buildPollRequest,
//     public buildReadRequest)
//   - 0x00 index8       — single-point Operate/Select control requests
//     (buildControlRequest)
//   - 0x28 range16      — ranged reads (buildRangedReadRequest)
//   - 0x07 count8       — event-class polls (buildPollRequest)
// Everything else must be rejected by al.ObjectHeader.Encode (see
// TestEncodeObjectHeaderUnsupportedQualifier in package al).
var v0RequestQualifiers = map[uint8]struct{}{
	al.QualAllObjects: {},
	al.QualIndex8:     {},
	al.QualRange16:    {},
	al.QualCount8:     {},
	al.QualCount16:    {},
}

// TestPollRequestQualifierAllowList asserts every object header emitted by
// buildPollRequest uses a qualifier from the v0 request allow-list, and that
// the Class-0/integrity poll specifically uses the canonical 0x06
// all-objects qualifier on G60V1 (MEXT-016).
func TestPollRequestQualifierAllowList(t *testing.T) {
	for _, pt := range []PollType{PollIntegrity, PollClass0, PollClass1, PollClass2, PollClass3, PollEvent} {
		data := buildPollRequest(pt)
		if data == nil {
			t.Fatalf("buildPollRequest(%d) returned nil", pt)
		}
		offset := 0
		for offset < len(data) {
			h, consumed, err := al.DecodeObjectHeader(data, offset)
			if err != nil {
				t.Fatalf("poll %d: decode header at %d: %v", pt, offset, err)
			}
			if _, ok := v0RequestQualifiers[h.Qualifier]; !ok {
				t.Fatalf("poll %d: qualifier 0x%02X not in v0 request allow-list", pt, h.Qualifier)
			}
			offset += consumed
		}
	}

	// Class-0 / integrity poll golden: G60V1, 0x06 all-objects.
	got := buildPollRequest(PollIntegrity)
	want := []byte{60, 1, al.QualAllObjects, 0x00}
	if string(got) != string(want) {
		t.Fatalf("integrity poll = % X, want % X", got, want)
	}
}

// TestControlRequestQualifierAllowList asserts the Operate/Select control
// request builder emits the 0x00 (index8) qualifier with a single point,
// locked as the v0 single-point control golden (MEXT-016).
func TestControlRequestQualifierAllowList(t *testing.T) {
	m := &Master{}
	req := m.buildControlRequest(al.FuncDirectOperate, 12, 1, 0x0001, uint8(CROBCodeLatchOn))
	if req == nil {
		t.Fatalf("buildControlRequest returned nil")
	}
	if len(req.Data) < 4 {
		t.Fatalf("control request data too short: % X", req.Data)
	}
	if req.Data[2] != al.QualIndex8 {
		t.Fatalf("control request qualifier = 0x%02X, want 0x00 (index8)", req.Data[2])
	}
	if req.Data[3] != 1 {
		t.Fatalf("control request count = %d, want 1", req.Data[3])
	}
}

// TestRangedReadRequestQualifierAllowList asserts ranged read requests use
// the 0x28 (range16) qualifier (MEXT-016).
func TestRangedReadRequestQualifierAllowList(t *testing.T) {
	data := buildReadRangeRequest(1, 1, 0, 3)
	h, _, err := al.DecodeObjectHeader(data, 0)
	if err != nil {
		t.Fatalf("decode ranged read header: %v", err)
	}
	if h.Qualifier != al.QualRange16 {
		t.Fatalf("ranged read qualifier = 0x%02X, want 0x28 (range16)", h.Qualifier)
	}
	if h.Start != 0 || h.Stop != 3 {
		t.Fatalf("ranged read range = [%d,%d], want [0,3]", h.Start, h.Stop)
	}
}
