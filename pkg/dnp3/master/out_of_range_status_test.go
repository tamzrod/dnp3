package master

import (
	"context"
	"testing"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/master"
	"dnp3/internal/tl"
	"dnp3/pkg/dnp3/types"
)

// AO-004 Appendix I: the public client must surface a G41 Direct Operate echo
// with per-point status 0x0C (OUT_OF_RANGE) as ControlOutOfRange — not
// ControlTimeout (the pre-fix default) and not ControlSuccess.

// pubG41EchoTransport echoes exactly one G41V1 record (index/value) with the
// configured per-point command-status octet. Used to drive the public
// client.Operate path.
type pubG41EchoTransport struct {
	lastSeq       uint8
	index         uint16
	value         int32
	commandStatus byte
}

func (t *pubG41EchoTransport) Send(data []byte) error {
	t.lastSeq = extractPubRequestSeq(data)
	return nil
}

func (t *pubG41EchoTransport) SetTimeout(ms int) {}

func (t *pubG41EchoTransport) Receive() ([]byte, error) {
	obj := []byte{
		41, 1, 0x28, 0x01, 0x00,
		byte(t.index), byte(t.index >> 8),
		byte(t.value), byte(t.value >> 8), byte(t.value >> 16), byte(t.value >> 24),
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

// TestPublicOperateOutOfRangeStatus asserts a G41V1 echo with status 0x0C is
// surfaced as ControlOutOfRange (not ControlTimeout, not ControlSuccess), with
// no error (definite remote rejection).
func TestPublicOperateOutOfRangeStatus(t *testing.T) {
	cc := newConnectedClientWithTransport(t, &pubG41EchoTransport{index: 1, value: 2300, commandStatus: 0x0C})
	resp, err := cc.Operate(context.Background(), &types.ControlOutput{
		Group: 41, Variation: 1, Index: 1,
		CommandType: types.DirectOperate,
		Value:       &types.AnalogInt32CommandValue{Value: 2300},
	})
	if err != nil {
		t.Fatalf("Operate error: %v", err)
	}
	if resp.Status != types.ControlOutOfRange {
		t.Fatalf("Status = %v (%d), want ControlOutOfRange (%d)", resp.Status, resp.Status, types.ControlOutOfRange)
	}
	if resp.Status == types.ControlTimeout {
		t.Fatalf("OUT_OF_RANGE must not be reported as ControlTimeout")
	}
	if resp.Status == types.ControlSuccess {
		t.Fatalf("OUT_OF_RANGE must not be reported as ControlSuccess")
	}
}

// TestMapCommandStatusOutOfRange pins the internal-to-public mapping directly.
func TestMapCommandStatusOutOfRange(t *testing.T) {
	if got := mapCommandStatus(master.CommandStatusOutOfRange); got != types.ControlOutOfRange {
		t.Fatalf("mapCommandStatus(OutOfRange) = %v, want ControlOutOfRange", got)
	}
}
