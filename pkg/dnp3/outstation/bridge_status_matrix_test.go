package outstation

import (
	"testing"

	"dnp3/internal/outstation"
	"dnp3/pkg/dnp3/types"
)

// MEXT-024 — the outstation command-handler bridge must not surface a
// non-success handler status (signalled via status-only, err==nil) as a false
// success. WriteBinaryOutput must return an error so the outstation emits an
// error response the master resolves to a failure status. (The end-to-end
// behaviour is locked in test/integration/operate_status_matrix_test.go; this
// unit test pins the bridge contract directly.)

// statusOnlyHandler returns a fixed ControlStatus with nil error.
type statusOnlyHandler struct {
	status types.ControlStatus
}

func (h *statusOnlyHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	s := h.status
	return &s, nil
}
func (h *statusOnlyHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	s := h.status
	return &s, nil
}

func TestBridgeBinaryOutputNonSuccessStatusIsFailure(t *testing.T) {
	rows := []struct {
		name   string
		status types.ControlStatus
		wantOK bool // wantOK true = bridge returns nil (accepted)
	}{
		{"success", types.ControlSuccess, true},
		{"not_supported", types.ControlNotSupported, false},
		{"blocked", types.ControlBlocked, false},
		{"no_select", types.ControlNoSelect, false},
		{"local", types.ControlLocal, false},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			h := &internalDataHandler{commandHandler: &statusOnlyHandler{status: row.status}}
			err := h.WriteBinaryOutput(0, &outstation.CROB{Code: 0x03, Count: 1, Status: 0})
			if row.wantOK {
				if err != nil {
					t.Fatalf("%s: expected accepted (nil error), got %v", row.name, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("%s: expected a failure error from the bridge for non-success status, got nil (false success)", row.name)
			}
		})
	}
}

func TestBridgeAnalogOutputNonSuccessStatusIsFailure(t *testing.T) {
	h := &internalDataHandler{commandHandler: &statusOnlyHandler{status: types.ControlNotSupported}}
	if err := h.WriteAnalogOutput(0, int16(7), 1); err == nil {
		t.Fatal("expected a failure error from the analog bridge for non-success status, got nil (false success)")
	}
}

// crobRecordingHandler records the boolean value the bridge maps from the
// CROB operation type.
type crobRecordingHandler struct {
	gotValue bool
	calls    int
}

func (h *crobRecordingHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	h.calls++
	if bc, ok := cmd.Value.(*types.BinaryCommandValue); ok {
		h.gotValue = bc.Value
	}
	s := types.ControlSuccess
	return &s, nil
}

func (h *crobRecordingHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	s := types.ControlSuccess
	return &s, nil
}

// TestBridgeCROBOpTypeBooleanMapping locks the public bridge decoder to the
// IEEE 1815 A.8.1 low-nibble enum. The retired one-hot values must not silently
// map to a boolean, and unsupported codes must be rejected explicitly (no
// `crob.Code & 0x1F` fallback that maps 0x03 to false).
func TestBridgeCROBOpTypeBooleanMapping(t *testing.T) {
	cases := []struct {
		name      string
		code      uint8
		wantErr   bool
		wantValue bool
	}{
		{"NUL 0x00 -> false", 0x00, false, false},
		{"PulseOn 0x01 -> true", 0x01, false, true},
		{"PulseOff 0x02 -> false", 0x02, false, false},
		{"LatchOn 0x03 -> true", 0x03, false, true},
		{"LatchOff 0x04 -> false", 0x04, false, false},
		{"retired 0x08 rejected", 0x08, true, false},
		{"retired 0x10 rejected", 0x10, true, false},
		{"trip-close 0x43 rejected", 0x43, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &crobRecordingHandler{}
			d := &internalDataHandler{commandHandler: h}
			err := d.WriteBinaryOutput(0, &outstation.CROB{Code: tc.code, Count: 1})
			if tc.wantErr {
				if err == nil {
					t.Fatalf("code 0x%02X: expected rejection, got nil", tc.code)
				}
				if h.calls != 0 {
					t.Fatalf("code 0x%02X: handler invoked %d times for a rejected code", tc.code, h.calls)
				}
				return
			}
			if err != nil {
				t.Fatalf("code 0x%02X: unexpected error %v", tc.code, err)
			}
			if h.calls != 1 {
				t.Fatalf("code 0x%02X: handler calls = %d, want 1", tc.code, h.calls)
			}
			if h.gotValue != tc.wantValue {
				t.Errorf("code 0x%02X: bridge value = %v, want %v", tc.code, h.gotValue, tc.wantValue)
			}
		})
	}
}
