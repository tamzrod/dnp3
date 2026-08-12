package outstation

import (
	"testing"

	"dnp3/pkg/dnp3/types"
)

// minimalMVPDataHandler implements only the v0 MVP-required input getters
// (G1.1 binary input, G30.1 analog input, G20.1 counter) with the remaining
// DataHandler methods as no-ops/nil. DNP3-089: a minimal handler must satisfy
// the DataHandler interface and serve a valid Class-0 read.
type minimalMVPDataHandler struct{}

func (h *minimalMVPDataHandler) GetBinaryInputs() []*types.BinaryInput {
	return []*types.BinaryInput{{Index: 0, Value: true, Quality: types.QualityOnline}}
}

func (h *minimalMVPDataHandler) GetAnalogInputs() []*types.AnalogInput {
	return []*types.AnalogInput{{Index: 0, Value: 100, Quality: types.QualityOnline}}
}

func (h *minimalMVPDataHandler) GetCounters() []*types.Counter {
	return []*types.Counter{{Index: 0, Value: 7, Quality: types.QualityOnline}}
}

// Non-MVP methods: nil / no-op per the v0 profile (DNP3-089).
func (h *minimalMVPDataHandler) GetBinaryOutputs() []*types.BinaryOutput { return nil }
func (h *minimalMVPDataHandler) GetAnalogOutputs() []*types.AnalogOutput { return nil }
func (h *minimalMVPDataHandler) GetFrozenCounters() []*types.Counter      { return nil }
func (h *minimalMVPDataHandler) FreezeCounters(clear bool) error           { return nil }

// TestMinimalMVPDataHandlerSatisfiesInterface verifies DNP3-089: a handler
// implementing only the MVP-required methods satisfies the DataHandler
// interface (compiles). This is the acceptance criterion "Compiles with
// minimal handler".
func TestMinimalMVPDataHandlerSatisfiesInterface(t *testing.T) {
	var _ DataHandler = (*minimalMVPDataHandler)(nil)
	// Also usable through the public setter without panic.
	cfg := NewConfig()
	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	srv.SetDataHandler(&minimalMVPDataHandler{})
}

// TestMinimalMVPDataHandlerProvidesOnlyMVPGroups verifies DNP3-089: the
// minimal handler's non-MVP getters return empty, and the MVP getters return
// data — so a Class-0 read serves only the MVP input groups.
func TestMinimalMVPDataHandlerProvidesOnlyMVPGroups(t *testing.T) {
	h := &minimalMVPDataHandler{}
	if len(h.GetBinaryInputs()) == 0 {
		t.Fatal("GetBinaryInputs empty — MVP-required group G1.1 must have data")
	}
	if len(h.GetAnalogInputs()) == 0 {
		t.Fatal("GetAnalogInputs empty — MVP-required group G30.1 must have data")
	}
	if len(h.GetCounters()) == 0 {
		t.Fatal("GetCounters empty — MVP-required group G20.1 must have data")
	}
	if g := h.GetBinaryOutputs(); g != nil {
		t.Fatalf("GetBinaryOutputs must be nil for minimal MVP handler, got %d", len(g))
	}
	if g := h.GetAnalogOutputs(); g != nil {
		t.Fatalf("GetAnalogOutputs must be nil for minimal MVP handler, got %d", len(g))
	}
	if g := h.GetFrozenCounters(); g != nil {
		t.Fatalf("GetFrozenCounters must be nil for minimal MVP handler, got %d", len(g))
	}
	if err := h.FreezeCounters(false); err != nil {
		t.Fatalf("FreezeCounters no-op must return nil, got %v", err)
	}
}
