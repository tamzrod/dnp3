package master

import (
	"context"
	"errors"
	"io"
	"testing"

	internalmaster "dnp3/internal/master"
	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/types"
)

// peerCloseTransport accepts Send but returns io.EOF from Receive, simulating a
// peer that closed the TCP connection mid-session (DNP3-031).
type peerCloseTransport struct{}

func (t *peerCloseTransport) Send(data []byte) error   { return nil }
func (t *peerCloseTransport) SetTimeout(ms int)        {}
func (t *peerCloseTransport) Receive() ([]byte, error) { return nil, io.EOF }

// TestReadDetectsTransportDisconnect asserts that when the transport reports a
// peer close (io.EOF) during a Read, the public client transitions to
// Disconnected and the error wraps ErrTransportDisconnected (DNP3-031).
func TestReadDetectsTransportDisconnect(t *testing.T) {
	cc := newConnectedClientWithTransport(t, &peerCloseTransport{})

	if got := cc.State(); got != dnp3.StateConnected {
		t.Fatalf("pre-Read state = %v, want Connected", got)
	}

	_, err := cc.Read(context.Background(), &types.ReadRequest{
		Groups: []types.GroupRequest{{Group: 1, Variation: 0}},
	})
	if err == nil {
		t.Fatal("expected error on peer close, got nil")
	}
	if !errors.Is(err, internalmaster.ErrTransportDisconnected) {
		t.Fatalf("error = %v, want ErrTransportDisconnected in chain", err)
	}

	if got := cc.State(); got != dnp3.StateDisconnected {
		t.Fatalf("post-Read state = %v, want Disconnected", got)
	}
}

// TestOperateDetectsTransportDisconnect asserts the same disconnect transition
// for the Operate path (DNP3-031).
func TestOperateDetectsTransportDisconnect(t *testing.T) {
	cc := newConnectedClientWithTransport(t, &peerCloseTransport{})

	_, err := cc.Operate(context.Background(), &types.ControlOutput{
		Group: 12, Variation: 1, Index: 0,
		CommandType: types.DirectOperate,
		Value:       &types.BinaryCommandValue{Value: true},
	})
	if err == nil {
		t.Fatal("expected error on peer close, got nil")
	}
	if !errors.Is(err, internalmaster.ErrTransportDisconnected) {
		t.Fatalf("error = %v, want ErrTransportDisconnected in chain", err)
	}
	if got := cc.State(); got != dnp3.StateDisconnected {
		t.Fatalf("post-Operate state = %v, want Disconnected", got)
	}
}

// TestReadAfterDisconnectReturnsNotConnected asserts that after a disconnect,
// a subsequent Read fails fast with ErrNotConnected (no wire traffic, no retry
// storm) (DNP3-031).
func TestReadAfterDisconnectReturnsNotConnected(t *testing.T) {
	cc := newConnectedClientWithTransport(t, &peerCloseTransport{})

	_, _ = cc.Read(context.Background(), &types.ReadRequest{
		Groups: []types.GroupRequest{{Group: 1, Variation: 0}},
	})

	_, err := cc.Read(context.Background(), &types.ReadRequest{
		Groups: []types.GroupRequest{{Group: 1, Variation: 0}},
	})
	if !errors.Is(err, dnp3.ErrNotConnected) {
		t.Fatalf("second Read error = %v, want ErrNotConnected", err)
	}
}

