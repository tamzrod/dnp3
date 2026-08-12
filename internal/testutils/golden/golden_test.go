package golden

import (
	"testing"
)

// TestLoadHex verifies the shared golden loader (DNP3-097) resolves the
// fixture directory and decodes a known fixture without comments/whitespace.
func TestLoadHex(t *testing.T) {
	raw, err := LoadHex("racom-dnp3-link-frame.hex")
	if err != nil {
		t.Fatalf("LoadHex: %v", err)
	}
	if len(raw) == 0 {
		t.Fatal("LoadHex returned empty bytes")
	}
	// The racom link frame begins with the DNP3 start bytes 0x05 0x64.
	if raw[0] != 0x05 || raw[1] != 0x64 {
		t.Fatalf("first two bytes = %02x %02x, want 05 64", raw[0], raw[1])
	}
}

// TestLoadHexMissingFile verifies the shared loader surfaces a clear error for
// an absent fixture rather than panicking.
func TestLoadHexMissingFile(t *testing.T) {
	if _, err := LoadHex("does-not-exist.hex"); err == nil {
		t.Fatal("expected error for missing fixture, got nil")
	}
}
