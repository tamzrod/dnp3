package outstation

import (
	"math"
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
		{"hardware_error", types.ControlHardwareError, false},
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

// aoRecordingHandler records the ControlOutput the analog bridge builds.
type aoRecordingHandler struct {
	got *types.ControlOutput
}

func (h *aoRecordingHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	s := types.ControlSuccess
	return &s, nil
}

func (h *aoRecordingHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	h.got = cmd
	s := types.ControlSuccess
	return &s, nil
}

// TestBridgeAnalogOutputG41V1Semantics locks the AO-002 bridge contract: a G41V1
// analog output is carried as a typed AnalogInt32CommandValue with DirectOperate
// semantics (not the previously hard-coded SelectThenOperate).
func TestBridgeAnalogOutputG41V1Semantics(t *testing.T) {
	h := &aoRecordingHandler{}
	d := &internalDataHandler{commandHandler: h}
	if err := d.WriteAnalogOutput(7, int32(2300), 1); err != nil {
		t.Fatalf("WriteAnalogOutput G41V1: %v", err)
	}
	if h.got == nil {
		t.Fatal("handler not called")
	}
	if h.got.Group != 41 || h.got.Variation != 1 {
		t.Fatalf("group/variation = %d/%d, want 41/1", h.got.Group, h.got.Variation)
	}
	if h.got.CommandType != types.DirectOperate {
		t.Fatalf("CommandType = %v, want DirectOperate (not SelectThenOperate)", h.got.CommandType)
	}
	if h.got.Index != 7 {
		t.Fatalf("index = %d, want 7", h.got.Index)
	}
	v, ok := h.got.Value.(*types.AnalogInt32CommandValue)
	if !ok {
		t.Fatalf("value type = %T, want *types.AnalogInt32CommandValue", h.got.Value)
	}
	if v.Value != 2300 {
		t.Fatalf("value = %d, want 2300", v.Value)
	}
}

// TestBridgeAnalogOutputG41V1RejectsWrongType asserts a G41V1 write with a value
// that is not an int32/int16 is rejected before reaching the handler.
func TestBridgeAnalogOutputG41V1RejectsWrongType(t *testing.T) {
	h := &aoRecordingHandler{}
	d := &internalDataHandler{commandHandler: h}
	if err := d.WriteAnalogOutput(0, float64(1.5), 1); err == nil {
		t.Fatal("expected rejection for a non-integer G41V1 value")
	}
	if h.got != nil {
		t.Fatal("handler invoked for a rejected G41V1 value")
	}
}

// TestBridgeAnalogOutputG41V2Semantics locks the AO-003 bridge contract: a G41V2
// analog output is carried as a typed AnalogInt16CommandValue with DirectOperate
// semantics.
func TestBridgeAnalogOutputG41V2Semantics(t *testing.T) {
	h := &aoRecordingHandler{}
	d := &internalDataHandler{commandHandler: h}
	if err := d.WriteAnalogOutput(9, int16(-42), 2); err != nil {
		t.Fatalf("WriteAnalogOutput G41V2: %v", err)
	}
	if h.got == nil {
		t.Fatal("handler not called")
	}
	if h.got.Group != 41 || h.got.Variation != 2 {
		t.Fatalf("group/variation = %d/%d, want 41/2", h.got.Group, h.got.Variation)
	}
	if h.got.CommandType != types.DirectOperate {
		t.Fatalf("CommandType = %v, want DirectOperate", h.got.CommandType)
	}
	if h.got.Index != 9 {
		t.Fatalf("index = %d, want 9", h.got.Index)
	}
	v, ok := h.got.Value.(*types.AnalogInt16CommandValue)
	if !ok {
		t.Fatalf("value type = %T, want *types.AnalogInt16CommandValue", h.got.Value)
	}
	if v.Value != -42 {
		t.Fatalf("value = %d, want -42", v.Value)
	}
}

// TestBridgeAnalogOutputG41V2RejectsWrongTypeAndRange asserts a G41V2 write with
// a non-integer value, or an int32 outside the int16 range, is rejected before
// reaching the handler (no silent narrowing/wrap).
func TestBridgeAnalogOutputG41V2RejectsWrongTypeAndRange(t *testing.T) {
	cases := []struct {
		name  string
		value interface{}
	}{
		{"float64", float64(1.5)},
		{"int32 overflow", int32(32768)},
		{"int32 underflow", int32(-32769)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &aoRecordingHandler{}
			d := &internalDataHandler{commandHandler: h}
			if err := d.WriteAnalogOutput(0, tc.value, 2); err == nil {
				t.Fatalf("expected rejection for G41V2 value %v", tc.value)
			}
			if h.got != nil {
				t.Fatal("handler invoked for a rejected G41V2 value")
			}
		})
	}
}

// TestBridgeAnalogOutputG41V4Semantics locks the AO-003 bridge contract: a G41V4
// analog output is carried as a typed AnalogFloat64CommandValue with DirectOperate
// semantics and its exact bits (including signed zero) preserved.
func TestBridgeAnalogOutputG41V4Semantics(t *testing.T) {
	h := &aoRecordingHandler{}
	d := &internalDataHandler{commandHandler: h}
	if err := d.WriteAnalogOutput(9, math.Copysign(0, -1), 4); err != nil {
		t.Fatalf("WriteAnalogOutput G41V4: %v", err)
	}
	if h.got == nil {
		t.Fatal("handler not called")
	}
	if h.got.Group != 41 || h.got.Variation != 4 {
		t.Fatalf("group/variation = %d/%d, want 41/4", h.got.Group, h.got.Variation)
	}
	if h.got.CommandType != types.DirectOperate {
		t.Fatalf("CommandType = %v, want DirectOperate", h.got.CommandType)
	}
	if h.got.Index != 9 {
		t.Fatalf("index = %d, want 9", h.got.Index)
	}
	v, ok := h.got.Value.(*types.AnalogFloat64CommandValue)
	if !ok {
		t.Fatalf("value type = %T, want *types.AnalogFloat64CommandValue", h.got.Value)
	}
	if math.Float64bits(v.Value) != math.Float64bits(math.Copysign(0, -1)) {
		t.Fatalf("value bits = %016X, want %016X (sign bit lost)", math.Float64bits(v.Value), math.Float64bits(math.Copysign(0, -1)))
	}
}

// TestBridgeAnalogOutputG41V4RejectsWrongTypeAndNonFinite asserts a G41V4 write
// with a non-float64 value or a NaN/±Inf value is rejected before reaching the
// handler (owner-approved AO-003 float policy; no silent widening).
func TestBridgeAnalogOutputG41V4RejectsWrongTypeAndNonFinite(t *testing.T) {
	cases := []struct {
		name  string
		value interface{}
	}{
		{"float32", float32(1.5)},
		{"int32", int32(1)},
		{"NaN", math.NaN()},
		{"+Inf", math.Inf(1)},
		{"-Inf", math.Inf(-1)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &aoRecordingHandler{}
			d := &internalDataHandler{commandHandler: h}
			if err := d.WriteAnalogOutput(0, tc.value, 4); err == nil {
				t.Fatalf("expected rejection for G41V4 value %v", tc.value)
			}
			if h.got != nil {
				t.Fatal("handler invoked for a rejected G41V4 value")
			}
		})
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
