package master

import (
	"context"
	"errors"
	"testing"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/tl"
	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/types"
)

// AO-003 — public master client gate and G41V2 status surface.
//
// These tests exercise the public (*client).Operate path for Group 41
// Variation 2 (int16) against deterministic in-repo fake transports (the same
// helpers the G41V1 public suite uses); no real socket is opened.

// pubG41V2EchoTransport returns a DirectOperate response carrying one G41V2
// object (index 0, int16 value 42) whose command-status octet is commandStatus.
type pubG41V2EchoTransport struct {
	lastSeq       uint8
	commandStatus byte
	index         uint16
	value         int16
}

func (t *pubG41V2EchoTransport) Send(data []byte) error {
	t.lastSeq = extractPubRequestSeq(data)
	return nil
}

func (t *pubG41V2EchoTransport) SetTimeout(ms int) {}

func (t *pubG41V2EchoTransport) Receive() ([]byte, error) {
	obj := []byte{
		0x29, 0x02, 0x28, 0x01, 0x00,
		byte(t.index), byte(t.index >> 8),
		byte(t.value), byte(t.value >> 8),
		t.commandStatus,
	}
	apdu := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: t.lastSeq},
		FuncCode: al.FuncResponse,
		Data:     append([]byte{0x00, 0x00}, obj...),
	}
	frag := tl.Fragment{FIR: true, FIN: true, Data: apdu.Encode()}
	dllFrame := &frame.Frame{
		Control:  frame.Control{DIR: false, PRM: false, FuncCode: frame.FuncConfirmedUserDataR},
		DestAddr: 1, SrcAddr: 2, Data: tl.EncodeFragment(frag),
	}
	raw, _ := frame.Encode(dllFrame)
	return raw, nil
}

// TestNewAnalogInt16Control locks the G41V2 constructor: Group 41, Variation 2,
// the given index and a typed *types.AnalogInt16CommandValue carrying the exact
// value (no conversion).
func TestNewAnalogInt16Control(t *testing.T) {
	cmd := types.NewAnalogInt16Control(7, -1234, types.DirectOperate)
	if cmd == nil {
		t.Fatal("NewAnalogInt16Control returned nil")
	}
	if cmd.Group != 41 || cmd.Variation != 2 || cmd.Index != 7 {
		t.Fatalf("group/var/index = %d/%d/%d, want 41/2/7", cmd.Group, cmd.Variation, cmd.Index)
	}
	if cmd.CommandType != types.DirectOperate {
		t.Fatalf("CommandType = %v, want DirectOperate", cmd.CommandType)
	}
	v, ok := cmd.Value.(*types.AnalogInt16CommandValue)
	if !ok {
		t.Fatalf("Value type = %T, want *types.AnalogInt16CommandValue", cmd.Value)
	}
	if v.Value != -1234 {
		t.Fatalf("value = %d, want -1234", v.Value)
	}
}

// TestPublicOperateG41V2Success asserts a G41V2 DirectOperate with a SUCCESS echo
// surfaces ControlSuccess through the public API.
func TestPublicOperateG41V2Success(t *testing.T) {
	cc := newConnectedClientWithTransport(t, &pubG41V2EchoTransport{commandStatus: 0, value: 42})
	resp, err := cc.Operate(context.Background(), types.NewAnalogInt16Control(0, 42, types.DirectOperate))
	if err != nil {
		t.Fatalf("Operate error: %v", err)
	}
	if resp.Status != types.ControlSuccess {
		t.Fatalf("Status = %v, want ControlSuccess", resp.Status)
	}
}

// TestPublicOperateG41V2NonSuccess asserts an explicit non-SUCCESS G41V2 echo is
// surfaced as the matching failure, never ControlSuccess.
func TestPublicOperateG41V2NonSuccess(t *testing.T) {
	cases := []struct {
		name   string
		status byte
		want   types.ControlStatus
	}{
		{"not_supported", 4, types.ControlNotSupported},
		{"hardware_error", 6, types.ControlHardwareError},
		{"no_select", 2, types.ControlNoSelect},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cc := newConnectedClientWithTransport(t, &pubG41V2EchoTransport{commandStatus: tc.status, value: 42})
			resp, err := cc.Operate(context.Background(), types.NewAnalogInt16Control(0, 42, types.DirectOperate))
			if err != nil {
				t.Fatalf("Operate error: %v", err)
			}
			if resp.Status != tc.want {
				t.Fatalf("Status = %v, want %v", resp.Status, tc.want)
			}
			if resp.Status == types.ControlSuccess {
				t.Fatalf("non-success G41V2 surfaced as ControlSuccess")
			}
		})
	}
}

// TestPublicOperateG41V2MismatchedEchoNotSuccess asserts a G41V2 echo whose
// value does not match the request is never ControlSuccess: the public path
// fails closed (error, no success status).
func TestPublicOperateG41V2MismatchedEchoNotSuccess(t *testing.T) {
	cc := newConnectedClientWithTransport(t, &pubG41V2EchoTransport{commandStatus: 0, value: 43})
	resp, err := cc.Operate(context.Background(), types.NewAnalogInt16Control(0, 42, types.DirectOperate))
	if err == nil && resp != nil && resp.Status == types.ControlSuccess {
		t.Fatalf("mismatched G41V2 echo surfaced as ControlSuccess")
	}
}

// TestPublicOperateG41V2UnsupportedModesBlockedBeforeSend asserts select-before-
// operate and direct-operate-no-response for G41V2 are rejected with
// ErrUnsupportedOption BEFORE any wire traffic.
func TestPublicOperateG41V2UnsupportedModesBlockedBeforeSend(t *testing.T) {
	cases := []struct {
		name string
		mode types.CommandType
	}{
		{"select_then_operate", types.SelectThenOperate},
		{"direct_operate_no_response", types.DirectOperateNoResponse},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tr := &recordingPubTransport{}
			cc := newConnectedClientWithTransport(t, tr)
			_, err := cc.Operate(context.Background(), types.NewAnalogInt16Control(0, 42, tc.mode))
			if !errors.Is(err, dnp3.ErrUnsupportedOption) {
				t.Fatalf("error = %v, want ErrUnsupportedOption", err)
			}
			if tr.sent {
				t.Fatalf("unsupported G41V2 command mode was transmitted despite the gate")
			}
		})
	}
}

// TestPublicOperateGateRejectsG41V5BeforeSend asserts the public gate still
// fails closed for the unimplemented variation G41V5: an ErrUnsupportedOption is
// returned before any wire traffic. (G41V3 and G41V4 are now implemented and
// covered by their own suites.)
func TestPublicOperateGateRejectsG41V5BeforeSend(t *testing.T) {
	cases := []struct {
		name      string
		variation uint8
	}{
		{"G41V5", 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tr := &recordingPubTransport{}
			cc := newConnectedClientWithTransport(t, tr)
			_, err := cc.Operate(context.Background(), &types.ControlOutput{
				Group: 41, Variation: tc.variation, Index: 0,
				CommandType: types.DirectOperate,
				Value:       &types.AnalogInt16CommandValue{Value: 42},
			})
			if !errors.Is(err, dnp3.ErrUnsupportedOption) {
				t.Fatalf("error = %v, want ErrUnsupportedOption", err)
			}
			if tr.sent {
				t.Fatalf("%s was transmitted despite the gate", tc.name)
			}
		})
	}
}

// TestPublicOperateG41V2RejectsWrongTypeAndRange asserts a G41V2 command whose
// value is a non-integer type or an out-of-range int32 is rejected without any
// wire traffic (no silent narrowing/wrap), and never surfaces as success.
func TestPublicOperateG41V2RejectsWrongTypeAndRange(t *testing.T) {
	cases := []struct {
		name  string
		value types.CommandValue
	}{
		{"float64 value", &types.AnalogCommandValue{Value: 1.5}},
		{"int32 overflow", &types.AnalogInt32CommandValue{Value: 32768}},
		{"int32 underflow", &types.AnalogInt32CommandValue{Value: -32769}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tr := &recordingPubTransport{}
			cc := newConnectedClientWithTransport(t, tr)
			resp, err := cc.Operate(context.Background(), &types.ControlOutput{
				Group: 41, Variation: 2, Index: 0,
				CommandType: types.DirectOperate,
				Value:       tc.value,
			})
			if err == nil {
				t.Fatalf("expected rejection for %s, got resp=%v", tc.name, resp)
			}
			if resp != nil && resp.Status == types.ControlSuccess {
				t.Fatalf("rejected G41V2 value surfaced as ControlSuccess")
			}
			if tr.sent {
				t.Fatalf("invalid G41V2 value was transmitted despite the pre-send rejection")
			}
		})
	}
}
