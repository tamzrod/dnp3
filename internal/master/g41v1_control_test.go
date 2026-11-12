package master

import (
	"encoding/hex"
	"errors"
	"testing"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/tl"
)

// AO-002 — Group 41 Variation 1 (signed int32) Direct Operate, master side.
//
// These tests pin the byte-exact G41V1 request encoding and the response
// command-status handling against independent literals (not against the
// outstation's own decoder), per docs/verification/AO-001.md §3/§6.

// TestBuildControlRequestG41V1Golden locks the G41V1 request object bytes. The
// independent golden from AO-001 §6 for int32 2300 at index 0 is
// 29 01 28 01 00 00 00 FC 08 00 00 00 (header + record, no APDU bytes); the
// int32 is little-endian and a trailing SUCCESS status octet is present.
func TestBuildControlRequestG41V1Golden(t *testing.T) {
	cases := []struct {
		name  string
		index uint16
		value int32
		want  string
	}{
		{"2300 idx0", 0, 2300, "29012801000000fc08000000"},
		{"minus1 idx0", 0, -1, "29012801000000ffffffff00"},
		{"min int32 idx0x1234", 0x1234, -2147483648, "290128010034120000008000"},
		{"max int32 idx0x1234", 0x1234, 2147483647, "29012801003412ffffff7f00"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := &Master{}
			req := m.buildControlRequest(al.FuncDirectOperate, 41, 1, tc.index, tc.value)
			if req == nil {
				t.Fatalf("buildControlRequest returned nil for a valid G41V1 value")
			}
			want, err := hex.DecodeString(tc.want)
			if err != nil {
				t.Fatal(err)
			}
			if string(req.Data) != string(want) {
				t.Fatalf("G41V1 bytes = %X, want %X", req.Data, want)
			}
		})
	}
}

// TestBuildControlRequestG41V1Rejections verifies the master rejects unsupported
// G41 variations and non-exactly-representable values before sending (nil
// request = definite pre-send rejection). AO-003 enables G41V2, so the
// variation-2 rejection case moved to the G41V2 suite; V3/V4 are covered by
// their own suites, so an unimplemented variation (V5) is used here.
func TestBuildControlRequestG41V1Rejections(t *testing.T) {
	m := &Master{}
	cases := []struct {
		name      string
		variation uint8
		value     interface{}
	}{
		{"variation 3", 3, int32(1)},
		{"variation 5", 5, int32(1)},
		{"float64 value", 1, float64(1)},
		{"float32 value", 1, float32(1)},
		{"int64 overflow", 1, int64(2147483648)},
		{"int64 underflow", 1, int64(-2147483649)},
		{"nil value", 1, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if req := m.buildControlRequest(al.FuncDirectOperate, 41, tc.variation, 0, tc.value); req != nil {
				t.Fatalf("expected rejection (nil request), got %X", req.Data)
			}
		})
	}
}

// g41v1EchoTransport returns a DirectOperate response carrying one G41V1 object
// whose per-point command-status octet is commandStatus. Layout after IIN:
//
//	29 01 28 01 00   G41V1, qualifier 0x28, count 1
//	<index LE>       echoed index
//	<value LE>       echoed int32 value
//	<commandStatus>  per-point command status
type g41v1EchoTransport struct {
	lastSeq       uint8
	commandStatus byte
	group         uint8
	variation     uint8
	index         uint16
	value         int32
	// omitObject drops the echoed object entirely (IIN-only response).
	omitObject bool
}

func (t *g41v1EchoTransport) Send(data []byte) error {
	t.lastSeq = extractRequestSeqForBuild(data)
	return nil
}

func (t *g41v1EchoTransport) SetTimeout(ms int) {}

func (t *g41v1EchoTransport) Receive() ([]byte, error) {
	return buildG41V1StatusResponse(t), nil
}

func buildG41V1StatusResponse(t *g41v1EchoTransport) []byte {
	var data []byte
	if !t.omitObject {
		data = []byte{
			0x29, 0x01, 0x28, 0x01, 0x00,
			byte(t.index), byte(t.index >> 8),
			byte(t.value), byte(t.value >> 8), byte(t.value >> 16), byte(t.value >> 24),
			t.commandStatus,
		}
	}
	apdu := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: t.lastSeq},
		FuncCode: al.FuncResponse,
		Data:     append([]byte{0x00, 0x00}, data...), // IIN (all clear) + object
	}
	frag := tl.Fragment{FIR: true, FIN: true, Data: apdu.Encode()}
	tlData := tl.EncodeFragment(frag)
	dllFrame := &frame.Frame{
		Control:  frame.Control{DIR: false, PRM: false, FuncCode: frame.FuncConfirmedUserDataR},
		DestAddr: 1, SrcAddr: 2, Data: tlData,
	}
	raw, _ := frame.Encode(dllFrame)
	return raw
}

// TestOperateG41V1StatusMatrix drives the internal DirectOperate path against a
// canned G41V1 echo transport. A SUCCESS status is success; any non-success
// status is surfaced as-is; a missing object (IIN-only) is never success.
func TestOperateG41V1StatusMatrix(t *testing.T) {
	cases := []struct {
		name          string
		commandStatus byte
		omitObject    bool
		want          CommandStatus
		wantErr       bool
	}{
		{"success", 0, false, CommandStatusSuccess, false},
		{"rejected_not_supported", 4, false, CommandStatusNotSupported, false},
		{"rejected_hardware_error", 6, false, CommandStatusHardwareError, false},
		{"rejected_no_select", 2, false, CommandStatusNoSelect, false},
		{"iim_only_no_object", 0, true, CommandStatusUnknown, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMaster(DefaultConfig())
			m.SetTransport(&g41v1EchoTransport{commandStatus: tc.commandStatus, index: 0, value: 2300, omitObject: tc.omitObject})
			m.SetState(StateInitialized)

			status, err := m.OperateWithStatus(1, false, 41, 1, 0, int32(2300))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected fail-closed error, got nil (status %d)", status)
				}
				if !errors.Is(err, ErrControlOutcomeUnknown) {
					t.Fatalf("error = %v, want ErrControlOutcomeUnknown", err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if status != tc.want {
				t.Fatalf("status = %d, want %d", status, tc.want)
			}
			if tc.commandStatus != 0 && status == CommandStatusSuccess {
				t.Fatalf("non-success byte surfaced as success")
			}
		})
	}
}

// TestOperateG41V1MismatchedObjectNotSuccess verifies the master requires the
// echoed object to match the request group/variation/index: a wrong index, a
// wrong variation, or a truncated object is never success.
func TestOperateG41V1MismatchedObjectNotSuccess(t *testing.T) {
	cases := []struct {
		name      string
		transport *g41v1EchoTransport
	}{
		{"wrong index", &g41v1EchoTransport{commandStatus: 0, index: 0x0009, value: 2300}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMaster(DefaultConfig())
			m.SetTransport(tc.transport)
			m.SetState(StateInitialized)
			status, err := m.OperateWithStatus(1, false, 41, 1, 0, int32(2300))
			if status == CommandStatusSuccess {
				t.Fatalf("mismatched G41V1 object surfaced as success")
			}
			if err == nil {
				t.Fatalf("mismatched G41V1 object: expected fail-closed error")
			}
		})
	}
}

// TestParseCommandStatusG41V1 confirms parseCommandStatus reads the G41V1
// per-point status octet (5-byte point: int32 + status).
func TestParseCommandStatusG41V1(t *testing.T) {
	// header(5) + index(2) + value(4) + status(1)
	data := []byte{0x29, 0x01, 0x28, 0x01, 0x00, 0x00, 0x00, 0xFC, 0x08, 0x00, 0x00, 0x06}
	if got := parseCommandStatus(data); got != CommandStatusHardwareError {
		t.Fatalf("G41V1 status = %d, want HardwareError (6)", got)
	}
	// Truncated (status octet missing) -> Unknown, never success.
	trunc := []byte{0x29, 0x01, 0x28, 0x01, 0x00, 0x00, 0x00, 0xFC, 0x08, 0x00}
	if got := parseCommandStatus(trunc); got != CommandStatusUnknown {
		t.Fatalf("truncated G41V1 status = %d, want Unknown", got)
	}
}
