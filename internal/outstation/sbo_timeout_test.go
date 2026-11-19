package outstation

import (
	"testing"

	"dnp3/internal/al"
)

// DNP3-PULSE-REGRESSION-001 (library-level pin).
//
// The public TCP server used to build an internal Config literal without
// SBOTimeout. A zero timeout made handleOperate treat every OPERATE after a
// successful SELECT as expired, returning an IIN-only error response instead
// of the required G12V1 command-status echo. The adapter-level end-to-end
// regression in the EG628 repository reproduces the exact Windows tester
// bytes; this test pins the generic outstation fallback.
func TestZeroSBOTimeoutUsesDefaultAndEchoesStatus(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SBOTimeout = 0 // simulate a caller that leaves the field unset

	h := &crobEchoHandler{}
	o := NewOutstation(cfg)
	o.Initialize()
	o.Start()
	o.SetDataHandler(h)

	selectResp, err := o.ProcessRequest(g12ControlAPDU(al.FuncSelect, windowsTesterCROBIndex2LatchOff))
	if err != nil {
		t.Fatalf("SELECT failed: %v", err)
	}
	assertG12Echo(t, selectResp, windowsTesterCROBIndex2LatchOff)
	if h.calls != 0 {
		t.Fatalf("SELECT invoked handler %d times, want 0", h.calls)
	}

	operateResp, err := o.ProcessRequest(g12ControlAPDU(al.FuncOperate, windowsTesterCROBIndex2LatchOff))
	if err != nil {
		t.Fatalf("OPERATE failed: %v", err)
	}
	echo := assertG12Echo(t, operateResp, windowsTesterCROBIndex2LatchOff)
	if echo[len(echo)-1] != 0x00 {
		t.Fatalf("command status = %02X, want SUCCESS 00", echo[len(echo)-1])
	}
	if h.calls != 1 || h.lastIndex != 2 {
		t.Fatalf("handler calls=%d index=%d, want 1 and 2", h.calls, h.lastIndex)
	}
}
