package master

import (
	"testing"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/tl"
)

// g41EchoTransport returns a DirectOperate response echoing exactly one G41
// record of the configured variation/index/value with the given per-point
// command-status octet. It drives the AO-004 Appendix I OUT_OF_RANGE case: a
// G41 echo whose status octet is 0x0C must surface as the definite internal
// CommandStatusOutOfRange — never CommandStatusUnknown, never success, and
// without quarantining the outstation. Layout after IIN:
//
//	29 <var> 28 01 00  <index LE>  <value LE>  <commandStatus>
type g41EchoTransport struct {
	lastSeq       uint8
	variation     uint8
	index         uint16
	valueBytes    []byte
	commandStatus byte
}

func (t *g41EchoTransport) Send(data []byte) error {
	t.lastSeq = extractRequestSeqForBuild(data)
	return nil
}

func (t *g41EchoTransport) SetTimeout(ms int) {}

func (t *g41EchoTransport) Receive() ([]byte, error) {
	return buildG41EchoResponse(t), nil
}

func buildG41EchoResponse(t *g41EchoTransport) []byte {
	record := []byte{41, t.variation, 0x28, 0x01, 0x00, byte(t.index), byte(t.index >> 8)}
	record = append(record, t.valueBytes...)
	record = append(record, t.commandStatus)
	apdu := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: t.lastSeq},
		FuncCode: al.FuncResponse,
		Data:     append([]byte{0x00, 0x00}, record...), // IIN (clear) + object
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

// TestOperateG41OutOfRangeStatus drives the internal DirectOperate path for all
// four G41 variations with an index=1 request echoed as status 0x0C. The result
// must be a definite CommandStatusOutOfRange with no error and no quarantine
// (the remote rejection is known, not ambiguous).
func TestOperateG41OutOfRangeStatus(t *testing.T) {
	cases := []struct {
		name       string
		variation  uint8
		value      interface{}
		valueBytes []byte
	}{
		{"V1 int32", 1, int32(2300), []byte{0xFC, 0x08, 0x00, 0x00}},
		{"V2 int16", 2, int16(12345), []byte{0x39, 0x30}},
		{"V3 float32", 3, float32(1.5), []byte{0x00, 0x00, 0xC0, 0x3F}},
		{"V4 float64", 4, float64(1.5), []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF8, 0x3F}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMaster(DefaultConfig())
			m.SetTransport(&g41EchoTransport{
				variation: tc.variation, index: 1, valueBytes: tc.valueBytes, commandStatus: 0x0C,
			})
			m.SetState(StateInitialized)

			status, err := m.OperateWithStatus(1, false, 41, tc.variation, 1, tc.value)
			if err != nil {
				t.Fatalf("unexpected fail-closed error: %v (status %d)", err, status)
			}
			if status != CommandStatusOutOfRange {
				t.Fatalf("status = %d, want CommandStatusOutOfRange (%d)", status, CommandStatusOutOfRange)
			}
			if status == CommandStatusSuccess || status == CommandStatusUnknown {
				t.Fatalf("OUT_OF_RANGE must be a definite non-success, non-unknown status")
			}
			if m.IsQuarantined(1) {
				t.Fatalf("a definite OUT_OF_RANGE rejection must not quarantine the outstation")
			}
		})
	}
}

// TestParseG41RecordOutOfRangeAllVariations pins that parseG41Record reads the
// 0x0C status octet verbatim for every supported G41 variation.
func TestParseG41RecordOutOfRangeAllVariations(t *testing.T) {
	cases := []struct {
		variation uint8
		record    []byte
	}{
		{1, []byte{41, 1, 0x28, 0x01, 0x00, 0x01, 0x00, 0xFC, 0x08, 0x00, 0x00, 0x0C}},
		{2, []byte{41, 2, 0x28, 0x01, 0x00, 0x01, 0x00, 0x39, 0x30, 0x0C}},
		{3, []byte{41, 3, 0x28, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0xC0, 0x3F, 0x0C}},
		{4, []byte{41, 4, 0x28, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF8, 0x3F, 0x0C}},
	}
	for _, tc := range cases {
		idx, _, status, ok := parseG41Record(tc.record, 0, tc.variation)
		if !ok {
			t.Fatalf("V%d record not parsed", tc.variation)
		}
		if idx != 1 {
			t.Fatalf("V%d index = %d, want 1", tc.variation, idx)
		}
		if status != CommandStatusOutOfRange {
			t.Fatalf("V%d status = %d, want CommandStatusOutOfRange (%d)", tc.variation, status, CommandStatusOutOfRange)
		}
	}
}

// TestG41CommandStatusOutOfRange confirms g41CommandStatus returns the explicit
// OUT_OF_RANGE status (not UNKNOWN) when the echoed object matches the request.
func TestG41CommandStatusOutOfRange(t *testing.T) {
	record := []byte{41, 1, 0x28, 0x01, 0x00, 0x01, 0x00, 0xFC, 0x08, 0x00, 0x00, 0x0C}
	got := g41CommandStatus(record, 1, 1, uint64(2300), al.IIN{})
	if got != CommandStatusOutOfRange {
		t.Fatalf("g41CommandStatus = %d, want CommandStatusOutOfRange (%d)", got, CommandStatusOutOfRange)
	}
}
