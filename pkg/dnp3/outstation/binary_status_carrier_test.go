package outstation

import (
	"errors"
	"testing"

	"dnp3/internal/outstation"
	"dnp3/pkg/dnp3/types"
)

// binaryWireStatusHandler returns a definite non-success status with or without
// an accompanying error, as a host command handler does when it refuses a CROB.
type binaryWireStatusHandler struct {
	status types.ControlStatus
	err    error
}

func (h *binaryWireStatusHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	s := h.status
	return &s, h.err
}

func (h *binaryWireStatusHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	s := types.ControlNotSupported
	return &s, nil
}

// TestBridgeBinaryOutputCarriesWireStatus pins the G12V1 rejection path: a
// refused binary command must carry its IEEE 1815 command-status octet through
// the public bridge so the internal outstation can echo it. Without the carrier
// the master sees an error response with no command-status object.
func TestBridgeBinaryOutputCarriesWireStatus(t *testing.T) {
	base := errors.New("host refused command")
	for _, tc := range []struct {
		name   string
		err    error
		status types.ControlStatus
		want   uint8
	}{
		{"hardware_error_with_error", base, types.ControlHardwareError, 0x06},
		{"hardware_error_status_only", nil, types.ControlHardwareError, 0x06},
		{"not_authorized", base, types.ControlNotAuthorized, 0x09},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := &internalDataHandler{commandHandler: &binaryWireStatusHandler{status: tc.status, err: tc.err}}
			err := h.WriteBinaryOutput(0, &outstation.CROB{Code: 0x04, Count: 1, Status: 0})
			if err == nil {
				t.Fatal("expected a refusal error, got nil")
			}
			cse, ok := outstation.AsControlStatusError(err)
			if !ok {
				t.Fatalf("error %v does not carry a wire command status", err)
			}
			if cse.Status != tc.want {
				t.Fatalf("carried status = %02X, want %02X", cse.Status, tc.want)
			}
		})
	}
}
