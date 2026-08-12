package frame

import (
	"strings"
	"testing"

	"dnp3/internal/dll/crc"
)

// buildFrameBytes builds a minimal valid DNP3 link frame with the given data
// payload, used as a base for truncation/mismatch tests.
func buildFrameBytes(t *testing.T, data []byte) []byte {
	t.Helper()
	f := &Frame{
		Control:  Control{DIR: false, PRM: false, FuncCode: FuncConfirmedUserDataR},
		DestAddr: 1,
		SrcAddr:  2,
		Data:     data,
	}
	raw, err := Encode(f)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	return raw
}

// TestDecodeRejectsTruncated asserts that frames shorter than the minimum frame
// size and frames whose bytes are fewer than the claimed length are rejected
// with an error and never panic (DNP3-027).
func TestDecodeRejectsTruncated(t *testing.T) {
	t.Run("below_min", func(t *testing.T) {
		// Far too short to be a frame.
		if _, err := Decode([]byte{0x05, 0x64}); err == nil {
			t.Fatal("expected error for frame below minimum size")
		}
	})

	t.Run("truncated_payload", func(t *testing.T) {
		good := buildFrameBytes(t, []byte{0x01, 0x02, 0x03})
		// Slice off trailing bytes so the claimed length exceeds remaining data.
		trunc := good[:len(good)-2]
		_, err := Decode(trunc)
		if err == nil {
			t.Fatal("expected error for truncated frame")
		}
		if !strings.Contains(err.Error(), "short") {
			t.Fatalf("expected 'too short' error, got: %v", err)
		}
	})

	t.Run("no_panic_on_garbage", func(t *testing.T) {
		// Decoding arbitrary short garbage must not panic.
		_, _ = Decode([]byte{0x05, 0x64, 0xFF, 0x00})
		_, _ = Decode([]byte{0x05, 0x64, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	})
}

// TestDecodeRejectsClaimedLengthMismatch asserts a frame whose claimed length
// does not match the actual content (trailing bytes) is detected. We craft a
// frame that claims a larger payload than present.
func TestDecodeRejectsClaimedLengthMismatch(t *testing.T) {
	good := buildFrameBytes(t, []byte{0x01, 0x02, 0x03})
	// Tamper the length byte (offset 2) to claim a larger payload than present.
	bad := make([]byte, len(good))
	copy(bad, good)
	// length field: keep >=5 (min) but larger than the actual data allows.
	bad[2] = 20 // claims dataLen = 15, but the buffer is much smaller
	_, err := Decode(bad)
	if err == nil {
		t.Fatal("expected error for claimed-length mismatch")
	}
	if !strings.Contains(err.Error(), "short") {
		t.Fatalf("expected 'too short' error, got: %v", err)
	}
}

// TestDecodeRejectsOversize asserts the oversize guard is present and that the
// maximum legal frame (250 data bytes == MaxDataSize) decodes cleanly, while the
// MaxDataSize constant is the spec value (DNP3-027).
func TestDecodeRejectsOversize(t *testing.T) {
	// Build a max-size frame (250 data bytes) using the package's own CRC.
	const dataLen = MaxDataSize // 250
	data := make([]byte, dataLen) // zero payload
	// Header prefix: sync(2) + length(1) + control(1) + dest(2) + src(2) = 8.
	hdr := []byte{
		0x05, 0x64,                  // sync
		byte(5 + dataLen),           // length = 5 + 250 = 255
		0x44,                        // control: DIR=0,PRM=0,Func=4
		0x01, 0x00,                  // dest = 1
		0x02, 0x00,                  // src = 2
	}
	hdrCRC := crc.CRC16(hdr)
	buf := make([]byte, 0, EncodedSize(dataLen))
	buf = append(buf, hdr...)
	buf = append(buf, byte(hdrCRC), byte(hdrCRC>>8))
	// Append data blocks (16 bytes each) followed by their per-block CRC.
	for i := 0; i < dataLen; i += 16 {
		end := i + 16
		if end > dataLen {
			end = dataLen
		}
		block := data[i:end]
		buf = append(buf, block...)
		c := crc.CRC16(block)
		buf = append(buf, byte(c), byte(c>>8))
	}
	if _, err := Decode(buf); err != nil {
		t.Fatalf("max-size frame (250 data) should decode, got: %v", err)
	}

	if MaxDataSize != 250 {
		t.Fatalf("MaxDataSize = %d, want 250", MaxDataSize)
	}
}

