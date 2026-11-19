package outstation

import (
	"testing"

	"dnp3/internal/al"
)

// TestReadRejectsMalformedRequestHeader verifies that a READ whose object header
// cannot be walked is refused with IIN OBJECT_UNKNOWN instead of being answered
// with a complete integrity response. Before this gate, an unknown qualifier was
// treated as a 4-octet header and the read fell through to "all static data", so
// a request the outstation did not understand looked like a success.
func TestReadRejectsMalformedRequestHeader(t *testing.T) {
	cases := []struct {
		name string
		data []byte
	}{
		{"unknown qualifier", []byte{1, 1, 0x99}},
		{"unknown qualifier with range", []byte{1, 1, 0x99, 0x00, 0x04}},
		{"truncated count16 header", []byte{1, 1, al.QualCount16}},
		{"truncated start/stop header", []byte{1, 1, al.QualStartStop8, 0x00}},
		{"truncated range16 header", []byte{1, 1, al.QualRange16, 0x00, 0x00, 0x04}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ost := NewOutstation(nil)
			ost.Initialize()
			ost.SetDataHandler(&fullStaticHandler{
				binary: []BinaryInput{{Value: true, Quality: BinaryQualityOnline}},
			})

			resp, err := ost.ProcessRequest(&al.APDU{
				Control:  al.AppControl{FIR: true, FIN: true, Seq: 7},
				FuncCode: al.FuncRead,
				Data:     tc.data,
			})
			if err != nil {
				t.Fatalf("ProcessRequest: %v", err)
			}
			iin, err := al.DecodeIIN(resp.Data)
			if err != nil {
				t.Fatalf("decode IIN: %v", err)
			}
			if !iin.ObjectUnknown {
				t.Fatalf("IIN = % X, want OBJECT_UNKNOWN for a malformed READ", resp.Data[:2])
			}
			if len(resp.Data) != 2 {
				t.Fatalf("response carried %d object octets, want none for a refused READ", len(resp.Data)-2)
			}
		})
	}
}

// TestReadAcceptsKnownQualifiers is the positive control for the gate above: a
// well-formed header for every qualifier in the v0 allow-list must still be
// served, so the malformed rejection cannot be satisfied by rejecting all reads.
func TestReadAcceptsKnownQualifiers(t *testing.T) {
	cases := []struct {
		name string
		data []byte
	}{
		{"all objects", []byte{1, 1, al.QualAllObjects}},
		{"count8", []byte{1, 1, al.QualCount8, 0x01}},
		{"index8", []byte{1, 1, al.QualIndex8, 0x01}},
		{"start/stop", []byte{1, 1, al.QualStartStop8, 0x00, 0x00}},
		{"count16", []byte{1, 1, al.QualCount16, 0x01, 0x00}},
		{"range16", []byte{1, 1, al.QualRange16, 0x00, 0x00, 0x00, 0x00}},
		{"class 0", []byte{60, 1, al.QualAllObjects}},
		{"multi-header", []byte{60, 1, al.QualAllObjects, 1, 1, al.QualAllObjects}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ost := NewOutstation(nil)
			ost.Initialize()
			ost.SetDataHandler(&fullStaticHandler{
				binary: []BinaryInput{{Value: true, Quality: BinaryQualityOnline}},
			})

			resp, err := ost.ProcessRequest(&al.APDU{
				Control:  al.AppControl{FIR: true, FIN: true, Seq: 8},
				FuncCode: al.FuncRead,
				Data:     tc.data,
			})
			if err != nil {
				t.Fatalf("ProcessRequest: %v", err)
			}
			iin, err := al.DecodeIIN(resp.Data)
			if err != nil {
				t.Fatalf("decode IIN: %v", err)
			}
			if iin.ObjectUnknown || iin.ParameterError || iin.FuncUnknown {
				t.Fatalf("IIN = % X, want no error bit for a well-formed READ", resp.Data[:2])
			}
			if len(resp.Data) <= 2 {
				t.Fatal("well-formed READ returned no objects")
			}
		})
	}
}

// TestReadRejectsShortRequest verifies the independent-CHECK pushback: a
// nonempty 1- or 2-octet READ, and a zero-length READ, name no object at all.
// Before this gate they fell through to the "all static data" fallback and were
// answered with a full integrity response plus IIN 00 00, reporting success for
// a request the outstation never parsed.
func TestReadRejectsShortRequest(t *testing.T) {
	cases := []struct {
		name string
		data []byte
	}{
		{"zero length", []byte{}},
		{"nil", nil},
		{"one octet", []byte{1}},
		{"two octets", []byte{1, 1}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ost := NewOutstation(nil)
			ost.Initialize()
			ost.SetDataHandler(&fullStaticHandler{
				binary: []BinaryInput{{Value: true, Quality: BinaryQualityOnline}},
			})

			resp, err := ost.ProcessRequest(&al.APDU{
				Control:  al.AppControl{FIR: true, FIN: true, Seq: 9},
				FuncCode: al.FuncRead,
				Data:     tc.data,
			})
			if err != nil {
				t.Fatalf("ProcessRequest: %v", err)
			}
			iin, err := al.DecodeIIN(resp.Data)
			if err != nil {
				t.Fatalf("decode IIN: %v", err)
			}
			if !iin.ObjectUnknown {
				t.Fatalf("IIN = % X, want OBJECT_UNKNOWN for a short READ", resp.Data[:2])
			}
			if len(resp.Data) != 2 {
				t.Fatalf("response carried %d object octets, want none for a refused READ", len(resp.Data)-2)
			}
		})
	}
}

// TestReadRejectsTrailingPartialHeader verifies that surplus bytes after the
// last complete object header are refused. The header walk stops when fewer
// than 3 octets remain, so without an explicit check a valid header followed by
// 1 or 2 stray octets was answered as a complete read.
func TestReadRejectsTrailingPartialHeader(t *testing.T) {
	cases := []struct {
		name string
		data []byte
	}{
		{"one trailing octet after class 0", []byte{60, 1, al.QualAllObjects, 0x01}},
		{"two trailing octets after class 0", []byte{60, 1, al.QualAllObjects, 0x01, 0x01}},
		{"one trailing octet after binary input", []byte{1, 1, al.QualAllObjects, 0x01}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ost := NewOutstation(nil)
			ost.Initialize()
			ost.SetDataHandler(&fullStaticHandler{
				binary: []BinaryInput{{Value: true, Quality: BinaryQualityOnline}},
			})

			resp, err := ost.ProcessRequest(&al.APDU{
				Control:  al.AppControl{FIR: true, FIN: true, Seq: 10},
				FuncCode: al.FuncRead,
				Data:     tc.data,
			})
			if err != nil {
				t.Fatalf("ProcessRequest: %v", err)
			}
			iin, err := al.DecodeIIN(resp.Data)
			if err != nil {
				t.Fatalf("decode IIN: %v", err)
			}
			if !iin.ObjectUnknown {
				t.Fatalf("IIN = % X, want OBJECT_UNKNOWN for trailing partial header", resp.Data[:2])
			}
			if len(resp.Data) != 2 {
				t.Fatalf("response carried %d object octets, want none for a refused READ", len(resp.Data)-2)
			}
		})
	}
}

// TestRefusedReadDoesNotLeakIINIntoNextRequest verifies the error bit is
// request-scoped: a refused short READ must not leave OBJECT_UNKNOWN set for the
// following well-formed READ, which would make a healthy exchange look failed.
func TestRefusedReadDoesNotLeakIINIntoNextRequest(t *testing.T) {
	ost := NewOutstation(nil)
	ost.Initialize()
	ost.SetDataHandler(&fullStaticHandler{
		binary: []BinaryInput{{Value: true, Quality: BinaryQualityOnline}},
	})

	bad, err := ost.ProcessRequest(&al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 11},
		FuncCode: al.FuncRead,
		Data:     []byte{1, 1},
	})
	if err != nil {
		t.Fatalf("ProcessRequest (short READ): %v", err)
	}
	badIIN, err := al.DecodeIIN(bad.Data)
	if err != nil {
		t.Fatalf("decode IIN: %v", err)
	}
	if !badIIN.ObjectUnknown {
		t.Fatalf("short READ IIN = % X, want OBJECT_UNKNOWN", bad.Data[:2])
	}

	good, err := ost.ProcessRequest(&al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 12},
		FuncCode: al.FuncRead,
		Data:     []byte{60, 1, al.QualAllObjects},
	})
	if err != nil {
		t.Fatalf("ProcessRequest (valid READ): %v", err)
	}
	goodIIN, err := al.DecodeIIN(good.Data)
	if err != nil {
		t.Fatalf("decode IIN: %v", err)
	}
	if goodIIN.ObjectUnknown || goodIIN.ParameterError || goodIIN.FuncUnknown {
		t.Fatalf("valid READ after a refusal has IIN = % X, want no request-scoped error bit", good.Data[:2])
	}
	if len(good.Data) <= 2 {
		t.Fatal("valid READ after a refusal returned no objects")
	}
}
