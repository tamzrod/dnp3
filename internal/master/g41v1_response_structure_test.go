package master

import (
	"errors"
	"testing"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/tl"
)

// AO-002 F4/F5/F6 — master-side G41V1 response handling.
//
// rawG41Transport returns a DirectOperate response whose object payload (the
// bytes after the 2-byte IIN) and IIN bytes are supplied verbatim, so response
// structure and IIN contradictions can be exercised exactly. It never contacts a
// real outstation.
type rawG41Transport struct {
	lastSeq uint8
	object  []byte // bytes after the IIN
	iin     [2]byte
}

func (t *rawG41Transport) Send(data []byte) error {
	t.lastSeq = extractRequestSeqForBuild(data)
	return nil
}

func (t *rawG41Transport) SetTimeout(ms int) {}

func (t *rawG41Transport) Receive() ([]byte, error) {
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

// g41Record builds one G41V1 record: header(5) + index(2 LE) + int32(4 LE) + status(1).
func g41Record(group, variation, qualifier uint8, count uint16, index uint16, value int32, status byte) []byte {
	return []byte{
		group, variation, qualifier,
		byte(count), byte(count >> 8),
		byte(index), byte(index >> 8),
		byte(value), byte(value >> 8), byte(value >> 16), byte(value >> 24),
		status,
	}
}

func newG41RawMaster(t *rawG41Transport) *Master {
	m := NewMaster(DefaultConfig())
	m.SetTransport(t)
	m.SetState(StateInitialized)
	return m
}

// TestOperateG41V1F4EchoedValueMustMatch verifies F4: a SUCCESS echo with the
// correct index/variation but a WRONG int32 value is never success; it is
// fail-closed (UNKNOWN + ErrControlOutcomeUnknown, quarantined). The correct
// value is the positive control.
func TestOperateG41V1F4EchoedValueMustMatch(t *testing.T) {
	rows := []struct {
		name      string
		echoValue int32
		wantOK    bool
	}{
		{"correct value", 2300, true},
		{"wrong value", 2301, false},
		{"zero value", 0, false},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			m := newG41RawMaster(&rawG41Transport{
				object: g41Record(41, 1, 0x28, 1, 0, row.echoValue, 0x00),
			})
			status, err := m.OperateWithStatus(1, false, 41, 1, 0, int32(2300))
			if row.wantOK {
				if err != nil {
					t.Fatalf("correct echo: unexpected error %v", err)
				}
				if status != CommandStatusSuccess {
					t.Fatalf("correct echo: status = %d, want Success", status)
				}
				return
			}
			if status == CommandStatusSuccess {
				t.Fatalf("wrong echoed value surfaced as SUCCESS")
			}
			if !errors.Is(err, ErrControlOutcomeUnknown) {
				t.Fatalf("wrong echoed value: error = %v, want ErrControlOutcomeUnknown", err)
			}
			if !m.IsQuarantined(1) {
				t.Fatalf("wrong echoed value must quarantine the outstation")
			}
		})
	}
}

// TestOperateG41V1F5StructureRejections verifies F5: any response whose G41V1
// object is not exactly one qualifier-0x28 record with count 1, correct
// group/variation/index/value and a readable status is never success. Cases with
// a positive status octet must not surface as SUCCESS; malformed structure is
// fail-closed (UNKNOWN + ErrControlOutcomeUnknown).
func TestOperateG41V1F5StructureRejections(t *testing.T) {
	good := g41Record(41, 1, 0x28, 1, 0, 2300, 0x00)

	cases := []struct {
		name    string
		object  []byte
		wantErr bool // true => must be UNKNOWN + ErrControlOutcomeUnknown
	}{
		{"empty object (no record)", nil, true},
		{"missing record after header", []byte{0x29, 0x01, 0x28, 0x01, 0x00, 0x00, 0x00}, true},
		{"count zero", g41Record(41, 1, 0x28, 0, 0, 2300, 0x00), true},
		{"count two", g41Record(41, 1, 0x28, 2, 0, 2300, 0x00), true},
		{"two records", append(append([]byte{}, good...), good...), true},
		{"truncated status", good[:len(good)-1], true},
		{"surplus byte", append(append([]byte{}, good...), 0xEE), true},
		{"qualifier 0x00 (legacy) not valid", g41Record(41, 1, 0x00, 1, 0, 2300, 0x00), true},
		{"wrong group", g41Record(12, 1, 0x28, 1, 0, 2300, 0x00), true},
		{"wrong variation", g41Record(41, 2, 0x28, 1, 0, 2300, 0x00), true},
		{"wrong index", g41Record(41, 1, 0x28, 1, 9, 2300, 0x00), true},
		{"wrong value", g41Record(41, 1, 0x28, 1, 0, 999, 0x00), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newG41RawMaster(&rawG41Transport{object: tc.object})
			status, err := m.OperateWithStatus(1, false, 41, 1, 0, int32(2300))
			if status == CommandStatusSuccess {
				t.Fatalf("malformed/mismatched G41V1 response surfaced as SUCCESS")
			}
			if tc.wantErr && !errors.Is(err, ErrControlOutcomeUnknown) {
				t.Fatalf("error = %v, want ErrControlOutcomeUnknown", err)
			}
		})
	}
}

// TestOperateG41V1F5ExplicitNonSuccessStatus confirms a structurally valid echo
// carrying an explicit non-SUCCESS command-status octet is surfaced as-is (never
// success) without being mistaken for a structural failure.
func TestOperateG41V1F5ExplicitNonSuccessStatus(t *testing.T) {
	m := newG41RawMaster(&rawG41Transport{object: g41Record(41, 1, 0x28, 1, 0, 2300, 0x06)})
	status, err := m.OperateWithStatus(1, false, 41, 1, 0, int32(2300))
	if err != nil {
		t.Fatalf("explicit non-success status should classify (not error), got %v", err)
	}
	if status != CommandStatusBlocked {
		t.Fatalf("status = %d, want Blocked (6)", status)
	}
}

// TestOperateG41V1F6IINContradictsSuccessEcho CONFIRMS F6: when a syntactically
// matching G41V1 SUCCESS echo is accompanied by error-indicating IIN bits, the
// current implementation still reports CommandStatusSuccess — the IIN is not
// reconciled against the echoed command-status octet. This test documents the
// demonstrated current (defective) behavior; it is NOT an assertion of desired
// behavior. A fix requires separate owner authorization (see docs/verification/
// AO-002.md §9). A clear-IIN SUCCESS echo is the positive control.
func TestOperateG41V1F6IINContradictsSuccessEcho(t *testing.T) {
	successEcho := g41Record(41, 1, 0x28, 1, 0, 2300, 0x00)

	// (name, IIN bytes, current-behavior status)
	cases := []struct {
		name string
		iin  [2]byte
	}{
		{"clear IIN + SUCCESS echo (positive control)", [2]byte{0x00, 0x00}},
		{"ParameterError + SUCCESS echo", [2]byte{0x00, 0x20}},
		{"ObjectUnknown + SUCCESS echo", [2]byte{0x00, 0x40}},
		{"FuncUnknown + SUCCESS echo", [2]byte{0x00, 0x80}},
		{"BadConfig + SUCCESS echo", [2]byte{0x00, 0x04}},
		{"AlreadyExecuting + SUCCESS echo", [2]byte{0x00, 0x08}},
		{"LocalControl + SUCCESS echo", [2]byte{0x04, 0x00}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newG41RawMaster(&rawG41Transport{object: successEcho, iin: tc.iin})
			status, err := m.OperateWithStatus(1, false, 41, 1, 0, int32(2300))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Confirmed current behavior: the echoed SUCCESS octet wins regardless
			// of error IIN, and no quarantine is applied. Recorded as evidence of
			// F6, not as desired behavior.
			if status != CommandStatusSuccess {
				t.Fatalf("current F6 behavior changed: status = %d, want Success (defect still reproduced)", status)
			}
			if m.IsQuarantined(1) {
				t.Fatalf("F6: contradictory IIN SUCCESS echo quarantined; behavior changed")
			}
		})
	}
}
