package outstation

import (
	"errors"
	"testing"

	"dnp3/internal/al"
	"dnp3/internal/outstation"
	"dnp3/pkg/dnp3/types"
)

// AO-002 F2 — the analog bridge must fail closed unless the handler returns an
// explicit ControlSuccess. A nil status (handler returned (nil, nil)) or any
// non-success status must not be coerced into success; otherwise the outstation
// echoes a SUCCESS command status for a command the handler did not accept.

type scriptedAnalogHandler struct {
	status *types.ControlStatus
	err    error
	calls  int
}

func (h *scriptedAnalogHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	return nil, errors.New("unused")
}

func (h *scriptedAnalogHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	h.calls++
	return h.status, h.err
}

func TestBridgeAnalogOutputF2FailClosed(t *testing.T) {
	success := types.ControlSuccess
	notSupported := types.ControlNotSupported
	blocked := types.ControlHardwareError
	local := types.ControlLocal

	rows := []struct {
		name    string
		status  *types.ControlStatus
		err     error
		wantErr bool
	}{
		{"explicit success", &success, nil, false},
		{"nil status (nil, nil)", nil, nil, true},
		{"not supported", &notSupported, nil, true},
		{"hardware_error", &blocked, nil, true},
		{"local", &local, nil, true},
		{"nil status with error", nil, errors.New("boom"), true},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			h := &scriptedAnalogHandler{status: row.status, err: row.err}
			d := &internalDataHandler{commandHandler: h}
			err := d.WriteAnalogOutput(0, int32(2300), 1)
			if row.wantErr {
				if err == nil {
					t.Fatalf("%s: expected failure, got nil (false success)", row.name)
				}
				if h.calls != 1 {
					t.Fatalf("%s: handler calls = %d, want 1", row.name, h.calls)
				}
				return
			}
			if err != nil {
				t.Fatalf("%s: expected success, got %v", row.name, err)
			}
			if h.calls != 1 {
				t.Fatalf("%s: handler calls = %d, want 1", row.name, h.calls)
			}
		})
	}
}

// TestBridgeAnalogOutputF2FalseSuccessNotPropagated wires the public bridge into
// an internal outstation and drives a G41V1 DirectOperate end to end: a handler
// that returns (nil, nil) must not yield a SUCCESS echo object; the command must
// surface as a failure response.
func TestBridgeAnalogOutputF2FalseSuccessNotPropagated(t *testing.T) {
	inst := outstation.NewOutstation(nil)
	inst.Initialize()
	inst.Start()
	inst.SetDataHandler(&internalDataHandler{
		commandHandler: &scriptedAnalogHandler{status: nil, err: nil},
	})

	req := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncDirectOperate,
		Data:     []byte{0x29, 0x01, 0x28, 0x01, 0x00, 0x00, 0x00, 0xFC, 0x08, 0x00, 0x00, 0x00},
	}
	resp, err := inst.ProcessRequest(req)
	if err == nil {
		t.Fatalf("nil analog status must fail closed, got a success response: %v", resp)
	}
	if resp != nil {
		t.Fatalf("expected nil response (no SUCCESS echo), got %v", resp)
	}
}
