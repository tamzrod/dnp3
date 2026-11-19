package types

import "testing"

func TestNewPulseControlUsesG12V1(t *testing.T) {
	cmd := NewPulseControl(3, 1000, 500, DirectOperate)
	if cmd.Group != 12 || cmd.Variation != 1 {
		t.Fatalf("NewPulseControl object = G%dV%d, want G12V1", cmd.Group, cmd.Variation)
	}
	if cmd.CROB == nil {
		t.Fatal("NewPulseControl did not carry a full CROB value")
	}
	c := cmd.CROB
	if c.Code != CROBOpPulseOn {
		t.Errorf("operation code = %d, want PulseOn (%d)", c.Code, CROBOpPulseOn)
	}
	if c.Count != 1 || c.OnTime != 1000 || c.OffTime != 500 {
		t.Errorf("crob = count=%d on=%d off=%d, want 1/1000/500", c.Count, c.OnTime, c.OffTime)
	}
	if cmd.Count != 1 || cmd.OnTime != 1000 || cmd.OffTime != 500 {
		t.Errorf("legacy fields = count=%d on=%d off=%d, want 1/1000/500", cmd.Count, cmd.OnTime, cmd.OffTime)
	}
	if cmd.Index != 3 || cmd.CommandType != DirectOperate {
		t.Errorf("index/type = %d/%d, want 3/%d", cmd.Index, cmd.CommandType, DirectOperate)
	}
}

func TestNewCROBControlPreservesAllOperationFields(t *testing.T) {
	cmd := NewCROBControl(2, CROBOpPulseOff, 4, 250, 750, DirectOperate)
	if cmd.CROB == nil {
		t.Fatal("NewCROBControl did not carry a CROB value")
	}
	c := cmd.CROB
	if c.Code != CROBOpPulseOff || c.Count != 4 || c.OnTime != 250 || c.OffTime != 750 {
		t.Errorf("crob = %+v, want PulseOff/4/250/750", c)
	}
	// The collapsed carrier is retained for legacy latch-only handlers only.
	if bc, ok := cmd.Value.(*BinaryCommandValue); !ok || bc.Value {
		t.Errorf("collapsed value = %#v, want BinaryCommandValue{false} for PulseOff", cmd.Value)
	}
}

func TestIsSupportedCROBOp(t *testing.T) {
	supported := []uint8{CROBOpNUL, CROBOpPulseOn, CROBOpPulseOff, CROBOpLatchOn, CROBOpLatchOff}
	for _, op := range supported {
		if !IsSupportedCROBOp(op) {
			t.Errorf("op 0x%02X reported unsupported", op)
		}
	}
	for _, op := range []uint8{0x05, 0x06, 0x08, 0x10, 0x43, 0xFF} {
		if IsSupportedCROBOp(op) {
			t.Errorf("op 0x%02X reported supported", op)
		}
	}
}
