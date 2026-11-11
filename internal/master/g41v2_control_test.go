package master

import (
	"encoding/hex"
	"errors"
	"testing"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/tl"
)

// AO-003 — Group 41 Variation 2 (signed int16) Direct Operate, master side.
//
// These tests pin the byte-exact G41V2 request encoding and the response
// command-status handling against independent literals (not against the
// outstation's own decoder), per docs/verification/AO-001.md §3/§6.1.

// TestBuildControlRequestG41V2Golden locks the G41V2 request object bytes. The
// independent golden from AO-001 §6.1 for int16 42 at index 0 is
// 29 02 28 01 00 00 00 2A 00 00 (header + record); the int16 is little-endian
// and a trailing SUCCESS status octet is present.
func TestBuildControlRequestG41V2Golden(t *testing.T) {
	cases := []struct {
		name  string
		index uint16
		value int16
		want  string
	}{
		{"42 idx0", 0, 42, "290228010000002a0000"},
		{"min int16 idx0", 0, -32768, "29022801000000008000"},
		{"minus1 idx0", 0, -1, "29022801000000ffff00"},
		{"max int16 idx0x1234", 0x1234, 32767, "29022801003412ff7f00"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := &Master{}
			req := m.buildControlRequest(al.FuncDirectOperate, 41, 2, tc.index, tc.value)
			if req == nil {
				t.Fatalf("buildControlRequest returned nil for a valid G41V2 value")
			}
			want, err := hex.DecodeString(tc.want)
			if err != nil {
				t.Fatal(err)
			}
			if string(req.Data) != string(want) {
				t.Fatalf("G41V2 bytes = %X, want %X", req.Data, want)
			}
		})
	}
}

// TestBuildControlRequestG41V2AcceptsInRangeInt32 verifies an in-range int32 is
// accepted and encoded identically to the equivalent int16 (no narrowing is
// needed because the value is exactly representable).
func TestBuildControlRequestG41V2AcceptsInRangeInt32(t *testing.T) {
	m := &Master{}
	cases := []struct {
		name  string
		value int32
		want  string
	}{
		{"42", 42, "290228010000002a0000"},
		{"min int16", -32768, "29022801000000008000"},
		{"max int16", 32767, "29022801000000ff7f00"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := m.buildControlRequest(al.FuncDirectOperate, 41, 2, 0, tc.value)
			if req == nil {
				t.Fatalf("buildControlRequest returned nil for an in-range int32 G41V2 value")
			}
			want, _ := hex.DecodeString(tc.want)
			if string(req.Data) != string(want) {
				t.Fatalf("G41V2 int32 bytes = %X, want %X", req.Data, want)
			}
		})
	}
}

// TestBuildControlRequestG41V2Rejections verifies the master rejects unsupported
// G41 variations, out-of-range integers, non-integral floats, nil and wrong
// types before sending (nil request = definite pre-send rejection). A float
// source is rejected even when integral: the typed G41V2 path never silently
// truncates or rounds.
func TestBuildControlRequestG41V2Rejections(t *testing.T) {
	m := &Master{}
	cases := []struct {
		name      string
		variation uint8
		value     interface{}
	}{
		{"variation 3", 3, int16(1)},
		{"variation 4", 4, int16(1)},
		{"int32 overflow", 2, int32(32768)},
		{"int32 underflow", 2, int32(-32769)},
		{"int64 overflow", 2, int64(32768)},
		{"int64 underflow", 2, int64(-32769)},
		{"int overflow", 2, int(32768)},
		{"int underflow", 2, int(-32769)},
		{"nonintegral float64", 2, float64(1.5)},
		{"integral float64", 2, float64(42)},
		{"float32", 2, float32(42)},
		{"nil value", 2, nil},
		{"string value", 2, "42"},
		{"bool value", 2, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if req := m.buildControlRequest(al.FuncDirectOperate, 41, tc.variation, 0, tc.value); req != nil {
				t.Fatalf("expected rejection (nil request), got %X", req.Data)
			}
		})
	}
}

// TestBuildControlRequestG41V1Unchanged confirms the AO-003 G41V2 addition did
// not alter the verified G41V1 encoding: int32 2300 at index 0 is still
// 29 01 28 01 00 00 00 FC 08 00 00 00.
func TestBuildControlRequestG41V1Unchanged(t *testing.T) {
	m := &Master{}
	req := m.buildControlRequest(al.FuncDirectOperate, 41, 1, 0, int32(2300))
	if req == nil {
		t.Fatal("G41V1 request must still build")
	}
	want, _ := hex.DecodeString("29012801000000fc08000000")
	if string(req.Data) != string(want) {
		t.Fatalf("G41V1 bytes = %X, want %X", req.Data, want)
	}
}

// g41v2Record builds one G41V2 record: header(5) + index(2 LE) + int16(2 LE) +
// status(1). The object bytes are supplied verbatim to a canned transport so
// response structure can be exercised exactly.
func g41v2Record(qualifier uint8, count uint16, index uint16, value int16, status byte) []byte {
	return []byte{
		41, 2, qualifier,
		byte(count), byte(count >> 8),
		byte(index), byte(index >> 8),
		byte(value), byte(value >> 8),
		status,
	}
}

// g41v2RawTransport returns a DirectOperate response whose object payload (the
// bytes after the 2-byte IIN) and IIN bytes are supplied verbatim. It never
// contacts a real outstation.
type g41v2RawTransport struct {
	lastSeq uint8
	object  []byte
	iin     [2]byte
}

func (t *g41v2RawTransport) Send(data []byte) error {
	t.lastSeq = extractRequestSeqForBuild(data)
	return nil
}

func (t *g41v2RawTransport) SetTimeout(ms int) {}

func (t *g41v2RawTransport) Receive() ([]byte, error) {
	apdu := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: t.lastSeq},
		FuncCode: al.FuncResponse,
		Data:     append([]byte{t.iin[0], t.iin[1]}, t.object...),
	}
	frag := tl.Fragment{FIR: true, FIN: true, Data: apdu.Encode()}
	dllFrame := &frame.Frame{
		Control:  frame.Control{DIR: false, PRM: false, FuncCode: frame.FuncConfirmedUserDataR},
		DestAddr: 1, SrcAddr: 2, Data: tl.EncodeFragment(frag),
	}
	raw, _ := frame.Encode(dllFrame)
	return raw, nil
}

func newG41V2Master(t *g41v2RawTransport) *Master {
	m := NewMaster(DefaultConfig())
	m.SetTransport(t)
	m.SetState(StateInitialized)
	return m
}

// TestOperateG41V2StatusMatrix drives the internal DirectOperate path against a
// canned G41V2 echo. A SUCCESS status is success; any explicit non-success
// status is surfaced as-is; a missing object (IIN-only) is never success and
// fails closed with ErrControlOutcomeUnknown.
func TestOperateG41V2StatusMatrix(t *testing.T) {
	cases := []struct {
		name       string
		object     []byte
		want       CommandStatus
		wantErr    bool
		quarantine bool
	}{
		{"success", g41v2Record(0x28, 1, 0, 42, 0x00), CommandStatusSuccess, false, false},
		{"rejected_not_supported", g41v2Record(0x28, 1, 0, 42, 0x04), CommandStatusNotSupported, false, false},
		{"rejected_blocked", g41v2Record(0x28, 1, 0, 42, 0x06), CommandStatusBlocked, false, false},
		{"rejected_no_select", g41v2Record(0x28, 1, 0, 42, 0x02), CommandStatusNoSelect, false, false},
		{"iin_only_no_object", nil, CommandStatusUnknown, true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newG41V2Master(&g41v2RawTransport{object: tc.object})
			status, err := m.OperateWithStatus(1, false, 41, 2, 0, int16(42))
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
			if status == CommandStatusSuccess && tc.want != CommandStatusSuccess {
				t.Fatalf("non-success outcome surfaced as success")
			}
			if got := m.IsQuarantined(1); got != tc.quarantine {
				t.Fatalf("quarantined = %v, want %v", got, tc.quarantine)
			}
		})
	}
}

// TestOperateG41V2MismatchedObjectNotSuccess verifies the master accepts a G41V2
// response only when it is structurally exactly one qualifier-0x28 record with
// count 1 whose group/variation/index/value all match the request. Any
// deviation is never success and fails closed (UNKNOWN + ErrControlOutcomeUnknown).
func TestOperateG41V2MismatchedObjectNotSuccess(t *testing.T) {
	good := g41v2Record(0x28, 1, 0, 42, 0x00)
	cases := []struct {
		name   string
		object []byte
	}{
		{"missing record after header", []byte{0x29, 0x02, 0x28, 0x01, 0x00, 0x00, 0x00}},
		{"truncated status", good[:len(good)-1]},
		{"truncated value", good[:len(good)-2]},
		{"surplus byte", append(append([]byte{}, good...), 0xEE)},
		{"count zero", g41v2Record(0x28, 0, 0, 42, 0x00)},
		{"count two", g41v2Record(0x28, 2, 0, 42, 0x00)},
		{"two records", append(append([]byte{}, good...), good...)},
		{"bad qualifier", g41v2Record(0x00, 1, 0, 42, 0x00)},
		{"wrong group", []byte{12, 2, 0x28, 0x01, 0x00, 0x00, 0x00, 42, 0x00, 0x00}},
		{"wrong index", g41v2Record(0x28, 1, 9, 42, 0x00)},
		{"wrong value", g41v2Record(0x28, 1, 0, 43, 0x00)},
		{"unsupported variation", []byte{41, 3, 0x28, 0x01, 0x00, 0x00, 0x00, 42, 0x00, 0x00}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newG41V2Master(&g41v2RawTransport{object: tc.object})
			status, err := m.OperateWithStatus(1, false, 41, 2, 0, int16(42))
			if status == CommandStatusSuccess {
				t.Fatalf("malformed/mismatched G41V2 object surfaced as success")
			}
			if !errors.Is(err, ErrControlOutcomeUnknown) {
				t.Fatalf("error = %v, want ErrControlOutcomeUnknown", err)
			}
			if !m.IsQuarantined(1) {
				t.Fatalf("malformed/mismatched G41V2 response must quarantine the outstation")
			}
		})
	}
}

// TestOperateG41V2ConflictingIIN verifies a structurally valid G41V2 SUCCESS
// echo is reconciled against the response IIN: an error-indicating IIN overrides
// the SUCCESS octet with the implied failure status, an unclassifiable error IIN
// (BufferOverflow) fails closed as UNKNOWN + quarantine, and a clean or
// non-error IIN keeps SUCCESS.
func TestOperateG41V2ConflictingIIN(t *testing.T) {
	successEcho := g41v2Record(0x28, 1, 0, 42, 0x00)
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
			status, err := m.OperateWithStatus(1, false, 41, 2, 0, int16(42))
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

// TestParseG41RecordG41V2 confirms parseG41Record reads the G41V2 int16 value
// little-endian (returned as the raw uint16 payload) and the per-point status
// octet, and rejects an unsupported variation (V3) or a truncated record.
func TestParseG41RecordG41V2(t *testing.T) {
	// 42 -> value bytes 2A 00, status 0x06 (Blocked).
	rec := g41v2Record(0x28, 1, 0x0004, 42, 0x06)
	idx, value, status, ok := parseG41Record(rec, 0, 2)
	if !ok {
		t.Fatalf("parseG41Record rejected a valid G41V2 record")
	}
	if idx != 4 {
		t.Fatalf("index = %d, want 4", idx)
	}
	if value != 42 {
		t.Fatalf("value = %d, want 42", value)
	}
	if status != CommandStatusBlocked {
		t.Fatalf("status = %d, want Blocked (6)", status)
	}

	// Negative boundary: -32768 -> 0x8000 -> 32768 as the raw uint16 payload.
	neg := g41v2Record(0x28, 1, 0, -32768, 0x00)
	if _, v, _, ok := parseG41Record(neg, 0, 2); !ok || v != 0x8000 {
		t.Fatalf("parseG41Record(-32768) = (%d, ok=%v), want (32768, true)", v, ok)
	}

	// Unsupported variation V3.
	if _, _, _, ok := parseG41Record(g41v2Record(0x28, 1, 0, 42, 0x00), 0, 3); ok {
		t.Fatalf("parseG41Record accepted unsupported variation 3")
	}

	// Truncated (status octet missing).
	if _, _, _, ok := parseG41Record(rec[:len(rec)-1], 0, 2); ok {
		t.Fatalf("parseG41Record accepted a truncated G41V2 record")
	}
}
