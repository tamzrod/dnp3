package outstation

import (
	"errors"
	"strings"
	"testing"

	"dnp3/pkg/dnp3/types"
)

// minimalMVPCommandHandler implements the v0 MVP command contract (DNP3-090):
// it accepts Group 12 Variation 1 binary (CROB) control and rejects analog
// commands (Group 41+) with ControlNotSupported + a clear error.
type minimalMVPCommandHandler struct{}

func (h *minimalMVPCommandHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	if cmd.Group != 12 || cmd.Variation != 1 {
		status := types.ControlNotSupported
		return &status, errors.New("only Group 12 Variation 1 (CROB) control is supported in the v0 MVP profile")
	}
	status := types.ControlSuccess
	return &status, nil
}

func (h *minimalMVPCommandHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	// Analog output control (Group 41+) is outside the v0 MVP profile.
	status := types.ControlNotSupported
	return &status, errors.New("analog control is not supported in the v0 MVP profile (Group 41+ out of scope)")
}

// TestMinimalMVPCommandHandlerSatisfiesInterface verifies DNP3-090: a handler
// implementing the MVP command contract satisfies the CommandHandler interface
// (acceptance: "Compiles with minimal handler").
func TestMinimalMVPCommandHandlerSatisfiesInterface(t *testing.T) {
	var _ CommandHandler = (*minimalMVPCommandHandler)(nil)
	cfg := NewConfig()
	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	srv.SetCommandHandler(&minimalMVPCommandHandler{})
}

// TestMinimalMVPCommandHandlerAcceptsCROB verifies DNP3-090: a G12V1 CROB
// command is accepted (ControlSuccess).
func TestMinimalMVPCommandHandlerAcceptsCROB(t *testing.T) {
	h := &minimalMVPCommandHandler{}
	status, err := h.HandleBinaryCommand(&types.ControlOutput{
		Group:       12,
		Variation:   1,
		Index:       0,
		Value:       &types.BinaryCommandValue{Value: true},
		CommandType: types.DirectOperate,
	})
	if err != nil {
		t.Fatalf("HandleBinaryCommand G12V1: unexpected error %v", err)
	}
	if status == nil || *status != types.ControlSuccess {
		t.Fatalf("G12V1 CROB must succeed, got status=%v", status)
	}
}

// TestMinimalMVPCommandHandlerRejectsAnalog verifies DNP3-090: an analog
// command is rejected with ControlNotSupported + a clear error (acceptance:
// "Analog command rejected" / "Clear error").
func TestMinimalMVPCommandHandlerRejectsAnalog(t *testing.T) {
	h := &minimalMVPCommandHandler{}
	cmd := &types.ControlOutput{
		Group:       41,
		Variation:   1,
		Index:       0,
		Value:       &types.AnalogCommandValue{Value: 42},
		CommandType: types.DirectOperate,
	}
	status, err := h.HandleAnalogCommand(cmd)
	if err == nil {
		t.Fatal("HandleAnalogCommand must return a clear error for v0-out-of-scope analog control")
	}
	if status == nil || *status != types.ControlNotSupported {
		t.Fatalf("analog command must return ControlNotSupported, got status=%v", status)
	}
	if !strings.Contains(err.Error(), "analog") {
		t.Fatalf("analog rejection error must be clear, got: %v", err)
	}
}
