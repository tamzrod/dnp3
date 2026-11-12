// AO-004 reverse-interop Go outstation fixture — offline unit tests.
//
// These tests exercise the memory-only command handler and the fail-closed
// listen-target guard directly. They open NO socket, start NO server and send NO
// DNP3 traffic.
package main

import (
	"math"
	"testing"

	"dnp3/pkg/dnp3/types"
)

func newHandler() *fixtureCommandHandler { return &fixtureCommandHandler{} }

func TestReadinessConstantsMatchMasterHarness(t *testing.T) {
	if labOutstationAddress != "10.77.0.2" || labOutstationPort != 20000 {
		t.Fatalf("lab endpoint drifted: %s:%d", labOutstationAddress, labOutstationPort)
	}
	if linkOutstationAddr != 4 || linkMasterAddr != 3 {
		t.Fatalf("link addresses drifted: outstation=%d master=%d", linkOutstationAddr, linkMasterAddr)
	}
	if commandIndex != 0 {
		t.Fatalf("command index drifted: %d", commandIndex)
	}
}

func TestValidateListenTarget(t *testing.T) {
	if err := validateListenTarget("10.77.0.2", 20000); err != nil {
		t.Fatalf("expected lab target accepted, got %v", err)
	}
	for _, tc := range []struct {
		host string
		port int
	}{
		{"0.0.0.0", 20000},
		{"10.77.0.2", 20001},
		{"127.0.0.1", 20000},
		{"10.77.0.1", 20000},
	} {
		if err := validateListenTarget(tc.host, tc.port); err == nil {
			t.Fatalf("expected %s:%d to be refused", tc.host, tc.port)
		}
	}
}

func TestAcceptG41V1Int32Index0(t *testing.T) {
	h := newHandler()
	status, err := h.HandleAnalogCommand(&types.ControlOutput{
		Group: 41, Variation: 1, Index: 0,
		CommandType: types.DirectOperate,
		Value:       &types.AnalogInt32CommandValue{Value: -123456},
	})
	if err != nil || status == nil || *status != types.ControlSuccess {
		t.Fatalf("V1 index0: status=%v err=%v", status, err)
	}
	rec, count, seen := h.Snapshot()
	if !seen || count != 1 || rec.Variation != 1 || rec.Int32 != -123456 {
		t.Fatalf("V1 stored wrong: rec=%+v count=%d seen=%v", rec, count, seen)
	}
}

func TestAcceptG41V2Int16Index0(t *testing.T) {
	h := newHandler()
	status, err := h.HandleAnalogCommand(&types.ControlOutput{
		Group: 41, Variation: 2, Index: 0,
		CommandType: types.DirectOperate,
		Value:       &types.AnalogInt16CommandValue{Value: -300},
	})
	if err != nil || status == nil || *status != types.ControlSuccess {
		t.Fatalf("V2 index0: status=%v err=%v", status, err)
	}
	rec, _, _ := h.Snapshot()
	if rec.Variation != 2 || rec.Int16 != -300 {
		t.Fatalf("V2 stored wrong: %+v", rec)
	}
}

func TestAcceptG41V3Float32Index0PreservesBits(t *testing.T) {
	h := newHandler()
	// Signed zero must survive exactly (bit pattern preserved).
	negZero := math.Float32frombits(0x80000000)
	status, err := h.HandleAnalogCommand(&types.ControlOutput{
		Group: 41, Variation: 3, Index: 0,
		CommandType: types.DirectOperate,
		Value:       &types.AnalogFloat32CommandValue{Value: negZero},
	})
	if err != nil || status == nil || *status != types.ControlSuccess {
		t.Fatalf("V3 index0: status=%v err=%v", status, err)
	}
	rec, _, _ := h.Snapshot()
	if rec.Variation != 3 || math.Float32bits(rec.Float32) != 0x80000000 {
		t.Fatalf("V3 signed-zero bits not preserved: %08x", math.Float32bits(rec.Float32))
	}
}

func TestAcceptG41V4Float64Index0PreservesBits(t *testing.T) {
	h := newHandler()
	negZero := math.Float64frombits(0x8000000000000000)
	status, err := h.HandleAnalogCommand(&types.ControlOutput{
		Group: 41, Variation: 4, Index: 0,
		CommandType: types.DirectOperate,
		Value:       &types.AnalogFloat64CommandValue{Value: negZero},
	})
	if err != nil || status == nil || *status != types.ControlSuccess {
		t.Fatalf("V4 index0: status=%v err=%v", status, err)
	}
	rec, _, _ := h.Snapshot()
	if rec.Variation != 4 || math.Float64bits(rec.Float64) != 0x8000000000000000 {
		t.Fatalf("V4 signed-zero bits not preserved: %016x", math.Float64bits(rec.Float64))
	}
}

func TestRejectNonZeroIndexWithOutOfRange(t *testing.T) {
	h := newHandler()
	status, err := h.HandleAnalogCommand(&types.ControlOutput{
		Group: 41, Variation: 1, Index: 7,
		CommandType: types.DirectOperate,
		Value:       &types.AnalogInt32CommandValue{Value: 1},
	})
	if err == nil || status == nil || *status != types.ControlOutOfRange {
		t.Fatalf("expected OUT_OF_RANGE, got status=%v err=%v", status, err)
	}
	if _, count, seen := h.Snapshot(); seen || count != 0 {
		t.Fatalf("rejected command must not mutate state: count=%d seen=%v", count, seen)
	}
}

func TestRejectNonFiniteFloats(t *testing.T) {
	h := newHandler()
	for _, v := range []float32{float32(math.NaN()), float32(math.Inf(1)), float32(math.Inf(-1))} {
		status, err := h.HandleAnalogCommand(&types.ControlOutput{
			Group: 41, Variation: 3, Index: 0,
			CommandType: types.DirectOperate,
			Value:       &types.AnalogFloat32CommandValue{Value: v},
		})
		if err == nil || status == nil || *status == types.ControlSuccess {
			t.Fatalf("non-finite V3 must be rejected: value=%v status=%v err=%v", v, status, err)
		}
	}
}

func TestRejectBinaryCommand(t *testing.T) {
	h := newHandler()
	status, err := h.HandleBinaryCommand(&types.ControlOutput{
		Group: 12, Variation: 1, Index: 0,
		CommandType: types.DirectOperate,
		Value:       &types.BinaryCommandValue{Value: true},
	})
	if err == nil || status == nil || *status != types.ControlNotSupported {
		t.Fatalf("binary must be NOT_SUPPORTED: status=%v err=%v", status, err)
	}
}

func TestRejectUnsupportedVariation(t *testing.T) {
	h := newHandler()
	status, err := h.HandleAnalogCommand(&types.ControlOutput{
		Group: 41, Variation: 9, Index: 0,
		CommandType: types.DirectOperate,
		Value:       &types.AnalogInt32CommandValue{Value: 1},
	})
	if err == nil || status == nil || *status != types.ControlNotSupported {
		t.Fatalf("variation 9 must be NOT_SUPPORTED: status=%v err=%v", status, err)
	}
}

func TestResetClearsState(t *testing.T) {
	h := newHandler()
	if _, err := h.HandleAnalogCommand(&types.ControlOutput{
		Group: 41, Variation: 1, Index: 0,
		CommandType: types.DirectOperate,
		Value:       &types.AnalogInt32CommandValue{Value: 5},
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	h.Reset()
	if _, count, seen := h.Snapshot(); seen || count != 0 {
		t.Fatalf("reset must clear state: count=%d seen=%v", count, seen)
	}
}
