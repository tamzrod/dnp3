package master

import (
	"context"
	"errors"
	"testing"

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

// TestUnsolicitedRejected asserts the unsolicited enable/disable APIs return
// ErrUnsupportedOption (DNP3-030).
func TestUnsolicitedRejected(t *testing.T) {
	cc := newConnectedClientWithTransport(t, &pubStatusEchoTransport{commandStatus: 0})

	if err := cc.EnableUnsolicited(context.Background()); !errors.Is(err, dnp3.ErrUnsupportedOption) {
		t.Fatalf("EnableUnsolicited error = %v, want ErrUnsupportedOption", err)
	}
	if err := cc.DisableUnsolicited(context.Background()); !errors.Is(err, dnp3.ErrUnsupportedOption) {
		t.Fatalf("DisableUnsolicited error = %v, want ErrUnsupportedOption", err)
	}
}
