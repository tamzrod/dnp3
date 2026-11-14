package al

import "testing"

// TestUnsolicitedFunctionCodesCanonical pins the IEEE 1815 wire values that the
// pinned OpenDNP3 3.1.2 FunctionCode.h uses: 0x14 is ENABLE_UNSOLICITED and 0x15
// is DISABLE_UNSOLICITED. The removed legacy "get identifier" symbol must not
// reappear as a second meaning for 0x15.
func TestUnsolicitedFunctionCodesCanonical(t *testing.T) {
	if FuncEnableUnsolicited != 0x14 {
		t.Fatalf("FuncEnableUnsolicited = 0x%02X, want 0x14", FuncEnableUnsolicited)
	}
	if FuncDisableUnsolicited != 0x15 {
		t.Fatalf("FuncDisableUnsolicited = 0x%02X, want 0x15", FuncDisableUnsolicited)
	}
	if FuncEnableUnsolicited == FuncDisableUnsolicited {
		t.Fatal("enable and disable share a wire value")
	}
}

// TestFunctionCodeNameUnsolicitedMapping pins the canonical names so 0x14 always
// reads ENABLE_UNSOLICITED and 0x15 always reads DISABLE_UNSOLICITED.
func TestFunctionCodeNameUnsolicitedMapping(t *testing.T) {
	if got := FunctionCodeName(0x14); got != "ENABLE_UNSOLICITED" {
		t.Fatalf("FunctionCodeName(0x14) = %q, want ENABLE_UNSOLICITED", got)
	}
	if got := FunctionCodeName(0x15); got != "DISABLE_UNSOLICITED" {
		t.Fatalf("FunctionCodeName(0x15) = %q, want DISABLE_UNSOLICITED", got)
	}
}

// TestUnsolicitedFunctionCodesAreConfiguration pins that both unsolicited codes
// categorize as configuration, and that the only request functions on 0x14/0x15
// are enable/disable.
func TestUnsolicitedFunctionCodesAreConfiguration(t *testing.T) {
	for _, code := range []uint8{FuncEnableUnsolicited, FuncDisableUnsolicited} {
		if got := Category(code); got != CategoryConfiguration {
			t.Fatalf("Category(0x%02X) = %v, want CategoryConfiguration", code, got)
		}
	}
}
