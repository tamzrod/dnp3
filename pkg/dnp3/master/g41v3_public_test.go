package master

import (
	"context"
	"errors"
	"math"
	"testing"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/tl"
	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/types"
)

// AO-003 — public master client gate and G41V3 status surface.
//
// These tests exercise the public (*client).Operate path for Group 41
// Variation 3 (single-precision float) against deterministic in-repo fake
// transports (the same helpers the G41V1/G41V2 public suites use); no real
// socket is opened.

// pubG41V3EchoTransport returns a DirectOperate response carrying one G41V3
// object (index 0, float32 value) whose command-status octet is commandStatus.
type pubG41V3EchoTransport struct {
	lastSeq       uint8
	commandStatus byte
	index         uint16
	valueBits     uint32
}

func (t *pubG41V3EchoTransport) Send(data []byte) error {
	t.lastSeq = extractPubRequestSeq(data)
	return nil
}

func (t *pubG41V3EchoTransport) SetTimeout(ms int) {}

func (t *pubG41V3EchoTransport) Receive() ([]byte, error) {
	obj := []byte{
		0x29, 0x03, 0x28, 0x01, 0x00,
		byte(t.index), byte(t.index >> 8),
		byte(t.valueBits), byte(t.valueBits >> 8), byte(t.valueBits >> 16), byte(t.valueBits >> 24),
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

// TestNewAnalogFloat32Control locks the G41V3 constructor: Group 41, Variation 3,
// the given index and a typed *types.AnalogFloat32CommandValue carrying the
// exact value (no conversion).
func TestNewAnalogFloat32Control(t *testing.T) {
	cmd := types.NewAnalogFloat32Control(7, -0.5, types.DirectOperate)
	if cmd == nil {
		t.Fatal("NewAnalogFloat32Control returned nil")
	}
	if cmd.Group != 41 || cmd.Variation != 3 || cmd.Index != 7 {
		t.Fatalf("group/var/index = %d/%d/%d, want 41/3/7", cmd.Group, cmd.Variation, cmd.Index)
	}
	if cmd.CommandType != types.DirectOperate {
		t.Fatalf("CommandType = %v, want DirectOperate", cmd.CommandType)
	}
	v, ok := cmd.Value.(*types.AnalogFloat32CommandValue)
	if !ok {
		t.Fatalf("Value type = %T, want *types.AnalogFloat32CommandValue", cmd.Value)
	}
	if math.Float32bits(v.Value) != math.Float32bits(-0.5) {
		t.Fatalf("value bits = %08X, want %08X", math.Float32bits(v.Value), math.Float32bits(-0.5))
	}
}

// TestPublicOperateG41V3Success asserts a G41V3 DirectOperate with a SUCCESS echo
// surfaces ControlSuccess through the public API, for representative finite
// values including signed zero.
func TestPublicOperateG41V3Success(t *testing.T) {
	cases := []struct {
		name string
		bits uint32
	}{
		{"+1.0", math.Float32bits(1.0)},
		{"-0.5", math.Float32bits(-0.5)},
		{"-0.0", math.Float32bits(float32(math.Copysign(0, -1)))},
		{"smallest positive subnormal", 0x00000001},
		{"max finite", 0x7F7FFFFF},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cc := newConnectedClientWithTransport(t, &pubG41V3EchoTransport{commandStatus: 0, valueBits: tc.bits})
			resp, err := cc.Operate(context.Background(), types.NewAnalogFloat32Control(0, math.Float32frombits(tc.bits), types.DirectOperate))
			if err != nil {
				t.Fatalf("Operate error: %v", err)
			}
			if resp.Status != types.ControlSuccess {
				t.Fatalf("Status = %v, want ControlSuccess", resp.Status)
			}
		})
	}
}

// TestPublicOperateG41V3NonSuccess asserts an explicit non-SUCCESS G41V3 echo is
// surfaced as the matching failure, never ControlSuccess.
func TestPublicOperateG41V3NonSuccess(t *testing.T) {
	cases := []struct {
		name   string
		status byte
		want   types.ControlStatus
	}{
		{"not_supported", 4, types.ControlNotSupported},
		{"blocked", 6, types.ControlBlocked},
		{"no_select", 2, types.ControlNoSelect},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cc := newConnectedClientWithTransport(t, &pubG41V3EchoTransport{commandStatus: tc.status, valueBits: math.Float32bits(1.0)})
			resp, err := cc.Operate(context.Background(), types.NewAnalogFloat32Control(0, 1.0, types.DirectOperate))
			if err != nil {
				t.Fatalf("Operate error: %v", err)
			}
			if resp.Status != tc.want {
				t.Fatalf("Status = %v, want %v", resp.Status, tc.want)
			}
			if resp.Status == types.ControlSuccess {
				t.Fatalf("non-success G41V3 surfaced as ControlSuccess")
			}
		})
	}
}

// TestPublicOperateG41V3SignedZeroEchoNotSuccess asserts the public path compares
// the echoed setpoint by raw IEEE-754 bits: a matching -0.0 echo is success,
// while an echo that flips the sign bit (+0.0 for a -0.0 request) is never
// ControlSuccess.
func TestPublicOperateG41V3SignedZeroEchoNotSuccess(t *testing.T) {
	negZero := math.Float32bits(float32(math.Copysign(0, -1)))
	posZero := math.Float32bits(float32(0))

	t.Run("matching -0.0 echo is success", func(t *testing.T) {
		cc := newConnectedClientWithTransport(t, &pubG41V3EchoTransport{commandStatus: 0, valueBits: negZero})
		resp, err := cc.Operate(context.Background(), types.NewAnalogFloat32Control(0, float32(math.Copysign(0, -1)), types.DirectOperate))
		if err != nil {
			t.Fatalf("Operate error: %v", err)
		}
		if resp.Status != types.ControlSuccess {
			t.Fatalf("Status = %v, want ControlSuccess", resp.Status)
		}
	})

	t.Run("sign-flipped +0.0 echo is not success", func(t *testing.T) {
		cc := newConnectedClientWithTransport(t, &pubG41V3EchoTransport{commandStatus: 0, valueBits: posZero})
		resp, err := cc.Operate(context.Background(), types.NewAnalogFloat32Control(0, float32(math.Copysign(0, -1)), types.DirectOperate))
		if err == nil && resp != nil && resp.Status == types.ControlSuccess {
			t.Fatalf("sign-flipped G41V3 echo surfaced as ControlSuccess")
		}
	})
}

// TestPublicOperateG41V3MismatchedEchoNotSuccess asserts a G41V3 echo whose value
// does not match the request is never ControlSuccess: the public path fails
// closed (error, no success status).
func TestPublicOperateG41V3MismatchedEchoNotSuccess(t *testing.T) {
	cc := newConnectedClientWithTransport(t, &pubG41V3EchoTransport{commandStatus: 0, valueBits: math.Float32bits(2.0)})
	resp, err := cc.Operate(context.Background(), types.NewAnalogFloat32Control(0, 1.0, types.DirectOperate))
	if err == nil && resp != nil && resp.Status == types.ControlSuccess {
		t.Fatalf("mismatched G41V3 echo surfaced as ControlSuccess")
	}
}

// TestPublicOperateG41V3UnsupportedModesBlockedBeforeSend asserts select-before-
// operate and direct-operate-no-response for G41V3 are rejected with
// ErrUnsupportedOption BEFORE any wire traffic.
func TestPublicOperateG41V3UnsupportedModesBlockedBeforeSend(t *testing.T) {
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
			_, err := cc.Operate(context.Background(), types.NewAnalogFloat32Control(0, 1.0, tc.mode))
			if !errors.Is(err, dnp3.ErrUnsupportedOption) {
				t.Fatalf("error = %v, want ErrUnsupportedOption", err)
			}
			if tr.sent {
				t.Fatalf("unsupported G41V3 command mode was transmitted despite the gate")
			}
		})
	}
}

// TestPublicOperateG41V3RejectsNonFiniteBeforeSend asserts a G41V3 command whose
// value is NaN or ±Inf is rejected without any wire traffic (fail closed) and
// never surfaces as success.
func TestPublicOperateG41V3RejectsNonFiniteBeforeSend(t *testing.T) {
	cases := []struct {
		name  string
		value float32
	}{
		{"NaN", float32(math.NaN())},
		{"+Inf", float32(math.Inf(1))},
		{"-Inf", float32(math.Inf(-1))},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tr := &recordingPubTransport{}
			cc := newConnectedClientWithTransport(t, tr)
			resp, err := cc.Operate(context.Background(), types.NewAnalogFloat32Control(0, tc.value, types.DirectOperate))
			if err == nil {
				t.Fatalf("expected rejection for %s, got resp=%v", tc.name, resp)
			}
			if resp != nil && resp.Status == types.ControlSuccess {
				t.Fatalf("rejected G41V3 value surfaced as ControlSuccess")
			}
			if tr.sent {
				t.Fatalf("non-finite G41V3 value was transmitted despite the pre-send rejection")
			}
		})
	}
}

// TestPublicOperateG41V3RejectsInexactFloat64BeforeSend asserts a G41V3 command
// carrying a float64 that is not exactly representable as float32 is rejected
// without any wire traffic (no silent narrowing), and never surfaces as success.
func TestPublicOperateG41V3RejectsInexactFloat64BeforeSend(t *testing.T) {
	tr := &recordingPubTransport{}
	cc := newConnectedClientWithTransport(t, tr)
	resp, err := cc.Operate(context.Background(), &types.ControlOutput{
		Group: 41, Variation: 3, Index: 0,
		CommandType: types.DirectOperate,
		Value:       &types.AnalogCommandValue{Value: 0.1},
	})
	if err == nil {
		t.Fatalf("expected rejection for an inexact float64, got resp=%v", resp)
	}
	if resp != nil && resp.Status == types.ControlSuccess {
		t.Fatalf("inexact float64 surfaced as ControlSuccess")
	}
	if tr.sent {
		t.Fatalf("inexact float64 was transmitted despite the pre-send rejection")
	}
}
