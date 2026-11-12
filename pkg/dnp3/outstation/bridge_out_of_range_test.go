package outstation

import (
	"errors"
	"testing"

	"dnp3/internal/al"
	"dnp3/internal/outstation"
	"dnp3/pkg/dnp3/types"
)

// AO-004 Appendix T — public-bridge OUT_OF_RANGE propagation.
//
// The public bridge (internalDataHandler.WriteAnalogOutput) must preserve a
// handler's definite non-success ControlStatus so the internal outstation echoes
// the G41 object with the actual wire command status (OUT_OF_RANGE = 0x0C) rather
// than emitting a generic error response. These tests fail on the pre-fix
// behavior (generic error response, IIN FuncUnknown, no echo) and pass after.

// TestWireCommandStatusMapping pins the public->wire command-status mapping used
// by the bridge: public values 0..10 map 1:1 to IEEE 1815 codes 0..10, and
// ControlOutOfRange maps to wire 12 (0x0C). The public synch values 11/12 are
// deliberately unmapped (no collision with the distinct wire meanings).
func TestWireCommandStatusMapping(t *testing.T) {
	cases := []struct {
		in   types.ControlStatus
		want uint8
		ok   bool
	}{
		{types.ControlSuccess, 0, true},
		{types.ControlTimeout, 1, true},
		{types.ControlNoSelect, 2, true},
		{types.ControlBadFormat, 3, true},
		{types.ControlNotSupported, 4, true},
		{types.ControlAlreadyActive, 5, true},
		{types.ControlHardwareError, 6, true},
		{types.ControlLocal, 7, true},
		{types.ControlTooMany, 8, true},
		{types.ControlNotAuthorized, 9, true},
		{types.ControlAutonomous, 10, true},
		{types.ControlOutOfRange, 12, true},
		{types.ControlSynchTimeout, 0, false},
		{types.ControlSynchBlocked, 0, false},
	}
	for _, tc := range cases {
		got, ok := wireCommandStatus(tc.in)
		if ok != tc.ok || (ok && got != tc.want) {
			t.Errorf("wireCommandStatus(%d) = (%d, %v), want (%d, %v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

// bridgeOutOfRangeHandler returns (ControlOutOfRange, error) — the lab fixture's
// contract for an invalid index.
type bridgeOutOfRangeHandler struct{ calls int }

func (h *bridgeOutOfRangeHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	return nil, errors.New("unused")
}
func (h *bridgeOutOfRangeHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	h.calls++
	s := types.ControlOutOfRange
	return &s, errors.New("index out of range")
}

// TestBridgeAnalogOutOfRangeCarriesStatus confirms the bridge returns a
// ControlStatusError carrying wire 0x0C for a (ControlOutOfRange, error) refusal,
// and still returns an error (fail closed).
func TestBridgeAnalogOutOfRangeCarriesStatus(t *testing.T) {
	h := &bridgeOutOfRangeHandler{}
	d := &internalDataHandler{commandHandler: h}
	err := d.WriteAnalogOutput(1, int32(2300), 1)
	if err == nil {
		t.Fatal("expected a failure error from the bridge, got nil")
	}
	cse, ok := outstation.AsControlStatusError(err)
	if !ok {
		t.Fatalf("bridge error %v is not a ControlStatusError", err)
	}
	if cse.Status != 0x0C {
		t.Fatalf("carried status = 0x%02X, want 0x0C (OUT_OF_RANGE)", cse.Status)
	}
	if h.calls != 1 {
		t.Fatalf("handler calls = %d, want 1", h.calls)
	}
}

// TestBridgeAnalogOutOfRangeStatusOnlyNoError confirms a status-only refusal
// ((ControlOutOfRange, nil)) also carries wire 0x0C — the handler is not
// required to raise an error to convey the status.
func TestBridgeAnalogOutOfRangeStatusOnlyNoError(t *testing.T) {
	d := &internalDataHandler{commandHandler: &statusOnlyHandler{status: types.ControlOutOfRange}}
	err := d.WriteAnalogOutput(1, int32(2300), 1)
	if err == nil {
		t.Fatal("expected a failure error for a status-only OUT_OF_RANGE refusal, got nil (false success)")
	}
	cse, ok := outstation.AsControlStatusError(err)
	if !ok || cse.Status != 0x0C {
		t.Fatalf("carried status = %+v (ok=%v), want 0x0C", cse, ok)
	}
}

// TestBridgeAnalogUnmappedStatusFailsClosed confirms a refusal with a status that
// has no wire mapping (e.g. ControlSynchBlocked) still fails closed with a plain
// error (no ControlStatusError) — the bridge never guesses a wire status.
func TestBridgeAnalogUnmappedStatusFailsClosed(t *testing.T) {
	d := &internalDataHandler{commandHandler: &statusOnlyHandler{status: types.ControlSynchBlocked}}
	err := d.WriteAnalogOutput(1, int32(2300), 1)
	if err == nil {
		t.Fatal("expected a failure error, got nil (false success)")
	}
	if _, ok := outstation.AsControlStatusError(err); ok {
		t.Fatalf("unmapped status unexpectedly carried a wire status: %v", err)
	}
}

// TestBridgeToOutstationOutOfRangeEchoEndToEnd drives the full path: the public
// bridge refusing with OUT_OF_RANGE must produce a wire G41 echo object carrying
// index 1, the request variation/value and status 0x0C, with a clean IIN.
func TestBridgeToOutstationOutOfRangeEchoEndToEnd(t *testing.T) {
	cases := []struct {
		name      string
		variation uint8
		value     []byte
	}{
		{"V1 int32", 1, []byte{0xFC, 0x08, 0x00, 0x00}},
		{"V2 int16", 2, []byte{0x39, 0x30}},
		{"V3 float32", 3, []byte{0x00, 0x00, 0xC0, 0x3F}},
		{"V4 float64", 4, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF8, 0x3F}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inst := outstation.NewOutstation(nil)
			inst.Initialize()
			inst.Start()
			inst.SetDataHandler(&internalDataHandler{commandHandler: &bridgeOutOfRangeHandler{}})

			data := []byte{0x29, tc.variation, 0x28, 0x01, 0x00, 0x01, 0x00}
			data = append(data, tc.value...)
			data = append(data, 0x00)
			resp, err := inst.ProcessRequest(&al.APDU{
				Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
				FuncCode: al.FuncDirectOperate,
				Data:     data,
			})
			if err != nil {
				t.Fatalf("expected a G41 echo, got error: %v", err)
			}
			if resp == nil {
				t.Fatal("expected a non-nil response")
			}
			want := append([]byte{0x00, 0x00}, byte(0x29), tc.variation, 0x28, 0x01, 0x00, 0x01, 0x00)
			want = append(want, tc.value...)
			want = append(want, 0x0C)
			if string(resp.Data) != string(want) {
				t.Fatalf("echo = %X, want %X", resp.Data, want)
			}
		})
	}
}
