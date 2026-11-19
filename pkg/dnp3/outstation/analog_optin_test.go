package outstation

import (
	"testing"

	"dnp3/pkg/dnp3/types"
)

// countingAnalogHandler counts how often the host analog handler is invoked.
type countingAnalogHandler struct {
	calls int
}

func (h *countingAnalogHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	return nil, nil
}

func (h *countingAnalogHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	h.calls++
	success := types.ControlSuccess
	return &success, nil
}

// TestWithAnalogControlConfigFlag asserts the stable opt-in sets the same
// underlying outstation config flag as the experimental alias and is off by
// default.
func TestWithAnalogControlConfigFlag(t *testing.T) {
	if DefaultConfig().ExperimentalAnalogControl {
		t.Fatalf("outstation analog control must be off by default")
	}
	if !NewConfig(WithAnalogControl()).ExperimentalAnalogControl {
		t.Fatalf("WithAnalogControl did not enable the analog-control flag")
	}
	if !NewConfig(WithExperimentalAnalogControl()).ExperimentalAnalogControl {
		t.Fatalf("WithExperimentalAnalogControl did not enable the analog-control flag")
	}
}

// TestStableOptInGatesAnalogBridge asserts the stable opt-in controls the same
// fail-closed bridge gate as the experimental alias: with it off the host
// handler is never invoked, and with it on the handler is invoked once.
func TestStableOptInGatesAnalogBridge(t *testing.T) {
	t.Run("off by default", func(t *testing.T) {
		h := &countingAnalogHandler{}
		d := &internalDataHandler{analogControlEnabled: DefaultConfig().ExperimentalAnalogControl, commandHandler: h}
		if err := d.WriteAnalogOutput(0, int32(2300), 1); err == nil {
			t.Fatalf("expected fail-closed rejection with the opt-in off")
		}
		if h.calls != 0 {
			t.Fatalf("host analog handler invoked %d times with the opt-in off, want 0", h.calls)
		}
	})

	t.Run("stable opt-in on", func(t *testing.T) {
		h := &countingAnalogHandler{}
		d := &internalDataHandler{analogControlEnabled: NewConfig(WithAnalogControl()).ExperimentalAnalogControl, commandHandler: h}
		if err := d.WriteAnalogOutput(0, int32(2300), 1); err != nil {
			t.Fatalf("expected success with the stable opt-in on, got %v", err)
		}
		if h.calls != 1 {
			t.Fatalf("host analog handler invoked %d times with the opt-in on, want 1", h.calls)
		}
	})
}
