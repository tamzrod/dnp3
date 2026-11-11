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

// pubG41V1EchoTransport returns a DirectOperate response carrying one G41V1
// object (index 0, int32 value 2300) whose command-status octet is commandStatus.
type pubG41V1EchoTransport struct {
	lastSeq       uint8
	commandStatus byte
}

func (t *pubG41V1EchoTransport) Send(data []byte) error {
	t.lastSeq = extractPubRequestSeq(data)
	return nil
}

func (t *pubG41V1EchoTransport) SetTimeout(ms int) {}

func (t *pubG41V1EchoTransport) Receive() ([]byte, error) {
	obj := []byte{
		0x29, 0x01, 0x28, 0x01, 0x00,
		0x00, 0x00,
		0xFC, 0x08, 0x00, 0x00,
		t.commandStatus,
	}
	apdu := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: t.lastSeq},
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
	return raw, nil
}

// AO-002 — public master client gate and G41V1 status surface.

// TestPublicOperateGateRejectsUnsupportedObjects asserts the public Operate gate
// fails closed for control objects outside the verified set (G12V1 and G41V1):
// an unsupported group/variation is rejected with ErrUnsupportedOption BEFORE any
// wire traffic (no request is sent).
func TestPublicOperateGateRejectsUnsupportedObjects(t *testing.T) {
	cases := []struct {
		name      string
		group     uint8
		variation uint8
	}{
		{"G41V2", 41, 2},
		{"G41V3", 41, 3},
		{"G41V4", 41, 4},
		{"G42V1", 42, 1},
		{"G30V1 (not a control)", 30, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tr := &recordingPubTransport{}
			cc := newConnectedClientWithTransport(t, tr)
			_, err := cc.Operate(context.Background(), &types.ControlOutput{
				Group: tc.group, Variation: tc.variation, Index: 0,
				CommandType: types.DirectOperate,
				Value:       &types.AnalogCommandValue{Value: 1},
			})
			if !errors.Is(err, dnp3.ErrUnsupportedOption) {
				t.Fatalf("error = %v, want ErrUnsupportedOption", err)
			}
			if tr.sent {
				t.Fatalf("unsupported control object was transmitted despite the gate")
			}
		})
	}
}

// recordingPubTransport records whether any request was transmitted.
type recordingPubTransport struct {
	sent bool
}

func (t *recordingPubTransport) Send(data []byte) error {
	t.sent = true
	return nil
}
func (t *recordingPubTransport) SetTimeout(ms int) {}
func (t *recordingPubTransport) Receive() ([]byte, error) {
	return nil, errors.New("no response expected")
}

// TestPublicOperateG41V1Success asserts a G41V1 DirectOperate with a SUCCESS
// echo surfaces ControlSuccess through the public API.
func TestPublicOperateG41V1Success(t *testing.T) {
	cc := newConnectedClientWithTransport(t, &pubG41V1EchoTransport{commandStatus: 0})
	resp, err := cc.Operate(context.Background(), &types.ControlOutput{
		Group: 41, Variation: 1, Index: 0,
		CommandType: types.DirectOperate,
		Value:       &types.AnalogInt32CommandValue{Value: 2300},
	})
	if err != nil {
		t.Fatalf("Operate error: %v", err)
	}
	if resp.Status != types.ControlSuccess {
		t.Fatalf("Status = %v, want ControlSuccess", resp.Status)
	}
}

// TestPublicOperateG41V1NonSuccess asserts a non-SUCCESS G41V1 echo is surfaced
// as the matching failure, never ControlSuccess.
func TestPublicOperateG41V1NonSuccess(t *testing.T) {
	cases := []struct {
		name   string
		status byte
		want   types.ControlStatus
	}{
		{"not_supported", 4, types.ControlNotSupported},
		{"blocked", 6, types.ControlBlocked},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cc := newConnectedClientWithTransport(t, &pubG41V1EchoTransport{commandStatus: tc.status})
			resp, err := cc.Operate(context.Background(), &types.ControlOutput{
				Group: 41, Variation: 1, Index: 0,
				CommandType: types.DirectOperate,
				Value:       &types.AnalogInt32CommandValue{Value: 2300},
			})
			if err != nil {
				t.Fatalf("Operate error: %v", err)
			}
			if resp.Status != tc.want {
				t.Fatalf("Status = %v, want %v", resp.Status, tc.want)
			}
			if resp.Status == types.ControlSuccess {
				t.Fatalf("non-success G41V1 surfaced as ControlSuccess")
			}
		})
	}
}
