package outstation

import (
	"testing"

	"dnp3/internal/outstation"
	"dnp3/pkg/dnp3/types"
)

// recordingCROBHandler captures the public command received by the bridge.
type recordingCROBHandler struct {
	seen *types.ControlOutput
}

func (h *recordingCROBHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	h.seen = cmd
	s := types.ControlSuccess
	return &s, nil
}

func (h *recordingCROBHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	s := types.ControlNotSupported
	return &s, nil
}

// v0CROBHandler models a consumer that accepts a single latch write but
// rejects pulse repetition and non-zero on/off durations. This is the policy
// that produced a NOT_SUPPORTED status for a Latch On carrying the master's
// default OnTime=1000 before latch times were normalised in the bridge.
type v0CROBHandler struct {
	calls int
}

func (h *v0CROBHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	if cmd.Count > 1 || cmd.OnTime != 0 || cmd.OffTime != 0 {
		s := types.ControlNotSupported
		return &s, nil
	}
	h.calls++
	s := types.ControlSuccess
	return &s, nil
}

func (h *v0CROBHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	s := types.ControlNotSupported
	return &s, nil
}

// TestBridgeNormalisesLatchTimes pins IEEE 1815 A.8.1 latch semantics:
// OnTime/OffTime are pulse parameters and must not reach a command handler
// for Latch On (0x03) or Latch Off (0x04), even when a master sends non-zero
// defaults. Pulse and NUL codes retain their raw times so timed operations
// remain explicitly rejectable rather than being collapsed into a latch.
func TestBridgeNormalisesLatchTimes(t *testing.T) {
	cases := []struct {
		name        string
		code        uint8
		count       uint8
		onTime      uint32
		offTime     uint32
		wantValue   bool
		wantOnTime  uint32
		wantOffTime uint32
	}{
		{"latch_on_with_default_ontime", 0x03, 1, 1000, 0, true, 0, 0},
		{"latch_off_with_default_ontime", 0x04, 1, 1000, 0, false, 0, 0},
		{"latch_on_with_both_times", 0x03, 1, 1000, 500, true, 0, 0},
		{"pulse_on_keeps_ontime", 0x01, 1, 1000, 0, true, 1000, 0},
		{"pulse_off_keeps_offtime", 0x02, 1, 0, 1000, false, 0, 1000},
		{"nul_keeps_times", 0x00, 1, 1000, 0, false, 1000, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &recordingCROBHandler{}
			d := &internalDataHandler{commandHandler: h}
			crob := &outstation.CROB{
				Code: tc.code, Count: tc.count,
				OnTime: tc.onTime, OffTime: tc.offTime,
			}
			if err := d.WriteBinaryOutput(1, crob); err != nil {
				t.Fatalf("WriteBinaryOutput: %v", err)
			}
			if h.seen == nil {
				t.Fatal("command handler was not called")
			}
			if h.seen.Index != 1 || h.seen.Group != 12 || h.seen.Variation != 1 {
				t.Fatalf("unexpected command identity: %+v", h.seen)
			}
			val, ok := h.seen.Value.(*types.BinaryCommandValue)
			if !ok {
				t.Fatalf("value is %T, want *BinaryCommandValue", h.seen.Value)
			}
			if val.Value != tc.wantValue {
				t.Fatalf("value = %v, want %v", val.Value, tc.wantValue)
			}
			if h.seen.OnTime != tc.wantOnTime || h.seen.OffTime != tc.wantOffTime {
				t.Fatalf("times = (%d,%d), want (%d,%d)",
					h.seen.OnTime, h.seen.OffTime, tc.wantOnTime, tc.wantOffTime)
			}
			if h.seen.Count != tc.count {
				t.Fatalf("count = %d, want %d", h.seen.Count, tc.count)
			}
		})
	}
}

// TestBridgeLatchWithDefaultOnTimeNotRejectedByV0Handler reproduces the
// captured interoperability failure at the bridge boundary: a Latch On or
// Latch Off carrying OnTime=1000 must be accepted by a v0 handler, while a
// genuine pulse or NUL command with the same timing must still be rejected.
func TestBridgeLatchWithDefaultOnTimeNotRejectedByV0Handler(t *testing.T) {
	cases := []struct {
		name string
		code uint8
		want int
	}{
		{"latch_on_index1", 0x03, 1},
		{"latch_off_index1", 0x04, 1},
		{"pulse_on_index1", 0x01, 0},
		{"pulse_off_index1", 0x02, 0},
		{"nul_index1", 0x00, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &v0CROBHandler{}
			d := &internalDataHandler{commandHandler: h}
			crob := &outstation.CROB{Code: tc.code, Count: 1, OnTime: 1000}
			err := d.WriteBinaryOutput(1, crob)
			if tc.want == 1 && err != nil {
				t.Fatalf("latch command rejected: %v", err)
			}
			if tc.want == 0 && err == nil {
				t.Fatal("pulse/NUL command with non-zero OnTime was accepted")
			}
			if h.calls != tc.want {
				t.Fatalf("accepted calls = %d, want %d", h.calls, tc.want)
			}
		})
	}
}
