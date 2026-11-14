package master

import (
	"encoding/hex"
	"errors"
	"math"
	"testing"

	"dnp3/internal/al"
)

// AO-003 — Group 41 Variation 4 (double-precision IEEE-754 float) Direct
// Operate, master side.
//
// These tests pin the byte-exact G41V4 request encoding and the response
// command-status handling against independent literals (not against the
// outstation's own decoder), per docs/verification/AO-001.md §3.1/§6.1. The
// owner-approved AO-003 float policy applies: NaN/±Inf are rejected before
// transmission, finite values (including signed zero and subnormals) keep their
// exact IEEE-754 bits, and a response that changes any echoed value bit
// (including the sign of zero) is never success.

// TestBuildControlRequestG41V4Golden locks the G41V4 request object bytes. The
// value is an eight-octet little-endian float64 followed by a SUCCESS status
// octet; goldens are the independent IEEE-754 literals (AO-001 §6.1).
func TestBuildControlRequestG41V4Golden(t *testing.T) {
	cases := []struct {
		name  string
		index uint16
		value float64
		want  string
	}{
		{"+1.0 idx0", 0, 1.0, "29042801000000000000000000f03f00"},
		{"-0.0 idx0", 0, math.Copysign(0, -1), "29042801000000000000000000008000"},
		{"smallest positive subnormal idx0", 0, math.Float64frombits(0x0000000000000001), "29042801000000010000000000000000"},
		{"smallest negative subnormal idx0", 0, math.Float64frombits(0x8000000000000001), "29042801000000010000000000008000"},
		{"max finite idx0", 0, math.Float64frombits(0x7FEFFFFFFFFFFFFF), "29042801000000ffffffffffffef7f00"},
		{"min normal idx0", 0, math.Float64frombits(0x0010000000000000), "29042801000000000000000000100000"},
		{"123.456789 idx0", 0, 123.456789, "290428010000000b0bee073cdd5e4000"},
		{"-0.5 idx0", 0, -0.5, "29042801000000000000000000e0bf00"},
		{"+1.0 idx0x1234", 0x1234, 1.0, "29042801003412000000000000f03f00"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := &Master{}
			req := m.buildControlRequest(al.FuncDirectOperate, 41, 4, tc.index, tc.value)
			if req == nil {
				t.Fatalf("buildControlRequest returned nil for a valid G41V4 value")
			}
			want, err := hex.DecodeString(tc.want)
			if err != nil {
				t.Fatal(err)
			}
			if string(req.Data) != string(want) {
				t.Fatalf("G41V4 bytes = %X, want %X", req.Data, want)
			}
		})
	}
}

// TestBuildControlRequestG41V4SignedZeroBits verifies +0 and -0 encode with
// distinct IEEE-754 bit patterns (they are numerically equal but differ on the
// wire), so the encoder never collapses them via a numeric comparison.
func TestBuildControlRequestG41V4SignedZeroBits(t *testing.T) {
	m := &Master{}
	pos := m.buildControlRequest(al.FuncDirectOperate, 41, 4, 0, float64(0))
	neg := m.buildControlRequest(al.FuncDirectOperate, 41, 4, 0, math.Copysign(0, -1))
	if pos == nil || neg == nil {
		t.Fatal("signed zero must be accepted")
	}
	if string(pos.Data) == string(neg.Data) {
		t.Fatalf("+0 and -0 encoded identically (%X); sign bit not preserved", pos.Data)
	}
	if string(pos.Data) != mustHex("29042801000000000000000000000000") {
		t.Fatalf("+0 bytes = %X", pos.Data)
	}
	if string(neg.Data) != mustHex("29042801000000000000000000008000") {
		t.Fatalf("-0 bytes = %X", neg.Data)
	}
}

// TestBuildControlRequestG41V4Rejections verifies the master rejects, before
// sending (nil request = definite pre-send rejection), non-finite floats and
// non-float64 source types. Unlike G41V3 there is no narrower carrier to widen:
// a float32 source is rejected so an accidental width mismatch at the API
// boundary is not silently promoted. G41V5 remains rejected.
func TestBuildControlRequestG41V4Rejections(t *testing.T) {
	m := &Master{}
	cases := []struct {
		name      string
		variation uint8
		value     interface{}
	}{
		{"float64 NaN", 4, math.NaN()},
		{"float64 +Inf", 4, math.Inf(1)},
		{"float64 -Inf", 4, math.Inf(-1)},
		{"float32 value", 4, float32(1)},
		{"int32 value", 4, int32(1)},
		{"int16 value", 4, int16(1)},
		{"int value", 4, int(1)},
		{"string value", 4, "1.0"},
		{"bool value", 4, true},
		{"nil value", 4, nil},
		{"variation 5 (unsupported)", 5, float64(1)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if req := m.buildControlRequest(al.FuncDirectOperate, 41, tc.variation, 0, tc.value); req != nil {
				t.Fatalf("expected rejection (nil request), got %X", req.Data)
			}
		})
	}
}

// TestBuildControlRequestG41V1V2V3Unchanged confirms the G41V4 addition did not
// alter the verified G41V1 (int32), G41V2 (int16) or G41V3 (float32) encodings.
func TestBuildControlRequestG41V1V2V3Unchanged(t *testing.T) {
	m := &Master{}
	v1 := m.buildControlRequest(al.FuncDirectOperate, 41, 1, 0, int32(2300))
	if v1 == nil || string(v1.Data) != mustHex("29012801000000fc08000000") {
		t.Fatalf("G41V1 bytes changed: %X", reqDataOrNil(v1))
	}
	v2 := m.buildControlRequest(al.FuncDirectOperate, 41, 2, 0, int16(42))
	if v2 == nil || string(v2.Data) != mustHex("290228010000002a0000") {
		t.Fatalf("G41V2 bytes changed: %X", reqDataOrNil(v2))
	}
	v3 := m.buildControlRequest(al.FuncDirectOperate, 41, 3, 0, float32(1.0))
	if v3 == nil || string(v3.Data) != mustHex("290328010000000000803f00") {
		t.Fatalf("G41V3 bytes changed: %X", reqDataOrNil(v3))
	}
}

// TestOperateG41V4StatusMatrix drives the internal DirectOperate path against a
// canned G41V4 echo. A SUCCESS status is success; any explicit non-success
// status is surfaced as-is; a missing object (IIN-only) is never success and
// fails closed with ErrControlOutcomeUnknown.
func TestOperateG41V4StatusMatrix(t *testing.T) {
	cases := []struct {
		name       string
		object     []byte
		want       CommandStatus
		wantErr    bool
		quarantine bool
	}{
		{"success", g41v4Record(0x28, 1, 0, math.Float64bits(1.0), 0x00), CommandStatusSuccess, false, false},
		{"rejected_not_supported", g41v4Record(0x28, 1, 0, math.Float64bits(1.0), 0x04), CommandStatusNotSupported, false, false},
		{"rejected_hardware_error", g41v4Record(0x28, 1, 0, math.Float64bits(1.0), 0x06), CommandStatusHardwareError, false, false},
		{"rejected_no_select", g41v4Record(0x28, 1, 0, math.Float64bits(1.0), 0x02), CommandStatusNoSelect, false, false},
		{"iin_only_no_object", nil, CommandStatusUnknown, true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newG41V2Master(&g41v2RawTransport{object: tc.object})
			status, err := m.OperateWithStatus(1, false, 41, 4, 0, float64(1.0))
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

// TestOperateG41V4SignedZeroEcho verifies the echoed setpoint is compared by raw
// IEEE-754 bits: a matching -0.0 echo is success, while an echo that flips the
// sign bit (+0.0 for a -0.0 request) is never success and is quarantined.
func TestOperateG41V4SignedZeroEcho(t *testing.T) {
	negZero := math.Float64bits(math.Copysign(0, -1))
	posZero := math.Float64bits(0.0)

	t.Run("matching -0.0 echo is success", func(t *testing.T) {
		m := newG41V2Master(&g41v2RawTransport{object: g41v4Record(0x28, 1, 0, negZero, 0x00)})
		status, err := m.OperateWithStatus(1, false, 41, 4, 0, math.Copysign(0, -1))
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
		m := newG41V2Master(&g41v2RawTransport{object: g41v4Record(0x28, 1, 0, posZero, 0x00)})
		status, err := m.OperateWithStatus(1, false, 41, 4, 0, math.Copysign(0, -1))
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

// TestOperateG41V4MismatchedObjectNotSuccess verifies the master accepts a G41V4
// response only when it is structurally exactly one qualifier-0x28 record with
// count 1 whose group/variation/index/value bits all match the request. Any
// deviation is never success and fails closed (UNKNOWN + ErrControlOutcomeUnknown).
func TestOperateG41V4MismatchedObjectNotSuccess(t *testing.T) {
	good := g41v4Record(0x28, 1, 0, math.Float64bits(1.0), 0x00)
	cases := []struct {
		name   string
		object []byte
	}{
		{"missing record after header", []byte{0x29, 0x04, 0x28, 0x01, 0x00, 0x00, 0x00}},
		{"truncated status", good[:len(good)-1]},
		{"truncated value", good[:len(good)-2]},
		{"surplus byte", append(append([]byte{}, good...), 0xEE)},
		{"count zero", g41v4Record(0x28, 0, 0, math.Float64bits(1.0), 0x00)},
		{"count two", g41v4Record(0x28, 2, 0, math.Float64bits(1.0), 0x00)},
		{"two records", append(append([]byte{}, good...), good...)},
		{"bad qualifier", g41v4Record(0x00, 1, 0, math.Float64bits(1.0), 0x00)},
		{"wrong group", []byte{12, 4, 0x28, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF0, 0x3F, 0x00}},
		{"wrong index", g41v4Record(0x28, 1, 9, math.Float64bits(1.0), 0x00)},
		{"wrong value (single low bit)", g41v4Record(0x28, 1, 0, math.Float64bits(1.0)^1, 0x00)},
		{"unsupported variation (V3)", []byte{41, 3, 0x28, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80, 0x3F, 0x00}},
		{"unsupported variation (V5)", []byte{41, 5, 0x28, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF0, 0x3F, 0x00}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newG41V2Master(&g41v2RawTransport{object: tc.object})
			status, err := m.OperateWithStatus(1, false, 41, 4, 0, float64(1.0))
			if status == CommandStatusSuccess {
				t.Fatalf("malformed/mismatched G41V4 object surfaced as success")
			}
			if !errors.Is(err, ErrControlOutcomeUnknown) {
				t.Fatalf("error = %v, want ErrControlOutcomeUnknown", err)
			}
			if !m.IsQuarantined(1) {
				t.Fatalf("malformed/mismatched G41V4 response must quarantine the outstation")
			}
		})
	}
}

// TestOperateG41V4ConflictingIIN verifies a structurally valid G41V4 SUCCESS
// echo is reconciled against the response IIN: an error-indicating IIN overrides
// the SUCCESS octet with the implied failure status, an unclassifiable error IIN
// (BufferOverflow) fails closed as UNKNOWN + quarantine, and a clean or
// non-error IIN keeps SUCCESS.
func TestOperateG41V4ConflictingIIN(t *testing.T) {
	successEcho := g41v4Record(0x28, 1, 0, math.Float64bits(1.0), 0x00)
	cases := []struct {
		name        string
		iin         [2]byte
		want        CommandStatus
		wantErr     bool
		wantQuarant bool
	}{
		{"clear IIN + SUCCESS echo", [2]byte{0x00, 0x00}, CommandStatusSuccess, false, false},
		{"non-error IIN (Class1Events)", [2]byte{0x02, 0x00}, CommandStatusSuccess, false, false},
		{"ParameterError contradicts SUCCESS", [2]byte{0x00, 0x04}, CommandStatusBadFormat, false, false},
		{"ObjectUnknown contradicts SUCCESS", [2]byte{0x00, 0x02}, CommandStatusNotSupported, false, false},
		{"LocalControl contradicts SUCCESS", [2]byte{0x20, 0x00}, CommandStatusLocal, false, false},
		{"BufferOverflow ambiguity fails closed", [2]byte{0x00, 0x08}, CommandStatusUnknown, true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newG41V2Master(&g41v2RawTransport{object: successEcho, iin: tc.iin})
			status, err := m.OperateWithStatus(1, false, 41, 4, 0, float64(1.0))
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

// g41v4Record builds one G41V4 record: header(5) + index(2 LE) + float64
// bits(8 LE) + status(1). The object bytes are supplied verbatim to a canned
// transport so response structure can be exercised exactly.
func g41v4Record(qualifier uint8, count uint16, index uint16, valueBits uint64, status byte) []byte {
	out := []byte{
		41, 4, qualifier,
		byte(count), byte(count >> 8),
		byte(index), byte(index >> 8),
	}
	for i := 0; i < 8; i++ {
		out = append(out, byte(valueBits>>(8*uint(i))))
	}
	return append(out, status)
}
