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

// TestOperateRejectsSBO asserts Operate rejects SelectThenOperate commands with
// ErrUnsupportedOption before any wire traffic (DNP3-030).
func TestOperateRejectsSBO(t *testing.T) {
	cc := newConnectedClientWithTransport(t, &pubStatusEchoTransport{commandStatus: 0})

	_, err := cc.Operate(context.Background(), &types.ControlOutput{
		Group: 12, Variation: 1, Index: 0,
		CommandType: types.SelectThenOperate,
		Value:       &types.BinaryCommandValue{Value: true},
	})
	if err == nil {
		t.Fatal("expected error for select-before-operate, got nil")
	}
	if !errors.Is(err, dnp3.ErrUnsupportedOption) {
		t.Fatalf("error = %v, want ErrUnsupportedOption", err)
	}
}

// TestOperateRejectsDirectOperateNoResponse asserts Operate rejects
// DirectOperateNoResponse commands with ErrUnsupportedOption (DNP3-030).
func TestOperateRejectsDirectOperateNoResponse(t *testing.T) {
	cc := newConnectedClientWithTransport(t, &pubStatusEchoTransport{commandStatus: 0})

	_, err := cc.Operate(context.Background(), &types.ControlOutput{
		Group: 12, Variation: 1, Index: 0,
		CommandType: types.DirectOperateNoResponse,
		Value:       &types.BinaryCommandValue{Value: true},
	})
	if err == nil {
		t.Fatal("expected error for direct-operate-no-response, got nil")
	}
	if !errors.Is(err, dnp3.ErrUnsupportedOption) {
		t.Fatalf("error = %v, want ErrUnsupportedOption", err)
	}
}

// TestOperateAcceptsDirectOperate asserts the supported DirectOperate command
// type is not rejected at the gate (it proceeds to the request path).
func TestOperateAcceptsDirectOperate(t *testing.T) {
	cc := newConnectedClientWithTransport(t, &pubStatusEchoTransport{commandStatus: 0})

	_, err := cc.Operate(context.Background(), &types.ControlOutput{
		Group: 12, Variation: 1, Index: 0,
		CommandType: types.DirectOperate,
		Value:       &types.BinaryCommandValue{Value: true},
	})
	if errors.Is(err, dnp3.ErrUnsupportedOption) {
		t.Fatalf("DirectOperate should be supported, got %v", err)
	}
}

// fnCaptureTransport records application function codes and answers with an
// empty response so enable/disable can complete.
type fnCaptureTransport struct {
	pubStatusEchoTransport
	funcs []uint8
}

func (t *fnCaptureTransport) Send(data []byte) error {
	if f, err := frame.Decode(data); err == nil {
		if frag, err := tl.DecodeFragment(f.Data); err == nil {
			if apdu, err := al.Decode(frag.Data); err == nil {
				t.funcs = append(t.funcs, apdu.FuncCode)
			}
		}
	}
	return t.pubStatusEchoTransport.Send(data)
}

// TestUnsolicitedSendsIEEEFunctionCodes asserts enable/disable are real
// requests on the explicit, non-default experimental event delivery path. The
// published supported profile is unchanged; this test only locks the wire
// function codes 0x14 and 0x15 when the experimental switch is enabled.
func TestUnsolicitedSendsIEEEFunctionCodes(t *testing.T) {
	tr := &fnCaptureTransport{}
	cc := newConnectedExperimentalClientWithTransport(t, tr)
	if err := cc.EnableUnsolicited(context.Background()); err != nil {
		t.Fatalf("EnableUnsolicited: %v", err)
	}
	if err := cc.DisableUnsolicited(context.Background()); err != nil {
		t.Fatalf("DisableUnsolicited: %v", err)
	}
	if len(tr.funcs) != 2 || tr.funcs[0] != 0x14 || tr.funcs[1] != 0x15 {
		t.Fatalf("function codes = %v, want [0x14 0x15]", tr.funcs)
	}
}
