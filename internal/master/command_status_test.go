package master

import (
	"testing"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/tl"
)

// statusEchoTransport records the SEQ of the most recent request and returns a
// DirectOperate response carrying a G12V1 object whose per-point status byte is
// the given commandStatus. This mirrors the IEEE 1815 response: the request
// header is echoed with the CROB status byte replaced by the command status.
type statusEchoTransport struct {
	lastSeq       uint8
	commandStatus byte
}

func (t *statusEchoTransport) Send(data []byte) error {
	t.lastSeq = extractRequestSeqForBuild(data)
	return nil
}

func (t *statusEchoTransport) SetTimeout(ms int) {}

func (t *statusEchoTransport) Receive() ([]byte, error) {
	return buildG12V1StatusResponse(t.lastSeq, t.commandStatus), nil
}

// buildG12V1StatusResponse builds a DLL+TL+APDU DirectOperate response whose
// application data carries one G12V1 CROB object with the given command status
// in the per-point status byte. Layout (after IIN):
//
//	0C 01 00 01          G12V1, qualifier 0x00, count 1
//	34 12                index 0x1234 (LSB)
//	07 01                code=LATCH_ON, count=1
//	00 00 00 00          onTime=0
//	00 00 00 00          offTime=0
//	<commandStatus>      per-point command status (CTRL-01)
func buildG12V1StatusResponse(seq uint8, commandStatus byte) []byte {
	obj := []byte{
		0x0C, 0x01, 0x00, 0x01,
		0x34, 0x12,
		0x07, 0x01,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		commandStatus,
	}
	apdu := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: seq},
		FuncCode: al.FuncResponse,
		Data:     append([]byte{0x00, 0x00}, obj...), // IIN (all clear) + object
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

// TestParseCommandStatusVectors locks the per-point command-status mapping for
// the G12V1 control response (DNP3-020). A failed point must never map to
// CommandStatusSuccess.
func TestParseCommandStatusVectors(t *testing.T) {
	cases := []struct {
		name string
		byte byte
		want CommandStatus
	}{
		{"success", 0, CommandStatusSuccess},
		{"timeout", 1, CommandStatusTimeout},
		{"no_select", 2, CommandStatusNoSelect},
		{"bad_format", 3, CommandStatusBadFormat},
		{"not_supported", 4, CommandStatusNotSupported},
		{"already_active", 5, CommandStatusAlreadyActive},
		{"blocked", 6, CommandStatusBlocked},
		{"local", 7, CommandStatusLocal},
		{"too_many", 8, CommandStatusTooMany},
		{"not_authorized", 9, CommandStatusNotAuthorized},
		{"autonomous", 10, CommandStatusAutonomous},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := []byte{0x0C, 0x01, 0x00, 0x01, 0x34, 0x12, 0x07, 0x01,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, tc.byte}
			got := parseCommandStatus(data)
			if got != tc.want {
				t.Fatalf("status byte %d -> %d, want %d", tc.byte, got, tc.want)
			}
			// The core acceptance criterion: a non-success byte is never success.
			if tc.byte != 0 && got == CommandStatusSuccess {
				t.Fatalf("failed point mapped to success")
			}
		})
	}
}

// TestParseCommandStatusMissingObject confirms that a response with no G12V1
// object yields CommandStatusUnknown, never success.
func TestParseCommandStatusMissingObject(t *testing.T) {
	// Empty object data.
	if got := parseCommandStatus(nil); got != CommandStatusUnknown {
		t.Fatalf("nil data -> %d, want Unknown", got)
	}
	// Unrelated object (G1V1) only.
	data := []byte{0x01, 0x01, 0x07, 0x01, 0x01}
	if got := parseCommandStatus(data); got != CommandStatusUnknown {
		t.Fatalf("non-G12 data -> %d, want Unknown", got)
	}
}

// TestParseCommandStatusTruncated confirms a truncated G12V1 object yields
// Unknown rather than a misread byte.
func TestParseCommandStatusTruncated(t *testing.T) {
	// Header + index + only 9 of 11 CROB bytes (missing the status byte).
	data := []byte{0x0C, 0x01, 0x00, 0x01, 0x34, 0x12, 0x07, 0x01,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if got := parseCommandStatus(data); got != CommandStatusUnknown {
		t.Fatalf("truncated -> %d, want Unknown", got)
	}
}

// TestOperateWithStatusSuccessRejected drives OperateWithStatus against a
// canned transport that returns a G12V1 status response, and asserts the
// master surfaces the real status — success for a 0 byte, and the matching
// failure code otherwise (never success on failure).
func TestOperateWithStatusSuccessRejected(t *testing.T) {
	cases := []struct {
		name string
		byte byte
		want CommandStatus
	}{
		{"success", 0, CommandStatusSuccess},
		{"rejected_not_supported", 4, CommandStatusNotSupported},
		{"rejected_blocked", 6, CommandStatusBlocked},
		{"rejected_no_select", 2, CommandStatusNoSelect},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMaster(DefaultConfig())
			m.SetTransport(&statusEchoTransport{commandStatus: tc.byte})
			m.SetState(StateInitialized)

			status, err := m.OperateWithStatus(1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn))
			if err != nil {
				t.Fatalf("OperateWithStatus returned error: %v", err)
			}
			if status != tc.want {
				t.Fatalf("status = %d, want %d", status, tc.want)
			}
			if tc.byte != 0 && status == CommandStatusSuccess {
				t.Fatalf("failed command surfaced as success")
			}
		})
	}
}

// TestOperateWithStatusIINOnlyClearSuccess confirms that an outstation
// returning an IIN-only response (no G12V1 status object) with a clear IIN is
// reported as success — real outstations may omit the G12V1 status echo on a
// valid Direct-Operate success (MEXT-012, fixing R1). This must NOT false-pass
// when the IIN carries error bits (see TestOperateWithStatusIINOnlyError).
func TestOperateWithStatusIINOnlyClearSuccess(t *testing.T) {
	m := NewMaster(DefaultConfig())
	// echoSeqTransport returns an IIN-only response with a clear IIN.
	m.SetTransport(&echoSeqTransport{})
	m.SetState(StateInitialized)

	status, err := m.OperateWithStatus(1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn))
	if err != nil {
		t.Fatalf("OperateWithStatus returned error: %v", err)
	}
	if status != CommandStatusSuccess {
		t.Fatalf("IIN-only clear response status = %d, want Success (R1 fix)", status)
	}
}

// iinOnlyErrorTransport returns an IIN-only response (no G12V1 object) whose
// IIN carries an error bit (ParameterError), exercising the MEXT-012 rule
// that an IIN-only response with error IIN is a failure, never success.
type iinOnlyErrorTransport struct {
	lastSeq uint8
	iin     al.IIN
}

func (t *iinOnlyErrorTransport) Send(data []byte) error {
	t.lastSeq = extractRequestSeqForBuild(data)
	return nil
}

func (t *iinOnlyErrorTransport) SetTimeout(ms int) {}

func (t *iinOnlyErrorTransport) Receive() ([]byte, error) {
	return buildResponseFrameWithIIN(t.lastSeq, t.iin), nil
}

// TestOperateWithStatusIINOnlyError confirms that an IIN-only response with an
// error IIN bit is a failure (never success), per MEXT-012.
func TestOperateWithStatusIINOnlyError(t *testing.T) {
	cases := []struct {
		name string
		iin  al.IIN
		want CommandStatus
	}{
		{"parameter_error", al.IIN{ParameterError: true}, CommandStatusBadFormat},
		{"func_unknown", al.IIN{FuncUnknown: true}, CommandStatusNotSupported},
		{"object_unknown", al.IIN{ObjectUnknown: true}, CommandStatusNotSupported},
		{"local_control", al.IIN{LocalControl: true}, CommandStatusLocal},
		{"already_executing", al.IIN{AlreadyExecuting: true}, CommandStatusAlreadyActive},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMaster(DefaultConfig())
			m.SetTransport(&iinOnlyErrorTransport{iin: tc.iin})
			m.SetState(StateInitialized)

			status, err := m.OperateWithStatus(1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn))
			if err != nil {
				t.Fatalf("OperateWithStatus returned error: %v", err)
			}
			if status == CommandStatusSuccess {
				t.Fatalf("error IIN (%s) reported as success; want non-success", tc.name)
			}
			if status != tc.want {
				t.Fatalf("status = %d, want %d", status, tc.want)
			}
		})
	}
}

// truncatedG12V1Transport returns a response carrying a G12V1 object whose
// CROB value is truncated (status byte missing), exercising the MEXT-012 rule
// that a truncated G12V1 object is a failure (CommandStatusUnknown), never
// success.
type truncatedG12V1Transport struct {
	lastSeq uint8
}

func (t *truncatedG12V1Transport) Send(data []byte) error {
	t.lastSeq = extractRequestSeqForBuild(data)
	return nil
}

func (t *truncatedG12V1Transport) SetTimeout(ms int) {}

func (t *truncatedG12V1Transport) Receive() ([]byte, error) {
	return buildTruncatedG12V1Response(t.lastSeq), nil
}

// buildTruncatedG12V1Response builds a response with a G12V1 header + index
// but only 9 of 11 CROB bytes (the per-point status byte is missing).
func buildTruncatedG12V1Response(seq uint8) []byte {
	obj := []byte{
		0x0C, 0x01, 0x00, 0x01,
		0x34, 0x12,
		0x08, 0x01,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, // only 7 of 9 remaining CROB bytes; status byte missing
	}
	apdu := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: seq},
		FuncCode: al.FuncResponse,
		Data:     append([]byte{0x00, 0x00}, obj...),
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

// TestOperateWithStatusTruncatedNotSuccess confirms a truncated G12V1 object
// (status byte missing) yields a failure, never success (MEXT-012).
func TestOperateWithStatusTruncatedNotSuccess(t *testing.T) {
	m := NewMaster(DefaultConfig())
	m.SetTransport(&truncatedG12V1Transport{})
	m.SetState(StateInitialized)

	status, err := m.OperateWithStatus(1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn))
	if err != nil {
		t.Fatalf("OperateWithStatus returned error: %v", err)
	}
	if status == CommandStatusSuccess {
		t.Fatalf("truncated G12V1 reported as success; want non-success")
	}
	if status != CommandStatusUnknown {
		t.Fatalf("truncated status = %d, want Unknown", status)
	}
}
