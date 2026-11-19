package outstation

import (
	"testing"

	"dnp3/internal/outstation"
	"dnp3/pkg/dnp3/types"
)

// fullCROBHandler captures the complete public ControlOutput so the bridge
// contract (operation code, count, timing, qualifier and request status) can
// be asserted without collapsing it to a boolean.
type fullCROBHandler struct {
	seen  *types.ControlOutput
	calls int
}

func (h *fullCROBHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	h.calls++
	copyCmd := *cmd
	h.seen = &copyCmd
	s := types.ControlSuccess
	return &s, nil
}

func (h *fullCROBHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	s := types.ControlNotSupported
	return &s, nil
}

func TestBridgePreservesFullG12V1CROB(t *testing.T) {
	cases := []struct {
		name           string
		crob           outstation.CROB
		wantCode       uint8
		wantTopOnTime  uint32
		wantTopOffTime uint32
	}{
		{
			name:           "pulse_on_all_fields",
			crob:           outstation.CROB{Code: 0x01, Count: 2, OnTime: 1000, OffTime: 500, Qualifier: 0x17, Status: 0x00},
			wantCode:       0x01,
			wantTopOnTime:  1000,
			wantTopOffTime: 500,
		},
		{
			name:           "pulse_off_all_fields",
			crob:           outstation.CROB{Code: 0x02, Count: 3, OnTime: 250, OffTime: 750, Qualifier: 0x28, Status: 0x00},
			wantCode:       0x02,
			wantTopOnTime:  250,
			wantTopOffTime: 750,
		},
		{
			name:           "latch_keeps_raw_crob_times_but_normalises_legacy",
			crob:           outstation.CROB{Code: 0x03, Count: 1, OnTime: 1000, OffTime: 900, Qualifier: 0x28},
			wantCode:       0x03,
			wantTopOnTime:  0,
			wantTopOffTime: 0,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &fullCROBHandler{}
			d := &internalDataHandler{commandHandler: h}
			if err := d.WriteBinaryOutput(7, &tc.crob); err != nil {
				t.Fatalf("WriteBinaryOutput: %v", err)
			}
			if h.calls != 1 || h.seen == nil {
				t.Fatalf("handler calls = %d, want 1", h.calls)
			}
			got := h.seen.CROB
			if got == nil {
				t.Fatal("public ControlOutput.CROB is nil; full CROB was collapsed")
			}
			if got.Code != tc.wantCode || got.Count != tc.crob.Count {
				t.Errorf("CROB code/count = (%d,%d), want (%d,%d)", got.Code, got.Count, tc.wantCode, tc.crob.Count)
			}
			if got.OnTime != tc.crob.OnTime || got.OffTime != tc.crob.OffTime {
				t.Errorf("CROB times = (%d,%d), want (%d,%d)", got.OnTime, got.OffTime, tc.crob.OnTime, tc.crob.OffTime)
			}
			if got.Qualifier != tc.crob.Qualifier || got.Status != tc.crob.Status {
				t.Errorf("CROB qualifier/status = (%d,%d), want (%d,%d)", got.Qualifier, got.Status, tc.crob.Qualifier, tc.crob.Status)
			}
			if h.seen.OnTime != tc.wantTopOnTime || h.seen.OffTime != tc.wantTopOffTime {
				t.Errorf("legacy top-level times = (%d,%d), want (%d,%d)", h.seen.OnTime, h.seen.OffTime, tc.wantTopOnTime, tc.wantTopOffTime)
			}
			if h.seen.Group != 12 || h.seen.Variation != 1 || h.seen.Index != 7 {
				t.Errorf("command identity = %d/%d/%d, want 12/1/7", h.seen.Group, h.seen.Variation, h.seen.Index)
			}
		})
	}
}

func TestBridgeNULIsTrueNoOp(t *testing.T) {
	h := &fullCROBHandler{}
	d := &internalDataHandler{commandHandler: h}
	if err := d.WriteBinaryOutput(0, &outstation.CROB{Code: 0x00, Count: 1, OnTime: 1000, OffTime: 0}); err != nil {
		t.Fatalf("NUL must be acknowledged, got %v", err)
	}
	if h.calls != 0 {
		t.Fatalf("NUL reached the command handler %d times, want 0", h.calls)
	}
}

func TestBridgeRejectsMalformedAndUnsupportedCROB(t *testing.T) {
	cases := []struct {
		name string
		crob outstation.CROB
	}{
		{"unsupported_code", outstation.CROB{Code: 0x43, Count: 1}},
		{"pulse_zero_count", outstation.CROB{Code: 0x01, Count: 0, OnTime: 1000}},
		{"pulse_off_zero_count", outstation.CROB{Code: 0x02, Count: 0, OffTime: 500}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &fullCROBHandler{}
			d := &internalDataHandler{commandHandler: h}
			err := d.WriteBinaryOutput(0, &tc.crob)
			if err == nil {
				t.Fatal("malformed/unsupported CROB was accepted")
			}
			if h.calls != 0 {
				t.Fatalf("handler invoked %d times for refused command", h.calls)
			}
			cse, ok := outstation.AsControlStatusError(err)
			if !ok {
				t.Fatalf("refusal has no definite wire command status: %T %v", err, err)
			}
			if cse.Status != uint8(types.ControlBadFormat) {
				t.Errorf("wire status = %d, want %d (bad format)", cse.Status, types.ControlBadFormat)
			}
		})
	}
}
