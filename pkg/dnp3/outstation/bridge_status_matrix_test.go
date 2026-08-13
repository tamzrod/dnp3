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
			err := h.WriteBinaryOutput(0, &outstation.CROB{Code: 0x08, Count: 1, Status: 0})
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
