package outstation

import (
	"testing"

	"dnp3/internal/al"
)

// TestEG628RequestScopedIINDoesNotPersist pins the correction for IIN error bits
// that leaked from one request into every later reply. IEEE 1815 IIN describes
// the response it is carried in, so a rejected request must not leave
// ParameterError/ObjectUnknown/FuncUnknown set on a subsequent READ.
//
// This is the reported EG628 symptom shape: a master saw IIN PARAMETER_ERROR on
// a Class 0+1+2+3 integrity READ with no parameter error in that request.
func TestEG628RequestScopedIINDoesNotPersist(t *testing.T) {
	rejections := []struct {
		name string
		req  *al.APDU
	}{
		{
			name: "G41 WRITE rejected",
			req: &al.APDU{
				Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
				FuncCode: al.FuncWrite,
				Data:     []byte{41, 1, 0x28, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			},
		},
		{
			name: "unsupported WRITE group",
			req: &al.APDU{
				Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
				FuncCode: al.FuncWrite,
				Data:     []byte{99, 1, 0x28, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			},
		},
		{
			name: "malformed WRITE header",
			req: &al.APDU{
				Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
				FuncCode: al.FuncWrite,
				Data:     []byte{12, 1},
			},
		},
		{
			name: "unsupported function code",
			req: &al.APDU{
				Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
				FuncCode: 0x2B,
			},
		},
	}

	for _, tc := range rejections {
		t.Run(tc.name, func(t *testing.T) {
			ost := NewOutstation(nil)
			if err := ost.Initialize(); err != nil {
				t.Fatalf("Initialize: %v", err)
			}
			ost.SetDataHandler(NewDefaultDataHandler())

			// The rejected request may or may not return a response; either way it
			// must not contaminate the next reply.
			_, _ = ost.ProcessRequest(tc.req)

			resp, err := ost.ProcessRequest(&al.APDU{
				Control:  al.AppControl{FIR: true, FIN: true, Seq: 2},
				FuncCode: al.FuncRead,
				Data:     []byte{0x3C, 0x01, 0x06},
			})
			if err != nil {
				t.Fatalf("subsequent READ: %v", err)
			}
			if resp == nil {
				t.Fatal("subsequent READ returned no response")
			}
			iin, err := al.DecodeIIN(resp.Data)
			if err != nil {
				t.Fatalf("DecodeIIN: %v", err)
			}
			if iin.ParameterError {
				t.Errorf("ParameterError from a rejected request leaked into the next READ")
			}
			if iin.ObjectUnknown {
				t.Errorf("ObjectUnknown from a rejected request leaked into the next READ")
			}
			if iin.FuncUnknown {
				t.Errorf("FuncUnknown from a rejected request leaked into the next READ")
			}
		})
	}
}

// TestEG628StationIINStillReported guards the other half of the contract: bits
// that describe station conditions must still be reported, so the per-request
// clear must not blank the whole IIN.
func TestEG628StationIINStillReported(t *testing.T) {
	ost := NewOutstation(nil)
	if err := ost.Initialize(); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	ost.SetDataHandler(NewDefaultDataHandler())

	ost.mu.Lock()
	ost.iin.DeviceTrouble = true
	ost.iin.NeedTime = true
	ost.mu.Unlock()

	resp, err := ost.ProcessRequest(&al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncRead,
		Data:     []byte{0x3C, 0x01, 0x06},
	})
	if err != nil {
		t.Fatalf("READ: %v", err)
	}
	iin, err := al.DecodeIIN(resp.Data)
	if err != nil {
		t.Fatalf("DecodeIIN: %v", err)
	}
	if !iin.DeviceTrouble || !iin.NeedTime {
		t.Fatalf("station IIN conditions were cleared: %+v", iin)
	}
}

// TestEG628RejectedRequestStillReportsParameterError confirms the clear happens
// before dispatch, so the response to the rejected request itself still carries
// the error bit rather than being silently cleaned.
func TestEG628RejectedRequestStillReportsParameterError(t *testing.T) {
	ost := NewOutstation(nil)
	if err := ost.Initialize(); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	ost.SetDataHandler(NewDefaultDataHandler())

	_, err := ost.ProcessRequest(&al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncWrite,
		Data:     []byte{41, 1, 0x28, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	})
	if err == nil {
		t.Fatal("expected the G41 WRITE to be rejected")
	}
	if !ost.IIN().ParameterError {
		t.Fatal("the rejected request must still set ParameterError on the station IIN")
	}
}
