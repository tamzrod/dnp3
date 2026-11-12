package master

import (
	"encoding/hex"
	"errors"
	"math"
	"testing"

	"dnp3/internal/al"
)

// AO-003 — Group 41 Variation 3 (single-precision IEEE-754 float) Direct
// Operate, master side.
//
// These tests pin the byte-exact G41V3 request encoding and the response
// command-status handling against independent literals (not against the
// outstation's own decoder), per docs/verification/AO-001.md §3/§6.1. The
// owner-approved AO-003 float policy applies: NaN/±Inf are rejected before
// transmission, finite values (including signed zero and subnormals) keep their
// exact IEEE-754 bits, and a response that changes any echoed value bit
// (including the sign of zero) is never success.

// TestBuildControlRequestG41V3Golden locks the G41V3 request object bytes. The
// value is a four-octet little-endian float32 followed by a SUCCESS status
// octet; goldens are the independent IEEE-754 literals (AO-001 §6.1).
func TestBuildControlRequestG41V3Golden(t *testing.T) {
	cases := []struct {
		name  string
		index uint16
		value float32
		want  string
	}{
		{"+1.0 idx0", 0, 1.0, "290328010000000000803f00"},
		{"-0.0 idx0", 0, float32(math.Copysign(0, -1)), "290328010000000000008000"},
		{"smallest positive subnormal idx0", 0, math.Float32frombits(0x00000001), "290328010000000100000000"},
		{"smallest negative subnormal idx0", 0, math.Float32frombits(0x80000001), "290328010000000100008000"},
		{"max finite idx0", 0, math.Float32frombits(0x7F7FFFFF), "29032801000000ffff7f7f00"},
		{"min normal idx0", 0, math.Float32frombits(0x00800000), "290328010000000000800000"},
		{"2300.0 idx0", 0, 2300.0, "2903280100000000c00f4500"},
		{"-0.5 idx0", 0, -0.5, "29032801000000000000bf00"},
		{"+1.0 idx0x1234", 0x1234, 1.0, "290328010034120000803f00"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := &Master{}
			req := m.buildControlRequest(al.FuncDirectOperate, 41, 3, tc.index, tc.value)
			if req == nil {
				t.Fatalf("buildControlRequest returned nil for a valid G41V3 value")
			}
			want, err := hex.DecodeString(tc.want)
			if err != nil {
				t.Fatal(err)
			}
			if string(req.Data) != string(want) {
				t.Fatalf("G41V3 bytes = %X, want %X", req.Data, want)
			}
		})
	}
}

// TestBuildControlRequestG41V3SignedZeroBits verifies +0 and -0 encode with
// distinct IEEE-754 bit patterns (they are numerically equal but differ on the
// wire), so the encoder never collapses them via a numeric comparison.
func TestBuildControlRequestG41V3SignedZeroBits(t *testing.T) {
	m := &Master{}
	pos := m.buildControlRequest(al.FuncDirectOperate, 41, 3, 0, float32(0))
	neg := m.buildControlRequest(al.FuncDirectOperate, 41, 3, 0, float32(math.Copysign(0, -1)))
	if pos == nil || neg == nil {
		t.Fatal("signed zero must be accepted")
	}
	if string(pos.Data) == string(neg.Data) {
		t.Fatalf("+0 and -0 encoded identically (%X); sign bit not preserved", pos.Data)
	}
	if string(pos.Data) != mustHex("290328010000000000000000") {
		t.Fatalf("+0 bytes = %X", pos.Data)
	}
	if string(neg.Data) != mustHex("290328010000000000008000") {
		t.Fatalf("-0 bytes = %X", neg.Data)
	}
}

// TestBuildControlRequestG41V3AcceptsExactFloat64 verifies a float64 source is
// accepted only when it is exactly representable as float32 (the owner-approved
// policy: no silent narrowing/precision loss). An exact float64 encodes
// identically to the equivalent float32.
func TestBuildControlRequestG41V3AcceptsExactFloat64(t *testing.T) {
	m := &Master{}
	for _, v := range []float64{0.5, 1.0, -0.0, 2300.0, 65536.0} {
		req := m.buildControlRequest(al.FuncDirectOperate, 41, 3, 0, v)
		if req == nil {
			t.Fatalf("exact float64 %v must be accepted for G41V3", v)
		}
		wantBits := math.Float32bits(float32(v))
		gotBits := uint32(req.Data[7]) | uint32(req.Data[8])<<8 | uint32(req.Data[9])<<16 | uint32(req.Data[10])<<24
		if gotBits != wantBits {
			t.Fatalf("float64 %v encoded bits = %08X, want %08X", v, gotBits, wantBits)
		}
	}
}

// TestBuildControlRequestG41V3Rejections verifies the master rejects, before
// sending (nil request = definite pre-send rejection), non-finite floats,
// float64 sources that lose precision on narrowing or overflow/underflow, and
// non-float types, and still rejects an unimplemented variation (V5).
func TestBuildControlRequestG41V3Rejections(t *testing.T) {
	m := &Master{}
	nan32 := float32(math.NaN())
	inf32 := float32(math.Inf(1))
	ninf32 := float32(math.Inf(-1))
	cases := []struct {
		name      string
		variation uint8
		value     interface{}
	}{
		{"float32 NaN", 3, nan32},
		{"float32 +Inf", 3, inf32},
		{"float32 -Inf", 3, ninf32},
		{"float64 NaN", 3, math.NaN()},
		{"float64 +Inf", 3, math.Inf(1)},
		{"float64 inexact 0.1", 3, 0.1},
		{"float64 overflow 1e40", 3, 1e40},
		{"float64 subnormal underflow 1e-46", 3, 1e-46},
		{"int32 value", 3, int32(1)},
		{"int16 value", 3, int16(1)},
		{"int value", 3, int(1)},
		{"string value", 3, "1.0"},
		{"bool value", 3, true},
		{"nil value", 3, nil},
		{"variation 5 (unsupported)", 5, float32(1)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if req := m.buildControlRequest(al.FuncDirectOperate, 41, tc.variation, 0, tc.value); req != nil {
				t.Fatalf("expected rejection (nil request), got %X", req.Data)
			}
		})
	}
}

// TestBuildControlRequestG41V1V2Unchanged confirms the G41V3 addition did not
// alter the verified G41V1 (int32) or G41V2 (int16) encodings.
func TestBuildControlRequestG41V1V2Unchanged(t *testing.T) {
	m := &Master{}
	v1 := m.buildControlRequest(al.FuncDirectOperate, 41, 1, 0, int32(2300))
	if v1 == nil || string(v1.Data) != mustHex("29012801000000fc08000000") {
		t.Fatalf("G41V1 bytes changed: %X", reqDataOrNil(v1))
	}
	v2 := m.buildControlRequest(al.FuncDirectOperate, 41, 2, 0, int16(42))
	if v2 == nil || string(v2.Data) != mustHex("290228010000002a0000") {
		t.Fatalf("G41V2 bytes changed: %X", reqDataOrNil(v2))
	}
}

// TestOperateG41V3StatusMatrix drives the internal DirectOperate path against a
// canned G41V3 echo. A SUCCESS status is success; any explicit non-success
// status is surfaced as-is; a missing object (IIN-only) is never success and
// fails closed with ErrControlOutcomeUnknown.
func TestOperateG41V3StatusMatrix(t *testing.T) {
	cases := []struct {
		name       string
		object     []byte
		want       CommandStatus
		wantErr    bool
		quarantine bool
	}{
		{"success", g41v3Record(0x28, 1, 0, math.Float32bits(1.0), 0x00), CommandStatusSuccess, false, false},
		{"rejected_not_supported", g41v3Record(0x28, 1, 0, math.Float32bits(1.0), 0x04), CommandStatusNotSupported, false, false},
		{"rejected_hardware_error", g41v3Record(0x28, 1, 0, math.Float32bits(1.0), 0x06), CommandStatusHardwareError, false, false},
		{"rejected_no_select", g41v3Record(0x28, 1, 0, math.Float32bits(1.0), 0x02), CommandStatusNoSelect, false, false},
		{"iin_only_no_object", nil, CommandStatusUnknown, true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newG41V2Master(&g41v2RawTransport{object: tc.object})
			status, err := m.OperateWithStatus(1, false, 41, 3, 0, float32(1.0))
			if tc.wantErr {
				if !errors.Is(err, ErrControlOutcomeUnknown) {
					t.Fatalf("error = %v, want ErrControlOutcomeUnknown", err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if status != tc.want {
				t.Fatalf("status = %d, want %d", status, tc.want)
			}
			if got := m.IsQuarantined(1); got != tc.quarantine {
				t.Fatalf("quarantined = %v, want %v", got, tc.quarantine)
			}
		})
	}
}

// TestOperateG41V3SignedZeroEcho verifies the echoed setpoint is compared by raw
// IEEE-754 bits: a matching -0.0 echo is success, while an echo that flips the
// sign bit (+0.0 for a -0.0 request) is never success and is quarantined.
func TestOperateG41V3SignedZeroEcho(t *testing.T) {
	negZero := math.Float32bits(float32(math.Copysign(0, -1)))
	posZero := math.Float32bits(float32(0))

	t.Run("matching -0.0 echo is success", func(t *testing.T) {
		m := newG41V2Master(&g41v2RawTransport{object: g41v3Record(0x28, 1, 0, negZero, 0x00)})
		status, err := m.OperateWithStatus(1, false, 41, 3, 0, float32(math.Copysign(0, -1)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status != CommandStatusSuccess {
			t.Fatalf("status = %d, want Success", status)
		}
		if m.IsQuarantined(1) {
			t.Fatalf("a matching echo must not quarantine")
		}
	})

	t.Run("sign-flipped +0.0 echo is unknown and quarantined", func(t *testing.T) {
		m := newG41V2Master(&g41v2RawTransport{object: g41v3Record(0x28, 1, 0, posZero, 0x00)})
		status, err := m.OperateWithStatus(1, false, 41, 3, 0, float32(math.Copysign(0, -1)))
		if status == CommandStatusSuccess {
			t.Fatalf("sign-flipped echo surfaced as success")
		}
		if !errors.Is(err, ErrControlOutcomeUnknown) {
			t.Fatalf("error = %v, want ErrControlOutcomeUnknown", err)
		}
		if !m.IsQuarantined(1) {
			t.Fatalf("a value-bit mismatch must quarantine the outstation")
		}
	})
}

// TestOperateG41V3MismatchedObjectNotSuccess verifies the master accepts a G41V3
// response only when it is structurally exactly one qualifier-0x28 record with
// count 1 whose group/variation/index/value bits all match the request. Any
// deviation is never success and fails closed (UNKNOWN + ErrControlOutcomeUnknown).
func TestOperateG41V3MismatchedObjectNotSuccess(t *testing.T) {
	good := g41v3Record(0x28, 1, 0, math.Float32bits(1.0), 0x00)
	cases := []struct {
		name   string
		object []byte
	}{
		{"missing record after header", []byte{0x29, 0x03, 0x28, 0x01, 0x00, 0x00, 0x00}},
		{"truncated status", good[:len(good)-1]},
		{"truncated value", good[:len(good)-2]},
		{"surplus byte", append(append([]byte{}, good...), 0xEE)},
		{"count zero", g41v3Record(0x28, 0, 0, math.Float32bits(1.0), 0x00)},
		{"count two", g41v3Record(0x28, 2, 0, math.Float32bits(1.0), 0x00)},
		{"two records", append(append([]byte{}, good...), good...)},
		{"bad qualifier", g41v3Record(0x00, 1, 0, math.Float32bits(1.0), 0x00)},
		{"wrong group", []byte{12, 3, 0x28, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80, 0x3F, 0x00}},
		{"wrong index", g41v3Record(0x28, 1, 9, math.Float32bits(1.0), 0x00)},
		{"wrong value (single low bit)", g41v3Record(0x28, 1, 0, math.Float32bits(1.0)^1, 0x00)},
		{"unsupported variation (V2)", []byte{41, 2, 0x28, 0x01, 0x00, 0x00, 0x00, 0x2A, 0x00, 0x00}},
		{"unsupported variation (V5)", []byte{41, 5, 0x28, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF0, 0x3F, 0x00}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newG41V2Master(&g41v2RawTransport{object: tc.object})
			status, err := m.OperateWithStatus(1, false, 41, 3, 0, float32(1.0))
			if status == CommandStatusSuccess {
				t.Fatalf("malformed/mismatched G41V3 object surfaced as success")
			}
			if !errors.Is(err, ErrControlOutcomeUnknown) {
				t.Fatalf("error = %v, want ErrControlOutcomeUnknown", err)
			}
			if !m.IsQuarantined(1) {
				t.Fatalf("malformed/mismatched G41V3 response must quarantine the outstation")
			}
		})
	}
}

// TestOperateG41V3ConflictingIIN verifies a structurally valid G41V3 SUCCESS
// echo is reconciled against the response IIN: an error-indicating IIN overrides
// the SUCCESS octet with the implied failure status, an unclassifiable error IIN
// (BufferOverflow) fails closed as UNKNOWN + quarantine, and a clean or
// non-error IIN keeps SUCCESS.
func TestOperateG41V3ConflictingIIN(t *testing.T) {
	successEcho := g41v3Record(0x28, 1, 0, math.Float32bits(1.0), 0x00)
	cases := []struct {
		name        string
		iin         [2]byte
		want        CommandStatus
		wantErr     bool
		wantQuarant bool
	}{
		{"clear IIN + SUCCESS echo", [2]byte{0x00, 0x00}, CommandStatusSuccess, false, false},
		{"non-error IIN (Class1Events)", [2]byte{0x40, 0x00}, CommandStatusSuccess, false, false},
		{"ParameterError contradicts SUCCESS", [2]byte{0x00, 0x20}, CommandStatusBadFormat, false, false},
		{"ObjectUnknown contradicts SUCCESS", [2]byte{0x00, 0x40}, CommandStatusNotSupported, false, false},
		{"LocalControl contradicts SUCCESS", [2]byte{0x04, 0x00}, CommandStatusLocal, false, false},
		{"BufferOverflow ambiguity fails closed", [2]byte{0x00, 0x10}, CommandStatusUnknown, true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newG41V2Master(&g41v2RawTransport{object: successEcho, iin: tc.iin})
			status, err := m.OperateWithStatus(1, false, 41, 3, 0, float32(1.0))
			if tc.wantErr {
				if !errors.Is(err, ErrControlOutcomeUnknown) {
					t.Fatalf("error = %v, want ErrControlOutcomeUnknown", err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if status != tc.want {
				t.Fatalf("status = %d, want %d", status, tc.want)
			}
			if got := m.IsQuarantined(1); got != tc.wantQuarant {
				t.Fatalf("quarantined = %v, want %v", got, tc.wantQuarant)
			}
		})
	}
}

// g41v3Record builds one G41V3 record: header(5) + index(2 LE) + float32
// bits(4 LE) + status(1). The object bytes are supplied verbatim to a canned
// transport so response structure can be exercised exactly.
func g41v3Record(qualifier uint8, count uint16, index uint16, valueBits uint32, status byte) []byte {
	return []byte{
		41, 3, qualifier,
		byte(count), byte(count >> 8),
		byte(index), byte(index >> 8),
		byte(valueBits), byte(valueBits >> 8), byte(valueBits >> 16), byte(valueBits >> 24),
		status,
	}
}

func mustHex(s string) string {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func reqDataOrNil(req *al.APDU) []byte {
	if req == nil {
		return nil
	}
	return req.Data
}
