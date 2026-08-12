package outstation

import (
	"testing"

	"dnp3/internal/al"
)

// TestReadEventClassReturnsEmpty verifies DNP3-088: a READ of an event class
// (G60 V2/V3/V4 = Class 1/2/3) against the empty event-buffer stub returns an
// empty (object-less) response deterministically, and does not crash, and does
// NOT fall back to returning static (Class 0) data.
func TestReadEventClassReturnsEmpty(t *testing.T) {
	ost := NewOutstation(nil)
	ost.Initialize()

	cases := []struct {
		name string
		v    uint8
	}{
		{"class1_G60V2", 2},
		{"class2_G60V3", 3},
		{"class3_G60V4", 4},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := &al.APDU{
				Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
				FuncCode: al.FuncRead,
				Data:     []byte{60, tc.v, 0x06, 0x00}, // G60, V<n>, all-objects
			}
			resp, err := ost.ProcessRequest(req)
			if err != nil {
				t.Fatalf("ProcessRequest Class %d events: %v", tc.v-1, err)
			}
			if resp == nil {
				t.Fatal("expected a response (empty event class still returns a response)")
			}
			if resp.FuncCode != al.FuncResponse {
				t.Fatalf("FuncCode = %d, want FuncResponse", resp.FuncCode)
			}
			// resp.Data is [IIN(2)] + object data. Object data must be empty
			// (no event objects) for the empty event-buffer stub.
			if len(resp.Data) < 2 {
				t.Fatalf("response Data too short: %d bytes (need IIN)", len(resp.Data))
			}
			objData := resp.Data[2:]
			if len(objData) != 0 {
				t.Fatalf("Class %d event poll must return empty object data, got %d bytes (DNP3-088 empty stub)",
					tc.v-1, len(objData))
			}
		})
	}
}

// TestReadClass0IntegrityReturnsStaticData verifies DNP3-088 did not regress the
// Class 0 integrity scan (G60 V1): it still returns static data, not empty.
func TestReadClass0IntegrityReturnsStaticData(t *testing.T) {
	ost := NewOutstation(nil)
	ost.Initialize()

	req := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncRead,
		Data:     []byte{60, 1, 0x06, 0x00}, // G60 V1 all-objects (Class 0)
	}
	resp, err := ost.ProcessRequest(req)
	if err != nil {
		t.Fatalf("ProcessRequest Class 0: %v", err)
	}
	if resp == nil {
		t.Fatal("expected a response")
	}
	if len(resp.Data) < 3 {
		t.Fatalf("Class 0 response too short: %d (expected IIN + static object data)", len(resp.Data))
	}
	if len(resp.Data[2:]) == 0 {
		t.Fatal("Class 0 integrity scan must return static data, got empty (regression)")
	}
}

// TestReadEventClassDoesNotPolluteIIN verifies DNP3-088: an empty event-class
// poll does not set any error IIN bits (no crash, no error indication).
func TestReadEventClassDoesNotPolluteIIN(t *testing.T) {
	ost := NewOutstation(nil)
	ost.Initialize()

	req := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncRead,
		Data:     []byte{60, 2, 0x06, 0x00}, // Class 1 events
	}
	resp, err := ost.ProcessRequest(req)
	if err != nil {
		t.Fatalf("ProcessRequest: %v", err)
	}
	iin, err := al.DecodeIIN(resp.Data)
	if err != nil {
		t.Fatalf("DecodeIIN: %v", err)
	}
	if iin.FuncUnknown || iin.ObjectUnknown || iin.ParameterError {
		t.Fatalf("empty event-class poll must not set error IIN, got %+v", iin)
	}
}
